// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package bulkclone

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestStatePathSafetyWindowsComponentLimitUsesUTF16Units(t *testing.T) {
	const maximum = 255
	require.False(t, stateComponentExceedsWindowsLimit(strings.Repeat("a", maximum), maximum))
	require.True(t, stateComponentExceedsWindowsLimit(strings.Repeat("a", maximum+1), maximum))
	require.False(t, stateComponentExceedsWindowsLimit(strings.Repeat("😀", maximum/2), maximum))
	require.True(t, stateComponentExceedsWindowsLimit(strings.Repeat("😀", maximum/2+1), maximum))
}

func TestStatePathSafetyWindowsLongLegacyComponentIsOptionalOnly(t *testing.T) {
	root, err := os.OpenRoot(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, root.Close()) })

	component := strings.Repeat("x", 300)
	_, lookupErr := root.Lstat(component)
	require.ErrorIs(t, lookupErr, windows.ERROR_INVALID_NAME)

	_, present, err := inspectStateComponent(root, component, component, true)
	require.NoError(t, err)
	require.False(t, present)

	_, _, err = inspectStateComponent(root, component, component, false)
	require.ErrorIs(t, err, windows.ERROR_INVALID_NAME)
}

func TestStatePathSafetyWindowsLengthVerificationKeepsLookupError(t *testing.T) {
	root, err := os.OpenRoot(t.TempDir())
	require.NoError(t, err)
	require.NoError(t, root.Close())

	missing, err := optionalLegacyComponentMissing(root, "invalid?name", windows.ERROR_INVALID_NAME)
	require.False(t, missing)
	require.Error(t, err)
	require.True(t, errors.Is(err, windows.ERROR_INVALID_NAME))
}

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
