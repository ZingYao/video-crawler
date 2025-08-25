# Wails 桌面模式说明 | [English](README_WAILS_EN.md) | [返回项目中文README](README.md)

本项目支持两种运行形态：

- HTTP 服务（独立后端 + 浏览器前端）
- Wails 桌面应用（内置前端，后端以随机端口运行并通过 HTTP 通信）

## 关键特性

- 桌面端启动一个独立的 Gin HTTP 服务，监听“随机端口”，前端通过该端口访问原有的 HTTP API，原有路由保持不变
- 前端在 Wails 环境下会自动调用 `GetServerPort()` 获取端口并组装 BaseURL
- 已开启 CORS 以便桌面端 WebView 访问本地 HTTP 服务
- 登录可选：当 `auth.require_login=false` 时，所有接口跳过鉴权，前端隐藏与登录相关菜单
- 数据目录统一：视频源配置 `video-source.json`、用户 `users.json`、搜索/观看/登录历史均存储于应用数据目录
  - macOS: `~/Library/Application Support/video-crawler/`
  - Windows: `%AppData%/video-crawler/`
  - Linux: `~/.local/share/video-crawler/`

## 运行

```bash
# 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 开发模式
$GOPATH/bin/wails dev

# 构建桌面应用
$GOPATH/bin/wails build
```

也可使用 Makefile 目标（如已配置）：

```bash
# 桌面应用（开发/构建）
make run-wails
make build-wails

# 纯 HTTP 服务（开发/构建）
make run-http
make build-http
```

## 后端端口

- 应用启动后，会在日志中打印当前 HTTP 服务端口，例如：`Starting gin server on 0.0.0.0:57124`
- 前端也会在“接口文档”页面顶部显示当前端口，方便用 curl 直接调用

## 接口文档（Wails 专属菜单）

- 在桌面应用左侧菜单进入“接口文档”，或访问路由 `/api-docs`
- 顶部展示当前端口；文档按分类列出所有接口的调用方式、入参、响应示例，并提供 curl / JS / Python 示例

> 说明：Wails 下前端并不会直接调用 Go 绑定方法，而是复用与 HTTP 模式一致的 REST API（通过随机端口），因此接口与调用方式与 HTTP 服务完全一致。

## 配置文件

- 默认优先读取环境变量 `VIDEO_CRAWLER_CONFIG_DIR/config.yaml`
- 未设置时使用项目内 `configs/config.yaml`
- 缺省时 `auth.require_login=false`

## 常见问题

- 端口 404：Wails 重启会更换端口，前端已改为每次请求动态读取端口；若仍旧 404，请重启应用
- 权限提示：当无需登录时，后端 JWT 中间件自动跳过；前端也不再提示“未登录”

---

## 集成概要（来自 WAILS_INTEGRATION_SUMMARY）

### 架构与完成项
- 服务层复用（`internal/services/`），功能一致
- 双入口：Wails 主入口 `main.go` 与 HTTP 入口 `cmd/http-server/main.go`
- 配置读取适配 OS 应用目录

### 前端适配
- 自动环境检测（Wails/HTTP）与统一 API 工具（`frontend/src/utils/api.ts`）
- 路由守卫与菜单在“无需登录模式”下的差异化展示

### 构建与运行
- `make run-wails` / `wails dev` 开发
- `make build-wails` / `wails build` 构建
- `make run-http` / `go run cmd/http-server/main.go` 开发
- `make build-http` 构建

### 统一 API 用法（TS）
```ts
import { configAPI, videoSourceAPI, scriptAPI } from '@/utils/api'
const cfg = await configAPI.getConfig()
const list = await videoSourceAPI.getList()
const save = await videoSourceAPI.saveVideoSource(data)
const lua = await scriptAPI.testLua(script, method, params)
```

### 登录控制与路由
- 无需登录：隐藏用户菜单、登录/注册/用户管理 → 404，其余可访问
- 需要登录：完整鉴权与角色控制


