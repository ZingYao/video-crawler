# Makefile 使用说明（中文） | [English](MAKEFILE_README_EN.md) | [返回主README](README.md)

本项目提供完善的 Makefile 以统一前端与后端的构建、运行、打包流程。支持 HTTP 服务与 Wails 桌面应用两条构建线，并提供多平台二进制产物构建。

## 约定
- 构建前端：所有后端构建目标都会自动依赖并执行 `frontend` 的打包（`npm run build`）。
- 产物目录：默认输出到 `bin/`。
- 版本信息：自动注入 `Version/BuildTime/GitCommit`。

## 目标摘要

- 基础
  - `deps`：安装 Go 和前端依赖
  - `fmt` / `vet` / `lint`：格式化与检查
  - `clean`：清理构建产物

- 运行
  - `run-http`：运行独立 HTTP 服务（开发）
  - `run-wails`：Wails 开发模式（热重载）
  - `dev`：前端 dev + 后端 main.go 同时启动（仅开发）

- 构建（当前平台）
  - `build-http`：HTTP 服务二进制（当前平台）
  - `build-wails`：Wails 桌面应用（当前平台）

- 矩阵构建
  - `build-http-all`：HTTP 服务 Linux/macOS/Windows 常见架构矩阵
    - linux/amd64, linux/arm64, linux/386, linux/arm
    - darwin/amd64, darwin/arm64
    - windows/amd64, windows/386, windows/arm64
  - `build-wails-all`：Wails 桌面（Linux/macOS/Windows 常见架构）
  - `build-wails-ios` / `build-wails-android`：移动端占位目标，提示所需 SDK（详见 Wails 文档）

- 发布
  - `release`：基于 `build-http-all` 生成跨平台压缩包（zip/tar.gz）

## 常用命令
```bash
# 安装依赖
make deps

# HTTP 服务：当前平台
make build-http

# HTTP 服务：多平台
make build-http-all

# Wails 桌面：当前平台
make build-wails

# Wails 桌面：常见桌面平台
make build-wails-all

# 运行 Wails 开发模式
make run-wails

# 运行 HTTP 开发服务
make run-http

# 生成发布包（HTTP）
make release
```

## 说明
- Wails 的 iOS 与 Android 构建需要本机安装对应 SDK/NDK 及签名/打包工具链，Makefile 仅提供入口与提示；实际产线配置请参考 `README_WAILS.md`。
- 在 CI/CD 中可直接调用上述目标组合以产出多平台产物。
