// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package validation

import (
	"fmt"
	"strconv"
)

// SyncCloneOptions represents the options that need validation for synclone operations.
type SyncCloneOptions struct {
	TargetPath     string
	OrgName        string
	Strategy       string
	ConfigFile     string
	Parallel       int
	MaxRetries     int
	Token          string
	MemoryLimit    string
	ProgressMode   string
	RedisAddr      string
	IncludePattern string
	ExcludePattern string
	IncludeTopics  []string
	ExcludeTopics  []string
	LanguageFilter string
	MinStars       int
	MaxStars       int
	UpdatedAfter   string
	UpdatedBefore  string
}

// SyncCloneValidator provides validation specifically for synclone operations.
type SyncCloneValidator struct {
	validator *Validator
}

// NewSyncCloneValidator creates a new synclone validator.
func NewSyncCloneValidator() *SyncCloneValidator {
	return &SyncCloneValidator{
		validator: New(),
	}
}

// ValidateOptions performs comprehensive validation of synclone options.
func (v *SyncCloneValidator) ValidateOptions(opts *SyncCloneOptions) error {
	// Required field validation
	if err := v.validateRequiredFields(opts); err != nil {
		return fmt.Errorf("required field validation failed: %w", err)
	}

	// Path validation
	if err := v.validatePaths(opts); err != nil {
		return fmt.Errorf("path validation failed: %w", err)
	}

	// Organization and strategy validation
	if err := v.validateOrganizationAndStrategy(opts); err != nil {
		return fmt.Errorf("organization/strategy validation failed: %w", err)
	}

	// Numeric parameters validation
	if err := v.validateNumericParams(opts); err != nil {
		return fmt.Errorf("numeric parameter validation failed: %w", err)
	}

	// Token validation
	if err := v.validateToken(opts); err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	// Performance settings validation
	if err := v.validatePerformanceSettings(opts); err != nil {
		return fmt.Errorf("performance settings validation failed: %w", err)
	}

	// Filter validation
	if err := v.validateFilters(opts); err != nil {
		return fmt.Errorf("filter validation failed: %w", err)
	}

	// Date range validation
	if err := v.validateDateRanges(opts); err != nil {
		return fmt.Errorf("date range validation failed: %w", err)
	}

	return nil
}

func (v *SyncCloneValidator) validateRequiredFields(opts *SyncCloneOptions) error {
	if opts.TargetPath == "" {
		return fmt.Errorf("target path is required")
	}

	if opts.OrgName == "" {
		return fmt.Errorf("organization name is required")
	}

	return nil
}

func (v *SyncCloneValidator) validatePaths(opts *SyncCloneOptions) error {
	// Validate target path
	if err := v.validator.ValidatePath(opts.TargetPath); err != nil {
		return fmt.Errorf("invalid target path: %w", err)
	}

	// Validate config file path if provided
	if opts.ConfigFile != "" {
		if err := v.validator.ValidatePath(opts.ConfigFile); err != nil {
			return fmt.Errorf("invalid config file path: %w", err)
		}
	}

	return nil
}

func (v *SyncCloneValidator) validateOrganizationAndStrategy(opts *SyncCloneOptions) error {
	// Validate organization name
	if err := v.validator.ValidateOrganizationName(opts.OrgName); err != nil {
		return fmt.Errorf("invalid organization name: %w", err)
	}

	// Validate strategy
	if opts.Strategy == "" {
		opts.Strategy = "reset" // Set default
	}
	if err := v.validator.ValidateStrategy(opts.Strategy); err != nil {
		return fmt.Errorf("invalid strategy: %w", err)
	}

	return nil
}

func (v *SyncCloneValidator) validateNumericParams(opts *SyncCloneOptions) error {
	// Validate parallelism
	if opts.Parallel <= 0 {
		opts.Parallel = 5 // Set default
	}
	if err := v.validator.ValidateParallelism(opts.Parallel); err != nil {
		return fmt.Errorf("invalid parallelism: %w", err)
	}

	// Validate max retries
	if opts.MaxRetries < 0 {
		opts.MaxRetries = 3 // Set default
	}
	if err := v.validator.ValidateRetries(opts.MaxRetries); err != nil {
		return fmt.Errorf("invalid max retries: %w", err)
	}

	// Validate star ranges
	if opts.MinStars < 0 {
		return fmt.Errorf("minimum stars cannot be negative")
	}

	if opts.MaxStars < 0 {
		return fmt.Errorf("maximum stars cannot be negative")
	}

	if opts.MinStars > 0 && opts.MaxStars > 0 && opts.MinStars > opts.MaxStars {
		return fmt.Errorf("minimum stars (%d) cannot be greater than maximum stars (%d)", opts.MinStars, opts.MaxStars)
	}

	return nil
}

func (v *SyncCloneValidator) validateToken(opts *SyncCloneOptions) error {
	if err := v.validator.ValidateToken(opts.Token); err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	return nil
}

func (v *SyncCloneValidator) validatePerformanceSettings(opts *SyncCloneOptions) error {
	// Validate memory limit
	if err := v.validator.ValidateMemoryLimit(opts.MemoryLimit); err != nil {
		return fmt.Errorf("invalid memory limit: %w", err)
	}

	// Validate progress mode
	if err := v.validator.ValidateProgressMode(opts.ProgressMode); err != nil {
		return fmt.Errorf("invalid progress mode: %w", err)
	}

	// Validate Redis address
	if err := v.validator.ValidateRedisAddress(opts.RedisAddr); err != nil {
		return fmt.Errorf("invalid Redis address: %w", err)
	}

	return nil
}

func (v *SyncCloneValidator) validateFilters(opts *SyncCloneOptions) error {
	// Validate include pattern
	if err := v.validator.ValidatePattern(opts.IncludePattern); err != nil {
		return fmt.Errorf("invalid include pattern: %w", err)
	}

	// Validate exclude pattern
	if err := v.validator.ValidatePattern(opts.ExcludePattern); err != nil {
		return fmt.Errorf("invalid exclude pattern: %w", err)
	}

	// Validate topics
	if err := v.validator.ValidateTopics(opts.IncludeTopics); err != nil {
		return fmt.Errorf("invalid include topics: %w", err)
	}

	if err := v.validator.ValidateTopics(opts.ExcludeTopics); err != nil {
		return fmt.Errorf("invalid exclude topics: %w", err)
	}

	// Validate language filter
	if opts.LanguageFilter != "" {
		if len(opts.LanguageFilter) > 50 {
			return fmt.Errorf("language filter too long: maximum 50 characters")
		}

		// Language should be alphanumeric with basic special characters
		if !v.validator.IsSecureInput(opts.LanguageFilter) {
			return fmt.Errorf("language filter contains potentially dangerous characters")
		}
	}

	return nil
}

func (v *SyncCloneValidator) validateDateRanges(opts *SyncCloneOptions) error {
	// Validate updated after date
	if err := v.validator.ValidateDateRange(opts.UpdatedAfter); err != nil {
		return fmt.Errorf("invalid updated after date: %w", err)
	}

	// Validate updated before date
	if err := v.validator.ValidateDateRange(opts.UpdatedBefore); err != nil {
		return fmt.Errorf("invalid updated before date: %w", err)
	}

	return nil
}

// SanitizeOptions sanitizes all string inputs in synclone options.
func (v *SyncCloneValidator) SanitizeOptions(opts *SyncCloneOptions) *SyncCloneOptions {
	sanitized := &SyncCloneOptions{
		TargetPath:     v.validator.SanitizeString(opts.TargetPath),
		OrgName:        v.validator.SanitizeString(opts.OrgName),
		Strategy:       v.validator.SanitizeString(opts.Strategy),
		ConfigFile:     v.validator.SanitizeString(opts.ConfigFile),
		Parallel:       opts.Parallel,
		MaxRetries:     opts.MaxRetries,
		Token:          opts.Token, // Don't sanitize tokens as they have specific formats
		MemoryLimit:    v.validator.SanitizeString(opts.MemoryLimit),
		ProgressMode:   v.validator.SanitizeString(opts.ProgressMode),
		RedisAddr:      v.validator.SanitizeString(opts.RedisAddr),
		IncludePattern: v.validator.SanitizeString(opts.IncludePattern),
		ExcludePattern: v.validator.SanitizeString(opts.ExcludePattern),
		LanguageFilter: v.validator.SanitizeString(opts.LanguageFilter),
		MinStars:       opts.MinStars,
		MaxStars:       opts.MaxStars,
		UpdatedAfter:   v.validator.SanitizeString(opts.UpdatedAfter),
		UpdatedBefore:  v.validator.SanitizeString(opts.UpdatedBefore),
	}

	// Sanitize topic arrays
	for _, topic := range opts.IncludeTopics {
		sanitized.IncludeTopics = append(sanitized.IncludeTopics, v.validator.SanitizeString(topic))
	}

	for _, topic := range opts.ExcludeTopics {
		sanitized.ExcludeTopics = append(sanitized.ExcludeTopics, v.validator.SanitizeString(topic))
	}

	return sanitized
}

// ValidateFlagValue validates individual flag values as they're parsed.
func (v *SyncCloneValidator) ValidateFlagValue(flagName, value string) error {
	switch flagName {
	case "org", "organization":
		return v.validator.ValidateOrganizationName(value)
	case "strategy":
		return v.validator.ValidateStrategy(value)
	case "parallel":
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("parallel must be a number")
		}
		return v.validator.ValidateParallelism(intValue)
	case "max-retries":
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("max-retries must be a number")
		}
		return v.validator.ValidateRetries(intValue)
	case "token":
		return v.validator.ValidateToken(value)
	case "memory-limit":
		return v.validator.ValidateMemoryLimit(value)
	case "progress-mode":
		return v.validator.ValidateProgressMode(value)
	case "redis-addr":
		return v.validator.ValidateRedisAddress(value)
	case "include-pattern", "exclude-pattern":
		return v.validator.ValidatePattern(value)
	case "updated-after", "updated-before":
		return v.validator.ValidateDateRange(value)
	case "target-path", "config":
		return v.validator.ValidatePath(value)
	default:
		// For unknown flags, just check for basic security
		if !v.validator.IsSecureInput(value) {
			return fmt.Errorf("flag value contains potentially dangerous characters")
		}
	}

	return nil
}
