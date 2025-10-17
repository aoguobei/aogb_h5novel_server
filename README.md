# 小说H5网页配置后端

一个基于Go语言的小说H5网页配置管理系统后端，提供品牌管理、网站配置、文件生成、远程部署等功能。

## 项目概述

本项目是一个小说H5网页配置管理系统的后端服务，主要功能包括：

- **品牌管理**：创建、编辑、删除品牌信息
- **网站配置**：管理不同平台的网站配置（H5、TT、KS等）
- **配置文件生成**：自动生成各种配置文件
- **远程部署**：支持Nginx远程部署和H5项目构建
- **Git操作**：代码提交、拉取、分支管理等
- **WebSocket**：实时进度反馈和状态管理
- **回滚机制**：完整的操作回滚和错误恢复

## 技术栈

- **语言**：Go 1.21+
- **Web框架**：Gin
- **ORM**：GORM
- **数据库**：MySQL 8.0+
- **中间件**：CORS、日志、请求ID等
- **架构模式**：分层架构 + 服务层 + 回滚机制

## 项目结构

```
h5ManagerServer/
├── config/                 # 配置管理
│   └── config.go          # 应用配置
├── database/              # 数据库相关
│   └── database.go        # 数据库连接
├── handlers/              # HTTP处理器
│   ├── brand_handler.go   # 品牌处理器
│   ├── client_handler.go  # 客户端处理器
│   ├── website_handler.go # 网站处理器
│   ├── deploy_handler.go  # 部署处理器
│   ├── build_handler.go   # 构建处理器
│   ├── git_handler.go     # Git处理器
│   └── websocket_handler.go # WebSocket处理器
├── middleware/            # 中间件
│   └── middleware.go      # 通用中间件
├── models/                # 数据模型
│   ├── brand.go          # 品牌模型
│   ├── client.go         # 客户端模型
│   └── configs.go        # 配置模型
├── routes/                # 路由定义
│   ├── routes.go         # 主路由配置
│   ├── deploy_routes.go  # 部署路由
│   ├── build_routes.go   # 构建路由
│   └── git_routes.go     # Git路由
├── services/              # 业务逻辑层
│   ├── brand_service.go   # 品牌服务
│   ├── client_service.go  # 客户端服务
│   ├── base_config_service.go    # 基础配置服务
│   ├── common_config_service.go  # 通用配置服务
│   ├── pay_config_service.go     # 支付配置服务
│   ├── ui_config_service.go      # UI配置服务
│   ├── novel_config_service.go   # 小说配置服务
│   ├── file_service.go    # 文件服务
│   ├── website_service.go # 网站服务
│   ├── deploy_service.go  # 部署服务
│   └── build_service.go   # 构建服务
├── types/                 # 类型定义
│   └── types.go          # 通用类型定义
├── utils/                 # 工具函数
│   ├── response_utils.go  # 响应工具
│   ├── file_utils.go      # 文件工具
│   ├── websocket.go       # WebSocket管理
│   ├── task_manager.go    # 任务管理
│   └── rollback/          # 回滚机制
│       ├── database_rollback.go  # 数据库回滚
│       ├── file_rollback.go      # 文件回滚
│       └── rollback_manager.go   # 回滚管理器
├── go.mod                 # Go模块文件
├── go.sum                 # 依赖校验文件
├── main.go               # 主程序入口
└── README.md             # 项目说明
```

## 核心功能模块

### 1. 品牌管理 (Brand)
- 创建、编辑、删除品牌
- 品牌信息管理（名称、代码、描述等）
- 品牌关联的客户端管理

### 2. 客户端管理 (Client)
- 客户端创建和管理
- 支持多种平台（H5、TT、KS等）
- 与品牌的关联关系
- 智能Host过滤（避免重复创建）

### 3. 配置管理 (Config)
- **基础配置 (BaseConfig)**：应用基本信息（app_name, platform, app_code等）
- **通用配置 (CommonConfig)**：通用业务配置（协议、脚本等）
- **支付配置 (PayConfig)**：支付相关配置（网关、开关等）
- **UI配置 (UIConfig)**：界面主题配置（主题色、背景色、文字色）
- **小说配置 (NovelConfig)**：小说相关配置（抖音跳转、登录回调等）

### 4. 网站服务 (Website)
- **7步骤创建流程**：基本信息→基础配置→通用配置→支付配置→UI配置→小说配置→生成文件
- 网站创建流程（原子操作）
- 配置文件自动生成
- 文件操作和备份
- 完整的回滚机制
- 支持多种平台（H5、TTH5、KSH5等）
- 实时进度回调
- 特殊逻辑：tth5/ksh5自动创建对应的tt/ks小程序配置

### 5. 部署服务 (Deploy)
- **Nginx远程部署**：支持SSH连接远程服务器部署Nginx配置
- **H5项目构建**：支持H5项目的构建和部署
- **WebSocket实时反馈**：部署过程的实时进度和日志输出
- **任务管理**：支持并发任务管理和状态跟踪

### 6. 构建服务 (Build)
- **H5项目构建**：支持Vue/UniApp项目的构建
- **多平台支持**：支持不同平台的构建配置
- **实时进度**：构建过程的实时反馈
- **错误处理**：完整的错误处理和恢复机制

### 7. Git操作 (Git)
- **代码提交**：支持代码的提交和推送
- **分支管理**：支持分支的创建、切换、合并
- **代码拉取**：支持从远程仓库拉取代码
- **状态查询**：支持Git状态的查询和显示

### 8. 文件服务 (File)
- 项目文件管理
- 配置文件生成和修改
- 文件备份和恢复
- JSON文件操作（package.json、vite.config.js、manifest.json等）
- 目录结构管理
- 品牌配置文件更新
- Prebuild文件创建

### 9. 回滚机制 (Rollback)
- **回滚管理器 (RollbackManager)**：统一管理数据库和文件回滚
- **数据库回滚 (DatabaseRollback)**：数据库事务回滚
- **文件回滚 (FileRollback)**：文件操作回滚
- 原子操作保证
- 完整的备份恢复机制
- 事务上下文管理

### 10. WebSocket通信
- **实时通信**：支持部署和构建过程的实时通信
- **任务状态**：实时任务状态更新
- **日志输出**：实时日志输出和显示
- **错误处理**：连接错误和异常处理

## 数据库设计

### 主要数据表

1. **brands** - 品牌表
2. **clients** - 客户端表
3. **base_configs** - 基础配置表
4. **common_configs** - 通用配置表
5. **pay_configs** - 支付配置表
6. **ui_configs** - UI配置表
7. **novel_configs** - 小说配置表

### 表关系

- `brands` 1:N `clients`
- `clients` 1:1 `base_configs`
- `clients` 1:1 `common_configs`
- `clients` 1:1 `pay_configs`
- `clients` 1:1 `ui_configs`
- `clients` 1:1 `novel_configs`

## 重难点技术解析

### 1. Nginx远程部署系统 

#### 技术难点
- **SSH连接管理**：远程服务器连接、认证、超时处理
- **脚本上传与执行**：部署脚本的远程上传和权限执行
- **流式输出处理**：实时读取远程命令的stdout/stderr
- **配置回滚机制**：部署失败时的完整配置回滚

#### 核心实现
```bash
# 部署流程
1. SSH连接建立 → 2. 脚本文件上传 → 3. 权限检查 → 4. 配置备份 
→ 5. Nginx配置生成 → 6. 配置验证 → 7. 服务重载 → 8. DNS配置
```

#### 关键特性
- **智能配置检测**：自动检测现有配置，避免重复部署
- **SSL自动判断**：根据端口号自动判断是否启用HTTPS
- **权限自动提升**：非root用户自动使用sudo执行
- **信号处理**：支持中断信号，确保回滚操作
- **多环境支持**：Linux服务器和本地测试环境

### 2. DNS配置管理系统 🌐

#### 技术难点
- **域名解析配置**：dnsmasq配置文件的动态修改
- **IP地址获取**：跨平台的本机IP地址获取
- **服务重启管理**：DNS服务的自动重启和状态检查
- **配置验证**：DNS配置的正确性验证

#### 核心实现
```bash
# DNS配置流程
1. 域名检查 → 2. 配置备份 → 3. IP获取 → 4. 配置添加 
→ 5. 服务重启 → 6. 配置验证 → 7. 清理机制
```

#### 关键特性
- **智能判断**：只有HTTPS且需要创建新server块时才配置DNS
- **重复检测**：避免重复添加相同的域名配置
- **多平台支持**：Windows和Linux环境的不同处理逻辑
- **服务管理**：支持systemctl、service等多种服务管理方式
- **自动清理**：支持配置的自动清理和恢复

### 3. 7步骤网站创建流程 ⚙️

#### 技术难点
- **原子性操作**：11个步骤的原子性执行，任何失败都要完全回滚
- **状态管理**：每个步骤的精确状态跟踪和进度回调
- **依赖关系**：步骤间的复杂依赖关系和错误传播
- **特殊逻辑**：tth5/ksh5自动创建对应小程序配置

#### 核心实现
```go
// 11步骤创建流程
1. 验证数据 → 2. 创建客户端 → 3. 创建基础配置 → 4. 创建通用配置 
→ 5. 创建支付配置 → 6. 创建UI配置 → 7. 创建额外客户端 → 8. 创建小说配置 
→ 9. 更新项目配置 → 10. 创建预构建文件 → 11. 创建静态资源
```

#### 关键特性
- **回滚机制**：使用RollbackManager统一管理数据库和文件回滚
- **进度回调**：实时进度反馈，支持WebSocket推送
- **错误处理**：完整的错误分类和处理策略
- **事务管理**：数据库事务和文件操作的统一管理

### 4. 实时WebSocket通信系统 📡

#### 技术难点
- **连接管理**：多任务并发时的连接状态管理
- **消息队列**：连接建立前消息的缓存和重发
- **并发写入**：防止多个goroutine同时写入同一连接
- **错误恢复**：连接断开时的自动重连和消息恢复

#### 核心实现
```go
// WebSocket管理
type WebSocketManager struct {
    connections  map[string]*websocket.Conn
    messageQueue map[string][]interface{}
    writeMutex   map[string]*sync.Mutex
}
```

#### 关键特性
- **消息队列**：连接建立前消息的自动缓存
- **写锁机制**：每个连接独立的写锁，防止并发冲突
- **超时处理**：写入超时和连接状态检查
- **广播支持**：支持消息的广播和单播

### 5. 复杂回滚机制 

#### 技术难点
- **数据库回滚**：GORM事务的自动回滚和错误处理
- **文件回滚**：文件操作的备份、恢复和清理
- **统一管理**：数据库和文件回滚的统一协调
- **并发安全**：多goroutine环境下的线程安全

#### 核心实现
```go
// 回滚管理器
type RollbackManager struct {
    dbManager  *DatabaseRollback
    fileManager *FileRollback
}

// 执行事务操作
func (rm *RollbackManager) ExecuteWithTransaction(
    operation func(*TransactionContext) error, 
    progressCallback types.ProgressCallback
) error
```

#### 关键特性
- **原子操作**：确保操作的原子性，要么全部成功要么全部回滚
- **进度回调**：回滚过程的实时进度反馈
- **错误分类**：不同错误类型的分类处理
- **资源清理**：回滚后的资源清理和状态恢复

## 环境配置

### 环境变量

```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=h5novel_config

# 服务器配置
PORT=8080
GIN_MODE=debug

# 基础路径配置（重要：设置你的项目根路径）
BASE_PATH=C:/F_explorer/h5projects/jianruiH5/novel_h5config
```

### 启动步骤

1. **安装依赖**
```bash
go mod download
```

2. **配置数据库**
```bash
# 创建数据库
mysql -u root -p < database/init.sql

# 初始化表结构
mysql -u root -p h5novel_config < database/novel_configs.sql

# 插入初始数据（可选）
mysql -u root -p h5novel_config < database/seed.sql
```

3. **启动服务**
```bash
go run main.go
```

## 开发指南

### 代码规范

- 使用Go官方代码规范
- 函数和变量使用驼峰命名
- 结构体字段使用snake_case的JSON标签
- 错误处理要完整
- 使用统一的响应格式

### 架构说明

- **Handler层**：处理HTTP请求，参数验证
- **Service层**：业务逻辑处理，事务管理
- **Model层**：数据模型定义
- **Utils层**：工具函数和辅助功能
- **Rollback层**：回滚机制和错误恢复

### 错误处理

- 使用统一的错误响应格式
- 数据库操作要有事务处理
- 文件操作要有回滚机制
- 原子操作保证数据一致性
- 完整的日志记录和错误追踪
- 进度回调支持实时状态反馈

## 部署说明

### 生产环境部署

1. **编译**
```bash
go build -o brand-config-api main.go
```

2. **配置环境变量**
```bash
export DB_HOST=production_host
export DB_PORT=3306
export DB_USER=prod_user
export DB_PASSWORD=prod_password
export DB_NAME=prod_database
export PORT=8080
export BASE_PATH=/path/to/novel_h5config
```

3. **启动服务**
```bash
./brand-config-api
```

### Docker部署

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o brand-config-api main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/brand-config-api .
EXPOSE 8080
CMD ["./brand-config-api"]
```

## 维护说明

### 日志管理

- 使用结构化日志
- 记录关键操作和错误信息
- 定期清理日志文件
- 支持进度回调日志

### 备份策略

- 数据库定期备份
- 配置文件版本管理
- 重要文件备份
- 支持自动回滚恢复

### 监控告警

- 服务健康检查
- 数据库连接监控
- 错误率监控
- 进度状态监控

## 常见问题

### Q: 回滚机制如何工作？
A: 使用 `RollbackManager` 统一管理，支持数据库事务回滚和文件操作回滚，确保操作的原子性。

### Q: 如何实现进度回调？
A: 在服务方法中接收 `ProgressCallback` 函数，在关键步骤调用回调函数传递进度信息。

### Q: Nginx部署失败如何处理？
A: 系统会自动回滚到部署前的配置状态，包括nginx配置和DNS配置的完整恢复。

### Q: DNS配置什么时候生效？
A: DNS配置在nginx部署成功后自动执行，只有HTTPS且需要创建新server块时才会配置DNS。

## 更新日志

### v1.3.0
- **新增**: 远程部署功能（Nginx部署、H5构建）
- **新增**: WebSocket实时通信
- **新增**: 任务管理系统
- **新增**: Git操作支持
- **优化**: 错误处理和恢复机制
- **优化**: 实时进度反馈

### v1.2.0
- **新增**: 7步骤网站创建流程
- **新增**: 完整的回滚机制
- **新增**: 实时进度回调
- **新增**: 小说配置支持
- **新增**: 抖音跳转URL生成
- **优化**: 文件服务架构
- **优化**: 错误处理和恢复

### v1.1.0
- 完善回滚机制
- 优化文件操作
- 增强错误处理
- 支持多种平台配置

### v1.0.0
- 初始版本发布
- 基础CRUD功能
- 文件生成功能
- 基础回滚机制

## 贡献指南

1. Fork项目
2. 创建功能分支
3. 提交代码
4. 创建Pull Request

## 许可证

MIT License
