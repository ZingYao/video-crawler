package middleware

import "github.com/gin-gonic/gin"

func CustomMiddleware(ctx *gin.Context, middle ...func(ctx *gin.Context)) {
	if len(middle) == 0 {
		return
	}
	last := len(middle) - 1
	// 1) 先执行除最后一个外的准备中间件（如 RequestId、Logger、JWT 等）
	for i := 0; i < last; i++ {
		middle[i](ctx)
		if ctx.IsAborted() {
			return
		}
	}
	// 2) 在最终处理函数前，写入角色信息到响应头
	if v, ok := ctx.Get("is_admin"); ok {
		if b, ok2 := v.(bool); ok2 {
			if b {
				ctx.Header("X-User-Is-Admin", "true")
			} else {
				ctx.Header("X-User-Is-Admin", "false")
			}
		}
	}
	if v, ok := ctx.Get("is_site_admin"); ok {
		if b, ok2 := v.(bool); ok2 {
			if b {
				ctx.Header("X-User-Is-Site-Admin", "true")
			} else {
				ctx.Header("X-User-Is-Site-Admin", "false")
			}
		}
	}
	// 3) 执行最终处理函数（通常为具体的控制器处理）
	middle[last](ctx)
}
