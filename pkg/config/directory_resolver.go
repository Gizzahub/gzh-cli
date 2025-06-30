package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

// DirectoryResolver handles directory structure resolution for repository cloning
type DirectoryResolver struct {
	target BulkCloneTarget
}

// NewDirectoryResolver creates a new directory resolver
func NewDirectoryResolver(target BulkCloneTarget) *DirectoryResolver {
	return &DirectoryResolver{
		target: target,
	}
}

// ResolveRepositoryPath resolves the full path for a specific repository
func (d *DirectoryResolver) ResolveRepositoryPath(repositoryName string) string {
	basePath := d.target.CloneDir

	if d.target.Flatten {
		// Flatten structure: base/{repository}
		return filepath.Join(basePath, repositoryName)
	} else {
		// Normal structure: base/{organization}/{repository}
		return filepath.Join(basePath, d.target.Name, repositoryName)
	}
}

// GetBasePath returns the base directory path for the target
func (d *DirectoryResolver) GetBasePath() string {
	return d.target.CloneDir
}

// GetOrganizationPath returns the organization/group directory path
func (d *DirectoryResolver) GetOrganizationPath() string {
	if d.target.Flatten {
		// For flattened structure, org path is the same as base path
		return d.target.CloneDir
	} else {
		// For normal structure, org path includes the organization name
		return filepath.Join(d.target.CloneDir, d.target.Name)
	}
}

// GetDirectoryStructure returns a description of the directory structure
func (d *DirectoryResolver) GetDirectoryStructure() DirectoryStructure {
	return DirectoryStructure{
		BasePath:         d.GetBasePath(),
		OrganizationPath: d.GetOrganizationPath(),
		IsFlattened:      d.target.Flatten,
		Provider:         d.target.Provider,
		TargetName:       d.target.Name,
	}
}

// DirectoryStructure represents the resolved directory structure
type DirectoryStructure struct {
	BasePath         string `json:"base_path"`
	OrganizationPath string `json:"organization_path"`
	IsFlattened      bool   `json:"is_flattened"`
	Provider         string `json:"provider"`
	TargetName       string `json:"target_name"`
}

// GetDescription returns a human-readable description of the structure
func (d *DirectoryStructure) GetDescription() string {
	if d.IsFlattened {
		return "Flattened structure: all repositories in " + d.OrganizationPath
	} else {
		return "Nested structure: repositories in " + d.OrganizationPath
	}
}

// GetExamplePath returns an example repository path
func (d *DirectoryStructure) GetExamplePath(repoName string) string {
	if d.IsFlattened {
		return filepath.Join(d.OrganizationPath, repoName)
	} else {
		return filepath.Join(d.OrganizationPath, repoName)
	}
}

// DirectoryPathGenerator provides utilities for generating directory paths
type DirectoryPathGenerator struct{}

// NewDirectoryPathGenerator creates a new directory path generator
func NewDirectoryPathGenerator() *DirectoryPathGenerator {
	return &DirectoryPathGenerator{}
}

// GenerateRepositoryPaths generates repository paths for a list of repositories
func (g *DirectoryPathGenerator) GenerateRepositoryPaths(target BulkCloneTarget, repositories []Repository) []RepositoryPath {
	resolver := NewDirectoryResolver(target)
	var paths []RepositoryPath

	for _, repo := range repositories {
		path := RepositoryPath{
			Repository:   repo,
			FullPath:     resolver.ResolveRepositoryPath(repo.Name),
			RelativePath: g.getRelativePath(target, repo.Name),
		}
		paths = append(paths, path)
	}

	return paths
}

// getRelativePath returns the relative path within the clone directory
func (g *DirectoryPathGenerator) getRelativePath(target BulkCloneTarget, repoName string) string {
	if target.Flatten {
		return repoName
	} else {
		return filepath.Join(target.Name, repoName)
	}
}

// RepositoryPath represents the resolved path for a repository
type RepositoryPath struct {
	Repository   Repository `json:"repository"`
	FullPath     string     `json:"full_path"`
	RelativePath string     `json:"relative_path"`
}

// GetParentDirectory returns the parent directory of the repository
func (r *RepositoryPath) GetParentDirectory() string {
	return filepath.Dir(r.FullPath)
}

// IsValid checks if the repository path is valid
func (r *RepositoryPath) IsValid() bool {
	return r.FullPath != "" && r.Repository.Name != ""
}

// DirectoryStructureValidator validates directory structures
type DirectoryStructureValidator struct{}

// NewDirectoryStructureValidator creates a new directory structure validator
func NewDirectoryStructureValidator() *DirectoryStructureValidator {
	return &DirectoryStructureValidator{}
}

// ValidateStructure validates a directory structure configuration
func (v *DirectoryStructureValidator) ValidateStructure(target BulkCloneTarget) error {
	// Basic validation
	if target.CloneDir == "" {
		return ErrInvalidCloneDir
	}

	if target.Name == "" {
		return ErrMissingName
	}

	// Check for problematic characters in paths
	if strings.Contains(target.CloneDir, "..") {
		return ErrUnsafePath
	}

	return nil
}

// ValidateRepositoryPaths validates a list of repository paths
func (v *DirectoryStructureValidator) ValidateRepositoryPaths(paths []RepositoryPath) []ValidationIssue {
	var issues []ValidationIssue
	pathMap := make(map[string]string) // path -> repository name

	for _, path := range paths {
		// Check for path conflicts
		if existingRepo, exists := pathMap[path.FullPath]; exists {
			issues = append(issues, ValidationIssue{
				Type:        "path_conflict",
				Severity:    "error",
				Description: "Path conflict between repositories: " + path.Repository.Name + " and " + existingRepo,
				Path:        path.FullPath,
			})
		} else {
			pathMap[path.FullPath] = path.Repository.Name
		}

		// Check for invalid paths
		if !path.IsValid() {
			issues = append(issues, ValidationIssue{
				Type:        "invalid_path",
				Severity:    "error",
				Description: "Invalid repository path for: " + path.Repository.Name,
				Path:        path.FullPath,
			})
		}
	}

	return issues
}

// ValidationIssue represents a directory structure validation issue
type ValidationIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Path        string `json:"path"`
}

// DirectoryStructureAnalyzer analyzes directory structures
type DirectoryStructureAnalyzer struct{}

// NewDirectoryStructureAnalyzer creates a new directory structure analyzer
func NewDirectoryStructureAnalyzer() *DirectoryStructureAnalyzer {
	return &DirectoryStructureAnalyzer{}
}

// AnalyzeStructure analyzes the directory structure for a target
func (a *DirectoryStructureAnalyzer) AnalyzeStructure(target BulkCloneTarget, repositories []Repository) *StructureAnalysis {
	generator := NewDirectoryPathGenerator()
	validator := NewDirectoryStructureValidator()
	resolver := NewDirectoryResolver(target)

	// Generate paths
	paths := generator.GenerateRepositoryPaths(target, repositories)

	// Validate structure
	structureErr := validator.ValidateStructure(target)
	pathIssues := validator.ValidateRepositoryPaths(paths)

	// Calculate statistics
	stats := a.calculateStatistics(paths)

	return &StructureAnalysis{
		Structure:       resolver.GetDirectoryStructure(),
		RepositoryPaths: paths,
		Statistics:      stats,
		StructureError:  structureErr,
		PathIssues:      pathIssues,
		IsValid:         structureErr == nil && len(pathIssues) == 0,
	}
}

// calculateStatistics calculates statistics about the directory structure
func (a *DirectoryStructureAnalyzer) calculateStatistics(paths []RepositoryPath) StructureStatistics {
	stats := StructureStatistics{
		TotalRepositories: len(paths),
	}

	// Count unique parent directories
	parentDirs := make(map[string]bool)
	for _, path := range paths {
		parentDirs[path.GetParentDirectory()] = true
	}
	stats.UniqueDirectories = len(parentDirs)

	return stats
}

// StructureAnalysis contains the results of directory structure analysis
type StructureAnalysis struct {
	Structure       DirectoryStructure  `json:"structure"`
	RepositoryPaths []RepositoryPath    `json:"repository_paths"`
	Statistics      StructureStatistics `json:"statistics"`
	StructureError  error               `json:"structure_error,omitempty"`
	PathIssues      []ValidationIssue   `json:"path_issues,omitempty"`
	IsValid         bool                `json:"is_valid"`
}

// StructureStatistics provides statistics about directory structure
type StructureStatistics struct {
	TotalRepositories int `json:"total_repositories"`
	UniqueDirectories int `json:"unique_directories"`
}

// GetSummary returns a summary of the structure analysis
func (s *StructureAnalysis) GetSummary() string {
	if !s.IsValid {
		return "Invalid directory structure configuration"
	}

	if s.Structure.IsFlattened {
		return fmt.Sprintf("Flattened structure: %d repositories in %d directories",
			s.Statistics.TotalRepositories, s.Statistics.UniqueDirectories)
	} else {
		return fmt.Sprintf("Nested structure: %d repositories in %d directories",
			s.Statistics.TotalRepositories, s.Statistics.UniqueDirectories)
	}
}
