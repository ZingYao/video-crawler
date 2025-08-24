package handler

import (
	"video-crawler/internal/config"
	"video-crawler/internal/controllers"
	"video-crawler/internal/services"

	"github.com/gin-gonic/gin"
)

// Handler HTTP处理器
type Handler struct {
	config             *config.Config
	userService        services.UserServiceInterface
	videoSourceService services.VideoSourceService
	historyService     services.HistoryService
	luaTestService     services.LuaTestService
}

// New 创建新的处理器实例
func New(cfg *config.Config, userService services.UserServiceInterface, videoSourceService services.VideoSourceService, historyService services.HistoryService, luaTestService services.LuaTestService) *Handler {
	return &Handler{
		config:             cfg,
		userService:        userService,
		videoSourceService: videoSourceService,
		historyService:     historyService,
		luaTestService:     luaTestService,
	}
}

// HandleApi 处理 API 请求
func (h *Handler) HandleApi(c *gin.Context) {
	userController := controllers.NewUserController(h.userService, h.historyService)
	videoSourceController := controllers.NewVideoSourceController(h.videoSourceService)
	videoController := controllers.NewVideoController(h.videoSourceService, h.historyService, h.userService)
	historyController := controllers.NewHistoryController(h.historyService, h.userService)
	switch c.Request.URL.Path {
	case "/api":
		c.JSON(200, gin.H{
			"message": "API 接口",
			"routes": []string{
				"GET / - 首页",
				"GET /api - API 信息",
				"GET /health - 健康检查",
				"GET /api/video-source/list - 站点列表",
				"GET /api/video-source/detail - 站点详情",
				"POST /api/video-source/save - 保存站点",
				"POST /api/video-source/delete - 删除站点",
				"POST /api/video-source/set-status - 设置站点状态",
				"GET /api/video/home/list - 视频首页推荐",
				"GET /api/video/search - 视频搜索",
				"GET /api/video/detail - 视频详情",
				"GET /api/video/url - 视频URL",
				"GET /api/history/search - 历史搜索",
				"GET /api/history/video - 视频观看历史",
				"GET /api/history/login - 登录历史",
				"POST /api/user/login - 用户登录",
				"POST /api/user/register - 用户注册",
				"GET /api/user/detail - 用户详情",
				"POST /api/user/save - 用户保存",
				"GET /api/user/list - 用户列表",
				"POST /api/lua/test - Lua脚本测试(流式)",
				"POST /api/lua/test-sse - Lua脚本测试(SSE)",
			},
		})
	case "/api/video-source/list":
		// 站点列表
		videoSourceController.List(c)
	case "/api/video-source/detail":
		// 站点详情
		videoSourceController.Detail(c)
	case "/api/video-source/save":
		// 保存站点
		videoSourceController.Save(c)
	case "/api/video-source/delete":
		// 删除站点
		videoSourceController.Delete(c)
	case "/api/video-source/check-status":
		// 检查站点状态
		videoSourceController.CheckStatus(c)
	case "/api/video-source/set-status":
		// 设置站点状态
		videoSourceController.SetStatus(c)
	case "/api/video/search":
		// 视频搜索
		videoController.Search(c)
	case "/api/video/detail":
		// 视频详情
		videoController.Detail(c)
	case "/api/video/url":
		// 视频URL
		videoController.PlayURL(c)
	case "/api/history/search":
		// 历史搜索
		historyController.GetSearchHistory(c)
	case "/api/history/video":
		// 视频观看历史
		historyController.GetVideoHistory(c)
	case "/api/history/login":
		// 登录历史
		historyController.GetLoginHistory(c)
	case "/api/user/login":
		// 用户登录
		userController.Login(c)
	case "/api/user/register":
		// 用户注册
		userController.Register(c)
	case "/api/user/detail":
		// 用户详情
		userController.UserDetail(c)
	case "/api/user/save":
		// 用户保存
		userController.Save(c)
	case "/api/user/list":
		// 用户列表
		userController.UserList(c)
	case "/api/user/delete":
		// 删除用户
		userController.Delete(c)
	case "/api/user/allow-login-status-change":
		// 用户能否登录状态修改
		userController.AllowLoginStatusChange(c)
	case "/api/user/admin-impersonate-login":
		// 管理员伪登录（不记录登录历史）
		userController.AdminImpersonateLogin(c)
	case "/api/lua/test":
		// Lua脚本测试(流式)
		luaTestController := controllers.NewLuaTestController(h.luaTestService)
		luaTestController.TestScript(c)
	case "/api/lua/test-sse":
		// Lua脚本测试(SSE)
		luaTestController := controllers.NewLuaTestController(h.luaTestService)
		luaTestController.TestScriptSSE(c)
	case "/api/js/test":
		// JS脚本测试(流式)
		luaTestController := controllers.NewLuaTestController(h.luaTestService)
		luaTestController.TestJSScript(c)
	case "/api/js/advanced-test":
		// JS高级调试
		luaTestController := controllers.NewLuaTestController(h.luaTestService)
		luaTestController.AdvancedTestJSScript(c)
	case "/api/lua/advanced-test":
		// Lua高级调试
		luaTestController := controllers.NewLuaTestController(h.luaTestService)
		luaTestController.AdvancedTestLuaScript(c)
	}
}

// HandleHealth 处理健康检查请求
func (h *Handler) HandleHealth(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":    "ok",
		"message":   "服务运行正常",
		"framework": "Gin",
	})
}
