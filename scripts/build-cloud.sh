#!/bin/bash
# 云端服务器构建脚本
# 使用方法：在项目根目录运行 ./scripts/build-cloud.sh

# 检查是否在项目根目录
if [ ! -f "go.mod" ] || [ ! -d "cloud" ]; then
    echo "错误：请在项目根目录运行此脚本"
    exit 1
fi

echo "开始构建云端服务器..."

# 设置变量
CLOUD_DIR="cloud"
OUTPUT_DIR="dist"
BINARY_NAME="manage-agent-cloud"

# 创建输出目录
mkdir -p $OUTPUT_DIR

# 构建 Linux 版本（服务器常用）
echo "构建 Linux amd64 版本..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $OUTPUT_DIR/${BINARY_NAME}-linux-amd64 ./$CLOUD_DIR/main.go

# 构建 Linux arm64 版本（ARM 服务器）
echo "构建 Linux arm64 版本..."
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $OUTPUT_DIR/${BINARY_NAME}-linux-arm64 ./$CLOUD_DIR/main.go

# 构建 Windows 版本（用于 Windows 服务器）
echo "构建 Windows amd64 版本..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $OUTPUT_DIR/${BINARY_NAME}-windows-amd64.exe ./$CLOUD_DIR/main.go

echo "构建完成！"
echo "输出文件："
ls -lh $OUTPUT_DIR/${BINARY_NAME}*
