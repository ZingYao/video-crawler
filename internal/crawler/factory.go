package crawler

import (
	"fmt"
)

// BrowserType 浏览器类型
type BrowserType string

const (
	// CollyBrowserType Colly浏览器类型
	CollyBrowserType BrowserType = "colly"
)

// NewBrowser 创建新的浏览器实例
func NewBrowser(browserType BrowserType, config *BrowserConfig) (BrowserRequest, error) {
	switch browserType {
	case CollyBrowserType:
		return NewCollyBrowser(config)
	default:
		return nil, fmt.Errorf("unsupported browser type: %s", browserType)
	}
}

// NewDefaultBrowser 创建默认浏览器实例
func NewDefaultBrowser() (BrowserRequest, error) {
	config := DefaultConfig()
	browser, err := NewCollyBrowser(config)
	if err != nil {
		return nil, err
	}
	browser.SetRandomUserAgent()
	return browser, nil
}
