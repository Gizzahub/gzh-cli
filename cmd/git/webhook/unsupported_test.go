// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package webhook

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/pkg/github"
)

// 독립 공개 Cobra 경로에서 정책 파일을 읽거나 가짜 성공을 출력하지 않는다.
func TestWebhookUnsupportedCommands(t *testing.T) {
	cases := [][]string{
		{"config", "policy", "create", "org", "missing.yaml"},
		{"config", "policy", "create", "org", "malformed.yaml"},
		{"config", "policy", "list", "org"},
		{"config", "policy", "apply", "org"},
		{"config", "policy", "apply", "org", "--dry-run", "--force", "--policies", "fixture", "--repos", "repo"},
		{"config", "policy", "preview", "org"},
		{"config", "org", "get", "org"},
		{"config", "org", "update", "org", "missing.yaml"},
		{"config", "org", "update", "org", "malformed.yaml"},
		{"config", "org", "validate", "missing.yaml"},
		{"config", "org", "validate", "malformed.yaml"},
		{"config", "report", "compliance", "org"},
		{"config", "report", "inventory", "org"},
		{"config", "report", "sync", "org"},
		{"monitor", "test", "owner", "repo", "7"},
		{"monitor", "deliveries", "owner", "repo", "7"},
	}
	for _, args := range cases {
		t.Run(args[0]+"/"+args[1]+"/"+args[2], func(t *testing.T) {
			dir := t.TempDir()
			t.Chdir(dir)
			original := []byte("malformed: [\n")
			require.NoError(t, os.WriteFile("malformed.yaml", original, 0o600))
			var output bytes.Buffer
			command := NewWebhookCmd()
			command.SetOut(&output)
			command.SetErr(&output)
			command.SilenceErrors, command.SilenceUsage = true, true
			command.SetArgs(args)
			finishCapture := captureUnsupportedStdout(t)
			err := command.Execute()
			stdout := finishCapture()
			require.ErrorIs(t, err, github.ErrNotImplemented)
			require.Empty(t, stdout, "fmt.Printf must not emit fabricated success, scores, or deliveries")
			require.Empty(t, output.String())
			data, err := os.ReadFile("malformed.yaml")
			require.NoError(t, err)
			require.Equal(t, original, data)
			entries, err := os.ReadDir(dir)
			require.NoError(t, err)
			require.Len(t, entries, 1)
			_, err = os.Stat(filepath.Join(dir, "missing.yaml"))
			require.ErrorIs(t, err, os.ErrNotExist)
		})
	}
}

func TestWebhookUnsupportedHelpAndValidationRemain(t *testing.T) {
	for _, path := range [][]string{{"config", "--help"}, {"monitor", "--help"}, {"config", "policy", "apply", "--help"}} {
		command := NewWebhookCmd()
		var output bytes.Buffer
		command.SetOut(&output)
		command.SetErr(&output)
		command.SetArgs(path)
		require.NoError(t, command.Execute())
		require.NotEmpty(t, output.String())
	}
	command := NewWebhookCmd()
	command.SilenceErrors, command.SilenceUsage = true, true
	command.SetArgs([]string{"monitor", "test", "owner", "repo", "not-an-id"})
	require.ErrorContains(t, command.Execute(), "invalid webhook ID")
}

// 테스트는 병렬 실행하지 않으며 실제 fmt.Printf 출력도 pipe로 수집한다.
func captureUnsupportedStdout(t *testing.T) func() string {
	t.Helper()
	original := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = writer
	type captured struct {
		data []byte
		err  error
	}
	done := make(chan captured, 1)
	go func() {
		data, readErr := io.ReadAll(reader)
		done <- captured{data: data, err: readErr}
	}()
	finished := false
	var output string
	finish := func() string {
		if finished {
			return output
		}
		finished = true
		os.Stdout = original
		closeErr := writer.Close()
		result := <-done
		readerErr := reader.Close()
		output = string(result.data)
		require.NoError(t, closeErr)
		require.NoError(t, result.err)
		require.NoError(t, readerErr)
		return output
	}
	t.Cleanup(func() { finish() })
	return finish
}
