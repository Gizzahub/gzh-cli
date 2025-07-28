// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package git

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-manager-go/pkg/git/provider"
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
	// TODO: Implement backup functionality
	// This could involve:
	// 1. Cloning the repository locally
	// 2. Creating an archive/export
	// 3. Saving to a backup location
	return fmt.Errorf("backup functionality not implemented yet")
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
