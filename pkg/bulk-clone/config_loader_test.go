package bulkclone

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindConfigFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Test with environment variable
	t.Run("environment variable", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "env-config.yaml")
		err := os.WriteFile(configPath, []byte("version: 0.1"), 0o644)
		require.NoError(t, err)

		os.Setenv("GZH_CONFIG_PATH", configPath)
		defer os.Unsetenv("GZH_CONFIG_PATH")

		found, err := FindConfigFile()
		assert.NoError(t, err)
		assert.Equal(t, configPath, found)
	})

	// Test with non-existent environment variable path
	t.Run("invalid environment variable", func(t *testing.T) {
		os.Setenv("GZH_CONFIG_PATH", "/non/existent/path.yaml")
		defer os.Unsetenv("GZH_CONFIG_PATH")

		_, err := FindConfigFile()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	// Test current directory search
	t.Run("current directory", func(t *testing.T) {
		// Change to temp directory
		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)

		os.Chdir(tempDir)

		configPath := filepath.Join(tempDir, "bulk-clone.yaml")
		err := os.WriteFile(configPath, []byte("version: 0.1"), 0o644)
		require.NoError(t, err)

		found, err := FindConfigFile()
		assert.NoError(t, err)
		// The found path may be relative "./bulk-clone.yaml" instead of absolute
		assert.True(t, found == configPath || found == "./bulk-clone.yaml")
	})
}

func TestLoadConfig(t *testing.T) {
	tempDir := t.TempDir()

	// Create a valid config file
	configContent := `version: "0.1"
default:
  protocol: https
  github:
    root_path: "$HOME/test-repos"
    org_name: "test-org"
  gitlab:
    root_path: "$HOME/test-repos"
    group_name: "test-group"
repo_roots:
  - root_path: "$HOME/my-projects"
    provider: "github"
    protocol: "https"
    org_name: "my-org"
`
	configPath := filepath.Join(tempDir, "test-config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	t.Run("load specific file", func(t *testing.T) {
		cfg, err := LoadConfig(configPath)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "0.1", cfg.Version)
		assert.Equal(t, "https", cfg.Default.Protocol)
		assert.Equal(t, "test-org", cfg.Default.Github.OrgName)
		assert.Len(t, cfg.RepoRoots, 1)
	})

	t.Run("load with invalid path", func(t *testing.T) {
		_, err := LoadConfig("/non/existent/config.yaml")
		assert.Error(t, err)
	})
}

func TestGetGithubOrgConfig(t *testing.T) {
	cfg := &bulkCloneConfig{
		Version: "0.1",
		Default: bulkCloneDefault{
			Protocol: "https",
			Github: bulkCloneDefaultGithub{
				RootPath: "/default/path",
				OrgName:  "default-org",
			},
		},
		RepoRoots: []BulkCloneGithub{
			{
				RootPath: "/specific/path",
				Provider: "github",
				Protocol: "ssh",
				OrgName:  "specific-org",
			},
		},
	}

	t.Run("get specific org from repo_roots", func(t *testing.T) {
		orgConfig, err := cfg.GetGithubOrgConfig("specific-org")
		assert.NoError(t, err)
		assert.Equal(t, "/specific/path", orgConfig.RootPath)
		assert.Equal(t, "ssh", orgConfig.Protocol)
	})

	t.Run("get org from defaults", func(t *testing.T) {
		orgConfig, err := cfg.GetGithubOrgConfig("default-org")
		assert.NoError(t, err)
		assert.Equal(t, "/default/path", orgConfig.RootPath)
		assert.Equal(t, "https", orgConfig.Protocol)
	})

	t.Run("org not found", func(t *testing.T) {
		_, err := cfg.GetGithubOrgConfig("unknown-org")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no configuration found")
	})
}

func TestExpandPath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "expand home directory",
			input:    "~/test/path",
			expected: filepath.Join(homeDir, "test/path"),
		},
		{
			name:     "expand environment variable",
			input:    "$HOME/test/path",
			expected: filepath.Join(homeDir, "test/path"),
		},
		{
			name:     "no expansion needed",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "empty path",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandPath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadConfigWithOverlays(t *testing.T) {
	tempDir := t.TempDir()

	// Create base config file
	baseConfig := `version: "0.1"
default:
  protocol: https
  github:
    root_path: "$HOME/base-repos"
    org_name: "base-org"
  gitlab:
    root_path: "$HOME/base-gitlab"
    group_name: "base-group"
repo_roots:
  - root_path: "$HOME/base-work"
    provider: "github"
    protocol: "https"
    org_name: "base-company"
ignore_names:
  - "test-.*"
`
	basePath := filepath.Join(tempDir, "base-config.yaml")
	err := os.WriteFile(basePath, []byte(baseConfig), 0o644)
	require.NoError(t, err)

	// Create home overlay config file
	homeOverlay := `default:
  protocol: ssh
  github:
    root_path: "$HOME/home-repos"
    org_name: "home-org"
repo_roots:
  - root_path: "$HOME/home-work"
    provider: "github"
    protocol: "ssh"
    org_name: "base-company"  # Override existing
  - root_path: "$HOME/home-personal"
    provider: "github"
    protocol: "ssh"
    org_name: "personal-org"  # New entry
ignore_names:
  - "home-.*"
`
	homePath := filepath.Join(tempDir, "home-overlay.yaml")
	err = os.WriteFile(homePath, []byte(homeOverlay), 0o644)
	require.NoError(t, err)

	// Create work overlay config file
	workOverlay := `default:
  github:
    root_path: "$HOME/work-repos"
repo_roots:
  - root_path: "$HOME/work-specific"
    provider: "github"
    protocol: "https"
    org_name: "work-org"
ignore_names:
  - "work-.*"
`
	workPath := filepath.Join(tempDir, "work-overlay.yaml")
	err = os.WriteFile(workPath, []byte(workOverlay), 0o644)
	require.NoError(t, err)

	t.Run("load config with home overlay", func(t *testing.T) {
		cfg, err := LoadConfigWithOverlays(basePath, homePath)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		// Check that protocol was overridden
		assert.Equal(t, "ssh", cfg.Default.Protocol)

		// Check that GitHub root path was overridden
		assert.Equal(t, "$HOME/home-repos", cfg.Default.Github.RootPath)
		assert.Equal(t, "home-org", cfg.Default.Github.OrgName)

		// Check that GitLab config from base is preserved
		assert.Equal(t, "$HOME/base-gitlab", cfg.Default.Gitlab.RootPath)
		assert.Equal(t, "base-group", cfg.Default.Gitlab.GroupName)

		// Check repo_roots merging
		assert.Len(t, cfg.RepoRoots, 2)

		// base-company should be overridden
		var (
			baseCompanyRepo *BulkCloneGithub
			personalRepo    *BulkCloneGithub
		)

		for i := range cfg.RepoRoots {
			if cfg.RepoRoots[i].OrgName == "base-company" {
				baseCompanyRepo = &cfg.RepoRoots[i]
			}

			if cfg.RepoRoots[i].OrgName == "personal-org" {
				personalRepo = &cfg.RepoRoots[i]
			}
		}

		require.NotNil(t, baseCompanyRepo)
		assert.Equal(t, "$HOME/home-work", baseCompanyRepo.RootPath)
		assert.Equal(t, "ssh", baseCompanyRepo.Protocol)

		require.NotNil(t, personalRepo)
		assert.Equal(t, "$HOME/home-personal", personalRepo.RootPath)

		// Check ignore patterns were appended
		assert.Len(t, cfg.IgnoreNameRegexes, 2)
		assert.Contains(t, cfg.IgnoreNameRegexes, "test-.*")
		assert.Contains(t, cfg.IgnoreNameRegexes, "home-.*")
	})

	t.Run("load config with multiple overlays", func(t *testing.T) {
		cfg, err := LoadConfigWithOverlays(basePath, homePath, workPath)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		// Check that work overlay further overrode the GitHub root path
		assert.Equal(t, "$HOME/work-repos", cfg.Default.Github.RootPath)
		// But protocol should still be ssh from home overlay
		assert.Equal(t, "ssh", cfg.Default.Protocol)

		// Check repo_roots - should have 3 entries now
		assert.Len(t, cfg.RepoRoots, 3)

		var workRepo *BulkCloneGithub

		for i := range cfg.RepoRoots {
			if cfg.RepoRoots[i].OrgName == "work-org" {
				workRepo = &cfg.RepoRoots[i]
				break
			}
		}

		require.NotNil(t, workRepo)
		assert.Equal(t, "$HOME/work-specific", workRepo.RootPath)

		// Check ignore patterns - should have all 3
		assert.Len(t, cfg.IgnoreNameRegexes, 3)
		assert.Contains(t, cfg.IgnoreNameRegexes, "test-.*")
		assert.Contains(t, cfg.IgnoreNameRegexes, "home-.*")
		assert.Contains(t, cfg.IgnoreNameRegexes, "work-.*")
	})

	t.Run("load config with non-existent overlay", func(t *testing.T) {
		cfg, err := LoadConfigWithOverlays(basePath, "/non/existent/overlay.yaml")
		assert.NoError(t, err) // Should not error for non-existent overlay
		assert.NotNil(t, cfg)

		// Should be same as base config
		assert.Equal(t, "https", cfg.Default.Protocol)
		assert.Equal(t, "$HOME/base-repos", cfg.Default.Github.RootPath)
	})
}

func TestGetOverlayConfigPaths(t *testing.T) {
	paths := GetOverlayConfigPaths()
	assert.NotEmpty(t, paths)

	// Should include current directory overlays
	assert.Contains(t, paths, "./bulk-clone.home.yaml")
	assert.Contains(t, paths, "./bulk-clone.home.yml")
	assert.Contains(t, paths, "./bulk-clone.work.yaml")
	assert.Contains(t, paths, "./bulk-clone.work.yml")

	// Should include home directory overlays if home directory exists
	homeDir, err := os.UserHomeDir()
	if err == nil {
		assert.Contains(t, paths, filepath.Join(homeDir, ".config", "gzh-manager", "bulk-clone.home.yaml"))
		assert.Contains(t, paths, filepath.Join(homeDir, ".config", "gzh-manager", "bulk-clone.work.yaml"))
	}
}

func TestMergeConfig(t *testing.T) {
	base := &bulkCloneConfig{
		Version: "0.1",
		Default: bulkCloneDefault{
			Protocol: "https",
			Github: bulkCloneDefaultGithub{
				RootPath: "/base/github",
				OrgName:  "base-org",
			},
			Gitlab: bulkCloneDefaultGitlab{
				RootPath:  "/base/gitlab",
				GroupName: "base-group",
				Recursive: false,
			},
		},
		IgnoreNameRegexes: []string{"base-.*"},
		RepoRoots: []BulkCloneGithub{
			{
				RootPath: "/base/work",
				Provider: "github",
				Protocol: "https",
				OrgName:  "work-org",
			},
		},
	}

	overlay := &bulkCloneConfig{
		Version: "0.2",
		Default: bulkCloneDefault{
			Protocol: "ssh",
			Github: bulkCloneDefaultGithub{
				RootPath: "/overlay/github",
			},
			Gitlab: bulkCloneDefaultGitlab{
				Recursive: true,
			},
		},
		IgnoreNameRegexes: []string{"overlay-.*"},
		RepoRoots: []BulkCloneGithub{
			{
				RootPath: "/overlay/work",
				Provider: "github",
				Protocol: "ssh",
				OrgName:  "work-org", // Same org name - should override
			},
			{
				RootPath: "/overlay/personal",
				Provider: "github",
				Protocol: "ssh",
				OrgName:  "personal-org", // New org - should append
			},
		},
	}

	base.mergeConfig(overlay)

	// Check version override
	assert.Equal(t, "0.2", base.Version)

	// Check default overrides
	assert.Equal(t, "ssh", base.Default.Protocol)
	assert.Equal(t, "/overlay/github", base.Default.Github.RootPath)
	assert.Equal(t, "base-org", base.Default.Github.OrgName)      // Should be preserved
	assert.Equal(t, "/base/gitlab", base.Default.Gitlab.RootPath) // Should be preserved
	assert.Equal(t, "base-group", base.Default.Gitlab.GroupName)  // Should be preserved
	assert.True(t, base.Default.Gitlab.Recursive)                 // Should be overridden

	// Check ignore patterns were appended
	assert.Len(t, base.IgnoreNameRegexes, 2)
	assert.Contains(t, base.IgnoreNameRegexes, "base-.*")
	assert.Contains(t, base.IgnoreNameRegexes, "overlay-.*")

	// Check repo_roots merging
	assert.Len(t, base.RepoRoots, 2)

	var workRepo, personalRepo *BulkCloneGithub

	for i := range base.RepoRoots {
		if base.RepoRoots[i].OrgName == "work-org" {
			workRepo = &base.RepoRoots[i]
		}

		if base.RepoRoots[i].OrgName == "personal-org" {
			personalRepo = &base.RepoRoots[i]
		}
	}

	require.NotNil(t, workRepo)
	assert.Equal(t, "/overlay/work", workRepo.RootPath) // Should be overridden
	assert.Equal(t, "ssh", workRepo.Protocol)

	require.NotNil(t, personalRepo)
	assert.Equal(t, "/overlay/personal", personalRepo.RootPath) // Should be new
}
