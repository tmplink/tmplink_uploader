package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

const (
	VERSION_URL           = "https://raw.githubusercontent.com/tmplink/tmplink_uploader/main/version.json"
	DOWNLOAD_BASE_URL     = "https://github.com/tmplink/tmplink_uploader/releases/download"
	GITHUB_REPO          = "tmplink/tmplink_uploader"
	UPDATE_CHECK_INTERVAL = 1 * time.Hour
)

type VersionInfo struct {
	CLIVersion string `json:"cli_version"`
	GUIVersion string `json:"gui_version"`
}

type UpdateInfo struct {
	HasUpdate      bool
	LatestVersion  string
	DownloadURL    string
	CurrentVersion string
}

// SharedConfig represents the shared configuration structure for both GUI and CLI
type SharedConfig struct {
	Token              string    `json:"token"`
	UploadServer       string    `json:"upload_server"`
	SelectedServerName string    `json:"selected_server_name"`
	ChunkSize          int       `json:"chunk_size"`
	MaxConcurrent      int       `json:"max_concurrent"`
	QuickUpload        bool      `json:"quick_upload"`
	SkipUpload         bool      `json:"skip_upload"`
	LastUpdateCheck    time.Time `json:"last_update_check"`
	// CLI专用字段
	Model int    `json:"model"`
	MrID  string `json:"mr_id"`
}

// GetPlatformSuffix returns the platform suffix based on runtime.GOOS and runtime.GOARCH
func GetPlatformSuffix() string {
	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			return "linux-amd64"
		case "386":
			return "linux-386"
		case "arm64":
			return "linux-arm64"
		default:
			return "linux-amd64"
		}
	case "windows":
		switch runtime.GOARCH {
		case "amd64":
			return "windows-amd64"
		case "386":
			return "windows-386"
		default:
			return "windows-amd64"
		}
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			return "darwin-amd64"
		case "arm64":
			return "darwin-arm64"
		default:
			return "darwin-arm64"
		}
	default:
		return "linux-amd64"
	}
}

// GetBinaryName returns the binary name for the given program type with platform suffix
func GetBinaryName(programType string) string {
	platformSuffix := GetPlatformSuffix()

	if runtime.GOOS == "windows" {
		if programType == "cli" {
			return fmt.Sprintf("tmplink-cli-%s.exe", platformSuffix)
		}
		return fmt.Sprintf("tmplink-%s.exe", platformSuffix)
	}

	if programType == "cli" {
		return fmt.Sprintf("tmplink-cli-%s", platformSuffix)
	}
	return fmt.Sprintf("tmplink-%s", platformSuffix)
}

// CheckForUpdate checks if there's a newer version available
func CheckForUpdate(programType string, currentVersion string) (*UpdateInfo, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Fetch version info from GitHub
	resp, err := client.Get(VERSION_URL)
	if err != nil {
		return nil, fmt.Errorf("无法获取版本信息: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取版本信息失败，状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取版本信息失败: %v", err)
	}

	var versionInfo VersionInfo
	if err := json.Unmarshal(body, &versionInfo); err != nil {
		return nil, fmt.Errorf("解析版本信息失败: %v", err)
	}

	var latestVersion string
	if programType == "cli" {
		latestVersion = versionInfo.CLIVersion
	} else {
		latestVersion = versionInfo.GUIVersion
	}

	updateInfo := &UpdateInfo{
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		HasUpdate:      latestVersion != currentVersion,
	}

	if updateInfo.HasUpdate {
		binaryName := GetBinaryName(programType)
		updateInfo.DownloadURL = fmt.Sprintf("%s/v%s/%s",
			DOWNLOAD_BASE_URL, latestVersion, binaryName)
	}

	return updateInfo, nil
}

// DownloadUpdate downloads the latest version of the program
func DownloadUpdate(updateInfo *UpdateInfo, targetPath string) error {
	if !updateInfo.HasUpdate {
		return fmt.Errorf("没有可用的更新")
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	fmt.Printf("正在从 %s 下载更新...\n", updateInfo.DownloadURL)

	resp, err := client.Get(updateInfo.DownloadURL)
	if err != nil {
		return fmt.Errorf("下载失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// Create target directory if it doesn't exist
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// Create temporary file
	tempPath := targetPath + ".tmp"
	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer file.Close()

	// Copy downloaded content to file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("写入文件失败: %v", err)
	}

	// Make file executable (for Unix-like systems)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tempPath, 0755); err != nil {
			os.Remove(tempPath)
			return fmt.Errorf("设置文件权限失败: %v", err)
		}
	}

	// Replace old file with new one
	if err := os.Rename(tempPath, targetPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("替换文件失败: %v", err)
	}

	fmt.Printf("更新完成，已下载到: %s\n", targetPath)
	return nil
}

// CheckForUpdateSilently performs a silent update check without output
func CheckForUpdateSilently(programType string, currentVersion string) (*UpdateInfo, error) {
	// Create HTTP client with shorter timeout for startup check
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Fetch version info from GitHub
	resp, err := client.Get(VERSION_URL)
	if err != nil {
		return nil, err // Silent fail, don't show error to user
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var versionInfo VersionInfo
	if err := json.Unmarshal(body, &versionInfo); err != nil {
		return nil, err
	}

	var latestVersion string
	if programType == "cli" {
		latestVersion = versionInfo.CLIVersion
	} else {
		latestVersion = versionInfo.GUIVersion
	}

	updateInfo := &UpdateInfo{
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		HasUpdate:      latestVersion != currentVersion,
	}

	if updateInfo.HasUpdate {
		binaryName := GetBinaryName(programType)
		updateInfo.DownloadURL = fmt.Sprintf("%s/v%s/%s",
			DOWNLOAD_BASE_URL, latestVersion, binaryName)
	}

	return updateInfo, nil
}

// AutoUpdate performs automatic update check and download
func AutoUpdate(programType string, currentVersion string) error {
	fmt.Println("正在检查更新...")

	updateInfo, err := CheckForUpdate(programType, currentVersion)
	if err != nil {
		return fmt.Errorf("检查更新失败: %v", err)
	}

	if !updateInfo.HasUpdate {
		fmt.Printf("当前版本 %s 已是最新版本\n", updateInfo.CurrentVersion)
		return nil
	}

	fmt.Printf("发现新版本 %s (当前版本: %s)\n",
		updateInfo.LatestVersion, updateInfo.CurrentVersion)

	// Get current executable path
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取程序路径失败: %v", err)
	}

	// Download and install update
	if err := DownloadUpdate(updateInfo, exePath); err != nil {
		return fmt.Errorf("更新失败: %v", err)
	}

	fmt.Printf("更新成功! 请重新启动程序以使用新版本 %s\n", updateInfo.LatestVersion)
	return nil
}

// getSharedConfigPath returns the shared config file path
func getSharedConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".tmplink_config.json"
	}
	return filepath.Join(homeDir, ".tmplink_config.json")
}

// shouldCheckUpdate checks if enough time has passed since last update check
func shouldCheckUpdate(programType string) bool {
	configPath := getSharedConfigPath()
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Config file doesn't exist, should check
		return true
	}
	
	var config SharedConfig
	if err := json.Unmarshal(data, &config); err != nil {
		// Invalid format, should check
		return true
	}
	
	// Check if enough time has passed
	if config.LastUpdateCheck.IsZero() {
		return true
	}
	
	return time.Since(config.LastUpdateCheck) > UPDATE_CHECK_INTERVAL
}

// saveUpdateCheckTime saves the current time as last update check time
func saveUpdateCheckTime(programType string) error {
	configPath := getSharedConfigPath()
	now := time.Now()
	
	// Try to load existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Config file doesn't exist, create a new one with default values
		configDir := filepath.Dir(configPath)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return err
		}
		
		config := SharedConfig{
			ChunkSize:       3,
			MaxConcurrent:   5,
			QuickUpload:     true,
			SkipUpload:      false,
			Model:           0,
			MrID:            "0",
			LastUpdateCheck: now,
		}
		newData, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(configPath, newData, 0644)
	}
	
	// Update existing config
	var config SharedConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}
	config.LastUpdateCheck = now
	newData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, newData, 0644)
}

// CheckUpdateOnStartup performs a background update check on program startup and auto-updates if found
func CheckUpdateOnStartup(programType string, currentVersion string, args []string) {
	go func() {
		// Wait a moment to ensure the program has started properly
		time.Sleep(2 * time.Second)
		
		// Check if we should skip the update check
		if !shouldCheckUpdate(programType) {
			// Skip update check as it was done recently
			return
		}
		
		// Save the check time immediately to prevent multiple checks
		saveUpdateCheckTime(programType)
		
		updateInfo, err := CheckForUpdateSilently(programType, currentVersion)
		if err != nil {
			// Silent fail, don't interrupt user experience
			return
		}

		if updateInfo.HasUpdate {
			fmt.Printf("\n🔄 发现新版本 %s，当前版本 %s\n",
				updateInfo.LatestVersion, updateInfo.CurrentVersion)
			fmt.Println("📥 开始自动更新...")
			
			// Get current executable path
			exePath, err := os.Executable()
			if err != nil {
				fmt.Printf("❌ 获取程序路径失败: %v\n", err)
				fmt.Printf("💡 请手动运行 --auto-update 参数更新\n\n")
				return
			}

			// Download and install update
			if err := DownloadUpdate(updateInfo, exePath); err != nil {
				fmt.Printf("❌ 自动更新失败: %v\n", err)
				fmt.Printf("💡 请手动运行 --auto-update 参数更新\n\n")
				return
			}

			fmt.Printf("✅ 更新成功! 正在重启程序...\n")
			
			// Restart the program with the same arguments
			if err := RestartProgram(exePath, args); err != nil {
				fmt.Printf("❌ 重启程序失败: %v\n", err)
				fmt.Printf("💡 请手动重启程序\n")
			}
			
			// Exit current process
			os.Exit(0)
		}
	}()
}

// RestartProgram restarts the current program with the same arguments
func RestartProgram(exePath string, args []string) error {
	// Prepare arguments (skip the first argument which is the program name)
	var newArgs []string
	if len(args) > 1 {
		newArgs = args[1:]
	}

	if runtime.GOOS == "windows" {
		// On Windows, use cmd.exe to start the new process
		cmd := exec.Command("cmd", "/c", "start", "", exePath)
		cmd.Args = append(cmd.Args, newArgs...)
		cmd.Env = os.Environ()
		return cmd.Start()
	} else {
		// On Unix-like systems, use syscall.Exec to replace the current process
		// Add a small delay to ensure the file write is complete
		time.Sleep(100 * time.Millisecond)
		
		// Prepare full arguments including program path
		fullArgs := append([]string{exePath}, newArgs...)
		
		// Use syscall.Exec to replace current process
		return syscall.Exec(exePath, fullArgs, os.Environ())
	}
}
