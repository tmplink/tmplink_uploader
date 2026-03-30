package i18n

import "fmt"

// Language constants
const (
	LangZhCN = "zh-CN"
	LangZhTW = "zh-TW"
	LangEn   = "en"
	LangJa   = "ja"
	LangRu   = "ru"
	LangFr   = "fr"
	LangMs   = "ms"
)

// DefaultLanguage is the fallback language
const DefaultLanguage = LangZhCN

// SupportedLanguages lists all supported languages in display order
var SupportedLanguages = []string{LangZhCN, LangZhTW, LangEn, LangJa, LangRu, LangFr, LangMs}

// LanguageNames maps language codes to their display names
var LanguageNames = map[string]string{
	LangZhCN: "简体中文",
	LangZhTW: "繁體中文",
	LangEn:   "English",
	LangJa:   "日本語",
	LangRu:   "Русский",
	LangFr:   "Français",
	LangMs:   "Bahasa Melayu",
}

var translations = map[string]map[string]string{
	LangZhCN: {
		// App
		"app.title": "钛盘文件上传工具",

		// Language select screen
		"lang.select_title":  "选择语言",
		"lang.select_hint":   "↑↓:选择  Enter:确认  Ctrl+C:退出",

		// Auth / Token input
		"auth.subtitle":       "安全认证",
		"auth.validating":     "%s 正在验证Token，请稍候...",
		"auth.wait":           "请耐心等待验证完成...",
		"auth.instructions":   "请输入您的钛盘 Token 来开始使用：\n\n1. 访问 https://tmp.link/ 并登录您的账户\n2. 点击\"上传文件\"按钮，然后点击\"重新设定\"按钮，滑动到窗口底部，点击 \"使用 CLI 上传\"\n3. 点击 Token 以复制到剪贴板",
		"auth.token_label":    "Token:",
		"auth.help":           "💡 Enter: 验证并保存  •  Ctrl+C: 退出程序",
		"auth.error_box":      "❌ 验证失败\n\n%s",
		"auth.placeholder":    "请输入钛盘 API Token",

		// Token validation failed
		"auth.failed_title":   "❌ 验证失败",
		"auth.failed_subtitle":"Token 认证出现问题",
		"auth.error_details":  "错误详情：\n\n%s",
		"auth.auto_return":    "%s 3秒后自动返回输入界面...",
		"auth.any_key_return": "💡 按任意键立即返回  •  Ctrl+C: 退出程序",

		// Loading
		"loading.token": "\n%s 正在验证Token...",

		// Status bar
		"status.user":          "用户: %s",
		"status.not_logged_in": "用户: 未登录",
		"status.sponsor":       " ✨ (赞助者)",
		"status.regular":       " (普通用户)",
		"status.storage":       "存储: %.1fGB/%.1fGB (%.1f%%)",
		"status.no_storage":    "存储: 无私有空间",
		"status.uploading":     " | 上传中: %d个文件",
		"status.speed_mb":      " (%.1fMB/s)",
		"status.speed_kb":      " (%.1fKB/s)",

		// Navigation hint keys
		"nav.enter":           "进入",
		"nav.upload":          "上传",
		"nav.keys_with_parent":"↑↓:选择 ←→:上级 Enter:%s t:隐藏文件 Tab:设置 Q:退出",
		"nav.keys_no_parent":  "↑↓:选择 Enter:%s t:隐藏文件 Tab:设置 Q:退出",
		"settings.keys":       "↑↓:选择 Enter:保存 Tab:上传管理 Esc:返回 Q:退出",
		"upload_list.keys":    "↑↓:选择 d:删除 t:清除完成 y:清除全部 Tab:文件浏览 Esc:返回 Q:退出",
		"error.keys":          "操作: Enter:重试 Esc:返回 Q:退出",
		"default.keys":        "操作: Q:退出",

		// Menu
		"menu.title":                "功能菜单",
		"menu.file_browser":         "文件浏览器",
		"menu.file_browser_desc":    "选择要上传的文件",
		"menu.settings":             "上传设置",
		"menu.settings_desc":        "配置上传参数",
		"menu.upload_manager":       "上传管理器",
		"menu.upload_manager_desc":  "查看和管理上传任务",

		// File browser
		"filebrowser.title":       "文件浏览器",
		"filebrowser.show_hidden": " (显示隐藏文件)",
		"filebrowser.current_dir": "当前目录: %s\n",
		"filebrowser.legend":      "📁目录 📄文件 🟡等待 🔵上传中 🟢已完成 🔴失败\n\n",
		"filebrowser.empty":       "目录为空或正在加载...",
		"filebrowser.scroll":      "\n[显示 %d-%d / 共 %d 项]",

		// Settings
		"settings.title":           "上传设置",
		"settings.sponsor_only":    "✨ 赞助者专享设置\n\n",
		"settings.require_sponsor": "⚠️  部分设置需要赞助者权限\n\n",
		"settings.chunk_size":      "分块大小 (MB):",
		"settings.concurrency":     "并发数:",
		"settings.server":          "上传服务器:",
		"settings.quick_upload":    "快速上传:",
		"settings.language":        "界面语言:",
		"settings.default_server":  "默认",
		"settings.on":              "开启",
		"settings.off":             "关闭",
		"settings.switch_lr":       "(←/→ 切换)",
		"settings.toggle_space":    "(Space 切换)",
		"settings.read_only":       "(只读)",
		"settings.chunk_placeholder":       "分块大小(MB)",
		"settings.concurrency_placeholder": "并发数",

		// Upload manager
		"upload_list.title":       "上传管理器",
		"upload_list.empty":       "暂无上传任务",
		"upload_list.col_filename":"文件名",
		"upload_list.col_size":    "大小",
		"upload_list.col_progress":"进度",
		"upload_list.col_speed":   "速度",
		"upload_list.col_server":  "服务器",
		"upload_list.col_status":  "状态",

		// Error screen
		"error.title": "错误",
		"error.retry": "• Enter: 重试 • Esc: 返回",

		// Task statuses
		"task.starting":   "启动中",
		"task.pending":    "等待中",
		"task.uploading":  "上传中",
		"task.completed":  "已完成",
		"task.failed":     "失败",
		"task.unknown_server": "未知",

		// Upload check messages
		"upload.in_progress": "文件正在上传中",
		"upload.completed":   "文件已上传完成",
		"upload.in_list":     "文件已在上传列表中",

		// Error simplification
		"err.invalid_token":   "Token无效，请检查后重新输入",
		"err.timeout":         "网络连接超时，请检查网络后重试",
		"err.connection":      "无法连接到服务器，请检查网络连接",
		"err.parse":           "无法读取到信息，请检查Token是否正确",
		"err.generic":         "Token验证失败，请重新输入",

		// Misc
		"unknown_state": "未知状态",
	},

	LangEn: {
		// App
		"app.title": "TMPLINK File Uploader",

		// Language select screen
		"lang.select_title": "Select Language",
		"lang.select_hint":  "↑↓:Select  Enter:Confirm  Ctrl+C:Quit",

		// Auth / Token input
		"auth.subtitle":    "Authentication",
		"auth.validating":  "%s Validating token, please wait...",
		"auth.wait":        "Please wait for validation to complete...",
		"auth.instructions":"Enter your TMPLINK Token to get started:\n\n1. Visit https://tmp.link/ and log in\n2. Click \"Upload File\", then \"Reset\", scroll to the bottom and click \"Use CLI Upload\"\n3. Click the Token to copy it to the clipboard",
		"auth.token_label": "Token:",
		"auth.help":        "💡 Enter: Verify & Save  •  Ctrl+C: Quit",
		"auth.error_box":   "❌ Verification Failed\n\n%s",
		"auth.placeholder": "Enter your TMPLINK API Token",

		// Token validation failed
		"auth.failed_title":    "❌ Verification Failed",
		"auth.failed_subtitle": "Token authentication issue",
		"auth.error_details":   "Error details:\n\n%s",
		"auth.auto_return":     "%s Returning to input screen in 3 seconds...",
		"auth.any_key_return":  "💡 Press any key to return immediately  •  Ctrl+C: Quit",

		// Loading
		"loading.token": "\n%s Validating token...",

		// Status bar
		"status.user":          "User: %s",
		"status.not_logged_in": "User: Not logged in",
		"status.sponsor":       " ✨ (Sponsor)",
		"status.regular":       " (Free User)",
		"status.storage":       "Storage: %.1fGB/%.1fGB (%.1f%%)",
		"status.no_storage":    "Storage: No private space",
		"status.uploading":     " | Uploading: %d file(s)",
		"status.speed_mb":      " (%.1fMB/s)",
		"status.speed_kb":      " (%.1fKB/s)",

		// Navigation hint keys
		"nav.enter":            "Open",
		"nav.upload":           "Upload",
		"nav.keys_with_parent": "↑↓:Select ←→:Parent Enter:%s t:Hidden Tab:Settings Q:Quit",
		"nav.keys_no_parent":   "↑↓:Select Enter:%s t:Hidden Tab:Settings Q:Quit",
		"settings.keys":        "↑↓:Select Enter:Save Tab:Uploads Esc:Back Q:Quit",
		"upload_list.keys":     "↑↓:Select d:Delete t:ClearDone y:ClearAll Tab:Files Esc:Back Q:Quit",
		"error.keys":           "Actions: Enter:Retry Esc:Back Q:Quit",
		"default.keys":         "Actions: Q:Quit",

		// Menu
		"menu.title":               "Menu",
		"menu.file_browser":        "File Browser",
		"menu.file_browser_desc":   "Select files to upload",
		"menu.settings":            "Upload Settings",
		"menu.settings_desc":       "Configure upload parameters",
		"menu.upload_manager":      "Upload Manager",
		"menu.upload_manager_desc": "View and manage upload tasks",

		// File browser
		"filebrowser.title":       "File Browser",
		"filebrowser.show_hidden": " (Showing hidden files)",
		"filebrowser.current_dir": "Current directory: %s\n",
		"filebrowser.legend":      "📁Dir 📄File 🟡Waiting 🔵Uploading 🟢Done 🔴Failed\n\n",
		"filebrowser.empty":       "Directory is empty or loading...",
		"filebrowser.scroll":      "\n[Showing %d-%d of %d items]",

		// Settings
		"settings.title":           "Upload Settings",
		"settings.sponsor_only":    "✨ Sponsor-exclusive settings\n\n",
		"settings.require_sponsor": "⚠️  Some settings require sponsor access\n\n",
		"settings.chunk_size":      "Chunk Size (MB):",
		"settings.concurrency":     "Concurrency:",
		"settings.server":          "Upload Server:",
		"settings.quick_upload":    "Quick Upload:",
		"settings.language":        "Interface Language:",
		"settings.default_server":  "Default",
		"settings.on":              "On",
		"settings.off":             "Off",
		"settings.switch_lr":       "(←/→ Switch)",
		"settings.toggle_space":    "(Space Toggle)",
		"settings.read_only":       "(Read-only)",
		"settings.chunk_placeholder":       "Chunk size (MB)",
		"settings.concurrency_placeholder": "Concurrency",

		// Upload manager
		"upload_list.title":        "Upload Manager",
		"upload_list.empty":        "No upload tasks",
		"upload_list.col_filename": "Filename",
		"upload_list.col_size":     "Size",
		"upload_list.col_progress": "Progress",
		"upload_list.col_speed":    "Speed",
		"upload_list.col_server":   "Server",
		"upload_list.col_status":   "Status",

		// Error screen
		"error.title": "Error",
		"error.retry": "• Enter: Retry • Esc: Back",

		// Task statuses
		"task.starting":       "Starting",
		"task.pending":        "Pending",
		"task.uploading":      "Uploading",
		"task.completed":      "Completed",
		"task.failed":         "Failed",
		"task.unknown_server": "Unknown",

		// Upload check messages
		"upload.in_progress": "File is being uploaded",
		"upload.completed":   "File has been uploaded",
		"upload.in_list":     "File is already in the upload queue",

		// Error simplification
		"err.invalid_token": "Invalid token, please check and re-enter",
		"err.timeout":       "Network timeout, please check your connection",
		"err.connection":    "Cannot connect to server, please check your network",
		"err.parse":         "Cannot read response, please check your token",
		"err.generic":       "Token validation failed, please re-enter",

		// Misc
		"unknown_state": "Unknown state",
	},

	LangJa: {
		// App
		"app.title": "TMPLINK ファイルアップローダー",

		// Language select screen
		"lang.select_title": "言語を選択",
		"lang.select_hint":  "↑↓:選択  Enter:確認  Ctrl+C:終了",

		// Auth / Token input
		"auth.subtitle":    "認証",
		"auth.validating":  "%s Tokenを検証中です...",
		"auth.wait":        "検証が完了するまでお待ちください...",
		"auth.instructions":"TMPLINK Tokenを入力して開始してください：\n\n1. https://tmp.link/ にアクセスしてログイン\n2. 「ファイルをアップロード」→「リセット」をクリックし、下にスクロールして「CLIでアップロード」をクリック\n3. Tokenをクリックしてクリップボードにコピー",
		"auth.token_label": "Token:",
		"auth.help":        "💡 Enter: 検証して保存  •  Ctrl+C: 終了",
		"auth.error_box":   "❌ 検証失敗\n\n%s",
		"auth.placeholder": "TMPLINK API Tokenを入力",

		// Token validation failed
		"auth.failed_title":    "❌ 検証失敗",
		"auth.failed_subtitle": "Token認証に問題が発生しました",
		"auth.error_details":   "エラーの詳細：\n\n%s",
		"auth.auto_return":     "%s 3秒後に入力画面に戻ります...",
		"auth.any_key_return":  "💡 任意のキーで即時に戻る  •  Ctrl+C: 終了",

		// Loading
		"loading.token": "\n%s Tokenを検証中...",

		// Status bar
		"status.user":          "ユーザー: %s",
		"status.not_logged_in": "ユーザー: 未ログイン",
		"status.sponsor":       " ✨ (スポンサー)",
		"status.regular":       " (一般ユーザー)",
		"status.storage":       "ストレージ: %.1fGB/%.1fGB (%.1f%%)",
		"status.no_storage":    "ストレージ: プライベートスペースなし",
		"status.uploading":     " | アップロード中: %dファイル",
		"status.speed_mb":      " (%.1fMB/s)",
		"status.speed_kb":      " (%.1fKB/s)",

		// Navigation hint keys
		"nav.enter":            "開く",
		"nav.upload":           "アップロード",
		"nav.keys_with_parent": "↑↓:選択 ←→:上へ Enter:%s t:隠しファイル Tab:設定 Q:終了",
		"nav.keys_no_parent":   "↑↓:選択 Enter:%s t:隠しファイル Tab:設定 Q:終了",
		"settings.keys":        "↑↓:選択 Enter:保存 Tab:アップロード Esc:戻る Q:終了",
		"upload_list.keys":     "↑↓:選択 d:削除 t:完了クリア y:全クリア Tab:ファイル Esc:戻る Q:終了",
		"error.keys":           "操作: Enter:再試行 Esc:戻る Q:終了",
		"default.keys":         "操作: Q:終了",

		// Menu
		"menu.title":               "メニュー",
		"menu.file_browser":        "ファイルブラウザ",
		"menu.file_browser_desc":   "アップロードするファイルを選択",
		"menu.settings":            "アップロード設定",
		"menu.settings_desc":       "アップロードパラメータを設定",
		"menu.upload_manager":      "アップロードマネージャー",
		"menu.upload_manager_desc": "アップロードタスクを管理",

		// File browser
		"filebrowser.title":       "ファイルブラウザ",
		"filebrowser.show_hidden": " (隠しファイルを表示)",
		"filebrowser.current_dir": "現在のディレクトリ: %s\n",
		"filebrowser.legend":      "📁ディレクトリ 📄ファイル 🟡待機中 🔵アップロード中 🟢完了 🔴失敗\n\n",
		"filebrowser.empty":       "ディレクトリが空またはロード中...",
		"filebrowser.scroll":      "\n[%d-%d / 全%d件を表示]",

		// Settings
		"settings.title":           "アップロード設定",
		"settings.sponsor_only":    "✨ スポンサー専用設定\n\n",
		"settings.require_sponsor": "⚠️  一部の設定はスポンサー権限が必要です\n\n",
		"settings.chunk_size":      "チャンクサイズ (MB):",
		"settings.concurrency":     "同時接続数:",
		"settings.server":          "アップロードサーバー:",
		"settings.quick_upload":    "クイックアップロード:",
		"settings.language":        "インターフェース言語:",
		"settings.default_server":  "デフォルト",
		"settings.on":              "オン",
		"settings.off":             "オフ",
		"settings.switch_lr":       "(←/→ 切替)",
		"settings.toggle_space":    "(スペース 切替)",
		"settings.read_only":       "(読み取り専用)",
		"settings.chunk_placeholder":       "チャンクサイズ(MB)",
		"settings.concurrency_placeholder": "同時接続数",

		// Upload manager
		"upload_list.title":        "アップロードマネージャー",
		"upload_list.empty":        "アップロードタスクなし",
		"upload_list.col_filename": "ファイル名",
		"upload_list.col_size":     "サイズ",
		"upload_list.col_progress": "進捗",
		"upload_list.col_speed":    "速度",
		"upload_list.col_server":   "サーバー",
		"upload_list.col_status":   "状態",

		// Error screen
		"error.title": "エラー",
		"error.retry": "• Enter: 再試行 • Esc: 戻る",

		// Task statuses
		"task.starting":       "起動中",
		"task.pending":        "待機中",
		"task.uploading":      "アップロード中",
		"task.completed":      "完了",
		"task.failed":         "失敗",
		"task.unknown_server": "不明",

		// Upload check messages
		"upload.in_progress": "ファイルはアップロード中です",
		"upload.completed":   "ファイルはアップロード済みです",
		"upload.in_list":     "ファイルはすでにキューにあります",

		// Error simplification
		"err.invalid_token": "Tokenが無効です。確認して再入力してください",
		"err.timeout":       "ネットワークタイムアウト。接続を確認してください",
		"err.connection":    "サーバーに接続できません。ネットワークを確認してください",
		"err.parse":         "情報を読み取れません。Tokenを確認してください",
		"err.generic":       "Token検証に失敗しました。再入力してください",

		// Misc
		"unknown_state": "不明な状態",
	},

	LangRu: {
		// App
		"app.title": "TMPLINK Загрузчик файлов",

		// Language select screen
		"lang.select_title": "Выбор языка",
		"lang.select_hint":  "↑↓:Выбор  Enter:Подтвердить  Ctrl+C:Выход",

		// Auth / Token input
		"auth.subtitle":     "Аутентификация",
		"auth.validating":   "%s Проверка токена, пожалуйста подождите...",
		"auth.wait":         "Дождитесь завершения проверки...",
		"auth.instructions": "Введите ваш TMPLINK Token для начала работы:\n\n1. Перейдите на https://tmp.link/ и войдите в аккаунт\n2. Нажмите \"Загрузить файл\", затем \"Сбросить\", прокрутите вниз и нажмите \"Использовать CLI\"\n3. Нажмите на Token, чтобы скопировать его",
		"auth.token_label":  "Token:",
		"auth.help":         "💡 Enter: Проверить и сохранить  •  Ctrl+C: Выход",
		"auth.error_box":    "❌ Ошибка проверки\n\n%s",
		"auth.placeholder":  "Введите TMPLINK API Token",

		// Token validation failed
		"auth.failed_title":    "❌ Ошибка проверки",
		"auth.failed_subtitle": "Проблема аутентификации токена",
		"auth.error_details":   "Подробности ошибки:\n\n%s",
		"auth.auto_return":     "%s Возврат к вводу через 3 секунды...",
		"auth.any_key_return":  "💡 Нажмите любую клавишу для немедленного возврата  •  Ctrl+C: Выход",

		// Loading
		"loading.token": "\n%s Проверка токена...",

		// Status bar
		"status.user":          "Пользователь: %s",
		"status.not_logged_in": "Пользователь: Не авторизован",
		"status.sponsor":       " ✨ (Спонсор)",
		"status.regular":       " (Обычный пользователь)",
		"status.storage":       "Хранилище: %.1fГБ/%.1fГБ (%.1f%%)",
		"status.no_storage":    "Хранилище: Нет личного пространства",
		"status.uploading":     " | Загрузка: %d файл(ов)",
		"status.speed_mb":      " (%.1fМБ/с)",
		"status.speed_kb":      " (%.1fКБ/с)",

		// Navigation hint keys
		"nav.enter":            "Открыть",
		"nav.upload":           "Загрузить",
		"nav.keys_with_parent": "↑↓:Выбор ←→:Назад Enter:%s t:Скрытые Tab:Настройки Q:Выход",
		"nav.keys_no_parent":   "↑↓:Выбор Enter:%s t:Скрытые Tab:Настройки Q:Выход",
		"settings.keys":        "↑↓:Выбор Enter:Сохранить Tab:Загрузки Esc:Назад Q:Выход",
		"upload_list.keys":     "↑↓:Выбор d:Удалить t:Очистить y:Удалить всё Tab:Файлы Esc:Назад Q:Выход",
		"error.keys":           "Действия: Enter:Повторить Esc:Назад Q:Выход",
		"default.keys":         "Действия: Q:Выход",

		// Menu
		"menu.title":               "Меню",
		"menu.file_browser":        "Файловый менеджер",
		"menu.file_browser_desc":   "Выбрать файлы для загрузки",
		"menu.settings":            "Настройки загрузки",
		"menu.settings_desc":       "Настроить параметры загрузки",
		"menu.upload_manager":      "Менеджер загрузок",
		"menu.upload_manager_desc": "Просмотр и управление загрузками",

		// File browser
		"filebrowser.title":       "Файловый менеджер",
		"filebrowser.show_hidden": " (Показаны скрытые файлы)",
		"filebrowser.current_dir": "Текущая папка: %s\n",
		"filebrowser.legend":      "📁Папка 📄Файл 🟡Ожидание 🔵Загрузка 🟢Готово 🔴Ошибка\n\n",
		"filebrowser.empty":       "Папка пуста или загружается...",
		"filebrowser.scroll":      "\n[Показано %d-%d из %d элементов]",

		// Settings
		"settings.title":           "Настройки загрузки",
		"settings.sponsor_only":    "✨ Настройки только для спонсоров\n\n",
		"settings.require_sponsor": "⚠️  Некоторые настройки требуют статуса спонсора\n\n",
		"settings.chunk_size":      "Размер фрагмента (МБ):",
		"settings.concurrency":     "Параллельные потоки:",
		"settings.server":          "Сервер загрузки:",
		"settings.quick_upload":    "Быстрая загрузка:",
		"settings.language":        "Язык интерфейса:",
		"settings.default_server":  "По умолчанию",
		"settings.on":              "Вкл",
		"settings.off":             "Выкл",
		"settings.switch_lr":       "(←/→ Переключить)",
		"settings.toggle_space":    "(Пробел Переключить)",
		"settings.read_only":       "(Только чтение)",
		"settings.chunk_placeholder":       "Размер фрагмента (МБ)",
		"settings.concurrency_placeholder": "Параллельные потоки",

		// Upload manager
		"upload_list.title":        "Менеджер загрузок",
		"upload_list.empty":        "Нет активных загрузок",
		"upload_list.col_filename": "Имя файла",
		"upload_list.col_size":     "Размер",
		"upload_list.col_progress": "Прогресс",
		"upload_list.col_speed":    "Скорость",
		"upload_list.col_server":   "Сервер",
		"upload_list.col_status":   "Статус",

		// Error screen
		"error.title": "Ошибка",
		"error.retry": "• Enter: Повторить • Esc: Назад",

		// Task statuses
		"task.starting":       "Запуск",
		"task.pending":        "Ожидание",
		"task.uploading":      "Загрузка",
		"task.completed":      "Завершено",
		"task.failed":         "Ошибка",
		"task.unknown_server": "Неизвестно",

		// Upload check messages
		"upload.in_progress": "Файл уже загружается",
		"upload.completed":   "Файл уже загружен",
		"upload.in_list":     "Файл уже в очереди",

		// Error simplification
		"err.invalid_token": "Недействительный токен, проверьте и введите заново",
		"err.timeout":       "Превышено время ожидания сети, проверьте подключение",
		"err.connection":    "Не удаётся подключиться к серверу, проверьте сеть",
		"err.parse":         "Не удаётся прочитать ответ, проверьте токен",
		"err.generic":       "Ошибка проверки токена, введите заново",

		// Misc
		"unknown_state": "Неизвестное состояние",
	},

	LangZhTW: {
		// App
		"app.title": "鈦盤文件上傳工具",

		// Language select screen
		"lang.select_title": "選擇語言",
		"lang.select_hint":  "↑↓:選擇  Enter:確認  Ctrl+C:退出",

		// Auth / Token input
		"auth.subtitle":     "安全認證",
		"auth.validating":   "%s 正在驗證Token，請稍候...",
		"auth.wait":         "請耐心等待驗證完成...",
		"auth.instructions": "請輸入您的鈦盤 Token 以開始使用：\n\n1. 訪問 https://tmp.link/ 並登入您的帳戶\n2. 點擊「上傳文件」按鈕，再點擊「重新設定」，滑動到頁面底部，點擊「使用 CLI 上傳」\n3. 點擊 Token 以複製到剪貼簿",
		"auth.token_label":  "Token:",
		"auth.help":         "💡 Enter: 驗證並儲存  •  Ctrl+C: 退出程式",
		"auth.error_box":    "❌ 驗證失敗\n\n%s",
		"auth.placeholder":  "請輸入鈦盤 API Token",

		// Token validation failed
		"auth.failed_title":    "❌ 驗證失敗",
		"auth.failed_subtitle": "Token 認證出現問題",
		"auth.error_details":   "錯誤詳情：\n\n%s",
		"auth.auto_return":     "%s 3秒後自動返回輸入介面...",
		"auth.any_key_return":  "💡 按任意鍵立即返回  •  Ctrl+C: 退出程式",

		// Loading
		"loading.token": "\n%s 正在驗證Token...",

		// Status bar
		"status.user":          "用戶: %s",
		"status.not_logged_in": "用戶: 未登入",
		"status.sponsor":       " ✨ (贊助者)",
		"status.regular":       " (一般用戶)",
		"status.storage":       "儲存空間: %.1fGB/%.1fGB (%.1f%%)",
		"status.no_storage":    "儲存空間: 無私人空間",
		"status.uploading":     " | 上傳中: %d個文件",
		"status.speed_mb":      " (%.1fMB/s)",
		"status.speed_kb":      " (%.1fKB/s)",

		// Navigation hint keys
		"nav.enter":            "進入",
		"nav.upload":           "上傳",
		"nav.keys_with_parent": "↑↓:選擇 ←→:上層 Enter:%s t:隱藏文件 Tab:設定 Q:退出",
		"nav.keys_no_parent":   "↑↓:選擇 Enter:%s t:隱藏文件 Tab:設定 Q:退出",
		"settings.keys":        "↑↓:選擇 Enter:儲存 Tab:上傳管理 Esc:返回 Q:退出",
		"upload_list.keys":     "↑↓:選擇 d:刪除 t:清除完成 y:清除全部 Tab:文件瀏覽 Esc:返回 Q:退出",
		"error.keys":           "操作: Enter:重試 Esc:返回 Q:退出",
		"default.keys":         "操作: Q:退出",

		// Menu
		"menu.title":               "功能選單",
		"menu.file_browser":        "文件瀏覽器",
		"menu.file_browser_desc":   "選擇要上傳的文件",
		"menu.settings":            "上傳設定",
		"menu.settings_desc":       "配置上傳參數",
		"menu.upload_manager":      "上傳管理器",
		"menu.upload_manager_desc": "查看和管理上傳任務",

		// File browser
		"filebrowser.title":       "文件瀏覽器",
		"filebrowser.show_hidden": " (顯示隱藏文件)",
		"filebrowser.current_dir": "當前目錄: %s\n",
		"filebrowser.legend":      "📁目錄 📄文件 🟡等待 🔵上傳中 🟢已完成 🔴失敗\n\n",
		"filebrowser.empty":       "目錄為空或正在載入...",
		"filebrowser.scroll":      "\n[顯示 %d-%d / 共 %d 項]",

		// Settings
		"settings.title":           "上傳設定",
		"settings.sponsor_only":    "✨ 贊助者專屬設定\n\n",
		"settings.require_sponsor": "⚠️  部分設定需要贊助者權限\n\n",
		"settings.chunk_size":      "分塊大小 (MB):",
		"settings.concurrency":     "並發數:",
		"settings.server":          "上傳伺服器:",
		"settings.quick_upload":    "快速上傳:",
		"settings.language":        "介面語言:",
		"settings.default_server":  "預設",
		"settings.on":              "開啟",
		"settings.off":             "關閉",
		"settings.switch_lr":       "(←/→ 切換)",
		"settings.toggle_space":    "(Space 切換)",
		"settings.read_only":       "(唯讀)",
		"settings.chunk_placeholder":       "分塊大小(MB)",
		"settings.concurrency_placeholder": "並發數",

		// Upload manager
		"upload_list.title":        "上傳管理器",
		"upload_list.empty":        "暫無上傳任務",
		"upload_list.col_filename": "文件名",
		"upload_list.col_size":     "大小",
		"upload_list.col_progress": "進度",
		"upload_list.col_speed":    "速度",
		"upload_list.col_server":   "伺服器",
		"upload_list.col_status":   "狀態",

		// Error screen
		"error.title": "錯誤",
		"error.retry": "• Enter: 重試 • Esc: 返回",

		// Task statuses
		"task.starting":       "啟動中",
		"task.pending":        "等待中",
		"task.uploading":      "上傳中",
		"task.completed":      "已完成",
		"task.failed":         "失敗",
		"task.unknown_server": "未知",

		// Upload check messages
		"upload.in_progress": "文件正在上傳中",
		"upload.completed":   "文件已上傳完成",
		"upload.in_list":     "文件已在上傳清單中",

		// Error simplification
		"err.invalid_token": "Token無效，請檢查後重新輸入",
		"err.timeout":       "網路連線逾時，請檢查網路後重試",
		"err.connection":    "無法連線到伺服器，請檢查網路連線",
		"err.parse":         "無法讀取到資訊，請檢查Token是否正確",
		"err.generic":       "Token驗證失敗，請重新輸入",

		// Misc
		"unknown_state": "未知狀態",
	},

	LangFr: {
		// App
		"app.title": "TMPLINK Gestionnaire de fichiers",

		// Language select screen
		"lang.select_title": "Choisir la langue",
		"lang.select_hint":  "↑↓:Sélectionner  Enter:Confirmer  Ctrl+C:Quitter",

		// Auth / Token input
		"auth.subtitle":     "Authentification",
		"auth.validating":   "%s Validation du token, veuillez patienter...",
		"auth.wait":         "Veuillez attendre la fin de la validation...",
		"auth.instructions": "Entrez votre Token TMPLINK pour commencer :\n\n1. Rendez-vous sur https://tmp.link/ et connectez-vous\n2. Cliquez sur \"Envoyer un fichier\", puis \"Réinitialiser\", faites défiler vers le bas et cliquez sur \"Utiliser le CLI\"\n3. Cliquez sur le Token pour le copier",
		"auth.token_label":  "Token :",
		"auth.help":         "💡 Entrée : Vérifier et enregistrer  •  Ctrl+C : Quitter",
		"auth.error_box":    "❌ Échec de la vérification\n\n%s",
		"auth.placeholder":  "Entrez votre Token API TMPLINK",

		// Token validation failed
		"auth.failed_title":    "❌ Échec de la vérification",
		"auth.failed_subtitle": "Problème d'authentification du token",
		"auth.error_details":   "Détails de l'erreur :\n\n%s",
		"auth.auto_return":     "%s Retour à la saisie dans 3 secondes...",
		"auth.any_key_return":  "💡 Appuyez sur une touche pour revenir immédiatement  •  Ctrl+C : Quitter",

		// Loading
		"loading.token": "\n%s Validation du token...",

		// Status bar
		"status.user":          "Utilisateur : %s",
		"status.not_logged_in": "Utilisateur : Non connecté",
		"status.sponsor":       " ✨ (Sponsor)",
		"status.regular":       " (Utilisateur gratuit)",
		"status.storage":       "Stockage : %.1fGo/%.1fGo (%.1f%%)",
		"status.no_storage":    "Stockage : Aucun espace privé",
		"status.uploading":     " | Envoi : %d fichier(s)",
		"status.speed_mb":      " (%.1fMo/s)",
		"status.speed_kb":      " (%.1fKo/s)",

		// Navigation hint keys
		"nav.enter":            "Ouvrir",
		"nav.upload":           "Envoyer",
		"nav.keys_with_parent": "↑↓:Sélect ←→:Parent Entrée:%s t:Cachés Tab:Param Q:Quitter",
		"nav.keys_no_parent":   "↑↓:Sélect Entrée:%s t:Cachés Tab:Param Q:Quitter",
		"settings.keys":        "↑↓:Sélect Entrée:Sauv Tab:Envois Échap:Retour Q:Quitter",
		"upload_list.keys":     "↑↓:Sélect d:Supp t:Vider y:Tout supp Tab:Fichiers Échap:Retour Q:Quitter",
		"error.keys":           "Actions : Entrée:Réessayer Échap:Retour Q:Quitter",
		"default.keys":         "Actions : Q:Quitter",

		// Menu
		"menu.title":               "Menu",
		"menu.file_browser":        "Explorateur de fichiers",
		"menu.file_browser_desc":   "Sélectionner des fichiers à envoyer",
		"menu.settings":            "Paramètres d'envoi",
		"menu.settings_desc":       "Configurer les paramètres d'envoi",
		"menu.upload_manager":      "Gestionnaire d'envois",
		"menu.upload_manager_desc": "Voir et gérer les envois",

		// File browser
		"filebrowser.title":       "Explorateur de fichiers",
		"filebrowser.show_hidden": " (Fichiers cachés visibles)",
		"filebrowser.current_dir": "Dossier courant : %s\n",
		"filebrowser.legend":      "📁Dossier 📄Fichier 🟡Attente 🔵Envoi 🟢Terminé 🔴Erreur\n\n",
		"filebrowser.empty":       "Dossier vide ou chargement...",
		"filebrowser.scroll":      "\n[Affichage %d-%d sur %d éléments]",

		// Settings
		"settings.title":           "Paramètres d'envoi",
		"settings.sponsor_only":    "✨ Paramètres réservés aux sponsors\n\n",
		"settings.require_sponsor": "⚠️  Certains paramètres nécessitent le statut sponsor\n\n",
		"settings.chunk_size":      "Taille des fragments (Mo) :",
		"settings.concurrency":     "Connexions simultanées :",
		"settings.server":          "Serveur d'envoi :",
		"settings.quick_upload":    "Envoi rapide :",
		"settings.language":        "Langue de l'interface :",
		"settings.default_server":  "Par défaut",
		"settings.on":              "Activé",
		"settings.off":             "Désactivé",
		"settings.switch_lr":       "(←/→ Changer)",
		"settings.toggle_space":    "(Espace Basculer)",
		"settings.read_only":       "(Lecture seule)",
		"settings.chunk_placeholder":       "Taille des fragments (Mo)",
		"settings.concurrency_placeholder": "Connexions simultanées",

		// Upload manager
		"upload_list.title":        "Gestionnaire d'envois",
		"upload_list.empty":        "Aucun envoi en cours",
		"upload_list.col_filename": "Nom du fichier",
		"upload_list.col_size":     "Taille",
		"upload_list.col_progress": "Progression",
		"upload_list.col_speed":    "Vitesse",
		"upload_list.col_server":   "Serveur",
		"upload_list.col_status":   "Statut",

		// Error screen
		"error.title": "Erreur",
		"error.retry": "• Entrée : Réessayer • Échap : Retour",

		// Task statuses
		"task.starting":       "Démarrage",
		"task.pending":        "En attente",
		"task.uploading":      "Envoi en cours",
		"task.completed":      "Terminé",
		"task.failed":         "Échec",
		"task.unknown_server": "Inconnu",

		// Upload check messages
		"upload.in_progress": "Le fichier est en cours d'envoi",
		"upload.completed":   "Le fichier a déjà été envoyé",
		"upload.in_list":     "Le fichier est déjà dans la file d'attente",

		// Error simplification
		"err.invalid_token": "Token invalide, vérifiez et ressaisissez",
		"err.timeout":       "Délai réseau dépassé, vérifiez votre connexion",
		"err.connection":    "Impossible de se connecter au serveur, vérifiez le réseau",
		"err.parse":         "Impossible de lire la réponse, vérifiez le token",
		"err.generic":       "Échec de la validation du token, ressaisissez",

		// Misc
		"unknown_state": "État inconnu",
	},

	LangMs: {
		// App
		"app.title": "TMPLINK Pemuat Naik Fail",

		// Language select screen
		"lang.select_title": "Pilih Bahasa",
		"lang.select_hint":  "↑↓:Pilih  Enter:Sahkan  Ctrl+C:Keluar",

		// Auth / Token input
		"auth.subtitle":     "Pengesahan",
		"auth.validating":   "%s Mengesahkan token, sila tunggu...",
		"auth.wait":         "Sila tunggu sehingga pengesahan selesai...",
		"auth.instructions": "Masukkan Token TMPLINK anda untuk memulakan:\n\n1. Layari https://tmp.link/ dan log masuk\n2. Klik \"Muat Naik Fail\", kemudian \"Set Semula\", tatal ke bawah dan klik \"Guna CLI\"\n3. Klik Token untuk menyalinnya",
		"auth.token_label":  "Token:",
		"auth.help":         "💡 Enter: Sahkan & Simpan  •  Ctrl+C: Keluar",
		"auth.error_box":    "❌ Pengesahan Gagal\n\n%s",
		"auth.placeholder":  "Masukkan Token API TMPLINK anda",

		// Token validation failed
		"auth.failed_title":    "❌ Pengesahan Gagal",
		"auth.failed_subtitle": "Masalah pengesahan token",
		"auth.error_details":   "Butiran ralat:\n\n%s",
		"auth.auto_return":     "%s Kembali ke skrin input dalam 3 saat...",
		"auth.any_key_return":  "💡 Tekan mana-mana kekunci untuk kembali segera  •  Ctrl+C: Keluar",

		// Loading
		"loading.token": "\n%s Mengesahkan token...",

		// Status bar
		"status.user":          "Pengguna: %s",
		"status.not_logged_in": "Pengguna: Belum log masuk",
		"status.sponsor":       " ✨ (Penaja)",
		"status.regular":       " (Pengguna Biasa)",
		"status.storage":       "Storan: %.1fGB/%.1fGB (%.1f%%)",
		"status.no_storage":    "Storan: Tiada ruang peribadi",
		"status.uploading":     " | Memuat naik: %d fail",
		"status.speed_mb":      " (%.1fMB/s)",
		"status.speed_kb":      " (%.1fKB/s)",

		// Navigation hint keys
		"nav.enter":            "Buka",
		"nav.upload":           "Muat Naik",
		"nav.keys_with_parent": "↑↓:Pilih ←→:Induk Enter:%s t:Tersembunyi Tab:Tetapan Q:Keluar",
		"nav.keys_no_parent":   "↑↓:Pilih Enter:%s t:Tersembunyi Tab:Tetapan Q:Keluar",
		"settings.keys":        "↑↓:Pilih Enter:Simpan Tab:Muat Naik Esc:Kembali Q:Keluar",
		"upload_list.keys":     "↑↓:Pilih d:Padam t:Bersih y:Padam Semua Tab:Fail Esc:Kembali Q:Keluar",
		"error.keys":           "Tindakan: Enter:Cuba Lagi Esc:Kembali Q:Keluar",
		"default.keys":         "Tindakan: Q:Keluar",

		// Menu
		"menu.title":               "Menu",
		"menu.file_browser":        "Pelayar Fail",
		"menu.file_browser_desc":   "Pilih fail untuk dimuat naik",
		"menu.settings":            "Tetapan Muat Naik",
		"menu.settings_desc":       "Konfigurasi parameter muat naik",
		"menu.upload_manager":      "Pengurus Muat Naik",
		"menu.upload_manager_desc": "Lihat dan urus tugas muat naik",

		// File browser
		"filebrowser.title":       "Pelayar Fail",
		"filebrowser.show_hidden": " (Menunjukkan fail tersembunyi)",
		"filebrowser.current_dir": "Direktori semasa: %s\n",
		"filebrowser.legend":      "📁Folder 📄Fail 🟡Menunggu 🔵Memuat Naik 🟢Selesai 🔴Gagal\n\n",
		"filebrowser.empty":       "Direktori kosong atau sedang dimuatkan...",
		"filebrowser.scroll":      "\n[Menunjukkan %d-%d daripada %d item]",

		// Settings
		"settings.title":           "Tetapan Muat Naik",
		"settings.sponsor_only":    "✨ Tetapan eksklusif penaja\n\n",
		"settings.require_sponsor": "⚠️  Sesetengah tetapan memerlukan status penaja\n\n",
		"settings.chunk_size":      "Saiz Serpihan (MB):",
		"settings.concurrency":     "Sambungan Serentak:",
		"settings.server":          "Pelayan Muat Naik:",
		"settings.quick_upload":    "Muat Naik Pantas:",
		"settings.language":        "Bahasa Antara Muka:",
		"settings.default_server":  "Lalai",
		"settings.on":              "Hidup",
		"settings.off":             "Mati",
		"settings.switch_lr":       "(←/→ Tukar)",
		"settings.toggle_space":    "(Ruang Togol)",
		"settings.read_only":       "(Baca Sahaja)",
		"settings.chunk_placeholder":       "Saiz serpihan (MB)",
		"settings.concurrency_placeholder": "Sambungan serentak",

		// Upload manager
		"upload_list.title":        "Pengurus Muat Naik",
		"upload_list.empty":        "Tiada tugas muat naik",
		"upload_list.col_filename": "Nama Fail",
		"upload_list.col_size":     "Saiz",
		"upload_list.col_progress": "Kemajuan",
		"upload_list.col_speed":    "Kelajuan",
		"upload_list.col_server":   "Pelayan",
		"upload_list.col_status":   "Status",

		// Error screen
		"error.title": "Ralat",
		"error.retry": "• Enter: Cuba Lagi • Esc: Kembali",

		// Task statuses
		"task.starting":       "Bermula",
		"task.pending":        "Menunggu",
		"task.uploading":      "Memuat Naik",
		"task.completed":      "Selesai",
		"task.failed":         "Gagal",
		"task.unknown_server": "Tidak Diketahui",

		// Upload check messages
		"upload.in_progress": "Fail sedang dimuat naik",
		"upload.completed":   "Fail telah dimuat naik",
		"upload.in_list":     "Fail sudah dalam baris gilir",

		// Error simplification
		"err.invalid_token": "Token tidak sah, semak dan masukkan semula",
		"err.timeout":       "Tamat masa rangkaian, semak sambungan anda",
		"err.connection":    "Tidak dapat menyambung ke pelayan, semak rangkaian",
		"err.parse":         "Tidak dapat membaca respons, semak token anda",
		"err.generic":       "Pengesahan token gagal, masukkan semula",

		// Misc
		"unknown_state": "Keadaan tidak diketahui",
	},
}

// current holds the active language code
var current = DefaultLanguage

// SetLanguage sets the active language. Falls back to DefaultLanguage for unknown codes.
func SetLanguage(lang string) {
	if _, ok := translations[lang]; ok {
		current = lang
	} else {
		current = DefaultLanguage
	}
}

// GetLanguage returns the current active language code.
func GetLanguage() string {
	return current
}

// T returns the translation for key in the current language.
// Falls back to zh-CN if the key is missing in the current language.
func T(key string) string {
	if lang, ok := translations[current]; ok {
		if val, ok := lang[key]; ok {
			return val
		}
	}
	// fallback to zh-CN
	if val, ok := translations[DefaultLanguage][key]; ok {
		return val
	}
	return key
}

// Tf is a shorthand for fmt.Sprintf(T(key), args...)
func Tf(key string, args ...interface{}) string {
	return fmt.Sprintf(T(key), args...)
}
