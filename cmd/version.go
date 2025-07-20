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
		Run: func(cmd *cobra.Command, _ []string) {
			if version == "" {
				version = "dev"
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "gz version %s\n", version) // Ignore error
		},
	}
}
