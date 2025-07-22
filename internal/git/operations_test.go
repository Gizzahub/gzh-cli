//nolint:testpackage // White-box testing needed for internal function access
package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-manager-go/internal/gitplatform"
)

func TestOperations_Clone(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		path    string
		verbose bool
		wantErr bool
	}{
		{
			name:    "clone with valid URL and path",
			url:     "https://github.com/example/repo.git",
			path:    "test-repo",
			verbose: false,
			wantErr: true, // Will fail without real repo
		},
		{
			name:    "clone with verbose output",
			url:     "https://github.com/example/repo.git",
			path:    "test-repo",
			verbose: true,
			wantErr: true, // Will fail without real repo
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir, err := os.MkdirTemp("", "git-test-*")
			require.NoError(t, err)

			defer func() {
				if err := os.RemoveAll(tmpDir); err != nil {
					t.Logf("Warning: failed to remove temp dir: %v", err)
				}
			}()

			ops := NewOperations(tt.verbose)
			targetPath := filepath.Join(tmpDir, tt.path)

			// For actual testing, we would need to mock the git command
			// This is a basic structure test
			err = ops.Clone(tt.url, targetPath)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOperations_ExecuteStrategy(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		strategy gitplatform.CloneStrategy
		setup    func(string) error
		wantErr  bool
	}{
		{
			name:     "execute reset strategy",
			path:     "test-repo",
			strategy: gitplatform.StrategyReset,
			setup:    nil,
			wantErr:  true, // Will fail without real repo
		},
		{
			name:     "execute pull strategy",
			path:     "test-repo",
			strategy: gitplatform.StrategyPull,
			setup:    nil,
			wantErr:  true, // Will fail without real repo
		},
		{
			name:     "execute fetch strategy",
			path:     "test-repo",
			strategy: gitplatform.StrategyFetch,
			setup:    nil,
			wantErr:  true, // Will fail without real repo
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir, err := os.MkdirTemp("", "git-test-*")
			require.NoError(t, err)

			defer func() {
				if err := os.RemoveAll(tmpDir); err != nil {
					t.Logf("Warning: failed to remove temp dir: %v", err)
				}
			}()

			ops := NewOperations(false)
			targetPath := filepath.Join(tmpDir, tt.path)

			if tt.setup != nil {
				err := tt.setup(targetPath)
				require.NoError(t, err)
			}

			err = ops.ExecuteStrategy(targetPath, tt.strategy)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
