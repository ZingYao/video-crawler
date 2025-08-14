# 浏览器请求接口

这是一个基于 Colly 框架的浏览器请求接口，支持完全真实的浏览器请求，包括随机 User-Agent、自定义请求头、Cookie 管理等功能。

## 功能特性

- 🚀 基于 Colly 框架的高性能爬虫
- 🎭 随机 User-Agent 生成
- 🔧 自定义请求头和 Cookie 设置
- ⏱️ 超时和重试机制
- 🌐 代理支持
- 📡 GET 和 POST 请求支持
- 🔄 自动重定向处理

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "video-crawler/internal/crawler"
)

func main() {
    // 创建默认浏览器实例
    browser, err := crawler.NewDefaultBrowser()
    if err != nil {
        log.Fatal("创建浏览器失败:", err)
    }
    defer browser.Close()

    // 设置随机User-Agent
    browser.SetRandomUserAgent()

    // 发送GET请求
    response, err := browser.Get("https://httpbin.org/get")
    if err != nil {
        log.Fatal("请求失败:", err)
    }

    fmt.Printf("状态码: %d\n", response.StatusCode)
    fmt.Printf("响应体长度: %d 字节\n", len(response.Body))
}
```

### 高级配置

```go
// 创建自定义配置
config := &crawler.BrowserConfig{
    Timeout:         60 * time.Second,
    UserAgent:       "Custom User Agent",
    Proxy:           "http://proxy.example.com:8080",
    Headers:         make(map[string]string),
    Cookies:         make(map[string]string),
    MaxRetries:      5,
    RetryDelay:      2 * time.Second,
    FollowRedirects: true,
}

// 创建浏览器实例
browser, err := crawler.NewBrowser(crawler.CollyBrowserType, config)
if err != nil {
    log.Fatal("创建浏览器失败:", err)
}
defer browser.Close()

// 设置真实浏览器请求头
headers := map[string]string{
    "Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
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

// 设置Cookie
cookies := map[string]string{
    "session_id": "your-session-id",
    "user_id":    "12345",
}
browser.SetCookies(cookies)

// 设置超时
browser.SetTimeout(30 * time.Second)

// 发送POST请求
postData := map[string]interface{}{
    "name":    "测试用户",
    "message": "Hello World",
    "timestamp": time.Now().Unix(),
}

response, err := browser.Post("https://httpbin.org/post", postData)
if err != nil {
    log.Fatal("POST请求失败:", err)
}

fmt.Printf("状态码: %d\n", response.StatusCode)
```

## API 接口

### BrowserRequest 接口

```go
type BrowserRequest interface {
    // Get 发送GET请求
    Get(url string) (*Response, error)
    
    // Post 发送POST请求
    Post(url string, data map[string]interface{}) (*Response, error)
    
    // SetHeaders 设置请求头
    SetHeaders(headers map[string]string)
    
    // SetCookies 设置Cookie
    SetCookies(cookies map[string]string)
    
    // SetTimeout 设置超时时间
    SetTimeout(timeout time.Duration)
    
    // SetProxy 设置代理
    SetProxy(proxy string)
    
    // SetUserAgent 设置User-Agent
    SetUserAgent(userAgent string)
    
    // SetRandomUserAgent 设置随机User-Agent
    SetRandomUserAgent()
    
    // Close 关闭浏览器实例
    Close() error
}
```

### Response 结构

```go
type Response struct {
    StatusCode int               // HTTP状态码
    Headers    map[string]string // 响应头
    Body       []byte            // 响应体
    URL        string            // 请求URL
    Cookies    map[string]string // 响应Cookie
}
```

### BrowserConfig 配置

```go
type BrowserConfig struct {
    Timeout          time.Duration // 超时时间
    UserAgent        string        // User-Agent
    Proxy            string        // 代理地址
    Headers          map[string]string // 请求头
    Cookies          map[string]string // Cookie
    MaxRetries       int           // 最大重试次数
    RetryDelay       time.Duration // 重试延迟
    FollowRedirects  bool          // 是否跟随重定向
}
```

## 工厂函数

### NewDefaultBrowser()

创建默认配置的浏览器实例，自动设置随机User-Agent。

### NewBrowser(browserType, config)

创建指定类型和配置的浏览器实例。

```go
browser, err := crawler.NewBrowser(crawler.CollyBrowserType, config)
```

## 随机User-Agent

使用 `github.com/EDDYCJY/fake-useragent` 库生成真实的浏览器User-Agent：

```go
browser.SetRandomUserAgent()
```

生成的User-Agent包括：
- Chrome
- Firefox
- Safari
- Edge
- 移动端浏览器

## 真实浏览器请求头

为了模拟真实浏览器，建议设置以下请求头：

```go
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
```

## 错误处理

```go
response, err := browser.Get("https://example.com")
if err != nil {
    // 处理错误
    log.Printf("请求失败: %v", err)
    return
}

if response.StatusCode != 200 {
    log.Printf("HTTP错误: %d", response.StatusCode)
    return
}
```

## 性能优化

1. **复用浏览器实例**：避免频繁创建和销毁浏览器实例
2. **设置合理的超时**：根据目标网站响应时间设置合适的超时
3. **使用代理池**：轮换使用多个代理避免IP被封
4. **控制请求频率**：添加请求间隔避免过于频繁的请求

## 示例程序

运行演示程序：

```bash
go run cmd/simple-demo/main.go
```

## 依赖库

- `github.com/gocolly/colly/v2` - 爬虫框架
- `github.com/EDDYCJY/fake-useragent` - 随机User-Agent生成
- `github.com/gocolly/colly/v2/extensions` - Colly扩展功能

## 注意事项

1. 请遵守网站的robots.txt规则
2. 合理控制请求频率，避免对目标网站造成压力
3. 在生产环境中使用代理和更复杂的反爬虫策略
4. 定期更新User-Agent库以获取最新的浏览器标识
