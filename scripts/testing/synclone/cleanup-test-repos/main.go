package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		log.Fatal(err)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup-test-repos",
		Short: "Clean up test repositories created for synclone testing",
		Long: `Safely remove test repositories created by setup-test-repos.
Includes confirmation prompts and verification to prevent accidental deletion.`,
	}

	cmd.AddCommand(newCleanCmd())
	cmd.AddCommand(newVerifyCmd())

	return cmd
}

func newCleanCmd() *cobra.Command {
	var (
		baseDir string
		force   bool
		dryRun  bool
	)

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Remove test repositories",
		RunE: func(_ *cobra.Command, args []string) error {
			return cleanupRepos(context.Background(), baseDir, force, dryRun)
		},
	}

	cmd.Flags().StringVar(&baseDir, "base-dir", "./test-repos", "Base directory containing test repositories")
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompts")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be deleted without actually deleting")

	return cmd
}

func newVerifyCmd() *cobra.Command {
	var baseDir string

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify test repository structure",
		RunE: func(_ *cobra.Command, args []string) error {
			return verifyRepos(context.Background(), baseDir)
		},
	}

	cmd.Flags().StringVar(&baseDir, "base-dir", "./test-repos", "Base directory to verify")

	return cmd
}

// cleanupRepos safely removes test repositories.
func cleanupRepos(_ context.Context, baseDir string, force, dryRun bool) error {
	// Check if base directory exists
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		log.Printf("Base directory %s does not exist", baseDir)
		return nil
	}

	// Find all potential test repositories
	repos, err := findTestRepos(baseDir)
	if err != nil {
		return fmt.Errorf("failed to find test repositories: %w", err)
	}

	if len(repos) == 0 {
		log.Printf("No test repositories found in %s", baseDir)
		return nil
	}

	// Show what will be deleted
	log.Printf("Found %d test repositories:", len(repos))
	for _, repo := range repos {
		log.Printf("  - %s", repo)
	}

	if dryRun {
		log.Println("DRY RUN: No repositories were deleted")
		return nil
	}

	// Confirmation prompt
	if !force {
		if !confirmDeletion(len(repos)) {
			log.Println("Cleanup cancelled by user")
			return nil
		}
	}

	// Delete repositories
	for _, repo := range repos {
		if err := os.RemoveAll(repo); err != nil {
			log.Printf("❌ Failed to delete %s: %v", repo, err)
			continue
		}
		log.Printf("✓ Deleted: %s", repo)
	}

	log.Printf("Cleanup completed. Removed %d repositories", len(repos))
	return nil
}

// verifyRepos checks the structure of test repositories.
func verifyRepos(_ context.Context, baseDir string) error {
	repos, err := findTestRepos(baseDir)
	if err != nil {
		return fmt.Errorf("failed to find test repositories: %w", err)
	}

	log.Printf("Verifying %d test repositories:", len(repos))

	for _, repo := range repos {
		if err := verifyGitRepo(repo); err != nil {
			log.Printf("❌ %s: %v", filepath.Base(repo), err)
		} else {
			log.Printf("✓ %s: Valid Git repository", filepath.Base(repo))
		}
	}

	return nil
}

// findTestRepos locates test repositories based on naming patterns.
func findTestRepos(baseDir string) ([]string, error) {
	var repos []string

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", baseDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Match test repository naming patterns
		if strings.HasPrefix(name, "basic-repo-") ||
			strings.HasPrefix(name, "conflict-") ||
			strings.HasPrefix(name, "special-") {
			repoPath := filepath.Join(baseDir, name)
			repos = append(repos, repoPath)
		}
	}

	return repos, nil
}

// verifyGitRepo checks if a directory is a valid Git repository.
func verifyGitRepo(repoPath string) error {
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("not a Git repository")
	}
	return nil
}

// confirmDeletion prompts user for confirmation.
func confirmDeletion(count int) bool {
	fmt.Printf("\n⚠️  This will permanently delete %d test repositories.\n", count)
	fmt.Print("Are you sure you want to continue? (yes/no): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "yes" || response == "y"
}
