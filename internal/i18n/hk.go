package i18n

// initHK initializes the Traditional Chinese (Hong Kong) translations
func initHK() {
	messages[LanguageHK] = map[string]string{
		// Common
		"app_name":                 "鐵盤檔案上傳工具",
		"loading":                  "正在載入中...",
		"error":                    "錯誤",
		"warning":                  "警告",
		"success":                  "成功",
		"yes":                      "是",
		"no":                       "否",
		"enabled":                  "開啟",
		"disabled":                 "關閉",
		"default":                  "預設",
		"unknown":                  "未知",
		"locked":                   "🔒",
		"readonly":                 "唯讀",
		
		// Auth & User
		"enter_api_token":          "請輸入鐵盤 API Token",
		"user_info_section":        "用戶資料",
		"user_email":               "電子郵件",
		"content_count":            "內容統計",
		"files":                    "檔案",
		"folders":                  "資料夾",
		"config_section":           "設定資料",
		"token_validation_success": "Token 已成功儲存並驗證 (UID: %s)",
		"token_validation_error":   "Token 驗證失敗: %v",
		"token_validation_failed":  "❌ Token 驗證失敗!",
		"token_error_message":      "❗ 錯誤訊息: %v",
		"token_help_message":       "💡 請使用 -set-token 命令重新設置有效的 API Token",
		"validating_token":         "正在驗證 Token 有效性...",
		"token_validation_success_mark": " ✅",
		"user_not_logged_in":       "用戶: 未登入",
		"user_info":                "用戶",
		"user_sponsored":           " ✨ (贊助者)",
		"user_regular":             " (普通用戶)",
			"user_level_info":          "級別: %s",
			"user_since":               "註冊時間: %s",
			"sponsor_expires":          "到期時間: %s",
		"storage_info":             "儲存: %.2fGB/%.2fGB (%.1f%%)",
		"storage_loading":          "儲存訊息: 載入中...",
		"get_user_info_failed":     "獲取用戶信息失敗: %s",
		
		// Main Menu
		"menu_title":               "功能選單",
		"menu_file_browser":        "檔案瀏覽器",
		"menu_upload_settings":     "上傳設置",
		"menu_upload_manager":      "上傳管理器",
		
		// File Browser
		"file_browser_title":       "檔案瀏覽器",
		"show_hidden_files":        " (顯示隱藏檔案)",
		"file_browser_legend":      "📁目錄 📄檔案 🟡等待 🔵上傳中 🟢已完成 🔴失敗",
		"directory_empty_loading":  "目錄為空或正在載入...",
		
		// Upload Settings
		"settings_title":           "上傳設置",
		"settings_sponsored_only":  "✨ 贊助者專享設置",
		"settings_some_sponsored":  "⚠️  部分設置需要贊助者權限",
		"chunk_size_mb":            "分塊大小 (MB):",
		"concurrency":              "並發數:",
		"upload_server":            "上傳伺服器:",
		"quick_upload":             "快速上傳:",
		
		// Upload Manager & Status
		"upload_manager_title":     "上傳管理器",
		"no_upload_tasks":          "暫無上傳任務",
		"files_uploading":          "上傳中: %d個檔案",
		"file_uploading":           "檔案正在上傳中",
		"file_upload_complete":     "檔案已上傳完成",
		"file_already_in_list":     "檔案已在上傳列表中",
		"file_upload_failed":       "上傳失敗: %s",
		"upload_failed":            "❌ 上傳失敗!",
		"upload_file_name":         "📁 檔案名: %s",
		"upload_error_message":     "❗ 錯誤訊息: %v",
		"upload_complete":          "✅ 上傳完成!",
		"upload_file_size":         "📊 檔案大小: %s",
		"upload_average_speed":     "⚡ 平均速度: %.2f MB/s",
		"upload_total_time":        "⏱️  總耗時: %v",
		"upload_download_link":     "🔗 下載連結: %s",
		"upload_in_progress":       "📤 上傳中",
		"resuming_upload":          "🔄 檢測到斷點續傳: 已完成 %d/%d 分片 (%.1f%%)",
		"cannot_get_file_info":     "無法獲取檔案訊息: %v",
		"file_size_exceeded":       "檔案大小超出限制，最大支持50GB，當前檔案: %.2fGB",
		
		// Table Columns
		"column_filename":          "檔案名",
		"column_size":              "大小",
		"column_progress":          "進度",
		"column_speed":             "速度",
		"column_server":            "伺服器",
		"column_status":            "狀態",
		
		// Error Messages
		"sync_user_settings":       "正在同步用戶設定...",
		"cli_usage":                "鐵盤 CLI - 鐵盤文件上傳工具",
		"cli_language":             "介面語言 (cn/en/hk/jp)",
		"cli_set_language":         "設定並儲存介面語言 (cn/en/hk/jp)",
		"language_set_to":          "語言設定已儲存為",
		"api_validation_failed":     "API 驗證失敗: %s",
		"create_request_failed":    "創建請求失敗: %v",
		"send_request_failed":     "發送請求失敗: %v",
		"server_error_status":     "伺服器返回錯誤狀態碼: %d",
		"read_response_failed":    "讀取回應失敗: %v",
		"parse_response_failed":   "解析回應失敗: %v (原始回應: %s)",
		"unknown_error":           "未知錯誤",
		"error_missing_file":      "錯誤: 缺少必需參數 -file",

		// Status Translations
		"status_starting":          "啟動中",
		"status_waiting":           "等待中",
		"status_uploading":         "上傳中",
		"status_completed":         "已完成",
		"status_failed":            "失敗",
		
		// Navigation & Controls
		"nav_file_browser":         "↑↓:選擇 ←:上級 →:進入 t:隱藏檔案 Tab:設置 Q:退出",
		"nav_settings":             "↑↓:選擇 Enter:儲存 Tab:上傳管理 Esc:返回 Q:退出",
		"nav_upload_manager":       "↑↓:選擇 d:刪除 t:清除完成 y:清除全部 Tab:檔案瀏覽 Esc:返回 Q:退出",
		"nav_error":                "操作: Enter:重試 Esc:返回 Q:退出",
		"nav_quit":                 "操作: Q:退出",
		"nav_error_hints":          "• Enter: 重試 • Esc: 返回",
		
		// Model values set
		"model_set":                "預設檔案有效期已設置為: %s",
		"dir_id_set":               "預設目錄ID已設置為: %s",
		
		// Config status
		"config_status_title":      "--- 鐵盤 CLI 配置狀態 ---",
		"config_token":             "Token",
		"config_token_valid_short":       "✅ 有效 (UID: %s)",
		"config_token_invalid_short":     "❌ 無效",
		"config_token_not_set":     "Token: ❓ 未設置",
		"config_model":             "預設檔案有效期",
		"config_dir_id_default":            "預設目錄ID",
		"config_language_setting":  "語言設置",
		"config_language_current":  "當前使用",
		"config_language_auto":     "自動檢測",
			"current_directory":         "當前目錄",
			"file_uploading_status":      "檔案正在上傳中",
			"file_completed_status":      "檔案已上傳完成",
			"file_in_list_status":        "檔案已在上傳列表中",
			"token_input_help":          "• Enter: 繼續 • Ctrl+C: 退出",
	}
}