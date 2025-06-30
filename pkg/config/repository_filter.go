package config

import (
	"fmt"
	"strings"
)

// RepositoryFilterConfig represents repository filtering configuration from gzh.yaml
type RepositoryFilterConfig struct {
	Visibility string   `json:"visibility"` // public, private, all
	Match      string   `json:"match"`      // regex pattern for repository names
	Exclude    []string `json:"exclude"`    // patterns to exclude repositories
}

// NewRepositoryFilterConfig creates a repository filter config from GitTarget
func NewRepositoryFilterConfig(target GitTarget) *RepositoryFilterConfig {
	return &RepositoryFilterConfig{
		Visibility: target.Visibility,
		Match:      target.Match,
		Exclude:    target.Exclude,
	}
}

// CreateRepositoryFilter creates a RepositoryFilter from config
func (r *RepositoryFilterConfig) CreateRepositoryFilter() (*RepositoryFilter, error) {
	return NewRepositoryFilter(r.Visibility, r.Match, r.Exclude)
}

// RepositoryMatcher provides high-level repository matching functionality
type RepositoryMatcher struct {
	filter *RepositoryFilter
	config *RepositoryFilterConfig
}

// NewRepositoryMatcher creates a new repository matcher from config
func NewRepositoryMatcher(config *RepositoryFilterConfig) (*RepositoryMatcher, error) {
	filter, err := config.CreateRepositoryFilter()
	if err != nil {
		return nil, fmt.Errorf("failed to create repository filter: %w", err)
	}

	return &RepositoryMatcher{
		filter: filter,
		config: config,
	}, nil
}

// ShouldCloneRepository determines if a repository should be cloned based on configuration
func (m *RepositoryMatcher) ShouldCloneRepository(repo Repository) bool {
	return m.filter.ShouldIncludeRepository(repo)
}

// FilterRepositoryList filters a list of repositories based on configuration
func (m *RepositoryMatcher) FilterRepositoryList(repos []Repository) []Repository {
	return m.filter.FilterRepositories(repos)
}

// GetFilterSummary returns a summary of the filtering configuration
func (m *RepositoryMatcher) GetFilterSummary() string {
	return m.filter.GetFilterSummary()
}

// GetStatistics returns filtering statistics for a repository list
func (m *RepositoryMatcher) GetStatistics(repos []Repository) *FilteringStatistics {
	originalStats := CalculateVisibilityStatistics(repos)
	filteredRepos := m.FilterRepositoryList(repos)
	filteredStats := CalculateVisibilityStatistics(filteredRepos)

	return &FilteringStatistics{
		OriginalStats: originalStats,
		FilteredStats: filteredStats,
		FilterConfig:  m.config,
	}
}

// FilteringStatistics provides statistics about repository filtering results
type FilteringStatistics struct {
	OriginalStats VisibilityStatistics    `json:"original_stats"`
	FilteredStats VisibilityStatistics    `json:"filtered_stats"`
	FilterConfig  *RepositoryFilterConfig `json:"filter_config"`
}

// GetFilteringRatio returns the percentage of repositories that passed filtering
func (f *FilteringStatistics) GetFilteringRatio() float64 {
	if f.OriginalStats.TotalRepositories == 0 {
		return 0
	}
	return float64(f.FilteredStats.TotalRepositories) / float64(f.OriginalStats.TotalRepositories) * 100
}

// GetFilteringSummary returns a human-readable summary of filtering results
func (f *FilteringStatistics) GetFilteringSummary() string {
	ratio := f.GetFilteringRatio()
	excluded := f.OriginalStats.TotalRepositories - f.FilteredStats.TotalRepositories

	return fmt.Sprintf("Filtered %d/%d repositories (%.1f%% included, %d excluded)",
		f.FilteredStats.TotalRepositories, f.OriginalStats.TotalRepositories, ratio, excluded)
}

// RepositoryNameMatcher provides simple repository name matching utilities
type RepositoryNameMatcher struct{}

// NewRepositoryNameMatcher creates a new repository name matcher
func NewRepositoryNameMatcher() *RepositoryNameMatcher {
	return &RepositoryNameMatcher{}
}

// MatchesPattern checks if a repository name matches a pattern (glob or regex)
func (m *RepositoryNameMatcher) MatchesPattern(repoName, pattern string) bool {
	if pattern == "" {
		return true // No pattern means match all
	}

	// Handle glob patterns
	if strings.Contains(pattern, "*") {
		return m.matchesGlob(repoName, pattern)
	}

	// Try regex pattern
	if regex, err := CompileRegex(pattern); err == nil {
		return regex.MatchString(repoName)
	}

	// Fallback to literal string match
	return strings.Contains(repoName, pattern)
}

// MatchesExcludePatterns checks if a repository name matches any exclude pattern
func (m *RepositoryNameMatcher) MatchesExcludePatterns(repoName string, excludePatterns []string) bool {
	for _, pattern := range excludePatterns {
		if m.MatchesPattern(repoName, pattern) {
			return true // Found a match in exclude patterns
		}
	}
	return false // No exclude patterns matched
}

// matchesGlob provides simple glob pattern matching
func (m *RepositoryNameMatcher) matchesGlob(name, pattern string) bool {
	// Simple glob implementation
	// Convert glob to regex
	regexPattern := strings.ReplaceAll(pattern, "*", ".*")
	regexPattern = "^" + regexPattern + "$"

	if regex, err := CompileRegex(regexPattern); err == nil {
		return regex.MatchString(name)
	}

	return false
}

// ValidateFilterConfig validates a repository filter configuration
func ValidateFilterConfig(config *RepositoryFilterConfig) error {
	// Validate visibility
	if config.Visibility != "" && !IsValidVisibility(config.Visibility) {
		return fmt.Errorf("invalid visibility '%s'", config.Visibility)
	}

	// Validate regex pattern
	if config.Match != "" {
		if _, err := CompileRegex(config.Match); err != nil {
			return fmt.Errorf("invalid match pattern '%s': %w", config.Match, err)
		}
	}

	// Validate exclude patterns (basic validation)
	for i, pattern := range config.Exclude {
		if pattern == "" {
			return fmt.Errorf("exclude pattern %d is empty", i)
		}
		// For complex patterns, try to compile as regex
		if !strings.Contains(pattern, "*") {
			if _, err := CompileRegex(pattern); err != nil {
				// Not a valid regex, but that's okay - it might be a literal string
				continue
			}
		}
	}

	return nil
}

// CreateRepositoryMatcherFromGitTarget creates a repository matcher from GitTarget
func CreateRepositoryMatcherFromGitTarget(target GitTarget) (*RepositoryMatcher, error) {
	config := NewRepositoryFilterConfig(target)
	if err := ValidateFilterConfig(config); err != nil {
		return nil, fmt.Errorf("invalid filter config: %w", err)
	}
	return NewRepositoryMatcher(config)
}
