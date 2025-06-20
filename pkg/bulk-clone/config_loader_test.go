package bulk_clone

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
		Default: BulkCloneDefault{
			Protocol: "https",
			Github: BulkCloneDefaultGithub{
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
