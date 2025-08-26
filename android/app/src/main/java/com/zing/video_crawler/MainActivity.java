package com.zing.video_crawler;

import android.os.Bundle;
import android.graphics.Color;
import android.graphics.drawable.GradientDrawable;
import android.view.View;
import android.view.ViewGroup;
import android.widget.FrameLayout;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;

import androidx.appcompat.app.AppCompatActivity;

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

        // 开启沉浸式：状态栏透明，内容延伸到状态栏区域
        getWindow().setStatusBarColor(Color.TRANSPARENT);
        getWindow().getDecorView().setSystemUiVisibility(
                View.SYSTEM_UI_FLAG_LAYOUT_FULLSCREEN | View.SYSTEM_UI_FLAG_LAYOUT_STABLE
        );

        // 创建 WebView 并为其添加顶部内边距，避免被顶部渐变遮挡
        WebView webView = new WebView(this);
        WebSettings settings = webView.getSettings();
        settings.setJavaScriptEnabled(true);
        settings.setDomStorageEnabled(true);
        webView.setWebViewClient(new WebViewClient());

        int statusBarHeight = getStatusBarHeight();
        int topOverlayHeight = statusBarHeight; // 仅占据系统状态栏高度，避免覆盖页面自有标题
        webView.setPadding(0, topOverlayHeight, 0, 0);
        webView.setClipToPadding(false);

        // 顶部渐变覆盖层（高度=状态栏+标题高度），实现“渐变状态栏”的视觉效果
        View gradientOverlay = new View(this);
        GradientDrawable gradient = new GradientDrawable(
                GradientDrawable.Orientation.TL_BR,
                new int[]{Color.parseColor("#10b981"), Color.parseColor("#059669")}
        );
        gradientOverlay.setBackground(gradient);
        FrameLayout.LayoutParams overlayLp = new FrameLayout.LayoutParams(
                ViewGroup.LayoutParams.MATCH_PARENT,
                topOverlayHeight
        );
        gradientOverlay.setLayoutParams(overlayLp);

        // 使用 FrameLayout 将 WebView 作为底层，渐变层叠在其上方
        FrameLayout root = new FrameLayout(this);
        root.addView(webView, new FrameLayout.LayoutParams(
                ViewGroup.LayoutParams.MATCH_PARENT,
                ViewGroup.LayoutParams.MATCH_PARENT
        ));
        root.addView(gradientOverlay);
        setContentView(root);

        webView.loadUrl("http://127.0.0.1:" + actualPort + "/");
    }

    private int getStatusBarHeight() {
        int result = 0;
        int resourceId = getResources().getIdentifier("status_bar_height", "dimen", "android");
        if (resourceId > 0) {
            result = getResources().getDimensionPixelSize(resourceId);
        }
        return result;
    }

    // 如需转 dp -> px 可使用该方法
    // private int dpToPx(int dp) {
    //     float density = getResources().getDisplayMetrics().density;
    //     return Math.round(dp * density);
    // }
}
