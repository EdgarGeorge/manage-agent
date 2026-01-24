@echo off
REM 云端服务器构建脚本（Windows）
REM 使用方法：在项目根目录运行 scripts\build-cloud.bat

REM 检查是否在项目根目录
if not exist "go.mod" (
    echo 错误：请在项目根目录运行此脚本
    pause
    exit /b 1
)
if not exist "cloud" (
    echo 错误：请在项目根目录运行此脚本
    pause
    exit /b 1
)

echo 开始构建云端服务器...

REM 设置变量
set CLOUD_DIR=cloud
set OUTPUT_DIR=dist
set BINARY_NAME=manage-agent-cloud

REM 创建输出目录
if not exist %OUTPUT_DIR% mkdir %OUTPUT_DIR%

REM 构建 Linux 版本（服务器常用）
echo 构建 Linux amd64 版本...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o %OUTPUT_DIR%\%BINARY_NAME%-linux-amd64.exe %CLOUD_DIR%\main.go

REM 构建 Linux arm64 版本（ARM 服务器）
echo 构建 Linux arm64 版本...
set GOOS=linux
set GOARCH=arm64
go build -ldflags="-s -w" -o %OUTPUT_DIR%\%BINARY_NAME%-linux-arm64.exe %CLOUD_DIR%\main.go

REM 构建 Windows 版本（用于 Windows 服务器）
echo 构建 Windows amd64 版本...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o %OUTPUT_DIR%\%BINARY_NAME%-windows-amd64.exe %CLOUD_DIR%\main.go

echo 构建完成！
echo 输出文件：
dir %OUTPUT_DIR%\%BINARY_NAME%*
