// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package common

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Common message types used across TUI components.

// RefreshMsg signals a component to refresh its data.
type RefreshMsg struct {
	Component string
	Force     bool
}

// ErrorMsg represents an error that occurred.
type ErrorMsg struct {
	Err       error
	Component string
	Timestamp time.Time
}

// StatusMsg represents a status update.
type StatusMsg struct {
	Message   string
	Level     StatusLevel
	Component string
	Timestamp time.Time
}

// StatusLevel represents the level of a status message.
type StatusLevel int

const (
	StatusInfo StatusLevel = iota
	StatusSuccess
	StatusWarning
	StatusError
)

// String returns the string representation of a status level.
func (sl StatusLevel) String() string {
	switch sl {
	case StatusInfo:
		return "info"
	case StatusSuccess:
		return "success"
	case StatusWarning:
		return "warning"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

// LoadingMsg indicates that a component is loading.
type LoadingMsg struct {
	Component string
	Message   string
	Show      bool
}

// DataUpdateMsg signals that data has been updated.
type DataUpdateMsg struct {
	Component string
	Data      interface{}
	Timestamp time.Time
}

// TabChangeMsg signals a tab change.
type TabChangeMsg struct {
	PreviousTab int
	CurrentTab  int
	TabName     string
}

// SectionChangeMsg signals a section change.
type SectionChangeMsg struct {
	PreviousSection int
	CurrentSection  int
	SectionName     string
}

// KeyPressMsg wraps a key press with additional context.
type KeyPressMsg struct {
	Key       string
	Component string
	Context   map[string]interface{}
}

// ResizeMsg signals that the terminal has been resized.
type ResizeMsg struct {
	Width  int
	Height int
}

// ConfigChangeMsg signals that configuration has changed.
type ConfigChangeMsg struct {
	Component  string
	ConfigPath string
	Changes    map[string]interface{}
}

// Helper functions for creating common messages.

// NewRefreshMsg creates a new refresh message.
func NewRefreshMsg(component string, force bool) RefreshMsg {
	return RefreshMsg{
		Component: component,
		Force:     force,
	}
}

// NewErrorMsg creates a new error message.
func NewErrorMsg(component string, err error) ErrorMsg {
	return ErrorMsg{
		Err:       err,
		Component: component,
		Timestamp: time.Now(),
	}
}

// NewStatusMsg creates a new status message.
func NewStatusMsg(component, message string, level StatusLevel) StatusMsg {
	return StatusMsg{
		Message:   message,
		Level:     level,
		Component: component,
		Timestamp: time.Now(),
	}
}

// NewLoadingMsg creates a new loading message.
func NewLoadingMsg(component, message string, show bool) LoadingMsg {
	return LoadingMsg{
		Component: component,
		Message:   message,
		Show:      show,
	}
}

// NewDataUpdateMsg creates a new data update message.
func NewDataUpdateMsg(component string, data interface{}) DataUpdateMsg {
	return DataUpdateMsg{
		Component: component,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// Command helpers for common operations.

// RefreshCmd returns a command that sends a refresh message.
func RefreshCmd(component string, force bool) tea.Cmd {
	return func() tea.Msg {
		return NewRefreshMsg(component, force)
	}
}

// ErrorCmd returns a command that sends an error message.
func ErrorCmd(component string, err error) tea.Cmd {
	return func() tea.Msg {
		return NewErrorMsg(component, err)
	}
}

// StatusCmd returns a command that sends a status message.
func StatusCmd(component, message string, level StatusLevel) tea.Cmd {
	return func() tea.Msg {
		return NewStatusMsg(component, message, level)
	}
}

// LoadingCmd returns a command that sends a loading message.
func LoadingCmd(component, message string, show bool) tea.Cmd {
	return func() tea.Msg {
		return NewLoadingMsg(component, message, show)
	}
}

// DelayCmd returns a command that sends a message after a delay.
func DelayCmd(delay time.Duration, msg tea.Msg) tea.Cmd {
	return tea.Tick(delay, func(time.Time) tea.Msg {
		return msg
	})
}

// BatchCmd returns a command that sends multiple messages.
func BatchCmd(msgs ...tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	for _, msg := range msgs {
		cmds = append(cmds, func(m tea.Msg) tea.Cmd {
			return func() tea.Msg { return m }
		}(msg))
	}
	return tea.Batch(cmds...)
}
