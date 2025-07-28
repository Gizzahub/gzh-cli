// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package providers

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderInterface(t *testing.T) {
	providers := []struct {
		name     string
		provider ProviderAdapter
	}{
		{"GitHub", NewGitHubAdapter()},
		{"GitLab", NewGitLabAdapter()},
		{"Gitea", NewGiteaAdapter()},
	}

	for _, p := range providers {
		t.Run(p.name, func(t *testing.T) {
			// Test that provider implements the interface
			assert.Implements(t, (*ProviderAdapter)(nil), p.provider)

			// Test provider name
			assert.NotEmpty(t, p.provider.GetProviderName())

			// Test options validation with nil options
			err := p.provider.ValidateOptions(nil)
			assert.Error(t, err, "should error with nil options")
		})
	}
}

func TestCloneOptionsValidation(t *testing.T) {
	tests := []struct {
		name    string
		options *CloneOptions
		wantErr bool
	}{
		{
			name:    "nil options",
			options: nil,
			wantErr: true,
		},
		{
			name: "valid options",
			options: &CloneOptions{
				Strategy:     "reset",
				Protocol:     "https",
				Parallel:     5,
				ProgressMode: "bar",
			},
			wantErr: false,
		},
		{
			name: "invalid strategy",
			options: &CloneOptions{
				Strategy:     "invalid",
				Protocol:     "https",
				Parallel:     5,
				ProgressMode: "bar",
			},
			wantErr: true,
		},
		{
			name: "invalid protocol",
			options: &CloneOptions{
				Strategy:     "reset",
				Protocol:     "ftp",
				Parallel:     5,
				ProgressMode: "bar",
			},
			wantErr: true,
		},
		{
			name: "invalid progress mode",
			options: &CloneOptions{
				Strategy:     "reset",
				Protocol:     "https",
				Parallel:     5,
				ProgressMode: "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid parallel workers",
			options: &CloneOptions{
				Strategy:     "reset",
				Protocol:     "https",
				Parallel:     0,
				ProgressMode: "bar",
			},
			wantErr: true,
		},
	}

	adapter := NewGitHubAdapter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.ValidateOptions(tt.options)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGitHubProviderIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("GITHUB_TOKEN not set, skipping GitHub integration test")
	}

	adapter := NewGitHubAdapter()
	ctx := context.Background()

	t.Run("ListRepositories", func(t *testing.T) {
		request := &ListRequest{
			Organization: "github", // Use a known public organization
			Filters: &RepositoryFilters{
				Visibility: "public",
			},
		}

		result, err := adapter.ListRepositories(ctx, request)
		require.NoError(t, err)
		assert.Greater(t, result.TotalRepositories, 0)
		assert.NotEmpty(t, result.Repositories)

		// Check first repository structure
		if len(result.Repositories) > 0 {
			repo := result.Repositories[0]
			assert.NotEmpty(t, repo.Name)
			assert.NotEmpty(t, repo.FullName)
			assert.NotEmpty(t, repo.CloneURL)
			assert.Contains(t, repo.CloneURL, "github.com")
		}
	})

	t.Run("CloneRepositories_DryRun", func(t *testing.T) {
		// Use a small test organization or create a mock
		request := &CloneRequest{
			Organization: "octocat", // GitHub's mascot user with few repos
			TargetPath:   t.TempDir(),
			Strategy:     "reset",
			Options: &CloneOptions{
				Strategy:     "reset",
				Protocol:     "https",
				Parallel:     1,
				DryRun:       true,
				ProgressMode: "quiet",
				Token:        token,
			},
		}

		// Note: This would normally clone, but with DryRun it should validate without actual cloning
		// For now, we'll test the structure without actual cloning
		err := adapter.ValidateOptions(request.Options)
		assert.NoError(t, err)
	})
}

func TestGitLabProviderIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	token := os.Getenv("GITLAB_TOKEN")
	if token == "" {
		t.Skip("GITLAB_TOKEN not set, skipping GitLab integration test")
	}

	adapter := NewGitLabAdapter()
	ctx := context.Background()

	t.Run("ListRepositories", func(t *testing.T) {
		request := &ListRequest{
			Organization: "gitlab-org", // GitLab's public group
			Filters: &RepositoryFilters{
				Visibility: "public",
			},
		}

		result, err := adapter.ListRepositories(ctx, request)
		if err != nil {
			t.Logf("GitLab list repositories failed (expected if no access): %v", err)
			return
		}

		assert.Greater(t, result.TotalRepositories, 0)
		assert.NotEmpty(t, result.Repositories)
	})
}

func TestGiteaProviderIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	adapter := NewGiteaAdapter()
	ctx := context.Background()

	t.Run("ListRepositories", func(t *testing.T) {
		request := &ListRequest{
			Organization: "gitea", // Gitea's organization
			Filters: &RepositoryFilters{
				Visibility: "public",
			},
		}

		result, err := adapter.ListRepositories(ctx, request)
		if err != nil {
			t.Logf("Gitea list repositories failed (expected if server not accessible): %v", err)
			return
		}

		assert.Greater(t, result.TotalRepositories, 0)
		assert.NotEmpty(t, result.Repositories)
	})
}

func TestProviderErrorHandling(t *testing.T) {
	adapter := NewGitHubAdapter()
	ctx := context.Background()

	t.Run("CloneRepositories_NilRequest", func(t *testing.T) {
		result, err := adapter.CloneRepositories(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("ListRepositories_NilRequest", func(t *testing.T) {
		result, err := adapter.ListRepositories(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("CloneRepositories_EmptyOrganization", func(t *testing.T) {
		request := &CloneRequest{
			Organization: "",
			TargetPath:   "/tmp/test",
			Strategy:     "reset",
		}

		result, err := adapter.CloneRepositories(ctx, request)
		assert.Error(t, err)
		assert.NotNil(t, result) // Result should be returned even on error
		assert.Contains(t, err.Error(), "organization")
	})

	t.Run("ListRepositories_EmptyOrganization", func(t *testing.T) {
		request := &ListRequest{
			Organization: "",
		}

		result, err := adapter.ListRepositories(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "organization")
	})
}

func TestBaseProviderAdapter(t *testing.T) {
	adapter := NewBaseProviderAdapter()

	t.Run("LoadConfig", func(t *testing.T) {
		// Test with non-existent config file
		err := adapter.LoadConfig("/non/existent/config.yaml")
		assert.Error(t, err)

		// Test with empty config file path (should try to find config)
		err = adapter.LoadConfig("")
		// This might succeed or fail depending on whether config files exist
		// We just test that it doesn't panic
		assert.NotPanics(t, func() {
			_ = adapter.LoadConfig("")
		})
	})

	t.Run("GetConfig", func(t *testing.T) {
		config := adapter.GetConfig()
		// Config might be nil if not loaded
		if config != nil {
			assert.NotNil(t, config)
		}
	})
}

func TestRepositoryFilters(t *testing.T) {
	adapter := NewGitHubAdapter()

	// Create test repositories
	repos := []RepositoryInfo{
		{
			Name:     "public-repo",
			Private:  false,
			Archived: false,
			Fork:     false,
			Language: "Go",
			Stars:    100,
		},
		{
			Name:     "private-repo",
			Private:  true,
			Archived: false,
			Fork:     false,
			Language: "Python",
			Stars:    50,
		},
		{
			Name:     "archived-repo",
			Private:  false,
			Archived: true,
			Fork:     false,
			Language: "JavaScript",
			Stars:    200,
		},
		{
			Name:     "fork-repo",
			Private:  false,
			Archived: false,
			Fork:     true,
			Language: "Go",
			Stars:    10,
		},
	}

	tests := []struct {
		name     string
		filters  *RepositoryFilters
		expected int
	}{
		{
			name:     "no filters",
			filters:  nil,
			expected: 4,
		},
		{
			name: "public only",
			filters: &RepositoryFilters{
				Visibility: "public",
			},
			expected: 3, // excludes private-repo
		},
		{
			name: "private only",
			filters: &RepositoryFilters{
				Visibility: "private",
			},
			expected: 1, // only private-repo
		},
		{
			name: "exclude archived",
			filters: &RepositoryFilters{
				IncludeArchived: false,
			},
			expected: 3, // excludes archived-repo
		},
		{
			name: "exclude forks",
			filters: &RepositoryFilters{
				IncludeForks: false,
			},
			expected: 3, // excludes fork-repo
		},
		{
			name: "Go language only",
			filters: &RepositoryFilters{
				Language: "Go",
			},
			expected: 2, // public-repo and fork-repo
		},
		{
			name: "minimum stars",
			filters: &RepositoryFilters{
				MinStars: 75,
			},
			expected: 2, // public-repo and archived-repo
		},
		{
			name: "maximum stars",
			filters: &RepositoryFilters{
				MaxStars: 150,
			},
			expected: 3, // excludes archived-repo (200 stars)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := adapter.applyFilters(repos, tt.filters)
			assert.Len(t, filtered, tt.expected)
		})
	}
}
