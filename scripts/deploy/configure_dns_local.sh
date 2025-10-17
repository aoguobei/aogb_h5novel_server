#!/bin/bash
# scripts/configure_dns_local.sh
# 本地测试DNS配置脚本 - Windows环境适配

set -e

# ========================================
# DNS配置相关路径
# ========================================
DNSMASQ_CONF_PATH="C:/F_explorer/env/dns-test/dnsmasq.conf"                                               # dnsmasq配置文件路径
DNSMASQ_BACKUP_ROOT="C:/F_explorer/env/opt/backups/dns_backup"                             # DNS配置备份目录
MAX_DNS_BACKUPS=5                                                                           # 保留最近5个DNS备份

# ========================================
# DNS配置管理函数
# ========================================

# 1. 判断是否需要DNS配置
need_dns_config() {
    local domain="$1"
    local ssl_enabled="$2"
    local deploy_strategy="$3"
    
    # localhost 和 IP 地址不需要DNS配置
    if [ "$domain" = "localhost" ] || [ "$domain" = "127.0.0.1" ] || \
       [[ "$domain" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "false"
        return
    fi
    
    # 必须是HTTPS且需要创建新server块才配置DNS
    if [ "$ssl_enabled" = "true" ] && [ "$deploy_strategy" = "create_server" ]; then
        # 检查是否是内网域名（包含点号）
        if [[ "$domain" =~ \. ]]; then
            echo "true"
        else
            echo "false"
        fi
    else
        echo "false"
    fi
}

# 2. 获取本机IP地址
get_local_ip() {
    # Windows环境下获取本机IP
    local ip=$(ipconfig | grep -A 5 "以太网适配器" | grep "IPv4" | head -1 | awk '{print $NF}')
    if [ -z "$ip" ]; then
        # 备用方法
        ip=$(ipconfig | grep "IPv4" | head -1 | awk '{print $NF}')
    fi
    echo "$ip"
}

# 3. 备份DNS配置
backup_dns_config() {
    local domain="$1"
    
    # 检查是否需要DNS配置
    if [ "$(need_dns_config "$domain" "true" "create_server")" = "false" ]; then
        return 0
    fi
    
    mkdir -p "$DNSMASQ_BACKUP_ROOT"
    
    if [ -f "$DNSMASQ_CONF_PATH" ]; then
        local dns_backup_dir="$DNSMASQ_BACKUP_ROOT/$(date +%Y%m%d_%H%M%S)"
        mkdir -p "$dns_backup_dir"
        cp "$DNSMASQ_CONF_PATH" "$dns_backup_dir/"
        
        # 备份完成后执行备份恢复系统（包括清理旧备份）
        backup_recovery_system
    fi
}

# 4. 备份恢复机制
backup_recovery_system() {
    local max_backups="$MAX_DNS_BACKUPS"
    local backup_root="$DNSMASQ_BACKUP_ROOT"
    
    if [ -d "$backup_root" ]; then
        local backup_count=$(find "$backup_root" -maxdepth 1 -type d -name "20*" | wc -l)
        
        if [ "$backup_count" -gt "$max_backups" ]; then
            local to_delete=$((backup_count - max_backups))
            # 删除最旧的备份
            find "$backup_root" -maxdepth 1 -type d -name "20*" -printf '%T@ %p\n' | \
            sort -n | head -n "$to_delete" | \
            while read timestamp path; do
                rm -rf "$path"
            done
        fi
    fi
}


# 4.1 自动备份恢复（直接恢复最新备份）
auto_backup_recovery() {
    local domain="$1"
    local backup_root="$DNSMASQ_BACKUP_ROOT"
    
    if [ ! -d "$backup_root" ]; then
        echo "❌ 备份目录不存在，无法进行自动恢复"
        return 1
    fi
    
    # 查找最近的可用备份，按时间倒序排列（最新的在前）
    local latest_backup=$(find "$backup_root" -maxdepth 1 -type d -name "20*" | sort -r | head -1)
    
    if [ -z "$latest_backup" ]; then
        echo "❌ 没有找到可用的备份进行恢复"
        return 1
    fi
    
    local backup_name=$(basename "$latest_backup")
    echo "🎯 直接恢复最新备份: $backup_name"
    
    # 直接恢复最新备份，无需用户选择
    local backup_path="$backup_root/$backup_name"
    local backup_file="$backup_path/dnsmasq.conf"
    
    if [ ! -f "$backup_file" ]; then
        echo "❌ 备份文件不存在: $backup_file"
        return 1
    fi
    
    echo "🔄 开始恢复DNS配置..."
    echo "📁 恢复源: $backup_path"
    
    # 创建恢复前的备份
    local recovery_backup_dir="$backup_root/recovery_backup_$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$recovery_backup_dir"
    
    if [ -f "$DNSMASQ_CONF_PATH" ]; then
        cp "$DNSMASQ_CONF_PATH" "$recovery_backup_dir/"
        echo "✅ 当前配置已备份到: $recovery_backup_dir"
    fi
    
    # 恢复配置
    cp "$backup_file" "$DNSMASQ_CONF_PATH"
    
    # 验证配置
    if validate_dns_config; then
        echo "✅ DNS配置恢复成功"
        
        # 重启DNS服务
        if restart_dns_service; then
            echo "✅ DNS服务重启成功"
            echo "🎉 备份恢复完成！"
            echo "💡 如需回滚，当前配置已备份到: $recovery_backup_dir"
            return 0
        else
            echo "❌ DNS服务重启失败，开始回滚..."
            # 回滚到恢复前的配置
            if [ -f "$recovery_backup_dir/dnsmasq.conf" ]; then
                cp "$recovery_backup_dir/dnsmasq.conf" "$DNSMASQ_CONF_PATH"
                echo "✅ 已回滚到恢复前的配置"
            fi
            return 1
        fi
    else
        echo "❌ DNS配置验证失败，开始回滚..."
        # 回滚到恢复前的配置
        if [ -f "$recovery_backup_dir/dnsmasq.conf" ]; then
            cp "$recovery_backup_dir/dnsmasq.conf" "$DNSMASQ_CONF_PATH"
            echo "✅ 已回滚到恢复前的配置"
        fi
        return 1
    fi
}

# 5. 添加DNS配置
add_dns_config() {
    local domain="$1"
    local local_ip="$2"
    
    # 检查是否需要DNS配置
    if [ "$(need_dns_config "$domain" "true" "create_server")" = "false" ]; then
        return 0
    fi
    
    # 检查dnsmasq配置文件是否存在
    if [ ! -f "$DNSMASQ_CONF_PATH" ]; then
        echo "❌ dnsmasq配置文件不存在: $DNSMASQ_CONF_PATH"
        echo "💡 请检查dnsmasq是否已安装，或修改脚本中的DNSMASQ_CONF_PATH变量"
        return 1
    fi
    
    # 检查是否已存在该域名的配置
    if grep -q "server=/$domain/" "$DNSMASQ_CONF_PATH" || \
       grep -q "address=/$domain/" "$DNSMASQ_CONF_PATH"; then
        echo "ℹ️  DNS配置已存在: $domain"
        return 0
    fi
    
    # 添加DNS配置
    echo "" >> "$DNSMASQ_CONF_PATH"
    echo "# 脚本添加的本地测试域名配置 - $domain - $(date)" >> "$DNSMASQ_CONF_PATH"
    echo "server=/$domain/127.0.0.1" >> "$DNSMASQ_CONF_PATH"
    echo "address=/$domain/$local_ip" >> "$DNSMASQ_CONF_PATH"
    
    echo "✅ DNS配置已添加: $domain -> $local_ip"
    
    # 重启dnsmasq服务（如果可能）
    restart_dnsmasq_service
}

# 6. 重启DNS服务
restart_dnsmasq_service() {
    # 尝试不同的服务管理方式
    if command -v systemctl >/dev/null 2>&1; then
        # Linux systemd
        if systemctl is-active --quiet dnsmasq; then
            systemctl restart dnsmasq 2>/dev/null
        fi
    elif command -v service >/dev/null 2>&1; then
        # Linux service
        if service dnsmasq status >/dev/null 2>&1; then
            service dnsmasq restart 2>/dev/null
        fi
    elif command -v net stop >/dev/null 2>&1; then
        # Windows服务管理
        if sc query dnsmasq >/dev/null 2>&1; then
            net stop dnsmasq >/dev/null 2>&1
            net start dnsmasq >/dev/null 2>&1
        fi
    fi
}

# 7. 清理DNS配置
cleanup_dns_config() {
    local domain="$1"
    
    # 检查是否需要清理DNS配置
    if [ "$(need_dns_config "$domain" "true" "create_server")" = "false" ]; then
        return 0
    fi
    
    echo "🧹 清理DNS配置: $domain"
    
    if [ -f "$DNSMASQ_CONF_PATH" ]; then
        # 创建临时文件，直接过滤掉相关配置
        local temp_file="$(dirname "$DNSMASQ_CONF_PATH")/dnsmasq_temp_$$"
        
        # 简单直接的方法：过滤掉包含域名的行和注释模板行
        grep -v "server=/$domain/" "$DNSMASQ_CONF_PATH" | \
        grep -v "address=/$domain/" | \
        grep -v "脚本添加的本地测试域名配置.*$domain" > "$temp_file"
        
        # 替换原文件
        mv "$temp_file" "$DNSMASQ_CONF_PATH"
        echo "✅ DNS配置已清理: $domain"
        
        # 重启DNS服务
        restart_dnsmasq_service
    fi
}

# 8. 验证DNS配置
verify_dns_config() {
    local domain="$1"
    local expected_ip="$2"
    
    # 检查是否需要验证DNS配置
    if [ "$(need_dns_config "$domain" "true" "create_server")" = "false" ]; then
        return 0
    fi
    
    # 等待DNS生效
    sleep 2
    
    # 测试DNS解析
    local resolved_ip=""
    if command -v nslookup >/dev/null 2>&1; then
        resolved_ip=$(nslookup "$domain" 2>/dev/null | grep "Address:" | tail -1 | awk '{print $NF}')
    elif command -v dig >/dev/null 2>&1; then
        resolved_ip=$(dig +short "$domain" 2>/dev/null | head -1)
    else
        echo "⚠️  无法验证DNS配置，请手动检查"
        return 0
    fi
    
    if [ "$resolved_ip" = "$expected_ip" ]; then
        echo "✅ DNS配置验证成功: $domain -> $resolved_ip"
        return 0
    else
        echo "❌ DNS配置验证失败: 期望 $expected_ip，实际 $resolved_ip"
        return 1
    fi
}

# 9. 主函数
main() {
    local domain="$1"
    local ssl_enabled="$2"
    local deploy_strategy="$3"
    
    # 特殊处理：清理模式
    if [ "$ssl_enabled" = "cleanup" ] && [ "$deploy_strategy" = "rollback" ]; then
        echo "🧹 清理DNS配置: $domain"
        cleanup_dns_config "$domain"
        echo "✅ DNS配置清理完成"
        exit 0
    fi
    
    # 特殊处理：备份恢复模式
    if [ "$ssl_enabled" = "backup" ]; then
        echo "🎯 恢复最新备份..."
        auto_backup_recovery "auto"
        exit 0
    fi
    
    # 检查是否需要DNS配置
    if [ "$(need_dns_config "$domain" "$ssl_enabled" "$deploy_strategy")" = "false" ]; then
        echo "ℹ️  无需DNS配置，跳过"
        exit 0
    fi
    
    # 1. 备份DNS配置
    backup_dns_config "$domain"
    
    # 2. 获取本机IP地址
    local local_ip=$(get_local_ip)
    if [ -z "$local_ip" ]; then
        echo "❌ 无法获取本机IP地址"
        exit 1
    fi
    
    # 3. 添加DNS配置
    if ! add_dns_config "$domain" "$local_ip"; then
        echo "❌ DNS配置失败"
        exit 1
    fi
    
    # 4. 验证DNS配置
    if ! verify_dns_config "$domain" "$local_ip"; then
        echo "⚠️  DNS配置验证失败，但继续执行"
    fi
    
    echo "✅ DNS配置完成: $domain -> $local_ip"
}

# 10. 使用说明
show_usage() {
    echo "使用方法: $0 <域名> <SSL启用状态> <部署策略>"
    echo ""
    echo "参数说明:"
    echo "  域名: 要配置的域名"
    echo "  SSL启用状态: true/false/cleanup/backup"
    echo "  部署策略: create_server/add_location/skip_deployment/rollback/list/restore"
    echo ""
    echo "常规DNS配置:"
    echo "  $0 \"test.funshion.tv\" \"true\" \"create_server\""
    echo "  $0 \"localhost\" \"false\" \"add_location\""
    echo ""
    echo "清理DNS配置:"
    echo "  $0 \"test.funshion.tv\" \"cleanup\" \"rollback\""
    echo ""
        echo "备份恢复操作:"
    echo "  $0 \"\" \"backup\"                           # 直接恢复最新备份"
}

# 11. 脚本入口
if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    show_usage
    exit 0
fi

# 验证参数
if [ $# -lt 3 ]; then
    echo "❌ 参数不足！需要3个参数：域名、SSL启用状态、部署策略"
    show_usage
    exit 1
fi

# 执行主流程
main "$@" 