// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package common

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme represents a color theme for TUI components.
type Theme struct {
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Success    lipgloss.Color
	Warning    lipgloss.Color
	Error      lipgloss.Color
	Text       lipgloss.Color
	Subtle     lipgloss.Color
	Background lipgloss.Color
	Border     lipgloss.Color
	Highlight  lipgloss.Color
}

// DefaultTheme returns the default Go-inspired theme.
func DefaultTheme() Theme {
	return Theme{
		Primary:    lipgloss.Color("#00ADD8"), // Go blue
		Secondary:  lipgloss.Color("#5E81AC"), // Muted blue
		Success:    lipgloss.Color("#A3BE8C"), // Green
		Warning:    lipgloss.Color("#EBCB8B"), // Yellow
		Error:      lipgloss.Color("#BF616A"), // Red
		Text:       lipgloss.Color("#D8DEE9"), // Light gray
		Subtle:     lipgloss.Color("#4C566A"), // Dark gray
		Background: lipgloss.Color("#2E3440"), // Dark background
		Border:     lipgloss.Color("#434C5E"), // Border color
		Highlight:  lipgloss.Color("#88C0D0"), // Highlight color
	}
}

// NetworkTheme returns a theme optimized for network components.
func NetworkTheme() Theme {
	return Theme{
		Primary:    lipgloss.Color("#00A8E8"), // Network blue
		Secondary:  lipgloss.Color("#0066CC"), // Darker blue
		Success:    lipgloss.Color("#28A745"), // Green
		Warning:    lipgloss.Color("#FFC107"), // Yellow
		Error:      lipgloss.Color("#DC3545"), // Red
		Text:       lipgloss.Color("#FFFFFF"), // White
		Subtle:     lipgloss.Color("#6C757D"), // Gray
		Background: lipgloss.Color("#1A1A1A"), // Very dark
		Border:     lipgloss.Color("#495057"), // Border color
		Highlight:  lipgloss.Color("#17A2B8"), // Info color
	}
}

// StyleSet represents a complete set of styles for TUI components.
type StyleSet struct {
	Theme       Theme
	Base        lipgloss.Style
	Title       lipgloss.Style
	Header      lipgloss.Style
	StatusBar   lipgloss.Style
	Footer      lipgloss.Style
	Button      lipgloss.Style
	ActiveBtn   lipgloss.Style
	InactiveBtn lipgloss.Style
	Success     lipgloss.Style
	Warning     lipgloss.Style
	Error       lipgloss.Style
	Border      lipgloss.Style
	Card        lipgloss.Style
	List        lipgloss.Style
	ListItem    lipgloss.Style
	Selected    lipgloss.Style
}

// NewStyleSet creates a new style set with the given theme.
func NewStyleSet(theme Theme) StyleSet {
	return StyleSet{
		Theme: theme,
		Base: lipgloss.NewStyle().
			Foreground(theme.Text).
			Background(theme.Background),
		Title: lipgloss.NewStyle().
			Foreground(theme.Primary).
			Bold(true).
			Padding(0, 1).
			Margin(0, 0, 1, 0),
		Header: lipgloss.NewStyle().
			Foreground(theme.Primary).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Border),
		StatusBar: lipgloss.NewStyle().
			Foreground(theme.Text).
			Background(theme.Border).
			Padding(0, 1),
		Footer: lipgloss.NewStyle().
			Foreground(theme.Subtle).
			Padding(0, 1),
		Button: lipgloss.NewStyle().
			Foreground(theme.Text).
			Background(theme.Secondary).
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Border),
		ActiveBtn: lipgloss.NewStyle().
			Foreground(theme.Background).
			Background(theme.Primary).
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Primary).
			Bold(true),
		InactiveBtn: lipgloss.NewStyle().
			Foreground(theme.Subtle).
			Background(theme.Background).
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Subtle),
		Success: lipgloss.NewStyle().
			Foreground(theme.Success).
			Bold(true),
		Warning: lipgloss.NewStyle().
			Foreground(theme.Warning).
			Bold(true),
		Error: lipgloss.NewStyle().
			Foreground(theme.Error).
			Bold(true),
		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Border).
			Padding(1),
		Card: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Border).
			Padding(1).
			Margin(0, 1, 1, 0),
		List: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Border).
			Padding(0, 1),
		ListItem: lipgloss.NewStyle().
			Padding(0, 1),
		Selected: lipgloss.NewStyle().
			Foreground(theme.Background).
			Background(theme.Highlight).
			Bold(true),
	}
}

// DefaultStyles returns a style set with the default theme.
func DefaultStyles() StyleSet {
	return NewStyleSet(DefaultTheme())
}

// NetworkStyles returns a style set with the network theme.
func NetworkStyles() StyleSet {
	return NewStyleSet(NetworkTheme())
}