//nolint:testpackage // White-box testing needed for internal function access
package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/internal/gitplatform"
)

func TestOperations_Clone(t *testing.T) {
	for _, verbose := range []bool{false, true} {
		t.Run("local fixture", func(t *testing.T) {
			setupIsolatedGitConfig(t)
			root := t.TempDir()
			source := filepath.Join(root, "source")
			initFixtureRepo(t, source)
			target := filepath.Join(root, "clone")

			err := NewOperations(verbose).Clone(context.Background(), source, target)
			require.NoError(t, err)

			contents, err := os.ReadFile(filepath.Join(target, "README.md"))
			require.NoError(t, err)
			assert.Equal(t, "initial\n", string(contents))
		})
	}
}

func TestOperations_ExecuteStrategy(t *testing.T) {
	tests := []struct {
		name     string
		strategy gitplatform.CloneStrategy
	}{
		{
			name:     "execute reset strategy",
			strategy: gitplatform.StrategyReset,
		},
		{
			name:     "execute pull strategy",
			strategy: gitplatform.StrategyPull,
		},
		{
			name:     "execute fetch strategy",
			strategy: gitplatform.StrategyFetch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupIsolatedGitConfig(t)
			root := t.TempDir()
			source := filepath.Join(root, "source")
			target := filepath.Join(root, "clone")
			initFixtureRepo(t, source)
			require.NoError(t, NewOperations(false).Clone(context.Background(), source, target))

			commitFixtureFile(t, source, "README.md", "remote\n", "remote update")
			if tt.strategy == gitplatform.StrategyReset {
				require.NoError(t, os.WriteFile(filepath.Join(target, "README.md"), []byte("local\n"), 0o600))
			}

			err := NewOperations(false).ExecuteStrategy(context.Background(), target, tt.strategy)
			require.NoError(t, err)

			if tt.strategy == gitplatform.StrategyFetch {
				assert.NotEqual(t, runFixtureGit(t, target, "rev-parse", "HEAD"), runFixtureGit(t, source, "rev-parse", "HEAD"))
				assert.Equal(t, runFixtureGit(t, source, "rev-parse", "HEAD"), runFixtureGit(t, target, "rev-parse", "origin/HEAD"))
				return
			}

			contents, err := os.ReadFile(filepath.Join(target, "README.md"))
			require.NoError(t, err)
			assert.Equal(t, "remote\n", string(contents))
		})
	}
}
