#!/bin/bash

# 钛盘上传工具 macOS 安装脚本
# TmpLink Uploader macOS Installation Script

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="/usr/local/bin"
BUILD_DIR="$SCRIPT_DIR/build"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}=====================================${NC}"
    echo -e "${BLUE}   钛盘上传工具 macOS 安装程序      ${NC}"
    echo -e "${BLUE}   TmpLink Uploader macOS Installer ${NC}"
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
    
    # 检查是否为 macOS
    if [[ "$OSTYPE" != "darwin"* ]]; then
        print_error "此脚本仅适用于 macOS 系统"
        exit 1
    fi
    
    # 检查 Go 是否安装
    if ! command -v go &> /dev/null; then
        print_error "Go 未安装，请先安装 Go (https://golang.org/dl/)"
        exit 1
    fi
    
    print_info "Go 版本: $(go version)"
    
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
    
    print_success "系统要求检查通过"
}

detect_architecture() {
    print_step "检测系统架构..."
    
    local arch=$(uname -m)
    case $arch in
        x86_64)
            ARCH_DIR="macos-intel"
            print_info "检测到 Intel Mac (x86_64)"
            ;;
        arm64)
            ARCH_DIR="macos-arm64"
            print_info "检测到 Apple Silicon Mac (ARM64)"
            ;;
        *)
            print_error "不支持的架构: $arch"
            exit 1
            ;;
    esac
}

build_binaries() {
    print_step "构建程序..."
    
    cd "$SCRIPT_DIR"
    
    # 安装依赖
    print_info "安装 Go 依赖..."
    go mod download
    go mod tidy
    
    # 构建发布版本
    print_info "构建 macOS 版本..."
    make build-release
    
    if [[ ! -d "$BUILD_DIR/$ARCH_DIR" ]]; then
        print_error "构建失败，找不到目标目录: $BUILD_DIR/$ARCH_DIR"
        exit 1
    fi
    
    print_success "程序构建完成"
}

install_binaries() {
    print_step "安装程序到系统..."
    
    local source_dir="$BUILD_DIR/$ARCH_DIR"
    local gui_binary="tmplink"
    local cli_binary="tmplink-cli"
    
    # 检查二进制文件是否存在
    if [[ ! -f "$source_dir/$gui_binary" ]] || [[ ! -f "$source_dir/$cli_binary" ]]; then
        print_error "二进制文件不存在，请检查构建过程"
        exit 1
    fi
    
    # 安装二进制文件
    print_info "复制 $gui_binary 到 $INSTALL_DIR..."
    if [[ -w "$INSTALL_DIR" ]]; then
        cp "$source_dir/$gui_binary" "$INSTALL_DIR/"
        cp "$source_dir/$cli_binary" "$INSTALL_DIR/"
    else
        sudo cp "$source_dir/$gui_binary" "$INSTALL_DIR/"
        sudo cp "$source_dir/$cli_binary" "$INSTALL_DIR/"
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

verify_installation() {
    print_step "验证安装..."
    
    if command -v tmplink &> /dev/null && command -v tmplink-cli &> /dev/null; then
        print_success "安装验证成功"
        print_info "GUI 程序版本: $(tmplink --version 2>/dev/null || echo '可用')"
        print_info "CLI 程序版本: $(tmplink-cli --version 2>/dev/null || echo '可用')"
    else
        print_error "安装验证失败，请检查 PATH 环境变量"
        exit 1
    fi
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
    echo "更多信息请查看："
    echo "  README.md"
    echo "  docs/usage.md"
    echo
}

cleanup() {
    print_step "清理临时文件..."
    # 这里可以添加清理逻辑，如果需要的话
    print_success "清理完成"
}

main() {
    print_header
    
    check_requirements
    detect_architecture
    build_binaries
    install_binaries
    verify_installation
    show_usage
    cleanup
    
    echo
    print_success "钛盘上传工具安装完成！"
    echo
}

# 捕获中断信号
trap 'print_error "安装被中断"; exit 1' INT TERM

# 运行主函数
main "$@"