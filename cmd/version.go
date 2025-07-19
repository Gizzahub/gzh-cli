package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:          "version",
		Short:        "gz version information",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			if version == "" {
				version = "dev"
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "gz version %s\n", version); err != nil {
				// Error writing version info - silently fail
			}
		},
	}
}
