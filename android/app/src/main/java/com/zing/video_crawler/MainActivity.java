package com.zing.video_crawler;

import android.os.Bundle;
import android.graphics.Color;
import android.graphics.drawable.GradientDrawable;
import android.view.View;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;

import androidx.appcompat.app.AppCompatActivity;
import androidx.core.view.ViewCompat;
import androidx.core.view.WindowInsetsCompat;

public class MainActivity extends AppCompatActivity {
    static {
        System.loadLibrary("go_video_crawler_jni");
    }

    private native int startServer(String baseDir, int port);

    private int actualPort = 0;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        // 设置状态栏颜色为与前端标题一致的绿色
        getWindow().setStatusBarColor(Color.parseColor("#10b981"));

        int p = startServer(getFilesDir().getAbsolutePath() + "/configs", 10086);
        actualPort = p > 0 ? p : 8089;

        // 还原：使用普通状态栏，不做沉浸式与覆盖层
        WebView webView = new WebView(this);
        WebSettings settings = webView.getSettings();
        settings.setJavaScriptEnabled(true);
        settings.setDomStorageEnabled(true);
        webView.setWebViewClient(new WebViewClient());

        setContentView(webView);

        webView.loadUrl("http://127.0.0.1:" + actualPort + "/");
    }

    private int getStatusBarHeight() { return 0; }
    private int dpToPx(int dp) { return dp; }
}
