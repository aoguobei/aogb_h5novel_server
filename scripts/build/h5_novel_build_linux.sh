#!/bin/bash

# H5小说项目构建脚本 - Linux专用版本
# 基于Jenkins流水线: h5_novel_pipline 和 h5_novel_pipline_local
# 专为Linux环境优化，移除Windows和macOS相关代码
# 集成SSH部署功能，无需外部脚本依赖
#
# 依赖工具:
#   - Node.js (v20.18.1)
#   - yarn
#   - git
#   - sshpass (用于SSH自动认证)
#     安装命令: sudo apt-get install sshpass (Ubuntu/Debian)
#               sudo yum install sshpass (CentOS/RHEL)

set -e  # 遇到错误立即退出

# 默认配置
DEFAULT_BRANCH="uni/funNovel/devNew"
DEFAULT_VERSION="1.0.0"
DEFAULT_ENV="master"
DEFAULT_PROJECTS="tth5-xingchen,ksh5-xingchen,tth5-qudu"

# Git仓库配置
GIT_REPO="***"
GIT_REPO_PUBLISH="***"

# SSH部署配置 - 从环境变量获取，提供默认值
SSH_HOST="${SSH_HOST:-***}"
SSH_USER="${SSH_USER:-fun}"
SSH_PASSWORD="${SSH_PASSWORD:-}"
REMOTE_BASE_PATH="/opt/website"

# Node环境配置
NODE_HOME="/home/fun/.nvm/versions/node/v20.18.1/bin"
export PATH="${NODE_HOME}:${PATH}"

# 获取脚本所在目录的绝对路径
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

# 工作目录
WORKSPACE="${PROJECT_ROOT}/workspace"
GIT_PROJECT_DIR="${WORKSPACE}/funNovel"
GIT_PROJECT_DIR_PUBLISH="${WORKSPACE}/publish"

# 日志函数
write_log() {
    local file=$1
    local type=$2
    local proj=$3
    local env=$4
    local msg=$5

    touch "${WORKSPACE}/${file}.txt"
    echo "${type}: ${env}, ${proj}, ${msg}" >> "${WORKSPACE}/${file}.txt"
}

# 简化日志函数 - 合并重复代码
log_msg() {
    local level="$1" msg="$2" icon="$3"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    local output="[${timestamp}] ${icon} ${msg}"

    case "$level" in
        "error") echo "$output" >&2; echo "$output" >> "${WORKSPACE}/error.log" ;;
        *) echo "$output"; echo "$output" >> "${WORKSPACE}/realtime.log" ;;
    esac
}

# 简化日志调用
log_output() { log_msg "info" "$1" ""; }
log_error() { log_msg "error" "$1" "❌ ERROR:"; }
log_success() { log_msg "info" "$1" "✅"; }
log_warning() { log_msg "info" "$1" "⚠️ WARNING:"; }

# SSH部署相关日志函数
log_info() {
    local message="$1"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[${timestamp}] ℹ️ ${message}"
}

# 获取推送分支
get_push_branch() {
    local website=$1
    local platform=$2
    local env=$3

    case "${website}" in
        "xingchen")
            case "${platform}" in
                "tt")
                    case "${env}" in
                        "master") echo "master" ;;
                        "release") echo "releaase" ;;
                        *) echo "${env}_${platform}_${website}" ;;
                    esac
                    ;;
                "ks")
                    case "${env}" in
                        "master") echo "master_ks" ;;
                        "release") echo "release_ks" ;;
                        *) echo "${env}_${platform}_${website}" ;;
                    esac
                    ;;
                *)
                    echo "${env}_${platform}_${website}"
                    ;;
            esac
            ;;
        *)
            echo "${env}_${platform}_${website}"
            ;;
    esac
}

# 检测操作系统类型（Linux专用）
detect_current_os() {
    log_output " 检测当前操作系统类型..."
    OS_TYPE="linux"
    log_output "✅ 检测到操作系统: ${OS_TYPE}"
}

# 简化压缩包创建
create_archive() {
    local proj=$1
    log_output "📦 创建压缩包: ${proj}"

    command -v zip >/dev/null || { log_error "未找到zip命令"; return 1; }
    (cd "$proj" && zip -q -r "../${proj}.zip" .) || { log_error "zip包创建失败"; return 1; }
    rm -rf "$proj"
    log_success "zip包创建成功"
}

# 简化的SSH连接设置
setup_ssh_connection() {
    # 检查SSH密码是否提供
    if [ -z "${SSH_PASSWORD}" ]; then
        log_error "SSH密码未设置，请通过环境变量SSH_PASSWORD提供"
        log_info "示例: export SSH_PASSWORD='your_password'"
        return 1
    fi

    # 检查sshpass工具
    if ! command -v sshpass >/dev/null 2>&1; then
        log_error "未安装sshpass工具，无法进行密码认证"
        log_info "请安装sshpass: sudo apt-get install sshpass"
        return 1
    fi

    # 检查ssh和scp命令是否可用
    if ! command -v ssh >/dev/null 2>&1 || ! command -v scp >/dev/null 2>&1; then
        log_error "未找到ssh或scp命令，请安装openssh-client"
        log_info "Ubuntu/Debian: sudo apt-get install openssh-client"
        return 1
    fi

    # 测试sshpass连接
    log_info "测试SSH连接..."
    if sshpass -p "${SSH_PASSWORD}" ssh -o StrictHostKeyChecking=no -o ConnectTimeout=5 -o UserKnownHostsFile=/dev/null "${SSH_USER}@${SSH_HOST}" "echo 'SSH连接测试成功'" >/dev/null 2>&1; then
        export SSHPASS="${SSH_PASSWORD}"
        SSH_CMD="ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
        SCP_CMD="scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
        USE_SSHPASS=true
        log_success "SSH连接测试成功，使用密码认证"
        return 0
    else
        log_error "SSH连接测试失败，无法连接到服务器"
        log_error "请检查SSH_HOST、SSH_USER、SSH_PASSWORD环境变量是否正确"
        return 1
    fi
}

# 简化的SSH命令执行
execute_ssh() {
    local command="$1"
    sshpass -p "${SSH_PASSWORD}" ${SSH_CMD} "${SSH_USER}@${SSH_HOST}" "$command"
}

# 简化的SCP命令执行
execute_scp() {
    local source="$1"
    local destination="$2"
    sshpass -p "${SSH_PASSWORD}" ${SCP_CMD} "$source" "$destination"
}

# 上传zip文件并部署
upload_and_deploy() {
    local project_name=$1
    local zip_file="${WORKSPACE}/dist_backup/${project_name}.zip"
    local zip_filename="${project_name}.zip"
    local remote_project_path="${REMOTE_BASE_PATH}/${project_name}"
    local remote_dist_path="${remote_project_path}/dist"

    # 检查本地zip文件
    if [ ! -f "$zip_file" ]; then
        log_error "本地zip文件不存在: $zip_file"
        return 1
    fi

    log_info "准备上传文件: $zip_file"

    # Step 1: 上传zip文件
    log_info "上传zip文件到远程服务器..."
    if execute_scp "${zip_file}" "${SSH_USER}@${SSH_HOST}:${REMOTE_BASE_PATH}/${zip_filename}"; then
        log_success "zip文件上传完成"
    else
        log_error "zip文件上传失败"
        return 1
    fi

    # Step 2: 解压文件到dist目录
    log_info "在远程服务器解压文件到dist目录..."
    local extract_script=$(cat << EOF_REMOTE_SCRIPT
        # 进入远程基础目录
        cd "${REMOTE_BASE_PATH}" || { echo '❌ 无法进入远程基础目录'; exit 1; }

        # 确保基础目录权限正确
        sudo chmod -R 755 "${REMOTE_BASE_PATH}" 2>/dev/null || true
        sudo chown -R "${SSH_USER}":"${SSH_USER}" "${REMOTE_BASE_PATH}" 2>/dev/null || true

        # 删除旧的项目目录（如果存在）
        if [ -d "${project_name}" ]; then
            sudo rm -rf "${project_name}" 2>/dev/null || true
        fi

        # 创建项目目录
        mkdir -p "${project_name}" || { echo '❌ 无法创建项目目录'; exit 1; }
        sudo chown -R "${SSH_USER}":"${SSH_USER}" "${project_name}" 2>/dev/null || true
        chmod -R 755 "${project_name}" 2>/dev/null || true

        # 进入项目目录
        cd "${project_name}" || { echo '❌ 无法进入项目目录'; exit 1; }

        # 清空并创建最终的dist目录
        if [ -d 'dist' ]; then
            sudo rm -rf 'dist' 2>/dev/null || true
        fi
        mkdir -p dist || { echo '❌ 无法创建dist目录'; exit 1; }
        chmod 755 dist 2>/dev/null || true

        # 检查zip文件
        if [ ! -f "../${zip_filename}" ]; then
            echo "❌ zip文件不存在: ../${zip_filename}" && exit 1
        fi

        zip_size=\$(stat -c%s "../${zip_filename}" 2>/dev/null || echo '0')
        if [ "\$zip_size" -lt "100" ]; then
            echo '❌ zip文件太小，可能是空文件或损坏' && exit 1
        fi

        # 尝试解压（使用-q参数静默解压，避免打印inflating信息）到dist目录
        if unzip -q -o '../${zip_filename}' -d dist 2>&1 | tee unzip_output.log; then
            echo '✅ zip文件解压成功'
        else
            echo '❌ zip文件解压失败，显示错误详情:'
            cat unzip_output.log 2>/dev/null || echo '无法读取解压日志'
            rm -f unzip_output.log 2>/dev/null || true
            echo '解压失败，终止部署' &&
            exit 1
        fi

        # 清理日志文件
        rm -f unzip_output.log 2>/dev/null || true

        # 修复解压后的权限（不使用sudo）
        chmod -R u+rwx dist 2>/dev/null || chmod -R 755 dist 2>/dev/null || true

        # 进入dist目录并处理文件
        cd dist || { echo '❌ 无法进入dist目录'; exit 1; }

        # 智能检测zip文件结构并处理
        if [ -f 'index.html' ]; then
            # 验证关键文件
            if [ -d 'assets' ] || [ -d 'static' ] || [ -d 'js' ] || [ -d 'css' ]; then
                echo '✅ 发现资源目录，确认为有效的网站文件'
            fi
        else
            echo '❌ 未发现网站文件或项目目录，部署失败'
            echo '当前目录结构:'
            find . -type d 2>/dev/null | head -10 || echo '无法列出目录'
            echo '期望的文件: index.html 或目录: ${project_name}' &&
            exit 1
        fi

        # 清理zip文件
        cd .. || { echo '❌ 无法返回上一级目录'; exit 1; }
        rm -f '../${zip_filename}' || true

        # 验证关键文件是否存在
        if [ -f 'dist/index.html' ]; then
            echo '✅ index.html 存在于dist目录'
        else
            echo '❌ index.html 不存在于dist目录，部署失败'
            echo 'DEBUG: dist目录内容:'
            ls -F dist || true
            exit 1
        fi
        if [ -d 'dist/assets' ]; then
            echo '✅ assets目录存在'
        elif [ -d 'dist/static' ]; then
            echo '✅ static目录存在（可能是assets的替代）'
        else
            echo '❌ assets目录不存在，查找可能的资源目录...'
            find dist/ -type d -name 'assets' -o -name 'static' -o -name 'js' -o -name 'css' 2>/dev/null | head -5 || echo '未找到资源目录'
        fi
        echo '✅ 文件解压和部署完成'
EOF_REMOTE_SCRIPT
)

    if execute_ssh "$extract_script"; then
        log_success "文件解压部署完成"
    else
        log_error "文件解压部署失败"
        return 1
    fi

    # Step 3: 验证部署结果
    log_info "验证部署结果..."
    local file_count=$(execute_ssh "
        if [ -d '${remote_dist_path}' ]; then
            find '${remote_dist_path}' -type f | wc -l
        else
            echo '0'
        fi
    ")

    if [ "${file_count}" -gt "0" ]; then
        log_success "远程部署完成，共部署 ${file_count} 个文件"
        log_success " 部署路径: ${SSH_HOST}:${remote_dist_path}"
        log_success "🌐 访问地址: http://${SSH_HOST}/${project_name}/dist/"
        return 0
    else
        log_error "部署验证失败，目标目录为空"
        return 1
    fi
}

# 帮助信息
show_help() {
    cat << EOF
H5小说项目构建脚本 (Linux专用版本)

用法: $0 [选项]

选项:
  -b, --branch BRANCH           Git分支 (默认: ${DEFAULT_BRANCH})
  -v, --version VERSION         版本号 (默认: ${DEFAULT_VERSION})
  -e, --env ENV                 环境列表，逗号分隔 (默认: ${DEFAULT_ENV})
  -p, --projects PROJECTS       项目列表，逗号分隔 (默认: ${DEFAULT_PROJECTS})
  -f, --force-foreign          强制使用外网套餐 (仅local环境)
  -d, --deploy                 构建后自动部署 (仅local环境)
  -w, --workspace DIR          工作目录 (默认: 当前目录)
  -h, --help                   显示帮助信息

环境选项:
  master   - 主分支环境
  release  - 发布环境
  local    - 远程服务器环境 (***)

项目选项:
  tth5-xingchen, ksh5-xingchen, tth5-fun, tth5-yunyou, tth5-xinyue,
  tth5-qudu, tth5-yuejie, tth5-jinse, tth5-yuexiang, tth5-shuxiang,
  tth5-shiguang, tth5-jiutian, tth5-yunchuan

示例:
  $0 -b uni/funNovel/devNew -v 1.2.0 -e master,release -p tth5-xingchen,ksh5-xingchen
  $0 -e local -f -d -p tth5-qudu  # 构建并自动部署到***
EOF
}

# 解析命令行参数
parse_args() {
    BRANCH="${DEFAULT_BRANCH}"
    VERSION="${DEFAULT_VERSION}"
    ENV="${DEFAULT_ENV}"
    PROJECTS="${DEFAULT_PROJECTS}"
    FORCE_FOREIGN=""
    AUTO_DEPLOY=""

    while [[ $# -gt 0 ]]; do
        case $1 in
            -b|--branch)
                BRANCH="$2"
                shift 2
                ;;
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -e|--env)
                ENV="$2"
                shift 2
                ;;
            -p|--projects)
                PROJECTS="$2"
                shift 2
                ;;
            -f|--force-foreign)
                FORCE_FOREIGN="true"
                shift
                ;;
            -d|--deploy)
                AUTO_DEPLOY="true"
                shift
                ;;
            -w|--workspace)
                WORKSPACE="$(cd "$2" && pwd)"  # 转换为绝对路径
                GIT_PROJECT_DIR="${WORKSPACE}/funNovel"
                GIT_PROJECT_DIR_PUBLISH="${WORKSPACE}/publish"
                shift 2
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                echo "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 简化工作目录初始化
init_workspace() {
    echo " 初始化工作目录..."
    cd "${WORKSPACE}"

    # 清理并创建目录
    rm -rf dist_backup log.txt realtime.log error.log
    mkdir dist_backup

    log_success "工作目录初始化完成"
}

# 简化目录操作
safe_clean() {
    local dir="$1"
    [ -d "$dir" ] && rm -rf "$dir"
    mkdir -p "$dir"
}

# 准备Git仓库
prepare_git_repo() {
    log_output "📦 准备Git仓库..."

    if [ ! -d "${GIT_PROJECT_DIR}" ]; then
        mkdir -p "${GIT_PROJECT_DIR}"
        log_output "📁 创建Git项目目录: ${GIT_PROJECT_DIR}"
    fi

    # 如果.git目录不存在，克隆仓库
    if [ ! -d "${GIT_PROJECT_DIR}/.git" ]; then
        log_output "🔄 克隆仓库: ${GIT_REPO}"
        cd "$(dirname "${GIT_PROJECT_DIR}")"
        rm -rf "${GIT_PROJECT_DIR}"
        if git clone "${GIT_REPO}" "${GIT_PROJECT_DIR}"; then
            log_success "仓库克隆完成"
        else
            log_error "仓库克隆失败"
            return 1
        fi
    else
        log_output " Git仓库已存在，跳过克隆"
    fi

    log_success "Git仓库准备完成"
}

# git_checkout_and_clean函数
git_checkout_and_clean() {
    local branch=$1
    log_output "🔄 Git检出分支: ${branch}"
    cd "${GIT_PROJECT_DIR}"

    # 保存当前工作区修改
    git stash 2>/dev/null || true

    # 获取最新代码
    git fetch origin

    # 切换到目标分支
    git checkout "${branch}" || { log_error "分支切换失败: ${branch}"; return 1; }

    # 拉取目标分支最新代码
    git pull origin "${branch}" || { log_error "分支拉取失败: ${branch}"; return 1; }

    # 强制检出目标分支的所有文件，覆盖工作区修改
    git checkout -f .

    # 再次保存修改（如果有的话）
    git stash 2>/dev/null || true

    # 重置到远程分支，确保工作区完全干净
    git reset --hard "origin/${branch}" || { log_error "分支重置失败: ${branch}"; return 1; }

    # 清理未跟踪的文件和目录
    git clean -df

    # 最后拉取确保最新
    git pull origin "${branch}" 2>/dev/null || true

    log_success "Git检出完成"
}

# 修改配置文件
modify_config() {
    local website=$1
    local platform=$2
    local env=$3
    local version=$4
    local force_foreign=$5

    log_output "⚙️ 修改配置文件: ${website} (${platform}) - ${env}"
    cd "${GIT_PROJECT_DIR}"

    # 修改版本号
    sed -i "s#\"version\": \".*\"#\"version\": \"${version}\"#g" "src/appConfig/baseConfigs/${website}.js"

    if [ "${env}" = "master" ]; then
        # 打开webLogin
        sed -i 's#"webLogin": false#"webLogin": true#' src/appConfig/localConfigs/base.js

        # 测试环境下，默认打开console控制台
        sed -i "s#const vconsole_enabled = ret != '' ? ret : false#const vconsole_enabled = true#g" src/modules/base/antiDebug.js

        # 测试环境下，展示"测试环境"字样
        sed -i 's#"test_enabled": false#"test_enabled": true#' src/appConfig/localConfigs/base.js

        # 测试环境下，展示"测试环境+版本号"
        sed -i "s#<view v-if=\"test_enabled\" class=\"absolute testIcon\">测试环境</view>#<view v-if=\"test_enabled\" class=\"absolute testIcon\">测试环境${version}</view>#" src/pages/readerPage/readerPage.vue
        sed -i "s#<view v-if=\"test_enabled\" class=\"absolute testIcon\">测试环境</view>#<view v-if=\"test_enabled\" class=\"absolute testIcon\">测试环境${version}</view>#" src/pages/userInfo/userInfo.vue

    elif [ "${env}" = "local" ]; then
        # 打开webLogin
        sed -i 's#"webLogin": false#"webLogin": true#' src/appConfig/localConfigs/base.js

        # 测试环境下，默认打开console控制台
        sed -i "s#const vconsole_enabled = ret != '' ? ret : false#const vconsole_enabled = true#g" src/modules/base/antiDebug.js

        # 测试环境下，展示"测试环境"字样
        sed -i 's#"test_enabled": false#"test_enabled": true#' src/appConfig/localConfigs/base.js

        # 公司内网，展示"测试环境+版本号"
        sed -i "s#<view v-if=\"test_enabled\" class=\"absolute testIcon\">测试环境</view>#<view v-if=\"test_enabled\" class=\"absolute testIcon\">公司内网-测试环境${version}</view>#" src/pages/readerPage/readerPage.vue
        sed -i "s#<view v-if=\"test_enabled\" class=\"absolute testIcon\">测试环境</view>#<view v-if=\"test_enabled\" class=\"absolute testIcon\">公司内网-测试环境${version}</view>#" src/pages/userInfo/userInfo.vue

        # 内网环境下，强制使用外网套餐
        if [ "${force_foreign}" = "true" ]; then
            sed -i 's/"force_foreign": false,/"force_foreign": true,/' src/appConfig/localConfigs/base.js
        fi

        # 替换测试策略广告位
        if [ "${platform}" = "tt" ]; then
            sed -i 's/tt_h5_xingchen_business_type/tt_h5_xingchen_product_test/g' "src/appConfig/commonConfigs/${website}.js"
        fi
        if [ "${platform}" = "ks" ]; then
            sed -i 's/ks_h5_xingchen_business_type/tt_h5_xingchen_product_test/g' "src/appConfig/commonConfigs/${website}.js"
        fi
    fi

    # ks平台下，注释掉douyin_open.umd.js文件
    if [ "${platform}" = "ks" ]; then
        sed -i 's#<script src="/douyin_open.umd.js"></script>#<!-- <script src="/douyin_open.umd.js"></script> -->#' index.html
    fi

    echo "✅ 配置修改完成"
}

# 简化构建项目
build_project() {
    local compile_cmd=$1
    log_output "🔨 构建项目: ${compile_cmd}"
    cd "${GIT_PROJECT_DIR}"

    rm -rf dist
    yarn install || { log_error "依赖包安装失败"; return 1; }
    yarn "build:${compile_cmd}" || { log_error "项目构建失败"; return 1; }

    log_success "项目构建完成"
}

# 简化拷贝构建产物
copy_build_artifacts() {
    local proj=$1 website=$2 env=$3 is_local=$4
    local proj_dir=$([ "$is_local" = "true" ] && echo "$proj" || echo "${env}-${proj}")

    log_output " 拷贝构建产物: ${proj_dir}"
    cd "${WORKSPACE}"

    safe_clean "dist_backup/${proj_dir}"

    if [ -d "${GIT_PROJECT_DIR}/dist/build/${website}/h5" ] && [ "$(ls -A "${GIT_PROJECT_DIR}/dist/build/${website}/h5")" ]; then
        cp -rf "${GIT_PROJECT_DIR}/dist/build/${website}/h5/"* "dist_backup/${proj_dir}/"

        if [ "$is_local" = "true" ]; then
            cd dist_backup && create_archive "$proj"
        fi

        log_success "构建产物拷贝完成"
    else
        log_error "构建失败: dist/build/${website}/h5 目录不存在或为空"
        return 1
    fi
}

# 发布到远程仓库
publish_to_remote() {
    local proj=$1
    local website=$2
    local platform=$3
    local env=$4
    local version=$5

    local branch
    branch=$(get_push_branch "${website}" "${platform}" "${env}")
    local proj_dir="${env}-${proj}"
    local h5_dir="dist_backup/${proj_dir}"

    log_output "🚀 发布到远程仓库: ${proj} -> ${branch}"

    if [ ! -d "${h5_dir}" ]; then
        log_error "发布失败: ${h5_dir} 目录不存在"
        write_log 'log' 'Publish' "${proj}" "${env}" "fail"
        return 1
    fi

    # 清理发布目录
    if [ -d "${GIT_PROJECT_DIR_PUBLISH}" ]; then
        rm -rf "${GIT_PROJECT_DIR_PUBLISH}"
    fi
    mkdir -p "${GIT_PROJECT_DIR_PUBLISH}"

    # 克隆发布仓库
    log_output " 克隆发布仓库分支: ${branch}"
    if git clone "${GIT_REPO_PUBLISH}" "${GIT_PROJECT_DIR_PUBLISH}" -b "${branch}" --depth=1; then
        log_success "发布仓库克隆成功"
    else
        log_error "发布仓库克隆失败"
        return 1
    fi

    # 进入发布目录并重置Git状态
    cd "${GIT_PROJECT_DIR_PUBLISH}"
    log_output "⚙️ 重置Git状态"

    # 强制检出当前分支的所有文件
    git checkout -f .

    # 保存所有修改
    git stash 2>/dev/null || true

    # 重置到HEAD
    git reset --hard HEAD

    # 清理未跟踪的文件
    git clean -df

    # 拉取最新代码
    git pull origin "${branch}" 2>/dev/null || true

    # 配置Git用户信息
    log_output "⚙️ 配置Git用户信息"
    git config user.email "aogb@example.com"
    git config user.name "aogb"

    # 拷贝文件到发布仓库
    log_output "📋 拷贝文件到发布仓库"
    rm -rf dist
    mkdir dist
    cp -rf "${WORKSPACE}/dist_backup/${proj_dir}/"* "${GIT_PROJECT_DIR_PUBLISH}/dist/"

    # 提交并推送
    log_output "📤 提交并推送到远程仓库"
    git add -A

    # 检查是否有变更
    if git diff --cached --quiet; then
        log_warning "没有变更需要提交"
    else
        if git commit -m "${version}"; then
            log_success "提交成功"
        else
            log_error "提交失败"
            return 1
        fi
    fi

    if git push origin "${branch}"; then
        log_success "推送成功"
    else
        log_error "推送失败"
        return 1
    fi

    log_success "发布完成"
    write_log 'log' 'Publish' "${proj}" "${env}" "success"
}

# 部署到远程服务器 (local环境指向***)
deploy_to_local() {
    local proj=$1

    log_output " 部署到远程服务器: ***"
    log_output "📂 项目: ${proj}"

    # 确保在正确的工作目录中查找压缩包
    cd "${WORKSPACE}"

    # 查找压缩包文件（只支持zip格式）
    local archive_file=""
    if [ -f "dist_backup/${proj}.zip" ]; then
        archive_file="${WORKSPACE}/dist_backup/${proj}.zip"
        log_output "📦 找到zip压缩包: ${archive_file}"
    else
        log_error "部署失败: 未找到zip压缩包文件"
        log_output "📂 当前工作目录: $(pwd)"
        log_output "📂 查找路径: ${WORKSPACE}/dist_backup/${proj}.zip"
        if [ -d "dist_backup" ]; then
            log_output " dist_backup目录内容:"
            ls -la dist_backup/ | while IFS= read -r line; do
                log_output "   ${line}"
            done
        else
            log_error "dist_backup目录不存在"
        fi
        write_log 'log' 'deploy' "${proj}" "local" "fail"
        return 1
    fi

    # 检查本地文件是否存在
    if [ ! -f "${archive_file}" ]; then
        log_error "本地zip文件不存在: ${archive_file}"
        write_log 'log' 'deploy' "${proj}" "local" "fail"
        return 1
    fi

    # 设置SSH连接
    if ! setup_ssh_connection; then
        log_error "SSH连接设置失败"
        write_log 'log' 'deploy' "${proj}" "local" "fail"
        return 1
    fi

    # 执行部署
    if upload_and_deploy "${proj}"; then
        log_success "项目部署成功"
        write_log 'log' 'deploy' "${proj}" "local" "success"
        return 0
    else
        log_error "项目部署失败"
        write_log 'log' 'deploy' "${proj}" "local" "fail"
        return 1
    fi
}

# 主构建函数 - 修改为Env在外循环
main_build() {
    local env_list=(${ENV//,/ })
    local project_list=(${PROJECTS//,/ })

    log_output "🎯 开始构建 ${#project_list[@]} 个项目，${#env_list[@]} 个环境"
    write_log 'log' '>>>>>>' 'Build Projects sizes' '' "${#project_list[@]}"

    # 准备Git仓库
    prepare_git_repo

    # 环境在外循环，项目在内循环
    for env in "${env_list[@]}"; do
        log_output ""
        log_output "📋 环境: ${env}"

        for proj in "${project_list[@]}"; do
            # 解析项目名称 - 使用 pipeline 的方式，按 h5- 分割
            local arr=(${proj//h5-/ })
            local platform="${arr[0]}"
            local website="${arr[1]}"
            local compile_cmd="${proj}"

            log_output ""
            log_output "🔄 处理项目: ${proj} (platform: ${platform}, website: ${website})"

            # 每个项目都清理工作区
            git_checkout_and_clean "${BRANCH}"

            # 修改配置
            modify_config "${website}" "${platform}" "${env}" "${VERSION}" "${FORCE_FOREIGN}"

            # 构建项目
            if build_project "${compile_cmd}"; then
                # 拷贝构建产物
                local is_local="false"
                if [ "${env}" = "local" ]; then
                    is_local="true"
                fi

                if copy_build_artifacts "${proj}" "${website}" "${env}" "${is_local}"; then
                    write_log 'log' 'build' "${proj}" "${env}" "success"

                    # 如果不是local环境，发布到远程仓库
                    if [ "${env}" != "local" ]; then
                        publish_to_remote "${proj}" "${website}" "${platform}" "${env}" "${VERSION}"
                    elif [ "${AUTO_DEPLOY}" = "true" ]; then
                        # local环境且开启自动部署
                        deploy_to_local "${proj}"
                    fi
                else
                    write_log 'log' 'build' "${proj}" "${env}" "fail"
                fi
            else
                log_error "项目构建失败: ${proj}"
                write_log 'log' 'build' "${proj}" "${env}" "fail"
            fi
        done
    done
}

# 显示构建结果
show_build_results() {
    log_output ""
    log_output "📊 构建结果:"
    log_output "============================================"

    if [ -f "${WORKSPACE}/log.txt" ]; then
        while IFS= read -r line; do
            log_output "${line}"
        done < "${WORKSPACE}/log.txt"
    else
        log_warning "未找到构建日志"
    fi

    log_output ""
    log_output "📁 构建产物目录:"
    if [ -d "${WORKSPACE}/dist_backup" ]; then
        ls -la "${WORKSPACE}/dist_backup" | while IFS= read -r line; do
            log_output "${line}"
        done
    else
        echo "未找到构建产物"
    fi

    if [ -f "${WORKSPACE}/error.log" ]; then
        echo ""
        echo "🔍 错误日志:"
        cat "${WORKSPACE}/error.log"
    fi
}

# 主函数
main() {
    echo "🎉 H5小说项目构建脚本启动 (Linux专用版本 - 集成SSH部署)"
    echo "=========================================="

    # 解析命令行参数
    parse_args "$@"

    # 显示配置信息
    log_output "📋 构建配置:"
    log_output "  分支: ${BRANCH}"
    log_output "  版本: ${VERSION}"
    log_output "  环境: ${ENV}"
    log_output "  项目: ${PROJECTS}"
    log_output "  工作目录: ${WORKSPACE}"
    log_output "  SSH主机: ${SSH_HOST}"
    log_output "  SSH用户: ${SSH_USER}"
    if [ -n "${SSH_PASSWORD}" ]; then
        log_output "  SSH密码: [已设置]"
    else
        log_output "  SSH密码: [未设置]"
    fi
    if [ "${FORCE_FOREIGN}" = "true" ]; then
        log_output "  强制外网套餐: 是"
    fi
    if [ "${AUTO_DEPLOY}" = "true" ]; then
        log_output "  自动部署: 是 (SSH: ${SSH_HOST})"
    fi
    log_output ""

    # 初始化工作目录
    log_output "🚀 开始执行构建脚本..."
    init_workspace

    # 开始构建
    main_build

    # 显示结果
    show_build_results

    log_output ""
    log_success " 构建脚本执行完成！"
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
