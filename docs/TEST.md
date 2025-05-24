# 测试指南

本项目包含全面的单元测试、集成测试和性能测试。

## 测试结构

### 测试文件

- `config_test.go` - 配置管理测试
- `client_test.go` - HTTP客户端测试
- `uploader_test.go` - 文件上传核心逻辑测试
- `app_test.go` - 应用程序逻辑测试
- `integration_test.go` - 集成测试和性能测试
- `test_helpers.go` - 测试辅助工具
- `interfaces.go` - 接口定义（用于mock测试）

### 测试数据

- `testdata/` - 测试数据文件夹
  - `sample_config.json` - 示例配置文件
  - `test_files.txt` - 测试文件内容

## 运行测试

### 基本测试命令

```bash
# 运行所有测试
make test

# 运行单元测试（快速）
make test-short

# 运行特定类型的测试
make test-unit           # 单元测试
make test-integration    # 集成测试

# 生成覆盖率报告
make test-coverage

# 运行性能测试
make test-bench

# 检查竞态条件
make test-race
```

### Go原生命令

```bash
# 运行特定测试
go test -run TestConfigSaveAndLoad -v

# 运行特定包的测试
go test -v ./...

# 生成详细输出
go test -v -cover ./...

# 运行基准测试
go test -bench=. -benchmem
```

## 测试覆盖率

当前测试覆盖率：**23.1%**

### 覆盖的主要功能

✅ **配置管理**
- 配置加载和保存
- 默认值设置
- JSON序列化/反序列化
- 配置验证

✅ **HTTP客户端**
- API请求处理
- 错误响应处理
- 登录流程
- 上传服务器获取
- 文件准备API

✅ **文件上传器**
- SHA1计算
- 队列管理
- 任务状态转换
- 并发安全性
- 配置影响（秒传/跳过上传）

✅ **应用程序逻辑**
- 字节格式化
- 配置切换逻辑
- 文件扫描
- 错误处理

### 需要更多覆盖的区域

⚠️ **用户界面交互**
- promptui交互（难以自动化测试）
- 菜单导航逻辑

⚠️ **文件I/O操作**
- 大文件处理
- 网络分片上传
- 进度条显示

⚠️ **错误恢复**
- 网络中断处理
- 上传重试机制

## 测试类型

### 1. 单元测试

测试单个函数和方法：

```go
func TestCalculateSHA1(t *testing.T) {
    // 测试SHA1计算功能
    helper := NewTestHelper(t)
    defer helper.Cleanup()
    
    testFile := helper.CreateTestFile("test.txt", "Hello, World!")
    uploader := NewUploader(mockClient, config)
    
    result, err := uploader.calculateSHA1(testFile)
    require.NoError(t, err)
    assert.Len(t, result, 40) // SHA1 hex length
}
```

### 2. 集成测试

测试组件间交互：

```go
func TestIntegrationUploadFlow(t *testing.T) {
    // 测试完整上传流程
    server := createMockTmpLinkServer(t)
    defer server.Close()
    
    // 创建真实的配置、客户端、上传器
    // 测试从文件添加到上传完成的整个流程
}
```

### 3. 性能测试

测试性能和资源使用：

```go
func TestPerformanceLargeFile(t *testing.T) {
    // 测试大文件处理性能
    largeFile := helper.CreateTestFileWithSize("large.bin", 1024*1024)
    
    start := time.Now()
    sha1Hash, err := uploader.calculateSHA1(largeFile)
    duration := time.Since(start)
    
    assert.Less(t, duration, 5*time.Second)
}
```

### 4. 并发测试

测试并发安全性：

```go
func TestUploaderConcurrency(t *testing.T) {
    // 测试多goroutine并发访问
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            // 并发操作队列
        }(i)
    }
    wg.Wait()
}
```

## Mock对象

使用testify/mock创建模拟对象：

```go
type MockClient struct {
    mock.Mock
}

func (m *MockClient) PrepareUpload(sha1, filename string, filesize int64, model, skipUpload int) (*PrepareResponse, error) {
    args := m.Called(sha1, filename, filesize, model, skipUpload)
    return args.Get(0).(*PrepareResponse), args.Error(1)
}
```

### Mock设置示例

```go
mockClient := &MockClient{}
mockClient.On("PrepareUpload", 
    mock.AnythingOfType("string"),
    "test.txt",
    int64(1024),
    0, 0).Return(&PrepareResponse{Status: 1}, nil)
```

## 测试数据管理

### 临时文件创建

```go
helper := NewTestHelper(t)
defer helper.Cleanup()

testFile := helper.CreateTestFile("test.txt", "content")
largeFile := helper.CreateTestFileWithSize("large.bin", 1024*1024)
```

### 配置管理

```go
config := helper.CreateTestConfig()
config.QuickUpload = false
config.ChunkSize = 512
```

## 持续集成

测试可以集成到CI/CD流水线中：

```yaml
# .github/workflows/test.yml
- name: Run tests
  run: |
    make test-short
    make test-coverage
    
- name: Upload coverage
  uses: codecov/codecov-action@v1
  with:
    file: ./coverage.out
```

## 故障排除

### 常见测试问题

1. **临时文件清理**
   ```go
   defer helper.Cleanup() // 确保清理临时文件
   ```

2. **Mock期望设置**
   ```go
   mockClient.AssertExpectations(t) // 验证所有mock调用
   ```

3. **并发测试稳定性**
   ```go
   time.Sleep(time.Millisecond) // 必要时添加小延迟
   ```

### 性能基准

运行基准测试来监控性能：

```bash
go test -bench=BenchmarkFormatBytes -benchmem
```

期望结果：
- SHA1计算：< 100ms/MB
- 配置加载：< 1ms
- 队列操作：< 1μs

## 测试最佳实践

1. **测试命名** - 使用描述性名称
2. **单一职责** - 每个测试只验证一个功能
3. **独立性** - 测试之间不应相互依赖
4. **可重复性** - 测试结果应该一致
5. **快速反馈** - 单元测试应该快速运行
6. **真实场景** - 集成测试应该模拟真实使用场景