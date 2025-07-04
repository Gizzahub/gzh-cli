package config

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// StartupValidator provides comprehensive configuration validation at application startup
type StartupValidator struct {
	validator *validator.Validate
	errors    []StartupValidationError
	warnings  []StartupValidationWarning
}

// StartupValidationError represents a validation error with context
type StartupValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// StartupValidationWarning represents a validation warning
type StartupValidationWarning struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// StartupValidationResult contains the results of startup validation
type StartupValidationResult struct {
	IsValid  bool                       `json:"is_valid"`
	Errors   []StartupValidationError   `json:"errors"`
	Warnings []StartupValidationWarning `json:"warnings"`
	Summary  string                     `json:"summary"`
}

// NewStartupValidator creates a new startup configuration validator
func NewStartupValidator() *StartupValidator {
	v := validator.New()

	sv := &StartupValidator{
		validator: v,
		errors:    make([]StartupValidationError, 0),
		warnings:  make([]StartupValidationWarning, 0),
	}

	// Register custom validation functions
	sv.registerCustomValidators()

	return sv
}

// registerCustomValidators registers custom validation functions
func (sv *StartupValidator) registerCustomValidators() {
	// Register custom strategy validator
	sv.validator.RegisterValidation("strategy", sv.validateStrategy)

	// Register custom provider validator
	sv.validator.RegisterValidation("provider", sv.validateProvider)

	// Register custom visibility validator
	sv.validator.RegisterValidation("visibility", sv.validateVisibility)

	// Register custom directory path validator
	sv.validator.RegisterValidation("dirpath", sv.validateDirectoryPath)

	// Register custom regex pattern validator
	sv.validator.RegisterValidation("regexpattern", sv.validateRegexPattern)

	// Register custom environment token validator
	sv.validator.RegisterValidation("envtoken", sv.validateEnvironmentToken)

	// Register custom timeout duration validator
	sv.validator.RegisterValidation("timeout", sv.validateTimeout)

	// Register custom concurrency validator
	sv.validator.RegisterValidation("concurrency", sv.validateConcurrency)
}

// ValidateUnifiedConfig performs comprehensive validation of unified configuration
func (sv *StartupValidator) ValidateUnifiedConfig(config *UnifiedConfig) *StartupValidationResult {
	sv.reset()

	// Basic struct validation using tags
	if err := sv.validator.Struct(config); err != nil {
		sv.processValidationErrors(err)
	}

	// Custom business logic validation
	sv.validateBusinessRules(config)

	// Check for warnings
	sv.checkConfigurationWarnings(config)

	return sv.buildResult()
}

// ValidateConfig performs validation of regular Config struct
func (sv *StartupValidator) ValidateConfig(config *Config) *StartupValidationResult {
	sv.reset()

	// Basic struct validation
	if err := sv.validator.Struct(config); err != nil {
		sv.processValidationErrors(err)
	}

	// Custom validation for Config
	sv.validateConfigBusinessRules(config)

	return sv.buildResult()
}

// validateBusinessRules performs custom business logic validation
func (sv *StartupValidator) validateBusinessRules(config *UnifiedConfig) {
	// Validate version format
	if config.Version != "" && !isValidVersionFormat(config.Version) {
		sv.addError("Version", "version_format", config.Version, "version must be in semantic versioning format (e.g., 1.0.0)")
	}

	// Validate default provider exists in providers map
	if config.DefaultProvider != "" {
		if _, exists := config.Providers[config.DefaultProvider]; !exists {
			sv.addError("DefaultProvider", "provider_exists", config.DefaultProvider, "default provider must exist in providers configuration")
		}
	}

	// Validate provider configurations
	for providerName, provider := range config.Providers {
		sv.validateProviderConfig(providerName, provider)
	}

	// Validate global settings
	if config.Global != nil {
		sv.validateGlobalSettings(config.Global)
	}
}

// validateConfigBusinessRules performs validation for regular Config struct
func (sv *StartupValidator) validateConfigBusinessRules(config *Config) {
	// Validate version
	if config.Version == "" {
		sv.addError("Version", "required", "", "version is required")
	}

	// Validate providers
	for providerName, provider := range config.Providers {
		sv.validateBasicProviderConfig(providerName, provider)
	}
}

// validateProviderConfig validates a provider configuration
func (sv *StartupValidator) validateProviderConfig(providerName string, provider *ProviderConfig) {
	fieldPrefix := fmt.Sprintf("Providers[%s]", providerName)

	// Validate token format
	if provider.Token == "" {
		sv.addError(fieldPrefix+".Token", "required", "", "provider token is required")
	} else {
		sv.validateTokenFormat(fieldPrefix+".Token", provider.Token)
	}

	// Validate API URL if provided
	if provider.APIURL != "" {
		if !isValidURL(provider.APIURL) {
			sv.addError(fieldPrefix+".APIURL", "url", provider.APIURL, "API URL must be a valid URL")
		}
	}

	// Validate organizations
	for i, org := range provider.Organizations {
		sv.validateOrganizationConfig(fmt.Sprintf("%s.Organizations[%d]", fieldPrefix, i), org)
	}
}

// validateBasicProviderConfig validates basic provider configuration
func (sv *StartupValidator) validateBasicProviderConfig(providerName string, provider Provider) {
	fieldPrefix := fmt.Sprintf("Providers[%s]", providerName)

	if provider.Token == "" {
		sv.addError(fieldPrefix+".Token", "required", "", "provider token is required")
	}

	// Validate orgs
	for i, org := range provider.Orgs {
		sv.validateBasicGitTarget(fmt.Sprintf("%s.Orgs[%d]", fieldPrefix, i), org)
	}

	// Validate groups
	for i, group := range provider.Groups {
		sv.validateBasicGitTarget(fmt.Sprintf("%s.Groups[%d]", fieldPrefix, i), group)
	}
}

// validateBasicGitTarget validates basic GitTarget configuration
func (sv *StartupValidator) validateBasicGitTarget(fieldPrefix string, target GitTarget) {
	if target.Name == "" {
		sv.addError(fieldPrefix+".Name", "required", "", "target name is required")
	}

	// Validate visibility
	if target.Visibility != "" && !isValidVisibility(target.Visibility) {
		sv.addError(fieldPrefix+".Visibility", "visibility", target.Visibility, "visibility must be 'public', 'private', or 'all'")
	}

	// Validate strategy
	if target.Strategy != "" && !isValidStrategy(target.Strategy) {
		sv.addError(fieldPrefix+".Strategy", "strategy", target.Strategy, "strategy must be 'reset', 'pull', or 'fetch'")
	}
}

// validateOrganizationConfig validates an organization configuration
func (sv *StartupValidator) validateOrganizationConfig(fieldPrefix string, org *OrganizationConfig) {
	if org.Name == "" {
		sv.addError(fieldPrefix+".Name", "required", "", "organization name is required")
	}

	if org.CloneDir == "" {
		sv.addError(fieldPrefix+".CloneDir", "required", "", "clone directory is required")
	} else {
		sv.validateDirectoryPathValue(fieldPrefix+".CloneDir", org.CloneDir)
	}

	// Validate visibility
	if org.Visibility != "" && !isValidVisibility(org.Visibility) {
		sv.addError(fieldPrefix+".Visibility", "visibility", org.Visibility, "visibility must be 'public', 'private', or 'all'")
	}

	// Validate strategy
	if org.Strategy != "" && !isValidStrategy(org.Strategy) {
		sv.addError(fieldPrefix+".Strategy", "strategy", org.Strategy, "strategy must be 'reset', 'pull', or 'fetch'")
	}

	// Validate include pattern
	if org.Include != "" {
		if _, err := regexp.Compile(org.Include); err != nil {
			sv.addError(fieldPrefix+".Include", "regexpattern", org.Include, fmt.Sprintf("include pattern is not a valid regex: %v", err))
		}
	}

	// Validate exclude patterns
	for i, pattern := range org.Exclude {
		if _, err := regexp.Compile(pattern); err != nil {
			sv.addError(fmt.Sprintf("%s.Exclude[%d]", fieldPrefix, i), "regexpattern", pattern, fmt.Sprintf("exclude pattern is not a valid regex: %v", err))
		}
	}
}

// validateGlobalSettings validates global settings
func (sv *StartupValidator) validateGlobalSettings(global *GlobalSettings) {
	// Validate default strategy
	if global.DefaultStrategy != "" && !isValidStrategy(global.DefaultStrategy) {
		sv.addError("Global.DefaultStrategy", "strategy", global.DefaultStrategy, "default strategy must be 'reset', 'pull', or 'fetch'")
	}

	// Validate default visibility
	if global.DefaultVisibility != "" && !isValidVisibility(global.DefaultVisibility) {
		sv.addError("Global.DefaultVisibility", "visibility", global.DefaultVisibility, "default visibility must be 'public', 'private', or 'all'")
	}

	// Validate clone base directory
	if global.CloneBaseDir != "" {
		sv.validateDirectoryPathValue("Global.CloneBaseDir", global.CloneBaseDir)
	}

	// Validate timeout settings
	if global.Timeouts != nil {
		sv.validateTimeoutSettings(global.Timeouts)
	}

	// Validate concurrency settings
	if global.Concurrency != nil {
		sv.validateConcurrencySettings(global.Concurrency)
	}
}

// validateTimeoutSettings validates timeout settings
func (sv *StartupValidator) validateTimeoutSettings(timeouts *TimeoutSettings) {
	if timeouts.HTTPTimeout > 0 && timeouts.HTTPTimeout < time.Second {
		sv.addWarning("Global.Timeouts.HTTPTimeout", "HTTP timeout is very short (< 1s), this may cause request failures")
	}

	if timeouts.GitTimeout > 0 && timeouts.GitTimeout < 30*time.Second {
		sv.addWarning("Global.Timeouts.GitTimeout", "Git timeout is short (< 30s), this may cause issues with large repositories")
	}

	if timeouts.RateLimitTimeout > 0 && timeouts.RateLimitTimeout < 5*time.Minute {
		sv.addWarning("Global.Timeouts.RateLimitTimeout", "Rate limit timeout is short (< 5m), this may cause frequent rate limit errors")
	}
}

// validateConcurrencySettings validates concurrency settings
func (sv *StartupValidator) validateConcurrencySettings(concurrency *ConcurrencySettings) {
	if concurrency.CloneWorkers > 50 {
		sv.addWarning("Global.Concurrency.CloneWorkers", "Very high clone worker count (> 50) may overwhelm the system or API rate limits")
	}

	if concurrency.UpdateWorkers > 50 {
		sv.addWarning("Global.Concurrency.UpdateWorkers", "Very high update worker count (> 50) may overwhelm the system or API rate limits")
	}

	if concurrency.APIWorkers > 20 {
		sv.addWarning("Global.Concurrency.APIWorkers", "Very high API worker count (> 20) may trigger rate limits")
	}
}

// checkConfigurationWarnings checks for potential configuration issues
func (sv *StartupValidator) checkConfigurationWarnings(config *UnifiedConfig) {
	// Check if no organizations are configured
	totalOrgs := 0
	for _, provider := range config.Providers {
		totalOrgs += len(provider.Organizations)
	}

	if totalOrgs == 0 {
		sv.addWarning("Providers", "No organizations configured, the application will not have any targets to process")
	}

	// Check for environment variable tokens
	for providerName, provider := range config.Providers {
		if provider.Token != "" && strings.HasPrefix(provider.Token, "${") && strings.HasSuffix(provider.Token, "}") {
			envVar := strings.Trim(provider.Token[2:len(provider.Token)-1], " ")
			if os.Getenv(envVar) == "" {
				sv.addWarning(fmt.Sprintf("Providers[%s].Token", providerName), fmt.Sprintf("Environment variable %s is not set", envVar))
			}
		}
	}
}

// Custom validation functions for validator tags

func (sv *StartupValidator) validateStrategy(fl validator.FieldLevel) bool {
	return isValidStrategy(fl.Field().String())
}

func (sv *StartupValidator) validateProvider(fl validator.FieldLevel) bool {
	return isValidProvider(fl.Field().String())
}

func (sv *StartupValidator) validateVisibility(fl validator.FieldLevel) bool {
	return isValidVisibility(fl.Field().String())
}

func (sv *StartupValidator) validateDirectoryPath(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	return path != "" && (strings.HasPrefix(path, "/") || strings.HasPrefix(path, "~") || strings.HasPrefix(path, "$") || strings.Contains(path, ":"))
}

func (sv *StartupValidator) validateRegexPattern(fl validator.FieldLevel) bool {
	pattern := fl.Field().String()
	if pattern == "" {
		return true // Empty patterns are allowed
	}
	_, err := regexp.Compile(pattern)
	return err == nil
}

func (sv *StartupValidator) validateEnvironmentToken(fl validator.FieldLevel) bool {
	token := fl.Field().String()
	if token == "" {
		return false
	}
	// Allow direct tokens or environment variable format
	return !strings.Contains(token, " ") // Basic check for no spaces
}

func (sv *StartupValidator) validateTimeout(fl validator.FieldLevel) bool {
	// Timeout should be positive
	duration := fl.Field().Interface().(time.Duration)
	return duration > 0
}

func (sv *StartupValidator) validateConcurrency(fl validator.FieldLevel) bool {
	// Concurrency should be positive and reasonable
	count := int(fl.Field().Int())
	return count > 0 && count <= 100
}

// Helper functions

func isValidVersionFormat(version string) bool {
	// Simple semantic version validation
	matched, _ := regexp.MatchString(`^\d+\.\d+\.\d+(-[\w.-]+)?(\+[\w.-]+)?$`, version)
	return matched
}

func isValidURL(urlStr string) bool {
	_, err := url.Parse(urlStr)
	return err == nil
}

func isValidStrategy(strategy string) bool {
	switch strategy {
	case "reset", "pull", "fetch":
		return true
	default:
		return false
	}
}

func isValidProvider(provider string) bool {
	switch provider {
	case "github", "gitlab", "gitea", "gogs":
		return true
	default:
		return false
	}
}

func isValidVisibility(visibility string) bool {
	switch visibility {
	case "public", "private", "all":
		return true
	default:
		return false
	}
}

func (sv *StartupValidator) validateTokenFormat(field, token string) {
	// Check if it's an environment variable format
	if strings.HasPrefix(token, "${") && strings.HasSuffix(token, "}") {
		envVar := strings.Trim(token[2:len(token)-1], " ")
		if envVar == "" {
			sv.addError(field, "envtoken", token, "environment variable name cannot be empty")
		}
		return
	}

	// For direct tokens, check basic format (should be alphanumeric with some special chars)
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_.-]+$`, token); !matched {
		sv.addWarning(field, "Token contains special characters that may cause issues")
	}
}

func (sv *StartupValidator) validateDirectoryPathValue(field, path string) {
	// Check for environment variable expansion
	if strings.Contains(path, "$") {
		return // Environment variables are valid
	}

	// Check if path is absolute or relative with home directory
	if !strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "~") && !strings.HasPrefix(path, ".") {
		sv.addWarning(field, "Directory path should be absolute or start with ~, $, or .")
	}
}

// Utility methods

func (sv *StartupValidator) reset() {
	sv.errors = make([]StartupValidationError, 0)
	sv.warnings = make([]StartupValidationWarning, 0)
}

func (sv *StartupValidator) addError(field, tag, value, message string) {
	sv.errors = append(sv.errors, StartupValidationError{
		Field:   field,
		Tag:     tag,
		Value:   value,
		Message: message,
	})
}

func (sv *StartupValidator) addWarning(field, message string) {
	sv.warnings = append(sv.warnings, StartupValidationWarning{
		Field:   field,
		Message: message,
	})
}

func (sv *StartupValidator) processValidationErrors(err error) {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, ve := range validationErrors {
			sv.addError(
				ve.StructField(),
				ve.Tag(),
				fmt.Sprintf("%v", ve.Value()),
				sv.getErrorMessage(ve),
			)
		}
	}
}

func (sv *StartupValidator) getErrorMessage(ve validator.FieldError) string {
	switch ve.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", ve.Field())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", ve.Field(), ve.Param())
	case "min":
		return fmt.Sprintf("%s must have at least %s items", ve.Field(), ve.Param())
	case "max":
		return fmt.Sprintf("%s must have at most %s items", ve.Field(), ve.Param())
	case "strategy":
		return fmt.Sprintf("%s must be 'reset', 'pull', or 'fetch'", ve.Field())
	case "provider":
		return fmt.Sprintf("%s must be 'github', 'gitlab', 'gitea', or 'gogs'", ve.Field())
	case "visibility":
		return fmt.Sprintf("%s must be 'public', 'private', or 'all'", ve.Field())
	case "dirpath":
		return fmt.Sprintf("%s must be a valid directory path", ve.Field())
	case "regexpattern":
		return fmt.Sprintf("%s must be a valid regular expression", ve.Field())
	case "envtoken":
		return fmt.Sprintf("%s must be a valid token or environment variable", ve.Field())
	default:
		return fmt.Sprintf("%s failed validation for tag '%s'", ve.Field(), ve.Tag())
	}
}

func (sv *StartupValidator) buildResult() *StartupValidationResult {
	isValid := len(sv.errors) == 0

	var summary string
	if isValid {
		if len(sv.warnings) > 0 {
			summary = fmt.Sprintf("Configuration is valid with %d warnings", len(sv.warnings))
		} else {
			summary = "Configuration is valid"
		}
	} else {
		summary = fmt.Sprintf("Configuration is invalid with %d errors", len(sv.errors))
		if len(sv.warnings) > 0 {
			summary += fmt.Sprintf(" and %d warnings", len(sv.warnings))
		}
	}

	return &StartupValidationResult{
		IsValid:  isValid,
		Errors:   sv.errors,
		Warnings: sv.warnings,
		Summary:  summary,
	}
}
