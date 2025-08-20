package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"video-crawler/internal/services"
	"video-crawler/internal/utils"

	"github.com/gin-gonic/gin"
)

type LuaTestController struct {
	luaTestService services.LuaTestService
}

func NewLuaTestController(luaTestService services.LuaTestService) *LuaTestController {
	return &LuaTestController{
		luaTestService: luaTestService,
	}
}

// TestScript 测试Lua脚本接口
func (c *LuaTestController) TestScript(ctx *gin.Context) {
	// 检查请求方法
	if ctx.Request.Method != "POST" {
		utils.SendResponse(ctx, http.StatusMethodNotAllowed, "只支持POST方法", nil)
		return
	}

	// 解析请求体
	var request struct {
		Script string `json:"script" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "参数错误: "+err.Error(), nil)
		return
	}

	// 设置响应头为流式传输
	ctx.Header("Content-Type", "text/plain; charset=utf-8")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Transfer-Encoding", "chunked")

	// 设置响应状态码
	ctx.Status(http.StatusOK)

	// 从请求头提取 UA，写入上下文
	ua := ctx.GetHeader("User-Agent")
	reqCtx := ctx.Request.Context()
	if ua != "" {
		reqCtx = context.WithValue(reqCtx, services.CtxKeyRequestUA, ua)
	}

	// 获取输出通道
	outputChan, err := c.luaTestService.ExecuteScript(reqCtx, request.Script)
	if err != nil {
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("[ERROR] 启动脚本执行失败: %v\n", err))
		return
	}

	// 创建响应写入器
	writer := ctx.Writer
	flusher, ok := writer.(http.Flusher)
	if !ok {
		ctx.String(http.StatusInternalServerError, "[ERROR] 服务器不支持流式响应\n")
		return
	}

	// 流式返回输出
	for {
		select {
		case msg, ok := <-outputChan:
			if !ok {
				// 通道关闭，发送结束标记
				writer.Write([]byte("[END] 脚本执行结束\n"))
				flusher.Flush()
				return
			}

			// 发送消息并刷新
			writer.Write([]byte(msg + "\n"))
			flusher.Flush()

		case <-ctx.Request.Context().Done():
			// 客户端断开连接
			return
		}
	}
}

// TestScriptSSE 使用Server-Sent Events的测试接口
func (c *LuaTestController) TestScriptSSE(ctx *gin.Context) {
	// 检查请求方法
	if ctx.Request.Method != "POST" {
		utils.SendResponse(ctx, http.StatusMethodNotAllowed, "只支持POST方法", nil)
		return
	}

	// 解析请求体
	var request struct {
		Script string `json:"script" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "参数错误: "+err.Error(), nil)
		return
	}

	// 设置SSE响应头
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Headers", "Cache-Control")

	// 设置响应状态码
	ctx.Status(http.StatusOK)

	// 从请求头提取 UA，写入上下文
	ua := ctx.GetHeader("User-Agent")
	reqCtx := ctx.Request.Context()
	if ua != "" {
		reqCtx = context.WithValue(reqCtx, services.CtxKeyRequestUA, ua)
	}

	// 获取输出通道
	outputChan, err := c.luaTestService.ExecuteScript(reqCtx, request.Script)
	if err != nil {
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("event: error\ndata: %s\n\n", err.Error()))
		return
	}

	// 创建响应写入器
	writer := ctx.Writer
	flusher, ok := writer.(http.Flusher)
	if !ok {
		ctx.String(http.StatusInternalServerError, "event: error\ndata: 服务器不支持流式响应\n\n")
		return
	}

	// 发送连接建立事件
	writer.Write([]byte("event: connected\ndata: 连接已建立\n\n"))
	flusher.Flush()

	// 流式返回输出
	for {
		select {
		case msg, ok := <-outputChan:
			if !ok {
				// 通道关闭，发送结束事件
				writer.Write([]byte("event: end\ndata: 脚本执行结束\n\n"))
				flusher.Flush()
				return
			}

			// 若是结果行，作为独立 result 事件（支持多行JSON）
			if strings.HasPrefix(msg, "[RESULT] ") {
				payload := strings.TrimPrefix(msg, "[RESULT] ")
				writer.Write([]byte("event: result\n"))
				for _, line := range strings.Split(payload, "\n") {
					writer.Write([]byte("data: " + line + "\n"))
				}
				writer.Write([]byte("\n"))
				flusher.Flush()
				continue
			}

			// 发送消息事件
			eventData := fmt.Sprintf("event: message\ndata: %s\n\n", jsonEscape(msg))
			writer.Write([]byte(eventData))
			flusher.Flush()

		case <-ctx.Request.Context().Done():
			// 客户端断开连接
			return
		}
	}
}

// jsonEscape 转义JSON字符串
func jsonEscape(s string) string {
	escaped, _ := json.Marshal(s)
	return string(escaped)
}
