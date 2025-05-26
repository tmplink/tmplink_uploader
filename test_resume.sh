#!/bin/bash

# 续传功能测试脚本

echo "🧪 测试续传功能"
echo "================="

# 创建测试文件
TEST_FILE="test/resume_test.bin"
TEST_SIZE_MB=10

echo "📁 创建测试文件: ${TEST_FILE} (${TEST_SIZE_MB}MB)"
mkdir -p test
dd if=/dev/zero of="${TEST_FILE}" bs=1M count=${TEST_SIZE_MB} 2>/dev/null

if [ ! -f "${TEST_FILE}" ]; then
    echo "❌ 测试文件创建失败"
    exit 1
fi

echo "✅ 测试文件创建成功: $(ls -lh ${TEST_FILE})"

# 检查是否有保存的token
if [ ! -f ~/.tmplink_cli_config.json ]; then
    echo "❌ 未找到保存的token配置"
    echo "请先运行: ./tmplink-cli -set-token YOUR_TOKEN"
    exit 1
fi

echo ""
echo "🚀 开始上传测试 (将在几秒后中断以测试续传)"
echo "==============================================="

# 启动上传（后台运行）
./tmplink-cli -file "${TEST_FILE}" -chunk-size 1 -debug > test/resume_test.log 2>&1 &
UPLOAD_PID=$!

echo "📤 上传进程启动: PID ${UPLOAD_PID}"

# 等待几秒让上传开始
sleep 5

echo "⏸️  中断上传进程..."
kill -TERM ${UPLOAD_PID} 2>/dev/null || kill -KILL ${UPLOAD_PID} 2>/dev/null

# 等待进程结束
sleep 2

echo ""
echo "🔄 重新启动上传 (测试续传功能)"
echo "================================"

# 重新启动上传
./tmplink-cli -file "${TEST_FILE}" -chunk-size 1 -debug

echo ""
echo "📊 测试结果分析"
echo "==============="

if [ -f "test/resume_test.log" ]; then
    echo "检查日志中的续传关键信息:"
    echo ""
    
    # 检查是否有续传检测信息
    if grep -q "检测到断点续传" test/resume_test.log; then
        echo "✅ 检测到续传功能已触发"
        grep "检测到断点续传" test/resume_test.log
    else
        echo "❌ 未检测到续传功能"
    fi
    
    echo ""
    echo "分片状态信息:"
    grep -E "(总分片数|已完成分片数|待上传分片数)" test/resume_test.log | tail -5
    
    echo ""
    echo "进度信息:"
    grep "进度更新" test/resume_test.log | tail -5
else
    echo "❌ 未找到测试日志文件"
fi

echo ""
echo "🧹 清理测试文件"
rm -f "${TEST_FILE}"
rm -f test/resume_test.log

echo "✅ 测试完成"