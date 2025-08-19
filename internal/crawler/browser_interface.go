package crawler

import (
	"time"
)

// BrowserRequest 浏览器请求接口
type BrowserRequest interface {
	// Get 发送GET请求
	Get(url string) (*Response, error)

	// Post 发送POST请求
	Post(url string, data map[string]interface{}) (*Response, error)

	// SetHeaders 设置请求头
	SetHeaders(headers map[string]string)

	// SetCookies 设置Cookie
	SetCookies(cookies map[string]string)

	// SetTimeout 设置超时时间
	SetTimeout(timeout time.Duration)

	// SetProxy 设置代理
	SetProxy(proxy string)

	// SetUserAgent 设置User-Agent
	SetUserAgent(userAgent string)

	// SetRandomUserAgent 设置随机User-Agent
	SetRandomUserAgent()

	// GetUserAgent 获取当前User-Agent
	GetUserAgent() string

	// Close 关闭浏览器实例
	Close() error
}

// Response 响应结构
type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
	URL        string
	Cookies    map[string]string
}

// BrowserConfig 浏览器配置
type BrowserConfig struct {
	Timeout         time.Duration
	UserAgent       string
	Proxy           string
	Headers         map[string]string
	Cookies         map[string]string
	MaxRetries      int
	RetryDelay      time.Duration
	FollowRedirects bool
}

// DefaultConfig 默认配置
func DefaultConfig() *BrowserConfig {
	// 设置真实的浏览器请求头
	headers := map[string]string{
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
	}

	return &BrowserConfig{
		Timeout:         30 * time.Second,
		UserAgent:       "",
		Proxy:           "",
		Headers:         headers,
		Cookies:         make(map[string]string),
		MaxRetries:      3,
		RetryDelay:      1 * time.Second,
		FollowRedirects: true,
	}
}
