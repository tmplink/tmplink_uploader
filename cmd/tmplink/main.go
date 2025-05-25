package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/bubbletea"
	"tmplink_uploader/internal/gui/tui"
)

func main() {
	// 获取CLI程序路径
	cliPath := getCLIPath()
	
	// 验证CLI程序存在
	if err := validateCLIPath(cliPath); err != nil {
		log.Fatalf("CLI程序验证失败: %v\n请确保tmplink-cli程序位于: %s", err, cliPath)
	}

	// 创建TUI模型
	model := tui.NewModel(cliPath)

	// 启动TUI程序
	p := tea.NewProgram(model, tea.WithAltScreen())
	
	if _, err := p.Run(); err != nil {
		log.Fatalf("程序运行失败: %v", err)
	}
}

// getCLIPath 获取CLI程序路径
func getCLIPath() string {
	// 首先尝试当前目录
	currentDir, err := os.Getwd()
	if err == nil {
		cliPath := filepath.Join(currentDir, "tmplink-cli")
		if _, err := os.Stat(cliPath); err == nil {
			return cliPath
		}
	}

	// 尝试可执行文件同目录
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		cliPath := filepath.Join(execDir, "tmplink-cli")
		if _, err := os.Stat(cliPath); err == nil {
			return cliPath
		}
	}

	// 尝试PATH环境变量
	if path, err := exec.LookPath("tmplink-cli"); err == nil {
		return path
	}

	// 默认假设在同一目录
	return "./tmplink-cli"
}

// validateCLIPath 验证CLI程序路径
func validateCLIPath(cliPath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(cliPath); os.IsNotExist(err) {
		return fmt.Errorf("CLI程序不存在: %s", cliPath)
	}

	// 检查是否可执行
	if err := exec.Command(cliPath, "-h").Start(); err != nil {
		return fmt.Errorf("CLI程序无法执行: %v", err)
	}

	return nil
}

func init() {
	// 设置日志输出到文件，避免干扰TUI
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/tmp"
	}
	
	logDir := filepath.Join(homeDir, ".tmplink")
	os.MkdirAll(logDir, 0755)
	
	logFile, err := os.OpenFile(filepath.Join(logDir, "tmplink.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logFile)
	}
}