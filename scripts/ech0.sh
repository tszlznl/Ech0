#!/bin/bash

RED='\033[1;31m'
GREEN='\033[1;32m'
YELLOW='\033[1;33m'
RESET='\033[0m'

REPO="${ECH0_GITHUB_REPO:-lin-snow/Ech0}"
SERVICE_NAME="ech0"
INSTALL_PATH_DEFAULT="/opt/ech0"
TMP_DIR="/tmp/ech0-install"
DOWNLOAD_FILE="/tmp/ech0.tar.gz"
MANAGER_PATH="/usr/local/sbin/em"
COMMAND_LINK="/usr/local/bin/em"
COMMAND_LINK_COMPAT="/usr/local/bin/ech0"
LEGACY_MANAGER_PATH="/usr/local/sbin/ech0-manager"

ARCH=""
LATEST_TAG=""

handle_error() {
  local code=$1
  shift
  echo -e "${RED}错误: $*${RESET}"
  exit "$code"
}

log_info() {
  echo -e "${GREEN}$*${RESET}"
}

log_warn() {
  echo -e "${YELLOW}$*${RESET}"
}

ensure_command() {
  command -v "$1" >/dev/null 2>&1 || handle_error 1 "缺少依赖: $1"
}

ensure_root() {
  if [ "$(id -u)" -ne 0 ]; then
    handle_error 1 "请使用 root 权限运行此命令"
  fi
}

detect_arch() {
  local machine
  if command -v arch >/dev/null 2>&1; then
    machine=$(arch)
  else
    machine=$(uname -m)
  fi

  case "$machine" in
    x86_64|amd64)
      ARCH="amd64"
      ;;
    aarch64|arm64)
      ARCH="arm64"
      ;;
    *)
      handle_error 1 "当前架构(${machine})不支持，仅支持 amd64 与 arm64"
      ;;
  esac
}

resolve_install_path() {
  local target="${1:-$INSTALL_PATH_DEFAULT}"
  target="${target%/}"
  if [[ $target != */ech0 ]]; then
    target="$target/ech0"
  fi
  echo "$target"
}

get_installed_path() {
  if [ -f "/etc/systemd/system/${SERVICE_NAME}.service" ]; then
    local service_path
    service_path=$(grep -m1 '^WorkingDirectory=' \
      "/etc/systemd/system/${SERVICE_NAME}.service" | cut -d'=' -f2)
    if [ -n "$service_path" ] && [ -f "$service_path/ech0" ]; then
      echo "$service_path"
      return 0
    fi
  fi

  if [ -f "$INSTALL_PATH_DEFAULT/ech0" ]; then
    echo "$INSTALL_PATH_DEFAULT"
    return 0
  fi

  echo "$INSTALL_PATH_DEFAULT"
  return 1
}

is_ech0_installed() {
  local installed_path
  installed_path=$(get_installed_path)
  [ -f "$installed_path/ech0" ]
}

cleanup_tmp() {
  rm -rf "$TMP_DIR"
  rm -f "$DOWNLOAD_FILE"
}

trap cleanup_tmp EXIT

resolve_latest_tag() {
  local effective
  effective=$(curl -sIL -o /dev/null -w '%{url_effective}' \
    "https://github.com/${REPO}/releases/latest")
  if [ -n "$effective" ]; then
    LATEST_TAG=$(basename "$effective")
  else
    LATEST_TAG="latest"
  fi
}

download_package() {
  detect_arch
  resolve_latest_tag
  local asset="ech0-linux-${ARCH}.tar.gz"
  local url="https://github.com/${REPO}/releases/latest/download/${asset}"

  log_info "最新版本: ${LATEST_TAG}"
  log_info "下载 ${asset}"

  if ! curl -fL --retry 3 --retry-delay 3 --connect-timeout 10 \
    "$url" -o "$DOWNLOAD_FILE"; then
    handle_error 1 "下载 Ech0 失败，请检查网络连接"
  fi
}

prepare_install_dir() {
  local install_path=$1
  local parent
  parent=$(dirname "$install_path")

  mkdir -p "$parent" || handle_error 1 "无法创建目录: $parent"

  if [ -f "$install_path/ech0" ]; then
    handle_error 1 "目标位置已存在 Ech0，请使用 update"
  fi

  mkdir -p "$install_path" || handle_error 1 "无法创建安装目录: $install_path"
}

extract_package() {
  local install_path=$1
  local package_dir
  package_dir=$(extract_release_package_dir)

  cp -a "$package_dir"/. "$install_path"/ || handle_error 1 "复制文件失败"

  local arch_binary="$install_path/ech0-linux-${ARCH}"
  if [ -f "$arch_binary" ]; then
    mv "$arch_binary" "$install_path/ech0" || handle_error 1 "重命名二进制失败"
  fi

  chmod +x "$install_path/ech0" || handle_error 1 "设置执行权限失败"
}

extract_release_package_dir() {
  rm -rf "$TMP_DIR"
  mkdir -p "$TMP_DIR"

  if ! tar -xzf "$DOWNLOAD_FILE" -C "$TMP_DIR"; then
    handle_error 1 "解压失败"
  fi

  local binary_path
  binary_path=$(find "$TMP_DIR" -mindepth 1 -maxdepth 5 -type f \
    \( -name "ech0" -o -name "ech0-linux-${ARCH}" \) -print -quit)
  if [ -z "$binary_path" ]; then
    handle_error 1 "未找到 Ech0 二进制文件"
  fi

  dirname "$binary_path"
}

ensure_service() {
  local install_path=$1
  cat <<EOF >/etc/systemd/system/${SERVICE_NAME}.service
[Unit]
Description=Ech0 Service
After=network.target
Wants=network.target

[Service]
Type=simple
WorkingDirectory=${install_path}
ExecStart=${install_path}/ech0 serve
Restart=on-failure
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

  systemctl daemon-reload
  systemctl enable "${SERVICE_NAME}" >/dev/null 2>&1
}

install_cli_helper() {
  local script_path
  if command -v readlink >/dev/null 2>&1; then
    script_path=$(readlink -f "$0")
  else
    script_path="$0"
  fi

  if [ ! -f "$script_path" ]; then
    log_warn "未找到脚本本身，跳过管理命令安装"
    return
  fi

  mkdir -p "$(dirname "$MANAGER_PATH")"
  cp "$script_path" "$MANAGER_PATH" || log_warn "复制管理脚本失败"
  chmod 755 "$MANAGER_PATH" 2>/dev/null || true

  mkdir -p "$(dirname "$COMMAND_LINK")"
  ln -sf "$MANAGER_PATH" "$COMMAND_LINK" || log_warn "创建命令链接失败"
  ln -sf "$MANAGER_PATH" "$COMMAND_LINK_COMPAT" || log_warn "创建兼容命令链接失败"
}

should_install_cli_helper() {
  if [ ! -t 0 ]; then
    return 0
  fi

  local choice
  read -r -p "是否安装管理命令(em，并兼容 ech0)? [Y/n]: " choice
  case "$choice" in
    n|N)
      return 1
      ;;
    *)
      return 0
      ;;
  esac
}

install_ech0() {
  ensure_root
  ensure_command curl
  ensure_command tar
  ensure_command systemctl

  local installed_path
  installed_path=$(get_installed_path)
  if [ -f "$installed_path/ech0" ]; then
    handle_error 1 "检测到 Ech0 已安装在 ${installed_path}，请使用 update 或先 uninstall"
  fi

  local target_path
  target_path=$(resolve_install_path "$1")

  prepare_install_dir "$target_path"
  download_package
  extract_package "$target_path"
  ensure_service "$target_path"

  local cli_installed="false"
  if should_install_cli_helper; then
    install_cli_helper
    cli_installed="true"
  else
    log_warn "已跳过管理命令安装，可稍后手动执行安装脚本来安装"
  fi

  systemctl restart "$SERVICE_NAME"

  log_info "Ech0 安装完成"
  echo "安装路径: $target_path"
  echo "服务命令: systemctl {start|stop|restart} ${SERVICE_NAME}"
  if [ "$cli_installed" = "true" ]; then
    echo "管理命令: ${COMMAND_LINK} (兼容命令: ${COMMAND_LINK_COMPAT})"
  fi
}

update_ech0() {
  ensure_root
  ensure_command curl
  ensure_command tar
  ensure_command systemctl

  local target_path
  target_path=$(get_installed_path)

  if [ ! -f "$target_path/ech0" ]; then
    handle_error 1 "未检测到已安装的 Ech0"
  fi

  download_package

  local package_dir
  package_dir=$(extract_release_package_dir)

  systemctl stop "$SERVICE_NAME" >/dev/null 2>&1 || true

  if [ -f "$target_path/ech0" ]; then
    cp "$target_path/ech0" "$target_path/ech0.bak" || true
  fi

  shopt -s dotglob nullglob
  for item in "$package_dir"/*; do
    local name
    name=$(basename "$item")
    if [ "$name" = "data" ]; then
      continue
    fi
    rm -rf "$target_path/$name"
    cp -a "$item" "$target_path/" || handle_error 1 "更新 $name 失败"
  done
  shopt -u dotglob nullglob

  local arch_binary="$target_path/ech0-linux-${ARCH}"
  if [ -f "$arch_binary" ]; then
    mv "$arch_binary" "$target_path/ech0" || handle_error 1 "重命名二进制失败"
  fi

  chmod +x "$target_path/ech0" || handle_error 1 "设置执行权限失败"
  rm -f "$target_path/ech0.bak"

  systemctl restart "$SERVICE_NAME"
  log_info "Ech0 更新完成"
}

uninstall_ech0() {
  ensure_root
  ensure_command systemctl

  local target_path
  target_path=$(get_installed_path)

  if [ ! -f "$target_path/ech0" ]; then
    handle_error 1 "未检测到 Ech0 安装"
  fi

  read -r -p "确认卸载 Ech0? [y/N]: " choice
  case "$choice" in
    y|Y)
      local delete_data_choice
      local remove_data="false"
      local data_dir="$target_path/data"
      if [ -d "$data_dir" ]; then
        read -r -p "检测到数据目录 ${data_dir}，是否同时删除(删除后不可恢复)? [y/N]: " delete_data_choice
        case "$delete_data_choice" in
          y|Y)
            remove_data="true"
            ;;
        esac
      fi

      systemctl stop "$SERVICE_NAME" >/dev/null 2>&1 || true
      systemctl disable "$SERVICE_NAME" >/dev/null 2>&1 || true
      rm -f "/etc/systemd/system/${SERVICE_NAME}.service"
      systemctl daemon-reload

      if [ "$remove_data" = "true" ]; then
        rm -rf "$target_path"
        log_info "已删除程序与数据目录"
      else
        shopt -s dotglob nullglob
        for item in "$target_path"/*; do
          local name
          name=$(basename "$item")
          if [ "$name" = "data" ]; then
            continue
          fi
          rm -rf "$item"
        done
        shopt -u dotglob nullglob
        log_warn "已保留数据目录: ${data_dir}"
      fi

      rm -f "$MANAGER_PATH" "$COMMAND_LINK" "$COMMAND_LINK_COMPAT" "$LEGACY_MANAGER_PATH"

      log_info "Ech0 已卸载"
      ;;
    *)
      log_warn "已取消卸载"
      ;;
  esac
}

show_status() {
  ensure_command systemctl
  if systemctl is-active "$SERVICE_NAME" >/dev/null 2>&1; then
    log_info "Ech0 服务运行中"
  else
    echo -e "${RED}Ech0 服务未运行${RESET}"
  fi
}

run_cli_command() {
  local subcommand=$1
  local target_path
  target_path=$(get_installed_path)
  if [ ! -f "$target_path/ech0" ]; then
    handle_error 1 "未检测到 Ech0 安装"
  fi

  (cd "$target_path" && "$target_path/ech0" "$subcommand")
}

start_service() {
  ensure_root
  ensure_command systemctl
  systemctl start "$SERVICE_NAME"
  log_info "Ech0 已启动"
}

stop_service() {
  ensure_root
  ensure_command systemctl
  systemctl stop "$SERVICE_NAME"
  log_info "Ech0 已停止"
}

restart_service() {
  ensure_root
  ensure_command systemctl
  systemctl restart "$SERVICE_NAME"
  log_info "Ech0 已重启"
}

show_menu() {
  local install_label="1) 安装 Ech0"
  if is_ech0_installed; then
    install_label="1) 安装 Ech0 (已安装)"
  fi

  echo
  echo "欢迎使用 Ech0 部署脚本"
  echo "$install_label"
  echo "2) 更新 Ech0"
  echo "3) 删除 Ech0"
  echo "4) 查看服务状态"
  echo "5) 查看当前信息"
  echo "6) 启动服务"
  echo "7) 停止服务"
  echo "8) 重启服务"
  echo "0) 退出脚本"
  echo
  read -r -p "请选择 [0-8]: " choice

  case "$choice" in
    1)
      if is_ech0_installed; then
        local installed_path
        installed_path=$(get_installed_path)
        log_warn "已检测到安装: ${installed_path}，请使用更新或先卸载"
        return
      fi
      read -r -p "安装路径(默认: ${INSTALL_PATH_DEFAULT}): " path_choice
      install_ech0 "$path_choice"
      ;;
    2)
      update_ech0
      ;;
    3)
      uninstall_ech0
      ;;
    4)
      show_status
      ;;
    5)
      run_cli_command info
      ;;
    6)
      start_service
      ;;
    7)
      stop_service
      ;;
    8)
      restart_service
      ;;
    0)
      exit 0
      ;;
    *)
      echo -e "${RED}无效的选项${RESET}"
      ;;
  esac
}

main() {
  case "$1" in
    install)
      install_ech0 "$2"
      ;;
    update)
      update_ech0
      ;;
    uninstall)
      uninstall_ech0
      ;;
    status)
      show_status
      ;;
    info)
      run_cli_command info
      ;;
    start)
      start_service
      ;;
    tui)
      run_cli_command tui
      ;;
    stop)
      stop_service
      ;;
    restart)
      restart_service
      ;;
    "")
      while true; do
        show_menu
      done
      ;;
    *)
      cat <<EOF
用法: $0 [命令]
  install [路径]   安装 Ech0 (默认安装到 ${INSTALL_PATH_DEFAULT})
  update          更新到最新版本
  uninstall       删除 Ech0 及 systemd 服务
  status          查看服务状态
  info            执行 "ech0 info"
  start           启动服务
  stop            停止服务
  restart         重启服务
  (无参数)        进入交互菜单
EOF
      ;;
  esac
}

main "$@"
