package i18n

// ModelToString converts a model number to a localized string description
func ModelToString(model int) string {
	switch model {
	case 0:
		return T("model_24h")
	case 1:
		return T("model_3d")
	case 2:
		return T("model_7d")
	case 99:
		return T("model_unlimited")
	default:
		return T("unknown")
	}
}

// DebugModeToString converts a debug mode (bool) to a localized string
func DebugModeToString(debugMode bool) string {
	if debugMode {
		return T("config_debug_on")
	}
	return T("config_debug_off")
}

// BoolToEnabledDisabled converts a boolean to enabled/disabled string
func BoolToEnabledDisabled(value bool) string {
	if value {
		return T("enabled")
	}
	return T("disabled")
}

// BoolToYesNo converts a boolean to yes/no string
func BoolToYesNo(value bool) string {
	if value {
		return T("yes")
	}
	return T("no")
}