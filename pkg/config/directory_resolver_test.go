package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDirectoryResolver(t *testing.T) {
	target := BulkCloneTarget{
		Provider: "github",
		Name:     "test-org",
		CloneDir: "/home/user/repos/github",
		Flatten:  false,
	}

	resolver := NewDirectoryResolver(target)

	assert.NotNil(t, resolver)
	assert.Equal(t, target, resolver.target)
}

func TestDirectoryResolver_ResolveRepositoryPath(t *testing.T) {
	tests := []struct {
		name           string
		target         BulkCloneTarget
		repositoryName string
		expected       string
	}{
		{
			name: "normal structure",
			target: BulkCloneTarget{
				Provider: "github",
				Name:     "test-org",
				CloneDir: "/home/user/repos/github",
				Flatten:  false,
			},
			repositoryName: "test-repo",
			expected:       "/home/user/repos/github/test-org/test-repo",
		},
		{
			name: "flattened structure",
			target: BulkCloneTarget{
				Provider: "github",
				Name:     "test-org",
				CloneDir: "/home/user/repos/github",
				Flatten:  true,
			},
			repositoryName: "test-repo",
			expected:       "/home/user/repos/github/test-repo",
		},
		{
			name: "custom clone directory with flatten",
			target: BulkCloneTarget{
				Provider: "gitlab",
				Name:     "my-group",
				CloneDir: "/custom/path",
				Flatten:  true,
			},
			repositoryName: "project-a",
			expected:       "/custom/path/project-a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewDirectoryResolver(tt.target)
			result := resolver.ResolveRepositoryPath(tt.repositoryName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDirectoryResolver_GetBasePath(t *testing.T) {
	target := BulkCloneTarget{
		CloneDir: "/home/user/repos/github",
	}

	resolver := NewDirectoryResolver(target)
	basePath := resolver.GetBasePath()

	assert.Equal(t, "/home/user/repos/github", basePath)
}

func TestDirectoryResolver_GetOrganizationPath(t *testing.T) {
	tests := []struct {
		name     string
		target   BulkCloneTarget
		expected string
	}{
		{
			name: "normal structure",
			target: BulkCloneTarget{
				Name:     "test-org",
				CloneDir: "/home/user/repos/github",
				Flatten:  false,
			},
			expected: "/home/user/repos/github/test-org",
		},
		{
			name: "flattened structure",
			target: BulkCloneTarget{
				Name:     "test-org",
				CloneDir: "/home/user/repos/github",
				Flatten:  true,
			},
			expected: "/home/user/repos/github",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewDirectoryResolver(tt.target)
			orgPath := resolver.GetOrganizationPath()
			assert.Equal(t, tt.expected, orgPath)
		})
	}
}

func TestDirectoryResolver_GetDirectoryStructure(t *testing.T) {
	target := BulkCloneTarget{
		Provider: "github",
		Name:     "test-org",
		CloneDir: "/home/user/repos/github",
		Flatten:  true,
	}

	resolver := NewDirectoryResolver(target)
	structure := resolver.GetDirectoryStructure()

	assert.Equal(t, "/home/user/repos/github", structure.BasePath)
	assert.Equal(t, "/home/user/repos/github", structure.OrganizationPath)
	assert.True(t, structure.IsFlattened)
	assert.Equal(t, "github", structure.Provider)
	assert.Equal(t, "test-org", structure.TargetName)
}

func TestDirectoryStructure_GetDescription(t *testing.T) {
	tests := []struct {
		name        string
		structure   DirectoryStructure
		expectedMsg string
	}{
		{
			name: "flattened structure",
			structure: DirectoryStructure{
				OrganizationPath: "/home/user/repos/github",
				IsFlattened:      true,
			},
			expectedMsg: "Flattened structure: all repositories in /home/user/repos/github",
		},
		{
			name: "nested structure",
			structure: DirectoryStructure{
				OrganizationPath: "/home/user/repos/github/test-org",
				IsFlattened:      false,
			},
			expectedMsg: "Nested structure: repositories in /home/user/repos/github/test-org",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			description := tt.structure.GetDescription()
			assert.Equal(t, tt.expectedMsg, description)
		})
	}
}

func TestDirectoryStructure_GetExamplePath(t *testing.T) {
	tests := []struct {
		name      string
		structure DirectoryStructure
		repoName  string
		expected  string
	}{
		{
			name: "flattened example",
			structure: DirectoryStructure{
				OrganizationPath: "/home/user/repos/github",
				IsFlattened:      true,
			},
			repoName: "test-repo",
			expected: "/home/user/repos/github/test-repo",
		},
		{
			name: "nested example",
			structure: DirectoryStructure{
				OrganizationPath: "/home/user/repos/github/test-org",
				IsFlattened:      false,
			},
			repoName: "test-repo",
			expected: "/home/user/repos/github/test-org/test-repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			examplePath := tt.structure.GetExamplePath(tt.repoName)
			assert.Equal(t, tt.expected, examplePath)
		})
	}
}

func TestDirectoryPathGenerator_GenerateRepositoryPaths(t *testing.T) {
	generator := NewDirectoryPathGenerator()

	target := BulkCloneTarget{
		Provider: "github",
		Name:     "test-org",
		CloneDir: "/home/user/repos/github",
		Flatten:  false,
	}

	repositories := []Repository{
		{Name: "repo1"},
		{Name: "repo2"},
	}

	paths := generator.GenerateRepositoryPaths(target, repositories)

	assert.Len(t, paths, 2)
	assert.Equal(t, "repo1", paths[0].Repository.Name)
	assert.Equal(t, "/home/user/repos/github/test-org/repo1", paths[0].FullPath)
	assert.Equal(t, "test-org/repo1", paths[0].RelativePath)
	assert.Equal(t, "repo2", paths[1].Repository.Name)
	assert.Equal(t, "/home/user/repos/github/test-org/repo2", paths[1].FullPath)
	assert.Equal(t, "test-org/repo2", paths[1].RelativePath)
}

func TestDirectoryPathGenerator_GenerateRepositoryPaths_Flattened(t *testing.T) {
	generator := NewDirectoryPathGenerator()

	target := BulkCloneTarget{
		Provider: "github",
		Name:     "test-org",
		CloneDir: "/home/user/repos/github",
		Flatten:  true,
	}

	repositories := []Repository{
		{Name: "repo1"},
		{Name: "repo2"},
	}

	paths := generator.GenerateRepositoryPaths(target, repositories)

	assert.Len(t, paths, 2)
	assert.Equal(t, "repo1", paths[0].Repository.Name)
	assert.Equal(t, "/home/user/repos/github/repo1", paths[0].FullPath)
	assert.Equal(t, "repo1", paths[0].RelativePath)
	assert.Equal(t, "repo2", paths[1].Repository.Name)
	assert.Equal(t, "/home/user/repos/github/repo2", paths[1].FullPath)
	assert.Equal(t, "repo2", paths[1].RelativePath)
}

func TestRepositoryPath_GetParentDirectory(t *testing.T) {
	path := RepositoryPath{
		FullPath: "/home/user/repos/github/test-org/test-repo",
	}

	parentDir := path.GetParentDirectory()
	expectedParent := "/home/user/repos/github/test-org"

	assert.Equal(t, expectedParent, parentDir)
}

func TestRepositoryPath_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		path     RepositoryPath
		expected bool
	}{
		{
			name: "valid path",
			path: RepositoryPath{
				Repository: Repository{Name: "test-repo"},
				FullPath:   "/home/user/repos/github/test-repo",
			},
			expected: true,
		},
		{
			name: "empty repository name",
			path: RepositoryPath{
				Repository: Repository{Name: ""},
				FullPath:   "/home/user/repos/github/test-repo",
			},
			expected: false,
		},
		{
			name: "empty full path",
			path: RepositoryPath{
				Repository: Repository{Name: "test-repo"},
				FullPath:   "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.path.IsValid()
			assert.Equal(t, tt.expected, isValid)
		})
	}
}

func TestDirectoryStructureValidator_ValidateStructure(t *testing.T) {
	validator := NewDirectoryStructureValidator()

	tests := []struct {
		name        string
		target      BulkCloneTarget
		expectError bool
		expectedErr error
	}{
		{
			name: "valid structure",
			target: BulkCloneTarget{
				Name:     "test-org",
				CloneDir: "/home/user/repos/github",
			},
			expectError: false,
		},
		{
			name: "missing clone directory",
			target: BulkCloneTarget{
				Name:     "test-org",
				CloneDir: "",
			},
			expectError: true,
			expectedErr: ErrInvalidCloneDir,
		},
		{
			name: "missing name",
			target: BulkCloneTarget{
				Name:     "",
				CloneDir: "/home/user/repos/github",
			},
			expectError: true,
			expectedErr: ErrMissingName,
		},
		{
			name: "unsafe path",
			target: BulkCloneTarget{
				Name:     "test-org",
				CloneDir: "/home/user/../etc/passwd",
			},
			expectError: true,
			expectedErr: ErrUnsafePath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateStructure(tt.target)

			if tt.expectError {
				assert.Error(t, err)

				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDirectoryStructureValidator_ValidateRepositoryPaths(t *testing.T) {
	validator := NewDirectoryStructureValidator()

	tests := []struct {
		name           string
		paths          []RepositoryPath
		expectedIssues int
	}{
		{
			name: "valid paths",
			paths: []RepositoryPath{
				{
					Repository: Repository{Name: "repo1"},
					FullPath:   "/home/user/repos/github/repo1",
				},
				{
					Repository: Repository{Name: "repo2"},
					FullPath:   "/home/user/repos/github/repo2",
				},
			},
			expectedIssues: 0,
		},
		{
			name: "path conflict",
			paths: []RepositoryPath{
				{
					Repository: Repository{Name: "repo1"},
					FullPath:   "/home/user/repos/github/same-path",
				},
				{
					Repository: Repository{Name: "repo2"},
					FullPath:   "/home/user/repos/github/same-path",
				},
			},
			expectedIssues: 1,
		},
		{
			name: "invalid path",
			paths: []RepositoryPath{
				{
					Repository: Repository{Name: ""},
					FullPath:   "/home/user/repos/github/repo1",
				},
			},
			expectedIssues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := validator.ValidateRepositoryPaths(tt.paths)
			assert.Len(t, issues, tt.expectedIssues)
		})
	}
}

func TestDirectoryStructureAnalyzer_AnalyzeStructure(t *testing.T) {
	analyzer := NewDirectoryStructureAnalyzer()

	target := BulkCloneTarget{
		Provider: "github",
		Name:     "test-org",
		CloneDir: "/home/user/repos/github",
		Flatten:  false,
	}

	repositories := []Repository{
		{Name: "repo1"},
		{Name: "repo2"},
	}

	analysis := analyzer.AnalyzeStructure(target, repositories)

	assert.NotNil(t, analysis)
	assert.True(t, analysis.IsValid)
	assert.NoError(t, analysis.StructureError)
	assert.Len(t, analysis.PathIssues, 0)
	assert.Len(t, analysis.RepositoryPaths, 2)
	assert.Equal(t, 2, analysis.Statistics.TotalRepositories)
	assert.Equal(t, 1, analysis.Statistics.UniqueDirectories) // All repos in same org directory
}

func TestDirectoryStructureAnalyzer_AnalyzeStructure_Flattened(t *testing.T) {
	analyzer := NewDirectoryStructureAnalyzer()

	target := BulkCloneTarget{
		Provider: "github",
		Name:     "test-org",
		CloneDir: "/home/user/repos/github",
		Flatten:  true,
	}

	repositories := []Repository{
		{Name: "repo1"},
		{Name: "repo2"},
	}

	analysis := analyzer.AnalyzeStructure(target, repositories)

	assert.NotNil(t, analysis)
	assert.True(t, analysis.IsValid)
	assert.True(t, analysis.Structure.IsFlattened)
	assert.Len(t, analysis.RepositoryPaths, 2)
	assert.Equal(t, 2, analysis.Statistics.TotalRepositories)
	assert.Equal(t, 1, analysis.Statistics.UniqueDirectories) // All repos in base directory
}

func TestStructureAnalysis_GetSummary(t *testing.T) {
	tests := []struct {
		name        string
		analysis    StructureAnalysis
		expectedMsg string
	}{
		{
			name: "valid flattened structure",
			analysis: StructureAnalysis{
				Structure: DirectoryStructure{IsFlattened: true},
				Statistics: StructureStatistics{
					TotalRepositories: 5,
					UniqueDirectories: 1,
				},
				IsValid: true,
			},
			expectedMsg: "Flattened structure: 5 repositories in 1 directories",
		},
		{
			name: "valid nested structure",
			analysis: StructureAnalysis{
				Structure: DirectoryStructure{IsFlattened: false},
				Statistics: StructureStatistics{
					TotalRepositories: 3,
					UniqueDirectories: 2,
				},
				IsValid: true,
			},
			expectedMsg: "Nested structure: 3 repositories in 2 directories",
		},
		{
			name: "invalid structure",
			analysis: StructureAnalysis{
				IsValid: false,
			},
			expectedMsg: "Invalid directory structure configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := tt.analysis.GetSummary()
			assert.Equal(t, tt.expectedMsg, summary)
		})
	}
}

func TestIntegration_FlattenDirectoryStructure(t *testing.T) {
	// Integration test to verify the complete directory structure workflow
	config := &Config{
		Version: "1.0.0",
		Providers: map[string]Provider{
			"github": {
				Token: "test-token",
				Orgs: []GitTarget{
					{
						Name:     "test-org",
						CloneDir: "/tmp/test-repos/github",
						Flatten:  true,
					},
				},
			},
		},
	}

	integration := NewBulkCloneIntegration(config)
	targets, err := integration.GetAllTargets()

	assert.NoError(t, err)
	assert.Len(t, targets, 1)

	target := targets[0]
	assert.True(t, target.Flatten)
	assert.Equal(t, "/tmp/test-repos/github", target.CloneDir)

	// Test directory resolver with the target
	resolver := NewDirectoryResolver(target)
	repoPath := resolver.ResolveRepositoryPath("test-repo")
	expectedPath := filepath.Join("/tmp/test-repos/github", "test-repo")

	assert.Equal(t, expectedPath, repoPath)

	// Test structure analysis
	analyzer := NewDirectoryStructureAnalyzer()
	repositories := []Repository{
		{Name: "repo1"},
		{Name: "repo2"},
	}

	analysis := analyzer.AnalyzeStructure(target, repositories)
	assert.True(t, analysis.IsValid)
	assert.True(t, analysis.Structure.IsFlattened)
	assert.Contains(t, analysis.GetSummary(), "Flattened structure")
}
