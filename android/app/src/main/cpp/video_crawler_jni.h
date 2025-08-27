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
