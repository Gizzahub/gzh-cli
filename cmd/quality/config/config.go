// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"
)

// Config represents the quality command configuration
type Config struct {
	// DefaultWorkers sets the default number of parallel workers
	DefaultWorkers int `yaml:"default_workers"`

	// Timeout sets the default timeout for tool execution
	Timeout string `yaml:"timeout"`

	// Tools contains tool-specific configurations
	Tools map[string]ToolConfig `yaml:"tools"`

	// Languages contains language-specific configurations
	Languages map[string]LanguageConfig `yaml:"languages"`

	// Exclude contains patterns to exclude from processing
	Exclude []string `yaml:"exclude"`

	// Include contains patterns to include in processing
	Include []string `yaml:"include"`
}

// ToolConfig represents configuration for a specific tool
type ToolConfig struct {
	// Enabled controls whether the tool should be used
	Enabled bool `yaml:"enabled"`

	// ConfigFile specifies a custom config file path
	ConfigFile string `yaml:"config_file"`

	// Args contains additional arguments to pass to the tool
	Args []string `yaml:"args"`

	// Env contains environment variables for the tool
	Env map[string]string `yaml:"env"`

	// Priority affects execution order (higher = earlier)
	Priority int `yaml:"priority"`
}

// LanguageConfig represents configuration for a language
type LanguageConfig struct {
	// Enabled controls whether to process this language
	Enabled bool `yaml:"enabled"`

	// PreferredTools lists preferred tools for this language
	PreferredTools []string `yaml:"preferred_tools"`

	// Extensions lists file extensions for this language
	Extensions []string `yaml:"extensions"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultWorkers: 4,
		Timeout:        "10m",
		Tools: map[string]ToolConfig{
			"gofumpt": {
				Enabled:  true,
				Priority: 10,
			},
			"goimports": {
				Enabled:  true,
				Priority: 9,
			},
			"golangci-lint": {
				Enabled:  true,
				Priority: 5,
			},
			"black": {
				Enabled:  true,
				Priority: 10,
			},
			"ruff": {
				Enabled:  true,
				Priority: 7,
			},
			"pylint": {
				Enabled:  false, // Disabled by default as ruff is preferred
				Priority: 5,
			},
			"prettier": {
				Enabled:  true,
				Priority: 10,
			},
			"eslint": {
				Enabled:  true,
				Priority: 5,
			},
			"tsc": {
				Enabled:  true,
				Priority: 3,
			},
		},
		Languages: map[string]LanguageConfig{
			"Go": {
				Enabled:        true,
				PreferredTools: []string{"gofumpt", "goimports", "golangci-lint"},
				Extensions:     []string{".go"},
			},
			"Python": {
				Enabled:        true,
				PreferredTools: []string{"black", "ruff"},
				Extensions:     []string{".py", ".pyi"},
			},
			"JavaScript": {
				Enabled:        true,
				PreferredTools: []string{"prettier", "eslint"},
				Extensions:     []string{".js", ".jsx"},
			},
			"TypeScript": {
				Enabled:        true,
				PreferredTools: []string{"prettier", "eslint", "tsc"},
				Extensions:     []string{".ts", ".tsx"},
			},
		},
		Exclude: []string{
			"node_modules/**",
			"vendor/**",
			".git/**",
			"dist/**",
			"build/**",
			"**/*.min.js",
			"**/*.min.css",
		},
	}
}

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*Config, error) {
	// Start with default config
	config := DefaultConfig()

	// If no config file specified, try to find one
	if configPath == "" {
		configPath = FindConfigFile()
	}

	// If still no config file, return default
	if configPath == "" {
		return config, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	return config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, configPath string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write file
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// FindConfigFile searches for a quality config file in the current directory and up the directory tree
func FindConfigFile() string {
	configNames := []string{
		".gzquality.yml",
		".gzquality.yaml",
		"gzquality.yml",
		"gzquality.yaml",
	}

	// Start from current directory
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Search up the directory tree
	for {
		for _, name := range configNames {
			configPath := filepath.Join(dir, name)
			if _, err := os.Stat(configPath); err == nil {
				return configPath
			}
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	return ""
}

// GetToolConfig returns configuration for a specific tool
func (c *Config) GetToolConfig(toolName string) ToolConfig {
	if config, exists := c.Tools[toolName]; exists {
		return config
	}

	// Return default tool config
	return ToolConfig{
		Enabled:  true,
		Priority: 5,
	}
}

// GetLanguageConfig returns configuration for a specific language
func (c *Config) GetLanguageConfig(language string) LanguageConfig {
	if config, exists := c.Languages[language]; exists {
		return config
	}

	// Return default language config
	return LanguageConfig{
		Enabled: true,
	}
}

// IsToolEnabled checks if a tool is enabled
func (c *Config) IsToolEnabled(toolName string) bool {
	return c.GetToolConfig(toolName).Enabled
}

// IsLanguageEnabled checks if a language is enabled
func (c *Config) IsLanguageEnabled(language string) bool {
	return c.GetLanguageConfig(language).Enabled
}

// GetPreferredTools returns preferred tools for a language
func (c *Config) GetPreferredTools(language string) []string {
	return c.GetLanguageConfig(language).PreferredTools
}

// ShouldExclude checks if a file path should be excluded
func (c *Config) ShouldExclude(filePath string) bool {
	for _, pattern := range c.Exclude {
		if matched, _ := filepath.Match(pattern, filePath); matched {
			return true
		}

		// Also check if any parent directory matches
		dir := filepath.Dir(filePath)
		for dir != "." && dir != "/" {
			if matched, _ := filepath.Match(pattern, dir); matched {
				return true
			}
			dir = filepath.Dir(dir)
		}
	}

	return false
}

// ShouldInclude checks if a file path should be included
func (c *Config) ShouldInclude(filePath string) bool {
	// If no include patterns, include everything (subject to exclude)
	if len(c.Include) == 0 {
		return !c.ShouldExclude(filePath)
	}

	// Check include patterns
	for _, pattern := range c.Include {
		if matched, _ := filepath.Match(pattern, filePath); matched {
			return !c.ShouldExclude(filePath)
		}
	}

	return false
}
