#!/bin/bash

# 钛盘上传工具 Linux 安装脚本
# TmpLink Uploader Linux Installation Script

set -e

INSTALL_DIR="/usr/local/bin"
GITHUB_REPO="tmplink/tmplink_uploader"
API_BASE="https://api.github.com/repos/$GITHUB_REPO"
DOWNLOAD_BASE="https://github.com/$GITHUB_REPO/releases/download"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}=====================================${NC}"
    echo -e "${BLUE}   钛盘上传工具 Linux 安装程序      ${NC}"
    echo -e "${BLUE}   TmpLink Uploader Linux Installer ${NC}"
    echo -e "${BLUE}=====================================${NC}"
    echo
}

print_step() {
    echo -e "${YELLOW}[步骤] $1${NC}"
}

print_success() {
    echo -e "${GREEN}[成功] $1${NC}"
}

print_error() {
    echo -e "${RED}[错误] $1${NC}"
}

print_info() {
    echo -e "${BLUE}[信息] $1${NC}"
}

check_requirements() {
    print_step "检查系统要求..."
    
    # 检查是否为 Linux
    if [[ "$OSTYPE" != "linux-gnu"* ]]; then
        print_error "此脚本仅适用于 Linux 系统"
        exit 1
    fi
    
    # 检查发行版
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        print_info "检测到 Linux 发行版: $NAME $VERSION"
    else
        print_info "未知 Linux 发行版"
    fi
    
    # 检查是否有写入权限（或者可以 sudo）
    if [[ ! -w "$INSTALL_DIR" ]] && ! sudo -n true 2>/dev/null; then
        print_info "需要管理员权限来安装到 $INSTALL_DIR"
        echo -n "是否继续？ [y/N]: "
        read -r response
        if [[ ! "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
            print_info "安装已取消"
            exit 0
        fi
    fi
    
    # 检查必要的工具
    local missing_tools=()
    for tool in curl wget; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        fi
    done
    
    if [[ ${#missing_tools[@]} -eq 2 ]]; then
        print_error "需要 curl 或 wget 来下载文件"
        print_info "请使用包管理器安装其中一个工具"
        exit 1
    fi
    
    print_success "系统要求检查通过"
}

detect_architecture() {
    print_step "检测系统架构..."
    
    local arch=$(uname -m)
    case $arch in
        x86_64)
            ARCH_SUFFIX="linux-amd64"
            print_info "检测到 64位 x86 架构"
            ;;
        i386|i686)
            ARCH_SUFFIX="linux-386"
            print_info "检测到 32位 x86 架构"
            ;;
        aarch64|arm64)
            ARCH_SUFFIX="linux-arm64"
            print_info "检测到 ARM64 架构"
            ;;
        arm*)
            # ARM32 不在当前构建目标中，但可以尝试 ARM64
            ARCH_SUFFIX="linux-arm64"
            print_info "检测到 ARM 架构，将使用 ARM64 版本"
            ;;
        *)
            print_error "不支持的架构: $arch"
            print_info "支持的架构: x86_64, i386/i686, aarch64/arm64"
            exit 1
            ;;
    esac
}

get_latest_version() {
    print_step "获取最新版本信息..."
    
    local version_cmd=""
    if command -v curl &> /dev/null; then
        version_cmd="curl -fsSL"
    elif command -v wget &> /dev/null; then
        version_cmd="wget -q -O -"
    else
        print_error "找不到 curl 或 wget"
        exit 1
    fi
    
    # 获取最新 release 信息
    local release_info
    if ! release_info=$($version_cmd "$API_BASE/releases/latest"); then
        print_error "获取版本信息失败"
        exit 1
    fi
    
    # 解析版本号 (提取 tag_name 字段)
    LATEST_VERSION=$(echo "$release_info" | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
    
    if [[ -z "$LATEST_VERSION" ]]; then
        print_error "解析版本信息失败"
        exit 1
    fi
    
    print_info "最新版本: $LATEST_VERSION"
}

download_binaries() {
    print_step "下载二进制文件..."
    
    local temp_dir=$(mktemp -d)
    local gui_binary="tmplink"
    local cli_binary="tmplink-cli"
    local gui_remote="tmplink-$ARCH_SUFFIX"
    local cli_remote="tmplink-cli-$ARCH_SUFFIX"
    
    # 选择下载工具
    local download_cmd=""
    if command -v curl &> /dev/null; then
        download_cmd="curl -fsSL"
    elif command -v wget &> /dev/null; then
        download_cmd="wget -q -O -"
    else
        print_error "找不到 curl 或 wget"
        exit 1
    fi
    
    print_info "下载 $gui_binary..."
    if ! $download_cmd "$DOWNLOAD_BASE/$LATEST_VERSION/$gui_remote" > "$temp_dir/$gui_binary"; then
        print_error "下载 $gui_binary 失败"
        rm -rf "$temp_dir"
        exit 1
    fi
    
    print_info "下载 $cli_binary..."
    if ! $download_cmd "$DOWNLOAD_BASE/$LATEST_VERSION/$cli_remote" > "$temp_dir/$cli_binary"; then
        print_error "下载 $cli_binary 失败"
        rm -rf "$temp_dir"
        exit 1
    fi
    
    # 验证文件
    if [[ ! -s "$temp_dir/$gui_binary" ]] || [[ ! -s "$temp_dir/$cli_binary" ]]; then
        print_error "下载的文件无效"
        rm -rf "$temp_dir"
        exit 1
    fi
    
    TEMP_DIR="$temp_dir"
    print_success "二进制文件下载完成"
}

install_binaries() {
    print_step "安装程序到系统..."
    
    local gui_binary="tmplink"
    local cli_binary="tmplink-cli"
    
    # 安装二进制文件
    print_info "安装 $gui_binary 到 $INSTALL_DIR..."
    if [[ -w "$INSTALL_DIR" ]]; then
        cp "$TEMP_DIR/$gui_binary" "$INSTALL_DIR/"
        cp "$TEMP_DIR/$cli_binary" "$INSTALL_DIR/"
    else
        sudo cp "$TEMP_DIR/$gui_binary" "$INSTALL_DIR/"
        sudo cp "$TEMP_DIR/$cli_binary" "$INSTALL_DIR/"
    fi
    
    # 设置执行权限
    if [[ -w "$INSTALL_DIR" ]]; then
        chmod +x "$INSTALL_DIR/$gui_binary"
        chmod +x "$INSTALL_DIR/$cli_binary"
    else
        sudo chmod +x "$INSTALL_DIR/$gui_binary"
        sudo chmod +x "$INSTALL_DIR/$cli_binary"
    fi
    
    print_success "程序安装完成"
}

create_desktop_entry() {
    print_step "创建桌面快捷方式..."
    
    local desktop_dir="$HOME/.local/share/applications"
    local desktop_file="$desktop_dir/tmplink.desktop"
    
    # 创建桌面应用目录
    mkdir -p "$desktop_dir"
    
    # 创建桌面文件
    cat > "$desktop_file" << EOF
[Desktop Entry]
Name=TmpLink Uploader
Name[zh_CN]=钛盘上传工具
Comment=Upload files to TmpLink
Comment[zh_CN]=上传文件到钛盘
Exec=$INSTALL_DIR/tmplink
Icon=cloud-upload
Terminal=false
Type=Application
Categories=Network;FileTransfer;
StartupNotify=true
EOF
    
    chmod +x "$desktop_file"
    
    print_info "桌面快捷方式已创建: $desktop_file"
    print_success "桌面快捷方式创建完成"
}

verify_installation() {
    print_step "验证安装..."
    
    if command -v tmplink &> /dev/null && command -v tmplink-cli &> /dev/null; then
        print_success "安装验证成功"
        print_info "GUI 程序: $(which tmplink)"
        print_info "CLI 程序: $(which tmplink-cli)"
    else
        print_error "安装验证失败，请检查 PATH 环境变量"
        print_info "您可能需要重新加载 shell 配置或重新登录"
        exit 1
    fi
}

setup_environment() {
    print_step "配置环境..."
    
    # 检查 PATH 是否包含 /usr/local/bin
    if [[ ":$PATH:" != *":/usr/local/bin:"* ]]; then
        print_info "添加 /usr/local/bin 到 PATH"
        
        # 根据不同的 shell 添加到配置文件
        local shell_config=""
        case "$SHELL" in
            */bash)
                shell_config="$HOME/.bashrc"
                ;;
            */zsh)
                shell_config="$HOME/.zshrc"
                ;;
            */fish)
                shell_config="$HOME/.config/fish/config.fish"
                ;;
            *)
                shell_config="$HOME/.profile"
                ;;
        esac
        
        if [[ -n "$shell_config" ]]; then
            if [[ "$SHELL" == *"fish" ]]; then
                echo 'set -gx PATH /usr/local/bin $PATH' >> "$shell_config"
            else
                echo 'export PATH="/usr/local/bin:$PATH"' >> "$shell_config"
            fi
            print_info "已添加 PATH 配置到 $shell_config"
            print_info "请运行 'source $shell_config' 或重新打开终端"
        fi
    fi
    
    print_success "环境配置完成"
}

show_usage() {
    print_step "使用说明"
    echo
    echo "安装完成！您现在可以使用以下命令："
    echo
    echo "  tmplink      - 启动图形界面版本"
    echo "  tmplink-cli  - 使用命令行版本"
    echo
    echo "获取帮助："
    echo "  tmplink --help"
    echo "  tmplink-cli --help"
    echo
    echo "配置文件位置："
    echo "  ~/.tmplink_config.json"
    echo
    echo "桌面快捷方式位置："
    echo "  ~/.local/share/applications/tmplink.desktop"
    echo
    echo "更多信息请查看："
    echo "  https://github.com/$GITHUB_REPO"
    echo
    
    # 显示特定发行版的额外信息
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        case "$ID" in
            ubuntu|debian)
                echo "Ubuntu/Debian 用户提示："
                echo "  如果在图形界面中无法找到应用，请运行："
                echo "  update-desktop-database ~/.local/share/applications/"
                ;;
            centos|rhel|fedora)
                echo "CentOS/RHEL/Fedora 用户提示："
                echo "  如果遇到权限问题，可能需要配置 SELinux："
                echo "  sudo setsebool -P allow_execmem 1"
                ;;
        esac
        echo
    fi
}

cleanup() {
    print_step "清理临时文件..."
    if [[ -n "$TEMP_DIR" ]] && [[ -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
    print_success "清理完成"
}

main() {
    print_header
    
    check_requirements
    detect_architecture
    get_latest_version
    download_binaries
    install_binaries
    create_desktop_entry
    setup_environment
    verify_installation
    show_usage
    cleanup
    
    echo
    print_success "钛盘上传工具安装完成！"
    echo
}

# 捕获中断信号
trap 'print_error "安装被中断"; cleanup; exit 1' INT TERM

# 运行主函数
main "$@"