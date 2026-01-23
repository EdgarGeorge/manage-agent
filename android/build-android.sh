#!/bin/bash
# Android 构建脚本

set -e

echo "开始构建 Android APK..."

# 1. 编译 Go 后端为 Android ARM64
echo "编译 Go 后端..."
cd "$(dirname "$0")/.."
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o android/app/src/main/assets/manage-agent bin/main.go

# 2. 复制前端资源
echo "复制前端资源..."
mkdir -p android/app/src/main/assets/html
cp -r bin/html/* android/app/src/main/assets/html/

# 3. 复制配置文件（如果需要）
if [ -f bin/config.json ]; then
    cp bin/config.json android/app/src/main/assets/
fi

# 4. 检查 local.properties
if [ ! -f "android/local.properties" ]; then
    echo ""
    echo "警告: 未找到 local.properties 文件"
    echo "请创建 android/local.properties 并设置 Android SDK 路径，例如："
    echo "sdk.dir=/home/username/Android/sdk"
    echo ""
    echo "或者设置 ANDROID_HOME 环境变量"
    echo ""
    exit 1
fi

# 5. 构建 APK
echo "构建 Android APK..."
cd android
chmod +x gradlew
./gradlew assembleRelease

echo ""
echo "构建完成！APK 位置："
echo "android/app/build/outputs/apk/release/app-release.apk"
