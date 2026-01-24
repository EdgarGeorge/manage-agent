# 项目结构说明

## 📁 目录结构

```
manage-agent/
├── README.md                 # 项目主文档
├── PROJECT_STRUCTURE.md      # 本文件
│
├── bin/                      # 本地客户端
│   ├── main.go              # 客户端主程序
│   ├── config.json          # 配置文件
│   ├── config.example.json  # 配置示例
│   └── html/                 # 前端资源（已嵌入）
│       ├── index.html
│       ├── assets/
│       └── api/
│
├── cloud/                    # 云端服务器
│   ├── main.go              # 云端服务器主程序
│   └── cloud-config.json    # 云端配置
│
├── server/                   # 共享服务层
│   ├── model.go             # 数据模型
│   ├── database.go          # 数据库连接
│   ├── router.go            # 路由注册
│   ├── sync.go              # 同步 API
│   ├── sync_client.go       # 同步客户端
│   ├── ctl.go               # 业务逻辑
│   ├── views.go             # 视图处理
│   └── ...                  # 其他服务文件
│
├── scripts/                  # 构建脚本
│   ├── build.sh             # Linux/Mac 构建脚本
│   ├── build.bat            # Windows 构建脚本
│   ├── build-cloud.sh       # 云端服务器构建（Linux/Mac）
│   ├── build-cloud.bat      # 云端服务器构建（Windows）
│   └── migrate_mysql_to_sqlite.py  # 数据迁移脚本
│
├── docs/                     # 文档目录
│   ├── README.md            # 文档索引
│   ├── BUILD.md             # 构建说明
│   ├── DEPLOY.md            # 部署指南
│   └── SYNC.md              # 同步使用指南
│
├── android/                  # Android 项目
│   ├── README.md
│   ├── build-android.sh
│   └── ...
│
├── dist/                     # 构建输出目录
├── log/                      # 日志目录
└── go.mod                    # Go 模块定义
```

## 📂 目录说明

### bin/
本地客户端主程序目录，包含：
- `main.go`: 客户端入口，使用 SQLite 数据库
- `config.json`: 客户端配置文件
- `html/`: 前端资源（使用 embed 嵌入到可执行文件）

### cloud/
云端服务器目录，包含：
- `main.go`: 云端服务器入口，使用 MySQL 数据库
- `cloud-config.json`: 云端服务器配置文件

### server/
共享服务层，包含所有业务逻辑：
- 数据模型定义
- 数据库连接管理
- API 路由和处理器
- 同步功能实现

### scripts/
所有构建和工具脚本：
- 客户端构建脚本
- 云端服务器构建脚本
- 数据迁移脚本

### docs/
所有项目文档：
- 构建说明
- 部署指南
- 使用指南

## 🔍 文件查找指南

- **想构建项目？** → 查看 `docs/BUILD.md` 或运行 `scripts/build.sh`
- **想部署云端？** → 查看 `docs/DEPLOY.md`
- **想配置同步？** → 查看 `docs/SYNC.md`
- **想了解项目？** → 查看 `README.md`

## 📝 代码组织原则

1. **分离关注点**：客户端和云端服务器分离
2. **共享代码**：业务逻辑放在 `server/` 目录
3. **配置分离**：配置文件与代码分离
4. **文档集中**：所有文档放在 `docs/` 目录
5. **脚本统一**：所有构建脚本放在 `scripts/` 目录
