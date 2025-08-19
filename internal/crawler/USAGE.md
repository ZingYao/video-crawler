# 浏览器请求接口使用说明

## 快速开始

```go
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
```

## 主要功能

- ✅ 随机User-Agent生成
- ✅ 自定义请求头和Cookie
- ✅ GET/POST请求支持
- ✅ 超时和重试机制
- ✅ 代理支持
- ✅ 自动重定向处理

## 运行演示

```bash
go run cmd/simple-demo/main.go
```

## 依赖库

- `github.com/gocolly/colly/v2` - 爬虫框架
- `github.com/lib4u/fake-useragent` - 随机User-Agent
