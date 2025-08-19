package crawler

import (
	"fmt"
)

// BrowserType 浏览器类型
type BrowserType string

const (
	// HTTPBrowserType HTTP浏览器类型
	HTTPBrowserType BrowserType = "http"
)

// NewBrowser 创建新的浏览器实例
func NewBrowser(browserType BrowserType, config *BrowserConfig) (BrowserRequest, error) {
	switch browserType {
	case HTTPBrowserType:
		return NewHTTPBrowser(config)
	default:
		return nil, fmt.Errorf("unsupported browser type: %s", browserType)
	}
}

// NewDefaultBrowser 创建默认浏览器实例
func NewDefaultBrowser() (BrowserRequest, error) {
	config := DefaultConfig()
	browser, err := NewHTTPBrowser(config)
	if err != nil {
		return nil, err
	}
	browser.SetRandomUserAgent()
	return browser, nil
}
