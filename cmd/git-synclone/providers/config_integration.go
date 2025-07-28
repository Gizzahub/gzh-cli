// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package providers

import (
	"fmt"

	bulkclone "github.com/gizzahub/gzh-manager-go/pkg/synclone"
)

// ConfigIntegration provides utilities for integrating synclone configuration
// with the Git extension commands.
type ConfigIntegration struct {
	config *bulkclone.BulkCloneConfig
}

// NewConfigIntegration creates a new config integration instance.
func NewConfigIntegration() *ConfigIntegration {
	return &ConfigIntegration{}
}

// LoadConfig loads the synclone configuration from file or standard locations.
func (c *ConfigIntegration) LoadConfig(configFile string) error {
	var err error
	if configFile != "" {
		c.config, err = bulkclone.LoadConfig(configFile)
	} else {
		c.config, err = bulkclone.LoadConfig("")
	}
	return err
}

// GetGitHubOrgConfig extracts GitHub organization configuration.
func (c *ConfigIntegration) GetGitHubOrgConfig(orgName string) (*GitHubConfigResult, error) {
	if c.config == nil {
		return nil, fmt.Errorf("configuration not loaded")
	}

	ghConfig, err := c.config.GetGithubOrgConfig(orgName)
	if err != nil {
		return nil, err
	}

	return &GitHubConfigResult{
		OrgName:  ghConfig.OrgName,
		RootPath: bulkclone.ExpandPath(ghConfig.RootPath),
		Provider: ghConfig.Provider,
		Protocol: ghConfig.Protocol,
	}, nil
}

// GetGitLabGroupConfig extracts GitLab group configuration.
func (c *ConfigIntegration) GetGitLabGroupConfig(groupName string) (*GitLabConfigResult, error) {
	if c.config == nil {
		return nil, fmt.Errorf("configuration not loaded")
	}

	glConfig, err := c.config.GetGitlabGroupConfig(groupName)
	if err != nil {
		return nil, err
	}

	return &GitLabConfigResult{
		GroupName: glConfig.GroupName,
		RootPath:  bulkclone.ExpandPath(glConfig.RootPath),
		Provider:  glConfig.Provider,
		Protocol:  glConfig.Protocol,
		URL:       glConfig.URL,
		Recursive: glConfig.Recursive,
	}, nil
}

// BuildCloneOptionsFromConfig creates CloneOptions from loaded configuration.
func (c *ConfigIntegration) BuildCloneOptionsFromConfig(provider, orgName string) (*CloneOptions, error) {
	if c.config == nil {
		return nil, fmt.Errorf("configuration not loaded")
	}

	switch provider {
	case "github":
		ghConfig, err := c.GetGitHubOrgConfig(orgName)
		if err != nil {
			return nil, err
		}
		return &CloneOptions{
			Protocol:       ghConfig.Protocol,
			Strategy:       "reset", // default strategy
			Parallel:       1,       // default parallel
			MaxRetries:     3,       // default retries
			Resume:         false,
			DryRun:         false,
			ProgressMode:   "bar",
			Token:          "",
			ConfigFile:     "",
			UseConfig:      true,
			CleanupOrphans: false,
		}, nil

	case "gitlab":
		glConfig, err := c.GetGitLabGroupConfig(orgName)
		if err != nil {
			return nil, err
		}
		return &CloneOptions{
			Protocol:       glConfig.Protocol,
			Strategy:       "reset", // default strategy
			Parallel:       1,       // default parallel
			MaxRetries:     3,       // default retries
			Resume:         false,
			DryRun:         false,
			ProgressMode:   "bar",
			Token:          "",
			ConfigFile:     "",
			UseConfig:      true,
			CleanupOrphans: false,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// GitHubConfigResult represents extracted GitHub configuration.
type GitHubConfigResult struct {
	OrgName  string
	RootPath string
	Provider string
	Protocol string
}

// GitLabConfigResult represents extracted GitLab configuration.
type GitLabConfigResult struct {
	GroupName string
	RootPath  string
	Provider  string
	Protocol  string
	URL       string
	Recursive bool
}

// GetConfigPaths returns the configuration file search paths.
func (c *ConfigIntegration) GetConfigPaths() []string {
	return bulkclone.GetConfigPaths()
}

// GetOverlayConfigPaths returns the overlay configuration file paths.
func (c *ConfigIntegration) GetOverlayConfigPaths() []string {
	return bulkclone.GetOverlayConfigPaths()
}

// FindConfigFile searches for config file in predefined locations.
func (c *ConfigIntegration) FindConfigFile() (string, error) {
	return bulkclone.FindConfigFile()
}

// ValidateConfig validates the loaded configuration.
func (c *ConfigIntegration) ValidateConfig() error {
	if c.config == nil {
		return fmt.Errorf("configuration not loaded")
	}
	// The configuration is already validated during loading
	// Additional validation can be added here if needed
	return nil
}
