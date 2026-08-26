// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"reflect"
	"runtime"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/cmd"
	"github.com/gizzahub/gzh-cli/internal/app"
)

func TestRunUsesReleaseRootFactory(t *testing.T) {
	require.Equal(t, functionName(cmd.NewRootCmdForGeneration), functionName(newRootCommand))

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

func functionName(function any) string {
	return runtime.FuncForPC(reflect.ValueOf(function).Pointer()).Name()
}
