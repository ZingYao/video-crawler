package utils

import (
	"net/http"
	"video-crawler/internal/consts"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func SuccessResponse(ctx *gin.Context, data any) {
	SendResponse(ctx, consts.ResponseCodeSuccess, "success", data)
}

func SendResponse(ctx *gin.Context, code int, msg string, data any) {
	ctx.JSON(http.StatusOK, Response{
		Code:    code,
		Message: msg,
		Data:    data,
	})
}
