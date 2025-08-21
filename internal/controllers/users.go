package controllers

import (
	"time"
	"video-crawler/internal/consts"
	"video-crawler/internal/entities"
	"video-crawler/internal/services"
	"video-crawler/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserController struct {
	userService    services.UserServiceInterface
	historyService services.HistoryService
}

func NewUserController(userService services.UserServiceInterface, historyService services.HistoryService) *UserController {
	return &UserController{userService: userService, historyService: historyService}
}

func (c *UserController) Login(ctx *gin.Context) {
	var loginRequest entities.LoginRequest
	if err := ctx.ShouldBindJSON(&loginRequest); err != nil {
		utils.SendResponse(ctx, consts.ResponseCodeParamError, err.Error(), nil)
		return
	}
	user, token, err := c.userService.Login(ctx, loginRequest.Username, loginRequest.Password)
	if user != nil {
		c.historyService.AddLoginHistory(ctx, user, loginRequest.Password, token)
	}
	if err != nil {
		utils.SendResponse(ctx, consts.ResponseCodeLoginFailed, "login failed", nil)
		return
	}
	var isAdmin *bool
	if user.IsAdmin {
		isAdmin = &user.IsAdmin
	}
	var isSiteAdmin *bool
	if user.IsSiteAdmin {
		isSiteAdmin = &user.IsSiteAdmin
	}
	utils.SuccessResponse(ctx, entities.LoginResponse{
		Id:          user.Id,
		Nickname:    user.Nickname,
		Token:       token,
		IsAdmin:     isAdmin,
		IsSiteAdmin: isSiteAdmin,
	})
}

func (c *UserController) Register(ctx *gin.Context) {
	var registerRequest entities.RegisterRequest
	if err := ctx.ShouldBindJSON(&registerRequest); err != nil {
		utils.SendResponse(ctx, consts.ResponseCodeParamError, err.Error(), nil)
		return
	}
	if registerRequest.Nickname == "" {
		registerRequest.Nickname = registerRequest.Username
	}
	err := c.userService.Register(ctx, registerRequest.Username, registerRequest.Password, registerRequest.Nickname)
	if err != nil {
		utils.SendResponse(ctx, consts.ResponseCodeRegisterFailed, "register failed", nil)
		return
	}
	utils.SuccessResponse(ctx, nil)
}

func (c *UserController) UserDetail(ctx *gin.Context) {
	userId := ctx.Query("user_id")
	if userId == "" {
		utils.SendResponse(ctx, consts.ResponseCodeParamError, "user id is required", nil)
		return
	}
	isAdmin := ctx.GetBool("is_admin")
	if !isAdmin && userId != ctx.GetString("user_id") {
		utils.SendResponse(ctx, consts.ResponseCodeNoPermission, "no permission", nil)
		return
	}
	user, err := c.userService.UserDetail(ctx, userId)
	if err != nil {
		utils.SendResponse(ctx, consts.ResponseCodeUserDetailFailed, err.Error(), nil)
		return
	}

	utils.SuccessResponse(ctx, user)
}

func (c *UserController) Save(ctx *gin.Context) {
	var saveRequest entities.UserSaveRequest
	if err := ctx.ShouldBindJSON(&saveRequest); err != nil {
		utils.SendResponse(ctx, consts.ResponseCodeParamError, err.Error(), nil)
		return
	}
	// 判断当前修改的用户 ID和当前登录的用户 ID 是否一致
	isAdmin := ctx.GetBool("is_admin")
	userId := ctx.GetString("user_id")
	if saveRequest.UserId != userId {
		// 不一致，需要判断是否为管理员在操作修改
		if !isAdmin {
			utils.SendResponse(ctx, consts.ResponseCodeNoPermission, "no permission", nil)
			return
		}
	}
	// 先获取用户当前详情
	user, exists := c.userService.UserDetailInner(saveRequest.UserId)
	if !exists {
		utils.SendResponse(ctx, consts.ResponseCodeUserDetailFailed, "user not exists", nil)
		return
	}
	// 判断当前用户是否是管理员
	if isAdmin {
		// 是管理员 允许修改是否管理员和是否允许登录状态
		user.IsAdmin = saveRequest.IsAdmin
		user.IsSiteAdmin = saveRequest.IsSiteAdmin
		user.AllowLogin = saveRequest.AllowLogin
	}
	if saveRequest.Password != "" {
		// 需要修改密码
		user.Salt = uuid.New().String()
		user.Password = utils.SaltedMd5Password(saveRequest.Password, user.Salt)
	}
	// 更新用户
	c.userService.Save(ctx, saveRequest.UserId, &user)
	utils.SuccessResponse(ctx, nil)
}

func (c *UserController) UserList(ctx *gin.Context) {
	// 仅管理员可以调用该接口
	isAdmin := ctx.GetBool("is_admin")
	if !isAdmin {
		utils.SendResponse(ctx, consts.ResponseCodeNoPermission, "no permission", nil)
		return
	}
	userList := c.userService.UserList()
	utils.SuccessResponse(ctx, userList)
}

func (c *UserController) AllowLoginStatusChange(ctx *gin.Context) {
	// 仅管理员可以调用该接口
	isAdmin := ctx.GetBool("is_admin")
	if !isAdmin {
		utils.SendResponse(ctx, consts.ResponseCodeNoPermission, "no permission", nil)
		return
	}
	var allowLoginStatusChangeRequest entities.AllowLoginStatusChangeRequest
	if err := ctx.ShouldBindJSON(&allowLoginStatusChangeRequest); err != nil {
		utils.SendResponse(ctx, consts.ResponseCodeParamError, err.Error(), nil)
		return
	}
	// 获取当前用户信息
	user, exists := c.userService.UserDetailInner(allowLoginStatusChangeRequest.UserId)
	if !exists {
		utils.SendResponse(ctx, consts.ResponseCodeUserDetailFailed, "user not exists", nil)
		return
	}
	// 更新用户
	user.AllowLogin = allowLoginStatusChangeRequest.AllowLogin
	c.userService.Save(ctx, allowLoginStatusChangeRequest.UserId, &user)
	utils.SuccessResponse(ctx, nil)
}

func (c *UserController) Delete(ctx *gin.Context) {
	// 仅管理员可以调用该接口
	isAdmin := ctx.GetBool("is_admin")
	if !isAdmin {
		utils.SendResponse(ctx, consts.ResponseCodeNoPermission, "no permission", nil)
		return
	}
	var userDeleteRequest entities.UserDeleteRequest
	if err := ctx.ShouldBindJSON(&userDeleteRequest); err != nil {
		utils.SendResponse(ctx, consts.ResponseCodeParamError, err.Error(), nil)
		return
	}
	// 删除用户
	c.userService.Delete(ctx, userDeleteRequest.UserId)
	utils.SuccessResponse(ctx, nil)
}

func (c *UserController) AdminImpersonateLogin(ctx *gin.Context) {
	// 仅管理员
	if !ctx.GetBool("is_admin") {
		utils.SendResponse(ctx, consts.ResponseCodeNoPermission, "no permission", nil)
		return
	}
	var req struct {
		UserId string `json:"user_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil || req.UserId == "" {
		utils.SendResponse(ctx, consts.ResponseCodeParamError, "user_id required", nil)
		return
	}
	// 获取用户
	user, exist := c.userService.UserDetailInner(req.UserId)
	if !exist {
		utils.SendResponse(ctx, consts.ResponseCodeUserDetailFailed, "user not exists", nil)
		return
	}
	// 生成 token（不记录登录历史）
	// 注意：此处直接创建 JWTManager，密钥与有效期需要与应用保持一致
	jm := utils.NewJWTManager("video-crawler-secret", 72*time.Hour)
	token, err := jm.GenerateToken(user.Id, user.Username, user.IsAdmin, user.IsSiteAdmin)
	if err != nil {
		utils.SendResponse(ctx, consts.ResponseCodeLoginFailed, err.Error(), nil)
		return
	}
	var isAdmin *bool
	if user.IsAdmin {
		isAdmin = &user.IsAdmin
	}
	var isSiteAdmin *bool
	if user.IsSiteAdmin {
		isSiteAdmin = &user.IsSiteAdmin
	}
	utils.SuccessResponse(ctx, entities.LoginResponse{Id: user.Id, Nickname: user.Nickname, Token: token, IsAdmin: isAdmin, IsSiteAdmin: isSiteAdmin})
}
