# 配置说明

## 环境变量配置

### 基础路径配置（重要）
```bash
# 设置你的项目基础路径，所有其他路径都会基于这个路径自动生成
BASE_PATH=C:/F_explorer/h5projects/jianruiH5/novel_h5config
```

### 数据库配置
```bash
DB_HOST=localhost          # 数据库主机地址
DB_PORT=3306              # 数据库端口
DB_USER=root              # 数据库用户名
DB_PASSWORD=your_password # 数据库密码
DB_NAME=h5novel_config    # 数据库名称
```

### 服务器配置
```bash
PORT=8080                 # 服务器端口
GIN_MODE=debug           # Gin运行模式
```

## 路径自动生成

设置 `BASE_PATH` 后，以下路径会自动生成：

```bash
# 项目根目录
PROJECT_ROOT = BASE_PATH/funNovel

# 配置文件目录
CONFIG_DIR = BASE_PATH/funNovel/src/appConfig

# 预构建目录
PREBUILD_DIR = BASE_PATH/funNovel/prebuild/build

# 静态资源目录
STATIC_DIR = BASE_PATH/funNovel/src/static

# 各种配置文件目录
BASE_CONFIGS_DIR = BASE_PATH/funNovel/src/appConfig/baseConfigs
COMMON_CONFIGS_DIR = BASE_PATH/funNovel/src/appConfig/commonConfigs
PAY_CONFIGS_DIR = BASE_PATH/funNovel/src/appConfig/payConfigs
UI_CONFIGS_DIR = BASE_PATH/funNovel/src/appConfig/uiConfigs
LOCAL_CONFIGS_DIR = BASE_PATH/funNovel/src/appConfig/localConfigs
NOVEL_CONFIG_FILE = BASE_PATH/funNovel/src/appConfig/localConfigs/novelConfig.js
```

## 快速开始

1. 复制项目到你的本地目录
2. 设置 `BASE_PATH` 环境变量指向你的项目根目录
3. 设置数据库配置
4. 运行项目

## 示例配置

### Windows系统
```cmd
set BASE_PATH=D:\your\project\path
set DB_HOST=localhost
set DB_PORT=3306
set DB_USER=root
set DB_PASSWORD=your_password
set DB_NAME=h5novel_config
```

### Linux/Mac系统
```bash
export BASE_PATH=/path/to/your/project
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=your_password
export DB_NAME=h5novel_config
```

### 使用.env文件
创建 `.env` 文件在项目根目录：
```bash
BASE_PATH=D:\your\project\path
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=h5novel_config
PORT=8080
GIN_MODE=debug
```

## 项目结构要求

你的项目应该有以下结构：
```
your_project_path/
├── funNovel/
│   ├── src/
│   │   ├── appConfig/
│   │   │   ├── baseConfigs/
│   │   │   ├── commonConfigs/
│   │   │   ├── payConfigs/
│   │   │   ├── uiConfigs/
│   │   │   ├── localConfigs/
│   │   │   ├── _host.js
│   │   │   └── index.js
│   │   └── static/
│   ├── prebuild/
│   ├── vite.config.js
│   └── package.json
```

## 注意事项

1. `BASE_PATH` 必须指向包含 `funNovel` 目录的路径
2. 数据库配置必须正确，否则应用无法启动
3. 确保应用有足够的权限访问配置的目录
4. 所有路径都支持相对路径和绝对路径 