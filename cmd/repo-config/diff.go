package repoconfig

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newDiffCmd creates the diff subcommand
func newDiffCmd() *cobra.Command {
	var flags GlobalFlags
	var (
		filter     string
		format     string
		showValues bool
	)

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Show differences between current and target configuration",
		Long: `Show differences between current repository configuration and target configuration.

This command compares the current state of repositories with the desired
configuration defined in templates and configuration files. It helps
identify what changes would be made before applying them.

Comparison Features:
- Side-by-side configuration comparison
- Change detection and categorization
- Impact analysis for proposed changes
- Template compliance checking

Output Formats:
- table: Human-readable diff table (default)
- json: JSON format for programmatic use
- unified: Unified diff format

Examples:
  gz repo-config diff --org myorg                # Show all differences
  gz repo-config diff --filter "^api-.*"        # Filter by repository pattern
  gz repo-config diff --format unified          # Unified diff format
  gz repo-config diff --show-values             # Include current values`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiffCommand(flags, filter, format, showValues)
		},
	}

	// Add global flags
	addGlobalFlags(cmd, &flags)

	// Add diff-specific flags
	cmd.Flags().StringVar(&filter, "filter", "", "Filter repositories by name pattern (regex)")
	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json, unified)")
	cmd.Flags().BoolVar(&showValues, "show-values", false, "Include current values in output")

	return cmd
}

// runDiffCommand executes the diff command
func runDiffCommand(flags GlobalFlags, filter, format string, showValues bool) error {
	if flags.Organization == "" {
		return fmt.Errorf("organization is required (use --org flag)")
	}

	if flags.Verbose {
		fmt.Printf("🔍 Comparing repository configurations for organization: %s\n", flags.Organization)
		if filter != "" {
			fmt.Printf("Filter pattern: %s\n", filter)
		}
		fmt.Printf("Format: %s\n", format)
		fmt.Println()
	}

	// TODO: Implement actual configuration comparison logic
	fmt.Printf("📊 Repository Configuration Comparison\n")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Organization: %s\n", flags.Organization)
	fmt.Println()

	// Mock differences for demonstration
	differences := []ConfigurationDifference{
		{
			Repository:   "api-server",
			Setting:      "branch_protection.main.required_reviews",
			CurrentValue: "1",
			TargetValue:  "2",
			ChangeType:   "update",
			Impact:       "medium",
			Template:     "microservice",
			Compliant:    false,
		},
		{
			Repository:   "web-frontend",
			Setting:      "features.wiki",
			CurrentValue: "true",
			TargetValue:  "false",
			ChangeType:   "update",
			Impact:       "low",
			Template:     "frontend",
			Compliant:    false,
		},
		{
			Repository:   "legacy-service",
			Setting:      "security.delete_head_branches",
			CurrentValue: "",
			TargetValue:  "true",
			ChangeType:   "create",
			Impact:       "high",
			Template:     "none",
			Compliant:    false,
		},
		{
			Repository:   "admin-tools",
			Setting:      "visibility",
			CurrentValue: "public",
			TargetValue:  "private",
			ChangeType:   "update",
			Impact:       "high",
			Template:     "security",
			Compliant:    false,
		},
	}

	if len(differences) == 0 {
		fmt.Println("✅ No configuration differences found - all repositories are compliant")
		return nil
	}

	switch format {
	case "table":
		displayDiffTable(differences, showValues)
	case "json":
		displayDiffJSON(differences)
	case "unified":
		displayDiffUnified(differences)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Summary
	fmt.Println()
	displayDiffSummary(differences)

	return nil
}

// ConfigurationDifference represents a difference between current and target config
type ConfigurationDifference struct {
	Repository   string `json:"repository"`
	Setting      string `json:"setting"`
	CurrentValue string `json:"current_value"`
	TargetValue  string `json:"target_value"`
	ChangeType   string `json:"change_type"` // create, update, delete
	Impact       string `json:"impact"`      // low, medium, high
	Template     string `json:"template"`
	Compliant    bool   `json:"compliant"`
}

// displayDiffTable displays differences in table format
func displayDiffTable(differences []ConfigurationDifference, showValues bool) {
	if showValues {
		fmt.Printf("%-20s %-30s %-15s %-15s %-10s %s\n",
			"REPOSITORY", "SETTING", "CURRENT", "TARGET", "IMPACT", "ACTION")
	} else {
		fmt.Printf("%-20s %-30s %-10s %-10s %s\n",
			"REPOSITORY", "SETTING", "IMPACT", "ACTION", "TEMPLATE")
	}
	fmt.Println("────────────────────────────────────────────────────────────────────────────────")

	for _, diff := range differences {
		actionSymbol := getActionSymbol(diff.ChangeType)
		impactSymbol := getImpactSymbol(diff.Impact)

		if showValues {
			currentDisplay := truncateString(diff.CurrentValue, 15)
			if currentDisplay == "" {
				currentDisplay = "-"
			}

			fmt.Printf("%-20s %-30s %-15s %-15s %-10s %s\n",
				diff.Repository,
				truncateString(diff.Setting, 30),
				currentDisplay,
				truncateString(diff.TargetValue, 15),
				impactSymbol,
				actionSymbol,
			)
		} else {
			fmt.Printf("%-20s %-30s %-10s %-10s %s\n",
				diff.Repository,
				truncateString(diff.Setting, 30),
				impactSymbol,
				actionSymbol,
				diff.Template,
			)
		}
	}
}

// displayDiffJSON displays differences in JSON format
func displayDiffJSON(differences []ConfigurationDifference) {
	// TODO: Implement proper JSON serialization
	fmt.Println("JSON diff output not yet implemented")
}

// displayDiffUnified displays differences in unified diff format
func displayDiffUnified(differences []ConfigurationDifference) {
	for _, diff := range differences {
		fmt.Printf("--- %s (current)\n", diff.Repository)
		fmt.Printf("+++ %s (target)\n", diff.Repository)
		fmt.Printf("@@ %s @@\n", diff.Setting)

		switch diff.ChangeType {
		case "create":
			fmt.Printf("+%s: %s\n", diff.Setting, diff.TargetValue)
		case "update":
			fmt.Printf("-%s: %s\n", diff.Setting, diff.CurrentValue)
			fmt.Printf("+%s: %s\n", diff.Setting, diff.TargetValue)
		case "delete":
			fmt.Printf("-%s: %s\n", diff.Setting, diff.CurrentValue)
		}
		fmt.Println()
	}
}

// displayDiffSummary displays a summary of differences
func displayDiffSummary(differences []ConfigurationDifference) {
	repoCount := len(getAffectedRepositories(differences))

	impactCounts := map[string]int{
		"low":    0,
		"medium": 0,
		"high":   0,
	}

	actionCounts := map[string]int{
		"create": 0,
		"update": 0,
		"delete": 0,
	}

	for _, diff := range differences {
		impactCounts[diff.Impact]++
		actionCounts[diff.ChangeType]++
	}

	fmt.Printf("📊 Summary\n")
	fmt.Printf("Repositories affected: %d\n", repoCount)
	fmt.Printf("Total changes: %d\n", len(differences))
	fmt.Println()

	fmt.Printf("Impact distribution:\n")
	fmt.Printf("  🔴 High: %d\n", impactCounts["high"])
	fmt.Printf("  🟡 Medium: %d\n", impactCounts["medium"])
	fmt.Printf("  🟢 Low: %d\n", impactCounts["low"])
	fmt.Println()

	fmt.Printf("Change types:\n")
	fmt.Printf("  ➕ Create: %d\n", actionCounts["create"])
	fmt.Printf("  🔄 Update: %d\n", actionCounts["update"])
	fmt.Printf("  ➖ Delete: %d\n", actionCounts["delete"])
}

// getImpactSymbol returns the symbol for impact level
func getImpactSymbol(impact string) string {
	switch impact {
	case "high":
		return "🔴 High"
	case "medium":
		return "🟡 Med"
	case "low":
		return "🟢 Low"
	default:
		return "❓ Unknown"
	}
}

// getAffectedRepositories returns unique repository names from differences
func getAffectedRepositories(differences []ConfigurationDifference) []string {
	repos := make(map[string]bool)
	for _, diff := range differences {
		repos[diff.Repository] = true
	}

	var result []string
	for repo := range repos {
		result = append(result, repo)
	}
	return result
}
