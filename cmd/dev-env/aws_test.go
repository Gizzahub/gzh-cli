//nolint:testpackage // White-box testing needed for internal function access
package devenv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestDefaultAwsOptions(t *testing.T) {
	opts := defaultAwsOptions()

	assert.NotEmpty(t, opts.configPath)
	assert.NotEmpty(t, opts.storePath)
	assert.False(t, opts.force)
	assert.False(t, opts.listAll)
}

func TestNewAwsCmd(t *testing.T) {
	cmd := newAwsCmd()

	assert.Equal(t, "aws", cmd.Use)
	assert.Equal(t, "Manage AWS configuration files", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check subcommands
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 3)

	var saveCmd, loadCmd, listCmd bool

	for _, subcmd := range subcommands {
		switch subcmd.Use {
		case "save":
			saveCmd = true
		case "load":
			loadCmd = true
		case "list":
			listCmd = true
		}
	}

	assert.True(t, saveCmd, "save subcommand should exist")
	assert.True(t, loadCmd, "load subcommand should exist")
	assert.True(t, listCmd, "list subcommand should exist")
}

func TestAwsSaveCmd(t *testing.T) {
	cmd := newAwsSaveCmd()

	assert.Equal(t, "save", cmd.Use)
	assert.Equal(t, "Save current AWS config", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("name"))
	assert.NotNil(t, cmd.Flags().Lookup("description"))
	assert.NotNil(t, cmd.Flags().Lookup("config-path"))
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
	assert.NotNil(t, cmd.Flags().Lookup("force"))
}

func TestAwsLoadCmd(t *testing.T) {
	cmd := newAwsLoadCmd()

	assert.Equal(t, "load", cmd.Use)
	assert.Equal(t, "Load a saved AWS config", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("name"))
	assert.NotNil(t, cmd.Flags().Lookup("config-path"))
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
	assert.NotNil(t, cmd.Flags().Lookup("force"))
}

func TestAwsListCmd(t *testing.T) {
	cmd := newAwsListCmd()

	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List saved AWS configs", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
}

func TestAwsSaveLoad(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	storeDir := filepath.Join(tempDir, "store")

	err := os.MkdirAll(configDir, 0o755)
	require.NoError(t, err)

	// Create a test AWS config file
	testConfigContent := `[default]
region = us-west-2
output = json

[profile dev]
region = us-east-1
output = table

[profile production]
region = eu-west-1
output = json
sso_start_url = https://example.awsapps.com/start
sso_region = us-east-1
sso_account_id = 123456789012
sso_role_name = PowerUserAccess
`

	configPath := filepath.Join(configDir, "config")
	err = os.WriteFile(configPath, []byte(testConfigContent), 0o644)
	require.NoError(t, err)

	t.Run("save AWS config", func(t *testing.T) {
		opts := &awsOptions{
			name:        "test-config",
			description: "Test AWS configuration",
			configPath:  configPath,
			storePath:   storeDir,
			force:       false,
		}

		err := opts.runSave(nil, nil)
		assert.NoError(t, err)

		// Check if file was saved
		savedPath := filepath.Join(storeDir, "test-config.config")
		assert.FileExists(t, savedPath)

		// Check if metadata was saved
		metadataPath := filepath.Join(storeDir, "test-config.meta")
		assert.FileExists(t, metadataPath)

		// Verify saved content
		savedContent, err := os.ReadFile(savedPath)
		require.NoError(t, err)
		assert.Equal(t, testConfigContent, string(savedContent))
	})

	t.Run("load AWS config", func(t *testing.T) {
		// Create a different target path
		targetPath := filepath.Join(tempDir, "loaded", "config")

		opts := &awsOptions{
			name:       "test-config",
			configPath: targetPath,
			storePath:  storeDir,
			force:      true, // Skip backup for test
		}

		err := opts.runLoad(nil, nil)
		assert.NoError(t, err)

		// Check if file was loaded
		assert.FileExists(t, targetPath)

		// Check content matches
		loadedContent, err := os.ReadFile(targetPath)
		require.NoError(t, err)
		assert.Equal(t, testConfigContent, string(loadedContent))
	})

	t.Run("list AWS configs", func(t *testing.T) {
		opts := &awsOptions{
			storePath: storeDir,
		}

		err := opts.runList(nil, nil)
		assert.NoError(t, err)
	})
}

func TestAwsMetadata(t *testing.T) {
	tempDir := t.TempDir()

	opts := &awsOptions{
		name:        "test-metadata",
		description: "Test description",
		storePath:   tempDir,
	}

	// Test save metadata
	err := opts.saveMetadata()
	assert.NoError(t, err)

	metadataPath := filepath.Join(tempDir, "test-metadata.meta")
	assert.FileExists(t, metadataPath)

	// Test load metadata
	metadata := opts.loadMetadata("test-metadata")
	assert.Equal(t, "test-metadata", metadata.Name)
	assert.Equal(t, "Test description", metadata.Description)
	assert.False(t, metadata.SavedAt.IsZero())
}

func TestAwsCopyFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create source file
	srcPath := filepath.Join(tempDir, "source.config")
	content := `[default]
region = us-west-2
output = json`
	err := os.WriteFile(srcPath, []byte(content), 0o644)
	require.NoError(t, err)

	// Test copy
	opts := &awsOptions{}
	dstPath := filepath.Join(tempDir, "destination.config")

	err = opts.copyFile(srcPath, dstPath)
	assert.NoError(t, err)

	// Check content
	copiedContent, err := os.ReadFile(dstPath)
	require.NoError(t, err)
	assert.Equal(t, content, string(copiedContent))

	// Check permissions
	srcInfo, err := os.Stat(srcPath)
	require.NoError(t, err)
	dstInfo, err := os.Stat(dstPath)
	require.NoError(t, err)
	assert.Equal(t, srcInfo.Mode(), dstInfo.Mode())
}

func TestAwsDisplayConfigInfo(t *testing.T) {
	tempDir := t.TempDir()

	// Create test config
	testConfigContent := `[default]
region = us-west-2
output = json

[profile dev]
region = us-east-1
output = table

[profile production]
region = eu-west-1
output = json
`

	configPath := filepath.Join(tempDir, "config")
	err := os.WriteFile(configPath, []byte(testConfigContent), 0o644)
	require.NoError(t, err)

	opts := &awsOptions{}
	err = opts.displayConfigInfo(configPath)
	assert.NoError(t, err)
}

func TestAwsParseConfig(t *testing.T) {
	opts := &awsOptions{}

	testConfig := `[default]
region = us-west-2
output = json

[profile dev]
region = us-east-1
output = table

[profile production]
region = eu-west-1
output = json
sso_start_url = https://example.awsapps.com/start
`

	profiles := opts.parseAwsConfig(testConfig)

	assert.Len(t, profiles, 3)

	// Check default profile
	defaultProfile := profiles[0]
	assert.Equal(t, "default", defaultProfile.Name)
	assert.Equal(t, "us-west-2", defaultProfile.Region)
	assert.Equal(t, "json", defaultProfile.Output)

	// Check dev profile
	devProfile := profiles[1]
	assert.Equal(t, "dev", devProfile.Name)
	assert.Equal(t, "us-east-1", devProfile.Region)
	assert.Equal(t, "table", devProfile.Output)

	// Check production profile
	prodProfile := profiles[2]
	assert.Equal(t, "production", prodProfile.Name)
	assert.Equal(t, "eu-west-1", prodProfile.Region)
	assert.Equal(t, "json", prodProfile.Output)
}

func TestAwsErrorCases(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("save non-existent config", func(t *testing.T) {
		opts := &awsOptions{
			name:       "test",
			configPath: "/non/existent/path",
			storePath:  tempDir,
		}

		err := opts.runSave(nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "AWS config file not found")
	})

	t.Run("load non-existent config", func(t *testing.T) {
		opts := &awsOptions{
			name:      "non-existent",
			storePath: tempDir,
		}

		err := opts.runLoad(nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("save duplicate without force", func(t *testing.T) {
		// Create a test config
		configPath := filepath.Join(tempDir, "config")
		err := os.WriteFile(configPath, []byte("[default]\nregion=us-west-2"), 0o644)
		require.NoError(t, err)

		opts := &awsOptions{
			name:       "duplicate-test",
			configPath: configPath,
			storePath:  tempDir,
			force:      false,
		}

		// Save first time
		err = opts.runSave(nil, nil)
		assert.NoError(t, err)

		// Try to save again without force
		err = opts.runSave(nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}
