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
	return &BrowserConfig{
		Timeout:         30 * time.Second,
		UserAgent:       "",
		Proxy:           "",
		Headers:         make(map[string]string),
		Cookies:         make(map[string]string),
		MaxRetries:      3,
		RetryDelay:      1 * time.Second,
		FollowRedirects: true,
	}
}
