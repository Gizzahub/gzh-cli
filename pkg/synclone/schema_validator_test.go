//nolint:testpackage // White-box testing needed for internal function access
package bulkclone

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateConfigWithSchema(t *testing.T) {
	t.Run("valid simple config", func(t *testing.T) {
		err := ValidateConfigWithSchema("../../examples/bulk-clone/bulk-clone-simple.yaml")
		assert.NoError(t, err, "Simple config should be valid against schema")
	})

	t.Run("valid comprehensive config", func(t *testing.T) {
		err := ValidateConfigWithSchema("../../examples/bulk-clone/bulk-clone-example.yaml")
		assert.NoError(t, err, "Example config should be valid against schema")
	})

	t.Run("invalid config - missing version", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "invalid.yaml")

		invalidConfig := `
default:
  protocol: https
repo_roots:
  - root_path: "/tmp"
    provider: "github"
    protocol: "https"
    org_name: "test"
`
		err := os.WriteFile(configPath, []byte(invalidConfig), 0o644)
		require.NoError(t, err)

		err = ValidateConfigWithSchema(configPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "version")
	})

	t.Run("invalid config - wrong protocol", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "invalid-protocol.yaml")

		invalidConfig := `
version: "1.0"
default:
  protocol: ftp  # Invalid protocol
`
		err := os.WriteFile(configPath, []byte(invalidConfig), 0o644)
		require.NoError(t, err)

		err = ValidateConfigWithSchema(configPath)
		assert.Error(t, err)
	})

	t.Run("invalid config - missing required fields in repo_root", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "invalid-repo.yaml")

		invalidConfig := `
version: "1.0"
repo_roots:
  - root_path: "/tmp"
    provider: "github"
    # Missing protocol and org_name
`
		err := os.WriteFile(configPath, []byte(invalidConfig), 0o644)
		require.NoError(t, err)

		err = ValidateConfigWithSchema(configPath)
		assert.Error(t, err)
	})

	t.Run("invalid config - wrong provider", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "invalid-provider.yaml")

		invalidConfig := `
version: "1.0"
repo_roots:
  - root_path: "/tmp"
    provider: "bitbucket"  # Not supported
    protocol: "https"
    org_name: "test"
`
		err := os.WriteFile(configPath, []byte(invalidConfig), 0o644)
		require.NoError(t, err)

		err = ValidateConfigWithSchema(configPath)
		assert.Error(t, err)
	})
}

func TestLoadSchemaFromFile(t *testing.T) {
	schema, err := LoadSchemaFromFile()
	assert.NoError(t, err)
	assert.NotEmpty(t, schema)
	assert.Contains(t, schema, `"$schema"`)
	assert.Contains(t, schema, "Bulk Clone Configuration Schema")
}

func TestConfigToJSON(t *testing.T) {
	cfg := &bulkCloneConfig{
		Version: "1.0",
		Default: bulkCloneDefault{
			Protocol: "https",
		},
		RepoRoots: []BulkCloneGithub{
			{
				RootPath: "/tmp",
				Provider: "github",
				Protocol: "ssh",
				OrgName:  "test",
			},
		},
		IgnoreNameRegexes: []string{"test-.*"},
	}

	jsonData, err := configToJSON(cfg)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Check that the JSON contains expected fields
	jsonStr := string(jsonData)
	assert.Contains(t, jsonStr, `"version":"1.0"`)
	assert.Contains(t, jsonStr, `"protocol":"https"`)
	assert.Contains(t, jsonStr, `"org_name":"test"`)
}
