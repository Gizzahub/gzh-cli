//nolint:testpackage // White-box testing needed for internal function access
package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVisibilityFilter(t *testing.T) {
	tests := []struct {
		name        string
		filter      string
		expected    string
		expectError bool
	}{
		{
			name:        "public filter",
			filter:      "public",
			expected:    VisibilityPublic,
			expectError: false,
		},
		{
			name:        "private filter",
			filter:      "private",
			expected:    VisibilityPrivate,
			expectError: false,
		},
		{
			name:        "all filter",
			filter:      "all",
			expected:    VisibilityAll,
			expectError: false,
		},
		{
			name:        "empty filter defaults to all",
			filter:      "",
			expected:    VisibilityAll,
			expectError: false,
		},
		{
			name:        "uppercase filter",
			filter:      "PUBLIC",
			expected:    VisibilityPublic,
			expectError: false,
		},
		{
			name:        "whitespace filter",
			filter:      "  private  ",
			expected:    VisibilityPrivate,
			expectError: false,
		},
		{
			name:        "invalid filter",
			filter:      "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := NewVisibilityFilter(tt.filter)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, filter)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, filter)
				assert.Equal(t, tt.expected, filter.Filter)
			}
		})
	}
}

func TestVisibilityFilter_ShouldIncludeRepository(t *testing.T) {
	tests := []struct {
		name      string
		filter    string
		isPrivate bool
		expected  bool
	}{
		{
			name:      "all filter includes public repos",
			filter:    VisibilityAll,
			isPrivate: false,
			expected:  true,
		},
		{
			name:      "all filter includes private repos",
			filter:    VisibilityAll,
			isPrivate: true,
			expected:  true,
		},
		{
			name:      "public filter includes public repos",
			filter:    VisibilityPublic,
			isPrivate: false,
			expected:  true,
		},
		{
			name:      "public filter excludes private repos",
			filter:    VisibilityPublic,
			isPrivate: true,
			expected:  false,
		},
		{
			name:      "private filter excludes public repos",
			filter:    VisibilityPrivate,
			isPrivate: false,
			expected:  false,
		},
		{
			name:      "private filter includes private repos",
			filter:    VisibilityPrivate,
			isPrivate: true,
			expected:  true,
		},
		{
			name:      "invalid filter defaults to include",
			filter:    "invalid",
			isPrivate: true,
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &VisibilityFilter{Filter: tt.filter}
			result := filter.ShouldIncludeRepository(tt.isPrivate)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidVisibility(t *testing.T) {
	tests := []struct {
		name       string
		visibility string
		expected   bool
	}{
		{"valid public", "public", true},
		{"valid private", "private", true},
		{"valid all", "all", true},
		{"uppercase valid", "PUBLIC", true},
		{"whitespace valid", "  private  ", true},
		{"invalid", "invalid", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidVisibility(tt.visibility)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeVisibility(t *testing.T) {
	tests := []struct {
		name       string
		visibility string
		expected   string
	}{
		{"public", "public", "public"},
		{"uppercase", "PUBLIC", "public"},
		{"whitespace", "  private  ", "private"},
		{"empty defaults to all", "", "all"},
		{"mixed case", "PrIvAtE", "private"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeVisibility(tt.visibility)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewRepositoryFilter(t *testing.T) {
	tests := []struct {
		name            string
		visibility      string
		namePattern     string
		excludePatterns []string
		expectError     bool
	}{
		{
			name:            "valid filter with all parameters",
			visibility:      "public",
			namePattern:     "test.*",
			excludePatterns: []string{"exclude-*", "test-exclude"},
			expectError:     false,
		},
		{
			name:        "valid filter with minimal parameters",
			visibility:  "all",
			expectError: false,
		},
		{
			name:        "invalid visibility",
			visibility:  "invalid",
			expectError: true,
		},
		{
			name:        "invalid name pattern regex",
			visibility:  "public",
			namePattern: "[invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := NewRepositoryFilter(tt.visibility, tt.namePattern, tt.excludePatterns)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, filter)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, filter)
				assert.Equal(t, tt.visibility, filter.VisibilityFilter.Filter)
				assert.Equal(t, tt.namePattern, filter.NamePattern)
				assert.Equal(t, tt.excludePatterns, filter.ExcludePatterns)
			}
		})
	}
}

func TestRepositoryFilter_ShouldIncludeRepository(t *testing.T) {
	tests := []struct {
		name            string
		visibility      string
		namePattern     string
		excludePatterns []string
		repo            VisibilityRepository
		expected        bool
	}{
		{
			name:       "public repo passes public filter",
			visibility: "public",
			repo:       VisibilityRepository{Name: "test-repo", IsPrivate: false},
			expected:   true,
		},
		{
			name:       "private repo fails public filter",
			visibility: "public",
			repo:       VisibilityRepository{Name: "test-repo", IsPrivate: true},
			expected:   false,
		},
		{
			name:        "repo matches name pattern",
			visibility:  "all",
			namePattern: "test-.*",
			repo:        VisibilityRepository{Name: "test-repo", IsPrivate: false},
			expected:    true,
		},
		{
			name:        "repo doesn't match name pattern",
			visibility:  "all",
			namePattern: "prod-.*",
			repo:        VisibilityRepository{Name: "test-repo", IsPrivate: false},
			expected:    false,
		},
		{
			name:            "repo excluded by exact match",
			visibility:      "all",
			excludePatterns: []string{"test-repo", "other-repo"},
			repo:            VisibilityRepository{Name: "test-repo", IsPrivate: false},
			expected:        false,
		},
		{
			name:            "repo excluded by glob pattern",
			visibility:      "all",
			excludePatterns: []string{"test-*"},
			repo:            VisibilityRepository{Name: "test-repo", IsPrivate: false},
			expected:        false,
		},
		{
			name:            "repo not excluded",
			visibility:      "all",
			excludePatterns: []string{"other-*"},
			repo:            VisibilityRepository{Name: "test-repo", IsPrivate: false},
			expected:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := NewRepositoryFilter(tt.visibility, tt.namePattern, tt.excludePatterns)
			assert.NoError(t, err)

			result := filter.ShouldIncludeRepository(tt.repo)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRepositoryFilter_FilterRepositories(t *testing.T) {
	repos := []VisibilityRepository{
		{Name: "public-repo", IsPrivate: false},
		{Name: "private-repo", IsPrivate: true},
		{Name: "test-public", IsPrivate: false},
		{Name: "test-private", IsPrivate: true},
		{Name: "excluded-repo", IsPrivate: false},
	}

	tests := []struct {
		name            string
		visibility      string
		namePattern     string
		excludePatterns []string
		expectedNames   []string
	}{
		{
			name:          "filter only public repos",
			visibility:    "public",
			expectedNames: []string{"public-repo", "test-public", "excluded-repo"},
		},
		{
			name:          "filter only private repos",
			visibility:    "private",
			expectedNames: []string{"private-repo", "test-private"},
		},
		{
			name:          "filter by name pattern",
			visibility:    "all",
			namePattern:   "test-.*",
			expectedNames: []string{"test-public", "test-private"},
		},
		{
			name:            "exclude specific repos",
			visibility:      "all",
			excludePatterns: []string{"excluded-*"},
			expectedNames:   []string{"public-repo", "private-repo", "test-public", "test-private"},
		},
		{
			name:            "complex filter",
			visibility:      "public",
			namePattern:     ".*-.*",
			excludePatterns: []string{"excluded-*"},
			expectedNames:   []string{"public-repo", "test-public"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := NewRepositoryFilter(tt.visibility, tt.namePattern, tt.excludePatterns)
			assert.NoError(t, err)

			filtered := filter.FilterRepositories(repos)

			var actualNames []string
			for _, repo := range filtered {
				actualNames = append(actualNames, repo.Name)
			}

			assert.ElementsMatch(t, tt.expectedNames, actualNames)
		})
	}
}

func TestRepositoryFilter_GetFilterSummary(t *testing.T) {
	tests := []struct {
		name            string
		visibility      string
		namePattern     string
		excludePatterns []string
		expected        string
	}{
		{
			name:       "visibility only",
			visibility: "public",
			expected:   "visibility=public",
		},
		{
			name:        "visibility and pattern",
			visibility:  "all",
			namePattern: "test-.*",
			expected:    "visibility=all, pattern=test-.*",
		},
		{
			name:            "all filters",
			visibility:      "private",
			namePattern:     "prod-.*",
			excludePatterns: []string{"exclude-*", "test-*"},
			expected:        "visibility=private, pattern=prod-.*, exclude=exclude-*,test-*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := NewRepositoryFilter(tt.visibility, tt.namePattern, tt.excludePatterns)
			assert.NoError(t, err)

			summary := filter.GetFilterSummary()
			assert.Equal(t, tt.expected, summary)
		})
	}
}

func TestCalculateVisibilityStatistics(t *testing.T) {
	repos := []VisibilityRepository{
		{Name: "public1", IsPrivate: false},
		{Name: "public2", IsPrivate: false},
		{Name: "private1", IsPrivate: true},
		{Name: "private2", IsPrivate: true},
		{Name: "private3", IsPrivate: true},
	}

	stats := CalculateVisibilityStatistics(repos)

	assert.Equal(t, 5, stats.TotalRepositories)
	assert.Equal(t, 2, stats.PublicRepositories)
	assert.Equal(t, 3, stats.PrivateRepositories)

	publicPercent, privatePercent := stats.GetVisibilityPercentage()
	assert.Equal(t, 40.0, publicPercent)
	assert.Equal(t, 60.0, privatePercent)

	summary := stats.GetSummary()
	expected := "Total: 5 repositories (40.0% public, 60.0% private)"
	assert.Equal(t, expected, summary)
}

func TestVisibilityStatistics_EmptyRepos(t *testing.T) {
	stats := CalculateVisibilityStatistics([]VisibilityRepository{})

	assert.Equal(t, 0, stats.TotalRepositories)
	assert.Equal(t, 0, stats.PublicRepositories)
	assert.Equal(t, 0, stats.PrivateRepositories)

	publicPercent, privatePercent := stats.GetVisibilityPercentage()
	assert.Equal(t, 0.0, publicPercent)
	assert.Equal(t, 0.0, privatePercent)

	summary := stats.GetSummary()
	assert.Equal(t, "No repositories found", summary)
}

func TestCompileRegex(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		expectError bool
	}{
		{
			name:        "valid regex",
			pattern:     "test-.*",
			expectError: false,
		},
		{
			name:        "complex valid regex",
			pattern:     "^[a-zA-Z0-9_-]+$",
			expectError: false,
		},
		{
			name:        "invalid regex",
			pattern:     "[invalid",
			expectError: true,
		},
		{
			name:        "empty pattern",
			pattern:     "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex, err := CompileRegex(tt.pattern)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, regex)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, regex)
			}
		})
	}
}

func TestRepositoryFilter_MatchesPattern(t *testing.T) {
	filter, err := NewRepositoryFilter("all", "", []string{})
	assert.NoError(t, err)

	tests := []struct {
		name     string
		repoName string
		pattern  string
		expected bool
	}{
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
			repoName: "test-repo",
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
			result := filter.matchesPattern(tt.repoName, tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}
