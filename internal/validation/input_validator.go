// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package validation provides comprehensive input validation and sanitization
// for security purposes throughout the application.
package validation

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Gizzahub/gzh-manager-go/internal/constants"
)

// Validator provides comprehensive input validation for security.
type Validator struct {
	// URL validation patterns
	urlPattern    *regexp.Regexp
	gitURLPattern *regexp.Regexp

	// Path validation patterns
	pathPattern *regexp.Regexp

	// Token validation patterns
	tokenPattern *regexp.Regexp

	// General security patterns
	injectionPattern *regexp.Regexp

	// Organization/repository patterns
	namePattern *regexp.Regexp
}

// New creates a new input validator with security patterns.
func New() *Validator {
	return &Validator{
		urlPattern:       regexp.MustCompile(`^https?://[a-zA-Z0-9.-]+(/.*)?$`),
		gitURLPattern:    regexp.MustCompile(`^(https?://[a-zA-Z0-9.-]+/[a-zA-Z0-9._/-]+\.git|git@[a-zA-Z0-9.-]+:[a-zA-Z0-9._/-]+\.git)$`),
		pathPattern:      regexp.MustCompile(`^[a-zA-Z0-9._/\-~\s]+$`),
		tokenPattern:     regexp.MustCompile(`^[a-zA-Z0-9_-]+$`),
		injectionPattern: regexp.MustCompile(`[;&|<>$()` + "`" + `]`),
		namePattern:      regexp.MustCompile(`^[a-zA-Z0-9._-]+$`),
	}
}

// ValidateURL validates HTTP/HTTPS URLs for security.
func (v *Validator) ValidateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// Check for injection patterns
	if v.injectionPattern.MatchString(rawURL) {
		return fmt.Errorf("URL contains potentially dangerous characters")
	}

	// Parse URL
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Validate scheme
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("only HTTP and HTTPS URLs are allowed")
	}

	// Validate host
	if parsed.Host == "" {
		return fmt.Errorf("URL must have a valid host")
	}

	return nil
}

// ValidateGitURL validates Git repository URLs.
func (v *Validator) ValidateGitURL(gitURL string) error {
	if gitURL == "" {
		return fmt.Errorf("git URL cannot be empty")
	}

	// Check for injection patterns
	if v.injectionPattern.MatchString(gitURL) {
		return fmt.Errorf("git URL contains potentially dangerous characters")
	}

	// Validate against Git URL pattern
	if !v.gitURLPattern.MatchString(gitURL) {
		return fmt.Errorf("invalid Git URL format")
	}

	return nil
}

// ValidatePath validates file system paths for security.
func (v *Validator) ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Check for injection patterns in the raw path
	if v.injectionPattern.MatchString(path) {
		return fmt.Errorf("path contains potentially dangerous characters")
	}

	// Check for path traversal
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal detected")
	}

	// Clean and validate the path
	cleanPath := filepath.Clean(path)

	// Check for access to system directories
	systemDirs := []string{"/etc", "/usr", "/sys", "/proc", "C:\\Windows", "C:\\Program Files"}
	for _, sysDir := range systemDirs {
		if strings.HasPrefix(cleanPath, sysDir) {
			return fmt.Errorf("access to system directory not allowed: %s", sysDir)
		}
	}

	return nil
}

// ValidateToken validates authentication tokens.
func (v *Validator) ValidateToken(token string) error {
	if token == "" {
		return nil // Empty token is allowed
	}

	// Check minimum length
	if len(token) < constants.MinTokenLength {
		return fmt.Errorf("token too short (minimum %d characters)", constants.MinTokenLength)
	}

	// Check maximum length (prevent DoS)
	if len(token) > constants.MaxTokenLength {
		return fmt.Errorf("token too long (maximum %d characters)", constants.MaxTokenLength)
	}

	// Validate token pattern - allow GitHub and GitLab token formats
	if !v.tokenPattern.MatchString(token) && !strings.HasPrefix(token, "ghp_") && !strings.HasPrefix(token, "glpat-") {
		return fmt.Errorf("token contains invalid characters")
	}

	return nil
}

// ValidateOrganizationName validates organization/user names.
func (v *Validator) ValidateOrganizationName(name string) error {
	if name == "" {
		return fmt.Errorf("organization name cannot be empty")
	}

	// Check length constraints
	if len(name) < 1 || len(name) > constants.MaxOrganizationNameLength {
		return fmt.Errorf("organization name must be 1-%d characters", constants.MaxOrganizationNameLength)
	}

	// Check for injection patterns
	if v.injectionPattern.MatchString(name) {
		return fmt.Errorf("organization name contains potentially dangerous characters")
	}

	// Validate against name pattern
	if !v.namePattern.MatchString(name) {
		return fmt.Errorf("organization name contains invalid characters")
	}

	// Additional restrictions
	if strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") {
		return fmt.Errorf("organization name cannot start or end with a period")
	}

	return nil
}

// ValidateRepositoryName validates repository names.
func (v *Validator) ValidateRepositoryName(name string) error {
	if name == "" {
		return fmt.Errorf("repository name cannot be empty")
	}

	// Check length constraints
	if len(name) < 1 || len(name) > constants.MaxRepositoryNameLength {
		return fmt.Errorf("repository name must be 1-%d characters", constants.MaxRepositoryNameLength)
	}

	// Check for injection patterns
	if v.injectionPattern.MatchString(name) {
		return fmt.Errorf("repository name contains potentially dangerous characters")
	}

	// Validate against name pattern
	if !v.namePattern.MatchString(name) {
		return fmt.Errorf("repository name contains invalid characters")
	}

	return nil
}

// ValidateStrategy validates clone/update strategies.
func (v *Validator) ValidateStrategy(strategy string) error {
	validStrategies := map[string]bool{
		"reset": true,
		"pull":  true,
		"fetch": true,
	}

	if !validStrategies[strategy] {
		return fmt.Errorf("invalid strategy: %s (valid: reset, pull, fetch)", strategy)
	}

	return nil
}

// ValidateInteger validates integer values within bounds.
func (v *Validator) ValidateInteger(value string, minValue, maxValue int, fieldName string) error {
	if value == "" {
		return nil // Empty is often allowed for optional fields
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid %s: must be a number", fieldName)
	}

	if intValue < minValue || intValue > maxValue {
		return fmt.Errorf("invalid %s: must be between %d and %d", fieldName, minValue, maxValue)
	}

	return nil
}

// ValidateParallelism validates parallelism settings.
func (v *Validator) ValidateParallelism(parallel int) error {
	if parallel < 1 || parallel > constants.MaxParallelism {
		return fmt.Errorf("invalid parallelism: must be between 1 and %d", constants.MaxParallelism)
	}

	return nil
}

// ValidateRetries validates retry count settings.
func (v *Validator) ValidateRetries(retries int) error {
	if retries < 0 || retries > constants.MaxRetryAttempts {
		return fmt.Errorf("invalid retry count: must be between 0 and %d", constants.MaxRetryAttempts)
	}

	return nil
}

// ValidateMemoryLimit validates memory limit strings.
func (v *Validator) ValidateMemoryLimit(limit string) error {
	if limit == "" {
		return nil // Empty is allowed
	}

	// Validate format like "100MB", "2GB", etc.
	pattern := regexp.MustCompile(`^\d+[KMGT]?B?$`)
	if !pattern.MatchString(strings.ToUpper(limit)) {
		return fmt.Errorf("invalid memory limit format: use format like '100MB' or '2GB'")
	}

	return nil
}

// ValidateProgressMode validates progress display modes.
func (v *Validator) ValidateProgressMode(mode string) error {
	if mode == "" {
		return nil
	}

	validModes := map[string]bool{
		"bar":     true,
		"dots":    true,
		"spinner": true,
		"quiet":   true,
	}

	if !validModes[mode] {
		return fmt.Errorf("invalid progress mode: %s (valid: bar, dots, spinner, quiet)", mode)
	}

	return nil
}

// ValidateRedisAddress validates Redis connection addresses.
func (v *Validator) ValidateRedisAddress(addr string) error {
	if addr == "" {
		return nil
	}

	// Basic format validation for host:port
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid Redis address format: use host:port")
	}

	host, portStr := parts[0], parts[1]

	// Validate host
	if host == "" {
		return fmt.Errorf("redis host cannot be empty")
	}

	// Validate port
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid Redis port: must be a number")
	}

	if port < constants.MinPortNumber || port > constants.MaxPortNumber {
		return fmt.Errorf("invalid Redis port: must be between %d and %d", constants.MinPortNumber, constants.MaxPortNumber)
	}

	return nil
}

// ValidateDateRange validates date strings in various formats.
func (v *Validator) ValidateDateRange(dateStr string) error {
	if dateStr == "" {
		return nil
	}

	// Try multiple date formats
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if _, err := time.Parse(format, dateStr); err == nil {
			return nil
		}
	}

	return fmt.Errorf("invalid date format: use YYYY-MM-DD or RFC3339 format")
}

// ValidatePattern validates regex patterns for include/exclude filters.
func (v *Validator) ValidatePattern(pattern string) error {
	if pattern == "" {
		return nil
	}

	// Check if it's a valid regex
	_, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	return nil
}

// ValidateTopics validates topic lists.
func (v *Validator) ValidateTopics(topics []string) error {
	for _, topic := range topics {
		if topic == "" {
			return fmt.Errorf("topic cannot be empty")
		}

		if len(topic) > 50 {
			return fmt.Errorf("topic too long: maximum 50 characters")
		}

		// Topics should be alphanumeric with hyphens
		topicPattern := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
		if !topicPattern.MatchString(topic) {
			return fmt.Errorf("invalid topic format: %s", topic)
		}
	}

	return nil
}

// SanitizeString removes potentially dangerous characters from input.
func (v *Validator) SanitizeString(input string) string {
	// Remove control characters
	sanitized := strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, input)

	// Remove potentially dangerous sequences
	dangerous := []string{
		"../", "..\\",
		"<script>", "</script>",
		"javascript:", "vbscript:",
		"onload=", "onerror=",
	}

	for _, danger := range dangerous {
		sanitized = strings.ReplaceAll(sanitized, danger, "")
	}

	return strings.TrimSpace(sanitized)
}

// IsSecureInput performs a comprehensive security check on input.
func (v *Validator) IsSecureInput(input string) bool {
	// Check for injection patterns
	if v.injectionPattern.MatchString(input) {
		return false
	}

	// Check for common attack patterns
	attackPatterns := []string{
		"<script", "javascript:", "vbscript:",
		"union select", "drop table", "delete from",
		"xp_cmdshell", "/bin/sh", "cmd.exe",
		"$(", "${", "`",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range attackPatterns {
		if strings.Contains(lowerInput, pattern) {
			return false
		}
	}

	return true
}
