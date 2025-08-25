package main

import (
	"log"

	"video-crawler/internal/app"
	"video-crawler/internal/config"
)

func main() {
	// 加载配置
	cfg, err := config.Load(true)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建应用实例
	app := app.New(cfg)

	// 启动HTTP服务
	log.Printf("启动HTTP服务在端口 %d", cfg.Server.Port)
	if err := app.Run(); err != nil {
		log.Fatalf("HTTP服务启动失败: %v", err)
	}
}
