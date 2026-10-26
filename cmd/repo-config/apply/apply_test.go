// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package apply

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyModesUnsupportedWithoutSuccessOrMutation(t *testing.T) {
	for _, args := range [][]string{
		{},
		{"--org", "fixture"},
		{"--org", "fixture", "--dry-run"},
		{"--org", "fixture", "--force"},
		{"--org", "fixture", "--interactive"},
		{"--org", "fixture", "--dry-run", "--force", "--interactive", "--filter", ".*", "--template", "security"},
	} {
		t.Run(applyTestName(args), func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			t.Setenv("HOME", root)
			t.Setenv("USERPROFILE", root)
			t.Setenv("GITHUB_TOKEN", "")
			emptyPath := filepath.Join(root, "empty-path")
			require.NoError(t, os.Mkdir(emptyPath, 0o700))
			t.Setenv("PATH", emptyPath)
			fixture := filepath.Join(root, "config.yaml")
			original := []byte("must not parse or change this fixture")
			require.NoError(t, os.WriteFile(fixture, original, 0o600))
			command := NewCmd()
			var captured bytes.Buffer
			command.SetOut(&captured)
			command.SetErr(&captured)
			command.SilenceUsage, command.SilenceErrors = true, true
			command.SetArgs(append(append([]string{}, args...), "--config", fixture))
			stdout, err := captureApplyStdout(t, command.Execute)
			require.ErrorContains(t, err, "not supported")
			require.Empty(t, stdout, "no fabricated preview or completion on process stdout")
			require.Empty(t, captured.String())
			after, err := os.ReadFile(fixture)
			require.NoError(t, err)
			require.Equal(t, original, after)
			entries, err := os.ReadDir(root)
			require.NoError(t, err)
			require.Len(t, entries, 2)
		})
	}
}

func applyTestName(args []string) string {
	if len(args) == 0 {
		return "bare"
	}
	var name bytes.Buffer
	for _, arg := range args {
		name.WriteString(arg)
		name.WriteByte(' ')
	}
	return name.String()
}

// fmt.Printf 기반의 기존 가짜 성공도 포착한다. 외부 프로세스는 실행하지 않는다.
func captureApplyStdout(t *testing.T, run func() error) (string, error) {
	t.Helper()
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	previous := os.Stdout
	os.Stdout = writer
	defer func() { os.Stdout = previous }()
	runErr := run()
	require.NoError(t, writer.Close())
	output, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())
	return string(output), runErr
}

func TestApplyHelpAndFlagsRemainAvailable(t *testing.T) {
	command := NewCmd()
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	command.SetArgs([]string{"--help"})
	require.NoError(t, command.Execute())
	require.Contains(t, command.Long, "currently unsupported")
	for _, name := range []string{"org", "config", "token", "dry-run", "verbose", "parallel", "timeout", "filter", "template", "interactive", "force"} {
		require.NotNil(t, command.Flags().Lookup(name))
	}
}
