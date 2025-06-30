package config

import (
	"fmt"
	"regexp"
	"strings"
)

// Validator provides configuration validation functionality
type Validator struct {
	errors   []string
	warnings []string
}

// NewValidator creates a new configuration validator
func NewValidator() *Validator {
	return &Validator{
		errors:   make([]string, 0),
		warnings: make([]string, 0),
	}
}

// ValidateConfig performs comprehensive validation of a configuration
func (v *Validator) ValidateConfig(config *Config) error {
	v.reset()

	// Validate required fields
	v.validateRequiredFields(config)

	// Validate version format
	v.validateVersion(config.Version)

	// Validate default provider
	v.validateDefaultProvider(config.DefaultProvider)

	// Validate providers
	v.validateProviders(config.Providers)

	// Return error if validation failed
	if len(v.errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n%s", strings.Join(v.errors, "\n"))
	}

	return nil
}

// GetWarnings returns validation warnings
func (v *Validator) GetWarnings() []string {
	return v.warnings
}

// reset clears previous validation results
func (v *Validator) reset() {
	v.errors = v.errors[:0]
	v.warnings = v.warnings[:0]
}

// validateRequiredFields checks for required configuration fields
func (v *Validator) validateRequiredFields(config *Config) {
	if config.Version == "" {
		v.addError("missing required field 'version'")
	}

	if len(config.Providers) == 0 {
		v.addWarning("no providers configured")
	}
}

// validateVersion validates the version format
func (v *Validator) validateVersion(version string) {
	if version == "" {
		return // Already handled in required fields
	}

	// Validate semantic version format (major.minor.patch)
	versionRegex := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
	if !versionRegex.MatchString(version) {
		v.addError(fmt.Sprintf("invalid version format '%s', expected semantic version (e.g., '1.0.0')", version))
	}
}

// validateDefaultProvider validates the default provider setting
func (v *Validator) validateDefaultProvider(provider string) {
	if provider == "" {
		return // Optional field
	}

	validProviders := []string{ProviderGitHub, ProviderGitLab, ProviderGitea, ProviderGogs}
	if !contains(validProviders, provider) {
		v.addError(fmt.Sprintf("invalid default_provider '%s', must be one of: %s",
			provider, strings.Join(validProviders, ", ")))
	}
}

// validateProviders validates all provider configurations
func (v *Validator) validateProviders(providers map[string]Provider) {
	for name, provider := range providers {
		v.validateProvider(name, provider)
	}
}

// validateProvider validates a single provider configuration
func (v *Validator) validateProvider(name string, provider Provider) {
	// Validate provider name
	validProviders := []string{ProviderGitHub, ProviderGitLab, ProviderGitea, ProviderGogs}
	if !contains(validProviders, name) {
		v.addError(fmt.Sprintf("invalid provider name '%s', must be one of: %s",
			name, strings.Join(validProviders, ", ")))
	}

	// Validate token
	if provider.Token == "" {
		v.addError(fmt.Sprintf("provider '%s': missing required field 'token'", name))
	} else {
		v.validateToken(name, provider.Token)
	}

	// Validate organizations/groups based on provider type
	switch name {
	case ProviderGitLab:
		if len(provider.Groups) == 0 && len(provider.Orgs) > 0 {
			v.addWarning(fmt.Sprintf("provider '%s': 'orgs' field should be 'groups' for GitLab", name))
		}
		for i, group := range provider.Groups {
			v.validateGitTarget(fmt.Sprintf("%s.groups[%d]", name, i), group)
		}
	default:
		if len(provider.Orgs) == 0 && len(provider.Groups) > 0 {
			v.addWarning(fmt.Sprintf("provider '%s': 'groups' field should be 'orgs' for %s", name, name))
		}
		for i, org := range provider.Orgs {
			v.validateGitTarget(fmt.Sprintf("%s.orgs[%d]", name, i), org)
		}
	}

	// Check if no targets are configured
	if len(provider.Orgs) == 0 && len(provider.Groups) == 0 {
		v.addWarning(fmt.Sprintf("provider '%s': no organizations or groups configured", name))
	}
}

// validateToken validates token format and provides warnings
func (v *Validator) validateToken(providerName, token string) {
	// Check for environment variable format
	if strings.HasPrefix(token, "${") && strings.HasSuffix(token, "}") {
		envVar := token[2 : len(token)-1]
		if envVar == "" {
			v.addError(fmt.Sprintf("provider '%s': empty environment variable name in token", providerName))
		}
		// Don't validate actual env var value as it may not be set during validation
		return
	}

	// Validate direct token format (basic checks)
	if len(token) < 10 {
		v.addWarning(fmt.Sprintf("provider '%s': token appears to be too short", providerName))
	}

	// Provider-specific token validation
	switch providerName {
	case ProviderGitHub:
		if !strings.HasPrefix(token, "ghp_") && !strings.HasPrefix(token, "github_pat_") {
			v.addWarning(fmt.Sprintf("provider '%s': token does not match expected GitHub token format", providerName))
		}
	case ProviderGitLab:
		if !strings.HasPrefix(token, "glpat-") {
			v.addWarning(fmt.Sprintf("provider '%s': token does not match expected GitLab token format", providerName))
		}
	}
}

// validateGitTarget validates a GitTarget configuration
func (v *Validator) validateGitTarget(path string, target GitTarget) {
	// Validate required fields
	if target.Name == "" {
		v.addError(fmt.Sprintf("%s: missing required field 'name'", path))
	}

	// Validate visibility
	if target.Visibility != "" {
		validVisibilities := []string{VisibilityPublic, VisibilityPrivate, VisibilityAll}
		if !contains(validVisibilities, target.Visibility) {
			v.addError(fmt.Sprintf("%s: invalid visibility '%s', must be one of: %s",
				path, target.Visibility, strings.Join(validVisibilities, ", ")))
		}
	}

	// Validate strategy
	if target.Strategy != "" {
		validStrategies := []string{StrategyReset, StrategyPull, StrategyFetch}
		if !contains(validStrategies, target.Strategy) {
			v.addError(fmt.Sprintf("%s: invalid strategy '%s', must be one of: %s",
				path, target.Strategy, strings.Join(validStrategies, ", ")))
		}
	}

	// Validate regex pattern
	if target.Match != "" {
		if _, err := regexp.Compile(target.Match); err != nil {
			v.addError(fmt.Sprintf("%s: invalid regex pattern '%s': %v", path, target.Match, err))
		}
	}

	// Validate clone directory
	if target.CloneDir != "" {
		if strings.Contains(target.CloneDir, "..") {
			v.addWarning(fmt.Sprintf("%s: clone_dir contains '..' which may be unsafe", path))
		}
	}

	// Validate exclude patterns
	for i, exclude := range target.Exclude {
		if strings.Contains(exclude, "*") {
			// Basic glob pattern - just warn if it looks suspicious
			if strings.Count(exclude, "*") > 2 {
				v.addWarning(fmt.Sprintf("%s.exclude[%d]: complex glob pattern '%s' may not work as expected",
					path, i, exclude))
			}
		}
	}
}

// addError adds a validation error
func (v *Validator) addError(message string) {
	v.errors = append(v.errors, fmt.Sprintf("ERROR: %s", message))
}

// addWarning adds a validation warning
func (v *Validator) addWarning(message string) {
	v.warnings = append(v.warnings, fmt.Sprintf("WARNING: %s", message))
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ValidateConfigFile validates a configuration file and returns detailed results
func ValidateConfigFile(filename string) (*ValidationResult, error) {
	config, err := LoadConfigFromFile(filename)
	if err != nil {
		return &ValidationResult{
			Valid:    false,
			Errors:   []string{err.Error()},
			Warnings: []string{},
		}, nil
	}

	validator := NewValidator()
	err = validator.ValidateConfig(config)

	return &ValidationResult{
		Valid:    err == nil,
		Errors:   validator.errors,
		Warnings: validator.warnings,
	}, nil
}

// ValidationResult contains the results of configuration validation
type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// HasIssues returns true if there are any errors or warnings
func (r *ValidationResult) HasIssues() bool {
	return len(r.Errors) > 0 || len(r.Warnings) > 0
}
