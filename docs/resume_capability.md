# 断点续传功能说明

## 功能概述

本项目已实现完整的断点续传功能，当上传过程中因网络中断、进程崩溃等原因中断时，重新启动上传会自动检测并继续从中断点开始上传。

## 实现原理

### 1. 续传检测机制

基于 JavaScript 版本的实现，Go 程序在分片上传时会解析服务器返回的完整分片状态：

```go
// 解析 API 响应中的分片信息
type SliceStatus struct {
    Total int `json:"total"` // 总分片数
    Wait  int `json:"wait"`  // 待上传分片数  
    Next  int `json:"next"`  // 下一个要上传的分片编号
}

// 计算已完成分片数
uploadedSlices := totalSlices - waitingSlices
```

### 2. 进度恢复算法

当检测到存在已上传分片时，自动计算并恢复进度显示：

```go
if !resumeTracker.initialized && uploadedSlices > 0 {
    // 基于已完成分片估算已上传字节数
    estimatedBytes := int64(uploadedSlices) * int64(chunkSize)
    
    // 更新进度显示
    progressCallback(estimatedBytes, fileSize)
    
    // 标记续传已初始化
    resumeTracker.initialized = true
}
```

### 3. 与 JavaScript 版本对比

| 功能特性 | JavaScript 版本 | Go 版本 (修复后) |
|----------|----------------|------------------|
| 分片状态解析 | ✅ 完整解析 total/wait/next | ✅ 完整解析 total/wait/next |
| 续传检测 | ✅ 自动检测已上传分片 | ✅ 自动检测已上传分片 |
| 进度恢复 | ✅ 基于分片数估算字节数 | ✅ 基于分片数估算字节数 |
| 状态初始化 | ✅ 一次性初始化标记 | ✅ 一次性初始化标记 |

## 使用示例

### 正常上传
```bash
# 完整上传一个大文件
./tmplink-cli -file large_file.bin -chunk-size 3
```

### 模拟续传场景
```bash
# 1. 开始上传
./tmplink-cli -file large_file.bin -chunk-size 1 &
UPLOAD_PID=$!

# 2. 几秒后中断
sleep 5
kill $UPLOAD_PID

# 3. 重新启动 - 自动续传
./tmplink-cli -file large_file.bin -chunk-size 1
```

## 调试信息

启用调试模式可以看到详细的续传过程：

```bash
./tmplink-cli -file test.bin -chunk-size 1 -debug
```

### 关键日志输出

**续传检测：**
```
🔄 检测到断点续传: 已完成 5/10 分片 (50.0%)
🔄 估算已上传字节数: 5242880/10485760 (5.0MB/10.0MB)
```

**分片状态：**
```
总分片数: 10
待上传分片数: 5, 已完成分片数: 5  
下一个分片编号: 5
```

**进度更新：**
```
进度更新: 分片#5完成, 总进度: 6291456/10485760 bytes (60.0%)
进度更新: 分片#6完成, 总进度: 7340032/10485760 bytes (70.0%)
```

## 技术细节

### 1. ResumeTracker 结构

```go
type ResumeTracker struct {
    initialized   bool  // 防止重复初始化
    totalSlices   int   // 总分片数
    uploadedBytes int64 // 已上传字节数估算值
}
```

### 2. 核心改进点

**修复前的问题：**
- 只解析 `next` 字段，忽略 `total` 和 `wait`
- 进度计算不考虑已上传分片
- 每次从 0% 开始显示

**修复后的改进：**
- 完整解析服务器分片状态
- 自动检测和初始化续传进度
- 基于分片状态准确计算进度

### 3. 兼容性

- 保持与原有 API 完全兼容
- 不影响正常上传流程
- 续传功能完全透明，无需用户干预

## 测试验证

使用提供的测试脚本验证续传功能：

```bash
./test_resume.sh
```

该脚本会：
1. 创建测试文件
2. 启动上传并中断
3. 重新启动验证续传
4. 分析日志输出
5. 清理测试文件

## 注意事项

1. **分块大小一致性**：续传时必须使用相同的分块大小
2. **文件完整性**：确保文件内容未发生变化
3. **网络稳定性**：续传功能不能解决持续的网络问题
4. **服务器支持**：依赖钛盘服务器的分片状态维护

## 性能优化

- 续传检测只在第一次状态3响应时执行
- 进度计算使用简化算法，减少计算开销  
- 调试信息可选，不影响生产环境性能