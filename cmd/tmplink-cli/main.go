package main

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/term"
	"tmplink_uploader/internal/i18n"
	"tmplink_uploader/internal/updater"
)

// CLI配置文件
type CLIConfig struct {
	Token string `json:"token"`
	Model int    `json:"model"`
	MrID  string `json:"mr_id"`
	Language string `json:"language"` // 界面语言设置，如"zh-CN"或"en-US"
}

// 上传配置
type Config struct {
	Token        string
	Server       string
	UploadServer string // 分片上传服务器
	ChunkSize    int
	Model        int
	MrID         string
	SkipUpload   int
	Debug        bool   // 调试模式
	Language     string // 界面语言设置
}

// getCLIConfigPath 获取CLI配置文件路径
func getCLIConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".tmplink_cli_config.json"
	}
	return filepath.Join(homeDir, ".tmplink_cli_config.json")
}

// loadCLIConfig 加载保存的CLI配置
func loadCLIConfig() CLIConfig {
	configPath := getCLIConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return CLIConfig{Model: 0, MrID: "0", Language: ""} // 返回默认值
	}

	var config CLIConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return CLIConfig{Model: 0, MrID: "0", Language: ""} // 返回默认值
	}

	// 确保MrID有默认值
	if config.MrID == "" {
		config.MrID = "0"
	}

	return config
}

// saveCLIConfig 保存CLI配置到文件
func saveCLIConfig(config CLIConfig) error {
	configPath := getCLIConfigPath()

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600) // 设置较严格的权限
}

// 兼容性函数：用于加载token
func loadSavedToken() string {
	config := loadCLIConfig()
	return config.Token
}

// 兼容性函数：用于保存token
func saveToken(token string) error {
	config := loadCLIConfig()
	config.Token = token
	return saveCLIConfig(config)
}

// debugPrint 调试输出函数
func debugPrint(config *Config, format string, args ...interface{}) {
	if config.Debug {
		logMsg := fmt.Sprintf("[DEBUG] "+format+"\n", args...)
		fmt.Print(logMsg)

		// 同时写入日志文件
		if logFile, err := os.OpenFile("api_requests.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			logFile.WriteString(fmt.Sprintf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), logMsg))
			logFile.Close()
		}
	}
}

// 任务状态
type TaskStatus struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	FilePath    string    `json:"file_path"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	Progress    float64   `json:"progress"`
	UploadSpeed float64   `json:"upload_speed,omitempty"` // KB/s
	ServerName  string    `json:"server_name,omitempty"`  // 上传服务器名称
	ProcessID   int       `json:"process_id,omitempty"`   // CLI进程号
	DownloadURL string    `json:"download_url,omitempty"`
	ErrorMsg    string    `json:"error_msg,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 上传结果
type UploadResult struct {
	DownloadURL string
	FileID      string
}

// 速度计算器
type SpeedCalculator struct {
	startTime    time.Time
	lastTime     time.Time
	lastBytes    int64
	totalBytes   int64
	currentSpeed float64 // KB/s
}

// NewSpeedCalculator 创建新的速度计算器
func NewSpeedCalculator(totalBytes int64) *SpeedCalculator {
	now := time.Now()
	return &SpeedCalculator{
		startTime:    now,
		lastTime:     now,
		lastBytes:    0,
		totalBytes:   totalBytes,
		currentSpeed: 0,
	}
}

// UpdateSpeed 更新上传速度
func (sc *SpeedCalculator) UpdateSpeed(uploadedBytes int64) float64 {
	now := time.Now()
	timeDiff := now.Sub(sc.lastTime).Seconds()

	// 降低时间间隔要求，对小文件更友好（0.5秒而不是1秒）
	if timeDiff >= 0.5 {
		bytesDiff := uploadedBytes - sc.lastBytes
		if bytesDiff > 0 && timeDiff > 0 {
			// 计算瞬时速度 (KB/s)
			instantSpeed := float64(bytesDiff) / 1024.0 / timeDiff

			// 使用加权平均平滑速度波动
			if sc.currentSpeed == 0 {
				sc.currentSpeed = instantSpeed
			} else {
				sc.currentSpeed = sc.currentSpeed*0.7 + instantSpeed*0.3
			}
		}

		sc.lastTime = now
		sc.lastBytes = uploadedBytes
	}

	return sc.currentSpeed
}

// GetFinalSpeed 计算最终平均速度（用于上传完成时）
func (sc *SpeedCalculator) GetFinalSpeed() float64 {
	now := time.Now()
	totalTime := now.Sub(sc.startTime).Seconds()

	// 如果总时间太短（小于0.1秒），计算理论最大速度
	if totalTime < 0.1 {
		totalTime = 0.1 // 假设最少0.1秒
	}

	// 计算总体平均速度 (KB/s)
	if totalTime > 0 && sc.totalBytes > 0 {
		avgSpeed := float64(sc.totalBytes) / 1024.0 / totalTime
		// 返回当前速度和平均速度中较大的那个（更准确）
		if sc.currentSpeed > 0 {
			return sc.currentSpeed
		}
		return avgSpeed
	}

	return sc.currentSpeed
}

// isFlagSet 检查flag是否被用户显式设置
func isFlagSet(f *flag.Flag) bool {
	found := false
	flag.Visit(func(flag *flag.Flag) {
		if flag.Name == f.Name {
			found = true
		}
	})
	return found
}

// getProgressBarWidth 根据终端宽度计算进度条宽度
func getProgressBarWidth() int {
	// 尝试获取终端宽度
	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 {
		// 计算进度条的宽度
		// 预留空间给其他元素：百分比(4) + 空格(2) + 括号(2) + 文件大小显示(~25) + 速度显示(~15) + 描述(~8) = ~56字符
		// 额外预留10字符空间以确保不会超出
		reservedSpace := 66
		barWidth := width - reservedSpace

		// 最小宽度20，最大宽度80
		if barWidth < 20 {
			return 20
		}
		if barWidth > 80 {
			return 80
		}
		return barWidth
	}

	// 无法获取终端宽度时使用默认值
	return 40
}

func main() {
	// 定义命令行参数
	var (
		filePath     = flag.String("file", "", "要上传的文件路径 (必需)")
		token        = flag.String("token", "", "TmpLink API token (可选，优先使用已保存的token)")
		setToken     = flag.String("set-token", "", "设置并保存API token")
		setModel     = flag.Int("set-model", -1, "设置并保存默认文件有效期 (0=24小时, 1=3天, 2=7天, 99=无限期)")
		setMrID      = flag.String("set-mr-id", "", "设置并保存默认目录ID")
		setLanguage  = flag.String("set-language", "", "设置并保存界面语言 (cn/en/hk/jp)")
		uploadServer = flag.String("upload-server", "", "强制指定上传服务器地址 (可选，留空自动选择)")
		serverName   = flag.String("server-name", "", "上传服务器名称 (用于显示)")
		chunkSizeMB  = flag.Int("chunk-size", 3, "分块大小(MB, 1-99)")
		statusFile   = flag.String("status-file", "", "任务状态文件路径 (可选，自动生成)")
		taskID       = flag.String("task-id", "", "任务ID (可选，自动生成)")
		model        = flag.Int("model", 0, "文件有效期 (0=24小时, 1=3天, 2=7天, 99=无限期)")
		mrID         = flag.String("mr-id", "0", "目录ID (默认0=根目录)")
		skipUpload   = flag.Int("skip-upload", 1, "跳过上传标志 (1=检查秒传)")
		language     = flag.String("language", "", "界面语言 (cn/en/hk/jp)")
		debugMode    = flag.Bool("debug", false, "调试模式，输出详细运行信息")
		showStatus   = flag.Bool("status", false, "显示当前配置状态和token有效性")
		checkUpdate  = flag.Bool("check-update", false, "检查是否有新版本可用")
		autoUpdate   = flag.Bool("auto-update", false, "自动检查并下载更新")
		showVersion  = flag.Bool("version", false, "显示当前版本号")
	)

	flag.Parse()

	// 处理版本相关的情况
	if *showVersion {
		fmt.Printf("tmplink-cli 版本: %s\n", updater.CURRENT_VERSION)
		return
	}

	if *checkUpdate {
		updateInfo, err := updater.CheckForUpdate("cli")
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
		if err := updater.AutoUpdate("cli"); err != nil {
			fmt.Printf("自动更新失败: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// 处理设置参数的情况
	if *setToken != "" || *setModel >= 0 || *setMrID != "" || *setLanguage != "" {
		config := loadCLIConfig()
		updated := false

		if *setToken != "" {
			// 验证Token有效性
			fmt.Print("正在验证Token有效性...")
			server := "https://tmplink-sec.vxtrans.com/api_v2"
			if uid, err := validateTokenAndGetUID(*setToken, server); err != nil {
				fmt.Printf("\n错误: Token验证失败: %v\n", err)
				fmt.Println("请确保Token正确且有效")
				os.Exit(1)
			} else {
				fmt.Printf(" ✅\n")
				config.Token = *setToken
				fmt.Printf("Token已成功保存并验证 (UID: %s)\n", uid)
				updated = true
			}
		}

		if *setModel >= 0 {
			if *setModel == 0 || *setModel == 1 || *setModel == 2 || *setModel == 99 {
				config.Model = *setModel
				modelDesc := map[int]string{0: "24小时", 1: "3天", 2: "7天", 99: "无限期"}
				fmt.Printf("默认文件有效期已设置为: %s\n", modelDesc[*setModel])
				updated = true
			} else {
				fmt.Fprintf(os.Stderr, "错误: 无效的文件有效期值，支持的值: 0, 1, 2, 99\n")
				os.Exit(1)
			}
		}

		if *setMrID != "" {
			config.MrID = *setMrID
			fmt.Printf("默认目录ID已设置为: %s\n", *setMrID)
			updated = true
		}
		
		if *setLanguage != "" {
			// 验证语言参数
			isValid := false
			for _, lang := range []string{"cn", "en", "hk", "jp", ""} {
				if *setLanguage == lang {
					isValid = true
					break
				}
			}
			
			if !isValid {
				fmt.Printf("错误: 无效的语言参数，支持的值: cn, en, hk, jp\n")
				os.Exit(1)
			}
			
			config.Language = *setLanguage
			updated = true
			
			var langDisplay string
			switch *setLanguage {
			case "cn":
				langDisplay = "中文"
			case "en":
				langDisplay = "English"
			case "hk":
				langDisplay = "繁體中文"
			case "jp":
				langDisplay = "日本語"
			default:
				langDisplay = "自动检测"
			}
			
			fmt.Printf("语言设置已保存为: %s\n", langDisplay)
		}

		if updated {
			if err := saveCLIConfig(config); err != nil {
				fmt.Fprintf(os.Stderr, "错误: 保存配置失败: %v\n", err)
				os.Exit(1)
			}
		}

		return
	}

	// 处理状态查询的情况
	if *showStatus {
		showConfigStatus()
		return
	}

	// 加载保存的配置作为默认值
	savedConfig := loadCLIConfig()
	
	// 初始化语言
	langSetting := *language
	if langSetting == "" {
		langSetting = savedConfig.Language
	}
	i18n.InitLanguage(i18n.Language(langSetting))

	// 参数优先级处理: 命令行参数 > 保存的配置 > 默认值
	finalToken := *token
	if finalToken == "" {
		finalToken = savedConfig.Token
	}

	// 检查model参数是否被显式设置
	modelFlag := flag.Lookup("model")
	finalModel := *model
	if !isFlagSet(modelFlag) {
		// 如果用户没有指定model参数，使用保存的配置
		finalModel = savedConfig.Model
	}

	// 检查mr-id参数是否被显式设置
	mrIDFlag := flag.Lookup("mr-id")
	finalMrID := *mrID
	if !isFlagSet(mrIDFlag) {
		// 如果用户没有指定mr-id参数，使用保存的配置
		finalMrID = savedConfig.MrID
	}

	// 验证必需参数
	if *filePath == "" {
		fmt.Fprintf(os.Stderr, i18n.T("error_missing_file")+"\n")
		flag.Usage()
		os.Exit(1)
	}

	if finalToken == "" {
		fmt.Fprintf(os.Stderr, i18n.T("error_token_not_found")+"\n")
		flag.Usage()
		os.Exit(1)
	}

	// 检测是否为CLI模式（用户未提供task-id）
	cliMode := *taskID == ""

	// 自动生成task-id (如果未提供)
	if cliMode {
		*taskID = fmt.Sprintf("upload_%d", time.Now().Unix())
	}

	// 自动生成status-file (如果未提供)
	if *statusFile == "" {
		*statusFile = fmt.Sprintf("%s_status.json", *taskID)
	}

	// 验证分块大小
	if *chunkSizeMB < 1 || *chunkSizeMB > 99 {
		fmt.Fprintf(os.Stderr, i18n.T("error_chunk_size_range", *chunkSizeMB)+"\n")
		os.Exit(1)
	}

	// 验证文件存在
	if _, err := os.Stat(*filePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, i18n.T("error_file_not_exist", *filePath)+"\n")
		os.Exit(1)
	}

	// 启动时检查更新（后台进行，不阻塞用户操作）
	updater.CheckUpdateOnStartup("cli")

	// 获取文件信息
	fileInfo, err := os.Stat(*filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, i18n.T("error_get_file_info", err)+"\n")
		os.Exit(1)
	}

	// 验证文件大小限制 (50GB)
	const maxFileSize = 50 * 1024 * 1024 * 1024 // 50GB
	if fileInfo.Size() > maxFileSize {
		fmt.Fprintf(os.Stderr, i18n.T("error_file_size_limit", 
			float64(fileInfo.Size())/(1024*1024*1024))+"\n")
		os.Exit(1)
	}

	// 初始化任务状态
	task := &TaskStatus{
		ID:         *taskID,
		Status:     "pending",
		FilePath:   *filePath,
		FileName:   filepath.Base(*filePath),
		FileSize:   fileInfo.Size(),
		Progress:   0.0,
		ServerName: *serverName,
		ProcessID:  os.Getpid(), // 记录当前进程号
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 只有在GUI模式下才保存初始状态到文件
	if !cliMode {
		if err := saveTaskStatus(*statusFile, task); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("error_save_task_status", err)+"\n")
			os.Exit(1)
		}
	}

	// 转换分块大小从MB到字节
	chunkSizeBytes := *chunkSizeMB * 1024 * 1024

	// 创建上传配置
	config := &Config{
		Token:        finalToken,                               // 使用最终确定的token
		Server:       "https://tmplink-sec.vxtrans.com/api_v2", // 固定API服务器地址
		UploadServer: *uploadServer,                            // 用户指定的上传服务器
		ChunkSize:    chunkSizeBytes,
		Model:        finalModel, // 使用最终确定的model
		MrID:         finalMrID,  // 使用最终确定的mrID
		SkipUpload:   *skipUpload,
		Debug:        *debugMode,
		Language:     langSetting,
	}

	debugPrint(config, "启动CLI上传程序")
	debugPrint(config, "文件路径: %s", *filePath)
	debugPrint(config, "分片大小: %d bytes (%dMB)", chunkSizeBytes, *chunkSizeMB)
	debugPrint(config, "API服务器: %s", config.Server)

	// 创建速度计算器
	speedCalc := NewSpeedCalculator(fileInfo.Size())

	// 设置进度回调
	progressCallback := createProgressCallback(cliMode, fileInfo.Size(), speedCalc, task, *statusFile)

	// 验证Token有效性
	debugPrint(config, "验证Token有效性...")
	if _, err := validateTokenAndGetUID(finalToken, config.Server); err != nil {
		task.Status = "failed"
		task.ErrorMsg = fmt.Sprintf("Token验证失败: %v", err)
		task.UpdatedAt = time.Now()

		// CLI模式：显示失败信息
		if cliMode {
			fmt.Printf(i18n.T("token_validation_failed") + "\n")
			fmt.Printf(i18n.T("token_error_message", err) + "\n")
			fmt.Println(i18n.T("token_help_message"))
		} else {
			// GUI模式：保存状态到文件
			if saveErr := saveTaskStatus(*statusFile, task); saveErr != nil {
				fmt.Fprintf(os.Stderr, i18n.T("save_complete_failed", saveErr)+"\n")
			}
			fmt.Fprintf(os.Stderr, i18n.T("token_validation_error", err)+"\n")
		}
		os.Exit(1)
	}
	debugPrint(config, "Token验证成功")

	// 开始上传
	task.Status = "uploading"
	task.UpdatedAt = time.Now()
	// 只有在GUI模式下才保存状态到文件
	if !cliMode {
		saveTaskStatus(*statusFile, task)
	}

	ctx := context.Background()

	result, err := uploadFile(ctx, config, *filePath, progressCallback)
	if err != nil {
		// 上传失败
		task.Status = "failed"
		task.ErrorMsg = err.Error()
		task.UpdatedAt = time.Now()

		// CLI模式：显示失败信息
		if cliMode {
			clearProgressBar() // 清除进度条残留
			fmt.Printf(i18n.T("upload_failed") + "\n")
			fmt.Printf(i18n.T("upload_file_name", task.FileName) + "\n")
			fmt.Printf(i18n.T("upload_error_message", err) + "\n")
		} else {
			// GUI模式：保存状态到文件
			if saveErr := saveTaskStatus(*statusFile, task); saveErr != nil {
				fmt.Fprintf(os.Stderr, i18n.T("save_complete_failed", saveErr)+"\n")
			}
			// GUI模式下仍然输出到stderr，供调试使用
			fmt.Fprintf(os.Stderr, i18n.T("file_upload_failed", err)+"\n")
		}

		os.Exit(1)
	}

	// 上传成功
	task.Status = "completed"
	task.Progress = 100.0
	task.DownloadURL = result.DownloadURL
	task.UpdatedAt = time.Now()

	// CLI模式：显示完成信息
	if cliMode {
		clearProgressBar() // 清除进度条残留
		fmt.Printf(i18n.T("upload_complete") + "\n")
		fmt.Printf(i18n.T("upload_file_name", task.FileName) + "\n")

		// 格式化文件大小
		var sizeStr string
		if task.FileSize < 1024 {
			sizeStr = fmt.Sprintf("%dB", task.FileSize)
		} else if task.FileSize < 1024*1024 {
			sizeStr = fmt.Sprintf("%.1fKB", float64(task.FileSize)/1024)
		} else if task.FileSize < 1024*1024*1024 {
			sizeStr = fmt.Sprintf("%.1fMB", float64(task.FileSize)/(1024*1024))
		} else {
			sizeStr = fmt.Sprintf("%.1fGB", float64(task.FileSize)/(1024*1024*1024))
		}
		fmt.Printf(i18n.T("upload_file_size", sizeStr) + "\n")

		// 显示平均速度
		finalSpeed := speedCalc.GetFinalSpeed()
		if finalSpeed >= 1024 {
			fmt.Printf(i18n.T("upload_average_speed", finalSpeed/1024) + "\n")
		} else {
			fmt.Printf(i18n.T("upload_average_speed", finalSpeed) + "\n")
		}

		// 显示总耗时
		totalTime := time.Since(task.CreatedAt).Round(time.Second)
		fmt.Printf(i18n.T("upload_total_time", totalTime) + "\n")

		// 显示下载链接
		fmt.Printf(i18n.T("upload_download_link", result.DownloadURL) + "\n")
	} else {
		// GUI模式：保存状态到文件
		if err := saveTaskStatus(*statusFile, task); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("save_complete_failed", err)+"\n")
		}
	}
}