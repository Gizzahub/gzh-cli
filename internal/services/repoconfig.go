// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

// Package services provides business logic services separated from CLI handlers.
// This package implements the service layer pattern to decouple business logic
// from command-line interface concerns.
package services

import (
	"context"
	"fmt"
	"regexp"

	"github.com/gizzahub/gzh-cli/internal/env"
	"github.com/gizzahub/gzh-cli/pkg/config"
	"github.com/gizzahub/gzh-cli/pkg/github"
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
func (s *RepoConfigService) ApplyConfiguration(_ context.Context, _ ApplyOptions) error {
	return unsupportedRepositoryApply()
}

// 실제 병합·확인·부분 실패 계약 없이 고정 변경이나 적용 성공을 만들지 않는다.
func unsupportedRepositoryApply() error {
	return fmt.Errorf("repository configuration apply and change planning are not supported")
}

// ListRepositories lists repositories with their configuration status.
func (s *RepoConfigService) ListRepositories(ctx context.Context, opts ListOptions) ([]RepositoryInfo, error) {
	// 명시한 입력의 오류는 인증·HTTP 조회보다 먼저 반환한다.
	var filterRegex *regexp.Regexp
	if opts.Filter != "" {
		var err error
		filterRegex, err = regexp.Compile(opts.Filter)
		if err != nil {
			return nil, fmt.Errorf("invalid filter regex %q: %w", opts.Filter, err)
		}
	}
	var repoConfig *config.RepoConfig
	if opts.ConfigFile != "" {
		var err error
		repoConfig, err = config.LoadRepoConfig(opts.ConfigFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load config file %s: %w", opts.ConfigFile, err)
		}
	}
	if err := s.setupClient(opts.Token); err != nil {
		return nil, fmt.Errorf("failed to setup client: %w", err)
	}

	// List repositories
	repos, err := s.client.ListRepositories(ctx, opts.Organization, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	// Apply filters and limits
	repos = s.applyFiltersAndLimits(repos, filterRegex, opts.Limit)

	// Convert to RepositoryInfo format
	return s.convertToRepositoryInfo(ctx, repos, opts, repoConfig)
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
func (s *RepoConfigService) getConfigurationChanges(_ context.Context, _, _, _ string) ([]ConfigurationChange, error) {
	return nil, unsupportedRepositoryApply()
}

// applyConfigurationChange는 연결되지 않은 적용을 성공으로 반환하지 않는다.
func (s *RepoConfigService) applyConfigurationChange(_ context.Context, _ ConfigurationChange) error {
	return unsupportedRepositoryApply()
}

// applyChanges는 빈 목록과 preview에서도 미구현 적용을 숨기지 않는다.
func (s *RepoConfigService) applyChanges(_ context.Context, _ []ConfigurationChange, _ bool) error {
	return unsupportedRepositoryApply()
}

// applyFiltersAndLimits applies filtering and limiting to the repository list.
func (s *RepoConfigService) applyFiltersAndLimits(repos []*github.Repository, filterRegex *regexp.Regexp, limit int) []*github.Repository {
	if filterRegex != nil {
		var filtered []*github.Repository
		for _, repo := range repos {
			if filterRegex.MatchString(repo.Name) {
				filtered = append(filtered, repo)
			}
		}
		repos = filtered
	}

	// Apply limit if specified
	if limit > 0 && len(repos) > limit {
		repos = repos[:limit]
	}

	return repos
}

// convertToRepositoryInfo converts GitHub repositories to RepositoryInfo format.
func (s *RepoConfigService) convertToRepositoryInfo(ctx context.Context, repos []*github.Repository, opts ListOptions, repoConfig *config.RepoConfig) ([]RepositoryInfo, error) {
	repositories := make([]RepositoryInfo, 0, len(repos))

	for _, repo := range repos {
		info := s.createRepositoryInfo(repo, repoConfig)

		if opts.ShowConfig {
			if err := s.addDetailedConfiguration(ctx, opts.Organization, &info); err != nil {
				return nil, err
			}
		}

		repositories = append(repositories, info)
	}

	return repositories, nil
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
func (s *RepoConfigService) addDetailedConfiguration(ctx context.Context, organization string, info *RepositoryInfo) error {
	repoConfig, err := s.client.GetRepositoryConfiguration(ctx, organization, info.Name)
	if err != nil {
		return fmt.Errorf("failed to get configuration for %s/%s: %w", organization, info.Name, err)
	}
	info.Config = repoConfig
	return nil
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
	if pattern != "" && pattern[0] == '*' || pattern[len(pattern)-1] == '*' {
		// Convert simple glob to regex
		regexPattern := regexp.QuoteMeta(pattern)
		regexPattern = regexp.MustCompile(`\\\*`).ReplaceAllString(regexPattern, ".*")
		regexPattern = "^" + regexPattern + "$"

		return regexp.MatchString(regexPattern, str)
	}

	return str == pattern, nil
}
