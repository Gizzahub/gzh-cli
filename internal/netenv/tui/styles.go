// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package tui

import (
	"github.com/gizzahub/gzh-cli/internal/tui/common"
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

// NetEnvStyles holds the styles for network environment TUI.
type NetEnvStyles struct {
	common.StyleSet
	ComponentActive    lipgloss.Style
	ComponentInactive  lipgloss.Style
	StatusConnected    lipgloss.Style
	StatusDisconnected lipgloss.Style
	StatusWarning      lipgloss.Style
	MetricGood         lipgloss.Style
	MetricBad          lipgloss.Style
	TabActive          lipgloss.Style
	TabInactive        lipgloss.Style
}

// NewNetEnvStyles creates a new set of styles for network environment TUI.
func NewNetEnvStyles() NetEnvStyles {
	baseStyles := common.NetworkStyles()

	return NetEnvStyles{
		StyleSet: baseStyles,
		ComponentActive: lipgloss.NewStyle().
			Foreground(baseStyles.Theme.Success).
			Bold(true),
		ComponentInactive: lipgloss.NewStyle().
			Foreground(baseStyles.Theme.Subtle),
		StatusConnected: lipgloss.NewStyle().
			Foreground(baseStyles.Theme.Success).
			Bold(true),
		StatusDisconnected: lipgloss.NewStyle().
			Foreground(baseStyles.Theme.Error).
			Bold(true),
		StatusWarning: lipgloss.NewStyle().
			Foreground(baseStyles.Theme.Warning).
			Bold(true),
		MetricGood: lipgloss.NewStyle().
			Foreground(baseStyles.Theme.Success),
		MetricBad: lipgloss.NewStyle().
			Foreground(baseStyles.Theme.Error),
		TabActive: lipgloss.NewStyle().
			Foreground(baseStyles.Theme.Background).
			Background(baseStyles.Theme.Primary).
			Padding(0, 2).
			Bold(true),
		TabInactive: lipgloss.NewStyle().
			Foreground(baseStyles.Theme.Subtle).
			Padding(0, 2),
	}
}

// Legacy style variables for backward compatibility.
var (
	styles = NewNetEnvStyles()

	// Backward compatibility exports.
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

	// Base styles.
	BaseStyle      = styles.Base
	TitleStyle     = styles.Title
	HeaderStyle    = styles.Header
	StatusBarStyle = styles.StatusBar
	FooterStyle    = styles.Footer

	// Component styles.
	ComponentActiveStyle    = styles.ComponentActive
	ComponentInactiveStyle  = styles.ComponentInactive
	StatusConnectedStyle    = styles.StatusConnected
	StatusDisconnectedStyle = styles.StatusDisconnected
	StatusWarningStyle      = styles.StatusWarning
	MetricGoodStyle         = styles.MetricGood
	MetricBadStyle          = styles.MetricBad
	TabActiveStyle          = styles.TabActive
	TabInactiveStyle        = styles.TabInactive

	// Additional styles for compatibility.
	SpinnerStyle    = styles.Base.Foreground(styles.Theme.Primary)
	ErrorStyle      = styles.Base.Foreground(styles.Theme.Error).Bold(true)
	WarningStyle    = styles.Base.Foreground(styles.Theme.Warning).Bold(true)
	InfoStyle       = styles.Base.Foreground(styles.Theme.Primary).Bold(true)
	HelpHeaderStyle = styles.Base.Foreground(styles.Theme.Primary).Bold(true).Margin(1, 0)
)

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
		return IconWiFiConnected // Good
	case strength >= 40:
		return IconWiFiConnected // Fair
	case strength >= 20:
		return IconWiFiConnected // Poor
	default:
		return IconWiFiDisconnected // Very poor/no signal
	}
}
