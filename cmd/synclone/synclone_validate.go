// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package synclone

import (
	"fmt"

	"github.com/spf13/cobra"

	synclonepkg "github.com/gizzahub/gzh-manager-go/pkg/synclone"
)

type syncCloneValidateOptions struct {
	configFile string
}

func defaultSyncCloneValidateOptions() *syncCloneValidateOptions {
	return &syncCloneValidateOptions{}
}

func newSyncCloneValidateCmd() *cobra.Command {
	o := defaultSyncCloneValidateOptions()

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a synclone configuration file",
		Long: `Validate a synclone configuration file against the schema.

This command checks that your configuration file:
- Has all required fields
- Uses valid values for enums (protocol, provider, etc.)
- Follows the correct structure
- Has valid regex patterns in ignore_names`,
		Example: `  # Validate a config file
  gzh synclone validate -c synclone.yaml

  # Validate config from standard location
  gzh synclone validate --use-config`,
		Args: cobra.NoArgs,
		RunE: o.run,
	}

	cmd.Flags().StringVarP(&o.configFile, "config", "c", o.configFile, "Path to config file to validate")
	cmd.Flags().Bool("use-config", false, "Find and validate config from standard locations")

	return cmd
}

func (o *syncCloneValidateOptions) run(cmd *cobra.Command, args []string) error {
	useConfig, _ := cmd.Flags().GetBool("use-config")

	if o.configFile == "" && !useConfig {
		return fmt.Errorf("either --config or --use-config must be specified")
	}

	var configPath string

	if useConfig {
		path, err := synclonepkg.FindConfigFile()
		if err != nil {
			return fmt.Errorf("failed to find config file: %w", err)
		}

		configPath = path
		fmt.Printf("Found config file at: %s\n", configPath)
	} else {
		configPath = o.configFile
	}

	// First try to load the config using existing validation
	cfg, err := synclonepkg.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Then validate against schema
	if err := synclonepkg.ValidateConfigWithSchema(configPath); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	// Print summary
	fmt.Println("✓ Configuration is valid!")
	fmt.Printf("  Version: %s\n", cfg.Version)
	fmt.Printf("  Default protocol: %s\n", cfg.Default.Protocol)
	fmt.Printf("  Repository roots: %d\n", len(cfg.RepoRoots))
	fmt.Printf("  Ignore patterns: %d\n", len(cfg.IgnoreNameRegexes))

	return nil
}
