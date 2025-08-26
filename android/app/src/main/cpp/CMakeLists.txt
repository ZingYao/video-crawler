# CMake config for building libgo_video_crawler.so
cmake_minimum_required(VERSION 3.22.1)
project(go_video_crawler LANGUAGES C)

add_library(go_video_crawler_jni SHARED go_video_crawler_jni.c)

find_library(log-lib log)
find_library(dl-lib dl)

target_link_libraries(go_video_crawler_jni ${log-lib} ${dl-lib})
