#!/bin/bash

# 钛盘上传工具 macOS 安装脚本
# TmpLink Uploader macOS Installation Script

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
    
    # 获取 macOS 版本
    local macos_version=$(sw_vers -productVersion)
    print_info "检测到 macOS 版本: $macos_version"
    
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
    if ! command -v curl &> /dev/null; then
        print_error "需要 curl 来下载文件"
        print_info "curl 应该在 macOS 中预装，请检查系统配置"
        exit 1
    fi
    
    print_success "系统要求检查通过"
}

detect_architecture() {
    print_step "检测系统架构..."
    
    local arch=$(uname -m)
    case $arch in
        x86_64)
            ARCH_SUFFIX="darwin-amd64"
            print_info "检测到 Intel Mac (x86_64)"
            ;;
        arm64)
            ARCH_SUFFIX="darwin-arm64"
            print_info "检测到 Apple Silicon Mac (ARM64)"
            ;;
        *)
            print_error "不支持的架构: $arch"
            print_info "支持的架构: x86_64 (Intel), arm64 (Apple Silicon)"
            exit 1
            ;;
    esac
}

get_latest_version() {
    print_step "获取最新版本信息..."
    
    # 获取最新 release 信息
    local release_info
    if ! release_info=$(curl -fsSL "$API_BASE/releases/latest"); then
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
    
    print_info "下载 $gui_binary..."
    if ! curl -fsSL "$DOWNLOAD_BASE/$LATEST_VERSION/$gui_remote" -o "$temp_dir/$gui_binary"; then
        print_error "下载 $gui_binary 失败"
        rm -rf "$temp_dir"
        exit 1
    fi
    
    print_info "下载 $cli_binary..."
    if ! curl -fsSL "$DOWNLOAD_BASE/$LATEST_VERSION/$cli_remote" -o "$temp_dir/$cli_binary"; then
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

remove_quarantine() {
    print_step "移除 macOS 隔离属性..."
    
    local gui_binary="tmplink"
    local cli_binary="tmplink-cli"
    
    # 移除下载文件的隔离属性
    xattr -d com.apple.quarantine "$TEMP_DIR/$gui_binary" 2>/dev/null || true
    xattr -d com.apple.quarantine "$TEMP_DIR/$cli_binary" 2>/dev/null || true
    
    print_success "隔离属性已移除"
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
    
    # 移除安装后文件的隔离属性
    xattr -d com.apple.quarantine "$INSTALL_DIR/$gui_binary" 2>/dev/null || true
    xattr -d com.apple.quarantine "$INSTALL_DIR/$cli_binary" 2>/dev/null || true
    
    print_success "程序安装完成"
}

create_app_bundle() {
    print_step "创建应用程序包..."
    
    local app_dir="/Applications/TmpLink Uploader.app"
    local contents_dir="$app_dir/Contents"
    local macos_dir="$contents_dir/MacOS"
    local resources_dir="$contents_dir/Resources"
    
    # 创建应用程序目录结构
    sudo mkdir -p "$macos_dir"
    sudo mkdir -p "$resources_dir"
    
    # 创建符号链接到实际的二进制文件
    sudo ln -sf "$INSTALL_DIR/tmplink" "$macos_dir/tmplink"
    
    # 创建 Info.plist
    sudo tee "$contents_dir/Info.plist" > /dev/null << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>tmplink</string>
    <key>CFBundleIconFile</key>
    <string>icon</string>
    <key>CFBundleIdentifier</key>
    <string>com.tmplink.uploader</string>
    <key>CFBundleName</key>
    <string>TmpLink Uploader</string>
    <key>CFBundleDisplayName</key>
    <string>钛盘上传工具</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0</string>
    <key>CFBundleVersion</key>
    <string>1</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.15</string>
    <key>LSUIElement</key>
    <false/>
</dict>
</plist>
EOF
    
    # 移除应用程序包的隔离属性
    sudo xattr -rd com.apple.quarantine "$app_dir" 2>/dev/null || true
    
    print_info "应用程序包已创建: $app_dir"
    print_success "应用程序包创建完成"
}

verify_installation() {
    print_step "验证安装..."
    
    if command -v tmplink &> /dev/null && command -v tmplink-cli &> /dev/null; then
        print_success "安装验证成功"
        print_info "GUI 程序: $(which tmplink)"
        print_info "CLI 程序: $(which tmplink-cli)"
        print_info "应用程序包: /Applications/TmpLink Uploader.app"
    else
        print_error "安装验证失败，请检查 PATH 环境变量"
        exit 1
    fi
}

setup_environment() {
    print_step "配置环境..."
    
    # macOS 默认包含 /usr/local/bin 在 PATH 中
    if [[ ":$PATH:" != *":/usr/local/bin:"* ]]; then
        print_info "添加 /usr/local/bin 到 PATH"
        
        # 根据不同的 shell 添加到配置文件
        local shell_config=""
        case "$SHELL" in
            */bash)
                shell_config="$HOME/.bash_profile"
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
    echo "安装完成！您现在可以使用以下方式启动程序："
    echo
    echo "方式一：应用程序文件夹"
    echo "  打开 Finder → 应用程序 → 双击 'TmpLink Uploader'"
    echo
    echo "方式二：Spotlight 搜索"
    echo "  按 Cmd+Space，搜索 'TmpLink' 或'钛盘'"
    echo
    echo "方式三：命令行"
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
    echo "  https://github.com/$GITHUB_REPO"
    echo
    echo "macOS 用户提示："
    echo "  - 如果系统提示'无法验证开发者'，脚本已自动处理"
    echo "  - 如果仍有问题，请在终端运行："
    echo "    sudo spctl --master-disable"
    echo "  - 使用完成后可重新启用 Gatekeeper："
    echo "    sudo spctl --master-enable"
    echo
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
    remove_quarantine
    install_binaries
    create_app_bundle
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