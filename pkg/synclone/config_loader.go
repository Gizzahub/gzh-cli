// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package bulkclone

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gizzahub/gzh-manager-go/internal/env"
)

// BulkCloneConfig is the public interface for bulk clone configuration.
type BulkCloneConfig = bulkCloneConfig //nolint:revive // Type alias maintained for public API compatibility

// ConfigPaths defines the paths where config files are searched.
var ConfigPaths = []string{
	"./bulk-clone.yaml",
	"./bulk-clone.yml",
}

// OverlayConfigPaths defines overlay config files that can override base configuration.
var OverlayConfigPaths = []string{
	"./bulk-clone.home.yaml",
	"./bulk-clone.home.yml",
	"./bulk-clone.work.yaml",
	"./bulk-clone.work.yml",
}

// GetConfigPaths returns all possible config file paths in order of preference.
func GetConfigPaths() []string {
	paths := make([]string, 0, len(ConfigPaths)+2)

	// 1. Current directory
	paths = append(paths, ConfigPaths...)

	// 2. Home directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		paths = append(paths,
			filepath.Join(homeDir, ".config", "gzh-manager", "bulk-clone.yaml"),
			filepath.Join(homeDir, ".config", "gzh-manager", "bulk-clone.yml"),
		)
	}

	// 3. System-wide config
	paths = append(paths,
		"/etc/gzh-manager/bulk-clone.yaml",
		"/etc/gzh-manager/bulk-clone.yml",
	)

	return paths
}

// GetOverlayConfigPaths returns overlay config file paths in order of preference.
func GetOverlayConfigPaths() []string {
	paths := make([]string, 0, len(OverlayConfigPaths)+4)

	// 1. Current directory overlays
	paths = append(paths, OverlayConfigPaths...)

	// 2. Home directory overlays
	homeDir, err := os.UserHomeDir()
	if err == nil {
		paths = append(paths,
			filepath.Join(homeDir, ".config", "gzh-manager", "bulk-clone.home.yaml"),
			filepath.Join(homeDir, ".config", "gzh-manager", "bulk-clone.home.yml"),
			filepath.Join(homeDir, ".config", "gzh-manager", "bulk-clone.work.yaml"),
			filepath.Join(homeDir, ".config", "gzh-manager", "bulk-clone.work.yml"),
		)
	}

	return paths
}

// FindConfigFile searches for config file in predefined locations.
// Deprecated: Use config.ConfigFactory.FindConfigFile() for unified configuration loading.
func FindConfigFile() (string, error) {
	return FindConfigFileWithEnv(env.NewOSEnvironment())
}

// FindConfigFileWithEnv searches for config file using the provided environment.
// Deprecated: Use config.ConfigFactory.FindConfigFile() for unified configuration loading.
func FindConfigFileWithEnv(environment env.Environment) (string, error) {
	// Check environment variable first
	if envPath := environment.Get(env.CommonEnvironmentKeys.GZHConfigPath); envPath != "" {
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

// LoadConfig loads configuration from file with overlay support.
// Deprecated: Use config.ConfigFactory.LoadConfig() for unified configuration loading.
func LoadConfig(configPath string) (*BulkCloneConfig, error) {
	if configPath == "" {
		path, err := FindConfigFile()
		if err != nil {
			return nil, err
		}

		configPath = path
	}

	// Load base configuration
	cfg := &bulkCloneConfig{}
	if err := cfg.ReadConfig(configPath); err != nil {
		return nil, err
	}

	// Apply overlays if they exist
	if err := cfg.applyOverlays(); err != nil {
		return nil, fmt.Errorf("failed to apply overlay configurations: %w", err)
	}

	return cfg, nil
}

// LoadConfigWithOverlays loads base config and applies overlay configurations.
func LoadConfigWithOverlays(basePath string, overlayPaths ...string) (*BulkCloneConfig, error) {
	// Load base configuration
	cfg := &bulkCloneConfig{}
	if err := cfg.ReadConfig(basePath); err != nil {
		return nil, fmt.Errorf("failed to read base config: %w", err)
	}

	// Apply specified overlays
	for _, overlayPath := range overlayPaths {
		if fileExists(overlayPath) {
			if err := cfg.applyOverlay(overlayPath); err != nil {
				return nil, fmt.Errorf("failed to apply overlay %s: %w", overlayPath, err)
			}
		}
	}

	return cfg, nil
}

// GetGithubOrgConfig extracts GitHub organization config from the full config.
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

// GetGitlabGroupConfig extracts GitLab group config from the full config.
func (cfg *bulkCloneConfig) GetGitlabGroupConfig(groupName string) (*BulkCloneGitlab, error) {
	// Note: Current config structure doesn't have repo_roots for GitLab
	// Check defaults
	if cfg.Default.Gitlab.GroupName == groupName {
		return &BulkCloneGitlab{
			RootPath:  cfg.Default.Gitlab.RootPath,
			Provider:  cfg.Default.Gitlab.Provider,
			URL:       cfg.Default.Gitlab.URL,
			Protocol:  cfg.Default.Protocol,
			GroupName: groupName,
			Recursive: cfg.Default.Gitlab.Recursive,
		}, nil
	}

	return nil, fmt.Errorf("no configuration found for group: %s", groupName)
}

// ExpandPath expands environment variables and ~ in paths.
func ExpandPath(path string) string {
	return ExpandPathWithEnv(path, env.NewOSEnvironment())
}

// ExpandPathWithEnv expands environment variables and ~ in paths using the provided environment.
func ExpandPathWithEnv(path string, environment env.Environment) string {
	if path == "" {
		return path
	}

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		homeDir := environment.Get(env.CommonEnvironmentKeys.HomeDir)
		if homeDir == "" {
			// Fallback to os.UserHomeDir() for compatibility
			if h, err := os.UserHomeDir(); err == nil {
				homeDir = h
			}
		}

		if homeDir != "" {
			path = filepath.Join(homeDir, path[2:])
		}
	}

	// Expand environment variables
	path = environment.Expand(path)

	return path
}

// applyOverlays applies all found overlay configurations.
func (cfg *bulkCloneConfig) applyOverlays() error {
	overlayPaths := GetOverlayConfigPaths()

	for _, overlayPath := range overlayPaths {
		if fileExists(overlayPath) {
			if err := cfg.applyOverlay(overlayPath); err != nil {
				return fmt.Errorf("failed to apply overlay %s: %w", overlayPath, err)
			}
		}
	}

	return nil
}

// applyOverlay applies a single overlay configuration file.
func (cfg *bulkCloneConfig) applyOverlay(overlayPath string) error {
	overlay := &bulkCloneConfig{}
	if err := overlay.ReadConfigWithoutValidation(overlayPath); err != nil {
		return fmt.Errorf("failed to read overlay config: %w", err)
	}

	// Merge configurations
	cfg.mergeConfig(overlay)

	// Validate merged configuration
	if err := cfg.validateConfig(); err != nil {
		return fmt.Errorf("validation failed after applying overlay: %w", err)
	}

	return nil
}

// mergeConfig merges an overlay configuration into the base configuration.
func (cfg *bulkCloneConfig) mergeConfig(overlay *bulkCloneConfig) {
	// Merge version if specified in overlay
	if overlay.Version != "" {
		cfg.Version = overlay.Version
	}

	// Merge default settings
	cfg.mergeDefaults(&overlay.Default)

	// Merge ignore patterns (append new ones)
	cfg.IgnoreNameRegexes = append(cfg.IgnoreNameRegexes, overlay.IgnoreNameRegexes...)

	// Merge repo_roots (append new ones, override existing by org_name)
	cfg.mergeRepoRoots(overlay.RepoRoots)
}

// mergeDefaults merges default configuration settings.
func (cfg *bulkCloneConfig) mergeDefaults(overlayDefault *bulkCloneDefault) {
	// Merge protocol if specified
	if overlayDefault.Protocol != "" {
		cfg.Default.Protocol = overlayDefault.Protocol
	}

	// Merge GitHub defaults
	if overlayDefault.Github.RootPath != "" {
		cfg.Default.Github.RootPath = overlayDefault.Github.RootPath
	}

	if overlayDefault.Github.Provider != "" {
		cfg.Default.Github.Provider = overlayDefault.Github.Provider
	}

	if overlayDefault.Github.Protocol != "" {
		cfg.Default.Github.Protocol = overlayDefault.Github.Protocol
	}

	if overlayDefault.Github.OrgName != "" {
		cfg.Default.Github.OrgName = overlayDefault.Github.OrgName
	}

	// Merge GitLab defaults
	if overlayDefault.Gitlab.RootPath != "" {
		cfg.Default.Gitlab.RootPath = overlayDefault.Gitlab.RootPath
	}

	if overlayDefault.Gitlab.Provider != "" {
		cfg.Default.Gitlab.Provider = overlayDefault.Gitlab.Provider
	}

	if overlayDefault.Gitlab.URL != "" {
		cfg.Default.Gitlab.URL = overlayDefault.Gitlab.URL
	}

	if overlayDefault.Gitlab.Protocol != "" {
		cfg.Default.Gitlab.Protocol = overlayDefault.Gitlab.Protocol
	}

	if overlayDefault.Gitlab.GroupName != "" {
		cfg.Default.Gitlab.GroupName = overlayDefault.Gitlab.GroupName
	}
	// Recursive is a boolean, so we merge it unconditionally if it's true
	if overlayDefault.Gitlab.Recursive {
		cfg.Default.Gitlab.Recursive = overlayDefault.Gitlab.Recursive
	}
}

// mergeRepoRoots merges repo_roots, replacing existing entries with same org_name.
func (cfg *bulkCloneConfig) mergeRepoRoots(overlayRepoRoots []BulkCloneGithub) {
	for _, overlayRepo := range overlayRepoRoots {
		found := false
		// Replace existing repo with same org_name
		for i, existingRepo := range cfg.RepoRoots {
			if existingRepo.OrgName == overlayRepo.OrgName {
				cfg.RepoRoots[i] = overlayRepo
				found = true

				break
			}
		}
		// If not found, append as new entry
		if !found {
			cfg.RepoRoots = append(cfg.RepoRoots, overlayRepo)
		}
	}
}
