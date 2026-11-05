// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package bulkclone

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatePathSafetyWindowsInvalidLegacyNamePreservesOtherState(t *testing.T) {
	root := t.TempDir()
	other := filepath.Join(root, "github_other.json")
	before := writePathSafetyState(t, other, "github", "other")

	sm := NewStateManager(root)
	err := sm.SaveState(NewCloneState("github", "invalid?name", t.TempDir(), "pull", 1, 0))
	require.Error(t, err)
	assertPathSafetyBytes(t, other, before)
	entries, readErr := os.ReadDir(root)
	require.NoError(t, readErr)
	require.Len(t, entries, 1)
}
