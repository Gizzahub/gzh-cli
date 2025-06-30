package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_EndToEndConfiguration(t *testing.T) {
	// Test end-to-end configuration loading, parsing, and integration
	tmpDir, err := os.MkdirTemp("", "gzh-integration-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a comprehensive test configuration
	configContent := `# gzh-manager configuration
version: "1.0.0"
default_provider: github

providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "test-org-1"
        visibility: "public"
        clone_dir: "${HOME}/repos/github/test-org-1"
        match: "^test-.*"
        exclude: ["temp-*", "*-backup"]
        strategy: "reset"
        flatten: false
      - name: "test-org-2"
        visibility: "all"
        clone_dir: "${HOME}/repos/github-flat"
        flatten: true
        strategy: "pull"
  
  gitlab:
    token: "${GITLAB_TOKEN}"
    groups:
      - name: "group-1"
        visibility: "private"
        recursive: true
        clone_dir: "${HOME}/repos/gitlab/group-1"
        exclude: ["archived-*"]
      - name: "group-2"
        visibility: "public"
        recursive: false
        clone_dir: "${HOME}/repos/gitlab/group-2"
        match: "project-.*"

  gitea:
    token: "${GITEA_TOKEN}"
    orgs:
      - name: "gitea-org"
        visibility: "all"
        clone_dir: "${HOME}/repos/gitea"
        strategy: "fetch"
`

	configPath := filepath.Join(tmpDir, "gzh.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Set up environment variables
	os.Setenv("GITHUB_TOKEN", "real-github-token")
	os.Setenv("GITLAB_TOKEN", "real-gitlab-token")
	os.Setenv("GITEA_TOKEN", "real-gitea-token")
	os.Setenv("HOME", "/home/testuser")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITLAB_TOKEN")
		os.Unsetenv("GITEA_TOKEN")
		os.Unsetenv("HOME")
	}()

	// Load and parse the configuration
	config, err := ParseYAMLFile(configPath)
	require.NoError(t, err, "Failed to parse YAML file")
	require.NotNil(t, config)

	// Verify basic configuration
	assert.Equal(t, "1.0.0", config.Version)
	assert.Equal(t, "github", config.DefaultProvider)
	assert.Len(t, config.Providers, 3)

	// Test GitHub provider configuration
	github := config.Providers["github"]
	assert.Equal(t, "real-github-token", github.Token)
	assert.Len(t, github.Orgs, 2)

	// Test first GitHub org
	org1 := github.Orgs[0]
	assert.Equal(t, "test-org-1", org1.Name)
	assert.Equal(t, "public", org1.Visibility)
	assert.Equal(t, "/home/testuser/repos/github/test-org-1", org1.CloneDir)
	assert.Equal(t, "^test-.*", org1.Match)
	assert.Equal(t, []string{"temp-*", "*-backup"}, org1.Exclude)
	assert.Equal(t, "reset", org1.Strategy)
	assert.False(t, org1.Flatten)

	// Test second GitHub org
	org2 := github.Orgs[1]
	assert.Equal(t, "test-org-2", org2.Name)
	assert.Equal(t, "all", org2.Visibility)
	assert.Equal(t, "/home/testuser/repos/github-flat", org2.CloneDir)
	assert.True(t, org2.Flatten)
	assert.Equal(t, "pull", org2.Strategy)

	// Test GitLab provider configuration
	gitlab := config.Providers["gitlab"]
	assert.Equal(t, "real-gitlab-token", gitlab.Token)
	assert.Len(t, gitlab.Groups, 2)

	// Test first GitLab group
	group1 := gitlab.Groups[0]
	assert.Equal(t, "group-1", group1.Name)
	assert.Equal(t, "private", group1.Visibility)
	assert.True(t, group1.Recursive)
	assert.Equal(t, "/home/testuser/repos/gitlab/group-1", group1.CloneDir)
	assert.Equal(t, []string{"archived-*"}, group1.Exclude)

	// Test second GitLab group
	group2 := gitlab.Groups[1]
	assert.Equal(t, "group-2", group2.Name)
	assert.Equal(t, "public", group2.Visibility)
	assert.False(t, group2.Recursive)
	assert.Equal(t, "project-.*", group2.Match)

	// Test Gitea provider configuration
	gitea := config.Providers["gitea"]
	assert.Equal(t, "real-gitea-token", gitea.Token)
	assert.Len(t, gitea.Orgs, 1)

	giteaOrg := gitea.Orgs[0]
	assert.Equal(t, "gitea-org", giteaOrg.Name)
	assert.Equal(t, "all", giteaOrg.Visibility)
	assert.Equal(t, "/home/testuser/repos/gitea", giteaOrg.CloneDir)
	assert.Equal(t, "fetch", giteaOrg.Strategy)

	// Test integration with BulkCloneIntegration
	integration := NewBulkCloneIntegration(config)
	targets, err := integration.GetAllTargets()
	require.NoError(t, err)
	assert.Len(t, targets, 5) // 2 GitHub orgs + 2 GitLab groups + 1 Gitea org = 5 targets

	// Verify targets have correct flatten settings
	var flattenedTargets []BulkCloneTarget
	for _, target := range targets {
		if target.Flatten {
			flattenedTargets = append(flattenedTargets, target)
		}
	}
	assert.Len(t, flattenedTargets, 1) // Only test-org-2 has flatten=true
	assert.Equal(t, "test-org-2", flattenedTargets[0].Name)
}

func TestIntegration_ConfigurationValidation(t *testing.T) {
	tests := []struct {
		name          string
		config        string
		expectError   bool
		errorContains string
	}{
		{
			name: "valid minimal configuration",
			config: `
version: "1.0.0"
providers:
  github:
    token: "test-token"
    orgs:
      - name: "test-org"
`,
			expectError: false,
		},
		{
			name: "missing version",
			config: `
providers:
  github:
    token: "test-token"
    orgs:
      - name: "test-org"
`,
			expectError:   true,
			errorContains: "version",
		},
		{
			name: "missing token",
			config: `
version: "1.0.0"
providers:
  github:
    orgs:
      - name: "test-org"
`,
			expectError:   true,
			errorContains: "token",
		},
		{
			name: "invalid visibility",
			config: `
version: "1.0.0"
providers:
  github:
    token: "test-token"
    orgs:
      - name: "test-org"
        visibility: "invalid"
`,
			expectError:   true,
			errorContains: "visibility",
		},
		{
			name: "invalid strategy",
			config: `
version: "1.0.0"
providers:
  github:
    token: "test-token"
    orgs:
      - name: "test-org"
        strategy: "invalid"
`,
			expectError:   true,
			errorContains: "strategy",
		},
		{
			name: "invalid regex pattern",
			config: `
version: "1.0.0"
providers:
  github:
    token: "test-token"
    orgs:
      - name: "test-org"
        match: "[invalid"
`,
			expectError:   true,
			errorContains: "regex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "gzh-validation-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			configPath := filepath.Join(tmpDir, "gzh.yaml")
			err = os.WriteFile(configPath, []byte(tt.config), 0o644)
			require.NoError(t, err)

			config, err := ParseYAMLFile(configPath)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
			}
		})
	}
}

func TestIntegration_RepositoryFiltering(t *testing.T) {
	// Test repository filtering integration
	target := GitTarget{
		Name:       "test-org",
		Visibility: "public",
		Match:      "^test-.*",
		Exclude:    []string{"test-temp-*", "test-backup"},
	}

	matcher, err := CreateRepositoryMatcherFromGitTarget(target)
	require.NoError(t, err)

	repositories := []Repository{
		{Name: "test-repo1", IsPrivate: false},
		{Name: "test-repo2", IsPrivate: true},
		{Name: "prod-repo", IsPrivate: false},
		{Name: "test-temp-file", IsPrivate: false},
		{Name: "test-backup", IsPrivate: false},
		{Name: "test-valid", IsPrivate: false},
	}

	filtered := matcher.FilterRepositoryList(repositories)

	// Should include only: test-repo1, test-valid
	// Excluded: test-repo2 (private), prod-repo (doesn't match), test-temp-file (excluded), test-backup (excluded)
	assert.Len(t, filtered, 2)

	var names []string
	for _, repo := range filtered {
		names = append(names, repo.Name)
	}
	assert.Contains(t, names, "test-repo1")
	assert.Contains(t, names, "test-valid")

	// Test statistics
	stats := matcher.GetStatistics(repositories)
	assert.Equal(t, 6, stats.OriginalStats.TotalRepositories)
	assert.Equal(t, 2, stats.FilteredStats.TotalRepositories)
	assert.InDelta(t, 33.3, stats.GetFilteringRatio(), 0.1)
}

func TestIntegration_DirectoryStructure(t *testing.T) {
	// Test integration of directory structure with configuration
	config := &Config{
		Version: "1.0.0",
		Providers: map[string]Provider{
			"github": {
				Token: "test-token",
				Orgs: []GitTarget{
					{
						Name:     "normal-org",
						CloneDir: "/repos/github",
						Flatten:  false,
					},
					{
						Name:     "flat-org",
						CloneDir: "/repos/github-flat",
						Flatten:  true,
					},
				},
			},
		},
	}

	integration := NewBulkCloneIntegration(config)
	targets, err := integration.GetAllTargets()
	require.NoError(t, err)
	assert.Len(t, targets, 2)

	// Test normal directory structure
	normalTarget := targets[0]
	if normalTarget.Name == "flat-org" {
		normalTarget = targets[1]
	}

	normalResolver := NewDirectoryResolver(normalTarget)
	normalPath := normalResolver.ResolveRepositoryPath("test-repo")
	expectedNormalPath := "/repos/github/normal-org/test-repo"
	assert.Equal(t, expectedNormalPath, normalPath)

	// Test flattened directory structure
	flatTarget := targets[0]
	if flatTarget.Name == "normal-org" {
		flatTarget = targets[1]
	}

	flatResolver := NewDirectoryResolver(flatTarget)
	flatPath := flatResolver.ResolveRepositoryPath("test-repo")
	expectedFlatPath := "/repos/github-flat/test-repo"
	assert.Equal(t, expectedFlatPath, flatPath)

	// Test directory structure analysis
	analyzer := NewDirectoryStructureAnalyzer()
	repositories := []Repository{
		{Name: "repo1"},
		{Name: "repo2"},
		{Name: "repo3"},
	}

	normalAnalysis := analyzer.AnalyzeStructure(normalTarget, repositories)
	assert.True(t, normalAnalysis.IsValid)
	assert.False(t, normalAnalysis.Structure.IsFlattened)
	assert.Equal(t, 3, normalAnalysis.Statistics.TotalRepositories)
	assert.Equal(t, 1, normalAnalysis.Statistics.UniqueDirectories) // All in same org dir

	flatAnalysis := analyzer.AnalyzeStructure(flatTarget, repositories)
	assert.True(t, flatAnalysis.IsValid)
	assert.True(t, flatAnalysis.Structure.IsFlattened)
	assert.Equal(t, 3, flatAnalysis.Statistics.TotalRepositories)
	assert.Equal(t, 1, flatAnalysis.Statistics.UniqueDirectories) // All in base dir
}

func TestIntegration_ProviderConfiguration(t *testing.T) {
	// Test integration of provider configuration with bulk clone operations
	config := &Config{
		Version: "1.0.0",
		Providers: map[string]Provider{
			"github": {
				Token: "github-token",
				Orgs: []GitTarget{
					{Name: "github-org"},
				},
			},
			"gitlab": {
				Token: "gitlab-token",
				Groups: []GitTarget{
					{Name: "gitlab-group"},
				},
			},
			"gitea": {
				Token: "gitea-token",
				Orgs: []GitTarget{
					{Name: "gitea-org"},
				},
			},
		},
	}

	// Test bulk clone executor creation
	executor, err := NewBulkCloneExecutor(config)
	require.NoError(t, err)
	assert.NotNil(t, executor)
	assert.Len(t, executor.cloners, 3)

	// Verify all providers have cloners
	assert.Contains(t, executor.cloners, "github")
	assert.Contains(t, executor.cloners, "gitlab")
	assert.Contains(t, executor.cloners, "gitea")

	// Test provider-specific target retrieval
	integration := NewBulkCloneIntegration(config)

	githubTargets, err := integration.GetTargetsByProvider("github")
	require.NoError(t, err)
	assert.Len(t, githubTargets, 1)
	assert.Equal(t, "github-org", githubTargets[0].Name)

	gitlabTargets, err := integration.GetTargetsByProvider("gitlab")
	require.NoError(t, err)
	assert.Len(t, gitlabTargets, 1)
	assert.Equal(t, "gitlab-group", gitlabTargets[0].Name)

	giteaTargets, err := integration.GetTargetsByProvider("gitea")
	require.NoError(t, err)
	assert.Len(t, giteaTargets, 1)
	assert.Equal(t, "gitea-org", giteaTargets[0].Name)
}

func TestIntegration_ConfigurationSearchPaths(t *testing.T) {
	// Test configuration file search functionality
	tmpDir, err := os.MkdirTemp("", "gzh-search-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a config in the temporary directory
	configContent := `
version: "1.0.0"
providers:
  github:
    token: "test-token"
    orgs:
      - name: "test-org"
`

	configPath := filepath.Join(tmpDir, "gzh.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Test direct path loading
	config, err := LoadConfigFromFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", config.Version)

	// Test file existence checking
	assert.True(t, fileExists(configPath))
	assert.False(t, fileExists(filepath.Join(tmpDir, "nonexistent.yaml")))

	// Test path expansion
	expandedPath := expandPath(configPath)
	assert.True(t, filepath.IsAbs(expandedPath))
}

func TestIntegration_DefaultsAndFallbacks(t *testing.T) {
	// Test that defaults are properly applied in integration scenarios
	config := &Config{
		Version: "1.0.0",
		Providers: map[string]Provider{
			"github": {
				Token: "test-token",
				Orgs: []GitTarget{
					{Name: "test-org"}, // No explicit defaults
				},
			},
		},
	}

	// Apply defaults
	config.applyDefaults()

	// Verify defaults were applied
	assert.Equal(t, "github", config.DefaultProvider)

	org := config.Providers["github"].Orgs[0]
	assert.Equal(t, "all", org.Visibility)
	assert.Equal(t, "reset", org.Strategy)
	assert.False(t, org.Flatten)
	assert.False(t, org.Recursive)

	// Test integration with default clone directory resolution
	integration := NewBulkCloneIntegration(config)
	targets, err := integration.GetAllTargets()
	require.NoError(t, err)
	assert.Len(t, targets, 1)

	target := targets[0]
	assert.Contains(t, target.CloneDir, "repos")    // Should contain default repos path
	assert.Contains(t, target.CloneDir, "github")   // Should contain provider name
	assert.Contains(t, target.CloneDir, "test-org") // Should contain org name
}
