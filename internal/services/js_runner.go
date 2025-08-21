package services

import (
	"context"
	"fmt"
	"time"
	"video-crawler/internal/crawler"
	"video-crawler/internal/jsengine"
)

type JSTestService interface {
	ExecuteScript(ctx context.Context, script string) (<-chan string, error)
}

type jsTestService struct{}

func NewJSTestService() JSTestService { return &jsTestService{} }

func (s *jsTestService) ExecuteScript(ctx context.Context, script string) (<-chan string, error) {
	browser, err := crawler.NewDefaultBrowser()
	if err != nil {
		return nil, fmt.Errorf("创建浏览器实例失败: %w", err)
	}
	if v := ctx.Value(CtxKeyRequestUA); v != nil {
		if ua, ok := v.(string); ok && ua != "" {
			browser.SetUserAgent(ua)
		}
	}
	// 常用头
	headers := map[string]string{
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
		"Accept-Encoding": "gzip, deflate, br, zstd",
		"DNT":             "1",
	}
	browser.SetHeaders(headers)

	eng := jsengine.New(browser)

	out := make(chan string, 200)
	// 将 console.* 输出回流到前端
	eng.SetLogSink(func(line string) {
		select {
		case out <- line:
		default:
			// 避免阻塞：丢弃超量日志
		}
	})

	go func() {
		defer close(out)
		defer browser.Close()
		out <- fmt.Sprintf("[INFO][%s] 开始执行JS脚本...", time.Now().Format(time.RFC3339Nano))
		// 直接执行脚本，保留调用方在结尾 return 的 {data, err}
		m, err := eng.ExecuteWrapped(script)
		if err != nil {
			out <- fmt.Sprintf("[ERROR] %v", err)
			out <- "[END] 脚本执行结束"
			return
		}
		// 输出结果
		if m != nil {
			out <- fmt.Sprintf("[RESULT] %v", m)
		}
		out <- "[INFO] JS脚本执行完成"
		out <- "[END] 脚本执行结束"
	}()
	return out, nil
}
