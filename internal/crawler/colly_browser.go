package crawler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	fakeUserAgent "github.com/lib4u/fake-useragent"
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

	// 设置默认请求头
	collector.OnRequest(func(r *colly.Request) {
		// 设置所有配置的请求头
		for key, value := range config.Headers {
			r.Headers.Set(key, value)
		}

		// 设置Cookie
		if len(config.Cookies) > 0 {
			for key, value := range config.Cookies {
				r.Headers.Set("Cookie", fmt.Sprintf("%s=%s", key, value))
			}
		}
	})

	if config.FollowRedirects {
		collector.SetRedirectHandler(func(req *http.Request, via []*http.Request) error {
			return nil
		})
	}

	// 注意：不在这里调用extensions.RandomUserAgent，因为我们会手动设置UA

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

		//去 body 的转义 和首位的引号
		var body map[string]string
		bodyContent := string(r.Body)
		if strings.HasPrefix(bodyContent, "\"") {
			err := json.Unmarshal([]byte(fmt.Sprintf("{\"body\":%s}", bodyContent)), &body)
			if err != nil {
				body = map[string]string{
					"body": bodyContent,
				}
			}
		} else {
			body = map[string]string{
				"body": bodyContent,
			}
		}
		responseBody = []byte(body["body"])
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
	// 合并新的请求头到现有配置中
	for key, value := range headers {
		c.config.Headers[key] = value
	}
}

// SetCookies 设置Cookie
func (c *CollyBrowser) SetCookies(cookies map[string]string) {
	// 合并新的Cookie到现有配置中
	for key, value := range cookies {
		c.config.Cookies[key] = value
	}
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
func (c *CollyBrowser) GetUserAgent() string {
	return c.config.UserAgent
}

// Close 关闭浏览器实例
func (c *CollyBrowser) Close() error {
	return nil
}
