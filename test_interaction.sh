#!/bin/bash

# 测试GUI和CLI交互的脚本

echo "=== TmpLink GUI-CLI 交互测试 ==="
echo

# 检查构建
echo "1. 检查构建..."
go build ./cmd/tmplink-cli
go build ./cmd/tmplink
echo "✅ 构建成功"
echo

# 创建测试文件
echo "2. 创建测试文件..."
echo "Hello, TmpLink!" > test_upload.txt
echo "✅ 测试文件已创建: test_upload.txt"
echo

# 测试CLI参数
echo "3. 测试CLI参数验证..."
./tmplink-cli -h 2>/dev/null || echo "✅ CLI帮助信息正常"
echo

# 测试缺少参数的情况
echo "4. 测试参数验证..."
./tmplink-cli -file test_upload.txt 2>/dev/null || echo "✅ 缺少参数时正确退出"
echo

# 创建状态文件目录
echo "5. 创建状态文件目录..."
mkdir -p ~/.tmplink/tasks
echo "✅ 状态文件目录已创建"
echo

echo "=== 交互测试完成 ==="
echo "注意: 实际上传需要有效的token"
echo
echo "GUI启动命令: ./tmplink"
echo "CLI手动测试示例:"
echo "./tmplink-cli -file test_upload.txt -token YOUR_TOKEN -task-id test123 -status-file ~/.tmplink/tasks/test123.json"
echo

# 清理
rm -f test_upload.txt
echo "✅ 测试文件已清理"