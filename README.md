# 小说H5网页配置后端

一个基于Go语言的小说H5网页配置管理系统后端，提供品牌管理、网站配置、文件生成等功能。

## 项目概述

本项目是一个小说H5网页配置管理系统的后端服务，主要功能包括：

- **品牌管理**：创建、编辑、删除品牌信息
- **网站配置**：管理不同平台的网站配置（H5、TT、KS等）
- **配置文件生成**：自动生成各种配置文件
- **文件操作**：管理项目文件、备份、回滚等
- **数据库管理**：MySQL数据库的CRUD操作
- **回滚机制**：完整的操作回滚和错误恢复
- **进度管理**：实时进度反馈和状态管理

## 技术栈

- **语言**：Go 1.21+
- **Web框架**：Gin
- **ORM**：GORM
- **数据库**：MySQL 8.0+
- **中间件**：CORS、日志、请求ID等
- **架构模式**：分层架构 + 服务层 + 回滚机制

## 项目结构

```
backend/
├── config/                 # 配置管理
│   └── config.go          # 应用配置
├── database/              # 数据库相关
│   └── database.go        # 数据库连接
├── handlers/              # HTTP处理器
│   ├── brand_handler.go   # 品牌处理器
│   ├── client_handler.go  # 客户端处理器
│   └── config_handler.go  # 配置处理器
├── middleware/            # 中间件
│   └── middleware.go      # 通用中间件
├── models/                # 数据模型
│   ├── brand.go          # 品牌模型
│   ├── client.go         # 客户端模型
│   └── configs.go        # 配置模型
├── routes/                # 路由定义
│   └── routes.go         # 路由配置
├── services/              # 业务逻辑层
│   ├── brand_service.go   # 品牌服务
│   ├── client_service.go  # 客户端服务
│   ├── base_config_service.go    # 基础配置服务
│   ├── common_config_service.go  # 通用配置服务
│   ├── pay_config_service.go     # 支付配置服务
│   ├── ui_config_service.go      # UI配置服务
│   ├── novel_config_service.go   # 小说配置服务
│   ├── file_service.go    # 文件服务
│   └── website_service.go # 网站服务
├── types/                 # 类型定义
│   └── types.go          # 通用类型定义
├── utils/                 # 工具函数
│   ├── response_utils.go  # 响应工具
│   ├── file_utils.go      # 文件工具
│   ├── config_file_manager.go    # 配置文件管理器
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
- 实时进度回调
- 特殊逻辑：tth5/ksh5自动创建对应的tt/ks小程序配置

### 5. 文件服务 (File)
- 项目文件管理
- 配置文件生成和修改
- 文件备份和恢复
- JSON文件操作（package.json、vite.config.js、manifest.json等）
- 目录结构管理
- 品牌配置文件更新
- Prebuild文件创建

### 6. 回滚机制 (Rollback)
- **回滚管理器 (RollbackManager)**：统一管理数据库和文件回滚
- **数据库回滚 (DatabaseRollback)**：数据库事务回滚
- **文件回滚 (FileRollback)**：文件操作回滚
- 原子操作保证
- 完整的备份恢复机制
- 事务上下文管理

### 7. 类型系统 (Types)
- **进度回调类型**：ProgressCallback接口
- **通用数据结构**：统一的数据类型定义
- **错误处理类型**：标准化的错误类型
- 事务上下文管理

### 7. 类型系统 (Types)
- **进度回调类型**：ProgressCallback接口
- **通用数据结构**：统一的数据类型定义
- **错误处理类型**：标准化的错误类型

## API接口文档

### 品牌相关接口

#### 获取品牌列表
```
GET /api/brands
```

#### 获取单个品牌
```
GET /api/brands/:id
```

#### 创建品牌
```
POST /api/brands
Content-Type: application/json

{
  "name": "品牌名称",
  "code": "brand_code",
  "description": "品牌描述"
}
```

#### 更新品牌
```
PUT /api/brands/:id
```

#### 删除品牌
```
DELETE /api/brands/:id
```

### 客户端相关接口

#### 获取客户端列表
```
GET /api/clients
```

#### 获取品牌下的客户端
```
GET /api/brands/:brandId/clients
```

#### 创建客户端
```
POST /api/clients
Content-Type: application/json

{
  "brand_id": 1,
  "host": "h5"
}
```

#### 删除客户端
```
DELETE /api/clients/:id
```

### 配置相关接口

#### 基础配置
```
GET /api/base-configs
POST /api/base-configs
PUT /api/base-configs/:id
DELETE /api/base-configs/:id
```

#### 通用配置
```
GET /api/common-configs
POST /api/common-configs
PUT /api/common-configs/:id
DELETE /api/common-configs/:id
```

#### 支付配置
```
GET /api/pay-configs
POST /api/pay-configs
PUT /api/pay-configs/:id
DELETE /api/pay-configs/:id
```

#### UI配置
```
GET /api/ui-configs
POST /api/ui-configs
PUT /api/ui-configs/:id
DELETE /api/ui-configs/:id
```

#### 小说配置
```
GET /api/novel-configs
POST /api/novel-configs
PUT /api/novel-configs/:id
DELETE /api/novel-configs/:id
```

### 网站相关接口

#### 创建网站（7步骤流程）
```
POST /api/create-website
Content-Type: application/json

{
  "basic_info": {
    "brand_id": 1,
    "host": "h5"
  },
  "base_config": {
    "app_name": "应用名称",
    "platform": "h5",
    "app_code": "app_code",
    "product": "product",
    "customer": "customer",
    "appid": "appid",
    "version": "1.0.0",
    "cl": "cl",
    "uc": "uc"
  },
  "common_config": {
    "deliver_business_id_enable": false,
    "deliver_business_id": "",
    "deliver_switch_id_enable": false,
    "deliver_switch_id": "",
    "protocol_company": "",
    "protocol_about": "",
    "protocol_privacy": "",
    "protocol_vod": "",
    "protocol_user_cancel": "",
    "contact_url": "",
    "script_base": ""
  },
  "pay_config": {
    "normal_pay_enable": true,
    "normal_pay_gateway_android": 1,
    "normal_pay_gateway_ios": 1,
    "renew_pay_enable": true,
    "renew_pay_gateway_android": 1,
    "renew_pay_gateway_ios": 1
  },
  "ui_config": {
    "theme_bg_main": "#ffffff",
    "theme_bg_second": "#f5f5f5",
    "theme_text_main": "#333333"
  },
  "novel_config": {
    "tt_jump_home_url": "https://example.com",
    "tt_login_callback_domain": "example.com"
  }
}
```

#### 获取网站配置
```
GET /api/website-config/:clientId
```

#### 删除网站
```
DELETE /api/website/:clientId
```

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

## 回滚机制详解

### 回滚管理器 (RollbackManager)
- **统一接口**：管理数据库和文件回滚
- **事务上下文**：提供事务和文件操作的统一上下文
- **进度回调**：支持实时进度反馈

### 数据库回滚 (DatabaseRollback)
- **事务管理**：开始、提交、回滚数据库事务
- **错误处理**：自动回滚失败的事务
- **日志记录**：完整的操作日志

### 文件回滚 (FileRollback)
- **文件备份**：操作前自动备份文件
- **目录备份**：支持目录级别的备份
- **回滚恢复**：失败时自动恢复文件状态

## 回滚机制详解

### 回滚管理器 (RollbackManager)
- **统一接口**：管理数据库和文件回滚
- **事务上下文**：提供事务和文件操作的统一上下文
- **进度回调**：支持实时进度反馈

### 数据库回滚 (DatabaseRollback)
- **事务管理**：开始、提交、回滚数据库事务
- **错误处理**：自动回滚失败的事务
- **日志记录**：完整的操作日志

### 文件回滚 (FileRollback)
- **文件备份**：操作前自动备份文件
- **目录备份**：支持目录级别的备份
- **回滚恢复**：失败时自动恢复文件状态

## 环境配置

### 环境变量

详细的配置说明请参考 [CONFIG.md](./CONFIG.md)

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

### 新增功能开发

1. **添加新的配置类型**：
   - 在 `models/configs.go` 中添加结构体
   - 在 `services/` 中添加对应服务
   - 在 `handlers/` 中添加处理器
   - 在 `routes/routes.go` 中添加路由

2. **修改文件生成逻辑**：
   - 修改 `services/file_service.go` 中的相关方法
   - 确保支持回滚机制
   - 添加进度回调支持

3. **扩展API接口**：
   - 在 `routes/routes.go` 中添加新路由
   - 在 `handlers/` 中添加对应的处理方法
   - 遵循统一的响应格式
- 进度回调支持实时状态反馈

### 新增功能开发

1. **添加新的配置类型**：
   - 在 `models/configs.go` 中添加结构体
   - 在 `services/` 中添加对应服务
   - 在 `handlers/` 中添加处理器
   - 在 `routes/routes.go` 中添加路由

2. **修改文件生成逻辑**：
   - 修改 `services/file_service.go` 中的相关方法
   - 确保支持回滚机制
   - 添加进度回调支持

3. **扩展API接口**：
   - 在 `routes/routes.go` 中添加新路由
   - 在 `handlers/` 中添加对应的处理方法
   - 遵循统一的响应格式

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

### Q: 如何添加新的配置类型？
A: 在 `models/configs.go` 中添加新的结构体，在 `services/` 中添加对应的服务，在 `handlers/` 中添加处理器，在 `routes/routes.go` 中添加路由。

### Q: 如何修改文件生成逻辑？
A: 修改 `services/file_service.go` 中的相关方法，确保支持回滚机制，添加进度回调支持。

### Q: 如何扩展API接口？
A: 在 `routes/routes.go` 中添加新的路由，在 `handlers/` 中添加对应的处理方法，遵循统一的响应格式。

### Q: 回滚机制如何工作？
A: 使用 `RollbackManager` 统一管理，支持数据库事务回滚和文件操作回滚，确保操作的原子性。

### Q: 如何实现进度回调？
A: 在服务方法中接收 `ProgressCallback` 函数，在关键步骤调用回调函数传递进度信息。

## 更新日志

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
