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

func TestDefaultDockerOptions(t *testing.T) {
	baseCmd := NewBaseCommand(
		"docker",
		"json",
		".docker/config.json",
		"Docker config management",
		[]string{"example"},
	)
	opts := baseCmd.DefaultOptions()

	assert.NotEmpty(t, opts.ConfigPath)
	assert.NotEmpty(t, opts.StorePath)
	assert.False(t, opts.Force)
	assert.False(t, opts.ListAll)
}

func TestNewDockerCmd(t *testing.T) {
	cmd := newDockerCmd()

	assert.Equal(t, "docker", cmd.Use)
	assert.Equal(t, "Manage Docker configuration files", cmd.Short)
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

func TestDockerSaveCmd(t *testing.T) {
	baseCmd := NewBaseCommand(
		"docker",
		"json",
		".docker/config.json",
		"Docker config management",
		[]string{"example"},
	)
	cmd := baseCmd.CreateSaveCommand()

	assert.Equal(t, "save", cmd.Use)
	assert.Equal(t, "Save current docker configuration", cmd.Short)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("name"))
	assert.NotNil(t, cmd.Flags().Lookup("description"))
	assert.NotNil(t, cmd.Flags().Lookup("config-path"))
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
	assert.NotNil(t, cmd.Flags().Lookup("force"))
}

func TestDockerLoadCmd(t *testing.T) {
	baseCmd := NewBaseCommand(
		"docker",
		"json",
		".docker/config.json",
		"Docker config management",
		[]string{"example"},
	)
	cmd := baseCmd.CreateLoadCommand()

	assert.Equal(t, "load", cmd.Use)
	assert.Equal(t, "Load saved docker configuration", cmd.Short)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("name"))
	assert.NotNil(t, cmd.Flags().Lookup("config-path"))
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
	assert.NotNil(t, cmd.Flags().Lookup("force"))
}

func TestDockerListCmd(t *testing.T) {
	baseCmd := NewBaseCommand(
		"docker",
		"json",
		".docker/config.json",
		"Docker config management",
		[]string{"example"},
	)
	cmd := baseCmd.CreateListCommand()

	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List saved docker configurations", cmd.Short)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
}

func TestDockerSaveLoad(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	storeDir := filepath.Join(tempDir, "store")

	err := os.MkdirAll(configDir, 0o755)
	require.NoError(t, err)

	// Create a test Docker config file
	testConfigContent := `{
  "auths": {
    "https://index.docker.io/v1/": {
      "auth": "dGVzdDp0ZXN0"
    },
    "myregistry.com": {
      "auth": "dXNlcjpwYXNz",
      "email": "user@example.com"
    }
  },
  "credsStore": "desktop",
  "credHelpers": {
    "gcr.io": "gcloud"
  },
  "experimental": "enabled"
}`

	configPath := filepath.Join(configDir, "config.json")
	err = os.WriteFile(configPath, []byte(testConfigContent), 0o644)
	require.NoError(t, err)

	t.Run("save Docker config", func(t *testing.T) {
		baseCmd := NewBaseCommand(
			"docker",
			"json",
			".docker/config.json",
			"Docker config management",
			[]string{"example"},
		)
		opts := &BaseOptions{
			Name:        "test-config",
			Description: "Test Docker configuration",
			ConfigPath:  configPath,
			StorePath:   storeDir,
			Force:       false,
		}

		err := baseCmd.SaveConfig(opts)
		assert.NoError(t, err)

		// Check if file was saved
		savedPath := filepath.Join(storeDir, "test-config.json")
		assert.FileExists(t, savedPath)

		// Check if metadata was saved
		metadataPath := filepath.Join(storeDir, "test-config.metadata.json")
		assert.FileExists(t, metadataPath)

		// Verify saved content
		savedContent, err := os.ReadFile(savedPath)
		require.NoError(t, err)
		assert.JSONEq(t, testConfigContent, string(savedContent))
	})

	t.Run("load Docker config", func(t *testing.T) {
		// Create a different target path
		targetPath := filepath.Join(tempDir, "loaded", "config.json")

		baseCmd := NewBaseCommand(
			"docker",
			"json",
			".docker/config.json",
			"Docker config management",
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
		assert.JSONEq(t, testConfigContent, string(loadedContent))
	})

	t.Run("list Docker configs", func(t *testing.T) {
		baseCmd := NewBaseCommand(
			"docker",
			"json",
			".docker/config.json",
			"Docker config management",
			[]string{"example"},
		)
		opts := &BaseOptions{
			StorePath: storeDir,
		}

		err := baseCmd.ListConfigs(opts)
		assert.NoError(t, err)
	})
}

func TestDockerMetadata(t *testing.T) {
	tempDir := t.TempDir()

	baseCmd := NewBaseCommand(
		"docker",
		"json",
		".docker/config.json",
		"Docker config management",
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

func TestDockerCopyFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create source file
	srcPath := filepath.Join(tempDir, "source.json")
	content := `{"test": "content"}`
	err := os.WriteFile(srcPath, []byte(content), 0o644)
	require.NoError(t, err)

	// Test copy
	baseCmd := NewBaseCommand(
		"docker",
		"json",
		".docker/config.json",
		"Docker config management",
		[]string{"example"},
	)
	dstPath := filepath.Join(tempDir, "destination.json")

	err = baseCmd.copyFile(srcPath, dstPath)
	assert.NoError(t, err)

	// Check content
	copiedContent, err := os.ReadFile(dstPath)
	require.NoError(t, err)
	assert.Equal(t, content, string(copiedContent))
}

// TestDockerConfigFileExistence tests basic config file operations.
func TestDockerConfigFileExistence(t *testing.T) {
	tempDir := t.TempDir()

	// Create test config
	testConfigContent := `{
  "auths": {
    "https://index.docker.io/v1/": {
      "auth": "dGVzdDp0ZXN0"
    }
  },
  "credsStore": "desktop"
}`

	configPath := filepath.Join(tempDir, "config.json")
	err := os.WriteFile(configPath, []byte(testConfigContent), 0o644)
	require.NoError(t, err)

	// Test file exists and can be read
	assert.FileExists(t, configPath)

	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "auths")
	assert.Contains(t, string(content), "credsStore")
}

func TestDockerErrorCases(t *testing.T) {
	tempDir := t.TempDir()
	baseCmd := NewBaseCommand(
		"docker",
		"json",
		".docker/config.json",
		"Docker config management",
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
		configPath := filepath.Join(tempDir, "config.json")
		err := os.WriteFile(configPath, []byte(`{"test":"config"}`), 0o644)
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
