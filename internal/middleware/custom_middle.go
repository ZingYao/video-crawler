package middleware

import "github.com/gin-gonic/gin"

// CustomMiddleware 依次执行传入的处理函数；鉴权与响应头写入由各自的中间件负责
func CustomMiddleware(ctx *gin.Context, middle ...func(ctx *gin.Context)) {
	for _, m := range middle {
		m(ctx)
		if ctx.IsAborted() {
			return
		}
	}
}

