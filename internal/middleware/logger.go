package middleware

import (
	"bytes"
	"time"
	"video-crawler/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// responseWriter 包装原始的 ResponseWriter
type responseWriter struct {
	ctx *gin.Context
	gin.ResponseWriter
	beginTime time.Time
	body      *bytes.Buffer
	oldWrite  func([]byte) (int, error)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	// 在这里可以记录响应内容
	w.body.Write(b)
	// 获取响应信息并记录日志
	logger.CtxLogger(w.ctx).WithFields(logrus.Fields{
		"request": map[string]any{
			"path":   w.ctx.Request.URL.Path,
			"method": w.ctx.Request.Method,
			"params": w.ctx.Request.URL.Query(),
			"body":   w.ctx.Request.Body,
		},
		"response": map[string]any{
			"status_code": w.ctx.Writer.Status(),
			"body":        string(b),
		},
		"duration": time.Since(w.beginTime).Milliseconds(),
	}).Info("request_record")
	return w.oldWrite(b)
}

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 包装 ResponseWriter
		oldWrite := c.Writer.Write
		responseBody := &bytes.Buffer{}
		c.Writer = &responseWriter{
			ResponseWriter: c.Writer,
			body:           responseBody,
			oldWrite:       oldWrite,
			beginTime:      time.Now(),
			ctx:            c,
		}
	}
}
