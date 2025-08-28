package testlib

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GitLFSSimulator simulates Git LFS (Large File Storage) environments
// for testing synclone behavior with large files and LFS tracking.
type GitLFSSimulator struct {
	timeout time.Duration
}

// LFSConfig represents Git LFS configuration options
type LFSConfig struct {
	TrackPatterns []string          // File patterns to track with LFS
	LargeFiles    []LFSFile         // Large files to create and track
	Attributes    map[string]string // Custom .gitattributes entries
}

// LFSFile represents a large file managed by LFS
type LFSFile struct {
	Path    string // File path relative to repo root
	SizeMB  int64  // Size in megabytes
	Content string // Content type: "random", "text", "binary"
	Tracked bool   // Whether file should be tracked by LFS
}

// NewGitLFSSimulator creates a new GitLFSSimulator instance
func NewGitLFSSimulator() *GitLFSSimulator {
	return &GitLFSSimulator{
		timeout: 120 * time.Second, // Longer timeout for LFS operations
	}
}

// CreateLFSRepository creates a repository with Git LFS configured
func (lfs *GitLFSSimulator) CreateLFSRepository(ctx context.Context, repoPath string, config LFSConfig) error {
	// Initialize repository
	if err := lfs.initializeRepo(ctx, repoPath); err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}

	// Install and initialize Git LFS (if available)
	if err := lfs.initializeLFS(ctx, repoPath); err != nil {
		return fmt.Errorf("failed to initialize LFS: %w", err)
	}

	// Create .gitattributes file
	if err := lfs.createGitAttributes(ctx, repoPath, config); err != nil {
		return fmt.Errorf("failed to create .gitattributes: %w", err)
	}

	// Create and track large files
	for _, file := range config.LargeFiles {
		if err := lfs.createLFSFile(ctx, repoPath, file); err != nil {
			return fmt.Errorf("failed to create LFS file %s: %w", file.Path, err)
		}
	}

	return nil
}

// CreateSimpleLFSRepo creates a simple LFS repository with common patterns
func (lfs *GitLFSSimulator) CreateSimpleLFSRepo(ctx context.Context, repoPath string) error {
	config := LFSConfig{
		TrackPatterns: []string{
			"*.bin",
			"*.zip",
			"*.tar.gz",
			"assets/**",
		},
		LargeFiles: []LFSFile{
			{
				Path:    "assets/large-image.png",
				SizeMB:  25,
				Content: "binary",
				Tracked: true,
			},
			{
				Path:    "data/dataset.bin",
				SizeMB:  50,
				Content: "random",
				Tracked: true,
			},
			{
				Path:    "regular-file.txt",
				SizeMB:  1, // Small file, won't be tracked by LFS
				Content: "text",
				Tracked: false,
			},
		},
	}

	return lfs.CreateLFSRepository(ctx, repoPath, config)
}

// IsLFSAvailable checks if Git LFS is available on the system
func (lfs *GitLFSSimulator) IsLFSAvailable(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "git", "lfs", "version")
	return cmd.Run() == nil
}

// GetLFSStatus returns the LFS status of the repository
func (lfs *GitLFSSimulator) GetLFSStatus(ctx context.Context, repoPath string) (LFSStatus, error) {
	if !lfs.IsLFSAvailable(ctx) {
		return LFSStatus{Available: false}, nil
	}

	cmd := exec.CommandContext(ctx, "git", "lfs", "ls-files")
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return LFSStatus{Available: true, Initialized: false}, nil
	}

	status := LFSStatus{
		Available:    true,
		Initialized:  true,
		TrackedFiles: lfs.parseLFSFiles(string(output)),
	}

	return status, nil
}

// LFSStatus represents the Git LFS status
type LFSStatus struct {
	Available    bool     // Whether Git LFS is available
	Initialized  bool     // Whether LFS is initialized in the repo
	TrackedFiles []string // Files currently tracked by LFS
}

// initializeRepo creates a basic Git repository
func (lfs *GitLFSSimulator) initializeRepo(ctx context.Context, repoPath string) error {
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := lfs.runGitCommand(ctx, repoPath, "init"); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Configure git user
	if err := lfs.runGitCommand(ctx, repoPath, "config", "user.name", "Test User"); err != nil {
		return fmt.Errorf("failed to set git user name: %w", err)
	}
	if err := lfs.runGitCommand(ctx, repoPath, "config", "user.email", "test@example.com"); err != nil {
		return fmt.Errorf("failed to set git user email: %w", err)
	}

	// Create initial commit
	readmePath := filepath.Join(repoPath, "README.md")
	content := fmt.Sprintf("# %s\n\nTest repository with Git LFS support.\n", filepath.Base(repoPath))
	if err := os.WriteFile(readmePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}

	if err := lfs.runGitCommand(ctx, repoPath, "add", "README.md"); err != nil {
		return fmt.Errorf("failed to add README.md: %w", err)
	}

	if err := lfs.runGitCommand(ctx, repoPath, "commit", "-m", "Initial commit"); err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	return nil
}

// initializeLFS initializes Git LFS in the repository
func (lfs *GitLFSSimulator) initializeLFS(ctx context.Context, repoPath string) error {
	if !lfs.IsLFSAvailable(ctx) {
		fmt.Printf("Warning: Git LFS not available, simulating without actual LFS\n")
		return nil
	}

	if err := lfs.runGitCommand(ctx, repoPath, "lfs", "install", "--local"); err != nil {
		return fmt.Errorf("failed to install LFS: %w", err)
	}

	return nil
}

// createGitAttributes creates .gitattributes file with LFS tracking patterns
func (lfs *GitLFSSimulator) createGitAttributes(ctx context.Context, repoPath string, config LFSConfig) error {
	attributesPath := filepath.Join(repoPath, ".gitattributes")

	var content strings.Builder
	content.WriteString("# Git LFS tracking patterns\n")

	for _, pattern := range config.TrackPatterns {
		content.WriteString(fmt.Sprintf("%s filter=lfs diff=lfs merge=lfs -text\n", pattern))
	}

	// Add custom attributes
	for pattern, attr := range config.Attributes {
		content.WriteString(fmt.Sprintf("%s %s\n", pattern, attr))
	}

	if err := os.WriteFile(attributesPath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to create .gitattributes: %w", err)
	}

	// Track the attributes file with LFS if available
	if lfs.IsLFSAvailable(ctx) {
		for _, pattern := range config.TrackPatterns {
			if err := lfs.runGitCommand(ctx, repoPath, "lfs", "track", pattern); err != nil {
				// Continue if tracking fails, might already be tracked
				continue
			}
		}
	}

	// Commit .gitattributes
	if err := lfs.runGitCommand(ctx, repoPath, "add", ".gitattributes"); err != nil {
		return fmt.Errorf("failed to add .gitattributes: %w", err)
	}

	if err := lfs.runGitCommand(ctx, repoPath, "commit", "-m", "Add LFS tracking configuration"); err != nil {
		return fmt.Errorf("failed to commit .gitattributes: %w", err)
	}

	return nil
}

// createLFSFile creates a large file that may be tracked by LFS
func (lfs *GitLFSSimulator) createLFSFile(ctx context.Context, repoPath string, file LFSFile) error {
	filePath := filepath.Join(repoPath, file.Path)

	// Create directory if needed
	if dir := filepath.Dir(filePath); dir != repoPath {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Create file content based on type
	var content []byte
	switch file.Content {
	case "text":
		content = lfs.generateTextContent(file.SizeMB)
	case "binary", "random":
		content = lfs.generateRandomContent(file.SizeMB)
	default:
		content = lfs.generateTextContent(file.SizeMB)
	}

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Add to Git
	if err := lfs.runGitCommand(ctx, repoPath, "add", file.Path); err != nil {
		return fmt.Errorf("failed to add file to git: %w", err)
	}

	// Commit the file
	commitMsg := fmt.Sprintf("Add %s (%.1fMB)", file.Path, float64(file.SizeMB))
	if file.Tracked {
		commitMsg += " [LFS]"
	}

	if err := lfs.runGitCommand(ctx, repoPath, "commit", "-m", commitMsg); err != nil {
		return fmt.Errorf("failed to commit file: %w", err)
	}

	return nil
}

// generateTextContent generates text content of specified size
func (lfs *GitLFSSimulator) generateTextContent(sizeMB int64) []byte {
	sizeBytes := sizeMB * 1024 * 1024
	content := make([]byte, 0, sizeBytes)

	line := "This is a line of text content for LFS testing purposes.\n"
	lineBytes := []byte(line)

	for int64(len(content)) < sizeBytes {
		remaining := sizeBytes - int64(len(content))
		if remaining < int64(len(lineBytes)) {
			content = append(content, lineBytes[:remaining]...)
		} else {
			content = append(content, lineBytes...)
		}
	}

	return content
}

// generateRandomContent generates random binary content of specified size
func (lfs *GitLFSSimulator) generateRandomContent(sizeMB int64) []byte {
	sizeBytes := sizeMB * 1024 * 1024
	content := make([]byte, sizeBytes)

	// Simple pseudo-random content generation
	for i := range content {
		content[i] = byte(i % 256)
	}

	return content
}

// parseLFSFiles parses the output of `git lfs ls-files`
func (lfs *GitLFSSimulator) parseLFSFiles(output string) []string {
	var files []string
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		// LFS ls-files format: "hash * filename"
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			files = append(files, parts[2])
		}
	}

	return files
}

// runGitCommand executes a git command with timeout
func (lfs *GitLFSSimulator) runGitCommand(ctx context.Context, dir string, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, lfs.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %v failed: %w, output: %s", args, err, string(output))
	}

	return nil
}
