// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

//go:build !windows

package bulkclone

import "os"

func legacyWindowsComponentTooLong(_ *os.Root, _ string, _ error) (bool, error) {
	return false, nil
}
