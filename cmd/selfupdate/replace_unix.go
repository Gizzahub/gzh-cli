// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

//go:build !windows

package selfupdate

import (
	"fmt"
	"os"
	"path/filepath"
)

func replaceBinary(tempPath, currentPath string) error {
	if err := validateReplacementPaths(tempPath, currentPath); err != nil {
		return err
	}
	if err := os.Rename(tempPath, currentPath); err != nil {
		return fmt.Errorf("replacing current binary: %w", err)
	}

	// 파일 이름 교체까지 디스크에 반영되도록 대상 디렉터리를 동기화한다.
	dir, err := os.Open(filepath.Dir(currentPath))
	if err != nil {
		return fmt.Errorf("opening binary directory after replacement: %w", err)
	}
	if err := dir.Sync(); err != nil {
		_ = dir.Close()
		return fmt.Errorf("syncing binary directory after replacement: %w", err)
	}
	if err := dir.Close(); err != nil {
		return fmt.Errorf("closing binary directory after replacement: %w", err)
	}

	return nil
}
