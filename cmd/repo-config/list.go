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

	// Get GitHub token
	environment := env.NewOSEnvironment()

	token := flags.Token
	if token == "" {
		token = environment.Get(env.CommonEnvironmentKeys.GitHubToken)
	}

	if token == "" {
		return fmt.Errorf("GitHub token is required (use --token flag or GITHUB_TOKEN env var)")
	}

	if flags.Verbose {
		fmt.Printf("Listing repositories for organization: %s\n", flags.Organization)

		if filter != "" {
			fmt.Printf("Filter pattern: %s\n", filter)
		}

		if limit > 0 {
			fmt.Printf("Limit: %d\n", limit)
		}

		fmt.Println()
	}

	// Create GitHub client
	client := github.NewRepoConfigClient(token)
	ctx := context.Background()

	fmt.Printf("📋 Repository Configuration Status for %s\n", flags.Organization)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// List repositories
	repos, err := client.ListRepositories(ctx, flags.Organization, nil)
	if err != nil {
		return fmt.Errorf("failed to list repositories: %w", err)
	}

	// Filter repositories if pattern provided
	if filter != "" {
		filterRegex, err := regexp.Compile(filter)
		if err != nil {
			return fmt.Errorf("invalid filter pattern: %w", err)
		}

		var filtered []*github.Repository

		for _, repo := range repos {
			if filterRegex.MatchString(repo.Name) {
				filtered = append(filtered, repo)
			}
		}

		repos = filtered
	}

	// Apply limit if specified
	if limit > 0 && len(repos) > limit {
		repos = repos[:limit]
	}

	// Load repository configuration to check compliance
	var repoConfig *config.RepoConfig
	if flags.ConfigFile != "" {
		repoConfig, err = config.LoadRepoConfig(flags.ConfigFile)
		if err != nil && flags.Verbose {
			fmt.Printf("Warning: Could not load repo config: %v\n", err)
		}
	}

	// Convert to RepositoryInfo format
	repositories := make([]RepositoryInfo, 0, len(repos))

	for _, repo := range repos {
		visibility := visibilityPublic
		if repo.Private {
			visibility = visibilityPrivate
		}

		info := RepositoryInfo{
			Name:        repo.Name,
			Description: repo.Description,
			Visibility:  visibility,
			Template:    detectTemplate(repo, repoConfig),
			Compliant:   checkCompliance(repo, repoConfig),
			Issues:      0, // Could be calculated based on actual compliance checks
		}

		if showConfig {
			// Get detailed configuration
			config, err := client.GetRepositoryConfiguration(ctx, flags.Organization, repo.Name)
			if err == nil {
				info.Config = config
			} else if flags.Verbose {
				fmt.Printf("Warning: Could not get config for %s: %v\n", repo.Name, err)
			}
		}

		repositories = append(repositories, info)
	}

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
