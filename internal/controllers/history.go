package controllers

import (
	"video-crawler/internal/consts"
	"video-crawler/internal/services"
	"video-crawler/internal/utils"

	"github.com/gin-gonic/gin"
)

type HistoryController struct {
	historyService services.HistoryService
	userService    services.UserServiceInterface
}

func NewHistoryController(historyService services.HistoryService, userService services.UserServiceInterface) *HistoryController {
	return &HistoryController{
		historyService: historyService,
		userService:    userService,
	}
}

// checkUserPermission 检查用户权限
// 只有管理员可以查看其他用户的登录历史，普通用户只能查看自己的历史
func (c *HistoryController) checkUserPermission(ctx *gin.Context, targetUserId string) (bool, string) {
	// 从上下文中获取当前用户信息
	currentUserId, exists := ctx.Get("user_id")
	if !exists {
		return false, "未登录用户"
	}

	currentIsAdmin, exists := ctx.Get("is_admin")
	if !exists {
		return false, "无法获取用户权限信息"
	}

	// 如果是管理员，允许查看任何用户的历史
	if isAdmin, ok := currentIsAdmin.(bool); ok && isAdmin {
		return true, ""
	}

	// 普通用户只能查看自己的历史
	if currentUserId.(string) == targetUserId {
		return true, ""
	}

	return false, "权限不足，只能查看自己的登录历史"
}

// 获取搜索历史
func (c *HistoryController) GetSearchHistory(ctx *gin.Context) {
	userId := ctx.Query("user_id")
	if userId == "" {
		utils.SendResponse(ctx, consts.ResponseCodeParamError, "用户ID不能为空", nil)
		return
	}
	// 用户ID 获取用户信息
	userEntity, exists := c.userService.UserDetailInner(userId)
	if !exists {
		utils.SendResponse(ctx, consts.ResponseCodeParamError, "用户不存在", nil)
		return
	}

	response, err := c.historyService.GetSearchHistory(ctx, userEntity.Username)
	if err != nil {
		utils.SendResponse(ctx, consts.ResponseCodeSystemError, "获取搜索历史失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(ctx, response)
}

// 获取视频观看历史
func (c *HistoryController) GetVideoHistory(ctx *gin.Context) {
	userId := ctx.Query("user_id")
	if userId == "" {
		utils.SendResponse(ctx, consts.ResponseCodeParamError, "用户ID不能为空", nil)
		return
	}

	userEntity, exists := c.userService.UserDetailInner(userId)
	if !exists {
		utils.SendResponse(ctx, consts.ResponseCodeParamError, "用户不存在", nil)
		return
	}

	response, err := c.historyService.GetVideoHistory(ctx, userEntity.Username)
	if err != nil {
		utils.SendResponse(ctx, consts.ResponseCodeSystemError, "获取视频观看历史失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(ctx, response)
}

// 获取登录历史
func (c *HistoryController) GetLoginHistory(ctx *gin.Context) {
	userId := ctx.Query("user_id")
	if userId == "" {
		utils.SendResponse(ctx, consts.ResponseCodeParamError, "用户ID不能为空", nil)
		return
	}

	// 检查用户权限
	hasPermission, errMsg := c.checkUserPermission(ctx, userId)
	if !hasPermission {
		utils.SendResponse(ctx, consts.ResponseCodeParamError, errMsg, nil)
		return
	}

	userEntity, exists := c.userService.UserDetailInner(userId)
	if !exists {
		utils.SendResponse(ctx, consts.ResponseCodeParamError, "用户不存在", nil)
		return
	}

	response, err := c.historyService.GetLoginHistory(ctx, userEntity.Username)
	if err != nil {
		utils.SendResponse(ctx, consts.ResponseCodeSystemError, "获取登录历史失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(ctx, response)
}
