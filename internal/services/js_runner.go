package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"video-crawler/internal/crawler"
	"video-crawler/internal/entities"
	"video-crawler/internal/jsengine"
)

type JSTestService interface {
	ExecuteScript(ctx context.Context, script string) (<-chan string, error)
	ExecuteAdvancedTest(ctx context.Context, script string, method string, params map[string]interface{}) (*entities.AdvancedTestResult, string, error)
	ExecuteAdvancedTestSSE(ctx context.Context, script string, method string, params map[string]interface{}) (<-chan string, error)
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

	eng := jsengine.New(browser) // 测试服务不需要ctxlog，保持原有行为

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

// ExecuteAdvancedTest 执行高级调试
func (s *jsTestService) ExecuteAdvancedTest(ctx context.Context, script string, method string, params map[string]interface{}) (*entities.AdvancedTestResult, string, error) {
	browser, err := crawler.NewDefaultBrowser()
	if err != nil {
		return nil, "", fmt.Errorf("创建浏览器实例失败: %w", err)
	}
	defer browser.Close()

	if v := ctx.Value(CtxKeyRequestUA); v != nil {
		if ua, ok := v.(string); ok && ua != "" {
			browser.SetUserAgent(ua)
		}
	}

	headers := map[string]string{
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
		"Accept-Encoding": "gzip, deflate, br, zstd",
		"DNT":             "1",
	}
	browser.SetHeaders(headers)

	eng := jsengine.New(browser)

	// 构建测试脚本
	var testScript string
	switch method {
	case "search_video":
		if keyword, ok := params["keyword"].(string); ok {
			testScript = fmt.Sprintf(`
%s

// 执行测试
console.log("[TEST] 执行 search_video 方法");
console.log("[TEST] 参数: %s");
const result = search_video("%s");
console.log("[TEST] 结果:", JSON.stringify(result, null, 2));
return {data: result, err: null};
`, script, keyword, keyword)
		}
	case "get_video_detail":
		if videoURL, ok := params["video_url"].(string); ok {
			testScript = fmt.Sprintf(`
%s

// 执行测试
console.log("[TEST] 执行 get_video_detail 方法");
console.log("[TEST] 参数: %s");
const result = get_video_detail("%s");
console.log("[TEST] 结果:", JSON.stringify(result, null, 2));
return {data: result, err: null};
`, script, videoURL, videoURL)
		}
	case "get_play_video_detail":
		if videoURL, ok := params["video_url"].(string); ok {
			testScript = fmt.Sprintf(`
%s

// 执行测试
console.log("[TEST] 执行 get_play_video_detail 方法");
console.log("[TEST] 参数: %s");
const result = get_play_video_detail("%s");
console.log("[TEST] 结果:", JSON.stringify(result, null, 2));
return {data: result, err: null};
`, script, videoURL, videoURL)
		}
	default:
		return nil, "", fmt.Errorf("不支持的方法: %s", method)
	}

	// 收集console输出
	var consoleOutput []string
	eng.SetLogSink(func(line string) {
		consoleOutput = append(consoleOutput, line)
	})

	// 执行脚本
	result, err := eng.ExecuteWrapped(testScript)
	if err != nil {
		return nil, "", fmt.Errorf("脚本执行失败: %w", err)
	}

	// 获取原始结果
	var originalResult interface{}
	if result != nil {
		if data, ok := result["data"]; ok {
			originalResult = data
		}
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

// ExecuteAdvancedTestSSE 执行高级调试(SSE)
func (s *jsTestService) ExecuteAdvancedTestSSE(ctx context.Context, script string, method string, params map[string]interface{}) (<-chan string, error) {
	browser, err := crawler.NewDefaultBrowser()
	if err != nil {
		return nil, fmt.Errorf("创建浏览器实例失败: %w", err)
	}

	if v := ctx.Value(CtxKeyRequestUA); v != nil {
		if ua, ok := v.(string); ok && ua != "" {
			browser.SetUserAgent(ua)
		}
	}

	headers := map[string]string{
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
		"Accept-Encoding": "gzip, deflate, br, zstd",
		"DNT":             "1",
	}
	browser.SetHeaders(headers)

	eng := jsengine.New(browser)

	// 构建测试脚本
	var testScript string
	switch method {
	case "search_video":
		if keyword, ok := params["keyword"].(string); ok {
			testScript = fmt.Sprintf(`
%s

// 执行测试
console.log("[TEST] 执行 search_video 方法");
console.log("[TEST] 参数: %s");
const result = search_video("%s");
console.log("[TEST] 结果:", JSON.stringify(result, null, 2));
return {data: result, err: null};
`, script, keyword, keyword)
		}
	case "get_video_detail":
		if videoURL, ok := params["video_url"].(string); ok {
			testScript = fmt.Sprintf(`
%s

// 执行测试
console.log("[TEST] 执行 get_video_detail 方法");
console.log("[TEST] 参数: %s");
const result = get_video_detail("%s");
console.log("[TEST] 结果:", JSON.stringify(result, null, 2));
return {data: result, err: null};
`, script, videoURL, videoURL)
		}
	case "get_play_video_detail":
		if videoURL, ok := params["video_url"].(string); ok {
			testScript = fmt.Sprintf(`
%s

// 执行测试
console.log("[TEST] 执行 get_play_video_detail 方法");
console.log("[TEST] 参数: %s");
const result = get_play_video_detail("%s");
console.log("[TEST] 结果:", JSON.stringify(result, null, 2));
return {data: result, err: null};
`, script, videoURL, videoURL)
		}
	default:
		return nil, fmt.Errorf("不支持的方法: %s", method)
	}

	// 创建输出通道
	out := make(chan string, 200)

	// 将 console.* 输出回流到前端
	eng.SetLogSink(func(line string) {
		select {
		case out <- fmt.Sprintf("event: log\ndata: {\"message\":\"%s\"}\n\n", jsonEscape(line)):
		default:
			// 避免阻塞：丢弃超量日志
		}
	})

	go func() {
		defer close(out)
		defer browser.Close()

		out <- fmt.Sprintf("event: log\ndata: {\"message\":\"[INFO] 开始执行JS高级调试...\"}\n\n")

		// 直接执行脚本，保留调用方在结尾 return 的 {data, err}
		m, err := eng.ExecuteWrapped(testScript)
		if err != nil {
			out <- fmt.Sprintf("event: error\ndata: {\"message\":\"%s\"}\n\n", jsonEscape(err.Error()))
			return
		}

		// 获取原始结果
		var originalResult interface{}
		if m != nil {
			if data, ok := m["data"]; ok {
				originalResult = data
			}
		}

		// 根据方法类型验证和转换结果
		var convertedResult interface{}
		if originalResult != nil {
			switch method {
			case "search_video":
				if filtered, err := entities.FilterSearchVideoResult(originalResult); err == nil {
					convertedResult = filtered
				} else {
					convertedResult = originalResult
				}
			case "get_video_detail":
				if filtered, err := entities.FilterVideoDetailResult(originalResult); err == nil {
					convertedResult = filtered
				} else {
					convertedResult = originalResult
				}
			case "get_play_video_detail":
				if filtered, err := entities.FilterPlayVideoDetailResult(originalResult); err == nil {
					convertedResult = filtered
				} else {
					convertedResult = originalResult
				}
			default:
				convertedResult = originalResult
			}
		}

		// 发送结果事件
		resultData := map[string]interface{}{
			"original":  originalResult,
			"converted": convertedResult,
		}
		resultJSON, _ := json.Marshal(resultData)
		out <- fmt.Sprintf("event: result\ndata: %s\n\n", string(resultJSON))

		out <- fmt.Sprintf("event: log\ndata: {\"message\":\"[INFO] JS高级调试执行完成\"}\n\n")
	}()

	return out, nil
}
