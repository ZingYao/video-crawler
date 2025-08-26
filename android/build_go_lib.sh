#!/bin/bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ANDROID_DIR="${REPO_ROOT}/android"
GO_SRC_DIR="${REPO_ROOT}/android/go/mobile"
OUT_DIR="${ANDROID_DIR}/app/src/main/jniLibs/arm64-v8a"

: "${ANDROID_NDK_HOME:?请先设置 ANDROID_NDK_HOME 指向 Android NDK 根目录}"

API=21
TRIPLE=aarch64-linux-android
CC="${ANDROID_NDK_HOME}/toolchains/llvm/prebuilt/darwin-*/bin/${TRIPLE}${API}-clang"

mkdir -p "${OUT_DIR}"

export CGO_ENABLED=1
export GOOS=android
export GOARCH=arm64
export CC=$(echo ${CC})

cd "${REPO_ROOT}"

go mod tidy

cd "${GO_SRC_DIR}"

# 清理旧库以避免混淆
rm -f "${OUT_DIR}/libhello.so" || true

go build -buildmode=c-shared -o "${OUT_DIR}/libgo_video_crawler.so"

echo "生成完成: ${OUT_DIR}/libgo_video_crawler.so"
