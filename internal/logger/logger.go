package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	// 打印调用日志的文件路径
	logrus.SetReportCaller(true)
	// 使用 lestrrat-go/file-rotatelogs 实现日志文件自动过期
	rotator, err := rotatelogs.New(
		"./logs/app.%Y%m%d.log",
		rotatelogs.WithLinkName("./logs/app.log"), // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(7*24*time.Hour),     // 日志保留7天
		rotatelogs.WithRotationTime(24*time.Hour), // 每24小时切割一次
	)
	if err != nil {
		panic(err)
	}
	logrus.SetOutput(rotator)
}

func CtxLogger(ctx *gin.Context) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"request_id": ctx.GetHeader("X-Request-ID"),
		"client_ip":  ctx.ClientIP(),
		"user_name":  ctx.GetString("username"),
		"user_id":    ctx.GetString("user_id"),
	})
}
