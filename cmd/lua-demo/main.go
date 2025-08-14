package main

import (
	"fmt"
	"log"

	"video-crawler/internal/crawler"
	"video-crawler/internal/lua"
)

func main() {
	fmt.Println("=== Lua请求和解析演示 ===")

	// 创建浏览器实例
	browser, err := crawler.NewDefaultBrowser()
	if err != nil {
		log.Fatal("创建浏览器失败:", err)
	}
	defer browser.Close()

	// 设置随机User-Agent
	browser.SetRandomUserAgent()

	// 创建Lua引擎
	engine := lua.NewLuaEngine(browser)
	defer engine.Close()

	// 示例1: 简单的GET请求
	fmt.Println("\n1. 简单GET请求示例:")
	script1 := `
		-- 设置请求头
		set_headers({
			Accept = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			Accept_Language = "zh-CN,zh;q=0.9,en;q=0.8"
		})
		
		-- 发送GET请求
		response, err = http_get("https://httpbin.org/get")
		if err then
			print("请求失败:", err)
		else
			print("状态码:", response.status_code)
			print("响应体长度:", #response.body)
		end
	`

	if err := engine.Execute(script1); err != nil {
		log.Printf("执行脚本1失败: %v", err)
	}

	// 示例2: POST请求
	fmt.Println("\n2. POST请求示例:")
	script2 := `
		-- 发送POST请求
		data = {
			name = "测试用户",
			message = "Hello from Lua!",
			number = 123
		}
		
		response, err = http_post("https://httpbin.org/post", data)
		if err then
			print("POST请求失败:", err)
		else
			print("POST状态码:", response.status_code)
			print("POST响应体长度:", #response.body)
		end
	`

	if err := engine.Execute(script2); err != nil {
		log.Printf("执行脚本2失败: %v", err)
	}

	// 示例3: HTML解析
	fmt.Println("\n3. HTML解析示例:")
	script3 := `
		-- 获取一个包含HTML的页面
		response, err = http_get("https://httpbin.org/html")
		if err then
			print("获取HTML失败:", err)
		else
			-- 解析HTML
			doc, err = parse_html(response.body)
			if err then
				print("解析HTML失败:", err)
			else
				-- 选择h1元素
				h1, err = select_one(doc, "h1")
				if err then
					print("选择h1失败:", err)
				else
					h1_text = text(h1)
					print("H1文本:", h1_text)
				end
				
				-- 选择所有p元素
				paragraphs, err = select(doc, "p")
				if err then
					print("选择p元素失败:", err)
				else
					print("找到", #paragraphs, "个段落")
					for i, p in ipairs(paragraphs) do
						p_text = text(p)
						print("段落", i, ":", p_text)
					end
				end
			end
		end
	`

	if err := engine.Execute(script3); err != nil {
		log.Printf("执行脚本3失败: %v", err)
	}

	// 示例4: 复杂爬虫示例
	fmt.Println("\n4. 复杂爬虫示例:")
	script4 := `
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
		
		-- 获取页面
		response, err = http_get("https://httpbin.org/user-agent")
		if err then
			print("获取页面失败:", err)
		else
			print("页面状态码:", response.status_code)
			
			-- 解析JSON响应
			doc, err = parse_html(response.body)
			if err then
				print("解析响应失败:", err)
			else
				-- 尝试提取信息
				body, err = select_one(doc, "body")
				if err then
					print("选择body失败:", err)
				else
					body_text = text(body)
					print("页面内容:", body_text)
				end
			end
		end
	`

	if err := engine.Execute(script4); err != nil {
		log.Printf("执行脚本4失败: %v", err)
	}

	fmt.Println("\n=== 演示完成 ===")
}
