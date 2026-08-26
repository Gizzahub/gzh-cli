// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/internal/app"
)

func TestRunUsesReleaseRootFactory(t *testing.T) {
	originalFactory := newRootCommand
	t.Cleanup(func() {
		newRootCommand = originalFactory
	})

	called := false
	newRootCommand = func(context.Context, string, *app.AppContext) *cobra.Command {
		called = true
		root := &cobra.Command{Use: "gz"}
		root.AddCommand(&cobra.Command{Use: "standard", Run: func(*cobra.Command, []string) {}})
		return root
	}

	var output bytes.Buffer
	require.NoError(t, run([]string{"bash"}, &output))
	require.True(t, called)
	require.Contains(t, output.String(), "standard")
}

func TestDefaultReleaseRootExcludesUserExtensions(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	configDir := filepath.Join(home, ".config", "gzh-manager")
	require.NoError(t, os.MkdirAll(configDir, 0o750))
	require.NoError(t, os.WriteFile(
		filepath.Join(configDir, "extensions.yaml"),
		[]byte("aliases:\n  local-only:\n    command: git status\n    description: local fixture\nexternal: []\n"),
		0o600,
	))

	var output bytes.Buffer
	require.NoError(t, run([]string{"bash"}, &output))
	require.NotContains(t, output.String(), "local-only")
}
