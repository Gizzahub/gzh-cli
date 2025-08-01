// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines keyboard shortcuts for the net-env TUI.
type KeyMap struct {
	Up              key.Binding
	Down            key.Binding
	Left            key.Binding
	Right           key.Binding
	Enter           key.Binding
	Back            key.Binding
	Quit            key.Binding
	Help            key.Binding
	Refresh         key.Binding
	Search          key.Binding
	Filter          key.Binding
	SwitchProfile   key.Binding
	VPNToggle       key.Binding
	DNSSettings     key.Binding
	ProxyToggle     key.Binding
	Monitor         key.Binding
	Settings        key.Binding
	QuickAction1    key.Binding
	QuickAction2    key.Binding
	QuickAction3    key.Binding
	QuickConnect    key.Binding
	QuickDisconnect key.Binding
}

// DefaultKeyMap provides the default keyboard shortcuts for net-env.
var DefaultKeyMap = KeyMap{
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
		key.WithKeys("enter"),
		key.WithHelp("enter", "select/confirm"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "backspace"),
		key.WithHelp("esc", "go back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "Q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r", "ctrl+r"),
		key.WithHelp("r", "refresh"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Filter: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "filter"),
	),
	SwitchProfile: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "switch profile"),
	),
	VPNToggle: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "vpn toggle"),
	),
	DNSSettings: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "dns settings"),
	),
	ProxyToggle: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "proxy toggle"),
	),
	Monitor: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "monitor"),
	),
	Settings: key.NewBinding(
		key.WithKeys("P"),
		key.WithHelp("P", "settings"),
	),
	QuickAction1: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "quick action 1"),
	),
	QuickAction2: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "quick action 2"),
	),
	QuickAction3: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "quick action 3"),
	),
	QuickConnect: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "quick connect"),
	),
	QuickDisconnect: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "quick disconnect"),
	),
}

// ShortHelp returns key bindings to be shown in the mini help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns key bindings for the expanded help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},                                   // navigation
		{k.Enter, k.Back, k.Quit, k.Help},                                 // actions
		{k.Refresh, k.Search, k.Filter},                                   // utilities
		{k.SwitchProfile, k.VPNToggle, k.Monitor, k.Settings},             // views
		{k.DNSSettings, k.ProxyToggle, k.QuickConnect, k.QuickDisconnect}, // network actions
		{k.QuickAction1, k.QuickAction2, k.QuickAction3},                  // quick actions
	}
}

// Enabled returns whether the keymap is enabled.
func (k KeyMap) Enabled() bool {
	return true
}
