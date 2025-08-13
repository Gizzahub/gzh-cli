// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repoconfig

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-manager-go/internal/cli"
	"github.com/Gizzahub/gzh-manager-go/internal/services"
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
		showConfig bool
	)

	builder := cli.NewCommandBuilder(context.Background(), "list", "List repositories with current configuration").
		WithLongDescription(`List repositories in the organization with their current configuration status.

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
  gz repo-config list --limit 50                    # Limit results`).
		WithOrganizationFlag(true).
		WithTokenFlag().
		WithConfigFileFlag().
		WithVerboseFlag().
		WithFormatFlag("table", []string{"table", "json", "yaml"}).
		WithFilterFlag().
		WithLimitFlag(0).
		WithCustomBoolFlag("show-config", false, "Include detailed configuration", &showConfig).
		WithRunFuncE(func(ctx context.Context, commonFlags *cli.CommonFlags, args []string) error {
			// Map common flags to GlobalFlags
			flags.Organization = commonFlags.Organization
			flags.Token = commonFlags.Token
			flags.ConfigFile = commonFlags.ConfigFile
			flags.Verbose = commonFlags.Verbose

			return runListCommand(flags, commonFlags.Filter, commonFlags.Format, showConfig, commonFlags.Limit)
		})

	return builder.Build()
}

// runListCommand executes the list command.
func runListCommand(flags GlobalFlags, filter, format string, showConfig bool, limit int) error {
	if flags.Organization == "" {
		return fmt.Errorf("organization is required (use --org flag)")
	}

	// Display header
	displayListHeader(flags.Organization)

	// Create service and prepare options
	service := services.NewRepoConfigService()
	opts := services.ListOptions{
		Organization: flags.Organization,
		Filter:       filter,
		Format:       format,
		ShowConfig:   showConfig,
		Limit:        limit,
		Token:        flags.Token,
		ConfigFile:   flags.ConfigFile,
		Verbose:      flags.Verbose,
	}

	// Get repository list from service
	ctx := context.Background()
	repositories, err := service.ListRepositories(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list repositories: %w", err)
	}

	// Display results
	return displayRepositoryList(repositories, format, showConfig)
}

// displayListHeader displays the list command header.
func displayListHeader(organization string) {
	fmt.Printf("📋 Repository Configuration Status for %s\n", organization)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}

// printTableFormat prints repositories in table format.
func printTableFormat(repos []services.RepositoryInfo, showConfig bool) {
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

// displayRepositoryList displays the repository list in the specified format.
func displayRepositoryList(repositories []services.RepositoryInfo, format string, showConfig bool) error {
	formatter := cli.NewOutputFormatter(format)

	switch format {
	case "table":
		printTableFormat(repositories, showConfig)
	case "json", "yaml":
		return formatter.FormatOutput(repositories)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}
