package crawler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
)

// CollyBrowser Colly浏览器实现
type CollyBrowser struct {
	collector *colly.Collector
	config    *BrowserConfig
	client    *http.Client
}

// NewCollyBrowser 创建新的Colly浏览器实例
func NewCollyBrowser(config *BrowserConfig) (*CollyBrowser, error) {
	if config == nil {
		config = DefaultConfig()
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	collector := colly.NewCollector(
		colly.UserAgent(config.UserAgent),
		colly.AllowURLRevisit(),
	)

	collector.SetRequestTimeout(config.Timeout)

	if config.Proxy != "" {
		collector.SetProxy(config.Proxy)
	}

	for key, value := range config.Headers {
		collector.OnRequest(func(r *colly.Request) {
			r.Headers.Set(key, value)
		})
	}

	if len(config.Cookies) > 0 {
		collector.OnRequest(func(r *colly.Request) {
			for key, value := range config.Cookies {
				r.Headers.Set("Cookie", fmt.Sprintf("%s=%s", key, value))
			}
		})
	}

	if config.FollowRedirects {
		collector.SetRedirectHandler(func(req *http.Request, via []*http.Request) error {
			return nil
		})
	}

	extensions.RandomUserAgent(collector)

	return &CollyBrowser{
		collector: collector,
		config:    config,
		client:    client,
	}, nil
}

// Get 发送GET请求
func (c *CollyBrowser) Get(url string) (*Response, error) {
	var response *Response
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
func (c *CollyBrowser) doGet(url string) (*Response, error) {
	var responseBody []byte
	var statusCode int
	var responseHeaders map[string]string

	c.collector.OnResponse(func(r *colly.Response) {
		responseBody = r.Body
		statusCode = r.StatusCode
		responseHeaders = make(map[string]string)
		// 暂时跳过Headers处理，避免类型问题
	})

	err := c.collector.Visit(url)
	if err != nil {
		return nil, fmt.Errorf("failed to visit URL %s: %w", url, err)
	}

	return &Response{
		StatusCode: statusCode,
		Headers:    responseHeaders,
		Body:       responseBody,
		URL:        url,
		Cookies:    make(map[string]string),
	}, nil
}

// Post 发送POST请求
func (c *CollyBrowser) Post(url string, data map[string]interface{}) (*Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.config.UserAgent)

	for key, value := range c.config.Headers {
		req.Header.Set(key, value)
	}

	if len(c.config.Cookies) > 0 {
		for key, value := range c.config.Cookies {
			req.AddCookie(&http.Cookie{Name: key, Value: value})
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	cookies := make(map[string]string)
	for _, cookie := range resp.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       body,
		URL:        url,
		Cookies:    cookies,
	}, nil
}

// SetHeaders 设置请求头
func (c *CollyBrowser) SetHeaders(headers map[string]string) {
	c.config.Headers = headers
	c.collector.OnRequest(func(r *colly.Request) {
		for key, value := range headers {
			r.Headers.Set(key, value)
		}
	})
}

// SetCookies 设置Cookie
func (c *CollyBrowser) SetCookies(cookies map[string]string) {
	c.config.Cookies = cookies
	c.collector.OnRequest(func(r *colly.Request) {
		for key, value := range cookies {
			r.Headers.Set("Cookie", fmt.Sprintf("%s=%s", key, value))
		}
	})
}

// SetTimeout 设置超时时间
func (c *CollyBrowser) SetTimeout(timeout time.Duration) {
	c.config.Timeout = timeout
	c.collector.SetRequestTimeout(timeout)
	c.client.Timeout = timeout
}

// SetProxy 设置代理
func (c *CollyBrowser) SetProxy(proxy string) {
	c.config.Proxy = proxy
	c.collector.SetProxy(proxy)
}

// SetUserAgent 设置User-Agent
func (c *CollyBrowser) SetUserAgent(userAgent string) {
	c.config.UserAgent = userAgent
	c.collector.UserAgent = userAgent
}

// SetRandomUserAgent 设置随机User-Agent
func (c *CollyBrowser) SetRandomUserAgent() {
	randomUA := browser.Random()
	c.SetUserAgent(randomUA)
}

// GetUserAgent 获取当前User-Agent
func (c *CollyBrowser) GetUserAgent() string {
	return c.config.UserAgent
}

// Close 关闭浏览器实例
func (c *CollyBrowser) Close() error {
	return nil
}
