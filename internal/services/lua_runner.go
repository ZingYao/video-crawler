package services

import (
	"context"
	"fmt"
	"sync"
	"time"
	"video-crawler/internal/crawler"
	"video-crawler/internal/lua"

	"github.com/sirupsen/logrus"
)

// CtxKey 用于在 context 中存取附加信息的 key 类型
type CtxKey string

// CtxKeyRequestUA 上下文中存放前端请求 User-Agent 的 key
const CtxKeyRequestUA CtxKey = "request_ua"

type LuaTestService interface {
	// ExecuteScript 执行Lua脚本并返回流式输出
	ExecuteScript(ctx context.Context, script string) (<-chan string, error)
}

type luaTestService struct{}

func NewLuaTestService() LuaTestService {
	return &luaTestService{}
}

func (s *luaTestService) ExecuteScript(ctx context.Context, script string) (<-chan string, error) {
	// 创建浏览器实例
	browser, err := crawler.NewDefaultBrowser()
	if err != nil {
		return nil, fmt.Errorf("创建浏览器实例失败: %w", err)
	}

	// 如果上下文中携带了请求UA，优先使用它
	if v := ctx.Value(CtxKeyRequestUA); v != nil {
		if ua, ok := v.(string); ok && ua != "" {
			browser.SetUserAgent(ua)
		}
	}

	// 创建Lua引擎
	engine := lua.NewLuaEngine(browser)

	// 创建输出通道
	outputChan := make(chan string, 100)

	// 统一关闭
	var closeOnce sync.Once
	closeAll := func() {
		defer func() {
			_ = browser.Close()
		}()
		engine.Close()
	}

	formatMsg := func(level string, msg string) string {
		return fmt.Sprintf("[%s][%s] %s", level, time.Now().Format(time.RFC3339Nano), msg)
	}

	// 在goroutine中执行脚本
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("Lua脚本执行panic: %v", r)
				select {
				case outputChan <- formatMsg("ERROR", fmt.Sprintf("脚本执行panic: %v", r)):
				default:
				}
			}
			// 脚本结束后关闭引擎与浏览器
			closeOnce.Do(closeAll)
		}()

		// 发送开始执行的消息
		select {
		case outputChan <- formatMsg("INFO", "开始执行Lua脚本..."):
		default:
		}

		// 执行脚本
		if err := engine.Execute(script); err != nil {
			select {
			case outputChan <- formatMsg("ERROR", fmt.Sprintf("脚本执行失败: %v", err)):
			default:
			}
			return
		}

		// 发送执行完成的消息
		select {
		case outputChan <- formatMsg("INFO", "Lua脚本执行完成"):
		default:
		}
	}()

	// 在另一个goroutine中转发引擎输出
	go func() {
		defer close(outputChan)

		// 监听引擎输出通道
		for {
			select {
			case msg, ok := <-engine.GetOutputChannel():
				if !ok {
					return
				}
				select {
				case outputChan <- msg:
				default:
				}
			case <-ctx.Done():
				// 客户端中断，关闭引擎与浏览器
				closeOnce.Do(closeAll)
				return
			}
		}
	}()

	return outputChan, nil
}
