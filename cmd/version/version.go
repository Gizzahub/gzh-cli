package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewVersionCmd 버전 커맨드 생성
func NewVersionCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:          "version",
		Short:        "gz version information",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, _ []string) {
			if version == "" {
				version = "dev"
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "gz version %s\n", version)
		},
	}
}
