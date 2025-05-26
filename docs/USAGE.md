# 钛盘上传工具使用指南

本文档提供详细的使用说明，包括安装、配置和操作指引。

## 快速开始

### 获取API Token

使用前需要获取钛盘API Token：

1. 访问 [钛盘](https://tmp.link/) 并登录
2. 点击上传文件按钮
3. 点击重新设定按钮
4. 找到使用 CLI 上传工具上传，并在对应的界面中复制 Token

⚠️ **重要提示**: Token 与您的账号相互关联，绝对不能泄露给他人。同时，如果在网页端退出了登录，Token 关联的用户信息也会分离。

### 基本使用

#### GUI模式（推荐新手）

```bash
# 启动GUI程序
./tmplink

# 首次使用会提示输入API Token
# 然后可以通过界面选择文件并上传
```

#### CLI模式（推荐高级用户）

```bash
# 首次设置token
./tmplink-cli -set-token YOUR_TOKEN

# 上传文件
./tmplink-cli -file /path/to/file.txt
```

## GUI 程序详细说明

### 界面导航

#### 主菜单
- `↑/↓` - 导航菜单项
- `Enter` - 选择菜单项  
- `q` - 退出程序

#### 文件选择界面
- `↑/↓` - 浏览文件和目录
- `Enter` - 进入目录或选择文件上传
- `..` - 返回上级目录
- `t` - 切换显示隐藏文件
- `Tab` - 切换到设置界面
- `Esc` - 返回主菜单

### 权限系统

本工具支持基于用户赞助状态的分级功能：

#### 普通用户功能
- 文件上传和下载
- 基础设置和标准上传功能
- 自动服务器选择
- 标准上传速度监控

#### 赞助用户专享功能 ✨
- 高级上传设置：分块大小和并发数控制
- 手动服务器选择（从API动态获取）
- 快速上传开关（启用/禁用秒传检查）
- 优先服务器访问

### 设置界面

#### 基础设置（所有用户）
- Token配置
- 文件有效期设置

#### 高级设置（仅赞助用户）
- **分块大小**: 1-80MB（默认3MB）🔒
- **并发数**: 1-20（默认5）🔒  
- **服务器选择**: 从API动态获取的服务器列表🔒
- **快速上传**: 开启/关闭秒传检查🔒

#### 设置界面操作
- `↑/↓` - 在设置项间导航
- `←/→` - 切换服务器或快速上传开关（仅赞助用户）
- `Space` - 切换快速上传开关
- `Enter` - 保存设置
- `Tab` - 切换到上传管理界面
- `Esc` - 返回主菜单

### 上传管理界面
- `↑/↓` - 浏览上传任务列表
- `d` - 删除选中的上传任务
- `Tab` - 切换到主界面
- `Esc` - 返回主菜单

#### 上传任务显示信息
- **文件名**: 正在上传的文件名称
- **状态**: pending（等待）/uploading（上传中）/completed（完成）/failed（失败）
- **进度**: 上传进度百分比
- **速度**: 实时上传速度（MB/s）
- **完成时间**: 上传完成或失败的时间戳

#### 上传速度计算
- 使用加权平均算法确保速度显示稳定
- 显示当前活跃上传的实时速度
- 保留已完成上传的最终速度记录

## CLI 程序详细说明

### 运行模式

CLI程序根据是否提供 `-task-id` 参数自动选择运行模式：

#### CLI模式（推荐个人使用）
**触发条件**: 不提供 `-task-id` 参数
```bash
./tmplink-cli -file document.pdf  # 自动CLI模式
```

**特性**:
- 🎯 **实时进度条**: 类似wget/curl的可视化进度显示
- ⚡ **速度监控**: 显示当前上传速度和平均速度
- ⏱️ **ETA计算**: 预计剩余完成时间
- 🎨 **美观输出**: 使用emojis和颜色增强用户体验
- 📊 **详细统计**: 显示文件大小、总耗时等信息

#### GUI模式（程序内部调用）
**触发条件**: 提供 `-task-id` 参数
```bash
./tmplink-cli -file document.pdf -task-id upload_123  # GUI模式
```

**特性**:
- 📄 **状态文件输出**: 将进度写入JSON状态文件
- 🔄 **后台静默运行**: 适合被其他程序调用
- 📡 **进程间通信**: 通过状态文件与GUI程序通信

### 命令行参数

#### 必需参数
```bash
-file /path/to/file        # 文件路径
```

**Token要求：** 必须通过以下方式之一提供API token：
- 使用 `-set-token` 预先保存到配置文件
- 使用 `-token` 参数临时提供

#### 配置设置（首次使用）
```bash
-set-token YOUR_API_TOKEN  # 设置并保存API token
-set-model 2              # 设置默认文件有效期为7天
-set-mr-id folder123      # 设置默认目录ID
```

#### 可选参数

**上传控制参数**
```bash
-chunk-size 3             # 分片大小(MB, 1-99)，默认3MB  
-model 0                  # 文件有效期（默认: 已保存值或0=24小时）
-mr-id folder123          # 目录ID（默认: 已保存值或0=根目录）
-skip-upload 1            # 启用秒传检查（默认: 1=启用）
```

**服务器选择参数**
```bash
-upload-server URL        # 强制指定上传服务器地址（可选，留空自动选择）
-server-name Global       # 上传服务器名称，仅用于显示（可选）
```

**身份认证参数**
```bash
-token YOUR_API_TOKEN     # 临时使用的API token（默认: 使用已保存值）
```

**任务管理参数**
```bash
-task-id upload_123       # 任务标识符（默认: 自动生成）
-status-file status.json  # 状态文件路径（默认: 自动生成）
```

**调试参数**
```bash
-debug                    # 启用调试模式，输出详细日志（默认: false）
```

#### 服务器架构说明

**API服务器（固定）**：
- API服务器地址固定为 `https://tmplink-sec.vxtrans.com/api_v2`
- 用于token验证、文件信息提交、获取上传服务器列表等
- 不可通过参数修改，确保程序稳定性

**上传服务器选择**：
- **自动选择**: 不使用 `-upload-server` 参数，由API自动分配最佳服务器
- **手动选择**: 使用 `-upload-server` 强制指定特定服务器节点
- **可用服务器**: 通过 `upload_request_select2` API动态获取，包括：
  - Global（全球节点）
  - JP（日本节点）  
  - CN（中国节点）
  - HD1/HD2（高清节点）
  - C2（荷兰节点）

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
  "upload_speed": 2.5,
  "download_url": "",
  "error_msg": "",
  "created_at": "2023-12-31T12:00:00Z",
  "updated_at": "2023-12-31T12:01:00Z",
  "process_id": 12345
}
```

#### 状态值说明
- `pending`: 任务创建，准备开始
- `uploading`: 正在上传
- `completed`: 上传完成
- `failed`: 上传失败

#### 新增字段说明
- `upload_speed`: 实时上传速度（MB/s），使用加权平均算法计算
- `process_id`: CLI进程ID，用于进程管理
- 速度计算考虑最近10次测量的加权平均，确保显示稳定性
- 完成的上传保留最终速度，失败的上传速度为0

### 使用示例

#### 首次设置Token
```bash
./tmplink-cli -set-token your_token_here
```

#### 基本上传
```bash
./tmplink-cli -file document.pdf
```

#### 大文件上传 (10MB 分片)
```bash
./tmplink-cli -file largefile.zip -chunk-size 10
```

#### 使用临时Token（覆盖保存的Token）
```bash
./tmplink-cli -file document.pdf -token temporary_token
```

#### 配置设置示例
```bash
# 设置所有常用配置
./tmplink-cli -set-token your_token -set-model 2 -set-mr-id folder123

# 单独设置有效期
./tmplink-cli -set-model 99

# 单独设置目录ID  
./tmplink-cli -set-mr-id 0
```

#### 调试模式上传
```bash
./tmplink-cli -file test.txt -debug
```

#### CLI进度显示示例

当不提供 `-task-id` 参数时，CLI会显示类似wget/curl的进度条：

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

## 进阶功能

### 批量上传

#### 使用脚本批量上传
```bash
#!/bin/bash
# 首次设置token
./tmplink-cli -set-token your_token_here

FILES=("file1.txt" "file2.txt" "file3.txt")

for file in "${FILES[@]}"; do
  ./tmplink-cli -file "$file" &
done

wait  # 等待所有上传完成
```

### 监控上传进度

#### 实时监控状态文件
```bash
#!/bin/bash
STATUS_FILE=status.json

while true; do
  if [ -f "$STATUS_FILE" ]; then
    STATUS=$(cat "$STATUS_FILE" | jq -r '.status')
    PROGRESS=$(cat "$STATUS_FILE" | jq -r '.progress')
    
    echo "状态: $STATUS, 进度: $PROGRESS%"
    
    if [ "$STATUS" = completed ] || [ "$STATUS" = failed ]; then
      break
    fi
  fi
  
  sleep 2
done
```

### 错误处理

CLI程序采用快速失败策略，遇到错误立即退出并返回详细错误信息。如需重试，请使用外部脚本或重新执行命令。

## 配置管理

### 配置存储
CLI工具会将配置安全保存在用户配置目录：
- Linux/macOS: `~/.tmplink_cli_config.json`
- Windows: `%USERPROFILE%/.tmplink_cli_config.json`

### 参数使用优先级
1. **命令行参数**（最高优先级）
2. **已保存的配置文件**
3. **程序内置默认值**

### 各参数默认值说明
- **token**: 无内置默认值，使用已保存值，无保存值则必须通过命令行提供
- **model**: 内置默认值为0（24小时），优先使用已保存值
- **mr_id**: 内置默认值为"0"（根目录），优先使用已保存值
- **其他参数**: 使用程序内置默认值

### 配置管理命令
```bash
# 设置并保存token
./tmplink-cli -set-token your_new_token

# 设置默认文件有效期
./tmplink-cli -set-model 2

# 设置默认目录ID
./tmplink-cli -set-mr-id folder123

# 一次设置多个配置
./tmplink-cli -set-token token -set-model 99 -set-mr-id 0

# 查看保存的配置文件
cat ~/.tmplink_cli_config.json

# 删除保存的配置
rm ~/.tmplink_cli_config.json
```

### 配置文件格式
```json
{
  "token": "your_api_token",
  "model": 2,
  "mr_id": "folder123"
}
```

## 配置文件

### GUI配置文件位置
- Linux/macOS: `~/.tmplink_config.json`
- Windows: `%USERPROFILE%/.tmplink_config.json`

### 配置文件格式
```json
{
  "token": "your_api_token",
  "upload_server": "https://tmplink-sec.vxtrans.com/api_v2",
  "selected_server_name": "Global",
  "chunk_size": 3,
  "max_concurrent": 5,
  "quick_upload": true,
  "skip_upload": true
}
```

### 配置项说明
- `token`: API 访问令牌
- `upload_server`: 上传服务器地址
- `selected_server_name`: 选中的服务器名称
- `chunk_size`: 分片大小(MB)
- `max_concurrent`: 最大并发数
- `quick_upload`: 是否启用快速上传
- `skip_upload`: 是否启用秒传检查

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
./tmplink-cli -debug -file test.txt
```

#### 查看详细日志
调试模式会输出：
- API 请求和响应详情
- 文件处理进度信息
- 错误堆栈信息
- 网络连接状态

#### 状态文件检查
```bash
# 查看当前状态（文件名自动生成）
cat upload_*_status.json | jq '.'

# 监控状态变化
watch -n 1 'cat upload_*_status.json | jq ".status, .progress"'
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
  -server https://custom-server.example.com/api \
  [其他参数...]
```

### 环境变量配置
```bash
export TMPLINK_TOKEN=your_token
export TMPLINK_CHUNK_SIZE=5
export TMPLINK_DEBUG=true

./tmplink-cli -file test.txt
```

## 文件有效期说明

钛盘支持多种文件有效期选项：

| 值 | 说明 | 适用场景 |
|---|------|----------|
| `0` | 24小时 | 临时文件分享 |
| `1` | 3天 | 短期项目文件 |
| `2` | 7天 | 中期协作文件 |
| `99` | 无限期 | 重要文档存档 |

**使用建议**：
- 临时分享文件使用24小时
- 项目协作文件使用7天
- 重要文档可选择无限期
- 定期清理过期文件节省空间

## 最佳实践

### 安全建议

1. **Token管理**：
   - 定期更换API Token
   - 不要在公共环境保存Token
   - 使用 `-token` 参数进行临时操作

2. **文件安全**：
   - 上传前检查文件内容
   - 避免上传包含敏感信息的文件
   - 合理设置文件有效期

3. **网络安全**：
   - 在可信网络环境下使用
   - 注意防火墙和代理设置

### 性能建议

1. **大文件处理**：
   - 根据网络情况调整分片大小
   - 使用合适的并发数
   - 避免在网络高峰期上传

2. **批量操作**：
   - 使用脚本进行批量上传
   - 合理控制并发任务数量
   - 监控系统资源使用

3. **错误处理**：
   - 使用调试模式排查问题
   - 保存重要的错误日志
   - 建立重试机制（脚本层面）

---

更多技术细节请参考 [技术文档](technical.md)。