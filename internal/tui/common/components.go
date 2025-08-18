// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package common

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Component defines the interface for TUI components.
type Component interface {
	// Init initializes the component
	Init() tea.Cmd
	
	// Update handles messages and updates component state
	Update(tea.Msg) (Component, tea.Cmd)
	
	// View renders the component
	View() string
	
	// GetID returns the component's unique identifier
	GetID() string
	
	// SetSize sets the component's dimensions
	SetSize(width, height int)
	
	// GetSize returns the component's dimensions
	GetSize() (width, height int)
	
	// SetFocus sets whether the component has focus
	SetFocus(focused bool)
	
	// HasFocus returns whether the component has focus
	HasFocus() bool
}

// BaseComponent provides common functionality for TUI components.
type BaseComponent struct {
	id       string
	width    int
	height   int
	focused  bool
	styles   StyleSet
	keyMap   KeyMap
	disabled bool
	visible  bool
}

// NewBaseComponent creates a new base component.
func NewBaseComponent(id string, styles StyleSet, keyMap KeyMap) BaseComponent {
	return BaseComponent{
		id:      id,
		styles:  styles,
		keyMap:  keyMap,
		visible: true,
	}
}

// GetID returns the component's unique identifier.
func (bc *BaseComponent) GetID() string {
	return bc.id
}

// SetSize sets the component's dimensions.
func (bc *BaseComponent) SetSize(width, height int) {
	bc.width = width
	bc.height = height
}

// GetSize returns the component's dimensions.
func (bc *BaseComponent) GetSize() (width, height int) {
	return bc.width, bc.height
}

// SetFocus sets whether the component has focus.
func (bc *BaseComponent) SetFocus(focused bool) {
	bc.focused = focused
}

// HasFocus returns whether the component has focus.
func (bc *BaseComponent) HasFocus() bool {
	return bc.focused
}

// SetDisabled sets whether the component is disabled.
func (bc *BaseComponent) SetDisabled(disabled bool) {
	bc.disabled = disabled
}

// IsDisabled returns whether the component is disabled.
func (bc *BaseComponent) IsDisabled() bool {
	return bc.disabled
}

// SetVisible sets whether the component is visible.
func (bc *BaseComponent) SetVisible(visible bool) {
	bc.visible = visible
}

// IsVisible returns whether the component is visible.
func (bc *BaseComponent) IsVisible() bool {
	return bc.visible
}

// GetStyles returns the component's style set.
func (bc *BaseComponent) GetStyles() StyleSet {
	return bc.styles
}

// GetKeyMap returns the component's key map.
func (bc *BaseComponent) GetKeyMap() KeyMap {
	return bc.keyMap
}

// Container represents a component that can contain other components.
type Container interface {
	Component
	
	// AddChild adds a child component
	AddChild(Component)
	
	// RemoveChild removes a child component
	RemoveChild(string) // by ID
	
	// GetChild gets a child component by ID
	GetChild(string) Component
	
	// GetChildren returns all child components
	GetChildren() []Component
	
	// SetActiveChild sets the active child component
	SetActiveChild(string) // by ID
	
	// GetActiveChild returns the active child component
	GetActiveChild() Component
}

// BaseContainer provides common container functionality.
type BaseContainer struct {
	BaseComponent
	children    []Component
	activeChild string
}

// NewBaseContainer creates a new base container.
func NewBaseContainer(id string, styles StyleSet, keyMap KeyMap) BaseContainer {
	return BaseContainer{
		BaseComponent: NewBaseComponent(id, styles, keyMap),
		children:      make([]Component, 0),
	}
}

// AddChild adds a child component.
func (bc *BaseContainer) AddChild(child Component) {
	bc.children = append(bc.children, child)
	if bc.activeChild == "" {
		bc.activeChild = child.GetID()
	}
}

// RemoveChild removes a child component by ID.
func (bc *BaseContainer) RemoveChild(id string) {
	for i, child := range bc.children {
		if child.GetID() == id {
			bc.children = append(bc.children[:i], bc.children[i+1:]...)
			if bc.activeChild == id && len(bc.children) > 0 {
				bc.activeChild = bc.children[0].GetID()
			}
			break
		}
	}
}

// GetChild gets a child component by ID.
func (bc *BaseContainer) GetChild(id string) Component {
	for _, child := range bc.children {
		if child.GetID() == id {
			return child
		}
	}
	return nil
}

// GetChildren returns all child components.
func (bc *BaseContainer) GetChildren() []Component {
	return bc.children
}

// SetActiveChild sets the active child component.
func (bc *BaseContainer) SetActiveChild(id string) {
	bc.activeChild = id
}

// GetActiveChild returns the active child component.
func (bc *BaseContainer) GetActiveChild() Component {
	return bc.GetChild(bc.activeChild)
}

// Helper functions for common component operations.

// WrapInBorder wraps content in a border with the given title.
func WrapInBorder(content, title string, style lipgloss.Style) string {
	if title != "" {
		// Use lipgloss.JoinVertical to add title above content
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(style.GetForeground())
		titleLine := titleStyle.Render(title)
		content = lipgloss.JoinVertical(lipgloss.Left, titleLine, content)
	}
	return style.Render(content)
}

// TruncateText truncates text to fit within the given width.
func TruncateText(text string, width int, suffix string) string {
	if width <= 0 {
		return ""
	}
	
	if len(text) <= width {
		return text
	}
	
	if len(suffix) >= width {
		return suffix[:width]
	}
	
	return text[:width-len(suffix)] + suffix
}

// PadContent pads content to fit within the given dimensions.
func PadContent(content string, width, height int) string {
	lines := strings.Split(content, "\n")
	
	// Pad or truncate to height
	for len(lines) < height {
		lines = append(lines, "")
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	
	// Pad or truncate each line to width
	for i, line := range lines {
		if len(line) < width {
			lines[i] = line + strings.Repeat(" ", width-len(line))
		} else if len(line) > width {
			lines[i] = line[:width]
		}
	}
	
	return strings.Join(lines, "\n")
}

// CenterText centers text within the given width.
func CenterText(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}
	
	padding := (width - len(text)) / 2
	leftPad := strings.Repeat(" ", padding)
	rightPad := strings.Repeat(" ", width-len(text)-padding)
	
	return leftPad + text + rightPad
}

// StatusIcon returns an appropriate icon for the given status.
func StatusIcon(status string) string {
	switch strings.ToLower(status) {
	case "active", "connected", "enabled", "online", "running":
		return "✅"
	case "inactive", "disconnected", "disabled", "offline", "stopped":
		return "❌"
	case "warning", "degraded", "partial":
		return "⚠️"
	case "loading", "pending", "processing":
		return "⏳"
	case "unknown", "unavailable":
		return "❓"
	default:
		return "⚪"
	}
}