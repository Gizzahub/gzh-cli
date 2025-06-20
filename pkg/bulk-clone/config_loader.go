package bulk_clone

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConfigPaths defines the paths where config files are searched
var ConfigPaths = []string{
	"./bulk-clone.yaml",
	"./bulk-clone.yml",
}

// GetConfigPaths returns all possible config file paths in order of preference
func GetConfigPaths() []string {
	paths := make([]string, 0, len(ConfigPaths)+2)

	// 1. Current directory
	paths = append(paths, ConfigPaths...)

	// 2. Home directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		paths = append(paths, filepath.Join(homeDir, ".config", "gzh-manager", "bulk-clone.yaml"))
		paths = append(paths, filepath.Join(homeDir, ".config", "gzh-manager", "bulk-clone.yml"))
	}

	// 3. System-wide config
	paths = append(paths, "/etc/gzh-manager/bulk-clone.yaml")
	paths = append(paths, "/etc/gzh-manager/bulk-clone.yml")

	return paths
}

// FindConfigFile searches for config file in predefined locations
func FindConfigFile() (string, error) {
	// Check environment variable first
	if envPath := os.Getenv("GZH_CONFIG_PATH"); envPath != "" {
		if fileExists(envPath) {
			return envPath, nil
		}
		return "", fmt.Errorf("config file specified in GZH_CONFIG_PATH not found: %s", envPath)
	}

	// Check standard paths
	for _, path := range GetConfigPaths() {
		if fileExists(path) {
			return path, nil
		}
	}

	return "", fmt.Errorf("no config file found in standard locations")
}

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*bulkCloneConfig, error) {
	if configPath == "" {
		path, err := FindConfigFile()
		if err != nil {
			return nil, err
		}
		configPath = path
	}

	cfg := &bulkCloneConfig{}
	if err := cfg.ReadConfig(configPath); err != nil {
		return nil, err
	}
	return cfg, nil
}

// GetGithubOrgConfig extracts GitHub organization config from the full config
func (cfg *bulkCloneConfig) GetGithubOrgConfig(orgName string) (*BulkCloneGithub, error) {
	// Check in repo_roots first
	for _, repo := range cfg.RepoRoots {
		if repo.OrgName == orgName {
			return &repo, nil
		}
	}

	// If not found, create from defaults
	if cfg.Default.Github.OrgName == orgName {
		return &BulkCloneGithub{
			RootPath: cfg.Default.Github.RootPath,
			Provider: cfg.Default.Github.Provider,
			Protocol: cfg.Default.Protocol,
			OrgName:  orgName,
		}, nil
	}

	return nil, fmt.Errorf("no configuration found for organization: %s", orgName)
}

// GetGitlabGroupConfig extracts GitLab group config from the full config
func (cfg *bulkCloneConfig) GetGitlabGroupConfig(groupName string) (*BulkCloneGitlab, error) {
	// Note: Current config structure doesn't have repo_roots for GitLab
	// Check defaults
	if cfg.Default.Gitlab.GroupName == groupName {
		return &BulkCloneGitlab{
			RootPath:  cfg.Default.Gitlab.RootPath,
			Provider:  cfg.Default.Gitlab.Provider,
			Url:       cfg.Default.Gitlab.Url,
			Protocol:  cfg.Default.Protocol,
			GroupName: groupName,
			Recursive: cfg.Default.Gitlab.Recursive,
		}, nil
	}

	return nil, fmt.Errorf("no configuration found for group: %s", groupName)
}

// ExpandPath expands environment variables and ~ in paths
func ExpandPath(path string) string {
	if path == "" {
		return path
	}

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		if homeDir, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(homeDir, path[2:])
		}
	}

	// Expand environment variables
	path = os.ExpandEnv(path)

	return path
}
