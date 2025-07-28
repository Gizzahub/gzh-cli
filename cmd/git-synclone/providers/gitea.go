// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package providers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gizzahub/gzh-manager-go/pkg/gitea"
)

// GiteaAdapter adapts the existing Gitea synclone functionality for Git extensions.
type GiteaAdapter struct {
	*BaseProviderAdapter
}

// NewGiteaAdapter creates a new Gitea provider adapter.
func NewGiteaAdapter() *GiteaAdapter {
	return &GiteaAdapter{
		BaseProviderAdapter: NewBaseProviderAdapter(),
	}
}

// GetProviderName returns the provider name for identification.
func (g *GiteaAdapter) GetProviderName() string {
	return "gitea"
}

// ValidateOptions validates Gitea-specific options.
func (g *GiteaAdapter) ValidateOptions(options *CloneOptions) error {
	// First validate common options
	if err := g.ValidateCommonOptions(options); err != nil {
		return err
	}

	// Gitea-specific validations can be added here
	// For now, Gitea uses all common validations

	return nil
}

// CloneRepositories clones repositories from Gitea using the existing synclone implementation.
func (g *GiteaAdapter) CloneRepositories(ctx context.Context, request *CloneRequest) (*CloneResult, error) {
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

	// Create absolute path for target
	absTarget, err := filepath.Abs(request.TargetPath)
	if err != nil {
		return nil, fmt.Errorf("invalid target path: %w", err)
	}

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(absTarget, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	// Use the Gitea package function (Gitea RefreshAll has 3 parameters)
	var cloneErr error
	cloneErr = gitea.RefreshAll(ctx, absTarget, request.Organization)

	// Convert result to our format
	result := &CloneResult{
		TotalRepositories: 0,                    // TODO: Get actual count from gitea package
		ClonesSuccessful:  0,                    // TODO: Get actual count from gitea package
		ClonesFailed:      0,                    // TODO: Get actual count from gitea package
		ClonesSkipped:     0,                    // TODO: Get actual count from gitea package
		Repositories:      []RepositoryResult{}, // TODO: Get actual results from gitea package
		Errors:            []error{},
	}

	if cloneErr != nil {
		result.Errors = append(result.Errors, cloneErr)
		return result, cloneErr
	}

	return result, nil
}

// ListRepositories lists repositories from Gitea without cloning them.
func (g *GiteaAdapter) ListRepositories(ctx context.Context, request *ListRequest) (*ListResult, error) {
	if request == nil {
		return nil, fmt.Errorf("list request cannot be nil")
	}

	if request.Organization == "" {
		return nil, fmt.Errorf("organization name is required")
	}

	// Get repository list from Gitea (List function returns []string)
	repoNames, err := gitea.List(ctx, request.Organization)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	// Convert []string to our RepositoryInfo format
	repositoryInfos := make([]RepositoryInfo, len(repoNames))
	for i, repoName := range repoNames {
		// Note: For custom Gitea instances, URL construction would need configuration
		repositoryInfos[i] = RepositoryInfo{
			Name:        repoName,
			FullName:    fmt.Sprintf("%s/%s", request.Organization, repoName),
			CloneURL:    fmt.Sprintf("https://gitea.example.com/%s/%s.git", request.Organization, repoName), // TODO: Make configurable
			SSHURL:      fmt.Sprintf("git@gitea.example.com:%s/%s.git", request.Organization, repoName),     // TODO: Make configurable
			Description: "",                                                                                 // Not available from List function
			Language:    "",                                                                                 // Not available from List function
			Private:     false,                                                                              // TODO: Determine visibility if needed
			Archived:    false,                                                                              // TODO: Determine if archived if needed
			Fork:        false,                                                                              // TODO: Determine if fork if needed
			Stars:       0,                                                                                  // Not available from List function
			Topics:      []string{},                                                                         // Not available from List function
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

// convertGiteaHTTPSToSSH converts a Gitea HTTPS clone URL to SSH format.
func convertGiteaHTTPSToSSH(httpsURL string) string {
	// Handle custom Gitea instances
	if len(httpsURL) > 8 && httpsURL[:8] == "https://" {
		// Extract domain and path
		withoutProtocol := httpsURL[8:]
		slashIndex := len(withoutProtocol)
		for i, c := range withoutProtocol {
			if c == '/' {
				slashIndex = i
				break
			}
		}
		if slashIndex < len(withoutProtocol) {
			domain := withoutProtocol[:slashIndex]
			path := withoutProtocol[slashIndex+1:]
			return fmt.Sprintf("git@%s:%s", domain, path)
		}
	}
	return httpsURL // Return original if conversion not applicable
}

// applyFilters applies repository filters to the list.
func (g *GiteaAdapter) applyFilters(repos []RepositoryInfo, filters *RepositoryFilters) []RepositoryInfo {
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
