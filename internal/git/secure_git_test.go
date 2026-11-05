//nolint:testpackage // The executor test calls the shipped entry in this package.
package git

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSecureGitExecutorPreservesProcessError(t *testing.T) {
	setupIsolatedGitConfig(t)
	executor, err := NewSecureGitExecutor()
	require.NoError(t, err)
	repo := t.TempDir()
	secretURL := "https://user:s3cret-token@example.com/org/repo.git"

	t.Run("nonzero exit", func(t *testing.T) {
		err := executor.ExecuteSecure(t.Context(), repo, "remote", "add", "origin", secretURL)
		require.Error(t, err)
		var exitErr *exec.ExitError
		require.ErrorAs(t, err, &exitErr)
		require.NotContains(t, err.Error(), "s3cret-token")
		require.NotContains(t, err.Error(), "user:s3cret")
		require.Contains(t, err.Error(), "<masked>")
		require.Contains(t, err.Error(), "git remote failed")
	})

	t.Run("canceled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		err := executor.ExecuteSecure(ctx, repo, "remote", "add", "origin", secretURL)
		require.Error(t, err)
		require.ErrorIs(t, err, context.Canceled)
		require.NotContains(t, err.Error(), "s3cret-token")
		require.NotContains(t, err.Error(), "user:s3cret")
		require.Contains(t, err.Error(), "<masked>")
		require.False(t, strings.Contains(err.Error(), secretURL))
	})
}
