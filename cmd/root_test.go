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

// 하위 명령만 가진 그룹이 모르는 인자를 받으면 도움말을 찍고 0으로 끝나지 않고
// 뿌리와 같은 오류를 돌려줘야 한다. 인자가 없으면 지금처럼 도움말만 찍는다.
func TestGroupCommandRejectsUnknownSubcommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "bare group prints help", args: []string{"profile"}},
		{name: "unknown subcommand", args: []string{"profile", "nosuch"}, wantErr: `unknown command "nosuch" for "gz profile"`},
		{name: "typo suggests the subcommand", args: []string{"profile", "cpuu"}, wantErr: "Did you mean this?\n\tcpu\n"},
		{name: "known subcommand help", args: []string{"profile", "cpu", "--help"}},
		{name: "nested group", args: []string{"git", "config", "nosuch"}, wantErr: `unknown command "nosuch" for "gz git config"`},
		{name: "bare nested group prints help", args: []string{"git", "repo"}},
		{name: "synclone takes no positional args", args: []string{"synclone", "githb"}, wantErr: `unknown command "githb" for "gz synclone"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := NewRootCmdForGeneration(context.Background(), "dev", app.NewTestAppContext())
			root.SetArgs(tt.args)
			root.SetOut(&bytes.Buffer{})
			root.SetErr(&bytes.Buffer{})

			_, err := root.ExecuteC()
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

// 새 그룹 명령도 빠짐없이 덮이는지 나무 전체를 확인한다.
func TestEveryGroupCommandIsRunnable(t *testing.T) {
	root := NewRootCmdForGeneration(context.Background(), "dev", app.NewTestAppContext())
	var walk func(*cobra.Command)
	walk = func(parent *cobra.Command) {
		for _, child := range parent.Commands() {
			if child.HasSubCommands() {
				require.True(t, child.Runnable(), "%s must reject unknown subcommands", child.CommandPath())
			}
			walk(child)
		}
	}
	walk(root)
}
