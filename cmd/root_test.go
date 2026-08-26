//nolint:testpackage // White-box testing needed for internal function access
package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/internal/app"
)

func TestRootCommandOutput(t *testing.T) {
	cmd := NewRootCmd(context.Background(), "", app.NewTestAppContext())
	b := bytes.NewBufferString("")

	cmd.SetArgs([]string{"-h"})
	cmd.SetOut(b)

	cmdErr := cmd.RunE(cmd, nil)
	require.NoError(t, cmdErr)
}

func TestNewRootCmdExtensionSelection(t *testing.T) {
	ctx := context.Background()
	appCtx := app.NewTestAppContext()
	loaded := false
	registerExtension := func(root *cobra.Command) error {
		loaded = true
		root.AddCommand(&cobra.Command{Use: "local-only"})
		return nil
	}

	productRoot := newRootCmd(ctx, "dev", appCtx, registerExtension)
	require.True(t, loaded)
	require.NotNil(t, findCommand(productRoot, "local-only"))

	loaded = false
	generationRoot := newRootCmd(ctx, "dev", appCtx, nil)
	require.False(t, loaded)
	require.Nil(t, findCommand(generationRoot, "local-only"))
}

func findCommand(root *cobra.Command, name string) *cobra.Command {
	for _, child := range root.Commands() {
		if child.Name() == name {
			return child
		}
	}

	return nil
}
