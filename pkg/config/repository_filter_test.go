package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRepositoryFilterConfig(t *testing.T) {
	target := GitTarget{
		Name:       "test-org",
		Visibility: "public",
		Match:      "test-.*",
		Exclude:    []string{"test-exclude", "temp-*"},
	}

	config := NewRepositoryFilterConfig(target)

	assert.Equal(t, "public", config.Visibility)
	assert.Equal(t, "test-.*", config.Match)
	assert.Equal(t, []string{"test-exclude", "temp-*"}, config.Exclude)
}

func TestRepositoryFilterConfig_CreateRepositoryFilter(t *testing.T) {
	tests := []struct {
		name        string
		config      *RepositoryFilterConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: &RepositoryFilterConfig{
				Visibility: "public",
				Match:      "test-.*",
				Exclude:    []string{"exclude-*"},
			},
			expectError: false,
		},
		{
			name: "invalid visibility",
			config: &RepositoryFilterConfig{
				Visibility: "invalid",
				Match:      "",
				Exclude:    []string{},
			},
			expectError: true,
		},
		{
			name: "invalid regex pattern",
			config: &RepositoryFilterConfig{
				Visibility: "public",
				Match:      "[invalid",
				Exclude:    []string{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := tt.config.CreateRepositoryFilter()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, filter)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, filter)
			}
		})
	}
}

func TestNewRepositoryMatcher(t *testing.T) {
	config := &RepositoryFilterConfig{
		Visibility: "public",
		Match:      "test-.*",
		Exclude:    []string{"test-exclude"},
	}

	matcher, err := NewRepositoryMatcher(config)

	assert.NoError(t, err)
	assert.NotNil(t, matcher)
	assert.NotNil(t, matcher.filter)
	assert.Equal(t, config, matcher.config)
}

func TestRepositoryMatcher_ShouldCloneRepository(t *testing.T) {
	config := &RepositoryFilterConfig{
		Visibility: "public",
		Match:      "test-.*",
		Exclude:    []string{"test-exclude"},
	}

	matcher, err := NewRepositoryMatcher(config)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		repo     Repository
		expected bool
	}{
		{
			name: "public repo matching pattern",
			repo: Repository{
				Name:    "test-repo",
				Private: false,
			},
			expected: true,
		},
		{
			name: "private repo (filtered out)",
			repo: Repository{
				Name:    "test-private",
				Private: true,
			},
			expected: false,
		},
		{
			name: "public repo not matching pattern",
			repo: Repository{
				Name:    "prod-repo",
				Private: false,
			},
			expected: false,
		},
		{
			name: "excluded repo",
			repo: Repository{
				Name:    "test-exclude",
				Private: false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.ShouldCloneRepository(tt.repo)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRepositoryMatcher_FilterRepositoryList(t *testing.T) {
	repos := []Repository{
		{Name: "test-repo1", Private: false},
		{Name: "test-repo2", Private: true},
		{Name: "prod-repo1", Private: false},
		{Name: "test-exclude", Private: false},
	}

	config := &RepositoryFilterConfig{
		Visibility: "public",
		Match:      "test-.*",
		Exclude:    []string{"test-exclude"},
	}

	matcher, err := NewRepositoryMatcher(config)
	assert.NoError(t, err)

	filtered := matcher.FilterRepositoryList(repos)

	assert.Len(t, filtered, 1)
	assert.Equal(t, "test-repo1", filtered[0].Name)
}

func TestRepositoryMatcher_GetFilterSummary(t *testing.T) {
	config := &RepositoryFilterConfig{
		Visibility: "public",
		Match:      "test-.*",
		Exclude:    []string{"test-exclude"},
	}

	matcher, err := NewRepositoryMatcher(config)
	assert.NoError(t, err)

	summary := matcher.GetFilterSummary()
	expectedSummary := "visibility=public, pattern=test-.*, exclude=test-exclude"
	assert.Equal(t, expectedSummary, summary)
}

func TestRepositoryMatcher_GetStatistics(t *testing.T) {
	repos := []Repository{
		{Name: "test-repo1", Private: false},
		{Name: "test-repo2", Private: true},
		{Name: "prod-repo1", Private: false},
		{Name: "test-exclude", Private: false},
	}

	config := &RepositoryFilterConfig{
		Visibility: "public",
		Match:      "test-.*",
		Exclude:    []string{"test-exclude"},
	}

	matcher, err := NewRepositoryMatcher(config)
	assert.NoError(t, err)

	stats := matcher.GetStatistics(repos)

	assert.Equal(t, 4, stats.OriginalStats.TotalRepositories)
	assert.Equal(t, 1, stats.FilteredStats.TotalRepositories)
	assert.Equal(t, config, stats.FilterConfig)

	ratio := stats.GetFilteringRatio()
	assert.Equal(t, 25.0, ratio) // 1/4 = 25%

	summary := stats.GetFilteringSummary()
	expectedSummary := "Filtered 1/4 repositories (25.0% included, 3 excluded)"
	assert.Equal(t, expectedSummary, summary)
}

func TestFilteringStatistics_EmptyRepos(t *testing.T) {
	config := &RepositoryFilterConfig{
		Visibility: "all",
	}

	matcher, err := NewRepositoryMatcher(config)
	assert.NoError(t, err)

	stats := matcher.GetStatistics([]Repository{})

	assert.Equal(t, 0, stats.OriginalStats.TotalRepositories)
	assert.Equal(t, 0, stats.FilteredStats.TotalRepositories)
	assert.Equal(t, 0.0, stats.GetFilteringRatio())

	summary := stats.GetFilteringSummary()
	expectedSummary := "Filtered 0/0 repositories (0.0% included, 0 excluded)"
	assert.Equal(t, expectedSummary, summary)
}

func TestRepositoryNameMatcher_MatchesPattern(t *testing.T) {
	matcher := NewRepositoryNameMatcher()

	tests := []struct {
		name     string
		repoName string
		pattern  string
		expected bool
	}{
		{
			name:     "empty pattern matches all",
			repoName: "any-repo",
			pattern:  "",
			expected: true,
		},
		{
			name:     "glob pattern match",
			repoName: "test-repo",
			pattern:  "test-*",
			expected: true,
		},
		{
			name:     "glob pattern no match",
			repoName: "prod-repo",
			pattern:  "test-*",
			expected: false,
		},
		{
			name:     "regex pattern match",
			repoName: "test-repo-123",
			pattern:  "test-repo-[0-9]+",
			expected: true,
		},
		{
			name:     "regex pattern no match",
			repoName: "test-repo-abc",
			pattern:  "test-repo-[0-9]+",
			expected: false,
		},
		{
			name:     "literal string match",
			repoName: "my-test-repo",
			pattern:  "test",
			expected: true,
		},
		{
			name:     "literal string no match",
			repoName: "prod-repo",
			pattern:  "test",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.MatchesPattern(tt.repoName, tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRepositoryNameMatcher_MatchesExcludePatterns(t *testing.T) {
	matcher := NewRepositoryNameMatcher()

	tests := []struct {
		name            string
		repoName        string
		excludePatterns []string
		expected        bool
	}{
		{
			name:            "no exclude patterns",
			repoName:        "test-repo",
			excludePatterns: []string{},
			expected:        false,
		},
		{
			name:            "matches exclude pattern",
			repoName:        "test-exclude",
			excludePatterns: []string{"test-exclude", "other-*"},
			expected:        true,
		},
		{
			name:            "matches glob exclude pattern",
			repoName:        "temp-file",
			excludePatterns: []string{"temp-*", "test-exclude"},
			expected:        true,
		},
		{
			name:            "no match in exclude patterns",
			repoName:        "prod-repo",
			excludePatterns: []string{"temp-*", "test-exclude"},
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.MatchesExcludePatterns(tt.repoName, tt.excludePatterns)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateFilterConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *RepositoryFilterConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &RepositoryFilterConfig{
				Visibility: "public",
				Match:      "test-.*",
				Exclude:    []string{"temp-*", "test-exclude"},
			},
			expectError: false,
		},
		{
			name: "invalid visibility",
			config: &RepositoryFilterConfig{
				Visibility: "invalid",
			},
			expectError: true,
			errorMsg:    "invalid visibility",
		},
		{
			name: "invalid match pattern",
			config: &RepositoryFilterConfig{
				Visibility: "public",
				Match:      "[invalid",
			},
			expectError: true,
			errorMsg:    "invalid match pattern",
		},
		{
			name: "empty exclude pattern",
			config: &RepositoryFilterConfig{
				Visibility: "public",
				Exclude:    []string{"valid", "", "also-valid"},
			},
			expectError: true,
			errorMsg:    "exclude pattern 1 is empty",
		},
		{
			name: "minimal valid config",
			config: &RepositoryFilterConfig{
				Visibility: "",
				Match:      "",
				Exclude:    []string{},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilterConfig(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateRepositoryMatcherFromGitTarget(t *testing.T) {
	tests := []struct {
		name        string
		target      GitTarget
		expectError bool
	}{
		{
			name: "valid git target",
			target: GitTarget{
				Name:       "test-org",
				Visibility: "public",
				Match:      "test-.*",
				Exclude:    []string{"temp-*"},
			},
			expectError: false,
		},
		{
			name: "invalid visibility in git target",
			target: GitTarget{
				Name:       "test-org",
				Visibility: "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid match pattern in git target",
			target: GitTarget{
				Name:  "test-org",
				Match: "[invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher, err := CreateRepositoryMatcherFromGitTarget(tt.target)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, matcher)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, matcher)
			}
		})
	}
}

func TestRepositoryNameMatcher_ComplexPatterns(t *testing.T) {
	matcher := NewRepositoryNameMatcher()

	tests := []struct {
		name     string
		repoName string
		pattern  string
		expected bool
	}{
		{
			name:     "complex glob with multiple wildcards",
			repoName: "test-repo-123",
			pattern:  "test-*-*",
			expected: true,
		},
		{
			name:     "complex regex with groups",
			repoName: "api-v1-service",
			pattern:  "api-v[0-9]+-service",
			expected: true,
		},
		{
			name:     "case sensitive matching",
			repoName: "Test-Repo",
			pattern:  "test-*",
			expected: false,
		},
		{
			name:     "special characters in repo name",
			repoName: "repo.with.dots",
			pattern:  "repo.*dots",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.MatchesPattern(tt.repoName, tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}
