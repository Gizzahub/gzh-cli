package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	configpkg "github.com/gizzahub/gzh-manager-go/pkg/config"
)

// validateConfig performs configuration validation
func validateConfig(configFile string, strict bool, verbose bool) error {
	// Step 1: Determine config file path
	if configFile == "" {
		var err error
		configFile, err = findConfigFile()
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

	// Step 3: Parse and validate configuration
	config, err := configpkg.ParseYAMLFile(configFile)
	if err != nil {
		return fmt.Errorf("configuration parsing failed: %w", err)
	}

	if verbose {
		fmt.Println("✓ Configuration parsing successful")
		fmt.Printf("  - Version: %s\n", config.Version)
		fmt.Printf("  - Default Provider: %s\n", config.DefaultProvider)
		fmt.Printf("  - Providers: %d\n", len(config.Providers))
	}

	// Step 4: Perform additional validations
	if err := performAdditionalValidations(config, strict, verbose); err != nil {
		return fmt.Errorf("additional validation failed: %w", err)
	}

	// Step 5: Check environment variables
	if err := validateEnvironmentVariables(config, verbose); err != nil {
		if strict {
			return fmt.Errorf("environment variable validation failed: %w", err)
		}
		if verbose {
			fmt.Printf("⚠ Warning: %v\n", err)
		}
	}

	// Step 6: Success message
	fmt.Printf("✓ Configuration validation successful: %s\n", configFile)

	if verbose {
		printConfigurationSummary(config)
	}

	return nil
}

// findConfigFile searches for configuration file in standard locations
func findConfigFile() (string, error) {
	searchPaths := []string{
		"./gzh.yaml",
		"./gzh.yml",
		filepath.Join(os.Getenv("HOME"), ".config", "gzh.yaml"),
		filepath.Join(os.Getenv("HOME"), ".config", "gzh.yml"),
	}

	// Check environment variable
	if envPath := os.Getenv("GZH_CONFIG_PATH"); envPath != "" {
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

// performAdditionalValidations runs additional validation checks
func performAdditionalValidations(config *configpkg.Config, strict bool, verbose bool) error {
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

			if os.Getenv(varName) == "" {
				missingVars = append(missingVars, fmt.Sprintf("%s (for provider '%s')", varName, providerName))
			} else if verbose {
				fmt.Printf("✓ Environment variable found: %s\n", varName)
			}
		}

		// Check clone directory environment variables
		for _, org := range provider.Orgs {
			if err := checkPathEnvironmentVariables(org.CloneDir); err != nil {
				missingVars = append(missingVars, err.Error())
			}
		}

		for _, group := range provider.Groups {
			if err := checkPathEnvironmentVariables(group.CloneDir); err != nil {
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

		if os.Getenv(varName) == "" {
			return fmt.Errorf("environment variable '%s' not found in path: %s", varName, path)
		}

		start = endIdx + 1
	}

	return nil
}

// printConfigurationSummary prints a summary of the configuration
func printConfigurationSummary(config *configpkg.Config) {
	fmt.Println("\n📋 Configuration Summary:")
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
