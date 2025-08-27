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
cat > "$CPP_DIR/video_crawler_jni.h" << 'EOF'
#ifndef VIDEO_CRAWLER_JNI_H
#define VIDEO_CRAWLER_JNI_H

#include <jni.h>

#ifdef __cplusplus
extern "C" {
#endif

// JNI函数声明
JNIEXPORT jint JNICALL Java_com_zing_video_1crawler_MainActivity_startHttpService(JNIEnv *env, jobject obj, jint port);
JNIEXPORT void JNICALL Java_com_zing_video_1crawler_MainActivity_stopHttpService(JNIEnv *env, jobject obj);

#ifdef __cplusplus
}
#endif

#endif // VIDEO_CRAWLER_JNI_H
EOF

# 将头文件复制到所有ABI目录
for abi_dir in arm64-v8a x86_64 armeabi-v7a x86; do
  mkdir -p "$JNI_LIBS_DIR/$abi_dir"
  cp "$CPP_DIR/video_crawler_jni.h" "$JNI_LIBS_DIR/$abi_dir/"
  echo "Copied header file to $JNI_LIBS_DIR/$abi_dir/video_crawler_jni.h"
done

echo "Generated header file: $CPP_DIR/video_crawler_jni.h"

echo "Android builds and JNI setup completed"
