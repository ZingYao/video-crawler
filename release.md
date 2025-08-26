# 发布说明 V2.0

- 版本标签: v2.0
- 发布日期: 2025 年 8 月 26 日
- 形态: 纯 HTTP 服务 + 原生 Android 客户端（JNI）+ Wails 桌面应用

## 核心特性
- Android 客户端（JNI + WebView）
  - 内置 Go HTTP 服务：动态端口、自检健康、失败自动重启；启动/恢复自动检测端口存活
  - 文件上传/下载：`onShowFileChooser`、Blob 导出（Base64/JSON）、DownloadManager 下载、“另存为”
  - 外链策略：非 `127.0.0.1:{port}` 的链接自动用系统浏览器打开
  - 全屏播放：进入全屏自动横屏 + 沉浸式隐藏系统栏；退出恢复
  - 返回键：非根路径后退，根路径返回桌面
  - 启动提示：服务成功启动 Toast 显示端口；启动失败提示后延时退出
  - 持久化：Cookie 刷新；`AndroidSession` 持久化 `sessionStorage`；`AndroidKV` 跨端口持久化搜索/观看历史
  - 悬浮工具球：可拖拽，1 秒延迟吸附边缘（1/3 隐藏），径向菜单（刷新/首页/退出），随吸附边自适应布局
- 前端
  - 运行环境识别（Android/Wails/浏览器），自动获取服务端口
  - API 文档移动端优化：容器与表格 `overflow-x: auto` 防溢出
  - 历史记录：在 App 环境使用 `AndroidKV`，浏览器回退 `localStorage`
  - 移除“下拉触底刷新”（避免与正常交互冲突）

## 构建与分发
- Makefile
  - 一键构建：`make build-all-products`
    - HTTP（Linux/macOS/Windows/Android 全架构）+ Android APK（`android/build_android.sh`）+ Wails（Windows/Linux/macOS，自动打包 macOS DMG）
  - HTTP 全架构：`make build-http-all`（Android 平台强制 `CGO_ENABLED=1`，自动选择 NDK clang；`arm` 设置 `GOARM=7`）
  - Android 应用：`make build-android-app`（调用脚本构建 APK）
  - Wails 应用：`make build-wails`（Windows/Linux/macOS + DMG）
- 产物
  - HTTP 可执行：`bin/video-crawler-{os}-{arch}[.exe]`
  - Android APK：由 Gradle 生成（未纳入版本库，`.gitignore` 忽略 `android/app/release/`）
  - Wails 应用：`build/bin/*.app`、`*.exe`、`*`（macOS 同步产出 `.dmg`）

## 配置与运行
- 环境
  - Android 构建需 `ANDROID_NDK_HOME`、`JAVA_HOME`（JDK 17）
  - 首次推荐：`make deps && make build-all-products`
- 运行
  - HTTP 开发：`make dev`（前后端联动）或 `go run cmd/http-server/main.go`
  - 前端开发：`cd frontend && npm run dev`
  - Android：安装 APK 后自动在本地 `127.0.0.1:{port}` 提供服务并加载 WebView

## 变更摘要
- feat(android): JNI 集成、本地服务自检/自启、全屏横屏沉浸、外链外部浏览器
- feat(android): WebView 上传/下载、Blob 导出、“另存为”
- feat(android): Cookie/Session/历史记录持久化（`AndroidSession`/`AndroidKV`）
- feat(android): 悬浮工具球拖拽与延迟吸附、径向菜单（刷新/首页/退出）
- feat(frontend): API 文档移动端表格/容器横向滚动；环境识别与端口获取
- build(make): `build-all-products`/`build-http-all`/`build-android-app`/`build-wails` 新增；Android 强制 CGO + NDK clang 自动选择
- docs: 新增 `docs/ANDROID_BUILD.md`；`RELEASE.md/RELEASE_EN.md` 发布说明
- chore: `.gitignore` 忽略 `android/app/release/` 并清理历史大文件

---

# 发布说明 V1.0

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
