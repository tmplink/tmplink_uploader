# 钛盘上传工具安装指南

本文档提供了钛盘上传工具在不同操作系统上的安装方法。

## 系统要求

### 通用要求
- **Go 语言环境**: 1.19 或更高版本
- **Make 工具**: 用于构建程序
- **网络连接**: 下载依赖和上传文件

### 平台特定要求

#### macOS
- macOS 10.14+ (支持 Intel 和 Apple Silicon)
- Xcode Command Line Tools (包含 make)

#### Linux
- 任何现代 Linux 发行版
- 包管理器 (apt, yum, dnf, pacman 等)
- 基本开发工具 (make, tar, gzip)

#### Windows
- Windows 7/10/11 (32位/64位)
- PowerShell 5.0 或更高版本
- Git for Windows (推荐) 或其他 make 工具

## 一键安装

### macOS 安装

```bash
# 下载并运行安装脚本
chmod +x install-macos.sh
./install-macos.sh
```

**安装内容:**
- 二进制文件: `/usr/local/bin/tmplink` 和 `/usr/local/bin/tmplink-cli`
- 配置文件: `~/.tmplink_config.json`

**安装后使用:**
```bash
tmplink       # 启动图形界面
tmplink-cli   # 命令行版本
```

### Linux 安装

```bash
# 下载并运行安装脚本
chmod +x install-linux.sh
./install-linux.sh
```

**安装内容:**
- 二进制文件: `/usr/local/bin/tmplink` 和 `/usr/local/bin/tmplink-cli`
- 桌面快捷方式: `~/.local/share/applications/tmplink.desktop`
- 配置文件: `~/.tmplink_config.json`

**安装后使用:**
```bash
tmplink       # 启动图形界面
tmplink-cli   # 命令行版本
```

也可以在应用菜单中搜索"钛盘上传工具"启动图形界面。

### Windows 安装

```powershell
# 以管理员身份运行 PowerShell，然后执行
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
.\install-windows.ps1
```

**安装选项:**
```powershell
# 自定义安装路径
.\install-windows.ps1 -InstallPath "C:\MyApps\TmpLink"

# 静默安装（跳过确认）
.\install-windows.ps1 -Force

# 不创建桌面快捷方式
.\install-windows.ps1 -NoDesktop

# 不创建开始菜单快捷方式
.\install-windows.ps1 -NoStartMenu

# 不修改 PATH 环境变量
.\install-windows.ps1 -NoPath
```

**安装内容:**
- 程序文件: `C:\Program Files\TmpLink\` (管理员) 或 `%LOCALAPPDATA%\Programs\TmpLink\` (普通用户)
- 开始菜单快捷方式: "钛盘上传工具"
- 桌面快捷方式: "钛盘上传工具"
- 配置文件: `%USERPROFILE%\.tmplink_config.json`
- 卸载脚本: `uninstall.ps1`

**安装后使用:**
```cmd
tmplink.exe     # 启动图形界面
tmplink-cli.exe # 命令行版本
```

## 手动安装

如果一键安装脚本无法使用，可以进行手动安装：

### 1. 克隆代码库

```bash
git clone https://github.com/your-username/tmplink_uploader.git
cd tmplink_uploader
```

### 2. 安装依赖

```bash
go mod download
go mod tidy
```

### 3. 构建程序

```bash
# 构建当前平台版本
make build

# 或构建所有平台版本
make build-release
```

### 4. 安装到系统

#### Unix 系统 (macOS/Linux)
```bash
# 安装到 /usr/local/bin
sudo make install

# 或手动复制
sudo cp tmplink tmplink-cli /usr/local/bin/
sudo chmod +x /usr/local/bin/tmplink /usr/local/bin/tmplink-cli
```

#### Windows
```powershell
# 创建安装目录
New-Item -ItemType Directory -Path "C:\Program Files\TmpLink" -Force

# 复制文件
Copy-Item "tmplink.exe", "tmplink-cli.exe" -Destination "C:\Program Files\TmpLink\"

# 添加到 PATH (可选)
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
[Environment]::SetEnvironmentVariable("Path", "$currentPath;C:\Program Files\TmpLink", "User")
```

## 验证安装

安装完成后，验证程序是否正常工作：

```bash
# 检查版本
tmplink --version
tmplink-cli --version

# 检查帮助
tmplink --help
tmplink-cli --help
```

## 首次配置

1. **启动程序**:
   ```bash
   tmplink
   ```

2. **输入 API Token**:
   - 访问 https://tmp.link/ 并登录
   - 获取 API token (通常在浏览器 localStorage 中)
   - 在程序中输入 token

3. **配置上传参数** (可选):
   - 块大小: 1-80MB (默认 3MB)
   - 并发数: 1-10 (默认 5)
   - 超时时间: 60-600秒 (默认 300秒)

## 卸载程序

### macOS/Linux
```bash
# 使用 Makefile
sudo make uninstall

# 或手动删除
sudo rm -f /usr/local/bin/tmplink /usr/local/bin/tmplink-cli
rm -f ~/.tmplink_config.json
rm -f ~/.local/share/applications/tmplink.desktop  # Linux only
```

### Windows
```powershell
# 运行卸载脚本
& "C:\Program Files\TmpLink\uninstall.ps1"

# 或手动删除
Remove-Item -Path "C:\Program Files\TmpLink" -Recurse -Force
Remove-Item -Path "$env:USERPROFILE\.tmplink_config.json" -ErrorAction SilentlyContinue
```

## 故障排除

### 常见问题

#### 1. Go 未安装
**错误**: `command not found: go`

**解决方案**:
- macOS: `brew install go` 或从 https://golang.org/dl/ 下载
- Linux: `sudo apt install golang-go` (Ubuntu/Debian) 或对应发行版的包管理器
- Windows: 从 https://golang.org/dl/ 下载并安装

#### 2. Make 工具缺失
**错误**: `command not found: make`

**解决方案**:
- macOS: `xcode-select --install`
- Linux: `sudo apt install build-essential` (Ubuntu/Debian)
- Windows: 安装 Git for Windows 或 MSYS2

#### 3. 权限不足
**错误**: `Permission denied`

**解决方案**:
- 使用 `sudo` 运行安装脚本
- 或选择用户目录安装

#### 4. 防火墙拦截 (Windows)
**问题**: Windows Defender 或防火墙阻止程序运行

**解决方案**:
- 将程序添加到白名单
- 允许程序访问网络

#### 5. PATH 环境变量问题
**问题**: 安装后找不到命令

**解决方案**:
- 重新启动终端或重新登录
- 手动添加安装目录到 PATH
- 使用完整路径运行程序

### 获取帮助

如果遇到其他问题：

1. **查看日志**: 程序运行时的错误信息
2. **检查配置**: `~/.tmplink_config.json` 文件内容
3. **测试网络**: 确保可以访问 https://tmp.link/
4. **查看文档**: README.md 和 docs/ 目录
5. **提交问题**: 在项目仓库创建 issue

## 更新程序

要更新到最新版本：

```bash
# 进入项目目录
cd tmplink_uploader

# 拉取最新代码
git pull

# 重新运行安装脚本
./install-macos.sh     # macOS
./install-linux.sh     # Linux
.\install-windows.ps1  # Windows
```

## 开发环境设置

如果您想参与开发或自定义程序：

```bash
# 克隆项目
git clone https://github.com/your-username/tmplink_uploader.git
cd tmplink_uploader

# 安装开发依赖
make dev-setup

# 运行测试
make test

# 代码格式化和检查
make check

# 开发模式运行
make run
```

---

**注意**: 
- 所有安装脚本都会自动处理依赖检查和错误处理
- Windows 用户可能需要调整 PowerShell 执行策略
- 首次运行需要配置 API token 才能使用上传功能
- 建议定期更新程序以获得最新功能和安全修复