// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package providers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gizzahub/gzh-manager-go/internal/env"
	"github.com/gizzahub/gzh-manager-go/pkg/github"
)

// GitHubAdapter adapts the existing GitHub synclone functionality for Git extensions.
type GitHubAdapter struct {
	*BaseProviderAdapter
}

// NewGitHubAdapter creates a new GitHub provider adapter.
func NewGitHubAdapter() *GitHubAdapter {
	return &GitHubAdapter{
		BaseProviderAdapter: NewBaseProviderAdapter(),
	}
}

// GetProviderName returns the provider name for identification.
func (g *GitHubAdapter) GetProviderName() string {
	return "github"
}

// ValidateOptions validates GitHub-specific options.
func (g *GitHubAdapter) ValidateOptions(options *CloneOptions) error {
	// First validate common options
	if err := g.ValidateCommonOptions(options); err != nil {
		return err
	}

	// GitHub-specific validations can be added here
	// For now, GitHub uses all common validations

	return nil
}

// CloneRepositories clones repositories from GitHub using the existing synclone implementation.
func (g *GitHubAdapter) CloneRepositories(ctx context.Context, request *CloneRequest) (*CloneResult, error) {
	if request == nil {
		return nil, fmt.Errorf("clone request cannot be nil")
	}

	// Validate the request
	if request.Organization == "" {
		return nil, fmt.Errorf("organization name is required")
	}
	if request.TargetPath == "" {
		return nil, fmt.Errorf("target path is required")
	}
	if request.Strategy == "" {
		request.Strategy = "reset" // default strategy
	}

	// Load config if specified
	if request.Options != nil && (request.Options.ConfigFile != "" || request.Options.UseConfig) {
		if err := g.LoadConfig(request.Options.ConfigFile); err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	// Get GitHub token
	token := ""
	if request.Options != nil && request.Options.Token != "" {
		token = request.Options.Token
	} else {
		token = env.GetToken("github")
	}

	// Create absolute path for target
	absTarget, err := filepath.Abs(request.TargetPath)
	if err != nil {
		return nil, fmt.Errorf("invalid target path: %w", err)
	}

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(absTarget, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	// Use the appropriate GitHub function based on options
	var cloneErr error
	if request.Options != nil {
		if request.Options.Parallel > 1 {
			// Use resumable parallel cloning
			cloneErr = github.RefreshAllResumable(
				ctx,
				absTarget,
				request.Organization,
				request.Strategy,
				request.Options.Parallel,
				request.Options.MaxRetries,
				request.Options.Resume,
				request.Options.ProgressMode,
			)
		} else if token != "" {
			// Use optimized streaming for authenticated requests
			cloneErr = github.RefreshAllOptimizedStreaming(
				ctx,
				absTarget,
				request.Organization,
				request.Strategy,
				token,
			)
		} else {
			// Use standard approach
			cloneErr = github.RefreshAll(ctx, absTarget, request.Organization, request.Strategy)
		}
	} else {
		// Use standard approach when no options provided
		cloneErr = github.RefreshAll(ctx, absTarget, request.Organization, request.Strategy)
	}

	// Convert result to our format
	result := &CloneResult{
		TotalRepositories: 0,                    // TODO: Get actual count from github package
		ClonesSuccessful:  0,                    // TODO: Get actual count from github package
		ClonesFailed:      0,                    // TODO: Get actual count from github package
		ClonesSkipped:     0,                    // TODO: Get actual count from github package
		Repositories:      []RepositoryResult{}, // TODO: Get actual results from github package
		Errors:            []error{},
	}

	if cloneErr != nil {
		result.Errors = append(result.Errors, cloneErr)
		return result, cloneErr
	}

	return result, nil
}

// ListRepositories lists repositories from GitHub without cloning them.
func (g *GitHubAdapter) ListRepositories(ctx context.Context, request *ListRequest) (*ListResult, error) {
	if request == nil {
		return nil, fmt.Errorf("list request cannot be nil")
	}

	if request.Organization == "" {
		return nil, fmt.Errorf("organization name is required")
	}

	// Get repository list from GitHub
	repos, err := github.ListRepos(ctx, request.Organization)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	// Convert github.RepoInfo to our RepositoryInfo format
	repositoryInfos := make([]RepositoryInfo, len(repos))
	for i, repo := range repos {
		repositoryInfos[i] = RepositoryInfo{
			Name:        repo.Name,
			FullName:    fmt.Sprintf("%s/%s", request.Organization, repo.Name),
			CloneURL:    repo.CloneURL,
			SSHURL:      convertHTTPSToSSH(repo.CloneURL),
			Description: repo.Description,
			Language:    "", // TODO: Add language field to github.RepoInfo if available
			Private:     repo.Private,
			Archived:    repo.Archived,
			Fork:        repo.Fork,
			Stars:       0,          // TODO: Add stars field to github.RepoInfo if available
			Topics:      []string{}, // TODO: Add topics field to github.RepoInfo if available
		}
	}

	// Apply filters if specified
	if request.Filters != nil {
		repositoryInfos = g.applyFilters(repositoryInfos, request.Filters)
	}

	return &ListResult{
		TotalRepositories: len(repositoryInfos),
		Repositories:      repositoryInfos,
	}, nil
}

// convertHTTPSToSSH converts an HTTPS clone URL to SSH format.
func convertHTTPSToSSH(httpsURL string) string {
	// Simple conversion from https://github.com/org/repo.git to git@github.com:org/repo.git
	if len(httpsURL) > 19 && httpsURL[:19] == "https://github.com/" {
		return "git@github.com:" + httpsURL[19:]
	}
	return httpsURL // Return original if conversion not applicable
}

// applyFilters applies repository filters to the list.
func (g *GitHubAdapter) applyFilters(repos []RepositoryInfo, filters *RepositoryFilters) []RepositoryInfo {
	if filters == nil {
		return repos
	}

	filtered := make([]RepositoryInfo, 0, len(repos))

	for _, repo := range repos {
		// Apply visibility filter
		if filters.Visibility != "" && filters.Visibility != "all" {
			if filters.Visibility == "public" && repo.Private {
				continue
			}
			if filters.Visibility == "private" && !repo.Private {
				continue
			}
		}

		// Apply archived filter
		if !filters.IncludeArchived && repo.Archived {
			continue
		}

		// Apply fork filter
		if !filters.IncludeForks && repo.Fork {
			continue
		}

		// Apply language filter
		if filters.Language != "" && repo.Language != filters.Language {
			continue
		}

		// Apply star filters
		if filters.MinStars > 0 && repo.Stars < filters.MinStars {
			continue
		}
		if filters.MaxStars > 0 && repo.Stars > filters.MaxStars {
			continue
		}

		// TODO: Apply name pattern filter using regex
		// TODO: Apply topics filter

		filtered = append(filtered, repo)
	}

	return filtered
}
