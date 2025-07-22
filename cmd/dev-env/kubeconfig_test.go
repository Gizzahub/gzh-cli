//nolint:testpackage // White-box testing needed for internal function access
package devenv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultKubeconfigOptions(t *testing.T) {
	opts := defaultKubeconfigOptions()

	assert.NotEmpty(t, opts.configPath)
	assert.NotEmpty(t, opts.storePath)
	assert.False(t, opts.force)
	assert.False(t, opts.listAll)
}

func TestNewKubeconfigCmd(t *testing.T) {
	cmd := newKubeconfigCmd()

	assert.Equal(t, "kubeconfig", cmd.Use)
	assert.Equal(t, "Manage Kubernetes configuration files", cmd.Short)
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

func TestKubeconfigSaveCmd(t *testing.T) {
	cmd := newKubeconfigSaveCmd()

	assert.Equal(t, "save", cmd.Use)
	assert.Equal(t, "Save current kubeconfig", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("name"))
	assert.NotNil(t, cmd.Flags().Lookup("description"))
	assert.NotNil(t, cmd.Flags().Lookup("config-path"))
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
	assert.NotNil(t, cmd.Flags().Lookup("force"))
}

func TestKubeconfigLoadCmd(t *testing.T) {
	cmd := newKubeconfigLoadCmd()

	assert.Equal(t, "load", cmd.Use)
	assert.Equal(t, "Load a saved kubeconfig", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("name"))
	assert.NotNil(t, cmd.Flags().Lookup("config-path"))
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
	assert.NotNil(t, cmd.Flags().Lookup("force"))
}

func TestKubeconfigListCmd(t *testing.T) {
	cmd := newKubeconfigListCmd()

	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List saved kubeconfigs", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
}

func TestKubeconfigSaveLoad(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	storeDir := filepath.Join(tempDir, "store")

	err := os.MkdirAll(configDir, 0o755)
	require.NoError(t, err)

	// Create a test kubeconfig file
	testConfig := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://test-cluster.example.com
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token`

	configPath := filepath.Join(configDir, "config")
	err = os.WriteFile(configPath, []byte(testConfig), 0o644)
	require.NoError(t, err)

	t.Run("save kubeconfig", func(t *testing.T) {
		opts := &kubeconfigOptions{
			name:        "test-config",
			description: "Test configuration",
			configPath:  configPath,
			storePath:   storeDir,
			force:       false,
		}

		err := opts.runSave(nil, nil)
		assert.NoError(t, err)

		// Check if file was saved
		savedPath := filepath.Join(storeDir, "test-config.yaml")
		assert.FileExists(t, savedPath)

		// Check if metadata was saved
		metadataPath := filepath.Join(storeDir, "test-config.meta")
		assert.FileExists(t, metadataPath)
	})

	t.Run("load kubeconfig", func(t *testing.T) {
		// Create a different target path
		targetPath := filepath.Join(tempDir, "loaded", "config")

		opts := &kubeconfigOptions{
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
		assert.Equal(t, testConfig, string(loadedContent))
	})

	t.Run("list kubeconfigs", func(t *testing.T) {
		opts := &kubeconfigOptions{
			storePath: storeDir,
		}

		err := opts.runList(nil, nil)
		assert.NoError(t, err)
	})
}

func TestKubeconfigMetadata(t *testing.T) {
	tempDir := t.TempDir()

	opts := &kubeconfigOptions{
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

func TestKubeconfigCopyFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create source file
	srcPath := filepath.Join(tempDir, "source.txt")
	content := "test content"
	err := os.WriteFile(srcPath, []byte(content), 0o644)
	require.NoError(t, err)

	// Test copy
	opts := &kubeconfigOptions{}
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

func TestKubeconfigErrorCases(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("save non-existent config", func(t *testing.T) {
		opts := &kubeconfigOptions{
			name:       "test",
			configPath: "/non/existent/path",
			storePath:  tempDir,
		}

		err := opts.runSave(nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "kubeconfig file not found")
	})

	t.Run("load non-existent config", func(t *testing.T) {
		opts := &kubeconfigOptions{
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
		err := os.WriteFile(configPath, []byte("test"), 0o644)
		require.NoError(t, err)

		opts := &kubeconfigOptions{
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
