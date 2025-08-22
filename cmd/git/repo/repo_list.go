// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repo

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/Gizzahub/gzh-cli/pkg/git/provider"
)

// ListOptions contains options for repository listing.
type ListOptions struct {
	// Provider options
	Provider     string
	AllProviders bool
	Org          string

	// Filtering options
	Visibility   string
	ArchivedOnly bool
	NoArchived   bool
	Match        string
	Language     string
	MinStars     int
	MaxStars     int
	UpdatedSince string

	// Sorting options
	Sort  string
	Order string

	// Output options
	Format  string
	Limit   int
	Quiet   bool
	Verbose bool
}

// newRepoListCmd creates the repo list command.
func newRepoListCmd() *cobra.Command {
	opts := &ListOptions{
		Visibility: "all",
		Sort:       "name",
		Order:      "asc",
		Format:     "table",
	}

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
		Example: `  # List repositories from a GitHub organization
  gz git repo list --provider github --org myorg

  # List with JSON output
  gz git repo list --provider gitlab --org mygroup --format json

  # List from all configured providers
  gz git repo list --all-providers --format table

  # List with advanced filtering
  gz git repo list --provider github --org myorg --language Go --min-stars 100

  # List only archived repositories
  gz git repo list --provider github --org myorg --archived-only

  # List with sorting and limits
  gz git repo list --provider github --org myorg --sort stars --order desc --limit 10`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRepoList(cmd.Context(), opts)
		},
	}

	// Provider options
	cmd.Flags().StringVar(&opts.Provider, "provider", "", "Git provider (github, gitlab, gitea, gogs)")
	cmd.Flags().BoolVar(&opts.AllProviders, "all-providers", false, "List from all configured providers")
	cmd.Flags().StringVar(&opts.Org, "org", "", "Organization/Group name")

	// Filtering options
	cmd.Flags().StringVar(&opts.Visibility, "visibility", "all", "Filter by visibility (public, private, all)")
	cmd.Flags().BoolVar(&opts.ArchivedOnly, "archived-only", false, "Show only archived repositories")
	cmd.Flags().BoolVar(&opts.NoArchived, "no-archived", false, "Exclude archived repositories")
	cmd.Flags().StringVar(&opts.Match, "match", "", "Repository name pattern (regex)")
	cmd.Flags().StringVar(&opts.Language, "language", "", "Filter by primary language")
	cmd.Flags().IntVar(&opts.MinStars, "min-stars", 0, "Minimum star count")
	cmd.Flags().IntVar(&opts.MaxStars, "max-stars", 0, "Maximum star count (0 = no limit)")
	cmd.Flags().StringVar(&opts.UpdatedSince, "updated-since", "", "Filter by last update date (YYYY-MM-DD)")

	// Sorting options
	cmd.Flags().StringVar(&opts.Sort, "sort", "name", "Sort by field (name, created, updated, stars, forks)")
	cmd.Flags().StringVar(&opts.Order, "order", "asc", "Sort order (asc, desc)")

	// Output options
	cmd.Flags().StringVar(&opts.Format, "format", "table", "Output format (table, json, yaml, csv)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 0, "Limit number of results (0 = no limit)")
	cmd.Flags().BoolVar(&opts.Quiet, "quiet", false, "Suppress headers and extra output")
	cmd.Flags().BoolVar(&opts.Verbose, "verbose", false, "Include additional repository details")

	// Validation rules
	cmd.MarkFlagsMutuallyExclusive("archived-only", "no-archived")
	cmd.MarkFlagsMutuallyExclusive("provider", "all-providers")

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if !opts.AllProviders && opts.Provider == "" {
			return fmt.Errorf("either --provider or --all-providers must be specified")
		}
		if !opts.AllProviders && opts.Org == "" {
			return fmt.Errorf("--org is required when using --provider")
		}
		return nil
	}

	return cmd
}

// runRepoList executes the repository listing operation.
func runRepoList(ctx context.Context, opts *ListOptions) error {
	// Validate options
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	var allRepos []provider.Repository

	if opts.AllProviders {
		// Get repositories from all configured providers
		repos, err := opts.listFromAllProviders(ctx)
		if err != nil {
			return fmt.Errorf("failed to list from all providers: %w", err)
		}
		allRepos = repos
	} else {
		// Get repositories from single provider
		repos, err := opts.listFromProvider(ctx, opts.Provider, opts.Org)
		if err != nil {
			return fmt.Errorf("failed to list repositories: %w", err)
		}
		allRepos = repos
	}

	// Apply filtering
	filtered := opts.applyFilters(allRepos)

	// Apply sorting
	sorted := opts.applySorting(filtered)

	// Apply limit
	if opts.Limit > 0 && len(sorted) > opts.Limit {
		sorted = sorted[:opts.Limit]
	}

	// Output results
	return opts.outputRepositories(sorted)
}

// Validate validates the list options.
func (opts *ListOptions) Validate() error {
	// Validate visibility
	validVisibility := []string{"all", "public", "private"}
	if !contains(validVisibility, opts.Visibility) {
		return fmt.Errorf("invalid visibility: %s (valid: %s)", opts.Visibility, strings.Join(validVisibility, ", "))
	}

	// Validate match pattern if provided
	if opts.Match != "" {
		if _, err := regexp.Compile(opts.Match); err != nil {
			return fmt.Errorf("invalid match pattern: %w", err)
		}
	}

	// Validate sort field
	validSortFields := []string{"name", "created", "updated", "stars", "forks"}
	if !contains(validSortFields, opts.Sort) {
		return fmt.Errorf("invalid sort field: %s (valid: %s)", opts.Sort, strings.Join(validSortFields, ", "))
	}

	// Validate sort order
	validOrders := []string{"asc", "desc"}
	if !contains(validOrders, opts.Order) {
		return fmt.Errorf("invalid sort order: %s (valid: %s)", opts.Order, strings.Join(validOrders, ", "))
	}

	// Validate output format
	if !isValidOutputFormat(opts.Format) {
		return fmt.Errorf("invalid output format: %s", opts.Format)
	}

	// Validate star range
	if opts.MinStars < 0 {
		return fmt.Errorf("min-stars cannot be negative")
	}
	if opts.MaxStars > 0 && opts.MaxStars < opts.MinStars {
		return fmt.Errorf("max-stars must be greater than min-stars")
	}

	return nil
}

// listFromAllProviders gets repositories from all configured providers.
func (opts *ListOptions) listFromAllProviders(ctx context.Context) ([]provider.Repository, error) {
	// TODO: Implement listing from all providers
	// This would require:
	// 1. Getting all configured providers from config
	// 2. Listing repositories from each provider
	// 3. Aggregating results
	return nil, fmt.Errorf("listing from all providers not implemented yet")
}

// listFromProvider gets repositories from a single provider.
func (opts *ListOptions) listFromProvider(ctx context.Context, providerType, org string) ([]provider.Repository, error) {
	// Get provider
	gitProvider, err := getGitProvider(providerType, org)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Convert visibility
	var visibility provider.VisibilityType
	switch opts.Visibility {
	case "public":
		visibility = provider.VisibilityPublic
	case "private":
		visibility = provider.VisibilityPrivate
	default:
		visibility = "" // All
	}

	// Build list options
	listOpts := provider.ListOptions{
		Organization: org,
		Visibility:   visibility,
		Type:         "all",
		Sort:         opts.Sort,
		Direction:    opts.Order,
		PerPage:      100,
	}

	// Set archived filter
	if opts.ArchivedOnly {
		archived := true
		listOpts.Archived = &archived
	} else if opts.NoArchived {
		archived := false
		listOpts.Archived = &archived
	}

	// Language filter
	if opts.Language != "" {
		listOpts.Language = opts.Language
	}

	// Get repositories
	repoList, err := gitProvider.ListRepositories(ctx, listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	return repoList.Repositories, nil
}

// applyFilters applies client-side filtering to repositories.
func (opts *ListOptions) applyFilters(repos []provider.Repository) []provider.Repository {
	var filtered []provider.Repository

	for _, repo := range repos {
		// Name pattern filter
		if opts.Match != "" {
			pattern, err := regexp.Compile(opts.Match)
			if err == nil && !pattern.MatchString(repo.Name) {
				continue
			}
		}

		// Stars filter
		if opts.MinStars > 0 && repo.Stars < opts.MinStars {
			continue
		}
		if opts.MaxStars > 0 && repo.Stars > opts.MaxStars {
			continue
		}

		// TODO: Add updated-since filter
		// This would require parsing the date and comparing with repo.UpdatedAt

		filtered = append(filtered, repo)
	}

	return filtered
}

// applySorting sorts repositories according to options.
func (opts *ListOptions) applySorting(repos []provider.Repository) []provider.Repository {
	sorted := make([]provider.Repository, len(repos))
	copy(sorted, repos)

	sort.Slice(sorted, func(i, j int) bool {
		var less bool

		switch opts.Sort {
		case "name":
			less = sorted[i].Name < sorted[j].Name
		case "created":
			less = sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
		case "updated":
			less = sorted[i].UpdatedAt.Before(sorted[j].UpdatedAt)
		case "stars":
			less = sorted[i].Stars < sorted[j].Stars
		case "forks":
			less = sorted[i].Forks < sorted[j].Forks
		default:
			less = sorted[i].Name < sorted[j].Name
		}

		if opts.Order == "desc" {
			return !less
		}
		return less
	})

	return sorted
}

// outputRepositories outputs repositories in the specified format.
func (opts *ListOptions) outputRepositories(repos []provider.Repository) error {
	switch opts.Format {
	case "table":
		return opts.outputTable(repos)
	case "json":
		return opts.outputJSON(repos)
	case "yaml":
		return opts.outputYAML(repos)
	case "csv":
		return opts.outputCSV(repos)
	default:
		return fmt.Errorf("unsupported output format: %s", opts.Format)
	}
}

// outputTable outputs repositories in table format.
func (opts *ListOptions) outputTable(repos []provider.Repository) error {
	if len(repos) == 0 {
		if !opts.Quiet {
			fmt.Println("No repositories found")
		}
		return nil
	}

	// Header
	if !opts.Quiet {
		if opts.Verbose {
			fmt.Printf("%-40s %-10s %-15s %-10s %-8s %-8s %-12s\n",
				"NAME", "PRIVATE", "LANGUAGE", "STARS", "FORKS", "ISSUES", "UPDATED")
			fmt.Printf("%-40s %-10s %-15s %-10s %-8s %-8s %-12s\n",
				strings.Repeat("-", 40),
				strings.Repeat("-", 10),
				strings.Repeat("-", 15),
				strings.Repeat("-", 10),
				strings.Repeat("-", 8),
				strings.Repeat("-", 8),
				strings.Repeat("-", 12))
		} else {
			fmt.Printf("%-40s %-10s %-15s %-8s %-12s\n",
				"NAME", "PRIVATE", "LANGUAGE", "STARS", "UPDATED")
			fmt.Printf("%-40s %-10s %-15s %-8s %-12s\n",
				strings.Repeat("-", 40),
				strings.Repeat("-", 10),
				strings.Repeat("-", 15),
				strings.Repeat("-", 8),
				strings.Repeat("-", 12))
		}
	}

	// Rows
	for _, repo := range repos {
		private := "public"
		if repo.Private {
			private = "private"
		}

		language := repo.Language
		if language == "" {
			language = "n/a"
		}

		updated := repo.UpdatedAt.Format("2006-01-02")

		if opts.Verbose {
			fmt.Printf("%-40s %-10s %-15s %-10d %-8d %-8d %-12s\n",
				truncateString(repo.FullName, 40),
				private,
				truncateString(language, 15),
				repo.Stars,
				repo.Forks,
				repo.Issues,
				updated)
		} else {
			fmt.Printf("%-40s %-10s %-15s %-8d %-12s\n",
				truncateString(repo.FullName, 40),
				private,
				truncateString(language, 15),
				repo.Stars,
				updated)
		}
	}

	// Summary
	if !opts.Quiet {
		fmt.Printf("\nTotal: %d repositories\n", len(repos))
	}

	return nil
}

// outputJSON outputs repositories in JSON format.
func (opts *ListOptions) outputJSON(repos []provider.Repository) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(repos); err != nil {
		return fmt.Errorf("failed to encode repositories as JSON: %w", err)
	}

	return nil
}

// outputYAML outputs repositories in YAML format.
func (opts *ListOptions) outputYAML(repos []provider.Repository) error {
	yamlData, err := yaml.Marshal(repos)
	if err != nil {
		return fmt.Errorf("failed to marshal repositories as YAML: %w", err)
	}

	fmt.Print(string(yamlData))
	return nil
}

// outputCSV outputs repositories in CSV format.
func (opts *ListOptions) outputCSV(repos []provider.Repository) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write CSV header
	header := []string{"Name", "Full Name", "Default Branch", "Private", "Fork", "Language", "Description", "Stars", "Forks", "Clone URL", "SSH URL", "HTML URL", "Created At", "Updated At"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write repository data
	for _, repo := range repos {
		record := []string{
			repo.Name,
			repo.FullName,
			repo.DefaultBranch,
			fmt.Sprintf("%t", repo.Private),
			fmt.Sprintf("%t", repo.Fork),
			repo.Language,
			repo.Description,
			fmt.Sprintf("%d", repo.Stars),
			fmt.Sprintf("%d", repo.Forks),
			repo.CloneURL,
			repo.SSHURL,
			repo.HTMLURL,
			formatTime(repo.CreatedAt),
			formatTime(repo.UpdatedAt),
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}

// formatTime formats a time.Time for CSV output
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// Helper functions

// contains checks if a slice contains a string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// truncateString truncates a string to the specified length.
func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	if length < 3 {
		return s[:length]
	}
	return s[:length-3] + "..."
}
