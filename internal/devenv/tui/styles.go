// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package tui

import (
	"github.com/Gizzahub/gzh-cli/internal/tui/common"
	"github.com/charmbracelet/lipgloss"
)

// DevEnvStyles holds the styles for development environment TUI.
type DevEnvStyles struct {
	common.StyleSet
	ServiceActive   lipgloss.Style
	ServiceInactive lipgloss.Style
	ServiceWarning  lipgloss.Style
	ServiceError    lipgloss.Style
	ServiceUnknown  lipgloss.Style
	TableHeader     lipgloss.Style
	TableCell       lipgloss.Style
	TableSelected   lipgloss.Style
	TableEvenRow    lipgloss.Style
	TableOddRow     lipgloss.Style
}

// NewDevEnvStyles creates a new set of styles for development environment TUI.
func NewDevEnvStyles() DevEnvStyles {
	baseStyles := common.DefaultStyles()

	return DevEnvStyles{
		StyleSet: baseStyles,
		ServiceActive: lipgloss.NewStyle().
			Foreground(baseStyles.Theme.Success).
			Bold(true),
		ServiceInactive: lipgloss.NewStyle().
			Foreground(baseStyles.Theme.Subtle),
		ServiceWarning: lipgloss.NewStyle().
			Foreground(baseStyles.Theme.Warning).
			Bold(true),
		ServiceError: lipgloss.NewStyle().
			Foreground(baseStyles.Theme.Error).
			Bold(true),
		ServiceUnknown: lipgloss.NewStyle().
			Foreground(baseStyles.Theme.Subtle),
		TableHeader: lipgloss.NewStyle().
			Foreground(baseStyles.Theme.Primary).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(baseStyles.Theme.Border),
		TableCell: lipgloss.NewStyle().
			Padding(0, 1),
		TableSelected: lipgloss.NewStyle().
			Foreground(baseStyles.Theme.Background).
			Background(baseStyles.Theme.Highlight).
			Bold(true).
			Padding(0, 1),
		TableEvenRow: lipgloss.NewStyle().
			Background(lipgloss.Color("#3B4252")),
		TableOddRow: lipgloss.NewStyle().
			Background(baseStyles.Theme.Background),
	}
}

// Legacy style variables for backward compatibility.
var (
	styles = NewDevEnvStyles()

	// Backward compatibility exports
	ColorPrimary    = styles.Theme.Primary
	ColorSecondary  = styles.Theme.Secondary
	ColorSuccess    = styles.Theme.Success
	ColorWarning    = styles.Theme.Warning
	ColorError      = styles.Theme.Error
	ColorText       = styles.Theme.Text
	ColorSubtle     = styles.Theme.Subtle
	ColorBackground = styles.Theme.Background
	ColorBorder     = styles.Theme.Border
	ColorHighlight  = styles.Theme.Highlight

	// Base styles
	BaseStyle      = styles.Base
	TitleStyle     = styles.Title
	HeaderStyle    = styles.Header
	StatusBarStyle = styles.StatusBar
	FooterStyle    = styles.Footer

	// Service status styles
	ServiceActiveStyle   = styles.ServiceActive
	ServiceInactiveStyle = styles.ServiceInactive
	ServiceWarningStyle  = styles.ServiceWarning
	ServiceErrorStyle    = styles.ServiceError
	ServiceUnknownStyle  = styles.ServiceUnknown

	// Table styles
	TableHeaderStyle   = styles.TableHeader
	TableCellStyle     = styles.TableCell
	TableSelectedStyle = styles.TableSelected
	TableEvenRowStyle  = styles.TableEvenRow
	TableOddRowStyle   = styles.TableOddRow

	// Additional styles for compatibility
	SpinnerStyle    = styles.Base.Foreground(styles.Theme.Primary)
	ErrorStyle      = styles.Base.Foreground(styles.Theme.Error).Bold(true)
	InfoStyle       = styles.Base.Foreground(styles.Theme.Primary).Bold(true)
	HelpHeaderStyle = styles.Base.Foreground(styles.Theme.Primary).Bold(true).Margin(1, 0)
)

// GetStatusIcon returns the appropriate icon for a service status.
func GetStatusIcon(status string) string {
	switch status {
	case "active", "connected", "running", "online":
		return "✅"
	case "inactive", "disconnected", "stopped", "offline":
		return "❌"
	case "warning", "degraded", "partial":
		return "⚠️"
	case "error", "failed", "critical":
		return "🔴"
	default:
		return "❓"
	}
}
