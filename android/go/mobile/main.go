package main

// #include <stdlib.h>
import "C"

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"video-crawler/internal/app"
	"video-crawler/internal/config"
)

var (
	startOnce     sync.Once
	serverPort    int
	serverStarted bool
	serverError   string
)

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
	defer func() {
		r := recover()
		if r != nil {
			log.Printf("StartServer 发生恐慌: %v", r)
			serverError = fmt.Sprintf("(%v)StartServer 发生恐慌: %v", serverError, r)
		}
	}()
	log.Printf("StartServer 被调用: baseDir=%v, port=%d", baseDir, port)

	if baseDir != nil {
		configDir := C.GoString(baseDir)
		log.Printf("设置配置目录: %s", configDir)
		os.Setenv("VIDEO_CRAWLER_CONFIG_DIR", configDir)
	}

	startOnce.Do(func() {
		log.Printf("开始初始化配置...")
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
		} else {
			// 检查端口是否被占用
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", p), 1*time.Second)
			if err == nil {
				conn.Close()
				p = 0
			}
			// 如果端口被占用   则获取一个随机端口
			if p == 0 {
				fp, err := getFreePort()
				if err == nil {
					p = fp
				}
			}
		}
		cfg.Server.Port = p
		config.SetConfig(cfg)

		application := app.New(cfg)
		log.Printf("启动HTTP服务在端口 %d", cfg.Server.Port)

		// 使用 channel 来等待服务启动结果
		waitErr := make(chan error, 1)

		go func() {
			if err := application.Run(); err != nil {
				errorMsg := fmt.Sprintf("HTTP服务启动失败: %v", err)
				log.Printf(errorMsg)
				serverError = errorMsg
				waitErr <- err
			}
		}()

		// 等待服务启动或超时
		timeout := time.After(60 * time.Second) // 60秒超时
		for {
			select {
			case err := <-waitErr:
				log.Printf("HTTP服务启动失败: %v", err)
				serverError = fmt.Sprintf("HTTP服务启动失败: %v", err)
				return
			case <-timeout:
				log.Printf("HTTP服务启动超时")
				serverError = "HTTP服务启动超时"
				return
			case <-time.After(500 * time.Millisecond):
				// 每500ms检测端口是否成功被监听
				if _, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", cfg.Server.Port)); err == nil {
					log.Printf("HTTP服务启动成功，端口: %d", cfg.Server.Port)
					serverPort = cfg.Server.Port
					serverStarted = true
					log.Printf("StartServer 完成，返回端口: %d", serverPort)
					return
				}
			}
		}
	})

	if serverStarted {
		return C.int(serverPort)
	}
	return 0
}

//export GetServerError
func GetServerError() *C.char {
	if serverError != "" {
		return C.CString(serverError)
	}
	return C.CString("")
}

//export GetServerStatus
func GetServerStatus() *C.char {
	if serverStarted {
		return C.CString(fmt.Sprintf("running:%d", serverPort))
	}
	if serverError != "" {
		return C.CString(fmt.Sprintf("error:%s", serverError))
	}
	return C.CString("not_started")
}

func main() {}
