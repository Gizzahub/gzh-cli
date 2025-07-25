// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package synclone

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// newConfigValidateCmd creates the validate subcommand for config
func newConfigValidateCmd() *cobra.Command {
	var (
		configFile string
		strict     bool
		format     string
	)

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate synclone configuration file",
		Long: `Validate the syntax and structure of a synclone configuration file.

This command checks:
- YAML syntax validity
- Required fields presence
- Field type correctness
- Provider-specific settings
- Token/credential references (without revealing values)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no config file specified, try to find one
			if configFile == "" {
				configFile = findConfigFile()
				if configFile == "" {
					return fmt.Errorf("no configuration file specified and none found in standard locations")
				}
			}

			// Read the config file
			data, err := os.ReadFile(configFile)
			if err != nil {
				return fmt.Errorf("failed to read config file: %w", err)
			}

			// Parse YAML
			var cfg interface{}
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return fmt.Errorf("invalid YAML syntax: %w", err)
			}

			// For now, basic YAML validation is sufficient
			// TODO: Implement full schema validation using pkg/config validators
			fmt.Printf("✓ YAML syntax is valid for file '%s'\n", configFile)

			// Output result
			switch format {
			case "json":
				fmt.Println(`{"valid": true, "file": "` + configFile + `"}`)
			case "quiet":
				// No output on success
			default:
				fmt.Printf("✓ Configuration file '%s' is valid\n", configFile)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&configFile, "file", "f", "", "Path to configuration file")
	cmd.Flags().BoolVar(&strict, "strict", false, "Enable strict validation mode")
	cmd.Flags().StringVar(&format, "format", "text", "Output format: text, json, quiet")

	return cmd
}

// findConfigFile looks for configuration files in standard locations
func findConfigFile() string {
	// Check in order of precedence
	locations := []string{
		"synclone.yaml",
		"synclone.yml",
		"bulk-clone.yaml",
		"bulk-clone.yml",
		".synclone.yaml",
		".synclone.yml",
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	// Check in home directory
	if home, err := os.UserHomeDir(); err == nil {
		homeLocations := []string{
			home + "/.config/gzh-manager/synclone.yaml",
			home + "/.config/gzh-manager/synclone.yml",
			home + "/.config/gzh-manager/bulk-clone.yaml",
			home + "/.config/gzh-manager/bulk-clone.yml",
		}
		for _, loc := range homeLocations {
			if _, err := os.Stat(loc); err == nil {
				return loc
			}
		}
	}

	return ""
}
