// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package open

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
				Name: "VS Code",
			},
			expected: true,
		},
		{
			name:    "Sublime Text",
			options: openOptions{},
			ide: IDE{
				Type: "other",
				Name: "Sublime Text 4",
			},
			expected: true,
		},
		{
			name:    "Vim",
			options: openOptions{},
			ide: IDE{
				Type: "other",
				Name: "Vim",
			},
			expected: false,
		},
		{
			name:    "NeoVim",
			options: openOptions{},
			ide: IDE{
				Type: "other",
				Name: "neovim",
			},
			expected: false,
		},
		{
			name:    "Emacs",
			options: openOptions{},
			ide: IDE{
				Type: "other",
				Name: "Emacs",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.options.shouldRunInBackground(&tt.ide)
			assert.Equal(t, tt.expected, result, "Expected %v for %s", tt.expected, tt.name)
		})
	}
}

func TestPrepareIDEArgs(t *testing.T) {
	options := openOptions{}
	targetPath := "/Users/test/project"

	tests := []struct {
		name     string
		ide      IDE
		expected []string
	}{
		{
			name: "JetBrains IDE",
			ide: IDE{
				Type: "jetbrains",
				Name: "PyCharm",
			},
			expected: []string{targetPath},
		},
		{
			name: "VS Code",
			ide: IDE{
				Type: "vscode",
				Name: "VS Code",
			},
			expected: []string{targetPath},
		},
		{
			name: "Sublime Text",
			ide: IDE{
				Type: "other",
				Name: "Sublime Text 4",
			},
			expected: []string{"--project", targetPath},
		},
		{
			name: "Vim",
			ide: IDE{
				Type: "other",
				Name: "Vim",
			},
			expected: []string{targetPath},
		},
		{
			name: "Emacs",
			ide: IDE{
				Type: "other",
				Name: "Emacs",
			},
			expected: []string{targetPath},
		},
		{
			name: "Unknown IDE",
			ide: IDE{
				Type: "other",
				Name: "Unknown Editor",
			},
			expected: []string{targetPath},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := options.prepareIDEArgs(&tt.ide, targetPath)
			assert.Equal(t, tt.expected, result, "Args mismatch for %s", tt.name)
		})
	}
}

func TestGetAvailableIDENames(t *testing.T) {
	options := openOptions{}

	tests := []struct {
		name     string
		ides     []IDE
		expected string
	}{
		{
			name:     "Empty list",
			ides:     []IDE{},
			expected: "none found - run 'gz ide scan' first",
		},
		{
			name: "Single IDE",
			ides: []IDE{
				{
					Name:    "VS Code",
					Aliases: []string{"code", "vscode"},
				},
			},
			expected: "VS Code, code, vscode",
		},
		{
			name: "Multiple IDEs with aliases",
			ides: []IDE{
				{
					Name:    "VS Code",
					Aliases: []string{"code"},
				},
				{
					Name:    "PyCharm",
					Aliases: []string{"pycharm"},
				},
			},
			expected: "VS Code, code, PyCharm, pycharm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := options.getAvailableIDENames(tt.ides)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolvePath(t *testing.T) {
	options := openOptions{}

	tests := []struct {
		name      string
		inputPath string
		wantError bool
	}{
		{
			name:      "Current directory dot",
			inputPath: ".",
			wantError: false,
		},
		{
			name:      "Empty path",
			inputPath: "",
			wantError: false,
		},
		{
			name:      "Absolute path",
			inputPath: "/Users/test",
			wantError: false,
		},
		{
			name:      "Relative path",
			inputPath: "../test",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := options.resolvePath(tt.inputPath)

			if tt.wantError {
				assert.Error(t, err, "Expected error for %s", tt.name)
			} else {
				assert.NoError(t, err, "Unexpected error for %s", tt.name)
				assert.NotEmpty(t, result, "Result should not be empty for %s", tt.name)
			}
		})
	}
}

func TestMockDetector(t *testing.T) {
	detector := NewIDEDetector()

	// Test DetectIDEs
	ides, err := detector.DetectIDEs(true)
	assert.NoError(t, err)
	assert.NotEmpty(t, ides, "Should return some mock IDEs")

	// Test FindIDEByAlias
	foundIDE := detector.FindIDEByAlias(ides, "code")
	assert.NotNil(t, foundIDE, "Should find VS Code by alias 'code'")
	assert.Equal(t, "VS Code", foundIDE.Name)

	// Test case insensitive search
	foundIDE = detector.FindIDEByAlias(ides, "CODE")
	assert.NotNil(t, foundIDE, "Should find VS Code by alias 'CODE' (case insensitive)")

	// Test non-existent IDE
	foundIDE = detector.FindIDEByAlias(ides, "nonexistent")
	assert.Nil(t, foundIDE, "Should not find non-existent IDE")
}

// Example of IDE names truncation test.
func TestGetAvailableIDENamesTruncation(t *testing.T) {
	options := openOptions{}

	// Create more than 10 IDEs to test truncation
	var ides []IDE
	for i := 0; i < 15; i++ {
		ides = append(ides, IDE{
			Name: fmt.Sprintf("IDE%d", i),
		})
	}

	result := options.getAvailableIDENames(ides)
	assert.Contains(t, result, "... (and more)", "Should truncate and add '... (and more)' for long lists")
}
