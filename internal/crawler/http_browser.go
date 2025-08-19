package crawler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	fakeUserAgent "github.com/lib4u/fake-useragent"
)

// HTTPBrowser HTTP浏览器实现（使用原生net/http）
type HTTPBrowser struct {
	client *http.Client
	config *BrowserConfig
}

// NewHTTPBrowser 创建新的HTTP浏览器实例
func NewHTTPBrowser(config *BrowserConfig) (*HTTPBrowser, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: config.Timeout,
	}

	// 设置重定向策略
	if config.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return nil
		}
	} else {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// 设置代理
	if config.Proxy != "" {
		proxyURL, err := url.Parse(config.Proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	return &HTTPBrowser{
		client: client,
		config: config,
	}, nil
}

// Get 发送GET请求
func (c *HTTPBrowser) Get(url string) (*http.Response, error) {
	var response *http.Response
	var err error

	for i := 0; i <= c.config.MaxRetries; i++ {
		response, err = c.doGet(url)
		if err == nil {
			break
		}

		if i < c.config.MaxRetries {
			time.Sleep(c.config.RetryDelay)
		}
	}

	return response, err
}

// doGet 执行GET请求
func (c *HTTPBrowser) doGet(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置User-Agent
	if c.config.UserAgent != "" {
		req.Header.Set("User-Agent", c.config.UserAgent)
	}

	// 设置请求头
	for key, value := range c.config.Headers {
		req.Header.Set(key, value)
	}

	// 设置Cookie
	if len(c.config.Cookies) > 0 {
		cookieStrings := make([]string, 0, len(c.config.Cookies))
		for key, value := range c.config.Cookies {
			cookieStrings = append(cookieStrings, fmt.Sprintf("%s=%s", key, value))
		}
		req.Header.Set("Cookie", strings.Join(cookieStrings, "; "))
	}

	// 执行请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, nil
}

// Post 发送POST请求
func (c *HTTPBrowser) Post(url string, data map[string]interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置Content-Type
	req.Header.Set("Content-Type", "application/json")

	// 设置User-Agent
	if c.config.UserAgent != "" {
		req.Header.Set("User-Agent", c.config.UserAgent)
	}

	// 设置请求头
	for key, value := range c.config.Headers {
		req.Header.Set(key, value)
	}

	// 设置Cookie
	if len(c.config.Cookies) > 0 {
		cookieStrings := make([]string, 0, len(c.config.Cookies))
		for key, value := range c.config.Cookies {
			cookieStrings = append(cookieStrings, fmt.Sprintf("%s=%s", key, value))
		}
		req.Header.Set("Cookie", strings.Join(cookieStrings, "; "))
	}

	// 执行请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, nil
}

// SetHeaders 设置请求头
func (c *HTTPBrowser) SetHeaders(headers map[string]string) {
	// 合并新的请求头到现有配置中
	for key, value := range headers {
		c.config.Headers[key] = value
	}
}

// SetCookies 设置Cookie
func (c *HTTPBrowser) SetCookies(cookies map[string]string) {
	// 合并新的Cookie到现有配置中
	for key, value := range cookies {
		c.config.Cookies[key] = value
	}
}

// SetTimeout 设置超时时间
func (c *HTTPBrowser) SetTimeout(timeout time.Duration) {
	c.config.Timeout = timeout
	c.client.Timeout = timeout
}

// SetProxy 设置代理
func (c *HTTPBrowser) SetProxy(proxy string) {
	c.config.Proxy = proxy
	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err == nil {
			c.client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}
	} else {
		c.client.Transport = &http.Transport{}
	}
}

// SetUserAgent 设置User-Agent
func (c *HTTPBrowser) SetUserAgent(userAgent string) {
	if c.config.Headers == nil {
		c.config.Headers = make(map[string]string)
	}
	c.config.UserAgent = userAgent
	c.config.Headers["User-Agent"] = userAgent
}

// SetRandomUserAgent 设置随机User-Agent
func (c *HTTPBrowser) SetRandomUserAgent() {
	ua, err := fakeUserAgent.New()
	if err != nil {
		// 如果创建失败，使用默认UA
		c.SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
		return
	}
	randomUA := ua.GetRandom()
	c.SetUserAgent(randomUA)
}

// GetUserAgent 获取当前User-Agent
func (c *HTTPBrowser) GetUserAgent() string {
	return c.config.UserAgent
}

// Close 关闭浏览器实例
func (c *HTTPBrowser) Close() error {
	// HTTP客户端不需要显式关闭
	return nil
}
