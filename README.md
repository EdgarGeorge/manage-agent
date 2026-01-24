# Manage Agent - 财务管理代理系统

一个基于 Go + Gin 的财务管理系统，支持本地 SQLite 存储和云端 MySQL 同步。

## ✨ 特性

- 🚀 **零配置本地使用**：使用 SQLite，无需安装数据库服务器
- ☁️ **云端同步**：支持多设备数据同步（可选）
- 📦 **单文件部署**：前端资源已嵌入，只需一个可执行文件
- 🔄 **自动冲突解决**：基于时间戳自动合并数据
- 💻 **跨平台**：支持 Windows、Linux、macOS、Android

## 📁 项目结构

```
manage-agent/
├── bin/                    # 本地客户端主程序
│   ├── main.go            # 客户端入口
│   ├── config.json        # 配置文件
│   └── html/              # 前端资源（已嵌入）
├── cloud/                  # 云端服务器
│   ├── main.go            # 云端服务器入口
│   └── cloud-config.json   # 云端配置
├── server/                 # 共享服务层
│   ├── model.go           # 数据模型
│   ├── database.go        # 数据库连接
│   ├── sync.go            # 同步 API
│   ├── sync_client.go     # 同步客户端
│   └── ...
├── scripts/                # 构建脚本
│   ├── build.sh           # Linux/Mac 构建
│   ├── build.bat          # Windows 构建
│   ├── build-cloud.sh     # 云端服务器构建
│   └── build-cloud.bat
├── docs/                   # 文档
│   ├── BUILD.md           # 构建说明
│   ├── DEPLOY.md          # 部署指南
│   ├── SYNC.md            # 同步使用指南
│   └── ...
└── android/                # Android 项目
```

## 🚀 快速开始

### 方式1：仅本地使用（推荐）

1. **下载可执行文件**
   - 从 [Releases](https://github.com/your-repo/releases) 下载对应平台版本
   - 或使用构建脚本编译：`scripts/build.sh` 或 `scripts/build.bat`

2. **运行程序**
   ```bash
   # Windows
   manage-agent-windows-amd64.exe
   
   # Linux
   ./manage-agent-linux-amd64
   ```

3. **访问界面**
   - 程序会自动打开浏览器
   - 或手动访问：http://localhost:8090

**就这么简单！** 无需安装数据库，无需配置，开箱即用。

### 方式2：使用云端同步

1. **部署云端服务器**（参考 [docs/DEPLOY.md](./docs/DEPLOY.md)）
2. **配置本地客户端**（参考 [docs/SYNC.md](./docs/SYNC.md)）

## 📚 文档

- [构建说明](./docs/BUILD.md) - 如何编译和打包
- [部署指南](./docs/DEPLOY.md) - 云端服务器部署
- [同步使用指南](./docs/SYNC.md) - 数据同步配置和使用
- [Android 构建](./android/README.md) - Android 平台构建

## 🛠️ 开发

### 环境要求

- Go 1.21+
- MySQL 8.0+（仅云端服务器需要）

### 本地开发

```bash
# 克隆项目
git clone https://github.com/your-repo/manage-agent.git
cd manage-agent

# 安装依赖
go mod download

# 运行本地客户端
go run bin/main.go

# 运行云端服务器（需要 MySQL）
go run cloud/main.go -c cloud/cloud-config.json
```

### 构建

```bash
# 构建所有平台
./scripts/build.sh        # Linux/Mac
scripts\build.bat          # Windows

# 构建云端服务器
./scripts/build-cloud.sh  # Linux/Mac
scripts\build-cloud.bat    # Windows
```

## ⚙️ 配置

### 本地客户端配置

`bin/config.json`:

```json
{
    "database": {
        "type": "sqlite",
        "sqlite": {
            "path": "./data/manage-agent.db"
        }
    },
    "sync": {
        "enabled": false,
        "cloud_api_url": "http://your-server:8080",
        "auto_sync_interval": 300
    }
}
```

### 云端服务器配置

`cloud/cloud-config.json`:

```json
{
    "database": {
        "type": "mysql",
        "mysql": {
            "username": "root",
            "password": "your_password",
            "host": "127.0.0.1",
            "port": "3306",
            "schema": "manage-agent"
        }
    },
    "server": {
        "addr": ":8080"
    }
}
```

## 🔧 API 文档

### 财务 API

- `GET /api/v1/finance/profit/` - 查询利润
- `GET /api/v1/finance/profit/history/` - 查询历史利润
- `GET /api/v1/finance/enum/` - 获取枚举值
- `POST /api/v1/finance/worth/` - 创建现值记录
- `POST /api/v1/finance/flow-record/` - 创建流水记录

### 同步 API（仅云端）

- `POST /api/v1/sync/upload` - 上传数据到云端
- `GET /api/v1/sync/download` - 从云端下载数据

## 📝 许可证

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！
