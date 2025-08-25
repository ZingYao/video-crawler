package main

import (
	"context"
	"embed"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"video-crawler/internal/app"
	"video-crawler/internal/config"
	"video-crawler/internal/logger"
)

//go:embed frontend/dist
var assets embed.FS

// App struct
type App struct {
	ctx        context.Context
	config     *config.Config
	app        *app.App
	serverPort int
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 初始化配置
	cfg, err := config.Load(true)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	a.config = cfg

	// 设置Wails模式，禁用登录
	a.config.Auth.RequireLogin = false
	log.Printf("Wails模式: 已禁用登录验证, RequireLogin = %v", a.config.Auth.RequireLogin)

	// 初始化日志
	logger.Init(cfg)

	// 初始化应用
	a.app = app.New(cfg)

	// 启动HTTP服务在随机端口
	a.serverPort = a.startHTTPServer()
	log.Printf("Wails HTTP服务启动在端口: %d", a.serverPort)
}

// startHTTPServer 启动HTTP服务并返回端口号
func (a *App) startHTTPServer() int {
	// 查找可用端口
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("无法启动HTTP服务: %v", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// 设置端口并启动服务
	a.config.Server.Port = port

	// 在后台启动HTTP服务
	go func() {
		log.Printf("启动Wails HTTP服务在端口: %d", port)
		if err := a.app.Run(); err != nil {
			log.Printf("HTTP服务启动失败: %v", err)
		}
	}()

	return port
}

// domReady is called after the front-end dom has been loaded
func (a *App) domReady(ctx context.Context) {
	// Add your action here
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	// Perform your teardown here
}

// 获取应用配置目录
func getAppConfigDir() string {
	var configDir string

	switch runtime.GOOS {
	case "darwin":
		// macOS: ~/Library/Application Support/video-crawler
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		configDir = filepath.Join(homeDir, "Library", "Application Support", "video-crawler")
	case "windows":
		// Windows: %APPDATA%\video-crawler
		appData := os.Getenv("APPDATA")
		if appData == "" {
			log.Fatal("无法获取APPDATA环境变量")
		}
		configDir = filepath.Join(appData, "video-crawler")
	case "linux":
		// Linux: ~/.config/video-crawler
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		configDir = filepath.Join(homeDir, ".config", "video-crawler")
	default:
		log.Fatal("不支持的操作系统")
	}

	// 创建配置目录
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Fatalf("创建配置目录失败: %v", err)
	}

	return configDir
}

// ========== Wails API 桥接方法 ==========

// GetServerPort 获取HTTP服务端口
func (a *App) GetServerPort() int {
	return a.serverPort
}

// GetConfig 获取系统配置
func (a *App) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"require_login": a.config.Auth.RequireLogin,
		"env":           a.config.Env,
		"server_port":   a.serverPort,
	}
}

func main() {
	// 设置配置目录
	configDir := getAppConfigDir()
	os.Setenv("VIDEO_CRAWLER_CONFIG_DIR", configDir)

	// 复制默认配置文件到应用目录（如果不存在）
	defaultConfigPath := "configs/config.yaml"
	appConfigPath := filepath.Join(configDir, "config.yaml")

	if _, err := os.Stat(appConfigPath); os.IsNotExist(err) {
		if _, err := os.Stat(defaultConfigPath); err == nil {
			// 读取默认配置
			defaultConfig, err := os.ReadFile(defaultConfigPath)
			if err == nil {
				// 写入到应用配置目录
				if err := os.WriteFile(appConfigPath, defaultConfig, 0644); err != nil {
					log.Printf("复制配置文件失败: %v", err)
				}
			}
		}
	}

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:             "Video Crawler Desktop",
		Width:             1200,
		Height:            800,
		MinWidth:          800,
		MinHeight:         600,
		MaxWidth:          1920,
		MaxHeight:         1080,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		Menu:              nil,
		Logger:            nil,
		LogLevel:          0,
		OnStartup:         app.startup,
		OnDomReady:        app.domReady,
		OnBeforeClose:     nil,
		OnShutdown:        app.shutdown,
		WindowStartState:  options.Normal,
		Assets:            assets,
		Bind: []interface{}{
			app,
		},
		// Windows platform specific options
		Windows: &windows.Options{},
	},
	)
	if err != nil {
		log.Fatal(err)
	}
}
