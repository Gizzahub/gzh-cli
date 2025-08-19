// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package git

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/pkg/git/provider"
)

// DeleteOptions contains options for repository deletion.
type DeleteOptions struct {
	// Provider and target
	Provider string
	Repos    []string
	Org      string

	// Pattern matching
	Match string

	// Safety options
	Force  bool
	DryRun bool
	Backup bool

	// Backup options
	BackupPath   string
	BackupFormat string

	// Output options
	Format string
	Quiet  bool
}

// newRepoDeleteCmd creates the repo delete command.
func newRepoDeleteCmd() *cobra.Command {
	opts := &DeleteOptions{
		Format: "table",
	}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete repositories with safety checks",
		Long: `Delete one or more repositories with comprehensive safety checks and confirmation.

This command provides safe repository deletion with:
- Interactive confirmation prompts
- Pattern matching for bulk operations
- Dry run capability
- Safety checks to prevent accidental deletion
- Support for backing up repositories before deletion`,
		Example: `  # Delete a single repository
  gz git repo delete --provider github --repo myorg/oldrepo

  # Delete multiple repositories
  gz git repo delete --provider gitlab --repo myorg/repo1 --repo myorg/repo2

  # Delete with pattern matching (requires --force)
  gz git repo delete --provider github --org myorg --match "test-*" --force

  # Dry run to preview deletion
  gz git repo delete --provider github --org myorg --match "deprecated-*" --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRepoDelete(cmd.Context(), opts)
		},
	}

	// Provider and target flags
	cmd.Flags().StringVar(&opts.Provider, "provider", "", "Git provider (github, gitlab, gitea, gogs)")
	cmd.Flags().StringSliceVar(&opts.Repos, "repo", nil, "Repository to delete (org/repo format)")
	cmd.Flags().StringVar(&opts.Org, "org", "", "Organization for pattern matching")

	// Pattern matching
	cmd.Flags().StringVar(&opts.Match, "match", "", "Repository name pattern (regex)")

	// Safety options
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Skip confirmation prompts")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Preview without deleting")
	cmd.Flags().BoolVar(&opts.Backup, "backup", false, "Create backup before deletion")

	// Backup options
	cmd.Flags().StringVar(&opts.BackupPath, "backup-path", "", "Directory to store backups (required when backup is enabled)")
	cmd.Flags().StringVar(&opts.BackupFormat, "backup-format", "clone", "Backup format (clone, archive, bundle)")

	// Output options
	cmd.Flags().StringVar(&opts.Format, "format", "table", "Output format (table, json, yaml)")
	cmd.Flags().BoolVar(&opts.Quiet, "quiet", false, "Suppress output")

	// Mark required flags
	cmd.MarkFlagRequired("provider")

	// Validation rules
	cmd.MarkFlagsMutuallyExclusive("repo", "match")
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if len(opts.Repos) == 0 && opts.Match == "" {
			return fmt.Errorf("either --repo or --match must be specified")
		}
		if opts.Match != "" && opts.Org == "" {
			return fmt.Errorf("--org is required when using --match")
		}
		return nil
	}

	return cmd
}

// runRepoDelete executes the repository deletion operation.
func runRepoDelete(ctx context.Context, opts *DeleteOptions) error {
	// Validate options
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	// Safety check for pattern matching
	if opts.Match != "" && !opts.Force {
		return fmt.Errorf("pattern matching requires --force flag for safety")
	}

	// Get provider
	gitProvider, err := getGitProvider(opts.Provider, opts.Org)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Get repositories to delete
	repos, err := opts.getTargetRepositories(ctx, gitProvider)
	if err != nil {
		return fmt.Errorf("failed to get target repositories: %w", err)
	}

	if len(repos) == 0 {
		if !opts.Quiet {
			fmt.Println("No repositories found matching the criteria")
		}
		return nil
	}

	// Dry run
	if opts.DryRun {
		return opts.showDryRun(repos)
	}

	// Confirmation prompt
	if !opts.Force {
		if !confirmDeletion(repos) {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	// Execute deletion
	return opts.deleteRepositories(ctx, gitProvider, repos)
}

// Validate validates the delete options.
func (opts *DeleteOptions) Validate() error {
	if opts.Provider == "" {
		return fmt.Errorf("provider is required")
	}

	// Validate pattern if provided
	if opts.Match != "" {
		if _, err := regexp.Compile(opts.Match); err != nil {
			return fmt.Errorf("invalid match pattern: %w", err)
		}
	}

	// Validate repository names if provided
	for _, repo := range opts.Repos {
		if !strings.Contains(repo, "/") {
			return fmt.Errorf("repository must be in format 'org/repo': %s", repo)
		}
	}

	// Validate output format
	if !isValidOutputFormat(opts.Format) {
		return fmt.Errorf("invalid output format: %s", opts.Format)
	}

	return nil
}

// getTargetRepositories retrieves the repositories to be deleted.
func (opts *DeleteOptions) getTargetRepositories(ctx context.Context, gitProvider provider.GitProvider) ([]provider.Repository, error) {
	var repos []provider.Repository

	if len(opts.Repos) > 0 {
		// Get specific repositories
		for _, repoName := range opts.Repos {
			// Parse org/repo format
			parts := strings.SplitN(repoName, "/", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid repository format: %s (expected org/repo)", repoName)
			}

			// Get repository by full name
			repo, err := gitProvider.GetRepository(ctx, repoName)
			if err != nil {
				return nil, fmt.Errorf("failed to get repository %s: %w", repoName, err)
			}

			repos = append(repos, *repo)
		}
	} else if opts.Match != "" {
		// Get repositories by pattern
		listOpts := provider.ListOptions{
			Organization: opts.Org,
			PerPage:      100,
		}

		repoList, err := gitProvider.ListRepositories(ctx, listOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}

		// Filter by pattern
		pattern, err := regexp.Compile(opts.Match)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern: %w", err)
		}

		for _, repo := range repoList.Repositories {
			if pattern.MatchString(repo.Name) {
				repos = append(repos, repo)
			}
		}
	}

	return repos, nil
}

// showDryRun displays what would be deleted without actually deleting.
func (opts *DeleteOptions) showDryRun(repos []provider.Repository) error {
	if !opts.Quiet {
		fmt.Printf("Dry run - would delete %d repositories:\n\n", len(repos))

		for _, repo := range repos {
			fmt.Printf("  ❌ %s\n", repo.FullName)
			if repo.Description != "" {
				fmt.Printf("     Description: %s\n", repo.Description)
			}
			fmt.Printf("     Private: %v, Archived: %v\n", repo.Private, repo.Archived)
			fmt.Printf("     Last updated: %s\n", repo.UpdatedAt.Format("2006-01-02 15:04:05"))
			fmt.Println()
		}

		fmt.Printf("Total repositories to delete: %d\n", len(repos))
		fmt.Println("\nUse --force to skip confirmation when ready to proceed.")
	}

	return nil
}

// deleteRepositories executes the actual deletion.
func (opts *DeleteOptions) deleteRepositories(ctx context.Context, gitProvider provider.GitProvider, repos []provider.Repository) error {
	var errors []error
	successCount := 0

	for i, repo := range repos {
		if !opts.Quiet {
			fmt.Printf("[%d/%d] Deleting %s...", i+1, len(repos), repo.FullName)
		}

		// Create backup if requested
		if opts.Backup {
			if err := opts.createBackup(ctx, gitProvider, &repo); err != nil {
				if !opts.Quiet {
					fmt.Printf(" backup failed: %v\n", err)
				}
				errors = append(errors, fmt.Errorf("backup failed for %s: %w", repo.FullName, err))
				continue
			}
			if !opts.Quiet {
				fmt.Print(" backed up...")
			}
		}

		// Delete repository
		if err := gitProvider.DeleteRepository(ctx, repo.ID); err != nil {
			if !opts.Quiet {
				fmt.Printf(" ❌ failed: %v\n", err)
			}
			errors = append(errors, fmt.Errorf("failed to delete %s: %w", repo.FullName, err))
			continue
		}

		successCount++
		if !opts.Quiet {
			fmt.Printf(" ✅ deleted\n")
		}
	}

	// Summary
	if !opts.Quiet {
		fmt.Printf("\nDeletion summary:\n")
		fmt.Printf("  Successful: %d\n", successCount)
		fmt.Printf("  Failed: %d\n", len(errors))

		if len(errors) > 0 {
			fmt.Printf("\nErrors:\n")
			for _, err := range errors {
				fmt.Printf("  - %v\n", err)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("deletion completed with %d errors", len(errors))
	}

	return nil
}

// createBackup creates a backup of the repository before deletion.
func (opts *DeleteOptions) createBackup(ctx context.Context, gitProvider provider.GitProvider, repo *provider.Repository) error {
	if opts.BackupPath == "" {
		return fmt.Errorf("backup path is required when backup is enabled")
	}

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(opts.BackupPath, 0o755); err != nil {
		return fmt.Errorf("failed to create backup directory %s: %w", opts.BackupPath, err)
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("%s-%s", repo.Name, timestamp)

	// Choose backup method based on options
	switch opts.BackupFormat {
	case "clone":
		return opts.createCloneBackup(ctx, repo, backupName)
	case "archive":
		return opts.createArchiveBackup(ctx, gitProvider, repo, backupName)
	case "bundle":
		return opts.createBundleBackup(ctx, repo, backupName)
	default:
		return opts.createCloneBackup(ctx, repo, backupName) // Default to clone
	}
}

// createCloneBackup creates a backup by cloning the repository
func (opts *DeleteOptions) createCloneBackup(ctx context.Context, repo *provider.Repository, backupName string) error {
	backupPath := filepath.Join(opts.BackupPath, backupName)

	fmt.Printf("📦 Creating clone backup: %s\n", backupPath)

	// Use git clone command to create backup
	cmd := exec.CommandContext(ctx, "git", "clone", "--mirror", repo.CloneURL, backupPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to clone repository for backup: %w\nOutput: %s", err, string(output))
	}

	// Create metadata file
	metadataPath := filepath.Join(opts.BackupPath, backupName+".metadata.json")
	metadata := BackupMetadata{
		Repository:  *repo,
		BackupTime:  time.Now(),
		BackupType:  "clone",
		BackupPath:  backupPath,
		OriginalURL: repo.CloneURL,
	}

	if err := opts.saveBackupMetadata(metadataPath, metadata); err != nil {
		fmt.Printf("⚠️  Warning: Failed to save backup metadata: %v\n", err)
	}

	fmt.Printf("✅ Clone backup completed: %s\n", backupPath)
	return nil
}

// createArchiveBackup creates a backup using git archive
func (opts *DeleteOptions) createArchiveBackup(ctx context.Context, gitProvider provider.GitProvider, repo *provider.Repository, backupName string) error {
	// First clone to a temporary location
	tempDir, err := os.MkdirTemp("", "git-archive-backup-")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Clone repository
	cloneCmd := exec.CommandContext(ctx, "git", "clone", repo.CloneURL, tempDir)
	if output, err := cloneCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to clone repository for archive: %w\nOutput: %s", err, string(output))
	}

	// Create archive
	archivePath := filepath.Join(opts.BackupPath, backupName+".tar.gz")
	fmt.Printf("📦 Creating archive backup: %s\n", archivePath)

	archiveCmd := exec.CommandContext(ctx, "git", "-C", tempDir, "archive", "--format=tar.gz", "--output", archivePath, "HEAD")
	if output, err := archiveCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create git archive: %w\nOutput: %s", err, string(output))
	}

	// Create metadata file
	metadataPath := filepath.Join(opts.BackupPath, backupName+".metadata.json")
	metadata := BackupMetadata{
		Repository:  *repo,
		BackupTime:  time.Now(),
		BackupType:  "archive",
		BackupPath:  archivePath,
		OriginalURL: repo.CloneURL,
	}

	if err := opts.saveBackupMetadata(metadataPath, metadata); err != nil {
		fmt.Printf("⚠️  Warning: Failed to save backup metadata: %v\n", err)
	}

	fmt.Printf("✅ Archive backup completed: %s\n", archivePath)
	return nil
}

// createBundleBackup creates a backup using git bundle
func (opts *DeleteOptions) createBundleBackup(ctx context.Context, repo *provider.Repository, backupName string) error {
	// First clone to a temporary location
	tempDir, err := os.MkdirTemp("", "git-bundle-backup-")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Clone repository
	cloneCmd := exec.CommandContext(ctx, "git", "clone", repo.CloneURL, tempDir)
	if output, err := cloneCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to clone repository for bundle: %w\nOutput: %s", err, string(output))
	}

	// Create bundle
	bundlePath := filepath.Join(opts.BackupPath, backupName+".bundle")
	fmt.Printf("📦 Creating bundle backup: %s\n", bundlePath)

	bundleCmd := exec.CommandContext(ctx, "git", "-C", tempDir, "bundle", "create", bundlePath, "--all")
	if output, err := bundleCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create git bundle: %w\nOutput: %s", err, string(output))
	}

	// Create metadata file
	metadataPath := filepath.Join(opts.BackupPath, backupName+".metadata.json")
	metadata := BackupMetadata{
		Repository:  *repo,
		BackupTime:  time.Now(),
		BackupType:  "bundle",
		BackupPath:  bundlePath,
		OriginalURL: repo.CloneURL,
	}

	if err := opts.saveBackupMetadata(metadataPath, metadata); err != nil {
		fmt.Printf("⚠️  Warning: Failed to save backup metadata: %v\n", err)
	}

	fmt.Printf("✅ Bundle backup completed: %s\n", bundlePath)
	return nil
}

// BackupMetadata contains metadata about a repository backup
type BackupMetadata struct {
	Repository  provider.Repository `json:"repository"`
	BackupTime  time.Time           `json:"backup_time"`
	BackupType  string              `json:"backup_type"`
	BackupPath  string              `json:"backup_path"`
	OriginalURL string              `json:"original_url"`
}

// saveBackupMetadata saves backup metadata to a JSON file
func (opts *DeleteOptions) saveBackupMetadata(metadataPath string, metadata BackupMetadata) error {
	file, err := os.Create(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(metadata); err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}

	return nil
}

// confirmDeletion prompts the user for confirmation before deletion.
func confirmDeletion(repos []provider.Repository) bool {
	fmt.Printf("\n⚠️  You are about to delete %d repositories:\n\n", len(repos))

	for _, repo := range repos {
		status := "public"
		if repo.Private {
			status = "private"
		}
		if repo.Archived {
			status += ", archived"
		}

		fmt.Printf("  ❌ %s (%s)\n", repo.FullName, status)
	}

	fmt.Printf("\nThis action cannot be undone!\n")
	fmt.Printf("Type 'DELETE' to confirm deletion: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.TrimSpace(input)
	return input == "DELETE"
}
