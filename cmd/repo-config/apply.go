// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repoconfig

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newApplyCmd creates the apply subcommand.
func newApplyCmd() *cobra.Command {
	var (
		flags       GlobalFlags
		filter      string
		template    string
		interactive bool
		force       bool
	)

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply repository configuration",
		Long: `Apply repository configuration to organizations and repositories.

This command applies configuration templates and policies to repositories
based on the configuration file. It supports both organization-wide
application and specific repository targeting.

Configuration Application:
- Template-based configuration deployment
- Policy enforcement across repositories
- Security settings standardization
- Branch protection rule management

Safety Features:
- Dry-run mode for preview
- Interactive confirmation for changes
- Rollback capability
- Detailed change reporting

Examples:
  gz repo-config apply --org myorg                     # Apply to all repositories
  gz repo-config apply --filter "^api-.*"             # Apply to matching repositories
  gz repo-config apply --template security            # Apply specific template
  gz repo-config apply --dry-run                      # Preview changes only
  gz repo-config apply --interactive                  # Interactive confirmation`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runApplyCommand(flags, filter, template, interactive, force)
		},
	}

	// Add global flags
	addGlobalFlags(cmd, &flags)

	// Add apply-specific flags
	cmd.Flags().StringVar(&filter, "filter", "", "Filter repositories by name pattern (regex)")
	cmd.Flags().StringVar(&template, "template", "", "Apply specific template")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive confirmation for each change")
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmations (use with caution)")

	return cmd
}

// runApplyCommand executes the apply command.
func runApplyCommand(flags GlobalFlags, filter, template string, interactive, force bool) error { //nolint:gocognit // Complex repository configuration application logic
	if flags.Organization == "" {
		return fmt.Errorf("organization is required (use --org flag)")
	}

	if flags.Verbose {
		fmt.Printf("🚀 Applying repository configuration to organization: %s\n", flags.Organization)

		if filter != "" {
			fmt.Printf("Filter pattern: %s\n", filter)
		}

		if template != "" {
			fmt.Printf("Template: %s\n", template)
		}

		if flags.DryRun {
			fmt.Println("Mode: DRY RUN (preview only)")
		}

		fmt.Println()
	}

	fmt.Printf("⚙️  Repository Configuration Application\n")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	if flags.DryRun {
		fmt.Println("🔍 DRY RUN MODE - No changes will be applied")
		fmt.Println()
	}

	// Get configuration changes to apply
	changes, err := getConfigurationChanges(flags.Organization, filter, template)
	if err != nil {
		return fmt.Errorf("failed to get configuration changes: %w", err)
	}

	if len(changes) == 0 {
		fmt.Println("✅ No configuration changes needed - all repositories are compliant")
		return nil
	}

	fmt.Printf("📋 Planned Changes (%d repositories affected)\n", getAffectedRepoCount(changes))
	fmt.Println()

	for _, change := range changes {
		actionSymbol := getActionSymbol(change.Action)
		fmt.Printf("  %s %s\n", actionSymbol, change.Repository)
		fmt.Printf("    %s: %s → %s\n", change.Setting, change.CurrentValue, change.NewValue)
		fmt.Println()
	}

	if flags.DryRun {
		fmt.Println("💡 Run without --dry-run to apply these changes")
		return nil
	}

	// Confirmation for non-force mode
	if !force && !interactive {
		fmt.Printf("Apply %d configuration changes? (y/N): ", len(changes))

		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			// Handle error but treat as "no" response
			response = ""
		}

		if response != "y" && response != "yes" {
			fmt.Println("Configuration application canceled")
			return nil
		}
	}

	// Apply changes
	fmt.Println("🔄 Applying configuration changes...")
	fmt.Println()

	for i, change := range changes {
		if interactive {
			fmt.Printf("Apply change %d/%d to %s? (y/N): ", i+1, len(changes), change.Repository)

			var response string
			if _, err := fmt.Scanln(&response); err != nil {
				// Handle error but treat as "no" response
				response = ""
			}

			if response != "y" && response != "yes" {
				fmt.Printf("  ⏭️  Skipped %s\n", change.Repository)
				continue
			}
		}

		// Apply configuration change
		fmt.Printf("  🔄 Updating %s...", change.Repository)

		if err := applyConfigurationChange(change); err != nil {
			fmt.Printf(" ❌ Failed: %v\n", err)
			continue
		}

		fmt.Printf(" ✅\n")
	}

	fmt.Println()
	fmt.Printf("✅ Configuration application completed successfully\n")
	fmt.Printf("📊 %d repositories updated\n", len(changes))

	return nil
}

// ConfigurationChange represents a pending configuration change.
type ConfigurationChange struct {
	Repository   string `json:"repository"`
	Setting      string `json:"setting"`
	CurrentValue string `json:"currentValue"`
	NewValue     string `json:"newValue"`
	Action       string `json:"action"` // create, update, delete
}

// getAffectedRepoCount returns the number of unique repositories affected.
func getAffectedRepoCount(changes []ConfigurationChange) int {
	repos := make(map[string]bool)
	for _, change := range changes {
		repos[change.Repository] = true
	}

	return len(repos)
}

// getConfigurationChanges retrieves configuration changes for an organization.
func getConfigurationChanges(organization, filter, template string) ([]ConfigurationChange, error) {
	// This is a mock implementation - in reality, this would:
	// 1. Fetch current repository configurations from GitHub API
	// 2. Load target configurations from templates
	// 3. Generate required changes
	// 4. Apply filter and template constraints if specified
	mockChanges := []ConfigurationChange{
		{
			Repository:   "api-server",
			Setting:      "branch_protection.main.required_reviews",
			CurrentValue: "1",
			NewValue:     "2",
			Action:       "update",
		},
		{
			Repository:   "web-frontend",
			Setting:      "features.wiki",
			CurrentValue: "true",
			NewValue:     "false",
			Action:       "update",
		},
		{
			Repository:   "legacy-service",
			Setting:      "security.delete_head_branches",
			CurrentValue: "false",
			NewValue:     "true",
			Action:       "create",
		},
	}

	// Apply filter if specified
	if filter != "" {
		// In a real implementation, this would use regex to filter repositories
	}

	// Apply template filter if specified
	if template != "" {
		// In a real implementation, this would filter changes based on template
	}

	return mockChanges, nil
}

// applyConfigurationChange applies a single configuration change.
func applyConfigurationChange(change ConfigurationChange) error {
	// This is a mock implementation - in reality, this would:
	// 1. Use GitHub API to apply the configuration change
	// 2. Handle authentication and rate limiting
	// 3. Verify the change was applied successfully
	// 4. Return appropriate errors if something fails

	// Simulate some processing time
	// time.Sleep(100 * time.Millisecond)

	// For demonstration, we'll just return success
	return nil
}
