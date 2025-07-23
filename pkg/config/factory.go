// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

import (
	"context"
	"fmt"

	"github.com/gizzahub/gzh-manager-go/internal/env"
)

// ProviderFactory defines the interface for creating provider instances.
type ProviderFactory interface {
	// CreateCloner creates a provider cloner for the specified provider
	CreateCloner(ctx context.Context, providerName, token string) (ProviderCloner, error)

	// CreateClonerWithEnv creates a provider cloner with a specific environment
	CreateClonerWithEnv(ctx context.Context, providerName, token string, environment env.Environment) (ProviderCloner, error)

	// GetSupportedProviders returns a list of supported provider names
	GetSupportedProviders() []string

	// IsProviderSupported checks if a provider is supported
	IsProviderSupported(providerName string) bool
}

// providerFactoryImpl implements the ProviderFactory interface.
type providerFactoryImpl struct {
	environment env.Environment
	logger      Logger
}

// NewProviderFactory creates a new provider factory with dependencies.
func NewProviderFactory(environment env.Environment, logger Logger) ProviderFactory {
	if environment == nil {
		environment = env.NewOSEnvironment()
	}

	return &providerFactoryImpl{
		environment: environment,
		logger:      logger,
	}
}

// CreateCloner creates a provider cloner for the specified provider.
func (f *providerFactoryImpl) CreateCloner(ctx context.Context, providerName, token string) (ProviderCloner, error) {
	return f.CreateClonerWithEnv(ctx, providerName, token, f.environment)
}

// CreateClonerWithEnv creates a provider cloner with a specific environment.
func (f *providerFactoryImpl) CreateClonerWithEnv(_ context.Context, providerName, token string, environment env.Environment) (ProviderCloner, error) {
	f.logger.Debug("Creating provider cloner", "provider", providerName)

	if !f.IsProviderSupported(providerName) {
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}

	switch providerName {
	case ProviderGitHub:
		return NewGitHubClonerWithEnv(token, environment), nil
	case ProviderGitLab:
		return NewGitLabClonerWithEnv(token, environment), nil
	case ProviderGitea:
		return NewGiteaClonerWithEnv(token, environment), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}
}

// GetSupportedProviders returns a list of supported provider names.
func (f *providerFactoryImpl) GetSupportedProviders() []string {
	return []string{ProviderGitHub, ProviderGitLab, ProviderGitea}
}

// IsProviderSupported checks if a provider is supported.
func (f *providerFactoryImpl) IsProviderSupported(providerName string) bool {
	supportedProviders := f.GetSupportedProviders()
	for _, provider := range supportedProviders {
		if provider == providerName {
			return true
		}
	}

	return false
}

// ProviderFactoryConfig holds configuration for the provider factory.
type ProviderFactoryConfig struct {
	// DefaultEnvironment is the default environment to use when none is specified
	DefaultEnvironment env.Environment
	// EnableLogging enables factory operation logging
	EnableLogging bool
}

// DefaultProviderFactoryConfig returns default factory configuration.
func DefaultProviderFactoryConfig() *ProviderFactoryConfig {
	return &ProviderFactoryConfig{
		DefaultEnvironment: env.NewOSEnvironment(),
		EnableLogging:      true,
	}
}

// NewProviderFactoryWithConfig creates a new provider factory with configuration.
func NewProviderFactoryWithConfig(config *ProviderFactoryConfig, logger Logger) ProviderFactory {
	if config == nil {
		config = DefaultProviderFactoryConfig()
	}

	return NewProviderFactory(config.DefaultEnvironment, logger)
}
