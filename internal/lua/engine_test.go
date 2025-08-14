package lua

import (
	"testing"

	"video-crawler/internal/crawler"
)

func TestLuaEngineGetBaiduElement(t *testing.T) {
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

	// Lua脚本：获取百度首页指定元素的HTML
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
		
		-- 获取 #s-top-left > a:nth-child(2) 的下一个a标签
		print("正在查找目标元素...")
		
		-- 获取所有a标签然后选择第3个
		all_links, err = select(doc, "#s-top-left a")
		if err then
			print("获取链接失败:", err)
			return
		end
		
		print("找到", #all_links, "个a标签")
		
		-- 获取第3个a标签（索引从1开始，所以第2个的下一个是第3个）
		if #all_links >= 3 then
			target_link = all_links[3]
			link_html = html(target_link)
			link_text = text(target_link)
			href, err = attr(target_link, "href")
			
			print("目标元素HTML:", link_html)
			print("目标元素文本:", link_text)
			if err then
				print("获取href失败:", err)
			else
				print("目标元素链接:", href)
			end
		else
			print("没有找到足够的a标签，只有", #all_links, "个")
		end
	`

	// 执行Lua脚本
	if err := engine.Execute(script); err != nil {
		t.Fatalf("执行Lua脚本失败: %v", err)
	}

	t.Log("Lua脚本执行完成")
}

func TestLuaEngineGetBaiduElementDirect(t *testing.T) {
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

	// 更直接的Lua脚本：使用CSS选择器直接获取目标元素
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
		
		-- 解析HTML
		doc, err = parse_html(response.body)
		if err then
			print("解析HTML失败:", err)
			return
		end
		
		-- 尝试直接选择目标元素
		-- 方法1：使用CSS选择器
		target, err = select_one(doc, "#s-top-left > a:nth-child(3)")
		if err then
			print("方法1失败，尝试方法2...")
			
			-- 方法2：获取所有a标签然后选择第3个
			all_links, err = select(doc, "#s-top-left a")
			if err then
				print("获取所有链接失败:", err)
				return
			end
			
			print("找到", #all_links, "个链接")
			
			if #all_links >= 3 then
				target = all_links[3]
			else
				print("没有找到第3个链接")
				return
			end
		end
		
		-- 输出目标元素的HTML
		target_html = html(target)
		target_text = text(target)
		target_href, err = attr(target, "href")
		
		print("=== 目标元素信息 ===")
		print("HTML:", target_html)
		print("文本:", target_text)
		if err then
			print("href: 获取失败")
		else
			print("href:", target_href)
		end
		print("==================")
	`

	// 执行Lua脚本
	if err := engine.Execute(script); err != nil {
		t.Fatalf("执行Lua脚本失败: %v", err)
	}

	t.Log("Lua脚本执行完成")
}

func TestLuaEngineGetBaiduElementWithValidation(t *testing.T) {
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

	// 带验证的Lua脚本
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
		
		-- 验证响应
		if response.status_code ~= 200 then
			print("HTTP状态码错误:", response.status_code)
			return
		end
		
		-- 检查响应体是否包含百度特征
		if not string.find(response.body, "百度") then
			print("响应体不包含百度特征，可能不是百度首页")
			return
		end
		
		-- 解析HTML
		doc, err = parse_html(response.body)
		if err then
			print("解析HTML失败:", err)
			return
		end
		
		-- 获取目标元素
		target, err = select_one(doc, "#s-top-left > a:nth-child(3)")
		if err then
			print("直接选择失败，尝试遍历方式...")
			
			-- 获取所有链接
			all_links, err = select(doc, "#s-top-left a")
			if err then
				print("获取链接失败:", err)
				return
			end
			
			if #all_links < 3 then
				print("链接数量不足，只有", #all_links, "个")
				return
			end
			
			target = all_links[3]
		end
		
		-- 验证目标元素
		target_html = html(target)
		if target_html == "" then
			print("目标元素HTML为空")
			return
		end
		
		-- 输出结果
		print("=== 测试结果 ===")
		print("目标元素HTML:", target_html)
		print("目标元素文本:", text(target))
		
		href, err = attr(target, "href")
		if err then
			print("href: 获取失败")
		else
			print("href:", href)
		end
		
		-- 验证HTML包含a标签
		if string.find(target_html, "<a") then
			print("验证通过：HTML包含a标签")
		else
			print("验证失败：HTML不包含a标签")
		end
		print("================")
	`

	// 执行Lua脚本
	if err := engine.Execute(script); err != nil {
		t.Fatalf("执行Lua脚本失败: %v", err)
	}

	t.Log("带验证的Lua脚本执行完成")
}

// 测试Lua引擎的基本功能
func TestLuaEngineBasic(t *testing.T) {
	// 创建浏览器实例
	browser, err := crawler.NewDefaultBrowser()
	if err != nil {
		t.Fatalf("创建浏览器失败: %v", err)
	}
	defer browser.Close()

	// 创建Lua引擎
	engine := NewLuaEngine(browser)
	defer engine.Close()

	// 测试基本功能
	script := `
		-- 测试print函数
		print("测试开始")
		
		-- 测试设置User-Agent
		set_random_user_agent()
		
		-- 测试设置请求头
		set_headers({
			Test_Header = "test_value"
		})
		
		-- 测试GET请求
		response, err = http_get("https://httpbin.org/get")
		if err then
			print("GET请求失败:", err)
		else
			print("GET请求成功，状态码:", response.status_code)
		end
		
		-- 测试HTML解析
		test_html = "<html><body><h1>测试标题</h1><p>测试段落</p></body></html>"
		doc, err = parse_html(test_html)
		if err then
			print("HTML解析失败:", err)
		else
			h1, err = select_one(doc, "h1")
			if err then
				print("选择h1失败:", err)
			else
				h1_text = text(h1)
				print("H1文本:", h1_text)
			end
		end
		
		print("测试完成")
	`

	// 执行Lua脚本
	if err := engine.Execute(script); err != nil {
		t.Fatalf("执行Lua脚本失败: %v", err)
	}

	t.Log("基本功能测试完成")
}

// 测试错误处理
func TestLuaEngineErrorHandling(t *testing.T) {
	// 创建浏览器实例
	browser, err := crawler.NewDefaultBrowser()
	if err != nil {
		t.Fatalf("创建浏览器失败: %v", err)
	}
	defer browser.Close()

	// 创建Lua引擎
	engine := NewLuaEngine(browser)
	defer engine.Close()

	// 测试错误处理
	script := `
		-- 测试无效URL
		response, err = http_get("https://invalid-url-that-does-not-exist.com")
		if err then
			print("预期的错误:", err)
		else
			print("意外成功")
		end
		
		-- 测试无效HTML
		doc, err = parse_html("invalid html")
		if err then
			print("HTML解析错误:", err)
		else
			print("HTML解析意外成功")
		end
		
		-- 测试无效选择器
		doc, err = parse_html("<html><body><p>test</p></body></html>")
		if err then
			print("HTML解析失败:", err)
		else
			element, err = select_one(doc, "invalid-selector")
			if err then
				print("选择器错误:", err)
			else
				print("选择器意外成功")
			end
		end
	`

	// 执行Lua脚本
	if err := engine.Execute(script); err != nil {
		t.Fatalf("执行Lua脚本失败: %v", err)
	}

	t.Log("错误处理测试完成")
}
