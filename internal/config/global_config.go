// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// GlobalConfig represents the global configuration for the application.
type GlobalConfig struct {
	Logging GlobalLoggingConfig `yaml:"logging" json:"logging"`
}

// GlobalLoggingConfig represents global logging configuration.
type GlobalLoggingConfig struct {
	Enabled   bool   `yaml:"enabled" json:"enabled"`
	FilePath  string `yaml:"filePath" json:"filePath"`
	Level     string `yaml:"level" json:"level"`
	MaxSizeMB int    `yaml:"maxSizeMB" json:"maxSizeMB"`
	MaxFiles  int    `yaml:"maxFiles" json:"maxFiles"`
}

// DefaultGlobalConfig returns the default global configuration.
func DefaultGlobalConfig() *GlobalConfig {
	homeDir, _ := os.UserHomeDir()
	defaultLogPath := filepath.Join(homeDir, ".scripton", "gzh", "logs", "gzh.log")
	
	return &GlobalConfig{
		Logging: GlobalLoggingConfig{
			Enabled:   false, // Default disabled
			FilePath:  defaultLogPath,
			Level:     "info",
			MaxSizeMB: 100,
			MaxFiles:  5,
		},
	}
}

// LoadGlobalConfig loads global configuration from the standard location.
func LoadGlobalConfig() (*GlobalConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return DefaultGlobalConfig(), nil
	}

	configPath := filepath.Join(homeDir, ".scripton", "gzh", "config.yaml")
	
	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultGlobalConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultGlobalConfig(), nil
	}

	var config GlobalConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return DefaultGlobalConfig(), nil
	}

	// Merge with defaults for missing fields
	defaultConfig := DefaultGlobalConfig()
	if config.Logging.FilePath == "" {
		config.Logging.FilePath = defaultConfig.Logging.FilePath
	}
	if config.Logging.Level == "" {
		config.Logging.Level = defaultConfig.Logging.Level
	}
	if config.Logging.MaxSizeMB == 0 {
		config.Logging.MaxSizeMB = defaultConfig.Logging.MaxSizeMB
	}
	if config.Logging.MaxFiles == 0 {
		config.Logging.MaxFiles = defaultConfig.Logging.MaxFiles
	}

	return &config, nil
}