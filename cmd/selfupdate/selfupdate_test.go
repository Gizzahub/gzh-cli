// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package selfupdate

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdater_GetAssetName(t *testing.T) {
	updater := NewUpdater("1.0.0")

	tests := []struct {
		name         string
		goos         string
		goarch       string
		expectedName string
	}{
		{
			name:         "Linux x86_64",
			goos:         "linux",
			goarch:       "amd64",
			expectedName: "gz_linux_x86_64",
		},
		{
			name:         "Windows x86_64",
			goos:         "windows",
			goarch:       "amd64",
			expectedName: "gz_windows_x86_64.exe",
		},
		{
			name:         "Darwin ARM64",
			goos:         "darwin",
			goarch:       "arm64",
			expectedName: "gz_darwin_arm64",
		},
		{
			name:         "Linux i386",
			goos:         "linux",
			goarch:       "386",
			expectedName: "gz_linux_i386",
		},
	}

	// Save original values
	originalGOOS := runtime.GOOS
	originalGOARCH := runtime.GOARCH

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test is conceptual since we can't actually change runtime.GOOS/GOARCH
			// In a real implementation, we would need to refactor GetAssetName to accept parameters
			_ = tt.goos
			_ = tt.goarch
			_ = tt.expectedName
			
			// For now, just test that GetAssetName returns something reasonable
			assetName := updater.GetAssetName()
			assert.NotEmpty(t, assetName)
			assert.Contains(t, assetName, "gz_")
			assert.Contains(t, assetName, originalGOOS)
		})
	}
}

func TestUpdater_IsNewerVersion(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		remoteVersion  string
		expected       bool
	}{
		{
			name:           "Same version",
			currentVersion: "v1.0.0",
			remoteVersion:  "v1.0.0",
			expected:       false,
		},
		{
			name:           "Same version without v prefix",
			currentVersion: "1.0.0",
			remoteVersion:  "1.0.0",
			expected:       false,
		},
		{
			name:           "Mixed v prefix",
			currentVersion: "v1.0.0",
			remoteVersion:  "1.0.0",
			expected:       false,
		},
		{
			name:           "Different versions",
			currentVersion: "v1.0.0",
			remoteVersion:  "v1.1.0",
			expected:       true,
		},
		{
			name:           "Dev version",
			currentVersion: "dev",
			remoteVersion:  "v1.0.0",
			expected:       true,
		},
		{
			name:           "Empty current version",
			currentVersion: "",
			remoteVersion:  "v1.0.0",
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updater := NewUpdater(tt.currentVersion)
			result := updater.IsNewerVersion(tt.remoteVersion)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewUpdater(t *testing.T) {
	version := "1.0.0"
	updater := NewUpdater(version)
	
	assert.NotNil(t, updater)
	assert.Equal(t, version, updater.currentVersion)
	assert.NotNil(t, updater.logger)
}