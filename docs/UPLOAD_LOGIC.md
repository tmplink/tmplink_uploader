# TmpLink 上传组件逻辑规范

## 分片上传状态码定义

根据JavaScript参考实现，分片上传服务返回状态码含义如下：

- **1**: 上传完成
- **2**: 上传尚未完成，需要等待其他人完成上传（客户端每隔一段时间再次发起查询，如果用户无法完成上传，则重新分配）
- **3**: 进入上传流程，客户端将会获得一份分配的分片编号
- **4**: 分片任务不存在
- **5**: 分片上传完成
- **6**: 这个文件已经被其他人上传了，因此直接跳过（需要清理已上传的文件）
- **7**: 上传失败，原因将会写入到 data
- **8**: 分片合并完成
- **9**: 文件已经上传完成，但是文件合并进程正在进行中，处于锁定状态

## JavaScript实际处理流程

基于 `reference/uploader.js` 的实现：

### 1. 上传初始化
```javascript
// 生成uptoken: SHA1(uid + filename + filesize + slice_size)
let uptoken = CryptoJS.SHA1(this.parent_op.uid + file.name + file.size + this.slice_size).toString();
```

### 2. 分片查询循环 (worker_slice)
```javascript
$.post(server, {
    'token': this.parent_op.api_token, 
    'uptoken': uptoken,
    'action': 'prepare',
    'sha1': sha1, 
    'filename': filename, 
    'filesize': file.size, 
    'slice_size': this.slice_size,
    'utoken': utoken, 
    'mr_id': this.upload_mrid_get(), 
    'model': this.upload_model_get()
}, callback);
```

### 3. 状态码处理逻辑

#### 成功状态 (直接返回下载链接)
- **状态码 1**: `this.upload_final({ status: rsp.status, data: { ukey: rsp.data } }, file, id);`
- **状态码 6**: 重置status=1，调用upload_final
- **状态码 8**: 重置status=1，调用upload_final  
- **状态码 9**: 重置status=1，调用upload_final

#### 等待状态 (继续查询)
- **状态码 2**: `setTimeout(() => { this.worker_slice(...) }, 5000);` (等待5秒后重新查询)
- **状态码 9**: 等待合并进程完成

#### 上传状态 (执行分片上传)
- **状态码 3**: 获取分片编号，执行`worker_slice_uploader`，上传完成后继续查询

#### 失败状态 (错误处理)
- **状态码 7**: `this.upload_final({ status: rsp.data, data: { ukey: rsp.data } }, file, id, thread);`

## CLI实现规范

### 必需参数
- `-file`: 文件路径
- `-token`: API token
- `-task-id`: 任务ID
- `-status-file`: 状态文件路径

### 可选参数
- `-server`: API服务器 (默认: https://tmplink-sec.vxtrans.com/api_v2)
- `-chunk-size`: 分块大小 (默认: 3MB)
- `-timeout`: 超时时间 (默认: 300秒)
- `-model`: 上传模型 (默认: 1)
- `-mr-id`: 资源ID (默认: 空)

### 内部处理流程

1. **获取上传服务器**
   ```
   POST /api_v2/file
   action=upload_request_select2&sha1=...&filename=...&filesize=...&model=...&token=...
   ```

2. **检查秒传**
   ```
   POST /api_v2/file  
   action=prepare_v4&sha1=...&filename=...&filesize=...&model=...&skip_upload=1&token=...
   ```

3. **获取用户UID**
   ```
   POST /api_v2/user
   action=get_detail&token=...
   ```

4. **分片上传循环**
   ```
   POST {upload_server}/app/upload_slice
   token=...&uptoken=...&action=prepare&sha1=...&filename=...&filesize=...&slice_size=...&utoken=...&mr_id=...&model=...
   ```

### 状态码处理规范

```go
switch status {
case 1, 6, 8, 9:
    // 上传完成，返回下载链接
    return fmt.Sprintf("https://tmp.link/%s", ukey)
    
case 2:
    // 等待其他人完成，5秒后重试
    time.Sleep(5 * time.Second)
    continue
    
case 3:
    // 执行分片上传
    sliceIndex := data["next"]
    uploadSlice(sliceIndex)
    continue
    
case 7:
    // 上传失败，返回错误
    return fmt.Errorf("上传失败，错误码: %v", data)
    
default:
    // 未知状态码
    return fmt.Errorf("未知状态码: %d", status)
}
```

### 关键点

1. **uptoken生成**: 必须使用正确的UID，格式为 `SHA1(uid + filename + filesize + slice_size)`
2. **服务器选择**: 必须使用从upload_request_select2获取的上传服务器
3. **状态码8**: 是成功状态，不是错误
4. **重试机制**: 状态码2需要等待重试，状态码7才是真正的失败
5. **参数完整性**: 所有API调用必须包含完整的必需参数

## 错误代码定义 (状态码7的data字段)

当状态码为7时，data字段包含具体错误代码：
- **2**: 无效请求
- **3**: 不能上传空文件  
- **4**: 文件大小超出限制
- **5**: 超出单日上传量限制
- **6**: 没有权限上传到文件夹
- **7**: 超出私有存储空间限制
- **8**: 暂时无法为文件分配存储空间
- **9**: 无法获取节点信息
- **10**: 文件名包含不允许字符

## 实施检查清单

- [ ] uptoken计算使用正确的UID
- [ ] 使用正确的上传服务器地址
- [ ] 状态码8作为成功处理
- [ ] 实现状态码2的重试逻辑
- [ ] 完整的API参数传递
- [ ] 正确的错误码解释