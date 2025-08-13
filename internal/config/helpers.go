// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Gizzahub/gzh-manager-go/pkg/config"
)

// LoadCommandConfig provides a unified way to load configuration for commands
// It follows the standard precedence: explicit path > env var > default locations.
func LoadCommandConfig(ctx context.Context, configPath, configType string) (*config.UnifiedConfig, error) {
	// 1. Use explicit config path if provided
	if configPath != "" {
		return loadConfigFromPath(ctx, configPath)
	}

	// 2. Check environment variable
	envVar := fmt.Sprintf("GZH_%s_CONFIG", strings.ToUpper(strings.ReplaceAll(configType, "-", "_")))
	if envPath := os.Getenv(envVar); envPath != "" {
		return loadConfigFromPath(ctx, envPath)
	}

	// 3. Check standard locations
	configName := fmt.Sprintf("%s.yaml", configType)
	searchPaths := []string{
		// Current directory
		configName,
		fmt.Sprintf("%s.yml", configType),

		// User config directory
		filepath.Join(os.Getenv("HOME"), ".config", "gzh-manager", configName),
		filepath.Join(os.Getenv("HOME"), ".config", "gzh-manager", fmt.Sprintf("%s.yml", configType)),

		// System config directory
		filepath.Join(string(filepath.Separator), "etc", "gzh-manager", configName),
		filepath.Join(string(filepath.Separator), "etc", "gzh-manager", fmt.Sprintf("%s.yml", configType)),
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return loadConfigFromPath(ctx, path)
		}
	}

	// No config found - return empty config with defaults
	return &config.UnifiedConfig{
		Version: "1.0",
		Global: &config.GlobalSettings{
			DefaultStrategy: "reset",
		},
		DefaultProvider: "github",
		Providers:       make(map[string]*config.ProviderConfig),
	}, nil
}

// loadConfigFromPath loads configuration from a specific path.
func loadConfigFromPath(_ context.Context, path string) (*config.UnifiedConfig, error) {
	// Use unified config loader
	loader := config.NewUnifiedLoader()

	result, err := loader.LoadConfigFromPath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from %s: %w", path, err)
	}

	if result.Config == nil {
		return nil, fmt.Errorf("no valid configuration found in %s", path)
	}

	return result.Config, nil
}

// GetConfiguredProvider returns the provider configuration for the specified provider type.
func GetConfiguredProvider(cfg *config.UnifiedConfig, providerType string) (*config.ProviderConfig, error) {
	provider, exists := cfg.Providers[providerType]
	if !exists {
		return nil, fmt.Errorf("provider '%s' not configured", providerType)
	}

	return provider, nil
}

// GetConfiguredOrganization returns the organization configuration for the specified provider and org.
func GetConfiguredOrganization(cfg *config.UnifiedConfig, providerType, orgName string) (*config.GitTarget, error) {
	provider, err := GetConfiguredProvider(cfg, providerType)
	if err != nil {
		return nil, err
	}

	for _, org := range provider.Organizations {
		if org.Name == orgName {
			// Convert OrganizationConfig to GitTarget
			return &config.GitTarget{
				Name:     org.Name,
				CloneDir: org.CloneDir,
			}, nil
		}
	}

	return nil, fmt.Errorf("organization '%s' not found in %s provider configuration", orgName, providerType)
}

// MergeConfigWithFlags merges CLI flags with configuration file values
// CLI flags take precedence over config file values.
func MergeConfigWithFlags(_ *config.UnifiedConfig, _ map[string]interface{}) {
	// This is a placeholder for flag merging logic
	// Each command would pass its flags here to override config values
}
