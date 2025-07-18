package devenv

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultDockerOptions(t *testing.T) {
	opts := defaultDockerOptions()

	assert.NotEmpty(t, opts.configPath)
	assert.NotEmpty(t, opts.storePath)
	assert.False(t, opts.force)
	assert.False(t, opts.listAll)
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
	cmd := newDockerSaveCmd()

	assert.Equal(t, "save", cmd.Use)
	assert.Equal(t, "Save current Docker config", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("name"))
	assert.NotNil(t, cmd.Flags().Lookup("description"))
	assert.NotNil(t, cmd.Flags().Lookup("config-path"))
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
	assert.NotNil(t, cmd.Flags().Lookup("force"))
}

func TestDockerLoadCmd(t *testing.T) {
	cmd := newDockerLoadCmd()

	assert.Equal(t, "load", cmd.Use)
	assert.Equal(t, "Load a saved Docker config", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("name"))
	assert.NotNil(t, cmd.Flags().Lookup("config-path"))
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
	assert.NotNil(t, cmd.Flags().Lookup("force"))
}

func TestDockerListCmd(t *testing.T) {
	cmd := newDockerListCmd()

	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List saved Docker configs", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

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
	testConfig := dockerConfig{
		Auths: map[string]interface{}{
			"https://index.docker.io/v1/": map[string]string{
				"auth": "dGVzdDp0ZXN0",
			},
			"myregistry.com": map[string]string{
				"auth":  "dXNlcjpwYXNz",
				"email": "user@example.com",
			},
		},
		CredsStore: "desktop",
		CredHelpers: map[string]string{
			"gcr.io": "gcloud",
		},
		Experimental: "enabled",
	}

	configData, err := json.MarshalIndent(testConfig, "", "  ")
	require.NoError(t, err)

	configPath := filepath.Join(configDir, "config.json")
	err = os.WriteFile(configPath, configData, 0o644)
	require.NoError(t, err)

	t.Run("save Docker config", func(t *testing.T) {
		opts := &dockerOptions{
			name:        "test-config",
			description: "Test Docker configuration",
			configPath:  configPath,
			storePath:   storeDir,
			force:       false,
		}

		err := opts.runSave(nil, nil)
		assert.NoError(t, err)

		// Check if file was saved
		savedPath := filepath.Join(storeDir, "test-config.json")
		assert.FileExists(t, savedPath)

		// Check if metadata was saved
		metadataPath := filepath.Join(storeDir, "test-config.meta")
		assert.FileExists(t, metadataPath)

		// Verify saved content
		savedContent, err := os.ReadFile(savedPath)
		require.NoError(t, err)
		assert.JSONEq(t, string(configData), string(savedContent))
	})

	t.Run("load Docker config", func(t *testing.T) {
		// Create a different target path
		targetPath := filepath.Join(tempDir, "loaded", "config.json")

		opts := &dockerOptions{
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
		assert.JSONEq(t, string(configData), string(loadedContent))
	})

	t.Run("list Docker configs", func(t *testing.T) {
		opts := &dockerOptions{
			storePath: storeDir,
		}

		err := opts.runList(nil, nil)
		assert.NoError(t, err)
	})
}

func TestDockerMetadata(t *testing.T) {
	tempDir := t.TempDir()

	opts := &dockerOptions{
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

func TestDockerCopyFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create source file
	srcPath := filepath.Join(tempDir, "source.json")
	content := `{"test": "content"}`
	err := os.WriteFile(srcPath, []byte(content), 0o644)
	require.NoError(t, err)

	// Test copy
	opts := &dockerOptions{}
	dstPath := filepath.Join(tempDir, "destination.json")

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

func TestDockerDisplayConfigInfo(t *testing.T) {
	tempDir := t.TempDir()

	// Create test config
	testConfig := dockerConfig{
		Auths: map[string]interface{}{
			"https://index.docker.io/v1/": map[string]string{
				"auth": "dGVzdDp0ZXN0",
			},
			"myregistry.com": map[string]string{
				"auth": "dXNlcjpwYXNz",
			},
		},
		CredsStore: "desktop",
		CredHelpers: map[string]string{
			"gcr.io": "gcloud",
		},
	}

	configData, err := json.MarshalIndent(testConfig, "", "  ")
	require.NoError(t, err)

	configPath := filepath.Join(tempDir, "config.json")
	err = os.WriteFile(configPath, configData, 0o644)
	require.NoError(t, err)

	opts := &dockerOptions{}
	err = opts.displayConfigInfo(configPath)
	assert.NoError(t, err)
}

func TestDockerErrorCases(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("save non-existent config", func(t *testing.T) {
		opts := &dockerOptions{
			name:       "test",
			configPath: "/non/existent/path",
			storePath:  tempDir,
		}

		err := opts.runSave(nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Docker config file not found")
	})

	t.Run("load non-existent config", func(t *testing.T) {
		opts := &dockerOptions{
			name:      "non-existent",
			storePath: tempDir,
		}

		err := opts.runLoad(nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("save duplicate without force", func(t *testing.T) {
		// Create a test config
		configPath := filepath.Join(tempDir, "config.json")
		err := os.WriteFile(configPath, []byte(`{"test":"config"}`), 0o644)
		require.NoError(t, err)

		opts := &dockerOptions{
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

	t.Run("display info for invalid config", func(t *testing.T) {
		invalidConfigPath := filepath.Join(tempDir, "invalid.json")
		err := os.WriteFile(invalidConfigPath, []byte("invalid json"), 0o644)
		require.NoError(t, err)

		opts := &dockerOptions{}
		err = opts.displayConfigInfo(invalidConfigPath)
		assert.Error(t, err)
	})
}
