package repoconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/gizzahub/gzh-manager-go/pkg/config"
	"github.com/gizzahub/gzh-manager-go/pkg/github"
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

	fmt.Printf("📊 Repository Configuration Comparison\n")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Organization: %s\n", flags.Organization)
	fmt.Println()

	// Get configuration differences
	differences, err := getConfigurationDifferences(flags.Organization, filter)
	if err != nil {
		return fmt.Errorf("failed to get configuration differences: %w", err)
	}

	// If no differences found, return early
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
	jsonData := map[string]interface{}{
		"differences": differences,
		"summary": map[string]interface{}{
			"total_changes":  len(differences),
			"affected_repos": len(getAffectedRepositories(differences)),
		},
	}

	if jsonBytes, err := json.MarshalIndent(jsonData, "", "  "); err != nil {
		fmt.Printf("Error serializing JSON: %v\n", err)
	} else {
		fmt.Println(string(jsonBytes))
	}
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

// compareRepositoryConfigurations compares current and target configurations
func compareRepositoryConfigurations(
	repoName string,
	current *github.RepositoryConfig,
	targetSettings *config.RepoSettings,
	targetSecurity *config.SecuritySettings,
	targetPermissions *config.PermissionSettings,
	templateName string,
	exceptions []config.PolicyException,
) []ConfigurationDifference {
	var differences []ConfigurationDifference

	// Compare basic settings
	if targetSettings != nil {
		// Description
		if targetSettings.Description != nil && current.Description != *targetSettings.Description {
			differences = append(differences, ConfigurationDifference{
				Repository:   repoName,
				Setting:      "description",
				CurrentValue: current.Description,
				TargetValue:  *targetSettings.Description,
				ChangeType:   getChangeType(current.Description, *targetSettings.Description),
				Impact:       "low",
				Template:     templateName,
				Compliant:    false,
			})
		}

		// Homepage
		if targetSettings.Homepage != nil && current.Homepage != *targetSettings.Homepage {
			differences = append(differences, ConfigurationDifference{
				Repository:   repoName,
				Setting:      "homepage",
				CurrentValue: current.Homepage,
				TargetValue:  *targetSettings.Homepage,
				ChangeType:   getChangeType(current.Homepage, *targetSettings.Homepage),
				Impact:       "low",
				Template:     templateName,
				Compliant:    false,
			})
		}

		// Private/Visibility
		if targetSettings.Private != nil && current.Private != *targetSettings.Private {
			visibility := "public"
			targetVisibility := "public"
			if current.Private {
				visibility = "private"
			}
			if *targetSettings.Private {
				targetVisibility = "private"
			}
			differences = append(differences, ConfigurationDifference{
				Repository:   repoName,
				Setting:      "visibility",
				CurrentValue: visibility,
				TargetValue:  targetVisibility,
				ChangeType:   "update",
				Impact:       "high",
				Template:     templateName,
				Compliant:    false,
			})
		}

		// Features
		if targetSettings.HasIssues != nil && current.Settings.HasIssues != *targetSettings.HasIssues {
			differences = append(differences, ConfigurationDifference{
				Repository:   repoName,
				Setting:      "features.issues",
				CurrentValue: fmt.Sprintf("%t", current.Settings.HasIssues),
				TargetValue:  fmt.Sprintf("%t", *targetSettings.HasIssues),
				ChangeType:   "update",
				Impact:       "low",
				Template:     templateName,
				Compliant:    false,
			})
		}

		if targetSettings.HasWiki != nil && current.Settings.HasWiki != *targetSettings.HasWiki {
			differences = append(differences, ConfigurationDifference{
				Repository:   repoName,
				Setting:      "features.wiki",
				CurrentValue: fmt.Sprintf("%t", current.Settings.HasWiki),
				TargetValue:  fmt.Sprintf("%t", *targetSettings.HasWiki),
				ChangeType:   "update",
				Impact:       "low",
				Template:     templateName,
				Compliant:    false,
			})
		}

		if targetSettings.HasProjects != nil && current.Settings.HasProjects != *targetSettings.HasProjects {
			differences = append(differences, ConfigurationDifference{
				Repository:   repoName,
				Setting:      "features.projects",
				CurrentValue: fmt.Sprintf("%t", current.Settings.HasProjects),
				TargetValue:  fmt.Sprintf("%t", *targetSettings.HasProjects),
				ChangeType:   "update",
				Impact:       "low",
				Template:     templateName,
				Compliant:    false,
			})
		}

		// Merge settings
		if targetSettings.DeleteBranchOnMerge != nil && current.Settings.DeleteBranchOnMerge != *targetSettings.DeleteBranchOnMerge {
			differences = append(differences, ConfigurationDifference{
				Repository:   repoName,
				Setting:      "merge.delete_branch_on_merge",
				CurrentValue: fmt.Sprintf("%t", current.Settings.DeleteBranchOnMerge),
				TargetValue:  fmt.Sprintf("%t", *targetSettings.DeleteBranchOnMerge),
				ChangeType:   "update",
				Impact:       "medium",
				Template:     templateName,
				Compliant:    false,
			})
		}

		if targetSettings.AllowSquashMerge != nil && current.Settings.AllowSquashMerge != *targetSettings.AllowSquashMerge {
			differences = append(differences, ConfigurationDifference{
				Repository:   repoName,
				Setting:      "merge.allow_squash_merge",
				CurrentValue: fmt.Sprintf("%t", current.Settings.AllowSquashMerge),
				TargetValue:  fmt.Sprintf("%t", *targetSettings.AllowSquashMerge),
				ChangeType:   "update",
				Impact:       "medium",
				Template:     templateName,
				Compliant:    false,
			})
		}
	}

	// Compare security settings
	if targetSecurity != nil {
		// Branch protection
		if targetSecurity.BranchProtection != nil {
			for branch, targetRule := range targetSecurity.BranchProtection {
				currentRule, exists := current.BranchProtection[branch]
				if !exists {
					// Branch protection doesn't exist
					if targetRule.RequiredReviews != nil && *targetRule.RequiredReviews > 0 {
						differences = append(differences, ConfigurationDifference{
							Repository:   repoName,
							Setting:      fmt.Sprintf("branch_protection.%s.required_reviews", branch),
							CurrentValue: "0",
							TargetValue:  fmt.Sprintf("%d", *targetRule.RequiredReviews),
							ChangeType:   "create",
							Impact:       "high",
							Template:     templateName,
							Compliant:    false,
						})
					}
				} else {
					// Compare existing branch protection
					if targetRule.RequiredReviews != nil && currentRule.RequiredReviews != *targetRule.RequiredReviews {
						differences = append(differences, ConfigurationDifference{
							Repository:   repoName,
							Setting:      fmt.Sprintf("branch_protection.%s.required_reviews", branch),
							CurrentValue: fmt.Sprintf("%d", currentRule.RequiredReviews),
							TargetValue:  fmt.Sprintf("%d", *targetRule.RequiredReviews),
							ChangeType:   "update",
							Impact:       "medium",
							Template:     templateName,
							Compliant:    false,
						})
					}

					if targetRule.EnforceAdmins != nil && currentRule.EnforceAdmins != *targetRule.EnforceAdmins {
						differences = append(differences, ConfigurationDifference{
							Repository:   repoName,
							Setting:      fmt.Sprintf("branch_protection.%s.enforce_admins", branch),
							CurrentValue: fmt.Sprintf("%t", currentRule.EnforceAdmins),
							TargetValue:  fmt.Sprintf("%t", *targetRule.EnforceAdmins),
							ChangeType:   "update",
							Impact:       "high",
							Template:     templateName,
							Compliant:    false,
						})
					}
				}
			}
		}
	}

	// Compare permissions
	if targetPermissions != nil {
		// Team permissions
		for team, targetPerm := range targetPermissions.TeamPermissions {
			currentPerm, exists := current.Permissions.Teams[team]
			if !exists {
				differences = append(differences, ConfigurationDifference{
					Repository:   repoName,
					Setting:      fmt.Sprintf("permissions.team.%s", team),
					CurrentValue: "none",
					TargetValue:  targetPerm,
					ChangeType:   "create",
					Impact:       "medium",
					Template:     templateName,
					Compliant:    false,
				})
			} else if currentPerm != targetPerm {
				differences = append(differences, ConfigurationDifference{
					Repository:   repoName,
					Setting:      fmt.Sprintf("permissions.team.%s", team),
					CurrentValue: currentPerm,
					TargetValue:  targetPerm,
					ChangeType:   "update",
					Impact:       "medium",
					Template:     templateName,
					Compliant:    false,
				})
			}
		}
	}

	// Apply policy exceptions to differences
	differences = applyPolicyExceptions(differences, exceptions)

	return differences
}

// getChangeType determines the type of change
func getChangeType(current, target string) string {
	if current == "" && target != "" {
		return "create"
	}
	if current != "" && target == "" {
		return "delete"
	}
	return "update"
}

// findAppliedTemplate finds which template applies to a repository
func findAppliedTemplate(repoConfig *config.RepoConfig, repoName string) string {
	if repoConfig.Repositories != nil {
		// Check specific repositories
		for _, specific := range repoConfig.Repositories.Specific {
			if specific.Name == repoName && specific.Template != "" {
				return specific.Template
			}
		}

		// Check patterns
		for _, pattern := range repoConfig.Repositories.Patterns {
			if matched, _ := matchRepoPattern(repoName, pattern.Match); matched && pattern.Template != "" {
				return pattern.Template
			}
		}

		// Check default
		if repoConfig.Repositories.Default != nil && repoConfig.Repositories.Default.Template != "" {
			return repoConfig.Repositories.Default.Template
		}
	}

	// Check global defaults
	if repoConfig.Defaults != nil && repoConfig.Defaults.Template != "" {
		return repoConfig.Defaults.Template
	}

	return "none"
}

// matchRepoPattern checks if a repository name matches a pattern
func matchRepoPattern(name, pattern string) (bool, error) {
	// Convert simple glob patterns to regex
	if strings.Contains(pattern, "*") {
		pattern = strings.ReplaceAll(pattern, ".", "\\.")
		pattern = strings.ReplaceAll(pattern, "*", ".*")
		pattern = "^" + pattern + "$"
		return regexp.MatchString(pattern, name)
	}
	return name == pattern, nil
}

// applyPolicyExceptions applies policy exceptions to differences
func applyPolicyExceptions(differences []ConfigurationDifference, exceptions []config.PolicyException) []ConfigurationDifference {
	// For now, just return differences as-is
	// In a full implementation, this would check if any differences are covered by exceptions
	// and mark them as compliant or reduce their impact level
	return differences
}

// getGitHubToken retrieves the GitHub token from various sources
func getGitHubToken() string {
	// This should be implemented to get token from:
	// 1. Command line flags (if available)
	// 2. Environment variables
	// 3. Config file
	return os.Getenv("GITHUB_TOKEN")
}

// getConfigPath retrieves the configuration file path
func getConfigPath() string {
	// This should be implemented to get config path from:
	// 1. Command line flags (if available)
	// 2. Default locations
	// For now, return a default
	return "repo-config.yaml"
}

// getConfigurationDifferences retrieves configuration differences for an organization
func getConfigurationDifferences(organization, filter string) ([]ConfigurationDifference, error) {
	// Create a context
	ctx := context.Background()

	// Get GitHub token from environment or global flags
	token := getGitHubToken()
	if token == "" {
		return nil, fmt.Errorf("GitHub token not found. Set GITHUB_TOKEN environment variable or use --token flag")
	}

	// Create GitHub client
	client := github.NewRepoConfigClient(token)

	// Load repo config file
	configPath := getConfigPath()
	if configPath == "" {
		return nil, fmt.Errorf("configuration file not found. Use --config flag or create a repo-config.yaml file")
	}

	repoConfig, err := config.LoadRepoConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load repo config: %w", err)
	}

	// Validate that the organization matches
	if repoConfig.Organization != organization {
		return nil, fmt.Errorf("organization mismatch: config file is for '%s', but diff requested for '%s'", repoConfig.Organization, organization)
	}

	// List all repositories in the organization
	listOpts := &github.ListOptions{
		PerPage: 100,
	}
	repos, err := client.ListRepositories(ctx, organization, listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	// Apply filter if specified
	var filteredRepos []*github.Repository
	if filter != "" {
		filterRegex, err := regexp.Compile(filter)
		if err != nil {
			return nil, fmt.Errorf("invalid filter regex: %w", err)
		}
		for _, repo := range repos {
			if filterRegex.MatchString(repo.Name) {
				filteredRepos = append(filteredRepos, repo)
			}
		}
	} else {
		filteredRepos = repos
	}

	// Compare configurations
	var differences []ConfigurationDifference
	for _, repo := range filteredRepos {
		// Skip archived repositories
		if repo.Archived {
			continue
		}

		// Get current configuration from GitHub
		currentConfig, err := client.GetRepositoryConfiguration(ctx, organization, repo.Name)
		if err != nil {
			// Log error but continue with other repos
			fmt.Printf("Warning: Failed to get configuration for %s: %v\n", repo.Name, err)
			continue
		}

		// Get target configuration from repo config
		targetSettings, targetSecurity, targetPermissions, exceptions, err := repoConfig.GetEffectiveConfig(repo.Name)
		if err != nil {
			fmt.Printf("Warning: Failed to get target configuration for %s: %v\n", repo.Name, err)
			continue
		}

		// Find which template applies to this repository
		templateName := findAppliedTemplate(repoConfig, repo.Name)

		// Compare settings
		repoDiffs := compareRepositoryConfigurations(repo.Name, currentConfig, targetSettings, targetSecurity, targetPermissions, templateName, exceptions)
		differences = append(differences, repoDiffs...)
	}

	return differences, nil
}
