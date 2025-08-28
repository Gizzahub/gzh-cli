// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package diff

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/pkg/config"
	"github.com/Gizzahub/gzh-cli/pkg/github"
)

// Constants for change types.
const (
	changeTypeCreate = "create"
	changeTypeUpdate = "update"
	changeTypeDelete = "delete"
)

// Visibility constants.
const (
	visibilityPublic  = "public"
	visibilityPrivate = "private"
)

// GlobalFlags represents global flags for repo-config commands.
type GlobalFlags struct {
	Organization string
	ConfigFile   string
	Token        string
	DryRun       bool
	Verbose      bool
	Parallel     int
	Timeout      string
}

// addGlobalFlags adds common flags to a command.
func addGlobalFlags(cmd *cobra.Command, flags *GlobalFlags) {
	cmd.Flags().StringVarP(&flags.Organization, "org", "o", "", "GitHub organization name")
	cmd.Flags().StringVarP(&flags.ConfigFile, "config", "c", "", "Configuration file path")
	cmd.Flags().StringVarP(&flags.Token, "token", "t", "", "GitHub personal access token")
	cmd.Flags().BoolVar(&flags.DryRun, "dry-run", false, "Preview changes without applying")
	cmd.Flags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Verbose output")
	cmd.Flags().IntVar(&flags.Parallel, "parallel", 5, "Number of parallel operations")
	cmd.Flags().StringVar(&flags.Timeout, "timeout", "30s", "API timeout duration")
}

// getActionSymbolWithText returns the symbol with text for action type.
func getActionSymbolWithText(changeType string) string {
	switch changeType {
	case changeTypeCreate:
		return "➕ Create"
	case changeTypeUpdate:
		return "🔄 Update"
	case changeTypeDelete:
		return "➖ Delete"
	default:
		return "❓ Unknown"
	}
}

// truncateString truncates a string to the specified length.
func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}

	if maxLength <= 3 {
		return s[:maxLength]
	}

	return s[:maxLength-3] + "..."
}

// NewCmd creates the diff subcommand.
func NewCmd() *cobra.Command {
	var (
		flags            GlobalFlags
		filter           string
		format           string
		showValues       bool
		impactFilter     string
		onlyNonCompliant bool
		groupByImpact    bool
		detailed         bool
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
			return runDiffCommand(flags, filter, format, showValues, impactFilter, onlyNonCompliant, groupByImpact, detailed)
		},
	}

	// Add global flags
	addGlobalFlags(cmd, &flags)

	// Add diff-specific flags
	cmd.Flags().StringVar(&filter, "filter", "", "Filter repositories by name pattern (regex)")
	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json, unified)")
	cmd.Flags().BoolVar(&showValues, "show-values", false, "Include current values in output")
	cmd.Flags().StringVar(&impactFilter, "impact", "", "Filter by impact level (low, medium, high)")
	cmd.Flags().BoolVar(&onlyNonCompliant, "non-compliant", false, "Show only non-compliant configurations")
	cmd.Flags().BoolVar(&groupByImpact, "group-by-impact", false, "Group results by impact level")
	cmd.Flags().BoolVar(&detailed, "detailed", false, "Show detailed change analysis")

	return cmd
}

// runDiffCommand executes the diff command.
func runDiffCommand(flags GlobalFlags, filter, format string, showValues bool, impactFilter string, onlyNonCompliant, groupByImpact, detailed bool) error {
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
	differences, err := getConfigurationDifferences(flags.Organization, filter, flags.Token, flags.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to get configuration differences: %w", err)
	}

	// Apply additional filters
	if impactFilter != "" {
		differences = filterByImpact(differences, impactFilter)
	}

	if onlyNonCompliant {
		differences = filterNonCompliant(differences)
	}

	// If no differences found, return early
	if len(differences) == 0 {
		if impactFilter != "" || onlyNonCompliant {
			fmt.Println("✅ No differences match the specified filters")
		} else {
			fmt.Println("✅ No configuration differences found - all repositories are compliant")
		}

		return nil
	}

	switch format {
	case "table":
		if groupByImpact {
			displayDiffTableByImpact(differences, showValues, detailed)
		} else {
			displayDiffTable(differences, showValues)
		}
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

// ConfigurationDifference represents a difference between current and target config.
type ConfigurationDifference struct {
	Repository   string `json:"repository"`
	Setting      string `json:"setting"`
	CurrentValue string `json:"currentValue"`
	TargetValue  string `json:"targetValue"`
	ChangeType   string `json:"changeType"` // create, update, delete
	Impact       string `json:"impact"`     // low, medium, high
	Template     string `json:"template"`
	Compliant    bool   `json:"compliant"`
}

// displayDiffTable displays differences in table format.
func displayDiffTable(differences []ConfigurationDifference, showValues bool) {
	if showValues {
		fmt.Printf("%-20s %-30s %-15s %-15s %-12s %s\n",
			"REPOSITORY", "SETTING", "CURRENT", "TARGET", "IMPACT", "ACTION")
	} else {
		fmt.Printf("%-20s %-30s %-12s %-15s %s\n",
			"REPOSITORY", "SETTING", "IMPACT", "ACTION", "TEMPLATE")
	}

	fmt.Println("────────────────────────────────────────────────────────────────────────────────")

	// Group differences by repository for better readability
	groupedDiffs := groupDifferencesByRepository(differences)

	for _, repoName := range getSortedRepositoryNames(groupedDiffs) {
		repoDiffs := groupedDiffs[repoName]

		// Print repository header
		fmt.Printf("\n📁 %s\n", repoName)
		fmt.Println(strings.Repeat("─", 80))

		for _, diff := range repoDiffs {
			actionSymbol := getActionSymbolWithText(diff.ChangeType)
			impactSymbol := getImpactSymbol(diff.Impact)

			if showValues {
				currentDisplay := truncateString(diff.CurrentValue, 15)
				if currentDisplay == "" {
					currentDisplay = colorize("-", "dim")
				}

				fmt.Printf("  %-28s %-15s %-15s %-12s %s\n",
					truncateString(diff.Setting, 28),
					colorizeValue(currentDisplay, diff.ChangeType == changeTypeDelete),
					colorizeValue(truncateString(diff.TargetValue, 15), diff.ChangeType == changeTypeCreate),
					impactSymbol,
					actionSymbol,
				)
			} else {
				fmt.Printf("  %-28s %-12s %-15s %s\n",
					truncateString(diff.Setting, 28),
					impactSymbol,
					actionSymbol,
					colorize(diff.Template, "dim"),
				)
			}
		}
	}
}

// displayDiffJSON displays differences in JSON format.
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

// displayDiffUnified displays differences in unified diff format.
func displayDiffUnified(differences []ConfigurationDifference) {
	for _, diff := range differences {
		fmt.Printf("--- %s (current)\n", diff.Repository)
		fmt.Printf("+++ %s (target)\n", diff.Repository)
		fmt.Printf("@@ %s @@\n", diff.Setting)

		switch diff.ChangeType {
		case changeTypeCreate:
			fmt.Printf("+%s: %s\n", diff.Setting, diff.TargetValue)
		case changeTypeUpdate:
			fmt.Printf("-%s: %s\n", diff.Setting, diff.CurrentValue)
			fmt.Printf("+%s: %s\n", diff.Setting, diff.TargetValue)
		case changeTypeDelete:
			fmt.Printf("-%s: %s\n", diff.Setting, diff.CurrentValue)
		}

		fmt.Println()
	}
}

// displayDiffSummary displays a summary of differences.
func displayDiffSummary(differences []ConfigurationDifference) {
	repoCount := len(getAffectedRepositories(differences))

	impactCounts := map[string]int{
		"low":    0,
		"medium": 0,
		"high":   0,
	}

	actionCounts := map[string]int{
		changeTypeCreate: 0,
		"update":         0,
		changeTypeDelete: 0,
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
	fmt.Printf("  ➕ Create: %d\n", actionCounts[changeTypeCreate])
	fmt.Printf("  🔄 Update: %d\n", actionCounts["update"])
	fmt.Printf("  ➖ Delete: %d\n", actionCounts[changeTypeDelete])
}

// getImpactSymbol returns the symbol for impact level.
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

// groupDifferencesByRepository groups differences by repository name.
func groupDifferencesByRepository(differences []ConfigurationDifference) map[string][]ConfigurationDifference {
	grouped := make(map[string][]ConfigurationDifference)
	for _, diff := range differences {
		grouped[diff.Repository] = append(grouped[diff.Repository], diff)
	}

	return grouped
}

// getSortedRepositoryNames returns repository names sorted alphabetically.
func getSortedRepositoryNames(grouped map[string][]ConfigurationDifference) []string {
	repos := make([]string, 0, len(grouped))
	for repo := range grouped {
		repos = append(repos, repo)
	}

	// Simple sort
	for i := 0; i < len(repos); i++ {
		for j := i + 1; j < len(repos); j++ {
			if repos[i] > repos[j] {
				repos[i], repos[j] = repos[j], repos[i]
			}
		}
	}

	return repos
}

// colorize applies ANSI color codes to text.
func colorize(text, style string) string {
	// For now, return text as-is
	// In a more advanced implementation, we could add color support
	// based on terminal capabilities
	return text
}

// colorizeValue applies color based on whether it's being added or removed.
func colorizeValue(text string, isRemoving bool) string {
	// For now, return text as-is
	// In a more advanced implementation:
	// - Green for additions
	// - Red for removals
	// - Yellow for changes
	return text
}

// filterByImpact filters differences by impact level.
func filterByImpact(differences []ConfigurationDifference, impact string) []ConfigurationDifference {
	var filtered []ConfigurationDifference

	for _, diff := range differences {
		if diff.Impact == impact {
			filtered = append(filtered, diff)
		}
	}

	return filtered
}

// filterNonCompliant filters differences to show only non-compliant configurations.
func filterNonCompliant(differences []ConfigurationDifference) []ConfigurationDifference {
	var filtered []ConfigurationDifference

	for _, diff := range differences {
		if !diff.Compliant {
			filtered = append(filtered, diff)
		}
	}

	return filtered
}

// displayDiffTableByImpact displays differences grouped by impact level.
func displayDiffTableByImpact(differences []ConfigurationDifference, showValues, detailed bool) {
	impactGroups := map[string][]ConfigurationDifference{
		"high":   {},
		"medium": {},
		"low":    {},
	}

	// Group by impact
	for _, diff := range differences {
		if group, exists := impactGroups[diff.Impact]; exists {
			impactGroups[diff.Impact] = append(group, diff)
		}
	}

	// Display each impact level
	impactOrder := []string{"high", "medium", "low"}
	for _, impact := range impactOrder {
		diffs := impactGroups[impact]
		if len(diffs) == 0 {
			continue
		}

		fmt.Printf("\n%s Impact Changes (%d)\n",
			strings.ToUpper(impact), len(diffs))
		fmt.Println(strings.Repeat("═", 80))

		if detailed {
			displayDetailedDifferences(diffs, showValues)
		} else {
			displayDiffTable(diffs, showValues)
		}
	}
}

// displayDetailedDifferences displays differences with detailed analysis.
func displayDetailedDifferences(differences []ConfigurationDifference, showValues bool) {
	for i, diff := range differences {
		if i > 0 {
			fmt.Println()
		}

		fmt.Printf("📋 Change #%d\n", i+1)
		fmt.Printf("Repository: %s\n", diff.Repository)
		fmt.Printf("Setting: %s\n", diff.Setting)
		fmt.Printf("Template: %s\n", diff.Template)
		fmt.Printf("Change Type: %s\n", getActionSymbolWithText(diff.ChangeType))
		fmt.Printf("Impact: %s\n", getImpactSymbol(diff.Impact))

		if showValues {
			fmt.Printf("Current Value: %s\n", formatValue(diff.CurrentValue))
			fmt.Printf("Target Value: %s\n", formatValue(diff.TargetValue))
		}

		// Add detailed analysis based on the setting type
		analysis := analyzeSettingChange(diff)
		if analysis != "" {
			fmt.Printf("Analysis: %s\n", analysis)
		}

		fmt.Println(strings.Repeat("─", 40))
	}
}

// formatValue formats a value for display, handling empty values.
func formatValue(value string) string {
	if value == "" {
		return "(not set)"
	}

	return value
}

// analyzeSettingChange provides detailed analysis for specific setting changes.
func analyzeSettingChange(diff ConfigurationDifference) string {
	switch {
	case strings.Contains(diff.Setting, "visibility"):
		if diff.TargetValue == visibilityPrivate {
			return "Making repository private will restrict access to organization members only"
		}

		return "Making repository public will allow anyone to view the code"

	case strings.Contains(diff.Setting, "branch_protection"):
		if strings.Contains(diff.Setting, "required_reviews") {
			return "Changing review requirements affects code quality gates"
		}

		if strings.Contains(diff.Setting, "enforce_admins") {
			return "Admin enforcement affects repository admin bypass capabilities"
		}

		return "Branch protection rule changes affect repository security"

	case strings.Contains(diff.Setting, "permissions"):
		return "Permission changes affect team access levels to the repository"

	case strings.Contains(diff.Setting, "merge"):
		return "Merge setting changes affect workflow and branch management"

	default:
		return ""
	}
}

// getAffectedRepositories returns unique repository names from differences.
func getAffectedRepositories(differences []ConfigurationDifference) []string {
	repos := make(map[string]bool)
	for _, diff := range differences {
		repos[diff.Repository] = true
	}

	result := make([]string, 0, len(repos))
	for repo := range repos {
		result = append(result, repo)
	}

	return result
}

// compareRepositoryConfigurations compares current and target configurations.
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

	// Compare different configuration categories
	differences = append(differences, compareBasicSettings(repoName, current, targetSettings, templateName)...)
	differences = append(differences, compareSecuritySettings(repoName, current, targetSecurity, templateName)...)
	differences = append(differences, comparePermissionSettings(repoName, current, targetPermissions, templateName)...)

	// Apply policy exceptions to differences
	differences = applyPolicyExceptions(differences, exceptions)

	return differences
}

// getChangeType determines the type of change.
func getChangeType(current, target string) string {
	if current == "" && target != "" {
		return changeTypeCreate
	}

	if current != "" && target == "" {
		return changeTypeDelete
	}

	return "update"
}

// findAppliedTemplate finds which template applies to a repository.
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

// matchRepoPattern checks if a repository name matches a pattern.
func matchRepoPattern(name, pattern string) (bool, error) {
	// Convert simple glob patterns to regex
	if strings.Contains(pattern, "*") {
		// Escape special regex characters except *
		pattern = regexp.QuoteMeta(pattern)
		// Replace escaped \* back to .*
		pattern = strings.ReplaceAll(pattern, "\\*", ".*")
		pattern = "^" + pattern + "$"

		return regexp.MatchString(pattern, name)
	}

	return name == pattern, nil
}

// applyPolicyExceptions applies policy exceptions to differences.
func applyPolicyExceptions(differences []ConfigurationDifference, exceptions []config.PolicyException) []ConfigurationDifference {
	// For now, just return differences as-is
	// In a full implementation, this would check if any differences are covered by exceptions
	// and mark them as compliant or reduce their impact level
	return differences
}

// getConfigurationDifferences retrieves configuration differences for an organization.
func getConfigurationDifferences(organization, filter, token, configPath string) ([]ConfigurationDifference, error) {
	// Create a context
	ctx := context.Background()

	// Get GitHub token from environment or global flags
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	if token == "" {
		return nil, fmt.Errorf("GitHub token not found. Set GITHUB_TOKEN environment variable or use --token flag")
	}

	// Create GitHub client
	client := github.NewRepoConfigClient(token)

	// Load repo config file
	if configPath == "" {
		configPath = "repo-config.yaml"
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

// compareBasicSettings compares basic repository settings.
func compareBasicSettings(repoName string, current *github.RepositoryConfig, targetSettings *config.RepoSettings, templateName string) []ConfigurationDifference {
	var differences []ConfigurationDifference

	if targetSettings == nil {
		return differences
	}

	// Description
	if targetSettings.Description != nil && current.Description != *targetSettings.Description {
		differences = append(differences, createConfigurationDifference(
			repoName, "description", current.Description, *targetSettings.Description, "low", templateName))
	}

	// Homepage
	if targetSettings.Homepage != nil && current.Homepage != *targetSettings.Homepage {
		differences = append(differences, createConfigurationDifference(
			repoName, "homepage", current.Homepage, *targetSettings.Homepage, "low", templateName))
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

	// Repository features
	differences = append(differences, compareRepositoryFeatures(repoName, current, targetSettings, templateName)...)

	// Merge settings
	differences = append(differences, compareMergeSettings(repoName, current, targetSettings, templateName)...)

	return differences
}

// compareRepositoryFeatures compares repository feature settings.
func compareRepositoryFeatures(repoName string, current *github.RepositoryConfig, targetSettings *config.RepoSettings, templateName string) []ConfigurationDifference {
	var differences []ConfigurationDifference

	// Issues feature
	if targetSettings.HasIssues != nil && current.Settings.HasIssues != *targetSettings.HasIssues {
		differences = append(differences, createBooleanConfigurationDifference(
			repoName, "features.issues", current.Settings.HasIssues, *targetSettings.HasIssues, "low", templateName))
	}

	// Wiki feature
	if targetSettings.HasWiki != nil && current.Settings.HasWiki != *targetSettings.HasWiki {
		differences = append(differences, createBooleanConfigurationDifference(
			repoName, "features.wiki", current.Settings.HasWiki, *targetSettings.HasWiki, "low", templateName))
	}

	// Projects feature
	if targetSettings.HasProjects != nil && current.Settings.HasProjects != *targetSettings.HasProjects {
		differences = append(differences, createBooleanConfigurationDifference(
			repoName, "features.projects", current.Settings.HasProjects, *targetSettings.HasProjects, "low", templateName))
	}

	return differences
}

// compareMergeSettings compares merge-related settings.
func compareMergeSettings(repoName string, current *github.RepositoryConfig, targetSettings *config.RepoSettings, templateName string) []ConfigurationDifference {
	var differences []ConfigurationDifference

	// Delete branch on merge
	if targetSettings.DeleteBranchOnMerge != nil && current.Settings.DeleteBranchOnMerge != *targetSettings.DeleteBranchOnMerge {
		differences = append(differences, createBooleanConfigurationDifference(
			repoName, "merge.delete_branch_on_merge", current.Settings.DeleteBranchOnMerge, *targetSettings.DeleteBranchOnMerge, "medium", templateName))
	}

	// Allow squash merge
	if targetSettings.AllowSquashMerge != nil && current.Settings.AllowSquashMerge != *targetSettings.AllowSquashMerge {
		differences = append(differences, createBooleanConfigurationDifference(
			repoName, "merge.allow_squash_merge", current.Settings.AllowSquashMerge, *targetSettings.AllowSquashMerge, "medium", templateName))
	}

	return differences
}

// compareSecuritySettings compares security-related settings.
func compareSecuritySettings(repoName string, current *github.RepositoryConfig, targetSecurity *config.SecuritySettings, templateName string) []ConfigurationDifference {
	var differences []ConfigurationDifference

	if targetSecurity == nil || targetSecurity.BranchProtection == nil {
		return differences
	}

	// Compare branch protection settings
	for branch, targetRule := range targetSecurity.BranchProtection {
		currentRule, exists := current.BranchProtection[branch]
		if !exists {
			// Branch protection doesn't exist, check if we need to create it
			differences = append(differences, compareMissingBranchProtection(repoName, branch, targetRule, templateName)...)
		} else {
			// Compare existing branch protection
			differences = append(differences, compareExistingBranchProtection(repoName, branch, currentRule, targetRule, templateName)...)
		}
	}

	return differences
}

// compareMissingBranchProtection handles cases where branch protection doesn't exist.
func compareMissingBranchProtection(repoName, branch string, targetRule *config.BranchProtectionRule, templateName string) []ConfigurationDifference {
	var differences []ConfigurationDifference

	if targetRule.RequiredReviews != nil && *targetRule.RequiredReviews > 0 {
		differences = append(differences, ConfigurationDifference{
			Repository:   repoName,
			Setting:      fmt.Sprintf("branch_protection.%s.required_reviews", branch),
			CurrentValue: "0",
			TargetValue:  fmt.Sprintf("%d", *targetRule.RequiredReviews),
			ChangeType:   changeTypeCreate,
			Impact:       "high",
			Template:     templateName,
			Compliant:    false,
		})
	}

	return differences
}

// compareExistingBranchProtection compares existing branch protection settings.
func compareExistingBranchProtection(repoName, branch string, currentRule github.BranchProtectionConfig, targetRule *config.BranchProtectionRule, templateName string) []ConfigurationDifference {
	var differences []ConfigurationDifference

	// Required reviews comparison
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

	// Enforce admins comparison
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

	return differences
}

// comparePermissionSettings compares permission-related settings.
func comparePermissionSettings(repoName string, current *github.RepositoryConfig, targetPermissions *config.PermissionSettings, templateName string) []ConfigurationDifference {
	var differences []ConfigurationDifference

	if targetPermissions == nil {
		return differences
	}

	// Team permissions
	for team, targetPerm := range targetPermissions.TeamPermissions {
		currentPerm, exists := current.Permissions.Teams[team]
		if !exists {
			differences = append(differences, ConfigurationDifference{
				Repository:   repoName,
				Setting:      fmt.Sprintf("permissions.team.%s", team),
				CurrentValue: "none",
				TargetValue:  targetPerm,
				ChangeType:   changeTypeCreate,
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

	return differences
}

// createConfigurationDifference creates a configuration difference for string values.
func createConfigurationDifference(repoName, setting, currentValue, targetValue, impact, templateName string) ConfigurationDifference {
	return ConfigurationDifference{
		Repository:   repoName,
		Setting:      setting,
		CurrentValue: currentValue,
		TargetValue:  targetValue,
		ChangeType:   getChangeType(currentValue, targetValue),
		Impact:       impact,
		Template:     templateName,
		Compliant:    false,
	}
}

// createBooleanConfigurationDifference creates a configuration difference for boolean values.
func createBooleanConfigurationDifference(repoName, setting string, currentValue, targetValue bool, impact, templateName string) ConfigurationDifference {
	return ConfigurationDifference{
		Repository:   repoName,
		Setting:      setting,
		CurrentValue: fmt.Sprintf("%t", currentValue),
		TargetValue:  fmt.Sprintf("%t", targetValue),
		ChangeType:   "update",
		Impact:       impact,
		Template:     templateName,
		Compliant:    false,
	}
}
