package com.zing.video_crawler;

import android.os.Bundle;
import android.widget.Button;
import android.widget.ScrollView;
import android.widget.TextView;
import android.widget.LinearLayout;

import androidx.appcompat.app.AppCompatActivity;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URL;
import java.util.concurrent.Executors;

public class MainActivity extends AppCompatActivity {
    static {
        System.loadLibrary("go_video_crawler_jni");
    }

    private native int startServer(String baseDir, int port);

    private int actualPort = 0;
    private TextView logView;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        logView = new TextView(this);
        logView.setText("Ready\n");

        Button btnStart = new Button(this);
        btnStart.setText("启动服务并请求");
        btnStart.setOnClickListener(v -> {
            int p = startServer(getFilesDir().getAbsolutePath() + "/configs", 10086);
            actualPort = p > 0 ? p : 8089;
            appendLog("服务端口: " + actualPort);
            fetch("http://127.0.0.1:" + actualPort + "/health");
            fetch("http://127.0.0.1:" + actualPort + "/api");
            fetch("http://127.0.0.1:" + actualPort + "/");
        });

        LinearLayout layout = new LinearLayout(this);
        layout.setOrientation(LinearLayout.VERTICAL);
        int pad = (int) (16 * getResources().getDisplayMetrics().density);
        layout.setPadding(pad, pad, pad, pad);
        layout.addView(btnStart);

        ScrollView scroll = new ScrollView(this);
        scroll.addView(logView);
        layout.addView(scroll);

        setContentView(layout);
    }

    private void fetch(String url) {
        Executors.newSingleThreadExecutor().execute(() -> {
            try {
                HttpURLConnection conn = (HttpURLConnection) new URL(url).openConnection();
                conn.setConnectTimeout(3000);
                conn.setReadTimeout(3000);
                BufferedReader reader = new BufferedReader(new InputStreamReader(conn.getInputStream()));
                StringBuilder sb = new StringBuilder();
                String line;
                while ((line = reader.readLine()) != null) sb.append(line);
                reader.close();
                String result = sb.toString();
                runOnUiThread(() -> appendLog(url + " -> " + result));
            } catch (Exception e) {
                runOnUiThread(() -> appendLog(url + " -> Error: " + e.getMessage()));
            }
        });
    }

    private void appendLog(String s) {
        logView.append(s + "\n");
    }
}
