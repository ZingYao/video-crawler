# v2.0 发布说明（重点：新客户端支持）

本次发布引入全新的原生 Android 客户端（JNI 集成）与多平台构建能力，显著提升移动端体验与分发效率。

## 新增 Android 客户端（JNI + WebView）
- 内置 Go HTTP 服务（动态端口、健康检查、自动重启），启动与恢复时自动检查端口存活。
- WebView 文件上传/下载：支持 `onShowFileChooser`；支持 Blob 导出（Base64/JSON）、DownloadManager 下载与“另存为”。
- 外链处理：非 `127.0.0.1:{port}` 的链接自动外部浏览器打开。
- 全屏视频：进入全屏自动横屏并隐藏系统栏（沉浸式），退出后恢复。
- 返回键逻辑：非根路径后退；根路径返回桌面。
- 启动提示：JNI 服务成功启动后 Toast 显示端口；失败则提示并延时退出。
- 存储持久化：Cookie 刷新；`AndroidSession` 持久化 `sessionStorage`；`AndroidKV` 跨端口持久化本地搜索/观看历史。
- 悬浮工具球（FAB）：可拖拽、1 秒延迟靠边吸附（1/3 隐藏）、径向菜单（刷新/首页/退出），吸附侧自动调整菜单布局。
- 网络安全：允许本地明文流量（`127.0.0.1/localhost`）。
- 细节修复：首页图标资源修复；菜单/标题展示优化。

## 前端改进
- 运行环境识别：支持 Android WebView 与 Wails，动态获取服务端口。
- API 文档移动端优化：内容容器与表格启用 `overflow-x: auto`，避免小屏溢出。
- 本地历史：搜索/观看历史在 App 环境使用 `AndroidKV`，浏览器回退 `localStorage`。
- 移除“下拉触底刷新”：避免与正常滚动/交互冲突。

## 构建与发布
- Makefile 一键构建：
  - `make build-all-products`：构建 HTTP（Linux/macOS/Windows/Android 全架构）、Android APK（调用 `android/build_android.sh`）、Wails（Windows/Linux/macOS，并打包 macOS DMG）。
  - `make build-http-all`：HTTP 全架构（Android 平台默认开启 CGO）。
  - `make build-android-app`：调用脚本快速构建 APK。
  - `make build-wails`：Windows/Linux/macOS；自动打包 macOS DMG。
- Android 构建文档：`docs/ANDROID_BUILD.md`（NDK、`--skip-frontend`、多 ABI 打包等）。
- 忽略与历史清理：`.gitignore` 新增 `android/app/release/`，移除误提交的大体积 APK，保证仓库轻量可推送。
- 脚本与 NDK：`android/build_go_lib.sh` 支持 `--skip-frontend`；多 ABI `.so`；CMake 生成 `go_video_crawler_jni` 并动态加载 `libgo_video_crawler.so`。

## 重要变更
- HTTP 服务入口调整为 `cmd/http-server/main.go`。
- 移除实验性的“下拉触底刷新”。
- Android 外链默认由外部浏览器打开，避免 WebView 跨站与隐私风险。

## 升级指引
- 拉取最新代码并安装依赖（Go/Node/NDK）。
- 推荐首次执行：
  - `make deps`
  - `make build-all-products`
- Android 构建需要配置 `ANDROID_NDK_HOME` 与 `JAVA_HOME`（JDK 17）。
- Android 13+ 请允许通知权限以显示下载完成提示。

## 已知问题
- 个别 ROM 对后台下载限制较多，DownloadManager 通知可能需要手动开启。
- 首次启动端口探测可能受权限/系统策略影响，异常会提示并退出，可重试或检查权限。
