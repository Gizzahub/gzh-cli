// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repoconfig

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-manager-go/internal/services"
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
func runApplyCommand(flags GlobalFlags, filter, template string, interactive, force bool) error {
	if flags.Organization == "" {
		return fmt.Errorf("organization is required (use --org flag)")
	}

	// Setup and display header
	displayApplyHeader(flags, filter, template)

	// Create service and prepare options
	service := services.NewRepoConfigService()
	opts := services.ApplyOptions{
		Organization: flags.Organization,
		Filter:       filter,
		Template:     template,
		DryRun:       flags.DryRun,
		Interactive:  interactive,
		Force:        force,
		Token:        flags.Token,
		ConfigFile:   flags.ConfigFile,
		Verbose:      flags.Verbose,
	}

	// Get configuration changes to apply for display
	ctx := context.Background()
	changes, err := service.GetConfigurationChanges(ctx, flags.Organization, filter, template)
	if err != nil {
		return fmt.Errorf("failed to get configuration changes: %w", err)
	}

	if len(changes) == 0 {
		fmt.Println("✅ No configuration changes needed - all repositories are compliant")
		return nil
	}

	// Display planned changes
	displayPlannedChanges(changes, service)

	if flags.DryRun {
		fmt.Println("💡 Run without --dry-run to apply these changes")
		return nil
	}

	// Handle confirmation
	if !confirmApplyChanges(changes, force, interactive) {
		fmt.Println("Configuration application canceled")
		return nil
	}

	// Apply the changes using service
	if err := service.ApplyConfiguration(ctx, opts); err != nil {
		return fmt.Errorf("failed to apply configuration: %w", err)
	}

	fmt.Println()
	fmt.Printf("✅ Configuration application completed successfully\n")
	fmt.Printf("📊 %d repositories updated\n", service.GetAffectedRepoCount(changes))

	return nil
}

// displayApplyHeader displays the command header and verbose information.
func displayApplyHeader(flags GlobalFlags, filter, template string) {
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
}

// displayPlannedChanges displays the list of planned configuration changes.
func displayPlannedChanges(changes []services.ConfigurationChange, service *services.RepoConfigService) {
	fmt.Printf("📋 Planned Changes (%d repositories affected)\n", service.GetAffectedRepoCount(changes))
	fmt.Println()

	for _, change := range changes {
		actionSymbol := getActionSymbol(change.Action)
		fmt.Printf("  %s %s\n", actionSymbol, change.Repository)
		fmt.Printf("    %s: %s → %s\n", change.Setting, change.CurrentValue, change.NewValue)
		fmt.Println()
	}
}

// confirmApplyChanges handles user confirmation for applying changes.
func confirmApplyChanges(changes []services.ConfigurationChange, force, interactive bool) bool {
	// Skip confirmation for force mode or when interactive mode will handle individual confirmations
	if force || interactive {
		return true
	}

	fmt.Printf("Apply %d configuration changes? (y/N): ", len(changes))

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		// Handle error but treat as "no" response
		response = ""
	}

	return response == "y" || response == "yes"
}
