// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package ide

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldRunInBackground(t *testing.T) {
	tests := []struct {
		name     string
		options  openOptions
		ide      IDE
		expected bool
	}{
		{
			name: "Explicit background flag",
			options: openOptions{
				background: true,
			},
			ide: IDE{
				Type: "jetbrains",
			},
			expected: true,
		},
		{
			name: "Explicit wait flag",
			options: openOptions{
				wait: true,
			},
			ide: IDE{
				Type: "jetbrains",
			},
			expected: false,
		},
		{
			name:    "JetBrains IDE default",
			options: openOptions{},
			ide: IDE{
				Type: "jetbrains",
				Name: "PyCharm Professional",
			},
			expected: true,
		},
		{
			name:    "VS Code default",
			options: openOptions{},
			ide: IDE{
				Type: "vscode",
				Name: "Visual Studio Code",
			},
			expected: true,
		},
		{
			name:    "Vim should wait",
			options: openOptions{},
			ide: IDE{
				Type: "other",
				Name: "Vim",
			},
			expected: false,
		},
		{
			name:    "Neovim should wait",
			options: openOptions{},
			ide: IDE{
				Type: "other",
				Name: "Neovim",
			},
			expected: false,
		},
		{
			name:    "Emacs should wait",
			options: openOptions{},
			ide: IDE{
				Type: "other",
				Name: "Emacs",
			},
			expected: false,
		},
		{
			name:    "Sublime Text should run in background",
			options: openOptions{},
			ide: IDE{
				Type: "other",
				Name: "Sublime Text",
			},
			expected: true,
		},
		{
			name:    "Unknown other IDE default to background",
			options: openOptions{},
			ide: IDE{
				Type: "other",
				Name: "Unknown Editor",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.options.shouldRunInBackground(&tt.ide)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPrepareIDEArgs(t *testing.T) {
	options := &openOptions{}

	tests := []struct {
		name       string
		ide        IDE
		targetPath string
		expected   []string
	}{
		{
			name: "JetBrains IDE",
			ide: IDE{
				Type: "jetbrains",
				Name: "PyCharm Professional",
			},
			targetPath: "/home/user/project",
			expected:   []string{"/home/user/project"},
		},
		{
			name: "VS Code family",
			ide: IDE{
				Type: "vscode",
				Name: "Visual Studio Code",
			},
			targetPath: "/home/user/project",
			expected:   []string{"/home/user/project"},
		},
		{
			name: "Sublime Text",
			ide: IDE{
				Type: "other",
				Name: "Sublime Text",
			},
			targetPath: "/home/user/project",
			expected:   []string{"--project", "/home/user/project"},
		},
		{
			name: "Vim",
			ide: IDE{
				Type: "other",
				Name: "Vim",
			},
			targetPath: "/home/user/project",
			expected:   []string{"/home/user/project"},
		},
		{
			name: "Unknown editor",
			ide: IDE{
				Type: "other",
				Name: "Unknown Editor",
			},
			targetPath: "/home/user/project",
			expected:   []string{"/home/user/project"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := options.prepareIDEArgs(&tt.ide, tt.targetPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAvailableIDENames(t *testing.T) {
	options := &openOptions{}

	ides := []IDE{
		{
			Name:    "PyCharm Professional",
			Aliases: []string{"pycharm", "pycharm-pro"},
		},
		{
			Name:    "Visual Studio Code",
			Aliases: []string{"code", "vscode"},
		},
		{
			Name:    "GoLand",
			Aliases: []string{"goland"},
		},
	}

	result := options.getAvailableIDENames(ides)

	// Should contain all main names
	assert.Contains(t, result, "PyCharm Professional")
	assert.Contains(t, result, "pycharm")
	assert.Contains(t, result, "Visual Studio Code")
	assert.Contains(t, result, "code")
	assert.Contains(t, result, "GoLand")
	// Note: aliases are added to the list, but order may vary due to deduplication

	// Test with empty slice
	emptyResult := options.getAvailableIDENames([]IDE{})
	assert.Equal(t, "none found - run 'gz ide scan' first", emptyResult)
}

func TestGetAvailableIDENamesLimit(t *testing.T) {
	options := &openOptions{}

	// Create more than 10 IDEs to test truncation
	var ides []IDE
	for i := 0; i < 12; i++ {
		ides = append(ides, IDE{
			Name:    fmt.Sprintf("IDE %d", i),
			Aliases: []string{fmt.Sprintf("ide%d", i)},
		})
	}

	result := options.getAvailableIDENames(ides)
	assert.Contains(t, result, "... (and more)")
}
