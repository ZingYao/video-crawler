package lua

import (
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	lua "github.com/yuin/gopher-lua"

	"video-crawler/internal/crawler"
)

const (
	mtGoqueryDocument  = "goquery_document"
	mtGoquerySelection = "goquery_selection"
)

// Lua userdata wrappers for chainable HTML operations
type luaDocument struct{ doc *goquery.Document }
type luaSelection struct{ sel *goquery.Selection }

// LuaEngine Lua引擎
type LuaEngine struct {
	L       *lua.LState
	browser crawler.BrowserRequest
	output  chan string // 用于流式输出的通道
}

// NewLuaEngine 创建新的Lua引擎
func NewLuaEngine(browser crawler.BrowserRequest) *LuaEngine {
	L := lua.NewState()
	engine := &LuaEngine{
		L:       L,
		browser: browser,
		output:  make(chan string, 100), // 缓冲通道，避免阻塞
	}

	// 注册所有函数到Lua
	engine.registerFunctions()

	return engine
}

// 解压响应体，支持 gzip/deflate（若未压缩则直接读取）
func readDecompressedBody(resp *http.Response) ([]byte, error) {
	enc := strings.ToLower(strings.TrimSpace(resp.Header.Get("Content-Encoding")))
	if enc == "" || enc == "identity" {
		return io.ReadAll(resp.Body)
	}
	// 只取第一个编码
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
		// 少数服务端可能返回 raw deflate，尝试 flate
		fr := flate.NewReader(resp.Body)
		defer fr.Close()
		return io.ReadAll(fr)
	default:
		// 其他编码（如 br/zstd）暂不引入额外依赖，回退为直接读取
		return io.ReadAll(resp.Body)
	}
}

// GetOutputChannel 获取输出通道
func (e *LuaEngine) GetOutputChannel() <-chan string {
	return e.output
}

// CloseOutput 关闭输出通道
func (e *LuaEngine) CloseOutput() {
	close(e.output)
}

// registerFunctions 注册所有函数到Lua
func (e *LuaEngine) registerFunctions() {
	// 注册HTTP请求函数
	e.L.SetGlobal("http_get", e.L.NewFunction(e.luaHttpGet))
	e.L.SetGlobal("http_post", e.L.NewFunction(e.luaHttpPost))
	e.L.SetGlobal("set_headers", e.L.NewFunction(e.luaSetHeaders))
	e.L.SetGlobal("set_cookies", e.L.NewFunction(e.luaSetCookies))
	e.L.SetGlobal("set_user_agent", e.L.NewFunction(e.luaSetUserAgent))
	e.L.SetGlobal("set_random_user_agent", e.L.NewFunction(e.luaSetRandomUserAgent))
	// 将 HTTP 客户端当前 UA 应用到请求头，返回生效的 UA
	e.L.SetGlobal("set_ua_2_current_request_ua", e.L.NewFunction(e.luaSetUA2CurrentRequestUA))

	// 注册HTML解析函数（链式入口）
	e.L.SetGlobal("parse_html", e.L.NewFunction(e.luaParseHtml))

	// 注册工具函数
	e.L.SetGlobal("print", e.L.NewFunction(e.luaPrint))
	e.L.SetGlobal("log", e.L.NewFunction(e.luaLog))
	e.L.SetGlobal("sleep", e.L.NewFunction(e.luaSleep))
	e.L.SetGlobal("get_user_agent", e.L.NewFunction(e.luaGetUserAgent))
	// 字符串工具
	e.L.SetGlobal("split", e.L.NewFunction(e.luaSplit))
	e.L.SetGlobal("trim", e.L.NewFunction(e.luaTrim))
	// JSON 编解码
	e.L.SetGlobal("json_encode", e.L.NewFunction(e.luaJsonEncode))
	e.L.SetGlobal("json_decode", e.L.NewFunction(e.luaJsonDecode))

	// 禁用危险的系统函数，并提供禁用信息
	e.L.SetGlobal("io", e.L.NewFunction(e.luaDisabledTable("io")))
	e.L.SetGlobal("package", e.L.NewFunction(e.luaDisabledTable("package")))
	e.L.SetGlobal("require", e.L.NewFunction(e.luaDisabledFunction("require")))
	e.L.SetGlobal("dofile", e.L.NewFunction(e.luaDisabledFunction("dofile")))
	e.L.SetGlobal("loadfile", e.L.NewFunction(e.luaDisabledFunction("loadfile")))

	// 保留安全的 os 函数，禁用危险的 os 函数
	e.L.SetGlobal("os", e.L.NewFunction(e.luaSafeOs))

	// 链式类型注册
	e.registerGoqueryTypes()
}

// registerGoqueryTypes 注册链式 HTML 操作需要的类型元表
func (e *LuaEngine) registerGoqueryTypes() {
	// Document methods
	mtDoc := e.L.NewTypeMetatable(mtGoqueryDocument)
	e.L.SetField(mtDoc, "__index", e.L.SetFuncs(e.L.NewTable(), map[string]lua.LGFunction{
		"select":     e.luaSelect,
		"select_one": e.luaSelectOne,
		"html":       e.luaHtml,
		"text":       e.luaText,
	}))

	// Selection methods
	mtSel := e.L.NewTypeMetatable(mtGoquerySelection)
	e.L.SetField(mtSel, "__index", e.L.SetFuncs(e.L.NewTable(), map[string]lua.LGFunction{
		"select":     e.luaSelect,
		"select_one": e.luaSelectOne,
		"first":      e.luaFirst,
		"parent":     e.luaParent,
		"children":   e.luaChildren,
		"next":       e.luaNext,
		"prev":       e.luaPrev,
		"eq":         e.luaEq,
		"attr":       e.luaAttr,
		"html":       e.luaHtml,
		"text":       e.luaText,
	}))
}

// 选择器辅助：从参数1中取 selection 或 document
func (e *LuaEngine) getSel(L *lua.LState) (*goquery.Selection, bool) {
	if ud, ok := L.Get(1).(*lua.LUserData); ok {
		if v, ok := ud.Value.(*luaSelection); ok && v != nil {
			return v.sel, true
		}
		if d, ok := ud.Value.(*luaDocument); ok && d != nil {
			return d.doc.Selection, true
		}
	}
	return nil, false
}

func (e *LuaEngine) luaParent(L *lua.LState) int {
	sel, ok := e.getSel(L)
	if !ok {
		L.Push(lua.LNil)
		return 1
	}
	udSel := L.NewUserData()
	udSel.Value = &luaSelection{sel: sel.Parent()}
	L.SetMetatable(udSel, L.GetTypeMetatable(mtGoquerySelection))
	L.Push(udSel)
	return 1
}

func (e *LuaEngine) luaChildren(L *lua.LState) int {
	sel, ok := e.getSel(L)
	if !ok {
		L.Push(lua.LNil)
		return 1
	}
	udSel := L.NewUserData()
	udSel.Value = &luaSelection{sel: sel.Children()}
	L.SetMetatable(udSel, L.GetTypeMetatable(mtGoquerySelection))
	L.Push(udSel)
	return 1
}

func (e *LuaEngine) luaNext(L *lua.LState) int {
	sel, ok := e.getSel(L)
	if !ok {
		L.Push(lua.LNil)
		return 1
	}
	udSel := L.NewUserData()
	udSel.Value = &luaSelection{sel: sel.Next()}
	L.SetMetatable(udSel, L.GetTypeMetatable(mtGoquerySelection))
	L.Push(udSel)
	return 1
}

func (e *LuaEngine) luaPrev(L *lua.LState) int {
	sel, ok := e.getSel(L)
	if !ok {
		L.Push(lua.LNil)
		return 1
	}
	udSel := L.NewUserData()
	udSel.Value = &luaSelection{sel: sel.Prev()}
	L.SetMetatable(udSel, L.GetTypeMetatable(mtGoquerySelection))
	L.Push(udSel)
	return 1
}

func (e *LuaEngine) luaEq(L *lua.LState) int {
	sel, ok := e.getSel(L)
	if !ok {
		L.Push(lua.LNil)
		return 1
	}
	idx := L.CheckInt(2)
	udSel := L.NewUserData()
	udSel.Value = &luaSelection{sel: sel.Eq(idx)}
	L.SetMetatable(udSel, L.GetTypeMetatable(mtGoquerySelection))
	L.Push(udSel)
	return 1
}

// luaFirst selection:first()
func (e *LuaEngine) luaFirst(L *lua.LState) int {
	if ud, ok := L.Get(1).(*lua.LUserData); ok {
		if v, ok := ud.Value.(*luaSelection); ok && v != nil {
			newSel := v.sel.First()
			udSel := L.NewUserData()
			udSel.Value = &luaSelection{sel: newSel}
			L.SetMetatable(udSel, L.GetTypeMetatable(mtGoquerySelection))
			L.Push(udSel)
			return 1
		}
	}
	L.Push(lua.LNil)
	return 1
}

// luaHttpGet Lua中的http_get函数
func (e *LuaEngine) luaHttpGet(L *lua.LState) int {
	url := L.CheckString(1)

	response, err := e.browser.Get(url)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	defer response.Body.Close()

	// 读取响应体（自动解压）
	body, err := readDecompressedBody(response)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("failed to read response body: %v", err)))
		return 2
	}

	// 处理响应体，去除转义
	bodyStr := string(body)
	// 返回响应表
	responseTable := L.CreateTable(0, 4)
	responseTable.RawSetString("status_code", lua.LNumber(response.StatusCode))
	responseTable.RawSetString("body", lua.LString(bodyStr))
	responseTable.RawSetString("url", lua.LString(response.Request.URL.String()))

	// 设置响应头
	headersTable := L.CreateTable(0, len(response.Header))
	for key, values := range response.Header {
		if len(values) > 0 {
			headersTable.RawSetString(key, lua.LString(values[0]))
		}
	}
	responseTable.RawSetString("headers", headersTable)

	L.Push(responseTable)
	L.Push(lua.LNil) // 错误为nil
	return 2
}

// luaHttpPost Lua中的http_post函数
func (e *LuaEngine) luaHttpPost(L *lua.LState) int {
	url := L.CheckString(1)
	dataTable := L.CheckTable(2)

	// 将Lua表转换为Go map
	data := make(map[string]interface{})
	dataTable.ForEach(func(key, value lua.LValue) {
		keyStr := key.String()
		switch value.Type() {
		case lua.LTString:
			data[keyStr] = value.String()
		case lua.LTNumber:
			data[keyStr] = float64(value.(lua.LNumber))
		case lua.LTBool:
			data[keyStr] = bool(value.(lua.LBool))
		default:
			data[keyStr] = value.String()
		}
	})

	response, err := e.browser.Post(url, data)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	defer response.Body.Close()

	// 读取响应体（自动解压）
	body, err := readDecompressedBody(response)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("failed to read response body: %v", err)))
		return 2
	}

	bodyStr := string(body)

	// 返回响应表
	responseTable := L.CreateTable(0, 4)
	responseTable.RawSetString("status_code", lua.LNumber(response.StatusCode))
	responseTable.RawSetString("body", lua.LString(bodyStr))
	responseTable.RawSetString("url", lua.LString(response.Request.URL.String()))

	// 设置响应头
	headersTable := L.CreateTable(0, len(response.Header))
	for key, values := range response.Header {
		if len(values) > 0 {
			headersTable.RawSetString(key, lua.LString(values[0]))
		}
	}
	responseTable.RawSetString("headers", headersTable)

	L.Push(responseTable)
	L.Push(lua.LNil) // 错误为nil
	return 2
}

// luaSetHeaders Lua中的set_headers函数
func (e *LuaEngine) luaSetHeaders(L *lua.LState) int {
	headersTable := L.CheckTable(1)

	headers := make(map[string]string)
	headersTable.ForEach(func(key, value lua.LValue) {
		headers[key.String()] = value.String()
	})

	e.browser.SetHeaders(headers)
	return 0
}

// luaSetCookies Lua中的set_cookies函数
func (e *LuaEngine) luaSetCookies(L *lua.LState) int {
	cookiesTable := L.CheckTable(1)

	cookies := make(map[string]string)
	cookiesTable.ForEach(func(key, value lua.LValue) {
		cookies[key.String()] = value.String()
	})

	e.browser.SetCookies(cookies)
	return 0
}

// luaSetUserAgent Lua中的set_user_agent函数
func (e *LuaEngine) luaSetUserAgent(L *lua.LState) int {
	userAgent := L.CheckString(1)
	e.browser.SetUserAgent(userAgent)
	return 0
}

// luaSetRandomUserAgent Lua中的set_random_user_agent函数
func (e *LuaEngine) luaSetRandomUserAgent(L *lua.LState) int {
	e.browser.SetRandomUserAgent()
	return 0
}

// luaSetUA2CurrentRequestUA 将 HTTP 客户端当前 UA 应用到请求头，并返回最终 UA
func (e *LuaEngine) luaSetUA2CurrentRequestUA(L *lua.LState) int {
	ua := e.browser.GetUserAgent()
	if ua == "" {
		e.browser.SetRandomUserAgent()
		ua = e.browser.GetUserAgent()
	}
	// 再次设置，确保写入到 headers
	e.browser.SetUserAgent(ua)
	L.Push(lua.LString(ua))
	return 1
}

// luaGetUserAgent Lua中的get_user_agent函数
func (e *LuaEngine) luaGetUserAgent(L *lua.LState) int {
	userAgent := e.browser.GetUserAgent()
	L.Push(lua.LString(userAgent))
	return 1
}

// luaParseHtml Lua中的parse_html函数
func (e *LuaEngine) luaParseHtml(L *lua.LState) int {
	html := L.CheckString(1)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	ud := L.NewUserData()
	ud.Value = &luaDocument{doc: doc}
	L.SetMetatable(ud, L.GetTypeMetatable(mtGoqueryDocument))
	L.Push(ud)
	L.Push(lua.LNil)
	return 2
}

// luaSelect Lua中的select函数
func (e *LuaEngine) luaSelect(L *lua.LState) int {
	selector := L.CheckString(2)
	if ud, ok := L.Get(1).(*lua.LUserData); ok {
		switch v := ud.Value.(type) {
		case *luaDocument:
			selection := v.doc.Find(selector)
			udSel := L.NewUserData()
			udSel.Value = &luaSelection{sel: selection}
			L.SetMetatable(udSel, L.GetTypeMetatable(mtGoquerySelection))
			L.Push(udSel)
			L.Push(lua.LNil)
			return 2
		case *luaSelection:
			selection := v.sel.Find(selector)
			udSel := L.NewUserData()
			udSel.Value = &luaSelection{sel: selection}
			L.SetMetatable(udSel, L.GetTypeMetatable(mtGoquerySelection))
			L.Push(udSel)
			L.Push(lua.LNil)
			return 2
		}
	}
	// 兼容旧表结构
	if docTable, ok := L.Get(1).(*lua.LTable); ok {
		h := docTable.RawGetString("_doc").String()
		d, err := goquery.NewDocumentFromReader(strings.NewReader(h))
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		selection := d.Find(selector)
		udSel := L.NewUserData()
		udSel.Value = &luaSelection{sel: selection}
		L.SetMetatable(udSel, L.GetTypeMetatable(mtGoquerySelection))
		L.Push(udSel)
		L.Push(lua.LNil)
		return 2
	}
	L.Push(lua.LNil)
	L.Push(lua.LString("invalid context for select"))
	return 2
}

// luaSelectOne Lua中的select_one函数
func (e *LuaEngine) luaSelectOne(L *lua.LState) int {
	selector := L.CheckString(2)
	// 1) 新版链式：userdata(doc/selection)
	if ud, ok := L.Get(1).(*lua.LUserData); ok {
		switch v := ud.Value.(type) {
		case *luaDocument:
			selection := v.doc.Find(selector).First()
			if selection.Length() == 0 {
				L.Push(lua.LNil)
				L.Push(lua.LString("no element found"))
				return 2
			}
			udSel := L.NewUserData()
			udSel.Value = &luaSelection{sel: selection}
			L.SetMetatable(udSel, L.GetTypeMetatable(mtGoquerySelection))
			L.Push(udSel)
			L.Push(lua.LNil)
			return 2
		case *luaSelection:
			selection := v.sel.Find(selector).First()
			if selection.Length() == 0 {
				L.Push(lua.LNil)
				L.Push(lua.LString("no element found"))
				return 2
			}
			udSel := L.NewUserData()
			udSel.Value = &luaSelection{sel: selection}
			L.SetMetatable(udSel, L.GetTypeMetatable(mtGoquerySelection))
			L.Push(udSel)
			L.Push(lua.LNil)
			return 2
		}
	}
	// 2) 旧版：table {_doc}
	docTable := L.CheckTable(1)
	html := docTable.RawGetString("_doc").String()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	selection := doc.Find(selector).First()
	if selection.Length() == 0 {
		L.Push(lua.LNil)
		L.Push(lua.LString("no element found"))
		return 2
	}
	elementTable := L.CreateTable(0, 3)
	htmlContent, _ := selection.Html()
	// 封装旧结构
	elementTable.RawSetString("_html", lua.LString(htmlContent))
	elementTable.RawSetString("_text", lua.LString(selection.Text()))
	elementTable.RawSetString("_selection", lua.LString("goquery_selection"))
	L.Push(elementTable)
	L.Push(lua.LNil)
	return 2
}

// luaAttr Lua中的attr函数
func (e *LuaEngine) luaAttr(L *lua.LState) int {
	name := L.CheckString(2)
	// 新版：支持 Selection userdata
	if ud, ok := L.Get(1).(*lua.LUserData); ok {
		if v, ok := ud.Value.(*luaSelection); ok && v != nil {
			val, exists := v.sel.Attr(name)
			if !exists {
				L.Push(lua.LNil)
				L.Push(lua.LString("attribute not found"))
				return 2
			}
			L.Push(lua.LString(val))
			L.Push(lua.LNil)
			return 2
		}
	}
	// 旧版：table {_html}
	elementTable := L.CheckTable(1)
	h := elementTable.RawGetString("_html").String()
	d, err := goquery.NewDocumentFromReader(strings.NewReader(h))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	selection := d.Find("*").First()
	if selection.Length() == 0 {
		L.Push(lua.LNil)
		L.Push(lua.LString("no element found"))
		return 2
	}
	val, exists := selection.Attr(name)
	if !exists {
		L.Push(lua.LNil)
		L.Push(lua.LString("attribute not found"))
		return 2
	}
	L.Push(lua.LString(val))
	L.Push(lua.LNil)
	return 2
}

// luaText Lua中的text函数
func (e *LuaEngine) luaText(L *lua.LState) int {
	// 新版：Selection userdata
	if ud, ok := L.Get(1).(*lua.LUserData); ok {
		if v, ok := ud.Value.(*luaSelection); ok && v != nil {
			L.Push(lua.LString(v.sel.Text()))
			return 1
		}
	}
	// 旧版：table {_text}
	elementTable := L.CheckTable(1)
	text := elementTable.RawGetString("_text").String()
	L.Push(lua.LString(text))
	return 1
}

// luaHtml Lua中的html函数
func (e *LuaEngine) luaHtml(L *lua.LState) int {
	// 新版：支持 Selection 或 Document userdata
	if ud, ok := L.Get(1).(*lua.LUserData); ok {
		switch v := ud.Value.(type) {
		case *luaSelection:
			h, err := v.sel.Html()
			if err != nil {
				h = ""
			}
			L.Push(lua.LString(h))
			return 1
		case *luaDocument:
			h, err := v.doc.Html()
			if err != nil {
				h = ""
			}
			L.Push(lua.LString(h))
			return 1
		}
	}
	// 旧版：table {_html}
	elementTable := L.CheckTable(1)
	h := elementTable.RawGetString("_html").String()
	if h == "" {
		L.Push(lua.LString(""))
		return 1
	}
	d, err := goquery.NewDocumentFromReader(strings.NewReader(h))
	if err != nil {
		L.Push(lua.LString(h))
		return 1
	}
	selection := d.Find("*").First()
	if selection.Length() == 0 {
		L.Push(lua.LString(h))
		return 1
	}
	formattedHtml, err := selection.Html()
	if err != nil {
		L.Push(lua.LString(h))
		return 1
	}
	L.Push(lua.LString(formattedHtml))
	return 1
}

// luaPrint Lua中的print函数
func (e *LuaEngine) luaPrint(L *lua.LState) int {
	args := make([]string, L.GetTop())
	for i := 1; i <= L.GetTop(); i++ {
		args[i-1] = L.Get(i).String()
	}
	output := strings.Join(args, " ")

	// 发送到输出通道
	select {
	case e.output <- fmt.Sprintf("[PRINT][%s] %s", time.Now().Format("2006-01-02 15:04:05.000"), output):
	default:
		// 如果通道满了，丢弃消息
	}

	return 0
}

// luaLog Lua中的log函数
func (e *LuaEngine) luaLog(L *lua.LState) int {
	args := make([]string, L.GetTop())
	for i := 1; i <= L.GetTop(); i++ {
		args[i-1] = L.Get(i).String()
	}
	output := strings.Join(args, " ")

	// 发送到输出通道
	select {
	case e.output <- fmt.Sprintf("[LOG][%s] %s", time.Now().Format("2006-01-02 15:04:05.000"), output):
	default:
		// 如果通道满了，丢弃消息
	}

	return 0
}

// luaSleep Lua中的sleep函数，单位毫秒
func (e *LuaEngine) luaSleep(L *lua.LState) int {
	ms := L.CheckInt(1)
	if ms < 0 {
		ms = 0
	}
	time.Sleep(time.Duration(ms) * time.Millisecond)
	return 0
}

// luaSplit 将字符串按分隔符拆分为数组（1 起始）
func (e *LuaEngine) luaSplit(L *lua.LState) int {
	s := L.CheckString(1)
	sep := L.CheckString(2)
	// 允许空分隔符：按字符切分（rune 安全）
	var parts []string
	if sep == "" {
		parts = make([]string, 0, len(s))
		for _, r := range s {
			parts = append(parts, string(r))
		}
	} else {
		parts = strings.Split(s, sep)
	}

	tbl := L.NewTable()
	for i, p := range parts {
		tbl.RawSetInt(i+1, lua.LString(p))
	}
	L.Push(tbl)
	return 1
}

// luaTrim 去除字符串首尾空白
func (e *LuaEngine) luaTrim(L *lua.LState) int {
	s := L.CheckString(1)
	L.Push(lua.LString(strings.TrimSpace(s)))
	return 1
}

// luaJsonEncode 将 Lua 值编码为 JSON 字符串
func (e *LuaEngine) luaJsonEncode(L *lua.LState) int {
	v := L.CheckAny(1)
	goVal := luaToInterfaceJSON(v)

	var (
		data []byte
		err  error
	)

	// 可选第二参数：
	// - boolean: true 使用两个空格缩进；false 使用紧凑模式
	// - number: 使用给定个数空格缩进
	// - string: 使用该字符串作为缩进（例如 "\t"）
	if L.GetTop() >= 2 {
		opt := L.Get(2)
		switch opt.Type() {
		case lua.LTBool:
			if bool(opt.(lua.LBool)) {
				data, err = json.MarshalIndent(goVal, "", "  ")
			} else {
				data, err = json.Marshal(goVal)
			}
		case lua.LTNumber:
			n := int(opt.(lua.LNumber))
			if n < 0 {
				n = 0
			}
			indent := strings.Repeat(" ", n)
			data, err = json.MarshalIndent(goVal, "", indent)
		case lua.LTString:
			indent := opt.String()
			data, err = json.MarshalIndent(goVal, "", indent)
		default:
			data, err = json.Marshal(goVal)
		}
	} else {
		data, err = json.Marshal(goVal)
	}

	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LString(string(data)))
	L.Push(lua.LNil)
	return 2
}

// luaJsonDecode 将 JSON 字符串解码为 Lua 值
func (e *LuaEngine) luaJsonDecode(L *lua.LState) int {
	s := L.CheckString(1)
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(interfaceToLua(L, v))
	L.Push(lua.LNil)
	return 2
}

// luaToInterfaceJSON 更宽松地将 Lua 值转为 Go 值用于 JSON 序列化（数组按 1..n 判断）
func luaToInterfaceJSON(v lua.LValue) interface{} {
	switch v.Type() {
	case lua.LTNil:
		return nil
	case lua.LTBool:
		return bool(v.(lua.LBool))
	case lua.LTNumber:
		return float64(v.(lua.LNumber))
	case lua.LTString:
		return v.String()
	case lua.LTTable:
		// 判断是否为顺序数组（1..n）
		t := v.(*lua.LTable)
		length := t.Len()
		if length > 0 {
			arr := make([]interface{}, 0, length)
			for i := 1; i <= length; i++ {
				arr = append(arr, luaToInterfaceJSON(t.RawGetInt(i)))
			}
			return arr
		}
		// 非顺序数组当作对象处理
		obj := make(map[string]interface{})
		t.ForEach(func(k, vv lua.LValue) {
			obj[k.String()] = luaToInterfaceJSON(vv)
		})
		return obj
	default:
		return v.String()
	}
}

// interfaceToLua 将解码后的 Go 值转为 Lua 值
func interfaceToLua(L *lua.LState, v interface{}) lua.LValue {
	switch val := v.(type) {
	case nil:
		return lua.LNil
	case bool:
		return lua.LBool(val)
	case float64:
		return lua.LNumber(val)
	case float32:
		return lua.LNumber(val)
	case int, int8, int16, int32, int64:
		return lua.LNumber(reflectToFloat64(val))
	case uint, uint8, uint16, uint32, uint64:
		return lua.LNumber(reflectToFloat64(val))
	case string:
		return lua.LString(val)
	case []interface{}:
		tbl := L.NewTable()
		for i, item := range val {
			tbl.RawSetInt(i+1, interfaceToLua(L, item))
		}
		return tbl
	case map[string]interface{}:
		tbl := L.NewTable()
		for k, item := range val {
			tbl.RawSetString(k, interfaceToLua(L, item))
		}
		return tbl
	default:
		// 兜底：尝试 json 再转
		b, err := json.Marshal(val)
		if err == nil {
			var any interface{}
			if json.Unmarshal(b, &any) == nil {
				return interfaceToLua(L, any)
			}
		}
		return lua.LString(fmt.Sprintf("%v", val))
	}
}

// reflectToFloat64 辅助把整型/无符号整型转为 float64
func reflectToFloat64(v interface{}) float64 {
	switch n := v.(type) {
	case int:
		return float64(n)
	case int8:
		return float64(n)
	case int16:
		return float64(n)
	case int32:
		return float64(n)
	case int64:
		return float64(n)
	case uint:
		return float64(n)
	case uint8:
		return float64(n)
	case uint16:
		return float64(n)
	case uint32:
		return float64(n)
	case uint64:
		return float64(n)
	default:
		return 0
	}
}

// Execute 执行Lua脚本，返回顶层 return 的表（map[string]interface{}）。无返回或非表时返回空map。
func (e *LuaEngine) Execute(script string) (map[string]interface{}, error) {
	L := e.L
	fn, err := L.LoadString(script)
	if err != nil {
		return nil, fmt.Errorf("compile error: %w", err)
	}
	// 将编译后的函数压栈
	L.Push(fn)
	base := L.GetTop() - 1 // 函数压栈后，base 为函数之前的位置
	// 固定接收 1 个返回值
	if err := L.PCall(0, 1, nil); err != nil {
		return nil, fmt.Errorf("execute error: %w", err)
	}
	top := L.GetTop()
	nret := top - base
	var result map[string]interface{}
	if nret > 0 {
		if tbl, ok := L.Get(base + 1).(*lua.LTable); ok {
			result = luaTableToMap(tbl)
		} else {
			result = map[string]interface{}{}
		}
	}
	// 清栈，仅保留之前状态
	L.SetTop(base)
	if result == nil {
		result = map[string]interface{}{}
	}
	return result, nil
}

// ExecuteFile 执行Lua文件，返回顶层 return 的表（map[string]interface{}）。
func (e *LuaEngine) ExecuteFile(filename string) (map[string]interface{}, error) {
	L := e.L
	fn, err := L.LoadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("compile file error: %w", err)
	}
	L.Push(fn)
	base := L.GetTop() - 1
	// 固定接收 1 个返回值
	if err := L.PCall(0, 1, nil); err != nil {
		return nil, fmt.Errorf("execute file error: %w", err)
	}
	top := L.GetTop()
	nret := top - base
	var result map[string]interface{}
	if nret > 0 {
		if tbl, ok := L.Get(base + 1).(*lua.LTable); ok {
			result = luaTableToMap(tbl)
		} else {
			result = map[string]interface{}{}
		}
	}
	L.SetTop(base)
	if result == nil {
		result = map[string]interface{}{}
	}
	return result, nil
}

// luaTableToMap 将 *lua.LTable 转为 map[string]interface{}
func luaTableToMap(t *lua.LTable) map[string]interface{} {
	out := make(map[string]interface{})
	t.ForEach(func(k, v lua.LValue) {
		key := k.String()
		out[key] = luaToGo(v)
	})
	return out
}

// luaToGo 将 LValue 转为 Go 值
func luaToGo(v lua.LValue) interface{} {
	switch v.Type() {
	case lua.LTNil:
		return nil
	case lua.LTBool:
		return bool(v.(lua.LBool))
	case lua.LTNumber:
		return float64(v.(lua.LNumber))
	case lua.LTString:
		return v.String()
	case lua.LTTable:
		m := map[string]interface{}{}
		arr := []interface{}{}
		isArray := true
		idx := 1
		v.(*lua.LTable).ForEach(func(kk, vv lua.LValue) {
			// 判断是否为顺序数组（1..n）
			if isArray && kk.Type() == lua.LTNumber && int(lua.LVAsNumber(kk)) == idx {
				arr = append(arr, luaToGo(vv))
				idx++
			} else {
				isArray = false
				m[kk.String()] = luaToGo(vv)
			}
		})
		if isArray {
			return arr
		}
		return m
	default:
		return v.String()
	}
}

// Close 关闭Lua引擎
func (e *LuaEngine) Close() {
	e.CloseOutput()
	e.L.Close()
}

// Enqueue 将系统消息写入输出通道，保证顺序（阻塞写入）
func (e *LuaEngine) Enqueue(msg string) {
	// 如果有人读取，该写入将按顺序阻塞直到被消费
	e.output <- msg
}

// luaDisabledFunction 返回一个函数，用于输出禁用信息并返回错误值
func (e *LuaEngine) luaDisabledFunction(funcName string) func(*lua.LState) int {
	return func(L *lua.LState) int {
		errorMsg := fmt.Sprintf("[SECURITY] 函数 '%s' 已被禁用，出于安全考虑不允许执行", funcName)
		// 发送错误信息到输出通道
		select {
		case e.output <- fmt.Sprintf("[ERROR] %s", errorMsg):
		default:
			// 如果通道满了，丢弃消息
		}
		// 返回 nil 和错误信息，而不是抛出错误
		L.Push(lua.LNil)
		L.Push(lua.LString(errorMsg))
		return 2
	}
}

// luaDisabledTable 返回一个函数，用于创建包含禁用方法的表
func (e *LuaEngine) luaDisabledTable(tableName string) func(*lua.LState) int {
	return func(L *lua.LState) int {
		// 创建一个表，包含所有被禁用的方法
		table := L.CreateTable(0, 10)

		// 为每个可能的方法添加禁用函数
		disabledMethods := []string{"open", "popen", "close", "read", "write", "flush", "seek", "lines", "input", "output"}
		for _, method := range disabledMethods {
			table.RawSetString(method, L.NewFunction(func(L *lua.LState) int {
				errorMsg := fmt.Sprintf("[SECURITY] 函数 '%s.%s' 已被禁用，出于安全考虑不允许执行", tableName, method)
				// 发送错误信息到输出通道
				select {
				case e.output <- fmt.Sprintf("[ERROR] %s", errorMsg):
				default:
					// 如果通道满了，丢弃消息
				}
				// 返回 nil 和错误信息
				L.Push(lua.LNil)
				L.Push(lua.LString(errorMsg))
				return 2
			}))
		}

		L.Push(table)
		return 1
	}
}

// luaSafeOs 提供安全的 os 函数
func (e *LuaEngine) luaSafeOs(L *lua.LState) int {
	// 创建安全的 os 表
	osTable := L.CreateTable(0, 10)

	// os.time([t]) - 获取当前时间戳或从表创建时间戳
	osTable.RawSetString("time", L.NewFunction(func(L *lua.LState) int {
		if L.GetTop() == 0 {
			// 无参数：返回当前时间戳
			L.Push(lua.LNumber(time.Now().Unix()))
		} else {
			// 有参数：从表创建时间戳
			tbl := L.CheckTable(1)
			tm := time.Now()

			// 从表中读取时间字段
			if year := tbl.RawGetString("year"); year != lua.LNil {
				if y, ok := year.(lua.LNumber); ok {
					tm = time.Date(int(y), tm.Month(), tm.Day(), tm.Hour(), tm.Minute(), tm.Second(), 0, tm.Location())
				}
			}
			if month := tbl.RawGetString("month"); month != lua.LNil {
				if m, ok := month.(lua.LNumber); ok {
					tm = time.Date(tm.Year(), time.Month(int(m)), tm.Day(), tm.Hour(), tm.Minute(), tm.Second(), 0, tm.Location())
				}
			}
			if day := tbl.RawGetString("day"); day != lua.LNil {
				if d, ok := day.(lua.LNumber); ok {
					tm = time.Date(tm.Year(), tm.Month(), int(d), tm.Hour(), tm.Minute(), tm.Second(), 0, tm.Location())
				}
			}
			if hour := tbl.RawGetString("hour"); hour != lua.LNil {
				if h, ok := hour.(lua.LNumber); ok {
					tm = time.Date(tm.Year(), tm.Month(), tm.Day(), int(h), tm.Minute(), tm.Second(), 0, tm.Location())
				}
			}
			if min := tbl.RawGetString("min"); min != lua.LNil {
				if m, ok := min.(lua.LNumber); ok {
					tm = time.Date(tm.Year(), tm.Month(), tm.Day(), tm.Hour(), int(m), tm.Second(), 0, tm.Location())
				}
			}
			if sec := tbl.RawGetString("sec"); sec != lua.LNil {
				if s, ok := sec.(lua.LNumber); ok {
					tm = time.Date(tm.Year(), tm.Month(), tm.Day(), tm.Hour(), tm.Minute(), int(s), 0, tm.Location())
				}
			}

			L.Push(lua.LNumber(tm.Unix()))
		}
		return 1
	}))

	// os.date([format, t]) - 格式化时间
	osTable.RawSetString("date", L.NewFunction(func(L *lua.LState) int {
		var format string
		var timestamp int64

		top := L.GetTop()
		if top == 0 {
			// 无参数：使用默认格式和当前时间
			format = "%c"
			timestamp = time.Now().Unix()
		} else if top == 1 {
			// 一个参数：可能是格式或时间戳
			arg1 := L.Get(1)
			if formatStr, ok := arg1.(lua.LString); ok {
				// 第一个参数是格式字符串
				format = string(formatStr)
				timestamp = time.Now().Unix()
			} else if timeNum, ok := arg1.(lua.LNumber); ok {
				// 第一个参数是时间戳
				format = "%c"
				timestamp = int64(timeNum)
			} else {
				L.Push(lua.LString(""))
				return 1
			}
		} else {
			// 两个参数：格式和时间戳
			format = L.CheckString(1)
			timestamp = int64(L.CheckNumber(2))
		}

		// 转换时间戳为时间
		t := time.Unix(timestamp, 0)

		// 根据格式返回结果
		switch format {
		case "*t":
			// 返回时间表
			timeTable := L.CreateTable(0, 8)
			timeTable.RawSetString("year", lua.LNumber(t.Year()))
			timeTable.RawSetString("month", lua.LNumber(t.Month()))
			timeTable.RawSetString("day", lua.LNumber(t.Day()))
			timeTable.RawSetString("hour", lua.LNumber(t.Hour()))
			timeTable.RawSetString("min", lua.LNumber(t.Minute()))
			timeTable.RawSetString("sec", lua.LNumber(t.Second()))
			timeTable.RawSetString("wday", lua.LNumber(int(t.Weekday())+1)) // Lua 中周日是1
			timeTable.RawSetString("yday", lua.LNumber(t.YearDay()))
			L.Push(timeTable)
		default:
			// 简单格式化（支持基本格式）
			result := t.Format("2006-01-02 15:04:05")
			L.Push(lua.LString(result))
		}

		return 1
	}))

	// os.exit([code]) - 安全退出
	osTable.RawSetString("exit", L.NewFunction(func(L *lua.LState) int {
		// 安全的退出，不传递退出码给系统
		L.Close()
		return 0
	}))

	// os.clock() - 获取程序运行时间（秒）
	osTable.RawSetString("clock", L.NewFunction(func(L *lua.LState) int {
		// 返回程序启动以来的CPU时间（秒）
		L.Push(lua.LNumber(float64(time.Now().UnixNano()) / 1e9))
		return 1
	}))

	// 为不安全的 os 方法提供禁用提示
	osTable.RawSetString("execute", e.L.NewFunction(e.luaDisabledFunction("os.execute")))
	osTable.RawSetString("remove", e.L.NewFunction(e.luaDisabledFunction("os.remove")))
	osTable.RawSetString("rename", e.L.NewFunction(e.luaDisabledFunction("os.rename")))
	osTable.RawSetString("tmpname", e.L.NewFunction(e.luaDisabledFunction("os.tmpname")))
	osTable.RawSetString("getenv", e.L.NewFunction(e.luaDisabledFunction("os.getenv")))
	osTable.RawSetString("setlocale", e.L.NewFunction(e.luaDisabledFunction("os.setlocale")))

	L.Push(osTable)
	return 1
}
