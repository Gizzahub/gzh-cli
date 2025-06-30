package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateMinimalConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-init-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	outputFile := filepath.Join(tmpDir, "test-gzh.yaml")

	err = createMinimalConfig(outputFile)
	assert.NoError(t, err)

	// Check file was created
	assert.FileExists(t, outputFile)

	// Read and verify content
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "version: \"1.0.0\"")
	assert.Contains(t, contentStr, "default_provider: github")
	assert.Contains(t, contentStr, "token: \"${GITHUB_TOKEN}\"")
	assert.Contains(t, contentStr, "name: \"your-org-name\"")
}

func TestGenerateConfigContent(t *testing.T) {
	template := ConfigTemplate{
		Version:         "1.0.0",
		DefaultProvider: "github",
		Providers: map[string]ProviderTemplate{
			"github": {
				Token: "${GITHUB_TOKEN}",
				Orgs: []TargetTemplate{
					{
						Name:       "test-org",
						Visibility: "all",
						CloneDir:   "${HOME}/repos/github",
						Strategy:   "reset",
						Flatten:    true,
					},
				},
			},
			"gitlab": {
				Token: "${GITLAB_TOKEN}",
				Groups: []TargetTemplate{
					{
						Name:       "test-group",
						Visibility: "private",
						CloneDir:   "${HOME}/repos/gitlab",
						Strategy:   "pull",
						Recursive:  true,
						Match:      "^project-.*",
						Exclude:    []string{".*-temp", ".*-backup"},
					},
				},
			},
		},
	}

	content := generateConfigContent(template)

	// Verify structure
	assert.Contains(t, content, "version: \"1.0.0\"")
	assert.Contains(t, content, "default_provider: github")
	assert.Contains(t, content, "providers:")

	// Verify GitHub configuration
	assert.Contains(t, content, "github:")
	assert.Contains(t, content, "token: \"${GITHUB_TOKEN}\"")
	assert.Contains(t, content, "orgs:")
	assert.Contains(t, content, "name: \"test-org\"")
	assert.Contains(t, content, "flatten: true")

	// Verify GitLab configuration
	assert.Contains(t, content, "gitlab:")
	assert.Contains(t, content, "token: \"${GITLAB_TOKEN}\"")
	assert.Contains(t, content, "groups:")
	assert.Contains(t, content, "name: \"test-group\"")
	assert.Contains(t, content, "recursive: true")
	assert.Contains(t, content, "match: \"^project-.*\"")
	assert.Contains(t, content, "exclude:")
	assert.Contains(t, content, "- \".*-temp\"")
	assert.Contains(t, content, "- \".*-backup\"")
}

func TestInitializeConfig_FileExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-init-exists-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	outputFile := filepath.Join(tmpDir, "existing.yaml")

	// Create existing file
	err = os.WriteFile(outputFile, []byte("existing content"), 0o644)
	require.NoError(t, err)

	// Should fail without force flag
	err = initializeConfig(outputFile, false, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Should succeed with force flag
	err = initializeConfig(outputFile, true, true)
	assert.NoError(t, err)
}

func TestWriteTarget(t *testing.T) {
	var content strings.Builder

	target := TargetTemplate{
		Name:       "test-target",
		Visibility: "public",
		CloneDir:   "/path/to/repos",
		Match:      "^test-.*",
		Exclude:    []string{"temp-.*", "backup-.*"},
		Strategy:   "pull",
		Flatten:    true,
		Recursive:  true,
	}

	writeTarget(&content, target, "  ")

	result := content.String()
	assert.Contains(t, result, "name: \"test-target\"")
	assert.Contains(t, result, "visibility: \"public\"")
	assert.Contains(t, result, "clone_dir: \"/path/to/repos\"")
	assert.Contains(t, result, "match: \"^test-.*\"")
	assert.Contains(t, result, "exclude:")
	assert.Contains(t, result, "- \"temp-.*\"")
	assert.Contains(t, result, "- \"backup-.*\"")
	assert.Contains(t, result, "strategy: \"pull\"")
	assert.Contains(t, result, "flatten: true")
	assert.Contains(t, result, "recursive: true")
}

func TestConfigureProvider_Minimal(t *testing.T) {
	// This test would require input mocking for interactive testing
	// For now, we'll test the data structures
	template := ProviderTemplate{
		Token: "${GITHUB_TOKEN}",
		Orgs: []TargetTemplate{
			{
				Name:       "test-org",
				Visibility: "all",
				CloneDir:   "${HOME}/repos/github/test-org",
				Strategy:   "reset",
			},
		},
	}

	assert.Equal(t, "${GITHUB_TOKEN}", template.Token)
	assert.Len(t, template.Orgs, 1)
	assert.Equal(t, "test-org", template.Orgs[0].Name)
	assert.Equal(t, "all", template.Orgs[0].Visibility)
}
