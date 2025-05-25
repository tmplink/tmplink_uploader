# 钛盘文件上传工具

一个用于 [钛盘](https://tmp.link/) 的命令行文件上传工具，采用双进程架构设计。

## 项目概述

本项目实现了一个分离式的文件上传解决方案：
- **tmplink** (GUI): 图形界面程序，负责用户交互和任务管理
- **tmplink-cli** (CLI): 命令行上传程序，负责实际的文件上传操作

## 快速开始

### 构建程序

```bash
# 构建当前平台版本
make build

# 构建所有平台发布版本
make release

# 或分别构建
make build-cli    # 构建CLI程序
make build-gui    # 构建GUI程序
```

### 使用GUI程序

```bash
# 启动图形界面
./tmplink
```

### 直接使用CLI程序

```bash
./tmplink-cli \
  -file /path/to/file.txt \
  -token YOUR_API_TOKEN \
  -task-id unique-task-id \
  -status-file /path/to/status.json
```

## 架构特性

### 双进程架构

```
┌─────────────────┐    启动子进程    ┌─────────────────┐
│                 │  ───────────→   │                 │
│   tmplink       │                │  tmplink-cli    │
│   (GUI主程序)    │                │  (上传进程)      │
│                 │  ←─────────────  │                 │
└─────────────────┘    状态文件      └─────────────────┘
```

### 关键特性

1. **独立上传进程**: 每个文件上传对应一个独立的CLI进程
2. **无配置文件依赖**: CLI程序所有参数通过命令行传递
3. **一次性执行**: CLI进程完成上传后自动退出
4. **并发上传**: GUI可以启动多个CLI进程实现并发上传
5. **状态监控**: 通过状态文件实时监控上传进度

## 参数说明

### 必需参数

| 参数 | 说明 |
|------|------|
| `-file` | 要上传的文件路径 |
| `-token` | 钛盘 API Token |
| `-task-id` | 唯一任务ID |
| `-status-file` | 状态文件保存路径 |

### 可选参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-server` | 上传服务器地址 | https://tmplink-sec.vxtrans.com/api_v2 |
| `-chunk-size` | 分块大小(字节) | 3145728 (3MB) |
| `-max-retries` | 最大重试次数 | 3 |
| `-timeout` | 超时时间(秒) | 300 |
| `-model` | 文件有效期 | 0 (24小时) |
| `-mr-id` | 资源ID | 0 (根目录) |
| `-skip-upload` | 跳过上传检查 | 1 (启用秒传) |
| `-debug` | 调试模式 | false |

### 文件有效期选项

- `0`: 24小时 (默认)
- `1`: 3天
- `2`: 7天
- `99`: 永久

## 配置

### Token 获取

需要从钛盘网站获取 API Token，详细步骤请参考 [使用说明](docs/usage.md)。

## 开发命令

### 构建和运行
```bash
make build        # 构建当前平台版本
make release      # 构建所有平台发布版本
make run          # 直接运行
make dist         # 创建发布包
```

### 开发工具
```bash
make deps         # 安装依赖
make fmt          # 格式化代码
make vet          # 代码检查
make test         # 运行测试
make clean        # 清理构建产物
```

## 故障排除

### 常见问题

1. **Token 无效**
   - 检查 token 是否正确复制
   - 确认钛盘账户未过期
   - 重新获取 token

2. **上传失败**
   - 检查网络连接
   - 确认API Token有效
   - 查看状态文件中的错误信息
   - 使用 `-debug` 参数查看详细调试信息
   
3. **上传错误**
   - 检查参数配置是否正确
   - 查看详细错误信息请参考 [API文档](docs/api.md)

### 调试

```bash
# 查看详细日志
./tmplink-cli -debug -file test.txt -token YOUR_TOKEN -task-id test -status-file status.json

# 检查状态文件
cat status.json
```

## 技术栈

- **语言**: Go 1.19+
- **GUI框架**: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **构建工具**: Make

## 更新日志

### v1.0.1 (2025-05-25)

- 🐛 **修复关键bug**: mr_id参数默认值从空字符串改为"0"
- ✅ 修复文件夹查找错误 (status 7 data=8)
- ✅ 正确处理状态码8 (合并完成)
- ✅ 完善API状态码文档
- ✅ 增加测试验证和调试模式

### v1.0.0 (2025-05-24)

- ✅ 实现双进程架构设计
- ✅ 完成GUI和CLI程序分离
- ✅ 实现状态文件通信机制
- ✅ 支持文件分片上传
- ✅ 添加进度监控功能
- ✅ 完善错误处理和重试机制

## 许可证

本项目基于 MIT 许可证开源，详见 [LICENSE](LICENSE) 文件。

## 文档

- [设计文档](docs/design.md) - 系统架构和设计理念
- [API文档](docs/api.md) - 钛盘 API集成规范
- [使用说明](docs/usage.md) - 详细使用指南
