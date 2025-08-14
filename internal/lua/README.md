# Lua请求和HTML解析引擎

这是一个基于Lua的请求和HTML解析引擎，将Go的浏览器请求工具和goquery的select能力注入到Lua中，让您可以用Lua脚本进行网页爬取和解析。

## 功能特性

- 🚀 基于Gopher-Lua的高性能Lua引擎
- 🌐 完整的HTTP请求支持（GET/POST）
- 🎭 随机User-Agent生成
- 🔧 自定义请求头和Cookie设置
- 📄 强大的HTML解析能力（基于goquery）
- 🎯 CSS选择器支持
- 📝 灵活的Lua脚本编写

## 快速开始

### 基本使用

```go
package main

import (
    "log"
    "video-crawler/internal/crawler"
    "video-crawler/internal/lua"
)

func main() {
    // 创建浏览器实例
    browser, err := crawler.NewDefaultBrowser()
    if err != nil {
        log.Fatal("创建浏览器失败:", err)
    }
    defer browser.Close()

    // 创建Lua引擎
    engine := lua.NewLuaEngine(browser)
    defer engine.Close()

    // 执行Lua脚本
    script := `
        -- 发送GET请求
        response, err = http_get("https://httpbin.org/get")
        if err then
            print("请求失败:", err)
        else
            print("状态码:", response.status_code)
        end
    `
    
    if err := engine.Execute(script); err != nil {
        log.Fatal("执行脚本失败:", err)
    }
}
```

### 执行Lua文件

```go
// 执行Lua文件
if err := engine.ExecuteFile("scripts/example.lua"); err != nil {
    log.Fatal("执行文件失败:", err)
}
```

## Lua API 参考

### HTTP请求函数

#### `http_get(url)`
发送GET请求
```lua
response, err = http_get("https://example.com")
if err then
    print("请求失败:", err)
else
    print("状态码:", response.status_code)
    print("响应体:", response.body)
    print("URL:", response.url)
    -- 访问响应头
    print("Content-Type:", response.headers["Content-Type"])
end
```

#### `http_post(url, data)`
发送POST请求
```lua
data = {
    name = "测试用户",
    message = "Hello World",
    number = 123
}
response, err = http_post("https://httpbin.org/post", data)
```

#### `set_headers(headers)`
设置请求头
```lua
set_headers({
    Accept = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
    Accept_Language = "zh-CN,zh;q=0.9,en;q=0.8",
    User_Agent = "Custom User Agent"
})
```

#### `set_cookies(cookies)`
设置Cookie
```lua
set_cookies({
    session_id = "abc123",
    user_id = "456"
})
```

#### `set_user_agent(user_agent)`
设置User-Agent
```lua
set_user_agent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
```

#### `set_random_user_agent()`
设置随机User-Agent
```lua
set_random_user_agent()
```

### HTML解析函数

#### `parse_html(html)`
解析HTML字符串
```lua
doc, err = parse_html(response.body)
if err then
    print("解析失败:", err)
end
```

#### `select(doc, selector)`
选择多个元素
```lua
elements, err = select(doc, "div.item")
if err then
    print("选择失败:", err)
else
    print("找到", #elements, "个元素")
    for i, element in ipairs(elements) do
        element_text = text(element)
        print("元素", i, ":", element_text)
    end
end
```

#### `select_one(doc, selector)`
选择单个元素
```lua
element, err = select_one(doc, "h1.title")
if err then
    print("选择失败:", err)
else
    title = text(element)
    print("标题:", title)
end
```

#### `text(element)`
获取元素文本内容
```lua
element_text = text(element)
```

#### `html(element)`
获取元素HTML内容
```lua
element_html = html(element)
```

#### `attr(element, attribute_name)`
获取元素属性值
```lua
href, err = attr(link, "href")
if err then
    print("获取属性失败:", err)
else
    print("链接地址:", href)
end
```

### 工具函数

#### `print(...)`
打印输出
```lua
print("Hello", "World", 123)
```

#### `log(...)`
带标签的日志输出
```lua
log("调试信息:", "请求完成")
```

## 完整示例

### 爬取网页并解析

```lua
-- 设置随机User-Agent
set_random_user_agent()

-- 设置请求头
set_headers({
    Accept = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
    Accept_Language = "zh-CN,zh;q=0.9,en;q=0.8"
})

-- 发送请求
response, err = http_get("https://example.com")
if err then
    print("请求失败:", err)
    return
end

-- 解析HTML
doc, err = parse_html(response.body)
if err then
    print("解析失败:", err)
    return
end

-- 提取标题
title_element, err = select_one(doc, "h1")
if err then
    print("获取标题失败:", err)
else
    title = text(title_element)
    print("页面标题:", title)
end

-- 提取所有链接
links, err = select(doc, "a")
if err then
    print("获取链接失败:", err)
else
    print("找到", #links, "个链接:")
    for i, link in ipairs(links) do
        href, err = attr(link, "href")
        if err then
            print("  链接", i, ": 获取href失败")
        else
            link_text = text(link)
            print("  链接", i, ":", link_text, "->", href)
        end
    end
end
```

### 表单提交

```lua
-- 准备表单数据
form_data = {
    username = "testuser",
    password = "testpass",
    submit = "登录"
}

-- 发送POST请求
response, err = http_post("https://example.com/login", form_data)
if err then
    print("登录失败:", err)
else
    print("登录响应状态码:", response.status_code)
    
    -- 检查登录结果
    doc, err = parse_html(response.body)
    if err then
        print("解析响应失败:", err)
    else
        error_msg, err = select_one(doc, ".error-message")
        if err then
            print("登录成功!")
        else
            error_text = text(error_msg)
            print("登录失败:", error_text)
        end
    end
end
```

## 运行演示

```bash
# 运行Lua演示程序
go run cmd/lua-demo/main.go

# 执行Lua脚本文件
go run cmd/lua-demo/main.go scripts/example.lua
```

## 注意事项

1. **错误处理**: 所有函数都返回错误信息，请务必检查错误
2. **内存管理**: 大型HTML文档会占用较多内存，及时释放不需要的变量
3. **请求频率**: 合理控制请求频率，避免对目标网站造成压力
4. **选择器性能**: 复杂的选择器可能影响性能，尽量使用简单的选择器
5. **编码问题**: 确保HTML文档使用正确的字符编码

## 依赖库

- `github.com/yuin/gopher-lua` - Lua引擎
- `github.com/PuerkitoBio/goquery` - HTML解析
- `video-crawler/internal/crawler` - 浏览器请求工具
