package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	configservice "github.com/gizzahub/gzh-manager-go/internal/config"
	"github.com/gizzahub/gzh-manager-go/internal/env"
	configpkg "github.com/gizzahub/gzh-manager-go/pkg/config"
)

// validateConfig performs configuration validation
func validateConfig(configFile string, strict bool, verbose bool) error {
	return validateConfigWithEnv(configFile, strict, verbose, env.NewOSEnvironment())
}

// validateConfigWithEnv performs configuration validation with a specific environment
func validateConfigWithEnv(configFile string, strict bool, verbose bool, environment env.Environment) error {
	ctx := context.Background()

	// Step 1: Determine config file path
	if configFile == "" {
		var err error
		configFile, err = findConfigFileWithEnv(environment)
		if err != nil {
			return fmt.Errorf("failed to find configuration file: %w", err)
		}
	}

	if verbose {
		fmt.Printf("Validating configuration file: %s\n", configFile)
	}

	// Step 2: Check file existence and accessibility
	if err := validateFileAccess(configFile); err != nil {
		return fmt.Errorf("file access validation failed: %w", err)
	}

	if verbose {
		fmt.Println("✓ File access validation passed")
	}

	// Step 3: Create configuration service with enhanced validation
	serviceOptions := configservice.DefaultConfigServiceOptions()
	serviceOptions.ValidationEnabled = true
	service, err := configservice.NewConfigService(serviceOptions)
	if err != nil {
		return fmt.Errorf("failed to create configuration service: %w", err)
	}

	// Step 4: Load and validate configuration using the service
	config, err := service.LoadConfiguration(ctx, configFile)
	if err != nil {
		return fmt.Errorf("configuration loading and validation failed: %w", err)
	}

	if verbose {
		fmt.Println("✓ Configuration loading and startup validation successful")
		if config != nil {
			fmt.Printf("  - Version: %s\n", config.Version)
			fmt.Printf("  - Default Provider: %s\n", config.DefaultProvider)
			fmt.Printf("  - Providers: %d\n", len(config.Providers))
		}
	}

	// Step 5: Get detailed validation results for reporting
	validationResult := service.GetValidationResult()
	if validationResult != nil {
		// Report warnings
		for _, warning := range validationResult.Warnings {
			fmt.Printf("⚠ Warning [%s]: %s\n", warning.Field, warning.Message)
		}

		// In strict mode, treat warnings as errors
		if strict && len(validationResult.Warnings) > 0 {
			return fmt.Errorf("validation failed in strict mode due to %d warnings", len(validationResult.Warnings))
		}
	}

	// Step 6: Perform legacy additional validations if requested
	if strict {
		if err := performLegacyValidationsWithEnv(configFile, verbose, environment); err != nil {
			return fmt.Errorf("legacy validation failed: %w", err)
		}
	}

	// Step 7: Success message
	fmt.Printf("✓ Configuration validation successful: %s\n", configFile)

	if verbose && config != nil {
		printUnifiedConfigurationSummary(config)
	}

	return nil
}

// findConfigFile searches for configuration file in standard locations
func findConfigFile() (string, error) {
	return findConfigFileWithEnv(env.NewOSEnvironment())
}

// findConfigFileWithEnv searches for configuration file using provided environment
func findConfigFileWithEnv(environment env.Environment) (string, error) {
	homeDir := environment.Get(env.CommonEnvironmentKeys.HomeDir)
	searchPaths := []string{
		"./gzh.yaml",
		"./gzh.yml",
		filepath.Join(homeDir, ".config", "gzh.yaml"),
		filepath.Join(homeDir, ".config", "gzh.yml"),
	}

	// Check environment variable
	if envPath := environment.Get(env.CommonEnvironmentKeys.GZHConfigPath); envPath != "" {
		searchPaths = append([]string{envPath}, searchPaths...)
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no configuration file found in standard locations: %v", searchPaths)
}

// validateFileAccess checks if the file can be read
func validateFileAccess(configFile string) error {
	// Check if file exists
	info, err := os.Stat(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("configuration file does not exist: %s", configFile)
		}
		return fmt.Errorf("failed to access file: %w", err)
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return fmt.Errorf("not a regular file: %s", configFile)
	}

	// Check read permissions
	file, err := os.Open(configFile)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}
	file.Close()

	return nil
}

// performLegacyValidations runs legacy validation checks for backwards compatibility
func performLegacyValidations(configFile string, verbose bool) error {
	return performLegacyValidationsWithEnv(configFile, verbose, env.NewOSEnvironment())
}

// performLegacyValidationsWithEnv runs legacy validation checks with provided environment
func performLegacyValidationsWithEnv(configFile string, verbose bool, environment env.Environment) error {
	// Parse configuration using legacy parser
	config, err := configpkg.ParseYAMLFile(configFile)
	if err != nil {
		return fmt.Errorf("legacy configuration parsing failed: %w", err)
	}

	return performAdditionalValidationsWithEnv(config, true, verbose, environment)
}

// performAdditionalValidations runs additional validation checks
func performAdditionalValidations(config *configpkg.Config, strict bool, verbose bool) error {
	return performAdditionalValidationsWithEnv(config, strict, verbose, env.NewOSEnvironment())
}

// performAdditionalValidationsWithEnv runs additional validation checks with provided environment
func performAdditionalValidationsWithEnv(config *configpkg.Config, strict bool, verbose bool, environment env.Environment) error {
	var warnings []string
	var errors []string

	// Validate provider configurations
	for providerName, provider := range config.Providers {
		if verbose {
			fmt.Printf("  Validating provider: %s\n", providerName)
		}

		// Check if provider has organizations or groups
		if len(provider.Orgs) == 0 && len(provider.Groups) == 0 {
			warning := fmt.Sprintf("provider '%s' has no organizations or groups configured", providerName)
			warnings = append(warnings, warning)
		}

		// Validate each organization
		for i, org := range provider.Orgs {
			if err := validateOrganization(providerName, i, org, strict); err != nil {
				errors = append(errors, err.Error())
			}
		}

		// Validate each group
		for i, group := range provider.Groups {
			if err := validateGroup(providerName, i, group, strict); err != nil {
				errors = append(errors, err.Error())
			}
		}
	}

	// Print warnings
	for _, warning := range warnings {
		if verbose {
			fmt.Printf("⚠ Warning: %s\n", warning)
		}
	}

	// Return errors if any
	if len(errors) > 0 {
		return fmt.Errorf("validation errors:\n  - %s", strings.Join(errors, "\n  - "))
	}

	if verbose && len(warnings) == 0 && len(errors) == 0 {
		fmt.Println("✓ Additional validations passed")
	}

	return nil
}

// validateOrganization validates an organization configuration
func validateOrganization(providerName string, index int, org configpkg.GitTarget, strict bool) error {
	prefix := fmt.Sprintf("provider '%s' org[%d]", providerName, index)

	// Validate clone directory
	if org.CloneDir != "" {
		if err := validateCloneDirectory(org.CloneDir, strict); err != nil {
			return fmt.Errorf("%s: %w", prefix, err)
		}
	}

	// Validate regex patterns
	if org.Match != "" {
		if _, err := configpkg.CompileRegex(org.Match); err != nil {
			return fmt.Errorf("%s: invalid match pattern '%s': %w", prefix, org.Match, err)
		}
	}

	// Validate exclude patterns
	for i, pattern := range org.Exclude {
		if _, err := configpkg.CompileRegex(pattern); err != nil {
			return fmt.Errorf("%s: invalid exclude pattern[%d] '%s': %w", prefix, i, pattern, err)
		}
	}

	return nil
}

// validateGroup validates a group configuration (same as organization for now)
func validateGroup(providerName string, index int, group configpkg.GitTarget, strict bool) error {
	prefix := fmt.Sprintf("provider '%s' group[%d]", providerName, index)

	// Same validation as organization
	if group.CloneDir != "" {
		if err := validateCloneDirectory(group.CloneDir, strict); err != nil {
			return fmt.Errorf("%s: %w", prefix, err)
		}
	}

	if group.Match != "" {
		if _, err := configpkg.CompileRegex(group.Match); err != nil {
			return fmt.Errorf("%s: invalid match pattern '%s': %w", prefix, group.Match, err)
		}
	}

	for i, pattern := range group.Exclude {
		if _, err := configpkg.CompileRegex(pattern); err != nil {
			return fmt.Errorf("%s: invalid exclude pattern[%d] '%s': %w", prefix, i, pattern, err)
		}
	}

	return nil
}

// validateCloneDirectory validates clone directory path
func validateCloneDirectory(cloneDir string, strict bool) error {
	// Expand environment variables for validation
	expandedDir := os.ExpandEnv(cloneDir)

	// Check for potentially unsafe paths
	if strings.Contains(expandedDir, "..") {
		return fmt.Errorf("potentially unsafe path with '..' component: %s", cloneDir)
	}

	// In strict mode, check if directory can be created
	if strict {
		// Try to create the directory (or check if it exists)
		if err := os.MkdirAll(expandedDir, 0o755); err != nil {
			return fmt.Errorf("cannot create clone directory '%s': %w", expandedDir, err)
		}
	}

	return nil
}

// validateEnvironmentVariables checks if required environment variables are accessible
func validateEnvironmentVariables(config *configpkg.Config, verbose bool) error {
	return validateEnvironmentVariablesWithEnv(config, verbose, env.NewOSEnvironment())
}

// validateEnvironmentVariablesWithEnv checks if required environment variables are accessible using provided environment
func validateEnvironmentVariablesWithEnv(config *configpkg.Config, verbose bool, environment env.Environment) error {
	var missingVars []string

	for providerName, provider := range config.Providers {
		// Check token environment variables
		if strings.HasPrefix(provider.Token, "${") && strings.HasSuffix(provider.Token, "}") {
			// Extract variable name from ${VAR} or ${VAR:default}
			varExpr := provider.Token[2 : len(provider.Token)-1]
			varName := varExpr

			// Handle default syntax: ${VAR:default}
			if colonIndex := strings.Index(varExpr, ":"); colonIndex != -1 {
				varName = varExpr[:colonIndex]
			}

			if environment.Get(varName) == "" {
				missingVars = append(missingVars, fmt.Sprintf("%s (for provider '%s')", varName, providerName))
			} else if verbose {
				fmt.Printf("✓ Environment variable found: %s\n", varName)
			}
		}

		// Check clone directory environment variables
		for _, org := range provider.Orgs {
			if err := checkPathEnvironmentVariablesWithEnv(org.CloneDir, environment); err != nil {
				missingVars = append(missingVars, err.Error())
			}
		}

		for _, group := range provider.Groups {
			if err := checkPathEnvironmentVariablesWithEnv(group.CloneDir, environment); err != nil {
				missingVars = append(missingVars, err.Error())
			}
		}
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("missing environment variables:\n  - %s", strings.Join(missingVars, "\n  - "))
	}

	return nil
}

// checkPathEnvironmentVariables checks environment variables in path strings
func checkPathEnvironmentVariables(path string) error {
	return checkPathEnvironmentVariablesWithEnv(path, env.NewOSEnvironment())
}

// checkPathEnvironmentVariablesWithEnv checks environment variables in path strings using provided environment
func checkPathEnvironmentVariablesWithEnv(path string, environment env.Environment) error {
	if path == "" {
		return nil
	}

	// Simple check for ${VAR} patterns
	start := 0
	for {
		startIdx := strings.Index(path[start:], "${")
		if startIdx == -1 {
			break
		}
		startIdx += start

		endIdx := strings.Index(path[startIdx:], "}")
		if endIdx == -1 {
			break
		}
		endIdx += startIdx

		varExpr := path[startIdx+2 : endIdx]
		varName := varExpr

		// Handle default syntax: ${VAR:default}
		if colonIndex := strings.Index(varExpr, ":"); colonIndex != -1 {
			varName = varExpr[:colonIndex]
		}

		if environment.Get(varName) == "" {
			return fmt.Errorf("environment variable '%s' not found in path: %s", varName, path)
		}

		start = endIdx + 1
	}

	return nil
}

// printUnifiedConfigurationSummary prints a summary of the unified configuration
func printUnifiedConfigurationSummary(config *configpkg.UnifiedConfig) {
	fmt.Println("\n📋 Unified Configuration Summary:")
	fmt.Printf("  Version: %s\n", config.Version)
	fmt.Printf("  Default Provider: %s\n", config.DefaultProvider)
	fmt.Printf("  Total Providers: %d\n", len(config.Providers))

	if config.Global != nil {
		fmt.Println("\n  🌐 Global Settings:")
		if config.Global.CloneBaseDir != "" {
			fmt.Printf("    Clone Base Dir: %s\n", config.Global.CloneBaseDir)
		}
		if config.Global.DefaultStrategy != "" {
			fmt.Printf("    Default Strategy: %s\n", config.Global.DefaultStrategy)
		}
		if config.Global.DefaultVisibility != "" {
			fmt.Printf("    Default Visibility: %s\n", config.Global.DefaultVisibility)
		}
		if config.Global.Concurrency != nil {
			fmt.Printf("    Concurrency: clone=%d, update=%d, api=%d\n",
				config.Global.Concurrency.CloneWorkers,
				config.Global.Concurrency.UpdateWorkers,
				config.Global.Concurrency.APIWorkers)
		}
	}

	for providerName, provider := range config.Providers {
		fmt.Printf("\n  📌 Provider: %s\n", providerName)
		fmt.Printf("    Organizations: %d\n", len(provider.Organizations))
		if provider.APIURL != "" {
			fmt.Printf("    API URL: %s\n", provider.APIURL)
		}

		totalOrgs := len(provider.Organizations)
		if totalOrgs > 0 {
			fmt.Printf("    Total Organizations: %d\n", totalOrgs)
			for _, org := range provider.Organizations {
				fmt.Printf("      - %s (%s)\n", org.Name, org.CloneDir)
			}
		}
	}

	if config.Migration != nil {
		fmt.Println("\n  🔄 Migration Info:")
		fmt.Printf("    Source Format: %s\n", config.Migration.SourceFormat)
		if !config.Migration.MigrationDate.IsZero() {
			fmt.Printf("    Migration Date: %s\n", config.Migration.MigrationDate.Format("2006-01-02 15:04:05"))
		}
	}
}

// printConfigurationSummary prints a summary of the legacy configuration
func printConfigurationSummary(config *configpkg.Config) {
	fmt.Println("\n📋 Legacy Configuration Summary:")
	fmt.Printf("  Version: %s\n", config.Version)
	fmt.Printf("  Default Provider: %s\n", config.DefaultProvider)
	fmt.Printf("  Total Providers: %d\n", len(config.Providers))

	for providerName, provider := range config.Providers {
		fmt.Printf("\n  📌 Provider: %s\n", providerName)
		fmt.Printf("    Organizations: %d\n", len(provider.Orgs))
		fmt.Printf("    Groups: %d\n", len(provider.Groups))

		totalTargets := len(provider.Orgs) + len(provider.Groups)
		if totalTargets > 0 {
			fmt.Printf("    Total Targets: %d\n", totalTargets)
		}
	}
}
