package testlib

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// LargeRepoCreator creates repositories with large files to test
// performance and handling of repositories with significant disk usage.
type LargeRepoCreator struct {
	timeout time.Duration
}

// LargeRepoOptions defines options for creating large repositories
type LargeRepoOptions struct {
	RepoPath      string
	LargeFileSize int64 // Size in MB
	FileCount     int   // Number of large files
	WithHistory   bool  // Create commit history with large files
	ChunkSize     int   // Write chunk size in KB (for memory efficiency)
}

// NewLargeRepoCreator creates a new LargeRepoCreator instance
func NewLargeRepoCreator() *LargeRepoCreator {
	return &LargeRepoCreator{
		timeout: 300 * time.Second, // 5 minutes for large operations
	}
}

// CreateLargeRepo creates a repository with large files
func (c *LargeRepoCreator) CreateLargeRepo(ctx context.Context, opts LargeRepoOptions) error {
	if opts.RepoPath == "" {
		return fmt.Errorf("repository path is required")
	}
	if opts.LargeFileSize <= 0 {
		opts.LargeFileSize = 100 // Default 100MB
	}
	if opts.FileCount <= 0 {
		opts.FileCount = 1 // Default 1 file
	}
	if opts.ChunkSize <= 0 {
		opts.ChunkSize = 1024 // Default 1MB chunks
	}

	// Create base repository
	if err := c.createBaseRepo(ctx, opts.RepoPath); err != nil {
		return fmt.Errorf("failed to create base repository: %w", err)
	}

	// Create large files
	for i := 1; i <= opts.FileCount; i++ {
		filename := fmt.Sprintf("large-file-%d.bin", i)
		if err := c.createLargeFile(ctx, opts.RepoPath, filename, opts.LargeFileSize, opts.ChunkSize); err != nil {
			return fmt.Errorf("failed to create large file %s: %w", filename, err)
		}

		// Add to Git (this might take time)
		if err := c.runGitCommand(ctx, opts.RepoPath, "add", filename); err != nil {
			return fmt.Errorf("failed to add large file %s: %w", filename, err)
		}

		commitMsg := fmt.Sprintf("Add large file %s (%.1fMB)", filename, float64(opts.LargeFileSize))
		if err := c.runGitCommand(ctx, opts.RepoPath, "commit", "-m", commitMsg); err != nil {
			return fmt.Errorf("failed to commit large file %s: %w", filename, err)
		}
	}

	// Create history with large files if requested
	if opts.WithHistory {
		if err := c.createLargeFileHistory(ctx, opts.RepoPath, opts.ChunkSize); err != nil {
			return fmt.Errorf("failed to create large file history: %w", err)
		}
	}

	return nil
}

// CreateLargeRepoSimple creates a simple large repository with default options
func (c *LargeRepoCreator) CreateLargeRepoSimple(ctx context.Context, repoPath string, sizeMB int64) error {
	opts := LargeRepoOptions{
		RepoPath:      repoPath,
		LargeFileSize: sizeMB,
		FileCount:     1,
		WithHistory:   false,
		ChunkSize:     1024, // 1MB chunks
	}

	return c.CreateLargeRepo(ctx, opts)
}

// CreateLargeRepoWithHistory creates a repository with large files and commit history
func (c *LargeRepoCreator) CreateLargeRepoWithHistory(ctx context.Context, repoPath string) error {
	opts := LargeRepoOptions{
		RepoPath:      repoPath,
		LargeFileSize: 50, // 50MB files
		FileCount:     2,
		WithHistory:   true,
		ChunkSize:     512, // 512KB chunks for better memory usage
	}

	return c.CreateLargeRepo(ctx, opts)
}

// CreateRepoWithBinaryFiles creates a repository with various binary file types
func (c *LargeRepoCreator) CreateRepoWithBinaryFiles(ctx context.Context, repoPath string) error {
	if err := c.createBaseRepo(ctx, repoPath); err != nil {
		return fmt.Errorf("failed to create base repository: %w", err)
	}

	// Create different types of binary files
	binaryFiles := map[string]int64{
		"image.jpg":    5,  // 5MB
		"video.mp4":    20, // 20MB
		"archive.zip":  15, // 15MB
		"database.db":  10, // 10MB
		"compiled.exe": 8,  // 8MB
	}

	for filename, sizeMB := range binaryFiles {
		if err := c.createLargeFile(ctx, repoPath, filename, sizeMB, 512); err != nil {
			return fmt.Errorf("failed to create binary file %s: %w", filename, err)
		}

		if err := c.runGitCommand(ctx, repoPath, "add", filename); err != nil {
			return fmt.Errorf("failed to add binary file %s: %w", filename, err)
		}

		commitMsg := fmt.Sprintf("Add %s (%.1fMB)", filename, float64(sizeMB))
		if err := c.runGitCommand(ctx, repoPath, "commit", "-m", commitMsg); err != nil {
			return fmt.Errorf("failed to commit binary file %s: %w", filename, err)
		}
	}

	return nil
}

// GetRepoSize calculates the approximate repository size
func (c *LargeRepoCreator) GetRepoSize(repoPath string) (int64, error) {
	var totalSize int64

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to calculate repository size: %w", err)
	}

	return totalSize, nil
}

// createBaseRepo creates a basic repository structure
func (c *LargeRepoCreator) createBaseRepo(ctx context.Context, repoPath string) error {
	// Create directory
	if err := os.MkdirAll(repoPath, 0755); err != nil {
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
	content := fmt.Sprintf("# %s\n\nLarge test repository for synclone performance testing.\n\n⚠️ This repository contains large files.\n", filepath.Base(repoPath))
	if err := os.WriteFile(readmePath, []byte(content), 0644); err != nil {
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

// createLargeFile creates a large file with specified size using memory-efficient chunked writing
func (c *LargeRepoCreator) createLargeFile(ctx context.Context, repoPath, filename string, sizeMB int64, chunkSizeKB int) error {
	filePath := filepath.Join(repoPath, filename)

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	// Calculate total bytes and chunk size
	totalBytes := sizeMB * 1024 * 1024
	chunkSize := int64(chunkSizeKB) * 1024

	// Create a buffer for random data
	buffer := make([]byte, chunkSize)

	// Write data in chunks to manage memory usage
	var bytesWritten int64
	for bytesWritten < totalBytes {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Calculate how much to write in this chunk
		remainingBytes := totalBytes - bytesWritten
		writeSize := chunkSize
		if remainingBytes < chunkSize {
			writeSize = remainingBytes
			buffer = buffer[:writeSize]
		}

		// Fill buffer with random data
		if _, err := rand.Read(buffer); err != nil {
			return fmt.Errorf("failed to generate random data: %w", err)
		}

		// Write chunk to file
		if _, err := file.Write(buffer); err != nil {
			return fmt.Errorf("failed to write chunk to file: %w", err)
		}

		bytesWritten += writeSize
	}

	return nil
}

// createLargeFileHistory creates a history with modifications to large files
func (c *LargeRepoCreator) createLargeFileHistory(ctx context.Context, repoPath string, chunkSizeKB int) error {
	// Create a file that will be modified over time
	filename := "evolving-large-file.bin"

	// Create initial version (10MB)
	if err := c.createLargeFile(ctx, repoPath, filename, 10, chunkSizeKB); err != nil {
		return fmt.Errorf("failed to create initial large file: %w", err)
	}

	if err := c.runGitCommand(ctx, repoPath, "add", filename); err != nil {
		return fmt.Errorf("failed to add initial large file: %w", err)
	}

	if err := c.runGitCommand(ctx, repoPath, "commit", "-m", "Add initial large file (10MB)"); err != nil {
		return fmt.Errorf("failed to commit initial large file: %w", err)
	}

	// Create versions with increasing size
	sizes := []int64{20, 35, 50} // MB
	for i, size := range sizes {
		if err := c.createLargeFile(ctx, repoPath, filename, size, chunkSizeKB); err != nil {
			return fmt.Errorf("failed to create large file version %d: %w", i+2, err)
		}

		if err := c.runGitCommand(ctx, repoPath, "add", filename); err != nil {
			return fmt.Errorf("failed to add large file version %d: %w", i+2, err)
		}

		commitMsg := fmt.Sprintf("Update large file to %dMB", size)
		if err := c.runGitCommand(ctx, repoPath, "commit", "-m", commitMsg); err != nil {
			return fmt.Errorf("failed to commit large file version %d: %w", i+2, err)
		}
	}

	return nil
}

// configureGitUser sets up git user configuration for testing
func (c *LargeRepoCreator) configureGitUser(ctx context.Context, repoPath string) error {
	if err := c.runGitCommand(ctx, repoPath, "config", "user.name", "Test User"); err != nil {
		return fmt.Errorf("failed to set git user name: %w", err)
	}
	if err := c.runGitCommand(ctx, repoPath, "config", "user.email", "test@example.com"); err != nil {
		return fmt.Errorf("failed to set git user email: %w", err)
	}
	return nil
}

// runGitCommand executes a git command with timeout
func (c *LargeRepoCreator) runGitCommand(ctx context.Context, dir string, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %v failed: %w, output: %s", args, err, string(output))
	}

	return nil
}
