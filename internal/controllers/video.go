package controllers

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"video-crawler/internal/crawler"
	"video-crawler/internal/entities"
	"video-crawler/internal/jsengine"
	"video-crawler/internal/lua"
	"video-crawler/internal/services"
	"video-crawler/internal/utils"

	"github.com/gin-gonic/gin"
)

// VideoController 提供视频搜索 / 详情 / 播放地址能力
type VideoController struct {
	videoSourceService services.VideoSourceService
	historyService     services.HistoryService
	userService        services.UserServiceInterface
}

func NewVideoController(videoSourceService services.VideoSourceService, historyService services.HistoryService, userService services.UserServiceInterface) *VideoController {
	return &VideoController{videoSourceService: videoSourceService, historyService: historyService, userService: userService}
}

// Search 视频搜索
// GET /api/video/search?source_id=xxx&keyword=yyy
func (c *VideoController) Search(ctx *gin.Context) {
	sourceID := ctx.Query("source_id")
	keyword := ctx.Query("keyword")
	if sourceID == "" || keyword == "" {
		utils.SendResponse(ctx, http.StatusBadRequest, "参数错误: source_id 与 keyword 不能为空", nil)
		return
	}

	videoSource, err := c.videoSourceService.Detail(sourceID)
	if err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "获取视频源失败: "+err.Error(), nil)
		return
	}

	data, err := c.executeByEngine(ctx, &videoSource, "search_video", keyword)
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// 验证并规范化搜索结果
	validResults, err := entities.ValidateSearchVideoResult(data)
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "搜索结果格式错误: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(ctx, validResults)
}

// Detail 视频详情
// GET /api/video/detail?source_id=xxx&url=yyy
func (c *VideoController) Detail(ctx *gin.Context) {
	sourceID := ctx.Query("source_id")
	url := ctx.Query("url")
	if sourceID == "" || url == "" {
		utils.SendResponse(ctx, http.StatusBadRequest, "参数错误: source_id 与 url 不能为空", nil)
		return
	}

	videoSource, err := c.videoSourceService.Detail(sourceID)
	if err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "获取视频源失败: "+err.Error(), nil)
		return
	}

	data, err := c.executeByEngine(ctx, &videoSource, "get_video_detail", url)
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// 验证并规范化视频详情结果
	validResult, err := entities.ValidateVideoDetailResult(data)
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "视频详情格式错误: "+err.Error(), nil)
		return
	}
	// 使用验证后的结果
	utils.SuccessResponse(ctx, validResult)

	// 异步记录观看历史，失败不影响接口返回
	go func(sourceIDCopy, urlCopy string, dataCopy interface{}) {
		userIDVal, exists := ctx.Get("user_id")
		if !exists {
			return
		}
		userID, _ := userIDVal.(string)
		userEntity, ok := c.userService.UserDetailInner(userID)
		if !ok {
			return
		}

		videoTitle := ctx.Query("title")
		if m, ok := dataCopy.(map[string]interface{}); ok {
			if v, ok := m["name"]; ok && fmt.Sprint(v) != "" {
				videoTitle = fmt.Sprint(v)
			} else if v, ok := m["title"]; ok && fmt.Sprint(v) != "" {
				videoTitle = fmt.Sprint(v)
			}
		}

		h := md5.Sum([]byte(sourceIDCopy + "|" + urlCopy))
		videoID := hex.EncodeToString(h[:])

		_ = c.historyService.AddVideoHistory(
			ctx,
			&userEntity,
			videoID,
			videoTitle,
			urlCopy,
			sourceIDCopy,
			videoSource.Name,
			0,
			0,
		)
	}(sourceID, url, validResult)
}

// PlayURL 获取可播放地址
// GET /api/video/url?source_id=xxx&url=yyy
func (c *VideoController) PlayURL(ctx *gin.Context) {
	sourceID := ctx.Query("source_id")
	url := ctx.Query("url")
	if sourceID == "" || url == "" {
		utils.SendResponse(ctx, http.StatusBadRequest, "参数错误: source_id 与 url 不能为空", nil)
		return
	}

	videoSource, err := c.videoSourceService.Detail(sourceID)
	if err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "获取视频源失败: "+err.Error(), nil)
		return
	}

	data, err := c.executeByEngine(ctx, &videoSource, "get_play_video_detail", url)
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// 验证并规范化播放详情结果
	validResult, err := entities.ValidatePlayVideoDetailResult(data)
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "播放详情格式错误: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(ctx, validResult)
}

// executeLuaFunction 组合 Lua 脚本并执行指定函数，返回其返回的数据表
func executeLuaFunction(ctx *gin.Context, baseScript string, funcName string, arg string) (interface{}, error) {
	// 创建浏览器
	browser, err := crawler.NewDefaultBrowser()
	if err != nil {
		return nil, fmt.Errorf("创建浏览器实例失败: %w", err)
	}
	defer browser.Close()

	// 设置 UA：优先沿用前端请求头 UA
	if ua := ctx.GetHeader("User-Agent"); ua != "" {
		browser.SetUserAgent(ua)
	} else {
		browser.SetRandomUserAgent()
	}
	// 设置常用头
	browser.SetHeaders(map[string]string{
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Accept-Language":           "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
		"Accept-Encoding":           "gzip, deflate, br, zstd",
		"Cache-Control":             "max-age=0",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1",
		"Sec-Fetch-Dest":            "document",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-Site":            "none",
		"Sec-Fetch-User":            "?1",
		"sec-ch-ua":                 `"Not;A=Brand";v="99", "Microsoft Edge";v="139", "Chromium";v="139"`,
		"sec-ch-ua-mobile":          "?0",
		"sec-ch-ua-platform":        `"macOS"`,
		"DNT":                       "1",
	})

	// 组合脚本：原脚本 + 调用指定函数并 return { data=..., err=... }
	argLiteral := strconv.Quote(arg) // 作为 Lua 字面量字符串使用
	wrapped := baseScript + "\n\n" +
		fmt.Sprintf("local __data, __err = %s(%s)\nreturn { data = __data, err = __err }\n", funcName, argLiteral)

	// 执行
	engine := lua.NewLuaEngineWithContext(browser, ctx)
	defer engine.Close()
	ret, execErr := engine.Execute(wrapped)
	if execErr != nil {
		return nil, fmt.Errorf("脚本执行失败: %w", execErr)
	}

	// 解析返回
	if v, ok := ret["err"]; ok && v != nil && fmt.Sprint(v) != "" {
		return nil, fmt.Errorf("脚本返回错误: %v", v)
	}
	return ret["data"], nil
}

// executeByEngine 根据站点 engine_type 调用 Lua 或 JS 引擎
func (c *VideoController) executeByEngine(ctx *gin.Context, src *entities.VideoSourceEntity, funcName string, arg string) (interface{}, error) {
	if src.EngineType == 1 {
		// JS 引擎
		browser, err := crawler.NewDefaultBrowser()
		if err != nil {
			return nil, fmt.Errorf("创建浏览器实例失败: %w", err)
		}
		defer browser.Close()
		if ua := ctx.GetHeader("User-Agent"); ua != "" {
			browser.SetUserAgent(ua)
		} else {
			browser.SetRandomUserAgent()
		}
		e := jsengine.NewWithContext(browser, ctx)
		argLiteral := strconv.Quote(arg)
		wrapped := src.JsScript + "\n\n" +
			fmt.Sprintf("var __ret = (function(){ try { var r = %s(%s); return {data: r, err: null}; } catch(e){ return {data:null, err: String(e)} } })(); __ret;", funcName, argLiteral)
		m, err := e.ExecuteWrapped(wrapped)
		if err != nil {
			return nil, err
		}
		if v, ok := m["err"]; ok && v != nil && fmt.Sprint(v) != "" {
			return nil, fmt.Errorf("脚本返回错误: %v", v)
		}
		return m["data"], nil
	}
	// 默认 Lua
	return executeLuaFunction(ctx, src.LuaScript, funcName, arg)
}
