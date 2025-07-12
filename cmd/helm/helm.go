package helm

import (
	"github.com/spf13/cobra"
)

// HelmCmd represents the helm command
var HelmCmd = &cobra.Command{
	Use:   "helm",
	Short: "Helm chart generation and management",
	Long: `Helm chart generation and management for Kubernetes deployments.

Generate optimized Helm charts with:
- Chart template library
- Values file management system
- Dependency chart handling
- Best practices integration

Available commands:
  chart       Generate Helm chart for projects`,
}

func init() {
	HelmCmd.AddCommand(ChartCmd)
}
