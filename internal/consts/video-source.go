package consts

const (
	VideoSourceStatusDisabled = iota
	VideoSourceStatusNormal
	VideoSourceStatusUnavailable
)

// CrawlerEngine 爬虫脚本类型
const (
	CrawlerEngineLua        = iota // Lua 脚本
	CrawlerEngineJavaScript        // JavaScript (goja)
)
