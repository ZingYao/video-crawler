-- 测试Lua脚本
print("开始执行Lua脚本测试")

-- 测试基本输出
print("Hello, World!")
log("这是一条日志信息")

-- 测试HTTP请求
print("开始测试HTTP请求...")
local response, err = http_get("https://httpbin.org/get")
if err then
    print("HTTP请求失败:", err)
else
    print("HTTP请求成功，状态码:", response.status_code)
    print("响应URL:", response.url)
end

-- 测试HTML解析
print("开始测试HTML解析...")
local html = "<html><body><div class='test'>测试内容</div><p>段落内容</p></body></html>"
local doc, err = parse_html(html)
if err then
    print("HTML解析失败:", err)
else
    print("HTML解析成功")
    
    -- 测试选择器
    local elements, err = select(doc, "div")
    if err then
        print("选择器执行失败:", err)
    else
        print("找到", #elements, "个div元素")
        for i, element in ipairs(elements) do
            print("元素", i, "文本:", text(element))
            print("元素", i, "HTML:", html(element))
        end
    end
    
    -- 测试单个元素选择
    local element, err = select_one(doc, "p")
    if err then
        print("单个元素选择失败:", err)
    else
        print("找到段落元素，文本:", text(element))
    end
end

-- 测试循环和延迟
print("开始测试循环...")
for i = 1, 3 do
    print("循环次数:", i)
    log("日志循环:", i)
end

print("Lua脚本测试完成")
log("测试结束")
