# 打包构建说明

## 快速开始

### Windows 系统

直接运行构建脚本：

```bash
build.bat
```

或者手动构建：

```bash
# 64位版本
go build -ldflags="-s -w" -o manage-agent.exe bin/main.go

# 32位版本
set GOARCH=386
go build -ldflags="-s -w" -o manage-agent-32.exe bin/main.go
```

### Linux/Mac 系统

```bash
chmod +x build.sh
./build.sh
```

或者手动构建：

```bash
# Windows 64位
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o manage-agent-windows-amd64.exe bin/main.go

# Linux 64位
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o manage-agent-linux-amd64 bin/main.go

# MacOS 64位
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o manage-agent-darwin-amd64 bin/main.go
```

## 构建参数说明

### `-ldflags` 参数

- `-s`: 省略符号表和调试信息
- `-w`: 省略 DWARF 符号表
- 这两个参数可以显著减小可执行文件大小（通常减少 30-50%）

### 交叉编译

Go 支持交叉编译，可以在任何平台上构建其他平台的可执行文件：

```bash
# 设置目标操作系统和架构
GOOS=目标操作系统  GOARCH=目标架构  go build -o 输出文件名 源文件

# 常见平台组合：
# Windows 64位: GOOS=windows GOARCH=amd64
# Windows 32位: GOOS=windows GOARCH=386
# Linux 64位:   GOOS=linux GOARCH=amd64
# Linux ARM64:  GOOS=linux GOARCH=arm64
# MacOS 64位:   GOOS=darwin GOARCH=amd64
# MacOS ARM64:  GOOS=darwin GOARCH=arm64
```

## 打包后的文件结构

构建完成后，你会得到一个**单个可执行文件**，包含：

- ✅ 所有 Go 代码编译后的二进制
- ✅ 前端资源（HTML/CSS/JS）已嵌入
- ✅ SQLite 数据库驱动
- ✅ 所有依赖库

### 运行所需文件

虽然前端资源已嵌入，但运行时仍需要：

```
manage-agent.exe          # 可执行文件
config.json              # 配置文件（可选，有默认值）
data/                    # 数据库目录（程序自动创建）
  └── manage-agent.db    # SQLite 数据库文件（程序自动创建）
log/                     # 日志目录（程序自动创建）
  └── log-*.json        # 日志文件
```

## 部署步骤

### 1. 构建可执行文件

```bash
# Windows
build.bat

# Linux/Mac
./build.sh
```

### 2. 准备配置文件（可选）

如果需要自定义配置，创建 `config.json`：

```json
{
  "database": {
    "type": "sqlite",
    "sqlite": {
      "path": "./data/manage-agent.db"
    }
  }
}
```

### 3. 分发文件

只需要分发：
- `manage-agent.exe`（或对应平台的可执行文件）
- `config.json`（可选）

**不需要**：
- ❌ `html/` 目录（已嵌入）
- ❌ `go.mod`、`go.sum`
- ❌ 源代码文件

### 4. 运行

用户只需要：
1. 双击运行 `manage-agent.exe`
2. 程序会自动：
   - 创建 `data/` 目录（如果不存在）
   - 创建 SQLite 数据库文件
   - 创建数据表
   - 启动 Web 服务器
   - 自动打开浏览器

## 文件大小优化

### 当前优化

- 使用 `-ldflags="-s -w"` 减小文件大小
- 前端资源使用 `embed` 嵌入，无需外部文件

### 进一步优化（可选）

如果需要更小的文件，可以使用 UPX 压缩：

```bash
# 安装 UPX
# Windows: 下载 https://upx.github.io/
# Linux: sudo apt install upx
# Mac: brew install upx

# 压缩可执行文件（可减少 50-70% 大小）
upx --best manage-agent.exe
```

**注意**：UPX 压缩后，某些杀毒软件可能会误报，需要添加白名单。

## 常见问题

### Q: 为什么可执行文件这么大？

A: Go 编译的可执行文件包含所有依赖，通常 20-50MB 是正常的。使用 `-ldflags="-s -w"` 可以减小到 15-30MB。

### Q: 可以在没有 Go 环境的机器上运行吗？

A: 可以！Go 编译的是静态链接的可执行文件，不依赖任何外部库（除了系统库）。

### Q: 前端资源修改后需要重新编译吗？

A: 是的，因为前端资源已嵌入到可执行文件中，修改后需要重新编译。

### Q: 如何查看嵌入的前端资源？

A: 前端资源在编译时嵌入，运行时无法直接查看。如果需要修改，需要：
1. 修改 `bin/html/` 目录下的文件
2. 重新编译

### Q: 可以分离前端资源吗？

A: 可以，但需要修改代码：
1. 移除 `embed` 指令
2. 使用文件系统路径加载资源
3. 分发时需要同时包含 `html/` 目录

## 构建云端版本

云端版本（使用 MySQL）的构建方式相同：

```bash
# 构建云端服务器
go build -ldflags="-s -w" -o manage-agent-cloud cloud/main.go
```

云端版本只需要：
- `manage-agent-cloud`（可执行文件）
- `cloud-config.json`（配置文件）
