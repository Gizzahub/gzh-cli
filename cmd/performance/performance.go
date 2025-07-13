package performance

import (
	"github.com/spf13/cobra"
)

// NewPerformanceCmd creates the performance command
func NewPerformanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "performance",
		Short: "Performance monitoring and optimization tools",
		Long: `Performance monitoring and optimization tools for memory management,
garbage collection tuning, and profiling.`,
	}

	// Add subcommands
	cmd.AddCommand(newGCTuningCmd())

	return cmd
}
