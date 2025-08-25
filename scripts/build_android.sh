#!/usr/bin/env bash
set -euo pipefail

BUILD_DIR=${BUILD_DIR:-bin}
BINARY_NAME=${BINARY_NAME:-video-crawler}
MAIN_PATH=${MAIN_PATH:-cmd/video-crawler/main.go}
VERSION=${VERSION:-dev}
BUILD_TIME=${BUILD_TIME:-unknown}
GIT_COMMIT=${GIT_COMMIT:-unknown}

# construct ldflags safely
GO_LDFLAGS=(
  -ldflags
  "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT} -s -w"
)

mkdir -p "$BUILD_DIR"

for arch in amd64 arm64; do
  echo "android/${arch} (CGO_ENABLED=1)"
  NDK_ROOT="${ANDROID_NDK_HOME:-}"
  if [[ -z "$NDK_ROOT" ]]; then NDK_ROOT="${ANDROID_NDK_ROOT:-}"; fi
  if [[ -z "$NDK_ROOT" ]]; then
    echo "[INFO] ANDROID_NDK_HOME/ROOT not set. Preparing local NDK cache..."
    NDK_CACHE_DIR=".ndk"; mkdir -p "$NDK_CACHE_DIR"
    NDK_ZIP="$NDK_CACHE_DIR/android-ndk-r29-beta3-darwin.zip"
    if [[ ! -f "$NDK_ZIP" ]]; then
      echo "[INFO] Downloading NDK r29-beta3..."
      curl -L --retry 3 -o "$NDK_ZIP" "https://dl.google.com/android/repository/android-ndk-r29-beta3-darwin.zip"
    fi
    if [[ ! -d "$NDK_CACHE_DIR/android-ndk-r29-beta3" ]]; then
      echo "[INFO] Extracting NDK..."
      unzip -q -o "$NDK_ZIP" -d "$NDK_CACHE_DIR"
    fi
    NDK_ROOT="$NDK_CACHE_DIR/android-ndk-r29-beta3"
  fi

  host_arch=$(uname -m)
  if [[ -d "$NDK_ROOT/toolchains/llvm/prebuilt/darwin-${host_arch}/bin" ]]; then
    NDK_BIN="$NDK_ROOT/toolchains/llvm/prebuilt/darwin-${host_arch}/bin"
  elif [[ -d "$NDK_ROOT/toolchains/llvm/prebuilt/darwin-x86_64/bin" ]]; then
    NDK_BIN="$NDK_ROOT/toolchains/llvm/prebuilt/darwin-x86_64/bin"
  else
    echo "[ERROR] Cannot locate NDK prebuilt bin under $NDK_ROOT/toolchains/llvm/prebuilt"; exit 3
  fi

  if [[ "$arch" == "arm64" ]]; then CC_PATH="$NDK_BIN/aarch64-linux-android21-clang"; else CC_PATH="$NDK_BIN/x86_64-linux-android21-clang"; fi
  if [[ ! -x "$CC_PATH" ]]; then echo "[ERROR] NDK clang not found at $CC_PATH"; exit 2; fi
  echo "Using CC=$CC_PATH"

  CGO_ENABLED=1 CC="$CC_PATH" GOOS=android GOARCH="$arch" go build "${GO_LDFLAGS[@]}" -o "$BUILD_DIR/$BINARY_NAME-android-$arch" "$MAIN_PATH"

done

echo "Android builds completed"
