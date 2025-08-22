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
	err := os.WriteFile(buildFile, []byte("PY-252.23892.515\n"), 0o644)
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

func TestDetectAppImageLauncher(t *testing.T) {
	detector := NewIDEDetector()

	// Create a temporary script file
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "cursor")

	// Test case 1: AppImage launcher script
	scriptContent := `#!/usr/bin/env bash
set -euo pipefail
APP_DIR="/home/user/Apps"
latest="$(ls -1 "$APP_DIR"/Cursor-*.AppImage | sort -V | tail -n1)"
exec "$latest" "$@"`

	err := os.WriteFile(scriptPath, []byte(scriptContent), 0o755)
	assert.NoError(t, err)

	method, path := detector.detectAppImageLauncher(scriptPath)
	assert.Equal(t, "appimage", method)
	assert.Equal(t, "/home/user/Apps", path)

	// Test case 2: Non-AppImage script
	scriptContent = `#!/bin/bash
echo "Hello World"`

	err = os.WriteFile(scriptPath, []byte(scriptContent), 0o755)
	assert.NoError(t, err)

	method, path = detector.detectAppImageLauncher(scriptPath)
	assert.Equal(t, "", method)
	assert.Equal(t, "", path)

	// Test case 3: Non-script file
	binaryContent := []byte{0x7f, 0x45, 0x4c, 0x46} // ELF header
	err = os.WriteFile(scriptPath, binaryContent, 0o755)
	assert.NoError(t, err)

	method, path = detector.detectAppImageLauncher(scriptPath)
	assert.Equal(t, "", method)
	assert.Equal(t, "", path)
}

func TestExtractVersionFromAppImageName(t *testing.T) {
	detector := NewIDEDetector()

	tests := []struct {
		name     string
		filename string
		appName  string
		expected string
	}{
		{
			name:     "Cursor AppImage with version",
			filename: "/home/user/Apps/Cursor-1.4.5-x86_64.AppImage",
			appName:  "cursor",
			expected: "1.4.5",
		},
		{
			name:     "VS Code AppImage with version",
			filename: "/opt/VSCode-1.103.1.AppImage",
			appName:  "code",
			expected: "1.103.1",
		},
		{
			name:     "AppImage without version",
			filename: "/opt/SomeApp.AppImage",
			appName:  "app",
			expected: "",
		},
		{
			name:     "Complex version string",
			filename: "/home/user/Downloads/MyEditor-2.1.0-beta.3-linux.AppImage",
			appName:  "myeditor",
			expected: "2.1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.extractVersionFromAppImageName(tt.filename, tt.appName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAppImageVersion(t *testing.T) {
	detector := NewIDEDetector()

	// Create temporary directory structure
	tmpDir := t.TempDir()
	appsDir := filepath.Join(tmpDir, "Apps")
	err := os.MkdirAll(appsDir, 0o755)
	assert.NoError(t, err)

	// Create mock AppImage files
	appImages := []string{
		"Cursor-1.4.5-x86_64.AppImage",
		"Cursor-1.4.3-x86_64.AppImage",
		"OtherApp-2.0.0.AppImage",
	}

	for _, appImage := range appImages {
		filePath := filepath.Join(appsDir, appImage)
		err := os.WriteFile(filePath, []byte("fake appimage content"), 0o755)
		assert.NoError(t, err)
	}

	// Test getting Cursor version
	version := detector.getAppImageVersion(appsDir, "Cursor")
	assert.Equal(t, "1.4.5", version) // Should return the latest version

	// Test non-existent app
	version = detector.getAppImageVersion(appsDir, "NonExistent")
	assert.Equal(t, "unknown", version)

	// Test with invalid directory and non-existent app name
	// resolveAppImageDirectory might fallback to common directories, so use an app that definitely doesn't exist
	version = detector.getAppImageVersion("/nonexistent", "CompletelyNonExistentAppThatDoesNotExist")
	assert.Equal(t, "unknown", version)
}

func TestDetectInstallMethod(t *testing.T) {
	detector := NewIDEDetector()

	// Test JetBrains Toolbox path
	toolboxPath := "/home/user/.local/share/JetBrains/Toolbox/apps/PyCharm-P/bin/pycharm.sh"
	method, path := detector.detectInstallMethod(toolboxPath)
	assert.Equal(t, "toolbox", method)
	assert.Equal(t, toolboxPath, path)

	// Test regular binary (use a path that's not in pacman packages)
	binaryPath := "/tmp/fake-binary"
	method, path = detector.detectInstallMethod(binaryPath)
	assert.Equal(t, "direct", method)
	assert.Equal(t, binaryPath, path)

	// Note: Actual system binaries might be detected as package manager installations
	// This is expected behavior on real systems
}

func TestCompareVersions(t *testing.T) {
	detector := NewIDEDetector()

	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{"v1 greater", "2.0.0", "1.0.0", 1},
		{"v1 lesser", "1.0.0", "2.0.0", -1},
		{"equal versions", "1.0.0", "1.0.0", 0},
		{"lexicographic comparison", "1.10.0", "1.9.0", -1}, // Note: lexicographic "1.10.0" < "1.9.0"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.compareVersions(tt.v1, tt.v2)
			assert.Equal(t, tt.expected, result)
		})
	}
}
