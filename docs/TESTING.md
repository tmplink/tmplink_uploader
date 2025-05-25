# TmpLink CLI 测试指南

## 当前状态

tmplink-cli 已经完成上传功能实现，关键bug已修复。v1.0.1版本解决了mr_id参数导致的状态7(data=8)错误，现在能正确返回状态8合并完成。

## 测试命令

### 基本测试命令
```bash
# 创建测试文件
echo "test content $(date)" > test_file.txt

# 运行CLI测试
./tmplink-cli \
  -file "test_file.txt" \
  -token "你的token" \
  -task-id "test_$(date +%s)" \
  -status-file "/tmp/test_status.json"
```

### 使用你的token的测试命令
```bash
./tmplink-cli \
  -file "test_upload_new.txt" \
  -token "caz9qdckqizbeffmoqin" \
  -task-id "test_$(date +%s)" \
  -status-file "~/.tmplink/test/test_$(date +%s).json"
```

## 当前问题

1. ✅ Token验证 - 已修复（UID类型问题）
2. ✅ 服务器地址 - 已修复（API vs 上传服务器）
3. ✅ 分片上传 - 已修复（状态码和服务器选择）
4. ❌ 最终状态 - 仍有问题（状态码7，错误码8）

## 测试结果分析

### 成功的部分：
- ✅ 用户验证成功
- ✅ 获取上传服务器成功 
- ✅ 分片上传成功（进度100%）

### 问题部分：
- ❌ 最终状态检查返回错误码7/8

### 调试信息示例：
```
upload_request_select2 响应: {"data":{"utoken":"UM3WamJYGA","servers":[...],"src":"130.62.77.12"},"status":1,"debug":[]}
使用上传服务器: https://tmp-nl.vx-cdn.com
prepare_v4 响应: {"data":false,"status":0,"debug":["prepare_v4"]}
分片prepare响应: {"status":3,"data":{"next":0,"total":1,"wait":1,"wait_list":[],"uploading":0,"success":0}}
最终状态检查响应: {"status":7,"data":8,"debug":{...}}
```

## 下一步调试

需要研究：
1. 错误码8的具体含义
2. 是否需要等待服务器处理完成
3. 是否需要不同的API调用来获取最终结果

## 快速使用测试脚本

运行 `./test_direct.sh` 进行快速测试。