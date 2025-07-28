// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package git

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewGitRepoCmd creates the unified repository lifecycle management command.
func NewGitRepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Repository lifecycle management",
		Long: `Manage repositories across Git platforms including cloning, 
creating, archiving, and synchronizing repositories.

This command provides comprehensive repository management capabilities:
- Clone repositories with advanced features (bulk operations, parallel execution)
- List repositories from multiple Git platforms
- Create new repositories with templates and configurations
- Delete and archive repositories with safety checks
- Synchronize repositories across different platforms
- Migrate repositories between platforms
- Search repositories with advanced filtering`,
		Example: `
  # Clone repositories from an organization
  gz git repo clone --provider github --org myorg --target ./repos

  # List repositories from multiple providers
  gz git repo list --provider github --org myorg --format table

  # Create a new repository
  gz git repo create --provider github --org myorg --name newrepo --private

  # Synchronize a repository between platforms
  gz git repo sync --from github:myorg/repo --to gitlab:myorg/repo`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add subcommands for repository lifecycle management
	cmd.AddCommand(newRepoCloneCmd())
	cmd.AddCommand(newRepoListCmd())
	cmd.AddCommand(newRepoCreateCmd())
	cmd.AddCommand(newRepoDeleteCmd())
	cmd.AddCommand(newRepoArchiveCmd())
	cmd.AddCommand(newRepoSyncCmd())
	cmd.AddCommand(newRepoMigrateCmd())
	cmd.AddCommand(newRepoSearchCmd())

	return cmd
}

// Note: newRepoCloneCmd is now implemented in repo_clone.go

// newRepoListCmd creates the list subcommand.
func newRepoListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List repositories from Git platforms",
		Long: `List repositories from Git platforms with advanced filtering and formatting.

This command provides comprehensive repository listing capabilities including:
- Support for multiple Git platforms (GitHub, GitLab, Gitea)
- Advanced filtering by various criteria
- Multiple output formats (table, json, yaml, csv)
- Aggregation across multiple providers
- Real-time repository statistics`,
		Example: `
  # List repositories from a GitHub organization
  gz git repo list --provider github --org myorg

  # List with JSON output
  gz git repo list --provider gitlab --org mygroup --format json

  # List from all configured providers
  gz git repo list --all-providers --format table

  # List with advanced filtering
  gz git repo list --provider github --org myorg --language Go --min-stars 100`,
		RunE: runRepoList,
	}

	// Provider and organization flags
	cmd.Flags().String("provider", "", "Git provider (github, gitlab, gitea)")
	cmd.Flags().String("org", "", "Organization/Group name")
	cmd.Flags().Bool("all-providers", false, "List from all configured providers")

	// Filtering options
	cmd.Flags().String("match", "", "Repository name pattern (glob)")
	cmd.Flags().String("visibility", "all", "Repository visibility (public, private, all)")
	cmd.Flags().Bool("archived", true, "Include archived repositories")
	cmd.Flags().Bool("forks", true, "Include forked repositories")
	cmd.Flags().String("language", "", "Filter by primary language")
	cmd.Flags().Int("min-stars", 0, "Minimum star count")
	cmd.Flags().Int("max-stars", 0, "Maximum star count (0 = no limit)")
	cmd.Flags().String("updated-since", "", "Filter by last update date (YYYY-MM-DD)")

	// Output formatting
	cmd.Flags().String("format", "table", "Output format (table, json, yaml, csv)")
	cmd.Flags().String("columns", "", "Comma-separated list of columns to display")
	cmd.Flags().Bool("show-stats", false, "Include repository statistics")
	cmd.Flags().Int("limit", 0, "Limit number of results (0 = no limit)")
	cmd.Flags().String("sort", "name", "Sort by field (name, stars, updated, created)")
	cmd.Flags().Bool("reverse", false, "Reverse sort order")

	return cmd
}

// newRepoCreateCmd creates the create subcommand.
func newRepoCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new repository",
		Long: `Create a new repository on Git platforms with advanced configuration options.

This command provides comprehensive repository creation capabilities including:
- Template-based repository creation
- Advanced repository settings and permissions
- Automatic initialization with README, gitignore, and license
- Webhook and branch protection setup
- Integration with existing configuration templates`,
		Example: `
  # Create a basic repository
  gz git repo create --provider github --org myorg --name newrepo

  # Create a private repository with template
  gz git repo create --provider github --org myorg --name api-service --template api-template --private

  # Create with advanced settings
  gz git repo create --provider gitlab --org mygroup --name webapp --description "Web application" --auto-init --license MIT`,
		RunE: runRepoCreate,
	}

	// Basic repository information
	cmd.Flags().String("provider", "", "Git provider (github, gitlab, gitea)")
	cmd.Flags().String("org", "", "Organization/Group name")
	cmd.Flags().String("name", "", "Repository name")
	cmd.Flags().String("description", "", "Repository description")

	// Repository settings
	cmd.Flags().Bool("private", false, "Create private repository")
	cmd.Flags().String("template", "", "Template repository (org/repo)")
	cmd.Flags().Bool("auto-init", true, "Initialize with README")
	cmd.Flags().String("gitignore", "", "Gitignore template name")
	cmd.Flags().String("license", "", "License template (MIT, Apache-2.0, GPL-3.0, etc.)")

	// Advanced settings
	cmd.Flags().Bool("issues", true, "Enable issues")
	cmd.Flags().Bool("projects", false, "Enable projects")
	cmd.Flags().Bool("wiki", false, "Enable wiki")
	cmd.Flags().Bool("downloads", true, "Enable downloads")
	cmd.Flags().String("homepage", "", "Repository homepage URL")
	cmd.Flags().StringSlice("topics", []string{}, "Repository topics/tags")

	// Protection and webhooks
	cmd.Flags().Bool("protect-main", false, "Enable branch protection for main branch")
	cmd.Flags().StringSlice("webhooks", []string{}, "Webhook URLs to add")

	// Output options
	cmd.Flags().Bool("clone-after", false, "Clone repository after creation")
	cmd.Flags().String("clone-path", "", "Local path to clone to")

	// Mark required flags
	cmd.MarkFlagRequired("provider")
	cmd.MarkFlagRequired("org")
	cmd.MarkFlagRequired("name")

	return cmd
}

// newRepoDeleteCmd creates the delete subcommand.
func newRepoDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete repositories with safety checks",
		Long: `Delete repositories from Git platforms with comprehensive safety checks.

This command provides safe repository deletion with:
- Interactive confirmation prompts
- Backup options before deletion
- Bulk deletion with filters
- Archive-before-delete option
- Audit logging of deletions`,
		Example: `
  # Delete a single repository
  gz git repo delete --provider github --org myorg --name oldrepo

  # Delete with backup
  gz git repo delete --provider github --org myorg --name oldrepo --backup ./backups

  # Bulk delete archived repositories
  gz git repo delete --provider github --org myorg --archived-only --confirm`,
		RunE: runRepoDelete,
	}

	// Repository identification
	cmd.Flags().String("provider", "", "Git provider (github, gitlab, gitea)")
	cmd.Flags().String("org", "", "Organization/Group name")
	cmd.Flags().String("name", "", "Repository name (required for single deletion)")

	// Bulk deletion options
	cmd.Flags().String("match", "", "Repository name pattern for bulk deletion")
	cmd.Flags().Bool("archived-only", false, "Only delete archived repositories")
	cmd.Flags().Bool("empty-only", false, "Only delete empty repositories")
	cmd.Flags().String("older-than", "", "Delete repositories older than date (YYYY-MM-DD)")

	// Safety options
	cmd.Flags().Bool("confirm", false, "Skip interactive confirmation")
	cmd.Flags().Bool("dry-run", false, "Show what would be deleted without doing it")
	cmd.Flags().String("backup", "", "Backup directory before deletion")
	cmd.Flags().Bool("archive-first", false, "Archive repository before deletion")

	// Mark required flags for single deletion
	cmd.MarkFlagRequired("provider")
	cmd.MarkFlagRequired("org")

	return cmd
}

// newRepoArchiveCmd creates the archive subcommand.
func newRepoArchiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive",
		Short: "Archive or unarchive repositories",
		Long: `Archive or unarchive repositories on Git platforms.

This command provides repository archival management including:
- Single and bulk archive operations
- Unarchive operations
- Archive with reason and metadata
- Scheduled archival based on criteria`,
		Example: `
  # Archive a single repository
  gz git repo archive --provider github --org myorg --name oldrepo

  # Unarchive a repository
  gz git repo archive --provider github --org myorg --name repo --unarchive

  # Bulk archive inactive repositories
  gz git repo archive --provider github --org myorg --inactive-days 365 --reason "Inactive project"`,
		RunE: runRepoArchive,
	}

	// Repository identification
	cmd.Flags().String("provider", "", "Git provider (github, gitlab, gitea)")
	cmd.Flags().String("org", "", "Organization/Group name")
	cmd.Flags().String("name", "", "Repository name (required for single operation)")

	// Archive/unarchive options
	cmd.Flags().Bool("unarchive", false, "Unarchive instead of archive")
	cmd.Flags().String("reason", "", "Reason for archiving")

	// Bulk operations
	cmd.Flags().String("match", "", "Repository name pattern for bulk operations")
	cmd.Flags().Int("inactive-days", 0, "Archive repositories inactive for N days")
	cmd.Flags().Bool("no-activity", false, "Archive repositories with no commits")

	// Safety options
	cmd.Flags().Bool("dry-run", false, "Show what would be archived without doing it")
	cmd.Flags().Bool("confirm", false, "Skip interactive confirmation")

	// Mark required flags
	cmd.MarkFlagRequired("provider")
	cmd.MarkFlagRequired("org")

	return cmd
}

// newRepoSyncCmd creates the sync subcommand.
func newRepoSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize repositories across platforms",
		Long: `Synchronize repositories between different Git platforms.

This command provides comprehensive cross-platform synchronization including:
- Code synchronization between platforms
- Issue and pull request migration
- Wiki and documentation sync
- Metadata and settings synchronization
- Continuous synchronization with scheduling`,
		Example: `
  # Sync a single repository
  gz git repo sync --from github:myorg/repo --to gitlab:myorg/repo

  # Sync with issues and PRs
  gz git repo sync --from github:myorg/repo --to gitlab:myorg/repo --include-issues --include-prs

  # Dry run to preview sync
  gz git repo sync --from github:myorg/repo --to gitlab:myorg/repo --dry-run`,
		RunE: runRepoSync,
	}

	// Source and destination
	cmd.Flags().String("from", "", "Source repository (provider:org/repo)")
	cmd.Flags().String("to", "", "Destination repository (provider:org/repo)")

	// Sync options
	cmd.Flags().Bool("include-issues", false, "Synchronize issues")
	cmd.Flags().Bool("include-prs", false, "Synchronize pull/merge requests")
	cmd.Flags().Bool("include-wiki", false, "Synchronize wiki")
	cmd.Flags().Bool("include-releases", false, "Synchronize releases")
	cmd.Flags().Bool("include-settings", false, "Synchronize repository settings")

	// Sync behavior
	cmd.Flags().String("strategy", "merge", "Sync strategy (merge, overwrite, bidirectional)")
	cmd.Flags().Bool("create-missing", false, "Create destination repository if it doesn't exist")
	cmd.Flags().Bool("dry-run", false, "Preview sync without making changes")

	// Filtering
	cmd.Flags().String("since", "", "Only sync items since date (YYYY-MM-DD)")
	cmd.Flags().StringSlice("labels", []string{}, "Filter issues/PRs by labels")

	// Mark required flags
	cmd.MarkFlagRequired("from")
	cmd.MarkFlagRequired("to")

	return cmd
}

// newRepoMigrateCmd creates the migrate subcommand.
func newRepoMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate repositories between platforms",
		Long: `Migrate repositories between different Git platforms with full history preservation.

This command provides comprehensive repository migration including:
- Complete git history migration
- Issue and pull request migration with linking
- Wiki and documentation migration
- User and permission mapping
- Bulk organization migration`,
		Example: `
  # Migrate a single repository
  gz git repo migrate --from github:oldorg/repo --to gitlab:neworg/repo

  # Migrate entire organization
  gz git repo migrate --from-org github:oldorg --to-org gitlab:neworg

  # Migrate with user mapping
  gz git repo migrate --from github:oldorg/repo --to gitlab:neworg/repo --user-mapping users.yaml`,
		RunE: runRepoMigrate,
	}

	// Migration source and destination
	cmd.Flags().String("from", "", "Source repository (provider:org/repo)")
	cmd.Flags().String("to", "", "Destination repository (provider:org/repo)")
	cmd.Flags().String("from-org", "", "Source organization (provider:org)")
	cmd.Flags().String("to-org", "", "Destination organization (provider:org)")

	// Migration options
	cmd.Flags().Bool("include-history", true, "Migrate complete git history")
	cmd.Flags().Bool("include-issues", true, "Migrate issues")
	cmd.Flags().Bool("include-prs", true, "Migrate pull/merge requests")
	cmd.Flags().Bool("include-wiki", true, "Migrate wiki")
	cmd.Flags().Bool("include-releases", true, "Migrate releases")

	// User and permission mapping
	cmd.Flags().String("user-mapping", "", "User mapping file (YAML)")
	cmd.Flags().Bool("preserve-authors", true, "Preserve commit authors")

	// Migration behavior
	cmd.Flags().Bool("delete-source", false, "Delete source repository after migration")
	cmd.Flags().Bool("archive-source", false, "Archive source repository after migration")
	cmd.Flags().Bool("dry-run", false, "Preview migration without making changes")

	return cmd
}

// newRepoSearchCmd creates the search subcommand.
func newRepoSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search repositories with advanced filtering",
		Long: `Search repositories across Git platforms with advanced filtering and ranking.

This command provides comprehensive repository search including:
- Cross-platform search capabilities
- Advanced filtering by multiple criteria
- Code content search
- Topic and language-based search
- Custom ranking and sorting`,
		Example: `
  # Search by name
  gz git repo search --query "api" --provider github

  # Search by language and stars
  gz git repo search --language Go --min-stars 100 --provider github

  # Search by topic
  gz git repo search --topic "kubernetes" --provider github

  # Advanced search with code content
  gz git repo search --code "func main" --language Go --provider github`,
		RunE: runRepoSearch,
	}

	// Search parameters
	cmd.Flags().String("query", "", "Search query")
	cmd.Flags().String("provider", "", "Git provider (github, gitlab, gitea)")
	cmd.Flags().Bool("all-providers", false, "Search across all providers")

	// Content filters
	cmd.Flags().String("language", "", "Programming language")
	cmd.Flags().StringSlice("topics", []string{}, "Repository topics")
	cmd.Flags().String("code", "", "Search in code content")
	cmd.Flags().String("filename", "", "Search in filenames")

	// Quality filters
	cmd.Flags().Int("min-stars", 0, "Minimum star count")
	cmd.Flags().Int("min-forks", 0, "Minimum fork count")
	cmd.Flags().String("created-after", "", "Created after date (YYYY-MM-DD)")
	cmd.Flags().String("updated-after", "", "Updated after date (YYYY-MM-DD)")

	// Search scope
	cmd.Flags().String("user", "", "Limit to specific user/organization")
	cmd.Flags().Bool("fork", false, "Include forked repositories")
	cmd.Flags().String("license", "", "Repository license")

	// Output options
	cmd.Flags().String("format", "table", "Output format (table, json, yaml)")
	cmd.Flags().Int("limit", 30, "Maximum number of results")
	cmd.Flags().String("sort", "best-match", "Sort by (best-match, stars, forks, updated)")

	return cmd
}

// Command execution functions (placeholder implementations)
// Note: runRepoClone is now implemented in repo_clone.go

func runRepoList(cmd *cobra.Command, args []string) error {
	// TODO: Implement repository listing
	return fmt.Errorf("list command not yet implemented")
}

func runRepoCreate(cmd *cobra.Command, args []string) error {
	// TODO: Implement repository creation
	return fmt.Errorf("create command not yet implemented")
}

func runRepoDelete(cmd *cobra.Command, args []string) error {
	// TODO: Implement repository deletion
	return fmt.Errorf("delete command not yet implemented")
}

func runRepoArchive(cmd *cobra.Command, args []string) error {
	// TODO: Implement repository archiving
	return fmt.Errorf("archive command not yet implemented")
}

func runRepoSync(cmd *cobra.Command, args []string) error {
	// TODO: Implement repository synchronization
	return fmt.Errorf("sync command not yet implemented")
}

func runRepoMigrate(cmd *cobra.Command, args []string) error {
	// TODO: Implement repository migration
	return fmt.Errorf("migrate command not yet implemented")
}

func runRepoSearch(cmd *cobra.Command, args []string) error {
	// TODO: Implement repository search
	return fmt.Errorf("search command not yet implemented")
}
