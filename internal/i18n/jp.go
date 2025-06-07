package i18n

// initJP initializes the Japanese translations
func initJP() {
	messages[LanguageJP] = map[string]string{
		// Common
		"app_name":                 "TmpLink ファイルアップローダー",
		"loading":                  "読み込み中...",
		"error":                    "エラー",
		"warning":                  "警告",
		"success":                  "成功",
		"yes":                      "はい",
		"no":                       "いいえ",
		"enabled":                  "有効",
		"disabled":                 "無効",
		"default":                  "デフォルト",
		"unknown":                  "不明",
		"locked":                   "🔒",
		"readonly":                 "読み取り専用",
		
		// Auth & User
		"enter_api_token":          "TmpLink APIトークンを入力してください",
		"token_validation_success": "トークンが正常に保存され、検証されました (UID: %s)",
		"token_validation_error":   "トークン検証に失敗しました: %v",
		"token_validation_failed":  "❌ トークン検証に失敗しました!",
		"token_error_message":      "❗ エラーメッセージ: %v",
		"token_help_message":       "💡 -set-tokenコマンドを使用して、有効なAPIトークンを設定してください",
		"validating_token":         "トークンの検証中...",
		"token_validation_success_mark": " ✅",
		"user_not_logged_in":       "ユーザー: ログインしていません",
		"user_info":                "ユーザー: %s",
		"user_sponsored":           " ✨ (スポンサー)",
		"user_regular":             " (一般ユーザー)",
		"storage_info":             "ストレージ: %.1fGB/%.1fGB (%.1f%%)",
		"storage_loading":          "ストレージ情報: 読み込み中...",
		"get_user_info_failed":     "ユーザー情報の取得に失敗しました: %s",
		
		// Main Menu
		"menu_title":               "メインメニュー",
		"menu_file_browser":        "ファイルブラウザ",
		"menu_upload_settings":     "アップロード設定",
		"menu_upload_manager":      "アップロード管理",
		
		// File Browser
		"file_browser_title":       "ファイルブラウザ",
		"show_hidden_files":        " (隠しファイルを表示)",
		"file_browser_legend":      "📁ディレクトリ 📄ファイル 🟡待機中 🔵アップロード中 🟢完了 🔴失敗",
		"directory_empty_loading":  "ディレクトリが空か読み込み中です...",
		
		// Upload Settings
		"settings_title":           "アップロード設定",
		"settings_sponsored_only":  "✨ スポンサー専用設定",
		"settings_some_sponsored":  "⚠️  一部の設定にはスポンサー権限が必要です",
		"chunk_size_mb":            "チャンクサイズ (MB):",
		"concurrency":              "同時実行数:",
		"upload_server":            "アップロードサーバー:",
		"quick_upload":             "クイックアップロード:",
		
		// Upload Manager & Status
		"upload_manager_title":     "アップロード管理",
		"no_upload_tasks":          "アップロードタスクはありません",
		"files_uploading":          "アップロード中: %dファイル",
		"file_uploading":           "ファイルをアップロード中です",
		"file_upload_complete":     "ファイルのアップロードが完了しました",
		"file_already_in_list":     "ファイルはすでにアップロードリストにあります",
		"file_upload_failed":       "アップロード失敗: %s",
		"upload_failed":            "❌ アップロード失敗!",
		"upload_file_name":         "📁 ファイル名: %s",
		"upload_error_message":     "❗ エラーメッセージ: %v",
		"upload_complete":          "✅ アップロード完了!",
		"upload_file_size":         "📊 ファイルサイズ: %s",
		"upload_average_speed":     "⚡ 平均速度: %.2f MB/s",
		"upload_total_time":        "⏱️  合計時間: %v",
		"upload_download_link":     "🔗 ダウンロードリンク: %s",
		"upload_in_progress":       "📤 アップロード中",
		"resuming_upload":          "🔄 レジューム検出: %d/%dチャンク完了 (%.1f%%)",
		"cannot_get_file_info":     "ファイル情報を取得できません: %v",
		"file_size_exceeded":       "ファイルサイズが制限を超えています。最大50GB、現在のファイル: %.2fGB",
		
		// Table Columns
		"column_filename":          "ファイル名",
		"column_size":              "サイズ",
		"column_progress":          "進捗",
		"column_speed":             "速度",
		"column_server":            "サーバー",
		"column_status":            "状態",
		
		// Status Translations
		"status_starting":          "開始中",
		"status_waiting":           "待機中",
		"status_uploading":         "アップロード中",
		"status_completed":         "完了",
		"status_failed":            "失敗",
		
		// Navigation & Controls
		"nav_file_browser":         "↑↓:選択 ←:上へ →:開く t:隠しファイル Tab:設定 Q:終了",
		"nav_settings":             "↑↓:選択 Enter:保存 Tab:アップロード管理 Esc:戻る Q:終了",
		"nav_upload_manager":       "↑↓:選択 d:削除 t:完了をクリア y:全クリア Tab:ファイル閲覧 Esc:戻る Q:終了",
		"nav_error":                "操作: Enter:再試行 Esc:戻る Q:終了",
		"nav_quit":                 "操作: Q:終了",
		"nav_error_hints":          "• Enter: 再試行 • Esc: 戻る",
		
		// Model values set
		"model_set":                "デフォルトの有効期限が設定されました: %s",
		"dir_id_set":               "デフォルトのディレクトリIDが設定されました: %s",
	}
}