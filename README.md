# 钛盘文件上传工具

一个用于 [钛盘](https://tmp.link/) 的高效文件上传工具，支持分块上传、断点续传和实时进度监控。

## 项目介绍

本工具采用双进程架构设计，提供图形界面(GUI)和命令行(CLI)两种使用方式：

- **tmplink** (GUI): 终端图形界面，适合日常文件上传
- **tmplink-cli** (CLI): 命令行工具，适合脚本自动化和批量操作

### 主要特性

- ✅ **分块上传**: 大文件自动分块，支持1-99MB分块大小
- ✅ **断点续传**: 基于SHA1的文件去重和秒传
- ✅ **实时监控**: 上传进度和速度实时显示
- ✅ **并发上传**: 多线程并发，提高传输效率
- ✅ **智能重试**: 自动错误恢复和重试机制
- ✅ **权限分级**: 支持普通用户和赞助用户不同功能
- ✅ **文件大小限制**: 支持最大50GB单文件上传

## 安装

### 方式一：下载预编译版本 (推荐)

从 [Releases](https://github.com/your-repo/tmplink_uploader/releases) 下载对应平台的预编译版本：

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
# 1. 下载并解压对应版本
# 2. 设置执行权限 (macOS/Linux)
chmod +x tmplink tmplink-cli

# 3. 运行程序
./tmplink        # GUI模式
./tmplink-cli    # CLI模式
```

### 方式二：从源码编译

```bash
# 克隆仓库
git clone https://github.com/your-repo/tmplink_uploader.git
cd tmplink_uploader

# 编译
make build        # 编译当前平台
make release      # 编译所有平台
```

## 使用方法

### 获取API Token

使用前需要获取钛盘API Token：

1. 访问 [钛盘](https://tmp.link/) 并登录
2. 打开浏览器开发者工具 (F12)
3. 在控制台执行：`localStorage.getItem('token')`
4. 复制返回的token值

### GUI模式使用

适合日常文件上传，提供友好的交互界面：

```bash
# 启动GUI
./tmplink

# 首次使用会提示输入API Token
# 然后可以通过界面选择文件并上传
```

**GUI功能说明：**
- 📁 文件浏览和选择
- 📊 实时上传进度和速度显示
- ⚙️ 上传参数配置
- 🔄 多文件并发上传
- 📋 上传历史记录

### CLI模式使用

#### 基本用法

```bash
./tmplink-cli \
  -file /path/to/file.txt \
  -token YOUR_API_TOKEN \
  -task-id unique-task-id \
  -status-file /path/to/status.json
```

#### 参数详解

**必需参数：**

| 参数 | 说明 | 示例 |
|------|------|------|
| `-file` | 要上传的文件路径 | `/home/user/document.pdf` |
| `-token` | 钛盘 API Token | `abcd1234...` |
| `-task-id` | 唯一任务标识符 | `upload-001` |
| `-status-file` | 状态文件路径 | `/tmp/status.json` |

**可选参数：**

| 参数 | 说明 | 默认值 | 示例 |
|------|------|--------|------|
| `-api-server` | API服务器地址 | `https://tmplink-sec.vxtrans.com/api_v2` | |
| `-upload-server` | 指定上传服务器 | 自动选择 | - |
| `-chunk-size` | 分块大小(MB) | `3` | `5` |
| `-timeout` | 请求超时(秒) | `300` | `600` |
| `-model` | 文件有效期 | `0` (24小时) | `99` (永久) |
| `-mr-id` | 资源ID | `0` (根目录) | `12345` |
| `-skip-upload` | 启用秒传检查 | `1` | `0` |
| `-uid` | 用户ID | 自动获取 | `123456` |
| `-debug` | 调试模式 | `false` | `true` |

**文件有效期选项：**

| 值 | 说明 |
|---|------|
| `0` | 24小时 (默认) |
| `1` | 3天 |
| `2` | 7天 |
| `99` | 永久保存 |

#### 使用示例

**1. 基本上传**
```bash
./tmplink-cli \
  -file ./document.pdf \
  -token "your_token_here" \
  -task-id "doc-upload-001" \
  -status-file ./status.json
```

**2. 大文件上传（调整分块大小）**
```bash
./tmplink-cli \
  -file ./large-video.mp4 \
  -token "your_token_here" \
  -task-id "video-upload" \
  -status-file ./video-status.json \
  -chunk-size 10 \
  -timeout 600
```

**3. 永久保存文件**
```bash
./tmplink-cli \
  -file ./important.zip \
  -token "your_token_here" \
  -task-id "permanent-file" \
  -status-file ./status.json \
  -model 99
```

**4. 指定上传服务器**
```bash
./tmplink-cli \
  -file ./file.txt \
  -token "your_token_here" \
  -task-id "server-specific" \
  -status-file ./status.json \
  -upload-server "https://up-jp.tmp.link"
```

**5. 调试模式**
```bash
./tmplink-cli \
  -file ./test.txt \
  -token "your_token_here" \
  -task-id "debug-test" \
  -status-file ./debug-status.json \
  -debug
```

#### 状态文件监控

CLI通过JSON状态文件提供实时进度信息：

```json
{
  "task_id": "upload-001",
  "status": "uploading",
  "progress": 75.5,
  "speed": "2.5 MB/s",
  "uploaded": 7548160,
  "total": 10000000,
  "eta": "00:00:30",
  "error": null
}
```

**状态字段说明：**
- `task_id`: 任务ID
- `status`: 状态 (`starting`, `uploading`, `completed`, `failed`)
- `progress`: 进度百分比 (0-100)
- `speed`: 上传速度
- `uploaded`: 已上传字节数
- `total`: 文件总大小
- `eta`: 预计剩余时间
- `error`: 错误信息 (如有)

#### 批量上传脚本示例

```bash
#!/bin/bash
TOKEN="your_token_here"
FILES=("file1.txt" "file2.pdf" "file3.zip")

for i in "${!FILES[@]}"; do
    ./tmplink-cli \
        -file "${FILES[$i]}" \
        -token "$TOKEN" \
        -task-id "batch-$i" \
        -status-file "status-$i.json" &
done

wait  # 等待所有上传完成
echo "批量上传完成"
```

#### 错误处理

常见错误及解决方法：

| 错误信息 | 原因 | 解决方法 |
|----------|------|----------|
| `Token invalid` | Token过期或无效 | 重新获取Token |
| `File not found` | 文件路径错误 | 检查文件路径 |
| `Upload failed` | 网络或服务器问题 | 检查网络连接，重试上传 |
| `Chunk size too large` | 分块大小超限 | 设置chunk-size在1-99MB之间 |
| `File too large` | 文件大小超限 | 文件大小不能超过50GB |

## 架构设计

### 双进程架构

```
┌─────────────────┐    启动子进程    ┌─────────────────┐
│   tmplink       │  ───────────→   │   tmplink-cli   │
│   (GUI主程序)    │                │   (上传进程)     │
│                 │  ←─────────────  │                 │
└─────────────────┘    状态文件      └─────────────────┘
```

**设计优势：**
- 进程独立：上传进程崩溃不影响主界面
- 资源隔离：每个文件独立进程，避免相互影响
- 可扩展性：支持并发上传多个文件
- 监控能力：通过状态文件实时监控进度

### 权限系统

**普通用户功能：**
- ✅ 文件上传下载
- ✅ 基础设置
- ✅ 自动服务器选择

**赞助用户专享：**
- ⭐ 分块大小自定义
- ⭐ 并发数控制
- ⭐ 手动服务器选择
- ⭐ 快速上传开关

## 开发相关

### 构建命令

```bash
# 安装依赖
make deps

# 代码格式化
make fmt

# 代码检查
make vet

# 运行测试
make test

# 构建程序
make build        # 当前平台
make release      # 所有平台
make dist         # 创建发布包

# 清理
make clean
```

### 技术栈

- **语言**: Go 1.19+
- **GUI框架**: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **CLI框架**: [Cobra](https://github.com/spf13/cobra)
- **HTTP客户端**: Go标准库
- **构建工具**: Make

### 依赖库

```go
github.com/spf13/cobra                    // CLI框架
github.com/charmbracelet/bubbletea        // TUI框架
github.com/charmbracelet/bubbles          // TUI组件
github.com/charmbracelet/lipgloss         // TUI样式
github.com/mattn/go-runewidth            // Unicode宽度
github.com/schollz/progressbar/v3         // 进度条
```

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
- 查看状态文件中的详细错误信息
- 使用`-debug`参数获取更多信息

### 调试方法

```bash
# 启用详细日志
./tmplink-cli -debug -file test.txt -token TOKEN -task-id debug -status-file status.json

# 查看状态文件
cat status.json | jq .

# 监控实时状态
watch -n 1 'cat status.json | jq .'
```

## 更新日志

### v1.0.1 (2025-05-25)

**🐛 关键修复:**
- 修复GUI上传进程通信问题
- 解决变量遮蔽导致的空指针异常
- 修复mr_id参数默认值错误

**✅ 功能改进:**
- 完善CLI参数文档
- 优化错误处理机制
- 增强状态监控功能

### v1.0.0 (2025-05-24)

**🎉 首次发布:**
- 实现双进程架构
- 支持GUI和CLI双模式
- 完成分块上传功能
- 添加权限分级系统

## 许可证

本项目基于 Apache 2.0 许可证开源。详见 [LICENSE](LICENSE) 文件。

## 相关文档

- [设计文档](docs/design.md) - 系统架构和设计理念
- [API文档](docs/api.md) - 钛盘 API集成规范  
- [使用说明](docs/usage.md) - 详细使用指南
- [预编译版本说明](build/README.md) - 跨平台预编译版本使用指南

---

如有问题或建议，欢迎提交 [Issue](https://github.com/your-repo/tmplink_uploader/issues) 或 [Pull Request](https://github.com/your-repo/tmplink_uploader/pulls)。
