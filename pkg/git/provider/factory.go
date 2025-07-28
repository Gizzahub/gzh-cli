// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"sync"
)

// ProviderConfig represents configuration for a Git provider.
type ProviderConfig struct {
	Type      string                 `json:"type" yaml:"type"`
	Name      string                 `json:"name" yaml:"name"`
	BaseURL   string                 `json:"base_url" yaml:"base_url"`
	Token     string                 `json:"token" yaml:"token"`
	Username  string                 `json:"username" yaml:"username"`
	Password  string                 `json:"password" yaml:"password"`
	APIKey    string                 `json:"api_key" yaml:"api_key"`
	Enabled   bool                   `json:"enabled" yaml:"enabled"`
	Timeout   int                    `json:"timeout" yaml:"timeout"` // seconds
	RateLimit RateLimitConfig        `json:"rate_limit" yaml:"rate_limit"`
	Retry     RetryConfig            `json:"retry" yaml:"retry"`
	Features  map[string]bool        `json:"features" yaml:"features"`
	Extra     map[string]interface{} `json:"extra" yaml:"extra"`
}

// RateLimitConfig represents rate limiting configuration.
type RateLimitConfig struct {
	Enabled     bool `json:"enabled" yaml:"enabled"`
	RequestsPer int  `json:"requests_per" yaml:"requests_per"`
	Duration    int  `json:"duration" yaml:"duration"` // seconds
	Burst       int  `json:"burst" yaml:"burst"`
}

// RetryConfig represents retry configuration.
type RetryConfig struct {
	Enabled         bool     `json:"enabled" yaml:"enabled"`
	MaxAttempts     int      `json:"max_attempts" yaml:"max_attempts"`
	InitialDelay    int      `json:"initial_delay" yaml:"initial_delay"` // milliseconds
	MaxDelay        int      `json:"max_delay" yaml:"max_delay"`         // milliseconds
	BackoffFactor   float64  `json:"backoff_factor" yaml:"backoff_factor"`
	RetryableErrors []string `json:"retryable_errors" yaml:"retryable_errors"`
}

// ProviderFactory creates and manages Git provider instances.
type ProviderFactory struct {
	mu              sync.RWMutex
	configs         map[string]*ProviderConfig
	constructors    map[string]ProviderConstructor
	defaultSettings *DefaultSettings
}

// ProviderConstructor is a function that creates a provider instance.
type ProviderConstructor func(config *ProviderConfig) (GitProvider, error)

// DefaultSettings represents default settings applied to all providers.
type DefaultSettings struct {
	Timeout     int               `json:"timeout" yaml:"timeout"`
	RateLimit   RateLimitConfig   `json:"rate_limit" yaml:"rate_limit"`
	Retry       RetryConfig       `json:"retry" yaml:"retry"`
	UserAgent   string            `json:"user_agent" yaml:"user_agent"`
	HTTPHeaders map[string]string `json:"http_headers" yaml:"http_headers"`
}

// NewProviderFactory creates a new provider factory.
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{
		configs:      make(map[string]*ProviderConfig),
		constructors: make(map[string]ProviderConstructor),
		defaultSettings: &DefaultSettings{
			Timeout: 30, // 30 seconds default
			RateLimit: RateLimitConfig{
				Enabled:     true,
				RequestsPer: 100,
				Duration:    60, // per minute
				Burst:       10,
			},
			Retry: RetryConfig{
				Enabled:       true,
				MaxAttempts:   3,
				InitialDelay:  100,
				MaxDelay:      5000,
				BackoffFactor: 2.0,
				RetryableErrors: []string{
					"network_error",
					"timeout_error",
					"rate_limit_exceeded",
					"service_unavailable",
					"internal_error",
				},
			},
			UserAgent: "gzh-manager-go/1.0.0",
			HTTPHeaders: map[string]string{
				"Accept": "application/json",
			},
		},
	}
}

// RegisterProvider registers a provider constructor with the factory.
func (f *ProviderFactory) RegisterProvider(providerType string, constructor ProviderConstructor) error {
	if providerType == "" {
		return fmt.Errorf("provider type cannot be empty")
	}
	if constructor == nil {
		return fmt.Errorf("constructor cannot be nil")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.constructors[providerType] = constructor
	return nil
}

// RegisterConfig registers a provider configuration.
func (f *ProviderFactory) RegisterConfig(name string, config *ProviderConfig) error {
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Apply default settings
	f.applyDefaults(config)

	f.mu.Lock()
	defer f.mu.Unlock()

	f.configs[name] = config
	return nil
}

// LoadConfigs loads multiple provider configurations.
func (f *ProviderFactory) LoadConfigs(configs map[string]*ProviderConfig) error {
	for name, config := range configs {
		if err := f.RegisterConfig(name, config); err != nil {
			return fmt.Errorf("failed to register config for %s: %w", name, err)
		}
	}
	return nil
}

// CreateProvider creates a provider instance by name.
func (f *ProviderFactory) CreateProvider(name string) (GitProvider, error) {
	f.mu.RLock()
	config, exists := f.configs[name]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("provider '%s' not configured", name)
	}

	if !config.Enabled {
		return nil, fmt.Errorf("provider '%s' is disabled", name)
	}

	f.mu.RLock()
	constructor, exists := f.constructors[config.Type]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no constructor registered for provider type '%s'", config.Type)
	}

	provider, err := constructor(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider '%s': %w", name, err)
	}

	return provider, nil
}

// CreateProviderByType creates a provider instance by type with temporary config.
func (f *ProviderFactory) CreateProviderByType(providerType string, config *ProviderConfig) (GitProvider, error) {
	f.mu.RLock()
	constructor, exists := f.constructors[providerType]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no constructor registered for provider type '%s'", providerType)
	}

	// Apply defaults to temporary config
	f.applyDefaults(config)

	provider, err := constructor(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider of type '%s': %w", providerType, err)
	}

	return provider, nil
}

// GetProvider creates or retrieves a cached provider instance.
func (f *ProviderFactory) GetProvider(name string) (GitProvider, error) {
	return f.CreateProvider(name)
}

// ListProviders returns a list of configured provider names.
func (f *ProviderFactory) ListProviders() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	names := make([]string, 0, len(f.configs))
	for name, config := range f.configs {
		if config.Enabled {
			names = append(names, name)
		}
	}
	return names
}

// ListAllProviders returns a list of all provider names (including disabled).
func (f *ProviderFactory) ListAllProviders() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	names := make([]string, 0, len(f.configs))
	for name := range f.configs {
		names = append(names, name)
	}
	return names
}

// GetProviderTypes returns a list of registered provider types.
func (f *ProviderFactory) GetProviderTypes() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	types := make([]string, 0, len(f.constructors))
	for providerType := range f.constructors {
		types = append(types, providerType)
	}
	return types
}

// GetConfig returns the configuration for a provider.
func (f *ProviderFactory) GetConfig(name string) (*ProviderConfig, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	config, exists := f.configs[name]
	if !exists {
		return nil, fmt.Errorf("provider '%s' not configured", name)
	}

	// Return a copy to prevent modification
	configCopy := *config
	return &configCopy, nil
}

// UpdateConfig updates the configuration for a provider.
func (f *ProviderFactory) UpdateConfig(name string, config *ProviderConfig) error {
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	f.applyDefaults(config)

	f.mu.Lock()
	defer f.mu.Unlock()

	f.configs[name] = config
	return nil
}

// RemoveProvider removes a provider configuration.
func (f *ProviderFactory) RemoveProvider(name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.configs[name]; !exists {
		return fmt.Errorf("provider '%s' not found", name)
	}

	delete(f.configs, name)
	return nil
}

// EnableProvider enables a provider.
func (f *ProviderFactory) EnableProvider(name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	config, exists := f.configs[name]
	if !exists {
		return fmt.Errorf("provider '%s' not found", name)
	}

	config.Enabled = true
	return nil
}

// DisableProvider disables a provider.
func (f *ProviderFactory) DisableProvider(name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	config, exists := f.configs[name]
	if !exists {
		return fmt.Errorf("provider '%s' not found", name)
	}

	config.Enabled = false
	return nil
}

// ValidateConfig validates a provider configuration.
func (f *ProviderFactory) ValidateConfig(config *ProviderConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.Type == "" {
		return fmt.Errorf("provider type is required")
	}

	if config.Name == "" {
		return fmt.Errorf("provider name is required")
	}

	// Check if constructor exists for this type
	f.mu.RLock()
	_, exists := f.constructors[config.Type]
	f.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no constructor registered for provider type '%s'", config.Type)
	}

	// Validate retry configuration
	if config.Retry.Enabled {
		if config.Retry.MaxAttempts < 1 {
			return fmt.Errorf("retry max_attempts must be at least 1")
		}
		if config.Retry.BackoffFactor <= 0 {
			return fmt.Errorf("retry backoff_factor must be positive")
		}
	}

	// Validate rate limit configuration
	if config.RateLimit.Enabled {
		if config.RateLimit.RequestsPer < 1 {
			return fmt.Errorf("rate_limit requests_per must be at least 1")
		}
		if config.RateLimit.Duration < 1 {
			return fmt.Errorf("rate_limit duration must be at least 1")
		}
	}

	return nil
}

// TestProvider tests a provider configuration by creating and validating it.
func (f *ProviderFactory) TestProvider(name string) error {
	provider, err := f.CreateProvider(name)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Test authentication if the provider supports it
	ctx := context.Background()
	if _, err := provider.ValidateToken(ctx); err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	return nil
}

// GetDefaultSettings returns a copy of the default settings.
func (f *ProviderFactory) GetDefaultSettings() *DefaultSettings {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Return a copy
	defaults := *f.defaultSettings
	return &defaults
}

// UpdateDefaultSettings updates the default settings.
func (f *ProviderFactory) UpdateDefaultSettings(settings *DefaultSettings) {
	if settings == nil {
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.defaultSettings = settings
}

// applyDefaults applies default settings to a provider configuration.
func (f *ProviderFactory) applyDefaults(config *ProviderConfig) {
	if config.Timeout == 0 {
		config.Timeout = f.defaultSettings.Timeout
	}

	// Apply rate limit defaults
	if !config.RateLimit.Enabled && f.defaultSettings.RateLimit.Enabled {
		config.RateLimit = f.defaultSettings.RateLimit
	}

	// Apply retry defaults
	if !config.Retry.Enabled && f.defaultSettings.Retry.Enabled {
		config.Retry = f.defaultSettings.Retry
	}

	// Initialize features map if nil
	if config.Features == nil {
		config.Features = make(map[string]bool)
	}

	// Initialize extra map if nil
	if config.Extra == nil {
		config.Extra = make(map[string]interface{})
	}
}
