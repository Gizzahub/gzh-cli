// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package synclone

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	synclonepkg "github.com/gizzahub/gzh-cli/pkg/synclone"
)

func runStateCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()

	cmd := newSyncCloneStateCmd(nil)
	cmd.SetArgs(args)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	original := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = writer
	runErr := cmd.Execute()
	require.NoError(t, writer.Close())
	os.Stdout = original

	var buf bytes.Buffer
	_, err = io.Copy(&buf, reader)
	require.NoError(t, err)

	return buf.String(), runErr
}

func TestStateTargetReadsOnlyTheSelectedDirectory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	target := t.TempDir()

	global := synclonepkg.NewStateManager("")
	require.NoError(t, global.SaveState(synclonepkg.NewCloneState("github", "home-org", home, "pull", 1, 0)))
	local := synclonepkg.NewStateManager(filepath.Join(target, ".gzh", "state"))
	require.NoError(t, local.SaveState(synclonepkg.NewCloneState("github", "target-org", target, "pull", 1, 0)))

	selected, err := runStateCommand(t, "list", "--target", target)
	require.NoError(t, err)
	require.Contains(t, selected, "target-org")
	require.NotContains(t, selected, "home-org")

	fallback, err := runStateCommand(t, "list")
	require.NoError(t, err)
	require.Contains(t, fallback, "home-org")
	require.NotContains(t, fallback, "target-org")

	empty := t.TempDir()
	missing, err := runStateCommand(t, "list", "--target", empty)
	require.NoError(t, err)
	require.Contains(t, missing, "No saved clone states found")
	require.NotContains(t, missing, "home-org")

	shown, err := runStateCommand(t, "show", "--target", target, "--provider", "github", "--organization", "target-org")
	require.NoError(t, err)
	require.Contains(t, shown, "target-org")
	require.Contains(t, shown, target)

	_, err = runStateCommand(t, "show", "--target", target, "--provider", "github", "--organization", "home-org")
	require.Error(t, err)
}

func TestStateTargetCleanAllLeavesTheDefaultDirectory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	target := t.TempDir()

	global := synclonepkg.NewStateManager("")
	require.NoError(t, global.SaveState(synclonepkg.NewCloneState("github", "home-org", home, "pull", 1, 0)))
	local := synclonepkg.NewStateManager(filepath.Join(target, ".gzh", "state"))
	require.NoError(t, local.SaveState(synclonepkg.NewCloneState("github", "target-org", target, "pull", 1, 0)))
	homeFile := global.GetStateFilePath("github", "home-org")
	targetFile := local.GetStateFilePath("github", "target-org")

	out, err := runStateCommand(t, "clean", "--all", "--target", target)
	require.NoError(t, err)
	require.Contains(t, out, "Cleaned github/target-org\n")
	require.NotContains(t, out, "home-org")
	require.FileExists(t, homeFile)
	_, err = os.Stat(targetFile)
	require.ErrorIs(t, err, os.ErrNotExist)
}
