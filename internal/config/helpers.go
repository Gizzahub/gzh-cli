// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gizzahub/gzh-cli/pkg/config"
)

// LoadCommandConfig provides a unified way to load configuration for commands
// It follows the standard precedence: explicit path > env var > default locations.
//
// 맥락을 받지 않는다. 예전에는 ctx를 받아 loadConfigFromPath로 넘겼는데 그쪽이
// `_`로 버렸다. 취소할 수 있는 척만 하는 셈이라 부르는 쪽에서 넘긴
// context.Background()가 문제없어 보였다. 여기는 지역 파일을 읽고 끝나는
// 자리다. 나중에 원격 설정을 가져오게 되면 그때 맥락을 다시 받는다.
func LoadCommandConfig(configPath, configType string) (*config.UnifiedConfig, error) {
	// 1. Use explicit config path if provided
	if configPath != "" {
		return loadConfigFromPath(configPath)
	}

	// 2. Check environment variable
	envVar := fmt.Sprintf("GZH_%s_CONFIG", strings.ToUpper(strings.ReplaceAll(configType, "-", "_")))
	if envPath := os.Getenv(envVar); envPath != "" {
		return loadConfigFromPath(envPath)
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
			return loadConfigFromPath(path)
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
func loadConfigFromPath(path string) (*config.UnifiedConfig, error) {
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
func MergeConfigWithFlags(_ *config.UnifiedConfig, _ map[string]any) {
	// This is a placeholder for flag merging logic
	// Each command would pass its flags here to override config values
}
