// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/internal/git/sync"
	"github.com/gizzahub/gzh-cli/pkg/git/provider"
)

// newRepoSyncCmd defines the sync command under repo.
func newRepoSyncCmd() *cobra.Command {
	var opts sync.SyncOptions

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize repositories across Git platforms",
		Long: `Synchronize repositories between different Git platforms including:
- Repository code and branches
- Issues and pull requests (if supported)
- Wiki content
- Releases and tags
- Repository settings and metadata`,
		Example: `
  # Sync a single repository
  gz git repo sync --from github:myorg/repo --to gitlab:mygroup/repo

  # Sync entire organization
  gz git repo sync --from github:myorg --to gitea:myorg --create-missing

  # Sync with specific features
  gz git repo sync --from github:org/repo --to gitlab:group/repo \
    --include-issues --include-wiki --include-releases

  # Dry run to preview changes
  gz git repo sync --from github:org/repo --to gitlab:group/repo --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync(cmd.Context(), opts)
		},
	}

	// Source and destination
	cmd.Flags().StringVar(&opts.From, "from", "", "Source (provider:org/repo or provider:org)")
	cmd.Flags().StringVar(&opts.To, "to", "", "Destination (provider:org/repo or provider:org)")

	// Sync options
	cmd.Flags().BoolVar(&opts.CreateMissing, "create-missing", false, "Create repos that don't exist in destination")
	cmd.Flags().BoolVar(&opts.UpdateExisting, "update-existing", true, "Update existing repositories")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Force push (destructive)")

	// Include options
	cmd.Flags().BoolVar(&opts.IncludeCode, "include-code", true, "Sync repository code")
	cmd.Flags().BoolVar(&opts.IncludeIssues, "include-issues", false, "Sync issues")
	cmd.Flags().BoolVar(&opts.IncludePRs, "include-prs", false, "Sync pull/merge requests")
	cmd.Flags().BoolVar(&opts.IncludeWiki, "include-wiki", false, "Sync wiki")
	cmd.Flags().BoolVar(&opts.IncludeReleases, "include-releases", false, "Sync releases")
	cmd.Flags().BoolVar(&opts.IncludeSettings, "include-settings", false, "Sync repository settings")

	// Filtering
	cmd.Flags().StringVar(&opts.Match, "match", "", "Repository name pattern")
	cmd.Flags().StringVar(&opts.Exclude, "exclude", "", "Exclude pattern")

	// Execution options
	cmd.Flags().IntVar(&opts.Parallel, "parallel", 1, "Parallel sync workers")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Preview without making changes")
	cmd.Flags().BoolVar(&opts.Verbose, "verbose", false, "Verbose output")

	cmd.MarkFlagRequired("from")
	cmd.MarkFlagRequired("to")

	return cmd
}

func newRepoMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Migrate repositories between platforms",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("migrate command not yet implemented")
		},
	}
}

// SearchOptions holds flags for repository search.
type SearchOptions struct {
	Provider string
	Query    string
	Language string
	Org      string
	User     string
	Topic    string
	Stars    string
	Sort     string
	Order    string
	Format   string
	Limit    int
	Page     int
}

func newRepoSearchCmd() *cobra.Command {
	opts := &SearchOptions{
		Sort:   "best-match",
		Order:  "desc",
		Format: "table",
		Limit:  30,
		Page:   1,
	}

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search repositories with advanced filtering",
		Long: `Search repositories across a Git provider using the platform search API.

Supported providers: github, gitlab, gitea. Clear errors are returned for
unsupported providers or when the query is missing.`,
		Example: `  # Basic search
  gz git repo search --provider github --query "golang cli"

  # Search with language and sort
  gz git repo search --provider github --query api --language Go --sort stars

  # Limit to an organization
  gz git repo search --provider github --query terraform --org myorg --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRepoSearch(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.Provider, "provider", "", "Git provider (github, gitlab, gitea)")
	cmd.Flags().StringVar(&opts.Query, "query", "", "Search query string (required)")
	cmd.Flags().StringVar(&opts.Language, "language", "", "Filter by programming language")
	cmd.Flags().StringVar(&opts.Org, "org", "", "Limit search to an organization/group")
	cmd.Flags().StringVar(&opts.User, "user", "", "Limit search to a user")
	cmd.Flags().StringVar(&opts.Topic, "topic", "", "Filter by topic")
	cmd.Flags().StringVar(&opts.Stars, "stars", "", "Star count filter (e.g. >100, 10..50)")
	cmd.Flags().StringVar(&opts.Sort, "sort", "best-match", "Sort field (stars, forks, updated, best-match)")
	cmd.Flags().StringVar(&opts.Order, "order", "desc", "Sort order (asc, desc)")
	cmd.Flags().StringVar(&opts.Format, "format", "table", "Output format (table, json, yaml, csv)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 30, "Maximum results per page")
	cmd.Flags().IntVar(&opts.Page, "page", 1, "Result page number")

	_ = cmd.MarkFlagRequired("provider")
	_ = cmd.MarkFlagRequired("query")

	return cmd
}

func runRepoSearch(ctx context.Context, opts *SearchOptions) error {
	if opts == nil {
		return fmt.Errorf("search options are required")
	}

	query := strings.TrimSpace(opts.Query)
	if query == "" {
		return fmt.Errorf("--query is required")
	}

	providerType := strings.ToLower(strings.TrimSpace(opts.Provider))
	if providerType == "" {
		return fmt.Errorf("--provider is required")
	}

	switch providerType {
	case "github", "gitlab", "gitea":
		// supported
	default:
		return fmt.Errorf("unsupported provider %q (supported: github, gitlab, gitea)", opts.Provider)
	}

	sort := opts.Sort
	if sort == "best-match" {
		sort = ""
	}

	ownerHint := opts.Org
	if ownerHint == "" {
		ownerHint = opts.User
	}

	gitProvider, err := getGitProvider(providerType, ownerHint)
	if err != nil {
		return fmt.Errorf("failed to create %s provider: %w", providerType, err)
	}

	searchQuery := provider.SearchQuery{
		Query:        query,
		Sort:         sort,
		Order:        opts.Order,
		Language:     opts.Language,
		User:         opts.User,
		Organization: opts.Org,
		Topic:        opts.Topic,
		Stars:        opts.Stars,
		Page:         opts.Page,
		PerPage:      opts.Limit,
	}

	result, err := gitProvider.SearchRepositories(ctx, searchQuery)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if result == nil {
		return fmt.Errorf("search returned no result")
	}

	listOpts := &ListOptions{
		Format: opts.Format,
		Quiet:  false,
	}

	return listOpts.outputRepositories(result.Repositories)
}
