#!/usr/bin/env bash

# =========================================================
# V2bX 管理脚本 v1.0 (Sing-Box 1.14+ 专属重构版)
# 快速安装命令 (GitHub 官方直链):
# bash <(curl -fsSL "https://raw.githubusercontent.com/zwgundam/V2bX-Patched-By-Haman/main/v2bx.sh")
# =========================================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
RESET='\033[0m'

# 直链物理下载根路径 (直接从 GitHub Releases 下载预编译二进制与最新脚本)
BASE_URL="https://github.com/zwgundam/V2bX-Patched-By-Haman/releases/latest/download"
RAW_SCRIPT_URL="https://raw.githubusercontent.com/zwgundam/V2bX-Patched-By-Haman/main/v2bx.sh"
INSTALL_DIR="/usr/local/bin"
BINARY_PATH="${INSTALL_DIR}/V2bX"
SCRIPT_PATH="${INSTALL_DIR}/v2bx"
CONFIG_DIR="/etc/V2bX"
CONFIG_FILE="${CONFIG_DIR}/config.json"
SING_ORIGIN_FILE="${CONFIG_DIR}/sing_origin.json"
SERVICE_FILE="/etc/systemd/system/V2bX.service"

# 检查是否为 Root 用户
check_root() {
    if [[ $EUID -ne 0 ]]; then
        echo -e "${RED}错误：必须使用 root 权限运行此脚本！${RESET}"
        exit 1
    fi
}

# 检测系统架构
detect_arch() {
    local arch=$(uname -m)
    case "${arch}" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            echo -e "${RED}不支持的系统架构: ${arch}${RESET}"
            exit 1
            ;;
    esac
}

# 物理检查运行状态
get_status() {
    if [[ -f "${BINARY_PATH}" ]]; then
        IS_INSTALLED=true
    else
        IS_INSTALLED=false
    fi

    IS_RUNNING=false
    IS_ENABLED=false

    for svc in V2bX.service v2bx.service V2bX v2bx; do
        if systemctl is-active --quiet "${svc}" 2>/dev/null; then
            IS_RUNNING=true
            break
        fi
    done

    if pgrep -f "V2bX" >/dev/null 2>&1 || pgrep -f "v2bx server" >/dev/null 2>&1; then
        IS_RUNNING=true
    fi

    for svc in V2bX.service v2bx.service V2bX v2bx; do
        if systemctl is-enabled --quiet "${svc}" 2>/dev/null; then
            IS_ENABLED=true
            break
        fi
    done
}

# 强制覆盖生成 Sing-Box 1.14+ 规范模板文件
generate_default_config() {
    mkdir -p "${CONFIG_DIR}/cert"
    echo -e "${YELLOW}正在覆盖生成最新 Sing-Box 1.14+ 规则文件 -> ${SING_ORIGIN_FILE}${RESET}"
    cat <<'EOF_ORIGIN' > "${SING_ORIGIN_FILE}"
{
  "dns": {
    "servers": [
      {
        "tag": "cf",
        "type": "udp",
        "server": "1.1.1.1"
      }
    ],
    "strategy": "prefer_ipv4"
  },
  "outbounds": [
    {
      "tag": "direct",
      "type": "direct"
    },
    {
      "type": "block",
      "tag": "block"
    }
  ],
  "route": {
    "rules": [
      {
        "ip_is_private": true,
        "outbound": "block"
      },
      {
        "outbound": "direct",
        "network": [
          "udp",
          "tcp"
        ]
      }
    ]
  },
  "experimental": {
    "cache_file": {
      "enabled": true
    }
  }
}
EOF_ORIGIN

    if [[ ! -f "${CONFIG_FILE}" ]]; then
        echo -e "${YELLOW}配置文件不存在，生成初始模板 -> ${CONFIG_FILE}${RESET}"
        cat <<'EOF_CONF' > "${CONFIG_FILE}"
{
  "Log": {
    "Level": "info",
    "Output": ""
  },
  "Cores": [
    {
      "Type": "sing",
      "Log": {
        "Level": "info",
        "Timestamp": true
      },
      "NTP": {
        "Enable": false,
        "Server": "time.apple.com",
        "ServerPort": 0
      },
      "OriginalPath": "/etc/V2bX/sing_origin.json"
    }
  ],
  "Machines": [
    {
      "ApiHost": "https://your-panel.com",
      "ApiKey": "***",
      "MachineID": 1
    }
  ]
}
EOF_CONF
    fi
}

# 捕获 Ctrl+C 安全退出日志流
tail_service_log() {
    trap 'echo -e "\n${YELLOW}已退出日志流。${RESET}"; return 0 2>/dev/null || exit 0' INT
    journalctl -u V2bX -f -n 50
    trap - INT
}

# 启动并追踪日志
start_and_tail_log() {
    systemctl start V2bX && echo -e "${GREEN}V2bX 启动成功！正在载入实时日志...${RESET}"
    tail_service_log
}

# 重启并追踪日志
restart_and_tail_log() {
    systemctl restart V2bX && echo -e "${GREEN}V2bX 重启成功！正在载入实时日志...${RESET}"
    tail_service_log
}

# 交互式配置 V2bX
configure_v2bx() {
    check_root
    echo -e "${PURPLE}=================================================${RESET}"
    echo -e "${PURPLE}          V2bX 交互式面板对接配置向导          ${RESET}"
    echo -e "${PURPLE}=================================================${RESET}"

    local current_host="https://your-panel.com"
    local current_key="***"
    local current_id=1

    if [[ -f "${CONFIG_FILE}" ]]; then
        local existing_host=$(grep -oP '"ApiHost":\s*"\K[^"]+' "${CONFIG_FILE}" 2>/dev/null || echo "")
        local existing_key=$(grep -oP '"ApiKey":\s*"\K[^"]+' "${CONFIG_FILE}" 2>/dev/null || echo "")
        local existing_id=$(grep -oP '"MachineID":\s*\K[0-9]+' "${CONFIG_FILE}" 2>/dev/null || echo "")

        [[ -n "${existing_host}" ]] && current_host="${existing_host}"
        [[ -n "${existing_key}" ]] && current_key="${existing_key}"
        [[ -n "${existing_id}" ]] && current_id="${existing_id}"
    fi

    echo -e "当前配置 ApiHost: ${CYAN}${current_host}${RESET}"
    read -p "请输入面板 ApiHost [默认: ${current_host}]: " input_host
    input_host=${input_host:-$current_host}

    echo -e "当前配置 ApiKey: ${CYAN}${current_key}${RESET}"
    read -p "请输入面板 ApiKey / 通讯密钥 [默认: ${current_key}]: " input_key
    input_key=${input_key:-$current_key}

    echo -e "当前配置 MachineID: ${CYAN}${current_id}${RESET}"
    read -p "请输入 MachineID / 节点 ID [默认: ${current_id}]: " input_id
    input_id=${input_id:-$current_id}

    mkdir -p "${CONFIG_DIR}/cert"

    cat <<EOF_CFG > "${CONFIG_FILE}"
{
  "Log": {
    "Level": "info",
    "Output": ""
  },
  "Cores": [
    {
      "Type": "sing",
      "Log": {
        "Level": "info",
        "Timestamp": true
      },
      "NTP": {
        "Enable": false,
        "Server": "time.apple.com",
        "ServerPort": 0
      },
      "OriginalPath": "/etc/V2bX/sing_origin.json"
    }
  ],
  "Machines": [
    {
      "ApiHost": "${input_host}",
      "ApiKey": "${input_key}",
      "MachineID": ${input_id}
    }
  ]
}
EOF_CFG

    echo -e "${GREEN}=================================================${RESET}"
    echo -e "${GREEN} 配置文件写入完成 -> ${CONFIG_FILE}${RESET}"
    echo -e "${GREEN}=================================================${RESET}"

    read -p "是否现在重启 V2bX 服务以应用新配置？[Y/n]: " restart_choice
    if [[ "${restart_choice}" != "n" && "${restart_choice}" != "N" ]]; then
        restart_and_tail_log
    fi
}

# 安装 V2bX
install_v2bx() {
    check_root
    detect_arch
    echo -e "${GREEN}>>> 正在准备安装 V2bX (Sing-Box 1.14+ 重构版) [架构: ${ARCH}]...${RESET}"

    mkdir -p "${CONFIG_DIR}/cert"
    mkdir -p "${INSTALL_DIR}"

    local bin_url="${BASE_URL}/v2bx-linux-${ARCH}"
    echo -e "${CYAN}正在下载物理二进制: ${bin_url}${RESET}"
    curl -fsSL -o "${BINARY_PATH}" "${bin_url}" || wget -qO "${BINARY_PATH}" "${bin_url}"

    if [[ ! -f "${BINARY_PATH}" ]]; then
        echo -e "${RED}错误：下载二进制文件失败！请检查网络连通性。${RESET}"
        exit 1
    fi

    local file_size=$(wc -c < "${BINARY_PATH}" 2>/dev/null || echo 0)
    if [[ ${file_size} -lt 1048576 ]]; then
        echo -e "${RED}错误：下载到的二进制文件异常 (大小为 ${file_size} 字节，小于 1MB)！${RESET}"
        rm -f "${BINARY_PATH}"
        exit 1
    fi

    chmod +x "${BINARY_PATH}"

    echo -e "${CYAN}正在注册全局管理命令 'v2bx' 至 ${SCRIPT_PATH}...${RESET}"
    if [[ -f "$0" && "$0" != "bash" && "$0" != "/bin/bash" && "$0" != *"subshell"* ]]; then
        cp -f "$0" "${SCRIPT_PATH}"
    else
        curl -fsSL -o "${SCRIPT_PATH}" "${RAW_SCRIPT_URL}" || wget -qO "${SCRIPT_PATH}" "${RAW_SCRIPT_URL}"
    fi
    chmod +x "${SCRIPT_PATH}"

    generate_default_config

    cat <<EOF_SVC > "${SERVICE_FILE}"
[Unit]
Description=V2bX Service (Sing-Box 1.14+ Custom)
After=network.target nss-lookup.target
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=${CONFIG_DIR}
ExecStart=${BINARY_PATH} server -c ${CONFIG_FILE}
Restart=on-failure
RestartSec=5s
LimitNOFILE=512000

[Install]
WantedBy=multi-user.target
EOF_SVC

    systemctl daemon-reload
    systemctl enable V2bX >/dev/null 2>&1
    
    echo -e "${GREEN}=================================================${RESET}"
    echo -e "${GREEN} V2bX 安装成功！物理命令 'v2bx' 已注册至环境变量。${RESET}"
    echo -e "${GREEN} 配置文件路径: ${CONFIG_FILE}${RESET}"
    echo -e "${GREEN} 任意目录下输入 'v2bx' 即可唤出图形管理菜单！${RESET}"
    echo -e "${GREEN}=================================================${RESET}"
    
    configure_v2bx
}

# 更新 V2bX (内核 + 脚本 + 规则三合一更新)
update_v2bx() {
    check_root
    detect_arch
    echo -e "${CYAN}>>> 正在全量更新 V2bX...${RESET}"
    systemctl stop V2bX >/dev/null 2>&1

    # 1. 更新物理二进制
    local bin_url="${BASE_URL}/v2bx-linux-${ARCH}"
    echo -e "${CYAN}正在更新二进制文件: ${bin_url}${RESET}"
    curl -fsSL -o "${BINARY_PATH}" "${bin_url}" || wget -qO "${BINARY_PATH}" "${bin_url}"
    chmod +x "${BINARY_PATH}"

    # 2. 安全更新管理脚本
    local tmp_script=$(mktemp /tmp/v2bx_update.XXXXXX)
    echo -e "${CYAN}正在获取最新管理脚本: ${RAW_SCRIPT_URL}${RESET}"
    curl -fsSL -o "${tmp_script}" "${RAW_SCRIPT_URL}" || wget -qO "${tmp_script}" "${RAW_SCRIPT_URL}"

    local tmp_size=$(wc -c < "${tmp_script}" 2>/dev/null || echo 0)
    if [[ ${tmp_size} -lt 8000 ]]; then
        echo -e "${RED}错误：下载脚本不完整 (大小仅 ${tmp_size} 字节)，跳过覆盖！${RESET}"
        rm -f "${tmp_script}"
        return 1
    fi
    if ! head -n 1 "${tmp_script}" | grep -q "^#\!"; then
        echo -e "${RED}错误：下载内容缺少 Shebang 头 (#!)，跳过覆盖！${RESET}"
        rm -f "${tmp_script}"
        return 1
    fi
    if ! bash -n "${tmp_script}" 2>/dev/null; then
        echo -e "${RED}错误：下载脚本语法校验失败，跳过覆盖！${RESET}"
        rm -f "${tmp_script}"
        return 1
    fi
    if ! grep -q "V2bX 管理脚本" "${tmp_script}"; then
        echo -e "${RED}错误：下载内容不是合法的 v2bx 脚本 (特征标识缺失)，跳过覆盖！${RESET}"
        rm -f "${tmp_script}"
        return 1
    fi

    chmod +x "${tmp_script}"
    mv -f "${tmp_script}" "${SCRIPT_PATH}"

    # 3. 同步更新 Sing-Box 1.14+ 规则
    generate_default_config

    echo -e "${GREEN}=================================================${RESET}"
    echo -e "${GREEN} V2bX 内核、管理脚本及 Sing-Box 规则已全量更新！${RESET}"
    echo -e "${GREEN} 正在重启服务并调取最新运行日志...${RESET}"
    echo -e "${GREEN}=================================================${RESET}"
    
    exec "${SCRIPT_PATH}" restart
}

# 卸载 V2bX
uninstall_v2bx() {
    check_root
    echo -e "${RED}>>> 警告：正在物理卸载 V2bX...${RESET}"
    systemctl stop V2bX >/dev/null 2>&1
    systemctl disable V2bX >/dev/null 2>&1
    rm -f "${SERVICE_FILE}"
    systemctl daemon-reload
    rm -f "${BINARY_PATH}"
    rm -f "${SCRIPT_PATH}"
    read -p "是否保留配置文件目录 (${CONFIG_DIR})? [Y/n]: " keep_conf
    if [[ "${keep_conf}" == "n" || "${keep_conf}" == "N" ]]; then
        rm -rf "${CONFIG_DIR}"
        echo -e "${YELLOW}配置文件已被物理清理。${RESET}"
    fi
    echo -e "${GREEN}V2bX 卸载完成！${RESET}"
}

# 显示主菜单
show_menu() {
    get_status
    echo -e "${GREEN}=================================================${RESET}"
    echo -e "${GREEN}   V2bX 管理脚本 (Sing-Box 1.14+ 专属重构版)    ${RESET}"
    echo -e "${GREEN}=================================================${RESET}"

    if [[ "${IS_INSTALLED}" == "true" ]]; then
        echo -ne " 安装状态: ${GREEN}已安装${RESET}"
        if [[ "${IS_RUNNING}" == "true" ]]; then
            echo -e " | 运行状态: ${GREEN}运行中${RESET}"
        else
            echo -e " | 运行状态: ${RED}已停止${RESET}"
        fi

        if [[ "${IS_ENABLED}" == "true" ]]; then
            echo -e " 开机自启: ${GREEN}已开启${RESET}"
        else
            echo -e " 开机自启: ${YELLOW}未开启${RESET}"
        fi
    else
        echo -e " 安装状态: ${RED}未安装${RESET} | 运行状态: ${RED}无服务${RESET} | 开机自启: ${RED}未配置${RESET}"
    fi

    echo -e "-------------------------------------------------"
    echo -e " ${CYAN}0.${RESET} 退出脚本"
    echo -e "-------------------------------------------------"
    echo -e " ${CYAN}1.${RESET} 安装 V2bX"
    echo -e " ${CYAN}2.${RESET} 更新 V2bX"
    echo -e " ${CYAN}3.${RESET} 卸载 V2bX"
    echo -e "-------------------------------------------------"
    echo -e " ${CYAN}4.${RESET} 启动 V2bX"
    echo -e " ${CYAN}5.${RESET} 停止 V2bX"
    echo -e " ${CYAN}6.${RESET} 重启 V2bX"
    echo -e " ${CYAN}7.${RESET} 查看 V2bX 运行状态"
    echo -e " ${CYAN}8.${RESET} 查看 V2bX 实时日志"
    echo -e "-------------------------------------------------"
    echo -e " ${CYAN}9.${RESET} 交互式配置 V2bX (ApiHost/ApiKey/MachineID)"
    echo -e " ${CYAN}10.${RESET} 设置/取消开机自启"
    echo -e "${GREEN}=================================================${RESET}"

    read -p " 请选择操作 [0-10]: " choice
    case "${choice}" in
        0)
            exit 0
            ;;
        1)
            install_v2bx
            ;;
        2)
            update_v2bx
            ;;
        3)
            uninstall_v2bx
            ;;
        4)
            start_and_tail_log
            ;;
        5)
            systemctl stop V2bX && echo -e "${YELLOW}V2bX 已停止！${RESET}"
            ;;
        6)
            restart_and_tail_log
            ;;
        7)
            systemctl status V2bX
            ;;
        8)
            tail_service_log
            ;;
        9)
            configure_v2bx
            ;;
        10)
            if [[ "${IS_ENABLED}" == "true" ]]; then
                systemctl disable V2bX && echo -e "${YELLOW}开机自启已取消！${RESET}"
            else
                systemctl enable V2bX && echo -e "${GREEN}开机自启已开启！${RESET}"
            fi
            ;;
        *)
            echo -e "${RED}请输入有效选项 [0-10]！${RESET}"
            ;;
    esac
}

# 命令行 CLI 入口逻辑
if [[ $# -gt 0 ]]; then
    case "$1" in
        install)
            install_v2bx
            ;;
        update)
            update_v2bx
            ;;
        uninstall)
            uninstall_v2bx
            ;;
        start)
            start_and_tail_log
            ;;
        stop)
            systemctl stop V2bX && echo -e "${YELLOW}V2bX 已停止！${RESET}"
            ;;
        restart)
            restart_and_tail_log
            ;;
        status)
            systemctl status V2bX
            ;;
        log)
            tail_service_log
            ;;
        config)
            configure_v2bx
            ;;
        *)
            echo -e "用法: $0 {install|update|uninstall|start|stop|restart|status|log|config}"
            exit 1
            ;;
    esac
else
    get_status
    if [[ "${IS_INSTALLED}" == "false" ]]; then
        echo -e "${YELLOW}=================================================${RESET}"
        echo -e "${YELLOW} 检测到当前系统尚未安装 V2bX 服务。${RESET}"
        echo -e "${YELLOW}=================================================${RESET}"
        read -p "是否立即开始安装 V2bX？[Y/n]: " auto_install_choice
        if [[ "${auto_install_choice}" != "n" && "${auto_install_choice}" != "N" ]]; then
            install_v2bx
            exit 0
        fi
    fi
    show_menu
fi
