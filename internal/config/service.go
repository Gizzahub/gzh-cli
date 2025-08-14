// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"github.com/Gizzahub/gzh-cli/internal/env"
	"github.com/Gizzahub/gzh-cli/pkg/config"
)

// ConfigService provides centralized configuration management.
type ConfigService interface { //nolint:revive // Interface name maintained for clarity in configuration service API
	// LoadConfiguration loads configuration from the specified path or default locations
	LoadConfiguration(ctx context.Context, configPath string) (*config.UnifiedConfig, error)

	// GetConfiguration returns the currently loaded configuration
	GetConfiguration() *config.UnifiedConfig

	// ReloadConfiguration reloads configuration from disk
	ReloadConfiguration(ctx context.Context) error

	// SaveConfiguration saves configuration to disk
	SaveConfiguration(ctx context.Context, cfg *config.UnifiedConfig, path string) error

	// ValidateConfiguration validates the current configuration
	ValidateConfiguration(ctx context.Context) error

	// WatchConfiguration starts watching for configuration file changes
	WatchConfiguration(ctx context.Context, callback func(*config.UnifiedConfig)) error

	// StopWatching stops watching for configuration changes
	StopWatching()

	// GetConfigPath returns the path of the currently loaded configuration
	GetConfigPath() string

	// IsLoaded returns true if configuration is loaded
	IsLoaded() bool

	// GetWarnings returns any warnings from configuration loading
	GetWarnings() []string

	// GetRequiredActions returns any required actions from configuration loading
	GetRequiredActions() []string

	// GetBulkCloneTargets returns bulk clone targets for integration
	GetBulkCloneTargets(ctx context.Context, providerFilter string) ([]config.BulkCloneTarget, error)

	// GetValidationResult returns the latest validation result
	GetValidationResult() *config.StartupValidationResult
}

// DefaultConfigService implements ConfigService using Viper.
type DefaultConfigService struct {
	mu                sync.RWMutex
	config            *config.UnifiedConfig
	configPath        string
	viper             *viper.Viper
	watcher           *fsnotify.Watcher
	environment       env.Environment
	unifiedFacade     *config.UnifiedConfigFacade
	startupValidator  *config.StartupValidator
	watchCallback     func(*config.UnifiedConfig)
	watchingEnabled   bool
	validationEnabled bool
}

// Ensure DefaultConfigService implements ConfigService interface.
var _ ConfigService = (*DefaultConfigService)(nil)

// ConfigServiceOptions provides configuration options for the service.
type ConfigServiceOptions struct { //nolint:revive // Type name maintained for clarity in service configuration
	Environment       env.Environment
	AutoMigrate       bool
	WatchEnabled      bool
	ValidationEnabled bool
	SearchPaths       []string
	ConfigName        string
	ConfigTypes       []string
}

// DefaultConfigServiceOptions returns default configuration service options.
func DefaultConfigServiceOptions() *ConfigServiceOptions {
	return &ConfigServiceOptions{
		Environment:       env.NewOSEnvironment(),
		AutoMigrate:       true,
		WatchEnabled:      true,
		ValidationEnabled: true,
		SearchPaths: []string{
			".",
			"$HOME/.config/gzh-manager",
			"/etc/gzh-manager",
		},
		ConfigName:  "gzh",
		ConfigTypes: []string{"yaml", "yml"},
	}
}

// NewConfigService creates a new configuration service.
func NewConfigService(options *ConfigServiceOptions) (ConfigService, error) {
	if options == nil {
		options = DefaultConfigServiceOptions()
	}

	v := viper.New()

	// Configure Viper
	v.SetConfigName(options.ConfigName)
	v.SetConfigType("yaml")

	// Add search paths
	for _, path := range options.SearchPaths {
		expandedPath := options.Environment.Expand(path)
		v.AddConfigPath(expandedPath)
	}

	// Set environment variable prefix
	v.SetEnvPrefix("GZH")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Create unified facade
	unifiedFacade := config.NewUnifiedConfigFacade()
	unifiedFacade.SetAutoMigrate(options.AutoMigrate)

	service := &DefaultConfigService{
		viper:             v,
		environment:       options.Environment,
		unifiedFacade:     unifiedFacade,
		startupValidator:  config.NewStartupValidator(),
		watchingEnabled:   options.WatchEnabled,
		validationEnabled: options.ValidationEnabled,
	}

	return service, nil
}

// LoadConfiguration loads configuration from the specified path or default locations.
func (s *DefaultConfigService) LoadConfiguration(_ context.Context, configPath string) (*config.UnifiedConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var err error

	if configPath != "" {
		// Load from specific path
		s.configPath = configPath
		err = s.unifiedFacade.LoadConfigurationFromPath(configPath)
	} else {
		// Load from default locations using Viper search paths
		configFile, findErr := s.findConfigFile()
		if findErr != nil {
			// Try to load using unified facade auto-discovery
			err = s.unifiedFacade.LoadConfiguration()
		} else {
			s.configPath = configFile
			err = s.unifiedFacade.LoadConfigurationFromPath(configFile)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	s.config = s.unifiedFacade.GetConfiguration()

	// Perform startup validation if enabled
	if s.validationEnabled {
		if err := s.performStartupValidation(); err != nil {
			return nil, fmt.Errorf("startup validation failed: %w", err)
		}
	}

	return s.config, nil
}

// GetConfiguration returns the currently loaded configuration.
func (s *DefaultConfigService) GetConfiguration() *config.UnifiedConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.config
}

// ReloadConfiguration reloads configuration from disk.
func (s *DefaultConfigService) ReloadConfiguration(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.configPath == "" {
		return fmt.Errorf("no configuration path set, cannot reload")
	}

	err := s.unifiedFacade.LoadConfigurationFromPath(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to reload configuration: %w", err)
	}

	s.config = s.unifiedFacade.GetConfiguration()

	// Validate reloaded configuration
	if err := s.validateConfig(); err != nil {
		return fmt.Errorf("reloaded configuration validation failed: %w", err)
	}

	return nil
}

// SaveConfiguration saves configuration to disk.
func (s *DefaultConfigService) SaveConfiguration(_ context.Context, cfg *config.UnifiedConfig, path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update internal state
	s.config = cfg
	s.configPath = path

	// Save using unified facade
	err := s.unifiedFacade.SaveConfiguration(path)
	if err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

// ValidateConfiguration validates the current configuration.
func (s *DefaultConfigService) ValidateConfiguration(_ context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.validationEnabled {
		return s.performStartupValidationLocked()
	}

	return s.validateConfig()
}

// validateConfig performs the actual validation (no locking).
func (s *DefaultConfigService) validateConfig() error {
	if s.config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	return s.unifiedFacade.ValidateConfiguration()
}

// performStartupValidation performs startup validation with locking.
func (s *DefaultConfigService) performStartupValidation() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.performStartupValidationLocked()
}

// performStartupValidationLocked performs startup validation without locking.
func (s *DefaultConfigService) performStartupValidationLocked() error {
	if s.config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	if s.startupValidator == nil {
		return fmt.Errorf("startup validator not initialized")
	}

	result := s.startupValidator.ValidateUnifiedConfig(s.config)

	// Log warnings (could be configured to go to logger)
	for _, warning := range result.Warnings {
		fmt.Printf("⚠ Configuration Warning [%s]: %s\n", warning.Field, warning.Message)
	}

	// Return error if validation failed
	if !result.IsValid {
		errorMessages := make([]string, len(result.Errors))
		for i, err := range result.Errors {
			errorMessages[i] = fmt.Sprintf("[%s] %s", err.Field, err.Message)
		}

		return fmt.Errorf("configuration validation failed:\n%s", strings.Join(errorMessages, "\n"))
	}

	return nil
}

// WatchConfiguration starts watching for configuration file changes.
func (s *DefaultConfigService) WatchConfiguration(ctx context.Context, callback func(*config.UnifiedConfig)) error {
	if !s.watchingEnabled {
		return fmt.Errorf("configuration watching is disabled")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.configPath == "" {
		return fmt.Errorf("no configuration file to watch")
	}

	if s.watcher != nil {
		return fmt.Errorf("already watching configuration file")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	s.watcher = watcher
	s.watchCallback = callback

	// Add the configuration file to the watcher
	err = s.watcher.Add(s.configPath)
	if err != nil {
		if err := s.watcher.Close(); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: failed to close watcher: %v\n", err)
		}
		s.watcher = nil

		return fmt.Errorf("failed to watch configuration file: %w", err)
	}

	// Start watching in a goroutine
	go s.watchLoop(ctx)

	return nil
}

// watchLoop handles file system events.
func (s *DefaultConfigService) watchLoop(ctx context.Context) {
	defer func() {
		if s.watcher != nil {
			if err := s.watcher.Close(); err != nil {
				// Log error but don't fail the operation
				fmt.Printf("Warning: failed to close watcher during shutdown: %v\n", err)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				s.handleConfigChange(ctx)
			}
		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}

			fmt.Printf("Configuration watcher error: %v\n", err)
		}
	}
}

// handleConfigChange handles configuration file changes.
func (s *DefaultConfigService) handleConfigChange(ctx context.Context) {
	// Add a small delay to avoid multiple rapid changes
	time.Sleep(100 * time.Millisecond)

	err := s.ReloadConfiguration(ctx)
	if err != nil {
		// Log error but continue watching - configuration might be temporarily invalid during editing
		fmt.Printf("⚠️  Configuration reload failed: %v\n", err)
		// Still call callback with current config so watchers know about the attempt
		if s.watchCallback != nil && s.config != nil {
			s.watchCallback(s.config)
		}

		return
	}

	// Configuration reloaded successfully
	if s.watchCallback != nil {
		s.watchCallback(s.config)
	}
}

// StopWatching stops watching for configuration changes.
func (s *DefaultConfigService) StopWatching() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.watcher != nil {
		if err := s.watcher.Close(); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: failed to close watcher: %v\n", err)
		}
		s.watcher = nil
		s.watchCallback = nil
	}
}

// GetConfigPath returns the path of the currently loaded configuration.
func (s *DefaultConfigService) GetConfigPath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.configPath
}

// IsLoaded returns true if configuration is loaded.
func (s *DefaultConfigService) IsLoaded() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.config != nil
}

// findConfigFile uses Viper to find configuration file.
func (s *DefaultConfigService) findConfigFile() (string, error) {
	// Try reading configuration to trigger file discovery
	err := s.viper.ReadInConfig()
	if err != nil {
		return "", err
	}

	return s.viper.ConfigFileUsed(), nil
}

// GetMigrationInfo returns migration information if available.
func (s *DefaultConfigService) GetMigrationInfo() *config.MigrationResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.unifiedFacade == nil {
		return nil
	}

	loadResult := s.unifiedFacade.GetLoadResult()
	if loadResult == nil || !loadResult.WasMigrated {
		return nil
	}

	return &config.MigrationResult{
		Success:         true,
		SourcePath:      loadResult.ConfigPath,
		TargetPath:      loadResult.MigrationPath,
		Warnings:        loadResult.Warnings,
		RequiredActions: loadResult.RequiredActions,
	}
}

// GetWarnings returns any warnings from configuration loading.
func (s *DefaultConfigService) GetWarnings() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.unifiedFacade == nil {
		return []string{}
	}

	return s.unifiedFacade.GetWarnings()
}

// GetRequiredActions returns any required actions from configuration loading.
func (s *DefaultConfigService) GetRequiredActions() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.unifiedFacade == nil {
		return []string{}
	}

	return s.unifiedFacade.GetRequiredActions()
}

// GetValidationResult returns the latest validation result.
func (s *DefaultConfigService) GetValidationResult() *config.StartupValidationResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.config == nil || s.startupValidator == nil {
		return &config.StartupValidationResult{
			IsValid: false,
			Errors:  []config.StartupValidationError{{Field: "config", Message: "No configuration loaded"}},
			Summary: "No configuration available for validation",
		}
	}

	return s.startupValidator.ValidateUnifiedConfig(s.config)
}

// IsValidationEnabled returns whether startup validation is enabled.
func (s *DefaultConfigService) IsValidationEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.validationEnabled
}

// SetValidationEnabled enables or disables startup validation.
func (s *DefaultConfigService) SetValidationEnabled(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.validationEnabled = enabled
}

// CreateDefaultConfiguration creates a default configuration file.
func (s *DefaultConfigService) CreateDefaultConfiguration(_ context.Context, path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.unifiedFacade.CreateDefaultConfiguration(path)
	if err != nil {
		return fmt.Errorf("failed to create default configuration: %w", err)
	}

	s.configPath = path

	return s.unifiedFacade.LoadConfigurationFromPath(path)
}

// GetBulkCloneTargets returns bulk clone targets for integration.
func (s *DefaultConfigService) GetBulkCloneTargets(_ context.Context, providerFilter string) ([]config.BulkCloneTarget, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.config == nil {
		return nil, fmt.Errorf("no configuration loaded")
	}

	if providerFilter != "" {
		return s.unifiedFacade.GetProviderTargets(providerFilter)
	}

	return s.unifiedFacade.GetAllTargets()
}

// GetConfiguredProviders returns all configured providers.
func (s *DefaultConfigService) GetConfiguredProviders() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.unifiedFacade == nil {
		return []string{}
	}

	return s.unifiedFacade.GetConfiguredProviders()
}

// GenerateReport generates a configuration report.
func (s *DefaultConfigService) GenerateReport(_ context.Context) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.unifiedFacade == nil {
		return "", fmt.Errorf("no configuration loaded")
	}

	return s.unifiedFacade.GenerateConfigurationReport()
}
