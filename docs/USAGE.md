# TmpLink Uploader 使用说明

## 概述

TmpLink Uploader 是一个双进程架构的文件上传工具，支持钛盘 (TmpLink) 平台的文件上传服务。包含图形用户界面 (TUI) 和命令行接口 (CLI) 两个组件。

## 程序组件

### tmplink - 图形用户界面 (TUI)

基于 Bubble Tea 框架的终端用户界面，提供交互式文件管理和上传功能。

#### 启动程序

```bash
./tmplink
```

#### 功能特性

- **文件选择**: 浏览目录并选择要上传的文件
- **批量上传**: 支持多文件同时上传
- **实时进度**: 显示每个文件的上传进度
- **任务管理**: 查看上传任务状态和历史
- **配置管理**: 设置上传参数和 API token
- **错误处理**: 自动重试和错误提示

#### 界面操作

**主菜单**:
- `↑/↓` - 导航菜单项
- `Enter` - 选择菜单项
- `q` - 退出程序

**文件选择**:
- `↑/↓` - 浏览文件列表
- `Enter` - 选择文件进行上传
- `Esc` - 返回上级菜单
- `→` - 进入子目录
- `←` - 返回父目录

**上传列表**:
- `↑/↓` - 浏览上传任务
- `r` - 刷新任务状态
- `Esc` - 返回主菜单

**设置页面**:
- `↑/↓` - 选择设置项
- `Enter` - 编辑设置值
- `Tab` - 切换输入框
- `Esc` - 保存并返回

#### 配置参数

设置页面可配置以下参数：

- **Token**: TmpLink API 访问令牌
- **分片大小**: 文件上传的块大小 (1MB-80MB)
- **并发数**: 同时上传的文件数量
- **超时时间**: 网络请求超时时间 (秒)

### tmplink-cli - 命令行接口

独立的命令行上传工具，用于单文件上传操作。

#### 基本用法

```bash
./tmplink-cli -file <文件路径> -token <API令牌> -task-id <任务ID> -status-file <状态文件>
```

#### 必需参数

- `-file`: 要上传的文件路径
- `-token`: TmpLink API 访问令牌
- `-task-id`: 唯一任务标识符
- `-status-file`: JSON 状态文件路径，用于进程间通信

#### 可选参数

- `-server`: 上传服务器 URL (默认: https://tmplink-upload-acc.vxtrans.com/app/upload_slice)
- `-chunk-size`: 分片大小，字节 (默认: 3MB, 最大: 80MB)
- `-max-retries`: 最大重试次数 (默认: 3)
- `-timeout`: 请求超时时间，秒 (默认: 300)
- `-model`: 文件有效期 (默认: 0=24小时, 1=3天, 2=7天, 99=永久)
- `-mr-id`: 资源ID (默认: "0"=根目录)
- `-skip-upload`: 跳过上传标志 (默认: 1, 启用秒传检查)
- `-debug`: 调试模式，输出详细运行信息

**重要更新 (v1.0.1)**:
- **修复关键bug**: mr_id参数默认值从空字符串改为"0"
- **状态码更新**: 状态7(data=8)表示"文件夹未找到"错误，状态8表示"合并完成"
- 正确处理API响应状态码，避免将错误误认为成功

**注意**: 
- API 接口地址为: https://tmplink-sec.vxtrans.com/api_v2/file
- 上传服务器地址为: https://tmplink-upload-acc.vxtrans.com/app/upload_slice

#### 使用示例

**基本上传**:
```bash
./tmplink-cli -file document.pdf -token "your_token_here" -task-id "upload-001" -status-file "status.json"
```

**自定义分片大小**:
```bash
./tmplink-cli -file largefile.zip -token "your_token_here" -task-id "upload-002" -status-file "status.json" -chunk-size 10485760
```

**指定服务器**:
```bash
./tmplink-cli -file video.mp4 -token "your_token_here" -task-id "upload-003" -status-file "status.json" -server "https://tmplink-upload-acc.vxtrans.com/app/upload_slice"
```

#### 状态文件格式

CLI 通过 JSON 状态文件与 TUI 通信：

```json
{
  "task_id": "upload-001",
  "status": "uploading",
  "file_path": "/path/to/file.pdf",
  "file_name": "file.pdf",
  "file_size": 1048576,
  "uploaded": 524288,
  "progress": 50.0,
  "speed": "1.2 MB/s",
  "eta": "30s",
  "error": "",
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:30Z"
}
```

## Token 获取

1. 访问 https://tmp.link/ 并登录
2. 打开浏览器开发者工具 (F12)
3. 在 Console 中执行: `localStorage.getItem('token')`
4. 复制返回的 token 值

## 构建和安装

### 编译程序

```bash
# 编译所有组件
make build

# 仅编译 TUI
make build-gui

# 仅编译 CLI
make build-cli

# 跨平台编译
make build-all
```

### 安装依赖

```bash
make deps
```

### 代码检查

```bash
# 格式化代码
make fmt

# 静态分析
make vet

# 运行测试
make test
```

## 故障排除

### 常见问题

**Token 无效**:
- 检查 token 是否正确复制
- 确认 TmpLink 账户未过期
- 重新获取 token

**上传失败**:
- 检查网络连接
- 确认文件大小在限制范围内
- 查看错误日志

**进度显示异常**:
- 检查状态文件权限
- 确认磁盘空间充足
- 重新启动程序

### 日志和调试

程序运行时会输出详细的状态信息，包括：
- API 请求和响应
- 文件分析结果
- 上传进度和错误信息

## 注意事项

1. **文件大小限制**: 单文件最大支持 80MB 分片
2. **并发限制**: 建议同时上传文件数不超过 10 个
3. **网络要求**: 需要稳定的互联网连接
4. **存储空间**: 确保目标平台有足够存储空间
5. **Token 安全**: 妥善保管 API token，避免泄露

## 技术支持

如遇到问题，请检查：
1. 程序版本和依赖
2. 网络连接状态
3. API token 有效性
4. 文件权限设置
