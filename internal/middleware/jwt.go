package middleware

import (
	"fmt"
	"log"
	"slices"

	"video-crawler/internal/config"
	"video-crawler/internal/consts"
	"video-crawler/internal/logger"
	"video-crawler/internal/services"
	"video-crawler/internal/utils"

	"github.com/gin-gonic/gin"
)

// routerWhiteList 路由白名单,白名单路由不进行JWT认证
var routerWhiteList = []string{
	"/api/user/login",
	"/api/user/register",
	"/api/config",
}

// JWTAuthMiddleware JWT 认证中间件
func JWTAuthMiddleware(cfg *config.Config, jwtManager *utils.JWTManager, userService services.UserServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果配置为不需要登录，直接跳过鉴权
		if !cfg.Auth.RequireLogin {
			// 如果不需要登录，则设置为管理员，所有功能都开放使用
			c.Set("is_admin", true)
			log.Printf("跳过JWT验证: RequireLogin = false, 路径 = %s", c.Request.URL.Path)
			c.Next()
			return
		}

		if slices.Contains(routerWhiteList, c.Request.URL.Path) {
			c.Next()
			return
		}
		// 获取 Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.CtxLogger(c).WithError(fmt.Errorf("Authorization header is required")).Error("Login verify failed")
			utils.SendResponse(c, consts.ResponseCodeLoginExpired, "login required", nil)
			c.Abort()
			return
		}

		// 提取 token
		token, err := utils.ExtractTokenFromHeader(authHeader)
		if err != nil {
			logger.CtxLogger(c).WithError(err).Error("Login verify failed")
			utils.SendResponse(c, consts.ResponseCodeLoginExpired, "login required", nil)
			c.Abort()
			return
		}

		// 解析和验证 token
		claims, err := jwtManager.ParseToken(token)
		if err != nil {
			logger.CtxLogger(c).WithError(err).Error("Login verify failed")
			utils.SendResponse(c, consts.ResponseCodeLoginExpired, "login required", nil)
			c.Abort()
			return
		}

		// 判断用户是否允许登录
		user, err := userService.UserDetail(c, claims.UserID)
		if err != nil {
			logger.CtxLogger(c).WithError(err).Error("Login verify failed")
			utils.SendResponse(c, consts.ResponseCodeLoginExpired, "login required", nil)
			c.Abort()
			return
		}

		if !user.AllowLogin {
			logger.CtxLogger(c).WithError(err).Error("Login verify failed")
			utils.SendResponse(c, consts.ResponseCodeLoginExpired, "login required", nil)
			c.Abort()
			return
		}
		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("is_admin", user.IsAdmin)
		c.Set("is_site_admin", user.IsSiteAdmin)
		// 同步角色到响应头
		if user.IsAdmin {
			c.Header("X-User-Is-Admin", "true")
		} else {
			c.Header("X-User-Is-Admin", "false")
		}
		if user.IsSiteAdmin {
			c.Header("X-User-Is-Site-Admin", "true")
		} else {
			c.Header("X-User-Is-Site-Admin", "false")
		}
		c.Set("claims", claims)
		c.Next()
	}
}

// OptionalJWTAuthMiddleware 可选的 JWT 认证中间件（不强制要求认证）
func OptionalJWTAuthMiddleware(cfg *config.Config, jwtManager *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if slices.Contains(routerWhiteList, c.Request.URL.Path) {
			c.Next()
			return
		}
		// 获取 Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 如果没有 Authorization header，继续执行
			c.Next()
			return
		}

		// 提取 token
		token, err := utils.ExtractTokenFromHeader(authHeader)
		if err != nil {
			// 如果格式错误，继续执行但不设置用户信息
			c.Next()
			return
		}

		// 解析和验证 token
		claims, err := jwtManager.ParseToken(token)
		if err != nil {
			// 如果 token 无效，继续执行但不设置用户信息
			c.Next()
			return
		}

		isAdmin := false
		if claims.IsAdmin != nil {
			isAdmin = *claims.IsAdmin
		}
		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("is_admin", isAdmin)
		c.Set("claims", claims)

		c.Next()
	}
}
