// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package config provides configuration loading and management functionality.
// It supports loading configuration from multiple sources with a defined precedence:
// environment variables, configuration files in various standard locations,
// and default values.
package config

import (
	"os"
	"path/filepath"

	"github.com/gizzahub/gzh-cli/internal/env"
)

// convertUnifiedToLegacyConfig converts a UnifiedConfig to legacy Config format for backward compatibility.
func convertUnifiedToLegacyConfig(unified *UnifiedConfig) *Config {
	if unified == nil {
		return nil
	}

	config := &Config{
		Version:         unified.Version,
		DefaultProvider: unified.DefaultProvider,
		Providers:       make(map[string]Provider),
	}

	// Copy providers
	for name, provider := range unified.Providers {
		legacyProvider := Provider{
			Token: provider.Token,
		}

		// Convert organizations to GitTargets
		var orgs []GitTarget
		for _, org := range provider.Organizations {
			target := GitTarget{
				Name:       org.Name,
				Visibility: org.Visibility,
				CloneDir:   org.CloneDir,
				Strategy:   org.Strategy,
				Match:      org.Include,
				Exclude:    org.Exclude,
			}
			orgs = append(orgs, target)
		}

		// Set orgs or groups based on provider type
		if name == "gitlab" {
			legacyProvider.Groups = orgs
		} else {
			legacyProvider.Orgs = orgs
		}

		config.Providers[name] = legacyProvider
	}

	return config
}

// ConfigSearchPaths defines the search order for configuration files.
var ConfigSearchPaths = []string{
	"./gzh.yaml",
	"./gzh.yml",
	"~/.config/gzh.yaml",
	"~/.config/gzh.yml",
	"~/.config/gzh-manager/gzh.yaml",
	"~/.config/gzh-manager/gzh.yml",
	"/etc/gzh-manager/gzh.yaml",
	"/etc/gzh-manager/gzh.yml",
}

// For migration guidance, see: docs/migration-guides/config-loader-migration.md.
func LoadConfig() (*Config, error) {
	return LoadConfigWithEnv(env.NewOSEnvironment())
}

// LoadConfigWithEnv loads configuration using the provided environment.
// Deprecated: Use ConfigFactory.LoadConfigFromPath() instead for better dependency injection support.
func LoadConfigWithEnv(environment env.Environment) (*Config, error) {
	// Use unified factory for backward compatibility
	factory := NewConfigFactoryWithOptions(&ConfigFactoryOptions{
		Environment: environment,
	})

	unifiedConfig, err := factory.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Convert unified config to legacy Config struct for backward compatibility
	return convertUnifiedToLegacyConfig(unifiedConfig), nil
}

// LoadConfigFromFile loads configuration from a specific file.
// Deprecated: Use ConfigFactory.LoadConfigFromPath() instead for better dependency injection support.
func LoadConfigFromFile(filename string) (*Config, error) {
	return LoadConfigFromFileWithEnv(filename, env.NewOSEnvironment())
}

// LoadConfigFromFileWithEnv loads configuration from a specific file using the provided environment.
// Deprecated: Use ConfigFactory.LoadConfigFromPath() instead for better dependency injection support.
func LoadConfigFromFileWithEnv(filename string, environment env.Environment) (*Config, error) {
	// Use unified factory for backward compatibility
	factory := NewConfigFactoryWithOptions(&ConfigFactoryOptions{
		Environment: environment,
	})

	unifiedConfig, err := factory.LoadConfigFromPath(filename)
	if err != nil {
		return nil, err
	}

	// Convert unified config to legacy Config struct for backward compatibility
	return convertUnifiedToLegacyConfig(unifiedConfig), nil
}

// FindConfigFile finds the first available configuration file.
// Deprecated: Use ConfigFactory.FindConfigFile() instead for better dependency injection support.
func FindConfigFile() (string, error) {
	return FindConfigFileWithEnv(env.NewOSEnvironment())
}

// FindConfigFileWithEnv finds the first available configuration file using the provided environment.
// Deprecated: Use ConfigFactory.FindConfigFile() instead for better dependency injection support.
func FindConfigFileWithEnv(environment env.Environment) (string, error) {
	// Use unified factory for backward compatibility
	factory := NewConfigFactoryWithOptions(&ConfigFactoryOptions{
		Environment: environment,
	})

	return factory.FindConfigFile()
}

// GetConfigSearchPaths returns the list of paths where configuration files are searched.
func GetConfigSearchPaths() []string {
	paths := make([]string, len(ConfigSearchPaths))
	for i, path := range ConfigSearchPaths {
		paths[i] = expandPath(path)
	}

	return paths
}

// expandPath expands ~ to home directory and resolves relative paths.
func expandPath(path string) string {
	return expandPathWithEnv(path, env.NewOSEnvironment())
}

// expandPathWithEnv expands ~ to home directory and resolves relative paths using the provided environment.
func expandPathWithEnv(path string, environment env.Environment) string {
	if path != "" && path[0] == '~' {
		// Try to get home directory from environment first
		homeDir := environment.Get(env.CommonEnvironmentKeys.HomeDir)
		if homeDir == "" {
			// Fallback to os.UserHomeDir() for compatibility
			var err error

			homeDir, err = os.UserHomeDir()
			if err != nil {
				return path // Return original if we can't get home dir
			}
		}

		return filepath.Join(homeDir, path[1:])
	}

	// Convert to absolute path if relative
	if !filepath.IsAbs(path) {
		if abs, err := filepath.Abs(path); err == nil {
			return abs
		}
	}

	return path
}

// fileExists checks if a file exists and is readable.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

// CreateDefaultConfig creates a default configuration file at the specified path.
// Deprecated: Use ConfigFactory.CreateDefaultConfig() instead for better dependency injection support.
func CreateDefaultConfig(filename string) error {
	// Use unified factory for backward compatibility
	factory := NewConfigFactory()
	return factory.CreateDefaultConfig(filename)
}

// GetDefaultConfigPath returns the default path for creating new config files.
// Deprecated: Use ConfigFactory.GetDefaultConfigPath() instead for better dependency injection support.
func GetDefaultConfigPath() string {
	// Use unified factory for backward compatibility
	factory := NewConfigFactory()
	return factory.GetDefaultConfigPath()
}
