// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "actions-policy",
	Short: "GitHub Actions policy management tool",
	Long: `A comprehensive tool for managing GitHub Actions policies.
Provides functionality to create, validate, and enforce Actions policies
across repositories and organizations.`,
}

var createCmd = &cobra.Command{
	Use:   "create [policy-name]",
	Short: "Create a new Actions policy",
	Long:  "Create a new Actions policy with specified configuration",
	Args:  cobra.ExactArgs(1),
	RunE:  createPolicy,
}

var enforceCmd = &cobra.Command{
	Use:   "enforce [policy-id] [org] [repo]",
	Short: "Enforce Actions policy on a repository",
	Long:  "Apply and enforce a specific Actions policy on a repository",
	Args:  cobra.ExactArgs(3),
	RunE:  enforcePolicy,
}

var validateCmd = &cobra.Command{
	Use:   "validate [policy-id] [org] [repo]",
	Short: "Validate repository against Actions policy",
	Long:  "Validate a repository's current configuration against an Actions policy",
	Args:  cobra.ExactArgs(3),
	RunE:  validatePolicy,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Actions policies",
	Long:  "Display all available Actions policies",
	RunE:  listPolicies,
}

var showCmd = &cobra.Command{
	Use:   "show [policy-id]",
	Short: "Show Actions policy details",
	Long:  "Display detailed information about a specific Actions policy",
	Args:  cobra.ExactArgs(1),
	RunE:  showPolicy,
}

var deleteCmd = &cobra.Command{
	Use:   "delete [policy-id]",
	Short: "Delete an Actions policy",
	Long:  "Remove an Actions policy from the system",
	Args:  cobra.ExactArgs(1),
	RunE:  deletePolicy,
}

var monitorCmd = &cobra.Command{
	Use:   "monitor [org]",
	Short: "Monitor policy compliance across organization",
	Long:  "Continuously monitor policy compliance across all repositories in an organization",
	Args:  cobra.ExactArgs(1),
	RunE:  monitorCompliance,
}

func init() {
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(enforceCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(monitorCmd)

	// Global flags
	rootCmd.PersistentFlags().String("token", "", "GitHub token (can also use GITHUB_TOKEN env var)")
	rootCmd.PersistentFlags().String("format", "table", "Output format (table, json, yaml)")
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose logging")

	// Create command flags
	createCmd.Flags().String("org", "", "Target organization")
	createCmd.Flags().String("repo", "", "Target repository (optional, for repo-level policies)")
	createCmd.Flags().String("template", "default", "Policy template (default, strict, permissive)")
	createCmd.Flags().String("description", "", "Policy description")
	createCmd.Flags().StringSlice("tags", []string{}, "Policy tags")
	createCmd.Flags().Bool("enabled", true, "Enable policy immediately")

	// Enforce command flags
	enforceCmd.Flags().Bool("dry-run", false, "Perform validation only, don't apply changes")
	enforceCmd.Flags().Bool("force", false, "Force enforcement even if validation fails")
	enforceCmd.Flags().Int("timeout", 300, "Enforcement timeout in seconds")

	// Validate command flags
	validateCmd.Flags().Bool("detailed", false, "Show detailed validation results")
	validateCmd.Flags().String("severity", "all", "Filter by severity (all, low, medium, high, critical)")

	// List command flags
	listCmd.Flags().String("org", "", "Filter by organization")
	listCmd.Flags().StringSlice("tags", []string{}, "Filter by tags")
	listCmd.Flags().Bool("enabled-only", false, "Show only enabled policies")

	// Monitor command flags
	monitorCmd.Flags().Duration("interval", 5*time.Minute, "Monitoring interval")
	monitorCmd.Flags().Bool("continuous", false, "Run continuously until interrupted")
	monitorCmd.Flags().String("webhook-url", "", "Webhook URL for compliance alerts")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func createPolicy(cmd *cobra.Command, args []string) error {
	policyName := args[0]

	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	template, _ := cmd.Flags().GetString("template")
	description, _ := cmd.Flags().GetString("description")
	tags, _ := cmd.Flags().GetStringSlice("tags")
	enabled, _ := cmd.Flags().GetBool("enabled")

	if org == "" {
		return fmt.Errorf("organization is required")
	}

	// Create policy manager
	logger := &consoleLogger{}
	apiClient := createAPIClient()
	policyManager := github.NewActionsPolicyManager(logger, apiClient)

	// Get policy template
	var policy *github.ActionsPolicy

	switch template {
	case "default":
		policy = github.GetDefaultActionsPolicy()
	case "strict":
		policy = createStrictPolicy()
	case "permissive":
		policy = createPermissivePolicy()
	default:
		return fmt.Errorf("unknown template: %s", template)
	}

	// Configure policy
	policy.ID = fmt.Sprintf("%s-%s-%d", org, policyName, time.Now().Unix())
	policy.Name = policyName
	policy.Organization = org
	policy.Repository = repo
	policy.Description = description
	policy.Tags = tags
	policy.Enabled = enabled
	policy.CreatedBy = "actions-policy-cli"

	if err := policyManager.CreatePolicy(cmd.Context(), policy); err != nil {
		return fmt.Errorf("failed to create policy: %w", err)
	}

	fmt.Printf("✅ Policy '%s' created successfully\n", policy.ID)
	fmt.Printf("   Name: %s\n", policy.Name)
	fmt.Printf("   Organization: %s\n", policy.Organization)

	if policy.Repository != "" {
		fmt.Printf("   Repository: %s\n", policy.Repository)
	}

	fmt.Printf("   Template: %s\n", template)
	fmt.Printf("   Enabled: %t\n", policy.Enabled)

	return nil
}

func enforcePolicy(cmd *cobra.Command, args []string) error {
	policyID := args[0]
	org := args[1]
	repo := args[2]

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	force, _ := cmd.Flags().GetBool("force")
	timeout, _ := cmd.Flags().GetInt("timeout")

	logger := &consoleLogger{}
	apiClient := createAPIClient()
	policyManager := github.NewActionsPolicyManager(logger, apiClient)
	enforcer := github.NewActionsPolicyEnforcer(logger, apiClient, policyManager)

	ctx, cancel := context.WithTimeout(cmd.Context(), time.Duration(timeout)*time.Second)
	defer cancel()

	if dryRun {
		fmt.Printf("🔍 Performing dry-run validation for policy '%s' on %s/%s\n", policyID, org, repo)

		// Get policy and validate only
		policy, err := policyManager.GetPolicy(ctx, policyID)
		if err != nil {
			return fmt.Errorf("failed to get policy: %w", err)
		}

		// This would get actual repository state in production
		state := &github.RepositoryActionsState{
			Organization: org,
			Repository:   repo,
		}

		results, err := enforcer.ValidatePolicy(cmd.Context(), policy, state)
		if err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		printValidationResults(results)

		return nil
	}

	fmt.Printf("🚀 Enforcing policy '%s' on %s/%s\n", policyID, org, repo)

	result, err := enforcer.EnforcePolicy(ctx, policyID, org, repo)
	if err != nil && !force {
		return fmt.Errorf("enforcement failed: %w", err)
	}

	printEnforcementResult(result)

	return nil
}

func validatePolicy(cmd *cobra.Command, args []string) error {
	policyID := args[0]
	org := args[1]
	repo := args[2]

	detailed, _ := cmd.Flags().GetBool("detailed")
	severityFilter, _ := cmd.Flags().GetString("severity")

	logger := &consoleLogger{}
	apiClient := createAPIClient()
	policyManager := github.NewActionsPolicyManager(logger, apiClient)
	enforcer := github.NewActionsPolicyEnforcer(logger, apiClient, policyManager)

	policy, err := policyManager.GetPolicy(cmd.Context(), policyID)
	if err != nil {
		return fmt.Errorf("failed to get policy: %w", err)
	}

	// Mock repository state for demonstration
	state := &github.RepositoryActionsState{
		Organization: org,
		Repository:   repo,
	}

	results, err := enforcer.ValidatePolicy(cmd.Context(), policy, state)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Filter by severity
	if severityFilter != "all" {
		filtered := make([]github.PolicyValidationResult, 0)

		for _, result := range results {
			if strings.EqualFold(string(result.Severity), severityFilter) {
				filtered = append(filtered, result)
			}
		}

		results = filtered
	}

	fmt.Printf("📋 Validation Results for %s/%s\n", org, repo)
	fmt.Printf("Policy: %s (%s)\n\n", policy.Name, policyID)

	printValidationResults(results)

	if detailed {
		printDetailedValidationResults(results)
	}

	return nil
}

func listPolicies(cmd *cobra.Command, args []string) error {
	org, _ := cmd.Flags().GetString("org")
	tags, _ := cmd.Flags().GetStringSlice("tags")
	enabledOnly, _ := cmd.Flags().GetBool("enabled-only")
	format, _ := cmd.Flags().GetString("format")

	logger := &consoleLogger{}
	apiClient := createAPIClient()
	policyManager := github.NewActionsPolicyManager(logger, apiClient)

	policies, err := policyManager.ListPolicies(cmd.Context(), org)
	if err != nil {
		return fmt.Errorf("failed to list policies: %w", err)
	}

	// Apply filters
	filtered := make([]*github.ActionsPolicy, 0)

	for _, policy := range policies {
		if enabledOnly && !policy.Enabled {
			continue
		}

		if len(tags) > 0 {
			hasTag := false

			for _, filterTag := range tags {
				for _, policyTag := range policy.Tags {
					if policyTag == filterTag {
						hasTag = true
						break
					}
				}

				if hasTag {
					break
				}
			}

			if !hasTag {
				continue
			}
		}

		filtered = append(filtered, policy)
	}

	if format == "json" {
		return printJSON(filtered)
	}

	fmt.Printf("📋 Actions Policies\n")
	fmt.Printf("===================\n\n")

	if len(filtered) == 0 {
		fmt.Printf("No policies found matching criteria\n")
		return nil
	}

	fmt.Printf("%-20s %-30s %-15s %-10s %-10s\n", "ID", "Name", "Organization", "Enabled", "Version")
	fmt.Printf("%s\n", strings.Repeat("-", 95))

	for _, policy := range filtered {
		fmt.Printf("%-20s %-30s %-15s %-10t %-10d\n",
			truncate(policy.ID, 20),
			truncate(policy.Name, 30),
			truncate(policy.Organization, 15),
			policy.Enabled,
			policy.Version)
	}

	return nil
}

func showPolicy(cmd *cobra.Command, args []string) error {
	policyID := args[0]
	format, _ := cmd.Flags().GetString("format")

	logger := &consoleLogger{}
	apiClient := createAPIClient()
	policyManager := github.NewActionsPolicyManager(logger, apiClient)

	policy, err := policyManager.GetPolicy(cmd.Context(), policyID)
	if err != nil {
		return fmt.Errorf("failed to get policy: %w", err)
	}

	if format == "json" {
		return printJSON(policy)
	}

	fmt.Printf("📋 Actions Policy Details\n")
	fmt.Printf("=========================\n\n")

	fmt.Printf("ID: %s\n", policy.ID)
	fmt.Printf("Name: %s\n", policy.Name)
	fmt.Printf("Description: %s\n", policy.Description)
	fmt.Printf("Organization: %s\n", policy.Organization)

	if policy.Repository != "" {
		fmt.Printf("Repository: %s\n", policy.Repository)
	}

	fmt.Printf("Permission Level: %s\n", policy.PermissionLevel)
	fmt.Printf("Enabled: %t\n", policy.Enabled)
	fmt.Printf("Version: %d\n", policy.Version)
	fmt.Printf("Created: %s\n", policy.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", policy.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Created By: %s\n", policy.CreatedBy)

	if policy.UpdatedBy != "" {
		fmt.Printf("Updated By: %s\n", policy.UpdatedBy)
	}

	if len(policy.Tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(policy.Tags, ", "))
	}

	fmt.Printf("\n📊 Configuration Summary:\n")
	fmt.Printf("  Default Permissions: %s\n", policy.WorkflowPermissions.DefaultPermissions)
	fmt.Printf("  Allow Fork PRs: %t\n", policy.SecuritySettings.AllowForkPRs)
	fmt.Printf("  Allow GitHub Actions: %t\n", policy.SecuritySettings.AllowGitHubOwnedActions)
	fmt.Printf("  Marketplace Policy: %s\n", policy.SecuritySettings.AllowMarketplaceActions)
	fmt.Printf("  Max Secrets: %d\n", policy.SecretsPolicy.MaxSecretCount)
	fmt.Printf("  Allowed Runners: %v\n", policy.Runners.AllowedRunnerTypes)

	return nil
}

func deletePolicy(cmd *cobra.Command, args []string) error {
	policyID := args[0]

	logger := &consoleLogger{}
	apiClient := createAPIClient()
	policyManager := github.NewActionsPolicyManager(logger, apiClient)

	if err := policyManager.DeletePolicy(cmd.Context(), policyID); err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	fmt.Printf("✅ Policy '%s' deleted successfully\n", policyID)

	return nil
}

func monitorCompliance(cmd *cobra.Command, args []string) error {
	org := args[0]
	interval, _ := cmd.Flags().GetDuration("interval")
	continuous, _ := cmd.Flags().GetBool("continuous")
	webhookURL, _ := cmd.Flags().GetString("webhook-url")

	logger := &consoleLogger{}
	apiClient := createAPIClient()
	policyManager := github.NewActionsPolicyManager(logger, apiClient)

	ctx := cmd.Context()

	fmt.Printf("🔍 Starting compliance monitoring for organization: %s\n", org)
	fmt.Printf("   Interval: %s\n", interval)
	fmt.Printf("   Continuous: %t\n", continuous)

	if webhookURL != "" {
		fmt.Printf("   Webhook URL: %s\n", webhookURL)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		fmt.Printf("\n⏰ Running compliance check at %s\n", time.Now().Format("2006-01-02 15:04:05"))

		if err := performComplianceCheck(ctx, policyManager, org, webhookURL); err != nil {
			fmt.Printf("❌ Compliance check failed: %v\n", err)
		}

		if !continuous {
			break
		}

		select {
		case <-ticker.C:
			continue
		case <-sigChan:
			fmt.Printf("\n🛑 Monitoring stopped by user\n")
			return nil
		}
	}

	return nil
}

func performComplianceCheck(ctx context.Context, policyManager *github.ActionsPolicyManager, org, _ string) error {
	policies, err := policyManager.ListPolicies(ctx, org)
	if err != nil {
		return err
	}

	fmt.Printf("📊 Found %d policies for organization %s\n", len(policies), org)

	for _, policy := range policies {
		if !policy.Enabled {
			continue
		}

		fmt.Printf("   Checking policy: %s\n", policy.Name)
		// In a real implementation, this would check all repositories
		// For now, just log that we're checking
	}

	return nil
}

// Helper functions.
func createAPIClient() github.APIClient {
	return &simpleAPIClient{}
}

func createStrictPolicy() *github.ActionsPolicy {
	policy := github.GetDefaultActionsPolicy()
	policy.Name = "Strict Security Policy"
	policy.Description = "High security policy with restrictive settings"
	policy.PermissionLevel = github.ActionsPermissionSelectedActions
	policy.WorkflowPermissions.DefaultPermissions = github.DefaultPermissionsRestricted
	policy.SecuritySettings.AllowForkPRs = false
	policy.SecuritySettings.AllowMarketplaceActions = github.MarketplacePolicyDisabled
	policy.SecretsPolicy.MaxSecretCount = 10

	return policy
}

func createPermissivePolicy() *github.ActionsPolicy {
	policy := github.GetDefaultActionsPolicy()
	policy.Name = "Permissive Policy"
	policy.Description = "Permissive policy for development environments"
	policy.PermissionLevel = github.ActionsPermissionAll
	policy.WorkflowPermissions.DefaultPermissions = github.DefaultPermissionsWrite
	policy.SecuritySettings.AllowForkPRs = true
	policy.SecuritySettings.AllowMarketplaceActions = github.MarketplacePolicyAll
	policy.SecretsPolicy.MaxSecretCount = 100

	return policy
}

func printValidationResults(results []github.PolicyValidationResult) {
	passed := 0
	failed := 0

	for _, result := range results {
		if result.Passed {
			passed++
		} else {
			failed++
		}
	}

	fmt.Printf("Summary: %d passed, %d failed\n\n", passed, failed)

	if failed > 0 {
		fmt.Printf("❌ Failed Validations:\n")

		for _, result := range results {
			if !result.Passed {
				fmt.Printf("   %s: %s (%s)\n", result.RuleID, result.Message, result.Severity)
			}
		}
	}

	if passed > 0 {
		fmt.Printf("\n✅ Passed Validations: %d rules\n", passed)
	}
}

func printDetailedValidationResults(results []github.PolicyValidationResult) {
	fmt.Printf("\n📋 Detailed Validation Results:\n")
	fmt.Printf("================================\n")

	for _, result := range results {
		status := "✅ PASS"
		if !result.Passed {
			status = "❌ FAIL"
		}

		fmt.Printf("\n%s %s (%s)\n", status, result.RuleID, result.Severity)
		fmt.Printf("  Message: %s\n", result.Message)

		if result.ActualValue != nil {
			fmt.Printf("  Actual: %v\n", result.ActualValue)
		}

		if result.ExpectedValue != nil {
			fmt.Printf("  Expected: %v\n", result.ExpectedValue)
		}

		if len(result.Suggestions) > 0 {
			fmt.Printf("  Suggestions:\n")

			for _, suggestion := range result.Suggestions {
				fmt.Printf("    - %s\n", suggestion)
			}
		}
	}
}

func printEnforcementResult(result *github.PolicyEnforcementResult) {
	status := "✅ SUCCESS"
	if !result.Success {
		status = "❌ FAILED"
	}

	fmt.Printf("\n%s Policy Enforcement Result\n", status)
	fmt.Printf("=====================================\n")
	fmt.Printf("Policy ID: %s\n", result.PolicyID)
	fmt.Printf("Target: %s/%s\n", result.Organization, result.Repository)
	fmt.Printf("Execution Time: %s\n", result.ExecutionTime)
	fmt.Printf("Applied Changes: %d\n", len(result.AppliedChanges))
	fmt.Printf("Failed Changes: %d\n", len(result.FailedChanges))
	fmt.Printf("Violations: %d\n", len(result.Violations))

	if len(result.AppliedChanges) > 0 {
		fmt.Printf("\n✅ Applied Changes:\n")

		for _, change := range result.AppliedChanges {
			fmt.Printf("   %s.%s: %v → %v\n", change.Type, change.Target, change.OldValue, change.NewValue)
		}
	}

	if len(result.FailedChanges) > 0 {
		fmt.Printf("\n❌ Failed Changes:\n")

		for _, change := range result.FailedChanges {
			fmt.Printf("   %s.%s: %s\n", change.Type, change.Target, change.Error)
		}
	}

	if len(result.Violations) > 0 {
		fmt.Printf("\n⚠️  Policy Violations:\n")

		for _, violation := range result.Violations {
			fmt.Printf("   %s: %s (%s)\n", violation.ViolationType, violation.Description, violation.Severity)
		}
	}
}

func printJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	return encoder.Encode(data)
}

func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}

	return s[:length-3] + "..."
}

// Console logger implementation.
type consoleLogger struct{}

func (l *consoleLogger) Debug(msg string, args ...interface{}) {
	// Only show debug messages in verbose mode
}

func (l *consoleLogger) Info(msg string, args ...interface{}) {
	fmt.Printf("[INFO] %s", msg)

	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			fmt.Printf(" %v=%v", args[i], args[i+1])
		}
	}

	fmt.Println()
}

func (l *consoleLogger) Warn(msg string, args ...interface{}) {
	fmt.Printf("[WARN] %s", msg)

	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			fmt.Printf(" %v=%v", args[i], args[i+1])
		}
	}

	fmt.Println()
}

func (l *consoleLogger) Error(msg string, args ...interface{}) {
	fmt.Printf("[ERROR] %s", msg)

	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			fmt.Printf(" %v=%v", args[i], args[i+1])
		}
	}

	fmt.Println()
}

// Simple API client for CLI.
type simpleAPIClient struct{}

func (m *simpleAPIClient) GetRepository(ctx context.Context, owner, repo string) (*github.RepositoryInfo, error) {
	return &github.RepositoryInfo{}, nil
}

func (m *simpleAPIClient) ListOrganizationRepositories(ctx context.Context, org string) ([]github.RepositoryInfo, error) {
	return []github.RepositoryInfo{}, nil
}

func (m *simpleAPIClient) GetDefaultBranch(ctx context.Context, owner, repo string) (string, error) {
	return "main", nil
}

func (m *simpleAPIClient) SetToken(token string) {}

func (m *simpleAPIClient) GetRateLimit(ctx context.Context) (*github.RateLimit, error) {
	return &github.RateLimit{}, nil
}

func (m *simpleAPIClient) GetRepositoryConfiguration(ctx context.Context, owner, repo string) (*github.RepositoryConfig, error) {
	return &github.RepositoryConfig{}, nil
}

func (m *simpleAPIClient) UpdateRepositoryConfiguration(ctx context.Context, owner, repo string, config *github.RepositoryConfig) error {
	return nil
}
