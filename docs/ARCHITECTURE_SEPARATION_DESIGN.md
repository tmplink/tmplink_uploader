# 架构分离设计文档

## 概述

将原有的单体程序分离为两个独立的程序：
- `tmplink`: GUI界面程序（TUI界面）
- `tmplink-cli`: 核心上传功能的CLI程序

## 架构设计

### 整体架构图

```
┌─────────────────┐    文件通信    ┌─────────────────┐
│                 │ ◄------------- │                 │
│    tmplink      │                │  tmplink-cli    │
│   (GUI程序)     │ -------------► │  (核心上传)     │
│                 │    进程控制    │                 │
└─────────────────┘                └─────────────────┘
         │                                   │
         └───────── 共享任务状态文件 ─────────┘
```

### 分离原则

1. **职责分离**: GUI负责用户交互，CLI负责文件上传
2. **进程独立**: 两个程序可以独立运行和停止
3. **状态共享**: 通过文件系统共享任务状态
4. **接口标准**: 定义标准的通信协议

## 目录结构设计

```
tmplink_uploader/
├── cmd/
│   ├── tmplink/            # GUI程序入口
│   │   └── main.go
│   └── tmplink-cli/        # CLI程序入口
│       └── main.go
├── internal/
│   ├── common/             # 共享组件
│   │   ├── config/         # 配置管理
│   │   ├── client/         # API客户端
│   │   └── types/          # 共享类型定义
│   ├── gui/                # GUI相关代码
│   │   ├── tui/            # TUI界面
│   │   └── controller/     # GUI控制逻辑
│   ├── cli/                # CLI相关代码
│   │   ├── uploader/       # 上传核心逻辑
│   │   └── worker/         # 工作进程管理
│   └── shared/             # 共享通信组件
│       ├── taskfile/       # 任务状态文件管理
│       └── ipc/            # 进程间通信
├── pkg/                    # 公共包
└── docs/                   # 文档目录
```

## 通信协议设计

### 任务状态文件格式

每个上传任务创建一个JSON状态文件：`~/.tmplink/tasks/{task_id}.json`

```json
{
  "task_id": "uuid-string",
  "pid": 12345,
  "status": "running|completed|failed|paused",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:35:00Z",
  "file_info": {
    "path": "/path/to/file.txt",
    "name": "file.txt",
    "size": 1048576,
    "sha1": "abc123..."
  },
  "upload_info": {
    "server_url": "https://upload.server.com",
    "token": "upload_token",
    "chunk_size": 3145728,
    "total_chunks": 5,
    "completed_chunks": 3
  },
  "progress": {
    "bytes_uploaded": 524288,
    "bytes_total": 1048576,
    "speed_bps": 1048576,
    "eta_seconds": 10
  },
  "result": {
    "download_url": "https://tmp.link/abc123",
    "error_message": null
  }
}
```

### 控制命令接口

GUI通过命令行参数控制CLI：

```bash
# 启动上传任务
tmplink-cli upload --file="/path/to/file" --task-id="uuid" --config="/path/to/config"

# 暂停任务
tmplink-cli pause --task-id="uuid"

# 恢复任务
tmplink-cli resume --task-id="uuid"

# 取消任务
tmplink-cli cancel --task-id="uuid"

# 查询任务状态
tmplink-cli status --task-id="uuid"

# 清理完成的任务
tmplink-cli cleanup --task-id="uuid"
```

## 接口定义

### 共享类型定义

```go
// TaskStatus 任务状态
type TaskStatus string

const (
    TaskStatusPending   TaskStatus = "pending"
    TaskStatusRunning   TaskStatus = "running"
    TaskStatusPaused    TaskStatus = "paused"
    TaskStatusCompleted TaskStatus = "completed"
    TaskStatusFailed    TaskStatus = "failed"
    TaskStatusCanceled  TaskStatus = "canceled"
)

// TaskInfo 任务信息
type TaskInfo struct {
    TaskID    string    `json:"task_id"`
    PID       int       `json:"pid"`
    Status    TaskStatus `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    FileInfo  FileInfo  `json:"file_info"`
    UploadInfo UploadInfo `json:"upload_info"`
    Progress  ProgressInfo `json:"progress"`
    Result    ResultInfo `json:"result"`
}

// FileInfo 文件信息
type FileInfo struct {
    Path string `json:"path"`
    Name string `json:"name"`
    Size int64  `json:"size"`
    SHA1 string `json:"sha1"`
}

// UploadInfo 上传配置信息
type UploadInfo struct {
    ServerURL      string `json:"server_url"`
    Token          string `json:"token"`
    ChunkSize      int64  `json:"chunk_size"`
    TotalChunks    int    `json:"total_chunks"`
    CompletedChunks int   `json:"completed_chunks"`
}

// ProgressInfo 进度信息
type ProgressInfo struct {
    BytesUploaded int64 `json:"bytes_uploaded"`
    BytesTotal    int64 `json:"bytes_total"`
    SpeedBps      int64 `json:"speed_bps"`
    ETASeconds    int   `json:"eta_seconds"`
}

// ResultInfo 结果信息
type ResultInfo struct {
    DownloadURL  string `json:"download_url,omitempty"`
    ErrorMessage string `json:"error_message,omitempty"`
}
```

### 任务管理器接口

```go
// TaskManager 任务管理器接口
type TaskManager interface {
    // 创建新任务
    CreateTask(fileInfo FileInfo, uploadConfig UploadInfo) (*TaskInfo, error)
    
    // 获取任务状态
    GetTask(taskID string) (*TaskInfo, error)
    
    // 列出所有任务
    ListTasks() ([]*TaskInfo, error)
    
    // 更新任务状态
    UpdateTask(taskID string, updates TaskInfo) error
    
    // 删除任务
    DeleteTask(taskID string) error
    
    // 清理完成的任务
    CleanupCompletedTasks() error
}
```

### CLI控制器接口

```go
// CLIController CLI控制器接口
type CLIController interface {
    // 启动上传任务
    StartUpload(taskID string, filePath string) error
    
    // 暂停任务
    PauseTask(taskID string) error
    
    // 恢复任务
    ResumeTask(taskID string) error
    
    // 取消任务
    CancelTask(taskID string) error
    
    // 检查进程状态
    IsProcessRunning(taskID string) bool
    
    // 获取进程ID
    GetProcessID(taskID string) (int, error)
}
```

## 实现步骤

### 阶段1: 基础架构 (高优先级)
1. ✅ 设计架构分离方案和接口定义
2. 🔲 实现任务状态文件格式和管理
3. 🔲 创建tmplink-cli核心上传程序
4. 🔲 重构tmplink GUI程序与CLI通信

### 阶段2: 核心功能 (中优先级)  
5. 🔲 实现进程管理和任务控制
6. 🔲 编写单元测试覆盖核心功能
7. 🔲 实现错误处理和恢复机制

### 阶段3: 优化完善 (低优先级)
8. 🔲 优化性能和资源管理
9. 🔲 编写集成测试和文档

## 优势分析

### 功能优势
- **独立性**: CLI可以独立使用，适合自动化场景
- **稳定性**: GUI崩溃不影响上传任务继续执行
- **扩展性**: 可以轻松添加其他GUI前端或API接口
- **监控性**: 通过文件状态可以监控和管理任务

### 技术优势
- **解耦合**: GUI和上传逻辑完全分离
- **可测试**: 核心逻辑更容易进行单元测试
- **可维护**: 代码职责清晰，便于维护
- **可伸缩**: 支持多个GUI实例管理同一组上传任务

## 风险和挑战

### 技术风险
- **文件锁竞争**: 多进程访问同一状态文件
- **进程同步**: GUI和CLI之间的状态同步
- **错误恢复**: 异常情况下的状态恢复

### 解决方案
- 使用文件锁机制保证并发安全
- 实现心跳检测确保进程状态
- 添加完整的错误处理和恢复逻辑

---

*本设计文档将作为架构分离实施的指导文档，确保分离过程的有序进行。*