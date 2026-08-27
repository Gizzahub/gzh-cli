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

type windowsReplaceOps struct {
	mkdirTemp func(string, string) (string, error)
	rename    func(string, string) error
	remove    func(string) error
}

var defaultWindowsReplaceOps = windowsReplaceOps{
	mkdirTemp: os.MkdirTemp,
	rename:    os.Rename,
	remove:    os.Remove,
}

func replaceBinary(tempPath, currentPath string) (replacementResult, error) {
	return replaceBinaryWithOps(tempPath, currentPath, defaultWindowsReplaceOps)
}

func replaceBinaryWithOps(tempPath, currentPath string, ops windowsReplaceOps) (replacementResult, error) {
	if err := validateReplacementPaths(tempPath, currentPath); err != nil {
		return replacementResult{}, err
	}

	backupDir, err := ops.mkdirTemp(filepath.Dir(currentPath), ".gz-backup-*")
	if err != nil {
		return replacementResult{}, fmt.Errorf("creating backup directory: %w", err)
	}
	backupPath := filepath.Join(backupDir, filepath.Base(currentPath))

	// Windows에서는 실행 중인 파일 교체 실패에 대비해 먼저 복구본을 만든다.
	if err := ops.rename(currentPath, backupPath); err != nil {
		cleanupErr := ops.remove(backupDir)
		return replacementResult{}, errors.Join(fmt.Errorf("backing up current binary: %w", err), cleanupErr)
	}

	if err := ops.rename(tempPath, currentPath); err != nil {
		rollbackErr := ops.rename(backupPath, currentPath)
		if rollbackErr != nil {
			return replacementResult{}, errors.Join(
				fmt.Errorf("replacing current binary: %w", err),
				fmt.Errorf("restoring backup %s: %w", backupPath, rollbackErr),
			)
		}
		cleanupErr := ops.remove(backupDir)
		return replacementResult{}, errors.Join(fmt.Errorf("replacing current binary: %w", err), cleanupErr)
	}

	if err := ops.remove(backupPath); err != nil {
		return replacementResult{cleanupWarning: fmt.Errorf("binary replaced but backup retained at %s: %w", backupPath, err)}, nil
	}
	if err := ops.remove(backupDir); err != nil {
		return replacementResult{cleanupWarning: fmt.Errorf("binary replaced but backup directory retained at %s: %w", backupDir, err)}, nil
	}

	return replacementResult{}, nil
}
