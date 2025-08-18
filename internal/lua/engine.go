package lua

import (
	"encoding/json"
	"fmt"
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

	// 注册HTML解析函数（链式入口）
	e.L.SetGlobal("parse_html", e.L.NewFunction(e.luaParseHtml))

	// 注册工具函数
	e.L.SetGlobal("print", e.L.NewFunction(e.luaPrint))
	e.L.SetGlobal("log", e.L.NewFunction(e.luaLog))
	e.L.SetGlobal("sleep", e.L.NewFunction(e.luaSleep))

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

	// 返回响应表
	responseTable := L.CreateTable(0, 4)
	responseTable.RawSetString("status_code", lua.LNumber(response.StatusCode))
	//去 body 的转义 和首位的引号
	var body map[string]string
	bodyContent := string(response.Body)
	if strings.HasPrefix(bodyContent, "\"") {
		err = json.Unmarshal([]byte(fmt.Sprintf("{\"body\":%s}", bodyContent)), &body)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
	} else {
		body = map[string]string{
			"body": bodyContent,
		}
	}
	fmt.Println("response body:", body["body"])
	responseTable.RawSetString("body", lua.LString(body["body"]))
	responseTable.RawSetString("url", lua.LString(response.URL))

	// 设置响应头
	headersTable := L.CreateTable(0, len(response.Headers))
	for key, value := range response.Headers {
		headersTable.RawSetString(key, lua.LString(value))
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

	// 返回响应表
	responseTable := L.CreateTable(0, 4)
	responseTable.RawSetString("status_code", lua.LNumber(response.StatusCode))
	responseTable.RawSetString("body", lua.LString(normalizeResponseBody(response.Body)))
	responseTable.RawSetString("url", lua.LString(response.URL))

	// 设置响应头
	headersTable := L.CreateTable(0, len(response.Headers))
	for key, value := range response.Headers {
		headersTable.RawSetString(key, lua.LString(value))
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

// normalizeResponseBody 尝试将响应体以 UTF-8 文本返回，避免出现大量转义显示
// 1) 如果是 JSON 字节，解码后以紧凑字符串返回
// 2) 否则按 UTF-8 直接转换为字符串
func normalizeResponseBody(body []byte) string {
	var js interface{}
	if len(body) > 0 && (body[0] == '{' || body[0] == '[') {
		if err := json.Unmarshal(body, &js); err == nil {
			b, err := json.Marshal(js)
			if err == nil {
				return string(b)
			}
		}
	}
	return string(body)
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

// Execute 执行Lua脚本
func (e *LuaEngine) Execute(script string) error {
	// 执行脚本
	if err := e.L.DoString(script); err != nil {
		return fmt.Errorf("execute error: %w", err)
	}

	return nil
}

// ExecuteFile 执行Lua文件
func (e *LuaEngine) ExecuteFile(filename string) error {
	if err := e.L.DoFile(filename); err != nil {
		return fmt.Errorf("execute file error: %w", err)
	}
	return nil
}

// Close 关闭Lua引擎
func (e *LuaEngine) Close() {
	e.CloseOutput()
	e.L.Close()
}
