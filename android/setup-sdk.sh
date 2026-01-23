#!/bin/bash
# Android SDK 快速安装脚本（Linux/Mac）

set -e

echo "Android SDK 命令行工具安装脚本"
echo "================================"
echo ""

# 检查是否已安装
if [ -n "$ANDROID_HOME" ] && [ -d "$ANDROID_HOME" ]; then
    echo "检测到已安装的 Android SDK: $ANDROID_HOME"
    read -p "是否继续安装？(y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 0
    fi
fi

# 设置安装目录
SDK_DIR="${ANDROID_HOME:-$HOME/Android/sdk}"
echo "安装目录: $SDK_DIR"

# 创建目录
mkdir -p "$SDK_DIR/cmdline-tools"

# 下载命令行工具
echo ""
echo "正在下载 Android SDK 命令行工具..."
cd /tmp
LATEST_VERSION=$(curl -s https://developer.android.com/studio#command-tools | grep -oP 'commandlinetools-linux-\K[0-9]+' | head -1)

if [ -z "$LATEST_VERSION" ]; then
    echo "无法获取最新版本号，使用固定版本"
    LATEST_VERSION="11076708"
fi

CMD_TOOLS_URL="https://dl.google.com/android/repository/commandlinetools-linux-${LATEST_VERSION}_latest.zip"
echo "下载地址: $CMD_TOOLS_URL"

if command -v wget &> /dev/null; then
    wget "$CMD_TOOLS_URL" -O cmdline-tools.zip
elif command -v curl &> /dev/null; then
    curl -L "$CMD_TOOLS_URL" -o cmdline-tools.zip
else
    echo "错误: 需要 wget 或 curl"
    exit 1
fi

# 解压
echo "解压中..."
unzip -q cmdline-tools.zip -d "$SDK_DIR/cmdline-tools"
mv "$SDK_DIR/cmdline-tools/cmdline-tools" "$SDK_DIR/cmdline-tools/latest"
rm cmdline-tools.zip

# 设置环境变量
echo ""
echo "设置环境变量..."
SHELL_RC="$HOME/.bashrc"
if [ -n "$ZSH_VERSION" ]; then
    SHELL_RC="$HOME/.zshrc"
fi

if ! grep -q "ANDROID_HOME" "$SHELL_RC"; then
    echo "" >> "$SHELL_RC"
    echo "# Android SDK" >> "$SHELL_RC"
    echo "export ANDROID_HOME=$SDK_DIR" >> "$SHELL_RC"
    echo "export PATH=\$PATH:\$ANDROID_HOME/cmdline-tools/latest/bin" >> "$SHELL_RC"
    echo "export PATH=\$PATH:\$ANDROID_HOME/platform-tools" >> "$SHELL_RC"
    echo "已添加到 $SHELL_RC"
fi

export ANDROID_HOME="$SDK_DIR"
export PATH="$PATH:$ANDROID_HOME/cmdline-tools/latest/bin"
export PATH="$PATH:$ANDROID_HOME/platform-tools"

# 接受许可证
echo ""
echo "接受 Android SDK 许可证..."
yes | sdkmanager --licenses > /dev/null 2>&1 || true

# 安装必要的组件
echo ""
echo "安装 Android SDK 组件..."
sdkmanager "platform-tools" "platforms;android-34" "build-tools;34.0.0"

echo ""
echo "================================"
echo "安装完成！"
echo ""
echo "请运行以下命令使环境变量生效："
echo "source $SHELL_RC"
echo ""
echo "或者重新打开终端"
echo ""
echo "验证安装："
echo "sdkmanager --version"
