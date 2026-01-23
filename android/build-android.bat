@echo off
REM Android 构建脚本 (Windows)

echo 开始构建 Android APK...

REM 1. 编译 Go 后端为 Android ARM64
echo 编译 Go 后端...
cd /d "%~dp0\.."
set GOOS=linux
set GOARCH=arm64
set CGO_ENABLED=0
go build -ldflags="-s -w" -o android\app\src\main\assets\manage-agent bin\main.go

REM 2. 复制前端资源
echo 复制前端资源...
if not exist "android\app\src\main\assets\html" mkdir android\app\src\main\assets\html
xcopy /E /I /Y bin\html\* android\app\src\main\assets\html\

REM 3. 复制配置文件（如果需要）
if exist "bin\config.json" (
    copy /Y bin\config.json android\app\src\main\assets\
)

REM 4. 检查 local.properties
if not exist "android\local.properties" (
    echo.
    echo 警告: 未找到 local.properties 文件
    echo 请创建 android\local.properties 并设置 Android SDK 路径，例如：
    echo sdk.dir=C:\\Android\\sdk
    echo.
    echo 或者设置 ANDROID_HOME 环境变量
    echo.
    pause
    exit /b 1
)

REM 5. 构建 APK
echo 构建 Android APK...
cd android
if exist "gradlew.bat" (
    call gradlew.bat assembleRelease
) else (
    echo 错误: 未找到 gradlew.bat，请确保在 android 目录下运行
    pause
    exit /b 1
)

echo.
echo 构建完成！APK 位置：
echo android\app\build\outputs\apk\release\app-release.apk
pause
