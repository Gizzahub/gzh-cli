// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package idecore

import (
	"os"
	"path/filepath"
)

// DetectorConfig holds configuration for IDE detector.
type DetectorConfig struct {
	CacheDir string
}

// NewDetectorConfig creates default detector configuration.
func NewDetectorConfig() *DetectorConfig {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home dir is unavailable
		homeDir = "."
	}
	cacheDir := filepath.Join(homeDir, ".gz", "cache")

	return &DetectorConfig{
		CacheDir: cacheDir,
	}
}

// GetDefaultCacheDir returns the default cache directory for IDE detection.
func GetDefaultCacheDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home dir is unavailable
		homeDir = "."
	}
	return filepath.Join(homeDir, ".gz", "cache")
}
