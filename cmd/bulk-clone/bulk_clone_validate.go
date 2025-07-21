// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package bulkclone

import (
	"fmt"

	"github.com/spf13/cobra"

	bulkclonepkg "github.com/gizzahub/gzh-manager-go/pkg/bulk-clone"
)

type bulkCloneValidateOptions struct {
	configFile string
}

func defaultBulkCloneValidateOptions() *bulkCloneValidateOptions {
	return &bulkCloneValidateOptions{}
}

func newBulkCloneValidateCmd() *cobra.Command {
	o := defaultBulkCloneValidateOptions()

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a bulk-clone configuration file",
		Long: `Validate a bulk-clone configuration file against the schema.
		
This command checks that your configuration file:
- Has all required fields
- Uses valid values for enums (protocol, provider, etc.)
- Follows the correct structure
- Has valid regex patterns in ignore_names`,
		Example: `  # Validate a config file
  gzh bulk-clone validate -c bulk-clone.yaml
  
  # Validate config from standard location
  gzh bulk-clone validate --use-config`,
		Args: cobra.NoArgs,
		RunE: o.run,
	}

	cmd.Flags().StringVarP(&o.configFile, "config", "c", o.configFile, "Path to config file to validate")
	cmd.Flags().Bool("use-config", false, "Find and validate config from standard locations")

	return cmd
}

func (o *bulkCloneValidateOptions) run(cmd *cobra.Command, args []string) error {
	useConfig, _ := cmd.Flags().GetBool("use-config")

	if o.configFile == "" && !useConfig {
		return fmt.Errorf("either --config or --use-config must be specified")
	}

	var configPath string

	if useConfig {
		path, err := bulkclonepkg.FindConfigFile()
		if err != nil {
			return fmt.Errorf("failed to find config file: %w", err)
		}

		configPath = path
		fmt.Printf("Found config file at: %s\n", configPath)
	} else {
		configPath = o.configFile
	}

	// First try to load the config using existing validation
	cfg, err := bulkclonepkg.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Then validate against schema
	if err := bulkclonepkg.ValidateConfigWithSchema(configPath); err != nil {
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
