package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/bubbletea"
	"tmplink_uploader/internal/gui/tui"
	"tmplink_uploader/internal/updater"
)

func main() {
	// 定义命令行参数
	var (
		checkUpdate  = flag.Bool("check-update", false, "检查是否有新版本可用")
		autoUpdate   = flag.Bool("auto-update", false, "自动检查并下载更新")
		showVersion  = flag.Bool("version", false, "显示当前版本号")
	)

	flag.Parse()

	// 处理版本相关的情况
	if *showVersion {
		fmt.Printf("tmplink GUI 版本: %s\n", updater.CURRENT_VERSION)
		return
	}

	if *checkUpdate {
		updateInfo, err := updater.CheckForUpdate("gui")
		if err != nil {
			fmt.Printf("检查更新失败: %v\n", err)
			os.Exit(1)
		}
		
		if updateInfo.HasUpdate {
			fmt.Printf("发现新版本: %s (当前版本: %s)\n", 
				updateInfo.LatestVersion, updateInfo.CurrentVersion)
			fmt.Printf("下载地址: %s\n", updateInfo.DownloadURL)
			fmt.Println("使用 --auto-update 参数自动下载更新")
		} else {
			fmt.Printf("当前版本 %s 已是最新版本\n", updateInfo.CurrentVersion)
		}
		return
	}

	if *autoUpdate {
		if err := updater.AutoUpdate("gui"); err != nil {
			fmt.Printf("自动更新失败: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// 获取CLI程序路径
	cliPath := getCLIPath()
	
	// 验证CLI程序存在
	if err := validateCLIPath(cliPath); err != nil {
		log.Fatalf("CLI程序验证失败: %v\n请确保tmplink-cli程序位于: %s", err, cliPath)
	}

	// 启动时检查更新（后台进行，不阻塞用户操作）
	updater.CheckUpdateOnStartup("gui")

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