# Lua请求和HTML解析引擎

## 快速开始

```go
// 创建浏览器和Lua引擎
browser, _ := crawler.NewDefaultBrowser()
engine := lua.NewLuaEngine(browser)
defer engine.Close()

// 执行Lua脚本
script := `
    response, err = http_get("https://httpbin.org/get")
    if err then
        print("请求失败:", err)
    else
        print("状态码:", response.status_code)
    end
`
engine.Execute(script)
```

## 主要功能

- ✅ HTTP请求 (GET/POST)
- ✅ 随机User-Agent
- ✅ 自定义请求头和Cookie
- ✅ HTML解析 (基于goquery)
- ✅ CSS选择器
- ✅ 属性提取

## Lua API

### HTTP请求
- `http_get(url)` - GET请求
- `http_post(url, data)` - POST请求
- `set_headers(headers)` - 设置请求头
- `set_cookies(cookies)` - 设置Cookie
- `set_random_user_agent()` - 随机User-Agent

### HTML解析
- `parse_html(html)` - 解析HTML
- `select(doc, selector)` - 选择多个元素
- `select_one(doc, selector)` - 选择单个元素
- `text(element)` - 获取文本
- `html(element)` - 获取HTML
- `attr(element, name)` - 获取属性

## 运行演示

```bash
go run cmd/lua-demo/main.go
```
