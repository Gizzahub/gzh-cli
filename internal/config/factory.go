// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gizzahub/gzh-cli/internal/env"
	"github.com/gizzahub/gzh-cli/internal/errors"
)

// ServiceFactory creates configuration service instances.
type ServiceFactory interface {
	// CreateConfigService creates a new configuration service instance
	CreateConfigService(options *ConfigServiceOptions) (ConfigService, error)

	// CreateDefaultConfigService creates a configuration service with default options
	CreateDefaultConfigService() (ConfigService, error)

	// CreateConfigServiceWithEnvironment creates a configuration service with custom environment
	CreateConfigServiceWithEnvironment(environment env.Environment) (ConfigService, error)
}

// DefaultServiceFactory implements ServiceFactory.
type DefaultServiceFactory struct{}

// NewServiceFactory creates a new configuration service factory.
func NewServiceFactory() ServiceFactory {
	return &DefaultServiceFactory{}
}

// CreateConfigService creates a new configuration service instance.
func (f *DefaultServiceFactory) CreateConfigService(options *ConfigServiceOptions) (ConfigService, error) {
	if options == nil {
		return nil, fmt.Errorf("configuration service options cannot be nil")
	}

	service, err := NewConfigService(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create configuration service: %w", err)
	}

	return service, nil
}

// CreateDefaultConfigService creates a configuration service with default options.
func (f *DefaultServiceFactory) CreateDefaultConfigService() (ConfigService, error) {
	options := DefaultConfigServiceOptions()
	return f.CreateConfigService(options)
}

// CreateConfigServiceWithEnvironment creates a configuration service with custom environment.
func (f *DefaultServiceFactory) CreateConfigServiceWithEnvironment(environment env.Environment) (ConfigService, error) {
	options := DefaultConfigServiceOptions()
	options.Environment = environment

	return f.CreateConfigService(options)
}

// Global factory instance for convenience.
var globalFactory = NewServiceFactory()

// CreateConfigService creates a configuration service using the global factory.
func CreateConfigService(options *ConfigServiceOptions) (ConfigService, error) {
	return globalFactory.CreateConfigService(options)
}

// CreateDefaultConfigService creates a configuration service with default options using the global factory.
func CreateDefaultConfigService() (ConfigService, error) {
	return globalFactory.CreateDefaultConfigService()
}

// CreateConfigServiceWithEnvironment creates a configuration service with custom environment using the global factory.
func CreateConfigServiceWithEnvironment(environment env.Environment) (ConfigService, error) {
	return globalFactory.CreateConfigServiceWithEnvironment(environment)
}

// SetGlobalFactory sets the global factory instance (useful for testing).
func SetGlobalFactory(factory ServiceFactory) {
	globalFactory = factory
}

// Unified Configuration Management Functions

// GlobalConfigManager is a singleton configuration manager instance.
var GlobalConfigManager *ConfigManager

// InitializeGlobalManager initializes the global configuration manager.
func InitializeGlobalManager() error {
	if GlobalConfigManager != nil {
		return nil // Already initialized
	}

	options := ConfigOptions{
		Sources:         []Source{SourceDefaults, SourceFile, SourceEnvironment},
		WatchForChanges: true,
		ValidateOnLoad:  true,
	}

	GlobalConfigManager = NewConfigManager(options)
	RegisterDefaultValidators(GlobalConfigManager)

	return nil
}

// GetGlobalManager returns the global configuration manager, initializing it if needed.
func GetGlobalManager() (*ConfigManager, error) {
	if GlobalConfigManager == nil {
		if err := InitializeGlobalManager(); err != nil {
			return nil, err
		}
	}
	return GlobalConfigManager, nil
}

// LoadConfig loads configuration for a specific component with automatic path discovery.
func LoadConfig[T any](ctx context.Context, component string, target *T) error {
	manager, err := GetGlobalManager()
	if err != nil {
		return err
	}

	options := ConfigOptions{
		Sources:         []Source{SourceDefaults, SourceFile, SourceEnvironment},
		ConfigPaths:     discoverConfigPaths(component),
		EnvPrefix:       "GZH_",
		WatchForChanges: false,
		ValidateOnLoad:  true,
	}

	return manager.LoadConfiguration(ctx, component, target, options)
}

// LoadConfigWithOptions loads configuration with custom options.
func LoadConfigWithOptions[T any](ctx context.Context, component string, target *T, options ConfigOptions) error {
	manager, err := GetGlobalManager()
	if err != nil {
		return err
	}

	if len(options.ConfigPaths) == 0 {
		options.ConfigPaths = discoverConfigPaths(component)
	}

	if options.EnvPrefix == "" {
		options.EnvPrefix = "GZH_"
	}

	return manager.LoadConfiguration(ctx, component, target, options)
}

// ReloadConfig reloads configuration for a specific component.
func ReloadConfig(ctx context.Context, component string) error {
	manager, err := GetGlobalManager()
	if err != nil {
		return err
	}

	options := ConfigOptions{
		Sources:         []Source{SourceDefaults, SourceFile, SourceEnvironment},
		ConfigPaths:     discoverConfigPaths(component),
		EnvPrefix:       "GZH_",
		WatchForChanges: false,
		ValidateOnLoad:  true,
	}

	return manager.ReloadConfiguration(ctx, component, options)
}

// GetConfig retrieves stored configuration.
func GetConfig[T any](component string) (*T, error) {
	manager, err := GetGlobalManager()
	if err != nil {
		return nil, err
	}

	config, exists := manager.GetConfiguration(component)
	if !exists {
		return nil, errors.NewStandardError(errors.ErrorCodeConfigNotFound,
			"configuration not found for component: "+component, errors.SeverityMedium)
	}

	typedConfig, ok := config.(*T)
	if !ok {
		return nil, errors.NewStandardError(errors.ErrorCodeInvalidConfig,
			"configuration type mismatch for component: "+component, errors.SeverityHigh)
	}

	return typedConfig, nil
}

// UpdateConfig updates configuration and notifies watchers.
func UpdateConfig(component string, config interface{}) error {
	manager, err := GetGlobalManager()
	if err != nil {
		return err
	}

	return manager.UpdateConfiguration(component, config)
}

// RegisterWatcher registers a configuration change watcher.
func RegisterWatcher(component string, watcher Watcher) error {
	manager, err := GetGlobalManager()
	if err != nil {
		return err
	}

	manager.RegisterWatcher(component, watcher)
	return nil
}

// RegisterValidator registers a configuration validator.
func RegisterValidator(component string, validator Validator) error {
	manager, err := GetGlobalManager()
	if err != nil {
		return err
	}

	manager.RegisterValidator(component, validator)
	return nil
}

// discoverConfigPaths discovers configuration file paths for a component.
func discoverConfigPaths(component string) []string {
	paths := make([]string, 0, 20) // Pre-allocate for expected number of paths

	// Environment variable override
	if envPath := env.GetConfigPath(strings.ToUpper(component)); envPath != "" {
		paths = append(paths, envPath)
	}

	// Current directory
	for _, ext := range []string{".yaml", ".yml", ".json"} {
		paths = append(paths, "./"+component+ext)
	}

	// User config directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		configDir := filepath.Join(homeDir, ".config", "gzh-manager")
		for _, ext := range []string{".yaml", ".yml", ".json"} {
			paths = append(paths, filepath.Join(configDir, component+ext))
		}
	}

	// System config directory
	systemConfigDir := "/etc/gzh-manager"
	for _, ext := range []string{".yaml", ".yml", ".json"} {
		paths = append(paths, filepath.Join(systemConfigDir, component+ext))
	}

	// Legacy bulk-clone configs
	if component == "bulk-clone" {
		legacyPaths := []string{
			"./bulk-clone.yaml",
			"./bulk-clone.yml",
			"./bulk-clone.json",
		}

		if homeDir, err := os.UserHomeDir(); err == nil {
			legacyPaths = append(legacyPaths,
				filepath.Join(homeDir, ".config", "gzh-manager", "bulk-clone.yaml"),
				filepath.Join(homeDir, ".config", "gzh-manager", "bulk-clone.yml"),
			)
		}

		legacyPaths = append(legacyPaths,
			"/etc/gzh-manager/bulk-clone.yaml",
			"/etc/gzh-manager/bulk-clone.yml",
		)

		paths = append(paths, legacyPaths...)
	}

	return paths
}

// Shutdown gracefully shuts down the global configuration manager.
func Shutdown() error {
	if GlobalConfigManager != nil {
		err := GlobalConfigManager.Shutdown()
		GlobalConfigManager = nil
		return err
	}
	return nil
}

// Helper functions for common configuration loading patterns.

// LoadBulkCloneConfig loads bulk clone configuration.
func LoadBulkCloneConfig(ctx context.Context) (*BulkCloneConfig, error) {
	var config BulkCloneConfig
	if err := LoadConfig(ctx, "bulk-clone", &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// LoadHTTPClientConfig loads HTTP client configuration.
func LoadHTTPClientConfig(ctx context.Context) (*HTTPClientConfig, error) {
	var config HTTPClientConfig
	if err := LoadConfig(ctx, "http-client", &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// LoadAuthConfig loads authentication configuration.
func LoadAuthConfig(ctx context.Context) (*AuthConfig, error) {
	var config AuthConfig
	if err := LoadConfig(ctx, "auth", &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// LoadGitHubConfig loads GitHub configuration.
func LoadGitHubConfig(ctx context.Context) (*GitHubConfig, error) {
	var config GitHubConfig
	if err := LoadConfig(ctx, "github", &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// LoadSecurityConfig loads security configuration.
func LoadSecurityConfig(ctx context.Context) (*SecurityConfig, error) {
	var config SecurityConfig
	if err := LoadConfig(ctx, "security", &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// LoadLoggingConfig loads logging configuration.
func LoadLoggingConfig(ctx context.Context) (*LoggingConfig, error) {
	var config LoggingConfig
	if err := LoadConfig(ctx, "logging", &config); err != nil {
		return nil, err
	}
	return &config, nil
}
