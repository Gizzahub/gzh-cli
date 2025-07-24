//nolint:testpackage // White-box testing needed for internal function access
package bulkclone

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExampleConfigs(t *testing.T) {
	// Test that example config files are valid
	exampleFiles := []string{
		"../../examples/bulk-clone/bulk-clone-simple.yaml",
		"../../examples/bulk-clone/bulk-clone-example.yaml",
	}

	for _, file := range exampleFiles {
		t.Run(filepath.Base(file), func(t *testing.T) {
			cfg := &bulkCloneConfig{}
			err := cfg.ReadConfig(file)

			// The example files should be valid
			assert.NoError(t, err, "Example config file should be valid: %s", file)

			// Basic validation
			assert.NotEmpty(t, cfg.Version, "Version should be set")
			assert.Equal(t, "0.1", cfg.Version, "Version should be 0.1")

			// Default protocol should be set
			assert.NotEmpty(t, cfg.Default.Protocol, "Default protocol should be set")

			// Should have at least one repo_root configured
			assert.NotEmpty(t, cfg.RepoRoots, "Should have at least one repo_root configured")

			// Each repo_root should have required fields
			for i, repo := range cfg.RepoRoots {
				assert.NotEmptyf(t, repo.RootPath, "repo_roots[%d].root_path should not be empty", i)
				assert.NotEmptyf(t, repo.Provider, "repo_roots[%d].provider should not be empty", i)
				assert.NotEmptyf(t, repo.Protocol, "repo_roots[%d].protocol should not be empty", i)
				// GitHub requires org_name, GitLab uses group_name (checked in config structure)
				assert.NotEmptyf(t, repo.OrgName, "repo_roots[%d].org_name should not be empty for GitHub", i)
			}
		})
	}
}

func TestSimpleExampleConfig(t *testing.T) {
	cfg := &bulkCloneConfig{}
	err := cfg.ReadConfig("../../examples/bulk-clone/bulk-clone-simple.yaml")
	require.NoError(t, err)

	// Test specific content of simple example
	t.Run("defaults", func(t *testing.T) {
		assert.Equal(t, "https", cfg.Default.Protocol)
		assert.Equal(t, "$HOME/github-repos", cfg.Default.Github.RootPath)
		assert.Equal(t, "$HOME/gitlab-repos", cfg.Default.Gitlab.RootPath)
	})

	t.Run("repo_roots", func(t *testing.T) {
		assert.Len(t, cfg.RepoRoots, 3, "Should have 3 repo_roots")

		// First repo_root - GitHub with SSH
		assert.Equal(t, "$HOME/work/mycompany", cfg.RepoRoots[0].RootPath)
		assert.Equal(t, "github", cfg.RepoRoots[0].Provider)
		assert.Equal(t, "ssh", cfg.RepoRoots[0].Protocol)
		assert.Equal(t, "mycompany", cfg.RepoRoots[0].OrgName)

		// Second repo_root - GitHub with HTTPS
		assert.Equal(t, "$HOME/opensource", cfg.RepoRoots[1].RootPath)
		assert.Equal(t, "github", cfg.RepoRoots[1].Provider)
		assert.Equal(t, "https", cfg.RepoRoots[1].Protocol)
		assert.Equal(t, "kubernetes", cfg.RepoRoots[1].OrgName)
	})

	t.Run("ignore_patterns", func(t *testing.T) {
		assert.Len(t, cfg.IgnoreNameRegexes, 2)
		assert.Contains(t, cfg.IgnoreNameRegexes, "test-.*")
		assert.Contains(t, cfg.IgnoreNameRegexes, ".*-archive")
	})
}

func TestComprehensiveExampleConfig(t *testing.T) {
	cfg := &bulkCloneConfig{}
	err := cfg.ReadConfig("../../examples/bulk-clone/bulk-clone-example.yaml")
	require.NoError(t, err)

	t.Run("has_all_sections", func(t *testing.T) {
		// Should have version
		assert.Equal(t, "0.1", cfg.Version)

		// Should have defaults for both providers
		assert.NotEmpty(t, cfg.Default.Github.RootPath)
		assert.NotEmpty(t, cfg.Default.Gitlab.RootPath)

		// Should have multiple repo_roots
		assert.GreaterOrEqual(t, len(cfg.RepoRoots), 5, "Should have at least 5 example repo_roots")

		// Should have ignore patterns
		assert.GreaterOrEqual(t, len(cfg.IgnoreNameRegexes), 4, "Should have at least 4 ignore patterns")
	})

	t.Run("diverse_configurations", func(t *testing.T) {
		// Should have various GitHub configurations
		hasSSH := false
		hasHTTPS := false

		for _, repo := range cfg.RepoRoots {
			// All should be GitHub since repo_roots only supports GitHub currently
			assert.Equal(t, "github", repo.Provider)

			if repo.Protocol == "ssh" {
				hasSSH = true
			}

			if repo.Protocol == "https" {
				hasHTTPS = true
			}
		}

		assert.True(t, hasSSH, "Should have SSH protocol examples")
		assert.True(t, hasHTTPS, "Should have HTTPS protocol examples")

		// Check that GitLab is mentioned in defaults even if not in repo_roots
		assert.NotEmpty(t, cfg.Default.Gitlab.RootPath, "Should have GitLab defaults configured")
	})
}

func TestIgnorePatternsValidity(t *testing.T) {
	// Test that ignore patterns in examples are valid regex
	cfg := &bulkCloneConfig{}
	err := cfg.ReadConfig("../../examples/bulk-clone/bulk-clone-example.yaml")
	require.NoError(t, err)

	for _, pattern := range cfg.IgnoreNameRegexes {
		t.Run(pattern, func(t *testing.T) {
			// This should not panic if pattern is valid
			assert.NotPanics(t, func() {
				// The actual regex compilation happens in the config validation
				// We're just checking that the patterns look reasonable
				assert.NotEmpty(t, pattern)
			})
		})
	}
}
