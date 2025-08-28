package com.zing.video_crawler;

import android.os.Bundle;
import android.graphics.Color;
import android.app.DownloadManager;
import android.net.Uri;
import android.os.Environment;
import android.webkit.DownloadListener;
import android.webkit.ValueCallback;
import android.webkit.WebChromeClient;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;
import android.webkit.URLUtil;
import android.content.Intent;
import android.Manifest;
import android.content.pm.PackageManager;
import androidx.core.app.ActivityCompat;
import androidx.core.content.ContextCompat;

import androidx.appcompat.app.AppCompatActivity;

// added imports
import android.widget.Toast;
import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.URL;
import java.net.URI;
import java.net.MalformedURLException;
import android.content.SharedPreferences;
import android.view.View;
import android.widget.FrameLayout;
import android.view.ViewGroup;
import android.content.pm.ActivityInfo;
import android.widget.ProgressBar;
import android.view.Gravity;
import android.view.MotionEvent;
import java.util.ArrayList;
import java.util.List;
import android.os.Handler;
import android.os.Looper;
import android.webkit.CookieManager;
import android.os.Build;
import androidx.core.view.WindowCompat;
import androidx.core.view.WindowInsetsControllerCompat;
import androidx.core.view.WindowInsetsCompat;
import android.widget.ImageView;
import android.view.ViewTreeObserver;
import android.graphics.drawable.GradientDrawable;
import android.view.ViewConfiguration;
import android.view.KeyEvent;

public class MainActivity extends AppCompatActivity {
    static {
        System.loadLibrary("go_video_crawler_jni");
    }

    private native int startServer(String baseDir, int port);

    private int actualPort = 0;
    private static final int REQUEST_FILE_CHOOSER = 1001;
    private static final int REQUEST_STORAGE_PERMISSION = 1002;
    private static final int REQUEST_CREATE_DOCUMENT_BLOB = 1003;
    private static final int REQUEST_CREATE_DOCUMENT_URL = 1004;
    private static final String PREFS_NAME = "vc_prefs";
    private static final String PREF_KEY_LAST_PORT = "last_port";
    private static final String PREF_KEY_SESSION_PREFIX = "session_snapshot_";
    private ValueCallback<Uri[]> filePathCallback;

    // temp holders for save-as
    private String pendingBase64Data;
    private String pendingFilename;
    private String pendingDownloadUrl;

    // hold WebView instance for back handling
    private WebView webView;

    // fullscreen video support
    private View customView;
    private FrameLayout fullScreenContainer;
    private WebChromeClient.CustomViewCallback customViewCallback;

    // root and loading
    private FrameLayout rootLayout;
    private ProgressBar loadingBar;

    // Start serialization
    private final Object startLock = new Object();
    private boolean isStarting = false;
    private final List<Runnable> startWaiters = new ArrayList<>();

    // JS 接口：Session 持久化
    private class AndroidSession {
        @android.webkit.JavascriptInterface
        public void save(String origin, String json) {
            if (origin == null) origin = "";
            getSharedPreferences(PREFS_NAME, MODE_PRIVATE)
                .edit()
                .putString(PREF_KEY_SESSION_PREFIX + origin, json == null ? "" : json)
                .apply();
        }
        @android.webkit.JavascriptInterface
        public String load(String origin) {
            if (origin == null) origin = "";
            return getSharedPreferences(PREFS_NAME, MODE_PRIVATE)
                    .getString(PREF_KEY_SESSION_PREFIX + origin, "");
        }
    }

    // JS 接口：Key-Value 持久化（跨端口）
    private class AndroidKV {
        @android.webkit.JavascriptInterface
        public void setItem(String key, String value) {
            if (key == null) return;
            getSharedPreferences(PREFS_NAME, MODE_PRIVATE)
                .edit()
                .putString(key, value == null ? "" : value)
                .apply();
        }
        @android.webkit.JavascriptInterface
        public String getItem(String key) {
            if (key == null) return null;
            return getSharedPreferences(PREFS_NAME, MODE_PRIVATE).getString(key, null);
        }
        @android.webkit.JavascriptInterface
        public void removeItem(String key) {
            if (key == null) return;
            getSharedPreferences(PREFS_NAME, MODE_PRIVATE)
                .edit()
                .remove(key)
                .apply();
        }
        @android.webkit.JavascriptInterface
        public boolean hasKey(String key) {
            if (key == null) return false;
            return getSharedPreferences(PREFS_NAME, MODE_PRIVATE).contains(key);
        }
    }

    private boolean isLocalUrl(String url) {
        try {
            URL u = new URL(url);
            String host = (u.getHost() == null) ? "" : u.getHost();
            int port = (u.getPort() == -1) ? ("https".equalsIgnoreCase(u.getProtocol()) ? 443 : 80) : u.getPort();
            if (host.equals("127.0.0.1") || host.equals("localhost") || host.equals("::1")) {
                int targetPort = actualPort > 0 ? actualPort : getSharedPreferences(PREFS_NAME, MODE_PRIVATE).getInt(PREF_KEY_LAST_PORT, 0);
                return targetPort > 0 && port == targetPort;
            }
            return false;
        } catch (MalformedURLException e) {
            return false;
        }
    }

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        // 设置状态栏颜色为与前端标题一致的绿色
        getWindow().setStatusBarColor(Color.parseColor("#10b981"));

        // 请求存储权限（Android 13 以下需要）
        requestStoragePermission();

        // 根容器
        rootLayout = new FrameLayout(this);

        // 普通状态栏模式
        webView = new WebView(this);
        webView.setOverScrollMode(View.OVER_SCROLL_ALWAYS);
        WebSettings settings = webView.getSettings();
        settings.setJavaScriptEnabled(true);
        settings.setDomStorageEnabled(true);
        settings.setAllowFileAccess(true);
        settings.setAllowContentAccess(true);
        settings.setJavaScriptEnabled(true);
        settings.setMediaPlaybackRequiresUserGesture(false);
        settings.setDatabaseEnabled(true);
        
        // 启用WebView开发者工具
        if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.KITKAT) {
            android.util.Log.d("MainActivity", "启用WebView开发者工具...");
            WebView.setWebContentsDebuggingEnabled(true);
        }

        // Cookie 持久化策略
        CookieManager cm = CookieManager.getInstance();
        cm.setAcceptCookie(true);
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.LOLLIPOP) {
            CookieManager.getInstance().setAcceptThirdPartyCookies(webView, true);
        }

        // 加载动画
        loadingBar = new ProgressBar(this, null, android.R.attr.progressBarStyleLarge);
        FrameLayout.LayoutParams lp = new FrameLayout.LayoutParams(ViewGroup.LayoutParams.WRAP_CONTENT, ViewGroup.LayoutParams.WRAP_CONTENT);
        lp.gravity = Gravity.BOTTOM | Gravity.CENTER_HORIZONTAL;
        lp.bottomMargin = (int)(8 * getResources().getDisplayMetrics().density);
        loadingBar.setVisibility(View.GONE);

        rootLayout.addView(webView, new FrameLayout.LayoutParams(ViewGroup.LayoutParams.MATCH_PARENT, ViewGroup.LayoutParams.MATCH_PARENT));
        rootLayout.addView(loadingBar, lp);

        // JS 接口注册
        android.util.Log.d("MainActivity", "注册JavaScript接口...");
        webView.addJavascriptInterface(new AndroidSession(), "AndroidSession");
        webView.addJavascriptInterface(new AndroidKV(), "AndroidKV");
        
        // 添加按键事件接口
        webView.addJavascriptInterface(new Object() {
            @android.webkit.JavascriptInterface
            public void onKeyEvent(String eventType, int keyCode, String key, boolean ctrlKey, boolean shiftKey, boolean altKey, boolean metaKey) {
                android.util.Log.d("AndroidKeyEvent", String.format("KeyEvent: type=%s, keyCode=%d, key=%s, ctrl=%b, shift=%b, alt=%b, meta=%b", 
                    eventType, keyCode, key, ctrlKey, shiftKey, altKey, metaKey));
            }
            
            @android.webkit.JavascriptInterface
            public void injectKeyEvent(String eventType, int keyCode, String key) {
                android.util.Log.d("AndroidKeyEvent", "Injecting key event: " + eventType + " " + keyCode + " " + key);
                // 通过 JavaScript 注入来触发按键事件
                runOnUiThread(() -> {
                    String js = String.format(
                        "if (window.dispatchEvent) {" +
                        "  const event = new KeyboardEvent('%s', {" +
                        "    key: '%s'," +
                        "    code: 'Key%s'," +
                        "    keyCode: %d," +
                        "    which: %d," +
                        "    bubbles: true," +
                        "    cancelable: true" +
                        "  });" +
                        "  window.dispatchEvent(event);" +
                        "  document.dispatchEvent(event);" +
                        "}",
                        eventType, key, key, keyCode, keyCode
                    );
                    webView.evaluateJavascript(js, null);
                });
            }
        }, "AndroidKeyEvent");

        // 添加 JS 接口：AndroidDownload（已存在）
        android.util.Log.d("MainActivity", "注册AndroidDownload接口...");
        webView.addJavascriptInterface(new Object() {
            @android.webkit.JavascriptInterface
            public void downloadBlobFile(String base64Data, String filename) {
                try {
                    byte[] data = android.util.Base64.decode(base64Data, android.util.Base64.DEFAULT);
                    java.io.File tempFile = new java.io.File(getCacheDir(), filename);
                    java.io.FileOutputStream fos = new java.io.FileOutputStream(tempFile);
                    fos.write(data);
                    fos.close();

                    DownloadManager dm = (DownloadManager) getSystemService(DOWNLOAD_SERVICE);
                    DownloadManager.Request request = new DownloadManager.Request(Uri.fromFile(tempFile));
                    request.setTitle("导出配置");
                    request.setDescription("正在保存配置文件...");
                    request.setNotificationVisibility(DownloadManager.Request.VISIBILITY_VISIBLE_NOTIFY_COMPLETED);
                    request.setDestinationInExternalPublicDir(Environment.DIRECTORY_DOWNLOADS, filename);
                    request.allowScanningByMediaScanner();
                    request.setMimeType("application/json");
                    dm.enqueue(request);
                } catch (Exception e) {
                }
            }

            @android.webkit.JavascriptInterface
            public void downloadBlobFileWithPicker(String base64Data, String filename) {
                android.util.Log.d("AndroidDownload", "downloadBlobFileWithPicker called - filename: " + filename + ", data length: " + (base64Data != null ? base64Data.length() : 0));
                pendingBase64Data = base64Data;
                pendingFilename = filename != null && !filename.isEmpty() ? filename : "video_crawler_config.json";
                Intent intent = new Intent(Intent.ACTION_CREATE_DOCUMENT);
                intent.addCategory(Intent.CATEGORY_OPENABLE);
                intent.setType("application/json");
                intent.putExtra(Intent.EXTRA_TITLE, pendingFilename);
                android.util.Log.d("AndroidDownload", "Starting file picker activity...");
                startActivityForResult(intent, REQUEST_CREATE_DOCUMENT_BLOB);
            }

            @android.webkit.JavascriptInterface
            public void downloadFile(String url, String filename) {
                try {
                    DownloadManager dm = (DownloadManager) getSystemService(DOWNLOAD_SERVICE);
                    DownloadManager.Request request = new DownloadManager.Request(Uri.parse(url));
                    request.setTitle("导出配置");
                    request.setDescription("正在下载配置文件...");
                    request.setNotificationVisibility(DownloadManager.Request.VISIBILITY_VISIBLE_NOTIFY_COMPLETED);
                    request.setDestinationInExternalPublicDir(Environment.DIRECTORY_DOWNLOADS, filename);
                    request.allowScanningByMediaScanner();
                    request.setMimeType("application/json");
                    dm.enqueue(request);
                } catch (Exception e) {
                }
            }

            @android.webkit.JavascriptInterface
            public void downloadFileWithPicker(String url, String filename) {
                pendingDownloadUrl = url;
                pendingFilename = filename != null && !filename.isEmpty() ? filename : "video_crawler_config.json";
                Intent intent = new Intent(Intent.ACTION_CREATE_DOCUMENT);
                intent.addCategory(Intent.CATEGORY_OPENABLE);
                intent.setType("application/json");
                intent.putExtra(Intent.EXTRA_TITLE, pendingFilename);
                startActivityForResult(intent, REQUEST_CREATE_DOCUMENT_URL);
            }

            @android.webkit.JavascriptInterface
            public void saveBlobData(String jsonData, String filename) {
                android.util.Log.d("AndroidDownload", "saveBlobData called - filename: " + filename + ", data length: " + (jsonData != null ? jsonData.length() : 0));
                // 将数据转换为base64并调用文件选择器
                try {
                    String base64Data = android.util.Base64.encodeToString(jsonData.getBytes("UTF-8"), android.util.Base64.DEFAULT);
                    downloadBlobFileWithPicker(base64Data, filename);
                } catch (Exception e) {
                    android.util.Log.e("AndroidDownload", "saveBlobData failed to convert to base64: " + e.getMessage(), e);
                    Toast.makeText(MainActivity.this, "保存失败: " + e.getMessage(), Toast.LENGTH_SHORT).show();
                }
            }
        }, "AndroidDownload");

        // 设置 WebViewClient（去除注入的下拉刷新脚本）
        webView.setWebViewClient(new WebViewClient() {
            @Override
            public boolean shouldOverrideUrlLoading(WebView view, String url) {
                // 非本地 127.0.0.1:{actualPort} 的链接，改为外部浏览器打开
                if (!isLocalUrl(url)) {
                    try {
                        Intent intent = new Intent(Intent.ACTION_VIEW, Uri.parse(url));
                        startActivity(intent);
                        return true;
                    } catch (Exception ignored) {}
                }
                // 下载兜底
                if (url.contains("/api/export") || url.contains("download") || url.endsWith(".json")) {
                    try {
                        String filename = "video_crawler_config.json";
                        DownloadManager dm = (DownloadManager) getSystemService(DOWNLOAD_SERVICE);
                        DownloadManager.Request request = new DownloadManager.Request(Uri.parse(url));
                        request.setTitle("导出配置");
                        request.setDescription("正在下载配置文件...");
                        request.setNotificationVisibility(DownloadManager.Request.VISIBILITY_VISIBLE_NOTIFY_COMPLETED);
                        request.setDestinationInExternalPublicDir(Environment.DIRECTORY_DOWNLOADS, filename);
                        request.allowScanningByMediaScanner();
                        request.setMimeType("application/json");
                        dm.enqueue(request);
                        return true;
                    } catch (Exception e) { }
                }
                return false;
            }

            @Override
            public void onPageFinished(WebView view, String url) {
                super.onPageFinished(view, url);
                showLoading(false);
                try { CookieManager.getInstance().flush(); } catch (Exception ignored) {}
                
                android.util.Log.d("MainActivity", "页面加载完成: " + url);
                
                // 页面加载完成后，注入一些必要的 JavaScript
                String js = "console.log('页面加载完成，JavaScript 环境已就绪');";
                view.evaluateJavascript(js, null);
                
                // 测试 JavaScript 注入是否正常工作
                String testJs = "console.log('[ANDROID_INJECT] 页面加载完成，测试 JavaScript 注入');";
                view.evaluateJavascript(testJs, null);
                
                // 添加一个全局的按键事件测试函数
                String testFunctionJs = 
                    "window.testAndroidKeyInjection = function() {" +
                    "  console.log('[ANDROID_INJECT] 测试函数被调用');" +
                    "  return 'JavaScript 注入测试成功';" +
                    "};" +
                    "console.log('[ANDROID_INJECT] 测试函数已注册');";
                view.evaluateJavascript(testFunctionJs, null);
            }
        });

        // 全屏视频：自定义容器（保持）
        final FrameLayout decor = (FrameLayout) getWindow().getDecorView();
        fullScreenContainer = new FrameLayout(this);
        fullScreenContainer.setBackgroundColor(0xFF000000);
        fullScreenContainer.setVisibility(View.GONE);
        decor.addView(fullScreenContainer, new FrameLayout.LayoutParams(ViewGroup.LayoutParams.MATCH_PARENT, ViewGroup.LayoutParams.MATCH_PARENT));

        webView.setWebChromeClient(new WebChromeClient() {
            @Override
            public boolean onConsoleMessage(android.webkit.ConsoleMessage consoleMessage) {
                // 将WebView的console信息打印到Android日志
                String message = String.format("WebView Console [%s:%d] %s: %s", 
                    consoleMessage.sourceId(), 
                    consoleMessage.lineNumber(), 
                    consoleMessage.messageLevel().name(), 
                    consoleMessage.message());
                
                switch (consoleMessage.messageLevel()) {
                    case ERROR:
                        android.util.Log.e("WebView", message);
                        break;
                    case WARNING:
                        android.util.Log.w("WebView", message);
                        break;
                    case DEBUG:
                        android.util.Log.d("WebView", message);
                        break;
                    default:
                        android.util.Log.i("WebView", message);
                        break;
                }
                return true; // 返回true表示我们已经处理了这个消息
            }
            
            @Override
            public boolean onShowFileChooser(WebView webView, ValueCallback<Uri[]> filePathCallback, FileChooserParams fileChooserParams) {
                MainActivity.this.filePathCallback = filePathCallback;
                Intent intent = new Intent(Intent.ACTION_OPEN_DOCUMENT);
                intent.addCategory(Intent.CATEGORY_OPENABLE);
                intent.setType("application/json");
                intent.putExtra(Intent.EXTRA_ALLOW_MULTIPLE, false);
                startActivityForResult(Intent.createChooser(intent, "选择配置文件"), REQUEST_FILE_CHOOSER);
                return true;
            }

            @Override
            public void onShowCustomView(View view, CustomViewCallback callback) {
                if (customView != null) {
                    callback.onCustomViewHidden();
                    return;
                }
                customView = view;
                customViewCallback = callback;
                fullScreenContainer.addView(view, new FrameLayout.LayoutParams(ViewGroup.LayoutParams.MATCH_PARENT, ViewGroup.LayoutParams.MATCH_PARENT));
                fullScreenContainer.setVisibility(View.VISIBLE);
                webView.setVisibility(View.GONE);
                setRequestedOrientation(ActivityInfo.SCREEN_ORIENTATION_SENSOR_LANDSCAPE);
                enterImmersive();
            }

            @Override
            public void onHideCustomView() {
                if (customView == null) return;
                fullScreenContainer.removeView(customView);
                customView = null;
                fullScreenContainer.setVisibility(View.GONE);
                webView.setVisibility(View.VISIBLE);
                if (customViewCallback != null) customViewCallback.onCustomViewHidden();
                setRequestedOrientation(ActivityInfo.SCREEN_ORIENTATION_UNSPECIFIED);
                exitImmersive();
            }
        });

        // 文件下载（导出配置）兜底（保留）
        webView.setDownloadListener(new DownloadListener() {
            @Override
            public void onDownloadStart(String url, String userAgent, String contentDisposition, String mimeType, long contentLength) {
                try {
                    String filename = "video_crawler_config.json";
                    if (contentDisposition != null && contentDisposition.contains("filename=")) {
                        String[] parts = contentDisposition.split("filename=");
                        if (parts.length > 1) {
                            filename = parts[1].replace("\"", "").trim();
                        }
                    }
                    DownloadManager dm = (DownloadManager) getSystemService(DOWNLOAD_SERVICE);
                    DownloadManager.Request request = new DownloadManager.Request(Uri.parse(url));
                    request.setTitle("导出配置");
                    request.setDescription("正在下载配置文件...");
                    request.setNotificationVisibility(DownloadManager.Request.VISIBILITY_VISIBLE_NOTIFY_COMPLETED);
                    request.setDestinationInExternalPublicDir(Environment.DIRECTORY_DOWNLOADS, filename);
                    request.allowScanningByMediaScanner();
                    request.setMimeType("application/json");
                    dm.enqueue(request);
                } catch (Exception e) { }
            }
        });

        setContentView(rootLayout);
        ensureServerRunningAndLoad();
        addFloatingTool();
    }

    private void showLoading(boolean show) {
        if (loadingBar != null) {
            loadingBar.setVisibility(show ? View.VISIBLE : View.GONE);
        }
    }

    private void loadUrlForPort(int port) {
        String expected = "http://127.0.0.1:" + port + "/";
        String current = webView.getUrl();
        if (current == null || !current.startsWith(expected)) {
            try { webView.stopLoading(); } catch (Exception ignored) {}
            webView.loadUrl(expected);
            try { webView.clearHistory(); } catch (Exception ignored) {}
        }
    }

    // 串行化的探活与启动：其他调用等待当前执行完成后再继续
    private void ensureServerRunningAndLoad() {
        synchronized (startLock) {
            if (isStarting) {
                // 追加一个完成后的动作：加载到最新端口
                startWaiters.add(() -> loadUrlForPort(actualPort));
                return;
            }
            isStarting = true;
        }

        new Thread(() -> {
            try {
                int last = getSharedPreferences(PREFS_NAME, MODE_PRIVATE).getInt(PREF_KEY_LAST_PORT, 0);
                int candidate = actualPort > 0 ? actualPort : last;
                boolean alive = candidate > 0 && isPortAlive(candidate);
                if (!alive) {
                    int p = startServer(getFilesDir().getAbsolutePath() + "/configs", 0);
                    if (p == 0) {
                        runOnUiThread(() -> {
                            Toast.makeText(this, "服务启动失败，请重启应用", Toast.LENGTH_LONG).show();
                            new Handler(Looper.getMainLooper()).postDelayed(() -> {
                                finishAndRemoveTask();
                            }, 3000);
                        });
                        return;
                    }
                    candidate = p;
                    actualPort = candidate;
                    saveLastPort(candidate);
                    int finalPort = candidate;
                    runOnUiThread(() -> Toast.makeText(this, "服务已启动: 127.0.0.1:" + finalPort, Toast.LENGTH_SHORT).show());
                } else {
                    actualPort = candidate;
                }

                int finalPort = candidate;
                runOnUiThread(() -> loadUrlForPort(finalPort));
            } finally {
                List<Runnable> waitersCopy;
                synchronized (startLock) {
                    waitersCopy = new ArrayList<>(startWaiters);
                    startWaiters.clear();
                    isStarting = false;
                }
                if (!waitersCopy.isEmpty()) {
                    runOnUiThread(() -> {
                        for (Runnable r : waitersCopy) {
                            try { r.run(); } catch (Exception ignored) {}
                        }
                    });
                }
            }
        }).start();
    }

    private boolean isPortAlive(int port) {
        try {
            URL url = new URL("http://127.0.0.1:" + port + "/health");
            HttpURLConnection conn = (HttpURLConnection) url.openConnection();
            conn.setConnectTimeout(500);
            conn.setReadTimeout(700);
            conn.connect();
            int code = conn.getResponseCode();
            conn.disconnect();
            if (code >= 200 && code < 500) return true;
        } catch (Exception ignored) { }
        try {
            URL url = new URL("http://127.0.0.1:" + port + "/");
            HttpURLConnection conn = (HttpURLConnection) url.openConnection();
            conn.setConnectTimeout(500);
            conn.setReadTimeout(700);
            conn.connect();
            int code = conn.getResponseCode();
            conn.disconnect();
            return code >= 200 && code < 500;
        } catch (Exception e) {
            return false;
        }
    }

    private void saveLastPort(int port) {
        getSharedPreferences(PREFS_NAME, MODE_PRIVATE)
            .edit()
            .putInt(PREF_KEY_LAST_PORT, port)
            .apply();
    }

    @Override
    protected void onResume() {
        super.onResume();
        ensureServerRunningAndLoad();
    }

    private int getStatusBarHeight() { return 0; }
    private int dpToPx(int dp) { return dp; }

    private boolean isRootRoute(String url) {
        try {
            if (url == null || url.isEmpty()) return true;
            URI uri = new URI(url);
            String path = uri.getPath();
            String frag = uri.getFragment(); // for hash router
            boolean pathIsRoot = path == null || path.isEmpty() || "/".equals(path);
            boolean hashIsRoot = (frag == null || frag.isEmpty() || "/".equals(frag));
            return pathIsRoot && hashIsRoot;
        } catch (Exception e) {
            return false;
        }
    }

    @Override
    public void onBackPressed() {
        if (customView != null) {
            WebChromeClient wc = (WebChromeClient) webView.getWebChromeClient();
            try { wc.onHideCustomView(); } catch (Exception ignored) {}
            return;
        }
        if (webView != null) {
            String currentUrl = webView.getUrl();
            if (!isRootRoute(currentUrl) && webView.canGoBack()) {
                webView.goBack();
                return;
            }
            moveTaskToBack(true);
            return;
        }
        super.onBackPressed();
    }

    @Override
    protected void onActivityResult(int requestCode, int resultCode, Intent data) {
        super.onActivityResult(requestCode, resultCode, data);
        if (requestCode == REQUEST_FILE_CHOOSER) {
            if (filePathCallback == null) return;
            Uri[] results = null;
            if (resultCode == RESULT_OK && data != null) {
                Uri uri = data.getData();
                if (uri != null) {
                    try {
                        getContentResolver().takePersistableUriPermission(uri, Intent.FLAG_GRANT_READ_URI_PERMISSION);
                    } catch (Exception ignored) {}
                    results = new Uri[]{ uri };
                }
            }
            filePathCallback.onReceiveValue(results);
            filePathCallback = null;
            return;
        }

        if (resultCode == RESULT_OK && data != null) {
            Uri uri = data.getData();
            if (uri == null) return;
            try {
                if (requestCode == REQUEST_CREATE_DOCUMENT_BLOB && pendingBase64Data != null) {
                    android.util.Log.d("AndroidDownload", "Processing REQUEST_CREATE_DOCUMENT_BLOB result");
                    byte[] bytes = android.util.Base64.decode(pendingBase64Data, android.util.Base64.DEFAULT);
                    android.util.Log.d("AndroidDownload", "Decoded base64 data, bytes length: " + bytes.length);
                    OutputStream os = getContentResolver().openOutputStream(uri, "w");
                    if (os != null) {
                        os.write(bytes);
                        os.flush();
                        os.close();
                        android.util.Log.d("AndroidDownload", "File saved successfully via picker");
                        Toast.makeText(this, "配置文件已保存", Toast.LENGTH_SHORT).show();
                    } else {
                        android.util.Log.e("AndroidDownload", "Failed to open output stream");
                    }
                    pendingBase64Data = null;
                    pendingFilename = null;
                } else if (requestCode == REQUEST_CREATE_DOCUMENT_URL && pendingDownloadUrl != null) {
                    URL u = new URL(pendingDownloadUrl);
                    HttpURLConnection conn = (HttpURLConnection) u.openConnection();
                    conn.setConnectTimeout(10000);
                    conn.setReadTimeout(15000);
                    conn.connect();
                    java.io.InputStream is = conn.getInputStream();
                    OutputStream os = getContentResolver().openOutputStream(uri, "w");
                    if (os != null) {
                        byte[] buf = new byte[8192];
                        int n;
                        while ((n = is.read(buf)) >= 0) {
                            os.write(buf, 0, n);
                        }
                        os.flush();
                        os.close();
                    }
                    is.close();
                    conn.disconnect();
                    Toast.makeText(this, "配置文件已保存", Toast.LENGTH_SHORT).show();
                    pendingDownloadUrl = null;
                    pendingFilename = null;
                }
            } catch (Exception e) {
                Toast.makeText(this, "保存失败: " + e.getMessage(), Toast.LENGTH_LONG).show();
            }
        }
    }

    private void requestStoragePermission() {
        if (android.os.Build.VERSION.SDK_INT < android.os.Build.VERSION_CODES.TIRAMISU) {
            if (ContextCompat.checkSelfPermission(this, Manifest.permission.READ_EXTERNAL_STORAGE) 
                    != PackageManager.PERMISSION_GRANTED) {
                ActivityCompat.requestPermissions(this, 
                    new String[]{Manifest.permission.READ_EXTERNAL_STORAGE}, 
                    REQUEST_STORAGE_PERMISSION);
            }
        }
    }

    @Override
    public void onRequestPermissionsResult(int requestCode, String[] permissions, int[] grantResults) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults);
        if (requestCode == REQUEST_STORAGE_PERMISSION) {
            if (grantResults.length > 0 && grantResults[0] == PackageManager.PERMISSION_GRANTED) {
                // 权限获取成功
            } else {
                // 权限被拒绝，但应用仍可继续运行
            }
        }
    }

    private void enterImmersive() {
        try {
            WindowCompat.setDecorFitsSystemWindows(getWindow(), false);
            WindowInsetsControllerCompat controller = new WindowInsetsControllerCompat(getWindow(), getWindow().getDecorView());
            controller.hide(WindowInsetsCompat.Type.systemBars());
            controller.setSystemBarsBehavior(WindowInsetsControllerCompat.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE);
        } catch (Exception ignored) {}
    }

    private void exitImmersive() {
        try {
            WindowInsetsControllerCompat controller = new WindowInsetsControllerCompat(getWindow(), getWindow().getDecorView());
            controller.show(WindowInsetsCompat.Type.systemBars());
            WindowCompat.setDecorFitsSystemWindows(getWindow(), true);
        } catch (Exception ignored) {}
    }

    private void goHome() {
        int port = actualPort > 0 ? actualPort : getSharedPreferences(PREFS_NAME, MODE_PRIVATE).getInt(PREF_KEY_LAST_PORT, 0);
        if (port <= 0) {
            Toast.makeText(this, "端口未就绪，稍后重试", Toast.LENGTH_SHORT).show();
            ensureServerRunningAndLoad();
            return;
        }
        final String url = "http://127.0.0.1:" + port + "/";
        try { webView.stopLoading(); } catch (Exception ignored) {}
        webView.loadUrl(url);
    }

    private void addFloatingTool() {
        final int sizeDp = 52;
        final float density = getResources().getDisplayMetrics().density;
        final int sizePx = (int) (sizeDp * density);
        final int marginPx = (int) (8 * density);
        final int edgeThresholdPx = (int) (32 * density);
        final int hidePortionPx = (int) (sizePx / 3f);
        final int touchSlop = Math.max(1, ViewConfiguration.get(this).getScaledTouchSlop() / 2);
        final Handler handler = new Handler(Looper.getMainLooper());

        FrameLayout bubbleContainer = new FrameLayout(this);
        FrameLayout.LayoutParams blp = new FrameLayout.LayoutParams(sizePx, sizePx);
        blp.gravity = Gravity.CENTER_VERTICAL | Gravity.END;
        blp.rightMargin = marginPx;
        bubbleContainer.setLayoutParams(blp);
        bubbleContainer.setAlpha(0.6f);
        bubbleContainer.setClickable(true);
        bubbleContainer.setFocusable(true);

        GradientDrawable bg = new GradientDrawable();
        bg.setColor(Color.parseColor("#1F000000"));
        bg.setCornerRadius(sizePx);

        ImageView bubble = new ImageView(this);
        bubble.setLayoutParams(new FrameLayout.LayoutParams(ViewGroup.LayoutParams.MATCH_PARENT, ViewGroup.LayoutParams.MATCH_PARENT));
        bubble.setBackground(bg);
        bubble.setImageResource(android.R.drawable.ic_popup_sync);
        bubble.setColorFilter(Color.WHITE);
        bubble.setScaleType(ImageView.ScaleType.CENTER);
        bubbleContainer.addView(bubble);
        bubbleContainer.bringToFront();

        FrameLayout radialMenu = new FrameLayout(this);
        FrameLayout.LayoutParams rlp = new FrameLayout.LayoutParams(ViewGroup.LayoutParams.MATCH_PARENT, ViewGroup.LayoutParams.MATCH_PARENT);
        radialMenu.setLayoutParams(rlp);
        radialMenu.setVisibility(View.GONE);

        ImageView btnRefresh = buildCircleButton(android.R.drawable.ic_menu_rotate, sizePx);
        ImageView btnHome = buildCircleButton(android.R.drawable.ic_menu_mylocation, sizePx);
        ImageView btnExit = buildCircleButton(android.R.drawable.ic_lock_power_off, sizePx);
        radialMenu.addView(btnRefresh);
        radialMenu.addView(btnHome);
        radialMenu.addView(btnExit);

        btnRefresh.setOnClickListener(v -> { radialMenu.setVisibility(View.GONE); bubbleContainer.setAlpha(0.6f); if (webView != null) webView.reload(); });
        btnHome.setOnClickListener(v -> { radialMenu.setVisibility(View.GONE); bubbleContainer.setAlpha(0.6f); goHome(); });
        btnExit.setOnClickListener(v -> { radialMenu.setVisibility(View.GONE); bubbleContainer.setAlpha(0.6f); finishAndRemoveTask(); });

        bubble.setOnClickListener(v -> {
            boolean show = radialMenu.getVisibility() != View.VISIBLE;
            radialMenu.setVisibility(show ? View.VISIBLE : View.GONE);
            bubbleContainer.setAlpha(show ? 1.0f : 0.6f);
            if (show) positionRadialMenu(bubbleContainer, radialMenu, sizePx, marginPx);
        });

        final Runnable[] pendingSnap = new Runnable[1];
        View.OnTouchListener dragListener = new View.OnTouchListener() {
            float downX, downY, dX, dY; boolean dragging = false; int lastAction;
            @Override public boolean onTouch(View v, MotionEvent e) {
                View target = bubbleContainer;
                switch (e.getActionMasked()) {
                    case MotionEvent.ACTION_DOWN:
                        lastAction = MotionEvent.ACTION_DOWN;
                        downX = e.getRawX();
                        downY = e.getRawY();
                        dX = target.getX() - downX;
                        dY = target.getY() - downY;
                        if (pendingSnap[0] != null) { handler.removeCallbacks(pendingSnap[0]); pendingSnap[0] = null; }
                        return true;
                    case MotionEvent.ACTION_MOVE:
                        lastAction = MotionEvent.ACTION_MOVE;
                        float nx = e.getRawX() + dX;
                        float ny = e.getRawY() + dY;
                        if (!dragging) {
                            float dx = Math.abs(e.getRawX() - downX);
                            float dy = Math.abs(e.getRawY() - downY);
                            dragging = (dx > touchSlop) || (dy > touchSlop);
                        }
                        if (dragging) {
                            // 边界约束
                            float clampedX = Math.max(-target.getWidth() + marginPx, Math.min(nx, rootLayout.getWidth() - marginPx));
                            float clampedY = Math.max(marginPx, Math.min(ny, rootLayout.getHeight() - target.getHeight() - marginPx));
                            target.setX(clampedX);
                            target.setY(clampedY);
                            radialMenu.setVisibility(View.GONE);
                        }
                        return true;
                    case MotionEvent.ACTION_UP:
                        if (!dragging && lastAction != MotionEvent.ACTION_MOVE) {
                            v.performClick();
                        }
                        dragging = false;
                        View parent = (View) target.getParent();
                        if (parent != null) {
                            float leftDist = Math.abs(target.getX() - marginPx);
                            float rightDist = Math.abs(parent.getWidth() - (target.getX() + target.getWidth()) - marginPx);
                            boolean nearLeft = leftDist <= edgeThresholdPx;
                            boolean nearRight = rightDist <= edgeThresholdPx;
                            if (nearLeft || nearRight) {
                                if (pendingSnap[0] != null) handler.removeCallbacks(pendingSnap[0]);
                                pendingSnap[0] = () -> {
                                    boolean snapLeft = nearLeft && !(nearRight && rightDist < leftDist);
                                    float targetX = snapLeft ? (marginPx - hidePortionPx) : (parent.getWidth() - target.getWidth() + hidePortionPx - marginPx);
                                    float targetY = Math.max(marginPx, Math.min(target.getY(), parent.getHeight() - target.getHeight() - marginPx));
                                    target.animate().x(targetX).y(targetY).setDuration(180).withEndAction(() -> positionRadialMenu(bubbleContainer, radialMenu, sizePx, marginPx)).start();
                                };
                                handler.postDelayed(pendingSnap[0], 1000);
                            }
                        }
                        return true;
                }
                return false;
            }
        };
        bubbleContainer.setOnTouchListener(dragListener);
        bubble.setOnTouchListener(dragListener);

        rootLayout.addView(radialMenu);
        rootLayout.addView(bubbleContainer);
        bubbleContainer.bringToFront();

        rootLayout.getViewTreeObserver().addOnGlobalLayoutListener(new ViewTreeObserver.OnGlobalLayoutListener() {
            @Override public void onGlobalLayout() {
                try { rootLayout.getViewTreeObserver().removeOnGlobalLayoutListener(this); } catch (Exception ignored) {}
                float targetX = rootLayout.getWidth() - sizePx + hidePortionPx - marginPx;
                float targetY = rootLayout.getHeight() * 0.5f - sizePx * 0.5f;
                bubbleContainer.setX(targetX);
                bubbleContainer.setY(Math.max(marginPx, Math.min(targetY, rootLayout.getHeight() - sizePx - marginPx)));
                positionRadialMenu(bubbleContainer, radialMenu, sizePx, marginPx);
            }
        });
    }

    private ImageView buildCircleButton(int iconRes, int refSizePx) {
        int btnSize = (int) (refSizePx * 0.72f);
        ImageView iv = new ImageView(this);
        FrameLayout.LayoutParams lp = new FrameLayout.LayoutParams(btnSize, btnSize);
        iv.setLayoutParams(lp);
        GradientDrawable bg = new GradientDrawable();
        bg.setColor(Color.parseColor("#CC000000"));
        bg.setCornerRadius(btnSize);
        iv.setBackground(bg);
        iv.setImageResource(iconRes);
        iv.setColorFilter(Color.WHITE);
        iv.setScaleType(ImageView.ScaleType.CENTER);
        iv.setVisibility(View.VISIBLE);
        return iv;
    }

    private void positionRadialMenu(FrameLayout bubbleContainer, FrameLayout radialMenu, int sizePx, int marginPx) {
        float cx = bubbleContainer.getX() + sizePx / 2f;
        float cy = bubbleContainer.getY() + sizePx / 2f;
        int count = radialMenu.getChildCount();
        if (count == 0) return;
        float R = sizePx * 1.2f;
        boolean onLeft = cx < (rootLayout.getWidth() / 2f);
        double[] angles = onLeft ? new double[]{-60, 0, 60} : new double[]{240, 180, 120};
        for (int i = 0; i < Math.min(count, angles.length); i++) {
            View child = radialMenu.getChildAt(i);
            double angle = Math.toRadians(angles[i]);
            float x = (float)(cx + Math.cos(angle) * R) - child.getLayoutParams().width / 2f;
            float y = (float)(cy + Math.sin(angle) * R) - child.getLayoutParams().height / 2f;
            child.setX(x);
            child.setY(y);
        }
    }

    @Override
    protected void onPause() {
        super.onPause();
        try { CookieManager.getInstance().flush(); } catch (Exception ignored) {}
    }

    // 返回键处理状态
    private long lastBackPressTime = 0;
    private static final long BACK_PRESS_INTERVAL = 2000; // 2秒内按两次返回键退出
    
    // 重写按键事件处理，允许按键透传到 WebView
    @Override
    public boolean onKeyDown(int keyCode, KeyEvent event) {
        android.util.Log.d("MainActivity", "onKeyDown: keyCode=" + keyCode + ", event=" + event);
        
        // 处理返回键的特殊逻辑
        if (keyCode == KeyEvent.KEYCODE_BACK) {
            return handleBackKey();
        }
        
        // 让 WebView 优先处理其他按键事件
        if (webView != null) {
            // 将按键事件传递给 WebView
            boolean handled = webView.dispatchKeyEvent(event);
            android.util.Log.d("MainActivity", "WebView dispatchKeyEvent result: " + handled + " for keyCode: " + keyCode);
            
            if (handled) {
                android.util.Log.d("MainActivity", "WebView handled key event: " + keyCode);
                return true;
            }
            
            // 如果 WebView 没有处理，尝试通过 JavaScript 注入来触发
            String keyName = getKeyName(keyCode);
            if (keyName != null) {
                android.util.Log.d("MainActivity", "Injecting key event via JavaScript: " + keyName);
                
                // 先测试 JavaScript 是否能正常执行
                String testJs = "console.log('[ANDROID_INJECT] JavaScript 测试注入成功');";
                webView.evaluateJavascript(testJs, null);
                
                // 直接创建并分发 KeyboardEvent
                String js = String.format(
                    "try {" +
                    "  console.log('[ANDROID_INJECT] 开始注入按键事件:', {keyCode: %d, key: '%s'});" +
                    "  const event = new KeyboardEvent('keydown', {" +
                    "    key: '%s'," +
                    "    code: 'Key%s'," +
                    "    keyCode: %d," +
                    "    which: %d," +
                    "    bubbles: true," +
                    "    cancelable: true" +
                    "  });" +
                    "  console.log('[ANDROID_INJECT] 创建的事件对象:', event);" +
                    "  window.dispatchEvent(event);" +
                    "  document.dispatchEvent(event);" +
                    "  console.log('[ANDROID_INJECT] 事件已分发');" +
                    "} catch (error) {" +
                    "  console.error('[ANDROID_INJECT] 注入失败:', error);" +
                    "}",
                    keyCode, keyName, keyName, keyName, keyCode, keyCode
                );
                webView.evaluateJavascript(js, null);
            }
        }
        
        // 如果 WebView 没有处理，则使用默认处理
        return super.onKeyDown(keyCode, event);
    }
    
    // 处理返回键逻辑
    private boolean handleBackKey() {
        long currentTime = System.currentTimeMillis();
        
        // 先尝试让 WebView 处理返回键
        if (webView != null) {
            // 注入返回键事件到 WebView
            String js = String.format(
                "try {" +
                "  console.log('[ANDROID_INJECT] 开始注入返回键事件:', {keyCode: 4, key: 'Backspace'});" +
                "  const event = new KeyboardEvent('keydown', {" +
                "    key: 'Backspace'," +
                "    code: 'KeyBackspace'," +
                "    keyCode: 4," +
                "    which: 4," +
                "    bubbles: true," +
                "    cancelable: true" +
                "  });" +
                "  console.log('[ANDROID_INJECT] 创建返回键事件对象:', event);" +
                "  window.dispatchEvent(event);" +
                "  document.dispatchEvent(event);" +
                "  console.log('[ANDROID_INJECT] 返回键事件已分发');" +
                "} catch (error) {" +
                "  console.error('[ANDROID_INJECT] 返回键注入失败:', error);" +
                "}"
            );
            webView.evaluateJavascript(js, null);
            
            // 检查 WebView 是否能处理返回键（通过检查是否有历史记录）
            if (webView.canGoBack()) {
                android.util.Log.d("MainActivity", "WebView 可以返回，让 WebView 处理");
                webView.goBack();
                return true;
            }
        }
        
        // WebView 无法处理返回键，检查是否需要退出应用
        if (currentTime - lastBackPressTime < BACK_PRESS_INTERVAL) {
            // 第二次按返回键，退出应用
            android.util.Log.d("MainActivity", "第二次按返回键，退出应用");
            finish();
            return true;
        } else {
            // 第一次按返回键，显示提示
            android.util.Log.d("MainActivity", "第一次按返回键，显示退出提示");
            lastBackPressTime = currentTime;
            android.widget.Toast.makeText(this, "再按一次返回键退出应用", android.widget.Toast.LENGTH_SHORT).show();
            return true;
        }
    }

    @Override
    public boolean onKeyUp(int keyCode, KeyEvent event) {
        android.util.Log.d("MainActivity", "onKeyUp: keyCode=" + keyCode + ", event=" + event);
        
        // 返回键在 onKeyDown 中统一处理，这里跳过
        if (keyCode == KeyEvent.KEYCODE_BACK) {
            return true;
        }
        
        // 让 WebView 优先处理其他按键事件
        if (webView != null) {
            boolean handled = webView.dispatchKeyEvent(event);
            android.util.Log.d("MainActivity", "WebView dispatchKeyEvent result: " + handled + " for keyCode: " + keyCode);
            
            if (handled) {
                android.util.Log.d("MainActivity", "WebView handled key up event: " + keyCode);
                return true;
            }
            
            // 如果 WebView 没有处理，尝试通过 JavaScript 注入来触发
            String keyName = getKeyName(keyCode);
            if (keyName != null) {
                android.util.Log.d("MainActivity", "Injecting key up event via JavaScript: " + keyName);
                
                // 先测试 JavaScript 是否能正常执行
                String testJs = "console.log('[ANDROID_INJECT] JavaScript 测试注入成功 (keyup)');";
                webView.evaluateJavascript(testJs, null);
                
                // 直接创建并分发 KeyboardEvent
                String js = String.format(
                    "try {" +
                    "  console.log('[ANDROID_INJECT] 开始注入按键释放事件:', {keyCode: %d, key: '%s'});" +
                    "  const event = new KeyboardEvent('keyup', {" +
                    "    key: '%s'," +
                    "    code: 'Key%s'," +
                    "    keyCode: %d," +
                    "    which: %d," +
                    "    bubbles: true," +
                    "    cancelable: true" +
                    "  });" +
                    "  console.log('[ANDROID_INJECT] 创建的释放事件对象:', event);" +
                    "  window.dispatchEvent(event);" +
                    "  document.dispatchEvent(event);" +
                    "  console.log('[ANDROID_INJECT] 释放事件已分发');" +
                    "} catch (error) {" +
                    "  console.error('[ANDROID_INJECT] 注入释放事件失败:', error);" +
                    "}",
                    keyCode, keyName, keyName, keyName, keyCode, keyCode
                );
                webView.evaluateJavascript(js, null);
            }
        }
        
        return super.onKeyUp(keyCode, event);
    }

    // 处理系统按键（如返回键）
    @Override
    public boolean onKeyLongPress(int keyCode, KeyEvent event) {
        android.util.Log.d("MainActivity", "onKeyLongPress: keyCode=" + keyCode + ", event=" + event);
        
        if (webView != null) {
            boolean handled = webView.dispatchKeyEvent(event);
            if (handled) {
                return true;
            }
        }
        
        return super.onKeyLongPress(keyCode, event);
    }
    
    // 将 Android 按键代码转换为按键名称
    private String getKeyName(int keyCode) {
        switch (keyCode) {
            case KeyEvent.KEYCODE_BACK:
                return "Backspace";
            case KeyEvent.KEYCODE_DPAD_UP:
                return "ArrowUp";
            case KeyEvent.KEYCODE_DPAD_DOWN:
                return "ArrowDown";
            case KeyEvent.KEYCODE_DPAD_LEFT:
                return "ArrowLeft";
            case KeyEvent.KEYCODE_DPAD_RIGHT:
                return "ArrowRight";
            case KeyEvent.KEYCODE_DPAD_CENTER:
            case KeyEvent.KEYCODE_ENTER:
                return "Enter";
            case KeyEvent.KEYCODE_MENU:
                return "ContextMenu";
            case KeyEvent.KEYCODE_HOME:
                return "Home";
            case KeyEvent.KEYCODE_VOLUME_UP:
                return "VolumeUp";
            case KeyEvent.KEYCODE_VOLUME_DOWN:
                return "VolumeDown";
            case KeyEvent.KEYCODE_POWER:
                return "Power";
            case KeyEvent.KEYCODE_SEARCH:
                return "Search";
            case KeyEvent.KEYCODE_ESCAPE:
                return "Escape";
            default:
                return null;
        }
    }
}
