#!/bin/bash
# Linux/Mac 构建脚本
# 使用方法：在项目根目录运行 ./scripts/build.sh

# 检查是否在项目根目录
if [ ! -f "go.mod" ] || [ ! -d "bin" ]; then
    echo "错误：请在项目根目录运行此脚本"
    exit 1
fi

echo "正在构建 manage-agent..."

# 设置构建参数
BUILD_TIME=$(date +%Y%m%d)
VERSION="1.0.0"

# 禁用 CGO，使用纯 Go SQLite 驱动
export CGO_ENABLED=0

# 创建输出目录
mkdir -p dist

# 构建 Windows 64位版本
echo "构建 Windows 64位版本..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME" -o dist/manage-agent-windows-amd64.exe bin/main.go

# 构建 Linux 64位版本
echo "构建 Linux 64位版本..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME" -o dist/manage-agent-linux-amd64 bin/main.go

# 构建 Linux ARM64版本（用于 Android Termux）
echo "构建 Linux ARM64版本..."
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME" -o dist/manage-agent-linux-arm64 bin/main.go

# 构建 MacOS 64位版本
echo "构建 MacOS 64位版本..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME" -o dist/manage-agent-darwin-amd64 bin/main.go

# 构建 MacOS ARM64版本（Apple Silicon）
echo "构建 MacOS ARM64版本..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME" -o dist/manage-agent-darwin-arm64 bin/main.go

echo ""
echo "构建完成！可执行文件位于 dist 目录"
echo ""
