#!/bin/bash
# scripts/deploy_nginx_config_linux_server.sh
# Linux服务器版本的nginx配置部署脚本

# 检查是否以root身份运行，如果不是，尝试使用sudo重新执行
if [ "$EUID" -ne 0 ]; then
    echo "🔐 检测到非root用户，使用sudo权限执行..."
    # 重新执行脚本，但这次使用sudo
    exec sudo "$0" "$@"
fi

echo "✅ 正在以root权限执行脚本"

set -e

# ========================================
# Linux服务器配置文件路径配置
# ========================================
# Linux环境下的nginx配置路径（标准Linux路径）
NGINX_CONF_PATH="/usr/local/nginx/conf/nginx.conf"                                    # nginx配置文件路径
NGINX_BIN_PATH="/usr/local/nginx/sbin/nginx"                                            # nginx可执行文件路径
NGINX_LOGS_PATH="/usr/local/nginx/logs"                                     # nginx日志目录路径
BACKUP_ROOT="/opt/nginx_dns_backups/nginx_conf_backup"                      # Linux备份目录

# DNS配置脚本路径
DNS_SCRIPT_PATH="$(dirname "$0")/configure_dns_linux_server.sh"

echo "══════════════════════════════════════════════════════════════════════════════════"
echo "🔧 环境配置与检测"
echo "══════════════════════════════════════════════════════════════════════════════════"
echo "   nginx: $NGINX_CONF_PATH"
echo "   备份: $BACKUP_ROOT"

# 信号处理 - 确保中断时回滚
trap 'echo "⚠️  脚本被中断，开始回滚..."; rollback_configs; exit 1' INT TERM

# 接收参数
DOMAIN="$1"
PORT="$2"
ROOT_PATH="$3"
LOCATION_PATH="$4"
SSL_CERT_PATH="$5"
SSL_KEY_PATH="$6"

echo "🚀 Linux服务器部署配置: $DOMAIN -> $LOCATION_PATH (端口: $PORT)"

# ========================================
# nginx配置函数
# ========================================

# 1. 备份当前配置
# 注意：只备份到BACKUP_ROOT目录，避免在nginx/conf目录下创建重复备份
backup_dir="$BACKUP_ROOT/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$backup_dir"

if [ -f "$NGINX_CONF_PATH" ]; then
    cp "$NGINX_CONF_PATH" "$backup_dir/"
    echo "✅ nginx配置已备份到: $backup_dir"
fi

# 2. 生成简洁的location配置
generate_location_config() {
    # 特殊处理 location_path 为 / 的情况
    local try_files_path
    if [ "$LOCATION_PATH" = "/" ]; then
        try_files_path="/index.html"
    else
        try_files_path="$LOCATION_PATH/index.html"
    fi

    cat << EOF
        location $LOCATION_PATH {
            alias $ROOT_PATH;
            try_files \$uri \$uri/ $try_files_path;
            index index.html;
        }
EOF
}

# 3. 自动判断SSL状态
determine_ssl_status() {
    # 只有当端口是标准HTTPS端口时才启用SSL
    # 忽略证书文件参数，避免端口和SSL配置不匹配的问题
    if [ "$PORT" = "443" ] || [ "$PORT" = "8443" ] || [ "$PORT" = "9443" ]; then
        echo "true"
    else
        echo "false"
    fi
}

# 4. 生成完整的server配置（包含error_page和50x.html）
generate_server_config() {
    local ssl_enabled=$(determine_ssl_status)

    if [ "$ssl_enabled" = "true" ]; then
        # HTTPS配置 - 只有当端口是HTTPS时才使用SSL
        cat << EOF
server {
    listen       $PORT ssl;
    server_name  $DOMAIN;
EOF

        # 只有当提供了证书文件时才添加SSL配置
        if [ -n "$SSL_CERT_PATH" ] && [ -n "$SSL_KEY_PATH" ]; then
            cat << EOF
    ssl_certificate $SSL_CERT_PATH;
    ssl_certificate_key $SSL_KEY_PATH;
EOF
        fi

        cat << EOF

EOF
        generate_location_config

        cat << EOF

    error_page 404 /index.html;
    location = /50x.html {
        root   html;
    }
}
EOF
    else
        # HTTP配置
        cat << EOF
server {
    listen       $PORT;
    server_name  $DOMAIN;
EOF
        generate_location_config

        cat << EOF

    error_page 404 /index.html;
    location = /50x.html {
        root   html;
    }
}
EOF
    fi
}

# 5. 检查配置组合是否存在
check_config_exists() {
    local domain="$1"
    local port="$2"
    local location_path="$3"

    # 重新设计变量，含义更清晰
    local has_matching_server=false      # 是否有匹配的server块（域名+端口）
    local has_matching_location=false    # 是否有匹配的location路径
    local is_exact_duplicate=false      # 是否完全重复（server+端口+location都相同）

    local server_start_line=0
    local server_end_line=0
    local location_start_line=0
    local location_end_line=0
    local server_port=""

    # 查找域名和端口组合 - 单域名精确匹配
    # 使用单词边界确保精确匹配，避免 ap.funshion.com 匹配到 aap.funshion.com
    local server_blocks=""

    # 精确匹配单域名：server_name 后面直接是目标域名
    server_blocks=$(grep -n "server_name[[:space:]]\+\<$domain\>[[:space:]]*;" "$NGINX_CONF_PATH" | cut -d: -f1)

    # 如果没找到，匹配行末没有分号的情况
    if [ -z "$server_blocks" ]; then
        server_blocks=$(grep -n "server_name[[:space:]]\+\<$domain\>[[:space:]]*$" "$NGINX_CONF_PATH" | cut -d: -f1)
    fi

    if [ -n "$server_blocks" ]; then
        # 遍历所有匹配的server块，找到正确的端口
        while IFS= read -r server_name_line; do
            if [ -z "$server_name_line" ]; then
                continue
            fi

            # 向前查找server块的开始
            local line_num=$server_name_line
            while [ $line_num -gt 0 ]; do
                local line=$(sed -n "${line_num}p" "$NGINX_CONF_PATH")
                if [[ $line =~ server[[:space:]]*\{ ]]; then
                    server_start_line=$line_num
                    break
                fi
                ((line_num--))
            done

            # 如果找到了server块开始，查找结束位置和端口
            if [ $server_start_line -gt 0 ]; then
                local brace_count=0
                local in_server_block=false
                local found_port=false

                for ((line_num=server_start_line; line_num<=$(wc -l < "$NGINX_CONF_PATH"); line_num++)); do
                    local line=$(sed -n "${line_num}p" "$NGINX_CONF_PATH")

                    if [[ $line =~ server[[:space:]]*\{ ]]; then
                        in_server_block=true
                        brace_count=1
                    elif [[ $line =~ listen[[:space:]]+([0-9]+) ]] && [ "$in_server_block" = true ]; then
                        # 匹配各种 listen 指令格式：listen 80; listen 443 ssl; listen 80 default_server;
                        server_port="${BASH_REMATCH[1]}"
                        if [ "$server_port" = "$port" ]; then
                            found_port=true
                        fi
                    elif [[ $line =~ \{ ]] && [ "$in_server_block" = true ]; then
                        ((brace_count++))
                    elif [[ $line =~ \} ]] && [ "$in_server_block" = true ]; then
                        ((brace_count--))
                        if [ $brace_count -eq 0 ]; then
                            server_end_line=$line_num
                            break
                        fi
                    fi
                done

                # 如果端口匹配，检查location
                if [ "$found_port" = true ] && [ $server_end_line -gt $server_start_line ]; then
                    has_matching_server=true

                    # 在server块内查找location
                    # 匹配location后面去掉空格第一个是/的，提取/到{之间的路径
                    local location_found=""
                    local found_location_path=""

                    # 在server块范围内逐行查找location
                    for ((line_num=server_start_line; line_num<=server_end_line; line_num++)); do
                        local line=$(sed -n "${line_num}p" "$NGINX_CONF_PATH")

                        # 匹配location行：location后面去掉空格第一个字符是/
                        if [[ $line =~ location[[:space:]]+(/[^[:space:]]*)[[:space:]]*\{ ]]; then
                            found_location_path="${BASH_REMATCH[1]}"
                            # 去掉路径中的空格
                            found_location_path=$(echo "$found_location_path" | tr -d ' ')

                            # 检查是否匹配目标路径
                            if [ "$found_location_path" = "$LOCATION_PATH" ]; then
                                location_found="$line_num"
                                break
                            fi
                        fi
                    done

                    if [ -n "$location_found" ]; then
                        location_start_line="$location_found"

                        # 找到location块结束
                        local loc_brace_count=0
                        local in_location_block=false

                        for ((line_num=location_start_line; line_num<=server_end_line; line_num++)); do
                            local line=$(sed -n "${line_num}p" "$NGINX_CONF_PATH")

                            if [[ $line =~ location[[:space:]]+(/[^[:space:]]*)[[:space:]]*\{ ]]; then
                                in_location_block=true
                                loc_brace_count=1
                            elif [[ $line =~ \{ ]] && [ "$in_location_block" = true ]; then
                                ((loc_brace_count++))
                            elif [[ $line =~ \} ]] && [ "$in_location_block" = true ]; then
                                ((loc_brace_count--))
                                if [ $loc_brace_count -eq 0 ]; then
                                    location_end_line=$line_num
                                    break
                                fi
                            fi
                        done

                        has_matching_location=true
                    fi

                    # 找到匹配的server块，跳出循环
                    break
                else
                    # 重置变量，继续查找下一个server块
                    server_start_line=0
                    server_end_line=0
                    server_port=""
                fi
            fi
        done <<< "$server_blocks"
    fi

    # 只输出变量，不输出调试信息
    echo "has_matching_server=$has_matching_server"
    echo "has_matching_location=$has_matching_location"
    echo "server_start_line=$server_start_line"
    echo "server_end_line=$server_end_line"
    echo "location_start_line=$location_start_line"
    echo "location_end_line=$location_end_line"
    echo "server_port=$server_port"
}

# 6. 智能部署
smart_deploy() {
    local check_output=$(check_config_exists "$DOMAIN" "$PORT" "$LOCATION_PATH")

    # 解析检查结果
    local has_matching_server=$(echo "$check_output" | grep "^has_matching_server=" | cut -d= -f2)
    local has_matching_location=$(echo "$check_output" | grep "^has_matching_location=" | cut -d= -f2)
    local server_start_line=$(echo "$check_output" | grep "^server_start_line=" | cut -d= -f2)
    local server_end_line=$(echo "$check_output" | grep "^server_end_line=" | cut -d= -f2)
    local location_start_line=$(echo "$check_output" | grep "^location_start_line=" | cut -d= -f2)
    local location_end_line=$(echo "$check_output" | grep "^location_end_line=" | cut -d= -f2)

    # 确保变量是数字
    server_start_line=${server_start_line:-0}
    server_end_line=${server_end_line:-0}
    location_start_line=${location_start_line:-0}
    location_end_line=${location_end_line:-0}

    # 决策逻辑：根据检查结果决定部署策略
    local deploy_strategy=""

    if [ "$has_matching_server" = "true" ] && [ "$has_matching_location" = "true" ]; then
        # 情况1：完全重复，跳过部署
        echo "🎯 配置已存在 ($DOMAIN:$PORT$LOCATION_PATH)，跳过部署"
        deploy_strategy="skip_deployment"

    elif [ "$has_matching_server" = "true" ] && [ "$has_matching_location" = "false" ]; then
        # 情况2：有匹配的server块，但没有匹配的location，在现有server块中添加location
        echo "➕ 在现有server中添加location: $LOCATION_PATH"
        deploy_strategy="add_location"
        if ! add_location "$server_start_line" "$server_end_line"; then
            echo "❌ 添加location失败"
            return 1
        fi

    elif [ "$has_matching_server" = "false" ]; then
        # 情况3：没有匹配的server块，创建全新的server块
        echo "🆕 创建新server块"
        deploy_strategy="create_server"
        if ! create_server; then
            echo "❌ 创建server块失败"
            return 1
        fi

    else
        # 情况4：其他情况（理论上不应该到达这里）
        echo "⚠️  创建新server块"
        deploy_strategy="create_server"
        if ! create_server; then
            echo "❌ 创建server块失败"
            return 1
        fi
    fi

    # 将部署策略保存到全局变量，供主流程使用
    DEPLOY_STRATEGY="$deploy_strategy"
}

# 7. 添加location（如果不存在的话）
add_location() {
    local start_line="$1"
    local end_line="$2"
    local temp_file="/tmp/nginx_temp_$$"

    # 检查是否已经存在相同的location路径
    # 逐行检查，匹配location后面去掉空格第一个是/的
    local existing_location=""

    for ((line_num=start_line; line_num<=end_line; line_num++)); do
        local line=$(sed -n "${line_num}p" "$NGINX_CONF_PATH")

        # 匹配location行：location后面去掉空格第一个字符是/
        if [[ $line =~ location[[:space:]]+(/[^[:space:]]*)[[:space:]]*\{ ]]; then
            local found_path="${BASH_REMATCH[1]}"
            # 去掉路径中的空格
            found_path=$(echo "$found_path" | tr -d ' ')

            # 检查是否匹配目标路径
            if [ "$found_path" = "$LOCATION_PATH" ]; then
                existing_location="$line_num"
                break
            fi
        fi
    done

    if [ -n "$existing_location" ]; then
        return 0
    else
        # 创建临时文件：在server块结束前添加location
        head -n $((end_line - 1)) "$NGINX_CONF_PATH" > "$temp_file"
        echo "" >> "$temp_file"
        generate_location_config >> "$temp_file"
        echo "" >> "$temp_file"
        echo "    }" >> "$temp_file"
        tail -n +$((end_line + 1)) "$NGINX_CONF_PATH" >> "$temp_file"
    fi

    # 验证临时文件
    if [ ! -f "$temp_file" ]; then
        echo "❌ 临时文件创建失败"
        return 1
    fi

    # 直接替换原文件（已有BACKUP_ROOT下的完整备份）
    mv "$temp_file" "$NGINX_CONF_PATH"
}

# 8. 创建server
create_server() {
    # 查找http块结束位置
    local http_start_line=$(grep -n "^[[:space:]]*http[[:space:]]*{" "$NGINX_CONF_PATH" | head -1 | cut -d: -f1)

    if [ -z "$http_start_line" ]; then
        echo "❌ 无法找到http块开始位置"
        return 1
    fi

    # 查找最后一个}（通常是http块的结束）
    local last_brace_line=$(grep -n "}" "$NGINX_CONF_PATH" | tail -1 | cut -d: -f1)

    if [ -z "$last_brace_line" ]; then
        echo "❌ 无法找到任何结束括号"
        return 1
    fi

    # 验证这个}是否是http块的结束
    local brace_count=0
    local http_end_line=0

    # 从http块开始位置向后扫描
    for ((line_num=http_start_line; line_num<=last_brace_line; line_num++)); do
        local line=$(sed -n "${line_num}p" "$NGINX_CONF_PATH")

        if [[ $line =~ \{ ]]; then
            ((brace_count++))
        elif [[ $line =~ \} ]]; then
            ((brace_count--))
            if [ $brace_count -eq 0 ]; then
                http_end_line=$line_num
                break
            fi
        fi
    done

    if [ $http_end_line -eq 0 ]; then
        echo "⚠️  使用最后一个}作为http块结束位置"
        http_end_line=$last_brace_line
    fi

    # 在http块结束前插入新server块
    local temp_file="/tmp/nginx_temp_$$"

    # 创建临时文件：http块内容 + 新server块 + 结束括号
    head -n $((http_end_line - 1)) "$NGINX_CONF_PATH" > "$temp_file"
    echo "" >> "$temp_file"
    generate_server_config >> "$temp_file"
    echo "" >> "$temp_file"
    echo "}" >> "$temp_file"

    # 验证临时文件
    if [ ! -f "$temp_file" ]; then
        echo "❌ 临时文件创建失败"
        return 1
    fi

    # 替换原文件
    mv "$temp_file" "$NGINX_CONF_PATH"
}

# 9. 主流程
main() {
    echo "══════════════════════════════════════════════════════════════════════════════════"
    echo "🚀 NGINX配置部署流程"
    echo "══════════════════════════════════════════════════════════════════════════════════"

    # 阶段1：配置分析与策略决策
    echo "  📊 阶段1：配置分析与策略决策"
    if ! smart_deploy; then
        echo "❌ nginx配置部署失败，开始回滚"
        rollback_configs
        exit 1
    fi

    # 获取部署策略
    deploy_strategy="$DEPLOY_STRATEGY"

    # 检查是否跳过部署
    if [ "$deploy_strategy" = "skip_deployment" ]; then
        echo "✅ 部署完成（无需操作）"
        exit 0
    fi

    # 阶段2：配置验证
    echo "  🔍 阶段2：配置验证"

    # 创建临时日志目录避免权限问题（使用绝对路径）
    local temp_log_dir="$NGINX_LOGS_PATH/temp_nginx_logs"
    mkdir -p "$temp_log_dir"

    # 验证nginx配置（Linux使用标准nginx -t命令）
    if ! "$NGINX_BIN_PATH" -t 2>/dev/null; then
        echo "❌ nginx配置验证失败，开始回滚..."
        rollback_configs
        rm -rf "$temp_log_dir"
        exit 1
    fi
    echo "✅ nginx配置验证通过"

    # 清理临时日志目录
    rm -rf "$temp_log_dir"

    # 阶段3：配置重载
    echo "  🔄 阶段3：配置重载"

    # Linux使用nginx -s reload命令，更通用和可靠
    if ! "$NGINX_BIN_PATH" -s reload 2>/dev/null; then
        echo "❌ nginx配置重载失败，开始回滚"
        rollback_configs
        exit 1
    fi
    echo "✅ nginx配置重载成功"

    echo "══════════════════════════════════════════════════════════════════════════════════"
    echo "🌐 DNS配置流程"
    echo "══════════════════════════════════════════════════════════════════════════════════"

    # 阶段1：DNS配置执行
    echo "  🔧 阶段1：DNS配置执行"
    local ssl_enabled=$(determine_ssl_status)

    # 调用DNS配置脚本 - 只有在nginx部署完全成功后才配置DNS
    if [ -f "$DNS_SCRIPT_PATH" ]; then
        if ! "$DNS_SCRIPT_PATH" "$DOMAIN" "$ssl_enabled" "$deploy_strategy"; then
            echo "❌ DNS配置失败，开始回滚"
            rollback_configs
            exit 1
        fi
        echo "✅ DNS配置完成"
        # 标记DNS配置成功
        touch "$backup_dir/dns_configured"
    else
        echo "⚠️  DNS配置脚本不存在: $DNS_SCRIPT_PATH"
    fi

    echo "══════════════════════════════════════════════════════════════════════════════════"
    echo "✅ 部署完成"
    echo "══════════════════════════════════════════════════════════════════════════════════"

    # 最终部署状态检查
    if check_deployment_status; then
        echo "🌐 访问地址: http://$DOMAIN$LOCATION_PATH"
        local ssl_enabled=$(determine_ssl_status)
        if [ "$ssl_enabled" = "true" ]; then
            echo "🔒 HTTPS访问: https://$DOMAIN$LOCATION_PATH"
        fi
        echo "📁 备份位置: $backup_dir"

        # 清理临时文件和旧备份
        cleanup_temp_files
        cleanup_old_backups
    else
        echo "❌ 部署状态检查失败，开始回滚"
        rollback_configs
        exit 1
    fi
}

# 10. 回滚配置
rollback_configs() {
    echo "🔄 开始回滚配置..."
    echo "📁 回滚源: $backup_dir"

    # 回滚DNS配置（只有在DNS配置成功后才需要回滚）
    # 检查是否存在DNS配置成功的标记文件
    local dns_configured_marker="$backup_dir/dns_configured"
    if [ -f "$dns_configured_marker" ] && [ -f "$DNS_SCRIPT_PATH" ]; then
        echo "🔄 回滚DNS配置..."
        # 调用DNS脚本的清理功能（传递特殊参数表示清理）
        if ! "$DNS_SCRIPT_PATH" "$DOMAIN" "cleanup" "rollback"; then
            echo "⚠️  DNS配置回滚失败，但继续执行nginx回滚"
        fi
        # 清理标记文件
        rm -f "$dns_configured_marker"
    else
        echo "ℹ️  跳过DNS回滚（DNS未配置或标记文件不存在）"
    fi

    # 回滚nginx配置
    if [ -f "$backup_dir/nginx.conf" ]; then
        cp "$backup_dir/nginx.conf" "$NGINX_CONF_PATH"
        echo "✅ nginx配置已回滚"
    else
        echo "❌ 未找到nginx备份文件: $backup_dir/nginx.conf"
        echo "💡 请检查备份目录: $backup_dir"
        return 1
    fi

    # 重新加载nginx配置
    echo "🔄 重新加载nginx配置..."

    # Linux使用nginx -s reload命令，更通用和可靠
    if ! "$NGINX_BIN_PATH" -s reload; then
        echo "❌ nginx配置重载失败"
        return 1
    fi
    echo "✅ nginx配置重载成功"

    echo "🔄 回滚完成，请检查服务状态"
}

# 11. 清理临时文件
cleanup_temp_files() {
    # 清理nginx临时文件
    local temp_file="/tmp/nginx_temp_$$"
    if [ -f "$temp_file" ]; then
        rm -f "$temp_file"
    fi

    # 清理其他可能的临时文件
    find /tmp -name "nginx_temp_*" -mtime +1 -delete 2>/dev/null || true
}

# 12. 清理旧备份文件
cleanup_old_backups() {
    local max_backups=5  # 保留最近5个备份
    local backup_root="$BACKUP_ROOT"

    if [ -d "$backup_root" ]; then
        local backup_count=$(find "$backup_root" -maxdepth 1 -type d -name "20*" | wc -l)

        if [ "$backup_count" -gt "$max_backups" ]; then
            local to_delete=$((backup_count - max_backups))
            echo "🧹 清理旧备份文件，删除 $to_delete 个旧备份..."

            # 删除最旧的备份
            find "$backup_root" -maxdepth 1 -type d -name "20*" -printf '%T@ %p\n' | \
            sort -n | head -n "$to_delete" | \
            while read timestamp path; do
                rm -rf "$path"
            done

            echo "✅ 旧备份清理完成"
        fi
    fi
}

# 13. 部署状态检查
check_deployment_status() {
    # 检查配置文件语法
    if "$NGINX_BIN_PATH" -t > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# 14. 参数验证
validate_params() {
    if [ $# -lt 4 ]; then
        echo "❌ 参数不足！需要至少4个参数：域名、端口、根路径、location路径"
        exit 1
    fi

    # 检查location_path不能为空
    if [ -z "$LOCATION_PATH" ]; then
        echo "❌ location路径不能为空！"
        echo "💡 请提供有效的location路径，例如：/app (应用路径) 、/ (根路径)"
        exit 1
    fi

    # 检查端口和SSL配置的一致性
    local port="$PORT"
    local ssl_enabled=$(determine_ssl_status)

    if [ "$ssl_enabled" = "true" ]; then
        # HTTPS端口
        if [ -z "$SSL_CERT_PATH" ] || [ -z "$SSL_KEY_PATH" ]; then
            echo "❌ HTTPS端口 $port 需要SSL证书配置！"
            echo "💡 请提供SSL证书和密钥文件路径"
            echo "   示例: $0 \"$DOMAIN\" \"$PORT\" \"$ROOT_PATH\" \"$LOCATION_PATH\" \"/path/to/cert.pem\" \"/path/to/key.key\""
            exit 1
        fi
    else
        # HTTP端口
        if [ -n "$SSL_CERT_PATH" ] || [ -n "$SSL_KEY_PATH" ]; then
            echo "⚠️  警告：HTTP端口 $port 不需要SSL证书，但提供了证书文件"
            echo "💡 证书文件将被忽略，使用HTTP配置"
            echo "   如需HTTPS，请使用端口 443, 8443, 或 9443"
        fi
    fi

    # 检查根路径是否存在（Linux路径）
    if [ ! -d "$ROOT_PATH" ]; then
        echo "❌ 根路径不存在: $ROOT_PATH"
        echo "💡 请检查路径是否正确，Linux路径示例："
        echo "   - /var/www/html"
        echo "   - /opt/website/dist"
        echo "   - ./dist"
        exit 1
    fi

    # 检查SSL证书文件（如果提供且端口是HTTPS）
    if [ "$ssl_enabled" = "true" ] && [ -n "$SSL_CERT_PATH" ] && [ ! -f "$SSL_CERT_PATH" ]; then
        echo "❌ SSL证书文件不存在: $SSL_CERT_PATH"
        exit 1
    fi

    if [ "$ssl_enabled" = "true" ] && [ -n "$SSL_KEY_PATH" ] && [ ! -f "$SSL_KEY_PATH" ]; then
        echo "❌ SSL密钥文件不存在: $SSL_KEY_PATH"
        exit 1
    fi
}

# 15. Linux环境检测和适配
check_linux_environment() {
    # 检查nginx是否安装
    if [ -f "$NGINX_BIN_PATH" ]; then
        echo "✅ nginx已安装: $NGINX_BIN_PATH"
    elif command -v nginx >/dev/null 2>&1; then
        NGINX_BIN_PATH=$(which nginx)
        echo "✅ nginx已安装: $NGINX_BIN_PATH"
    else
        echo "❌ nginx未安装或路径不正确"
        echo "💡 请安装nginx或修改脚本中的NGINX_BIN_PATH变量"
        exit 1
    fi

    # 检查nginx配置文件
    if [ -f "$NGINX_CONF_PATH" ]; then
        echo "✅ nginx配置文件存在: $NGINX_CONF_PATH"

        # 检查配置文件权限
        if [ -r "$NGINX_CONF_PATH" ]; then
            echo "✅ nginx配置文件可读"
            # 检查写权限，如果没有直接写权限，检查是否可以通过sudo获得权限
            if [ -w "$NGINX_CONF_PATH" ]; then
                echo "✅ nginx配置文件可写"
            else
                echo "⚠️ nginx配置文件需要sudo权限写入"
                # 测试sudo权限
                if sudo -n test -w "$NGINX_CONF_PATH" 2>/dev/null; then
                    echo "✅ 确认可通过sudo写入配置文件"
                else
                    echo "💡 将使用sudo权限进行配置文件操作"
                fi
            fi
        else
            echo "❌ nginx配置文件权限不足，请检查文件权限"
            exit 1
        fi
    else
        echo "❌ nginx配置文件不存在: $NGINX_CONF_PATH"
        echo "💡 请检查路径或修改脚本中的NGINX_CONF_PATH变量"
        exit 1
    fi

    # 检查nginx进程是否运行
    if pgrep nginx > /dev/null 2>&1; then
        echo "✅ nginx进程运行中"
    else
        echo "⚠️  nginx进程未运行，尝试启动..."
        # 直接使用nginx命令启动，保持与reload命令的一致性
        if "$NGINX_BIN_PATH" 2>/dev/null; then
            echo "✅ nginx启动成功"
        else
            echo "❌ nginx启动失败"
            exit 1
        fi
    fi
}

# 16. 使用说明
show_usage() {
    echo "使用方法: $0 <域名> <端口> <根路径> <location路径> [SSL证书路径] [SSL密钥路径]"
    echo ""
    echo "详细使用说明请查看: scripts/README_DNS_CONFIG.md"
    echo ""
    echo "快速示例:"
    echo "  $0 \"localhost\" \"80\" \"/var/www/html\" \"/\""
    echo "  $0 \"test.funshion.tv\" \"443\" \"/opt/website/dist\" \"/\" \"/etc/ssl/cert.pem\" \"/etc/ssl/key.key\""
    echo ""
    echo "注意: location路径不能为空，必须明确指定，例如：/app (应用路径) 、/ (根路径)"
}

# 17. 脚本入口
if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    show_usage
    exit 0
fi

# 检查Linux环境
check_linux_environment

# 验证参数
validate_params "$@"

# 如果根路径不存在，直接报错退出
if [ ! -d "$ROOT_PATH" ]; then
    echo "❌ 根路径不存在: $ROOT_PATH"
    echo "💡 请先创建目录或提供正确的路径"
    exit 1
fi

# 执行主流程
main
