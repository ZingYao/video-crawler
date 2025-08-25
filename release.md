# 发布说明（master）

- 版本标签: v1.0
- 发布日期: 2025 年 8 月 25 日 16:13:26
- 形态: 纯 HTTP 服务

## 核心特性
- Gin 提供统一 REST API（/api/...），含健康检查与基础信息
- 视频源管理：列表/详情、保存/删除、状态检查与手动设置
- 脚本调试（Lua/JS）：流式输出（Chunked/SSE）与基础日志
- 前端（Vue3 + Vite）：视频源编辑/管理、观影/历史、统一样式
- 登录与鉴权（可选）：JWT，路由守卫与权限控制

## 构建与分发
- Makefile（HTTP 服务）
  - 单平台：`make build`
  - 多平台：`make build-all`
    - Linux: amd64/arm64/386/arm
    - macOS: amd64/arm64
    - Windows: amd64/386/arm64
    - Android: amd64/arm64（如需要）
  - 前端构建自动依赖 `build-frontend`
- 产物命名：`bin/video-crawler-{os}-{arch}[.exe]`

## 配置与运行
- 配置文件：`configs/config.yaml`（含 `env`, `auth.require_login` 等）
- 运行：
  - 开发一体：`make dev`
  - 仅后端：`go run cmd/video-crawler/main.go`
  - 仅前端：`cd frontend && npm run dev`

## 变更摘要
- feat: 视频源管理与脚本调试接口
- feat: 前端编辑页与基础观影/历史
- feat: Makefile 多平台构建与发布脚手架
- chore/docs: 目录结构与基础 README 完善
