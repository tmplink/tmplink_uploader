package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	VERSION_URL = "https://raw.githubusercontent.com/tmplink/tmplink_uploader/main/version.json"
	DOWNLOAD_BASE_URL = "https://github.com/tmplink/tmplink_uploader/releases/download"
	CURRENT_VERSION = "1.0.0"
)

type VersionInfo struct {
	CLIVersion string `json:"cli_version"`
	GUIVersion string `json:"gui_version"`
}

type UpdateInfo struct {
	HasUpdate    bool
	LatestVersion string
	DownloadURL   string
	CurrentVersion string
}

// GetPlatformString returns the platform string based on runtime.GOOS and runtime.GOARCH
func GetPlatformString() string {
	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			return "linux-64bit"
		case "386":
			return "linux-32bit"
		case "arm64":
			return "linux-arm64"
		default:
			return "linux-64bit"
		}
	case "windows":
		switch runtime.GOARCH {
		case "amd64":
			return "windows-64bit"
		case "386":
			return "windows-32bit"
		default:
			return "windows-64bit"
		}
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			return "macos-intel"
		case "arm64":
			return "macos-arm64"
		default:
			return "macos-arm64"
		}
	default:
		return "linux-64bit"
	}
}

// GetBinaryName returns the binary name for the given program type and platform
func GetBinaryName(programType string) string {
	platform := GetPlatformString()
	
	if strings.Contains(platform, "windows") {
		if programType == "cli" {
			return "tmplink-cli.exe"
		}
		return "tmplink.exe"
	}
	
	if programType == "cli" {
		return "tmplink-cli"
	}
	return "tmplink"
}

// CheckForUpdate checks if there's a newer version available
func CheckForUpdate(programType string) (*UpdateInfo, error) {
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
		CurrentVersion: CURRENT_VERSION,
		LatestVersion:  latestVersion,
		HasUpdate:      latestVersion != CURRENT_VERSION,
	}
	
	if updateInfo.HasUpdate {
		platform := GetPlatformString()
		binaryName := GetBinaryName(programType)
		updateInfo.DownloadURL = fmt.Sprintf("%s/v%s/%s/%s", 
			DOWNLOAD_BASE_URL, latestVersion, platform, binaryName)
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
func CheckForUpdateSilently(programType string) (*UpdateInfo, error) {
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
		CurrentVersion: CURRENT_VERSION,
		LatestVersion:  latestVersion,
		HasUpdate:      latestVersion != CURRENT_VERSION,
	}
	
	if updateInfo.HasUpdate {
		platform := GetPlatformString()
		binaryName := GetBinaryName(programType)
		updateInfo.DownloadURL = fmt.Sprintf("%s/v%s/%s/%s", 
			DOWNLOAD_BASE_URL, latestVersion, platform, binaryName)
	}
	
	return updateInfo, nil
}

// AutoUpdate performs automatic update check and download
func AutoUpdate(programType string) error {
	fmt.Println("正在检查更新...")
	
	updateInfo, err := CheckForUpdate(programType)
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

// CheckUpdateOnStartup performs a background update check on program startup
func CheckUpdateOnStartup(programType string) {
	go func() {
		updateInfo, err := CheckForUpdateSilently(programType)
		if err != nil {
			// Silent fail, don't interrupt user experience
			return
		}
		
		if updateInfo.HasUpdate {
			fmt.Printf("\n💡 提示: 发现新版本 %s，当前版本 %s\n", 
				updateInfo.LatestVersion, updateInfo.CurrentVersion)
			fmt.Printf("   使用 --auto-update 参数自动更新\n\n")
		}
	}()
}