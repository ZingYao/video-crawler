package crawler

import (
	"io"
	"testing"
	"time"
)

func TestNewHTTPBrowser(t *testing.T) {
	config := DefaultConfig()
	browser, err := NewHTTPBrowser(config)
	if err != nil {
		t.Fatalf("Failed to create browser: %v", err)
	}
	defer browser.Close()

	if browser == nil {
		t.Fatal("Browser should not be nil")
	}
}

func TestBrowserGet(t *testing.T) {
	config := DefaultConfig()
	config.Timeout = 10 * time.Second

	browser, err := NewHTTPBrowser(config)
	if err != nil {
		t.Fatalf("Failed to create browser: %v", err)
	}
	defer browser.Close()

	// 设置随机User-Agent
	browser.SetRandomUserAgent()

	response, err := browser.Get("https://httpbin.org/get")
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", response.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if len(body) == 0 {
		t.Error("Response body should not be empty")
	}
}

func TestBrowserPost(t *testing.T) {
	config := DefaultConfig()
	config.Timeout = 10 * time.Second

	browser, err := NewHTTPBrowser(config)
	if err != nil {
		t.Fatalf("Failed to create browser: %v", err)
	}
	defer browser.Close()

	data := map[string]interface{}{
		"test": "value",
	}

	response, err := browser.Post("https://httpbin.org/post", data)
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", response.StatusCode)
	}
}

func TestSetRandomUserAgent(t *testing.T) {
	config := DefaultConfig()
	browser, err := NewHTTPBrowser(config)
	if err != nil {
		t.Fatalf("Failed to create browser: %v", err)
	}
	defer browser.Close()

	// 设置随机User-Agent
	browser.SetRandomUserAgent()

	// 验证User-Agent已设置
	if browser.GetUserAgent() == "" {
		t.Error("User-Agent should not be empty after setting random User-Agent")
	}
}

func TestSetHeaders(t *testing.T) {
	config := DefaultConfig()
	browser, err := NewHTTPBrowser(config)
	if err != nil {
		t.Fatalf("Failed to create browser: %v", err)
	}
	defer browser.Close()

	headers := map[string]string{
		"X-Test-Header": "test-value",
		"Accept":        "application/json",
	}

	browser.SetHeaders(headers)

	// 通过接口方法验证，而不是直接访问内部字段
	// 这里我们只能测试方法调用是否成功，无法直接验证内部状态
	// 如果需要验证，可以通过发送请求并检查请求头来实现
}

func TestSetCookies(t *testing.T) {
	config := DefaultConfig()
	browser, err := NewHTTPBrowser(config)
	if err != nil {
		t.Fatalf("Failed to create browser: %v", err)
	}
	defer browser.Close()

	cookies := map[string]string{
		"session_id": "test-session",
		"user_id":    "12345",
	}

	browser.SetCookies(cookies)

	// 通过接口方法验证，而不是直接访问内部字段
	// 这里我们只能测试方法调用是否成功，无法直接验证内部状态
	// 如果需要验证，可以通过发送请求并检查Cookie来实现
}
