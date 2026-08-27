// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

//go:build windows

package selfupdate

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplaceBinaryWithOpsRetainsCommittedBackupAsWarning(t *testing.T) {
	dir := t.TempDir()
	currentPath := filepath.Join(dir, "gz.exe")
	tempPath := filepath.Join(dir, ".gz-update-stage.exe")
	require.NoError(t, os.WriteFile(currentPath, []byte("old"), 0o600))
	require.NoError(t, os.WriteFile(tempPath, []byte("new"), 0o600))
	backupDir := filepath.Join(dir, ".gz-backup-test")

	renameCalls := 0
	ops := windowsReplaceOps{
		mkdirTemp: func(string, string) (string, error) { return backupDir, nil },
		rename: func(string, string) error {
			renameCalls++
			return nil
		},
		remove: func(path string) error {
			if path == filepath.Join(backupDir, filepath.Base(currentPath)) {
				return errors.New("sharing violation")
			}
			return nil
		},
	}

	result, err := replaceBinaryWithOps(tempPath, currentPath, ops)
	require.NoError(t, err)
	require.Error(t, result.cleanupWarning)
	assert.ErrorContains(t, result.cleanupWarning, "backup retained")
	assert.Equal(t, 2, renameCalls)
}

func TestReplaceBinaryWithOpsRollback(t *testing.T) {
	dir := t.TempDir()
	currentPath := filepath.Join(dir, "gz.exe")
	tempPath := filepath.Join(dir, ".gz-update-stage.exe")
	require.NoError(t, os.WriteFile(currentPath, []byte("old"), 0o600))
	require.NoError(t, os.WriteFile(tempPath, []byte("new"), 0o600))
	backupDir := filepath.Join(dir, ".gz-backup-test")
	replaceErr := errors.New("replace failed")

	t.Run("rollback succeeds", func(t *testing.T) {
		renameCalls := 0
		ops := windowsReplaceOps{
			mkdirTemp: func(string, string) (string, error) { return backupDir, nil },
			rename: func(string, string) error {
				renameCalls++
				if renameCalls == 2 {
					return replaceErr
				}
				return nil
			},
			remove: func(string) error { return nil },
		}

		result, err := replaceBinaryWithOps(tempPath, currentPath, ops)
		require.NoError(t, result.cleanupWarning)
		require.ErrorIs(t, err, replaceErr)
		assert.Equal(t, 3, renameCalls)
	})

	t.Run("rollback failure exposes backup", func(t *testing.T) {
		renameCalls := 0
		rollbackErr := errors.New("rollback failed")
		ops := windowsReplaceOps{
			mkdirTemp: func(string, string) (string, error) { return backupDir, nil },
			rename: func(string, string) error {
				renameCalls++
				switch renameCalls {
				case 2:
					return replaceErr
				case 3:
					return rollbackErr
				default:
					return nil
				}
			},
			remove: func(string) error { return nil },
		}

		result, err := replaceBinaryWithOps(tempPath, currentPath, ops)
		require.NoError(t, result.cleanupWarning)
		require.ErrorIs(t, err, replaceErr)
		require.ErrorIs(t, err, rollbackErr)
		assert.ErrorContains(t, err, filepath.Join(backupDir, filepath.Base(currentPath)))
	})
}
