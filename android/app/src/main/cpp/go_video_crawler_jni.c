#include <jni.h>
#include <stdlib.h>
#include <dlfcn.h>
#include <sys/stat.h>
#include <sys/types.h>

// typedef for Go functions
typedef int (*StartServerFn)(char* baseDir, int port);
typedef char* (*GetServerErrorFn)();
typedef char* (*GetServerStatusFn)();

static void ensure_dir(const char* path) {
    if (!path || !*path) return;
    mkdir(path, 0700);
}

JNIEXPORT jint JNICALL
Java_com_zing_video_1crawler_MainActivity_startServer(JNIEnv* env, jobject thiz, jstring baseDir, jint port) {
    const char* base_c = (*env)->GetStringUTFChars(env, baseDir, 0);
    if (base_c && *base_c) {
        setenv("VIDEO_CRAWLER_CONFIG_DIR", base_c, 1);
        ensure_dir(base_c);
    }

    void* handle = dlopen("libgo_video_crawler.so", RTLD_NOW);
    if (!handle) {
        if (base_c) (*env)->ReleaseStringUTFChars(env, baseDir, base_c);
        return 0;
    }
    StartServerFn fn = (StartServerFn)dlsym(handle, "StartServer");
    if (!fn) {
        dlclose(handle);
        if (base_c) (*env)->ReleaseStringUTFChars(env, baseDir, base_c);
        return 0;
    }

    int actual = fn((char*)base_c, (int)port);
    dlclose(handle);
    if (base_c) (*env)->ReleaseStringUTFChars(env, baseDir, base_c);
    return (jint)actual;
}

JNIEXPORT jstring JNICALL
Java_com_zing_video_1crawler_MainActivity_getServerError(JNIEnv* env, jobject thiz) {
    void* handle = dlopen("libgo_video_crawler.so", RTLD_NOW);
    if (!handle) {
        return (*env)->NewStringUTF(env, "");
    }
    
    GetServerErrorFn fn = (GetServerErrorFn)dlsym(handle, "GetServerError");
    if (!fn) {
        dlclose(handle);
        return (*env)->NewStringUTF(env, "");
    }
    
    char* error = fn();
    jstring result = (*env)->NewStringUTF(env, error ? error : "");
    dlclose(handle);
    return result;
}

JNIEXPORT jstring JNICALL
Java_com_zing_video_1crawler_MainActivity_getServerStatus(JNIEnv* env, jobject thiz) {
    void* handle = dlopen("libgo_video_crawler.so", RTLD_NOW);
    if (!handle) {
        return (*env)->NewStringUTF(env, "not_started");
    }
    
    GetServerStatusFn fn = (GetServerStatusFn)dlsym(handle, "GetServerStatus");
    if (!fn) {
        dlclose(handle);
        return (*env)->NewStringUTF(env, "not_started");
    }
    
    char* status = fn();
    jstring result = (*env)->NewStringUTF(env, status ? status : "not_started");
    dlclose(handle);
    return result;
}
