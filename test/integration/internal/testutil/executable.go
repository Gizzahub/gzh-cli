// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package testutil

import (
	"path/filepath"
	"runtime"
)

// ExecutablePath는 현재 플랫폼에서 빌드된 실행 파일의 경로를 돌려준다.
func ExecutablePath(dir, name string) string {
	return executablePathForGOOS(dir, name, runtime.GOOS)
}

func executablePathForGOOS(dir, name, goos string) string {
	if goos == "windows" {
		name += ".exe"
	}

	return filepath.Join(dir, name)
}
