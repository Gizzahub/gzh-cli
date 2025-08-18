// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package common

import (
	"github.com/charmbracelet/bubbles/key"
)

// KeyMap represents the key bindings for TUI components.
type KeyMap struct {
	// Navigation
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
	Enter key.Binding
	Back  key.Binding

	// Actions
	Refresh key.Binding
	Save    key.Binding
	Load    key.Binding
	Delete  key.Binding
	Edit    key.Binding
	Copy    key.Binding

	// Tabs and sections
	NextTab     key.Binding
	PrevTab     key.Binding
	NextSection key.Binding
	PrevSection key.Binding

	// Global
	Help key.Binding
	Quit key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Navigation
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "move left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "move right"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "backspace"),
			key.WithHelp("esc", "back"),
		),

		// Actions
		Refresh: key.NewBinding(
			key.WithKeys("r", "F5"),
			key.WithHelp("r", "refresh"),
		),
		Save: key.NewBinding(
			key.WithKeys("s", "ctrl+s"),
			key.WithHelp("s", "save"),
		),
		Load: key.NewBinding(
			key.WithKeys("o", "ctrl+o"),
			key.WithHelp("o", "open/load"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d", "delete"),
			key.WithHelp("d", "delete"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e", "F2"),
			key.WithHelp("e", "edit"),
		),
		Copy: key.NewBinding(
			key.WithKeys("c", "ctrl+c"),
			key.WithHelp("c", "copy"),
		),

		// Tabs and sections
		NextTab: key.NewBinding(
			key.WithKeys("tab", "ctrl+right"),
			key.WithHelp("tab", "next tab"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("shift+tab", "ctrl+left"),
			key.WithHelp("shift+tab", "prev tab"),
		),
		NextSection: key.NewBinding(
			key.WithKeys("ctrl+down", "page_down"),
			key.WithHelp("ctrl+↓", "next section"),
		),
		PrevSection: key.NewBinding(
			key.WithKeys("ctrl+up", "page_up"),
			key.WithHelp("ctrl+↑", "prev section"),
		),

		// Global
		Help: key.NewBinding(
			key.WithKeys("?", "F1"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c", "ctrl+q"),
			key.WithHelp("q", "quit"),
		),
	}
}

// ShortHelp returns the short help for the key bindings.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up, k.Down, k.Enter, k.Back, k.Refresh, k.Help, k.Quit,
	}
}

// FullHelp returns the full help for the key bindings.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},    // Navigation
		{k.Enter, k.Back, k.Refresh},       // Basic actions
		{k.Save, k.Load, k.Edit, k.Delete}, // File operations
		{k.NextTab, k.PrevTab},             // Tab navigation
		{k.Copy, k.Help, k.Quit},           // Global actions
	}
}

// CustomKeyMap allows creating custom key bindings for specific components.
type CustomKeyMap struct {
	KeyMap
	Custom map[string]key.Binding
}

// NewCustomKeyMap creates a new custom key map based on the default one.
func NewCustomKeyMap() *CustomKeyMap {
	return &CustomKeyMap{
		KeyMap: DefaultKeyMap(),
		Custom: make(map[string]key.Binding),
	}
}

// AddCustomBinding adds a custom key binding.
func (ckm *CustomKeyMap) AddCustomBinding(name string, binding key.Binding) {
	ckm.Custom[name] = binding
}

// GetCustomBinding retrieves a custom key binding.
func (ckm *CustomKeyMap) GetCustomBinding(name string) (key.Binding, bool) {
	binding, exists := ckm.Custom[name]
	return binding, exists
}

// AllBindings returns all bindings including custom ones.
func (ckm *CustomKeyMap) AllBindings() []key.Binding {
	bindings := ckm.FullHelp()
	var all []key.Binding

	for _, row := range bindings {
		all = append(all, row...)
	}

	for _, binding := range ckm.Custom {
		all = append(all, binding)
	}

	return all
}
