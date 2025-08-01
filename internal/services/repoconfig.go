// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package services provides business logic services separated from CLI handlers.
// This package implements the service layer pattern to decouple business logic
// from command-line interface concerns.
package services

import (
	"context"
	"fmt"
	"regexp"

	"github.com/gizzahub/gzh-manager-go/internal/env"
	"github.com/gizzahub/gzh-manager-go/pkg/config"
	"github.com/gizzahub/gzh-manager-go/pkg/github"
)

const (
	templateNone = "none"
)

// ConfigurationChange represents a pending configuration change.
type ConfigurationChange struct {
	Repository   string `json:"repository"`
	Setting      string `json:"setting"`
	CurrentValue string `json:"currentValue"`
	NewValue     string `json:"newValue"`
	Action       string `json:"action"` // create, update, delete
}

// ApplyOptions contains options for applying repository configuration.
type ApplyOptions struct {
	Organization string
	Filter       string
	Template     string
	DryRun       bool
	Interactive  bool
	Force        bool
	Token        string
	ConfigFile   string
	Verbose      bool
}

// ListOptions contains options for listing repository configurations.
type ListOptions struct {
	Organization string
	Filter       string
	Format       string
	ShowConfig   bool
	Limit        int
	Token        string
	ConfigFile   string
	Verbose      bool
}

// RepositoryInfo represents repository information with configuration status.
type RepositoryInfo struct {
	Name        string                   `json:"name" yaml:"name"`
	Description string                   `json:"description" yaml:"description"`
	Visibility  string                   `json:"visibility" yaml:"visibility"`
	Template    string                   `json:"template" yaml:"template"`
	Compliant   bool                     `json:"compliant" yaml:"compliant"`
	Issues      int                      `json:"issues" yaml:"issues"`
	Config      *github.RepositoryConfig `json:"config,omitempty" yaml:"config,omitempty"`
}

// RepoConfigService provides repository configuration management functionality.
type RepoConfigService struct {
	client      *github.RepoConfigClient
	environment env.Environment
}

// NewRepoConfigService creates a new repository configuration service.
func NewRepoConfigService() *RepoConfigService {
	return &RepoConfigService{
		environment: env.NewOSEnvironment(),
	}
}

// NewRepoConfigServiceWithClient creates a new service with a custom client.
func NewRepoConfigServiceWithClient(client *github.RepoConfigClient) *RepoConfigService {
	return &RepoConfigService{
		client:      client,
		environment: env.NewOSEnvironment(),
	}
}

// ApplyConfiguration applies configuration changes to repositories.
func (s *RepoConfigService) ApplyConfiguration(ctx context.Context, opts ApplyOptions) error {
	// Setup client
	if err := s.setupClient(opts.Token); err != nil {
		return fmt.Errorf("failed to setup client: %w", err)
	}

	// Get configuration changes to apply
	changes, err := s.getConfigurationChanges(ctx, opts.Organization, opts.Filter, opts.Template)
	if err != nil {
		return fmt.Errorf("failed to get configuration changes: %w", err)
	}

	if len(changes) == 0 {
		return nil // No changes needed
	}

	if opts.DryRun {
		return nil // Dry run - don't apply changes
	}

	// Apply the changes
	return s.applyChanges(ctx, changes, opts.Interactive)
}

// ListRepositories lists repositories with their configuration status.
func (s *RepoConfigService) ListRepositories(ctx context.Context, opts ListOptions) ([]RepositoryInfo, error) {
	// Setup client
	if err := s.setupClient(opts.Token); err != nil {
		return nil, fmt.Errorf("failed to setup client: %w", err)
	}

	// Load repository configuration to check compliance
	var repoConfig *config.RepoConfig
	if opts.ConfigFile != "" {
		var err error
		repoConfig, err = config.LoadRepoConfig(opts.ConfigFile)
		if err != nil && opts.Verbose {
			// Log warning but continue without config
		}
	}

	// List repositories
	repos, err := s.client.ListRepositories(ctx, opts.Organization, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	// Apply filters and limits
	repos = s.applyFiltersAndLimits(repos, opts.Filter, opts.Limit, opts.Verbose)

	// Convert to RepositoryInfo format
	return s.convertToRepositoryInfo(ctx, repos, opts, repoConfig), nil
}

// GetConfigurationChanges retrieves configuration changes for preview.
func (s *RepoConfigService) GetConfigurationChanges(ctx context.Context, organization, filter, template string) ([]ConfigurationChange, error) {
	return s.getConfigurationChanges(ctx, organization, filter, template)
}

// GetAffectedRepoCount returns the number of unique repositories affected by changes.
func (s *RepoConfigService) GetAffectedRepoCount(changes []ConfigurationChange) int {
	repos := make(map[string]bool)
	for _, change := range changes {
		repos[change.Repository] = true
	}
	return len(repos)
}

// setupClient initializes the GitHub client with authentication.
func (s *RepoConfigService) setupClient(token string) error {
	if s.client != nil {
		return nil // Already initialized
	}

	if token == "" {
		token = s.environment.Get(env.CommonEnvironmentKeys.GitHubToken)
	}

	if token == "" {
		return fmt.Errorf("GitHub token is required (use --token flag or GITHUB_TOKEN env var)")
	}

	s.client = github.NewRepoConfigClient(token)
	return nil
}

// getConfigurationChanges retrieves configuration changes for an organization.
func (s *RepoConfigService) getConfigurationChanges(ctx context.Context, organization, filter, template string) ([]ConfigurationChange, error) {
	// This is a mock implementation - in reality, this would:
	// 1. Fetch current repository configurations from GitHub API
	// 2. Load target configurations from templates
	// 3. Generate required changes
	// 4. Apply filter and template constraints if specified
	mockChanges := []ConfigurationChange{
		{
			Repository:   "api-server",
			Setting:      "branch_protection.main.required_reviews",
			CurrentValue: "1",
			NewValue:     "2",
			Action:       "update",
		},
		{
			Repository:   "web-frontend",
			Setting:      "features.wiki",
			CurrentValue: "true",
			NewValue:     "false",
			Action:       "update",
		},
		{
			Repository:   "legacy-service",
			Setting:      "security.delete_head_branches",
			CurrentValue: "false",
			NewValue:     "true",
			Action:       "create",
		},
	}

	// Apply filter if specified
	if filter != "" {
		filteredChanges := []ConfigurationChange{}
		filterRegex, err := regexp.Compile(filter)
		if err != nil {
			return nil, fmt.Errorf("invalid filter pattern: %w", err)
		}

		for _, change := range mockChanges {
			if filterRegex.MatchString(change.Repository) {
				filteredChanges = append(filteredChanges, change)
			}
		}
		mockChanges = filteredChanges
	}

	// Apply template filter if specified
	if template != "" {
		// TODO: In a real implementation, this would filter changes based on template
		// For now, just document the intended behavior
	}

	return mockChanges, nil
}

// applyConfigurationChange applies a single configuration change.
func (s *RepoConfigService) applyConfigurationChange(ctx context.Context, change ConfigurationChange) error {
	// This is a mock implementation - in reality, this would:
	// 1. Use GitHub API to apply the configuration change
	// 2. Handle authentication and rate limiting
	// 3. Verify the change was applied successfully
	// 4. Return appropriate errors if something fails

	// For demonstration, we'll just return success
	return nil
}

// applyChanges applies the configuration changes.
func (s *RepoConfigService) applyChanges(ctx context.Context, changes []ConfigurationChange, interactive bool) error {
	appliedCount := 0

	for _, change := range changes {
		if err := s.applyConfigurationChange(ctx, change); err != nil {
			continue // Skip failed changes but continue with others
		}
		appliedCount++
	}

	return nil
}

// applyFiltersAndLimits applies filtering and limiting to the repository list.
func (s *RepoConfigService) applyFiltersAndLimits(repos []*github.Repository, filter string, limit int, verbose bool) []*github.Repository {
	// Filter repositories if pattern provided
	if filter != "" {
		filterRegex, err := regexp.Compile(filter)
		if err != nil && verbose {
			// Log warning but continue without filtering
		} else {
			var filtered []*github.Repository
			for _, repo := range repos {
				if filterRegex.MatchString(repo.Name) {
					filtered = append(filtered, repo)
				}
			}
			repos = filtered
		}
	}

	// Apply limit if specified
	if limit > 0 && len(repos) > limit {
		repos = repos[:limit]
	}

	return repos
}

// convertToRepositoryInfo converts GitHub repositories to RepositoryInfo format.
func (s *RepoConfigService) convertToRepositoryInfo(ctx context.Context, repos []*github.Repository, opts ListOptions, repoConfig *config.RepoConfig) []RepositoryInfo {
	repositories := make([]RepositoryInfo, 0, len(repos))

	for _, repo := range repos {
		info := s.createRepositoryInfo(repo, repoConfig)

		if opts.ShowConfig {
			s.addDetailedConfiguration(ctx, opts.Organization, &info)
		}

		repositories = append(repositories, info)
	}

	return repositories
}

// createRepositoryInfo creates a RepositoryInfo from a GitHub repository.
func (s *RepoConfigService) createRepositoryInfo(repo *github.Repository, repoConfig *config.RepoConfig) RepositoryInfo {
	visibility := "public"
	if repo.Private {
		visibility = "private"
	}

	return RepositoryInfo{
		Name:        repo.Name,
		Description: repo.Description,
		Visibility:  visibility,
		Template:    s.detectTemplate(repo, repoConfig),
		Compliant:   s.checkCompliance(repo, repoConfig),
		Issues:      0, // Could be calculated based on actual compliance checks
	}
}

// addDetailedConfiguration adds detailed repository configuration if requested.
func (s *RepoConfigService) addDetailedConfiguration(ctx context.Context, organization string, info *RepositoryInfo) {
	repoConfig, err := s.client.GetRepositoryConfiguration(ctx, organization, info.Name)
	if err == nil {
		info.Config = repoConfig
	}
}

// detectTemplate attempts to detect which template a repository is using.
func (s *RepoConfigService) detectTemplate(repo *github.Repository, repoConfig *config.RepoConfig) string {
	if repoConfig == nil || repoConfig.Repositories == nil {
		return templateNone
	}

	// Check specific repositories
	for _, specific := range repoConfig.Repositories.Specific {
		if specific.Name == repo.Name && specific.Template != "" {
			return specific.Template
		}
	}

	// Check patterns
	for _, pattern := range repoConfig.Repositories.Patterns {
		if matched, err := s.matchPattern(repo.Name, pattern.Match); err == nil && matched && pattern.Template != "" {
			return pattern.Template
		}
	}

	// Check default
	if repoConfig.Repositories.Default != nil && repoConfig.Repositories.Default.Template != "" {
		return repoConfig.Repositories.Default.Template
	}

	return templateNone
}

// checkCompliance checks if a repository is compliant with its template.
func (s *RepoConfigService) checkCompliance(repo *github.Repository, repoConfig *config.RepoConfig) bool {
	// Simple compliance check - can be expanded
	// For now, just check if it has a template assigned
	template := s.detectTemplate(repo, repoConfig)
	return template != "none"
}

// matchPattern checks if a string matches a pattern (simple glob support).
func (s *RepoConfigService) matchPattern(str, pattern string) (bool, error) {
	if len(pattern) > 0 && pattern[0] == '*' || pattern[len(pattern)-1] == '*' {
		// Convert simple glob to regex
		regexPattern := regexp.QuoteMeta(pattern)
		regexPattern = regexp.MustCompile(`\\\*`).ReplaceAllString(regexPattern, ".*")
		regexPattern = "^" + regexPattern + "$"

		return regexp.MatchString(regexPattern, str)
	}

	return str == pattern, nil
}
