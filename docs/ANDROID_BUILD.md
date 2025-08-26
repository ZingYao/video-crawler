# Android 构建说明（JNI + WebView）

本文档介绍如何在 Android 上构建并运行本项目的 JNI 后端与 APK。

## 1. 环境准备

- 操作系统：macOS / Linux / Windows（建议 macOS/Linux）
- JDK：17（必须）
  - macOS 可使用 `brew install temurin@17` 或者从 Azul/Adoptium 安装
  - 设置 `JAVA_HOME` 指向 JDK 17 目录
- Android SDK & NDK：
  - 安装 Android Studio，并通过 SDK Manager 安装：
    - Android SDK Platform 34（或兼容版本）
    - Android SDK Build-Tools（最新）
    - Android NDK（建议 r26d 或更高；如需 16KB page-size 支持，建议使用 r27+）
  - 设置环境：
    - `sdk.dir` 写入 `android/local.properties`（脚本会自动写入或可手工设置）
    - `ANDROID_NDK_HOME` 指向 NDK 根目录（非常重要）
- Node.js + npm（前端构建）
  - Node 18+/20+，npm 9+/10+
- Go：1.21+（推荐）

## 2. 目录概览

- `android/`：Android 工程与脚本
  - `app/src/main/cpp/`：JNI 桥接（C）
  - `app/src/main/jniLibs/`：JNI 产物（.so）输出目录
  - `go/mobile/`：JNI 动态库的 Go 源码（`c-shared`）
  - `build_go_lib.sh`：构建 JNI 的脚本
  - `build_android.sh`：一键构建 APK 的脚本
- `frontend/`：Web 前端（打包产物通过 go:embed 嵌入 / 或 WebView 加载）

## 3. 关键环境变量

- `JAVA_HOME`：JDK 17 路径
- `ANDROID_NDK_HOME`：Android NDK 路径（例如：`~/Library/Android/sdk/ndk/26.1.10909125`）
- 可选代理（如需）：`http_proxy` / `https_proxy`

## 4. 构建 JNI 动态库（libgo_video_crawler.so）

脚本位置：`android/build_go_lib.sh`

用法：

```bash
# 正常构建（包含前端构建）
./android/build_go_lib.sh

# 跳过前端构建（新增参数/环境变量）
./android/build_go_lib.sh --skip-frontend
# 或
SKIP_FRONTEND=1 ./android/build_go_lib.sh
```

说明：
- 默认先构建前端（供 go:embed 与静态资源使用），然后按 4 个 ABI 输出到：
  - `android/app/src/main/jniLibs/{arm64-v8a,armeabi-v7a,x86,x86_64}/libgo_video_crawler.so`
- 若你已手动构建前端或不需要更新前端，可使用 `--skip-frontend` 或 `SKIP_FRONTEND=1` 跳过前端构建以加快速度。

常见问题：
- 提示 `ANDROID_NDK_HOME` 未设置：请正确设置 NDK 根目录。
- r27+ NDK 可用于满足 Android 15 的 16KB 页面对齐要求（涉及链接器 flags）。

## 5. 构建 APK

一键脚本：`android/build_android.sh`

```bash
# 默认：构建 JNI（包含前端） + assembleDebug
./android/build_android.sh

# 结合跳过前端（加快调试效率）
SKIP_FRONTEND=1 ./android/build_android.sh
```

脚本行为：
- 调用 `build_go_lib.sh` 产出各 ABI 的 `libgo_video_crawler.so`
- 使用 gradle wrapper 构建 `:app:assembleDebug`
- 构建成功后在 `android/app/build/outputs/apk/debug/` 输出 APK

注意：
- 首次可能需要生成 gradle wrapper 或通过 Android Studio 同步依赖。
- `local.properties` 中的 `sdk.dir` 会被写入你的本机 SDK 路径。

## 6. 运行与调试

- 使用 Android Studio 打开 `android/`，选择真机或模拟器运行 `app`。
- 首次运行如果看到 `127.0.0.1 明文流量` 的提示：已在 `AndroidManifest.xml` 中开启了明文与 `network_security_config`，允许 WebView 访问本地服务。
- JNI 启动策略：
  - App 启动时检测上次端口是否存活；若不存活则通过 JNI 以端口 0 启动（自动分配），并重定向 WebView 到新端口。
  - 启动成功会有 Toast 显示实际端口。
  - 启动失败（端口返回 0）会提示用户重启，并于 3 秒后自动退出应用。

## 7. 功能与交互

- WebView 文件上传下载：
  - 上传：支持 `<input type="file">` 弹系统文件选择；
  - 下载：通过 `DownloadManager` 保存到“下载”目录，并显示通知。
- 权限：Android 13 以下会请求读存储权限；Android 13+ 使用分区存储。
- 悬浮工具球：
  - 可拖拽、吸边半隐藏；
  - 点击展开径向菜单（刷新 / 回首页 / 退出程序）。
- 持久化：
  - Cookie：开启并在生命周期 flush；
  - sessionStorage：注入脚本做影子持久化；
  - 搜索/观看历史：在 Android WebView 下使用原生 SharedPreferences（`AndroidKV`）跨端口持久化；非 Android 环境使用 `localStorage`。

## 8. 常见问题排查（FAQ）

1) Gradle 插件解析失败 / `com.android.application` 找不到：
- 确保 `android/settings.gradle` 中有 `pluginManagement { repositories { google(); gradlePluginPortal() } }`
- 移除顶层 `allprojects { repositories { ... } }` 冲突配置。

2) `SDK location not found.`：
- 在 `android/local.properties` 写入：`sdk.dir=/你的/Android/sdk`

3) `Manifest merger failed: android:exported`：
- 已在 `AndroidManifest.xml` 的 `MainActivity` 设置 `android:exported="true"`。

4) `Cleartext HTTP traffic to 127.0.0.1 not permitted`：
- 已开启 `usesCleartextTraffic=true` 并提供 `res/xml/network_security_config.xml`。

5) `UnsatisfiedLinkError: libgo_video_crawler_jni.so not found`：
- 确保 JNI 库与 CMake 产物命名一致并被正确打包；确认 `app/build.gradle` 中 `externalNativeBuild` 与 `CMakeLists.txt` 的目标一致。

6) 端口变化导致本地历史丢失：
- Android WebView 下已切换为 `AndroidKV`（SharedPreferences）持久化，端口变化（Origin 变化）也不会丢失。

7) 通知权限（Android 13+）：
- 若需下载通知，请在首次运行时授予 `POST_NOTIFICATIONS` 权限（已在 Manifest 中声明）。

## 9. 构建参数与可选项

- `./android/build_go_lib.sh --skip-frontend` 或 `SKIP_FRONTEND=1`：跳过前端构建。
- 若目标设备为 Android 15 且需要 16KB Page Size 对齐，可升级到 NDK r27+ 并在 C/C++ 链接参数中添加 `-Wl,--page-size=16384`（CMake/Go 分别处理）。

## 10. 目录大小与忽略项

- `.gitignore` 已忽略：
  - `android/.gradle/`、`android/build/`、`android/app/build/`、`android/app/.cxx/`
  - `android/app/src/main/jniLibs/**`
  - 前端 `frontend/node_modules/`、`frontend/dist/`
  - 下载的 NDK 压缩包及解压目录（避免误提交）

---

如需更多帮助或遇到未覆盖的问题，请在仓库提 Issue 或联系维护者。
