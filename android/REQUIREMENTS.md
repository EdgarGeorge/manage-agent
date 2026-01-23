# Android 构建依赖说明

## 必须安装的组件

### 1. Java SE (JDK) - 必须 ✅

**Android SDK 不包含 Java SE**，需要单独安装。

#### 为什么需要 Java SE？
- Gradle 构建工具需要 Java 运行环境
- Android SDK 工具（如 `sdkmanager`）需要 Java

#### 安装方式

**Windows:**
- 下载：https://adoptium.net/ 或 https://www.oracle.com/java/technologies/downloads/
- 推荐版本：JDK 11 或 JDK 17（LTS 版本）
- 安装后设置环境变量 `JAVA_HOME`

**Linux/Mac:**
```bash
# Ubuntu/Debian
sudo apt install openjdk-11-jdk

# macOS (使用 Homebrew)
brew install openjdk@11

# 验证安装
java -version
javac -version
```

### 2. Android SDK - 必须 ✅

**是的，必须安装 Android SDK**，因为：
- 需要 Android SDK Platform（API Level 34）
- 需要 Android SDK Build-Tools（编译工具）
- 需要 Android SDK Platform-Tools（adb 等工具）

#### Android SDK 不包含 Java SE
- Android SDK 和 Java SE 是**分开的**
- Android SDK 需要 Java SE 作为依赖
- 必须先安装 Java SE，再安装 Android SDK

#### 最小化安装（推荐）

只需要安装以下组件（约 200MB）：

```bash
sdkmanager "platform-tools" "platforms;android-34" "build-tools;34.0.0"
```

**不包含：**
- ❌ Android Studio IDE（不需要）
- ❌ Android Emulator（不需要，除非要测试）
- ❌ 其他 API Level（不需要）

## 完整安装流程

### 步骤 1：安装 Java SE (JDK)

**Windows:**
1. 下载 JDK 11 或 17：https://adoptium.net/
2. 安装（记住安装路径，例如：`C:\Program Files\Java\jdk-11`）
3. 设置环境变量：
   ```powershell
   $env:JAVA_HOME = "C:\Program Files\Java\jdk-11"
   $env:PATH += ";$env:JAVA_HOME\bin"
   ```
4. 验证：`java -version`

**Linux/Mac:**
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install openjdk-11-jdk

# macOS
brew install openjdk@11

# 验证
java -version
```

### 步骤 2：安装 Android SDK 命令行工具

**Windows:**
1. 下载：https://developer.android.com/studio#command-tools
2. 解压到：`C:\Android\sdk\cmdline-tools\latest`
3. 设置环境变量：
   ```powershell
   $env:ANDROID_HOME = "C:\Android\sdk"
   $env:PATH += ";$env:ANDROID_HOME\cmdline-tools\latest\bin"
   $env:PATH += ";$env:ANDROID_HOME\platform-tools"
   ```

**Linux/Mac:**
```bash
# 使用提供的脚本
chmod +x android/setup-sdk.sh
./android/setup-sdk.sh

# 或手动安装
mkdir -p ~/Android/sdk/cmdline-tools
cd ~/Android/sdk/cmdline-tools
wget https://dl.google.com/android/repository/commandlinetools-linux-*.zip
unzip commandlinetools-linux-*.zip
mv cmdline-tools latest
```

### 步骤 3：安装必要的 Android SDK 组件

```bash
# 接受许可证
yes | sdkmanager --licenses

# 安装最小化组件
sdkmanager "platform-tools" "platforms;android-34" "build-tools;34.0.0"
```

### 步骤 4：配置 local.properties

创建 `android/local.properties`：

**Windows:**
```properties
sdk.dir=C:\\Android\\sdk
```

**Linux/Mac:**
```properties
sdk.dir=/home/username/Android/sdk
```

## 依赖关系图

```
构建 APK
  │
  ├─→ Gradle (构建工具)
  │     └─→ 需要 Java SE (JDK)
  │
  └─→ Android SDK
        ├─→ platform-tools (adb 等)
        ├─→ platforms;android-34 (API)
        └─→ build-tools;34.0.0 (编译工具)
              └─→ 需要 Java SE (JDK)
```

## 总结

| 组件 | 是否必须 | 是否包含 Java SE | 大小 |
|------|---------|-----------------|------|
| **Java SE (JDK)** | ✅ 必须 | ✅ 是（这就是 Java SE） | ~200MB |
| **Android SDK** | ✅ 必须 | ❌ 否（需要 Java SE） | ~200MB |
| **Android Studio** | ❌ 不需要 | ❌ 否 | ~1GB+ |

## 最小化安装总大小

- Java SE (JDK 11): ~200MB
- Android SDK (最小化): ~200MB
- **总计：约 400MB**

相比 Android Studio（1GB+），节省约 60% 的空间！

## 验证安装

```bash
# 检查 Java
java -version
javac -version

# 检查 Android SDK
sdkmanager --version
adb version

# 检查环境变量
echo $JAVA_HOME      # Linux/Mac
echo $ANDROID_HOME   # Linux/Mac

echo %JAVA_HOME%     # Windows
echo %ANDROID_HOME%  # Windows
```

## 常见问题

### Q: 可以用 JRE 代替 JDK 吗？

**A:** 不可以。Gradle 和 Android SDK 工具需要 JDK（包含编译器），JRE（只有运行时）不够。

### Q: Android SDK 包含 Java 吗？

**A:** 不包含。Android SDK 需要系统已安装 Java SE (JDK) 作为依赖。

### Q: 必须安装 Android SDK 吗？

**A:** 是的，必须安装。因为：
- 需要 Android SDK Platform（定义 Android API）
- 需要 Android SDK Build-Tools（编译 APK）
- 需要 Android SDK Platform-Tools（adb 等工具）

### Q: 可以用其他方式构建吗？

**A:** 理论上可以，但不推荐：
- ❌ 手动编译：太复杂
- ❌ 使用其他工具：兼容性问题
- ✅ 使用 Gradle + Android SDK：标准方式，最可靠

## 一键安装脚本

我们提供了安装脚本：

**Linux/Mac:**
```bash
chmod +x android/setup-sdk.sh
./android/setup-sdk.sh
```

**Windows:**
- 需要手动安装（见上面的步骤）
- 或使用 Chocolatey：`choco install android-sdk`
