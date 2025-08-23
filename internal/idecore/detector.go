// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package idecore

import (
	"os"
	"path/filepath"
)

// DetectorConfig holds configuration for IDE detector
type DetectorConfig struct {
	CacheDir string
}

// NewDetectorConfig creates default detector configuration
func NewDetectorConfig() *DetectorConfig {
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".gz", "cache")

	return &DetectorConfig{
		CacheDir: cacheDir,
	}
}

// GetDefaultCacheDir returns the default cache directory for IDE detection
func GetDefaultCacheDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".gz", "cache")
}
