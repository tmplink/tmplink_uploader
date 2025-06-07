package i18n

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// Language represents a supported language
type Language string

const (
	// LanguageCN represents Simplified Chinese
	LanguageCN Language = "cn"
	// LanguageEN represents English
	LanguageEN Language = "en"
	// LanguageHK represents Traditional Chinese (Hong Kong)
	LanguageHK Language = "hk"
	// LanguageJP represents Japanese
	LanguageJP Language = "jp"
)

// SupportedLanguages contains all supported languages
var SupportedLanguages = []Language{
	LanguageCN,
	LanguageEN,
	LanguageHK,
	LanguageJP,
}

// DefaultLanguage is the fallback language
const DefaultLanguage = LanguageCN

var (
	currentLanguage Language = DefaultLanguage
	languageMutex   sync.RWMutex
)

// messages stores all translation templates
var messages = map[Language]map[string]string{}

// GetCurrentLanguage returns the currently set language
func GetCurrentLanguage() Language {
	languageMutex.RLock()
	defer languageMutex.RUnlock()
	return currentLanguage
}

// SetLanguage changes the current language
func SetLanguage(lang Language) bool {
	for _, supportedLang := range SupportedLanguages {
		if supportedLang == lang {
			languageMutex.Lock()
			currentLanguage = lang
			languageMutex.Unlock()
			return true
		}
	}
	return false
}

// DetectSystemLanguage attempts to detect the system language
// and returns a supported language closest to the system language
func DetectSystemLanguage() Language {
	langEnv := os.Getenv("LANG")
	if langEnv == "" {
		return DefaultLanguage
	}

	langEnv = strings.ToLower(langEnv)
	
	// Check for Chinese variants
	if strings.HasPrefix(langEnv, "zh_cn") || 
	   strings.HasPrefix(langEnv, "zh-cn") || 
	   strings.HasPrefix(langEnv, "zh_hans") {
		return LanguageCN
	}
	
	if strings.HasPrefix(langEnv, "zh_tw") || 
	   strings.HasPrefix(langEnv, "zh-tw") || 
	   strings.HasPrefix(langEnv, "zh_hk") || 
	   strings.HasPrefix(langEnv, "zh-hk") || 
	   strings.HasPrefix(langEnv, "zh_hant") {
		return LanguageHK
	}
	
	// Check for English
	if strings.HasPrefix(langEnv, "en") {
		return LanguageEN
	}
	
	// Check for Japanese
	if strings.HasPrefix(langEnv, "ja") || strings.HasPrefix(langEnv, "jp") {
		return LanguageJP
	}
	
	// Partial match for Chinese as fallback
	if strings.HasPrefix(langEnv, "zh") {
		return LanguageCN
	}
	
	return DefaultLanguage
}

// InitLanguage initializes the i18n system and sets the language
func InitLanguage(lang Language) {
	// Initialize message maps
	initCN() // Simplified Chinese
	initEN() // English
	initHK() // Traditional Chinese (Hong Kong)
	initJP() // Japanese
	
	// Set language (falls back to default if not supported)
	if lang == "" {
		lang = DetectSystemLanguage()
	}
	
	SetLanguage(lang)
}

// T translates a message key to the current language
func T(key string, args ...interface{}) string {
	languageMutex.RLock()
	lang := currentLanguage
	languageMutex.RUnlock()
	
	langMessages, ok := messages[lang]
	if !ok {
		langMessages = messages[DefaultLanguage]
	}
	
	msg, ok := langMessages[key]
	if !ok {
		// If not found in current language, try default language
		if lang != DefaultLanguage {
			msg, ok = messages[DefaultLanguage][key]
			if !ok {
				return key // Return key as fallback
			}
		} else {
			return key // Return key as fallback
		}
	}
	
	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}