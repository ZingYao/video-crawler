package middleware

import "github.com/gin-gonic/gin"

func CustomMiddleware(ctx *gin.Context, middle ...func(ctx *gin.Context)) {
	// 重写 Next 方法
	for _, m := range middle {
		m(ctx)
		if ctx.IsAborted() {
			return
		}
	}
}
