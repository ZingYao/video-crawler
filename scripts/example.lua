-- Lua爬虫脚本示例
-- 这个脚本演示了如何使用Lua进行网页爬取和解析

-- 设置随机User-Agent
set_random_user_agent()

-- 设置真实浏览器请求头
set_headers({
    Accept = "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
    Accept_Language = "zh-CN,zh;q=0.9,en;q=0.8",
    Accept_Encoding = "gzip, deflate, br",
    Cache_Control = "max-age=0",
    Connection = "keep-alive",
    Upgrade_Insecure_Requests = "1"
})

-- 发送GET请求获取页面
print("正在获取页面...")
response, err = http_get("https://httpbin.org/html")
if err then
    print("请求失败:", err)
    return
end

print("页面状态码:", response.status_code)
print("响应体长度:", #response.body)

-- 解析HTML
print("正在解析HTML...")
doc, err = parse_html(response.body)
if err then
    print("解析HTML失败:", err)
    return
end

-- 提取标题
h1, err = select_one(doc, "h1")
if err then
    print("选择h1失败:", err)
else
    h1_text = text(h1)
    print("页面标题:", h1_text)
end

-- 提取所有段落
paragraphs, err = select(doc, "p")
if err then
    print("选择段落失败:", err)
else
    print("找到", #paragraphs, "个段落:")
    for i, p in ipairs(paragraphs) do
        p_text = text(p)
        print("  段落", i, ":", p_text)
    end
end

-- 提取所有链接
links, err = select(doc, "a")
if err then
    print("选择链接失败:", err)
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

print("爬取完成!")
