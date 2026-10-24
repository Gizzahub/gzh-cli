package webhook

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBulkWebhookRejectsUnimplementedOperations(t *testing.T) {
	for _, operation := range []string{"create", "update", "delete", "sync"} {
		for _, dryRun := range []bool{false, true} {
			name := operation
			if dryRun {
				name += "/dry-run"
			}

			t.Run(name, func(t *testing.T) {
				command := newWebhookBulkCmd()
				var output bytes.Buffer
				command.SetOut(&output)
				command.SetErr(&output)
				command.SilenceErrors = true
				command.SilenceUsage = true
				args := []string{"--org", "example", "--operation", operation}
				if dryRun {
					args = append(args, "--dry-run-bulk")
				}
				command.SetArgs(args)

				// 기존 구현의 전역 stdout 성공 문구도 함께 검사한다.
				captured := captureWebhookOutput(t, func() error {
					err := command.Execute()
					require.ErrorContains(t, err, "bulk webhook "+operation+" is not implemented")
					return nil
				})
				require.NotContains(t, string(captured)+output.String(), "✅")
				require.NotContains(t, string(captured)+output.String(), "Would")
			})
		}
	}
}

func TestBulkWebhookHelpDisclosesUnsupportedStatus(t *testing.T) {
	command := newWebhookBulkCmd()
	var output bytes.Buffer
	command.SetOut(&output)
	command.SetArgs([]string{"--help"})
	require.NoError(t, command.Execute())
	require.Contains(t, output.String(), "not implemented")
	require.Contains(t, output.String(), "including requests with --dry-run-bulk")
	require.NotContains(t, output.String(), "Template-based webhook configuration")
}
