# 不使用 Android Studio 构建 APK

## 方案：使用命令行工具（无需 Android Studio）

只需要 Android SDK 的命令行工具，不需要完整的 Android Studio IDE。

## 前置要求

1. **Go 1.21+**（已安装）
2. **Android SDK**（只需要命令行工具）
3. **JDK 11+**

## 安装 Android SDK 命令行工具

### Windows

1. **下载 Android SDK Command Line Tools**
   - 访问：https://developer.android.com/studio#command-tools
   - 下载 Windows 版本的 `commandlinetools-win-*.zip`

2. **解压并设置环境变量**
   ```powershell
   # 解压到某个目录，例如：C:\Android\sdk
   # 设置环境变量
   $env:ANDROID_HOME = "C:\Android\sdk"
   $env:PATH += ";$env:ANDROID_HOME\cmdline-tools\latest\bin"
   $env:PATH += ";$env:ANDROID_HOME\platform-tools"
   ```

3. **安装必要的 SDK 组件**
   ```powershell
   sdkmanager "platform-tools" "platforms;android-34" "build-tools;34.0.0"
   ```

### Linux/Mac

1. **下载并解压**
   ```bash
   cd ~
   wget https://dl.google.com/android/repository/commandlinetools-linux-*.zip
   unzip commandlinetools-linux-*.zip
   mkdir -p ~/Android/sdk/cmdline-tools
   mv cmdline-tools ~/Android/sdk/cmdline-tools/latest
   ```

2. **设置环境变量**
   ```bash
   export ANDROID_HOME=~/Android/sdk
   export PATH=$PATH:$ANDROID_HOME/cmdline-tools/latest/bin
   export PATH=$PATH:$ANDROID_HOME/platform-tools
   
   # 添加到 ~/.bashrc 或 ~/.zshrc
   echo 'export ANDROID_HOME=~/Android/sdk' >> ~/.bashrc
   echo 'export PATH=$PATH:$ANDROID_HOME/cmdline-tools/latest/bin' >> ~/.bashrc
   echo 'export PATH=$PATH:$ANDROID_HOME/platform-tools' >> ~/.bashrc
   ```

3. **安装必要的 SDK 组件**
   ```bash
   sdkmanager "platform-tools" "platforms;android-34" "build-tools;34.0.0"
   ```

## 构建步骤

### 1. 编译 Go 后端

```bash
# Windows PowerShell
$env:GOOS="linux"
$env:GOARCH="arm64"
$env:CGO_ENABLED="0"
go build -ldflags="-s -w" -o android\app\src\main\assets\manage-agent bin\main.go

# Linux/Mac
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o android/app/src/main/assets/manage-agent bin/main.go
```

### 2. 复制前端资源

```bash
# Windows
xcopy /E /I /Y bin\html\* android\app\src\main\assets\html\

# Linux/Mac
mkdir -p android/app/src/main/assets/html
cp -r bin/html/* android/app/src/main/assets/html/
```

### 3. 配置 local.properties

创建 `android/local.properties` 文件：

**Windows:**
```properties
sdk.dir=C:\\Android\\sdk
```

**Linux/Mac:**
```properties
sdk.dir=/home/username/Android/sdk
```

### 4. 使用 Gradle 构建 APK

```bash
cd android

# Windows
gradlew.bat assembleRelease

# Linux/Mac
chmod +x gradlew
./gradlew assembleRelease
```

### 5. 安装 APK

```bash
adb install app/build/outputs/apk/release/app-release.apk
```

## 一键构建脚本（已更新）

构建脚本会自动处理上述步骤：

**Windows:**
```bash
.\android\build-android.bat
```

**Linux/Mac:**
```bash
./android/build-android.sh
```

## 验证安装

```bash
# 检查 Android SDK
sdkmanager --version

# 检查 Gradle（会自动下载）
cd android
./gradlew --version
```

## 常见问题

### Q: 找不到 sdkmanager 命令

**解决**：确保 `ANDROID_HOME` 环境变量已设置，并且 `cmdline-tools/latest/bin` 在 PATH 中。

### Q: Gradle 下载失败

**解决**：检查网络连接，或手动下载 Gradle 放到 `~/.gradle/wrapper/dists/` 目录。

### Q: 构建失败，提示找不到 SDK

**解决**：创建 `android/local.properties` 文件，设置 `sdk.dir` 路径。

## 最小化安装（推荐）

如果只需要构建 APK，最小化安装包括：

1. **Android SDK Platform 34**
2. **Android SDK Build-Tools 34.0.0**
3. **Android SDK Platform-Tools**

安装命令：
```bash
sdkmanager "platform-tools" "platforms;android-34" "build-tools;34.0.0"
```

这样就不需要安装完整的 Android Studio 了！
