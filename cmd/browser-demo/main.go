package main

import (
	"fmt"
	"log"
	"time"

	"video-crawler/internal/crawler"
)

func main() {
	fmt.Println("=== 浏览器请求演示 ===")

	// 创建默认浏览器实例
	browser, err := crawler.NewDefaultBrowser()
	if err != nil {
		log.Fatal("创建浏览器失败:", err)
	}
	defer browser.Close()

	// 设置随机User-Agent
	browser.SetRandomUserAgent()

	// 设置真实浏览器请求头
	headers := map[string]string{
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Accept-Language":           "zh-CN,zh;q=0.9,en;q=0.8",
		"Accept-Encoding":           "gzip, deflate, br",
		"Cache-Control":             "max-age=0",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1",
		"Sec-Fetch-Dest":            "document",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-Site":            "none",
		"Sec-Fetch-User":            "?1",
	}
	browser.SetHeaders(headers)

	// 设置超时
	browser.SetTimeout(30 * time.Second)

	fmt.Println("1. 测试GET请求...")
	response, err := browser.Get("https://httpbin.org/get")
	if err != nil {
		log.Fatal("GET请求失败:", err)
	}

	fmt.Printf("状态码: %d\n", response.StatusCode)
	fmt.Printf("响应体长度: %d 字节\n", len(response.Body))
	fmt.Printf("请求URL: %s\n", response.URL)

	// 显示响应头
	fmt.Println("\n响应头:")
	for key, value := range response.Headers {
		fmt.Printf("  %s: %s\n", key, value)
	}

	fmt.Println("\n2. 测试POST请求...")
	postData := map[string]interface{}{
		"name":      "测试用户",
		"message":   "Hello from Go crawler!",
		"timestamp": time.Now().Unix(),
	}

	postResponse, err := browser.Post("https://httpbin.org/post", postData)
	if err != nil {
		log.Fatal("POST请求失败:", err)
	}

	fmt.Printf("状态码: %d\n", postResponse.StatusCode)
	fmt.Printf("响应体长度: %d 字节\n", len(postResponse.Body))

	fmt.Println("\n3. 测试User-Agent...")
	uaResponse, err := browser.Get("https://httpbin.org/user-agent")
	if err != nil {
		log.Fatal("User-Agent请求失败:", err)
	}

	fmt.Printf("当前User-Agent: %s\n", string(uaResponse.Body))

	fmt.Println("\n=== 演示完成 ===")
}
