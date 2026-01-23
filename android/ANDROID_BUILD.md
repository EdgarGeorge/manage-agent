# Android APK 打包完整指南

## 方案概述

本方案将 Go Web 应用打包成 Android APK，采用 **WebView + 本地 Go 服务** 架构：

- ✅ **Go 后端**：编译为 Android ARM64 可执行文件，在应用内作为本地服务运行
- ✅ **前端界面**：完全复用现有 HTML/JS/CSS 代码，通过 WebView 加载
- ✅ **数据库**：SQLite 数据库存储在应用数据目录
- ✅ **零前端改动**：前端代码无需任何修改

## 前置要求

1. **Go 1.21+**（已安装）
2. **Android Studio**（用于构建 APK）
3. **Android SDK**（API Level 24+）
4. **JDK 11+**

## 快速开始

### 方法 1：使用构建脚本（推荐）

#### Windows:
```bash
cd c:\workplace\manage-agent
.\android\build-android.bat
```

#### Linux/Mac:
```bash
cd /path/to/manage-agent
chmod +x android/build-android.sh
./android/build-android.sh
```

### 方法 2：手动构建

#### 步骤 1：编译 Go 后端

```bash
# Windows PowerShell
$env:GOOS="linux"
$env:GOARCH="arm64"
$env:CGO_ENABLED="0"
go build -ldflags="-s -w" -o android\app\src\main\assets\manage-agent bin\main.go

# Linux/Mac
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o android/app/src/main/assets/manage-agent bin/main.go
```

#### 步骤 2：复制前端资源

```bash
# Windows
xcopy /E /I /Y bin\html\* android\app\src\main\assets\html\

# Linux/Mac
mkdir -p android/app/src/main/assets/html
cp -r bin/html/* android/app/src/main/assets/html/
```

#### 步骤 3：复制配置文件（可选）

```bash
# Windows
copy bin\config.json android\app\src\main\assets\

# Linux/Mac
cp bin/config.json android/app/src/main/assets/
```

#### 步骤 4：使用 Android Studio 构建

1. 打开 Android Studio
2. 选择 `Open` → 选择 `android` 目录
3. 等待 Gradle 同步完成
4. 点击 `Build` → `Build Bundle(s) / APK(s)` → `Build APK(s)`
5. 等待构建完成

#### 步骤 5：安装 APK

构建完成后，APK 位于：
```
android/app/build/outputs/apk/release/app-release.apk
```

可以通过以下方式安装：
- 使用 `adb install app-release.apk`
- 或直接传输到 Android 设备安装

## 项目结构说明

```
android/
├── app/
│   ├── src/main/
│   │   ├── assets/              # 资源文件（打包到 APK）
│   │   │   ├── manage-agent     # Go 可执行文件
│   │   │   └── html/            # 前端文件
│   │   ├── java/com/manageagent/
│   │   │   └── MainActivity.java # 主 Activity
│   │   └── res/                 # Android 资源
│   └── build.gradle
├── build.gradle
├── settings.gradle
└── build-android.bat/sh         # 构建脚本
```

## 工作原理

1. **应用启动**：
   - `MainActivity` 启动时，将 assets 中的 Go 可执行文件复制到应用数据目录
   - 启动 Go 进程，运行本地 HTTP 服务（127.0.0.1:8090）

2. **前端加载**：
   - WebView 加载 `http://127.0.0.1:8090`
   - 前端代码完全复用，无需修改

3. **数据存储**：
   - SQLite 数据库存储在应用数据目录
   - 日志文件也存储在应用数据目录

## 配置说明

### AndroidManifest.xml

- **权限**：需要网络权限（虽然只访问本地服务）
- **usesCleartextTraffic**：允许 HTTP（本地服务）

### MainActivity.java

- **端口**：Go 服务运行在 `127.0.0.1:8090`
- **数据目录**：通过环境变量 `ANDROID_DATA_DIR` 传递给 Go 程序
- **WebView 设置**：启用 JavaScript、DOM Storage 等

## 调试方法

### 1. 查看 Go 服务日志

```bash
adb logcat | grep ManageAgent
```

### 2. 查看 WebView 错误

在 Android Studio 中：
- `Run` → `Attach Debugger to Android Process`
- 选择应用进程
- 在 Chrome 中打开 `chrome://inspect` 可以调试 WebView

### 3. 检查文件

```bash
# 进入应用数据目录
adb shell
run-as com.manageagent
cd files
ls -la
```

## 常见问题

### Q1: 应用启动后白屏

**原因**：Go 服务未启动成功

**解决**：
1. 检查 logcat 日志：`adb logcat | grep ManageAgent`
2. 确认可执行文件权限：`chmod +x manage-agent`
3. 检查端口是否被占用

### Q2: 前端资源加载失败

**原因**：assets 目录结构不正确

**解决**：
1. 确认 `android/app/src/main/assets/html/` 目录存在
2. 确认所有前端文件都已复制
3. 检查 WebView 控制台错误

### Q3: 数据库无法访问

**原因**：路径权限问题

**解决**：
1. 确认应用有存储权限
2. 检查 `ANDROID_DATA_DIR` 环境变量
3. 查看 Go 程序的错误日志

### Q4: 构建失败

**原因**：Gradle 配置问题

**解决**：
1. 检查 `local.properties` 文件（Android SDK 路径）
2. 确认 Gradle 版本兼容
3. 清理重建：`./gradlew clean`

## 性能优化

1. **减小 APK 体积**：
   - 使用 `-ldflags="-s -w"` 减小 Go 二进制大小
   - 压缩前端资源（可选）

2. **启动速度**：
   - Go 服务在后台线程启动
   - 前端资源预加载

3. **内存优化**：
   - WebView 内存管理
   - Go 服务资源释放

## 发布准备

1. **签名配置**：
   - 创建 keystore
   - 配置 `app/build.gradle` 中的签名信息

2. **版本号**：
   - 更新 `versionCode` 和 `versionName`

3. **应用图标**：
   - 替换 `app/src/main/res/mipmap-*/ic_launcher.png`

4. **应用名称**：
   - 修改 `app/src/main/res/values/strings.xml`

## 下一步

- ✅ 测试所有功能
- ✅ 优化启动速度
- ✅ 添加应用图标和启动画面
- ✅ 配置应用签名
- ✅ 发布到 Google Play（可选）
