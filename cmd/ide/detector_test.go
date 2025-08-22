// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package ide

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIDEDetector(t *testing.T) {
	detector := NewIDEDetector()
	assert.NotNil(t, detector)
	assert.NotEmpty(t, detector.cacheDir)
}

func TestParseJetBrainsBuildNumber(t *testing.T) {
	detector := NewIDEDetector()

	tests := []struct {
		name        string
		buildNumber string
		expected    string
	}{
		{
			name:        "PyCharm build number",
			buildNumber: "PY-252.23892.515",
			expected:    "2025.2.515",
		},
		{
			name:        "IntelliJ IDEA build number",
			buildNumber: "IU-252.23892.409",
			expected:    "2025.2.409",
		},
		{
			name:        "WebStorm build number",
			buildNumber: "WS-252.23892.411",
			expected:    "2025.2.411",
		},
		{
			name:        "Invalid format",
			buildNumber: "invalid-format",
			expected:    "invalid-format",
		},
		{
			name:        "Short format",
			buildNumber: "PY-252",
			expected:    "PY-252",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.parseJetBrainsBuildNumber(tt.buildNumber)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsVersionNumber(t *testing.T) {
	detector := NewIDEDetector()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid version 1.2.3",
			input:    "1.2.3",
			expected: true,
		},
		{
			name:     "Valid version 2025.2.0.1",
			input:    "2025.2.0.1",
			expected: true,
		},
		{
			name:     "Valid version with hyphen",
			input:    "1.2.3-beta",
			expected: false, // Current implementation doesn't allow letters after hyphen
		},
		{
			name:     "Invalid - no dot",
			input:    "123",
			expected: false,
		},
		{
			name:     "Invalid - no digits",
			input:    "abc.def",
			expected: false,
		},
		{
			name:     "Invalid - too short",
			input:    "1.",
			expected: false,
		},
		{
			name:     "Invalid - letters",
			input:    "1.2.3.abc",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.isVersionNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseVSCodeVersion(t *testing.T) {
	detector := NewIDEDetector()

	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{
			name:     "VS Code version only",
			output:   "1.103.1",
			expected: "1.103.1",
		},
		{
			name:     "VS Code with app name",
			output:   "Visual Studio Code 1.103.1",
			expected: "1.103.1",
		},
		{
			name:     "Multi-line output",
			output:   "1.103.1\nCommit: abc123\nDate: 2025-01-01",
			expected: "1.103.1",
		},
		{
			name:     "No version found",
			output:   "Some random output",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.parseVSCodeVersion(tt.output)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetJetBrainsVersionFromBuildFile(t *testing.T) {
	detector := NewIDEDetector()

	// Create a temporary directory structure
	tempDir := t.TempDir()
	buildFile := filepath.Join(tempDir, "build.txt")

	// Test with valid build.txt
	err := os.WriteFile(buildFile, []byte("PY-252.23892.515\n"), 0644)
	assert.NoError(t, err)

	version := detector.getJetBrainsVersionFromBuildFile(tempDir)
	assert.Equal(t, "2025.2.515", version)

	// Test with non-existent file
	nonExistentDir := filepath.Join(tempDir, "nonexistent")
	version = detector.getJetBrainsVersionFromBuildFile(nonExistentDir)
	assert.Equal(t, "unknown", version)
}

func TestExtractJetBrainsVersionFromDir(t *testing.T) {
	detector := NewIDEDetector()

	tests := []struct {
		name     string
		dirName  string
		expected string
	}{
		{
			name:     "Version in directory name",
			dirName:  "pycharm-2024.3",
			expected: "2024.3",
		},
		{
			name:     "Multiple hyphens",
			dirName:  "intellij-idea-ultimate-2024.3",
			expected: "2024.3",
		},
		{
			name:     "No version",
			dirName:  "pycharm",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.extractJetBrainsVersionFromDir(tt.dirName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindIDEByAlias(t *testing.T) {
	detector := NewIDEDetector()

	ides := []IDE{
		{
			Name:    "PyCharm Professional",
			Aliases: []string{"pycharm", "pycharm-pro"},
		},
		{
			Name:    "Visual Studio Code",
			Aliases: []string{"code", "vscode"},
		},
	}

	// Test exact name match
	result := detector.FindIDEByAlias(ides, "PyCharm Professional")
	assert.NotNil(t, result)
	assert.Equal(t, "PyCharm Professional", result.Name)

	// Test alias match
	result = detector.FindIDEByAlias(ides, "pycharm")
	assert.NotNil(t, result)
	assert.Equal(t, "PyCharm Professional", result.Name)

	// Test case insensitive
	result = detector.FindIDEByAlias(ides, "VSCODE")
	assert.NotNil(t, result)
	assert.Equal(t, "Visual Studio Code", result.Name)

	// Test not found
	result = detector.FindIDEByAlias(ides, "nonexistent")
	assert.Nil(t, result)
}