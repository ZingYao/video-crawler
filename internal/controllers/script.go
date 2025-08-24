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
	jsTestService  services.JSTestService
}

func NewLuaTestController(luaTestService services.LuaTestService) *LuaTestController {
	return &LuaTestController{
		luaTestService: luaTestService,
		jsTestService:  services.NewJSTestService(),
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

// TestJSScript JS 脚本流式调试
func (c *LuaTestController) TestJSScript(ctx *gin.Context) {
	if ctx.Request.Method != "POST" {
		utils.SendResponse(ctx, http.StatusMethodNotAllowed, "只支持POST方法", nil)
		return
	}
	var request struct {
		Script string `json:"script" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "参数错误: "+err.Error(), nil)
		return
	}
	ctx.Header("Content-Type", "text/plain; charset=utf-8")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Transfer-Encoding", "chunked")
	ctx.Status(http.StatusOK)
	reqCtx := ctx.Request.Context()
	if ua := ctx.GetHeader("User-Agent"); ua != "" {
		reqCtx = context.WithValue(reqCtx, services.CtxKeyRequestUA, ua)
	}
	ch, err := c.jsTestService.ExecuteScript(reqCtx, request.Script)
	if err != nil {
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("[ERROR] 启动脚本执行失败: %v\n", err))
		return
	}
	writer := ctx.Writer
	flusher, ok := writer.(http.Flusher)
	if !ok {
		ctx.String(http.StatusInternalServerError, "[ERROR] 服务器不支持流式响应\n")
		return
	}
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				writer.Write([]byte("[END] 脚本执行结束\n"))
				flusher.Flush()
				return
			}
			writer.Write([]byte(msg + "\n"))
			flusher.Flush()
		case <-ctx.Request.Context().Done():
			return
		}
	}
}

// AdvancedTestJSScript JS 高级调试
func (c *LuaTestController) AdvancedTestJSScript(ctx *gin.Context) {
	if ctx.Request.Method != "POST" {
		utils.SendResponse(ctx, http.StatusMethodNotAllowed, "只支持POST方法", nil)
		return
	}

	var request struct {
		Script string                 `json:"script" binding:"required"`
		Method string                 `json:"method" binding:"required"`
		Params map[string]interface{} `json:"params" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "参数错误: "+err.Error(), nil)
		return
	}

	// 验证方法类型
	validMethods := map[string]bool{
		"search_video":           true,
		"get_video_detail":       true,
		"get_play_video_detail":  true,
	}
	if !validMethods[request.Method] {
		utils.SendResponse(ctx, http.StatusBadRequest, "不支持的方法类型: "+request.Method, nil)
		return
	}

	// 设置上下文
	reqCtx := ctx.Request.Context()
	if ua := ctx.GetHeader("User-Agent"); ua != "" {
		reqCtx = context.WithValue(reqCtx, services.CtxKeyRequestUA, ua)
	}

	// 执行高级调试
	result, consoleOutput, err := c.jsTestService.ExecuteAdvancedTest(reqCtx, request.Script, request.Method, request.Params)
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "执行失败: "+err.Error(), nil)
		return
	}

	// 返回结果
	response := map[string]interface{}{
		"original":  result.Original,
		"converted": result.Converted,
		"console":   consoleOutput,
	}

	utils.SendResponse(ctx, http.StatusOK, "执行成功", response)
}

// AdvancedTestLuaScript Lua 高级调试
func (c *LuaTestController) AdvancedTestLuaScript(ctx *gin.Context) {
	if ctx.Request.Method != "POST" {
		utils.SendResponse(ctx, http.StatusMethodNotAllowed, "只支持POST方法", nil)
		return
	}

	var request struct {
		Script string                 `json:"script" binding:"required"`
		Method string                 `json:"method" binding:"required"`
		Params map[string]interface{} `json:"params" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "参数错误: "+err.Error(), nil)
		return
	}

	// 验证方法类型
	validMethods := map[string]bool{
		"search_video":           true,
		"get_video_detail":       true,
		"get_play_video_detail":  true,
	}
	if !validMethods[request.Method] {
		utils.SendResponse(ctx, http.StatusBadRequest, "不支持的方法类型: "+request.Method, nil)
		return
	}

	// 设置上下文
	reqCtx := ctx.Request.Context()
	if ua := ctx.GetHeader("User-Agent"); ua != "" {
		reqCtx = context.WithValue(reqCtx, services.CtxKeyRequestUA, ua)
	}

	// 执行高级调试
	result, consoleOutput, err := c.luaTestService.ExecuteAdvancedTest(reqCtx, request.Script, request.Method, request.Params)
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "执行失败: "+err.Error(), nil)
		return
	}

	// 返回结果
	response := map[string]interface{}{
		"original":  result.Original,
		"converted": result.Converted,
		"console":   consoleOutput,
	}

	utils.SendResponse(ctx, http.StatusOK, "执行成功", response)
}

// jsonEscape 转义JSON字符串
func jsonEscape(s string) string {
	escaped, _ := json.Marshal(s)
	return string(escaped)
}

// AdvancedTestJSScriptSSE JS 高级调试(SSE)
func (c *LuaTestController) AdvancedTestJSScriptSSE(ctx *gin.Context) {
	if ctx.Request.Method != "GET" {
		utils.SendResponse(ctx, http.StatusMethodNotAllowed, "只支持GET方法", nil)
		return
	}

	// 从查询参数获取数据
	script := ctx.Query("script")
	method := ctx.Query("method")
	paramsStr := ctx.Query("params")

	if script == "" || method == "" || paramsStr == "" {
		utils.SendResponse(ctx, http.StatusBadRequest, "缺少必要参数", nil)
		return
	}

	// 解析参数
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "参数格式错误", nil)
		return
	}

	// 验证方法类型
	validMethods := map[string]bool{
		"search_video":           true,
		"get_video_detail":       true,
		"get_play_video_detail":  true,
	}
	if !validMethods[method] {
		utils.SendResponse(ctx, http.StatusBadRequest, "不支持的方法类型: "+method, nil)
		return
	}

	// 设置SSE响应头
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Headers", "Cache-Control")

	// 设置上下文
	reqCtx := ctx.Request.Context()
	if ua := ctx.GetHeader("User-Agent"); ua != "" {
		reqCtx = context.WithValue(reqCtx, services.CtxKeyRequestUA, ua)
	}

	// 获取输出通道
	outputChan, err := c.jsTestService.ExecuteAdvancedTestSSE(reqCtx, script, method, params)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "event: error\ndata: {\"message\":\"%s\"}\n\n", err.Error())
		return
	}

	// 创建响应写入器
	writer := ctx.Writer
	flusher, ok := writer.(http.Flusher)
	if !ok {
		ctx.String(http.StatusInternalServerError, "event: error\ndata: {\"message\":\"服务器不支持流式响应\"}\n\n")
		return
	}

	// 发送连接建立事件
	writer.Write([]byte("event: connected\ndata: {\"message\":\"连接已建立\"}\n\n"))
	flusher.Flush()

	// 流式返回输出
	for {
		select {
		case msg, ok := <-outputChan:
			if !ok {
				// 通道关闭，发送完成事件
				writer.Write([]byte("event: complete\ndata: {\"message\":\"执行完成\"}\n\n"))
				flusher.Flush()
				return
			}

			// 发送消息
			writer.Write([]byte(msg + "\n"))
			flusher.Flush()

		case <-ctx.Request.Context().Done():
			// 客户端断开连接
			return
		}
	}
}

// AdvancedTestLuaScriptSSE Lua 高级调试(SSE)
func (c *LuaTestController) AdvancedTestLuaScriptSSE(ctx *gin.Context) {
	if ctx.Request.Method != "GET" {
		utils.SendResponse(ctx, http.StatusMethodNotAllowed, "只支持GET方法", nil)
		return
	}

	// 从查询参数获取数据
	script := ctx.Query("script")
	method := ctx.Query("method")
	paramsStr := ctx.Query("params")

	if script == "" || method == "" || paramsStr == "" {
		utils.SendResponse(ctx, http.StatusBadRequest, "缺少必要参数", nil)
		return
	}

	// 解析参数
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "参数格式错误", nil)
		return
	}

	// 验证方法类型
	validMethods := map[string]bool{
		"search_video":           true,
		"get_video_detail":       true,
		"get_play_video_detail":  true,
	}
	if !validMethods[method] {
		utils.SendResponse(ctx, http.StatusBadRequest, "不支持的方法类型: "+method, nil)
		return
	}

	// 设置SSE响应头
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Headers", "Cache-Control")

	// 设置上下文
	reqCtx := ctx.Request.Context()
	if ua := ctx.GetHeader("User-Agent"); ua != "" {
		reqCtx = context.WithValue(reqCtx, services.CtxKeyRequestUA, ua)
	}

	// 获取输出通道
	outputChan, err := c.luaTestService.ExecuteAdvancedTestSSE(reqCtx, script, method, params)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "event: error\ndata: {\"message\":\"%s\"}\n\n", err.Error())
		return
	}

	// 创建响应写入器
	writer := ctx.Writer
	flusher, ok := writer.(http.Flusher)
	if !ok {
		ctx.String(http.StatusInternalServerError, "event: error\ndata: {\"message\":\"服务器不支持流式响应\"}\n\n")
		return
	}

	// 发送连接建立事件
	writer.Write([]byte("event: connected\ndata: {\"message\":\"连接已建立\"}\n\n"))
	flusher.Flush()

	// 流式返回输出
	for {
		select {
		case msg, ok := <-outputChan:
			if !ok {
				// 通道关闭，发送完成事件
				writer.Write([]byte("event: complete\ndata: {\"message\":\"执行完成\"}\n\n"))
				flusher.Flush()
				return
			}

			// 发送消息
			writer.Write([]byte(msg + "\n"))
			flusher.Flush()

		case <-ctx.Request.Context().Done():
			// 客户端断开连接
			return
		}
	}
}
