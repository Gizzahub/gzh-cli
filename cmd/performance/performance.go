package performance

import (
	"github.com/spf13/cobra"
)

// performanceCmd represents the performance command
var performanceCmd = &cobra.Command{
	Use:   "performance",
	Short: "Performance monitoring and optimization tools",
	Long: `Performance monitoring and optimization tools for memory management,
garbage collection tuning, and profiling.`,
}

// NewPerformanceCmd creates the performance command
func NewPerformanceCmd() *cobra.Command {
	// Add subcommands
	performanceCmd.AddCommand(newGCTuningCmd())

	return performanceCmd
}
