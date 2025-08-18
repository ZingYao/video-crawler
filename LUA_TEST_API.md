# Lua脚本测试API

## 概述

本项目实现了一个Lua脚本测试接口，支持流式返回执行结果。通过长连接的方式，可以实时查看Lua脚本的执行过程，包括`print`和`log`函数的输出。

## 功能特性

- **流式输出**: 支持实时查看脚本执行过程
- **HTTP请求**: 内置HTTP GET/POST请求功能
- **HTML解析**: 支持HTML解析和CSS选择器
- **多种输出方式**: 支持普通流式输出和Server-Sent Events (SSE)
- **错误处理**: 完善的错误处理和异常捕获

## API接口

### 1. 流式输出接口

**接口地址**: `POST /api/lua/test`

**请求参数**:
```json
{
    "script": "print('Hello, World!')"
}
```

**响应**: 流式文本输出

### 2. SSE接口

**接口地址**: `POST /api/lua/test-sse`

**请求参数**:
```json
{
    "script": "print('Hello, World!')"
}
```

**响应**: Server-Sent Events格式

## Lua脚本支持的功能

### 基本输出
```lua
print("Hello, World!")
log("这是一条日志信息")
```

### HTTP请求
```lua
-- GET请求
local response, err = http_get("https://httpbin.org/get")
if err then
    print("请求失败:", err)
else
    print("状态码:", response.status_code)
    print("响应体:", response.body)
end

-- POST请求
local data = {name = "test", value = 123}
local response, err = http_post("https://httpbin.org/post", data)
```

### 设置请求头和Cookie
```lua
set_headers({["User-Agent"] = "Custom Agent"})
set_cookies({session = "abc123"})
set_user_agent("My Browser")
set_random_user_agent()
```

### HTML解析
```lua
local html = "<html><body><div class='test'>内容</div></body></html>"
local doc, err = parse_html(html)
if not err then
    -- 选择所有div元素
    local elements, err = select(doc, "div")
    if not err then
        for i, element in ipairs(elements) do
            print("文本:", text(element))
            print("HTML:", html(element))
            print("属性:", attr(element, "class"))
        end
    end
    
    -- 选择单个元素
    local element, err = select_one(doc, "div.test")
    if not err then
        print("找到元素:", text(element))
    end
end
```

## 使用示例

### JavaScript客户端示例

```javascript
async function testLuaScript() {
    const script = `
        print("开始执行脚本")
        log("测试日志输出")
        
        local response, err = http_get("https://httpbin.org/get")
        if err then
            print("HTTP请求失败:", err)
        else
            print("请求成功，状态码:", response.status_code)
        end
        
        print("脚本执行完成")
    `;
    
    try {
        const response = await fetch('/api/lua/test', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ script: script })
        });
        
        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        
        while (true) {
            const { done, value } = await reader.read();
            if (done) break;
            
            const chunk = decoder.decode(value);
            console.log(chunk); // 实时输出
        }
    } catch (error) {
        console.error('执行失败:', error);
    }
}
```

### cURL示例

```bash
curl -X POST http://localhost:8080/api/lua/test \
  -H "Content-Type: application/json" \
  -d '{"script": "print(\"Hello from cURL!\")"}'
```

## 错误处理

脚本执行过程中的错误会被捕获并返回：

- **语法错误**: 会在执行时被捕获并返回错误信息
- **运行时错误**: 包括HTTP请求失败、HTML解析错误等
- **Panic错误**: 严重的运行时错误会被捕获并返回

## 注意事项

1. **超时设置**: 长时间运行的脚本可能会被中断
2. **资源限制**: 避免在脚本中执行过于复杂的操作
3. **并发限制**: 同时执行的脚本数量有限制
4. **内存使用**: 大型HTML文档解析会消耗较多内存

## 测试页面

项目根目录下的 `lua_test.html` 文件提供了一个简单的测试界面，可以直接在浏览器中测试Lua脚本功能。

## 开发说明

### 文件结构

```
internal/
├── lua/
│   └── engine.go          # Lua引擎核心实现
├── services/
│   └── lua_test.go        # Lua测试服务
├── controllers/
│   └── lua_test.go        # Lua测试控制器
└── handler/
    └── handler.go         # 路由处理
```

### 扩展功能

如需添加新的Lua函数，请在 `internal/lua/engine.go` 的 `registerFunctions` 方法中注册，并实现对应的处理函数。
