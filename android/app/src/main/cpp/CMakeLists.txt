# CMake config for building libgo_video_crawler.so
cmake_minimum_required(VERSION 3.22.1)
project(go_video_crawler LANGUAGES C)

# 设置jniLibs目录路径
set(JNI_LIBS_DIR ${CMAKE_CURRENT_SOURCE_DIR}/../jniLibs/${ANDROID_ABI})

# 添加头文件路径
include_directories(${JNI_LIBS_DIR})

add_library(go_video_crawler_jni SHARED go_video_crawler_jni.c)

find_library(log-lib log)
find_library(dl-lib dl)

# 链接Go构建的库
target_link_libraries(go_video_crawler_jni ${log-lib} ${dl-lib} ${JNI_LIBS_DIR}/libvideo_crawler.so)
