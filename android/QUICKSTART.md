# Android APK 快速打包指南

## 一键构建（推荐）

### Windows:
```bash
cd c:\workplace\manage-agent
.\android\build-android.bat
```

### Linux/Mac:
```bash
cd /path/to/manage-agent
chmod +x android/build-android.sh
./android/build-android.sh
```

## 手动构建步骤

### 1. 编译 Go 后端

```bash
# Windows PowerShell
$env:GOOS="linux"; $env:GOARCH="arm64"; $env:CGO_ENABLED="0"
go build -ldflags="-s -w" -o android\app\src\main\assets\manage-agent bin\main.go

# Linux/Mac
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o android/app/src/main/assets/manage-agent bin/main.go
```

### 2. 复制前端文件

```bash
# Windows
xcopy /E /I /Y bin\html\* android\app\src\main\assets\html\

# Linux/Mac  
mkdir -p android/app/src/main/assets/html
cp -r bin/html/* android/app/src/main/assets/html/
```

### 3. 使用 Android Studio 构建 APK

1. 打开 Android Studio
2. `File` → `Open` → 选择 `android` 目录
3. 等待 Gradle 同步
4. `Build` → `Build Bundle(s) / APK(s)` → `Build APK(s)`
5. APK 位置：`android/app/build/outputs/apk/release/app-release.apk`

## 安装测试

```bash
adb install android/app/build/outputs/apk/release/app-release.apk
```

## 工作原理

- Go 后端编译为 Android ARM64 可执行文件
- Android 应用启动本地 HTTP 服务（127.0.0.1:8090）
- WebView 加载前端页面
- **前端代码无需任何修改！**

## 详细文档

查看 `ANDROID_BUILD.md` 获取完整文档。
