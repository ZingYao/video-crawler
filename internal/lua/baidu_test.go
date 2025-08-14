package lua

import (
	"testing"

	"video-crawler/internal/crawler"
)

func TestGetBaiduElementHTML(t *testing.T) {
	// 创建浏览器实例
	browser, err := crawler.NewDefaultBrowser()
	if err != nil {
		t.Fatalf("创建浏览器失败: %v", err)
	}
	defer browser.Close()

	// 设置随机User-Agent
	browser.SetRandomUserAgent()

	// 创建Lua引擎
	engine := NewLuaEngine(browser)
	defer engine.Close()

	// 专门获取百度首页指定元素HTML的脚本
	script := `
		-- 设置真实浏览器请求头
		set_headers({
			Accept = "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			Accept_Language = "zh-CN,zh;q=0.9,en;q=0.8",
			Accept_Encoding = "gzip, deflate, br",
			Cache_Control = "max-age=0",
			Connection = "keep-alive",
			Upgrade_Insecure_Requests = "1"
		})
		
		-- 获取百度首页
		print("正在获取百度首页...")
		response, err = http_get("https://www.baidu.com")
		if err then
			print("获取百度首页失败:", err)
			return
		end
		
		print("百度首页状态码:", response.status_code)
		print("响应体长度:", #response.body)
		
		-- 解析HTML
		print("正在解析HTML...")
		doc, err = parse_html(response.body)
		if err then
			print("解析HTML失败:", err)
			return
		end
		
		-- 获取 #s-top-left 下的所有a标签
		print("正在查找 #s-top-left 下的a标签...")
		all_links, err = select(doc, "#s-top-left a")
		if err then
			print("获取链接失败:", err)
			return
		end
		
		print("找到", #all_links, "个a标签")
		
		-- 输出所有a标签的信息
		for i, link in ipairs(all_links) do
			link_html = html(link)
			link_text = text(link)
			href, err = attr(link, "href")
			
			print("=== 第", i, "个a标签 ===")
			print("HTML:", link_html)
			print("文本:", link_text)
			if err then
				print("href: 获取失败")
			else
				print("href:", href)
			end
			print("==================")
		end
		
		-- 特别关注第3个a标签（#s-top-left > a:nth-child(2) 的下一个）
		if #all_links >= 3 then
			target_link = all_links[3]
			target_html = html(target_link)
			target_text = text(target_link)
			target_href, err = attr(target_link, "href")
			
			print("=== 目标元素（第3个a标签）===")
			print("HTML:", target_html)
			print("文本:", target_text)
			if err then
				print("href: 获取失败")
			else
				print("href:", target_href)
			end
			print("==========================")
		else
			print("没有找到第3个a标签")
		end
	`

	// 执行Lua脚本
	if err := engine.Execute(script); err != nil {
		t.Fatalf("执行Lua脚本失败: %v", err)
	}

	t.Log("百度元素获取测试完成")
}

func TestGetBaiduElementWithCSSSelector(t *testing.T) {
	// 创建浏览器实例
	browser, err := crawler.NewDefaultBrowser()
	if err != nil {
		t.Fatalf("创建浏览器失败: %v", err)
	}
	defer browser.Close()

	// 设置随机User-Agent
	browser.SetRandomUserAgent()

	// 创建Lua引擎
	engine := NewLuaEngine(browser)
	defer engine.Close()

	// 使用CSS选择器直接获取目标元素
	script := `
		-- 设置真实浏览器请求头
		set_headers({
			Accept = "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			Accept_Language = "zh-CN,zh;q=0.9,en;q=0.8",
			Accept_Encoding = "gzip, deflate, br",
			Cache_Control = "max-age=0",
			Connection = "keep-alive",
			Upgrade_Insecure_Requests = "1"
		})
		
		-- 获取百度首页
		response, err = http_get("https://www.baidu.com")
		if err then
			print("获取百度首页失败:", err)
			return
		end
		
		-- 解析HTML
		doc, err = parse_html(response.body)
		if err then
			print("解析HTML失败:", err)
			return
		end
		
		-- 尝试不同的CSS选择器
		selectors = {
			"#s-top-left > a:nth-child(3)",
			"#s-top-left a:nth-child(3)",
			"#s-top-left > a:nth-of-type(3)",
			"#s-top-left a:nth-of-type(3)"
		}
		
		for i, selector in ipairs(selectors) do
			print("尝试选择器", i, ":", selector)
			element, err = select_one(doc, selector)
			if err then
				print("选择器", i, "失败:", err)
			else
				element_html = html(element)
				element_text = text(element)
				element_href, err = attr(element, "href")
				
				print("选择器", i, "成功:")
				print("HTML:", element_html)
				print("文本:", element_text)
				if err then
					print("href: 获取失败")
				else
					print("href:", element_href)
				end
				break
			end
		end
	`

	// 执行Lua脚本
	if err := engine.Execute(script); err != nil {
		t.Fatalf("执行Lua脚本失败: %v", err)
	}

	t.Log("CSS选择器测试完成")
}

func TestGetBaiduElementRawHTML(t *testing.T) {
	// 创建浏览器实例
	browser, err := crawler.NewDefaultBrowser()
	if err != nil {
		t.Fatalf("创建浏览器失败: %v", err)
	}
	defer browser.Close()

	// 设置随机User-Agent
	browser.SetRandomUserAgent()

	// 创建Lua引擎
	engine := NewLuaEngine(browser)
	defer engine.Close()

	// 获取原始HTML并分析
	script := `
		-- 设置真实浏览器请求头
		set_headers({
			Accept = "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			Accept_Language = "zh-CN,zh;q=0.9,en;q=0.8",
			Accept_Encoding = "gzip, deflate, br",
			Cache_Control = "max-age=0",
			Connection = "keep-alive",
			Upgrade_Insecure_Requests = "1"
		})
		
		-- 获取百度首页
		response, err = http_get("https://www.baidu.com")
		if err then
			print("获取百度首页失败:", err)
			return
		end
		
		-- 解析HTML
		doc, err = parse_html(response.body)
		if err then
			print("解析HTML失败:", err)
			return
		end
		
		-- 获取 #s-top-left 容器
		top_left, err = select_one(doc, "#s-top-left")
		if err then
			print("获取 #s-top-left 失败:", err)
			return
		end
		
		-- 获取容器的HTML
		container_html = html(top_left)
		print("=== #s-top-left 容器HTML ===")
		print(container_html)
		print("============================")
		
		-- 获取容器下的所有a标签
		links, err = select(top_left, "a")
		if err then
			print("获取a标签失败:", err)
			return
		end
		
		print("找到", #links, "个a标签")
		
		-- 输出每个a标签的详细信息
		for i, link in ipairs(links) do
			link_html = html(link)
			link_text = text(link)
			href, err = attr(link, "href")
			class, err2 = attr(link, "class")
			
			print("=== 第", i, "个a标签详细信息 ===")
			print("完整HTML:", link_html)
			print("文本内容:", link_text)
			if err then
				print("href: 获取失败")
			else
				print("href:", href)
			end
			if err2 then
				print("class: 获取失败")
			else
				print("class:", class)
			end
			print("=============================")
		end
		
		-- 特别输出第3个a标签（目标元素）
		if #links >= 3 then
			target = links[3]
			target_html = html(target)
			print("=== 目标元素（第3个a标签）完整HTML ===")
			print(target_html)
			print("=====================================")
		end
	`

	// 执行Lua脚本
	if err := engine.Execute(script); err != nil {
		t.Fatalf("执行Lua脚本失败: %v", err)
	}

	t.Log("原始HTML分析测试完成")
}
