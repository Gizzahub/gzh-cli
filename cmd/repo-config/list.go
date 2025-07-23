// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repoconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/gizzahub/gzh-manager-go/internal/env"
	"github.com/gizzahub/gzh-manager-go/pkg/config"
	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	visibilityPublic  = "public"
	visibilityPrivate = "private"
	templateNone      = "none"
)

// newListCmd creates the list subcommand.
func newListCmd() *cobra.Command {
	var (
		flags      GlobalFlags
		filter     string
		format     string
		showConfig bool
		limit      int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List repositories with current configuration",
		Long: `List repositories in the organization with their current configuration status.

This command shows repository details including:
- Basic information (name, description, visibility)
- Current configuration status
- Template compliance
- Security settings summary

Output formats:
- table: Human-readable table format (default)
- json: JSON format for programmatic use
- yaml: YAML format for configuration export

Examples:
  gz repo-config list --org myorg                    # List all repositories
  gz repo-config list --filter "^api-.*"            # Filter by name pattern
  gz repo-config list --format json                 # JSON output
  gz repo-config list --show-config                 # Include configuration details
  gz repo-config list --limit 50                    # Limit results`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListCommand(flags, filter, format, showConfig, limit)
		},
	}

	// Add global flags
	addGlobalFlags(cmd, &flags)

	// Add list-specific flags
	cmd.Flags().StringVar(&filter, "filter", "", "Filter repositories by name pattern (regex)")
	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json, yaml)")
	cmd.Flags().BoolVar(&showConfig, "show-config", false, "Include detailed configuration")
	cmd.Flags().IntVar(&limit, "limit", 0, "Limit number of results (0 = no limit)")

	return cmd
}

// runListCommand executes the list command.
func runListCommand(flags GlobalFlags, filter, format string, showConfig bool, limit int) error { //nolint:gocognit // Complex repository listing logic with multiple formatting options
	if flags.Organization == "" {
		return fmt.Errorf("organization is required (use --org flag)")
	}

	// Setup client and dependencies
	client, repoConfig, err := setupListCommand(flags)
	if err != nil {
		return err
	}

	// List and filter repositories
	repositories, err := listAndProcessRepositories(client, flags, filter, limit, showConfig, repoConfig)
	if err != nil {
		return err
	}

	// Display results
	return displayRepositoryList(repositories, format, showConfig)
}

// RepositoryInfo represents repository information.
type RepositoryInfo struct {
	Name        string                   `json:"name" yaml:"name"`
	Description string                   `json:"description" yaml:"description"`
	Visibility  string                   `json:"visibility" yaml:"visibility"`
	Template    string                   `json:"template" yaml:"template"`
	Compliant   bool                     `json:"compliant" yaml:"compliant"`
	Issues      int                      `json:"issues" yaml:"issues"`
	Config      *github.RepositoryConfig `json:"config,omitempty" yaml:"config,omitempty"`
}

// detectTemplate attempts to detect which template a repository is using.
func detectTemplate(repo *github.Repository, repoConfig *config.RepoConfig) string {
	if repoConfig == nil || repoConfig.Repositories == nil {
		return templateNone
	}

	// Check specific repositories
	for _, specific := range repoConfig.Repositories.Specific {
		if specific.Name == repo.Name && specific.Template != "" {
			return specific.Template
		}
	}

	// Check patterns
	for _, pattern := range repoConfig.Repositories.Patterns {
		if matched, _ := matchPattern(repo.Name, pattern.Match); matched && pattern.Template != "" {
			return pattern.Template
		}
	}

	// Check default
	if repoConfig.Repositories.Default != nil && repoConfig.Repositories.Default.Template != "" {
		return repoConfig.Repositories.Default.Template
	}

	return "none"
}

// checkCompliance checks if a repository is compliant with its template.
func checkCompliance(repo *github.Repository, repoConfig *config.RepoConfig) bool {
	// Simple compliance check - can be expanded
	// For now, just check if it has a template assigned
	template := detectTemplate(repo, repoConfig)
	return template != templateNone
}

// matchPattern checks if a string matches a pattern (simple glob support).
func matchPattern(str, pattern string) (bool, error) {
	if strings.Contains(pattern, "*") {
		// Convert simple glob to regex
		pattern = strings.ReplaceAll(pattern, ".", "\\.")
		pattern = strings.ReplaceAll(pattern, "*", ".*")
		pattern = "^" + pattern + "$"

		return regexp.MatchString(pattern, str)
	}

	return str == pattern, nil
}

// printTableFormat prints repositories in table format.
func printTableFormat(repos []RepositoryInfo, showConfig bool) {
	fmt.Printf("%-20s %-12s %-12s %-10s %s\n", "NAME", "VISIBILITY", "TEMPLATE", "COMPLIANT", "ISSUES")
	fmt.Println("────────────────────────────────────────────────────────────────────")

	for _, repo := range repos {
		compliantSymbol := "✓"
		if !repo.Compliant {
			compliantSymbol = "✗"
		}

		issuesStr := fmt.Sprintf("%d", repo.Issues)
		if repo.Issues == 0 {
			issuesStr = "-"
		}

		fmt.Printf("%-20s %-12s %-12s %-10s %s\n",
			repo.Name,
			repo.Visibility,
			repo.Template,
			compliantSymbol,
			issuesStr,
		)

		if showConfig {
			fmt.Printf("  Description: %s\n", repo.Description)
			fmt.Println()
		}
	}

	fmt.Println()
	fmt.Printf("Total repositories: %d\n", len(repos))

	compliantCount := 0

	for _, repo := range repos {
		if repo.Compliant {
			compliantCount++
		}
	}

	fmt.Printf("Compliant: %d/%d (%.1f%%)\n",
		compliantCount, len(repos),
		float64(compliantCount)/float64(len(repos))*100)
}

// printJSONFormat prints repositories in JSON format.
func printJSONFormat(repos []RepositoryInfo) {
	data, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	fmt.Println(string(data))
}

// printYAMLFormat prints repositories in YAML format.
func printYAMLFormat(repos []RepositoryInfo) {
	data, err := yaml.Marshal(repos) //nolint:musttag // RepositoryInfo struct already has yaml tags
	if err != nil {
		fmt.Printf("Error marshaling YAML: %v\n", err)
		return
	}

	fmt.Println(string(data))
}

// setupListCommand initializes client and configuration for the list command.
func setupListCommand(flags GlobalFlags) (*github.RepoConfigClient, *config.RepoConfig, error) {
	// Get GitHub token
	environment := env.NewOSEnvironment()

	token := flags.Token
	if token == "" {
		token = environment.Get(env.CommonEnvironmentKeys.GitHubToken)
	}

	if token == "" {
		return nil, nil, fmt.Errorf("GitHub token is required (use --token flag or GITHUB_TOKEN env var)")
	}

	if flags.Verbose {
		fmt.Printf("Listing repositories for organization: %s\n", flags.Organization)
		fmt.Println()
	}

	// Create GitHub client
	client := github.NewRepoConfigClient(token)

	// Load repository configuration to check compliance
	var repoConfig *config.RepoConfig
	if flags.ConfigFile != "" {
		var err error
		repoConfig, err = config.LoadRepoConfig(flags.ConfigFile)
		if err != nil && flags.Verbose {
			fmt.Printf("Warning: Could not load repo config: %v\n", err)
		}
	}

	return client, repoConfig, nil
}

// listAndProcessRepositories retrieves and processes repository information.
func listAndProcessRepositories(client *github.RepoConfigClient, flags GlobalFlags, filter string, limit int, showConfig bool, repoConfig *config.RepoConfig) ([]RepositoryInfo, error) {
	ctx := context.Background()

	fmt.Printf("📋 Repository Configuration Status for %s\n", flags.Organization)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// List repositories
	repos, err := client.ListRepositories(ctx, flags.Organization, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	// Apply filters and limits
	repos = applyFiltersAndLimits(repos, filter, limit, flags.Verbose)

	// Convert to RepositoryInfo format
	return convertToRepositoryInfo(repos, client, flags, showConfig, repoConfig), nil
}

// applyFiltersAndLimits applies filtering and limiting to the repository list.
func applyFiltersAndLimits(repos []*github.Repository, filter string, limit int, verbose bool) []*github.Repository {
	// Filter repositories if pattern provided
	if filter != "" {
		filterRegex, err := regexp.Compile(filter)
		if err != nil && verbose {
			fmt.Printf("Warning: Invalid filter pattern: %v\n", err)
		} else {
			var filtered []*github.Repository
			for _, repo := range repos {
				if filterRegex.MatchString(repo.Name) {
					filtered = append(filtered, repo)
				}
			}
			repos = filtered
		}
	}

	// Apply limit if specified
	if limit > 0 && len(repos) > limit {
		repos = repos[:limit]
	}

	return repos
}

// convertToRepositoryInfo converts GitHub repositories to RepositoryInfo format.
func convertToRepositoryInfo(repos []*github.Repository, client *github.RepoConfigClient, flags GlobalFlags, showConfig bool, repoConfig *config.RepoConfig) []RepositoryInfo {
	repositories := make([]RepositoryInfo, 0, len(repos))

	for _, repo := range repos {
		info := createRepositoryInfo(repo, repoConfig)

		if showConfig {
			addDetailedConfiguration(client, flags, &info)
		}

		repositories = append(repositories, info)
	}

	return repositories
}

// createRepositoryInfo creates a RepositoryInfo from a GitHub repository.
func createRepositoryInfo(repo *github.Repository, repoConfig *config.RepoConfig) RepositoryInfo {
	visibility := visibilityPublic
	if repo.Private {
		visibility = visibilityPrivate
	}

	return RepositoryInfo{
		Name:        repo.Name,
		Description: repo.Description,
		Visibility:  visibility,
		Template:    detectTemplate(repo, repoConfig),
		Compliant:   checkCompliance(repo, repoConfig),
		Issues:      0, // Could be calculated based on actual compliance checks
	}
}

// addDetailedConfiguration adds detailed repository configuration if requested.
func addDetailedConfiguration(client *github.RepoConfigClient, flags GlobalFlags, info *RepositoryInfo) {
	ctx := context.Background()
	repoConfig, err := client.GetRepositoryConfiguration(ctx, flags.Organization, info.Name)
	if err == nil {
		info.Config = repoConfig
	} else if flags.Verbose {
		fmt.Printf("Warning: Could not get config for %s: %v\n", info.Name, err)
	}
}

// displayRepositoryList displays the repository list in the specified format.
func displayRepositoryList(repositories []RepositoryInfo, format string, showConfig bool) error {
	switch format {
	case "table":
		printTableFormat(repositories, showConfig)
	case "json":
		printJSONFormat(repositories)
	case formatYAML:
		printYAMLFormat(repositories)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}
