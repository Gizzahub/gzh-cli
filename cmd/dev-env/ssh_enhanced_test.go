// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package devenv

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnhancedSSHCommand_SaveAndLoadConfig(t *testing.T) {
	// Create temporary directories
	tempDir := t.TempDir()
	sshDir := filepath.Join(tempDir, ".ssh")
	storeDir := filepath.Join(tempDir, "store")

	require.NoError(t, os.MkdirAll(sshDir, 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(sshDir, "config.d"), 0o755))

	// Create test SSH config
	mainConfig := `Include config.d/*

Host example.com
    HostName example.com
    User myuser
    IdentityFile id_rsa

Host test.com
    HostName test.com
    User testuser
    IdentityFile test_key`

	configPath := filepath.Join(sshDir, "config")
	require.NoError(t, os.WriteFile(configPath, []byte(mainConfig), 0o644))

	// Create include files
	includeContent := `Host work.internal
    HostName work.internal.com
    User worker
    IdentityFile work_key`

	require.NoError(t, os.WriteFile(
		filepath.Join(sshDir, "config.d", "work.conf"),
		[]byte(includeContent),
		0o644,
	))

	// Create key files
	keyFiles := map[string]struct {
		content     string
		permissions os.FileMode
	}{
		"id_rsa":       {"private key content", 0o600},
		"id_rsa.pub":   {"public key content", 0o644},
		"test_key":     {"test private key", 0o600},
		"test_key.pub": {"test public key", 0o644},
		"work_key":     {"work private key", 0o600},
		"work_key.pub": {"work public key", 0o644},
	}

	for keyName, info := range keyFiles {
		keyPath := filepath.Join(sshDir, keyName)
		require.NoError(t, os.WriteFile(keyPath, []byte(info.content), info.permissions))
	}

	// Create enhanced SSH command
	cmd := NewEnhancedSSHCommand()

	t.Run("Save enhanced config", func(t *testing.T) {
		opts := &EnhancedSSHOptions{
			Name:          "test-config",
			Description:   "Test configuration",
			ConfigPath:    configPath,
			StorePath:     storeDir,
			IncludeKeys:   true,
			IncludePublic: true,
		}

		err := cmd.SaveEnhancedConfig(opts)
		require.NoError(t, err)

		// Verify saved structure
		configDir := filepath.Join(storeDir, "test-config")
		assert.DirExists(t, configDir)
		assert.FileExists(t, filepath.Join(configDir, "config"))
		assert.FileExists(t, filepath.Join(configDir, "metadata.json"))
		assert.DirExists(t, filepath.Join(configDir, "includes"))
		assert.DirExists(t, filepath.Join(configDir, "keys"))

		// Check include files
		includesDir := filepath.Join(configDir, "includes")
		entries, err := os.ReadDir(includesDir)
		require.NoError(t, err)
		assert.Len(t, entries, 1) // One include file

		// Check key files
		keysDir := filepath.Join(configDir, "keys")
		keyEntries, err := os.ReadDir(keysDir)
		require.NoError(t, err)
		assert.Len(t, keyEntries, 6) // 3 private + 3 public keys

		// Verify metadata
		metadata, err := cmd.loadEnhancedMetadata(filepath.Join(configDir, "metadata.json"))
		require.NoError(t, err)
		assert.Equal(t, "Test configuration", metadata.Description)
		assert.True(t, metadata.HasIncludes)
		assert.True(t, metadata.HasKeys)
		assert.Len(t, metadata.IncludeFiles, 1)
		assert.Len(t, metadata.PrivateKeys, 3)
		assert.Len(t, metadata.PublicKeys, 3)

		configRoot, err := os.OpenRoot(configDir)
		require.NoError(t, err)
		defer configRoot.Close()

		metadataContent, err := configRoot.ReadFile("metadata.json")
		require.NoError(t, err)
		assert.Contains(t, string(metadataContent), `"private_keys"`)
		for _, info := range keyFiles {
			assert.NotContains(t, string(metadataContent), info.content)
		}

		var metadataJSON struct {
			PrivateKeys []string `json:"private_keys"`
		}
		require.NoError(t, json.Unmarshal(metadataContent, &metadataJSON))
		assert.Equal(t, []string{
			filepath.Join(sshDir, "id_rsa"),
			filepath.Join(sshDir, "test_key"),
			filepath.Join(sshDir, "work_key"),
		}, metadataJSON.PrivateKeys)
	})

	t.Run("Load enhanced config", func(t *testing.T) {
		// Create new target directory
		targetDir := filepath.Join(tempDir, "target")
		targetSSHDir := filepath.Join(targetDir, ".ssh")
		targetConfigPath := filepath.Join(targetSSHDir, "config")

		opts := &EnhancedSSHOptions{
			Name:       "test-config",
			ConfigPath: targetConfigPath,
			StorePath:  storeDir,
			Force:      true,
		}

		err := cmd.LoadEnhancedConfig(opts)
		require.NoError(t, err)

		// Verify loaded files
		assert.FileExists(t, targetConfigPath)
		assert.DirExists(t, filepath.Join(targetSSHDir, "config.d"))

		// Verify main config content
		content, err := os.ReadFile(targetConfigPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "Include config.d/*")
		assert.Contains(t, string(content), "Host example.com")

		// Verify include files are restored
		configDEntries, err := os.ReadDir(filepath.Join(targetSSHDir, "config.d"))
		require.NoError(t, err)
		assert.Len(t, configDEntries, 1)

		// Verify key files are restored
		keyEntries, err := os.ReadDir(targetSSHDir)
		require.NoError(t, err)
		keyFiles := 0
		for _, entry := range keyEntries {
			if !entry.IsDir() && (entry.Name() != "config") {
				keyFiles++
			}
		}
		assert.Equal(t, 6, keyFiles) // 3 private + 3 public keys

		// Verify file permissions where the platform exposes POSIX mode bits.
		info, err := os.Stat(filepath.Join(targetSSHDir, "id_rsa"))
		require.NoError(t, err)
		assertPrivateMode(t, info, 0o600)

		info, err = os.Stat(filepath.Join(targetSSHDir, "id_rsa.pub"))
		require.NoError(t, err)
		assertPrivateMode(t, info, 0o644)
	})

	t.Run("List enhanced configs", func(t *testing.T) {
		opts := &EnhancedSSHOptions{
			StorePath: storeDir,
			ListAll:   true,
		}

		err := cmd.ListEnhancedConfigs(opts)
		require.NoError(t, err)
	})
}

func TestEnhancedSSHCommand_SaveWithoutKeys(t *testing.T) {
	// Create temporary directories
	tempDir := t.TempDir()
	sshDir := filepath.Join(tempDir, ".ssh")
	storeDir := filepath.Join(tempDir, "store")

	require.NoError(t, os.MkdirAll(sshDir, 0o755))

	// Create simple SSH config without key references
	simpleConfig := `Host simple.com
    HostName simple.com
    User simpleuser`

	configPath := filepath.Join(sshDir, "config")
	require.NoError(t, os.WriteFile(configPath, []byte(simpleConfig), 0o644))

	// Create enhanced SSH command
	cmd := NewEnhancedSSHCommand()

	opts := &EnhancedSSHOptions{
		Name:          "simple-config",
		Description:   "Simple configuration without keys",
		ConfigPath:    configPath,
		StorePath:     storeDir,
		IncludeKeys:   false,
		IncludePublic: false,
	}

	err := cmd.SaveEnhancedConfig(opts)
	require.NoError(t, err)

	// Verify saved structure
	configDir := filepath.Join(storeDir, "simple-config")
	assert.DirExists(t, configDir)
	assert.FileExists(t, filepath.Join(configDir, "config"))
	assert.FileExists(t, filepath.Join(configDir, "metadata.json"))

	// Should not have keys directory since no keys were found
	assert.NoDirExists(t, filepath.Join(configDir, "keys"))

	// Verify metadata
	metadata, err := cmd.loadEnhancedMetadata(filepath.Join(configDir, "metadata.json"))
	require.NoError(t, err)
	assert.False(t, metadata.HasKeys)
	assert.Len(t, metadata.PrivateKeys, 0)
	assert.Len(t, metadata.PublicKeys, 0)
}

func TestEnhancedSSHCommand_ErrorCases(t *testing.T) {
	cmd := NewEnhancedSSHCommand()
	tempDir := t.TempDir()

	t.Run("Save with missing config file", func(t *testing.T) {
		opts := &EnhancedSSHOptions{
			Name:       "missing-config",
			ConfigPath: filepath.Join(tempDir, "nonexistent", "config"),
			StorePath:  filepath.Join(tempDir, "store"),
		}

		err := cmd.SaveEnhancedConfig(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config file not found")
	})

	t.Run("Save with empty name", func(t *testing.T) {
		opts := &EnhancedSSHOptions{
			Name:       "",
			ConfigPath: filepath.Join(tempDir, "config"),
			StorePath:  filepath.Join(tempDir, "store"),
		}

		err := cmd.SaveEnhancedConfig(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration name is required")
	})

	t.Run("Load non-existent config", func(t *testing.T) {
		opts := &EnhancedSSHOptions{
			Name:       "nonexistent",
			ConfigPath: filepath.Join(tempDir, "config"),
			StorePath:  filepath.Join(tempDir, "store"),
		}

		err := cmd.LoadEnhancedConfig(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration 'nonexistent' not found")
	})

	t.Run("Load with empty name", func(t *testing.T) {
		opts := &EnhancedSSHOptions{
			Name:       "",
			ConfigPath: filepath.Join(tempDir, "config"),
			StorePath:  filepath.Join(tempDir, "store"),
		}

		err := cmd.LoadEnhancedConfig(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration name is required")
	})
}

func TestEnhancedSSHCommand_OverwriteProtection(t *testing.T) {
	// Create temporary directories
	tempDir := t.TempDir()
	sshDir := filepath.Join(tempDir, ".ssh")
	storeDir := filepath.Join(tempDir, "store")

	require.NoError(t, os.MkdirAll(sshDir, 0o755))

	// Create test SSH config
	configContent := `Host example.com
    HostName example.com
    User myuser`

	configPath := filepath.Join(sshDir, "config")
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0o644))

	cmd := NewEnhancedSSHCommand()

	// Save config first time
	opts := &EnhancedSSHOptions{
		Name:       "overwrite-test",
		ConfigPath: configPath,
		StorePath:  storeDir,
	}

	err := cmd.SaveEnhancedConfig(opts)
	require.NoError(t, err)

	t.Run("Save without force should fail", func(t *testing.T) {
		err := cmd.SaveEnhancedConfig(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
		assert.Contains(t, err.Error(), "--force")
	})

	t.Run("Save with force should succeed", func(t *testing.T) {
		opts.Force = true
		err := cmd.SaveEnhancedConfig(opts)
		assert.NoError(t, err)
	})
}

func TestEnhancedSSHCommand_RestoreEnhancedFileDestinationContract(t *testing.T) {
	command := NewEnhancedSSHCommand()
	tempDir := t.TempDir()
	source := filepath.Join(tempDir, "saved")
	require.NoError(t, os.WriteFile(source, []byte("saved content"), 0o600))

	t.Run("absent destination publishes requested mode", func(t *testing.T) {
		destination := filepath.Join(tempDir, "absent")
		require.NoError(t, command.restoreEnhancedFile(source, destination, 0o600, false))
		content, err := os.ReadFile(destination)
		require.NoError(t, err)
		assert.Equal(t, "saved content", string(content))
		info, err := os.Stat(destination)
		require.NoError(t, err)
		assertPrivateMode(t, info, 0o600)
	})

	t.Run("existing regular file requires force and preserves victim", func(t *testing.T) {
		destination := filepath.Join(tempDir, "regular")
		require.NoError(t, os.WriteFile(destination, []byte("victim"), 0o600))

		err := command.restoreEnhancedFile(source, destination, 0o600, false)
		require.Error(t, err)
		content, readErr := os.ReadFile(destination)
		require.NoError(t, readErr)
		assert.Equal(t, "victim", string(content))

		require.NoError(t, command.restoreEnhancedFile(source, destination, 0o600, true))
		content, readErr = os.ReadFile(destination)
		require.NoError(t, readErr)
		assert.Equal(t, "saved content", string(content))
	})

	for _, force := range []bool{false, true} {
		t.Run("live symlink is refused regardless of force="+strconv.FormatBool(force), func(t *testing.T) {
			victim := filepath.Join(tempDir, "live-victim-"+strconv.FormatBool(force))
			destination := filepath.Join(tempDir, "live-link-"+strconv.FormatBool(force))
			require.NoError(t, os.WriteFile(victim, []byte("victim"), 0o600))
			require.NoError(t, os.Symlink(victim, destination))

			require.Error(t, command.restoreEnhancedFile(source, destination, 0o600, force))
			content, err := os.ReadFile(victim)
			require.NoError(t, err)
			assert.Equal(t, "victim", string(content))
		})

		t.Run("dangling symlink is refused regardless of force="+strconv.FormatBool(force), func(t *testing.T) {
			destination := filepath.Join(tempDir, "dangling-link-"+strconv.FormatBool(force))
			require.NoError(t, os.Symlink(filepath.Join(tempDir, "missing"), destination))

			require.Error(t, command.restoreEnhancedFile(source, destination, 0o600, force))
			info, err := os.Lstat(destination)
			require.NoError(t, err)
			assert.NotZero(t, info.Mode()&os.ModeSymlink)
		})
	}

	t.Run("directory is refused regardless of force", func(t *testing.T) {
		for _, force := range []bool{false, true} {
			destination := filepath.Join(tempDir, "directory-"+strconv.FormatBool(force))
			require.NoError(t, os.Mkdir(destination, 0o700))
			require.Error(t, command.restoreEnhancedFile(source, destination, 0o600, force))
			assert.DirExists(t, destination)
		}
	})

	t.Run("public key mode is explicit", func(t *testing.T) {
		destination := filepath.Join(tempDir, "id_ed25519.pub")
		require.NoError(t, command.restoreEnhancedFile(source, destination, 0o644, false))
		info, err := os.Stat(destination)
		require.NoError(t, err)
		assertPrivateMode(t, info, 0o644)
	})
}

func TestEnhancedSSHCommand_LoadRestoreModesAndFailureStopsFurtherFiles(t *testing.T) {
	command := NewEnhancedSSHCommand()
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "store")
	configDir := filepath.Join(storePath, "saved")
	require.NoError(t, os.MkdirAll(filepath.Join(configDir, "includes"), 0o700))
	require.NoError(t, os.MkdirAll(filepath.Join(configDir, "keys"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "metadata.json"), []byte(`{}`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config"), []byte("Host saved"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "includes", "include_0_hosts.conf"), []byte("Host included"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "keys", "id_ed25519"), []byte("private"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "keys", "id_ed25519.pub"), []byte("public"), 0o600))

	targetConfig := filepath.Join(tempDir, "target", ".ssh", "config")
	opts := &EnhancedSSHOptions{Name: "saved", ConfigPath: targetConfig, StorePath: storePath}
	require.NoError(t, command.LoadEnhancedConfig(opts))

	for path, mode := range map[string]os.FileMode{
		targetConfig: 0o600,
		filepath.Join(filepath.Dir(targetConfig), "config.d", "hosts.conf"): 0o600,
		filepath.Join(filepath.Dir(targetConfig), "id_ed25519"):             0o600,
		filepath.Join(filepath.Dir(targetConfig), "id_ed25519.pub"):         0o644,
	} {
		info, err := os.Stat(path)
		require.NoError(t, err, path)
		assertPrivateMode(t, info, mode)
	}

	// A failed main-file restore must stop before include and key publication.
	blockedTarget := filepath.Join(tempDir, "blocked", ".ssh", "config")
	require.NoError(t, os.MkdirAll(filepath.Dir(blockedTarget), 0o700))
	require.NoError(t, os.WriteFile(blockedTarget, []byte("victim"), 0o600))
	err := command.LoadEnhancedConfig(&EnhancedSSHOptions{Name: "saved", ConfigPath: blockedTarget, StorePath: storePath})
	require.Error(t, err)
	content, readErr := os.ReadFile(blockedTarget)
	require.NoError(t, readErr)
	assert.Equal(t, "victim", string(content))
	assert.NoFileExists(t, filepath.Join(filepath.Dir(blockedTarget), "id_ed25519"))
}

func TestEnhancedSSHCommand_LoadForceFlagContract(t *testing.T) {
	command := NewEnhancedSSHCommand()

	for _, args := range [][]string{nil, {"--force"}, {"-f"}} {
		cmd := command.CreateEnhancedLoadCommand()
		cmd.SetArgs(args)
		if len(args) == 0 {
			assert.Equal(t, "false", cmd.Flags().Lookup("force").DefValue)
			assert.False(t, cmd.Flags().Lookup("force").Changed)
			continue
		}
		require.NoError(t, cmd.ParseFlags(args))
		value, err := cmd.Flags().GetBool("force")
		require.NoError(t, err)
		assert.True(t, value)
	}
}
