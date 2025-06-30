package config

import (
	"github.com/spf13/cobra"
)

// NewConfigCmd creates the config command with subcommands
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management commands",
		Long: `Configuration management commands for gzh-manager.

This command provides utilities for managing gzh.yaml configuration files,
including validation, initialization, and migration tools.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(newValidateCmd())

	return cmd
}

// newValidateCmd creates the config validate subcommand
func newValidateCmd() *cobra.Command {
	var (
		configFile string
		strict     bool
		verbose    bool
	)

	cmd := &cobra.Command{
		Use:   "validate [config-file]",
		Short: "Validate gzh.yaml configuration file",
		Long: `Validate gzh.yaml configuration file for syntax and semantic correctness.

This command checks:
- YAML syntax and structure
- Required fields and data types
- Enum value validation
- Regex pattern validation
- Environment variable accessibility
- Provider-specific requirements

Examples:
  gz config validate                     # Validate default gzh.yaml
  gz config validate my-config.yaml     # Validate specific file
  gz config validate --strict           # Strict validation mode
  gz config validate --verbose          # Verbose output`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine config file path
			if len(args) > 0 {
				configFile = args[0]
			}

			return validateConfig(configFile, strict, verbose)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Configuration file path (default: auto-detect)")
	cmd.Flags().BoolVar(&strict, "strict", false, "Enable strict validation mode")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	return cmd
}
