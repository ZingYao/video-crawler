package controllers

import (
	"io"
	"strings"
	"video-crawler/internal/consts"
	"video-crawler/internal/crawler"
	"video-crawler/internal/entities"
	"video-crawler/internal/services"
	"video-crawler/internal/utils"

	"github.com/gin-gonic/gin"
)

type VideoSourceController struct {
	videoSourceService services.VideoSourceService
}

func NewVideoSourceController(videoSourceService services.VideoSourceService) *VideoSourceController {
	return &VideoSourceController{videoSourceService: videoSourceService}
}

func (c *VideoSourceController) List(ctx *gin.Context) {
	videoSourceList, err := c.videoSourceService.List()
	if err != nil {
		utils.SendResponse(ctx, consts.ResponseCodeGetVideoSourceListFailed, err.Error(), nil)
		return
	}
	utils.SuccessResponse(ctx, videoSourceList)
}

func (c *VideoSourceController) Detail(ctx *gin.Context) {
	videoSourceId := ctx.Query("id")
	videoSource, err := c.videoSourceService.Detail(videoSourceId)
	if err != nil {
		utils.SendResponse(ctx, consts.ResponseCodeGetVideoSourceDetailFailed, err.Error(), nil)
		return
	}
	utils.SuccessResponse(ctx, videoSource)
}

func (c *VideoSourceController) Save(ctx *gin.Context) {
	var videoSource entities.VideoSourceEntity
	if err := ctx.ShouldBindJSON(&videoSource); err != nil {
		utils.SendResponse(ctx, consts.ResponseCodeParamError, "参数错误: "+err.Error(), nil)
		return
	}

	err := c.videoSourceService.Save(videoSource)
	if err != nil {
		utils.SendResponse(ctx, consts.ResponseCodeSaveVideoSourceFailed, err.Error(), nil)
		return
	}

	utils.SuccessResponse(ctx, gin.H{
		"id":      videoSource.Id,
		"message": "保存成功",
	})
}

func (c *VideoSourceController) Delete(ctx *gin.Context) {
	videoSourceId := ctx.Query("id")
	if videoSourceId == "" {
		utils.SendResponse(ctx, consts.ResponseCodeParamError, "站点ID不能为空", nil)
		return
	}

	err := c.videoSourceService.Delete(videoSourceId)
	if err != nil {
		utils.SendResponse(ctx, consts.ResponseCodeDeleteVideoSourceFailed, err.Error(), nil)
		return
	}

	utils.SuccessResponse(ctx, gin.H{
		"message": "删除成功",
	})
}

func (c *VideoSourceController) CheckStatus(ctx *gin.Context) {
	videoSourceId := ctx.Query("id")
	videoSource, err := c.videoSourceService.Detail(videoSourceId)
	if err != nil {
		utils.SendResponse(ctx, consts.ResponseCodeGetVideoSourceDetailFailed, err.Error(), nil)
		return
	}

	// 创建爬虫浏览器实例
	browser, err := crawler.NewDefaultBrowser()
	if err != nil {
		utils.SendResponse(ctx, consts.ResponseCodeCheckVideoSourceStatusFailed, "创建浏览器实例失败: "+err.Error(), nil)
		return
	}
	defer browser.Close()
	browser.SetRandomUserAgent()

	// 将前端请求头透传到 crawler 请求（合并，不覆盖默认关键头）
	incoming := ctx.Request.Header
	headers := make(map[string]string)
	var ua string
	for key, values := range incoming {
		if len(values) == 0 {
			continue
		}
		val := values[0]
		lk := strings.ToLower(key)
		// 跳过不适合透传或由客户端自动设置的头
		switch lk {
		case "host", "content-length":
			continue
		case "user-agent":
			ua = val
			continue
		case "cookie":
			// 按需求：Cookie 不透传
			continue
		}
		headers[key] = val
	}
	if len(headers) > 0 {
		browser.SetHeaders(headers)
	}
	// 处理 UA
	if strings.TrimSpace(ua) != "" {
		browser.SetUserAgent(ua)
	}

	// 使用爬虫请求域名，如果返回200，则站点正常，否则站点不可用
	resp, err := browser.Get(videoSource.Domain)
	if err != nil {
		utils.SendResponse(ctx, consts.ResponseCodeCheckVideoSourceStatusFailed, err.Error(), nil)
		return
	}
	defer resp.Body.Close()

	// 读取响应体以确保连接正常
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		utils.SendResponse(ctx, consts.ResponseCodeCheckVideoSourceStatusFailed, "读取响应失败: "+err.Error(), nil)
		return
	}

	if resp.StatusCode == 200 {
		videoSource.Status = consts.VideoSourceStatusNormal
	} else {
		videoSource.Status = consts.VideoSourceStatusUnavailable
	}
	err = c.videoSourceService.Save(videoSource)
	if err != nil {
		utils.SendResponse(ctx, consts.ResponseCodeSaveVideoSourceFailed, err.Error(), nil)
		return
	}
	utils.SuccessResponse(ctx, videoSource.Status)
}
