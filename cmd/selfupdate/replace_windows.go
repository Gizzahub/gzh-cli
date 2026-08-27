// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

//go:build windows

package selfupdate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func replaceBinary(tempPath, currentPath string) error {
	if err := validateReplacementPaths(tempPath, currentPath); err != nil {
		return err
	}

	backupDir, err := os.MkdirTemp(filepath.Dir(currentPath), ".gz-backup-*")
	if err != nil {
		return fmt.Errorf("creating backup directory: %w", err)
	}
	backupPath := filepath.Join(backupDir, filepath.Base(currentPath))

	// Windows에서는 실행 중인 파일 교체 실패에 대비해 먼저 복구본을 만든다.
	if err := os.Rename(currentPath, backupPath); err != nil {
		cleanupErr := os.Remove(backupDir)
		return errors.Join(fmt.Errorf("backing up current binary: %w", err), cleanupErr)
	}

	if err := os.Rename(tempPath, currentPath); err != nil {
		rollbackErr := os.Rename(backupPath, currentPath)
		if rollbackErr != nil {
			return errors.Join(
				fmt.Errorf("replacing current binary: %w", err),
				fmt.Errorf("restoring backup %s: %w", backupPath, rollbackErr),
			)
		}
		cleanupErr := os.Remove(backupDir)
		return errors.Join(fmt.Errorf("replacing current binary: %w", err), cleanupErr)
	}

	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("binary replaced but backup retained at %s: %w", backupPath, err)
	}
	if err := os.Remove(backupDir); err != nil {
		return fmt.Errorf("binary replaced but backup directory retained at %s: %w", backupDir, err)
	}

	return nil
}
