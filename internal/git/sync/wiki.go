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

	"github.com/gizzahub/gzh-manager-go/pkg/git/provider"
)

// WikiSyncer handles synchronization of repository wikis.
type WikiSyncer struct {
	source      provider.GitProvider
	destination provider.GitProvider
}

// Sync synchronizes wiki content between source and destination repositories.
func (w *WikiSyncer) Sync(ctx context.Context, srcRepo, dstRepo provider.Repository) error {
	// Note: Wiki availability is checked during cloning
	// Some providers don't expose HasWiki field in Repository struct

	fmt.Printf("    📖 Synchronizing wiki for %s...\n", srcRepo.FullName)

	// Get wiki URLs
	srcWikiURL := w.getWikiURL(srcRepo)
	dstWikiURL := w.getWikiURL(dstRepo)

	// Create temporary directory for wiki operations
	tempDir, err := os.MkdirTemp("", "gzh-wiki-sync-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Clone source wiki
	if err := w.cloneSourceWiki(ctx, srcWikiURL, tempDir); err != nil {
		if w.isWikiNotFound(err) {
			fmt.Printf("    📖 Source wiki is empty or doesn't exist\n")
			return nil
		}
		return fmt.Errorf("failed to clone source wiki: %w", err)
	}

	// Add destination remote and push
	if err := w.pushToDestinationWiki(ctx, dstWikiURL, tempDir); err != nil {
		return fmt.Errorf("failed to push to destination wiki: %w", err)
	}

	// Get sync statistics
	stats, err := w.getWikiStats(ctx, tempDir)
	if err != nil {
		fmt.Printf("    📖 Wiki synchronized (unable to get stats: %v)\n", err)
	} else {
		fmt.Printf("    ✅ Wiki synchronized: %d pages, %s\n", stats.PageCount, stats.Size)
	}

	return nil
}

// getWikiURL constructs the wiki repository URL for a given repository.
func (w *WikiSyncer) getWikiURL(repo provider.Repository) string {
	// Most Git platforms follow the pattern: repo.wiki.git
	// GitHub: https://github.com/owner/repo.wiki.git
	// GitLab: https://gitlab.com/owner/repo.wiki.git
	// Gitea: https://gitea.com/owner/repo.wiki.git

	cloneURL := repo.CloneURL
	if strings.HasSuffix(cloneURL, ".git") {
		// Remove .git and add .wiki.git
		baseURL := strings.TrimSuffix(cloneURL, ".git")
		return baseURL + ".wiki.git"
	}

	// If no .git suffix, just add .wiki.git
	return cloneURL + ".wiki.git"
}

// cloneSourceWiki clones the source wiki repository.
func (w *WikiSyncer) cloneSourceWiki(ctx context.Context, wikiURL, tempDir string) error {
	wikiDir := filepath.Join(tempDir, "wiki")

	// Clone the wiki repository
	cmd := exec.CommandContext(ctx, "git", "clone", wikiURL, wikiDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if this is a "repository not found" error
		outputStr := string(output)
		if w.isWikiNotFound(fmt.Errorf("%s", outputStr)) {
			return fmt.Errorf("wiki not found")
		}
		return fmt.Errorf("git clone failed: %w\nOutput: %s", err, output)
	}

	return nil
}

// pushToDestinationWiki pushes wiki content to the destination.
func (w *WikiSyncer) pushToDestinationWiki(ctx context.Context, dstWikiURL, tempDir string) error {
	wikiDir := filepath.Join(tempDir, "wiki")

	// Add destination remote
	cmd := exec.CommandContext(ctx, "git", "remote", "add", "destination", dstWikiURL)
	cmd.Dir = wikiDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add destination remote: %w\nOutput: %s", err, output)
	}

	// Push all branches (usually just main/master)
	cmd = exec.CommandContext(ctx, "git", "push", "destination", "--all", "--force")
	cmd.Dir = wikiDir
	if output, err := cmd.CombinedOutput(); err != nil {
		// Check if this is because destination wiki doesn't exist yet
		outputStr := string(output)
		if strings.Contains(outputStr, "repository not found") ||
			strings.Contains(outputStr, "does not exist") {
			// Try to create the destination wiki by making an initial commit
			if err := w.initializeDestinationWiki(ctx, dstWikiURL, wikiDir); err != nil {
				return fmt.Errorf("failed to initialize destination wiki: %w", err)
			}
			// Retry the push
			cmd = exec.CommandContext(ctx, "git", "push", "destination", "--all", "--force")
			cmd.Dir = wikiDir
			if retryOutput, retryErr := cmd.CombinedOutput(); retryErr != nil {
				return fmt.Errorf("git push retry failed: %w\nOutput: %s", retryErr, retryOutput)
			}
		} else {
			return fmt.Errorf("git push failed: %w\nOutput: %s", err, output)
		}
	}

	return nil
}

// initializeDestinationWiki initializes an empty wiki at the destination.
func (w *WikiSyncer) initializeDestinationWiki(ctx context.Context, dstWikiURL, wikiDir string) error {
	// Create a temporary directory for initializing the destination wiki
	initDir, err := os.MkdirTemp("", "gzh-wiki-init-*")
	if err != nil {
		return fmt.Errorf("failed to create init directory: %w", err)
	}
	defer os.RemoveAll(initDir)

	// Initialize a new git repository
	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Dir = initDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init failed: %w\nOutput: %s", err, output)
	}

	// Create an initial wiki page
	homeFile := filepath.Join(initDir, "Home.md")
	homeContent := "# Wiki\n\nThis wiki was synchronized from another repository.\n"
	if err := os.WriteFile(homeFile, []byte(homeContent), 0o644); err != nil {
		return fmt.Errorf("failed to create Home.md: %w", err)
	}

	// Add and commit the initial page
	cmd = exec.CommandContext(ctx, "git", "add", "Home.md")
	cmd.Dir = initDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %w\nOutput: %s", err, output)
	}

	cmd = exec.CommandContext(ctx, "git", "commit", "-m", "Initialize wiki")
	cmd.Dir = initDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed: %w\nOutput: %s", err, output)
	}

	// Add destination remote and push
	cmd = exec.CommandContext(ctx, "git", "remote", "add", "origin", dstWikiURL)
	cmd.Dir = initDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add origin remote: %w\nOutput: %s", err, output)
	}

	cmd = exec.CommandContext(ctx, "git", "push", "origin", "main")
	cmd.Dir = initDir
	if output, err := cmd.CombinedOutput(); err != nil {
		// Try with master if main fails
		cmd = exec.CommandContext(ctx, "git", "push", "origin", "master")
		cmd.Dir = initDir
		if masterOutput, masterErr := cmd.CombinedOutput(); masterErr != nil {
			return fmt.Errorf("git push failed for both main and master: %w\nMain output: %s\nMaster output: %s",
				err, output, masterOutput)
		}
	}

	return nil
}

// isWikiNotFound checks if the error indicates that the wiki doesn't exist.
func (w *WikiSyncer) isWikiNotFound(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "repository not found") ||
		strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "does not exist") ||
		strings.Contains(errStr, "wiki not found")
}

// getWikiStats returns statistics about the wiki content.
func (w *WikiSyncer) getWikiStats(ctx context.Context, tempDir string) (*WikiSyncStats, error) {
	wikiDir := filepath.Join(tempDir, "wiki")
	stats := &WikiSyncStats{}

	// Count wiki pages (*.md files)
	err := filepath.Walk(wikiDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			stats.PageCount++
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk wiki directory: %w", err)
	}

	// Get repository size (approximate)
	cmd := exec.CommandContext(ctx, "du", "-sh", wikiDir)
	if output, err := cmd.Output(); err == nil {
		stats.Size = strings.TrimSpace(strings.Fields(string(output))[0])
	} else {
		stats.Size = "unknown"
	}

	// Get page list
	err = filepath.Walk(wikiDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			// Get relative path from wiki directory
			relPath, err := filepath.Rel(wikiDir, path)
			if err == nil {
				stats.Pages = append(stats.Pages, relPath)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get page list: %w", err)
	}

	return stats, nil
}

// ValidateWikiAccess checks if wiki repositories are accessible.
func (w *WikiSyncer) ValidateWikiAccess(ctx context.Context, srcRepo, dstRepo provider.Repository) error {
	// Check source wiki URL
	srcWikiURL := w.getWikiURL(srcRepo)
	if srcWikiURL == "" {
		return fmt.Errorf("source wiki URL is empty")
	}

	// Check destination wiki URL
	dstWikiURL := w.getWikiURL(dstRepo)
	if dstWikiURL == "" {
		return fmt.Errorf("destination wiki URL is empty")
	}

	// TODO: Add actual connectivity checks if needed
	// This could involve testing git ls-remote on the wiki URLs

	return nil
}

// WikiSyncStats represents statistics for wiki synchronization.
type WikiSyncStats struct {
	PageCount int      `json:"page_count"`
	Size      string   `json:"size"`
	Pages     []string `json:"pages"`
}
