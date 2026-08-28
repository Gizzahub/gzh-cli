// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package devenv

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureSSHSaveOutput(t *testing.T, action func() error) (string, error) {
	t.Helper()
	original := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = writer
	err = action()
	require.NoError(t, writer.Close())
	os.Stdout = original
	output, readErr := io.ReadAll(reader)
	require.NoError(t, readErr)
	require.NoError(t, reader.Close())
	return string(output), err
}

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
	require.NoError(t, os.WriteFile(configPath, []byte(mainConfig), 0o600))

	// Create include files
	includeContent := `Host work.internal
    HostName work.internal.com
    User worker
    IdentityFile work_key`

	require.NoError(t, os.WriteFile(
		filepath.Join(sshDir, "config.d", "work.conf"),
		[]byte(includeContent),
		0o600,
	))

	// Create key files
	keyFiles := map[string]struct {
		content     string
		permissions os.FileMode
	}{
		"id_rsa":       {"private key content", 0o600},
		"id_rsa.pub":   {"public key content", 0o600},
		"test_key":     {"test private key", 0o600},
		"test_key.pub": {"test public key", 0o600},
		"work_key":     {"work private key", 0o600},
		"work_key.pub": {"work public key", 0o600},
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
	require.NoError(t, os.WriteFile(configPath, []byte(simpleConfig), 0o600))

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
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0o600))

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

func TestEnhancedSSHCommand_SaveSnapshotContract(t *testing.T) {
	tempDir := t.TempDir()
	sshDir := filepath.Join(tempDir, "ssh")
	store := filepath.Join(tempDir, "store")
	require.NoError(t, os.MkdirAll(sshDir, 0o700))
	config := filepath.Join(sshDir, "config")
	key := filepath.Join(sshDir, "id_test")
	require.NoError(t, os.WriteFile(key, []byte("private"), 0o600))
	require.NoError(t, os.WriteFile(key+".pub", []byte("public"), 0o600))
	require.NoError(t, os.WriteFile(config, []byte("Host test\n IdentityFile id_test\n"), 0o600))

	command := NewEnhancedSSHCommand()
	opts := &EnhancedSSHOptions{Name: "snapshot", ConfigPath: config, StorePath: store, IncludeKeys: true, IncludePublic: true}
	require.NoError(t, command.SaveEnhancedConfig(opts))
	final := filepath.Join(store, "snapshot")
	for path, mode := range map[string]os.FileMode{
		store:                                       0o700,
		final:                                       0o700,
		filepath.Join(final, "keys"):                0o700,
		filepath.Join(final, "config"):              0o600,
		filepath.Join(final, "metadata.json"):       0o600,
		filepath.Join(final, "keys", "id_test"):     0o600,
		filepath.Join(final, "keys", "id_test.pub"): 0o644,
	} {
		info, err := os.Stat(path)
		require.NoError(t, err, path)
		assertPrivateMode(t, info, mode)
	}
	metadata, err := command.loadEnhancedMetadata(filepath.Join(final, "metadata.json"))
	require.NoError(t, err)
	assert.True(t, metadata.HasKeys)
	assert.Equal(t, []string{key}, metadata.PrivateKeys)
	assert.Equal(t, []string{key + ".pub"}, metadata.PublicKeys)

	err = command.SaveEnhancedConfig(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	require.NoError(t, os.WriteFile(config, []byte("Host replacement"), 0o600))
	opts.Force = true
	require.NoError(t, command.SaveEnhancedConfig(opts))
	content, err := os.ReadFile(filepath.Join(final, "config"))
	require.NoError(t, err)
	assert.Equal(t, "Host replacement", string(content))
}

func TestEnhancedSSHCommand_SaveRejectsUnsafeNameAndDestination(t *testing.T) {
	tempDir := t.TempDir()
	config := filepath.Join(tempDir, "config")
	require.NoError(t, os.WriteFile(config, []byte("Host example"), 0o600))
	command := NewEnhancedSSHCommand()
	for _, name := range []string{"../escape", "a/b", ".", ".."} {
		err := command.SaveEnhancedConfig(&EnhancedSSHOptions{Name: name, ConfigPath: config, StorePath: filepath.Join(tempDir, "store")})
		require.Error(t, err, name)
	}
	store := filepath.Join(tempDir, "store")
	require.NoError(t, os.MkdirAll(store, 0o700))
	final := filepath.Join(store, "entry")
	require.NoError(t, os.WriteFile(final, []byte("victim"), 0o600))
	err := command.SaveEnhancedConfig(&EnhancedSSHOptions{Name: "entry", ConfigPath: config, StorePath: store, Force: true})
	require.Error(t, err)
	content, readErr := os.ReadFile(final)
	require.NoError(t, readErr)
	assert.Equal(t, "victim", string(content))
}

func TestSSHSavePublishForceRestoresOldSnapshotOnPublishFailure(t *testing.T) {
	root := t.TempDir()
	final := filepath.Join(root, "saved")
	stage := filepath.Join(root, "stage")
	require.NoError(t, os.Mkdir(final, 0o700))
	require.NoError(t, os.Mkdir(stage, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(final, "config"), []byte("old"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(stage, "config"), []byte("new"), 0o600))

	publishErr := errors.New("publish failed")
	originalRename := sshSaveRename
	sshSaveRename = func(from, to string) error {
		if from == stage && to == final {
			return publishErr
		}
		return originalRename(from, to)
	}
	t.Cleanup(func() { sshSaveRename = originalRename })

	_, err := sshSavePublish(stage, final, true)
	require.ErrorIs(t, err, publishErr)
	content, readErr := os.ReadFile(filepath.Join(final, "config"))
	require.NoError(t, readErr)
	assert.Equal(t, "old", string(content))
	// The caller owns stage cleanup; publish only owns its backup wrapper.
	assert.DirExists(t, stage)
}

func TestSSHSavePublishFailureStatesAndOwnership(t *testing.T) {
	newSnapshot := func(t *testing.T) (string, string) {
		t.Helper()
		root := t.TempDir()
		final, stage := filepath.Join(root, "saved"), filepath.Join(root, "stage")
		require.NoError(t, os.Mkdir(final, 0o700))
		require.NoError(t, os.Mkdir(stage, 0o700))
		return final, stage
	}
	t.Run("no force publish failure leaves stage for caller", func(t *testing.T) {
		root := t.TempDir()
		stage, final := filepath.Join(root, "stage"), filepath.Join(root, "saved")
		require.NoError(t, os.Mkdir(stage, 0o700))
		original := sshSaveRename
		sshSaveRename = func(_, _ string) error { return errors.New("rename") }
		t.Cleanup(func() { sshSaveRename = original })
		result, err := sshSavePublish(stage, final, false)
		require.Error(t, err)
		assert.Equal(t, sshSaveStageCleanup, result.disposition)
		assert.DirExists(t, stage)
	})
	t.Run("old to backup failure reports all observed paths", func(t *testing.T) {
		final, stage := newSnapshot(t)
		original := sshSaveRename
		sshSaveRename = func(from, to string) error {
			if from == final {
				return errors.New("old-to-backup")
			}
			return original(from, to)
		}
		t.Cleanup(func() { sshSaveRename = original })
		_, err := sshSavePublish(stage, final, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "stage="+stage)
		assert.Contains(t, err.Error(), "final="+final)
		assert.DirExists(t, stage)
		assert.DirExists(t, final)
	})
	t.Run("rollback failure retains stage and backup", func(t *testing.T) {
		final, stage := newSnapshot(t)
		original := sshSaveRename
		sshSaveRename = func(from, to string) error {
			if from == stage && to == final {
				return errors.New("publish")
			}
			if to == final {
				return errors.New("rollback")
			}
			return original(from, to)
		}
		t.Cleanup(func() { sshSaveRename = original })
		result, err := sshSavePublish(stage, final, true)
		require.Error(t, err)
		assert.Equal(t, sshSaveStageRetained, result.disposition)
		assert.Contains(t, err.Error(), "wrapper=")
		assert.DirExists(t, stage)
		assert.NoDirExists(t, final)
	})
	t.Run("postcommit backup cleanup is reported without removing stage", func(t *testing.T) {
		final, stage := newSnapshot(t)
		original := sshSaveRemoveAll
		sshSaveRemoveAll = func(path string) error {
			if strings.Contains(filepath.Base(path), ".gzh-ssh-backup-") {
				return errors.New("cleanup")
			}
			return original(path)
		}
		t.Cleanup(func() { sshSaveRemoveAll = original })
		result, err := sshSavePublish(stage, final, true)
		require.Error(t, err)
		assert.Equal(t, sshSaveStageCommitted, result.disposition)
		assert.Contains(t, err.Error(), "stage="+stage)
		assert.Contains(t, err.Error(), "final="+final)
		assert.DirExists(t, final)
	})
}

func TestSSHSaveDedupeSymlinksAndUnicodeFold(t *testing.T) {
	root := t.TempDir()
	realFile := filepath.Join(root, "real")
	alias := filepath.Join(root, "alias")
	require.NoError(t, os.WriteFile(realFile, []byte("key"), 0o600))
	require.NoError(t, os.Symlink(realFile, alias))
	result, err := sshSaveDedupe([]string{realFile, alias}, "key")
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{realFile, alias}, result, "different restore basenames must be preserved even for one inode")
	dangling := filepath.Join(root, "dangling")
	require.NoError(t, os.Symlink(filepath.Join(root, "missing"), dangling))
	_, err = sshSaveDedupe([]string{dangling}, "key")
	require.Error(t, err)
	dir := filepath.Join(root, "directory")
	require.NoError(t, os.Mkdir(dir, 0o700))
	_, err = sshSaveDedupe([]string{dir}, "key")
	require.Error(t, err)
	upper := filepath.Join(root, "\u212aey") // Kelvin sign folds with ASCII K.
	lower := filepath.Join(root, "key")
	require.NoError(t, os.WriteFile(upper, []byte("one"), 0o600))
	require.NoError(t, os.WriteFile(lower, []byte("two"), 0o600))
	upperInfo, err := os.Stat(upper)
	require.NoError(t, err)
	lowerInfo, err := os.Stat(lower)
	require.NoError(t, err)
	if os.SameFile(upperInfo, lowerInfo) {
		t.Skip("case-insensitive filesystem aliases Unicode-folded basenames")
	}
	_, err = sshSaveDedupe([]string{upper, lower}, "key")
	require.Error(t, err)
}

func TestEnhancedSSHCommand_SaveCopyFaultMatrix(t *testing.T) {
	newOptions := func(t *testing.T) *EnhancedSSHOptions {
		t.Helper()
		root := t.TempDir()
		sshDir := filepath.Join(root, "ssh")
		require.NoError(t, os.Mkdir(sshDir, 0o700))
		includeDir := filepath.Join(sshDir, "config.d")
		require.NoError(t, os.Mkdir(includeDir, 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(includeDir, "extra"), []byte("Host extra"), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(sshDir, "id_key"), []byte("private"), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(sshDir, "id_key.pub"), []byte("public"), 0o600))
		config := filepath.Join(sshDir, "config")
		require.NoError(t, os.WriteFile(config, []byte("Include config.d/*\nIdentityFile id_key\n"), 0o600))
		return &EnhancedSSHOptions{Name: "snapshot", ConfigPath: config, StorePath: filepath.Join(root, "store"), IncludeKeys: true, IncludePublic: true}
	}
	for copyNumber, label := range []string{"main", "include", "private", "public"} {
		t.Run(label+" copy error leaves no published snapshot", func(t *testing.T) {
			opts := newOptions(t)
			original := sshSaveCopyBytes
			calls := 0
			sshSaveCopyBytes = func(dst io.Writer, src io.Reader) (int64, error) {
				calls++
				if calls == copyNumber+1 {
					return 0, errors.New(label + " copy")
				}
				return original(dst, src)
			}
			t.Cleanup(func() { sshSaveCopyBytes = original })
			output, err := captureSSHSaveOutput(t, func() error { return NewEnhancedSSHCommand().SaveEnhancedConfig(opts) })
			require.Error(t, err)
			assert.Contains(t, err.Error(), label+" copy")
			assert.NotContains(t, output, "saved successfully")
			assert.NoDirExists(t, filepath.Join(opts.StorePath, opts.Name))
		})
	}
}

func TestSSHSaveCopyAndMetadataFaultJoins(t *testing.T) {
	root := t.TempDir()
	source, destination := filepath.Join(root, "source"), filepath.Join(root, "destination")
	require.NoError(t, os.WriteFile(source, []byte("source"), 0o600))
	t.Run("source close failure", func(t *testing.T) {
		original := sshSaveClose
		sourceCloseErr := errors.New("source close")
		sshSaveClose = func(closer io.Closer) error {
			if file, ok := closer.(*os.File); ok && file.Name() == source {
				_ = file.Close()
				return sourceCloseErr
			}
			return original(closer)
		}
		t.Cleanup(func() { sshSaveClose = original })
		err := sshSaveCopy(source, destination, 0o600)
		require.ErrorIs(t, err, sourceCloseErr)
		assert.Contains(t, err.Error(), source)
	})
	t.Run("copy and both close failures join", func(t *testing.T) {
		dest := destination + "-copy"
		copyErr, sourceErr, destErr := errors.New("copy"), errors.New("source close"), errors.New("destination close")
		originalCopy, originalClose := sshSaveCopyBytes, sshSaveClose
		sshSaveCopyBytes = func(io.Writer, io.Reader) (int64, error) { return 0, copyErr }
		sshSaveClose = func(closer io.Closer) error {
			if file, ok := closer.(*os.File); ok {
				_ = file.Close()
				if file.Name() == source {
					return sourceErr
				}
				if file.Name() == dest {
					return destErr
				}
			}
			return originalClose(closer)
		}
		t.Cleanup(func() { sshSaveCopyBytes, sshSaveClose = originalCopy, originalClose })
		err := sshSaveCopy(source, dest, 0o600)
		require.ErrorIs(t, err, copyErr)
		require.ErrorIs(t, err, sourceErr)
		require.ErrorIs(t, err, destErr)
	})
	t.Run("chmod and destination close failures join", func(t *testing.T) {
		dest := destination + "-chmod"
		chmodErr, closeErr := errors.New("chmod"), errors.New("destination close")
		originalChmod, originalClose := sshSaveChmod, sshSaveClose
		sshSaveChmod = func(path string, mode os.FileMode) error {
			if path == dest {
				return chmodErr
			}
			return originalChmod(path, mode)
		}
		sshSaveClose = func(closer io.Closer) error {
			if file, ok := closer.(*os.File); ok && file.Name() == dest {
				_ = file.Close()
				return closeErr
			}
			return originalClose(closer)
		}
		t.Cleanup(func() { sshSaveChmod, sshSaveClose = originalChmod, originalClose })
		err := sshSaveCopy(source, dest, 0o600)
		require.ErrorIs(t, err, chmodErr)
		require.ErrorIs(t, err, closeErr)
		assert.Contains(t, err.Error(), dest)
	})
	t.Run("metadata write chmod and close errors are contextual", func(t *testing.T) {
		dest := filepath.Join(root, "metadata.json")
		writeErr, closeErr := errors.New("write"), errors.New("metadata close")
		originalWrite, originalClose := sshSaveWrite, sshSaveClose
		sshSaveWrite = func(io.Writer, []byte) (int, error) { return 0, writeErr }
		sshSaveClose = func(closer io.Closer) error { _ = closer.Close(); return closeErr }
		t.Cleanup(func() { sshSaveWrite, sshSaveClose = originalWrite, originalClose })
		err := sshSaveBytes(dest, []byte("metadata"), 0o600)
		require.ErrorIs(t, err, writeErr)
		require.ErrorIs(t, err, closeErr)
		assert.Contains(t, err.Error(), dest)
	})
	t.Run("metadata chmod and close failures join", func(t *testing.T) {
		dest := filepath.Join(root, "metadata-chmod.json")
		chmodErr, closeErr := errors.New("metadata chmod"), errors.New("metadata close")
		originalChmod, originalClose := sshSaveChmod, sshSaveClose
		sshSaveChmod = func(path string, mode os.FileMode) error {
			if path == dest {
				return chmodErr
			}
			return originalChmod(path, mode)
		}
		sshSaveClose = func(closer io.Closer) error { _ = closer.Close(); return closeErr }
		t.Cleanup(func() { sshSaveChmod, sshSaveClose = originalChmod, originalClose })
		err := sshSaveBytes(dest, []byte("metadata"), 0o600)
		require.ErrorIs(t, err, chmodErr)
		require.ErrorIs(t, err, closeErr)
		assert.Contains(t, err.Error(), dest)
	})
}

func TestEnhancedSSHCommand_SaveMetadataAndStageCleanupFaults(t *testing.T) {
	newOptions := func(t *testing.T) *EnhancedSSHOptions {
		t.Helper()
		root := t.TempDir()
		config := filepath.Join(root, "config")
		require.NoError(t, os.WriteFile(config, []byte("Host test"), 0o600))
		return &EnhancedSSHOptions{Name: "snapshot", ConfigPath: config, StorePath: filepath.Join(root, "store")}
	}
	t.Run("metadata create failure", func(t *testing.T) {
		opts := newOptions(t)
		original := sshSaveOpenFile
		sshSaveOpenFile = func(path string, flag int, mode os.FileMode) (*os.File, error) {
			if filepath.Base(path) == "metadata.json" {
				return nil, errors.New("metadata create")
			}
			return original(path, flag, mode)
		}
		t.Cleanup(func() { sshSaveOpenFile = original })
		err := NewEnhancedSSHCommand().SaveEnhancedConfig(opts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "metadata create")
		assert.NoDirExists(t, filepath.Join(opts.StorePath, opts.Name))
	})
	t.Run("stage cleanup failure names stage and final", func(t *testing.T) {
		opts := newOptions(t)
		originalCopy, originalRemove, originalTemp := sshSaveCopyBytes, sshSaveRemoveAll, sshSaveMkdirTemp
		var stage string
		sshSaveMkdirTemp = func(dir, pattern string) (string, error) {
			var err error
			stage, err = originalTemp(dir, pattern)
			return stage, err
		}
		sshSaveCopyBytes = func(io.Writer, io.Reader) (int64, error) { return 0, errors.New("copy") }
		sshSaveRemoveAll = func(path string) error {
			if path == stage {
				return errors.New("stage cleanup")
			}
			return originalRemove(path)
		}
		t.Cleanup(func() {
			sshSaveCopyBytes, sshSaveRemoveAll, sshSaveMkdirTemp = originalCopy, originalRemove, originalTemp
			_ = os.RemoveAll(stage)
		})
		err := NewEnhancedSSHCommand().SaveEnhancedConfig(opts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "stage="+stage)
		assert.Contains(t, err.Error(), "final="+filepath.Join(opts.StorePath, opts.Name))
	})
}

func TestPreflightEnhancedLoadIncludeContainerAndForceDestinations(t *testing.T) {
	root := t.TempDir()
	snapshot, target := filepath.Join(root, "snapshot"), filepath.Join(root, "target", "config")
	require.NoError(t, os.MkdirAll(snapshot, 0o700))
	t.Run("no includes does not reserve config.d", func(t *testing.T) {
		configD := filepath.Join(filepath.Dir(target), "config.d")
		require.NoError(t, os.MkdirAll(filepath.Dir(target), 0o700))
		require.NoError(t, os.WriteFile(configD, []byte("not a directory"), 0o600))
		require.NoError(t, preflightEnhancedLoad(snapshot, target, false))
	})
	t.Run("includes accept existing real container but inspect actual destinations", func(t *testing.T) {
		includeDir := filepath.Join(snapshot, "includes")
		require.NoError(t, os.MkdirAll(includeDir, 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(includeDir, "include_0_hosts"), []byte("Host test"), 0o600))
		configD := filepath.Join(filepath.Dir(target), "config.d")
		require.NoError(t, os.Remove(configD))
		require.NoError(t, os.Mkdir(configD, 0o700))
		require.NoError(t, preflightEnhancedLoad(snapshot, target, false))
		require.NoError(t, os.WriteFile(filepath.Join(configD, "hosts"), []byte("old"), 0o600))
		require.Error(t, preflightEnhancedLoad(snapshot, target, false))
	})
	t.Run("container symlink and force symlink destination are rejected", func(t *testing.T) {
		configD := filepath.Join(filepath.Dir(target), "config.d")
		require.NoError(t, os.RemoveAll(configD))
		require.NoError(t, os.Symlink(root, configD))
		require.Error(t, preflightEnhancedLoad(snapshot, target, true))
		require.NoError(t, os.Remove(configD))
		require.NoError(t, os.Symlink(filepath.Join(root, "elsewhere"), target))
		require.Error(t, preflightEnhancedLoad(snapshot, target, true))
	})
}

func TestSSHLoadKeyModeMetadataRoles(t *testing.T) {
	metadata := &EnhancedSSHMetadata{PrivateKeys: []string{"/keys/secret.pub"}, PublicKeys: []string{"/keys/public.pub"}}
	assert.Equal(t, os.FileMode(0o600), sshLoadKeyMode("secret.pub", metadata))
	assert.Equal(t, os.FileMode(0o644), sshLoadKeyMode("public.pub", metadata))
	assert.Equal(t, os.FileMode(0o644), sshLoadKeyMode("legacy.pub", &EnhancedSSHMetadata{}))
	// Private role wins if malformed metadata lists the same basename in both roles.
	conflict := &EnhancedSSHMetadata{PrivateKeys: []string{"/keys/KEY"}, PublicKeys: []string{"/keys/key"}}
	assert.Equal(t, os.FileMode(0o600), sshLoadKeyMode("KEY", conflict))
}

func TestEnhancedSSHCommand_SavePublishFailureIntegration(t *testing.T) {
	newForcedSave := func(t *testing.T) (*EnhancedSSHOptions, string) {
		t.Helper()
		root := t.TempDir()
		config, store := filepath.Join(root, "config"), filepath.Join(root, "store")
		require.NoError(t, os.WriteFile(config, []byte("Host new"), 0o600))
		final := filepath.Join(store, "snapshot")
		require.NoError(t, os.MkdirAll(final, 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(final, "config"), []byte("Host old"), 0o600))
		return &EnhancedSSHOptions{Name: "snapshot", ConfigPath: config, StorePath: store, Force: true}, final
	}
	t.Run("postcommit cleanup failure never prints success or removes committed stage", func(t *testing.T) {
		opts, final := newForcedSave(t)
		originalRemove, originalTemp := sshSaveRemoveAll, sshSaveMkdirTemp
		var stage string
		sshSaveMkdirTemp = func(dir, pattern string) (string, error) {
			path, err := originalTemp(dir, pattern)
			if strings.HasPrefix(pattern, ".gzh-ssh-stage-") {
				stage = path
			}
			return path, err
		}
		sshSaveRemoveAll = func(path string) error {
			if strings.Contains(filepath.Base(path), ".gzh-ssh-backup-") {
				return errors.New("backup cleanup")
			}
			return originalRemove(path)
		}
		t.Cleanup(func() { sshSaveRemoveAll, sshSaveMkdirTemp = originalRemove, originalTemp; _ = os.RemoveAll(stage) })
		output, err := captureSSHSaveOutput(t, func() error { return NewEnhancedSSHCommand().SaveEnhancedConfig(opts) })
		require.Error(t, err)
		assert.NotContains(t, output, "saved successfully")
		assert.DirExists(t, final)
		assert.NoDirExists(t, stage)
	})
	t.Run("rollback failure retains stage and suppresses success", func(t *testing.T) {
		opts, final := newForcedSave(t)
		originalRename, originalTemp := sshSaveRename, sshSaveMkdirTemp
		var stage string
		sshSaveMkdirTemp = func(dir, pattern string) (string, error) {
			path, err := originalTemp(dir, pattern)
			if strings.HasPrefix(pattern, ".gzh-ssh-stage-") {
				stage = path
			}
			return path, err
		}
		sshSaveRename = func(from, to string) error {
			if from == stage && to == final {
				return errors.New("publish")
			}
			if to == final {
				return errors.New("rollback")
			}
			return originalRename(from, to)
		}
		t.Cleanup(func() { sshSaveRename, sshSaveMkdirTemp = originalRename, originalTemp; _ = os.RemoveAll(stage) })
		output, err := captureSSHSaveOutput(t, func() error { return NewEnhancedSSHCommand().SaveEnhancedConfig(opts) })
		require.Error(t, err)
		assert.NotContains(t, output, "saved successfully")
		assert.DirExists(t, stage)
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
		assertNoRestoreTemps(t, tempDir)
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
		assertNoRestoreTemps(t, tempDir)
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

	t.Run("missing source leaves absent and forced existing destinations unchanged", func(t *testing.T) {
		missingSource := filepath.Join(tempDir, "missing-source")
		absentDestination := filepath.Join(tempDir, "missing-absent")
		require.Error(t, command.restoreEnhancedFile(missingSource, absentDestination, 0o600, false))
		assert.NoFileExists(t, absentDestination)

		existingDestination := filepath.Join(tempDir, "missing-existing")
		require.NoError(t, os.WriteFile(existingDestination, []byte("victim"), 0o600))
		require.Error(t, command.restoreEnhancedFile(missingSource, existingDestination, 0o600, true))
		info, err := os.Stat(existingDestination)
		require.NoError(t, err)
		assert.Equal(t, int64(len("victim")), info.Size())
		assertNoRestoreTemps(t, tempDir)
	})

	t.Run("invalid mode leaves destination unchanged", func(t *testing.T) {
		destination := filepath.Join(tempDir, "invalid-mode")
		require.NoError(t, os.WriteFile(destination, []byte("victim"), 0o600))
		require.Error(t, command.restoreEnhancedFile(source, destination, 0o640, true))
		info, err := os.Stat(destination)
		require.NoError(t, err)
		assert.Equal(t, int64(len("victim")), info.Size())
		assertNoRestoreTemps(t, tempDir)
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

	// main 파일 복원 실패 시 include와 key 발행 전에 중단해야 한다.
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

func assertNoRestoreTemps(t *testing.T, directory string) {
	t.Helper()
	entries, err := os.ReadDir(directory)
	require.NoError(t, err)
	for _, entry := range entries {
		assert.False(t, strings.HasPrefix(entry.Name(), ".gzh-ssh-restore-"), entry.Name())
	}
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
