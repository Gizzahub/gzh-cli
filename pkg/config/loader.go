package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// ConfigSearchPaths defines the search order for configuration files
var ConfigSearchPaths = []string{
	"./gzh.yaml",
	"./gzh.yml",
	"~/.config/gzh.yaml",
	"~/.config/gzh.yml",
	"~/.config/gzh-manager/gzh.yaml",
	"~/.config/gzh-manager/gzh.yml",
	"/etc/gzh-manager/gzh.yaml",
	"/etc/gzh-manager/gzh.yml",
}

// LoadConfig loads configuration from the first available file in search paths
func LoadConfig() (*Config, error) {
	// Check environment variable first
	if configPath := os.Getenv("GZH_CONFIG_PATH"); configPath != "" {
		return LoadConfigFromFile(configPath)
	}

	// Search in predefined paths
	for _, path := range ConfigSearchPaths {
		expandedPath := expandPath(path)
		if fileExists(expandedPath) {
			return LoadConfigFromFile(expandedPath)
		}
	}

	return nil, fmt.Errorf("no configuration file found in search paths: %v", ConfigSearchPaths)
}

// LoadConfigFromFile loads configuration from a specific file
func LoadConfigFromFile(filename string) (*Config, error) {
	expandedPath := expandPath(filename)
	return ParseYAMLFile(expandedPath)
}

// FindConfigFile finds the first available configuration file
func FindConfigFile() (string, error) {
	// Check environment variable first
	if configPath := os.Getenv("GZH_CONFIG_PATH"); configPath != "" {
		expandedPath := expandPath(configPath)
		if fileExists(expandedPath) {
			return expandedPath, nil
		}
		return "", fmt.Errorf("config file specified in GZH_CONFIG_PATH does not exist: %s", expandedPath)
	}

	// Search in predefined paths
	for _, path := range ConfigSearchPaths {
		expandedPath := expandPath(path)
		if fileExists(expandedPath) {
			return expandedPath, nil
		}
	}

	return "", fmt.Errorf("no configuration file found in search paths")
}

// GetConfigSearchPaths returns the list of paths where configuration files are searched
func GetConfigSearchPaths() []string {
	paths := make([]string, len(ConfigSearchPaths))
	for i, path := range ConfigSearchPaths {
		paths[i] = expandPath(path)
	}
	return paths
}

// expandPath expands ~ to home directory and resolves relative paths
func expandPath(path string) string {
	if path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path // Return original if we can't get home dir
		}
		return filepath.Join(homeDir, path[1:])
	}

	// Convert to absolute path if relative
	if !filepath.IsAbs(path) {
		if abs, err := filepath.Abs(path); err == nil {
			return abs
		}
	}

	return path
}

// fileExists checks if a file exists and is readable
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// CreateDefaultConfig creates a default configuration file at the specified path
func CreateDefaultConfig(filename string) error {
	defaultConfig := `version: "1.0.0"
default_provider: github

providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "your-org-name"
        visibility: all
        clone_dir: "./github"
        
  gitlab:
    token: "${GITLAB_TOKEN}"
    groups:
      - name: "your-group-name"
        visibility: public
        recursive: true
        clone_dir: "./gitlab"
`

	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config file
	if err := os.WriteFile(filename, []byte(defaultConfig), 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetDefaultConfigPath returns the default path for creating new config files
func GetDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./gzh.yaml" // Fallback to current directory
	}
	return filepath.Join(homeDir, ".config", "gzh.yaml")
}
