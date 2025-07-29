// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package synclone

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// newConfigConvertCmd creates the convert subcommand for config.
func newConfigConvertCmd() *cobra.Command {
	var (
		inputFile  string
		outputFile string
		fromFormat string
		toFormat   string
		backup     bool
	)

	cmd := &cobra.Command{
		Use:   "convert",
		Short: "Convert between configuration formats",
		Long: `Convert synclone configuration files between different formats and versions.

Supported conversions:
- bulk-clone v1 → synclone v2
- JSON → YAML
- YAML → JSON
- Legacy format → Modern format`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate input
			if inputFile == "" {
				return fmt.Errorf("input file required")
			}

			// Read input file
			data, err := os.ReadFile(inputFile)
			if err != nil {
				return fmt.Errorf("failed to read input file: %w", err)
			}

			// Create backup if requested
			if backup {
				backupFile := inputFile + ".bak"
				if err := os.WriteFile(backupFile, data, 0o644); err != nil {
					return fmt.Errorf("failed to create backup: %w", err)
				}
				fmt.Printf("Created backup: %s\n", backupFile)
			}

			// Perform conversion based on format
			var result []byte
			switch {
			case fromFormat == "v1" && toFormat == "v2":
				result, err = convertV1ToV2(data)
			case fromFormat == "json" && toFormat == "yaml":
				result, err = convertJSONToYAML(data)
			case fromFormat == "yaml" && toFormat == "json":
				result, err = convertYAMLToJSON(data)
			default:
				return fmt.Errorf("unsupported conversion: %s → %s", fromFormat, toFormat)
			}

			if err != nil {
				return fmt.Errorf("conversion failed: %w", err)
			}

			// Write output
			if outputFile == "" {
				outputFile = inputFile
			}
			if err := os.WriteFile(outputFile, result, 0o644); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}

			fmt.Printf("✓ Successfully converted %s → %s\n", inputFile, outputFile)
			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "file", "f", "", "Input configuration file")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (defaults to input file)")
	cmd.Flags().StringVar(&fromFormat, "from", "v1", "Source format: v1, json, yaml")
	cmd.Flags().StringVar(&toFormat, "to", "v2", "Target format: v2, json, yaml")
	cmd.Flags().BoolVar(&backup, "backup", true, "Create backup of original file")

	cmd.MarkFlagRequired("file")

	return cmd
}

// convertV1ToV2 converts bulk-clone v1 format to synclone v2 format.
func convertV1ToV2(data []byte) ([]byte, error) {
	// Parse v1 format
	var v1Config map[string]interface{}
	if err := yaml.Unmarshal(data, &v1Config); err != nil {
		return nil, fmt.Errorf("failed to parse v1 config: %w", err)
	}

	// Transform to v2 format
	v2Config := map[string]interface{}{
		"version": "2.0",
		"synclone": map[string]interface{}{
			"providers": transformProviders(v1Config),
		},
	}

	// Marshal to YAML
	return yaml.Marshal(v2Config)
}

// transformProviders transforms v1 provider configuration to v2 format.
func transformProviders(v1Config map[string]interface{}) []map[string]interface{} {
	var providers []map[string]interface{}

	// Handle bulk_clone section
	if bulkClone, ok := v1Config["bulk_clone"].(map[string]interface{}); ok {
		if roots, ok := bulkClone["repository_roots"].([]interface{}); ok {
			for _, root := range roots {
				if rootMap, ok := root.(map[string]interface{}); ok {
					provider := transformProvider(rootMap)
					if provider != nil {
						providers = append(providers, provider)
					}
				}
			}
		}
	}

	return providers
}

// transformProvider transforms a single provider configuration.
func transformProvider(v1Provider map[string]interface{}) map[string]interface{} {
	provider := make(map[string]interface{})

	// Copy basic fields
	if name, ok := v1Provider["name"].(string); ok {
		provider["name"] = name
	}
	if providerType, ok := v1Provider["provider"].(string); ok {
		provider["type"] = providerType
	}
	if targetDir, ok := v1Provider["target_dir"].(string); ok {
		provider["target_dir"] = targetDir
	}

	// Transform provider-specific settings
	switch provider["type"] {
	case "github":
		if org, ok := v1Provider["organization"].(string); ok {
			provider["organization"] = org
		}
		if token, ok := v1Provider["token"].(string); ok {
			provider["token"] = token
		}
	case "gitlab":
		if group, ok := v1Provider["group"].(string); ok {
			provider["group"] = group
		}
		if token, ok := v1Provider["token"].(string); ok {
			provider["token"] = token
		}
		if baseURL, ok := v1Provider["base_url"].(string); ok {
			provider["base_url"] = baseURL
		}
	}

	// Copy filters
	if filters, ok := v1Provider["filters"].(map[string]interface{}); ok {
		provider["filters"] = filters
	}

	return provider
}

// convertJSONToYAML converts JSON configuration to YAML.
func convertJSONToYAML(data []byte) ([]byte, error) {
	var config interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return yaml.Marshal(config)
}

// convertYAMLToJSON converts YAML configuration to JSON.
func convertYAMLToJSON(data []byte) ([]byte, error) {
	// For now, just return an error - would need JSON encoding
	return nil, fmt.Errorf("yaml to JSON conversion not yet implemented")
}
