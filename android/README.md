# Android APK 打包方案

## 📱 方案概述

将 Go Web 应用打包成 Android APK，采用 **WebView + 本地 Go 服务** 架构：

- ✅ Go 后端编译为 Android ARM64 可执行文件
- ✅ Android 应用内启动本地 HTTP 服务（127.0.0.1:8090）
- ✅ WebView 加载前端 HTML 页面
- ✅ **前端代码完全复用，无需修改**

## 🚀 快速开始

### 前置要求

**必须安装（两种方式都需要）：**
- ✅ **Java SE (JDK) 11+** - 必须（Android SDK 不包含 Java）
- ✅ **Android SDK** - 必须（最小化安装约 200MB）

**方式 1：使用 Android Studio（图形界面）**
- Android Studio（完整 IDE，~1GB+）

**方式 2：使用命令行工具（推荐，无需 Android Studio）**
- Android SDK 命令行工具（最小化安装约 200MB）
- JDK 11+（需单独安装）

📖 **详细说明**：查看 `REQUIREMENTS.md` 了解依赖关系

### 方式 1：使用 Android Studio

1. **安装 Android Studio**
   - 下载：https://developer.android.com/studio
   - 安装并打开

2. **一键构建**
   ```bash
   cd c:\workplace\manage-agent
   .\android\build-android.bat
   ```

3. **打开项目**
   - Android Studio → `File` → `Open` → 选择 `android` 目录
   - 等待 Gradle 同步
   - `Build` → `Build Bundle(s) / APK(s)` → `Build APK(s)`

### 方式 2：使用命令行工具（无需 Android Studio）⭐

1. **安装 Android SDK 命令行工具**
   
   **Windows:**
   - 下载：https://developer.android.com/studio#command-tools
   - 解压到 `C:\Android\sdk\cmdline-tools\latest`
   - 设置环境变量 `ANDROID_HOME=C:\Android\sdk`
   
   **Linux/Mac:**
   ```bash
   chmod +x android/setup-sdk.sh
   ./android/setup-sdk.sh
   ```

2. **一键构建**
   ```bash
   # Windows
   .\android\build-android.bat
   
   # Linux/Mac
   ./android/build-android.sh
   ```

3. **安装 APK**
   ```bash
   adb install android/app/build/outputs/apk/release/app-release.apk
   ```

📖 **详细说明**：查看 `BUILD_WITHOUT_STUDIO.md`

## 📁 项目结构

```
android/
├── app/
│   ├── src/main/
│   │   ├── assets/              # 打包到 APK 的资源
│   │   │   ├── manage-agent     # Go 可执行文件（需编译）
│   │   │   └── html/            # 前端文件（需复制）
│   │   ├── java/com/manageagent/
│   │   │   └── MainActivity.java # 主 Activity
│   │   └── res/                 # Android 资源
│   └── build.gradle
├── build.gradle
├── settings.gradle
└── build-android.bat/sh         # 构建脚本
```

## 🔧 工作原理

1. **应用启动**：
   - 将 assets 中的 Go 可执行文件复制到应用数据目录
   - 启动 Go 进程，运行本地 HTTP 服务

2. **前端加载**：
   - WebView 加载 `http://127.0.0.1:8090`
   - 前端代码完全复用

3. **数据存储**：
   - SQLite 数据库存储在应用数据目录
   - 通过环境变量 `ANDROID_DATA_DIR` 传递给 Go 程序

## 📝 详细文档

- **快速开始**：查看 `QUICKSTART.md`
- **完整指南**：查看 `ANDROID_BUILD.md`
- **常见问题**：查看 `ANDROID_BUILD.md` 中的 FAQ 部分

## ⚠️ 注意事项

1. **端口**：Go 服务运行在 `127.0.0.1:8090`（仅本地访问）
2. **权限**：应用需要网络权限（虽然只访问本地服务）
3. **数据目录**：数据库和日志存储在应用数据目录
4. **前端资源**：需要手动复制 `bin/html` 到 `android/app/src/main/assets/html/`

## 🐛 调试

### 查看 Go 服务日志
```bash
adb logcat | grep ManageAgent
```

### 调试 WebView
在 Chrome 中打开 `chrome://inspect` 可以调试 WebView

### 检查应用文件
```bash
adb shell
run-as com.manageagent
cd files
ls -la
```

## 📦 构建产物

构建完成后，APK 位于：
```
android/app/build/outputs/apk/release/app-release.apk
```

## 🎯 下一步

- ✅ 测试所有功能
- ✅ 添加应用图标
- ✅ 配置应用签名
- ✅ 优化启动速度
- ✅ 发布到应用商店（可选）
