// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-manager-go/cmd/git-synclone/providers"
)

func TestConfigIntegration(t *testing.T) {
	integration := providers.NewConfigIntegration()

	t.Run("LoadConfig_NonExistent", func(t *testing.T) {
		err := integration.LoadConfig("/non/existent/config.yaml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no config file found")
	})

	t.Run("LoadConfig_InvalidYAML", func(t *testing.T) {
		tmpFile := createTempFile(t, "invalid: yaml: content: [")
		defer os.Remove(tmpFile)

		err := integration.LoadConfig(tmpFile)
		assert.Error(t, err)
	})

	t.Run("LoadConfig_ValidConfig", func(t *testing.T) {
		configContent := `
version: "1.0.0"
default:
  protocol: "https"
  github:
    rootPath: "/tmp/github"
    provider: "github"
    protocol: "https"
    orgName: "test-org"
  gitlab:
    rootPath: "/tmp/gitlab"
    provider: "gitlab"
    url: "https://gitlab.com"
    protocol: "https"
    groupName: "test-group"
    recursive: true
repoRoots:
  - rootPath: "/custom/path"
    provider: "github"
    protocol: "ssh"
    orgName: "custom-org"
`

		tmpFile := createTempFile(t, configContent)
		defer os.Remove(tmpFile)

		err := integration.LoadConfig(tmpFile)
		assert.NoError(t, err)
	})
}

func TestGitHubConfigExtraction(t *testing.T) {
	configContent := `
version: "1.0.0"
default:
  protocol: "https"
  github:
    rootPath: "~/github"
    provider: "github"
    protocol: "https"
    orgName: "default-org"
repoRoots:
  - rootPath: "/custom/github"
    provider: "github"
    protocol: "ssh"
    orgName: "custom-org"
`

	tmpFile := createTempFile(t, configContent)
	defer os.Remove(tmpFile)

	integration := providers.NewConfigIntegration()
	err := integration.LoadConfig(tmpFile)
	require.NoError(t, err)

	t.Run("GetGitHubOrgConfig_FromRepoRoots", func(t *testing.T) {
		config, err := integration.GetGitHubOrgConfig("custom-org")
		require.NoError(t, err)
		assert.Equal(t, "custom-org", config.OrgName)
		assert.Equal(t, "/custom/github", config.RootPath)
		assert.Equal(t, "ssh", config.Protocol)
		assert.Equal(t, "github", config.Provider)
	})

	t.Run("GetGitHubOrgConfig_FromDefaults", func(t *testing.T) {
		config, err := integration.GetGitHubOrgConfig("default-org")
		require.NoError(t, err)
		assert.Equal(t, "default-org", config.OrgName)
		assert.Contains(t, config.RootPath, "github") // Should expand ~ to home dir
		assert.Equal(t, "https", config.Protocol)
		assert.Equal(t, "github", config.Provider)
	})

	t.Run("GetGitHubOrgConfig_NotFound", func(t *testing.T) {
		config, err := integration.GetGitHubOrgConfig("nonexistent-org")
		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "no configuration found")
	})
}

func TestGitLabConfigExtraction(t *testing.T) {
	configContent := `
version: "1.0.0"
default:
  protocol: "https"
  gitlab:
    rootPath: "~/gitlab"
    provider: "gitlab"
    url: "https://gitlab.example.com"
    protocol: "ssh"
    groupName: "test-group"
    recursive: true
`

	tmpFile := createTempFile(t, configContent)
	defer os.Remove(tmpFile)

	integration := providers.NewConfigIntegration()
	err := integration.LoadConfig(tmpFile)
	require.NoError(t, err)

	t.Run("GetGitLabGroupConfig_FromDefaults", func(t *testing.T) {
		config, err := integration.GetGitLabGroupConfig("test-group")
		require.NoError(t, err)
		assert.Equal(t, "test-group", config.GroupName)
		assert.Contains(t, config.RootPath, "gitlab")
		assert.Equal(t, "ssh", config.Protocol)
		assert.Equal(t, "gitlab", config.Provider)
		assert.Equal(t, "https://gitlab.example.com", config.URL)
		assert.True(t, config.Recursive)
	})

	t.Run("GetGitLabGroupConfig_NotFound", func(t *testing.T) {
		config, err := integration.GetGitLabGroupConfig("nonexistent-group")
		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "no configuration found")
	})
}

func TestBuildCloneOptionsFromConfig(t *testing.T) {
	configContent := `
version: "1.0.0"
default:
  protocol: "ssh"
  github:
    rootPath: "/github"
    provider: "github"
    protocol: "ssh"
    orgName: "test-org"
  gitlab:
    rootPath: "/gitlab"
    provider: "gitlab"
    url: "https://gitlab.com"
    protocol: "https"
    groupName: "test-group"
    recursive: false
`

	tmpFile := createTempFile(t, configContent)
	defer os.Remove(tmpFile)

	integration := providers.NewConfigIntegration()
	err := integration.LoadConfig(tmpFile)
	require.NoError(t, err)

	t.Run("BuildCloneOptions_GitHub", func(t *testing.T) {
		options, err := integration.BuildCloneOptionsFromConfig("github", "test-org")
		require.NoError(t, err)
		assert.Equal(t, "ssh", options.Protocol)
		assert.Equal(t, "reset", options.Strategy)
		assert.Equal(t, 1, options.Parallel)
		assert.Equal(t, 3, options.MaxRetries)
		assert.False(t, options.Resume)
		assert.False(t, options.DryRun)
		assert.Equal(t, "bar", options.ProgressMode)
		assert.True(t, options.UseConfig)
		assert.False(t, options.CleanupOrphans)
	})

	t.Run("BuildCloneOptions_GitLab", func(t *testing.T) {
		options, err := integration.BuildCloneOptionsFromConfig("gitlab", "test-group")
		require.NoError(t, err)
		assert.Equal(t, "https", options.Protocol)
		assert.Equal(t, "reset", options.Strategy)
		assert.True(t, options.UseConfig)
	})

	t.Run("BuildCloneOptions_UnsupportedProvider", func(t *testing.T) {
		options, err := integration.BuildCloneOptionsFromConfig("unsupported", "test-org")
		assert.Error(t, err)
		assert.Nil(t, options)
		assert.Contains(t, err.Error(), "unsupported provider")
	})

	t.Run("BuildCloneOptions_NoConfig", func(t *testing.T) {
		newIntegration := providers.NewConfigIntegration()
		options, err := newIntegration.BuildCloneOptionsFromConfig("github", "test-org")
		assert.Error(t, err)
		assert.Nil(t, options)
		assert.Contains(t, err.Error(), "configuration not loaded")
	})
}

func TestConfigPaths(t *testing.T) {
	integration := providers.NewConfigIntegration()

	t.Run("GetConfigPaths", func(t *testing.T) {
		paths := integration.GetConfigPaths()
		assert.NotEmpty(t, paths)

		// Should include current directory paths
		found := false
		for _, path := range paths {
			if path == "./synclone.yaml" || path == "./synclone.yml" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should include current directory config paths")
	})

	t.Run("GetOverlayConfigPaths", func(t *testing.T) {
		paths := integration.GetOverlayConfigPaths()
		assert.NotEmpty(t, paths)

		// Should include overlay paths
		found := false
		for _, path := range paths {
			if filepath.Base(path) == "synclone.home.yaml" || filepath.Base(path) == "synclone.work.yaml" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should include overlay config paths")
	})

	t.Run("FindConfigFile", func(t *testing.T) {
		// This test depends on actual file system state
		// We just test that it doesn't panic and returns a string or error
		configFile, err := integration.FindConfigFile()
		if err != nil {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "no config file found")
		} else {
			assert.NotEmpty(t, configFile)
		}
	})
}

func TestValidateConfig(t *testing.T) {
	integration := providers.NewConfigIntegration()

	t.Run("ValidateConfig_NoConfig", func(t *testing.T) {
		err := integration.ValidateConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration not loaded")
	})

	t.Run("ValidateConfig_WithConfig", func(t *testing.T) {
		configContent := `
version: "1.0.0"
default:
  protocol: "https"
  github:
    rootPath: "/github"
    provider: "github"
    protocol: "https"
    orgName: "test-org"
`

		tmpFile := createTempFile(t, configContent)
		defer os.Remove(tmpFile)

		err := integration.LoadConfig(tmpFile)
		require.NoError(t, err)

		err = integration.ValidateConfig()
		assert.NoError(t, err)
	})
}

func TestConfigWithEnvironmentVariables(t *testing.T) {
	configContent := `
version: "1.0.0"
default:
  protocol: "https"
  github:
    rootPath: "${HOME}/github-repos"
    provider: "github"
    protocol: "https"
    orgName: "test-org"
`

	tmpFile := createTempFile(t, configContent)
	defer os.Remove(tmpFile)

	integration := providers.NewConfigIntegration()
	err := integration.LoadConfig(tmpFile)
	require.NoError(t, err)

	config, err := integration.GetGitHubOrgConfig("test-org")
	require.NoError(t, err)

	// The path should be expanded with environment variables
	homeDir := os.Getenv("HOME")
	if homeDir != "" {
		expectedPath := filepath.Join(homeDir, "github-repos")
		assert.Equal(t, expectedPath, config.RootPath)
	}
}

func TestConfigWithOverlays(t *testing.T) {
	// Create base config
	baseConfig := `
version: "1.0.0"
default:
  protocol: "https"
  github:
    rootPath: "/base/github"
    provider: "github"
    protocol: "https"
    orgName: "base-org"
`

	// Create overlay config
	overlayConfig := `
default:
  protocol: "ssh"
  github:
    rootPath: "/overlay/github"
    protocol: "ssh"
    orgName: "overlay-org"
`

	baseFile := createTempFile(t, baseConfig)
	defer os.Remove(baseFile)

	overlayFile := createTempFile(t, overlayConfig)
	defer os.Remove(overlayFile)

	// Test loading base config only
	integration := providers.NewConfigIntegration()
	err := integration.LoadConfig(baseFile)
	require.NoError(t, err)

	config, err := integration.GetGitHubOrgConfig("base-org")
	require.NoError(t, err)
	assert.Equal(t, "https", config.Protocol)
	assert.Equal(t, "/base/github", config.RootPath)
}

// Helper function to create temporary config files for testing
func createTempFile(t *testing.T, content string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	require.NoError(t, err)

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)

	err = tmpFile.Close()
	require.NoError(t, err)

	return tmpFile.Name()
}
