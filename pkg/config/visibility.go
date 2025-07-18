package config

import (
	"fmt"
	"strings"
)

// VisibilityFilter represents a repository visibility filter.
type VisibilityFilter struct {
	Filter string `json:"filter"` // public, private, all
}

// NewVisibilityFilter creates a new visibility filter.
func NewVisibilityFilter(filter string) (*VisibilityFilter, error) {
	filter = strings.ToLower(strings.TrimSpace(filter))

	// Validate filter value
	switch filter {
	case VisibilityPublic, VisibilityPrivate, VisibilityAll, "":
		if filter == "" {
			filter = VisibilityAll
		}

		return &VisibilityFilter{Filter: filter}, nil
	default:
		return nil, fmt.Errorf("invalid visibility filter '%s': must be one of %s, %s, %s",
			filter, VisibilityPublic, VisibilityPrivate, VisibilityAll)
	}
}

// ShouldIncludeRepository determines if a repository should be included based on visibility.
func (v *VisibilityFilter) ShouldIncludeRepository(isPrivate bool) bool {
	switch v.Filter {
	case VisibilityAll:
		return true
	case VisibilityPublic:
		return !isPrivate
	case VisibilityPrivate:
		return isPrivate
	default:
		// Default to all if filter is invalid
		return true
	}
}

// GetFilterString returns the filter string.
func (v *VisibilityFilter) GetFilterString() string {
	return v.Filter
}

// IsValidVisibility checks if a visibility string is valid.
func IsValidVisibility(visibility string) bool {
	switch strings.ToLower(strings.TrimSpace(visibility)) {
	case VisibilityPublic, VisibilityPrivate, VisibilityAll:
		return true
	default:
		return false
	}
}

// NormalizeVisibility normalizes a visibility string to standard format.
func NormalizeVisibility(visibility string) string {
	normalized := strings.ToLower(strings.TrimSpace(visibility))
	if normalized == "" {
		return VisibilityAll
	}

	return normalized
}

// VisibilityRepository represents a repository with visibility information.
type VisibilityRepository struct {
	Name      string `json:"name"`
	FullName  string `json:"full_name"`
	IsPrivate bool   `json:"is_private"`
	CloneURL  string `json:"clone_url"`
	SSHURL    string `json:"ssh_url"`
	HTTPURL   string `json:"http_url"`
}

// RepositoryFilter provides filtering capabilities for repositories.
type RepositoryFilter struct {
	VisibilityFilter *VisibilityFilter
	NamePattern      string // regex pattern for name filtering
	ExcludePatterns  []string
}

// NewRepositoryFilter creates a new repository filter.
func NewRepositoryFilter(visibility, namePattern string, excludePatterns []string) (*RepositoryFilter, error) {
	visFilter, err := NewVisibilityFilter(visibility)
	if err != nil {
		return nil, fmt.Errorf("invalid visibility filter: %w", err)
	}

	// Validate name pattern if provided
	if namePattern != "" {
		if _, err := CompileRegex(namePattern); err != nil {
			return nil, fmt.Errorf("invalid name pattern '%s': %w", namePattern, err)
		}
	}

	// Validate exclude patterns
	for _, pattern := range excludePatterns {
		if strings.Contains(pattern, "*") {
			// Simple glob pattern validation
			continue
		}
		// For regex patterns, try to compile
		if _, err := CompileRegex(pattern); err != nil {
			// Not a valid regex, treat as literal string
			continue
		}
	}

	return &RepositoryFilter{
		VisibilityFilter: visFilter,
		NamePattern:      namePattern,
		ExcludePatterns:  excludePatterns,
	}, nil
}

// ShouldIncludeRepository determines if a repository should be included.
func (f *RepositoryFilter) ShouldIncludeRepository(repo VisibilityRepository) bool {
	// Check visibility filter
	if !f.VisibilityFilter.ShouldIncludeRepository(repo.IsPrivate) {
		return false
	}

	// Check name pattern filter
	if f.NamePattern != "" {
		if regex, err := CompileRegex(f.NamePattern); err == nil {
			if !regex.MatchString(repo.Name) {
				return false
			}
		}
	}

	// Check exclude patterns
	for _, pattern := range f.ExcludePatterns {
		if f.matchesPattern(repo.Name, pattern) {
			return false
		}
	}

	return true
}

// matchesPattern checks if a name matches a pattern (glob or regex).
func (f *RepositoryFilter) matchesPattern(name, pattern string) bool {
	// Handle glob patterns
	if strings.Contains(pattern, "*") {
		return f.matchesGlob(name, pattern)
	}

	// Try regex pattern
	if regex, err := CompileRegex(pattern); err == nil {
		return regex.MatchString(name)
	}

	// Fallback to literal string match
	return strings.Contains(name, pattern)
}

// matchesGlob provides simple glob pattern matching.
func (f *RepositoryFilter) matchesGlob(name, pattern string) bool {
	// Simple glob implementation
	// Convert glob to regex
	regexPattern := strings.ReplaceAll(pattern, "*", ".*")
	regexPattern = "^" + regexPattern + "$"

	if regex, err := CompileRegex(regexPattern); err == nil {
		return regex.MatchString(name)
	}

	return false
}

// FilterRepositories filters a list of repositories based on the filter criteria.
func (f *RepositoryFilter) FilterRepositories(repos []VisibilityRepository) []VisibilityRepository {
	var filtered []VisibilityRepository

	for _, repo := range repos {
		if f.ShouldIncludeRepository(repo) {
			filtered = append(filtered, repo)
		}
	}

	return filtered
}

// GetFilterSummary returns a summary of the filter configuration.
func (f *RepositoryFilter) GetFilterSummary() string {
	var parts []string

	parts = append(parts, fmt.Sprintf("visibility=%s", f.VisibilityFilter.GetFilterString()))

	if f.NamePattern != "" {
		parts = append(parts, fmt.Sprintf("pattern=%s", f.NamePattern))
	}

	if len(f.ExcludePatterns) > 0 {
		parts = append(parts, fmt.Sprintf("exclude=%s", strings.Join(f.ExcludePatterns, ",")))
	}

	return strings.Join(parts, ", ")
}

// VisibilityStatistics provides statistics about repository visibility.
type VisibilityStatistics struct {
	TotalRepositories   int `json:"total_repositories"`
	PublicRepositories  int `json:"public_repositories"`
	PrivateRepositories int `json:"private_repositories"`
}

// CalculateVisibilityStatistics calculates visibility statistics for a list of repositories.
func CalculateVisibilityStatistics(repos []VisibilityRepository) VisibilityStatistics {
	stats := VisibilityStatistics{
		TotalRepositories: len(repos),
	}

	for _, repo := range repos {
		if repo.IsPrivate {
			stats.PrivateRepositories++
		} else {
			stats.PublicRepositories++
		}
	}

	return stats
}

// GetVisibilityPercentage returns the percentage of repositories by visibility.
func (v *VisibilityStatistics) GetVisibilityPercentage() (publicPercent, privatePercent float64) {
	if v.TotalRepositories == 0 {
		return 0, 0
	}

	publicPercent = float64(v.PublicRepositories) / float64(v.TotalRepositories) * 100
	privatePercent = float64(v.PrivateRepositories) / float64(v.TotalRepositories) * 100

	return publicPercent, privatePercent
}

// GetSummary returns a human-readable summary of the statistics.
func (v *VisibilityStatistics) GetSummary() string {
	if v.TotalRepositories == 0 {
		return "No repositories found"
	}

	publicPercent, privatePercent := v.GetVisibilityPercentage()

	return fmt.Sprintf("Total: %d repositories (%.1f%% public, %.1f%% private)",
		v.TotalRepositories, publicPercent, privatePercent)
}
