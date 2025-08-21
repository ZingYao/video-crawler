package jsengine

import (
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dop251/goja"

	"video-crawler/internal/crawler"
)

// Engine 提供最小可用的 JS 执行环境（goja）
type Engine struct {
	vm      *goja.Runtime
	browser crawler.BrowserRequest
}

func New(browser crawler.BrowserRequest) *Engine {
	vm := goja.New()
	e := &Engine{vm: vm, browser: browser}
	e.bindApis()
	return e
}

// bindApis 绑定可用的全局函数到 JS
func (e *Engine) bindApis() {
	// JSON 助手
	e.vm.Set("json_encode", func(v interface{}) string {
		b, _ := json.Marshal(v)
		return string(b)
	})
	e.vm.Set("json_decode", func(s string) interface{} {
		var v interface{}
		_ = json.Unmarshal([]byte(s), &v)
		return v
	})

	// HTTP
	e.vm.Set("http_get", func(url string) map[string]interface{} {
		resp, err := e.browser.Get(url)
		if err != nil {
			return map[string]interface{}{"status_code": 0, "body": "", "url": url, "err": err.Error()}
		}
		defer resp.Body.Close()
		body, _ := readDecompressedBody(resp)
		headers := map[string]string{}
		for k, v := range resp.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}
		return map[string]interface{}{
			"status_code": resp.StatusCode,
			"body":        string(body),
			"url":         resp.Request.URL.String(),
			"headers":     headers,
		}
	})
	e.vm.Set("set_headers", func(m map[string]string) { e.browser.SetHeaders(m) })
	e.vm.Set("set_user_agent", func(ua string) { e.browser.SetUserAgent(ua) })
	e.vm.Set("set_random_user_agent", func() { e.browser.SetRandomUserAgent() })
}

// ExecuteWrapped 执行完整脚本文本，返回其最后一个表达式的值（用于 {data,err} 对象）
func (e *Engine) ExecuteWrapped(script string) (map[string]interface{}, error) {
	v, err := e.vm.RunString(script)
	if err != nil {
		return nil, fmt.Errorf("execute js error: %w", err)
	}
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return map[string]interface{}{}, nil
	}
	got := v.Export()
	if m, ok := got.(map[string]interface{}); ok {
		return m, nil
	}
	// 兜底：尝试 JSON 解析
	b, _ := json.Marshal(got)
	var out map[string]interface{}
	_ = json.Unmarshal(b, &out)
	if out == nil {
		out = map[string]interface{}{}
	}
	return out, nil
}

// readDecompressedBody 与 Lua 引擎一致：自动解压 gzip/deflate
func readDecompressedBody(resp *http.Response) ([]byte, error) {
	enc := strings.ToLower(strings.TrimSpace(resp.Header.Get("Content-Encoding")))
	if enc == "" || enc == "identity" {
		return io.ReadAll(resp.Body)
	}
	if idx := strings.Index(enc, ","); idx >= 0 {
		enc = strings.TrimSpace(enc[:idx])
	}
	switch enc {
	case "gzip":
		gr, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer gr.Close()
		return io.ReadAll(gr)
	case "deflate":
		zr, err := zlib.NewReader(resp.Body)
		if err == nil {
			defer zr.Close()
			return io.ReadAll(zr)
		}
		fr := flate.NewReader(resp.Body)
		defer fr.Close()
		return io.ReadAll(fr)
	default:
		return io.ReadAll(resp.Body)
	}
}
