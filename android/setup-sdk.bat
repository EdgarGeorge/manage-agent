@echo off
REM Android SDK 快速安装脚本（Windows）

echo Android SDK 命令行工具安装脚本
echo ================================
echo.

REM 检查是否已安装
if defined ANDROID_HOME (
    echo 检测到已安装的 Android SDK: %ANDROID_HOME%
    set /p continue="是否继续安装？(y/n) "
    if /i not "%continue%"=="y" exit /b 0
)

REM 设置安装目录
set SDK_DIR=%USERPROFILE%\Android\sdk
if defined ANDROID_HOME set SDK_DIR=%ANDROID_HOME%

echo 安装目录: %SDK_DIR%
echo.

REM 创建目录
if not exist "%SDK_DIR%\cmdline-tools" mkdir "%SDK_DIR%\cmdline-tools"

echo 请手动下载 Android SDK 命令行工具：
echo 1. 访问: https://developer.android.com/studio#command-tools
echo 2. 下载 Windows 版本的 commandlinetools-win-*.zip
echo 3. 解压到: %SDK_DIR%\cmdline-tools\latest
echo.
echo 或者使用 Chocolatey 安装：
echo choco install android-sdk
echo.
pause

REM 设置环境变量
echo.
echo 设置环境变量...
setx ANDROID_HOME "%SDK_DIR%"
setx PATH "%PATH%;%SDK_DIR%\cmdline-tools\latest\bin"
setx PATH "%PATH%;%SDK_DIR%\platform-tools"

echo.
echo ================================
echo 环境变量已设置（需要重新打开终端生效）
echo.
echo 然后运行以下命令安装组件：
echo sdkmanager "platform-tools" "platforms;android-34" "build-tools;34.0.0"
echo.
pause
