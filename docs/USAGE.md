# TmpLink Uploader 使用指南

本文档提供详细的使用说明，包括安装、配置和操作指引。

## 快速开始

### 安装

#### 从源码构建
```bash
git clone https://github.com/your-username/tmplink_uploader.git
cd tmplink_uploader
make build
```

#### 使用预构建版本
从 [Releases](https://github.com/your-username/tmplink_uploader/releases) 页面下载对应平台的预构建版本。

### 基本使用

#### 1. 获取 API Token
1. 访问 [TmpLink](https://tmp.link/) 并登录
2. 打开浏览器开发者工具 (F12)
3. 在控制台执行: `localStorage.getItem('token')`
4. 复制返回的 token 值

#### 2. 启动 GUI 程序
```bash
./tmplink
```

#### 3. 直接使用 CLI 程序
```bash
./tmplink-cli -file /path/to/file.txt -token YOUR_TOKEN -task-id unique-id -status-file status.json
```

## GUI 程序详细说明

### 界面导航

#### 主菜单
- `↑/↓` - 导航菜单项
- `Enter` - 选择菜单项  
- `q` - 退出程序

#### 文件选择界面
- `↑/↓` - 浏览文件列表
- `Enter` - 选择文件进行上传
- `→` - 进入子目录
- `←` - 返回父目录
- `Esc` - 返回主菜单

#### 上传列表界面  
- `↑/↓` - 浏览上传任务
- `r` - 刷新任务状态
- `d` - 复制下载链接
- `Esc` - 返回主菜单

#### 设置界面
- `↑/↓` - 选择设置项
- `Enter` - 编辑设置值
- `Tab` - 切换输入框
- `Esc` - 保存并返回

### 配置参数

#### Token 设置
- **必需**: TmpLink API 访问令牌
- **获取方式**: 从浏览器 localStorage 获取
- **注意**: Token 会加密保存到配置文件

#### 上传参数
- **分片大小**: 1MB-80MB，推荐 3MB
- **并发数**: 同时上传的文件数，推荐 3-5
- **超时时间**: 网络请求超时，默认 300 秒
- **文件有效期**: 0=24小时, 1=3天, 2=7天, 99=永久

#### 高级选项
- **调试模式**: 输出详细日志信息
- **跳过上传**: 启用秒传检查，默认开启
- **资源目录**: 指定上传目录，默认根目录

## CLI 程序详细说明

### 命令行参数

#### 必需参数
```bash
-file /path/to/file        # 文件路径
-token YOUR_API_TOKEN      # API 令牌
-task-id unique-task-id    # 任务ID
-status-file status.json   # 状态文件路径
```

#### 可选参数
```bash
-server URL               # 服务器地址
-chunk-size 3145728       # 分片大小(字节)
-max-retries 3            # 最大重试次数
-timeout 300              # 超时时间(秒)
-model 0                  # 文件有效期
-mr-id "0"                # 资源ID
-skip-upload 1            # 启用秒传
-debug                    # 调试模式
```

### 状态文件格式

CLI 程序通过 JSON 状态文件与 GUI 通信：

```json
{
  "id": "task_1640995200",
  "status": "uploading",
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

#### 状态值说明
- `pending`: 任务创建，准备开始
- `in_progress`: 正在上传
- `completed`: 上传完成
- `failed`: 上传失败

### 使用示例

#### 基本上传
```bash
./tmplink-cli \
  -file document.pdf \
  -token "your_token_here" \
  -task-id "upload-001" \
  -status-file "status.json"
```

#### 大文件上传 (10MB 分片)
```bash
./tmplink-cli \
  -file largefile.zip \
  -token "your_token_here" \
  -task-id "upload-002" \
  -status-file "status.json" \
  -chunk-size 10485760
```

#### 调试模式上传
```bash
./tmplink-cli \
  -file test.txt \
  -token "your_token_here" \
  -task-id "upload-003" \
  -status-file "status.json" \
  -debug
```

## 进阶功能

### 批量上传

#### 使用脚本批量上传
```bash
#!/bin/bash
TOKEN="your_token_here"
FILES=("file1.txt" "file2.txt" "file3.txt")

for i in "${!FILES[@]}"; do
  TASK_ID="batch_upload_$i"
  STATUS_FILE="status_$i.json"
  
  ./tmplink-cli \
    -file "${FILES[$i]}" \
    -token "$TOKEN" \
    -task-id "$TASK_ID" \
    -status-file "$STATUS_FILE" &
done

wait  # 等待所有上传完成
```

### 监控上传进度

#### 实时监控状态文件
```bash
#!/bin/bash
STATUS_FILE="status.json"

while true; do
  if [ -f "$STATUS_FILE" ]; then
    STATUS=$(cat "$STATUS_FILE" | jq -r '.status')
    PROGRESS=$(cat "$STATUS_FILE" | jq -r '.progress')
    
    echo "状态: $STATUS, 进度: $PROGRESS%"
    
    if [ "$STATUS" = "completed" ] || [ "$STATUS" = "failed" ]; then
      break
    fi
  fi
  
  sleep 2
done
```

### 错误处理和重试

#### 自动重试脚本
```bash
#!/bin/bash
MAX_ATTEMPTS=3
ATTEMPT=1

while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
  echo "尝试第 $ATTEMPT 次上传..."
  
  ./tmplink-cli \
    -file "$1" \
    -token "$2" \
    -task-id "retry_$ATTEMPT" \
    -status-file "status_retry.json"
  
  if [ $? -eq 0 ]; then
    echo "上传成功!"
    break
  else
    echo "上传失败，$(($MAX_ATTEMPTS - $ATTEMPT)) 次重试机会剩余"
    ATTEMPT=$((ATTEMPT + 1))
    sleep 5
  fi
done
```

## 配置文件

### 配置文件位置
- Linux/macOS: `~/.tmplink_config.json`
- Windows: `%USERPROFILE%/.tmplink_config.json`

### 配置文件格式
```json
{
  "token": "your_api_token",
  "upload_server": "https://tmplink-sec.vxtrans.com/api_v2",
  "chunk_size": 3145728,
  "max_concurrent": 5,
  "timeout": 300,
  "model": 0,
  "mr_id": "0",
  "skip_upload": 1,
  "debug": false
}
```

### 配置项说明
- `token`: API 访问令牌
- `upload_server`: 上传服务器地址
- `chunk_size`: 分片大小(字节)
- `max_concurrent`: 最大并发数
- `timeout`: 请求超时时间(秒)
- `model`: 文件有效期设置
- `mr_id`: 默认资源目录
- `skip_upload`: 是否启用秒传
- `debug`: 是否开启调试模式

## 故障排除

### 常见问题

#### Token 相关
**问题**: Token 无效错误
**解决**:
1. 确认从正确位置获取 token
2. 检查 token 是否完整复制
3. 确认账户未过期或被限制

#### 上传失败
**问题**: 文件上传失败
**解决**:
1. 检查网络连接稳定性
2. 确认文件大小在限制范围内
3. 使用 `-debug` 参数查看详细错误
4. 检查磁盘空间是否充足

#### 状态码 7 错误
**问题**: 收到状态码 7 错误
**解决**:
1. 检查 `mr_id` 参数设置
2. 确认目标文件夹存在
3. 验证用户权限

#### 进度显示异常
**问题**: 进度不更新或显示错误
**解决**:
1. 检查状态文件权限
2. 确认状态文件路径正确
3. 重启程序重新初始化

### 日志和调试

#### 启用调试模式
```bash
./tmplink-cli -debug -file test.txt -token YOUR_TOKEN -task-id test -status-file status.json
```

#### 查看详细日志
调试模式会输出：
- API 请求和响应详情
- 文件处理进度信息
- 错误堆栈信息
- 网络连接状态

#### 状态文件检查
```bash
# 查看当前状态
cat status.json | jq '.'

# 监控状态变化
watch -n 1 'cat status.json | jq ".status, .progress"'
```

## 性能优化

### 上传性能调优

#### 分片大小选择
- **小文件 (<10MB)**: 使用 1-3MB 分片
- **中等文件 (10MB-100MB)**: 使用 3-10MB 分片  
- **大文件 (>100MB)**: 使用 10-80MB 分片

#### 并发数设置
- **网络良好**: 可设置 5-10 并发
- **网络一般**: 建议 3-5 并发
- **网络较差**: 使用 1-2 并发

#### 超时时间设置
- **本地网络**: 60-120 秒
- **公共网络**: 180-300 秒
- **移动网络**: 300-600 秒

### 系统资源优化

#### 内存使用
- 分片大小直接影响内存占用
- 大文件建议分多次上传
- 监控系统内存使用情况

#### 网络带宽
- 根据带宽调整并发数
- 避免占满带宽影响其他应用
- 考虑设置速率限制

## 高级配置

### 网络代理设置
```bash
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=http://proxy.example.com:8080
./tmplink-cli [参数...]
```

### 自定义服务器
```bash
./tmplink-cli \
  -server "https://custom-server.example.com/api" \
  [其他参数...]
```

### 环境变量配置
```bash
export TMPLINK_TOKEN="your_token"
export TMPLINK_CHUNK_SIZE="5242880"
export TMPLINK_DEBUG="true"

./tmplink-cli -file test.txt -task-id test -status-file status.json
```