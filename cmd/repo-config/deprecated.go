// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repoconfig

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// MakeDeprecatedCommand modifies the repo-config command to show deprecation
func MakeDeprecatedCommand(cmd *cobra.Command) *cobra.Command {
	// Mark as deprecated
	cmd.Deprecated = "use 'gz repo-sync config' instead"

	// Override the run function
	originalRun := cmd.RunE
	cmd.RunE = func(c *cobra.Command, args []string) error {
		// Show deprecation warning
		fmt.Fprintf(os.Stderr, "\n"+
			"╔════════════════════════════════════════════════════════════════╗\n"+
			"║                    DEPRECATION WARNING                         ║\n"+
			"╠════════════════════════════════════════════════════════════════╣\n"+
			"║  'repo-config' is deprecated and will be removed in v3.0      ║\n"+
			"║                                                                ║\n"+
			"║  Please use: gz repo-sync config                              ║\n"+
			"║                                                                ║\n"+
			"║  For more information: gz help migrate                        ║\n"+
			"╚════════════════════════════════════════════════════════════════╝\n\n")

		// Set environment variable to indicate deprecated command usage
		os.Setenv("GZ_DEPRECATED_COMMAND", "repo-config")

		// Try to run the original command
		if originalRun != nil {
			return originalRun(c, args)
		}

		// If no original run function, suggest the new command
		newArgs := []string{"repo-sync", "config"}
		if len(args) > 0 {
			newArgs = append(newArgs, args...)
		}

		fmt.Fprintf(os.Stderr, "Suggested command: gz %s\n",
			stringSliceToString(newArgs))

		return fmt.Errorf("command has been restructured")
	}

	// Update help text
	cmd.Short = "(DEPRECATED) " + cmd.Short

	return cmd
}

func stringSliceToString(slice []string) string {
	result := ""
	for i, s := range slice {
		if i > 0 {
			result += " "
		}
		result += s
	}
	return result
}
