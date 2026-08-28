// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package devenv

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sshSaveFaultOptions(t *testing.T, include, keys bool) *EnhancedSSHOptions {
	t.Helper()
	root := t.TempDir()
	sshDir := filepath.Join(root, "ssh")
	require.NoError(t, os.Mkdir(sshDir, 0o700))
	lines := []string{"Host test"}
	if include {
		require.NoError(t, os.Mkdir(filepath.Join(sshDir, "config.d"), 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(sshDir, "config.d", "included"), []byte("Host included"), 0o600))
		lines = append(lines, "Include config.d/*")
	}
	if keys {
		require.NoError(t, os.WriteFile(filepath.Join(sshDir, "id_key"), []byte("private"), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(sshDir, "id_key.pub"), []byte("public"), 0o600))
		lines = append(lines, "IdentityFile id_key")
	}
	config := filepath.Join(sshDir, "config")
	require.NoError(t, os.WriteFile(config, []byte(strings.Join(lines, "\n")), 0o600))
	return &EnhancedSSHOptions{Name: "snapshot", ConfigPath: config, StorePath: filepath.Join(root, "store"), IncludeKeys: keys, IncludePublic: keys}
}

func requireSaveFault(t *testing.T, opts *EnhancedSSHOptions) error {
	t.Helper()
	output, err := captureSSHSaveOutput(t, func() error { return NewEnhancedSSHCommand().SaveEnhancedConfig(opts) })
	require.Error(t, err)
	assert.NotContains(t, output, "saved successfully")
	return err
}

func TestSSHSaveFaultCardFileOperations(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	require.NoError(t, os.WriteFile(source, []byte("source"), 0o600))
	t.Run("source-open failure", func(t *testing.T) {
		original := sshSaveOpen
		sshSaveOpen = func(string) (*os.File, error) { return nil, errors.New("source open") }
		t.Cleanup(func() { sshSaveOpen = original })
		require.Error(t, sshSaveCopy(source, filepath.Join(root, "dest"), 0o600))
	})
	t.Run("destination O_EXCL create failure plus source close join", func(t *testing.T) {
		dest := filepath.Join(root, "dest-create")
		createErr, closeErr := errors.New("create"), errors.New("source close")
		originalOpen, originalClose := sshSaveOpenFile, sshSaveClose
		sshSaveOpenFile = func(string, int, os.FileMode) (*os.File, error) { return nil, createErr }
		sshSaveClose = func(closer io.Closer) error { _ = closer.Close(); return closeErr }
		t.Cleanup(func() { sshSaveOpenFile, sshSaveClose = originalOpen, originalClose })
		err := sshSaveCopy(source, dest, 0o600)
		require.ErrorIs(t, err, createErr)
		require.ErrorIs(t, err, closeErr)
	})
	t.Run("successful copy destination-close failure", func(t *testing.T) {
		dest := filepath.Join(root, "dest-close")
		closeErr := errors.New("destination close")
		original := sshSaveClose
		sshSaveClose = func(closer io.Closer) error {
			if file, ok := closer.(*os.File); ok && file.Name() == dest {
				_ = file.Close()
				return closeErr
			}
			return original(closer)
		}
		t.Cleanup(func() { sshSaveClose = original })
		require.ErrorIs(t, sshSaveCopy(source, dest, 0o600), closeErr)
	})
}

func TestSSHSaveFaultCardFullSaveOutputSuppression(t *testing.T) {
	t.Run("source-open", func(t *testing.T) {
		opts := sshSaveFaultOptions(t, false, false)
		original := sshSaveOpen
		sshSaveOpen = func(string) (*os.File, error) { return nil, errors.New("source open") }
		t.Cleanup(func() { sshSaveOpen = original })
		requireSaveFault(t, opts)
	})
	t.Run("destination-create", func(t *testing.T) {
		opts := sshSaveFaultOptions(t, false, false)
		original := sshSaveOpenFile
		sshSaveOpenFile = func(string, int, os.FileMode) (*os.File, error) { return nil, errors.New("destination create") }
		t.Cleanup(func() { sshSaveOpenFile = original })
		requireSaveFault(t, opts)
	})
	t.Run("destination-close", func(t *testing.T) {
		opts := sshSaveFaultOptions(t, false, false)
		original := sshSaveClose
		sshSaveClose = func(closer io.Closer) error {
			if file, ok := closer.(*os.File); ok && filepath.Base(file.Name()) == "config" {
				_ = file.Close()
				return errors.New("destination close")
			}
			return original(closer)
		}
		t.Cleanup(func() { sshSaveClose = original })
		requireSaveFault(t, opts)
	})
	for _, fault := range []string{"create", "write", "chmod"} {
		t.Run("metadata-"+fault, func(t *testing.T) {
			opts := sshSaveFaultOptions(t, false, false)
			originalOpen, originalWrite, originalChmod := sshSaveOpenFile, sshSaveWrite, sshSaveChmod
			sshSaveOpenFile = func(path string, flag int, mode os.FileMode) (*os.File, error) {
				if filepath.Base(path) == "metadata.json" && fault == "create" {
					return nil, errors.New("metadata create")
				}
				return originalOpen(path, flag, mode)
			}
			sshSaveWrite = func(w io.Writer, data []byte) (int, error) {
				if fault == "write" {
					return 0, errors.New("metadata write")
				}
				return originalWrite(w, data)
			}
			sshSaveChmod = func(path string, mode os.FileMode) error {
				if filepath.Base(path) == "metadata.json" && fault == "chmod" {
					return errors.New("metadata chmod")
				}
				return originalChmod(path, mode)
			}
			t.Cleanup(func() { sshSaveOpenFile, sshSaveWrite, sshSaveChmod = originalOpen, originalWrite, originalChmod })
			requireSaveFault(t, opts)
		})
	}
	t.Run("parser-error", func(t *testing.T) {
		opts := sshSaveFaultOptions(t, false, false)
		original := sshParserOpen
		sshParserOpen = func(string) (sshParserFile, error) { return nil, errors.New("parser open") }
		t.Cleanup(func() { sshParserOpen = original })
		requireSaveFault(t, opts)
	})
}

func TestSSHSaveFaultCardStagingAndMetadata(t *testing.T) {
	t.Run("stage MkdirTemp failure", func(t *testing.T) {
		opts := sshSaveFaultOptions(t, false, false)
		original := sshSaveMkdirTemp
		sshSaveMkdirTemp = func(string, string) (string, error) { return "", errors.New("stage temp") }
		t.Cleanup(func() { sshSaveMkdirTemp = original })
		requireSaveFault(t, opts)
	})
	t.Run("stage chmod plus cleanup joined", func(t *testing.T) {
		opts := sshSaveFaultOptions(t, false, false)
		originalTemp, originalChmod, originalRemove := sshSaveMkdirTemp, sshSaveChmod, sshSaveRemoveAll
		var stage string
		sshSaveMkdirTemp = func(d, p string) (string, error) { var err error; stage, err = originalTemp(d, p); return stage, err }
		chmodErr, cleanupErr := errors.New("stage chmod"), errors.New("stage cleanup")
		sshSaveChmod = func(path string, mode os.FileMode) error {
			if path == stage {
				return chmodErr
			}
			return originalChmod(path, mode)
		}
		sshSaveRemoveAll = func(path string) error {
			if path == stage {
				return cleanupErr
			}
			return originalRemove(path)
		}
		t.Cleanup(func() {
			sshSaveMkdirTemp, sshSaveChmod, sshSaveRemoveAll = originalTemp, originalChmod, originalRemove
			_ = os.RemoveAll(stage)
		})
		err := requireSaveFault(t, opts)
		require.ErrorIs(t, err, chmodErr)
		require.ErrorIs(t, err, cleanupErr)
	})
	t.Run("includes mkdir and chmod", func(t *testing.T) {
		for _, fault := range []string{"mkdir", "chmod"} {
			t.Run(fault, func(t *testing.T) {
				opts := sshSaveFaultOptions(t, true, false)
				originalMkdir, originalChmod := sshSaveMkdirDir, sshSaveChmod
				sshSaveMkdirDir = func(path string, mode os.FileMode) error {
					if strings.HasSuffix(path, "includes") && fault == "mkdir" {
						return errors.New("includes mkdir")
					}
					return originalMkdir(path, mode)
				}
				sshSaveChmod = func(path string, mode os.FileMode) error {
					if strings.HasSuffix(path, "includes") && fault == "chmod" {
						return errors.New("includes chmod")
					}
					return originalChmod(path, mode)
				}
				t.Cleanup(func() { sshSaveMkdirDir, sshSaveChmod = originalMkdir, originalChmod })
				requireSaveFault(t, opts)
			})
		}
	})
	t.Run("keys mkdir and chmod", func(t *testing.T) {
		for _, fault := range []string{"mkdir", "chmod"} {
			t.Run(fault, func(t *testing.T) {
				opts := sshSaveFaultOptions(t, false, true)
				originalMkdir, originalChmod := sshSaveMkdirDir, sshSaveChmod
				sshSaveMkdirDir = func(path string, mode os.FileMode) error {
					if strings.HasSuffix(path, "keys") && fault == "mkdir" {
						return errors.New("keys mkdir")
					}
					return originalMkdir(path, mode)
				}
				sshSaveChmod = func(path string, mode os.FileMode) error {
					if strings.HasSuffix(path, "keys") && fault == "chmod" {
						return errors.New("keys chmod")
					}
					return originalChmod(path, mode)
				}
				t.Cleanup(func() { sshSaveMkdirDir, sshSaveChmod = originalMkdir, originalChmod })
				requireSaveFault(t, opts)
			})
		}
	})
	t.Run("metadata successful-write final-close through Save", func(t *testing.T) {
		opts := sshSaveFaultOptions(t, false, false)
		original := sshSaveClose
		sshSaveClose = func(closer io.Closer) error {
			if file, ok := closer.(*os.File); ok && filepath.Base(file.Name()) == "metadata.json" {
				_ = file.Close()
				return errors.New("metadata close")
			}
			return original(closer)
		}
		t.Cleanup(func() { sshSaveClose = original })
		requireSaveFault(t, opts)
	})
}

func TestSSHSaveFaultCardBackupWrapper(t *testing.T) {
	newForced := func(t *testing.T) (*EnhancedSSHOptions, string) {
		t.Helper()
		opts := sshSaveFaultOptions(t, false, false)
		final := filepath.Join(opts.StorePath, opts.Name)
		require.NoError(t, os.MkdirAll(final, 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(final, "config"), []byte("old"), 0o600))
		opts.Force = true
		return opts, final
	}
	t.Run("backup wrapper MkdirTemp", func(t *testing.T) {
		opts, _ := newForced(t)
		original := sshSaveMkdirTemp
		sshSaveMkdirTemp = func(dir, pattern string) (string, error) {
			if strings.HasPrefix(pattern, ".gzh-ssh-backup-") {
				return "", errors.New("wrapper temp")
			}
			return original(dir, pattern)
		}
		t.Cleanup(func() { sshSaveMkdirTemp = original })
		requireSaveFault(t, opts)
	})
	t.Run("wrapper chmod plus wrapper cleanup joined", func(t *testing.T) {
		opts, _ := newForced(t)
		originalChmod, originalRemove := sshSaveChmod, sshSaveRemoveAll
		chmodErr, cleanupErr := errors.New("wrapper chmod"), errors.New("wrapper cleanup")
		sshSaveChmod = func(path string, mode os.FileMode) error {
			if strings.Contains(filepath.Base(path), ".gzh-ssh-backup-") {
				return chmodErr
			}
			return originalChmod(path, mode)
		}
		sshSaveRemoveAll = func(path string) error {
			if strings.Contains(filepath.Base(path), ".gzh-ssh-backup-") {
				return cleanupErr
			}
			return originalRemove(path)
		}
		t.Cleanup(func() { sshSaveChmod, sshSaveRemoveAll = originalChmod, originalRemove })
		err := requireSaveFault(t, opts)
		require.ErrorIs(t, err, chmodErr)
		require.ErrorIs(t, err, cleanupErr)
	})
	t.Run("rollback success plus wrapper cleanup joined at caller", func(t *testing.T) {
		opts, final := newForced(t)
		originalRename, originalTemp, originalRemove := sshSaveRename, sshSaveMkdirTemp, sshSaveRemoveAll
		var stage string
		sshSaveMkdirTemp = func(dir, pattern string) (string, error) {
			path, err := originalTemp(dir, pattern)
			if strings.HasPrefix(pattern, ".gzh-ssh-stage-") {
				stage = path
			}
			return path, err
		}
		publishErr, cleanupErr := errors.New("publish"), errors.New("wrapper cleanup")
		sshSaveRename = func(from, to string) error {
			if from == stage && to == final {
				return publishErr
			}
			return originalRename(from, to)
		}
		sshSaveRemoveAll = func(path string) error {
			if strings.Contains(filepath.Base(path), ".gzh-ssh-backup-") {
				return cleanupErr
			}
			return originalRemove(path)
		}
		t.Cleanup(func() {
			sshSaveRename, sshSaveMkdirTemp, sshSaveRemoveAll = originalRename, originalTemp, originalRemove
			_ = os.RemoveAll(stage)
		})
		err := requireSaveFault(t, opts)
		require.ErrorIs(t, err, publishErr)
		require.ErrorIs(t, err, cleanupErr)
		assert.NoDirExists(t, stage)
		assert.DirExists(t, final)
	})
}
