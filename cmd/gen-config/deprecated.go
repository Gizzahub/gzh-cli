// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package genconfig

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// DeprecatedRedirect handles the redirection from gen-config to synclone config
func DeprecatedRedirect(args []string) error {
	// Show deprecation warning
	fmt.Fprintf(os.Stderr, "\n"+
		"╔════════════════════════════════════════════════════════════════╗\n"+
		"║                    DEPRECATION WARNING                         ║\n"+
		"╠════════════════════════════════════════════════════════════════╣\n"+
		"║  'gen-config' is deprecated and will be removed in v3.0       ║\n"+
		"║                                                                ║\n"+
		"║  Please use: gz synclone config generate                      ║\n"+
		"║                                                                ║\n"+
		"║  For more information: gz help migrate                        ║\n"+
		"╚════════════════════════════════════════════════════════════════╝\n\n")

	// Map old commands to new commands
	newArgs := []string{"synclone", "config", "generate"}
	
	// If there are subcommands, append them
	if len(args) > 0 {
		// Skip the "gen-config" part if it's in args
		startIdx := 0
		if args[0] == "gen-config" {
			startIdx = 1
		}
		newArgs = append(newArgs, args[startIdx:]...)
	}

	// Set environment variable to indicate deprecated command usage
	os.Setenv("GZ_DEPRECATED_COMMAND", "gen-config")

	// Execute the new command
	cmd := exec.Command("gz", newArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// MakeDeprecatedCommand modifies the existing gen-config command to show deprecation
func MakeDeprecatedCommand(cmd *cobra.Command) *cobra.Command {
	// Mark as deprecated
	cmd.Deprecated = "use 'gz synclone config generate' instead"
	
	// Override the run function to redirect
	originalRun := cmd.RunE
	cmd.RunE = func(c *cobra.Command, args []string) error {
		// For the root gen-config command, show help and deprecation
		if len(args) == 0 && len(os.Args) == 2 {
			fmt.Fprintf(os.Stderr, "\n"+
				"Error: 'gen-config' has been moved to 'gz synclone config generate'\n\n"+
				"Examples of new commands:\n"+
				"  gz synclone config generate init\n"+
				"  gz synclone config generate template simple\n"+
				"  gz synclone config generate discover ~/projects\n\n"+
				"Run 'gz synclone config generate --help' for more information.\n")
			return fmt.Errorf("command restructured")
		}
		
		// Otherwise, try to run the original command with deprecation warning
		fmt.Fprintf(os.Stderr, "\nWarning: 'gen-config' is deprecated. Use 'gz synclone config generate' instead.\n\n")
		
		if originalRun != nil {
			return originalRun(c, args)
		}
		return nil
	}
	
	// Update help text
	cmd.Short = "(DEPRECATED) " + cmd.Short
	
	return cmd
}