# 视频爬虫 (Video Crawler)

[English README](README_EN.md) | 中文说明

一个基于 Go + Vue3 的可视化视频爬虫/脚本运行平台，支持 Lua 实时调试、链式 HTML 解析、前端本地 Monaco 编辑器、SSE/Chunked 流式输出等能力。

## 技术栈（已更新）

- 后端（Go）
  - Gin（HTTP 服务）
  - 原生 net/http 爬虫（返回 *http.Response），默认模拟浏览器请求头；支持转发前端请求头（跳过 Cookie/Host/Content-Length）
  - gopher-lua（Lua 引擎）：
    - 流式输出通道，`print`/`log` 均带时间戳
    - 捕获顶层 `return`，转 `map[string]interface{}` 并格式化 JSON 以 `[RESULT]` 顺序输出
    - 注入函数：
      - HTTP：`http_get`、`http_post`、`set_headers`、`set_cookies`、`set_user_agent`、`set_random_user_agent`、`get_user_agent`、`set_ua_2_current_request_ua`
      - HTML 链式：`parse_html`，并在 `Document/Selection` 上提供 `select/select_one/first/eq/parent/children/next/prev/attr/text/html`
      - 工具：`sleep(ms)`、`split(s, sep)`、`trim(s)`、`json_encode(value, indent?)`（支持布尔/数字/字符串缩进）、`json_decode(json)`
    - 响应体自动解压 `gzip/deflate`
  - goquery（HTML 解析，Lua 侧以 userdata 暴露链式 API）
  - github.com/lib4u/fake-useragent（随机 UA）
  - 日志：按 `env=dev` 输出到控制台

- 前端（Vue3 + TS + Vite）
  - Ant Design Vue、Pinia
  - @guolao/vue-monaco-editor + 本地 Monaco 资源（无 CDN）
  - 统一绿色主题；编辑器/日志并排、支持拖拽分栏宽度与持久化、日志彩色高亮、清空日志、F5 调试、禁用 Cmd/Ctrl+S
  - 全局请求拦截：`code === 6` 提示登录过期并延迟跳转到登录页
  - Lua 文档组件（右侧抽屉，无遮罩，可同时编辑）

## 项目结构（已精简）

```
video-crawler/
├── cmd/video-crawler/          # 主程序入口
├── internal/
│   ├── app/                    # Gin 启动与路由
│   ├── config/                 # 配置（含 env）
│   ├── handler/                # 入口注册
│   ├── controllers/            # 控制器层（含 /api/lua/test 与 /test-sse）
│   ├── services/               # 业务逻辑（Lua 脚本执行、顺序输出保障）
│   ├── crawler/                # 爬虫（原生 net/http），默认浏览器头+随机 UA
│   ├── lua/                    # Lua 引擎（函数注入、链式 HTML、返回值捕获）
│   ├── entities/ middleware/ utils/ ...
│   └── static/dist/            # 前端构建产物（自动拷贝）
├── frontend/                   # Vue3 前端
│   ├── src/
│   ├── public/monaco/          # 本地 Monaco 资源
│   └── scripts/copy-monaco.js  # 构建时复制 Monaco
├── configs/                    # 配置与示例
├── scripts/
├── Makefile
└── README.md / README_EN.md
```

## 快速开始

1) 初始化配置
```bash
cp configs/config.example.yaml configs/config.yaml
cp configs/users.example.json configs/users.json
cp configs/video-source.example.json configs/video-source.json
```

2) 构建前端（会本地化 Monaco 资源）
```bash
cd frontend && npm install && npm run build
```

3) 启动后端
```bash
make dev
# 或
go run cmd/video-crawler/main.go
```

4) 访问
- 前端: http://localhost:8080
- API:   http://localhost:8080/api

## Lua 调试接口

- Chunked：`POST /api/lua/test`
- SSE：`POST /api/lua/test-sse`

请求体：
```json
{ "script": "print('hello'); return { x = 1 }" }
```

输出顺序严格：`[INFO]` → Lua `[PRINT]/[LOG]` → `[RESULT]` → `[INFO]` 完成

## 前端编辑页（视频源）

- 字段：站点名称、站点域名、排序值、资源类型
- Lua 编辑：默认模板、必需函数校验（`search_video`/`get_video_detail`/`get_play_video_detail`）
- 草稿：定时保存、进入页提示恢复/删除（删除二次确认）
- 交互：F5 调试、Cmd/Ctrl+S 禁用、可拖拽分栏并持久化、清空日志、自动滚动

## 配置

`configs/config.yaml` 关键项：
```yaml
env: dev  # dev 环境打印日志到控制台
```

## 许可

MIT License
