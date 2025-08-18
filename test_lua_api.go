package main

import (
	"fmt"
	"log"
	"time"

	"video-crawler/internal/crawler"
	"video-crawler/internal/lua"
)

func main() {
	fmt.Println("开始测试Lua引擎...")

	// 创建浏览器实例
	browser, err := crawler.NewDefaultBrowser()
	if err != nil {
		log.Fatalf("创建浏览器实例失败: %v", err)
	}

	// 创建Lua引擎
	engine := lua.NewLuaEngine(browser)
	defer engine.Close()

	// 获取输出通道
	outputChan := engine.GetOutputChannel()

	// 测试脚本
	script := `
print("Hello, World!")
log("这是一条测试日志")
print("测试完成")
`

	// 在goroutine中执行脚本
	go func() {
		if err := engine.Execute(script); err != nil {
			fmt.Printf("脚本执行失败: %v\n", err)
		}
	}()

	// 监听输出
	fmt.Println("脚本输出:")
	for {
		select {
		case msg, ok := <-outputChan:
			if !ok {
				fmt.Println("输出通道已关闭")
				return
			}
			fmt.Printf("收到: %s\n", msg)
		case <-time.After(5 * time.Second):
			fmt.Println("超时退出")
			return
		}
	}
}
