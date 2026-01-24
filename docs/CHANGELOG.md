# 项目整理记录

## 整理日期
2024-01-23

## 整理内容

### 1. 文档整理
- ✅ 创建 `docs/` 目录，统一存放所有文档
- ✅ 合并重复文档：
  - `BUILD.md` → `docs/BUILD.md`（构建说明）
  - `DEPLOY.md` → `docs/DEPLOY.md`（部署指南）
  - `SYNC_GUIDE.md` + `CLOUD_SERVER_SUMMARY.md` → `docs/SYNC.md`（同步指南）
- ✅ 创建 `docs/README.md` 作为文档索引
- ✅ 创建项目主 `README.md`

### 2. 脚本整理
- ✅ 创建 `scripts/` 目录，统一存放所有构建脚本
- ✅ 移动构建脚本：
  - `build.sh` → `scripts/build.sh`
  - `build.bat` → `scripts/build.bat`
  - `build-cloud.sh` → `scripts/build-cloud.sh`
  - `build-cloud.bat` → `scripts/build-cloud.bat`
- ✅ 合并 `script/` 目录到 `scripts/`
- ✅ 更新脚本，添加路径检查和错误处理

### 3. 目录结构优化
- ✅ 清理根目录，移除分散的文档和脚本
- ✅ 保持代码目录结构清晰：
  - `bin/` - 本地客户端
  - `cloud/` - 云端服务器
  - `server/` - 共享服务层
  - `android/` - Android 项目

### 4. 文档更新
- ✅ 更新所有文档中的路径引用
- ✅ 统一文档格式和风格
- ✅ 创建 `PROJECT_STRUCTURE.md` 说明项目结构

## 整理后的目录结构

```
manage-agent/
├── README.md                 # 项目主文档
├── PROJECT_STRUCTURE.md     # 项目结构说明
├── bin/                      # 本地客户端
├── cloud/                    # 云端服务器
├── server/                   # 共享服务层
├── scripts/                  # 构建脚本（统一）
├── docs/                     # 文档（统一）
└── android/                  # Android 项目
```

## 使用说明

### 构建项目
```bash
# Linux/Mac
./scripts/build.sh

# Windows
scripts\build.bat
```

### 查看文档
- 项目概述：`README.md`
- 构建说明：`docs/BUILD.md`
- 部署指南：`docs/DEPLOY.md`
- 同步指南：`docs/SYNC.md`
- 项目结构：`PROJECT_STRUCTURE.md`

## 注意事项

1. 所有构建脚本需要在项目根目录运行
2. 文档路径已更新，请使用新的路径引用
3. 配置文件位置保持不变：
   - 客户端：`bin/config.json`
   - 云端：`cloud/cloud-config.json`
