package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TaskStatus 任务状态结构（与CLI中的定义保持一致）
type TaskStatus struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	FilePath    string    `json:"file_path"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	Progress    float64   `json:"progress"`
	DownloadURL string    `json:"download_url,omitempty"`
	ErrorMsg    string    `json:"error_msg,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TestCLIBuild 测试CLI程序能否成功构建
func TestCLIBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过构建测试")
	}

	// 构建CLI程序
	cmd := exec.Command("go", "build", "-o", "test-tmplink-cli", "./cmd/tmplink-cli")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("构建CLI失败: %v\n输出: %s", err, output)
	}

	// 清理
	defer os.Remove("test-tmplink-cli")

	// 测试帮助信息
	cmd = exec.Command("./test-tmplink-cli", "-h")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("运行CLI帮助失败: %v\n输出: %s", err, output)
	}

	// 检查帮助信息包含关键参数
	helpText := string(output)
	requiredFlags := []string{"-file", "-token", "-task-id", "-status-file"}
	for _, flag := range requiredFlags {
		if !strings.Contains(helpText, flag) {
			t.Errorf("帮助信息缺少必需参数: %s", flag)
		}
	}
}

// TestGUIBuild 测试GUI程序能否成功构建
func TestGUIBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过构建测试")
	}

	// 构建GUI程序
	cmd := exec.Command("go", "build", "-o", "test-tmplink-gui", "./cmd/tmplink")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("构建GUI失败: %v\n输出: %s", err, output)
	}

	// 清理
	defer os.Remove("test-tmplink-gui")

	t.Log("GUI程序构建成功")
}

// TestCLIParameterValidation 测试CLI参数验证
func TestCLIParameterValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过CLI测试")
	}

	// 构建CLI程序
	cmd := exec.Command("go", "build", "-o", "test-tmplink-cli", "./cmd/tmplink-cli")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("构建CLI失败: %v", err)
	}
	defer os.Remove("test-tmplink-cli")

	// 测试缺少必需参数
	testCases := []struct {
		name string
		args []string
	}{
		{"无参数", []string{}},
		{"只有文件路径", []string{"-file", "test.txt"}},
		{"只有token", []string{"-token", "test-token"}},
		{"缺少状态文件", []string{"-file", "test.txt", "-token", "test-token", "-task-id", "test"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("./test-tmplink-cli")
			cmd.Args = append(cmd.Args, tc.args...)
			
			err := cmd.Run()
			if err == nil {
				t.Error("期望命令失败，但成功了")
			}
		})
	}
}

// TestStatusFileOperations 测试状态文件操作
func TestStatusFileOperations(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "status_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建测试文件
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "这是测试文件内容"
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 状态文件路径
	statusFile := filepath.Join(tmpDir, "task_status.json")

	// 构建CLI程序
	cmd := exec.Command("go", "build", "-o", "test-tmplink-cli", "./cmd/tmplink-cli")
	err = cmd.Run()
	if err != nil {
		t.Fatalf("构建CLI失败: %v", err)
	}
	defer os.Remove("test-tmplink-cli")

	// 使用无效token测试（应该会创建状态文件）
	cmd = exec.Command("./test-tmplink-cli",
		"-file", testFile,
		"-token", "invalid-token",
		"-task-id", "test-task",
		"-status-file", statusFile,
		"-timeout", "5", // 5秒超时
	)

	// 启动命令但不等待完成
	err = cmd.Start()
	if err != nil {
		t.Fatalf("启动CLI失败: %v", err)
	}

	// 等待状态文件创建
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(statusFile); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	// 检查状态文件是否存在
	if _, err := os.Stat(statusFile); os.IsNotExist(err) {
		t.Error("状态文件未创建")
		return
	}

	// 读取状态文件
	data, err := os.ReadFile(statusFile)
	if err != nil {
		t.Fatalf("读取状态文件失败: %v", err)
	}

	// 解析状态文件
	var status TaskStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		t.Fatalf("解析状态文件失败: %v", err)
	}

	// 验证状态文件内容
	if status.ID != "test-task" {
		t.Errorf("任务ID不匹配: 期望 'test-task', 得到 '%s'", status.ID)
	}

	if status.FilePath != testFile {
		t.Errorf("文件路径不匹配: 期望 '%s', 得到 '%s'", testFile, status.FilePath)
	}

	if status.FileSize != int64(len(testContent)) {
		t.Errorf("文件大小不匹配: 期望 %d, 得到 %d", len(testContent), status.FileSize)
	}

	// 停止命令
	if cmd.Process != nil {
		cmd.Process.Kill()
	}
	cmd.Wait()
}

// TestDualProcessArchitecture 测试双进程架构
func TestDualProcessArchitecture(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过架构测试")
	}

	// 构建两个程序
	cmd := exec.Command("go", "build", "-o", "test-tmplink-cli", "./cmd/tmplink-cli")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("构建CLI失败: %v", err)
	}
	defer os.Remove("test-tmplink-cli")

	cmd = exec.Command("go", "build", "-o", "test-tmplink-gui", "./cmd/tmplink")
	err = cmd.Run()
	if err != nil {
		t.Fatalf("构建GUI失败: %v", err)
	}
	defer os.Remove("test-tmplink-gui")

	// 验证两个程序是独立的
	cliInfo, err := os.Stat("test-tmplink-cli")
	if err != nil {
		t.Fatalf("CLI程序不存在: %v", err)
	}

	guiInfo, err := os.Stat("test-tmplink-gui")
	if err != nil {
		t.Fatalf("GUI程序不存在: %v", err)
	}

	// 检查文件大小不同（确保是不同的程序）
	if cliInfo.Size() == guiInfo.Size() {
		t.Log("警告: CLI和GUI程序大小相同，可能存在问题")
	}

	t.Logf("CLI程序大小: %d bytes", cliInfo.Size())
	t.Logf("GUI程序大小: %d bytes", guiInfo.Size())
}

// TestArchitectureCompliance 测试架构合规性
func TestArchitectureCompliance(t *testing.T) {
	t.Run("CLI程序独立性", func(t *testing.T) {
		// CLI程序应该是独立的，不依赖配置文件
		// 所有参数通过命令行传递
		
		// 构建CLI
		cmd := exec.Command("go", "build", "-o", "test-tmplink-cli", "./cmd/tmplink-cli")
		err := cmd.Run()
		if err != nil {
			t.Fatalf("构建CLI失败: %v", err)
		}
		defer os.Remove("test-tmplink-cli")

		// 检查CLI程序的帮助信息
		cmd = exec.Command("./test-tmplink-cli", "-h")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("获取CLI帮助失败: %v", err)
		}

		helpText := string(output)
		
		// 验证所有必需参数都通过命令行提供
		requiredParams := []string{
			"-file",        // 文件路径
			"-token",       // API token
			"-task-id",     // 任务ID
			"-status-file", // 状态文件
		}

		for _, param := range requiredParams {
			if !strings.Contains(helpText, param) {
				t.Errorf("CLI缺少必需参数: %s", param)
			}
		}

		// 验证CLI不应该依赖配置文件
		if strings.Contains(helpText, "config") || strings.Contains(helpText, "配置") {
			t.Log("警告: CLI帮助信息提到了配置，这可能违反了架构要求")
		}
	})

	t.Run("一次性进程设计", func(t *testing.T) {
		// 验证CLI是一次性运行的进程
		// 这个测试通过检查CLI进程的行为来验证
		
		// 构建CLI
		cmd := exec.Command("go", "build", "-o", "test-tmplink-cli", "./cmd/tmplink-cli")
		err := cmd.Run()
		if err != nil {
			t.Fatalf("构建CLI失败: %v", err)
		}
		defer os.Remove("test-tmplink-cli")

		// 用无效参数运行CLI，验证它会退出而不是持续运行
		cmd = exec.Command("./test-tmplink-cli")
		
		start := time.Now()
		err = cmd.Run()
		duration := time.Since(start)

		// CLI应该快速退出（因为参数无效）
		if duration > 5*time.Second {
			t.Error("CLI进程运行时间过长，可能不是一次性进程")
		}

		// CLI应该因为参数错误而退出
		if err == nil {
			t.Error("CLI应该因为缺少参数而失败")
		}

		t.Logf("CLI进程运行时间: %v", duration)
	})
}