// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package apply

import (
	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/internal/services"
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

// getActionSymbol returns the symbol for action type.
//
//nolint:unused // Utility function for future use
func getActionSymbol(changeType string) string {
	switch changeType {
	case "create":
		return "➕"
	case "update":
		return "🔄"
	case "delete":
		return "➖"
	default:
		return "📝"
	}
}

// NewCmd creates the apply subcommand.
func NewCmd() *cobra.Command {
	var (
		flags       GlobalFlags
		filter      string
		template    string
		interactive bool
		force       bool
	)

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply repository configuration (currently unsupported)",
		Long: `Repository configuration apply and change planning are currently unsupported.

Execution returns an explicit error without reading configuration files,
looking up credentials, contacting GitHub, or changing repositories.
Dry-run, force, and interactive options do not enable a working preview or apply.
Existing flags remain available for compatibility.`,
		Example: "gz repo-config apply --help",
		RunE: func(cmd *cobra.Command, args []string) error {
			return services.NewRepoConfigService().ApplyConfiguration(cmd.Context(), services.ApplyOptions{
				Organization: flags.Organization,
				Filter:       filter,
				Template:     template,
				DryRun:       flags.DryRun,
				Interactive:  interactive,
				Force:        force,
				Token:        flags.Token,
				ConfigFile:   flags.ConfigFile,
				Verbose:      flags.Verbose,
			})
		},
	}

	// Add global flags
	addGlobalFlags(cmd, &flags)

	// Add apply-specific flags
	cmd.Flags().StringVar(&filter, "filter", "", "Filter repositories by name pattern")
	cmd.Flags().StringVar(&template, "template", "", "Configuration template to apply")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Enable interactive mode")
	cmd.Flags().BoolVar(&force, "force", false, "Force apply without confirmation")

	return cmd
}
