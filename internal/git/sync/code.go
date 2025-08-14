// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package sync

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Gizzahub/gzh-cli/pkg/git/provider"
)

// CodeSyncer handles repository code synchronization.
type CodeSyncer struct {
	source      provider.Repository
	destination *provider.Repository
	options     Options
}

// Sync synchronizes repository code between source and destination.
func (c *CodeSyncer) Sync(ctx context.Context) error {
	if c.destination == nil {
		return fmt.Errorf("destination repository is required for code sync")
	}

	// Create temporary directory for git operations
	tempDir, err := os.MkdirTemp("", "gzh-sync-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Clone source repository
	if err := c.cloneSource(ctx, tempDir); err != nil {
		return fmt.Errorf("failed to clone source: %w", err)
	}

	// Add destination remote
	if err := c.addDestinationRemote(ctx, tempDir); err != nil {
		return fmt.Errorf("failed to add destination remote: %w", err)
	}

	// Push to destination
	if err := c.pushToDestination(ctx, tempDir); err != nil {
		return fmt.Errorf("failed to push to destination: %w", err)
	}

	return nil
}

// cloneSource clones the source repository using mirror mode.
func (c *CodeSyncer) cloneSource(ctx context.Context, tempDir string) error {
	repoDir := filepath.Join(tempDir, "repo")

	// Use mirror clone to get all branches and tags
	args := []string{"clone", "--mirror", c.source.CloneURL, repoDir}
	cmd := exec.CommandContext(ctx, "git", args...)

	if c.options.Verbose {
		fmt.Printf("Cloning source: git %s\n", joinArgs(args))
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %w\nOutput: %s", err, output)
	}

	return nil
}

// addDestinationRemote adds the destination as a remote.
func (c *CodeSyncer) addDestinationRemote(ctx context.Context, tempDir string) error {
	repoDir := filepath.Join(tempDir, "repo")

	// Add destination remote
	cmd := exec.CommandContext(ctx, "git", "remote", "add", "destination", c.destination.CloneURL)
	cmd.Dir = repoDir

	if c.options.Verbose {
		fmt.Printf("Adding destination remote: %s\n", c.destination.CloneURL)
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add destination remote: %w\nOutput: %s", err, output)
	}

	return nil
}

// pushToDestination pushes all branches and tags to the destination.
func (c *CodeSyncer) pushToDestination(ctx context.Context, tempDir string) error {
	repoDir := filepath.Join(tempDir, "repo")

	// Push all branches
	if err := c.pushBranches(ctx, repoDir); err != nil {
		return err
	}

	// Push all tags
	if err := c.pushTags(ctx, repoDir); err != nil {
		return err
	}

	return nil
}

// pushBranches pushes all branches to the destination.
func (c *CodeSyncer) pushBranches(ctx context.Context, repoDir string) error {
	args := []string{"push", "destination", "--all"}
	if c.options.Force {
		args = append(args, "--force")
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = repoDir

	if c.options.Verbose {
		fmt.Printf("Pushing branches: git %s\n", joinArgs(args))
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		// Check if this is because the destination is empty (common case)
		if !c.options.Force {
			// Try with force for initial push to empty repository
			forceArgs := append(args, "--force")
			forceCmd := exec.CommandContext(ctx, "git", forceArgs...)
			forceCmd.Dir = repoDir

			if c.options.Verbose {
				fmt.Printf("Retrying with force: git %s\n", joinArgs(forceArgs))
			}

			if forceOutput, forceErr := forceCmd.CombinedOutput(); forceErr != nil {
				return fmt.Errorf("git push branches failed: %w\nOutput: %s\nForce output: %s", err, output, forceOutput)
			}
			return nil
		}
		return fmt.Errorf("git push branches failed: %w\nOutput: %s", err, output)
	}

	return nil
}

// pushTags pushes all tags to the destination.
func (c *CodeSyncer) pushTags(ctx context.Context, repoDir string) error {
	args := []string{"push", "destination", "--tags"}
	if c.options.Force {
		args = append(args, "--force")
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = repoDir

	if c.options.Verbose {
		fmt.Printf("Pushing tags: git %s\n", joinArgs(args))
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		// Tags might not exist, which is OK
		if contains(string(output), "no refs in common") || contains(string(output), "Everything up-to-date") {
			if c.options.Verbose {
				fmt.Println("No tags to push")
			}
			return nil
		}
		return fmt.Errorf("git push tags failed: %w\nOutput: %s", err, output)
	}

	return nil
}

// SyncNewRepository creates and synchronizes a new repository.
func (c *CodeSyncer) SyncNewRepository(ctx context.Context, destProvider provider.GitProvider, destTarget *SyncTarget) (*provider.Repository, error) {
	// Create repository in destination
	createReq := provider.CreateRepoRequest{
		Name:        c.source.Name,
		Description: c.source.Description,
		Private:     c.source.Private,
		HasIssues:   true, // Default to true, can be configured later
		HasWiki:     true, // Default to true, can be configured later
		Topics:      c.source.Topics,
	}

	if c.options.Verbose {
		fmt.Printf("Creating repository: %s/%s\n", destTarget.Org, c.source.Name)
	}

	newRepo, err := destProvider.CreateRepository(ctx, createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	// Update destination for code sync
	c.destination = newRepo

	// Sync code to the new repository
	if err := c.Sync(ctx); err != nil {
		return nil, fmt.Errorf("failed to sync code to new repository: %w", err)
	}

	return newRepo, nil
}

// ValidateRepositories checks if source and destination repositories are accessible.
func (c *CodeSyncer) ValidateRepositories(ctx context.Context) error {
	// Check source repository
	if c.source.CloneURL == "" {
		return fmt.Errorf("source repository clone URL is empty")
	}

	// Check destination repository if provided
	if c.destination != nil && c.destination.CloneURL == "" {
		return fmt.Errorf("destination repository clone URL is empty")
	}

	return nil
}

// GetSyncStats returns statistics about the code sync operation.
func (c *CodeSyncer) GetSyncStats(ctx context.Context, tempDir string) (*CodeSyncStats, error) {
	repoDir := filepath.Join(tempDir, "repo")

	stats := &CodeSyncStats{}

	// Count branches
	cmd := exec.CommandContext(ctx, "git", "branch", "-r", "--list")
	cmd.Dir = repoDir
	if output, err := cmd.Output(); err == nil {
		stats.BranchCount = countLines(string(output))
	}

	// Count tags
	cmd = exec.CommandContext(ctx, "git", "tag", "--list")
	cmd.Dir = repoDir
	if output, err := cmd.Output(); err == nil {
		stats.TagCount = countLines(string(output))
	}

	// Get repository size (approximate)
	cmd = exec.CommandContext(ctx, "du", "-sh", repoDir)
	if output, err := cmd.Output(); err == nil {
		stats.Size = strings.TrimSpace(strings.Fields(string(output))[0])
	}

	return stats, nil
}

// CodeSyncStats represents statistics for code synchronization.
type CodeSyncStats struct {
	BranchCount int    `json:"branch_count"`
	TagCount    int    `json:"tag_count"`
	Size        string `json:"size"`
}

// Helper functions

// joinArgs joins command arguments for display.
func joinArgs(args []string) string {
	return strings.Join(args, " ")
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// countLines counts non-empty lines in a string.
func countLines(s string) int {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}
