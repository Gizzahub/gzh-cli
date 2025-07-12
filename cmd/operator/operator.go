package operator

import (
	"github.com/spf13/cobra"
)

// OperatorCmd represents the operator command
var OperatorCmd = &cobra.Command{
	Use:   "operator",
	Short: "Kubernetes operator development and management",
	Long: `Kubernetes operator development and management tools.

Generate and manage Kubernetes operators with:
- Custom Resource Definition (CRD) generation
- Controller implementation scaffolding
- Resource lifecycle management
- Operator SDK integration
- Best practices enforcement

Available commands:
  generate    Generate Kubernetes operator components
  deploy      Deploy operator to Kubernetes cluster
  validate    Validate operator configurations`,
}

func init() {
	OperatorCmd.AddCommand(GenerateCmd)
	OperatorCmd.AddCommand(DeployCmd)
	OperatorCmd.AddCommand(ValidateCmd)
}
