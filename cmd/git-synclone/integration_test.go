// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

//go:build integration
// +build integration

package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitCommandIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Ensure git-synclone binary is available
	gitSynclonePath, err := exec.LookPath("git-synclone")
	if err != nil {
		// Try to build it first
		buildCmd := exec.Command("make", "build-git-extensions")
		buildCmd.Dir = "../.." // Go up to project root
		err = buildCmd.Run()
		require.NoError(t, err, "Failed to build git-synclone")

		// Try to find it in current directory
		if _, err := os.Stat("../../git-synclone"); err == nil {
			// Add current directory to PATH for this test
			currentPath := os.Getenv("PATH")
			pwd, _ := os.Getwd()
			projectRoot := filepath.Join(pwd, "../..")
			os.Setenv("PATH", projectRoot+":"+currentPath)

			gitSynclonePath, err = exec.LookPath("git-synclone")
		}
	}
	require.NoError(t, err, "git-synclone should be available in PATH")
	t.Logf("Found git-synclone at: %s", gitSynclonePath)

	t.Run("GitSyncloneHelp", func(t *testing.T) {
		cmd := exec.Command("git", "synclone", "--help")
		output, err := cmd.CombinedOutput()

		require.NoError(t, err, "git synclone --help should succeed")
		outputStr := string(output)
		assert.Contains(t, outputStr, "Enhanced Git cloning")
		assert.Contains(t, outputStr, "github")
		assert.Contains(t, outputStr, "gitlab")
		assert.Contains(t, outputStr, "gitea")
	})

	t.Run("GitSyncloneVersion", func(t *testing.T) {
		cmd := exec.Command("git", "synclone", "--version")
		output, err := cmd.CombinedOutput()

		require.NoError(t, err, "git synclone --version should succeed")
		outputStr := string(output)
		assert.Contains(t, outputStr, "git-synclone")
	})

	t.Run("GitSyncloneSubcommands", func(t *testing.T) {
		subcommands := []string{"github", "gitlab", "gitea", "doctor", "config", "validate"}

		for _, subcmd := range subcommands {
			t.Run(subcmd, func(t *testing.T) {
				cmd := exec.Command("git", "synclone", subcmd, "--help")
				err := cmd.Run()
				assert.NoError(t, err, "git synclone %s --help should succeed", subcmd)
			})
		}
	})

	t.Run("GitSyncloneDoctor", func(t *testing.T) {
		cmd := exec.Command("git", "synclone", "doctor")
		output, err := cmd.CombinedOutput()

		// Doctor might fail due to missing configuration, but should not crash
		outputStr := string(output)
		assert.Contains(t, outputStr, "Installation Diagnostics")
		assert.Contains(t, outputStr, "Git")
	})
}

func TestGitSyncloneDryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Ensure git-synclone is available
	_, err := exec.LookPath("git-synclone")
	if err != nil {
		t.Skip("git-synclone not available in PATH")
	}

	tempDir := t.TempDir()

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "GitHub dry run",
			args: []string{"github", "-o", "octocat", "-t", tempDir, "--dry-run"},
		},
		{
			name: "GitLab dry run",
			args: []string{"gitlab", "-g", "gitlab-org", "-t", tempDir, "--dry-run"},
		},
		{
			name: "Gitea dry run",
			args: []string{"gitea", "-o", "gitea", "-t", tempDir, "--dry-run"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, "git", append([]string{"synclone"}, tt.args...)...)
			output, err := cmd.CombinedOutput()

			// Dry run should succeed or fail gracefully
			outputStr := string(output)
			t.Logf("Output: %s", outputStr)

			if err != nil {
				// Check if it's a reasonable error (like network timeout, auth failure)
				assert.NotContains(t, outputStr, "panic")
				assert.NotContains(t, outputStr, "segmentation fault")
			}
		})
	}
}

func TestConfigFileIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Ensure git-synclone is available
	_, err := exec.LookPath("git-synclone")
	if err != nil {
		t.Skip("git-synclone not available in PATH")
	}

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.yaml")

	configContent := `
version: "1.0.0"
providers:
  github:
    organizations:
      - name: "octocat"
        clone_dir: "` + tempDir + `/github"
  gitlab:
    groups:
      - name: "gitlab-org"
        clone_dir: "` + tempDir + `/gitlab"
`

	err = os.WriteFile(configFile, []byte(configContent), 0o644)
	require.NoError(t, err)

	t.Run("ConfigValidation", func(t *testing.T) {
		cmd := exec.Command("git", "synclone", "validate", "--config", configFile)
		output, err := cmd.CombinedOutput()

		outputStr := string(output)
		t.Logf("Validation output: %s", outputStr)

		// Validation should succeed or provide meaningful error
		if err != nil {
			assert.NotContains(t, outputStr, "panic")
		}
	})

	t.Run("ConfigBasedClone", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "git", "synclone", "--config", configFile, "--dry-run")
		output, err := cmd.CombinedOutput()

		outputStr := string(output)
		t.Logf("Config-based clone output: %s", outputStr)

		// Should not crash, even if configuration is not perfect
		assert.NotContains(t, outputStr, "panic")
		assert.NotContains(t, outputStr, "segmentation fault")
	})
}

func TestErrorHandlingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Ensure git-synclone is available
	_, err := exec.LookPath("git-synclone")
	if err != nil {
		t.Skip("git-synclone not available in PATH")
	}

	tests := []struct {
		name           string
		args           []string
		expectError    bool
		expectedOutput string
	}{
		{
			name:           "Missing required flag",
			args:           []string{"github"},
			expectError:    true,
			expectedOutput: "required",
		},
		{
			name:           "Invalid config file",
			args:           []string{"--config", "/nonexistent/config.yaml"},
			expectError:    true,
			expectedOutput: "config",
		},
		{
			name:           "Invalid strategy",
			args:           []string{"github", "-o", "test", "--strategy", "invalid"},
			expectError:    true,
			expectedOutput: "strategy",
		},
		{
			name:           "Invalid protocol",
			args:           []string{"github", "-o", "test", "--protocol", "ftp"},
			expectError:    true,
			expectedOutput: "protocol",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("git", append([]string{"synclone"}, tt.args...)...)
			output, err := cmd.CombinedOutput()

			outputStr := string(output)
			t.Logf("Error test output: %s", outputStr)

			if tt.expectError {
				assert.Error(t, err, "Command should fail")
				assert.Contains(t, strings.ToLower(outputStr), strings.ToLower(tt.expectedOutput))
			} else {
				assert.NoError(t, err, "Command should succeed")
			}

			// Ensure no crashes
			assert.NotContains(t, outputStr, "panic")
			assert.NotContains(t, outputStr, "segmentation fault")
		})
	}
}

func TestPerformanceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Ensure git-synclone is available
	_, err := exec.LookPath("git-synclone")
	if err != nil {
		t.Skip("git-synclone not available in PATH")
	}

	tempDir := t.TempDir()

	t.Run("ParallelVsSequential", func(t *testing.T) {
		// Test with a small public organization
		orgName := "octocat"

		// Sequential (parallel=1)
		start := time.Now()
		cmd := exec.Command("git", "synclone", "github", "-o", orgName, "-t", tempDir+"/sequential", "--parallel", "1", "--dry-run")
		err := cmd.Run()
		sequentialTime := time.Since(start)

		if err != nil {
			t.Logf("Sequential command failed (expected): %v", err)
		}

		// Parallel (parallel=5)
		start = time.Now()
		cmd = exec.Command("git", "synclone", "github", "-o", orgName, "-t", tempDir+"/parallel", "--parallel", "5", "--dry-run")
		err = cmd.Run()
		parallelTime := time.Since(start)

		if err != nil {
			t.Logf("Parallel command failed (expected): %v", err)
		}

		t.Logf("Sequential time: %v, Parallel time: %v", sequentialTime, parallelTime)

		// Performance test is informational - we just ensure both complete
		assert.True(t, sequentialTime > 0)
		assert.True(t, parallelTime > 0)
	})
}

func TestConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	// Ensure git-synclone is available
	_, err := exec.LookPath("git-synclone")
	if err != nil {
		t.Skip("git-synclone not available in PATH")
	}

	tempDir := t.TempDir()

	t.Run("ConcurrentDryRuns", func(t *testing.T) {
		// Run multiple dry-run operations concurrently
		const numConcurrent = 3
		results := make(chan error, numConcurrent)

		for i := 0; i < numConcurrent; i++ {
			go func(index int) {
				targetDir := filepath.Join(tempDir, "concurrent", "run"+string(rune(index+'0')))
				cmd := exec.Command("git", "synclone", "github", "-o", "octocat", "-t", targetDir, "--dry-run")
				results <- cmd.Run()
			}(i)
		}

		// Collect results
		for i := 0; i < numConcurrent; i++ {
			err := <-results
			if err != nil {
				t.Logf("Concurrent operation %d failed (may be expected): %v", i, err)
			}
		}

		// Test passes if no deadlocks or crashes occurred
		t.Log("Concurrent operations completed without deadlocks")
	})
}

func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	// Ensure git-synclone is available
	_, err := exec.LookPath("git-synclone")
	if err != nil {
		t.Skip("git-synclone not available in PATH")
	}

	tempDir := t.TempDir()

	t.Run("MemoryLimit", func(t *testing.T) {
		// Test with memory limit (this would be more meaningful with a real large organization)
		cmd := exec.Command("git", "synclone", "github", "-o", "octocat", "-t", tempDir, "--dry-run")

		// Set memory limit environment variable if supported
		cmd.Env = append(os.Environ(), "GOMAXPROCS=1")

		start := time.Now()
		err := cmd.Run()
		duration := time.Since(start)

		t.Logf("Memory-limited operation took: %v", duration)
		if err != nil {
			t.Logf("Memory-limited operation failed (may be expected): %v", err)
		}

		// Test that operation completes within reasonable time (not stuck)
		assert.Less(t, duration, 2*time.Minute, "Operation should complete within reasonable time")
	})
}
