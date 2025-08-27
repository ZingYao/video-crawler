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

echo "Building for Android..."

# 调用build_go_lib.sh构建.so文件
echo "Calling build_go_lib.sh to build .so files..."
cd android
./build_go_lib.sh
cd ..

# Android项目路径
ANDROID_PROJECT_DIR="android/app/src/main"
JNI_LIBS_DIR="$ANDROID_PROJECT_DIR/jniLibs"
CPP_DIR="$ANDROID_PROJECT_DIR/cpp"

# 确保目录存在
mkdir -p "$JNI_LIBS_DIR"
mkdir -p "$CPP_DIR"

# 生成头文件
echo "Generating header files..."

# 将头文件复制到所有ABI目录
for abi_dir in arm64-v8a x86_64 armeabi-v7a x86; do
  mkdir -p "$JNI_LIBS_DIR/$abi_dir"
done

echo "Android builds and JNI setup completed"
