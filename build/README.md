# 钛盘上传工具 - 编译版本

本目录包含了钛盘上传工具的跨平台编译版本。

## 目录结构

```
build/
├── macos/          # macOS 版本
├── windows/        # Windows 版本
├── linux/          # Linux 版本
└── README.md       # 本文档
```

## 平台支持

### macOS
- `tmplink-amd64` / `tmplink-cli-amd64` - Intel Mac (x86_64)
- `tmplink-arm64` / `tmplink-cli-arm64` - Apple Silicon Mac (ARM64)

### Windows
- `tmplink-amd64.exe` / `tmplink-cli-amd64.exe` - 64位 Windows
- `tmplink-386.exe` / `tmplink-cli-386.exe` - 32位 Windows

### Linux
- `tmplink-amd64` / `tmplink-cli-amd64` - 64位 Linux (x86_64)
- `tmplink-386` / `tmplink-cli-386` - 32位 Linux (i386)
- `tmplink-arm64` / `tmplink-cli-arm64` - ARM64 Linux (AArch64)

## 程序说明

### tmplink (GUI程序)
- 图形用户界面程序
- 基于 Bubble Tea 的终端界面
- 提供文件选择、上传管理、进度监控功能

### tmplink-cli (CLI程序)
- 命令行上传程序
- 独立的单文件上传工具
- 支持状态文件通信

## 使用方法

### 选择合适的版本

1. **确定您的操作系统**:
   - macOS: 选择 `macos/` 目录
   - Windows: 选择 `windows/` 目录  
   - Linux: 选择 `linux/` 目录

2. **确定您的架构**:
   - Intel/AMD 64位: 选择 `amd64` 版本
   - Intel 32位: 选择 `386` 版本
   - ARM64: 选择 `arm64` 版本

### 运行程序

#### macOS/Linux
```bash
# 添加执行权限
chmod +x tmplink-amd64
chmod +x tmplink-cli-amd64

# 运行GUI程序
./tmplink-amd64

# 运行CLI程序
./tmplink-cli-amd64 -file /path/to/file -token YOUR_TOKEN -task-id task1 -status-file status.json
```

#### Windows
```cmd
# 运行GUI程序
tmplink-amd64.exe

# 运行CLI程序
tmplink-cli-amd64.exe -file C:\path\to\file -token YOUR_TOKEN -task-id task1 -status-file status.json
```

## 版本信息

- **架构**: 双进程架构设计
- **语言**: Go 1.19+
- **GUI框架**: Bubble Tea
- **编译时间**: 自动构建

## 系统要求

### 最低要求
- **内存**: 64MB RAM
- **磁盘**: 20MB 可用空间
- **网络**: 稳定的互联网连接

### 推荐配置
- **内存**: 256MB+ RAM
- **磁盘**: 100MB+ 可用空间
- **网络**: 宽带连接

## 安全说明

1. **文件完整性**: 建议验证下载文件的哈希值
2. **权限设置**: Linux/macOS 用户需要设置执行权限
3. **防火墙**: 确保防火墙允许程序访问网络
4. **杀毒软件**: 首次运行时可能需要添加信任

## 故障排除

### 常见问题

1. **权限被拒绝 (Linux/macOS)**:
   ```bash
   chmod +x tmplink-*
   ```

2. **文件被阻止 (macOS)**:
   ```bash
   xattr -d com.apple.quarantine tmplink-*
   ```

3. **Windows Defender 警告**:
   - 点击"更多信息" → "仍要运行"
   - 或将程序添加到信任列表

4. **架构不匹配**:
   - 确认下载了正确的架构版本
   - 使用 `uname -m` (Linux/macOS) 或系统信息查看架构

### 获取帮助

- 项目文档: [README.md](../README.md)
- 使用指南: [docs/usage.md](../docs/usage.md)
- API文档: [docs/api.md](../docs/api.md)

## 更新说明

编译版本会定期更新，建议：

1. 定期检查项目仓库获取最新版本
2. 关注版本更新日志了解新功能
3. 备份重要配置文件后再更新

---

如有问题，请参考项目文档或提交 Issue。