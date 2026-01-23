@echo off
REM Windows build script

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
