// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"os"
	"path/filepath"
)

// GetConfigDirectory returns the configuration directory for net-env components
func GetConfigDirectory() string {
	if configDir := os.Getenv("GZH_CONFIG_DIR"); configDir != "" {
		return configDir
	}

	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "gzh-manager")
}

// EnsureConfigDirectory creates the configuration directory if it doesn't exist
func EnsureConfigDirectory() error {
	configDir := GetConfigDirectory()
	return os.MkdirAll(configDir, 0o755)
}

// GetProfilesPath returns the path to the network profiles configuration file
func GetProfilesPath() string {
	return filepath.Join(GetConfigDirectory(), "network-profiles.yaml")
}

// GetMetricsPath returns the path to the metrics storage directory
func GetMetricsPath() string {
	return filepath.Join(GetConfigDirectory(), "metrics")
}

// GetCachePath returns the path to the cache directory
func GetCachePath() string {
	return filepath.Join(GetConfigDirectory(), "cache")
}
