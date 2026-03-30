package tui

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"tmplink_uploader/internal/i18n"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// 应用状态
type State int

const (
	StateInit                  State = iota // 初始化状态
	StateLanguageSelect                     // 语言选择（首次运行）
	StateTokenInput                         // Token输入
	StateTokenValidationFailed              // Token验证失败等待状态
	StateMain                               // 主界面（文件浏览器）
	StateSettings                           // 上传设置
	StateUploadList                         // 上传管理器
	StateError                              // 错误状态
)

// 用户信息
type UserInfo struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	UsedSpace   int64  `json:"used_space"`
	TotalSpace  int64  `json:"total_space"`
	IsSponsored bool   `json:"is_sponsored"`
	UID         string `json:"uid"`
}

// 上传服务器信息
type ServerOption struct {
	Name string // 显示名称
	URL  string // 实际URL
}

// 配置结构
type Config struct {
	Token              string    `json:"token"`
	UploadServer       string    `json:"upload_server"`
	SelectedServerName string    `json:"selected_server_name"` // 选中的服务器名称
	ChunkSize          int       `json:"chunk_size"`           // 存储MB数
	MaxConcurrent      int       `json:"max_concurrent"`
	QuickUpload        bool      `json:"quick_upload"`
	SkipUpload         bool      `json:"skip_upload"`
	LastUpdateCheck    time.Time `json:"last_update_check"`    // 最后一次更新检查时间
	Language           string    `json:"language"`             // 界面语言
	// CLI专用字段
	Model int    `json:"model"` // CLI文件过期模式
	MrID  string `json:"mr_id"` // CLI目录ID
}

// getAvailableServers 从API获取可用的上传服务器列表
func getAvailableServers(token string) ([]ServerOption, error) {
	var servers []ServerOption

	// 如果没有token，返回空列表
	if token == "" {
		return servers, nil
	}

	// 调用API获取服务器列表
	apiServers, err := fetchServerListFromAPI(token)
	if err != nil {
		// 如果API调用失败，返回空列表和错误
		return servers, err
	}

	// 直接使用从API获取的服务器列表
	return apiServers, nil
}

// fetchServerListFromAPI 从API获取服务器列表
func fetchServerListFromAPI(token string) ([]ServerOption, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	// 使用upload_request_select2 API获取服务器列表
	// 需要提供一个虚拟文件信息来获取服务器列表
	formData := fmt.Sprintf("action=upload_request_select2&sha1=dummy&filename=dummy.txt&filesize=1024&model=1&token=%s", token)

	req, err := http.NewRequest("POST", "https://tmplink-sec.vxtrans.com/api_v2/file", strings.NewReader(formData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp struct {
		Status int `json:"status"`
		Data   struct {
			Servers interface{} `json:"servers"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析API响应失败: %w", err)
	}

	if apiResp.Status != 1 {
		return nil, fmt.Errorf("API返回错误状态: %d", apiResp.Status)
	}

	var servers []ServerOption

	// 解析servers字段
	if apiResp.Data.Servers != nil {
		if serverList, ok := apiResp.Data.Servers.([]interface{}); ok {
			for _, serverItem := range serverList {
				if serverObj, ok := serverItem.(map[string]interface{}); ok {
					if title, hasTitle := serverObj["title"].(string); hasTitle {
						if url, hasURL := serverObj["url"].(string); hasURL {
							servers = append(servers, ServerOption{
								Name: title,
								URL:  url,
							})
						}
					}
				}
			}
		}
	}

	return servers, nil
}

// 默认配置
func defaultConfig() Config {
	return Config{
		Token:              "",
		UploadServer:       "",
		SelectedServerName: "",
		ChunkSize:          3, // 3MB
		MaxConcurrent:      5,
		QuickUpload:        true,
		SkipUpload:         false,
		Language:           "", // 空表示尚未选择，首次运行时展示语言选择界面
		Model:              0,  // CLI默认过期模式
		MrID:               "0", // CLI默认根目录
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

// 文件信息
type FileInfo struct {
	Name    string
	Size    int64
	IsDir   bool
	ModTime time.Time
}

// Model TUI模型
type Model struct {
	// 基本状态
	state        State
	cliPath      string
	config       Config
	userInfo     UserInfo
	selectedFile string
	uploadTasks  []TaskStatus
	langIndex    int // 语言选择界面的当前选中索引

	// UI组件
	tokenInput  textinput.Model
	filePicker  filepicker.Model
	progress    progress.Model
	spinner     spinner.Model
	navigation  list.Model
	uploadTable table.Model
	viewport    viewport.Model

	// 文件浏览器状态
	currentDir    string
	files         []FileInfo
	selectedIndex int
	showHidden    bool

	// 设置界面状态
	settingsIndex    int
	settingsInputs   map[string]textinput.Model
	serverIndex      int            // 当前选中的服务器索引
	availableServers []ServerOption // 可用服务器列表

	// 界面状态
	err               error
	width             int
	height            int
	statusFiles       map[string]string // taskID -> statusFile path
	isLoading         bool
	isValidatingToken bool // 标记是否正在验证Token
	activeUploads     int
}

// 导航菜单项
type menuItem struct {
	title string
	desc  string
}

func (i menuItem) FilterValue() string { return i.title }
func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return i.desc }

// NewModel 创建新的TUI模型
func NewModel(cliPath string) Model {
	// 加载配置
	config := loadConfig()

	// 初始化token输入框
	tokenInput := textinput.New()
	tokenInput.Placeholder = i18n.T("auth.placeholder")
	tokenInput.Width = 50

	// 初始化状态
	// 如果尚未选择语言，先进入语言选择界面
	// 否则，如果没有有效Token，进入Token输入界面
	var initialState State
	if config.Language == "" {
		initialState = StateLanguageSelect
	} else if strings.TrimSpace(config.Token) != "" {
		initialState = StateInit
		tokenInput.Focus()
	} else {
		initialState = StateTokenInput
		tokenInput.Focus()
	}

	// 初始化文件选择器
	fp := filepicker.New()
	fp.AllowedTypes = []string{} // 允许所有文件类型
	fp.ShowHidden = false
	fp.DirAllowed = true
	// 设置为当前工作目录
	if currentDir, err := os.Getwd(); err == nil {
		fp.CurrentDirectory = currentDir
	} else {
		fp.CurrentDirectory, _ = os.UserHomeDir()
	}

	// 初始化进度条
	prog := progress.New(progress.WithDefaultGradient())

	// 初始化加载动画
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// 初始化导航菜单
	items := []list.Item{
		menuItem{title: i18n.T("menu.file_browser"), desc: i18n.T("menu.file_browser_desc")},
		menuItem{title: i18n.T("menu.settings"), desc: i18n.T("menu.settings_desc")},
		menuItem{title: i18n.T("menu.upload_manager"), desc: i18n.T("menu.upload_manager_desc")},
	}

	nav := list.New(items, list.NewDefaultDelegate(), 0, 0)
	nav.Title = i18n.T("menu.title")
	nav.SetShowStatusBar(false)
	nav.SetFilteringEnabled(false)
	nav.SetShowHelp(false)

	// 初始化上传任务表格
	columns := []table.Column{
		{Title: i18n.T("upload_list.col_filename"), Width: 25},
		{Title: i18n.T("upload_list.col_size"), Width: 10},
		{Title: i18n.T("upload_list.col_progress"), Width: 10},
		{Title: i18n.T("upload_list.col_speed"), Width: 10},
		{Title: i18n.T("upload_list.col_server"), Width: 12},
		{Title: i18n.T("upload_list.col_status"), Width: 10},
	}

	uploadTable := table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	// 初始化viewport
	vp := viewport.New(78, 20)

	// 获取当前目录
	currentDir, err := os.Getwd()
	if err != nil {
		currentDir, _ = os.UserHomeDir()
	}

	// 清理无效状态文件并加载有效任务
	var uploadTasks []TaskStatus
	var statusFiles map[string]string
	if config.Token != "" {
		validTasks, validStatusFiles, err := cleanupAndLoadTasks()
		if err == nil {
			uploadTasks = validTasks
			statusFiles = validStatusFiles
		} else {
			// 如果清理失败，使用空的任务列表
			uploadTasks = make([]TaskStatus, 0)
			statusFiles = make(map[string]string)
		}
	} else {
		uploadTasks = make([]TaskStatus, 0)
		statusFiles = make(map[string]string)
	}

	// 初始化设置输入框
	settingsInputs := make(map[string]textinput.Model)

	chunkSizeInput := textinput.New()
	chunkSizeInput.Placeholder = i18n.T("settings.chunk_placeholder")
	chunkSizeInput.Width = 20
	chunkSizeInput.SetValue(fmt.Sprintf("%d", config.ChunkSize))
	settingsInputs["chunk_size"] = chunkSizeInput

	concurrencyInput := textinput.New()
	concurrencyInput.Placeholder = i18n.T("settings.concurrency_placeholder")
	concurrencyInput.Width = 20
	concurrencyInput.SetValue(fmt.Sprintf("%d", config.MaxConcurrent))
	settingsInputs["concurrency"] = concurrencyInput

	// 默认设置焦点（在用户验证前假设非赞助用户）
	// 没有所有用户都可编辑的设置，所以先不设置焦点
	initialSettingsIndex := 0

	// 初始化服务器列表和索引（在没有token时为空列表）
	availableServers, _ := getAvailableServers("") // 空token，返回空列表
	serverIndex := 0
	// 如果有配置的服务器，根据配置的服务器URL或名称找到对应的索引
	if config.SelectedServerName != "" {
		for i, server := range availableServers {
			if server.URL == config.UploadServer || server.Name == config.SelectedServerName {
				serverIndex = i
				break
			}
		}
	}

	return Model{
		state:            initialState,
		cliPath:          cliPath,
		config:           config,
		tokenInput:       tokenInput,
		filePicker:       fp,
		progress:         prog,
		spinner:          s,
		navigation:       nav,
		uploadTable:      uploadTable,
		viewport:         vp,
		currentDir:       currentDir,
		files:            []FileInfo{},
		selectedIndex:    1,     // 跳过占位符，从第一个真实条目开始
		showHidden:       false, // 默认不显示隐藏文件
		settingsIndex:    initialSettingsIndex,
		settingsInputs:   settingsInputs,
		serverIndex:      serverIndex,
		availableServers: availableServers,
		uploadTasks:      uploadTasks,
		statusFiles:      statusFiles,
		langIndex:        0,
		isLoading:        strings.TrimSpace(config.Token) != "" && initialState == StateInit,
	}
}

// Init 初始化命令
func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd

	cmds = append(cmds, textinput.Blink)
	cmds = append(cmds, m.filePicker.Init())
	cmds = append(cmds, m.spinner.Tick)

	// 如果有有效token，开始获取用户信息
	if strings.TrimSpace(m.config.Token) != "" {
		cmds = append(cmds, m.fetchUserInfo())
	}

	// 加载文件列表
	cmds = append(cmds, m.loadFiles())

	// 为恢复的上传任务启动进度监控
	for _, task := range m.uploadTasks {
		if task.Status == "uploading" || task.Status == "pending" || task.Status == "starting" {
			cmds = append(cmds, m.startProgressTimer(task.ID))
		}
	}

	return tea.Batch(cmds...)
}

// Update 更新模型
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateComponentSizes()

	case UserInfoMsg:
		m.userInfo = msg.UserInfo
		m.isLoading = false
		if m.state == StateInit {
			m.state = StateMain
		}

		// 用户验证成功后，从API获取最新的服务器列表
		if updatedServers, err := getAvailableServers(m.config.Token); err == nil {
			m.availableServers = updatedServers

			// 如果没有配置的服务器，默认选择第一个可用服务器
			if m.config.SelectedServerName == "" && len(m.availableServers) > 0 {
				m.serverIndex = 0
				m.config.SelectedServerName = m.availableServers[0].Name
				m.config.UploadServer = m.availableServers[0].URL
			} else {
				// 根据配置查找对应的服务器索引
				found := false
				for i, server := range m.availableServers {
					if server.URL == m.config.UploadServer || server.Name == m.config.SelectedServerName {
						m.serverIndex = i
						// 更新配置以确保同步
						m.config.SelectedServerName = server.Name
						m.config.UploadServer = server.URL
						found = true
						break
					}
				}
				// 如果配置的服务器不在可用列表中，默认选择第一个
				if !found && len(m.availableServers) > 0 {
					m.serverIndex = 0
					m.config.SelectedServerName = m.availableServers[0].Name
					m.config.UploadServer = m.availableServers[0].URL
				}
			}
		}

		// 如果是赞助用户，设置设置界面的焦点和索引
		if m.userInfo.IsSponsored {
			// 设置chunk_size获得焦点
			if chunkSizeInput, exists := m.settingsInputs["chunk_size"]; exists {
				chunkSizeInput.Focus()
				m.settingsInputs["chunk_size"] = chunkSizeInput
				m.settingsIndex = 0 // 设置为第一个设置项
			}
		}

		return m, nil

	case UserInfoErrorMsg:
		// 如果是在token验证过程中失败，使用新的失败流程
		if m.isValidatingToken {
			m.isValidatingToken = false
			simplifiedError := simplifyErrorMessage(msg.Error)
			m.err = fmt.Errorf(simplifiedError)
			m.isLoading = false
			m.state = StateTokenValidationFailed
			return m, m.startReturnToTokenInputDelay()
		} else {
			m.err = fmt.Errorf("获取用户信息失败: %s", msg.Error)
			m.isLoading = false
			m.state = StateError
		}
		return m, nil

	case TokenValidatedMsg:
		m.config.Token = msg.Token
		m.userInfo = msg.UserInfo
		m.isLoading = false
		m.isValidatingToken = false // 清除验证标志
		m.state = StateMain

		// 用户验证成功后，从API获取最新的服务器列表
		if updatedServers, err := getAvailableServers(m.config.Token); err == nil {
			m.availableServers = updatedServers

			// 如果没有配置的服务器，默认选择第一个可用服务器
			if m.config.SelectedServerName == "" && len(m.availableServers) > 0 {
				m.serverIndex = 0
				m.config.SelectedServerName = m.availableServers[0].Name
				m.config.UploadServer = m.availableServers[0].URL
			} else {
				// 根据配置查找对应的服务器索引
				found := false
				for i, server := range m.availableServers {
					if server.URL == m.config.UploadServer || server.Name == m.config.SelectedServerName {
						m.serverIndex = i
						// 更新配置以确保同步
						m.config.SelectedServerName = server.Name
						m.config.UploadServer = server.URL
						found = true
						break
					}
				}
				// 如果配置的服务器不在可用列表中，默认选择第一个
				if !found && len(m.availableServers) > 0 {
					m.serverIndex = 0
					m.config.SelectedServerName = m.availableServers[0].Name
					m.config.UploadServer = m.availableServers[0].URL
				}
			}
		}

		// 如果是赞助用户，设置设置界面的焦点和索引
		if m.userInfo.IsSponsored {
			// 设置chunk_size获得焦点
			if chunkSizeInput, exists := m.settingsInputs["chunk_size"]; exists {
				chunkSizeInput.Focus()
				m.settingsInputs["chunk_size"] = chunkSizeInput
				m.settingsIndex = 0 // 设置为第一个设置项
			}
		}

		return m, nil

	case ReturnToTokenInputMsg:
		m.err = nil // 清除错误信息
		m.state = StateTokenInput
		m.tokenInput.Focus()
		return m, nil

	case FilesLoadedMsg:
		m.files = msg.Files
		return m, nil

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case UploadProgressMsg:
		return m.handleUploadProgress(msg)

	case UploadCompleteMsg:
		return m.handleUploadComplete(msg)

	case UploadErrorMsg:
		return m.handleUploadError(msg)

	case ProcessStartedMsg:
		return m.handleProcessStarted(msg)

	case CheckProgressTickMsg:
		return m.handleProgressTick(msg)
	}

	// 更新各组件
	return m.updateComponents(msg)
}

// updateComponentSizes 更新组件尺寸
func (m *Model) updateComponentSizes() {
	m.progress.Width = m.width - 4
	m.navigation.SetWidth(m.width)
	m.navigation.SetHeight(m.height - 7) // 为三行状态栏留空间
	m.uploadTable.SetWidth(m.width)
	m.viewport.Width = m.width
	m.viewport.Height = m.height - 7
}

// handleKeyPress 处理键盘输入
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	}

	switch m.state {
	case StateLanguageSelect:
		return m.handleLanguageSelect(msg)
	case StateTokenInput:
		return m.handleTokenInput(msg)
	case StateTokenValidationFailed:
		return m.handleTokenValidationFailed(msg)
	case StateMain:
		return m.handleMainView(msg)
	case StateSettings:
		return m.handleSettings(msg)
	case StateUploadList:
		return m.handleUploadList(msg)
	case StateError:
		return m.handleError(msg)
	}

	return m, nil
}

// handleLanguageSelect 处理语言选择界面输入
func (m Model) handleLanguageSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	langs := i18n.SupportedLanguages
	switch msg.String() {
	case "up", "k":
		if m.langIndex > 0 {
			m.langIndex--
		}
	case "down", "j":
		if m.langIndex < len(langs)-1 {
			m.langIndex++
		}
	case "enter":
		// 保存选择的语言
		selectedLang := langs[m.langIndex]
		i18n.SetLanguage(selectedLang)
		m.config.Language = selectedLang

		// 重建菜单和输入框以反映新语言
		items := []list.Item{
			menuItem{title: i18n.T("menu.file_browser"), desc: i18n.T("menu.file_browser_desc")},
			menuItem{title: i18n.T("menu.settings"), desc: i18n.T("menu.settings_desc")},
			menuItem{title: i18n.T("menu.upload_manager"), desc: i18n.T("menu.upload_manager_desc")},
		}
		m.navigation.SetItems(items)
		m.navigation.Title = i18n.T("menu.title")
		m.tokenInput.Placeholder = i18n.T("auth.placeholder")

		if chunkInput, ok := m.settingsInputs["chunk_size"]; ok {
			chunkInput.Placeholder = i18n.T("settings.chunk_placeholder")
			m.settingsInputs["chunk_size"] = chunkInput
		}
		if concInput, ok := m.settingsInputs["concurrency"]; ok {
			concInput.Placeholder = i18n.T("settings.concurrency_placeholder")
			m.settingsInputs["concurrency"] = concInput
		}

		// 更新表格列标题
		columns := []table.Column{
			{Title: i18n.T("upload_list.col_filename"), Width: 25},
			{Title: i18n.T("upload_list.col_size"), Width: 10},
			{Title: i18n.T("upload_list.col_progress"), Width: 10},
			{Title: i18n.T("upload_list.col_speed"), Width: 10},
			{Title: i18n.T("upload_list.col_server"), Width: 12},
			{Title: i18n.T("upload_list.col_status"), Width: 10},
		}
		m.uploadTable.SetColumns(columns)

		// 保存配置
		saveConfig(m.config)

		// 进入Token输入界面
		m.state = StateTokenInput
		m.tokenInput.Focus()
	}
	return m, nil
}

// handleTokenInput 处理token输入
func (m Model) handleTokenInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.tokenInput.Value() != "" {
			// 验证token，但不保存到配置直到验证成功
			tempToken := m.tokenInput.Value()
			m.isLoading = true
			m.isValidatingToken = true // 设置验证标志
			return m, m.validateAndSaveToken(tempToken)
		}
	}

	var cmd tea.Cmd
	m.tokenInput, cmd = m.tokenInput.Update(msg)
	return m, cmd
}

// handleTokenValidationFailed 处理Token验证失败状态的输入
func (m Model) handleTokenValidationFailed(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 在验证失败等待期间，任何按键都会立即返回Token输入界面
	switch msg.String() {
	case "enter", "esc", " ":
		m.err = nil
		m.state = StateTokenInput
		m.tokenInput.Focus()
		return m, nil
	}
	return m, nil
}

// handleMainView 处理主界面输入
func (m Model) handleMainView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		m.state = StateSettings
		return m, nil
	case "enter":
		return m.handleFileSelection()
	case "up":
		if m.selectedIndex > 1 { // 跳过第一个占位符条目
			m.selectedIndex--
		}
		return m, nil
	case "down":
		if m.selectedIndex < len(m.files)-1 {
			m.selectedIndex++
		} else if m.selectedIndex == 0 { // 如果在占位符上，移动到第一个真实条目
			m.selectedIndex = 1
		}
		return m, nil
	case "left", "right":
		// 返回上级目录 (仅在非根目录时有效)
		parentDir := filepath.Dir(m.currentDir)
		if parentDir != m.currentDir { // 确保不是根目录
			return m.navigateToParent()
		}
		return m, nil
	case "t":
		// 切换显示隐藏文件
		m.showHidden = !m.showHidden
		m.selectedIndex = 1 // 重置选择索引，跳过占位符
		return m, m.loadFiles()
	}

	return m, nil
}

// handleSettings 处理设置界面输入
func (m Model) handleSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 根据用户类型确定可用设置（需与 renderSettings 保持一致）
	var settingsKeys []string
	if m.userInfo.IsSponsored {
		settingsKeys = []string{"chunk_size", "concurrency", "server", "quick_upload", "language"}
	} else {
		settingsKeys = []string{"language"}
	}

	switch msg.String() {
	case "tab":
		m.state = StateUploadList
		m.updateUploadTable()
		return m, nil
	case "esc":
		m.state = StateMain
		return m, nil
	case "up":
		if m.settingsIndex > 0 {
			currentKey := settingsKeys[m.settingsIndex]
			if input, exists := m.settingsInputs[currentKey]; exists {
				input.Blur()
				m.settingsInputs[currentKey] = input
			}
			m.settingsIndex--
			newKey := settingsKeys[m.settingsIndex]
			if newInput, exists := m.settingsInputs[newKey]; exists {
				newInput.Focus()
				m.settingsInputs[newKey] = newInput
			}
		}
		return m, nil
	case "down":
		if m.settingsIndex < len(settingsKeys)-1 {
			currentKey := settingsKeys[m.settingsIndex]
			if input, exists := m.settingsInputs[currentKey]; exists {
				input.Blur()
				m.settingsInputs[currentKey] = input
			}
			m.settingsIndex++
			newKey := settingsKeys[m.settingsIndex]
			if newInput, exists := m.settingsInputs[newKey]; exists {
				newInput.Focus()
				m.settingsInputs[newKey] = newInput
			}
		}
		return m, nil
	case "left", "right":
		if m.settingsIndex < len(settingsKeys) {
			currentKey := settingsKeys[m.settingsIndex]
			switch currentKey {
			case "server":
				if m.userInfo.IsSponsored {
					if msg.String() == "left" {
						if m.serverIndex > 0 {
							m.serverIndex--
						} else {
							m.serverIndex = len(m.availableServers) - 1
						}
					} else {
						if m.serverIndex < len(m.availableServers)-1 {
							m.serverIndex++
						} else {
							m.serverIndex = 0
						}
					}
				}
			case "quick_upload":
				if m.userInfo.IsSponsored {
					m.config.QuickUpload = !m.config.QuickUpload
				}
			case "language":
				langs := i18n.SupportedLanguages
				curIdx := 0
				for i, l := range langs {
					if l == i18n.GetLanguage() {
						curIdx = i
						break
					}
				}
				if msg.String() == "left" {
					if curIdx > 0 {
						curIdx--
					} else {
						curIdx = len(langs) - 1
					}
				} else {
					if curIdx < len(langs)-1 {
						curIdx++
					} else {
						curIdx = 0
					}
				}
				m = m.applyLanguage(langs[curIdx])
			}
		}
		return m, nil
	case " ":
		if m.settingsIndex < len(settingsKeys) {
			currentKey := settingsKeys[m.settingsIndex]
			if currentKey == "quick_upload" && m.userInfo.IsSponsored {
				m.config.QuickUpload = !m.config.QuickUpload
			}
		}
		return m, nil
	case "enter":
		return m.saveSettings()
	}

	// 更新当前聚焦的输入框
	if m.settingsIndex < len(settingsKeys) {
		currentKey := settingsKeys[m.settingsIndex]
		if input, exists := m.settingsInputs[currentKey]; exists {
			newInput, cmd := input.Update(msg)
			m.settingsInputs[currentKey] = newInput
			return m, cmd
		}
	}

	return m, nil
}

// applyLanguage 切换界面语言并重建需要语言的 UI 组件
func (m Model) applyLanguage(lang string) Model {
	i18n.SetLanguage(lang)
	m.config.Language = lang

	items := []list.Item{
		menuItem{title: i18n.T("menu.file_browser"), desc: i18n.T("menu.file_browser_desc")},
		menuItem{title: i18n.T("menu.settings"), desc: i18n.T("menu.settings_desc")},
		menuItem{title: i18n.T("menu.upload_manager"), desc: i18n.T("menu.upload_manager_desc")},
	}
	m.navigation.SetItems(items)
	m.navigation.Title = i18n.T("menu.title")
	m.tokenInput.Placeholder = i18n.T("auth.placeholder")

	if chunkInput, ok := m.settingsInputs["chunk_size"]; ok {
		chunkInput.Placeholder = i18n.T("settings.chunk_placeholder")
		m.settingsInputs["chunk_size"] = chunkInput
	}
	if concInput, ok := m.settingsInputs["concurrency"]; ok {
		concInput.Placeholder = i18n.T("settings.concurrency_placeholder")
		m.settingsInputs["concurrency"] = concInput
	}

	columns := []table.Column{
		{Title: i18n.T("upload_list.col_filename"), Width: 25},
		{Title: i18n.T("upload_list.col_size"), Width: 10},
		{Title: i18n.T("upload_list.col_progress"), Width: 10},
		{Title: i18n.T("upload_list.col_speed"), Width: 10},
		{Title: i18n.T("upload_list.col_server"), Width: 12},
		{Title: i18n.T("upload_list.col_status"), Width: 10},
	}
	m.uploadTable.SetColumns(columns)

	return m
}

// handleUploadList 处理上传列表输入
func (m Model) handleUploadList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		m.state = StateMain
		return m, nil
	case "esc":
		m.state = StateMain
		return m, nil
	case "d":
		// 删除选中的上传任务
		return m.cancelSelectedUpload()
	case "t":
		// 清除所有已完成任务
		return m.clearCompletedTasks()
	case "y":
		// 清除所有任务
		return m.clearAllTasks()
	}

	var cmd tea.Cmd
	m.uploadTable, cmd = m.uploadTable.Update(msg)
	return m, cmd
}

// handleError 处理错误界面输入
func (m Model) handleError(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc":
		// 先检查错误信息再清除
		isTokenError := m.err != nil && strings.Contains(strings.ToLower(m.err.Error()), "token")
		m.err = nil

		// 如果是Token相关错误，返回Token输入界面
		if isTokenError {
			m.state = StateTokenInput
			m.tokenInput.Focus()
		} else {
			// 如果用户还没有有效的Token，也返回Token输入界面
			if m.config.Token == "" || m.userInfo.Username == "" {
				m.state = StateTokenInput
				m.tokenInput.Focus()
			} else {
				m.state = StateMain
			}
		}
		return m, nil
	}
	return m, nil
}

// getFileUploadStatus 获取文件的上传状态
func (m Model) getFileUploadStatus(filePath string) (string, bool) {
	// 规范化文件路径以便比较
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}

	for _, task := range m.uploadTasks {
		taskAbsPath, err := filepath.Abs(task.FilePath)
		if err != nil {
			taskAbsPath = task.FilePath
		}

		if taskAbsPath == absPath {
			return task.Status, true
		}
	}
	return "", false
}

// isFileUploadAllowed 检查文件是否允许上传
func (m Model) isFileUploadAllowed(filePath string) (bool, string) {
	status, exists := m.getFileUploadStatus(filePath)
	if !exists {
		return true, ""
	}

	// 只有上传失败的文件才允许重新上传
	if status == "failed" {
		return true, ""
	}

	// 其他状态都不允许重复上传
	switch status {
	case "starting", "pending", "uploading":
		return false, i18n.T("upload.in_progress")
	case "completed":
		return false, i18n.T("upload.completed")
	default:
		return false, i18n.T("upload.in_list")
	}
}

// handleFileSelection 处理文件选择
func (m Model) handleFileSelection() (tea.Model, tea.Cmd) {
	if len(m.files) == 0 || m.selectedIndex >= len(m.files) {
		return m, nil
	}

	selectedFile := m.files[m.selectedIndex]

	if selectedFile.IsDir {
		if selectedFile.Name == ".." {
			// 返回上级目录
			return m.navigateToParent()
		} else {
			// 进入目录
			newDir := filepath.Join(m.currentDir, selectedFile.Name)
			m.currentDir = newDir
			m.selectedIndex = 1
			return m, m.loadFiles()
		}
	} else {
		// 选择文件进行上传
		filePath := filepath.Join(m.currentDir, selectedFile.Name)

		// 检查文件是否允许上传
		allowed, reason := m.isFileUploadAllowed(filePath)
		if !allowed {
			m.err = fmt.Errorf("%s", reason)
			m.state = StateError
			return m, nil
		}

		// 验证文件大小限制 (50GB)
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			m.err = fmt.Errorf("无法获取文件信息: %v", err)
			m.state = StateError
			return m, nil
		}

		const maxFileSize = 50 * 1024 * 1024 * 1024 // 50GB
		if fileInfo.Size() > maxFileSize {
			m.err = fmt.Errorf("文件大小超出限制，最大支持50GB，当前文件: %.2fGB",
				float64(fileInfo.Size())/(1024*1024*1024))
			m.state = StateError
			return m, nil
		}

		return m.startFileUpload(filePath)
	}
}

// navigateToParent 返回上级目录
func (m Model) navigateToParent() (tea.Model, tea.Cmd) {
	parentDir := filepath.Dir(m.currentDir)
	if parentDir != m.currentDir { // 确保不是根目录
		m.currentDir = parentDir
		m.selectedIndex = 1
		return m, m.loadFiles()
	}
	return m, nil
}

// cancelSelectedUpload 取消选中的上传任务
func (m Model) cancelSelectedUpload() (tea.Model, tea.Cmd) {
	// 获取当前选中的任务索引
	selectedRow := m.uploadTable.Cursor()

	// 检查是否有任务可删除
	if selectedRow < 0 || selectedRow >= len(m.uploadTasks) {
		return m, nil
	}

	task := m.uploadTasks[selectedRow]

	// 如果任务正在运行，先尝试终止进程
	if task.Status == "uploading" || task.Status == "pending" || task.Status == "starting" {
		if task.ProcessID > 0 {
			// 尝试终止CLI进程
			if process, err := os.FindProcess(task.ProcessID); err == nil {
				// 先尝试优雅终止（SIGTERM）
				process.Signal(syscall.SIGTERM)

				// 等待短暂时间，然后强制终止
				go func() {
					time.Sleep(2 * time.Second)
					if isProcessRunning(task.ProcessID) {
						process.Kill() // 强制终止进程（SIGKILL）
					}
				}()
			}
		}

		// 更新活跃上传计数
		m.activeUploads--
		if m.activeUploads < 0 {
			m.activeUploads = 0
		}
	}

	// 删除状态文件
	if statusFile, exists := m.statusFiles[task.ID]; exists {
		os.Remove(statusFile)
		os.Remove(statusFile + ".log") // 同时删除日志文件
		delete(m.statusFiles, task.ID)
	}

	// 从任务列表中移除
	if selectedRow < len(m.uploadTasks) {
		m.uploadTasks = append(m.uploadTasks[:selectedRow], m.uploadTasks[selectedRow+1:]...)
	}

	// 更新表格选中位置
	if len(m.uploadTasks) > 0 && selectedRow >= len(m.uploadTasks) {
		m.uploadTable.SetCursor(len(m.uploadTasks) - 1)
	} else if len(m.uploadTasks) == 0 {
		m.uploadTable.SetCursor(0)
	}

	// 更新上传表格显示
	m.updateUploadTable()

	return m, nil
}

// clearCompletedTasks 清除所有已完成任务
func (m Model) clearCompletedTasks() (tea.Model, tea.Cmd) {
	var activeTasks []TaskStatus

	// 遍历任务，只保留未完成的任务
	for _, task := range m.uploadTasks {
		if task.Status != "completed" && task.Status != "failed" {
			// 保留进行中或等待中的任务
			activeTasks = append(activeTasks, task)
		} else {
			// 删除已完成/失败任务的状态文件
			if statusFile, exists := m.statusFiles[task.ID]; exists {
				os.Remove(statusFile)
				os.Remove(statusFile + ".log")
				delete(m.statusFiles, task.ID)
			}
		}
	}

	// 更新任务列表
	m.uploadTasks = activeTasks

	// 重置表格选中位置
	if len(m.uploadTasks) > 0 {
		m.uploadTable.SetCursor(0)
	}

	// 更新上传表格显示
	m.updateUploadTable()

	return m, nil
}

// clearAllTasks 清除所有任务
func (m Model) clearAllTasks() (tea.Model, tea.Cmd) {
	// 终止所有运行中的任务
	for _, task := range m.uploadTasks {
		if task.Status == "uploading" || task.Status == "pending" || task.Status == "starting" {
			if task.ProcessID > 0 {
				// 尝试终止CLI进程
				if process, err := os.FindProcess(task.ProcessID); err == nil {
					// 先尝试优雅终止（SIGTERM）
					process.Signal(syscall.SIGTERM)

					// 等待短暂时间，然后强制终止
					go func(pid int) {
						time.Sleep(2 * time.Second)
						if isProcessRunning(pid) {
							if proc, err := os.FindProcess(pid); err == nil {
								proc.Kill() // 强制终止进程（SIGKILL）
							}
						}
					}(task.ProcessID)
				}
			}
		}

		// 删除状态文件
		if statusFile, exists := m.statusFiles[task.ID]; exists {
			os.Remove(statusFile)
			os.Remove(statusFile + ".log")
			delete(m.statusFiles, task.ID)
		}
	}

	// 清空所有任务
	m.uploadTasks = []TaskStatus{}
	m.statusFiles = make(map[string]string)
	m.activeUploads = 0

	// 重置表格
	m.uploadTable.SetCursor(0)

	// 更新上传表格显示
	m.updateUploadTable()

	return m, nil
}

// handleUploadProgress 处理上传进度
func (m Model) handleUploadProgress(msg UploadProgressMsg) (tea.Model, tea.Cmd) {
	for i, task := range m.uploadTasks {
		if task.ID == msg.TaskID {
			m.uploadTasks[i].Progress = msg.Progress
			m.uploadTasks[i].UploadSpeed = msg.Speed
			if msg.Progress > 0 {
				m.uploadTasks[i].Status = "uploading"
			}
			m.uploadTasks[i].UpdatedAt = time.Now()
			break
		}
	}
	m.updateUploadTable()
	// 不需要继续调用 checkProgress，因为定时器会处理
	return m, nil
}

// handleUploadComplete 处理上传完成
func (m Model) handleUploadComplete(msg UploadCompleteMsg) (tea.Model, tea.Cmd) {
	for i, task := range m.uploadTasks {
		if task.ID == msg.TaskID {
			m.uploadTasks[i].Status = "completed"
			m.uploadTasks[i].Progress = 100.0 // CLI使用0-100的百分比
			m.uploadTasks[i].DownloadURL = msg.DownloadURL
			m.uploadTasks[i].UpdatedAt = time.Now()
			m.activeUploads--
			break
		}
	}
	m.updateUploadTable()
	return m, nil
}

// handleUploadError 处理上传错误
func (m Model) handleUploadError(msg UploadErrorMsg) (tea.Model, tea.Cmd) {
	// 如果有TaskID，更新对应任务状态
	if msg.TaskID != "" {
		for i, task := range m.uploadTasks {
			if task.ID == msg.TaskID {
				m.uploadTasks[i].Status = "failed"
				m.uploadTasks[i].ErrorMsg = msg.Error
				m.uploadTasks[i].UpdatedAt = time.Now()
				break
			}
		}
		m.updateUploadTable()
	} else {
		m.err = fmt.Errorf("上传失败: %s", msg.Error)
	}
	m.activeUploads--
	return m, nil
}

// handleProcessStarted 处理进程启动
func (m Model) handleProcessStarted(msg ProcessStartedMsg) (tea.Model, tea.Cmd) {
	// 更新任务状态，保存进程ID
	for i, task := range m.uploadTasks {
		if task.ID == msg.TaskID {
			m.uploadTasks[i].ProcessID = msg.ProcessID
			m.uploadTasks[i].Status = "pending"
			m.uploadTasks[i].UpdatedAt = time.Now()
			break
		}
	}
	m.updateUploadTable()

	// 启动定时器进行进度监控
	return m, m.startProgressTimer(msg.TaskID)
}

// handleProgressTick 处理进度检查定时器
func (m Model) handleProgressTick(msg CheckProgressTickMsg) (tea.Model, tea.Cmd) {
	// 检查任务是否还在运行
	taskExists := false
	var currentTask *TaskStatus
	for i, task := range m.uploadTasks {
		if task.ID == msg.TaskID {
			currentTask = &m.uploadTasks[i]
			if task.Status == "starting" || task.Status == "pending" || task.Status == "uploading" {
				taskExists = true
			}
			break
		}
	}

	if !taskExists || currentTask == nil {
		// 任务不存在或已完成，停止监控
		return m, nil
	}

	// 检查进程是否还在运行
	if currentTask.ProcessID > 0 && !isProcessRunning(currentTask.ProcessID) {
		// 进程已结束，进行最后一次状态检查
		return m, m.checkProgress(msg.TaskID)
	}

	// 检查进度并继续定时器
	var cmds []tea.Cmd
	cmds = append(cmds, m.checkProgress(msg.TaskID))
	cmds = append(cmds, m.startProgressTimer(msg.TaskID))

	return m, tea.Batch(cmds...)
}

// updateComponents 更新组件
func (m Model) updateComponents(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// 更新文件选择器
	if m.state == StateMain {
		m.filePicker, cmd = m.filePicker.Update(msg)
		cmds = append(cmds, cmd)

		// 检查文件选择
		if didSelect, path := m.filePicker.DidSelectFile(msg); didSelect {
			return m.startFileUpload(path)
		}
	}

	return m, tea.Batch(cmds...)
}

// startFileUpload 开始文件上传
func (m Model) startFileUpload(filePath string) (tea.Model, tea.Cmd) {
	m.selectedFile = filePath
	m.activeUploads++

	// 生成任务ID（包含纳秒确保唯一性）
	taskID := fmt.Sprintf("task_%d", time.Now().UnixNano())
	homeDir, _ := os.UserHomeDir()
	statusDir := filepath.Join(homeDir, ".tmplink", "tasks")
	os.MkdirAll(statusDir, 0755)
	statusFile := filepath.Join(statusDir, taskID+".json")
	m.statusFiles[taskID] = statusFile

	// 立即创建任务状态并添加到任务列表
	fileInfo, _ := os.Stat(filePath)

	// 获取当前选中的服务器名称
	selectedServerName := "未知"
	if m.serverIndex < len(m.availableServers) && len(m.availableServers) > 0 {
		selectedServerName = m.availableServers[m.serverIndex].Name
	}

	task := TaskStatus{
		ID:         taskID,
		Status:     "starting",
		FilePath:   filePath,
		FileName:   filepath.Base(filePath),
		FileSize:   fileInfo.Size(),
		Progress:   0.0,
		ServerName: selectedServerName, // 设置服务器名称
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 添加到任务列表
	m.uploadTasks = append(m.uploadTasks, task)

	// 更新上传表格
	m.updateUploadTable()

	return m, m.startUpload(filePath, taskID, statusFile)
}

// updateUploadTable 更新上传任务表格
func (m *Model) updateUploadTable() {
	var rows []table.Row

	for _, task := range m.uploadTasks {
		// 格式化文件大小
		sizeStr := ""
		if task.FileSize < 1024 {
			sizeStr = fmt.Sprintf("%dB", task.FileSize)
		} else if task.FileSize < 1024*1024 {
			sizeStr = fmt.Sprintf("%.1fKB", float64(task.FileSize)/1024)
		} else if task.FileSize < 1024*1024*1024 {
			sizeStr = fmt.Sprintf("%.1fMB", float64(task.FileSize)/(1024*1024))
		} else {
			sizeStr = fmt.Sprintf("%.1fGB", float64(task.FileSize)/(1024*1024*1024))
		}

		// 格式化进度 (CLI返回的是0-100的百分比，直接使用)
		progressStr := fmt.Sprintf("%.1f%%", task.Progress)

		// 状态翻译
		statusStr := task.Status
		switch task.Status {
		case "starting":
			statusStr = i18n.T("task.starting")
		case "pending":
			statusStr = i18n.T("task.pending")
		case "uploading":
			statusStr = i18n.T("task.uploading")
		case "completed":
			statusStr = i18n.T("task.completed")
		case "failed":
			statusStr = i18n.T("task.failed")
		}

		// 速度显示（上传中和已完成都显示最终速度）
		speedStr := ""
		if task.UploadSpeed > 0 && (task.Status == "uploading" || task.Status == "completed") {
			if task.UploadSpeed >= 1024 {
				speedStr = fmt.Sprintf("%.1fMB/s", task.UploadSpeed/1024)
			} else {
				speedStr = fmt.Sprintf("%.1fKB/s", task.UploadSpeed)
			}
		}

		// 服务器名称显示
		serverStr := task.ServerName
		if serverStr == "" {
			serverStr = i18n.T("task.unknown_server")
		}

		row := table.Row{
			task.FileName,
			sizeStr,
			progressStr,
			speedStr,
			serverStr,
			statusStr,
		}
		rows = append(rows, row)
	}

	m.uploadTable.SetRows(rows)
}

// View 渲染视图
func (m Model) View() string {
	// 在没有有效Token或用户信息的状态下，不显示状态栏
	shouldHideStatusBar := m.state == StateLanguageSelect ||
		m.state == StateTokenInput ||
		m.state == StateTokenValidationFailed ||
		m.state == StateInit ||
		m.userInfo.Username == ""

	if shouldHideStatusBar {
		return m.renderContent()
	}

	// 双区域布局：顶部状态栏 + 功能区域
	statusBar := m.renderStatusBar()
	content := m.renderContent()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		statusBar,
		content,
	)
}

// renderTokenInput 渲染token输入界面
func (m Model) renderTokenInput() string {
	// 计算居中位置
	windowWidth := 80
	if m.width > 80 {
		windowWidth = m.width - 20 // 留出边距
		if windowWidth > 100 {
			windowWidth = 100 // 最大宽度限制
		}
	}

	// 创建窗口边框样式
	windowStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(2, 4).
		Width(windowWidth - 10)

	// 标题样式
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Align(lipgloss.Center).
		Width(windowWidth - 18)

	// 子标题样式
	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("246")).
		Italic(true).
		Align(lipgloss.Center).
		Width(windowWidth - 18)

	// 说明文字样式
	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250")).
		Align(lipgloss.Left).
		Width(windowWidth - 18).
		MarginTop(1).
		MarginBottom(1)

	// 输入框容器样式
	inputContainerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		MarginTop(1).
		MarginBottom(2).
		Width(windowWidth - 22)

	// 帮助文字样式
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Center).
		Width(windowWidth - 18)

	var content strings.Builder

	// 标题区域
	content.WriteString(titleStyle.Render(i18n.T("app.title")))
	content.WriteString("\n")
	content.WriteString(subtitleStyle.Render(i18n.T("auth.subtitle")))
	content.WriteString("\n\n")

	// 显示错误信息（如果有）
	if m.err != nil {
		errorBoxStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("196")).
			Background(lipgloss.Color("52")).
			Foreground(lipgloss.Color("255")).
			Padding(1, 2).
			Width(windowWidth - 22).
			MarginBottom(2)

		content.WriteString(errorBoxStyle.Render(i18n.Tf("auth.error_box", m.err.Error())))
		content.WriteString("\n\n")
	}

	if m.isLoading {
		// 加载状态
		loadingStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")).
			Bold(true).
			Align(lipgloss.Center).
			Width(windowWidth - 18)

		content.WriteString(loadingStyle.Render(i18n.Tf("auth.validating", m.spinner.View())))
		content.WriteString("\n\n")
	} else {
		content.WriteString(instructionStyle.Render(i18n.T("auth.instructions")))
		content.WriteString("\n")

		// 输入框标签
		labelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")).
			Bold(true).
			MarginBottom(1)

		content.WriteString(labelStyle.Render(i18n.T("auth.token_label")))
		content.WriteString("\n")
	}

	// 输入框
	if !m.isLoading {
		// 设置输入框宽度
		m.tokenInput.Width = windowWidth - 26
		inputContainer := inputContainerStyle.Render(m.tokenInput.View())
		content.WriteString(inputContainer)
	}

	// 帮助信息
	if m.isLoading {
		content.WriteString(helpStyle.Render(i18n.T("auth.wait")))
	} else {
		content.WriteString(helpStyle.Render(i18n.T("auth.help")))
	}

	// 将内容放入窗口边框
	window := windowStyle.Render(content.String())

	// 使用 lipgloss.Place 实现完全居中
	availableHeight := m.height
	availableWidth := m.width
	if availableHeight <= 0 {
		availableHeight = 30
	}
	if availableWidth <= 0 {
		availableWidth = 80
	}

	// 使用 lipgloss.Place 进行居中定位
	centered := lipgloss.Place(
		availableWidth,
		availableHeight,
		lipgloss.Center,
		lipgloss.Center,
		window,
	)

	return centered
}

// renderStatusBar 渲染顶部状态栏（三行布局）
func (m Model) renderStatusBar() string {
	if m.isLoading {
		return statusBarStyle.Render(i18n.Tf("loading.token", m.spinner.View()))
	}

	// 计算可用宽度
	statusWidth := m.width
	if statusWidth <= 0 {
		statusWidth = 80
	}

	var lines []string

	// 第一行：用户信息和认证状态
	var line1 string
	if m.userInfo.Username != "" {
		userText := i18n.Tf("status.user", m.userInfo.Username)
		if m.userInfo.IsSponsored {
			userText += i18n.T("status.sponsor")
		} else {
			userText += i18n.T("status.regular")
		}
		line1 = userText
	} else {
		line1 = i18n.T("status.not_logged_in")
	}
	lines = append(lines, statusBarStyle.Width(statusWidth).Render(line1))

	// 第二行：存储信息
	var line2 string
	if m.userInfo.TotalSpace > 0 {
		usedGB := float64(m.userInfo.UsedSpace) / (1024 * 1024 * 1024)
		totalGB := float64(m.userInfo.TotalSpace) / (1024 * 1024 * 1024)
		usagePercent := float64(m.userInfo.UsedSpace) / float64(m.userInfo.TotalSpace) * 100
		line2 = i18n.Tf("status.storage", usedGB, totalGB, usagePercent)
	} else {
		line2 = i18n.T("status.no_storage")
	}
	if m.activeUploads > 0 {
		uploadText := i18n.Tf("status.uploading", m.activeUploads)
		totalSpeed := 0.0
		for _, task := range m.uploadTasks {
			if task.Status == "uploading" {
				totalSpeed += task.UploadSpeed
			}
		}
		if totalSpeed > 0 {
			if totalSpeed >= 1024 {
				uploadText += i18n.Tf("status.speed_mb", totalSpeed/1024)
			} else {
				uploadText += i18n.Tf("status.speed_kb", totalSpeed)
			}
		}
		line2 += uploadText
	}
	lines = append(lines, statusBarStyle.Width(statusWidth).Render(line2))

	// 第三行：操作提示
	var line3 string
	switch m.state {
	case StateMain:
		parentDir := filepath.Dir(m.currentDir)
		enterAction := i18n.T("nav.enter")
		if len(m.files) > 0 && m.selectedIndex < len(m.files) {
			selectedFile := m.files[m.selectedIndex]
			if !selectedFile.IsDir {
				enterAction = i18n.T("nav.upload")
			}
		}
		if parentDir != m.currentDir {
			line3 = i18n.Tf("nav.keys_with_parent", enterAction)
		} else {
			line3 = i18n.Tf("nav.keys_no_parent", enterAction)
		}
	case StateSettings:
		line3 = i18n.T("settings.keys")
	case StateUploadList:
		line3 = i18n.T("upload_list.keys")
	case StateError:
		line3 = i18n.T("error.keys")
	default:
		line3 = i18n.T("default.keys")
	}

	// 确保操作提示不超过宽度（按 Unicode 字符截断）
	if len([]rune(line3)) > statusWidth {
		runes := []rune(line3)
		if statusWidth > 3 {
			line3 = string(runes[:statusWidth-3]) + "..."
		}
	}
	lines = append(lines, statusBarStyle.Width(statusWidth).Render(line3))

	return strings.Join(lines, "\n")
}

// cleanupAndLoadTasks 清理无效状态文件并加载有效任务
func cleanupAndLoadTasks() ([]TaskStatus, map[string]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, err
	}

	tasksDir := filepath.Join(homeDir, ".tmplink", "tasks")
	statusFiles := make(map[string]string)
	var validTasks []TaskStatus

	// 检查任务目录是否存在
	if _, err := os.Stat(tasksDir); os.IsNotExist(err) {
		return validTasks, statusFiles, nil
	}

	// 读取所有状态文件
	files, err := os.ReadDir(tasksDir)
	if err != nil {
		return nil, nil, err
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		statusFile := filepath.Join(tasksDir, file.Name())

		// 读取状态文件
		data, err := os.ReadFile(statusFile)
		if err != nil {
			// 无法读取的文件直接删除
			os.Remove(statusFile)
			continue
		}

		var task TaskStatus
		if err := json.Unmarshal(data, &task); err != nil {
			// 无法解析的文件直接删除
			os.Remove(statusFile)
			continue
		}

		// 检查任务状态
		shouldKeep := false

		if task.Status == "completed" || task.Status == "failed" {
			// 已完成或失败的任务保留并加载到UI中
			shouldKeep = true
			validTasks = append(validTasks, task)
			statusFiles[task.ID] = statusFile
		} else if task.ProcessID > 0 {
			// 检查进程是否还在运行
			if isProcessRunning(task.ProcessID) {
				// 进程仍在运行，加入监控列表
				shouldKeep = true
				validTasks = append(validTasks, task)
				statusFiles[task.ID] = statusFile
			}
		}

		if !shouldKeep {
			// 删除无效的状态文件
			os.Remove(statusFile)
			// 同时删除对应的日志文件
			os.Remove(statusFile + ".log")
		}
	}

	return validTasks, statusFiles, nil
}

// isProcessRunning 检查进程是否正在运行
func isProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}

	// 在Unix系统上，发送信号0可以检查进程是否存在
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// 发送信号0检查进程是否存在（在Unix系统上）
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// renderContent 渲染主要内容区域
func (m Model) renderContent() string {
	switch m.state {
	case StateInit:
		return m.renderLoading()
	case StateLanguageSelect:
		return m.renderLanguageSelect()
	case StateTokenInput:
		return m.renderTokenInput()
	case StateTokenValidationFailed:
		return m.renderTokenValidationFailed()
	case StateMain:
		return m.renderMainView()
	case StateSettings:
		return m.renderSettings()
	case StateUploadList:
		return m.renderUploadList()
	case StateError:
		return m.renderError()
	default:
		return i18n.T("unknown_state")
	}
}

// renderLoading 渲染加载界面
func (m Model) renderLoading() string {
	return i18n.Tf("loading.token", m.spinner.View())
}

// renderLanguageSelect 渲染语言选择界面（首次运行）
func (m Model) renderLanguageSelect() string {
	windowWidth := 60
	if m.width > 0 && m.width < windowWidth+10 {
		windowWidth = m.width - 10
	}

	windowStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(2, 4).
		Width(windowWidth - 10)

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Align(lipgloss.Center).
		Width(windowWidth - 18)

	var content strings.Builder

	// 显示三种语言的标题，方便用户识别
	content.WriteString(titleStyle.Render("选择语言 / Select Language / 言語を選択"))
	content.WriteString("\n\n")

	for i, lang := range i18n.SupportedLanguages {
		name := i18n.LanguageNames[lang]
		if i == m.langIndex {
			line := lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				Bold(true).
				Render("> " + name)
			content.WriteString(line)
		} else {
			content.WriteString("  " + name)
		}
		content.WriteString("\n")
	}

	content.WriteString("\n")

	// 三语提示
	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Center).
		Width(windowWidth - 18)
	content.WriteString(hintStyle.Render("↑↓:选择/Select/選択  Enter:确认/Confirm/確認"))

	window := windowStyle.Render(content.String())

	availableHeight := m.height
	availableWidth := m.width
	if availableHeight <= 0 {
		availableHeight = 30
	}
	if availableWidth <= 0 {
		availableWidth = 80
	}

	return lipgloss.Place(availableWidth, availableHeight, lipgloss.Center, lipgloss.Center, window)
}

// renderTokenValidationFailed 渲染Token验证失败界面
func (m Model) renderTokenValidationFailed() string {
	// 计算居中位置
	windowWidth := 80
	if m.width > 80 {
		windowWidth = m.width - 20
		if windowWidth > 100 {
			windowWidth = 100
		}
	}

	// 创建窗口边框样式
	windowStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(2, 4).
		Width(windowWidth - 10)

	// 标题样式
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true).
		Align(lipgloss.Center).
		Width(windowWidth - 18)

	// 子标题样式
	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("246")).
		Italic(true).
		Align(lipgloss.Center).
		Width(windowWidth - 18)

	// 帮助文字样式
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Center).
		Width(windowWidth - 18)

	var content strings.Builder

	// 标题区域
	content.WriteString(titleStyle.Render(i18n.T("auth.failed_title")))
	content.WriteString("\n")
	content.WriteString(subtitleStyle.Render(i18n.T("auth.failed_subtitle")))
	content.WriteString("\n\n")

	// 显示错误信息
	if m.err != nil {
		errorBoxStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("196")).
			Background(lipgloss.Color("52")).
			Foreground(lipgloss.Color("255")).
			Padding(1, 2).
			Width(windowWidth - 22).
			MarginBottom(2)

		content.WriteString(errorBoxStyle.Render(i18n.Tf("auth.error_details", m.err.Error())))
		content.WriteString("\n\n")
	}

	// 自动返回提示
	loadingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("33")).
		Bold(true).
		Align(lipgloss.Center).
		Width(windowWidth - 18)

	content.WriteString(loadingStyle.Render(i18n.Tf("auth.auto_return", m.spinner.View())))
	content.WriteString("\n\n")

	// 帮助信息
	content.WriteString(helpStyle.Render(i18n.T("auth.any_key_return")))

	// 将内容放入窗口边框
	window := windowStyle.Render(content.String())

	// 使用 lipgloss.Place 实现完全居中
	availableHeight := m.height
	availableWidth := m.width
	if availableHeight <= 0 {
		availableHeight = 30
	}
	if availableWidth <= 0 {
		availableWidth = 80
	}

	// 使用 lipgloss.Place 进行居中定位
	centered := lipgloss.Place(
		availableWidth,
		availableHeight,
		lipgloss.Center,
		lipgloss.Center,
		window,
	)

	return centered
}

// renderMainView 渲染主界面（文件浏览器）
func (m Model) renderMainView() string {
	var s strings.Builder

	// 标题和当前路径
	title := i18n.T("filebrowser.title")
	if m.showHidden {
		title += i18n.T("filebrowser.show_hidden")
	}
	s.WriteString(titleStyle.Render(title))
	s.WriteString("\n")
	s.WriteString(i18n.Tf("filebrowser.current_dir", m.currentDir))
	s.WriteString(helpStyle.Render(i18n.T("filebrowser.legend")))

	// 文件列表
	if len(m.files) == 0 {
		s.WriteString(i18n.T("filebrowser.empty"))
	} else {
		// 显示文件列表
		maxHeight := m.height - 10 // 为三行状态栏和标题留空间
		if maxHeight < 5 || m.height == 0 {
			maxHeight = 10 // 为未初始化的终端提供合理的默认值
		}

		startIndex := 0
		if m.selectedIndex >= maxHeight {
			startIndex = m.selectedIndex - maxHeight + 1
		}

		endIndex := startIndex + maxHeight
		if endIndex > len(m.files) {
			endIndex = len(m.files)
		}

		for i := startIndex; i < endIndex; i++ {
			file := m.files[i]

			// 跳过空的占位符条目（第一个条目）
			if file.Name == "" {
				s.WriteString("\n") // 输出空行
				continue
			}

			prefix := "  "

			if i == m.selectedIndex {
				prefix = "> "
			}

			// 文件/目录图标和状态圆点
			var icon string
			var statusDot string

			if file.IsDir {
				icon = "📁"
				statusDot = ""
			} else {
				icon = "📄"
				// 检查文件上传状态并设置相应的状态圆点
				filePath := filepath.Join(m.currentDir, file.Name)
				status, exists := m.getFileUploadStatus(filePath)
				if exists {
					switch status {
					case "starting", "pending":
						statusDot = " 🟡" // 黄色圆点：等待中
					case "uploading":
						statusDot = " 🔵" // 蓝色圆点：上传中
					case "completed":
						statusDot = " 🟢" // 绿色圆点：已完成
					case "failed":
						statusDot = " 🔴" // 红色圆点：上传失败
					default:
						statusDot = ""
					}
				} else {
					statusDot = ""
				}
			}

			// 格式化大小
			sizeStr := ""
			if !file.IsDir {
				if file.Size < 1024 {
					sizeStr = fmt.Sprintf("%dB", file.Size)
				} else if file.Size < 1024*1024 {
					sizeStr = fmt.Sprintf("%.1fKB", float64(file.Size)/1024)
				} else if file.Size < 1024*1024*1024 {
					sizeStr = fmt.Sprintf("%.1fMB", float64(file.Size)/(1024*1024))
				} else {
					sizeStr = fmt.Sprintf("%.1fGB", float64(file.Size)/(1024*1024*1024))
				}
			}

			line := fmt.Sprintf("%s%s %s%s", prefix, icon, file.Name, statusDot)
			if sizeStr != "" {
				line += fmt.Sprintf(" (%s)", sizeStr)
			}

			// 根据选中状态设置颜色
			if i == m.selectedIndex {
				line = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(line)
			}

			s.WriteString(line)
			s.WriteString("\n")
		}

		// 显示滚动指示器
		if len(m.files) > maxHeight {
			s.WriteString(i18n.Tf("filebrowser.scroll", startIndex+1, endIndex, len(m.files)))
		}
	}

	return s.String()
}

// renderSettings 渲染设置界面
func (m Model) renderSettings() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(i18n.T("settings.title")))
	s.WriteString("\n\n")

	// 赞助者状态提示
	if m.userInfo.IsSponsored {
		s.WriteString(i18n.T("settings.sponsor_only"))
	} else {
		s.WriteString(i18n.T("settings.require_sponsor"))
	}

	// 所有用户都可以修改语言；赞助者还能修改上传参数
	var settingsKeys []string
	var settingsLabels []string
	var settingsSponsored []bool

	if m.userInfo.IsSponsored {
		settingsKeys = []string{"chunk_size", "concurrency", "server", "quick_upload", "language"}
		settingsLabels = []string{
			i18n.T("settings.chunk_size"),
			i18n.T("settings.concurrency"),
			i18n.T("settings.server"),
			i18n.T("settings.quick_upload"),
			i18n.T("settings.language"),
		}
		settingsSponsored = []bool{true, true, true, true, false}
	} else {
		settingsKeys = []string{"language"}
		settingsLabels = []string{i18n.T("settings.language")}
		settingsSponsored = []bool{false}
	}

	for i, key := range settingsKeys {
		prefix := "  "
		if i == m.settingsIndex {
			prefix = "> "
		}

		label := settingsLabels[i]

		// 检查权限
		isLocked := settingsSponsored[i] && !m.userInfo.IsSponsored
		if isLocked {
			label += " 🔒"
		}

		var line string

		if key == "server" && m.userInfo.IsSponsored {
			currentServer := i18n.T("settings.default_server")
			if m.serverIndex < len(m.availableServers) && len(m.availableServers) > 0 {
				currentServer = m.availableServers[m.serverIndex].Name
			}
			line = fmt.Sprintf("%s%s\n%s  %s %s", prefix, label, strings.Repeat(" ", len(prefix)), currentServer, i18n.T("settings.switch_lr"))
		} else if key == "quick_upload" && m.userInfo.IsSponsored {
			status := i18n.T("settings.off")
			if m.config.QuickUpload {
				status = i18n.T("settings.on")
			}
			line = fmt.Sprintf("%s%s\n%s  %s %s", prefix, label, strings.Repeat(" ", len(prefix)), status, i18n.T("settings.toggle_space"))
		} else if key == "language" {
			currentLang := i18n.LanguageNames[i18n.GetLanguage()]
			line = fmt.Sprintf("%s%s\n%s  %s %s", prefix, label, strings.Repeat(" ", len(prefix)), currentLang, i18n.T("settings.switch_lr"))
		} else if isLocked {
			var value string
			switch key {
			case "chunk_size":
				value = fmt.Sprintf("%d", m.config.ChunkSize)
			case "concurrency":
				value = fmt.Sprintf("%d", m.config.MaxConcurrent)
			}
			line = fmt.Sprintf("%s%s\n%s  %s %s", prefix, label, strings.Repeat(" ", len(prefix)), value, i18n.T("settings.read_only"))
		} else {
			input := m.settingsInputs[key]
			line = fmt.Sprintf("%s%s\n%s  %s", prefix, label, strings.Repeat(" ", len(prefix)), input.View())
		}

		if isLocked {
			line = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(line)
		} else if i == m.settingsIndex {
			line = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(line)
		}

		s.WriteString(line)
		s.WriteString("\n\n")
	}

	return s.String()
}

// renderUploadList 渲染上传管理器
func (m Model) renderUploadList() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(i18n.T("upload_list.title")))
	s.WriteString("\n\n")

	if len(m.uploadTasks) == 0 {
		s.WriteString(i18n.T("upload_list.empty"))
	} else {
		s.WriteString(m.uploadTable.View())
	}

	return s.String()
}

// renderError 渲染错误界面
func (m Model) renderError() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(i18n.T("error.title")))
	s.WriteString("\n\n")
	if m.err != nil {
		s.WriteString(errorStyle.Render(m.err.Error()))
	}
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render(i18n.T("error.retry")))

	return s.String()
}

// startUpload 开始上传文件
func (m Model) startUpload(filePath, taskID, statusFile string) tea.Cmd {
	return func() tea.Msg {
		// CLI现在是自包含的，不需要预先获取上传信息
		// 启动CLI进程，只传递CLI支持的参数

		skipUpload := "1"
		if !m.config.QuickUpload {
			skipUpload = "0"
		}

		// 获取当前选中的服务器信息
		selectedServerName := "未知"
		selectedServerURL := ""
		if m.serverIndex < len(m.availableServers) && len(m.availableServers) > 0 {
			selectedServer := m.availableServers[m.serverIndex]
			selectedServerName = selectedServer.Name
			selectedServerURL = selectedServer.URL
		}

		// 构建CLI命令参数
		args := []string{
			"-file", filePath,
			"-token", m.config.Token,
			"-task-id", taskID,
			"-status-file", statusFile,
			"-chunk-size", fmt.Sprintf("%d", m.config.ChunkSize),
			"-model", "1",
			"-mr-id", "0",
			"-skip-upload", skipUpload,
			"-server-name", selectedServerName,
		}

		// GUI模式下始终传递选中的上传服务器地址
		if selectedServerURL != "" {
			args = append(args, "-upload-server", selectedServerURL)
		}

		cmd := exec.Command(m.cliPath, args...)

		// 设置输出到文件，便于调试
		logFile := statusFile + ".log"
		if file, err := os.Create(logFile); err == nil {
			cmd.Stdout = file
			cmd.Stderr = file
		}

		// 启动进程但不等待完成
		err := cmd.Start()
		if err != nil {
			return UploadErrorMsg{Error: fmt.Sprintf("启动CLI失败: %v", err), TaskID: taskID}
		}

		// 获取进程ID
		processID := cmd.Process.Pid

		// 后台等待进程完成
		go func() {
			cmd.Wait() // 等待进程完成
		}()

		// 返回进程启动消息，包含进程ID
		return ProcessStartedMsg{TaskID: taskID, ProcessID: processID}
	}
}

// UploadInfo 上传信息
type UploadInfo struct {
	Server string
	UToken string
}

// getUploadInfo 获取上传服务器和token信息
func (m Model) getUploadInfo(filePath string) (*UploadInfo, error) {
	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 计算文件SHA1
	sha1Hash, err := calculateFileSHA1(filePath)
	if err != nil {
		return nil, fmt.Errorf("计算SHA1失败: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// 调用upload_request_select2获取上传服务器
	formData := fmt.Sprintf("action=upload_request_select2&sha1=%s&filename=%s&filesize=%d&model=1&token=%s",
		sha1Hash, filepath.Base(filePath), fileInfo.Size(), m.config.Token)

	req, err := http.NewRequest("POST", "https://tmplink-sec.vxtrans.com/api_v2/file", strings.NewReader(formData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var selectResp struct {
		Status int `json:"status"`
		Data   struct {
			UToken  string      `json:"utoken"`
			Servers interface{} `json:"servers"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &selectResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if selectResp.Status != 1 {
		return nil, fmt.Errorf("获取上传服务器失败，状态码: %d", selectResp.Status)
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

	if uploadServer == "" {
		return nil, fmt.Errorf("无法获取上传服务器地址")
	}

	return &UploadInfo{
		Server: uploadServer,
		UToken: selectResp.Data.UToken,
	}, nil
}

// calculateFileSHA1 计算文件SHA1
func calculateFileSHA1(filePath string) (string, error) {
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

// checkProgress 检查上传进度
func (m Model) checkProgress(taskID string) tea.Cmd {
	return func() tea.Msg {
		statusFile, exists := m.statusFiles[taskID]
		if !exists {
			return UploadErrorMsg{Error: "找不到状态文件", TaskID: taskID}
		}

		// 读取状态文件
		data, err := os.ReadFile(statusFile)
		if err != nil {
			// 文件可能还没创建，返回待检查消息
			return UploadProgressMsg{TaskID: taskID, Progress: 0.0, Speed: 0.0}
		}

		var task TaskStatus
		if err := json.Unmarshal(data, &task); err != nil {
			// JSON解析失败，返回待检查消息
			return UploadProgressMsg{TaskID: taskID, Progress: 0.0, Speed: 0.0}
		}

		switch task.Status {
		case "completed":
			return UploadCompleteMsg{TaskID: taskID, DownloadURL: task.DownloadURL}
		case "failed":
			return UploadErrorMsg{Error: task.ErrorMsg, TaskID: taskID}
		default:
			// 返回当前进度，继续监控
			return UploadProgressMsg{TaskID: taskID, Progress: task.Progress, Speed: task.UploadSpeed}
		}
	}
}

// startProgressTimer 启动进度检查定时器
func (m Model) startProgressTimer(taskID string) tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return CheckProgressTickMsg{TaskID: taskID}
	})
}

// 消息类型
type UploadProgressMsg struct {
	TaskID   string
	Progress float64
	Speed    float64 // KB/s
}

type UploadCompleteMsg struct {
	TaskID      string
	DownloadURL string
}

type UploadErrorMsg struct {
	Error  string
	TaskID string
}

type ProcessStartedMsg struct {
	TaskID    string
	ProcessID int
}

type CheckProgressTickMsg struct {
	TaskID string
}

// 样式
// getConfigPath 获取配置文件路径
func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/tmp"
	}
	return filepath.Join(homeDir, ".tmplink_config.json")
}

// loadConfig 加载配置
func loadConfig() Config {
	configPath := getConfigPath()

	// 如果配置文件不存在，返回默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return defaultConfig()
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return defaultConfig()
	}

	// 解析配置
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return defaultConfig()
	}

	// 清理旧的默认配置
	if config.UploadServer == "https://tmplink-sec.vxtrans.com/api_v2" {
		config.UploadServer = ""
	}
	if config.SelectedServerName == "默认 (自动选择)" || config.SelectedServerName == "默认服务器" {
		config.SelectedServerName = ""
	}

	// 应用已保存的语言设置
	if config.Language != "" {
		i18n.SetLanguage(config.Language)
	}

	return config
}

// saveConfig 保存配置
func saveConfig(config Config) error {
	configPath := getConfigPath()

	// 确保目录存在
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// 序列化配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(configPath, data, 0644)
}

// simplifyErrorMessage 简化错误信息，使其对用户更友好
func simplifyErrorMessage(err string) string {
	errLower := strings.ToLower(err)

	if strings.Contains(errLower, "unauthorized") || strings.Contains(errLower, "invalid") || strings.Contains(errLower, "status") {
		return i18n.T("err.invalid_token")
	}
	if strings.Contains(errLower, "timeout") || strings.Contains(errLower, "network") {
		return i18n.T("err.timeout")
	}
	if strings.Contains(errLower, "connection") {
		return i18n.T("err.connection")
	}
	if strings.Contains(errLower, "parse") || strings.Contains(errLower, "json") {
		return i18n.T("err.parse")
	}
	return i18n.T("err.generic")
}

// validateAndSaveToken 验证token并保存配置
func (m Model) validateAndSaveToken(token string) tea.Cmd {
	return func() tea.Msg {
		// 先验证token
		userInfo, err := callUserAPI(token)
		if err != nil {
			return UserInfoErrorMsg{Error: err.Error()}
		}

		// 验证成功后保存token到配置
		config := m.config
		config.Token = token
		if err := saveConfig(config); err != nil {
			return UserInfoErrorMsg{Error: fmt.Sprintf("保存配置失败: %v", err)}
		}

		return TokenValidatedMsg{Token: token, UserInfo: userInfo}
	}
}

// startReturnToTokenInputDelay 启动3秒延迟后返回Token输入界面
func (m Model) startReturnToTokenInputDelay() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return ReturnToTokenInputMsg{}
	})
}

// fetchUserInfo 获取用户信息
func (m Model) fetchUserInfo() tea.Cmd {
	return func() tea.Msg {
		// 调用实际API获取用户信息
		userInfo, err := callUserAPI(m.config.Token)
		if err != nil {
			return UserInfoErrorMsg{Error: err.Error()}
		}

		return UserInfoMsg{UserInfo: userInfo}
	}
}

// callUserAPI 调用用户信息API
func callUserAPI(token string) (UserInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	// 第一步：获取基本用户信息和存储信息
	detailData := fmt.Sprintf("action=get_detail&token=%s", token)
	detailReq, err := http.NewRequest("POST", "https://tmplink-sec.vxtrans.com/api_v2/user", strings.NewReader(detailData))
	if err != nil {
		return UserInfo{}, err
	}

	detailReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	detailResp, err := client.Do(detailReq)
	if err != nil {
		return UserInfo{}, err
	}
	defer detailResp.Body.Close()

	if detailResp.StatusCode != 200 {
		return UserInfo{}, fmt.Errorf("HTTP错误: %d", detailResp.StatusCode)
	}

	detailBody, err := io.ReadAll(detailResp.Body)
	if err != nil {
		return UserInfo{}, err
	}

	// 解析详细信息响应
	var detailApiResp struct {
		Status int `json:"status"`
		Data   struct {
			UID         int64 `json:"uid"`
			Storage     int64 `json:"storage"`
			StorageUsed int64 `json:"storage_used"`
			Sponsor     bool  `json:"sponsor"`
		} `json:"data"`
		Msg string `json:"msg"`
	}

	if err := json.Unmarshal(detailBody, &detailApiResp); err != nil {
		return UserInfo{}, fmt.Errorf("解析详细信息失败: %w", err)
	}

	if detailApiResp.Status != 1 {
		return UserInfo{}, fmt.Errorf("获取详细信息失败: %s", detailApiResp.Msg)
	}

	// 第二步：获取用户名信息
	userInfoData := fmt.Sprintf("action=pf_userinfo_get&token=%s", token)
	userInfoReq, err := http.NewRequest("POST", "https://tmplink-sec.vxtrans.com/api_v2/user", strings.NewReader(userInfoData))
	if err != nil {
		return UserInfo{}, err
	}

	userInfoReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	userInfoResp, err := client.Do(userInfoReq)
	if err != nil {
		return UserInfo{}, err
	}
	defer userInfoResp.Body.Close()

	if userInfoResp.StatusCode != 200 {
		return UserInfo{}, fmt.Errorf("HTTP错误: %d", userInfoResp.StatusCode)
	}

	userInfoBody, err := io.ReadAll(userInfoResp.Body)
	if err != nil {
		return UserInfo{}, err
	}

	// 解析用户信息响应
	var userInfoApiResp struct {
		Status int `json:"status"`
		Data   struct {
			Nickname string `json:"nickname"`
		} `json:"data"`
		Msg string `json:"msg"`
	}

	if err := json.Unmarshal(userInfoBody, &userInfoApiResp); err != nil {
		return UserInfo{}, fmt.Errorf("解析用户信息失败: %w", err)
	}

	// 如果获取用户名失败，使用默认值
	username := "用户"
	if userInfoApiResp.Status == 1 && userInfoApiResp.Data.Nickname != "" {
		username = userInfoApiResp.Data.Nickname
	}

	return UserInfo{
		Username:    username,
		Email:       "", // API似乎不返回邮箱
		UID:         fmt.Sprintf("%d", detailApiResp.Data.UID),
		IsSponsored: detailApiResp.Data.Sponsor,
		UsedSpace:   detailApiResp.Data.StorageUsed,
		TotalSpace:  detailApiResp.Data.Storage,
	}, nil
}

// 消息类型
type UserInfoMsg struct {
	UserInfo UserInfo
}

type UserInfoErrorMsg struct {
	Error string
}

type TokenValidatedMsg struct {
	Token    string
	UserInfo UserInfo
}

type TokenValidationFailedMsg struct {
	Error string
}

type ReturnToTokenInputMsg struct{}

type FilesLoadedMsg struct {
	Files []FileInfo
}

// loadFiles 加载当前目录的文件列表
func (m Model) loadFiles() tea.Cmd {
	return func() tea.Msg {
		files, err := loadDirectoryFiles(m.currentDir, m.showHidden)
		if err != nil {
			return UserInfoErrorMsg{Error: fmt.Sprintf("加载目录失败: %v", err)}
		}
		return FilesLoadedMsg{Files: files}
	}
}

// loadDirectoryFiles 读取目录中的文件
func loadDirectoryFiles(dirPath string, showHidden bool) ([]FileInfo, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var files []FileInfo

	// 添加空的占位符条目作为第一个条目（解决居中显示问题）
	files = append(files, FileInfo{
		Name:  "", // 空名称，在渲染时会被跳过
		IsDir: false,
	})

	// 添加返回上级目录选项（除非已在根目录）
	if dirPath != "/" && dirPath != filepath.VolumeName(dirPath) {
		files = append(files, FileInfo{
			Name:  "..",
			IsDir: true,
		})
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue // 跳过无法读取的文件
		}

		// 过滤隐藏文件（以.开头的文件）
		if !showHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		files = append(files, FileInfo{
			Name:    entry.Name(),
			Size:    info.Size(),
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime(),
		})
	}

	return files, nil
}

// saveSettings 保存设置
func (m Model) saveSettings() (tea.Model, tea.Cmd) {
	// 根据用户类型确定可用设置（需与 renderSettings / handleSettings 保持一致）
	var settingsKeys []string
	var settingsSponsored []bool

	if m.userInfo.IsSponsored {
		settingsKeys = []string{"chunk_size", "concurrency", "server", "quick_upload", "language"}
		settingsSponsored = []bool{true, true, true, true, false}
	} else {
		settingsKeys = []string{"language"}
		settingsSponsored = []bool{false}
	}

	// 解析和验证输入值
	for i, key := range settingsKeys {
		// 检查权限
		if settingsSponsored[i] && !m.userInfo.IsSponsored {
			continue // 跳过无权限的设置
		}

		// 处理特殊设置项
		if key == "server" && m.userInfo.IsSponsored {
			if m.serverIndex < len(m.availableServers) {
				selectedServer := m.availableServers[m.serverIndex]
				m.config.SelectedServerName = selectedServer.Name
				if selectedServer.URL != "" {
					m.config.UploadServer = selectedServer.URL
				}
			}
			continue
		} else if key == "quick_upload" && m.userInfo.IsSponsored {
			// 已在按键处理中直接修改
			continue
		} else if key == "language" {
			// 语言已在按键处理中通过 applyLanguage 修改，config.Language 已更新
			continue
		}

		// 处理常规输入框设置
		input := m.settingsInputs[key]
		value := input.Value()

		// 解析数值
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err != nil {
			m.err = fmt.Errorf("设置 %s 的值无效: %s", key, value)
			m.state = StateError
			return m, nil
		}

		// 验证范围并应用设置
		switch key {
		case "chunk_size":
			if intValue < 1 || intValue > 99 {
				m.err = fmt.Errorf("分块大小必须在 1-99 MB 之间")
				m.state = StateError
				return m, nil
			}
			m.config.ChunkSize = intValue
		case "concurrency":
			if intValue < 1 || intValue > 20 {
				m.err = fmt.Errorf("并发数必须在 1-20 之间")
				m.state = StateError
				return m, nil
			}
			m.config.MaxConcurrent = intValue
		}
	}

	// 保存配置到文件
	if err := saveConfig(m.config); err != nil {
		m.err = fmt.Errorf("保存配置失败: %w", err)
		m.state = StateError
		return m, nil
	}

	// 返回主界面
	m.state = StateMain
	return m, nil
}

// 样式
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Padding(1, 0)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("40"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1).
			Width(0) // 动态设置宽度
)
