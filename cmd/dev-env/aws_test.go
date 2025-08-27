//nolint:testpackage // White-box testing needed for internal function access
package devenv

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultAwsOptions(t *testing.T) {
	baseCmd := NewBaseCommand(
		"aws",
		"config",
		".aws/config",
		"AWS config management",
		[]string{"example"},
	)
	opts := baseCmd.DefaultOptions()

	assert.NotEmpty(t, opts.ConfigPath)
	assert.NotEmpty(t, opts.StorePath)
	assert.False(t, opts.Force)
	assert.False(t, opts.ListAll)
}

func TestNewAwsCmd(t *testing.T) {
	cmd := newAwsCmd()

	assert.Equal(t, "aws", cmd.Use)
	assert.Equal(t, "Manage Aws configuration files", cmd.Short)
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
	baseCmd := NewBaseCommand(
		"aws",
		"config",
		".aws/config",
		"AWS config management",
		[]string{"example"},
	)
	cmd := baseCmd.CreateSaveCommand()

	assert.Equal(t, "save", cmd.Use)
	assert.Equal(t, "Save current aws configuration", cmd.Short)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("name"))
	assert.NotNil(t, cmd.Flags().Lookup("description"))
	assert.NotNil(t, cmd.Flags().Lookup("config-path"))
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
	assert.NotNil(t, cmd.Flags().Lookup("force"))
}

func TestAwsLoadCmd(t *testing.T) {
	baseCmd := NewBaseCommand(
		"aws",
		"config",
		".aws/config",
		"AWS config management",
		[]string{"example"},
	)
	cmd := baseCmd.CreateLoadCommand()

	assert.Equal(t, "load", cmd.Use)
	assert.Equal(t, "Load saved aws configuration", cmd.Short)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("name"))
	assert.NotNil(t, cmd.Flags().Lookup("config-path"))
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
	assert.NotNil(t, cmd.Flags().Lookup("force"))
}

func TestAwsListCmd(t *testing.T) {
	baseCmd := NewBaseCommand(
		"aws",
		"config",
		".aws/config",
		"AWS config management",
		[]string{"example"},
	)
	cmd := baseCmd.CreateListCommand()

	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List saved aws configurations", cmd.Short)

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
		baseCmd := NewBaseCommand(
			"aws",
			"config",
			".aws/config",
			"AWS config management",
			[]string{"example"},
		)
		opts := &BaseOptions{
			Name:        "test-config",
			Description: "Test AWS configuration",
			ConfigPath:  configPath,
			StorePath:   storeDir,
			Force:       false,
		}

		err := baseCmd.SaveConfig(opts)
		assert.NoError(t, err)

		// Check if file was saved
		savedPath := filepath.Join(storeDir, "test-config.config")
		assert.FileExists(t, savedPath)

		// Check if metadata was saved
		metadataPath := filepath.Join(storeDir, "test-config.metadata.json")
		assert.FileExists(t, metadataPath)

		// Verify saved content
		savedContent, err := os.ReadFile(savedPath)
		require.NoError(t, err)
		assert.Equal(t, testConfigContent, string(savedContent))
	})

	t.Run("load AWS config", func(t *testing.T) {
		// Create a different target path
		targetPath := filepath.Join(tempDir, "loaded", "config")

		baseCmd := NewBaseCommand(
			"aws",
			"config",
			".aws/config",
			"AWS config management",
			[]string{"example"},
		)
		opts := &BaseOptions{
			Name:       "test-config",
			ConfigPath: targetPath,
			StorePath:  storeDir,
			Force:      true, // Skip backup for test
		}

		err := baseCmd.LoadConfig(opts)
		assert.NoError(t, err)

		// Check if file was loaded
		assert.FileExists(t, targetPath)

		// Check content matches
		loadedContent, err := os.ReadFile(targetPath)
		require.NoError(t, err)
		assert.Equal(t, testConfigContent, string(loadedContent))
	})

	t.Run("list AWS configs", func(t *testing.T) {
		baseCmd := NewBaseCommand(
			"aws",
			"config",
			".aws/config",
			"AWS config management",
			[]string{"example"},
		)
		opts := &BaseOptions{
			StorePath: storeDir,
		}

		err := baseCmd.ListConfigs(opts)
		assert.NoError(t, err)
	})
}

func TestAwsMetadata(t *testing.T) {
	tempDir := t.TempDir()

	baseCmd := NewBaseCommand(
		"aws",
		"config",
		".aws/config",
		"AWS config management",
		[]string{"example"},
	)

	metadata := ConfigMetadata{
		Description: "Test description",
		SavedAt:     time.Now(),
		SourcePath:  "/test/path",
	}

	metadataFile := filepath.Join(tempDir, "test-metadata.metadata.json")

	// Test save metadata
	err := baseCmd.saveMetadata(metadataFile, metadata)
	assert.NoError(t, err)
	assert.FileExists(t, metadataFile)

	// Test load metadata
	loadedMetadata, err := baseCmd.loadMetadata(metadataFile)
	assert.NoError(t, err)
	assert.Equal(t, "Test description", loadedMetadata.Description)
	assert.Equal(t, "/test/path", loadedMetadata.SourcePath)
	assert.False(t, loadedMetadata.SavedAt.IsZero())
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
	baseCmd := NewBaseCommand(
		"aws",
		"config",
		".aws/config",
		"AWS config management",
		[]string{"example"},
	)
	dstPath := filepath.Join(tempDir, "destination.config")

	err = baseCmd.copyFile(srcPath, dstPath)
	assert.NoError(t, err)

	// Check content
	copiedContent, err := os.ReadFile(dstPath)
	require.NoError(t, err)
	assert.Equal(t, content, string(copiedContent))
}

// TestAwsConfigFileExistence tests basic config file operations
func TestAwsConfigFileExistence(t *testing.T) {
	tempDir := t.TempDir()

	// Create test config
	testConfigContent := `[default]
region = us-west-2
output = json

[profile dev]
region = us-east-1
output = table
`

	configPath := filepath.Join(tempDir, "config")
	err := os.WriteFile(configPath, []byte(testConfigContent), 0o644)
	require.NoError(t, err)

	// Test file exists and can be read
	assert.FileExists(t, configPath)

	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "[default]")
	assert.Contains(t, string(content), "[profile dev]")
}

func TestAwsErrorCases(t *testing.T) {
	tempDir := t.TempDir()
	baseCmd := NewBaseCommand(
		"aws",
		"config",
		".aws/config",
		"AWS config management",
		[]string{"example"},
	)

	t.Run("save non-existent config", func(t *testing.T) {
		opts := &BaseOptions{
			Name:       "test",
			ConfigPath: "/non/existent/path",
			StorePath:  tempDir,
		}

		err := baseCmd.SaveConfig(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config file not found")
	})

	t.Run("load non-existent config", func(t *testing.T) {
		opts := &BaseOptions{
			Name:      "non-existent",
			StorePath: tempDir,
		}

		err := baseCmd.LoadConfig(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("save duplicate without force", func(t *testing.T) {
		// Create a test config
		configPath := filepath.Join(tempDir, "config")
		err := os.WriteFile(configPath, []byte("[default]\nregion=us-west-2"), 0o644)
		require.NoError(t, err)

		opts := &BaseOptions{
			Name:       "duplicate-test",
			ConfigPath: configPath,
			StorePath:  tempDir,
			Force:      false,
		}

		// Save first time
		err = baseCmd.SaveConfig(opts)
		assert.NoError(t, err)

		// Try to save again without force
		err = baseCmd.SaveConfig(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}
