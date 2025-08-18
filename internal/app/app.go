package app

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"video-crawler/internal/config"
	"video-crawler/internal/handler"
	"video-crawler/internal/logger"
	"video-crawler/internal/middleware"
	"video-crawler/internal/services"
	"video-crawler/internal/static"
	"video-crawler/internal/utils"
)

// App 应用结构
type App struct {
	config      *config.Config
	httpHandler *handler.Handler
	userService services.UserServiceInterface
	engine      *gin.Engine
}

var (
	fs                  http.FileSystem = static.GetStaticFS()
	indexFile, _                        = fs.Open("/index.html")
	indexFileContent, _                 = io.ReadAll(indexFile)
)

// New 创建新的应用实例
func New(cfg *config.Config) *App {
	// 设置 Gin 模式
	if cfg.Server.Host == "0.0.0.0" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化日志
	logger.Init(cfg)

	// 创建 Gin 引擎
	engine := gin.New()

	// 添加中间件
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	jwtManager := utils.NewJWTManager(cfg.Server.JwtSecret, time.Duration(cfg.Server.JwtExpire)*time.Hour)
	userService := services.NewUserService(jwtManager)
	videoSourceService := services.NewVideoSourceService()
	historyService := services.GetHistoryService()
	luaTestService := services.NewLuaTestService()
	return &App{
		config:      cfg,
		httpHandler: handler.New(cfg, userService, videoSourceService, historyService, luaTestService),
		userService: userService,
		engine:      engine,
	}
}

// Run 启动应用
func (a *App) Run() error {
	// 注册路由
	a.registerRoutes()

	addr := fmt.Sprintf("%s:%d", a.config.Server.Host, a.config.Server.Port)
	log.Printf("Starting gin server on %s", addr)

	return a.engine.Run(addr)
}

// registerRoutes 注册所有路由
func (a *App) registerRoutes() {

	// // API 路由
	// a.engine.GET("/api", a.httpHandler.HandleApi)

	// // 健康检查路由
	// a.engine.GET("/health", a.httpHandler.HandleHealth)
	jwtManager := utils.NewJWTManager(a.config.Server.JwtSecret, time.Duration(a.config.Server.JwtExpire)*time.Hour)

	// 静态文件处理 - 只处理静态资源
	a.engine.StaticFS("/", static.GetStaticFS())

	// 处理所有其他路由
	a.engine.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			middleware.CustomMiddleware(c, middleware.RequestIdMiddleware(), middleware.LoggerMiddleware(), middleware.JWTAuthMiddleware(jwtManager, a.userService), a.httpHandler.HandleApi)
			return
		} else if strings.HasPrefix(c.Request.URL.Path, "/health") {
			middleware.CustomMiddleware(c, middleware.RequestIdMiddleware(), middleware.LoggerMiddleware(), a.httpHandler.HandleHealth)
			return
		} else {
			// 对于所有其他路径，返回 index.html 以支持前端路由
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.Status(http.StatusOK)
			c.Writer.Write(indexFileContent)
			return
		}
	})
}
