package testlib

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// StandardRepoCreator creates repositories with standard Git structures
// including multiple commits, branches, and typical development patterns.
type StandardRepoCreator struct {
	timeout time.Duration
}

// StandardRepoOptions defines options for creating standard repositories
type StandardRepoOptions struct {
	RepoPath     string
	Branches     []string
	CommitsCount int
	WithTags     bool
	WithMerges   bool
}

// NewStandardRepoCreator creates a new StandardRepoCreator instance
func NewStandardRepoCreator() *StandardRepoCreator {
	return &StandardRepoCreator{
		timeout: 30 * time.Second,
	}
}

// CreateStandardRepo creates a repository with typical development patterns
func (c *StandardRepoCreator) CreateStandardRepo(ctx context.Context, opts StandardRepoOptions) error {
	if opts.RepoPath == "" {
		return fmt.Errorf("repository path is required")
	}

	// Create base repository
	if err := c.createBaseRepo(ctx, opts.RepoPath); err != nil {
		return fmt.Errorf("failed to create base repository: %w", err)
	}

	// Add multiple commits on main branch
	if err := c.createMultipleCommits(ctx, opts.RepoPath, opts.CommitsCount); err != nil {
		return fmt.Errorf("failed to create commits: %w", err)
	}

	// Create additional branches
	if err := c.createBranches(ctx, opts.RepoPath, opts.Branches); err != nil {
		return fmt.Errorf("failed to create branches: %w", err)
	}

	// Add tags if requested
	if opts.WithTags {
		if err := c.createTags(ctx, opts.RepoPath); err != nil {
			return fmt.Errorf("failed to create tags: %w", err)
		}
	}

	// Create merge commits if requested
	if opts.WithMerges {
		if err := c.createMergeCommits(ctx, opts.RepoPath); err != nil {
			return fmt.Errorf("failed to create merge commits: %w", err)
		}
	}

	return nil
}

// CreateDevelopmentRepo creates a repository simulating typical development workflow
func (c *StandardRepoCreator) CreateDevelopmentRepo(ctx context.Context, repoPath string) error {
	opts := StandardRepoOptions{
		RepoPath:     repoPath,
		Branches:     []string{"develop", "feature/user-auth", "hotfix/critical-bug"},
		CommitsCount: 8,
		WithTags:     true,
		WithMerges:   true,
	}

	return c.CreateStandardRepo(ctx, opts)
}

// CreateProjectRepo creates a repository with project-like structure
func (c *StandardRepoCreator) CreateProjectRepo(ctx context.Context, repoPath string, projectType string) error {
	if err := c.createBaseRepo(ctx, repoPath); err != nil {
		return fmt.Errorf("failed to create base repository: %w", err)
	}

	var files map[string]string
	switch projectType {
	case "go":
		files = c.getGoProjectFiles()
	case "javascript":
		files = c.getJavaScriptProjectFiles()
	case "python":
		files = c.getPythonProjectFiles()
	default:
		files = c.getGenericProjectFiles()
	}

	// Create project structure
	for filename, content := range files {
		filePath := filepath.Join(repoPath, filename)

		// Create directory if needed
		if dir := filepath.Dir(filePath); dir != "." {
			if err := os.MkdirAll(filepath.Join(repoPath, dir), 0o755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}

		if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("failed to create file %s: %w", filename, err)
		}

		if err := c.runGitCommand(ctx, repoPath, "add", filename); err != nil {
			return fmt.Errorf("failed to add file %s: %w", filename, err)
		}
	}

	// Commit project structure
	if err := c.runGitCommand(ctx, repoPath, "commit", "-m", fmt.Sprintf("Initial %s project structure", projectType)); err != nil {
		return fmt.Errorf("failed to commit project structure: %w", err)
	}

	return nil
}

// createBaseRepo creates a basic repository with initial commit
func (c *StandardRepoCreator) createBaseRepo(ctx context.Context, repoPath string) error {
	// Create directory
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		return fmt.Errorf("failed to create repository directory: %w", err)
	}

	// Initialize Git repository
	if err := c.runGitCommand(ctx, repoPath, "init"); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Configure git user for testing
	if err := c.configureGitUser(ctx, repoPath); err != nil {
		return fmt.Errorf("failed to configure git user: %w", err)
	}

	// Add initial README
	readmePath := filepath.Join(repoPath, "README.md")
	content := fmt.Sprintf("# %s\n\nStandard test repository for synclone testing.\n\nCreated with multiple commits and branches.\n", filepath.Base(repoPath))
	if err := os.WriteFile(readmePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}

	if err := c.runGitCommand(ctx, repoPath, "add", "README.md"); err != nil {
		return fmt.Errorf("failed to add README.md: %w", err)
	}

	if err := c.runGitCommand(ctx, repoPath, "commit", "-m", "Initial commit"); err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	return nil
}

// createMultipleCommits adds multiple commits to simulate development history
func (c *StandardRepoCreator) createMultipleCommits(ctx context.Context, repoPath string, count int) error {
	if count <= 0 {
		count = 5 // Default to 5 commits
	}

	for i := 1; i <= count; i++ {
		// Create a new file or modify existing one
		filename := fmt.Sprintf("file%d.txt", i)
		content := fmt.Sprintf("This is file number %d\nCreated for commit %d\nTimestamp: %s\n",
			i, i, time.Now().Format(time.RFC3339))

		filePath := filepath.Join(repoPath, filename)
		if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("failed to create file %s: %w", filename, err)
		}

		if err := c.runGitCommand(ctx, repoPath, "add", filename); err != nil {
			return fmt.Errorf("failed to add file %s: %w", filename, err)
		}

		commitMsg := fmt.Sprintf("Add %s - commit %d", filename, i)
		if err := c.runGitCommand(ctx, repoPath, "commit", "-m", commitMsg); err != nil {
			return fmt.Errorf("failed to commit %s: %w", filename, err)
		}
	}

	return nil
}

// createBranches creates additional branches
func (c *StandardRepoCreator) createBranches(ctx context.Context, repoPath string, branches []string) error {
	for _, branch := range branches {
		if err := c.runGitCommand(ctx, repoPath, "checkout", "-b", branch); err != nil {
			return fmt.Errorf("failed to create branch %s: %w", branch, err)
		}

		// Add a commit to the new branch
		filename := fmt.Sprintf("%s.txt", branch)
		content := fmt.Sprintf("Content for branch: %s\nCreated at: %s\n", branch, time.Now().Format(time.RFC3339))

		filePath := filepath.Join(repoPath, filename)
		if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("failed to create branch file %s: %w", filename, err)
		}

		if err := c.runGitCommand(ctx, repoPath, "add", filename); err != nil {
			return fmt.Errorf("failed to add branch file %s: %w", filename, err)
		}

		commitMsg := fmt.Sprintf("Add %s to branch %s", filename, branch)
		if err := c.runGitCommand(ctx, repoPath, "commit", "-m", commitMsg); err != nil {
			return fmt.Errorf("failed to commit to branch %s: %w", branch, err)
		}

		// Return to main branch
		if err := c.runGitCommand(ctx, repoPath, "checkout", "main"); err != nil {
			// Try master if main doesn't exist
			c.runGitCommand(ctx, repoPath, "checkout", "master")
		}
	}

	return nil
}

// createTags adds version tags to the repository
func (c *StandardRepoCreator) createTags(ctx context.Context, repoPath string) error {
	tags := []string{"v1.0.0", "v1.1.0", "v1.2.0"}

	for _, tag := range tags {
		if err := c.runGitCommand(ctx, repoPath, "tag", "-a", tag, "-m", fmt.Sprintf("Version %s", tag)); err != nil {
			return fmt.Errorf("failed to create tag %s: %w", tag, err)
		}
	}

	return nil
}

// createMergeCommits creates merge commits to simulate feature merging
func (c *StandardRepoCreator) createMergeCommits(ctx context.Context, repoPath string) error {
	// Create and merge a simple feature branch
	if err := c.runGitCommand(ctx, repoPath, "checkout", "-b", "temp-feature"); err != nil {
		return fmt.Errorf("failed to create temp feature branch: %w", err)
	}

	// Add commit to feature branch
	filename := "feature.txt"
	content := "Feature implementation\n"
	filePath := filepath.Join(repoPath, filename)

	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to create feature file: %w", err)
	}

	if err := c.runGitCommand(ctx, repoPath, "add", filename); err != nil {
		return fmt.Errorf("failed to add feature file: %w", err)
	}

	if err := c.runGitCommand(ctx, repoPath, "commit", "-m", "Implement new feature"); err != nil {
		return fmt.Errorf("failed to commit feature: %w", err)
	}

	// Switch back to main and merge
	if err := c.runGitCommand(ctx, repoPath, "checkout", "main"); err != nil {
		c.runGitCommand(ctx, repoPath, "checkout", "master")
	}

	if err := c.runGitCommand(ctx, repoPath, "merge", "--no-ff", "temp-feature", "-m", "Merge feature branch"); err != nil {
		return fmt.Errorf("failed to merge feature branch: %w", err)
	}

	// Clean up temporary branch
	c.runGitCommand(ctx, repoPath, "branch", "-d", "temp-feature")

	return nil
}

// getGoProjectFiles returns files for a Go project structure
func (c *StandardRepoCreator) getGoProjectFiles() map[string]string {
	return map[string]string{
		"go.mod":       "module example.com/test\n\ngo 1.23\n",
		"main.go":      "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}\n",
		"pkg/utils.go": "package pkg\n\n// Utils provides utility functions\ntype Utils struct{}\n",
		"cmd/cli.go":   "package main\n\n// CLI implementation\n",
		"README.md":    "# Go Project\n\nA test Go project.\n",
	}
}

// getJavaScriptProjectFiles returns files for a JavaScript/Node.js project
func (c *StandardRepoCreator) getJavaScriptProjectFiles() map[string]string {
	return map[string]string{
		"package.json": "{\n  \"name\": \"test-project\",\n  \"version\": \"1.0.0\",\n  \"main\": \"index.js\"\n}\n",
		"index.js":     "console.log('Hello, World!');\n",
		"src/app.js":   "// Main application logic\n",
		"test/test.js": "// Test files\n",
		"README.md":    "# JavaScript Project\n\nA test Node.js project.\n",
	}
}

// getPythonProjectFiles returns files for a Python project
func (c *StandardRepoCreator) getPythonProjectFiles() map[string]string {
	return map[string]string{
		"setup.py":           "from setuptools import setup\n\nsetup(name='test-project', version='1.0.0')\n",
		"requirements.txt":   "requests>=2.25.0\n",
		"main.py":            "#!/usr/bin/env python3\n\nif __name__ == '__main__':\n    print('Hello, World!')\n",
		"src/__init__.py":    "",
		"tests/test_main.py": "import unittest\n\nclass TestMain(unittest.TestCase):\n    pass\n",
		"README.md":          "# Python Project\n\nA test Python project.\n",
	}
}

// getGenericProjectFiles returns generic project files
func (c *StandardRepoCreator) getGenericProjectFiles() map[string]string {
	return map[string]string{
		"README.md":     "# Generic Project\n\nA generic test project.\n",
		"LICENSE":       "MIT License\n\nCopyright (c) 2024\n",
		"src/main.txt":  "Main source file\n",
		"docs/guide.md": "# User Guide\n\nProject documentation.\n",
		".gitignore":    "*.log\n.DS_Store\n",
	}
}

// configureGitUser sets up git user configuration for testing
func (c *StandardRepoCreator) configureGitUser(ctx context.Context, repoPath string) error {
	if err := c.runGitCommand(ctx, repoPath, "config", "user.name", "Test User"); err != nil {
		return fmt.Errorf("failed to set git user name: %w", err)
	}
	if err := c.runGitCommand(ctx, repoPath, "config", "user.email", "test@example.com"); err != nil {
		return fmt.Errorf("failed to set git user email: %w", err)
	}
	return nil
}

// runGitCommand executes a git command with timeout
func (c *StandardRepoCreator) runGitCommand(ctx context.Context, dir string, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %v failed: %w, output: %s", args, err, string(output))
	}

	return nil
}
