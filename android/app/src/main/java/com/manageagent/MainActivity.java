package com.manageagent;

import android.Manifest;
import android.content.pm.PackageManager;
import android.os.Build;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.util.Log;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;
import android.widget.Toast;

import androidx.annotation.NonNull;
import androidx.appcompat.app.AppCompatActivity;
import androidx.core.app.ActivityCompat;
import androidx.core.content.ContextCompat;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.util.concurrent.Executors;

public class MainActivity extends AppCompatActivity {
    private static final String TAG = "ManageAgent";
    private WebView webView;
    private Process goProcess;
    private Handler mainHandler;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);

        mainHandler = new Handler(Looper.getMainLooper());

        // 初始化应用
        initializeApp();
    }

    private void checkPermissions() {
        // Android 10+ 不需要 WRITE_EXTERNAL_STORAGE 权限来访问应用数据目录
        // 直接初始化应用
        initializeApp();
    }

    private void initializeApp() {
        Executors.newSingleThreadExecutor().execute(() -> {
            try {
                // 复制可执行文件到应用数据目录
                File appDir = getFilesDir();
                File binDir = new File(appDir, "bin");
                if (!binDir.exists()) {
                    binDir.mkdirs();
                }

                File executable = new File(binDir, "manage-agent");
                if (!executable.exists()) {
                    copyAsset("manage-agent", executable);
                    executable.setExecutable(true);
                }

                // 复制前端资源
                File htmlDir = new File(appDir, "html");
                if (!htmlDir.exists()) {
                    copyAssetsRecursive("html", htmlDir);
                }

                // 启动 Go 服务
                startGoServer(executable, appDir);

                // 等待服务启动
                Thread.sleep(2000);

                // 在主线程中加载 WebView
                mainHandler.post(() -> {
                    setupWebView();
                    webView.loadUrl("http://127.0.0.1:8090");
                });

            } catch (Exception e) {
                Log.e(TAG, "初始化失败", e);
                mainHandler.post(() -> {
                    Toast.makeText(this, "启动失败: " + e.getMessage(), Toast.LENGTH_LONG).show();
                });
            }
        });
    }

    private void copyAsset(String assetName, File dest) throws IOException {
        InputStream is = getAssets().open(assetName);
        OutputStream os = new FileOutputStream(dest);
        byte[] buffer = new byte[1024];
        int length;
        while ((length = is.read(buffer)) > 0) {
            os.write(buffer, 0, length);
        }
        os.flush();
        os.close();
        is.close();
    }

    private void copyAssetsRecursive(String assetPath, File destDir) throws IOException {
        String[] files = getAssets().list(assetPath);
        if (files != null && files.length > 0) {
            destDir.mkdirs();
            for (String file : files) {
                String assetFilePath = assetPath + "/" + file;
                String[] subFiles = getAssets().list(assetFilePath);
                if (subFiles != null && subFiles.length > 0) {
                    // 是目录
                    copyAssetsRecursive(assetFilePath, new File(destDir, file));
                } else {
                    // 是文件
                    File destFile = new File(destDir, file);
                    copyAsset(assetFilePath, destFile);
                }
            }
        }
    }

    private void startGoServer(File executable, File appDir) throws IOException {
        // 设置环境变量，让 Go 程序知道数据目录
        ProcessBuilder pb = new ProcessBuilder(
                executable.getAbsolutePath(),
                "-c", new File(appDir, "config.json").getAbsolutePath()
        );

        // 设置工作目录
        pb.directory(appDir);

        // 设置环境变量
        pb.environment().put("ANDROID_DATA_DIR", appDir.getAbsolutePath());

        // 重定向输出到 logcat
        pb.redirectErrorStream(true);

        goProcess = pb.start();

        // 读取输出到 logcat
        Executors.newSingleThreadExecutor().execute(() -> {
            try {
                InputStream is = goProcess.getInputStream();
                byte[] buffer = new byte[1024];
                int length;
                while ((length = is.read(buffer)) > 0) {
                    String line = new String(buffer, 0, length);
                    Log.d(TAG, line);
                }
            } catch (IOException e) {
                Log.e(TAG, "读取 Go 进程输出失败", e);
            }
        });
    }

    private void setupWebView() {
        webView = findViewById(R.id.webview);
        WebSettings settings = webView.getSettings();
        settings.setJavaScriptEnabled(true);
        settings.setDomStorageEnabled(true);
        settings.setDatabaseEnabled(true);
        settings.setAllowFileAccess(true);
        settings.setAllowContentAccess(true);
        settings.setLoadWithOverviewMode(true);
        settings.setUseWideViewPort(true);
        settings.setBuiltInZoomControls(false);
        settings.setDisplayZoomControls(false);

        webView.setWebViewClient(new WebViewClient() {
            @Override
            public boolean shouldOverrideUrlLoading(WebView view, String url) {
                return false;
            }
        });
    }

    @Override
    protected void onDestroy() {
        super.onDestroy();
        if (goProcess != null) {
            goProcess.destroy();
        }
    }

    @Override
    public void onBackPressed() {
        if (webView.canGoBack()) {
            webView.goBack();
        } else {
            super.onBackPressed();
        }
    }
}
