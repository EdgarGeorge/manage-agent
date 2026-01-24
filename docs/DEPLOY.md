# 云端服务器部署指南

## 一、概述

云端服务器用于集中存储和管理所有客户端的数据，支持多设备数据同步。

### 架构

```
┌─────────────┐         ┌─────────────┐
│ Windows客户端│────────▶│  云端服务器  │
│ (SQLite)    │  HTTP   │  (MySQL)    │
└─────────────┘         └─────────────┘
                               ▲
┌─────────────┐                │
│ Android客户端│────────────────┘
│ (SQLite)    │      HTTP
└─────────────┘
```

### 功能

1. **数据存储**：使用 MySQL 存储所有数据
2. **数据同步**：提供上传/下载 API
3. **冲突解决**：基于时间戳的自动冲突解决

## 二、部署步骤

### 步骤1：准备服务器环境

#### 1.1 安装 MySQL

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install mysql-server

# CentOS/RHEL
sudo yum install mysql-server

# 启动 MySQL
sudo systemctl start mysql
sudo systemctl enable mysql
```

#### 1.2 配置 MySQL

```bash
# 登录 MySQL
sudo mysql -u root -p

# 创建数据库和用户
CREATE DATABASE `manage-agent` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'manageagent'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON `manage-agent`.* TO 'manageagent'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

#### 1.3 安装 Go（如果服务器上没有）

```bash
# 下载 Go（以 1.21 为例）
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# 添加到 PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### 步骤2：编译云端服务器

#### 方式1：在本地编译（推荐）

```bash
# Windows 上编译 Linux 版本
set GOOS=linux
set GOARCH=amd64
go build -o manage-agent-cloud-linux-amd64 ./cloud/main.go

# 或使用构建脚本
./scripts/build-cloud.sh   # Linux/Mac
scripts\build-cloud.bat     # Windows
```

#### 方式2：在服务器上编译

```bash
# 上传代码到服务器
scp -r manage-agent user@your-server:/opt/

# SSH 登录服务器
ssh user@your-server

# 进入项目目录
cd /opt/manage-agent

# 编译
go build -o cloud/manage-agent-cloud ./cloud/main.go
```

### 步骤3：配置云端服务器

#### 3.1 创建配置文件

```bash
cd /opt/manage-agent/cloud
cp cloud-config.json cloud-config.json.bak
```

编辑 `cloud/cloud-config.json`：

```json
{
    "database": {
        "type": "mysql",
        "mysql": {
            "username": "manageagent",
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

### 步骤4：运行云端服务器

#### 4.1 直接运行（测试用）

```bash
cd /opt/manage-agent/cloud
./manage-agent-cloud -c ./cloud-config.json
```

#### 4.2 使用 systemd 服务（生产环境）

创建服务文件 `/etc/systemd/system/manage-agent-cloud.service`：

```ini
[Unit]
Description=Manage Agent Cloud Server
After=network.target mysql.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/manage-agent/cloud
ExecStart=/opt/manage-agent/cloud/manage-agent-cloud -c /opt/manage-agent/cloud/cloud-config.json
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable manage-agent-cloud
sudo systemctl start manage-agent-cloud
sudo systemctl status manage-agent-cloud
```

#### 4.3 使用 Nginx 反向代理（可选）

编辑 `/etc/nginx/sites-available/manage-agent`：

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

启用配置：

```bash
sudo ln -s /etc/nginx/sites-available/manage-agent /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 步骤5：配置防火墙

```bash
# 开放 8080 端口（如果直接访问）
sudo ufw allow 8080/tcp

# 或者只开放 80/443（如果使用 Nginx）
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
```

## 三、验证部署

### 3.1 检查服务状态

```bash
# 检查进程
ps aux | grep manage-agent-cloud

# 检查端口
netstat -tlnp | grep 8080

# 检查日志
tail -f /opt/manage-agent/log/cloud-latest.json
```

### 3.2 测试 API

```bash
# 测试健康检查（如果有）
curl http://your-server:8080/api/v1/sync/download

# 测试同步上传
curl -X POST http://your-server:8080/api/v1/sync/upload \
  -H "Content-Type: application/json" \
  -d '{"worth_records":[],"flow_records":[],"last_sync_time":"2024-01-01T00:00:00Z"}'
```

## 四、本地客户端配置

在本地客户端的 `bin/config.json` 中添加同步配置：

```json
{
    "database": {
        "type": "sqlite",
        "sqlite": {
            "path": "./data/manage-agent.db"
        }
    },
    "sync": {
        "enabled": true,
        "cloud_api_url": "http://your-server:8080",
        "auto_sync_interval": 300
    }
}
```

配置说明：
- `enabled`: 是否启用同步
- `cloud_api_url`: 云端服务器地址
- `auto_sync_interval`: 自动同步间隔（秒），0 表示不自动同步

## 五、数据同步流程

1. **本地创建数据** → `sync_status = 0`（未同步）
2. **自动同步触发** → 上传未同步数据到云端
3. **云端合并数据** → 根据时间戳解决冲突
4. **标记为已同步** → 本地 `sync_status = 1`
5. **下载云端数据** → 合并到本地数据库

## 六、故障排查

### 问题1：无法连接 MySQL

```bash
# 检查 MySQL 服务
sudo systemctl status mysql

# 检查 MySQL 端口
netstat -tlnp | grep 3306

# 测试连接
mysql -u manageagent -p -h 127.0.0.1
```

### 问题2：端口被占用

```bash
# 查看端口占用
sudo lsof -i :8080

# 修改配置文件中的端口
```

### 问题3：同步失败

- 检查网络连接
- 检查云端服务器日志
- 检查本地客户端日志
- 验证云端 API 地址是否正确

## 七、安全建议

1. **使用 HTTPS**：生产环境建议使用 Nginx + Let's Encrypt SSL
2. **数据库安全**：使用强密码，限制 MySQL 访问 IP
3. **防火墙**：只开放必要的端口
4. **定期备份**：定期备份 MySQL 数据库

```bash
# 备份数据库
mysqldump -u manageagent -p manage-agent > backup_$(date +%Y%m%d).sql

# 恢复数据库
mysql -u manageagent -p manage-agent < backup_20240101.sql
```

## 八、API 端点

### 同步 API

- `POST /api/v1/sync/upload` - 上传本地数据到云端
- `GET /api/v1/sync/download?last_sync_time=xxx` - 从云端下载数据

### 财务 API（与本地相同）

- `GET /api/v1/finance/profit/` - 查询利润
- `POST /api/v1/finance/worth/` - 创建现值记录
- `POST /api/v1/finance/flow-record/` - 创建流水记录
