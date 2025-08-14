package lua

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	lua "github.com/yuin/gopher-lua"

	"video-crawler/internal/crawler"
)

// LuaEngine Lua引擎
type LuaEngine struct {
	L       *lua.LState
	browser crawler.BrowserRequest
}

// NewLuaEngine 创建新的Lua引擎
func NewLuaEngine(browser crawler.BrowserRequest) *LuaEngine {
	L := lua.NewState()
	engine := &LuaEngine{
		L:       L,
		browser: browser,
	}

	// 注册所有函数到Lua
	engine.registerFunctions()

	return engine
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

	// 注册HTML解析函数
	e.L.SetGlobal("parse_html", e.L.NewFunction(e.luaParseHtml))
	e.L.SetGlobal("select", e.L.NewFunction(e.luaSelect))
	e.L.SetGlobal("select_one", e.L.NewFunction(e.luaSelectOne))
	e.L.SetGlobal("attr", e.L.NewFunction(e.luaAttr))
	e.L.SetGlobal("text", e.L.NewFunction(e.luaText))
	e.L.SetGlobal("html", e.L.NewFunction(e.luaHtml))

	// 注册工具函数
	e.L.SetGlobal("print", e.L.NewFunction(e.luaPrint))
	e.L.SetGlobal("log", e.L.NewFunction(e.luaLog))
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
	responseTable.RawSetString("body", lua.LString(string(response.Body)))
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
	responseTable.RawSetString("body", lua.LString(string(response.Body)))
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

// luaParseHtml Lua中的parse_html函数
func (e *LuaEngine) luaParseHtml(L *lua.LState) int {
	html := L.CheckString(1)

	_, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// 创建文档对象
	docTable := L.CreateTable(0, 1)
	docTable.RawSetString("_doc", lua.LString(html))                  // 存储原始HTML
	docTable.RawSetString("_goquery_doc", lua.LString("goquery_doc")) // 标记为goquery文档

	L.Push(docTable)
	L.Push(lua.LNil) // 错误为nil
	return 2
}

// luaSelect Lua中的select函数
func (e *LuaEngine) luaSelect(L *lua.LState) int {
	docTable := L.CheckTable(1)
	selector := L.CheckString(2)

	html := docTable.RawGetString("_doc").String()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// 执行选择器
	selection := doc.Find(selector)

	// 创建结果数组
	resultTable := L.CreateTable(selection.Length(), 0)
	selection.Each(func(i int, s *goquery.Selection) {
		elementTable := L.CreateTable(0, 3)
		html, _ := s.Html()
		fmt.Printf("select html: %q\n", html)

		// 如果html不包含标签，说明是文本内容，需要获取完整的元素HTML
		if !strings.Contains(html, "<") {
			// 获取标签名和所有属性
			node := s.Get(0)
			tagName := node.Data

			// 构建属性字符串
			var attrs []string
			for _, attr := range node.Attr {
				attrs = append(attrs, fmt.Sprintf(`%s="%s"`, attr.Key, attr.Val))
			}

			attrStr := ""
			if len(attrs) > 0 {
				attrStr = " " + strings.Join(attrs, " ")
			}

			html = fmt.Sprintf("<%s%s>%s</%s>", tagName, attrStr, html, tagName)
		}

		elementTable.RawSetString("_html", lua.LString(html))
		elementTable.RawSetString("_text", lua.LString(s.Text()))
		elementTable.RawSetString("_selection", lua.LString("goquery_selection"))
		resultTable.Append(elementTable)
	})

	L.Push(resultTable)
	L.Push(lua.LNil) // 错误为nil
	return 2
}

// luaSelectOne Lua中的select_one函数
func (e *LuaEngine) luaSelectOne(L *lua.LState) int {
	docTable := L.CheckTable(1)
	selector := L.CheckString(2)

	html := docTable.RawGetString("_doc").String()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// 执行选择器，只取第一个
	selection := doc.Find(selector).First()

	if selection.Length() == 0 {
		L.Push(lua.LNil)
		L.Push(lua.LString("no element found"))
		return 2
	}

	// 创建元素对象
	elementTable := L.CreateTable(0, 3)

	// 获取完整的HTML（包含元素本身）
	htmlContent, _ := selection.Html()
	fmt.Printf("htmlContent: %q ,selector: %s\n", htmlContent, selector)

	// 如果htmlContent不包含标签，说明是文本内容，需要获取完整的元素HTML
	if !strings.Contains(htmlContent, "<") {
		// 获取标签名和所有属性
		node := selection.Get(0)
		tagName := node.Data

		// 构建属性字符串
		var attrs []string
		for _, attr := range node.Attr {
			attrs = append(attrs, fmt.Sprintf(`%s="%s"`, attr.Key, attr.Val))
		}

		attrStr := ""
		if len(attrs) > 0 {
			attrStr = " " + strings.Join(attrs, " ")
		}

		htmlContent = fmt.Sprintf("<%s%s>%s</%s>", tagName, attrStr, htmlContent, tagName)
	}

	elementTable.RawSetString("_html", lua.LString(htmlContent))
	elementTable.RawSetString("_text", lua.LString(selection.Text()))
	elementTable.RawSetString("_selection", lua.LString("goquery_selection"))

	L.Push(elementTable)
	L.Push(lua.LNil) // 错误为nil
	return 2
}

// luaAttr Lua中的attr函数
func (e *LuaEngine) luaAttr(L *lua.LState) int {
	elementTable := L.CheckTable(1)
	attrName := L.CheckString(2)

	// 从元素表中获取HTML内容
	html := elementTable.RawGetString("_html").String()

	// 使用goquery解析HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// 获取第一个元素的属性值
	selection := doc.Find("*").First()
	if selection.Length() == 0 {
		L.Push(lua.LNil)
		L.Push(lua.LString("no element found"))
		return 2
	}

	// 使用goquery的Attr方法获取属性
	attrValue, exists := selection.Attr(attrName)
	if !exists {
		L.Push(lua.LNil)
		L.Push(lua.LString("attribute not found"))
		return 2
	}

	L.Push(lua.LString(attrValue))
	L.Push(lua.LNil) // 错误为nil
	return 2
}

// luaText Lua中的text函数
func (e *LuaEngine) luaText(L *lua.LState) int {
	elementTable := L.CheckTable(1)
	text := elementTable.RawGetString("_text").String()
	L.Push(lua.LString(text))
	return 1
}

// luaHtml Lua中的html函数
func (e *LuaEngine) luaHtml(L *lua.LState) int {
	elementTable := L.CheckTable(1)

	// 从元素表中获取HTML内容
	html := elementTable.RawGetString("_html").String()

	// 如果HTML为空，返回空字符串
	if html == "" {
		L.Push(lua.LString(""))
		return 1
	}
	fmt.Printf("html: %q\n", html)
	// 使用goquery解析HTML以确保格式正确
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		// 如果解析失败，返回原始HTML
		L.Push(lua.LString(html))
		return 1
	}

	// 获取第一个元素的HTML
	selection := doc.Find("*").First()
	if selection.Length() == 0 {
		L.Push(lua.LString(html))
		return 1
	}

	// 使用goquery的Html方法获取格式化后的HTML
	formattedHtml, err := selection.Html()
	if err != nil {
		L.Push(lua.LString(html))
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
	fmt.Println(strings.Join(args, " "))
	return 0
}

// luaLog Lua中的log函数
func (e *LuaEngine) luaLog(L *lua.LState) int {
	args := make([]string, L.GetTop())
	for i := 1; i <= L.GetTop(); i++ {
		args[i-1] = L.Get(i).String()
	}
	fmt.Printf("[LUA] %s\n", strings.Join(args, " "))
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
	e.L.Close()
}
