package config

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseYAML parses YAML content from a reader and returns a Config
func ParseYAML(reader io.Reader) (*Config, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML content: %w", err)
	}

	// Expand environment variables
	expandedContent := os.ExpandEnv(string(content))

	var config Config
	if err := yaml.Unmarshal([]byte(expandedContent), &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", ErrInvalidYAML)
	}

	// Set defaults for all GitTargets
	for providerName, provider := range config.Providers {
		for i := range provider.Orgs {
			provider.Orgs[i].SetDefaults()
		}
		for i := range provider.Groups {
			provider.Groups[i].SetDefaults()
		}
		config.Providers[providerName] = provider
	}

	// Validate the configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// ParseYAMLFile parses a YAML file and returns a Config
func ParseYAMLFile(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("configuration file not found: %s", filename)
		}
		return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	return ParseYAML(file)
}

// ExpandEnvironmentVariables expands environment variables in a string
// Supports both ${VAR} and $VAR formats
func ExpandEnvironmentVariables(input string) string {
	return os.ExpandEnv(input)
}

// LoadYAMLWithEnvSubstitution loads YAML with environment variable substitution
func LoadYAMLWithEnvSubstitution(filename string) (*Config, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Pre-process for custom environment variable handling
	processedContent := preprocessEnvVars(string(content))

	// Apply standard environment variable expansion
	expandedContent := os.ExpandEnv(processedContent)

	var config Config
	if err := yaml.Unmarshal([]byte(expandedContent), &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", ErrInvalidYAML)
	}

	// Apply defaults and validate
	config.applyDefaults()
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// preprocessEnvVars handles custom environment variable processing
func preprocessEnvVars(content string) string {
	// Handle ${VAR:default} syntax (fallback to default if VAR is not set)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, "${") && strings.Contains(line, ":") {
			lines[i] = processDefaultValues(line)
		}
	}
	return strings.Join(lines, "\n")
}

// processDefaultValues processes ${VAR:default} syntax
func processDefaultValues(line string) string {
	// Simple implementation for ${VAR:default} pattern
	start := strings.Index(line, "${")
	if start == -1 {
		return line
	}

	end := strings.Index(line[start:], "}")
	if end == -1 {
		return line
	}

	envExpr := line[start+2 : start+end]
	if colonIndex := strings.Index(envExpr, ":"); colonIndex != -1 {
		varName := envExpr[:colonIndex]
		defaultValue := envExpr[colonIndex+1:]

		if value := os.Getenv(varName); value != "" {
			return strings.Replace(line, "${"+envExpr+"}", value, 1)
		}
		return strings.Replace(line, "${"+envExpr+"}", defaultValue, 1)
	}

	return line
}

// applyDefaults applies default values to the configuration
func (c *Config) applyDefaults() {
	if c.DefaultProvider == "" {
		c.DefaultProvider = ProviderGitHub
	}

	for providerName, provider := range c.Providers {
		for i := range provider.Orgs {
			provider.Orgs[i].SetDefaults()
		}
		for i := range provider.Groups {
			provider.Groups[i].SetDefaults()
		}
		c.Providers[providerName] = provider
	}
}
