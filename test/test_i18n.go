package main

import (
	"fmt"
	"os"
	"tmplink_uploader/internal/i18n"
)

func main() {
	// Get language from command line argument or use auto-detection
	lang := ""
	if len(os.Args) > 1 {
		lang = os.Args[1]
	}
	
	// Initialize i18n
	i18n.InitLanguage(i18n.Language(lang))
	
	// Display the currently active language
	currentLang := i18n.GetCurrentLanguage()
	fmt.Printf("Active language: %s\n\n", currentLang)
	
	// Test various localized strings
	fmt.Println("=== Common Strings ===")
	fmt.Printf("App name: %s\n", i18n.T("app_name"))
	fmt.Printf("Loading: %s\n", i18n.T("loading"))
	fmt.Printf("Error: %s\n", i18n.T("error"))
	fmt.Printf("Success: %s\n", i18n.T("success"))
	
	fmt.Println("\n=== User Interface ===")
	fmt.Printf("Enter API Token: %s\n", i18n.T("enter_api_token"))
	fmt.Printf("Menu title: %s\n", i18n.T("menu_title"))
	fmt.Printf("File browser: %s\n", i18n.T("menu_file_browser"))
	fmt.Printf("Upload settings: %s\n", i18n.T("menu_upload_settings"))
	fmt.Printf("Upload manager: %s\n", i18n.T("menu_upload_manager"))
	
	fmt.Println("\n=== User Information ===")
	fmt.Printf("User info: %s\n", i18n.T("user_info", "TestUser"))
	fmt.Printf("User sponsored: %s\n", i18n.T("user_sponsored"))
	fmt.Printf("User regular: %s\n", i18n.T("user_regular"))
	fmt.Printf("Not logged in: %s\n", i18n.T("user_not_logged_in"))
	
	fmt.Println("\n=== Table Columns ===")
	fmt.Printf("Filename: %s\n", i18n.T("column_filename"))
	fmt.Printf("Size: %s\n", i18n.T("column_size"))
	fmt.Printf("Progress: %s\n", i18n.T("column_progress"))
	fmt.Printf("Speed: %s\n", i18n.T("column_speed"))
	fmt.Printf("Server: %s\n", i18n.T("column_server"))
	fmt.Printf("Status: %s\n", i18n.T("column_status"))
	
	fmt.Println("\n=== Status Messages ===")
	fmt.Printf("Starting: %s\n", i18n.T("status_starting"))
	fmt.Printf("Waiting: %s\n", i18n.T("status_waiting"))
	fmt.Printf("Uploading: %s\n", i18n.T("status_uploading"))
	fmt.Printf("Completed: %s\n", i18n.T("status_completed"))
	fmt.Printf("Failed: %s\n", i18n.T("status_failed"))
	
	fmt.Println("\n=== Upload Information ===")
	fmt.Printf("Upload complete: %s\n", i18n.T("upload_complete"))
	fmt.Printf("Upload failed: %s\n", i18n.T("upload_failed"))
	fmt.Printf("File name: %s\n", i18n.T("upload_file_name", "test.txt"))
	fmt.Printf("Error message: %s\n", i18n.T("upload_error_message", "Test error"))
	fmt.Printf("File size: %s\n", i18n.T("upload_file_size", "10MB"))
	fmt.Printf("Average speed: %s\n", i18n.T("upload_average_speed", 1.5))
	fmt.Printf("Total time: %s\n", i18n.T("upload_total_time", "1m30s"))
	fmt.Printf("Download link: %s\n", i18n.T("upload_download_link", "https://example.com"))
	
	fmt.Println("\n=== Settings ===")
	fmt.Printf("Settings title: %s\n", i18n.T("settings_title"))
	fmt.Printf("Sponsored only: %s\n", i18n.T("settings_sponsored_only"))
	fmt.Printf("Some sponsored: %s\n", i18n.T("settings_some_sponsored"))
	fmt.Printf("Chunk size: %s\n", i18n.T("chunk_size_mb"))
	fmt.Printf("Concurrency: %s\n", i18n.T("concurrency"))
	fmt.Printf("Upload server: %s\n", i18n.T("upload_server"))
	fmt.Printf("Quick upload: %s\n", i18n.T("quick_upload"))
	
	fmt.Println("\n=== Error Messages ===")
	fmt.Printf("Missing file: %s\n", i18n.T("error_missing_file"))
	fmt.Printf("Token not found: %s\n", i18n.T("error_token_not_found"))
	fmt.Printf("Chunk size range: %s\n", i18n.T("error_chunk_size_range", 100))
	fmt.Printf("Token validation: %s\n", i18n.T("token_validation_error", "Invalid token"))
}