# 钛盘上传工具 - 预编译版本

本目录包含了钛盘上传工具的跨平台预编译版本，可直接使用无需安装Go环境。

## 目录结构

```
build/
├── macos-intel/        # macOS Intel 处理器版本
├── macos-arm64/        # macOS Apple Silicon 版本  
├── windows-64bit/      # Windows 64位版本
├── windows-32bit/      # Windows 32位版本
├── linux-64bit/        # Linux 64位版本
├── linux-32bit/        # Linux 32位版本
├── linux-arm64/        # Linux ARM64版本
└── README.md           # 本文档
```

## 如何选择版本

### macOS 用户
- **Intel Mac** (2020年之前): 使用 `macos-intel/`
- **Apple Silicon Mac** (M1/M2/M3): 使用 `macos-arm64/`

### Windows 用户  
- **64位 Windows**: 使用 `windows-64bit/` (推荐)
- **32位 Windows**: 使用 `windows-32bit/`

### Linux 用户
- **64位 x86_64**: 使用 `linux-64bit/` (最常见)
- **32位 i386**: 使用 `linux-32bit/`
- **ARM64**: 使用 `linux-arm64/` (树莓派4, 服务器等)

## 程序说明

每个目录都包含两个程序：

### tmplink (GUI程序)
- 图形用户界面，基于终端的交互式界面
- 提供文件选择、上传管理、进度监控功能
- 适合日常使用

### tmplink-cli (CLI程序)  
- 命令行工具，适合脚本和自动化
- 独立的单文件上传功能
- 支持批处理和系统集成

## 快速开始

### 1. 下载对应版本
根据您的系统选择合适的目录，下载其中的程序文件。

### 2. 设置执行权限 (macOS/Linux)
```bash
chmod +x tmplink tmplink-cli
```

### 3. 运行程序

#### GUI模式 (推荐新用户)
```bash
# macOS/Linux
./tmplink

# Windows  
tmplink.exe
```

#### CLI模式 (适合脚本)
```bash
# macOS/Linux
./tmplink-cli -file /path/to/file -token YOUR_TOKEN -task-id task1 -status-file status.json

# Windows
tmplink-cli.exe -file C:\path\to\file -token YOUR_TOKEN -task-id task1 -status-file status.json
```

## 获取API Token

1. 访问 [钛盘](https://tmp.link/) 并登录
2. 打开浏览器开发者工具 (F12)  
3. 在控制台执行: `localStorage.getItem('token')`
4. 复制返回的token值

## 常见问题

### macOS用户

**问题**: "无法打开，因为无法验证开发者"
```bash
# 解决方法
xattr -d com.apple.quarantine tmplink tmplink-cli
```

**问题**: 权限被拒绝
```bash
# 解决方法
chmod +x tmplink tmplink-cli
```

### Windows用户

**问题**: Windows Defender阻止运行
- 点击"更多信息" → "仍要运行"
- 或将程序添加到Windows Defender信任列表

**问题**: 找不到程序
- 确保在正确的目录下运行
- 或将程序目录添加到系统PATH

### Linux用户

**问题**: 权限不足
```bash
chmod +x tmplink tmplink-cli
```

**问题**: 依赖库缺失
```bash
# 大多数现代Linux发行版无需额外依赖
# 如有问题，请安装基础C库
sudo apt-get install libc6  # Ubuntu/Debian
sudo yum install glibc      # CentOS/RHEL
```

## 系统要求

### 最低配置
- **内存**: 64MB RAM
- **存储**: 20MB 可用空间  
- **网络**: 互联网连接

### 推荐配置
- **内存**: 256MB+ RAM
- **存储**: 100MB+ 可用空间
- **网络**: 宽带连接

## 版本信息

- **构建**: 跨平台Go编译
- **架构**: 双进程设计
- **GUI框架**: Bubble Tea
- **文件大小**: 8-11MB

## 安全提示

1. **来源验证**: 仅从官方仓库下载
2. **权限控制**: 程序仅需要网络访问权限
3. **数据安全**: Token等敏感信息加密存储
4. **网络安全**: 所有API调用使用HTTPS

## 更多信息

- **项目主页**: [README.md](../README.md)
- **使用指南**: [docs/usage.md](../docs/usage.md)  
- **API文档**: [docs/api.md](../docs/api.md)
- **设计文档**: [docs/design.md](../docs/design.md)


---

如遇问题，请查看文档或提交Issue获得帮助。
