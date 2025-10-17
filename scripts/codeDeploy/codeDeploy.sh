#!/bin/bash

# 代码部署脚本
# 功能：前后端代码编译更新部署

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 项目根目录
PROJECT_ROOT="/opt/websites/novel_h5_webconfig"
FRONTEND_DIR="$PROJECT_ROOT/h5Manager"
BACKEND_DIR="$PROJECT_ROOT/h5ManagerServer"
BACKEND_PORT="8080"

# 检查目录是否存在
check_directories() {
    log_info "检查项目目录..."
    
    if [ ! -d "$PROJECT_ROOT" ]; then
        log_error "项目根目录不存在: $PROJECT_ROOT"
        exit 1
    fi
    
    if [ ! -d "$FRONTEND_DIR" ]; then
        log_error "前端目录不存在: $FRONTEND_DIR"
        exit 1
    fi
    
    if [ ! -d "$BACKEND_DIR" ]; then
        log_error "后端目录不存在: $BACKEND_DIR"
        exit 1
    fi
    
    log_success "目录检查完成"
}

# 停止后端服务
stop_backend_service() {
    log_info "停止后端服务 (优雅优先)..."

    # 先优雅停止（SIGTERM）
    pkill -TERM -f "go run main.go" 2>/dev/null || true
    pkill -TERM -f "h5ManagerServer" 2>/dev/null || true

    # 等待最多10秒让其自行退出
    for i in $(seq 1 10); do
        if ! pgrep -f "go run main.go\|h5ManagerServer" > /dev/null; then
            break
        fi
        sleep 1
    done

    # 若仍存活则强制停止（SIGKILL）
    if pgrep -f "go run main.go\|h5ManagerServer" > /dev/null; then
        log_warning "优雅停止超时，强制杀进程..."
        pkill -9 -f "go run main.go" 2>/dev/null || true
        pkill -9 -f "h5ManagerServer" 2>/dev/null || true
    fi

    log_success "后端服务已停止"
}

# 更新前端代码
update_frontend() {
    log_info "开始更新前端代码..."
    
    cd "$FRONTEND_DIR"
    
    # Git pull（优先 git pull，其次 git pull origin master）
    log_info "拉取最新前端代码..."
    git pull || git pull origin master
    
    # 检查yarn是否安装
    if ! command -v yarn &> /dev/null; then
        log_error "yarn 未安装，请先安装 yarn"
        exit 1
    fi
    
    # 安装依赖
    log_info "安装前端依赖..."
    yarn install
    
    # 构建前端
    log_info "构建前端项目..."
    yarn build
    
    log_success "前端代码更新完成"
}

# 更新后端代码
update_backend() {
    log_info "开始更新后端代码..."
    
    cd "$BACKEND_DIR"
    
    # Git pull（优先 git pull，其次 git pull origin master）
    log_info "拉取最新后端代码..."
    git pull || git pull origin master
    
    # 检查Go是否安装
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装，请先安装 Go"
        exit 1
    fi
    
    # 下载依赖
    log_info "下载Go依赖..."
    go mod tidy
    go mod download
    
    # 编译后端
    log_info "编译后端项目..."
    go build -o h5ManagerServer main.go
    
    log_success "后端代码更新完成"
}

# 启动后端服务
start_backend_service() {
    log_info "启动后端服务..."
    
    cd "$BACKEND_DIR"
    
    # 检查端口是否被占用
    if netstat -tlnp 2>/dev/null | grep -q ":$BACKEND_PORT "; then
        log_warning "端口 $BACKEND_PORT 已被占用"
        netstat -tlnp | grep ":$BACKEND_PORT "
    fi
    
    # 启动服务
    nohup ./h5ManagerServer > server.log 2>&1 &
    local pid=$!
    
    # 等待服务启动
    sleep 3
    
    # 检查服务是否启动成功
    if kill -0 "$pid" 2>/dev/null; then
        log_success "后端服务启动成功，PID: $pid"
        log_info "服务日志: $BACKEND_DIR/server.log"
    else
        log_error "后端服务启动失败"
        if [ -f "$BACKEND_DIR/server.log" ]; then
            log_error "错误日志:"
            tail -20 "$BACKEND_DIR/server.log"
        fi
        exit 1
    fi
}

# 检查服务状态
check_service_status() {
    log_info "检查服务状态..."
    
    # 检查后端服务
    if pgrep -f "go run main.go\|h5ManagerServer" > /dev/null; then
        local pid=$(pgrep -f "go run main.go\|h5ManagerServer" | head -1)
        log_success "后端服务运行正常，PID: $pid"
    else
        log_warning "后端服务未运行"
    fi
    
    # 检查端口
    if netstat -tlnp 2>/dev/null | grep -q ":$BACKEND_PORT "; then
        log_success "端口 $BACKEND_PORT 监听正常"
    else
        log_warning "端口 $BACKEND_PORT 未监听"
    fi
}

# 显示菜单
show_menu() {
    echo ""
    echo "=========================================="
    echo "           代码部署脚本"
    echo "=========================================="
    echo "1. 完整部署 (前端+后端)"
    echo "2. 仅更新前端"
    echo "3. 仅更新后端"
    echo "4. 停止后端服务"
    echo "5. 启动后端服务"
    echo "6. 检查服务状态"
    echo "7. 重启后端服务"
    echo "0. 退出"
    echo "=========================================="
}

# 完整部署
full_deploy() {
    log_info "开始完整部署..."
    check_directories
    stop_backend_service
    update_frontend
    update_backend
    start_backend_service
    check_service_status
    log_success "完整部署完成！"
}

# 重启后端服务
restart_backend() {
    log_info "重启后端服务..."
    stop_backend_service
    sleep 2
    start_backend_service
    check_service_status
    log_success "后端服务重启完成！"
}

# 主函数
main() {
    while true; do
        show_menu
        read -p "请选择功能 (0-7): " choice
        
        case $choice in
            1)
                full_deploy
                ;;
            2)
                log_info "更新前端..."
                check_directories
                update_frontend
                log_success "前端更新完成！"
                ;;
            3)
                log_info "更新后端..."
                check_directories
                stop_backend_service
                update_backend
                start_backend_service
                check_service_status
                log_success "后端更新完成！"
                ;;
            4)
                stop_backend_service
                ;;
            5)
                log_info "启动后端服务..."
                check_directories
                start_backend_service
                check_service_status
                ;;
            6)
                check_service_status
                ;;
            7)
                restart_backend
                ;;
            0)
                log_info "退出脚本"
                exit 0
                ;;
            *)
                log_error "无效选择，请输入 0-7"
                ;;
        esac
        
        echo ""
        read -p "按回车键继续..."
    done
}

# 启动脚本
main
