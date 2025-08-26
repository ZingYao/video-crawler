#include <jni.h>
#include <stdlib.h>
#include <dlfcn.h>
#include <sys/stat.h>
#include <sys/types.h>

// typedef for Go function
typedef int (*StartServerFn)(char* baseDir, int port);

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
