// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Status constants.
const (
	StatusActive       = "active"
	StatusConnected    = "connected"
	StatusEnabled      = "enabled"
	StatusCustom       = "custom"
	StatusDisabled     = "disabled"
	StatusInactive     = "inactive"
	StatusDisconnected = "disconnected"
)

// Component constants.
const (
	ComponentWiFi   = "wifi"
	ComponentVPN    = "vpn"
	ComponentDNS    = "dns"
	ComponentProxy  = "proxy"
	ComponentDocker = "docker"
)

// Icon constants.
const (
	IconWiFiConnected      = "📶"
	IconWiFiDisconnected   = "📵"
	IconVPNConnected       = "🔒"
	IconVPNDisconnected    = "🔓"
	IconDNSActive          = "🌐"
	IconDNSInactive        = "⚠️"
	IconProxyEnabled       = "🔀"
	IconProxyDisabled      = "➡️"
	IconDockerConnected    = "🐳"
	IconDockerDisconnected = "⭕"
	IconHealthy            = "✅"
	IconUnhealthy          = "❌"
	IconUnknown            = "❓"
)

// Color palette for network environment TUI.
var (
	ColorPrimary    = lipgloss.Color("#00A8E8") // Network blue
	ColorSecondary  = lipgloss.Color("#0066CC") // Darker blue
	ColorSuccess    = lipgloss.Color("#4CBB17") // Green for connected
	ColorWarning    = lipgloss.Color("#FFB347") // Orange for warnings
	ColorError      = lipgloss.Color("#DC143C") // Red for errors
	ColorText       = lipgloss.Color("#E8E8E8") // Light gray text
	ColorSubtle     = lipgloss.Color("#808080") // Gray for subtle text
	ColorBackground = lipgloss.Color("#1A1A1A") // Dark background
	ColorBorder     = lipgloss.Color("#404040") // Border color
	ColorHighlight  = lipgloss.Color("#00D4FF") // Highlight color
	ColorVPN        = lipgloss.Color("#8A2BE2") // Purple for VPN
	ColorWiFi       = lipgloss.Color("#32CD32") // Lime green for WiFi
	ColorDNS        = lipgloss.Color("#FF6347") // Tomato for DNS
	ColorProxy      = lipgloss.Color("#DDA0DD") // Plum for proxy
	ColorInactive   = lipgloss.Color("#696969") // Dim gray for inactive
)

// Base styles.
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

// Network component status styles.
var (
	WiFiConnectedStyle = lipgloss.NewStyle().
				Foreground(ColorWiFi).
				Bold(true)

	WiFiDisconnectedStyle = lipgloss.NewStyle().
				Foreground(ColorInactive)

	VPNConnectedStyle = lipgloss.NewStyle().
				Foreground(ColorVPN).
				Bold(true)

	VPNDisconnectedStyle = lipgloss.NewStyle().
				Foreground(ColorInactive)

	DNSActiveStyle = lipgloss.NewStyle().
			Foreground(ColorDNS).
			Bold(true)

	DNSInactiveStyle = lipgloss.NewStyle().
				Foreground(ColorInactive)

	ProxyActiveStyle = lipgloss.NewStyle().
				Foreground(ColorProxy).
				Bold(true)

	ProxyInactiveStyle = lipgloss.NewStyle().
				Foreground(ColorInactive)

	ConnectedStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	DisconnectedStyle = lipgloss.NewStyle().
				Foreground(ColorError).
				Bold(true)

	UnknownStyle = lipgloss.NewStyle().
			Foreground(ColorSubtle)
)

// Table styles.
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
				Background(lipgloss.Color("#2A2A2A"))

	TableOddRowStyle = lipgloss.NewStyle().
				Background(ColorBackground)
)

// Button and interactive element styles.
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

// Dialog and modal styles.
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

// Progress and loading styles.
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

// Message and notification styles.
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

// Help styles.
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

// Profile and network specific styles.
var (
	ProfileSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorBackground).
				Background(ColorPrimary).
				Bold(true).
				Padding(0, 1)

	ProfileStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Padding(0, 1)

	NetworkMetricStyle = lipgloss.NewStyle().
				Foreground(ColorHighlight).
				Bold(true)

	LatencyGoodStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess)

	LatencyOkStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)

	LatencyPoorStyle = lipgloss.NewStyle().
				Foreground(ColorError)

	SignalStrengthStyle = lipgloss.NewStyle().
				Foreground(ColorWiFi).
				Bold(true)
)

// Border styles.
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

// GetNetworkComponentStyle returns the appropriate style for a network component.
func GetNetworkComponentStyle(component, status string) lipgloss.Style {
	switch component {
	case ComponentWiFi:
		if status == StatusConnected {
			return WiFiConnectedStyle
		}
		return WiFiDisconnectedStyle
	case ComponentVPN:
		if status == StatusConnected {
			return VPNConnectedStyle
		}
		return VPNDisconnectedStyle
	case ComponentDNS:
		if status == StatusActive || status == StatusCustom {
			return DNSActiveStyle
		}
		return DNSInactiveStyle
	case ComponentProxy:
		if status == StatusEnabled {
			return ProxyActiveStyle
		}
		return ProxyInactiveStyle
	default:
		switch status {
		case StatusConnected, StatusActive, StatusEnabled:
			return ConnectedStyle
		case StatusDisconnected, StatusInactive, StatusDisabled:
			return DisconnectedStyle
		}
		return UnknownStyle
	}
}

// GetStatusIcon returns the appropriate icon for a network component status.
func GetStatusIcon(component, status string) string {
	switch component {
	case ComponentWiFi:
		if status == StatusConnected {
			return IconWiFiConnected
		}
		return IconWiFiDisconnected
	case ComponentVPN:
		if status == StatusConnected {
			return IconVPNConnected
		}
		return IconVPNDisconnected
	case ComponentDNS:
		if status == StatusActive || status == StatusCustom {
			return IconDNSActive
		}
		return IconDNSInactive
	case ComponentProxy:
		if status == StatusEnabled {
			return IconProxyEnabled
		}
		return IconProxyDisabled
	case ComponentDocker:
		if status == StatusConnected {
			return IconDockerConnected
		}
		return IconDockerDisconnected
	default:
		switch status {
		case StatusConnected, StatusActive, StatusEnabled:
			return IconHealthy
		case StatusDisconnected, StatusInactive, StatusDisabled:
			return IconUnhealthy
		}
		return IconUnknown
	}
}

// GetSignalStrengthIcon returns an icon representing WiFi signal strength.
func GetSignalStrengthIcon(strength int) string {
	switch {
	case strength >= 80:
		return IconWiFiConnected // Excellent
	case strength >= 60:
		return IconWiFiConnected // Good (could use different icon)
	case strength >= 40:
		return IconWiFiConnected // Fair (could use different icon)
	case strength >= 20:
		return IconWiFiConnected // Poor (could use different icon)
	default:
		return IconWiFiDisconnected // Very poor/no signal
	}
}

// GetLatencyStyle returns style based on latency value.
func GetLatencyStyle(latency int) lipgloss.Style {
	if latency < 50 {
		return LatencyGoodStyle
	} else if latency < 100 {
		return LatencyOkStyle
	}
	return LatencyPoorStyle
}

// AdaptiveStyle adjusts styles based on terminal capabilities.
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
