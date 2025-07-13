package terraform

import (
	"github.com/spf13/cobra"
)

// TerraformCmd represents the terraform command
var TerraformCmd = &cobra.Command{
	Use:   "terraform",
	Short: "Terraform Infrastructure as Code management",
	Long: `Terraform Infrastructure as Code management tools.

Generate and manage Terraform configurations with:
- Infrastructure module generation
- State management and backends
- Cloud provider specific modules
- Resource dependency management
- Terraform plan/apply automation
- Multi-environment deployment

Available commands:
  generate    Generate Terraform modules and configurations
  plan        Run Terraform plan operations
  apply       Apply Terraform configurations
  state       Manage Terraform state files`,
	Aliases: []string{"tf"},
}

func init() {
	TerraformCmd.AddCommand(GenerateCmd)
	TerraformCmd.AddCommand(PlanCmd)
	TerraformCmd.AddCommand(ApplyCmd)
	TerraformCmd.AddCommand(StateCmd)
}
