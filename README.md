# 钛盘文件上传工具

一个用于 [钛盘](https://tmp.link/) 的高效文件上传工具，支持分块上传、断点续传和实时进度监控。

## 项目介绍

本工具提供图形界面(GUI)和命令行(CLI)两种使用方式：

- **tmplink** (GUI): 终端图形界面，适合日常文件上传
- **tmplink-cli** (CLI): 命令行工具，适合脚本自动化和批量操作

### 主要特性

- ✅ **分块上传**: 大文件自动分块，支持1-99MB分块大小
- ✅ **断点续传**: 基于SHA1的文件去重和秒传
- ✅ **实时监控**: 上传进度和速度实时显示
- ✅ **并发上传**: 多线程并发，提高传输效率
- ✅ **权限分级**: 支持普通用户和赞助用户不同功能
- ✅ **文件大小限制**: 支持最大50GB单文件上传

## 安装

### 方式一：下载预编译版本 (推荐)

从项目的 `build` 目录下载对应平台的预编译版本：

| 平台 | 架构 | 目录 |
|------|------|------|
| macOS | Intel | `macos-intel/` |
| macOS | Apple Silicon | `macos-arm64/` |
| Windows | 64位 | `windows-64bit/` |
| Windows | 32位 | `windows-32bit/` |
| Linux | x86_64 | `linux-64bit/` |
| Linux | i386 | `linux-32bit/` |
| Linux | ARM64 | `linux-arm64/` |

**安装步骤：**

```bash
# 1. 进入对应平台目录 (如 build/macos-arm64/)
# 2. 设置执行权限 (macOS/Linux)
chmod +x tmplink tmplink-cli

# 3. 运行程序
./tmplink        # GUI模式
./tmplink-cli    # CLI模式
```

### 方式二：从源码编译

```bash
# 克隆仓库
git clone https://github.com/tmplink/tmplink_uploader.git
cd tmplink_uploader

# 编译
make build        # 编译当前平台
make release      # 编译所有平台
```

## 使用方法

### 获取API Token

使用前需要获取钛盘API Token：

1. 访问 [钛盘](https://tmp.link/) 并登录
2. 点开上传文件，然后点击 "重新设定"，然后在 "命令行上传" 界面复制 Token

### GUI模式使用

适合日常文件上传，提供友好的交互界面：

```bash
# 启动GUI
./tmplink

# 首次使用会提示输入API Token
# 然后可以通过界面选择文件并上传
```

**主要功能：**
- 📁 文件浏览和选择
- 📊 实时上传进度和速度显示  
- ⚙️ 上传参数配置（赞助用户）
- 🔄 多文件并发上传
- 📋 上传历史记录

### CLI模式使用

#### 基本用法

```bash
# 首次使用 - 设置Token
./tmplink-cli -set-token YOUR_API_TOKEN

# 上传文件 - 自动显示进度条
./tmplink-cli -file /path/to/file.txt
```

#### 常用示例

```bash
# 大文件上传（调整分块大小）
./tmplink-cli -file ./large-video.mp4 -chunk-size 10

# 无限期保存文件
./tmplink-cli -file ./important.zip -model 99

# 指定上传服务器
./tmplink-cli -file ./file.txt -upload-server https://up-jp.tmp.link

# 使用临时Token
./tmplink-cli -file ./file.txt -token temporary_token

# 设置默认配置
./tmplink-cli -set-model 2              # 7天有效期
./tmplink-cli -set-mr-id folder123      # 指定目录

# 调试模式
./tmplink-cli -file ./test.txt -debug
```

#### CLI进度显示

CLI模式提供类似wget/curl的进度显示：

```
🚀 开始上传文件: document.pdf
📊 文件大小: 2.5 MB

📤 上传中 [████████████████████████████████████████] 100% | 2.5 MB/2.5 MB | 1.2 MB/s | ETA: 0s

✅ 上传完成!
📁 文件名: document.pdf
📊 文件大小: 2.5 MB
⚡ 平均速度: 1.15 MB/s
⏱️ 总耗时: 2s
🔗 下载链接: https://tmp.link/f/abc123
```

## 参数说明

### 必需参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `-file` | 要上传的文件路径 | `/home/user/document.pdf` |

### 常用参数

| 参数 | 说明 | 默认值 | 示例 |
|------|------|--------|------|
| `-chunk-size` | 分块大小(MB,1-99) | `3` | `5` |
| `-model` | 文件有效期 | `0`(24小时) | `99`(无限期) |
| `-token` | API Token | 已保存值 | `your_token` |
| `-debug` | 调试模式 | `false` | - |

### 配置管理

| 参数 | 说明 | 示例 |
|------|------|------|
| `-set-token` | 设置并保存API Token | `your_token_here` |
| `-set-model` | 设置默认文件有效期 | `2` |
| `-set-mr-id` | 设置默认目录ID | `folder123` |

**文件有效期选项：**
- `0`: 24小时 (默认)
- `1`: 3天  
- `2`: 7天
- `99`: 无限期

## 故障排除

### 常见问题

**1. macOS安全提示**
```bash
# 解决"无法验证开发者"错误
xattr -d com.apple.quarantine tmplink tmplink-cli
```

**2. Linux权限问题**
```bash
# 设置执行权限
chmod +x tmplink tmplink-cli
```

**3. Windows Defender拦截**
- 点击"更多信息" → "仍要运行"
- 或添加到信任列表

**4. Token相关问题**
- 确认从正确位置复制Token
- 检查Token是否过期
- 重新登录钛盘获取新Token

**5. 上传失败**
- 检查网络连接
- 验证文件权限
- 使用`-debug`参数获取详细错误信息

### 获取帮助

```bash
# 查看所有参数
./tmplink-cli -h

# 启用调试模式
./tmplink-cli -debug -file test.txt

# 查看详细文档
cat docs/usage.md
```

## 相关文档

- 📚 [详细使用指南](docs/usage.md) - 完整的功能说明和使用示例
- 🔧 [技术文档](docs/technical.md) - 架构设计和开发指南
- 🔌 [API文档](docs/api.md) - 钛盘API集成规范
- 🎨 [设计文档](docs/design.md) - 系统设计理念和架构
- 📦 [预编译版本说明](build/README.md) - 跨平台版本使用指南

## 版本更新

### v1.0.1 (2025-05-25)
- 🐛 修复GUI上传进程通信问题
- ✅ 完善CLI参数文档和错误处理
- 🎯 新增CLI模式实时进度条显示

### v1.0.0 (2025-05-24)
- 🎉 首次发布，实现双进程架构
- ✅ 支持GUI和CLI双模式
- 📊 完成分块上传和权限分级系统

## 许可证

本项目基于 Apache 2.0 许可证开源。详见 [LICENSE](LICENSE) 文件。

---

💡 **提示**: 如有问题或建议，欢迎提交 [Issue](https://github.com/tmplink/tmplink_uploader/issues) 或 [Pull Request](https://github.com/tmplink/tmplink_uploader/pulls)。
