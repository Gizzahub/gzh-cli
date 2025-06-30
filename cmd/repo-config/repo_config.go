package repoconfig

import (
	"github.com/spf13/cobra"
)

// NewRepoConfigCmd creates the repo-config command with subcommands
func NewRepoConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo-config",
		Short: "GitHub repository configuration management",
		Long: `Manage GitHub repository configurations across organizations.

This command provides tools for managing repository settings, security policies,
and compliance across entire GitHub organizations using infrastructure-as-code
principles.

Key Features:
- Apply consistent configuration across repositories
- Manage security policies and branch protection rules
- Template-based configuration management
- Compliance auditing and reporting
- Dry-run mode for safe changes

Examples:
  gz repo-config list                    # List repositories with current settings
  gz repo-config apply                   # Apply configuration to repositories
  gz repo-config validate               # Validate configuration files
  gz repo-config diff                   # Show differences between current and target
  gz repo-config audit                  # Generate compliance audit report`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newApplyCmd())
	cmd.AddCommand(newValidateCmd())
	cmd.AddCommand(newDiffCmd())
	cmd.AddCommand(newAuditCmd())
	cmd.AddCommand(newTemplateCmd())

	return cmd
}

// Global flags for all repo-config commands
type GlobalFlags struct {
	Organization string
	ConfigFile   string
	Token        string
	DryRun       bool
	Verbose      bool
	Parallel     int
	Timeout      string
}

// addGlobalFlags adds common flags to a command
func addGlobalFlags(cmd *cobra.Command, flags *GlobalFlags) {
	cmd.Flags().StringVarP(&flags.Organization, "org", "o", "", "GitHub organization name")
	cmd.Flags().StringVarP(&flags.ConfigFile, "config", "c", "", "Configuration file path")
	cmd.Flags().StringVarP(&flags.Token, "token", "t", "", "GitHub personal access token")
	cmd.Flags().BoolVar(&flags.DryRun, "dry-run", false, "Preview changes without applying")
	cmd.Flags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Verbose output")
	cmd.Flags().IntVar(&flags.Parallel, "parallel", 5, "Number of parallel operations")
	cmd.Flags().StringVar(&flags.Timeout, "timeout", "30s", "API timeout duration")
}
