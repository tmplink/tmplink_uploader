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
	"tmplink_uploader/internal/updater"
)

// CLI配置文件
type CLIConfig struct {
	Token string `json:"token"`
	Model int    `json:"model"`
	MrID  string `json:"mr_id"`
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
	Debug        bool // 调试模式
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
		return CLIConfig{Model: 0, MrID: "0"} // 返回默认值
	}

	var config CLIConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return CLIConfig{Model: 0, MrID: "0"} // 返回默认值
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

func main() {
	// 定义命令行参数
	var (
		filePath     = flag.String("file", "", "要上传的文件路径 (必需)")
		token        = flag.String("token", "", "TmpLink API token (可选，优先使用已保存的token)")
		setToken     = flag.String("set-token", "", "设置并保存API token")
		setModel     = flag.Int("set-model", -1, "设置并保存默认文件有效期 (0=24小时, 1=3天, 2=7天, 99=无限期)")
		setMrID      = flag.String("set-mr-id", "", "设置并保存默认目录ID")
		uploadServer = flag.String("upload-server", "", "强制指定上传服务器地址 (可选，留空自动选择)")
		serverName   = flag.String("server-name", "", "上传服务器名称 (用于显示)")
		chunkSizeMB  = flag.Int("chunk-size", 3, "分块大小(MB, 1-99)")
		statusFile   = flag.String("status-file", "", "任务状态文件路径 (可选，自动生成)")
		taskID       = flag.String("task-id", "", "任务ID (可选，自动生成)")
		model        = flag.Int("model", 0, "文件有效期 (0=24小时, 1=3天, 2=7天, 99=无限期)")
		mrID         = flag.String("mr-id", "0", "目录ID (默认0=根目录)")
		skipUpload   = flag.Int("skip-upload", 1, "跳过上传标志 (1=检查秒传)")
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
	if *setToken != "" || *setModel >= 0 || *setMrID != "" {
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
		fmt.Fprintf(os.Stderr, "错误: 缺少必需参数 -file\n")
		flag.Usage()
		os.Exit(1)
	}

	if finalToken == "" {
		fmt.Fprintf(os.Stderr, "错误: 未找到token，请使用 -token 参数或先用 -set-token 保存token\n")
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
		fmt.Fprintf(os.Stderr, "错误: 分块大小必须在1-99MB之间，当前值: %dMB\n", *chunkSizeMB)
		os.Exit(1)
	}

	// 验证文件存在
	if _, err := os.Stat(*filePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "错误: 文件不存在: %s\n", *filePath)
		os.Exit(1)
	}

	// 启动时检查更新（后台进行，不阻塞用户操作）
	updater.CheckUpdateOnStartup("cli")

	// 获取文件信息
	fileInfo, err := os.Stat(*filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 获取文件信息失败: %v\n", err)
		os.Exit(1)
	}

	// 验证文件大小限制 (50GB)
	const maxFileSize = 50 * 1024 * 1024 * 1024 // 50GB
	if fileInfo.Size() > maxFileSize {
		fmt.Fprintf(os.Stderr, "错误: 文件大小超出限制，最大支持50GB，当前文件: %.2fGB\n",
			float64(fileInfo.Size())/(1024*1024*1024))
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
			fmt.Fprintf(os.Stderr, "错误: 保存任务状态失败: %v\n", err)
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
	}

	debugPrint(config, "启动CLI上传程序")
	debugPrint(config, "文件路径: %s", *filePath)
	debugPrint(config, "分片大小: %d bytes (%dMB)", chunkSizeBytes, *chunkSizeMB)
	debugPrint(config, "API服务器: %s", config.Server)

	// 创建速度计算器
	speedCalc := NewSpeedCalculator(fileInfo.Size())

	// 设置进度回调
	progressCallback := createProgressCallback(cliMode, fileInfo.Size(), speedCalc, task, *statusFile)

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
			fmt.Printf("❌ 上传失败!\n")
			fmt.Printf("📁 文件名: %s\n", task.FileName)
			fmt.Printf("❗ 错误信息: %v\n", err)
		} else {
			// GUI模式：保存状态到文件
			if !cliMode {
				if saveErr := saveTaskStatus(*statusFile, task); saveErr != nil {
					fmt.Fprintf(os.Stderr, "错误: 保存失败状态失败: %v\n", saveErr)
				}
			}
		}

		fmt.Fprintf(os.Stderr, "上传失败: %v\n", err)
		os.Exit(1)
	}

	// 上传成功
	task.Status = "completed"
	task.Progress = 100.0
	task.UpdatedAt = time.Now()
	task.DownloadURL = result.DownloadURL
	// 计算最终速度（确保小文件也有速度显示）
	task.UploadSpeed = speedCalc.GetFinalSpeed()

	// CLI模式：显示完成信息
	if cliMode {
		clearProgressBar() // 清除进度条残留
		fmt.Printf("✅ 上传完成!\n")
		fmt.Printf("📁 文件名: %s\n", task.FileName)
		fmt.Printf("📊 文件大小: %s\n", formatBytes(fileInfo.Size()))
		fmt.Printf("⚡ 平均速度: %.2f MB/s\n", task.UploadSpeed/1024) // 转换为MB/s
		duration := time.Since(speedCalc.startTime)
		fmt.Printf("⏱️  总耗时: %v\n", duration.Round(time.Second))
		fmt.Printf("🔗 下载链接: %s\n", result.DownloadURL)
	} else {
		// GUI模式：保存状态到文件
		if !cliMode {
			if err := saveTaskStatus(*statusFile, task); err != nil {
				fmt.Fprintf(os.Stderr, "警告: 保存完成状态失败: %v\n", err)
			}
		}
	}
}

// saveTaskStatus 保存任务状态到文件
func saveTaskStatus(statusFile string, task *TaskStatus) error {
	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return err
	}

	// 确保目录存在
	dir := filepath.Dir(statusFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(statusFile, data, 0644)
}

// uploadFile 上传文件 - 完全按照JavaScript逻辑
func uploadFile(ctx context.Context, config *Config, filePath string, progressCallback func(int64, int64)) (*UploadResult, error) {
	debugPrint(config, "开始上传文件: %s", filePath)

	// 计算文件SHA1
	debugPrint(config, "正在计算文件SHA1...")
	sha1Hash, err := calculateSHA1(filePath)
	if err != nil {
		return nil, fmt.Errorf("计算SHA1失败: %w", err)
	}
	debugPrint(config, "文件SHA1: %s", sha1Hash)

	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}
	debugPrint(config, "文件大小: %d bytes (%.2f MB)", fileInfo.Size(), float64(fileInfo.Size())/(1024*1024))

	fileName := filepath.Base(filePath)

	var uploadInfo *UploadInfo

	// 检查是否为GUI模式（已预设上传服务器）
	if config.UploadServer != "" {
		debugPrint(config, "GUI模式: 使用预设的上传服务器: %s", config.UploadServer)
		// GUI模式下仍需要获取UToken，但直接使用预设的上传服务器
		uploadInfo, err = getUTokenOnly(ctx, config, sha1Hash, fileName, fileInfo.Size())
		if err != nil {
			return nil, fmt.Errorf("获取UToken失败: %w", err)
		}
		// 使用预设的上传服务器
		uploadInfo.Server = config.UploadServer
		debugPrint(config, "获取到UToken: %s", uploadInfo.UToken)
		debugPrint(config, "使用预设上传服务器: %s", uploadInfo.Server)
	} else {
		// CLI独立模式：查找可用的上传服务器
		debugPrint(config, "CLI独立模式: 查找可用上传服务器...")
		uploadInfo, err = getUploadServers(ctx, config, sha1Hash, fileName, fileInfo.Size())
		if err != nil {
			return nil, fmt.Errorf("获取上传服务器失败: %w", err)
		}
		debugPrint(config, "获取到UToken: %s", uploadInfo.UToken)
		debugPrint(config, "找到上传服务器: %s", uploadInfo.Server)
	}

	// 第二步：调用prepare_v4检查是否可以秒传
	debugPrint(config, "步骤2: 检查是否支持秒传...")
	downloadURL, needUpload, err := checkQuickUpload(ctx, config, sha1Hash, fileName, fileInfo.Size())
	if err != nil {
		return nil, fmt.Errorf("检查秒传失败: %w", err)
	}

	if !needUpload {
		debugPrint(config, "秒传成功! 下载链接: %s", downloadURL)
		return &UploadResult{DownloadURL: downloadURL}, nil
	}
	debugPrint(config, "需要分片上传")

	// 第三步：执行分片上传逻辑
	debugPrint(config, "步骤3: 开始分片上传...")
	downloadURL, err = workerSlice(ctx, config, filePath, sha1Hash, fileName, fileInfo.Size(), uploadInfo.UToken, progressCallback)
	if err != nil {
		return nil, fmt.Errorf("分片上传失败: %w", err)
	}

	return &UploadResult{DownloadURL: downloadURL}, nil
}

// calculateSHA1 计算文件SHA1
func calculateSHA1(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha1.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// ResumeTracker 续传进度跟踪器
type ResumeTracker struct {
	initialized   bool  // 是否已初始化续传状态
	totalSlices   int   // 总分片数
	uploadedBytes int64 // 已上传字节数（基于已完成分片估算）
}

// workerSlice 分片上传核心逻辑，基于 JavaScript 实现，支持断点续传
func workerSlice(ctx context.Context, config *Config, filePath, sha1Hash, fileName string, fileSize int64, utoken string, progressCallback func(int64, int64)) (string, error) {
	// 生成uptoken (基于文件特征: SHA1(sha1 + filename + filesize + slice_size))
	upTokenData := fmt.Sprintf("%s%s%d%d", sha1Hash, fileName, fileSize, config.ChunkSize)
	upTokenHash := sha1.Sum([]byte(upTokenData))
	upToken := hex.EncodeToString(upTokenHash[:])

	debugPrint(config, "生成uptoken: %s -> %s", upTokenData, upToken)
	debugPrint(config, "开始分片上传状态机循环...")

	client := &http.Client{}

	// 初始化续传跟踪器
	resumeTracker := &ResumeTracker{
		initialized:   false,
		totalSlices:   0,
		uploadedBytes: 0,
	}

	// 添加循环计数器，防止无限循环
	loopCount := 0
	maxLoops := 1000 // 允许最多1000次状态机循环

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		// 检查循环次数，防止无限循环
		loopCount++
		if loopCount > maxLoops {
			return "", fmt.Errorf("上传超时，状态机循环次数过多（%d次）", loopCount)
		}

		// 查询分片信息 (prepare)
		debugPrint(config, "========== API请求 #%d ==========", loopCount)
		prepareData := fmt.Sprintf("token=%s&uptoken=%s&action=prepare&sha1=%s&filename=%s&filesize=%d&slice_size=%d&utoken=%s&mr_id=%s&model=%d",
			config.Token, upToken, sha1Hash, fileName, fileSize, config.ChunkSize, utoken, config.MrID, config.Model)

		debugPrint(config, "请求URL: %s", config.UploadServer+"/app/upload_slice")
		debugPrint(config, "请求方法: POST")
		debugPrint(config, "Content-Type: application/x-www-form-urlencoded")
		debugPrint(config, "请求参数: %s", prepareData)

		req, err := http.NewRequestWithContext(ctx, "POST", config.UploadServer+"/app/upload_slice", strings.NewReader(prepareData))
		if err != nil {
			debugPrint(config, "创建请求失败: %v", err)
			return "", err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := client.Do(req)
		if err != nil {
			debugPrint(config, "发送请求失败: %v", err)
			return "", fmt.Errorf("网络请求失败: %w", err)
		}
		defer resp.Body.Close()

		debugPrint(config, "HTTP状态码: %d", resp.StatusCode)

		if resp.StatusCode != http.StatusOK {
			debugPrint(config, "HTTP状态码错误: %d", resp.StatusCode)
			return "", fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			debugPrint(config, "读取响应失败: %v", err)
			return "", fmt.Errorf("读取响应失败: %w", err)
		}

		debugPrint(config, "响应内容: %s", string(body))

		var prepareResp struct {
			Status int         `json:"status"`
			Data   interface{} `json:"data"`
			Debug  interface{} `json:"debug,omitempty"`
		}

		if err := json.Unmarshal(body, &prepareResp); err != nil {
			debugPrint(config, "JSON解析失败: %v", err)
			return "", fmt.Errorf("JSON解析失败: %w", err)
		}

		debugPrint(config, "解析结果 - 状态码: %d, 数据: %v, Debug: %v", prepareResp.Status, prepareResp.Data, prepareResp.Debug)

		switch prepareResp.Status {
		case 1:
			// 上传完成
			debugPrint(config, "状态1: 上传完成")
			if ukey, ok := prepareResp.Data.(string); ok {
				return fmt.Sprintf("https://tmp.link/f/%s", ukey), nil
			}
			return "", fmt.Errorf("无法获取ukey")

		case 6:
			// 文件已被其他人上传，直接跳过
			debugPrint(config, "状态6: 文件已存在，直接返回")
			if ukey, ok := prepareResp.Data.(string); ok {
				return fmt.Sprintf("https://tmp.link/f/%s", ukey), nil
			}
			return "", fmt.Errorf("无法获取ukey")

		case 8:
			// 分片合并完成 - 按照JavaScript逻辑直接成功
			debugPrint(config, "状态8: 分片合并完成，上传成功")
			if ukey, ok := prepareResp.Data.(string); ok {
				return fmt.Sprintf("https://tmp.link/f/%s", ukey), nil
			}
			// 如果data是数字，也当作ukey处理
			if ukeyNum, ok := prepareResp.Data.(float64); ok {
				return fmt.Sprintf("https://tmp.link/f/%d", int64(ukeyNum)), nil
			}
			return "", fmt.Errorf("无法获取ukey")

		case 9:
			// 文件合并进程正在进行中，按照JavaScript逻辑直接成功
			debugPrint(config, "状态9: 合并进行中，按JS逻辑直接成功")
			if ukey, ok := prepareResp.Data.(string); ok {
				return fmt.Sprintf("https://tmp.link/f/%s", ukey), nil
			}
			// 如果没有ukey，等待一下再查询
			debugPrint(config, "状态9: 没有ukey，等待2秒...")
			time.Sleep(2 * time.Second)
			continue

		case 2:
			// 没有可上传分片，等待所有分片完成
			debugPrint(config, "状态2: 等待分片完成，等待5秒...")
			time.Sleep(5 * time.Second)
			continue

		case 3:
			// 获得一个需要上传的分片编号，开始处理上传
			debugPrint(config, "状态3: 需要上传分片")
			if dataMap, ok := prepareResp.Data.(map[string]interface{}); ok {
				// 解析完整的分片信息（支持续传检测）
				var totalSlices, waitingSlices, uploadedSlices int
				var nextSlice int = -1

				// 解析总分片数
				if total, ok := dataMap["total"].(float64); ok {
					totalSlices = int(total)
					resumeTracker.totalSlices = totalSlices
					debugPrint(config, "总分片数: %d", totalSlices)
				}

				// 解析待上传分片数
				if wait, ok := dataMap["wait"].(float64); ok {
					waitingSlices = int(wait)
					uploadedSlices = totalSlices - waitingSlices
					debugPrint(config, "待上传分片数: %d, 已完成分片数: %d", waitingSlices, uploadedSlices)
				}

				// 解析下一个要上传的分片编号
				if nextFloat, ok := dataMap["next"].(float64); ok {
					nextSlice = int(nextFloat)
					debugPrint(config, "下一个分片编号: %d", nextSlice)
				}

				// 断点续传初始化 - 只在第一次检测到续传时执行
				if !resumeTracker.initialized && uploadedSlices > 0 && totalSlices > 0 {
					resumeTracker.initialized = true

					// 计算已上传字节数（基于已完成分片估算）
					estimatedBytes := int64(uploadedSlices) * int64(config.ChunkSize)
					if estimatedBytes > fileSize {
						estimatedBytes = fileSize
					}
					resumeTracker.uploadedBytes = estimatedBytes

					// 计算续传进度百分比
					progressPercent := float64(uploadedSlices) / float64(totalSlices) * 100

					debugPrint(config, "🔄 检测到断点续传: 已完成 %d/%d 分片 (%.1f%%)",
						uploadedSlices, totalSlices, progressPercent)
					debugPrint(config, "🔄 估算已上传字节数: %d/%d (%s/%s)",
						estimatedBytes, fileSize,
						formatBytes(estimatedBytes), formatBytes(fileSize))

					// 调用进度回调更新显示
					if progressCallback != nil {
						progressCallback(estimatedBytes, fileSize)
					}
				}

				// 检查是否有下一个分片需要上传
				if nextSlice >= 0 {
					debugPrint(config, "上传分片 #%d", nextSlice)

					// 上传分片
					err := uploadSlice(ctx, client, config, filePath, fileName, upToken, nextSlice, resumeTracker, progressCallback)
					if err != nil {
						return "", fmt.Errorf("分片 %d 上传失败: %w", nextSlice, err)
					}
					debugPrint(config, "分片 #%d 上传完成", nextSlice)

					// 继续下一轮查询
					continue
				}
			}
			return "", fmt.Errorf("无法解析分片信息")

		case 7:
			// 按照JavaScript逻辑：rsp.data是错误代码，直接传递给upload_final
			debugPrint(config, "状态7: 上传失败，错误代码: %v", prepareResp.Data)

			// 检查是否是特殊情况：data为0、8或9（按照JavaScript逻辑直接成功）
			if dataFloat, ok := prepareResp.Data.(float64); ok {
				if dataFloat == 0 {
					debugPrint(config, "状态7但data=0: 上传成功，按JavaScript逻辑直接成功")
					// 根据debug信息构造下载链接
					if debugMap, ok := prepareResp.Debug.(map[string]interface{}); ok {
						if fileinfo, ok := debugMap["fileinfo"].(map[string]interface{}); ok {
							if sha1, ok := fileinfo["sha1"].(string); ok {
								return fmt.Sprintf("https://tmp.link/f/%s", sha1), nil
							}
						}
					}
					// 如果无法从debug获取，返回基于SHA1的链接
					return fmt.Sprintf("https://tmp.link/f/upload_success"), nil
				} else if dataFloat == 8 {
					debugPrint(config, "状态7但data=8: 合并完成，按JavaScript逻辑直接成功")
					// 根据debug信息构造下载链接
					if debugMap, ok := prepareResp.Debug.(map[string]interface{}); ok {
						if fileinfo, ok := debugMap["fileinfo"].(map[string]interface{}); ok {
							if sha1, ok := fileinfo["sha1"].(string); ok {
								return fmt.Sprintf("https://tmp.link/f/%s", sha1), nil
							}
						}
					}
					// 如果无法从debug获取，返回基于SHA1的链接
					return fmt.Sprintf("https://tmp.link/f/upload_success"), nil
				} else if dataFloat == 9 {
					debugPrint(config, "状态7但data=9: 合并进行中，按JavaScript逻辑直接成功")
					// 根据debug信息构造下载链接
					if debugMap, ok := prepareResp.Debug.(map[string]interface{}); ok {
						if fileinfo, ok := debugMap["fileinfo"].(map[string]interface{}); ok {
							if sha1, ok := fileinfo["sha1"].(string); ok {
								return fmt.Sprintf("https://tmp.link/f/%s", sha1), nil
							}
						}
					}
					// 如果无法从debug获取，返回基于SHA1的链接
					return fmt.Sprintf("https://tmp.link/f/upload_success"), nil
				}
			}

			return "", fmt.Errorf("服务器返回上传失败，错误码: %v", prepareResp.Data)

		default:
			debugPrint(config, "未知状态码: %d", prepareResp.Status)
			return "", fmt.Errorf("未知状态码: %d", prepareResp.Status)
		}
	}
}

// uploadSlice 上传单个分片，支持续传进度计算
func uploadSlice(ctx context.Context, client *http.Client, config *Config, filePath, fileName, upToken string, sliceIndex int, resumeTracker *ResumeTracker, progressCallback func(int64, int64)) error {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 计算分片偏移和大小
	chunkSize := int64(config.ChunkSize)
	offset := int64(sliceIndex) * chunkSize

	// 读取分片数据
	_, err = file.Seek(offset, 0)
	if err != nil {
		return fmt.Errorf("文件定位失败: %w", err)
	}

	buffer := make([]byte, chunkSize)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("读取分片数据失败: %w", err)
	}

	chunkData := buffer[:n]

	// 创建multipart表单
	debugPrint(config, "========== 上传分片 #%d ==========", sliceIndex)
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加文件数据
	fileWriter, err := writer.CreateFormFile("filedata", "slice")
	if err != nil {
		debugPrint(config, "创建文件字段失败: %v", err)
		return err
	}
	fileWriter.Write(chunkData)

	// 添加表单字段
	writer.WriteField("uptoken", upToken)
	writer.WriteField("filename", fileName)
	writer.WriteField("index", fmt.Sprintf("%d", sliceIndex))
	writer.WriteField("action", "upload_slice")
	writer.Close()

	debugPrint(config, "请求URL: %s", config.UploadServer+"/app/upload_slice")
	debugPrint(config, "请求方法: POST")
	debugPrint(config, "Content-Type: %s", writer.FormDataContentType())
	debugPrint(config, "分片索引: %d", sliceIndex)
	debugPrint(config, "分片大小: %d bytes", len(chunkData))
	debugPrint(config, "uptoken: %s", upToken)
	debugPrint(config, "filename: %s", fileName)

	// 发送上传请求
	uploadReq, err := http.NewRequestWithContext(ctx, "POST", config.UploadServer+"/app/upload_slice", &buf)
	if err != nil {
		debugPrint(config, "创建上传请求失败: %v", err)
		return err
	}
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())

	uploadResp, err := client.Do(uploadReq)
	if err != nil {
		debugPrint(config, "发送上传请求失败: %v", err)
		return err
	}
	defer uploadResp.Body.Close()

	debugPrint(config, "HTTP状态码: %d", uploadResp.StatusCode)

	if uploadResp.StatusCode != http.StatusOK {
		debugPrint(config, "HTTP请求失败: %d", uploadResp.StatusCode)
		return fmt.Errorf("HTTP请求失败: %d", uploadResp.StatusCode)
	}

	uploadBody, err := io.ReadAll(uploadResp.Body)
	if err != nil {
		debugPrint(config, "读取上传响应失败: %v", err)
		return err
	}

	debugPrint(config, "上传响应内容: %s", string(uploadBody))

	var uploadResult struct {
		Status int `json:"status"`
	}

	if err := json.Unmarshal(uploadBody, &uploadResult); err != nil {
		debugPrint(config, "JSON解析失败: %v", err)
		return fmt.Errorf("解析上传响应失败: %w", err)
	}

	debugPrint(config, "解析结果 - 状态码: %d", uploadResult.Status)

	// 根据JavaScript代码，状态5表示分片上传完成
	if uploadResult.Status != 1 && uploadResult.Status != 2 && uploadResult.Status != 3 && uploadResult.Status != 5 {
		debugPrint(config, "分片上传失败，状态码: %d", uploadResult.Status)
		return fmt.Errorf("分片上传失败，状态码: %d", uploadResult.Status)
	}

	debugPrint(config, "分片 #%d 上传成功", sliceIndex)

	// 更新进度（支持续传）
	if progressCallback != nil {
		fileInfo, _ := os.Stat(filePath)

		// 简化的进度计算：基于分片索引 + 1（已完成的分片数）
		completedSlices := int64(sliceIndex + 1)
		totalUploadedBytes := completedSlices * int64(config.ChunkSize)

		// 确保不超过文件总大小
		if totalUploadedBytes > fileInfo.Size() {
			totalUploadedBytes = fileInfo.Size()
		}

		debugPrint(config, "进度更新: 分片#%d完成, 总进度: %d/%d bytes (%.1f%%)",
			sliceIndex, totalUploadedBytes, fileInfo.Size(),
			float64(totalUploadedBytes)/float64(fileInfo.Size())*100)

		progressCallback(totalUploadedBytes, fileInfo.Size())
	}

	return nil
}

// UploadInfo 上传信息
type UploadInfo struct {
	UToken string
	Server string
}

// getUTokenOnly 仅获取UToken（GUI模式使用）
func getUTokenOnly(ctx context.Context, config *Config, sha1Hash, fileName string, fileSize int64) (*UploadInfo, error) {
	debugPrint(config, "========== 获取UToken API ==========")
	client := &http.Client{}

	formData := fmt.Sprintf("action=upload_request_select2&sha1=%s&filename=%s&filesize=%d&model=%d&token=%s",
		sha1Hash, fileName, fileSize, config.Model, config.Token)

	debugPrint(config, "请求URL: %s", config.Server+"/file")
	debugPrint(config, "请求方法: POST")
	debugPrint(config, "Content-Type: application/x-www-form-urlencoded")
	debugPrint(config, "请求参数: %s", formData)

	req, err := http.NewRequestWithContext(ctx, "POST", config.Server+"/file", strings.NewReader(formData))
	if err != nil {
		debugPrint(config, "创建请求失败: %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		debugPrint(config, "发送请求失败: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	debugPrint(config, "HTTP状态码: %d", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		debugPrint(config, "读取响应失败: %v", err)
		return nil, err
	}

	debugPrint(config, "响应内容: %s", string(body))

	var selectResp struct {
		Status int `json:"status"`
		Data   struct {
			UToken  string      `json:"utoken"`
			Servers interface{} `json:"servers"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &selectResp); err != nil {
		debugPrint(config, "JSON解析失败: %v", err)
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	debugPrint(config, "解析结果 - 状态码: %d, UToken: %s", selectResp.Status, selectResp.Data.UToken)

	if selectResp.Status != 1 {
		debugPrint(config, "API返回错误状态: %d", selectResp.Status)
		return nil, fmt.Errorf("获取UToken失败，状态码: %d", selectResp.Status)
	}

	return &UploadInfo{
		UToken: selectResp.Data.UToken,
		Server: "", // 在调用方设置
	}, nil
}

// getUploadServers 获取上传服务器列表
func getUploadServers(ctx context.Context, config *Config, sha1Hash, fileName string, fileSize int64) (*UploadInfo, error) {
	debugPrint(config, "========== 获取上传服务器API ==========")
	client := &http.Client{}

	formData := fmt.Sprintf("action=upload_request_select2&sha1=%s&filename=%s&filesize=%d&model=%d&token=%s",
		sha1Hash, fileName, fileSize, config.Model, config.Token)

	debugPrint(config, "请求URL: %s", config.Server+"/file")
	debugPrint(config, "请求方法: POST")
	debugPrint(config, "Content-Type: application/x-www-form-urlencoded")
	debugPrint(config, "请求参数: %s", formData)

	req, err := http.NewRequestWithContext(ctx, "POST", config.Server+"/file", strings.NewReader(formData))
	if err != nil {
		debugPrint(config, "创建请求失败: %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		debugPrint(config, "发送请求失败: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	debugPrint(config, "HTTP状态码: %d", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		debugPrint(config, "读取响应失败: %v", err)
		return nil, err
	}

	debugPrint(config, "响应内容: %s", string(body))

	var selectResp struct {
		Status int `json:"status"`
		Data   struct {
			UToken  string      `json:"utoken"`
			Servers interface{} `json:"servers"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &selectResp); err != nil {
		debugPrint(config, "JSON解析失败: %v", err)
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	debugPrint(config, "解析结果 - 状态码: %d, UToken: %s, Servers: %v", selectResp.Status, selectResp.Data.UToken, selectResp.Data.Servers)

	if selectResp.Status != 1 {
		debugPrint(config, "API返回错误状态: %d", selectResp.Status)
		return nil, fmt.Errorf("获取上传服务器失败，状态码: %d", selectResp.Status)
	}

	// 解析第一个可用的上传服务器
	var uploadServer string
	if selectResp.Data.Servers != nil {
		if serverList, ok := selectResp.Data.Servers.([]interface{}); ok && len(serverList) > 0 {
			if serverObj, ok := serverList[0].(map[string]interface{}); ok {
				if serverURL, ok := serverObj["url"].(string); ok {
					uploadServer = serverURL
				}
			}
		}
	}

	if uploadServer == "" {
		return nil, fmt.Errorf("无法获取上传服务器地址")
	}

	// 检查是否用户强制指定了上传服务器
	if config.UploadServer != "" {
		debugPrint(config, "使用用户指定的上传服务器: %s", config.UploadServer)
		uploadServer = config.UploadServer
	} else {
		// 设置从API获取的上传服务器到配置中
		config.UploadServer = uploadServer
	}

	return &UploadInfo{
		UToken: selectResp.Data.UToken,
		Server: uploadServer,
	}, nil
}

// checkQuickUpload 检查是否可以秒传
func checkQuickUpload(ctx context.Context, config *Config, sha1Hash, fileName string, fileSize int64) (string, bool, error) {
	debugPrint(config, "========== 检查秒传API ==========")
	client := &http.Client{}

	formData := fmt.Sprintf("action=prepare_v4&sha1=%s&filename=%s&filesize=%d&model=%d&skip_upload=%d&token=%s",
		sha1Hash, fileName, fileSize, config.Model, config.SkipUpload, config.Token)

	debugPrint(config, "请求URL: %s", config.Server+"/file")
	debugPrint(config, "请求方法: POST")
	debugPrint(config, "Content-Type: application/x-www-form-urlencoded")
	debugPrint(config, "请求参数: %s", formData)

	req, err := http.NewRequestWithContext(ctx, "POST", config.Server+"/file", strings.NewReader(formData))
	if err != nil {
		debugPrint(config, "创建请求失败: %v", err)
		return "", false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		debugPrint(config, "发送请求失败: %v", err)
		return "", false, err
	}
	defer resp.Body.Close()

	debugPrint(config, "HTTP状态码: %d", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		debugPrint(config, "读取响应失败: %v", err)
		return "", false, err
	}

	debugPrint(config, "响应内容: %s", string(body))

	var prepareResp struct {
		Status int         `json:"status"`
		Data   interface{} `json:"data"`
	}

	if err := json.Unmarshal(body, &prepareResp); err != nil {
		debugPrint(config, "JSON解析失败: %v", err)
		return "", false, fmt.Errorf("解析响应失败: %w", err)
	}

	debugPrint(config, "解析结果 - 状态码: %d, 数据: %v", prepareResp.Status, prepareResp.Data)

	switch prepareResp.Status {
	case 6, 8:
		// 秒传成功
		if dataMap, ok := prepareResp.Data.(map[string]interface{}); ok {
			if ukey, exists := dataMap["ukey"].(string); exists {
				return fmt.Sprintf("https://tmp.link/f/%s", ukey), false, nil
			}
		}
		return "", false, fmt.Errorf("秒传响应格式错误")
	case 1:
		// 需要分片上传
		return "", true, nil
	case 0:
		// 需要分片上传
		return "", true, nil
	default:
		return "", false, fmt.Errorf("准备上传失败，状态码: %d", prepareResp.Status)
	}
}

// validateTokenAndGetUID 验证token并获取用户UID
func validateTokenAndGetUID(token, server string) (string, error) {
	client := &http.Client{}

	// 调用/user API验证token并获取用户信息
	formData := fmt.Sprintf("action=get_detail&token=%s", token)

	req, err := http.NewRequest("POST", server+"/user", strings.NewReader(formData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP请求失败: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var userResp struct {
		Status int `json:"status"`
		Data   struct {
			UID int64 `json:"uid"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &userResp); err != nil {
		return "", fmt.Errorf("解析用户响应失败: %w", err)
	}

	if userResp.Status != 1 {
		return "", fmt.Errorf("token验证失败，状态码: %d", userResp.Status)
	}

	if userResp.Data.UID == 0 {
		return "", fmt.Errorf("无法获取用户UID")
	}

	return fmt.Sprintf("%d", userResp.Data.UID), nil
}

// UploadTokens 上传所需的tokens
type UploadTokens struct {
	UToken       string // 从服务器获取的上传token
	UpToken      string // 客户端生成的上传token
	UploadServer string // 上传服务器地址
}

// prepareUpload 准备上传
func prepareUpload(ctx context.Context, config *Config, filePath, sha1Hash string, fileSize int64) (string, bool, *UploadTokens, error) {
	client := &http.Client{}

	// 第一步：upload_request_select2 - 获取上传服务器信息
	formData := fmt.Sprintf("action=upload_request_select2&sha1=%s&filename=%s&filesize=%d&model=%d&token=%s",
		sha1Hash, filepath.Base(filePath), fileSize, config.Model, config.Token)

	if config.MrID != "" {
		formData += "&mr_id=" + config.MrID
	}

	req, err := http.NewRequestWithContext(ctx, "POST", config.Server+"/file", strings.NewReader(formData))
	if err != nil {
		return "", false, nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", false, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", false, nil, fmt.Errorf("HTTP请求失败: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", false, nil, err
	}

	var selectResp struct {
		Status int `json:"status"`
		Data   struct {
			UToken  string      `json:"utoken"`
			Servers interface{} `json:"servers"`
		} `json:"data"`
	}

	// 调试信息
	fmt.Fprintf(os.Stderr, "upload_request_select2 响应: %s\n", string(body))

	if err := json.Unmarshal(body, &selectResp); err != nil {
		return "", false, nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if selectResp.Status != 1 {
		return "", false, nil, fmt.Errorf("获取上传服务器失败，状态码: %d", selectResp.Status)
	}

	// 解析servers字段
	var uploadServer string
	if selectResp.Data.Servers != nil {
		if serverList, ok := selectResp.Data.Servers.([]interface{}); ok && len(serverList) > 0 {
			// servers是对象数组，每个对象有url字段
			if serverObj, ok := serverList[0].(map[string]interface{}); ok {
				if serverURL, ok := serverObj["url"].(string); ok {
					uploadServer = serverURL
				}
			}
		}
	}

	// 检查是否用户强制指定了上传服务器
	if config.UploadServer != "" {
		uploadServer = config.UploadServer
		fmt.Fprintf(os.Stderr, "使用用户指定的上传服务器: %s\n", uploadServer)
	} else {
		if uploadServer == "" {
			// 如果没有找到有效的上传服务器，使用默认服务器
			uploadServer = strings.TrimSuffix(config.Server, "/api_v2")
		}
		fmt.Fprintf(os.Stderr, "使用API分配的上传服务器: %s\n", uploadServer)
	}

	// 生成uptoken (基于文件特征: SHA1(sha1 + filename + filesize + slice_size))
	upTokenData := fmt.Sprintf("%s%s%d%d", sha1Hash, filepath.Base(filePath), fileSize, config.ChunkSize)
	upTokenHash := sha1.Sum([]byte(upTokenData))
	upToken := hex.EncodeToString(upTokenHash[:])

	tokens := &UploadTokens{
		UToken:       selectResp.Data.UToken,
		UpToken:      upToken,
		UploadServer: uploadServer,
	}

	// 第二步：prepare_v4 - 准备文件上传
	prepareData := fmt.Sprintf("action=prepare_v4&sha1=%s&filename=%s&filesize=%d&model=%d&skip_upload=%d&token=%s",
		sha1Hash, filepath.Base(filePath), fileSize, config.Model, config.SkipUpload, config.Token)

	if config.MrID != "" {
		prepareData += "&mr_id=" + config.MrID
	}

	req2, err := http.NewRequestWithContext(ctx, "POST", config.Server+"/file", strings.NewReader(prepareData))
	if err != nil {
		return "", false, nil, err
	}

	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp2, err := client.Do(req2)
	if err != nil {
		return "", false, nil, err
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		return "", false, nil, fmt.Errorf("HTTP请求失败: %d", resp2.StatusCode)
	}

	body2, err := io.ReadAll(resp2.Body)
	if err != nil {
		return "", false, nil, err
	}

	var prepareResp struct {
		Status int         `json:"status"`
		Data   interface{} `json:"data"`
	}

	if err := json.Unmarshal(body2, &prepareResp); err != nil {
		return "", false, nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 调试信息
	fmt.Fprintf(os.Stderr, "prepare_v4 响应: %s\n", string(body2))

	switch prepareResp.Status {
	case 6, 8:
		// 秒传成功
		if dataMap, ok := prepareResp.Data.(map[string]interface{}); ok {
			if ukey, exists := dataMap["ukey"].(string); exists {
				return fmt.Sprintf("https://tmp.link/f/%s", ukey), false, nil, nil
			}
		}
		return "", false, nil, fmt.Errorf("秒传响应格式错误")

	case 1:
		// 需要分片上传
		return "", true, tokens, nil

	case 0:
		// 状态0：需要分片上传（根据CLAUDE.md，这是正常状态）
		return "", true, tokens, nil

	default:
		return "", false, nil, fmt.Errorf("准备上传失败，状态码: %d", prepareResp.Status)
	}
}

// uploadChunks 分片上传
func uploadChunks(ctx context.Context, config *Config, filePath, sha1Hash string, fileSize int64, tokens *UploadTokens, progressCallback func(int64, int64)) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	client := &http.Client{}
	chunkSize := int64(config.ChunkSize)
	totalChunks := int((fileSize + chunkSize - 1) / chunkSize)

	var uploadedBytes int64

	// 逐个上传分片
	for chunkIndex := 0; chunkIndex < totalChunks; chunkIndex++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		// 读取分片数据
		chunkData, err := readChunk(file, chunkIndex, chunkSize)
		if err != nil {
			return "", fmt.Errorf("读取分片 %d 失败: %w", chunkIndex, err)
		}

		// 上传分片
		err = uploadChunk(ctx, client, config, chunkIndex, chunkData, sha1Hash, filepath.Base(filePath), tokens, fileSize)
		if err != nil {
			return "", fmt.Errorf("上传分片 %d 失败: %w", chunkIndex, err)
		}

		// 更新进度
		uploadedBytes += int64(len(chunkData))
		if progressCallback != nil {
			progressCallback(uploadedBytes, fileSize)
		}
	}

	// 所有分片上传完成，通过再次调用prepare来获取最终状态和下载链接
	return getFinalResult(ctx, client, config, sha1Hash, filepath.Base(filePath), fileSize, tokens)
}

// readChunk 读取分片数据
func readChunk(file *os.File, chunkIndex int, chunkSize int64) ([]byte, error) {
	offset := int64(chunkIndex) * chunkSize
	_, err := file.Seek(offset, 0)
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, chunkSize)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return buffer[:n], nil
}

// uploadChunk 上传单个分片
func uploadChunk(ctx context.Context, client *http.Client, config *Config, chunkIndex int, chunkData []byte, sha1Hash, fileName string, tokens *UploadTokens, totalFileSize int64) error {
	// 第一步：查询分片信息 (prepare) - 包含所有必需参数
	prepareData := fmt.Sprintf("token=%s&uptoken=%s&action=prepare&sha1=%s&filename=%s&filesize=%d&slice_size=%d&utoken=%s&mr_id=%s&model=%d",
		config.Token, tokens.UpToken, sha1Hash, fileName, totalFileSize, config.ChunkSize, tokens.UToken, config.MrID, config.Model)

	// API参数已准备

	// 使用tokens中的上传服务器地址
	req, err := http.NewRequestWithContext(ctx, "POST", tokens.UploadServer+"/app/upload_slice", strings.NewReader(prepareData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// 调试信息
	fmt.Fprintf(os.Stderr, "分片prepare响应: %s\n", string(body))

	var prepareResp struct {
		Status int `json:"status"`
		Data   struct {
			Next int `json:"next"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &prepareResp); err != nil {
		return fmt.Errorf("解析prepare响应失败: %w", err)
	}

	// 检查是否需要上传此分片
	if prepareResp.Status == 1 {
		return nil // 上传完成
	}

	if prepareResp.Status != 3 || prepareResp.Data.Next != chunkIndex {
		return fmt.Errorf("分片prepare失败，状态码: %d, 期待分片: %d, 实际: %d", prepareResp.Status, chunkIndex, prepareResp.Data.Next)
	}

	// 第二步：上传分片数据
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加文件数据
	fileWriter, err := writer.CreateFormFile("filedata", "slice")
	if err != nil {
		return err
	}
	fileWriter.Write(chunkData)

	// 添加表单字段
	writer.WriteField("uptoken", tokens.UpToken)
	writer.WriteField("filename", fileName)
	writer.WriteField("index", fmt.Sprintf("%d", chunkIndex))
	writer.WriteField("action", "upload_slice")
	writer.Close()

	uploadReq, err := http.NewRequestWithContext(ctx, "POST", tokens.UploadServer+"/app/upload_slice", &buf)
	if err != nil {
		return err
	}
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())

	uploadResp, err := client.Do(uploadReq)
	if err != nil {
		return err
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP请求失败: %d", uploadResp.StatusCode)
	}

	uploadBody, err := io.ReadAll(uploadResp.Body)
	if err != nil {
		return err
	}

	var uploadResult struct {
		Status int `json:"status"`
	}

	if err := json.Unmarshal(uploadBody, &uploadResult); err != nil {
		return fmt.Errorf("解析上传响应失败: %w", err)
	}

	if uploadResult.Status != 1 && uploadResult.Status != 2 && uploadResult.Status != 3 && uploadResult.Status != 5 {
		return fmt.Errorf("分片上传失败，状态码: %d", uploadResult.Status)
	}

	return nil
}

// getFinalResult 获取最终上传结果
func getFinalResult(ctx context.Context, client *http.Client, config *Config, sha1Hash, fileName string, fileSize int64, tokens *UploadTokens) (string, error) {
	// 调用prepare检查最终状态
	prepareData := fmt.Sprintf("token=%s&uptoken=%s&action=prepare&sha1=%s&filename=%s&filesize=%d&slice_size=%d&utoken=%s&mr_id=%s&model=%d",
		config.Token, tokens.UpToken, sha1Hash, fileName, fileSize, config.ChunkSize, tokens.UToken, config.MrID, config.Model)

	req, err := http.NewRequestWithContext(ctx, "POST", tokens.UploadServer+"/app/upload_slice", strings.NewReader(prepareData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 调试信息
	fmt.Fprintf(os.Stderr, "最终状态检查响应: %s\n", string(body))

	var finalResp struct {
		Status int         `json:"status"`
		Data   interface{} `json:"data"`
	}

	if err := json.Unmarshal(body, &finalResp); err != nil {
		return "", fmt.Errorf("解析最终响应失败: %w", err)
	}

	switch finalResp.Status {
	case 1, 6, 8, 9:
		// 上传完成，从data中获取ukey
		if dataMap, ok := finalResp.Data.(map[string]interface{}); ok {
			if ukey, exists := dataMap["ukey"].(string); exists {
				return fmt.Sprintf("https://tmp.link/f/%s", ukey), nil
			}
		}
		// 如果data是字符串（某些情况下）
		if ukey, ok := finalResp.Data.(string); ok && ukey != "" {
			return fmt.Sprintf("https://tmp.link/f/%s", ukey), nil
		}
		return "", fmt.Errorf("无法从响应中获取下载链接")
	default:
		return "", fmt.Errorf("上传未完成，状态码: %d", finalResp.Status)
	}
}

// getDownloadURL 获取下载链接（已弃用）
func getDownloadURL(ctx context.Context, client *http.Client, config *Config, sha1Hash, fileName string, fileSize int64) (string, error) {
	// 构造完成请求
	formData := fmt.Sprintf("action=upload_complete&sha1=%s&filename=%s&filesize=%d&token=%s",
		sha1Hash, fileName, fileSize, config.Token)

	req, err := http.NewRequestWithContext(ctx, "POST", config.Server+"/file", strings.NewReader(formData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP请求失败: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 调试信息
	fmt.Fprintf(os.Stderr, "upload_complete 响应: %s\n", string(body))

	var completeResp struct {
		Status int         `json:"status"`
		Data   interface{} `json:"data"`
	}

	if err := json.Unmarshal(body, &completeResp); err != nil {
		return "", fmt.Errorf("解析完成响应失败: %w", err)
	}

	if completeResp.Status == 1 {
		if dataMap, ok := completeResp.Data.(map[string]interface{}); ok {
			if ukey, exists := dataMap["ukey"].(string); exists {
				return fmt.Sprintf("https://tmp.link/f/%s", ukey), nil
			}
		}
	}

	return "", fmt.Errorf("获取下载链接失败，状态码: %d", completeResp.Status)
}

// clearProgressBar 清除进度条残留和开始信息
func clearProgressBar() {
	// 清除我们输出的内容：进度条 + 文件大小行 + 开始上传行（共3行）
	fmt.Print("\r\033[K")     // 清除当前行（进度条）
	fmt.Print("\033[1A\033[K") // 向上移动一行并清除（文件大小行）
	fmt.Print("\033[1A\033[K") // 向上移动一行并清除（开始上传行）
	// 现在光标在开始上传行的位置，准备输出完成信息
}

// createProgressCallback 创建进度回调函数
func createProgressCallback(cliMode bool, fileSize int64, speedCalc *SpeedCalculator, task *TaskStatus, statusFile string) func(int64, int64) {
	var bar *progressbar.ProgressBar

	// 如果是CLI模式，只显示开始信息，不立即创建进度条
	if cliMode {
		fmt.Printf("🚀 开始上传文件: %s\n", task.FileName)
		fmt.Printf("📊 文件大小: %s\n", formatBytes(fileSize))
	}

	return func(uploaded, total int64) {
		progress := float64(uploaded) / float64(total) * 100

		// 计算上传速度
		speed := speedCalc.UpdateSpeed(uploaded)

		// 更新任务状态
		task.Status = "uploading"
		task.Progress = progress
		task.UploadSpeed = speed
		task.UpdatedAt = time.Now()

		// CLI模式：惰性创建和更新进度条
		if cliMode {
			// 只在第一次调用时创建进度条
			if bar == nil {
				bar = progressbar.NewOptions64(
					total,
					progressbar.OptionSetDescription("📤 上传中"),
					progressbar.OptionSetWidth(40),
					progressbar.OptionShowBytes(true),
					progressbar.OptionSetTheme(progressbar.Theme{
						Saucer:        "█",
						SaucerHead:    "█",
						SaucerPadding: "░",
						BarStart:      "[",
						BarEnd:        "]",
					}),
					progressbar.OptionShowIts(),
					progressbar.OptionShowCount(),
					progressbar.OptionSetPredictTime(true),
					progressbar.OptionShowDescriptionAtLineEnd(),
					// 移除 OptionSetRenderBlankState 防止立即显示空进度条
				)
			}
			bar.Set64(uploaded)
		}

		// GUI模式：保存状态到文件
		if !cliMode {
			if err := saveTaskStatus(statusFile, task); err != nil {
				fmt.Fprintf(os.Stderr, "警告: 保存进度失败: %v\n", err)
			}
		}
	}
}

// showConfigStatus 显示当前配置状态和token有效性
func showConfigStatus() {
	fmt.Println("=== 钛盘上传工具配置状态 ===")
	fmt.Println()

	// 加载配置
	config := loadCLIConfig()
	configPath := getCLIConfigPath()

	// 获取已解析的命令行参数值
	chunkSizeFlag := flag.Lookup("chunk-size")
	modelFlag := flag.Lookup("model")
	mrIDFlag := flag.Lookup("mr-id")
	skipUploadFlag := flag.Lookup("skip-upload")
	debugFlag := flag.Lookup("debug")

	// 确定最终使用的值（命令行参数优先级高于配置文件）
	var finalChunkSize int = 3
	if chunkSizeFlag != nil {
		if val, err := strconv.Atoi(chunkSizeFlag.Value.String()); err == nil {
			finalChunkSize = val
		}
	}

	var finalModel int = config.Model
	if modelFlag != nil && isFlagSet(modelFlag) {
		if val, err := strconv.Atoi(modelFlag.Value.String()); err == nil {
			finalModel = val
		}
	}

	var finalMrID string = config.MrID
	if mrIDFlag != nil && isFlagSet(mrIDFlag) {
		finalMrID = mrIDFlag.Value.String()
	}

	var finalSkipUpload int = 1
	if skipUploadFlag != nil {
		if val, err := strconv.Atoi(skipUploadFlag.Value.String()); err == nil {
			finalSkipUpload = val
		}
	}

	var finalDebug bool = false
	if debugFlag != nil {
		if val, err := strconv.ParseBool(debugFlag.Value.String()); err == nil {
			finalDebug = val
		}
	}

	// 显示配置文件信息
	fmt.Printf("📁 配置文件路径: %s\n", configPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("⚠️  配置文件状态: 不存在 (使用默认值)\n")
	} else {
		fmt.Printf("✅ 配置文件状态: 存在\n")
	}
	fmt.Println()

	// 显示Token信息
	fmt.Println("🔑 Token配置:")
	if config.Token == "" {
		fmt.Printf("   状态: ❌ 未设置\n")
		fmt.Printf("   建议: 使用 -set-token 命令设置API Token\n")
	} else {
		fmt.Printf("   状态: ✅ 已设置\n")
		fmt.Printf("   长度: %d 字符\n", len(config.Token))
		fmt.Printf("   前缀: %s...\n", config.Token[:min(8, len(config.Token))])

		// 验证Token有效性
		fmt.Printf("   验证: ")
		server := "https://tmplink-sec.vxtrans.com/api_v2"
		if uid, err := validateTokenAndGetUID(config.Token, server); err != nil {
			fmt.Printf("❌ 无效 (%v)\n", err)
		} else {
			fmt.Printf("✅ 有效 (UID: %s)\n", uid)
		}
	}
	fmt.Println()

	// 显示其他配置
	fmt.Println("⚙️ 其他配置:")
	modelDesc := map[int]string{0: "24小时", 1: "3天", 2: "7天", 99: "无限期"}
	fmt.Printf("   文件有效期: %s (%d)\n", modelDesc[finalModel], finalModel)
	fmt.Printf("   目录ID: %s\n", finalMrID)
	fmt.Println()

	// 显示当前运行参数
	fmt.Println("🔧 当前运行参数:")
	fmt.Printf("   分块大小: %dMB\n", finalChunkSize)
	fmt.Printf("   跳过上传: %d (%s)\n", finalSkipUpload, map[int]string{0: "禁用秒传检查", 1: "启用秒传检查"}[finalSkipUpload])
	
	debugStatus := "关闭"
	if finalDebug {
		debugStatus = "开启"
	}
	fmt.Printf("   调试模式: %s\n", debugStatus)
	fmt.Println()

	// 显示使用建议
	if config.Token == "" {
		fmt.Println("💡 下一步建议:")
		fmt.Println("   1. 访问 https://tmp.link/ 并登录")
		fmt.Println("   2. 在上传界面点击'重新设定' -> '命令行上传'复制Token")
		fmt.Println("   3. 运行: ./tmplink-cli -set-token YOUR_TOKEN")
		fmt.Println("   4. 然后就可以上传文件了: ./tmplink-cli -file /path/to/file")
	} else {
		fmt.Println("✨ 配置完成，现在可以上传文件:")
		fmt.Println("   ./tmplink-cli -file /path/to/your/file")
	}
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// formatBytes 格式化字节数为可读格式
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
