package middleware

import (
	"fmt"
	"slices"

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
}

// JWTAuthMiddleware JWT 认证中间件
func JWTAuthMiddleware(jwtManager *utils.JWTManager, userService services.UserServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
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
		c.Set("claims", claims)
		c.Next()
	}
}

// OptionalJWTAuthMiddleware 可选的 JWT 认证中间件（不强制要求认证）
func OptionalJWTAuthMiddleware(jwtManager *utils.JWTManager) gin.HandlerFunc {
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
