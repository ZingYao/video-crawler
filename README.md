# 视频爬虫 (Video Crawler)

[English README](README_EN.md) | 中文说明 | [发布说明](release.md)

一个基于 Go + Vue3 的可视化视频爬虫/脚本运行平台，支持 Lua 与 JavaScript 两种脚本引擎、实时调试、链式 HTML 解析、前端本地 Monaco 编辑器、SSE/Chunked 流式输出等能力，并内置基于 Plyr + hls.js 的 HLS 播放器与观影页。

## 技术栈（已更新）

- 后端（Go）
  - Gin（HTTP 服务）
  - 原生 net/http 爬虫（返回 *http.Response），默认模拟浏览器请求头；支持转发前端请求头（跳过 Cookie/Host/Content-Length）
  - Lua 引擎（gopher-lua）：
    - 注入：`http_get/http_post/set_headers/set_cookies/set_user_agent/set_random_user_agent/get_user_agent/set_ua_2_current_request_ua`
    - HTML 解析：`parse_html` 与链式选择器（`select/select_one/first/eq/parent/children/next/prev/attr/text/html`）
    - 工具：`sleep/trim/split` 与 `json_encode/json_decode`
    - 安全：禁用 `io/os/package` 危险能力，仅允许 `os.time/os.exit/os.clock` 等安全方法，危险方法返回禁用提示
  - JavaScript 引擎（goja）：
    - 同步 `fetch(url, { method, headers, body, timeout, redirect })`，返回 Response：`ok/status/statusText/url/headers/text()/json()/arrayBuffer()/clone()`；Headers：`get/has/keys/values/entries/forEach`
    - HTTP/UA：`httpGet/httpPost/setHeaders/setCookies/setUserAgent/setRandomUserAgent/getUserAgent/setUaToCurrentRequestUa`
    - DOM：`parseHtml(html)` → Document/Element，支持 `querySelector/querySelectorAll/getElementById/getElementsByTagName/getElementsByClassName/text()/html()/attr()/innerText/innerHTML/getAttribute`
    - Console：完整 `console` API（`log/info/warn/error/debug/trace/time/timeEnd/assert/group/groupCollapsed/groupEnd/count/countReset/table/dir/dirxml/clear`）并流式回传前端
    - 安全：沙箱环境，无 `os/fs/child_process` 等本地能力
  - goquery（HTML 解析）
  - github.com/lib4u/fake-useragent（随机 UA）
  - 日志：按 `env=dev` 输出到控制台

- 前端（Vue3 + TS + Vite）
  - Ant Design Vue、Pinia
  - @guolao/vue-monaco-editor + 本地 Monaco 资源（无 CDN，按需加载，绝对路径 Worker 修复）
  - 播放器：Plyr + hls.js，支持 m3u8（HLS）、倍速选择（移动端下拉）、上下集、剧集 Tab、进度缓存与续播、移动端全屏自动横竖屏
  - 统一绿色主题；编辑器/日志并排、支持拖拽分栏宽度与持久化、日志彩色高亮、清空日志、F5 调试、禁用 Cmd/Ctrl+S
  - 全局请求拦截：`code === 6` 提示登录过期并延迟跳转到登录页
  - 文档组件：LuaDocs 与 JSDocs，随脚本类型自动切换

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

5) 播放/观影
- 搜索页：输入关键词，选择站点类型后搜索，仅使用“正常状态”的站点
- 结果卡片：整卡可点击开始观看，保留“原站点”按钮
- 播放页：自动选第一源第一集（若无缓存），支持切源/切集、自动播放与续播

## 调试接口

### 基础调试
- Lua (Chunked)：`POST /api/lua/test`
- Lua (SSE)：`POST /api/lua/test-sse`
- JavaScript (Chunked)：`POST /api/js/test`
- JavaScript (SSE)：`POST /api/js/test-sse`

请求体：
```json
{ "script": "print('hello'); return { x = 1 }" }
```

输出顺序严格：`[INFO]` → Lua `[PRINT]/[LOG]` → `[RESULT]` → `[INFO]` 完成

### 高级调试
- Lua 高级调试 (SSE)：`POST /api/lua/advanced-test-sse`
- JavaScript 高级调试 (SSE)：`POST /api/js/advanced-test-sse`

请求体：
```json
{
  "script": "脚本内容",
  "method": "search_video|get_video_detail|get_play_video_detail",
  "params": {
    "keyword": "搜索关键词" // 或 "video_url": "视频链接"
  }
}
```

高级调试功能：
- 支持三种方法：搜索视频、获取视频详情、获取播放链接
- 自动验证脚本必需函数
- 返回原始结果和结构体转换结果对比
- 支持调试参数缓存（按站点ID隔离）
- 实时日志输出，支持展开/收起和自动滚动
- 代码差异对比，支持折叠相同内容

## 前端编辑页（视频源）

- 字段：站点名称、站点域名、排序值、资源类型、爬虫引擎（Lua/JavaScript）与状态
- 脚本类型：Lua / JavaScript；编辑器语言与文档随引擎自动切换
- 模板：Lua 与 JS 均提供默认模板与 Demo，可一键填充；切换语言可自动填充 Demo（当该语言无历史草稿/脚本）
- 必需函数校验：`search_video` / `get_video_detail` / `get_play_video_detail`
- 草稿：定时保存、进入页提示恢复/删除（删除二次确认）
- 交互：F5 调试、Cmd/Ctrl+S 禁用、可拖拽分栏并持久化、清空日志、自动滚动

### 高级调试功能

- **实时调试**：支持基础调试和高级调试两种模式
- **参数缓存**：调试参数按站点ID自动缓存，切换站点时自动恢复
- **日志管理**：支持日志展开/收起、自动滚动、清空日志
- **结果对比**：原始结果与结构体转换结果并排显示
- **差异高亮**：使用自定义 TextDiffViewer 组件，支持：
  - GoLand 风格的左右对比视图
  - 红色/绿色背景高亮差异
  - 行号显示
  - 折叠相同内容（默认隐藏）
  - 上下文行显示（默认10行）
  - 移动端响应式适配
- **草稿对比**：发现草稿时显示代码对比弹窗，支持选择版本

### JavaScript 脚本规范

- 全局方法（驼峰命名）：`httpGet`、`httpPost`、`setHeaders`、`setCookies`、`setUserAgent`、`setRandomUserAgent`、`getUserAgent`、`setUaToCurrentRequestUa`、`fetch`
- DOM：`parseHtml(html)` → `Document`/`Element`，提供 `querySelector/querySelectorAll/.../text/html/attr` 等
- Console：完整 `console` API，输出回流到调试面板
- Demo：在“填充完整 Demo”按钮中包含所有 API 的调用示例

## 配置

`configs/config.yaml` 关键项：
```yaml
env: dev  # dev 环境打印日志到控制台
auth:
  require_login: true  # 是否需要登录注册: true 或 false
```

### 登录控制配置

系统支持通过配置文件控制是否需要登录注册功能：

- **启用登录功能**：`auth.require_login: true`
  - 需要用户登录才能访问系统
  - 显示用户信息、个人中心、用户管理等菜单
  - 所有接口都需要JWT认证

- **禁用登录功能**：`auth.require_login: false`
  - 无需登录即可访问系统
  - 隐藏用户相关菜单和界面元素
  - 所有接口跳过JWT认证
  - 自动重定向登录注册页面到首页

**注意**：配置获取不到或没有对应字段时，默认为 `false`（不需要登录）

## 许可

MIT License

---

## 播放器交互优化（面向中国用户的使用习惯）

为提升移动端与触屏场景的观影体验，我们对视频播放标签（Plyr + 原生 HTML5 video）进行了以下优化，这些交互更贴近中国用户的习惯：

- 双击播放/暂停：视频区域双击可快速切换播放/暂停（替代早期的双击快进/快退）。
- 长按倍速（2x）：按住视频区域 500ms 进入 2x 播放，松开后恢复原速；与进度拖动互斥，避免误触。
- 左右滑动调节进度：支持反复来回拖动；当纵向位移超过容器高度 1/4 或触发区位于顶部/底部 1/6 时忽略，避免下拉通知栏/上滑桌面造成误触。
- 进度条常驻策略：开始拖动即常驻显示进度条，手指抬起后才允许收起；拖动暂停不收起。
- 移动端控件精简：移动端隐藏音量滑块，保留静音键与倍速设置，使界面更干净。

上述优化同时作用于 Plyr 与原生 HTML5 video，避免不同内核在移动端上的体验割裂。

## 构建与跨平台编译（已更新）

- 常用命令：
  - `make build`：本机构建（CGO=0）
  - `make build-linux|build-darwin|build-windows`：分别构建对应系统（CGO=0）
  - `make build-android`：构建 Android amd64/arm64（CGO=1，使用 NDK clang）
  - `make build-all`：构建所有平台（Android 由脚本单独处理，详见下文）

- Android 构建说明：
  - 使用脚本 `scripts/build_android.sh`，自动检测 `ANDROID_NDK_HOME/ANDROID_NDK_ROOT`
  - 若未配置，将自动在项目根目录 `.ndk/` 下载并解压 `android-ndk-r29-beta3-darwin.zip`
  - 自动选择 NDK 预编译目录（`darwin-arm64` 或 `darwin-x86_64`）并设置 `CC`
  - 产物输出到 `bin/video-crawler-android-{amd64,arm64}`

- 注意事项：
  - 为避免 Shell 续行解析问题，Android 构建被提取到独立脚本中；`build-all` 中会跳过 `android/*`，并在循环结束后调用脚本统一生成 Android 产物
  - `.gitignore` 已忽略 `.ndk/` 与其中的 NDK 压缩包，避免提交到仓库
