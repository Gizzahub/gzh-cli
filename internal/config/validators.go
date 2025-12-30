// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/gizzahub/gzh-cli/internal/constants"
	"github.com/gizzahub/gzh-cli/internal/errors"
)

// BulkCloneConfigValidator validates bulk clone configuration.
type BulkCloneConfigValidator struct{}

// Validate validates bulk clone configuration.
func (v *BulkCloneConfigValidator) Validate(config interface{}) error {
	bulkConfig, ok := config.(*BulkCloneConfig)
	if !ok {
		return errors.NewStandardError(errors.ErrorCodeValidationFailed,
			"invalid configuration type for bulk clone validator", errors.SeverityHigh)
	}

	if bulkConfig.Concurrency <= 0 || bulkConfig.Concurrency > constants.MaxParallelism {
		return errors.NewValidationError(
			fmt.Sprintf("concurrency must be between 1 and %d", constants.MaxParallelism),
			"concurrency")
	}

	if bulkConfig.Timeout <= 0 || bulkConfig.Timeout > constants.VeryLongTimeout {
		return errors.NewValidationError(
			"timeout must be positive and reasonable", "timeout")
	}

	if bulkConfig.RetryAttempts < 0 || bulkConfig.RetryAttempts > constants.MaxRetryAttempts {
		return errors.NewValidationError(
			fmt.Sprintf("retry attempts must be between 0 and %d", constants.MaxRetryAttempts),
			"retry_attempts")
	}

	if bulkConfig.BufferSize <= 0 {
		return errors.NewValidationError(
			"buffer size must be positive", "buffer_size")
	}

	if bulkConfig.MaxRepositories <= 0 {
		return errors.NewValidationError(
			"max repositories must be positive", "max_repositories")
	}

	return nil
}

// HTTPClientConfigValidator validates HTTP client configuration.
type HTTPClientConfigValidator struct{}

// Validate validates HTTP client configuration.
func (v *HTTPClientConfigValidator) Validate(config interface{}) error {
	httpConfig, ok := config.(*HTTPClientConfig)
	if !ok {
		return errors.NewStandardError(errors.ErrorCodeValidationFailed,
			"invalid configuration type for HTTP client validator", errors.SeverityHigh)
	}

	if httpConfig.Timeout <= 0 || httpConfig.Timeout > constants.ExtraLongTimeout {
		return errors.NewValidationError(
			"timeout must be positive and reasonable", "timeout")
	}

	if httpConfig.MaxIdleConns < 0 {
		return errors.NewValidationError(
			"max idle connections must be non-negative", "max_idle_conns")
	}

	if httpConfig.IdleConnTimeout < 0 {
		return errors.NewValidationError(
			"idle connection timeout must be non-negative", "idle_conn_timeout")
	}

	validTLSVersions := map[string]bool{
		"1.0": true, "1.1": true, "1.2": true, "1.3": true,
	}

	if httpConfig.MinTLSVersion != "" && !validTLSVersions[httpConfig.MinTLSVersion] {
		return errors.NewValidationError(
			"min TLS version must be one of: 1.0, 1.1, 1.2, 1.3", "min_tls_version")
	}

	if httpConfig.UserAgent == "" {
		return errors.NewValidationError(
			"user agent cannot be empty", "user_agent")
	}

	return nil
}

// AuthConfigValidator validates authentication configuration.
type AuthConfigValidator struct{}

// Validate validates authentication configuration.
func (v *AuthConfigValidator) Validate(config interface{}) error {
	authConfig, ok := config.(*AuthConfig)
	if !ok {
		return errors.NewStandardError(errors.ErrorCodeValidationFailed,
			"invalid configuration type for auth validator", errors.SeverityHigh)
	}

	if authConfig.TokenMinLength < 4 || authConfig.TokenMinLength > constants.MaxTokenLength {
		return errors.NewValidationError(
			fmt.Sprintf("token min length must be between 4 and %d", constants.MaxTokenLength),
			"token_min_length")
	}

	if authConfig.ValidationTimeout <= 0 || authConfig.ValidationTimeout > constants.LongHTTPTimeout {
		return errors.NewValidationError(
			"validation timeout must be positive and reasonable", "validation_timeout")
	}

	if authConfig.CacheDuration < 0 {
		return errors.NewValidationError(
			"cache duration must be non-negative", "cache_duration")
	}

	if authConfig.MaxRetryAttempts < 0 || authConfig.MaxRetryAttempts > constants.MaxRetryAttempts {
		return errors.NewValidationError(
			fmt.Sprintf("max retry attempts must be between 0 and %d", constants.MaxRetryAttempts),
			"max_retry_attempts")
	}

	return nil
}

// GitHubConfigValidator validates GitHub-specific configuration.
type GitHubConfigValidator struct{}

// Validate validates GitHub configuration.
func (v *GitHubConfigValidator) Validate(config interface{}) error {
	githubConfig, ok := config.(*GitHubConfig)
	if !ok {
		return errors.NewStandardError(errors.ErrorCodeValidationFailed,
			"invalid configuration type for GitHub validator", errors.SeverityHigh)
	}

	// Validate token format (GitHub tokens start with ghp_ or github_pat_)
	if githubConfig.Token != "" {
		if !strings.HasPrefix(githubConfig.Token, "ghp_") &&
			!strings.HasPrefix(githubConfig.Token, "github_pat_") {
			return errors.NewValidationError(
				"GitHub token must start with 'ghp_' or 'github_pat_'", "token")
		}

		if len(githubConfig.Token) < constants.MinTokenLength {
			return errors.NewValidationError(
				fmt.Sprintf("token must be at least %d characters", constants.MinTokenLength), "token")
		}
	}

	// Validate base URL
	if githubConfig.BaseURL != "" {
		if _, err := url.Parse(githubConfig.BaseURL); err != nil {
			return errors.NewValidationError(
				"invalid base URL format", "base_url")
		}

		if !strings.HasPrefix(githubConfig.BaseURL, "https://") {
			return errors.NewValidationError(
				"base URL must use HTTPS", "base_url")
		}
	}

	// Validate timeouts
	if githubConfig.Timeout <= 0 || githubConfig.Timeout > constants.ExtraLongTimeout {
		return errors.NewValidationError(
			"timeout must be positive and reasonable", "timeout")
	}

	// Validate retry settings
	if githubConfig.RetryAttempts < 0 || githubConfig.RetryAttempts > constants.MaxRetryAttempts {
		return errors.NewValidationError(
			fmt.Sprintf("retry attempts must be between 0 and %d", constants.MaxRetryAttempts),
			"retry_attempts")
	}

	if githubConfig.RetryDelay < 0 {
		return errors.NewValidationError(
			"retry delay must be non-negative", "retry_delay")
	}

	// Validate rate limit
	if githubConfig.RateLimit < 0 {
		return errors.NewValidationError(
			"rate limit must be non-negative", "rate_limit")
	}

	return nil
}

// SecurityConfigValidator validates security configuration.
type SecurityConfigValidator struct{}

// Validate validates security configuration.
func (v *SecurityConfigValidator) Validate(config interface{}) error {
	secConfig, ok := config.(*SecurityConfig)
	if !ok {
		return errors.NewStandardError(errors.ErrorCodeValidationFailed,
			"invalid configuration type for security validator", errors.SeverityHigh)
	}

	if secConfig.TokenValidationTimeout <= 0 || secConfig.TokenValidationTimeout > constants.LongHTTPTimeout {
		return errors.NewValidationError(
			"token validation timeout must be positive and reasonable", "token_validation_timeout")
	}

	if secConfig.MaxRequestSize <= 0 {
		return errors.NewValidationError(
			"max request size must be positive", "max_request_size")
	}

	// Validate allowed hosts
	for _, host := range secConfig.AllowedHosts {
		if host == "" {
			return errors.NewValidationError(
				"allowed hosts cannot contain empty values", "allowed_hosts")
		}

		// Basic hostname validation
		hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
		if !hostnameRegex.MatchString(host) && host != "localhost" && host != "*" {
			return errors.NewValidationError(
				fmt.Sprintf("invalid hostname format: %s", host), "allowed_hosts")
		}
	}

	return nil
}

// LoggingConfigValidator validates logging configuration.
type LoggingConfigValidator struct{}

// Validate validates logging configuration.
func (v *LoggingConfigValidator) Validate(config interface{}) error {
	logConfig, ok := config.(*LoggingConfig)
	if !ok {
		return errors.NewStandardError(errors.ErrorCodeValidationFailed,
			"invalid configuration type for logging validator", errors.SeverityHigh)
	}

	validLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true, "fatal": true,
	}

	if !validLevels[strings.ToLower(logConfig.Level)] {
		return errors.NewValidationError(
			"log level must be one of: debug, info, warn, error, fatal", "level")
	}

	validFormats := map[string]bool{
		"json": true, "text": true, "console": true,
	}

	if !validFormats[strings.ToLower(logConfig.Format)] {
		return errors.NewValidationError(
			"log format must be one of: json, text, console", "format")
	}

	validOutputs := map[string]bool{
		"stdout": true, "stderr": true, "file": true,
	}

	if !validOutputs[strings.ToLower(logConfig.Output)] && !strings.HasPrefix(logConfig.Output, "/") {
		return errors.NewValidationError(
			"log output must be stdout, stderr, file, or a file path", "output")
	}

	if logConfig.MaxSize <= 0 {
		return errors.NewValidationError(
			"max size must be positive", "max_size")
	}

	if logConfig.MaxBackups < 0 {
		return errors.NewValidationError(
			"max backups must be non-negative", "max_backups")
	}

	if logConfig.MaxAge < 0 {
		return errors.NewValidationError(
			"max age must be non-negative", "max_age")
	}

	return nil
}

// CompositeValidator combines multiple validators.
type CompositeValidator struct {
	validators []Validator
}

// NewCompositeValidator creates a new composite validator.
func NewCompositeValidator(validators ...Validator) *CompositeValidator {
	return &CompositeValidator{validators: validators}
}

// Validate runs all validators.
func (cv *CompositeValidator) Validate(config interface{}) error {
	for _, validator := range cv.validators {
		if err := validator.Validate(config); err != nil {
			return err
		}
	}
	return nil
}

// RegisterDefaultValidators registers default validators for all config types.
func RegisterDefaultValidators(manager *ConfigManager) {
	manager.RegisterValidator("bulk-clone", &BulkCloneConfigValidator{})
	manager.RegisterValidator("http-client", &HTTPClientConfigValidator{})
	manager.RegisterValidator("auth", &AuthConfigValidator{})
	manager.RegisterValidator("github", &GitHubConfigValidator{})
	manager.RegisterValidator("security", &SecurityConfigValidator{})
	manager.RegisterValidator("logging", &LoggingConfigValidator{})
}
