package crawler

import (
	"net/http"
	"time"

	fakeUserAgent "github.com/lib4u/fake-useragent"
)

// BrowserRequest 浏览器请求接口
type BrowserRequest interface {
	// Get 发送GET请求
	Get(url string) (*http.Response, error)

	// Post 发送POST请求
	Post(url string, data map[string]interface{}) (*http.Response, error)

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
	ua, _ := fakeUserAgent.New()
	var userAgent string
	if ua != nil {
		userAgent = ua.GetRandom()
	}
	if userAgent == "" {
		userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36"
	}
	headers := map[string]string{
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Accept-Encoding":           "gzip, deflate, br, zstd",
		"Accept-Language":           "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
		"Cache-Control":             "no-cache",
		"Connection":                "keep-alive",
		"Pragma":                    "no-cache",
		"Upgrade-Insecure-Requests": "1",
		"User-Agent":                userAgent,
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
