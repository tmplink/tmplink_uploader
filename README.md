# TmpLink 文件上传工具

一个用于 [钛盘 (TmpLink)](https://tmp.link/) 的命令行文件上传工具，采用双进程架构设计。

## 项目概述

本项目实现了一个分离式的文件上传解决方案：
- **tmplink** (GUI): 图形界面程序，负责用户交互和任务管理
- **tmplink-cli** (CLI): 命令行上传程序，负责实际的文件上传操作

## 架构设计

### 双进程架构

```
┌─────────────────┐    启动子进程    ┌─────────────────┐
│                 │  ───────────→   │                 │
│   tmplink       │                │  tmplink-cli    │
│   (GUI主程序)    │                │  (上传进程)      │
│                 │  ←─────────────  │                 │
└─────────────────┘    状态文件      └─────────────────┘
```

### 进程间通信

- **命令行参数**: GUI通过命令行参数启动CLI进程
- **状态文件**: CLI通过JSON状态文件报告进度和结果
- **进程管理**: GUI监控CLI进程状态和生命周期

### 关键特性

1. **独立上传进程**: 每个文件上传对应一个独立的CLI进程
2. **无配置文件依赖**: CLI程序所有参数通过命令行传递
3. **一次性执行**: CLI进程完成上传后自动退出
4. **并发上传**: GUI可以启动多个CLI进程实现并发上传
5. **状态监控**: 通过状态文件实时监控上传进度

## 安装和使用

### 构建

```bash
# 构建所有程序
make build

# 或分别构建
make build-cli    # 构建CLI程序
make build-gui    # 构建GUI程序
```

### 使用GUI程序

```bash
# 启动图形界面
./tmplink
```

GUI程序提供：
- Token输入界面
- 文件选择器
- 实时上传进度显示
- 下载链接获取

### 直接使用CLI程序

```bash
# CLI参数说明
./tmplink-cli \
  -file /path/to/file.txt \
  -token YOUR_API_TOKEN \
  -task-id unique-task-id \
  -status-file /path/to/status.json \
  [-server https://tmplink-sec.vxtrans.com/api_v2] \
  [-chunk-size 3145728] \
  [-max-retries 3] \
  [-timeout 300] \
  [-model 0] \
  [-mr-id "0"] \
  [-skip-upload 1] \
  [-debug]
```

### 参数说明

| 参数 | 必需 | 说明 | 默认值 |
|------|------|------|--------|
| `-file` | ✓ | 要上传的文件路径 | - |
| `-token` | ✓ | TmpLink API Token | - |
| `-task-id` | ✓ | 唯一任务ID | - |
| `-status-file` | ✓ | 状态文件保存路径 | - |
| `-server` | - | 上传服务器地址 | https://tmplink-sec.vxtrans.com/api_v2 |
| `-chunk-size` | - | 分块大小(字节) | 3145728 (3MB) |
| `-max-retries` | - | 最大重试次数 | 3 |
| `-timeout` | - | 超时时间(秒) | 300 |
| `-model` | - | 文件有效期 | 0 (24小时) |
| `-mr-id` | - | 资源ID | 0 (根目录) |
| `-skip-upload` | - | 跳过上传检查 | 1 (启用秒传) |
| `-debug` | - | 调试模式 | false |

## 状态文件格式

CLI程序会创建和更新JSON格式的状态文件：

```json
{
  "id": "task_1640995200",
  "status": "in_progress",
  "file_path": "/path/to/file.txt",
  "file_name": "file.txt",
  "file_size": 1048576,
  "progress": 75.5,
  "download_url": "",
  "error_msg": "",
  "created_at": "2023-12-31T12:00:00Z",
  "updated_at": "2023-12-31T12:01:00Z"
}
```

### 状态值

- `pending`: 任务创建，准备开始
- `in_progress`: 正在上传
- `completed`: 上传完成
- `failed`: 上传失败

## 开发

### 项目结构

```
├── cmd/
│   ├── tmplink/         # GUI主程序
│   └── tmplink-cli/     # CLI上传程序
├── internal/
│   ├── gui/
│   │   └── tui/         # 终端用户界面
│   └── shared/          # 共享组件
├── Makefile             # 构建脚本
├── integration_test.go  # 集成测试
└── README.md           # 项目文档
```

### 运行测试

```bash
# 运行所有测试
make test

# 运行集成测试
go test -v ./integration_test.go

# 生成测试覆盖率报告
make test-coverage
```

### 代码检查

```bash
# 格式化代码
make fmt

# 运行代码检查
make vet

# 运行完整检查
make check
```

## API集成

### TmpLink API

本工具集成了TmpLink的REST API：

- **Base URL**: `https://tmplink-sec.vxtrans.com/api_v2`
- **认证方式**: Bearer Token
- **请求格式**: `application/x-www-form-urlencoded`

### 关键API端点

1. **上传准备**: `POST /file`
   - Action: `upload_request_select2`
   - 检查文件是否可以秒传

2. **分片上传**: `POST /app/upload_slice`
   - 上传单个文件分片

3. **完成上传**: `POST /file`
   - Action: `upload_complete`
   - 获取下载链接

### 上传流程

1. 计算文件SHA1哈希值
2. 发送准备请求检查秒传
3. 如果需要上传，按分片上传文件
4. 接收状态8合并完成响应
5. 获取下载链接 (ukey格式)

### API状态码

- **状态1**: 秒传成功
- **状态2**: 等待其他分片
- **状态3**: 准备上传下一分片
- **状态6**: 文件已存在(秒传)
- **状态7**: 上传错误 (data字段包含错误码，如data=8表示"文件夹未找到")
- **状态8**: 分片合并完成 (正常上传完成)

## 许可证

本项目基于 MIT 许可证开源，详见 [LICENSE](LICENSE) 文件。

## 贡献

欢迎提交问题和拉取请求。请确保：

1. 代码通过所有测试
2. 遵循现有的代码风格
3. 添加适当的测试覆盖
4. 更新相关文档

## 更新日志

### v1.0.1 (2024-12-31)

- 🐛 **修复关键bug**: mr_id参数默认值从空字符串改为"0"
- ✅ 修复文件夹查找错误 (status 7 data=8)
- ✅ 正确处理状态码8 (合并完成)
- ✅ 完善API状态码文档
- ✅ 增加测试验证和调试模式

### v1.0.0 (2024-12-31)

- ✅ 实现双进程架构设计
- ✅ 完成GUI和CLI程序分离
- ✅ 实现状态文件通信机制
- ✅ 支持文件分片上传
- ✅ 添加进度监控功能
- ✅ 完善错误处理和重试机制
- ✅ 完成集成测试覆盖

## 技术栈

- **语言**: Go 1.19+
- **GUI框架**: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **CLI框架**: Go标准库 `flag`
- **HTTP客户端**: Go标准库 `net/http`
- **构建工具**: Make

## 性能特性

- **分片上传**: 支持大文件分片上传
- **并发处理**: 支持多文件并发上传
- **断点续传**: 通过SHA1哈希实现重复文件检测
- **进度跟踪**: 实时显示上传进度
- **错误恢复**: 自动重试机制

## 故障排除

### 常见问题

1. **构建失败**
   ```bash
   # 更新依赖
   go mod tidy
   make build
   ```

2. **CLI程序找不到**
   ```bash
   # 确保CLI程序在PATH中或使用绝对路径
   export PATH=$PATH:/path/to/tmplink-cli
   ```

3. **上传失败**
   - 检查网络连接
   - 确认API Token有效
   - 查看状态文件中的错误信息
   - 使用 `-debug` 参数查看详细调试信息
   
4. **状态码7错误**
   - 检查mr_id参数设置 (默认应为"0")
   - 确认目标文件夹存在

### 调试

```bash
# 查看详细日志
tail -f ~/.tmplink/tmplink.log

# 检查状态文件
cat ~/.tmplink/tasks/task_*.json
```

## 联系方式

- 项目主页: [GitHub](https://github.com/your-username/tmplink_uploader)
- 问题反馈: [Issues](https://github.com/your-username/tmplink_uploader/issues)