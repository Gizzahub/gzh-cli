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
	MaxSizeMB int    `yaml:"maxSizeMb" json:"maxSizeMb"`
	MaxFiles  int    `yaml:"maxFiles" json:"maxFiles"`
	// CLI-specific logging settings
	CLILogging CLILoggingConfig `yaml:"cli" json:"cliLogging"`
}

// CLILoggingConfig represents CLI-specific logging configuration.
type CLILoggingConfig struct {
	Enabled    bool   `yaml:"enabled" json:"enabled"`       // Show logs in CLI by default
	Level      string `yaml:"level" json:"level"`           // CLI log level (error, warn, info, debug)
	OnlyErrors bool   `yaml:"onlyErrors" json:"onlyErrors"` // Show only errors and warnings
	Quiet      bool   `yaml:"quiet" json:"quiet"`           // Suppress all logs except critical errors
}

// DefaultGlobalConfig returns the default global configuration.
func DefaultGlobalConfig() *GlobalConfig {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to /tmp if home directory cannot be determined
		homeDir = "/tmp"
	}
	defaultLogPath := filepath.Join(homeDir, ".scripton", "gzh", "logs", "gzh.log")

	return &GlobalConfig{
		Logging: GlobalLoggingConfig{
			Enabled:   false, // Default disabled
			FilePath:  defaultLogPath,
			Level:     "info",
			MaxSizeMB: 100,
			MaxFiles:  5,
			CLILogging: CLILoggingConfig{
				Enabled:    false,   // CLI logs disabled by default
				Level:      "error", // Only show errors by default
				OnlyErrors: true,    // Show only errors and warnings
				Quiet:      false,   // Don't suppress critical errors
			},
		},
	}
}

// LoadGlobalConfig loads global configuration from the standard location.
func LoadGlobalConfig() (*GlobalConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return DefaultGlobalConfig(), err
	}

	configPath := filepath.Join(homeDir, ".scripton", "gzh", "config.yaml")

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultGlobalConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultGlobalConfig(), err
	}

	var config GlobalConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return DefaultGlobalConfig(), err
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

	// Merge CLI logging defaults
	if config.Logging.CLILogging.Level == "" {
		config.Logging.CLILogging.Level = defaultConfig.Logging.CLILogging.Level
	}

	return &config, nil
}
