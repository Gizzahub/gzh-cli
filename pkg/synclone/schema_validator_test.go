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
	// 배포되는 예시가 스키마와 로더 양쪽을 다 통과해야 한다. 이 둘이 갈라진
	// 적이 있다 -- 스키마는 snake_case를 요구하는데 구조체 태그는 camelCase여서,
	// 스키마 검사는 통과하고 로더는 내용을 통째로 버렸다. 그래서 gz synclone
	// validate가 읽지도 못한 설정을 "valid"라고 답했다.
	t.Run("valid comprehensive config", func(t *testing.T) {
		err := ValidateConfigWithSchema(legacyExampleConfig)
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
