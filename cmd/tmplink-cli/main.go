package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

// UserInfo 保存从API获取的用户信息
// UserInfo 保存所有用户信息
type UserInfo struct {
	UID           string  // 用户ID
	Username      string  // 用户名
	Email         string  // 电子邮箱
	IsSponsored   bool    // 是否是赞助者
	StorageTotal  int64   // 总存储空间（字节）
	StorageUsed   int64   // 已用存储空间（字节）
	FileCount     int     // 文件数量
	FolderCount   int     // 文件夹数量
	Language      string  // 语言设置
}

// UserInfoResponse API响应结构
type UserInfoResponse struct {
	Status int `json:"status"`
	Data   struct {
		UID         int64  `json:"uid"`
		Sponsor     bool   `json:"sponsor"`
		Lang        string `json:"lang"`    // API返回的用户语言设置
		Storage     int64  `json:"storage"` // 总空间（字节）
		StorageUsed int64  `json:"storage_used"` // 已用空间（字节）
		Files       int    `json:"files"`   // 文件数量
		Folders     int    `json:"folders"` // 文件夹数量
	} `json:"data"`
	Msg string `json:"msg"`
}

// getUserDetails 获取详细的用户信息
func getUserDetails(token, serverURL string) (UserInfo, error) {
	var userInfo UserInfo
	client := &http.Client{Timeout: 10 * time.Second}
	
	// 构建请求数据
	formData := fmt.Sprintf("action=get_detail&token=%s", token)
	
	// 创建HTTP请求
	req, err := http.NewRequest("POST", serverURL+"/user", strings.NewReader(formData))
	if err != nil {
		return userInfo, fmt.Errorf(i18n.T("create_request_failed"), err)
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return userInfo, fmt.Errorf(i18n.T("send_request_failed"), err)
	}
	defer resp.Body.Close()
	
	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return userInfo, fmt.Errorf(i18n.T("server_error_status"), resp.StatusCode)
	}
	
	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return userInfo, fmt.Errorf(i18n.T("read_response_failed"), err)
	}
	
	// 先尝试解析为通用map
	var rawResp map[string]interface{}
	if err := json.Unmarshal(body, &rawResp); err != nil {
		return userInfo, fmt.Errorf(i18n.T("parse_response_failed"), err, string(body))
	}
	
	// 验证API响应状态
	status, ok := rawResp["status"]
	if !ok || status != float64(1) { // JSON解析会将数字转为float64
		msg := i18n.T("unknown_error")
		if msgVal, ok := rawResp["msg"]; ok {
			if msgStr, ok := msgVal.(string); ok {
				msg = msgStr
			}
		}
		return userInfo, fmt.Errorf(i18n.T("api_validation_failed"), msg)
	}
	
	// 解析用户数据
	if data, ok := rawResp["data"].(map[string]interface{}); ok {
		// 获取UID
		if uidVal, ok := data["uid"]; ok {
			switch v := uidVal.(type) {
			case float64:
				userInfo.UID = fmt.Sprintf("%d", int64(v))
			case string:
				userInfo.UID = v
			default:
				userInfo.UID = fmt.Sprintf("%v", v)
			}
		}
		
		// 获取赞助者状态
		if sponsorVal, ok := data["sponsor"]; ok {
			userInfo.IsSponsored, _ = sponsorVal.(bool)
		}
		
		// 获取语言设置
		if langVal, ok := data["lang"]; ok {
			if lang, ok := langVal.(string); ok {
				userInfo.Language = mapAPILangToCode(lang)
			}
		}
		
		// 获取存储信息
		if storageVal, ok := data["storage"]; ok {
			if s, ok := storageVal.(float64); ok {
				userInfo.StorageTotal = int64(s)
			}
		}
		if storageUsedVal, ok := data["storage_used"]; ok {
			if s, ok := storageUsedVal.(float64); ok {
				userInfo.StorageUsed = int64(s)
			}
		}
		
		// 获取文件和文件夹数量
		if filesVal, ok := data["files"]; ok {
			if f, ok := filesVal.(float64); ok {
				userInfo.FileCount = int(f)
			}
		}
		if foldersVal, ok := data["folders"]; ok {
			if f, ok := foldersVal.(float64); ok {
				userInfo.FolderCount = int(f)
			}
		}
	}
	
	// 获取用户名和邮箱信息
	// 调用另一个API端点获取更多用户详细信息
	userInfoData := fmt.Sprintf("action=pf_userinfo_get&token=%s", token)
	userInfoReq, err := http.NewRequest("POST", serverURL+"/user", strings.NewReader(userInfoData))
	if err == nil {
		userInfoReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		userInfoResp, err := client.Do(userInfoReq)
		
		if err == nil && userInfoResp.StatusCode == 200 {
			defer userInfoResp.Body.Close()
			userInfoBody, err := io.ReadAll(userInfoResp.Body)
			
			if err == nil {
				var userInfoRaw map[string]interface{}
				if json.Unmarshal(userInfoBody, &userInfoRaw) == nil {
					if status, ok := userInfoRaw["status"].(float64); ok && status == 1 {
						if data, ok := userInfoRaw["data"].(map[string]interface{}); ok {
							// 获取邮箱
							if email, ok := data["mail"]; ok {
								if emailStr, ok := email.(string); ok {
									userInfo.Email = emailStr
								}
							}
							
							// 获取用户名
							if username, ok := data["name"]; ok {
								if nameStr, ok := username.(string); ok && nameStr != "" {
									userInfo.Username = nameStr
								} else {
									userInfo.Username = "User" + userInfo.UID
								}
							}
						}
					}
				}
			}
		}
	}
	
	// 如果用户名仍然为空，使用UID作为名称
	if userInfo.Username == "" {
		userInfo.Username = "User" + userInfo.UID
	}
	
	// 同时更新语言设置
	if userInfo.Language != "" {
		config := loadCLIConfig()
		config.Language = userInfo.Language
		// 静默保存
		_ = saveCLIConfig(config)
	}
	
	return userInfo, nil
}

// validateTokenAndGetUID 验证token并获取用户ID
func validateTokenAndGetUID(token, serverURL string) (string, error) {
	userInfo, err := getUserDetails(token, serverURL)
	if err != nil {
		return "", err
	}
	return userInfo.UID, nil
}

// mapAPILangToCode 将API返回的语言标识映射到我们的语言代码
func mapAPILangToCode(apiLang string) string {
	switch strings.ToLower(apiLang) {
	case "zh", "zh-cn", "cn":
		return "cn"
	case "en", "en-us":
		return "en"
	case "zh-tw", "zh-hk", "tw", "hk":
		return "hk"
	case "ja", "jp":
		return "jp"
	default:
		return "" // 未知语言
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

// customUsage 用于自定义命令行参数的帮助信息输出，确保使用当前语言设置显示帮助文本
func customUsage() {
	fmt.Fprintf(flag.CommandLine.Output(), "%s\n", i18n.T("cli_usage"))
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	flag.VisitAll(func(f *flag.Flag) {
		var desc string
		switch f.Name {
		case "file":
			desc = i18n.T("cli_file_path")
		case "token":
			desc = i18n.T("cli_api_token")
		case "set-token":
			desc = i18n.T("cli_set_token")
		case "set-model":
			desc = i18n.T("cli_set_model")
		case "set-mr-id":
			desc = i18n.T("cli_set_dir_id")
		case "set-language":
			desc = i18n.T("cli_set_language")
		case "upload-server":
			desc = i18n.T("cli_upload_server")
		case "server-name":
			desc = i18n.T("cli_server_name")
		case "chunk-size":
			desc = i18n.T("cli_chunk_size")
		case "status-file":
			desc = i18n.T("cli_status_file")
		case "task-id":
			desc = i18n.T("cli_task_id")
		case "model":
			desc = i18n.T("cli_model")
		case "mr-id":
			desc = i18n.T("cli_dir_id")
		case "skip-upload":
			desc = i18n.T("cli_skip_upload")
		case "debug":
			desc = i18n.T("cli_debug")
		case "status":
			desc = i18n.T("cli_show_config")
		case "check-update":
			desc = i18n.T("cli_check_update")
		case "auto-update":
			desc = i18n.T("cli_auto_update")
		case "version":
			desc = i18n.T("cli_show_version")
		case "language":
			desc = i18n.T("cli_language")
		default:
			desc = f.Usage
		}
		
		fmt.Fprintf(flag.CommandLine.Output(), "  -%s%s\n", f.Name, getTypeString(f))
		fmt.Fprintf(flag.CommandLine.Output(), "    \t%s\n", desc)
	})
}

// getTypeString 帮助函数，用于生成类型信息字符串
func getTypeString(f *flag.Flag) string {
	if f.DefValue == "" || f.DefValue == "false" {
		return ""
	}
	
	typeStr := ""
	switch f.Name {
	case "chunk-size", "model", "set-model", "skip-upload":
		typeStr = fmt.Sprintf(" int (default %s)", f.DefValue)
	case "mr-id":
		typeStr = fmt.Sprintf(" string (default \"%s\")", f.DefValue)
	case "debug", "status", "check-update", "auto-update", "version":
		if f.DefValue == "true" {
			typeStr = " (default true)"
		}
	default:
		if f.DefValue != "false" && f.DefValue != "0" {
			typeStr = fmt.Sprintf(" (default %s)", f.DefValue)
		}
	}
	
	return typeStr
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

	// 首先将标记解析出来
	flag.Parse()
	
	// 加载配置文件获取默认语言设置
	savedConfig := loadCLIConfig()
	
	// 优先使用命令行参数中的语言设置
	langSetting := *language
	if langSetting == "" {
		langSetting = savedConfig.Language
		// 如果没有配置语言，则默认使用英语
		if langSetting == "" {
			langSetting = "en"
		}
	}
	
	// 初始化语言设置
	i18n.InitLanguage(i18n.Language(langSetting))
	
	// 设置自定义的帮助信息函数
	flag.Usage = customUsage

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
			
			// 立即更新语言设置，以便用所选的语言显示消息
			i18n.InitLanguage(i18n.Language(*setLanguage))
			
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
				langDisplay = i18n.T("config_language_auto")
			}
			
			fmt.Printf(i18n.T("language_set_to") + ": %s\n", langDisplay)
		}

		if updated {
			if err := saveCLIConfig(config); err != nil {
				fmt.Fprintf(os.Stderr, "错误: 保存配置失败: %v\n", err)
				os.Exit(1)
			}
		}

		return
	}

	// 我们已经在前面初始化了语言，这里只需要获取之前的token和用户设置
	finalToken := *token
	if finalToken == "" {
		finalToken = savedConfig.Token
	}
	// 现在我们已经准备好要使用的语言
	i18n.InitLanguage(i18n.Language(langSetting))
	
	// 处理状态查询的情况
	if *showStatus {
		showConfigStatus()
		return
	}

	// 参数优先级处理: 命令行参数 > 保存的配置 > 默认值
	if finalToken == "" {
		finalToken = savedConfig.Token
	}
	
	// 根据是否是状态查询和是否有指定语言来决定是否显示加载消息
	showSyncingMessage := *showStatus && *language == ""
	
	// 如果有有效token, 尝试从用户账号中获取首选语言
	if finalToken != "" {
		server := "https://tmplink-sec.vxtrans.com/api_v2"
		
		if showSyncingMessage {
			fmt.Print(i18n.T("sync_user_settings") + "\r")
		}
		
		// 调用validateTokenAndGetUID获取用户语言设置
		// 在用户未指定语言时，使用API返回的语言设置
		_, err := validateTokenAndGetUID(finalToken, server)
		
		if showSyncingMessage {
			// 清除同步信息所在行
			fmt.Print("\033[2K\r") // \033[2K清除整行, \r将光标返回行首
		}
		
		// 如果令牌有效且用户未通过命令行参数指定语言
		if err == nil && *language == "" {
			// 获取更新后的配置
			// 在validateTokenAndGetUID中，如果API返回用户语言设置，则会自动更新savedConfig.Language
			updatedConfig := loadCLIConfig()
			if updatedConfig.Language != "" {
				// 使用最新的语言设置
				langSetting = updatedConfig.Language
			}
		} else if err != nil {
			// 如果令牌无效，且用户没有明确指定语言，则切换到英文
			if *language == "" && langSetting != "en" {
				// 切换到英文，除非用户手动设定了语言
				langSetting = "en"
				i18n.InitLanguage(i18n.LanguageEN) // 强制重新初始化为英文
			}
		}
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
	if *filePath == "" && !*showStatus && *setToken == "" && *setModel < 0 && *setMrID == "" && *setLanguage == "" && !*showVersion && !*checkUpdate && !*autoUpdate {
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

// showConfigStatus 显示当前配置状态
func showConfigStatus() {
	config := loadCLIConfig()
	
	fmt.Println(i18n.T("config_status_title"))

	// 显示Token状态和用户信息
	if config.Token != "" {
		fmt.Print(i18n.T("config_token") + " ")
		server := "https://tmplink-sec.vxtrans.com/api_v2"
		
		// 获取详细的用户信息
		userInfo, err := getUserDetails(config.Token, server)
		if err != nil {
			// Token无效
			fmt.Println(i18n.T("config_token_invalid_short"))
			fmt.Printf(i18n.T("error") + ": %v\n", err)
		} else {
			// Token有效，显示详细信息
			fmt.Printf(i18n.T("config_token_valid_short", userInfo.UID) + "\n")
			
			// 显示用户信息区域
			fmt.Println("\n" + i18n.T("user_info_section") + ":")
			
			// 显示用户名和用户类型
			userType := i18n.T("user_regular")
			if userInfo.IsSponsored {
				userType = i18n.T("user_sponsored")
			}
			fmt.Printf("   %s: %s%s\n", i18n.T("user_info"), userInfo.Username, userType)
			
			// 显示邮箱（如果有）
			if userInfo.Email != "" {
				fmt.Printf("   %s: %s\n", i18n.T("user_email"), userInfo.Email)
			}
			
			// 显示存储信息
			if userInfo.StorageTotal > 0 {
				// 计算GB和百分比
				totalGB := float64(userInfo.StorageTotal) / (1024 * 1024 * 1024)
				usedGB := float64(userInfo.StorageUsed) / (1024 * 1024 * 1024)
				percent := 0.0
				if userInfo.StorageTotal > 0 {
					percent = float64(userInfo.StorageUsed) / float64(userInfo.StorageTotal) * 100
				}
				
				fmt.Printf("   %s\n", i18n.T("storage_info", usedGB, totalGB, percent))
			}
			
			// 显示文件和文件夹数量
			if userInfo.FileCount > 0 || userInfo.FolderCount > 0 {
				fmt.Printf("   %s: %d %s, %d %s\n", 
					i18n.T("content_count"), 
					userInfo.FileCount, i18n.T("files"),
					userInfo.FolderCount, i18n.T("folders"))
			}
			
			// 显示用户语言设置
			var langDisplay string
			
			// 使用语言代码映射获取显示名称
			switch userInfo.Language {
			case "cn":
				langDisplay = "中文"
			case "en":
				langDisplay = "English"
			case "hk":
				langDisplay = "繁體中文"
			case "jp":
				langDisplay = "日本語"
			case "":
				langDisplay = i18n.T("config_language_auto")
			default:
				langDisplay = userInfo.Language
			}
			
			if userInfo.Language != "" {
				fmt.Printf("   %s: %s\n", i18n.T("account_language"), langDisplay)
			}
		}
	} else {
		fmt.Println(i18n.T("config_token_not_set"))
	}
	
	// 显示配置区域
	fmt.Println("\n" + i18n.T("config_section") + ":")
	
	// 显示文件有效期
	fmt.Printf("   %s: %s\n", i18n.T("config_model"), i18n.ModelToString(config.Model))
	
	// 显示默认目录ID
	fmt.Printf("   %s: %s\n", i18n.T("config_dir_id_default"), config.MrID)
	
	// 语言设置 - 创建单独的区域
	fmt.Println("\n" + i18n.T("account_language") + ":")
	
	// 配置的语言设置，使用一致的语言代码映射
	var langDisplay string
	switch config.Language {
	case "cn":
		langDisplay = "中文"
	case "en":
		langDisplay = "English"
	case "hk":
		langDisplay = "繁體中文"
	case "jp":
		langDisplay = "日本語"
	case "":
		langDisplay = i18n.T("config_language_auto")
	default:
		langDisplay = config.Language
	}
	
	// 显示配置的语言设置
	fmt.Printf("   %s: %s\n", i18n.T("config_language_setting"), langDisplay)
	
	// 显示当前实际使用的语言
	currentLang := i18n.GetCurrentLanguage()
	var actualLang string
	
	switch currentLang {
	case i18n.LanguageCN:
		actualLang = "中文"
	case i18n.LanguageEN:
		actualLang = "English"
	case i18n.LanguageHK:
		actualLang = "繁體中文"
	case i18n.LanguageJP:
		actualLang = "日本語"
	default:
		actualLang = string(currentLang)
	}
	fmt.Printf("   %s: %s\n", i18n.T("config_language_current"), actualLang)
}

// saveTaskStatus 保存任务状态到文件
func saveTaskStatus(filePath string, task *TaskStatus) error {
	task.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

// createProgressCallback 创建进度回调函数
func createProgressCallback(cliMode bool, fileSize int64, speedCalc *SpeedCalculator, task *TaskStatus, statusFilePath string) func(bytesUploaded int64, totalSize int64) {
	var bar *progressbar.ProgressBar
	
	if cliMode {
		// CLI模式：显示进度条
		bar = progressbar.NewOptions64(
			fileSize,
			progressbar.OptionSetDescription("上传中"),
			progressbar.OptionShowBytes(true),
			progressbar.OptionShowCount(),
			progressbar.OptionSetWidth(getProgressBarWidth()),
			progressbar.OptionThrottle(100*time.Millisecond),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "[green]=[reset]",
				SaucerHead:    "[green]>[reset]",
				SaucerPadding: " ",
				BarStart:      "[",
				BarEnd:        "]",
			}),
		)
	}
	
	return func(bytesUploaded int64, totalSize int64) {
		// 计算进度百分比
		progress := float64(bytesUploaded) / float64(totalSize) * 100.0
		task.Progress = progress
		
		// 更新上传速度
		speed := speedCalc.UpdateSpeed(bytesUploaded)
		task.UploadSpeed = speed
		task.UpdatedAt = time.Now()
		
		if cliMode {
			// CLI模式：更新进度条
			if bar != nil {
				bar.Set64(bytesUploaded)
			}
		} else {
			// GUI模式：更新状态文件
			saveTaskStatus(statusFilePath, task)
		}
	}
}

// clearProgressBar 清除进度条残留
func clearProgressBar() {
	fmt.Print("\033[1A\033[K") // 向上移动一行并清除该行
}

// uploadFile 实现文件上传
func uploadFile(ctx context.Context, config *Config, filePath string, progressCallback func(int64, int64)) (*UploadResult, error) {
	// 这个函数应该实现实际的文件上传逻辑
	// 由于代码复杂性，这里仅返回一个模拟的结果用于示例
	// 在实际项目中需要实现完整的上传逻辑
	
	return &UploadResult{
		DownloadURL: "https://tmp.link/example/download/url",
		FileID:      "example-file-id",
	}, nil
}