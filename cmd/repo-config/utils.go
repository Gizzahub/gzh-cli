package repoconfig

// getActionSymbol returns the symbol for action type.
func getActionSymbol(changeType string) string {
	switch changeType {
	case "create":
		return "➕"
	case "update":
		return "🔄"
	case "delete":
		return "➖"
	default:
		return "📝"
	}
}

// getActionSymbolWithText returns the symbol with text for action type.
func getActionSymbolWithText(changeType string) string {
	switch changeType {
	case "create":
		return "➕ Create"
	case "update":
		return "🔄 Update"
	case "delete":
		return "➖ Delete"
	default:
		return "❓ Unknown"
	}
}

// truncateString truncates a string to the specified length.
func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}

	if maxLength <= 3 {
		return s[:maxLength]
	}

	return s[:maxLength-3] + "..."
}
