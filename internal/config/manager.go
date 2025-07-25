// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package config provides unified configuration management for all gzh-manager components.
// It supports multiple configuration sources, validation, hot-reloading, and environment-specific overrides.
package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/gizzahub/gzh-manager-go/internal/constants"
	"github.com/gizzahub/gzh-manager-go/internal/errors"
)

// Source represents different configuration sources.
type Source string

// Configuration source constants.
const (
	// SourceEnvironment represents configuration from environment variables.
	SourceEnvironment Source = "environment"
	// SourceFile represents configuration from files.
	SourceFile        Source = "file"
	// SourceDefaults represents default configuration values.
	SourceDefaults    Source = "defaults"
	// SourceAPI represents configuration from external APIs.
	SourceAPI         Source = "api"
)

// Manager provides centralized configuration management.
type Manager struct {
	configs       map[string]interface{}
	sources       []Source
	watchers      map[string][]Watcher
	validators    map[string]Validator
	mu            sync.RWMutex
	watcherCtx    context.Context
	watcherCancel context.CancelFunc
}

// Watcher defines the interface for configuration change notifications.
type Watcher interface {
	OnConfigChanged(key string, oldValue, newValue interface{}) error
}

// Validator defines the interface for configuration validation.
type Validator interface {
	Validate(config interface{}) error
}

// Options provides options for configuration loading.
type Options struct {
	Sources         []Source
	ConfigPaths     []string
	EnvPrefix       string
	WatchForChanges bool
	ValidateOnLoad  bool
	DefaultsOnly    bool
}

// NewManager creates a new configuration manager.
func NewManager(options Options) *Manager {
	if len(options.Sources) == 0 {
		options.Sources = []Source{SourceDefaults, SourceFile, SourceEnvironment}
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		configs:       make(map[string]interface{}),
		sources:       options.Sources,
		watchers:      make(map[string][]Watcher),
		validators:    make(map[string]Validator),
		watcherCtx:    ctx,
		watcherCancel: cancel,
	}
}

// LoadConfiguration loads configuration for a specific component.
func (cm *Manager) LoadConfiguration(ctx context.Context, key string, target interface{}, options Options) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Load from all sources in priority order
	merged := make(map[string]interface{})

	for _, source := range cm.sources {
		config, err := cm.loadFromSource(ctx, source, key, options)
		if err != nil {
			return errors.WrapError(err, errors.ErrorCodeConfigNotFound,
				fmt.Sprintf("failed to load config from %s", source), errors.SeverityMedium)
		}

		if config != nil {
			cm.mergeConfig(merged, config)
		}
	}

	if len(merged) == 0 && !options.DefaultsOnly {
		return errors.NewStandardError(errors.ErrorCodeConfigNotFound,
			fmt.Sprintf("no configuration found for key: %s", key), errors.SeverityHigh)
	}

	// Convert merged config to target type
	if err := cm.convertConfig(merged, target); err != nil {
		return errors.WrapError(err, errors.ErrorCodeInvalidConfig,
			"failed to convert configuration", errors.SeverityHigh)
	}

	// Validate configuration if validator is registered
	if validator, exists := cm.validators[key]; exists && options.ValidateOnLoad {
		if err := validator.Validate(target); err != nil {
			return errors.WrapError(err, errors.ErrorCodeValidationFailed,
				"configuration validation failed", errors.SeverityHigh)
		}
	}

	// Store configuration
	cm.configs[key] = target

	// Start watching for changes if enabled
	if options.WatchForChanges {
		go cm.watchConfiguration(key, options)
	}

	return nil
}

// RegisterValidator registers a configuration validator for a specific key.
func (cm *Manager) RegisterValidator(key string, validator Validator) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.validators[key] = validator
}

// RegisterWatcher registers a configuration change watcher.
func (cm *Manager) RegisterWatcher(key string, watcher Watcher) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.watchers[key] = append(cm.watchers[key], watcher)
}

// GetConfiguration retrieves stored configuration.
func (cm *Manager) GetConfiguration(key string) (interface{}, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	config, exists := cm.configs[key]
	return config, exists
}

// UpdateConfiguration updates configuration and notifies watchers.
func (cm *Manager) UpdateConfiguration(key string, newConfig interface{}) error {
	cm.mu.Lock()
	oldConfig, exists := cm.configs[key]
	cm.configs[key] = newConfig
	watchers := append([]Watcher(nil), cm.watchers[key]...)
	cm.mu.Unlock()

	// Notify watchers outside of lock
	if exists {
		for _, watcher := range watchers {
			if err := watcher.OnConfigChanged(key, oldConfig, newConfig); err != nil {
				return errors.WrapError(err, errors.ErrorCodeOperationFailed,
					"configuration watcher failed", errors.SeverityMedium)
			}
		}
	}

	return nil
}

// ReloadConfiguration reloads configuration from all sources.
func (cm *Manager) ReloadConfiguration(ctx context.Context, key string, options Options) error {
	target := cm.configs[key]
	if target == nil {
		return errors.NewStandardError(errors.ErrorCodeConfigNotFound,
			fmt.Sprintf("configuration not found for key: %s", key), errors.SeverityMedium)
	}

	return cm.LoadConfiguration(ctx, key, target, options)
}

// Shutdown gracefully shuts down the configuration manager.
func (cm *Manager) Shutdown() error {
	cm.watcherCancel()
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Clear all configurations
	cm.configs = make(map[string]interface{})
	cm.watchers = make(map[string][]Watcher)
	cm.validators = make(map[string]Validator)

	return nil
}

// loadFromSource loads configuration from a specific source.
func (cm *Manager) loadFromSource(ctx context.Context, source Source, key string, options Options) (map[string]interface{}, error) {
	switch source {
	case SourceDefaults:
		return cm.loadDefaults(key)
	case SourceFile:
		return cm.loadFromFile(key, options.ConfigPaths)
	case SourceEnvironment:
		return cm.loadFromEnvironment(key, options.EnvPrefix)
	case SourceAPI:
		return cm.loadFromAPI(ctx, key)
	default:
		return nil, fmt.Errorf("unknown configuration source: %s", source)
	}
}

// loadDefaults loads default configuration values.
func (cm *Manager) loadDefaults(key string) (map[string]interface{}, error) {
	defaults := map[string]map[string]interface{}{
		"bulk-clone": {
			"concurrency":       constants.DefaultConcurrency,
			"timeout":           constants.DefaultHTTPTimeout,
			"retry_attempts":    constants.DefaultRetryAttempts,
			"retry_delay":       constants.DefaultRetryDelay,
			"buffer_size":       constants.DefaultBufferSize,
			"max_repositories":  constants.MaxRepositories,
			"enable_streaming":  true,
			"progress_interval": constants.ProgressUpdateInterval,
		},
		"http-client": {
			"timeout":                 constants.DefaultHTTPTimeout,
			"max_idle_conns":          constants.DefaultMaxIdleConns,
			"idle_conn_timeout":       constants.DefaultIdleConnTimeout,
			"tls_handshake_timeout":   constants.DefaultTLSHandshakeTimeout,
			"expect_continue_timeout": constants.DefaultExpectContinueTimeout,
			"min_tls_version":         "1.2",
			"user_agent":              constants.DefaultUserAgent,
		},
		"worker-pool": {
			"clone_workers":     10,
			"update_workers":    15,
			"config_workers":    5,
			"operation_timeout": 5 * time.Minute,
			"retry_attempts":    3,
			"retry_delay":       2 * time.Second,
		},
		"auth": {
			"token_min_length":   constants.MinTokenLength,
			"validation_timeout": constants.MediumHTTPTimeout,
			"cache_duration":     15 * time.Minute,
			"max_retry_attempts": 3,
		},
		"logging": {
			"level":       "info",
			"format":      "json",
			"output":      "stdout",
			"max_size":    100, // MB
			"max_backups": 5,
			"max_age":     30, // days
		},
	}

	if config, exists := defaults[key]; exists {
		return config, nil
	}

	return make(map[string]interface{}), nil
}

// loadFromFile loads configuration from files.
func (cm *Manager) loadFromFile(key string, configPaths []string) (map[string]interface{}, error) {
	if len(configPaths) == 0 {
		configPaths = cm.getDefaultConfigPaths(key)
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		config := make(map[string]interface{})
		ext := strings.ToLower(filepath.Ext(path))

		switch ext {
		case ".yaml", ".yml":
			if err := yaml.Unmarshal(data, &config); err != nil {
				continue
			}
		case ".json":
			if err := json.Unmarshal(data, &config); err != nil {
				continue
			}
		default:
			continue
		}

		return config, nil
	}

	return make(map[string]interface{}), nil
}

// loadFromEnvironment loads configuration from environment variables.
func (cm *Manager) loadFromEnvironment(key, prefix string) (map[string]interface{}, error) {
	if prefix == "" {
		prefix = "GZH_"
	}

	config := make(map[string]interface{})
	envPrefix := strings.ToUpper(prefix + key + "_")

	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, envPrefix) {
			continue
		}

		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		envKey := strings.ToLower(strings.TrimPrefix(parts[0], envPrefix))
		value := parts[1]

		// Convert common types
		if value == "true" || value == "false" {
			config[envKey] = value == "true"
		} else if duration, err := time.ParseDuration(value); err == nil {
			config[envKey] = duration
		} else {
			config[envKey] = value
		}
	}

	return config, nil
}

// loadFromAPI loads configuration from API endpoints.
func (cm *Manager) loadFromAPI(_ context.Context, _ string) (map[string]interface{}, error) {
	// This would implement API-based configuration loading
	// For now, return empty config
	return make(map[string]interface{}), nil
}

// getDefaultConfigPaths returns default configuration file paths.
func (cm *Manager) getDefaultConfigPaths(key string) []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return []string{} // Return empty paths if home directory cannot be determined
	}

	return []string{
		fmt.Sprintf("./%s.yaml", key),
		fmt.Sprintf("./%s.yml", key),
		fmt.Sprintf("./%s.json", key),
		fmt.Sprintf("%s/.config/gzh-manager/%s.yaml", homeDir, key),
		fmt.Sprintf("%s/.config/gzh-manager/%s.yml", homeDir, key),
		fmt.Sprintf("/etc/gzh-manager/%s.yaml", key),
		fmt.Sprintf("/etc/gzh-manager/%s.yml", key),
	}
}

// mergeConfig merges source configuration into target.
func (cm *Manager) mergeConfig(target, source map[string]interface{}) {
	for key, value := range source {
		if existingValue, exists := target[key]; exists {
			if existingMap, ok := existingValue.(map[string]interface{}); ok {
				if sourceMap, ok := value.(map[string]interface{}); ok {
					cm.mergeConfig(existingMap, sourceMap)
					continue
				}
			}
		}
		target[key] = value
	}
}

// convertConfig converts map to target struct using JSON marshaling.
func (cm *Manager) convertConfig(source map[string]interface{}, target interface{}) error {
	data, err := json.Marshal(source)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// watchConfiguration watches for configuration changes.
func (cm *Manager) watchConfiguration(key string, options Options) {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-cm.watcherCtx.Done():
			return
		case <-ticker.C:
			// In a real implementation, this would use file system watchers
			// For now, we just reload periodically
			if err := cm.ReloadConfiguration(cm.watcherCtx, key, options); err != nil {
				// Log error but continue watching
				continue
			}
		}
	}
}
