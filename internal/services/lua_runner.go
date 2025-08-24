package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"video-crawler/internal/crawler"
	"video-crawler/internal/entities"
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
	// ExecuteAdvancedTest 执行高级调试
	ExecuteAdvancedTest(ctx context.Context, script string, method string, params map[string]interface{}) (*entities.AdvancedTestResult, string, error)
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

	// 设置更完整的真实浏览器请求头，确保与测试脚本一致
	headers := map[string]string{
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Accept-Language":           "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
		"Accept-Encoding":           "gzip, deflate, br, zstd",
		"Cache-Control":             "max-age=0",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1",
		"Sec-Fetch-Dest":            "document",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-Site":            "none",
		"Sec-Fetch-User":            "?1",
		"sec-ch-ua":                 `"Not;A=Brand";v="99", "Microsoft Edge";v="139", "Chromium";v="139"`,
		"sec-ch-ua-mobile":          "?0",
		"sec-ch-ua-platform":        `"macOS"`,
		"DNT":                       "1",
	}
	browser.SetHeaders(headers)

	// 创建Lua引擎
	engine := lua.NewLuaEngine(browser) // 测试服务不需要ctxlog，保持原有行为

	// 创建输出通道
	outputChan := make(chan string, 100)

	formatMsg := func(level string, msg string) string {
		return fmt.Sprintf("[%s][%s] %s", level, time.Now().Format(time.RFC3339Nano), msg)
	}

	// 单协程串行写入，严格保证顺序
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("Lua脚本执行panic: %v", r)
				outputChan <- formatMsg("ERROR", fmt.Sprintf("脚本执行panic: %v", r))
			}
			// 最终清理
			_ = browser.Close()
			close(outputChan)
		}()

		// 先发送开始消息
		outputChan <- formatMsg("INFO", "开始执行Lua脚本...")

		// 后台执行脚本；主循环继续串行转发输出
		done := make(chan struct{})
		var ret map[string]interface{}
		var execErr error
		go func() {
			ret, execErr = engine.Execute(script)
			close(done)
		}()

		engOut := engine.GetOutputChannel()
		for {
			select {
			case msg, ok := <-engOut:
				if !ok {
					engOut = nil
				} else {
					outputChan <- msg
				}
			case <-done:
				// 关闭引擎，令输出通道完结，然后将剩余日志全部转发，最后输出结果
				engine.Close()
				if engOut != nil {
					for msg := range engOut {
						outputChan <- msg
					}
					engOut = nil
				}
				if execErr != nil {
					outputChan <- formatMsg("ERROR", fmt.Sprintf("脚本执行失败: %v", execErr))
				} else if ret != nil {
					if data, mErr := json.MarshalIndent(ret, "", "  "); mErr == nil {
						outputChan <- fmt.Sprintf("[RESULT] %s", string(data))
					}
				}
				outputChan <- formatMsg("INFO", "Lua脚本执行完成")
				return
			case <-ctx.Done():
				outputChan <- formatMsg("ERROR", "执行已取消")
				return
			}
		}
	}()

	return outputChan, nil
}

// ExecuteAdvancedTest 执行高级调试
func (s *luaTestService) ExecuteAdvancedTest(ctx context.Context, script string, method string, params map[string]interface{}) (*entities.AdvancedTestResult, string, error) {
	// 创建浏览器实例
	browser, err := crawler.NewDefaultBrowser()
	if err != nil {
		return nil, "", fmt.Errorf("创建浏览器实例失败: %w", err)
	}
	defer browser.Close()

	// 如果上下文中携带了请求UA，优先使用它
	if v := ctx.Value(CtxKeyRequestUA); v != nil {
		if ua, ok := v.(string); ok && ua != "" {
			browser.SetUserAgent(ua)
		}
	}

	// 设置请求头
	headers := map[string]string{
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Accept-Language":           "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
		"Accept-Encoding":           "gzip, deflate, br, zstd",
		"Cache-Control":             "max-age=0",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1",
		"Sec-Fetch-Dest":            "document",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-Site":            "none",
		"Sec-Fetch-User":            "?1",
		"sec-ch-ua":                 `"Not;A=Brand";v="99", "Microsoft Edge";v="139", "Chromium";v="139"`,
		"sec-ch-ua-mobile":          "?0",
		"sec-ch-ua-platform":        `"macOS"`,
		"DNT":                       "1",
	}
	browser.SetHeaders(headers)

	// 创建Lua引擎
	engine := lua.NewLuaEngine(browser)

	// 构建测试脚本
	var testScript string
	switch method {
	case "search_video":
		if keyword, ok := params["keyword"].(string); ok {
			testScript = fmt.Sprintf(`
%s

-- 执行测试
local result = search_video("%s")
print("[TEST] 执行 search_video 方法")
print("[TEST] 参数: %s")
print("[TEST] 结果: " .. json.encode(result))
return {data = result, err = nil}
`, script, keyword, keyword)
		}
	case "get_video_detail":
		if videoURL, ok := params["video_url"].(string); ok {
			testScript = fmt.Sprintf(`
%s

-- 执行测试
local result = get_video_detail("%s")
print("[TEST] 执行 get_video_detail 方法")
print("[TEST] 参数: %s")
print("[TEST] 结果: " .. json.encode(result))
return {data = result, err = nil}
`, script, videoURL, videoURL)
		}
	case "get_play_video_detail":
		if videoURL, ok := params["video_url"].(string); ok {
			testScript = fmt.Sprintf(`
%s

-- 执行测试
local result = get_play_video_detail("%s")
print("[TEST] 执行 get_play_video_detail 方法")
print("[TEST] 参数: %s")
print("[TEST] 结果: " .. json.encode(result))
return {data = result, err = nil}
`, script, videoURL, videoURL)
		}
	default:
		return nil, "", fmt.Errorf("不支持的方法: %s", method)
	}

	// 收集console输出
	var consoleOutput []string
	outputChan := make(chan string, 100)
	go func() {
		for msg := range outputChan {
			consoleOutput = append(consoleOutput, msg)
		}
	}()

	// 执行脚本
	result, err := engine.Execute(testScript)
	if err != nil {
		return nil, "", fmt.Errorf("脚本执行失败: %w", err)
	}

	// 获取原始结果
	var originalResult interface{}
	if data, ok := result["data"]; ok {
		originalResult = data
	}

	// 转换结果
	var convertedResult interface{}
	if originalResult != nil {
		// 这里可以根据方法类型进行结构体转换
		// 暂时返回原始结果
		convertedResult = originalResult
	}

	// 构建console输出字符串
	consoleStr := ""
	for _, line := range consoleOutput {
		consoleStr += line + "\n"
	}

	return &entities.AdvancedTestResult{
		Original:  originalResult,
		Converted: convertedResult,
	}, consoleStr, nil
}
