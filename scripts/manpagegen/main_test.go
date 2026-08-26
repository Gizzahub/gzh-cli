// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package main

import (
	"compress/gzip"
	"context"
	"io"
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
		root := &cobra.Command{Use: "gz", Short: "test command", Long: "test command tree"}
		root.PersistentFlags().Bool("verbose", false, "show verbose output")
		root.AddCommand(&cobra.Command{
			Use:   "standard",
			Short: "standard command",
			Run:   func(*cobra.Command, []string) {},
		})
		return root
	}
	t.Setenv("SOURCE_DATE_EPOCH", "1787756400")

	outputPath := filepath.Join(t.TempDir(), "gz.1.gz")
	require.NoError(t, run([]string{outputPath}))
	require.True(t, called)

	compressed, err := os.Open(outputPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, compressed.Close())
	})
	reader, err := gzip.NewReader(compressed)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, reader.Close())
	})
	page, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Contains(t, string(page), ".TH \"GZ\" \"1\"")
	require.Contains(t, string(page), ".B \"gz standard\"")
	require.Contains(t, string(page), "\\-\\-verbose")
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
	t.Setenv("SOURCE_DATE_EPOCH", "1787756400")

	outputPath := filepath.Join(t.TempDir(), "gz.1.gz")
	require.NoError(t, run([]string{outputPath}))

	compressed, err := os.Open(outputPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, compressed.Close())
	})
	reader, err := gzip.NewReader(compressed)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, reader.Close())
	})
	page, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NotContains(t, string(page), "local-only")
}
