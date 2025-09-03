// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package devenv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnhancedSSHCommand_SaveAndLoadConfig(t *testing.T) {
	// Create temporary directories
	tempDir := t.TempDir()
	sshDir := filepath.Join(tempDir, ".ssh")
	storeDir := filepath.Join(tempDir, "store")
	
	require.NoError(t, os.MkdirAll(sshDir, 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(sshDir, "config.d"), 0755))

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
	require.NoError(t, os.WriteFile(configPath, []byte(mainConfig), 0644))

	// Create include files
	includeContent := `Host work.internal
    HostName work.internal.com
    User worker
    IdentityFile work_key`
	
	require.NoError(t, os.WriteFile(
		filepath.Join(sshDir, "config.d", "work.conf"), 
		[]byte(includeContent), 
		0644,
	))

	// Create key files
	keyFiles := map[string]struct {
		content     string
		permissions os.FileMode
	}{
		"id_rsa":        {"private key content", 0600},
		"id_rsa.pub":    {"public key content", 0644},
		"test_key":      {"test private key", 0600},
		"test_key.pub":  {"test public key", 0644},
		"work_key":      {"work private key", 0600},
		"work_key.pub":  {"work public key", 0644},
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

		// Verify file permissions
		if info, err := os.Stat(filepath.Join(targetSSHDir, "id_rsa")); err == nil {
			assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
		}
		if info, err := os.Stat(filepath.Join(targetSSHDir, "id_rsa.pub")); err == nil {
			assert.Equal(t, os.FileMode(0644), info.Mode().Perm())
		}
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
	
	require.NoError(t, os.MkdirAll(sshDir, 0755))

	// Create simple SSH config without key references
	simpleConfig := `Host simple.com
    HostName simple.com
    User simpleuser`

	configPath := filepath.Join(sshDir, "config")
	require.NoError(t, os.WriteFile(configPath, []byte(simpleConfig), 0644))

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
	
	require.NoError(t, os.MkdirAll(sshDir, 0755))

	// Create test SSH config
	configContent := `Host example.com
    HostName example.com
    User myuser`

	configPath := filepath.Join(sshDir, "config")
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

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