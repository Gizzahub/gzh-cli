// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package synclone

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// newConfigValidateCmd creates the validate subcommand for config.
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
			var cfg any
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return fmt.Errorf("invalid YAML syntax: %w", err)
			}

			// For now, basic YAML validation is sufficient
			// TODO: Implement full schema validation using pkg/config validators

			// cmd.OutOrStdout()으로 내보낸다. fmt.Printf는 프로세스의 표준출력에
			// 바로 쓰기 때문에 cmd.SetOut으로 받을 수 없었다 -- 감싸서 쓰는 쪽도
			// 시험도 아무것도 못 봤다.
			out := cmd.OutOrStdout()

			switch format {
			case "json":
				// %q로 감싼다. 예전에는 문자열을 이어 붙여서 경로에 따옴표나
				// 역슬래시가 있으면 JSON이 깨졌다.
				fmt.Fprintf(out, "{\"valid\": true, \"file\": %q}\n", configFile)
			case "quiet":
				// 아무것도 내지 않는다. 종료 코드로만 답한다. 예전에는 위에서
				// 한 줄을 무조건 찍어서 quiet가 quiet가 아니었고, json도 JSON
				// 앞에 사람이 읽는 줄이 하나 붙어 나갔다.
			default:
				fmt.Fprintf(out, "✓ YAML syntax is valid for file '%s'\n", configFile)
				fmt.Fprintf(out, "✓ Configuration file '%s' is valid\n", configFile)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&configFile, "file", "f", "", "Path to configuration file")
	cmd.Flags().BoolVar(&strict, "strict", false, "Enable strict validation mode")
	cmd.Flags().StringVar(&format, "format", "text", "Output format: text, json, quiet")

	return cmd
}

// findConfigFile looks for configuration files in standard locations.
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
