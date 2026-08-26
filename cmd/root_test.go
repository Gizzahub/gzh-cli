//nolint:testpackage // White-box testing needed for internal function access
package cmd

import (
	"bytes"
	"context"
	"go/ast"
	"go/parser"
	"go/token"
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

func TestPublicRootConstructorsSelectExtensionRegistrar(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "root.go", nil, 0)
	require.NoError(t, err)

	require.Equal(t, "registerUserExtensions", rootConstructorRegistrar(t, file, "NewRootCmd"))
	require.Equal(t, "nil", rootConstructorRegistrar(t, file, "NewRootCmdForGeneration"))
}

func rootConstructorRegistrar(t *testing.T, file *ast.File, constructorName string) string {
	t.Helper()

	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name.Name != constructorName {
			continue
		}

		require.NotNil(t, function.Body)
		require.Len(t, function.Body.List, 1)
		returnStatement, ok := function.Body.List[0].(*ast.ReturnStmt)
		require.True(t, ok, "%s must return the shared root constructor", constructorName)
		require.Len(t, returnStatement.Results, 1)
		call, ok := returnStatement.Results[0].(*ast.CallExpr)
		require.True(t, ok, "%s must call the shared root constructor", constructorName)
		callee, ok := call.Fun.(*ast.Ident)
		require.True(t, ok)
		require.Equal(t, "newRootCmd", callee.Name)
		require.Len(t, call.Args, 4)
		registrar, ok := call.Args[3].(*ast.Ident)
		require.True(t, ok)

		return registrar.Name
	}

	require.FailNow(t, "constructor not found", constructorName)

	return ""
}

func findCommand(root *cobra.Command, name string) *cobra.Command {
	for _, child := range root.Commands() {
		if child.Name() == name {
			return child
		}
	}

	return nil
}
