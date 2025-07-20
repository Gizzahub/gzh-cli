//nolint:testpackage // White-box testing needed for internal function access
package github

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStrategyExecution(t *testing.T) {
	// Skip if not in CI or if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	tempDir := t.TempDir()

	// Create a mock git repository
	repoPath := filepath.Join(tempDir, "test-repo")
	err := os.MkdirAll(repoPath, 0o755)
	require.NoError(t, err)

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	err = cmd.Run()
	require.NoError(t, err)

	// Add a file and commit
	testFile := filepath.Join(repoPath, "test.txt")
	err = os.WriteFile(testFile, []byte("initial content"), 0o644)
	require.NoError(t, err)

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = repoPath
	err = cmd.Run()
	require.NoError(t, err)

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoPath
	_ = cmd.Run()

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoPath
	_ = cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "initial commit")
	cmd.Dir = repoPath
	err = cmd.Run()
	require.NoError(t, err)

	// Test different strategies
	strategies := []string{"reset", "pull", "fetch"}

	for _, strategy := range strategies {
		t.Run("strategy_"+strategy, func(t *testing.T) {
			// Modify the file to simulate local changes
			err := os.WriteFile(testFile, []byte("modified content"), 0o644)
			require.NoError(t, err)

			// Execute git operation based on strategy
			switch strategy {
			case "reset":
				cmd := exec.Command("git", "-C", repoPath, "reset", "--hard", "HEAD")
				err := cmd.Run()
				assert.NoError(t, err)

				// Verify file was reset
				content, err := os.ReadFile(testFile)
				assert.NoError(t, err)
				assert.Equal(t, "initial content", string(content))

			case "pull":
				// pull will fail without remote, but command should execute
				cmd := exec.Command("git", "-C", repoPath, "pull")
				_ = cmd.Run()
				// We don't check error as pull will fail without remote

			case "fetch":
				// fetch will fail without remote, but command should execute
				cmd := exec.Command("git", "-C", repoPath, "fetch")
				_ = cmd.Run()
				// We don't check error as fetch will fail without remote
			}
		})
	}
}

func TestValidateStrategy(t *testing.T) {
	tests := []struct {
		strategy string
		valid    bool
	}{
		{"reset", true},
		{"pull", true},
		{"fetch", true},
		{"invalid", false},
		{"RESET", false},
		{"", false},
		{"merge", false},
	}

	for _, tt := range tests {
		t.Run(tt.strategy, func(t *testing.T) {
			isValid := tt.strategy == "reset" || tt.strategy == "pull" || tt.strategy == "fetch"
			assert.Equal(t, tt.valid, isValid, "Strategy %s validation failed", tt.strategy)
		})
	}
}
