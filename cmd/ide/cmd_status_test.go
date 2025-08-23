// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package ide

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatInstallMethod(t *testing.T) {
	options := &statusOptions{}

	tests := []struct {
		name     string
		method   string
		path     string
		expected string
	}{
		{
			name:     "AppImage installation",
			method:   "appimage",
			path:     "/home/user/Apps",
			expected: "AppImage",
		},
		{
			name:     "Pacman installation",
			method:   "pacman",
			path:     "vim 9.1.1-1",
			expected: "Pacman (Arch Linux)",
		},
		{
			name:     "Snap installation",
			method:   "snap",
			path:     "cursor",
			expected: "Snap",
		},
		{
			name:     "Flatpak installation",
			method:   "flatpak",
			path:     "com.cursor.Cursor",
			expected: "Flatpak",
		},
		{
			name:     "JetBrains Toolbox",
			method:   "toolbox",
			path:     "/home/user/.local/share/JetBrains/Toolbox/apps/PyCharm-P",
			expected: "JetBrains Toolbox",
		},
		{
			name:     "Direct installation",
			method:   "direct",
			path:     "/usr/bin/vim",
			expected: "Direct Installation",
		},
		{
			name:     "Unknown method",
			method:   "homebrew",
			path:     "/opt/homebrew/bin/vim",
			expected: "homebrew",
		},
		{
			name:     "Empty method",
			method:   "",
			path:     "/usr/bin/vim",
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := options.formatInstallMethod(tt.method, tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatPath(t *testing.T) {
	options := &statusOptions{}

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "Home directory path",
			path:     "/home/user/Apps/cursor",
			expected: "~/Apps/cursor",
		},
		{
			name:     "System path",
			path:     "/usr/bin/vim",
			expected: "/usr/bin/vim",
		},
		{
			name:     "Short home path",
			path:     "/home/user",
			expected: "~/", // Actually gets converted to ~/
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := options.formatPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatLastUpdated(t *testing.T) {
	options := &statusOptions{}
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "Zero time",
			time:     time.Time{},
			expected: "unknown",
		},
		{
			name:     "Just now",
			time:     now.Add(-30 * time.Second),
			expected: "just now",
		},
		{
			name:     "Minutes ago",
			time:     now.Add(-5 * time.Minute),
			expected: "5m ago",
		},
		{
			name:     "Hours ago",
			time:     now.Add(-3 * time.Hour),
			expected: "3h ago",
		},
		{
			name:     "Days ago",
			time:     now.Add(-2 * 24 * time.Hour),
			expected: "2d ago",
		},
		{
			name:     "Weeks ago",
			time:     now.Add(-10 * 24 * time.Hour),
			expected: "1w ago",
		},
		{
			name:     "Months ago",
			time:     now.Add(-60 * 24 * time.Hour),
			expected: "2m ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := options.formatLastUpdated(tt.time)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatDetailedTime(t *testing.T) {
	options := &statusOptions{}

	// Test zero time
	result := options.formatDetailedTime(time.Time{})
	assert.Equal(t, "unknown", result)

	// Test actual time
	testTime := time.Date(2025, 8, 14, 13, 37, 6, 0, time.UTC)
	result = options.formatDetailedTime(testTime)
	assert.Equal(t, "2025-08-14 13:37:06", result)
}

func TestTruncateString(t *testing.T) {
	options := &statusOptions{}

	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "Short string",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "Exact length",
			input:    "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "Long string",
			input:    "this is a very long string",
			maxLen:   10,
			expected: "this is...",
		},
		{
			name:     "Very short max length",
			input:    "hello",
			maxLen:   3,
			expected: "...", // maxLen=3 means only "..." fits (3-3=0 chars + ...)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := options.truncateString(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEscapeJSON(t *testing.T) {
	options := &statusOptions{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple string",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "String with quotes",
			input:    "hello \"world\"",
			expected: "hello \\\"world\\\"",
		},
		{
			name:     "String with backslashes",
			input:    "C:\\Program Files\\IDE",
			expected: "C:\\\\Program Files\\\\IDE",
		},
		{
			name:     "String with newlines",
			input:    "line1\nline2\r\nline3",
			expected: "line1\\nline2\\r\\nline3",
		},
		{
			name:     "String with tabs",
			input:    "column1\tcolumn2",
			expected: "column1\\tcolumn2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := options.escapeJSON(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEscapeYAML(t *testing.T) {
	options := &statusOptions{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple string",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "String with quotes",
			input:    "hello \"world\"",
			expected: "hello \\\"world\\\"",
		},
		{
			name:     "String with newlines",
			input:    "line1\nline2\r\nline3",
			expected: "line1\\nline2\\r\\nline3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := options.escapeYAML(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
