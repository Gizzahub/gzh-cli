// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"time"
)

// BaseProvider provides common functionality for all Git providers.
type BaseProvider struct {
	name        string
	baseURL     string
	token       string
	httpTimeout time.Duration
}

// NewBaseProvider creates a new base provider with common configuration.
func NewBaseProvider(name, baseURL, token string) *BaseProvider {
	return &BaseProvider{
		name:        name,
		baseURL:     baseURL,
		token:       token,
		httpTimeout: 30 * time.Second,
	}
}

// GetName returns the provider name.
func (b *BaseProvider) GetName() string {
	return b.name
}

// GetBaseURL returns the provider base URL.
func (b *BaseProvider) GetBaseURL() string {
	return b.baseURL
}

// SetToken sets the authentication token.
func (b *BaseProvider) SetToken(token string) {
	b.token = token
}

// GetToken returns the authentication token.
func (b *BaseProvider) GetToken() string {
	return b.token
}

// SetHTTPTimeout sets the HTTP client timeout.
func (b *BaseProvider) SetHTTPTimeout(timeout time.Duration) {
	b.httpTimeout = timeout
}

// GetHTTPTimeout returns the HTTP client timeout.
func (b *BaseProvider) GetHTTPTimeout() time.Duration {
	return b.httpTimeout
}

// ValidateConfig validates the provider configuration.
func (b *BaseProvider) ValidateConfig() error {
	if b.name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}
	if b.baseURL == "" {
		return fmt.Errorf("provider base URL cannot be empty")
	}
	return nil
}

// HealthCheck performs a basic health check against the provider.
func (b *BaseProvider) HealthCheck(ctx context.Context) error {
	// Default implementation - can be overridden by specific providers
	if err := b.ValidateConfig(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}
	return nil
}

// FormatError formats provider-specific errors with context.
func (b *BaseProvider) FormatError(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s provider %s operation failed: %w", b.name, operation, err)
}

// IsAuthenticated checks if the provider has authentication configured.
func (b *BaseProvider) IsAuthenticated() bool {
	return b.token != ""
}

// DefaultRepositoryMetadata returns default metadata for repositories.
func (b *BaseProvider) DefaultRepositoryMetadata() map[string]interface{} {
	return map[string]interface{}{
		"provider":     b.name,
		"provider_url": b.baseURL,
		"created_at":   time.Now().UTC(),
	}
}
