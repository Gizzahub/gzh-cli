package repoconfig

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// newListCmd creates the list subcommand
func newListCmd() *cobra.Command {
	var flags GlobalFlags
	var (
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

// runListCommand executes the list command
func runListCommand(flags GlobalFlags, filter, format string, showConfig bool, limit int) error {
	if flags.Organization == "" {
		return fmt.Errorf("organization is required (use --org flag)")
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

	// TODO: Implement actual repository listing logic
	fmt.Printf("📋 Repository Configuration Status for %s\n", flags.Organization)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Mock data for demonstration
	repositories := []RepositoryInfo{
		{
			Name:        "api-server",
			Description: "Main API server",
			Visibility:  "private",
			Template:    "microservice",
			Compliant:   true,
			Issues:      0,
		},
		{
			Name:        "web-frontend",
			Description: "React frontend application",
			Visibility:  "private",
			Template:    "frontend",
			Compliant:   true,
			Issues:      0,
		},
		{
			Name:        "legacy-service",
			Description: "Legacy service (needs migration)",
			Visibility:  "private",
			Template:    "none",
			Compliant:   false,
			Issues:      3,
		},
	}

	switch format {
	case "table":
		printTableFormat(repositories, showConfig)
	case "json":
		printJSONFormat(repositories)
	case "yaml":
		printYAMLFormat(repositories)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}

// RepositoryInfo represents repository information
type RepositoryInfo struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Visibility  string `json:"visibility" yaml:"visibility"`
	Template    string `json:"template" yaml:"template"`
	Compliant   bool   `json:"compliant" yaml:"compliant"`
	Issues      int    `json:"issues" yaml:"issues"`
}

// printTableFormat prints repositories in table format
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

// printJSONFormat prints repositories in JSON format
func printJSONFormat(repos []RepositoryInfo) {
	data, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

// printYAMLFormat prints repositories in YAML format
func printYAMLFormat(repos []RepositoryInfo) {
	data, err := yaml.Marshal(repos)
	if err != nil {
		fmt.Printf("Error marshaling YAML: %v\n", err)
		return
	}
	fmt.Println(string(data))
}
