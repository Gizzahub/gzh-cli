//nolint:testpackage // White-box testing needed for internal function access
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

			switch {
			case tt.path == "~/test.yaml":
				homeDir, _ := os.UserHomeDir()
				assert.Equal(t, filepath.Join(homeDir, "test.yaml"), result)
			case tt.expected != "":
				assert.Equal(t, tt.expected, result)
			default:
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

	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Warning: failed to remove temp file: %v", err)
		}
	}()

	if err := tmpFile.Close(); err != nil {
		t.Logf("Warning: failed to close temp file: %v", err)
	}

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
    organizations:
      - name: "test-org"
        clone_dir: "./repos/test-org"
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	require.NoError(t, err)

	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Warning: failed to remove temp file: %v", err)
		}
	}()

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	if err := tmpFile.Close(); err != nil {
		t.Logf("Warning: failed to close temp file: %v", err)
	}

	// require: assert는 실패해도 계속 진행하므로, 로드가 깨지면 다음 줄에서 nil을
	// 역참조해 SIGSEGV로 번진다. 어설션 실패로 끝나야 원인이 보인다.
	config, err := LoadConfigFromFile(tmpFile.Name())
	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, "1.0.0", config.Version)
	assert.Equal(t, "github", config.DefaultProvider)
}

func TestCreateDefaultConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-config-*")
	require.NoError(t, err)

	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	configPath := filepath.Join(tmpDir, "gzh.yaml")

	// t.Setenv: 이전에는 defer에서 Unsetenv로 되돌렸는데, 개발자 셸에 이미
	// GITHUB_TOKEN이 있으면 그것까지 지워버려 같은 패키지의 뒤따르는 테스트가
	// 오염됐다. TestStartupValidator_ValidateUnifiedConfig가 "환경변수 없음"
	// 경고 개수를 세는데, 이 테스트가 먼저 도느냐에 따라 결과가 뒤집혔다.
	// t.Setenv는 원래 값을 -- 없었다는 사실까지 포함해 -- 복원한다.
	t.Setenv("GITHUB_TOKEN", "test-github-token")
	t.Setenv("GITLAB_TOKEN", "test-gitlab-token")

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

	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Warning: failed to remove temp file: %v", err)
		}
	}()

	if err := tmpFile.Close(); err != nil {
		t.Logf("Warning: failed to close temp file: %v", err)
	}

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
    organizations:
      - name: "test-org"
        clone_dir: "./repos/test-org"
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	require.NoError(t, err)

	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Warning: failed to remove temp file: %v", err)
		}
	}()

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	if err := tmpFile.Close(); err != nil {
		t.Logf("Warning: failed to close temp file: %v", err)
	}

	// Set environment variable
	if err := os.Setenv("GZH_CONFIG_PATH", tmpFile.Name()); err != nil {
		t.Logf("Warning: failed to set GZH_CONFIG_PATH: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("GZH_CONFIG_PATH"); err != nil {
			t.Logf("Warning: failed to unset GZH_CONFIG_PATH: %v", err)
		}
	}()

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

// TestLoadConfigOrganizationKeyAliases verifies gzh.yaml SSoT is organizations
// while deprecated orgs/groups aliases still load (issue 32).
func TestLoadConfigOrganizationKeyAliases(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		provider    string
		wantOrgName string
		wantClone   string
	}{
		{
			name: "canonical organizations key",
			content: `
version: "1.0.0"
default_provider: github
providers:
  github:
    token: "test-token"
    organizations:
      - name: "canonical-org"
        clone_dir: "./repos/canonical"
`,
			provider:    "github",
			wantOrgName: "canonical-org",
			wantClone:   "./repos/canonical",
		},
		{
			name: "deprecated orgs alias",
			content: `
version: "1.0.0"
default_provider: github
providers:
  github:
    token: "test-token"
    orgs:
      - name: "alias-org"
        clone_dir: "./repos/alias"
`,
			provider:    "github",
			wantOrgName: "alias-org",
			wantClone:   "./repos/alias",
		},
		{
			name: "deprecated groups alias for gitlab",
			content: `
version: "1.0.0"
default_provider: gitlab
providers:
  gitlab:
    token: "test-token"
    groups:
      - name: "alias-group"
        clone_dir: "./repos/group"
`,
			provider:    "gitlab",
			wantOrgName: "alias-group",
			wantClone:   "./repos/group",
		},
		{
			name: "organizations wins over orgs when both present",
			content: `
version: "1.0.0"
providers:
  github:
    token: "test-token"
    organizations:
      - name: "canonical-wins"
        clone_dir: "./repos/canonical"
    orgs:
      - name: "alias-ignored"
        clone_dir: "./repos/ignored"
`,
			provider:    "github",
			wantOrgName: "canonical-wins",
			wantClone:   "./repos/canonical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "test-org-alias-*.yaml")
			require.NoError(t, err)
			t.Cleanup(func() {
				_ = os.Remove(tmpFile.Name())
			})

			_, err = tmpFile.WriteString(tt.content)
			require.NoError(t, err)
			require.NoError(t, tmpFile.Close())

			// Unified loader path (live SSoT)
			loader := NewUnifiedLoader()
			result, err := loader.LoadConfigFromPath(tmpFile.Name())
			require.NoError(t, err)
			require.NotNil(t, result)
			require.NotNil(t, result.Config)

			provider, ok := result.Config.Providers[tt.provider]
			require.True(t, ok, "provider %s present", tt.provider)
			require.Len(t, provider.Organizations, 1)
			assert.Equal(t, tt.wantOrgName, provider.Organizations[0].Name)
			assert.Equal(t, tt.wantClone, provider.Organizations[0].CloneDir)
			// Aliases must be cleared after normalize so re-marshal stays canonical
			assert.Nil(t, provider.OrgsAlias)
			assert.Nil(t, provider.GroupsAlias)

			// Public LoadConfigFromFile also accepts both shapes (converts to legacy Config)
			legacy, err := LoadConfigFromFile(tmpFile.Name())
			require.NoError(t, err)
			require.NotNil(t, legacy)
			legacyProvider, ok := legacy.Providers[tt.provider]
			require.True(t, ok)

			targets := legacyProvider.Orgs
			if tt.provider == "gitlab" {
				targets = legacyProvider.Groups
			}
			require.Len(t, targets, 1)
			assert.Equal(t, tt.wantOrgName, targets[0].Name)
			assert.Equal(t, tt.wantClone, targets[0].CloneDir)
		})
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
    organizations:
      - name: "test"
        clone_dir: "./repos/test"
`,
			wantErr: false,
		},
		{
			name: "valid config with orgs alias",
			content: `
version: "1.0.0"
providers:
  github:
    token: "test"
    orgs:
      - name: "test"
        clone_dir: "./repos/test"
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

			defer func() {
				if err := os.Remove(tmpFile.Name()); err != nil {
					t.Logf("Warning: failed to remove temp file: %v", err)
				}
			}()

			_, err = tmpFile.WriteString(tt.content)
			require.NoError(t, err)
			if err := tmpFile.Close(); err != nil {
				t.Logf("Warning: failed to close temp file: %v", err)
			}

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
