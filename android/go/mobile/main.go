package main

// #include <stdlib.h>
import "C"

import (
	"log"
	"net"
	"os"
	"sync"
	"time"

	"video-crawler/internal/app"
	"video-crawler/internal/config"
)

var startOnce sync.Once

func getFreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

//export StartServer
func StartServer(baseDir *C.char, port C.int) C.int {
	if baseDir != nil {
		os.Setenv("VIDEO_CRAWLER_CONFIG_DIR", C.GoString(baseDir))
	}
	var retPort C.int
	startOnce.Do(func() {
		cfg := &config.Config{
			Server: config.ServerConfig{
				Port: int(port),
				Host: "127.0.0.1",
			},
			Env: "dev",
			Auth: config.AuthConfig{
				RequireLogin: false,
			},
		}

		p := int(port)
		if p == 0 {
			fp, err := getFreePort()
			if err == nil {
				p = fp
			}
		}
		cfg.Server.Port = p
		config.SetConfig(cfg)

		application := app.New(cfg)
		log.Printf("启动HTTP服务在端口 %d", cfg.Server.Port)
		go func() {
			if err := application.Run(); err != nil {
				log.Fatalf("HTTP服务启动失败: %v", err)
			}
		}()
		time.Sleep(500 * time.Millisecond)
		retPort = C.int(cfg.Server.Port)
	})
	return retPort
}

func main() {}
