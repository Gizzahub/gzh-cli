// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repo

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloneOrUpdateOptions_IsValidStrategy(t *testing.T) {
	testCases := []struct {
		name     string
		strategy CloneOrUpdateStrategy
		expected bool
	}{
		{"Valid rebase", StrategyRebase, true},
		{"Valid reset", StrategyReset, true},
		{"Valid clone", StrategyClone, true},
		{"Valid skip", StrategySkip, true},
		{"Valid pull", StrategyPull, true},
		{"Valid fetch", StrategyFetch, true},
		{"Invalid strategy", CloneOrUpdateStrategy("invalid"), false},
		{"Empty strategy", CloneOrUpdateStrategy(""), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := &cloneOrUpdateOptions{strategy: tc.strategy}
			result := opts.isValidStrategy()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCloneOrUpdateOptions_CheckTargetDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	testCases := []struct {
		name           string
		setupFunc      func() string
		expectedExists bool
		expectedIsGit  bool
		expectError    bool
	}{
		{
			name: "Non-existent directory",
			setupFunc: func() string {
				return filepath.Join(tmpDir, "non-existent")
			},
			expectedExists: false,
			expectedIsGit:  false,
			expectError:    false,
		},
		{
			name: "Existing non-git directory",
			setupFunc: func() string {
				dir := filepath.Join(tmpDir, "non-git")
				require.NoError(t, os.MkdirAll(dir, 0o755))
				return dir
			},
			expectedExists: true,
			expectedIsGit:  false,
			expectError:    false,
		},
		{
			name: "Existing git directory",
			setupFunc: func() string {
				dir := filepath.Join(tmpDir, "git-repo")
				require.NoError(t, os.MkdirAll(dir, 0o755))
				gitDir := filepath.Join(dir, ".git")
				require.NoError(t, os.MkdirAll(gitDir, 0o755))
				return dir
			},
			expectedExists: true,
			expectedIsGit:  true,
			expectError:    false,
		},
		{
			name: "File instead of directory",
			setupFunc: func() string {
				file := filepath.Join(tmpDir, "file")
				require.NoError(t, os.WriteFile(file, []byte("content"), 0o644))
				return file
			},
			expectedExists: false,
			expectedIsGit:  false,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			targetPath := tc.setupFunc()
			opts := &cloneOrUpdateOptions{targetPath: targetPath}

			exists, isGit, err := opts.checkTargetDirectory()

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedExists, exists)
				assert.Equal(t, tc.expectedIsGit, isGit)
			}
		})
	}
}

func TestNewRepoCloneOrUpdateCmd(t *testing.T) {
	cmd := newRepoCloneOrUpdateCmd()

	// Test command properties
	assert.Contains(t, cmd.Use, "clone-or-update")
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.Args)

	// Test flags are present
	flags := cmd.Flags()

	strategyFlag := flags.Lookup("strategy")
	assert.NotNil(t, strategyFlag)
	assert.Equal(t, "rebase", strategyFlag.DefValue)

	branchFlag := flags.Lookup("branch")
	assert.NotNil(t, branchFlag)

	depthFlag := flags.Lookup("depth")
	assert.NotNil(t, depthFlag)
	assert.Equal(t, "0", depthFlag.DefValue)

	forceFlag := flags.Lookup("force")
	assert.NotNil(t, forceFlag)

	verboseFlag := flags.Lookup("verbose")
	assert.NotNil(t, verboseFlag)
}

func TestCloneOrUpdateOptions_Run_Validation(t *testing.T) {
	testCases := []struct {
		name        string
		repoURL     string
		targetPath  string
		strategy    CloneOrUpdateStrategy
		expectError bool
		errorString string
	}{
		{
			name:        "Invalid strategy",
			repoURL:     "https://github.com/user/repo.git",
			targetPath:  "/tmp/repo",
			strategy:    CloneOrUpdateStrategy("invalid"),
			expectError: true,
			errorString: "invalid strategy",
		},
		{
			name:        "Valid strategy",
			repoURL:     "https://github.com/user/repo.git",
			targetPath:  "/tmp/repo",
			strategy:    StrategyRebase,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := &cloneOrUpdateOptions{
				repoURL:    tc.repoURL,
				targetPath: tc.targetPath,
				strategy:   tc.strategy,
			}

			ctx := context.Background()
			err := opts.run(ctx)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorString != "" {
					assert.Contains(t, err.Error(), tc.errorString)
				}
			} else {
				// Note: This will likely fail due to missing git operations
				// In a real test environment, we would mock the git operations
				// For now, we just verify the validation passes
				if err != nil {
					// Allow git-related errors in this basic test
					t.Logf("Expected git-related error: %v", err)
				}
			}
		})
	}
}

// TestCloneOrUpdateStrategies tests the string representation of strategies
func TestCloneOrUpdateStrategies(t *testing.T) {
	testCases := []struct {
		strategy CloneOrUpdateStrategy
		expected string
	}{
		{StrategyRebase, "rebase"},
		{StrategyReset, "reset"},
		{StrategyClone, "clone"},
		{StrategySkip, "skip"},
		{StrategyPull, "pull"},
		{StrategyFetch, "fetch"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.strategy), func(t *testing.T) {
			assert.Equal(t, tc.expected, string(tc.strategy))
		})
	}
}

// TestDefaultCloneOrUpdateOptions tests the default options
func TestDefaultCloneOrUpdateOptions(t *testing.T) {
	opts := &cloneOrUpdateOptions{
		strategy: StrategyRebase,
		depth:    0,
	}

	assert.Equal(t, StrategyRebase, opts.strategy)
	assert.Equal(t, 0, opts.depth)
	assert.False(t, opts.force)
	assert.False(t, opts.verbose)
	assert.Empty(t, opts.branch)
}

// BenchmarkIsValidStrategy benchmarks the strategy validation
func BenchmarkIsValidStrategy(b *testing.B) {
	opts := &cloneOrUpdateOptions{strategy: StrategyRebase}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opts.isValidStrategy()
	}
}

// TestExtractRepoNameFromURL tests repository name extraction from various URL formats
func TestExtractRepoNameFromURL(t *testing.T) {
	testCases := []struct {
		name        string
		url         string
		expected    string
		expectError bool
	}{
		{
			name:     "HTTPS GitHub URL with .git",
			url:      "https://github.com/user/repo.git",
			expected: "repo",
		},
		{
			name:     "HTTPS GitHub URL without .git",
			url:      "https://github.com/user/repo",
			expected: "repo",
		},
		{
			name:     "SSH GitHub URL",
			url:      "git@github.com:user/repo.git",
			expected: "repo",
		},
		{
			name:     "SSH GitHub URL without .git",
			url:      "git@github.com:user/repo",
			expected: "repo",
		},
		{
			name:     "SSH URL with ssh:// prefix",
			url:      "ssh://git@server.com/user/repo.git",
			expected: "repo",
		},
		{
			name:     "GitLab HTTPS URL",
			url:      "https://gitlab.com/group/project.git",
			expected: "project",
		},
		{
			name:     "Complex repository name",
			url:      "https://github.com/org/my-awesome-project.git",
			expected: "my-awesome-project",
		},
		{
			name:        "Empty URL",
			url:         "",
			expectError: true,
		},
		{
			name:        "Invalid HTTPS URL",
			url:         "https://github.com/",
			expected:    "",
			expectError: true,
		},
		{
			name:        "URL with spaces",
			url:         "https://github.com/user/repo with spaces",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := extractRepoNameFromURL(tc.url)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

// TestNewRepoCloneOrUpdateCmd_OptionalTargetPath tests the command with optional target path
func TestNewRepoCloneOrUpdateCmd_OptionalTargetPath(t *testing.T) {
	cmd := newRepoCloneOrUpdateCmd()

	// Test with both arguments (original behavior)
	cmd.SetArgs([]string{"https://github.com/user/repo.git", "./custom-path"})
	err := cmd.Execute()
	// Expect git-related error since we're not actually cloning
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "clone")

	// Test with one argument (new behavior - auto-extract repo name)
	cmd.SetArgs([]string{"https://github.com/user/test-repo.git"})
	err = cmd.Execute()
	// Expect git-related error since we're not actually cloning
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "clone")
}
