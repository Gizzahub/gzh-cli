package devenv

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultGcloudOptions(t *testing.T) {
	opts := defaultGcloudOptions()

	assert.NotEmpty(t, opts.configPath)
	assert.NotEmpty(t, opts.storePath)
	assert.False(t, opts.force)
	assert.False(t, opts.listAll)
}

func TestNewGcloudCmd(t *testing.T) {
	cmd := newGcloudCmd()

	assert.Equal(t, "gcloud", cmd.Use)
	assert.Equal(t, "Manage Google Cloud configuration", cmd.Short)
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

func TestGcloudSaveCmd(t *testing.T) {
	cmd := newGcloudSaveCmd()

	assert.Equal(t, "save", cmd.Use)
	assert.Equal(t, "Save current gcloud config", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("name"))
	assert.NotNil(t, cmd.Flags().Lookup("description"))
	assert.NotNil(t, cmd.Flags().Lookup("config-path"))
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
	assert.NotNil(t, cmd.Flags().Lookup("force"))
}

func TestGcloudLoadCmd(t *testing.T) {
	cmd := newGcloudLoadCmd()

	assert.Equal(t, "load", cmd.Use)
	assert.Equal(t, "Load a saved gcloud config", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("name"))
	assert.NotNil(t, cmd.Flags().Lookup("config-path"))
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
	assert.NotNil(t, cmd.Flags().Lookup("force"))
}

func TestGcloudListCmd(t *testing.T) {
	cmd := newGcloudListCmd()

	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List saved gcloud configs", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
}

func TestGcloudSaveLoad(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	storeDir := filepath.Join(tempDir, "store")

	// Create a test gcloud config directory structure
	err := os.MkdirAll(configDir, 0o755)
	require.NoError(t, err)

	// Create active_config file
	activeConfigContent := "default"
	err = os.WriteFile(filepath.Join(configDir, "active_config"), []byte(activeConfigContent), 0o644)
	require.NoError(t, err)

	// Create configurations directory structure
	configsDir := filepath.Join(configDir, "configurations")
	err = os.MkdirAll(configsDir, 0o755)
	require.NoError(t, err)

	// Create default configuration
	defaultConfigDir := filepath.Join(configsDir, "default")
	err = os.MkdirAll(defaultConfigDir, 0o755)
	require.NoError(t, err)

	// Create properties file (JSON format)
	properties := map[string]interface{}{
		"core": map[string]interface{}{
			"project": "my-test-project",
			"account": "test@example.com",
		},
		"compute": map[string]interface{}{
			"region": "us-central1",
			"zone":   "us-central1-a",
		},
	}
	propertiesData, err := json.MarshalIndent(properties, "", "  ")
	require.NoError(t, err)

	propertiesPath := filepath.Join(defaultConfigDir, "properties")
	err = os.WriteFile(propertiesPath, propertiesData, 0o644)
	require.NoError(t, err)

	// Create dev configuration
	devConfigDir := filepath.Join(configsDir, "dev")
	err = os.MkdirAll(devConfigDir, 0o755)
	require.NoError(t, err)

	devProperties := map[string]interface{}{
		"core": map[string]interface{}{
			"project": "my-dev-project",
			"account": "dev@example.com",
		},
		"compute": map[string]interface{}{
			"region": "us-west1",
		},
	}
	devPropertiesData, err := json.MarshalIndent(devProperties, "", "  ")
	require.NoError(t, err)

	devPropertiesPath := filepath.Join(devConfigDir, "properties")
	err = os.WriteFile(devPropertiesPath, devPropertiesData, 0o644)
	require.NoError(t, err)

	t.Run("save gcloud config", func(t *testing.T) {
		opts := &gcloudOptions{
			name:        "test-config",
			description: "Test gcloud configuration",
			configPath:  configDir,
			storePath:   storeDir,
			force:       false,
		}

		err := opts.runSave(nil, nil)
		assert.NoError(t, err)

		// Check if directory was saved
		savedPath := filepath.Join(storeDir, "test-config")
		assert.DirExists(t, savedPath)

		// Check if metadata was saved
		metadataPath := filepath.Join(storeDir, "test-config.meta")
		assert.FileExists(t, metadataPath)

		// Verify active_config was copied
		savedActiveConfigPath := filepath.Join(savedPath, "active_config")
		assert.FileExists(t, savedActiveConfigPath)
		savedActiveConfig, err := os.ReadFile(savedActiveConfigPath)
		require.NoError(t, err)
		assert.Equal(t, activeConfigContent, string(savedActiveConfig))

		// Verify configurations were copied
		savedConfigsPath := filepath.Join(savedPath, "configurations", "default", "properties")
		assert.FileExists(t, savedConfigsPath)
	})

	t.Run("load gcloud config", func(t *testing.T) {
		// Create a different target path
		targetPath := filepath.Join(tempDir, "loaded")

		opts := &gcloudOptions{
			name:       "test-config",
			configPath: targetPath,
			storePath:  storeDir,
			force:      true, // Skip backup for test
		}

		err := opts.runLoad(nil, nil)
		assert.NoError(t, err)

		// Check if directory was loaded
		assert.DirExists(t, targetPath)

		// Check if active_config was loaded
		loadedActiveConfigPath := filepath.Join(targetPath, "active_config")
		assert.FileExists(t, loadedActiveConfigPath)
		loadedActiveConfig, err := os.ReadFile(loadedActiveConfigPath)
		require.NoError(t, err)
		assert.Equal(t, activeConfigContent, string(loadedActiveConfig))

		// Check if configurations were loaded
		loadedConfigsPath := filepath.Join(targetPath, "configurations", "default", "properties")
		assert.FileExists(t, loadedConfigsPath)
	})

	t.Run("list gcloud configs", func(t *testing.T) {
		opts := &gcloudOptions{
			storePath: storeDir,
		}

		err := opts.runList(nil, nil)
		assert.NoError(t, err)
	})
}

func TestGcloudMetadata(t *testing.T) {
	tempDir := t.TempDir()

	opts := &gcloudOptions{
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

func TestGcloudCopyFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create source file
	srcPath := filepath.Join(tempDir, "source.txt")
	content := "test content"
	err := os.WriteFile(srcPath, []byte(content), 0o644)
	require.NoError(t, err)

	// Test copy
	opts := &gcloudOptions{}
	dstPath := filepath.Join(tempDir, "destination.txt")

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

func TestGcloudCopyDir(t *testing.T) {
	tempDir := t.TempDir()

	// Create source directory with files
	srcDir := filepath.Join(tempDir, "source")
	err := os.MkdirAll(filepath.Join(srcDir, "subdir"), 0o755)
	require.NoError(t, err)

	// Create files
	err = os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(srcDir, "subdir", "file2.txt"), []byte("content2"), 0o644)
	require.NoError(t, err)

	// Test copy directory
	opts := &gcloudOptions{}
	dstDir := filepath.Join(tempDir, "destination")

	err = opts.copyDir(srcDir, dstDir)
	assert.NoError(t, err)

	// Check copied files
	assert.FileExists(t, filepath.Join(dstDir, "file1.txt"))
	assert.FileExists(t, filepath.Join(dstDir, "subdir", "file2.txt"))

	// Check content
	content1, err := os.ReadFile(filepath.Join(dstDir, "file1.txt"))
	require.NoError(t, err)
	assert.Equal(t, "content1", string(content1))

	content2, err := os.ReadFile(filepath.Join(dstDir, "subdir", "file2.txt"))
	require.NoError(t, err)
	assert.Equal(t, "content2", string(content2))
}

func TestGcloudDisplayConfigInfo(t *testing.T) {
	tempDir := t.TempDir()

	// Create test gcloud config structure
	configDir := filepath.Join(tempDir, "gcloud")
	err := os.MkdirAll(configDir, 0o755)
	require.NoError(t, err)

	// Create active_config
	err = os.WriteFile(filepath.Join(configDir, "active_config"), []byte("default"), 0o644)
	require.NoError(t, err)

	// Create configurations
	configsDir := filepath.Join(configDir, "configurations", "default")
	err = os.MkdirAll(configsDir, 0o755)
	require.NoError(t, err)

	properties := map[string]interface{}{
		"core": map[string]interface{}{
			"project": "test-project",
			"account": "test@example.com",
		},
		"compute": map[string]interface{}{
			"region": "us-central1",
		},
	}
	propertiesData, err := json.MarshalIndent(properties, "", "  ")
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(configsDir, "properties"), propertiesData, 0o644)
	require.NoError(t, err)

	opts := &gcloudOptions{}
	err = opts.displayConfigInfo(configDir)
	assert.NoError(t, err)
}

func TestGcloudParseConfigurations(t *testing.T) {
	tempDir := t.TempDir()

	// Create configurations directory
	configsDir := filepath.Join(tempDir, "configurations")

	// Create default config
	defaultDir := filepath.Join(configsDir, "default")
	err := os.MkdirAll(defaultDir, 0o755)
	require.NoError(t, err)

	defaultProps := map[string]interface{}{
		"core": map[string]interface{}{
			"project": "default-project",
			"account": "default@example.com",
		},
		"compute": map[string]interface{}{
			"region": "us-central1",
			"zone":   "us-central1-a",
		},
	}
	defaultData, err := json.MarshalIndent(defaultProps, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(defaultDir, "properties"), defaultData, 0o644)
	require.NoError(t, err)

	// Create dev config
	devDir := filepath.Join(configsDir, "dev")
	err = os.MkdirAll(devDir, 0o755)
	require.NoError(t, err)

	devProps := map[string]interface{}{
		"core": map[string]interface{}{
			"project": "dev-project",
			"account": "dev@example.com",
		},
	}
	devData, err := json.MarshalIndent(devProps, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(devDir, "properties"), devData, 0o644)
	require.NoError(t, err)

	opts := &gcloudOptions{}
	configs, err := opts.parseGcloudConfigurations(configsDir)
	assert.NoError(t, err)
	assert.Len(t, configs, 2)

	// Check default config
	var defaultConfig, devConfig *gcloudConfiguration
	for i := range configs {
		if configs[i].Name == "default" {
			defaultConfig = &configs[i]
		} else if configs[i].Name == "dev" {
			devConfig = &configs[i]
		}
	}

	require.NotNil(t, defaultConfig)
	assert.Equal(t, "default-project", defaultConfig.Project)
	assert.Equal(t, "default@example.com", defaultConfig.Account)
	assert.Equal(t, "us-central1", defaultConfig.Region)
	assert.Equal(t, "us-central1-a", defaultConfig.Zone)

	require.NotNil(t, devConfig)
	assert.Equal(t, "dev-project", devConfig.Project)
	assert.Equal(t, "dev@example.com", devConfig.Account)
}

func TestGcloudParsePropertiesINI(t *testing.T) {
	tempDir := t.TempDir()

	// Create INI format properties file
	iniContent := `[core]
project = ini-project
account = ini@example.com

[compute]
region = europe-west1
zone = europe-west1-b
`

	propertiesPath := filepath.Join(tempDir, "properties")
	err := os.WriteFile(propertiesPath, []byte(iniContent), 0o644)
	require.NoError(t, err)

	opts := &gcloudOptions{}
	config := gcloudConfiguration{Name: "test"}
	err = opts.parseGcloudProperties(propertiesPath, &config)
	assert.NoError(t, err)

	assert.Equal(t, "ini-project", config.Project)
	assert.Equal(t, "ini@example.com", config.Account)
	assert.Equal(t, "europe-west1", config.Region)
	assert.Equal(t, "europe-west1-b", config.Zone)
}

func TestGcloudErrorCases(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("save non-existent config", func(t *testing.T) {
		opts := &gcloudOptions{
			name:       "test",
			configPath: "/non/existent/path",
			storePath:  tempDir,
		}

		err := opts.runSave(nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "gcloud config directory not found")
	})

	t.Run("load non-existent config", func(t *testing.T) {
		opts := &gcloudOptions{
			name:      "non-existent",
			storePath: tempDir,
		}

		err := opts.runLoad(nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("save duplicate without force", func(t *testing.T) {
		// Create a test config directory
		configPath := filepath.Join(tempDir, "gcloud")
		err := os.MkdirAll(configPath, 0o755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(configPath, "active_config"), []byte("default"), 0o644)
		require.NoError(t, err)

		opts := &gcloudOptions{
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

	t.Run("getDirSize", func(t *testing.T) {
		// Create test directory with files
		testDir := filepath.Join(tempDir, "size-test")
		err := os.MkdirAll(testDir, 0o755)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(testDir, "file1.txt"), []byte("hello"), 0o644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(testDir, "file2.txt"), []byte("world"), 0o644)
		require.NoError(t, err)

		opts := &gcloudOptions{}
		size, err := opts.getDirSize(testDir)
		assert.NoError(t, err)
		assert.Equal(t, int64(10), size) // "hello" (5) + "world" (5)
	})
}
