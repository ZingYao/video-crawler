package crawler

import (
	"testing"
	"time"
)

func TestNewCollyBrowser(t *testing.T) {
	config := DefaultConfig()
	browser, err := NewCollyBrowser(config)
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

	browser, err := NewCollyBrowser(config)
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

	if response.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", response.StatusCode)
	}

	if len(response.Body) == 0 {
		t.Error("Response body should not be empty")
	}
}

func TestBrowserPost(t *testing.T) {
	config := DefaultConfig()
	config.Timeout = 10 * time.Second

	browser, err := NewCollyBrowser(config)
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

	if response.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", response.StatusCode)
	}
}

func TestSetRandomUserAgent(t *testing.T) {
	config := DefaultConfig()
	browser, err := NewCollyBrowser(config)
	if err != nil {
		t.Fatalf("Failed to create browser: %v", err)
	}
	defer browser.Close()

	// 设置随机User-Agent
	browser.SetRandomUserAgent()

	// 验证User-Agent已设置
	collyBrowser, ok := browser.(*CollyBrowser)
	if !ok {
		t.Fatal("Failed to cast to CollyBrowser")
	}

	if collyBrowser.config.UserAgent == "" {
		t.Error("User-Agent should not be empty after setting random User-Agent")
	}
}

func TestSetHeaders(t *testing.T) {
	config := DefaultConfig()
	browser, err := NewCollyBrowser(config)
	if err != nil {
		t.Fatalf("Failed to create browser: %v", err)
	}
	defer browser.Close()

	headers := map[string]string{
		"X-Test-Header": "test-value",
		"Accept":        "application/json",
	}

	browser.SetHeaders(headers)

	collyBrowser, ok := browser.(*CollyBrowser)
	if !ok {
		t.Fatal("Failed to cast to CollyBrowser")
	}

	for key, value := range headers {
		if collyBrowser.config.Headers[key] != value {
			t.Errorf("Header %s should be %s, got %s", key, value, collyBrowser.config.Headers[key])
		}
	}
}

func TestSetCookies(t *testing.T) {
	config := DefaultConfig()
	browser, err := NewCollyBrowser(config)
	if err != nil {
		t.Fatalf("Failed to create browser: %v", err)
	}
	defer browser.Close()

	cookies := map[string]string{
		"session_id": "test-session",
		"user_id":    "12345",
	}

	browser.SetCookies(cookies)

	collyBrowser, ok := browser.(*CollyBrowser)
	if !ok {
		t.Fatal("Failed to cast to CollyBrowser")
	}

	for key, value := range cookies {
		if collyBrowser.config.Cookies[key] != value {
			t.Errorf("Cookie %s should be %s, got %s", key, value, collyBrowser.config.Cookies[key])
		}
	}
}
