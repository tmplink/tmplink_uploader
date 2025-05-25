# 钛盘上传工具设计文档

## 系统架构

### 双进程架构设计

钛盘上传工具采用双进程架构，将用户界面和文件上传逻辑分离：

```
┌─────────────────┐    JSON 状态文件    ┌─────────────────┐
│   tmplink       │ ◄─────────────────► │  tmplink-cli    │
│   (TUI进程)     │                     │  (上传进程)     │
│                 │                     │                 │
│ • 文件选择      │                     │ • 文件上传      │
│ • 进度显示      │                     │ • 分片处理      │
│ • 任务管理      │                     │ • 状态更新      │
│ • 配置管理      │                     │ • 错误处理      │
└─────────────────┘                     └─────────────────┘
        │                                       │
        └─── 启动和监控 ────────────────────────┘
```

### 设计优势

1. **进程隔离**: 上传进程独立运行，崩溃不影响界面
2. **资源管理**: 每个上传任务独立管理内存和网络资源
3. **并发控制**: 多个 CLI 进程可同时处理不同文件
4. **可维护性**: GUI 和上传逻辑解耦，便于独立开发和测试

## 组件设计

### TUI 组件 (tmplink)

#### 架构模式

采用 Elm 架构模式 (Model-Update-View)：

```go
type TUIModel struct {
    state           ViewState
    list            list.Model
    viewport        viewport.Model
    textInput       textinput.Model
    uploadTasks     []TaskStatus
    config          *Config
    // ...
}

func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m TUIModel) View() string
```

#### 状态管理

```
stateLoading ──→ stateMain ──┬──→ stateFileSelect ──→ stateUploadList
                             │
                             ├──→ stateSettings
                             │
                             └──→ stateUploadList
```

#### 组件层次

```
TUIModel
├── Navigation (list.Model)
├── Content Display (viewport.Model)
├── Input Forms (textinput.Model)
├── Progress Bars (progress.Model)
└── Status Indicators (spinner.Model)
```

### CLI 组件 (tmplink-cli)

#### 处理流程

```
参数解析 ──→ Token验证 ──→ 文件分析 ──→ 服务器选择 ──→ 上传处理 ──→ 状态更新
    │           │           │           │             │           │
    │           │           │           │             │           └─→ JSON状态文件
    │           │           │           │             │
    │           │           │           │             └─→ 分片上传/秒传检查
    │           │           │           │
    │           │           │           └─→ 上传服务器发现
    │           │           │
    │           │           └─→ SHA1计算/文件大小获取
    │           │
    │           └─→ UID获取/Token有效性验证
    │
    └─→ 命令行参数验证
```

#### 核心算法

**Uptoken 生成**:
```go
uptoken := sha1.Sum([]byte(uid + filename + filesize + sliceSize))
```

**分片上传状态机**:
```
准备阶段 ──→ 服务器查询 ──┬──→ 秒传成功 (状态1/6/8)
                        │
                        ├──→ 等待重试 (状态2)
                        │
                        ├──→ 分片上传 (状态3)
                        │
                        └──→ 上传失败 (状态7)
```

## 数据结构设计

### 配置数据

```go
type Config struct {
    Token          string `json:"token"`
    UploadServer   string `json:"upload_server"`
    ChunkSize      int64  `json:"chunk_size"`
    MaxConcurrent  int    `json:"max_concurrent"`
    QuickUpload    bool   `json:"quick_upload"`
    SkipUpload     bool   `json:"skip_upload"`
    Timeout        int    `json:"timeout"`
}
```

### 任务状态

```go
type TaskStatus struct {
    ID         string    `json:"task_id"`
    Status     string    `json:"status"`
    FilePath   string    `json:"file_path"`
    FileName   string    `json:"file_name"`
    FileSize   int64     `json:"file_size"`
    Uploaded   int64     `json:"uploaded"`
    Progress   float64   `json:"progress"`
    Speed      string    `json:"speed"`
    ETA        string    `json:"eta"`
    Error      string    `json:"error"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}
```

### API 响应

```go
type PrepareResponse struct {
    Status int         `json:"status"`
    Data   interface{} `json:"data"`
    Msg    string      `json:"msg"`
}

type UploadResult struct {
    FileKey   string `json:"file_key"`
    FileName  string `json:"file_name"`
    FileSize  int64  `json:"file_size"`
    UploadURL string `json:"upload_url"`
}
```

## API 集成设计

### 认证流程

```
浏览器 localStorage ──→ Token获取 ──→ API验证 ──→ UID获取
                                    │
                                    └─→ 用户信息缓存
```

### 上传流程

```
文件选择 ──→ SHA1计算 ──→ 上传准备 ──→ 服务器分配 ──→ 分片上传
    │           │           │           │          │
    │           │           │           │          └─→ 进度回调
    │           │           │           │
    │           │           │           └─→ 服务器URL获取
    │           │           │
    │           │           └─→ Uptoken生成
    │           │
    │           └─→ 文件完整性校验
    │
    └─→ 文件大小/类型验证
```

### 错误处理机制

采用快速失败策略，遇到错误立即返回详细错误信息：

```go
type ErrorConfig struct {
    EnableDebug   bool
    LogLevel      string
    TimeoutMs     int
}

func fastFail(err error, config ErrorConfig) error {
    if config.EnableDebug {
        log.Printf("ERROR: %v", err)
    }
    return fmt.Errorf("操作失败: %w", err)
}
```

## 并发设计

### TUI 并发模型

```go
// 主 UI 协程
go func() {
    for {
        select {
        case msg := <-uiChannel:
            model, cmd := model.Update(msg)
        case <-ticker.C:
            // 定期刷新状态
        }
    }
}()

// 状态监控协程
go func() {
    for {
        // 扫描状态文件
        // 更新任务状态
        time.Sleep(refreshInterval)
    }
}()

// CLI 进程管理
for _, file := range selectedFiles {
    go startCLIProcess(file, config)
}
```

### CLI 并发模型

```go
// 分片上传并发
semaphore := make(chan struct{}, maxConcurrent)
var wg sync.WaitGroup

for _, chunk := range chunks {
    wg.Add(1)
    go func(chunk ChunkData) {
        defer wg.Done()
        semaphore <- struct{}{}
        defer func() { <-semaphore }()
        
        uploadChunk(chunk)
    }(chunk)
}
wg.Wait()
```

## 错误处理设计

### 错误分类

1. **系统错误**: 文件系统、网络连接问题
2. **API 错误**: 服务器返回的错误状态
3. **用户错误**: 参数配置、权限问题
4. **业务错误**: 文件大小、格式限制

### 错误处理策略

```go
type ErrorHandler struct {
    Logger    *log.Logger
    Notifier  func(error)
}

func (h *ErrorHandler) Handle(err error) error {
    switch e := err.(type) {
    case *NetworkError:
        return h.logAndReturn(e)
    case *APIError:
        return h.logAndReturn(e)
    case *UserError:
        return h.logAndReturn(e)
    default:
        return h.logAndReturn(e)
    }
}
```

## 性能优化设计

### 内存管理

1. **流式处理**: 大文件分块读取，避免全部加载到内存
2. **缓冲池**: 复用分片缓冲区，减少内存分配
3. **及时释放**: 上传完成后立即释放资源

### 网络优化

1. **连接复用**: HTTP 连接池管理
2. **并发控制**: 限制同时连接数，避免服务器拒绝
3. **断点续传**: 支持网络中断后继续上传

### UI 响应性

1. **异步操作**: 所有 IO 操作在后台协程执行
2. **渐进式加载**: 大列表分页显示
3. **状态缓存**: 避免重复计算和 API 调用

## 安全设计

### Token 管理

```go
type TokenManager struct {
    token       string
    encryptedAt time.Time
    validUntil  time.Time
}

func (tm *TokenManager) Encrypt() error {
    // 使用系统密钥加密存储
}

func (tm *TokenManager) Validate() error {
    // API 调用验证 token 有效性
}
```

### 文件安全

1. **路径验证**: 防止路径遍历攻击
2. **大小限制**: 限制单文件和总大小
3. **类型检查**: 验证文件格式合法性

## 扩展设计

### 插件架构

```go
type Uploader interface {
    Upload(file string, config Config) (*UploadResult, error)
    GetProgress() float64
    Cancel() error
}

type UploaderRegistry struct {
    uploaders map[string]Uploader
}

func (r *UploaderRegistry) Register(name string, uploader Uploader) {
    r.uploaders[name] = uploader
}
```

### 配置扩展

```go
type ProviderConfig interface {
    Validate() error
    GetEndpoint() string
    GetCredentials() interface{}
}

type ConfigManager struct {
    providers map[string]ProviderConfig
}
```

## 测试设计

### 单元测试

- API 客户端测试
- 文件处理逻辑测试
- 状态管理测试

### 集成测试

- TUI-CLI 通信测试
- 完整上传流程测试
- 错误场景测试

### 性能测试

- 大文件上传测试
- 并发上传测试
- 内存泄漏测试

## 部署设计

### 构建流程

```makefile
build: deps fmt vet
    go build -ldflags "$(LDFLAGS)" -o ./tmplink ./cmd/tmplink
    go build -ldflags "$(LDFLAGS)" -o ./tmplink-cli ./cmd/tmplink-cli

release: build-all
    tar -czf tmplink-$(VERSION)-$(OS)-$(ARCH).tar.gz tmplink tmplink-cli docs/
```

### 发布策略

1. **版本管理**: 语义化版本控制
2. **兼容性**: 配置文件向后兼容
3. **更新机制**: 自动检查和提示更新
