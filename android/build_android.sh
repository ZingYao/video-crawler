#!/bin/bash
set -euo pipefail

# 路径
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ANDROID_DIR="${REPO_ROOT}/android"
FRONTEND_DIR="${REPO_ROOT}/frontend"

# 环境检查
command -v node >/dev/null 2>&1 || { echo "需要 Node.js，请先安装"; exit 1; }
command -v npm >/dev/null 2>&1 || { echo "需要 npm，请先安装"; exit 1; }

# 1) 构建前端
echo "[1/3] 构建前端..."
cd "${FRONTEND_DIR}"
if [ -f package-lock.json ]; then
  npm ci --no-audit --no-fund
else
  npm install --no-audit --no-fund
fi
npm run build

# 2) 构建 Go JNI 动态库
echo "[2/3] 构建 Go JNI 库..."
bash "${ANDROID_DIR}/build_go_lib.sh"

# 3) 构建 Android APK
echo "[3/3] 构建 Android APK..."
cd "${ANDROID_DIR}"
if [ ! -x ./gradlew ]; then
  echo "未检测到 gradlew，尝试生成 wrapper..."
  command -v gradle >/dev/null 2>&1 || { echo "需要 Gradle 或 Android Studio 生成 gradle wrapper"; exit 1; }
  gradle wrapper --gradle-version 8.7
fi
./gradlew :app:assembleDebug

APK_PATH="${ANDROID_DIR}/app/build/outputs/apk/debug/app-debug.apk"
if [ -f "${APK_PATH}" ]; then
  echo "✅ 构建完成: ${APK_PATH}"
else
  echo "⚠️ 未找到 APK，请检查 Gradle 输出"
fi
