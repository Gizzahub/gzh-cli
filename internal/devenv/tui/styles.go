// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette inspired by Go's branding and modern terminal themes
var (
	ColorPrimary    = lipgloss.Color("#00ADD8") // Go blue
	ColorSecondary  = lipgloss.Color("#5E81AC") // Muted blue
	ColorSuccess    = lipgloss.Color("#A3BE8C") // Green
	ColorWarning    = lipgloss.Color("#EBCB8B") // Yellow
	ColorError      = lipgloss.Color("#BF616A") // Red
	ColorText       = lipgloss.Color("#D8DEE9") // Light gray
	ColorSubtle     = lipgloss.Color("#4C566A") // Dark gray
	ColorBackground = lipgloss.Color("#2E3440") // Dark background
	ColorBorder     = lipgloss.Color("#434C5E") // Border color
	ColorHighlight  = lipgloss.Color("#88C0D0") // Highlight color
)

// Base styles
var (
	BaseStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorBackground)

	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			Padding(0, 1).
			Margin(0, 0, 1, 0)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorBorder).
			Padding(0, 1)

	FooterStyle = lipgloss.NewStyle().
			Foreground(ColorSubtle).
			Padding(0, 1)
)

// Service status styles
var (
	ServiceActiveStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Bold(true)

	ServiceInactiveStyle = lipgloss.NewStyle().
				Foreground(ColorSubtle)

	ServiceWarningStyle = lipgloss.NewStyle().
				Foreground(ColorWarning).
				Bold(true)

	ServiceErrorStyle = lipgloss.NewStyle().
				Foreground(ColorError).
				Bold(true)

	ServiceUnknownStyle = lipgloss.NewStyle().
				Foreground(ColorSubtle)
)

// Table styles
var (
	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				Padding(0, 1).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(ColorBorder)

	TableCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	TableSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorBackground).
				Background(ColorHighlight).
				Bold(true).
				Padding(0, 1)

	TableEvenRowStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#3B4252"))

	TableOddRowStyle = lipgloss.NewStyle().
				Background(ColorBackground)
)

// Button and interactive element styles
var (
	ButtonStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorBorder).
			Padding(0, 2).
			Margin(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder)

	ButtonActiveStyle = lipgloss.NewStyle().
				Foreground(ColorBackground).
				Background(ColorPrimary).
				Padding(0, 2).
				Margin(0, 1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Bold(true)

	ButtonHoverStyle = lipgloss.NewStyle().
				Foreground(ColorText).
				Background(ColorHighlight).
				Padding(0, 2).
				Margin(0, 1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorHighlight)
)

// Dialog and modal styles
var (
	DialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2).
			Background(ColorBackground).
			Foreground(ColorText)

	DialogTitleStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				Align(lipgloss.Center).
				Margin(0, 0, 1, 0)

	DialogContentStyle = lipgloss.NewStyle().
				Foreground(ColorText).
				Margin(0, 0, 1, 0)
)

// Progress and loading styles
var (
	ProgressBarStyle = lipgloss.NewStyle().
				Background(ColorBorder).
				Foreground(ColorPrimary)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary)

	LoadingStyle = lipgloss.NewStyle().
			Foreground(ColorSubtle).
			Italic(true)
)

// Message and notification styles
var (
	InfoStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)
)

// Help styles
var (
	HelpHeaderStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			Margin(1, 0)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorHighlight).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	HelpSectionStyle = lipgloss.NewStyle().
				Margin(0, 0, 1, 2)
)

// Border styles
var (
	NormalBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
	}

	ThickBorder = lipgloss.Border{
		Top:         "━",
		Bottom:      "━",
		Left:        "┃",
		Right:       "┃",
		TopLeft:     "┏",
		TopRight:    "┓",
		BottomLeft:  "┗",
		BottomRight: "┛",
	}
)

// GetServiceStatusStyle returns the appropriate style for a service status
func GetServiceStatusStyle(status string) lipgloss.Style {
	switch status {
	case "active", "connected", "running", "online":
		return ServiceActiveStyle
	case "inactive", "disconnected", "stopped", "offline":
		return ServiceInactiveStyle
	case "warning", "degraded", "partial":
		return ServiceWarningStyle
	case "error", "failed", "critical":
		return ServiceErrorStyle
	default:
		return ServiceUnknownStyle
	}
}

// GetStatusIcon returns the appropriate icon for a service status
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

// AdaptiveStyle adjusts styles based on terminal capabilities
func AdaptiveStyle(width, height int) lipgloss.Style {
	base := BaseStyle

	// Adjust for smaller terminals
	if width < 80 {
		base = base.Padding(0)
	}

	if height < 24 {
		base = base.Margin(0)
	}

	return base
}
