package jsengine

import (
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/dop251/goja"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"video-crawler/internal/crawler"
	"video-crawler/internal/logger"
)

// Engine 提供最小可用的 JS 执行环境（goja）
type Engine struct {
	vm      *goja.Runtime
	browser crawler.BrowserRequest
	logSink func(string)
	ctx     *gin.Context // 添加gin.Context支持
}

func New(browser crawler.BrowserRequest) *Engine {
	vm := goja.New()
	e := &Engine{vm: vm, browser: browser}
	e.bindApis()
	return e
}

// NewWithContext 创建带有gin.Context的引擎实例
func NewWithContext(browser crawler.BrowserRequest, ctx *gin.Context) *Engine {
	vm := goja.New()
	e := &Engine{vm: vm, browser: browser, ctx: ctx}
	e.bindApis()
	return e
}

// SetLogSink 设置日志输出回调，用于回流到前端调试面板
func (e *Engine) SetLogSink(sink func(string)) { e.logSink = sink }

func (e *Engine) emit(line string) {
	// 如果有gin.Context，同步输出到ctxlog
	if e.ctx != nil {
		logger.CtxLogger(e.ctx).WithFields(logrus.Fields{
			"engine": "javascript",
			"output": line,
		}).Info("script_output")
	}

	// 如果有logSink，输出到前端调试面板
	if e.logSink != nil {
		e.logSink(line)
		return
	}

	// 默认输出到控制台
	fmt.Println(line)
}

// ---- DOM 封装 ----
func (e *Engine) wrapSelection(sel *goquery.Selection) *goja.Object {
	obj := e.vm.NewObject()
	_ = obj.Set("text", func() string { return strings.TrimSpace(sel.Text()) })
	_ = obj.Set("innerText", func() string { return strings.TrimSpace(sel.Text()) })
	_ = obj.Set("html", func() string { h, _ := sel.Html(); return h })
	_ = obj.Set("innerHTML", func() string { h, _ := sel.Html(); return h })
	_ = obj.Set("attr", func(name string) goja.Value {
		if v, ok := sel.Attr(name); ok {
			return e.vm.ToValue(v)
		}
		return goja.Undefined()
	})
	_ = obj.Set("getAttribute", func(name string) goja.Value {
		if v, ok := sel.Attr(name); ok {
			return e.vm.ToValue(v)
		}
		return goja.Undefined()
	})
	_ = obj.Set("querySelector", func(css string) goja.Value {
		s := sel.Find(css).First()
		if s.Length() == 0 {
			return goja.Undefined()
		}
		return e.wrapSelection(s)
	})
	_ = obj.Set("querySelectorAll", func(css string) *goja.Object {
		arr := e.vm.NewArray()
		var idx int64 = 0
		sel.Find(css).Each(func(i int, s *goquery.Selection) {
			arr.Set(strconv.FormatInt(idx, 10), e.wrapSelection(s))
			idx++
		})
		arr.Set("length", idx)
		return arr
	})
	_ = obj.Set("getElementsByTagName", func(tag string) *goja.Object { return e.wrapSelection(sel.Find(tag)) })
	_ = obj.Set("getElementsByClassName", func(cls string) *goja.Object { return e.wrapSelection(sel.Find("." + cls)) })
	_ = obj.Set("getElementById", func(id string) goja.Value {
		s := sel.Find("#" + id).First()
		if s.Length() == 0 {
			return goja.Undefined()
		}
		return e.wrapSelection(s)
	})

	// 链式DOM操作方法
	_ = obj.Set("parent", func() goja.Value {
		parent := sel.Parent()
		if parent.Length() == 0 {
			return goja.Undefined()
		}
		return e.wrapSelection(parent)
	})
	_ = obj.Set("children", func() *goja.Object {
		children := sel.Children()
		arr := e.vm.NewArray()
		var idx int64 = 0
		children.Each(func(i int, s *goquery.Selection) {
			arr.Set(strconv.FormatInt(idx, 10), e.wrapSelection(s))
			idx++
		})
		arr.Set("length", idx)
		return arr
	})
	_ = obj.Set("next", func() goja.Value {
		next := sel.Next()
		if next.Length() == 0 {
			return goja.Undefined()
		}
		return e.wrapSelection(next)
	})
	_ = obj.Set("prev", func() goja.Value {
		prev := sel.Prev()
		if prev.Length() == 0 {
			return goja.Undefined()
		}
		return e.wrapSelection(prev)
	})
	_ = obj.Set("eq", func(index int) goja.Value {
		eq := sel.Eq(index)
		if eq.Length() == 0 {
			return goja.Undefined()
		}
		return e.wrapSelection(eq)
	})
	_ = obj.Set("first", func() goja.Value {
		first := sel.First()
		if first.Length() == 0 {
			return goja.Undefined()
		}
		return e.wrapSelection(first)
	})
	return obj
}

func (e *Engine) wrapDocument(doc *goquery.Document) *goja.Object {
	// Document 方法代理到根 selection
	root := doc.Selection
	obj := e.vm.NewObject()
	_ = obj.Set("querySelector", func(css string) goja.Value {
		s := root.Find(css).First()
		if s.Length() == 0 {
			return goja.Undefined()
		}
		return e.wrapSelection(s)
	})
	_ = obj.Set("querySelectorAll", func(css string) *goja.Object {
		arr := e.vm.NewArray()
		var idx int64 = 0
		root.Find(css).Each(func(i int, s *goquery.Selection) {
			arr.Set(strconv.FormatInt(idx, 10), e.wrapSelection(s))
			idx++
		})
		arr.Set("length", idx)
		return arr
	})
	_ = obj.Set("getElementsByTagName", func(tag string) *goja.Object { return e.wrapSelection(root.Find(tag)) })
	_ = obj.Set("getElementsByClassName", func(cls string) *goja.Object { return e.wrapSelection(root.Find("." + cls)) })
	_ = obj.Set("getElementById", func(id string) goja.Value {
		s := root.Find("#" + id).First()
		if s.Length() == 0 {
			return goja.Undefined()
		}
		return e.wrapSelection(s)
	})
	_ = obj.Set("text", func() string { return strings.TrimSpace(root.Text()) })
	_ = obj.Set("html", func() string { h, _ := root.Html(); return h })
	return obj
}

// bindApis 绑定可用的全局函数到 JS
func (e *Engine) bindApis() {
	// HTTP（兼容）
	e.vm.Set("httpGet", func(url string) map[string]interface{} {
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
	// httpPost(url, data)
	e.vm.Set("httpPost", func(call goja.FunctionCall) goja.Value {
		var url string
		if len(call.Arguments) > 0 {
			url = call.Arguments[0].String()
		}
		var resp *http.Response
		var err error
		if len(call.Arguments) > 1 {
			if m, ok := call.Arguments[1].Export().(map[string]interface{}); ok {
				resp, err = e.browser.Post(url, m)
			} else {
				b := []byte(call.Arguments[1].String())
				resp, err = e.browser.Do("POST", url, b, map[string]string{"Content-Type": "text/plain"})
			}
		} else {
			resp, err = e.browser.Post(url, map[string]interface{}{})
		}
		if err != nil {
			return e.vm.ToValue(map[string]interface{}{"status_code": 0, "body": "", "url": url, "err": err.Error()})
		}
		defer resp.Body.Close()
		body, _ := readDecompressedBody(resp)
		h := map[string]string{}
		for k, v := range resp.Header {
			if len(v) > 0 {
				h[k] = v[0]
			}
		}
		return e.vm.ToValue(map[string]interface{}{
			"status_code": resp.StatusCode,
			"body":        string(body),
			"url":         resp.Request.URL.String(),
			"headers":     h,
		})
	})

	// 头/UA/Cookie 设置（驼峰命名）
	e.vm.Set("setHeaders", func(m map[string]string) { e.browser.SetHeaders(m) })
	e.vm.Set("setCookies", func(m map[string]string) { e.browser.SetCookies(m) })
	e.vm.Set("setUserAgent", func(ua string) { e.browser.SetUserAgent(ua) })
	e.vm.Set("setRandomUserAgent", func() { e.browser.SetRandomUserAgent() })
	e.vm.Set("getUserAgent", func() string { return e.browser.GetUserAgent() })
	e.vm.Set("setUaToCurrentRequestUa", func() string {
		ua := e.browser.GetUserAgent()
		if ua == "" {
			e.browser.SetRandomUserAgent()
			ua = e.browser.GetUserAgent()
		}
		// 再次设置，确保写入 headers
		e.browser.SetUserAgent(ua)
		return ua
	})

	// DOMParser 与 parseHtml
	e.vm.Set("parseHtml", func(html string) *goja.Object {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			return e.vm.NewObject()
		}
		return e.wrapDocument(doc)
	})
	e.vm.Set("DOMParser", func(call goja.ConstructorCall) *goja.Object {
		obj := e.vm.NewObject()
		_ = obj.Set("parseFromString", func(html string, _type string) *goja.Object {
			d, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				return e.vm.NewObject()
			}
			return e.wrapDocument(d)
		})
		return obj
	})

	// console：常用方法绑定（输出回流到 logSink）
	{
		timers := map[string]time.Time{}
		counters := map[string]int{}
		indentLevel := 0
		indent := func() string {
			if indentLevel <= 0 {
				return ""
			}
			return strings.Repeat("  ", indentLevel)
		}
		stringify := func(v goja.Value) string {
			if goja.IsUndefined(v) || goja.IsNull(v) {
				return ""
			}
			x := v.Export()
			// 如果是字符串，直接返回，不添加引号
			if s, ok := x.(string); ok {
				return s
			}
			// 其他类型使用 JSON 格式化
			b, err := json.Marshal(x)
			if err != nil {
				return fmt.Sprint(x)
			}
			return string(b)
		}
		printArgs := func(level string, args []goja.Value) {
			parts := make([]string, 0, len(args))
			for _, a := range args {
				parts = append(parts, stringify(a))
			}
			e.emit("[" + level + "] " + indent() + strings.Join(parts, " "))
		}
		console := e.vm.NewObject()
		_ = console.Set("log", func(call goja.FunctionCall) goja.Value { printArgs("LOG", call.Arguments); return goja.Undefined() })
		_ = console.Set("info", func(call goja.FunctionCall) goja.Value { printArgs("INFO", call.Arguments); return goja.Undefined() })
		_ = console.Set("debug", func(call goja.FunctionCall) goja.Value { printArgs("DEBUG", call.Arguments); return goja.Undefined() })
		_ = console.Set("warn", func(call goja.FunctionCall) goja.Value { printArgs("WARN", call.Arguments); return goja.Undefined() })
		_ = console.Set("error", func(call goja.FunctionCall) goja.Value { printArgs("ERROR", call.Arguments); return goja.Undefined() })
		_ = console.Set("trace", func(call goja.FunctionCall) goja.Value {
			printArgs("TRACE", call.Arguments)
			e.emit(string(debug.Stack()))
			return goja.Undefined()
		})
		_ = console.Set("time", func(call goja.FunctionCall) goja.Value {
			label := "default"
			if len(call.Arguments) > 0 {
				label = call.Arguments[0].String()
			}
			timers[label] = time.Now()
			e.emit("[TIME] start " + label)
			return goja.Undefined()
		})
		_ = console.Set("timeEnd", func(call goja.FunctionCall) goja.Value {
			label := "default"
			if len(call.Arguments) > 0 {
				label = call.Arguments[0].String()
			}
			if t, ok := timers[label]; ok {
				d := time.Since(t)
				e.emit("[TIME] " + label + " " + d.String())
				delete(timers, label)
			} else {
				e.emit("[TIME] " + label + " no such label")
			}
			return goja.Undefined()
		})
		_ = console.Set("assert", func(call goja.FunctionCall) goja.Value {
			cond := true
			if len(call.Arguments) > 0 {
				cond = !goja.IsUndefined(call.Arguments[0]) && !goja.IsNull(call.Arguments[0]) && call.Arguments[0].Export() != nil
			}
			if !cond {
				printArgs("ASSERT", call.Arguments[1:])
			}
			return goja.Undefined()
		})
		_ = console.Set("group", func(call goja.FunctionCall) goja.Value {
			printArgs("GROUP", call.Arguments)
			indentLevel++
			return goja.Undefined()
		})
		_ = console.Set("groupCollapsed", func(call goja.FunctionCall) goja.Value {
			printArgs("GROUP-COLLAPSED", call.Arguments)
			indentLevel++
			return goja.Undefined()
		})
		_ = console.Set("groupEnd", func(call goja.FunctionCall) goja.Value {
			if indentLevel > 0 {
				indentLevel--
			}
			e.emit("[GROUP] end")
			return goja.Undefined()
		})
		_ = console.Set("count", func(call goja.FunctionCall) goja.Value {
			label := "default"
			if len(call.Arguments) > 0 {
				label = call.Arguments[0].String()
			}
			counters[label]++
			e.emit(fmt.Sprintf("[COUNT] %s %d", label, counters[label]))
			return goja.Undefined()
		})
		_ = console.Set("countReset", func(call goja.FunctionCall) goja.Value {
			label := "default"
			if len(call.Arguments) > 0 {
				label = call.Arguments[0].String()
			}
			counters[label] = 0
			e.emit(fmt.Sprintf("[COUNT] %s 0", label))
			return goja.Undefined()
		})
		_ = console.Set("table", func(call goja.FunctionCall) goja.Value { printArgs("TABLE", call.Arguments); return goja.Undefined() })
		_ = console.Set("dir", func(call goja.FunctionCall) goja.Value { printArgs("DIR", call.Arguments); return goja.Undefined() })
		_ = console.Set("dirxml", func(call goja.FunctionCall) goja.Value { printArgs("DIRXML", call.Arguments); return goja.Undefined() })
		_ = console.Set("clear", func(call goja.FunctionCall) goja.Value { e.emit("[CLEAR]\n\n"); return goja.Undefined() })
		e.vm.Set("console", console)
	}

	// URL 库
	urlLib := e.vm.NewObject()
	_ = urlLib.Set("encode", func(str string) string {
		return url.QueryEscape(str)
	})
	_ = urlLib.Set("decode", func(str string) string {
		decoded, err := url.QueryUnescape(str)
		if err != nil {
			return str // 解码失败时返回原字符串
		}
		return decoded
	})
	_ = urlLib.Set("parse", func(urlStr string) map[string]interface{} {
		parsed, err := url.Parse(urlStr)
		if err != nil {
			return map[string]interface{}{"error": err.Error()}
		}
		return map[string]interface{}{
			"scheme":   parsed.Scheme,
			"host":     parsed.Host,
			"path":     parsed.Path,
			"query":    parsed.RawQuery,
			"fragment": parsed.Fragment,
			"raw":      parsed.String(),
		}
	})
	_ = urlLib.Set("build", func(components map[string]interface{}) string {
		parsed := &url.URL{}
		if scheme, ok := components["scheme"].(string); ok {
			parsed.Scheme = scheme
		}
		if host, ok := components["host"].(string); ok {
			parsed.Host = host
		}
		if path, ok := components["path"].(string); ok {
			parsed.Path = path
		}
		if query, ok := components["query"].(string); ok {
			parsed.RawQuery = query
		}
		if fragment, ok := components["fragment"].(string); ok {
			parsed.Fragment = fragment
		}
		return parsed.String()
	})
	e.vm.Set("url", urlLib)

	// Unicode 库
	unicodeLib := e.vm.NewObject()
	_ = unicodeLib.Set("encode", func(str string) string {
		var result strings.Builder
		for _, r := range str {
			if r < 128 {
				result.WriteRune(r)
			} else {
				result.WriteString(fmt.Sprintf("\\u%04X", r))
			}
		}
		return result.String()
	})
	_ = unicodeLib.Set("decode", func(str string) string {
		re := regexp.MustCompile(`\\u([0-9a-fA-F]{4})`)
		result := re.ReplaceAllStringFunc(str, func(match string) string {
			hex := match[2:6] // 提取 XXXX 部分
			if code, err := strconv.ParseUint(hex, 16, 32); err == nil {
				return string(rune(code))
			}
			return match // 如果解析失败，保持原样
		})
		return result
	})
	_ = unicodeLib.Set("isAscii", func(str string) bool {
		for _, r := range str {
			if r >= 128 {
				return false
			}
		}
		return true
	})
	_ = unicodeLib.Set("length", func(str string) int64 {
		return int64(len([]rune(str)))
	})
	e.vm.Set("unicode", unicodeLib)

	// fetch：同步实现（与 await 兼容：await 非 thenable 值将立即返回）
	e.vm.Set("fetch", func(call goja.FunctionCall) goja.Value {
		var url string
		if len(call.Arguments) > 0 {
			url = call.Arguments[0].String()
		}
		method := "GET"
		headers := map[string]string{}
		var bodyBytes []byte
		var contentType string
		var timeoutMs int64 = -1
		redirect := "follow" // follow | manual | error
		if len(call.Arguments) > 1 {
			opt := call.Arguments[1].Export()
			if m, ok := opt.(map[string]interface{}); ok {
				if v, ok := m["method"].(string); ok && v != "" {
					method = strings.ToUpper(v)
				}
				if hv, ok := m["headers"]; ok {
					switch h := hv.(type) {
					case map[string]interface{}:
						for k, vv := range h {
							headers[k] = fmt.Sprint(vv)
						}
					case map[string]string:
						for k, vv := range h {
							headers[k] = vv
						}
					}
				}
				if b, ok := m["body"]; ok {
					// 推断 body 类型：字符串 => 原样；对象/数组 => JSON；其它 => toString
					switch bb := b.(type) {
					case string:
						bodyBytes = []byte(bb)
						if contentType == "" {
							contentType = headers["Content-Type"]
						}
					case map[string]interface{}, []interface{}:
						js, _ := json.Marshal(b)
						bodyBytes = js
						if contentType == "" {
							contentType = "application/json"
						}
					default:
						bodyBytes = []byte(fmt.Sprint(b))
					}
				}
				// Node/Web 常见选项：timeout(毫秒) & redirect
				if t, ok := m["timeout"].(int64); ok {
					timeoutMs = t
				}
				if t2, ok := m["timeout"].(float64); ok {
					timeoutMs = int64(t2)
				}
				if rv, ok := m["redirect"].(string); ok {
					redirect = strings.ToLower(rv)
				}
			}
		}

		// 处理超时与重定向策略
		var restoreTimeout time.Duration
		if timeoutMs > 0 {
			restoreTimeout = e.browser.GetTimeout()
			e.browser.SetTimeout(time.Duration(timeoutMs) * time.Millisecond)
		}
		var restoreFollow *bool
		if redirect == "manual" || redirect == "error" {
			prev := e.browser.GetFollowRedirects()
			restoreFollow = &prev
			e.browser.SetFollowRedirects(false)
		}

		// 合成请求头
		h := map[string]string{}
		for k, v := range headers {
			h[k] = v
		}
		if contentType != "" && h["Content-Type"] == "" {
			h["Content-Type"] = contentType
		}

		// 发起请求（统一使用 Do）
		resp, err := e.browser.Do(method, url, bodyBytes, h)

		// 恢复配置
		if timeoutMs > 0 {
			e.browser.SetTimeout(restoreTimeout)
		}
		if restoreFollow != nil {
			e.browser.SetFollowRedirects(*restoreFollow)
		}

		if err != nil {
			return e.vm.ToValue(map[string]interface{}{"error": err.Error()})
		}
		defer resp.Body.Close()

		// 构造 Headers 对象
		hdrObj := e.vm.NewObject()
		hdrMap := map[string]string{}
		for k, v := range resp.Header {
			if len(v) > 0 {
				hdrMap[strings.ToLower(k)] = v[0]
			}
		}
		_ = hdrObj.Set("get", func(name string) goja.Value {
			if v, ok := hdrMap[strings.ToLower(name)]; ok {
				return e.vm.ToValue(v)
			}
			return goja.Undefined()
		})
		_ = hdrObj.Set("has", func(name string) bool { _, ok := hdrMap[strings.ToLower(name)]; return ok })
		_ = hdrObj.Set("keys", func() []string {
			ks := make([]string, 0, len(hdrMap))
			for k := range hdrMap {
				ks = append(ks, k)
			}
			return ks
		})
		_ = hdrObj.Set("values", func() []string {
			vs := make([]string, 0, len(hdrMap))
			for _, v := range hdrMap {
				vs = append(vs, v)
			}
			return vs
		})
		_ = hdrObj.Set("entries", func() []map[string]string {
			arr := []map[string]string{}
			for k, v := range hdrMap {
				arr = append(arr, map[string]string{"0": k, "1": v})
			}
			return arr
		})
		_ = hdrObj.Set("forEach", func(cb goja.Callable) {
			for k, v := range hdrMap {
				_, _ = cb(nil, e.vm.ToValue(v), e.vm.ToValue(k))
			}
		})

		// 读取 body（自动解压）
		b, _ := readDecompressedBody(resp)

		// 构造 Response 对象（同步）
		respObj := e.vm.NewObject()
		_ = respObj.Set("ok", resp.StatusCode >= 200 && resp.StatusCode < 300)
		_ = respObj.Set("status", resp.StatusCode)
		_ = respObj.Set("statusText", resp.Status)
		_ = respObj.Set("url", resp.Request.URL.String())
		_ = respObj.Set("headers", hdrObj)
		_ = respObj.Set("redirected", resp.Request.URL.String() != url)
		_ = respObj.Set("type", "basic")
		_ = respObj.Set("text", func() string { return string(b) })
		_ = respObj.Set("json", func() goja.Value {
			var v interface{}
			if err := json.Unmarshal(b, &v); err != nil {
				return goja.Undefined()
			}
			return e.vm.ToValue(v)
		})
		_ = respObj.Set("arrayBuffer", func() []byte { return b })

		// 重定向策略（manual | error）
		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			loc := resp.Header.Get("Location")
			if redirect == "error" {
				return e.vm.ToValue(map[string]interface{}{"error": fmt.Sprintf("redirect not allowed: %d %s", resp.StatusCode, loc)})
			}
			if redirect == "manual" {
				_ = respObj.Set("location", loc)
			}
		}

		return respObj
	})
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
