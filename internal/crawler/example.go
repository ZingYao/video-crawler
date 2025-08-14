package crawler

import (
	"fmt"
	"log"
	"time"
)

// ExampleUsage 使用示例
func ExampleUsage() {
	// 创建默认配置的浏览器
	browser, err := NewDefaultBrowser()
	if err != nil {
		log.Fatal("Failed to create browser:", err)
	}
	defer browser.Close()

	// 设置随机User-Agent
	browser.SetRandomUserAgent()

	// 设置自定义请求头
	headers := map[string]string{
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		"Accept-Language":           "zh-CN,zh;q=0.8,en-US;q=0.5,en;q=0.3",
		"Accept-Encoding":           "gzip, deflate",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1",
	}
	browser.SetHeaders(headers)

	// 设置Cookie
	cookies := map[string]string{
		"session_id": "example_session",
		"user_id":    "12345",
	}
	browser.SetCookies(cookies)

	// 设置超时
	browser.SetTimeout(30 * time.Second)

	// 发送GET请求
	fmt.Println("Sending GET request...")
	response, err := browser.Get("https://httpbin.org/get")
	if err != nil {
		log.Fatal("GET request failed:", err)
	}

	fmt.Printf("Status Code: %d\n", response.StatusCode)
	fmt.Printf("Response Body Length: %d\n", len(response.Body))
	fmt.Printf("Response URL: %s\n", response.URL)

	// 发送POST请求
	fmt.Println("\nSending POST request...")
	postData := map[string]interface{}{
		"name":    "test",
		"message": "Hello World",
	}
	postResponse, err := browser.Post("https://httpbin.org/post", postData)
	if err != nil {
		log.Fatal("POST request failed:", err)
	}

	fmt.Printf("Status Code: %d\n", postResponse.StatusCode)
	fmt.Printf("Response Body Length: %d\n", len(postResponse.Body))
}

// ExampleWithCustomConfig 使用自定义配置的示例
func ExampleWithCustomConfig() {
	// 创建自定义配置
	config := &BrowserConfig{
		Timeout:         60 * time.Second,
		UserAgent:       "Custom User Agent",
		Proxy:           "", // 可以设置代理
		Headers:         make(map[string]string),
		Cookies:         make(map[string]string),
		MaxRetries:      5,
		RetryDelay:      2 * time.Second,
		FollowRedirects: true,
	}

	// 创建浏览器实例
	browser, err := NewBrowser(CollyBrowserType, config)
	if err != nil {
		log.Fatal("Failed to create browser:", err)
	}
	defer browser.Close()

	// 设置随机User-Agent
	browser.SetRandomUserAgent()

	// 发送请求
	response, err := browser.Get("https://httpbin.org/user-agent")
	if err != nil {
		log.Fatal("Request failed:", err)
	}

	fmt.Printf("User-Agent: %s\n", string(response.Body))
}
