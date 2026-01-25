@echo off
REM Windows build script
REM 使用方法：可以在项目根目录或 scripts 目录运行此脚本

REM 如果当前在 scripts 目录，切换到项目根目录
if exist "build.bat" (
    cd ..
)

REM 检查是否在项目根目录
if not exist "go.mod" (
    echo 错误：请在项目根目录运行此脚本，或从 scripts 目录运行
    pause
    exit /b 1
)
if not exist "bin" (
    echo 错误：请在项目根目录运行此脚本，或从 scripts 目录运行
    pause
    exit /b 1
)

echo Building manage-agent...

REM Set build parameters
set BUILD_TIME=%date:~0,4%%date:~5,2%%date:~8,2%
set VERSION=1.0.0

REM Disable CGO for pure Go SQLite driver
set CGO_ENABLED=0

REM Create output directory
if not exist "dist" mkdir dist

REM Build Windows 64-bit version
echo Building Windows 64-bit...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w -X main.Version=%VERSION% -X main.BuildTime=%BUILD_TIME%" -o dist\manage-agent-windows-amd64.exe bin\main.go

if %errorlevel% neq 0 (
    echo.
    echo Build failed.
    goto end
)

echo.
echo Build complete! Executables are in the dist directory.

:end
echo.
pause
