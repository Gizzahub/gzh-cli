package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name: "home directory expansion",
			path: "~/test.yaml",
		},
		{
			name:     "absolute path unchanged",
			path:     "/etc/config.yaml",
			expected: "/etc/config.yaml",
		},
		{
			name: "relative path to absolute",
			path: "./config.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.path)

			if tt.path == "~/test.yaml" {
				homeDir, _ := os.UserHomeDir()
				assert.Equal(t, filepath.Join(homeDir, "test.yaml"), result)
			} else if tt.expected != "" {
				assert.Equal(t, tt.expected, result)
			} else {
				// For relative paths, just check it's now absolute
				assert.True(t, filepath.IsAbs(result))
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "existing file",
			filename: tmpFile.Name(),
			expected: true,
		},
		{
			name:     "non-existing file",
			filename: "/path/that/does/not/exist.yaml",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fileExists(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	// Create a temporary config file
	configContent := `
version: "1.0.0"
default_provider: github
providers:
  github:
    token: "test-token"
    orgs:
      - name: "test-org"
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	config, err := LoadConfigFromFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "1.0.0", config.Version)
	assert.Equal(t, "github", config.DefaultProvider)
}

func TestCreateDefaultConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-config-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "gzh.yaml")

	// Set environment variables for test
	os.Setenv("GITHUB_TOKEN", "test-github-token")
	os.Setenv("GITLAB_TOKEN", "test-gitlab-token")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITLAB_TOKEN")
	}()

	err = CreateDefaultConfig(configPath)
	assert.NoError(t, err)

	// Verify file was created
	assert.True(t, fileExists(configPath))

	// Verify content can be loaded (validation will happen in separate tests)
	config, err := LoadConfigFromFile(configPath)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "1.0.0", config.Version)
}

func TestFindConfigFile(t *testing.T) {
	// Create a temporary config file in current directory
	tmpFile, err := os.CreateTemp(".", "gzh-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Temporarily modify ConfigSearchPaths to include our test file
	originalPaths := ConfigSearchPaths
	ConfigSearchPaths = []string{tmpFile.Name()}
	defer func() { ConfigSearchPaths = originalPaths }()

	foundPath, err := FindConfigFile()
	assert.NoError(t, err)
	assert.Contains(t, foundPath, filepath.Base(tmpFile.Name()))
}

func TestLoadConfigWithEnvVar(t *testing.T) {
	// Create a temporary config file
	configContent := `
version: "1.0.0"
providers:
  github:
    token: "test-token"
    orgs:
      - name: "test-org"
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Set environment variable
	os.Setenv("GZH_CONFIG_PATH", tmpFile.Name())
	defer os.Unsetenv("GZH_CONFIG_PATH")

	config, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "1.0.0", config.Version)
}

func TestGetConfigSearchPaths(t *testing.T) {
	paths := GetConfigSearchPaths()
	assert.Greater(t, len(paths), 0)

	// All paths should be absolute after expansion
	for _, path := range paths {
		assert.True(t, filepath.IsAbs(path), "Path should be absolute: %s", path)
	}
}

func TestValidateConfigFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "valid config",
			content: `
version: "1.0.0"
providers:
  github:
    token: "test"
    orgs:
      - name: "test"
`,
			wantErr: false,
		},
		{
			name: "invalid config - missing version",
			content: `
providers:
  github:
    token: "test"
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "test-*.yaml")
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(tt.content)
			require.NoError(t, err)
			tmpFile.Close()

			result, err := ValidateConfigFile(tmpFile.Name())
			assert.NoError(t, err) // ValidateConfigFile should not return an error, validation results are in the result
			if tt.wantErr {
				assert.False(t, result.Valid)
				assert.NotEmpty(t, result.Errors)
			} else {
				assert.True(t, result.Valid)
				assert.Empty(t, result.Errors)
			}
		})
	}
}
