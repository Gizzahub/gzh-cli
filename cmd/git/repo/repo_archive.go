// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repo

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/pkg/git/provider"
)

// ArchiveOptions contains options for repository archiving.
type ArchiveOptions struct {
	// Provider and target
	Provider string
	Repos    []string
	Org      string

	// Pattern matching
	Match string

	// Operation mode
	Unarchive bool

	// Safety options
	DryRun bool
	Force  bool

	// Output options
	Format string
	Quiet  bool
}

// newRepoArchiveCmd creates the repo archive command.
func newRepoArchiveCmd() *cobra.Command {
	opts := &ArchiveOptions{
		Format: "table",
	}

	cmd := &cobra.Command{
		Use:   "archive",
		Short: "Archive or unarchive repositories",
		Long: `Archive or unarchive repositories to make them read-only while preserving all data.

This command provides repository archival management including:
- Single and bulk archive operations
- Unarchive operations to restore repository activity
- Pattern matching for bulk operations
- Dry run capability to preview changes

Archived repositories are read-only but preserve all git history, issues,
pull requests, and other repository data.`,
		Example: `  # Archive a single repository
  gz git repo archive --provider github --repo myorg/oldproject

  # Archive multiple repositories
  gz git repo archive --provider github --repo myorg/proj1 --repo myorg/proj2

  # Archive repositories matching pattern
  gz git repo archive --provider github --org myorg --match "deprecated-*"

  # Unarchive a repository
  gz git repo archive --provider github --repo myorg/project --unarchive

  # Dry run to preview archival
  gz git repo archive --provider github --org myorg --match "old-*" --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRepoArchive(cmd.Context(), opts)
		},
	}

	// Provider and target flags
	cmd.Flags().StringVar(&opts.Provider, "provider", "", "Git provider (github, gitlab, gitea, gogs)")
	cmd.Flags().StringSliceVar(&opts.Repos, "repo", nil, "Repository to archive (org/repo format)")
	cmd.Flags().StringVar(&opts.Org, "org", "", "Organization for pattern matching")

	// Pattern matching
	cmd.Flags().StringVar(&opts.Match, "match", "", "Repository name pattern (regex)")

	// Operation mode
	cmd.Flags().BoolVar(&opts.Unarchive, "unarchive", false, "Unarchive instead of archive")

	// Safety options
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Preview without archiving")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Skip confirmation prompts")

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

// runRepoArchive executes the repository archive operation.
func runRepoArchive(ctx context.Context, opts *ArchiveOptions) error {
	// Validate options
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	// Get provider
	gitProvider, err := getGitProvider(opts.Provider, opts.Org)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Get repositories to archive/unarchive
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

	// Filter repositories based on current archive status
	filteredRepos := opts.filterByArchiveStatus(repos)
	if len(filteredRepos) == 0 {
		if !opts.Quiet {
			if opts.Unarchive {
				fmt.Println("No archived repositories found to unarchive")
			} else {
				fmt.Println("No active repositories found to archive")
			}
		}
		return nil
	}

	// Dry run
	if opts.DryRun {
		return opts.showDryRun(filteredRepos)
	}

	// Confirmation prompt for bulk operations
	if !opts.Force && len(filteredRepos) > 1 {
		if !opts.confirmOperation(filteredRepos) {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	// Execute archive/unarchive operation
	return opts.executeOperation(ctx, gitProvider, filteredRepos)
}

// Validate validates the archive options.
func (opts *ArchiveOptions) Validate() error {
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

// getTargetRepositories retrieves the repositories to be archived/unarchived.
func (opts *ArchiveOptions) getTargetRepositories(ctx context.Context, gitProvider provider.GitProvider) ([]provider.Repository, error) {
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

// filterByArchiveStatus filters repositories based on their current archive status.
func (opts *ArchiveOptions) filterByArchiveStatus(repos []provider.Repository) []provider.Repository {
	var filtered []provider.Repository

	for _, repo := range repos {
		if opts.Unarchive {
			// For unarchive, only include archived repositories
			if repo.Archived {
				filtered = append(filtered, repo)
			}
		} else {
			// For archive, only include non-archived repositories
			if !repo.Archived {
				filtered = append(filtered, repo)
			}
		}
	}

	return filtered
}

// showDryRun displays what would be archived/unarchived without actually doing it.
func (opts *ArchiveOptions) showDryRun(repos []provider.Repository) error {
	if !opts.Quiet {
		operation := "archive"
		if opts.Unarchive {
			operation = "unarchive"
		}

		fmt.Printf("Dry run - would %s %d repositories:\n\n", operation, len(repos))

		for _, repo := range repos {
			emoji := "📦"
			if opts.Unarchive {
				emoji = "🔓"
			}

			fmt.Printf("  %s %s\n", emoji, repo.FullName)
			if repo.Description != "" {
				fmt.Printf("     Description: %s\n", repo.Description)
			}
			fmt.Printf("     Private: %v, Current status: %s\n",
				repo.Private,
				map[bool]string{true: "archived", false: "active"}[repo.Archived])
			fmt.Printf("     Last updated: %s\n", repo.UpdatedAt.Format("2006-01-02 15:04:05"))
			fmt.Println()
		}

		fmt.Printf("Total repositories to %s: %d\n", operation, len(repos))
	}

	return nil
}

// confirmOperation prompts for confirmation before bulk operations.
func (opts *ArchiveOptions) confirmOperation(repos []provider.Repository) bool {
	operation := "archive"
	emoji := "📦"
	if opts.Unarchive {
		operation = "unarchive"
		emoji = "🔓"
	}

	fmt.Printf("\n%s You are about to %s %d repositories:\n\n", emoji, operation, len(repos))

	for _, repo := range repos {
		status := "public"
		if repo.Private {
			status = "private"
		}

		fmt.Printf("  %s %s (%s)\n", emoji, repo.FullName, status)
	}

	fmt.Printf("\nProceed with %s operation? [y/N]: ", operation)

	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// executeOperation executes the archive/unarchive operation.
func (opts *ArchiveOptions) executeOperation(ctx context.Context, gitProvider provider.GitProvider, repos []provider.Repository) error {
	var errors []error
	successCount := 0

	operation := "Archiving"
	emoji := "📦"
	if opts.Unarchive {
		operation = "Unarchiving"
		emoji = "🔓"
	}

	for i, repo := range repos {
		if !opts.Quiet {
			fmt.Printf("[%d/%d] %s %s...", i+1, len(repos), operation, repo.FullName)
		}

		var err error
		if opts.Unarchive {
			err = gitProvider.UnarchiveRepository(ctx, repo.ID)
		} else {
			err = gitProvider.ArchiveRepository(ctx, repo.ID)
		}

		if err != nil {
			if !opts.Quiet {
				fmt.Printf(" ❌ failed: %v\n", err)
			}
			errors = append(errors, fmt.Errorf("failed to %s %s: %w",
				strings.ToLower(operation[:len(operation)-3]), repo.FullName, err))
			continue
		}

		successCount++
		if !opts.Quiet {
			fmt.Printf(" %s success\n", emoji)
		}
	}

	// Summary
	if !opts.Quiet {
		fmt.Printf("\n%s summary:\n", operation)
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
		return fmt.Errorf("%s completed with %d errors", strings.ToLower(operation), len(errors))
	}

	return nil
}
