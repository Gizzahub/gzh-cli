package cloudformation

import (
	"github.com/spf13/cobra"
)

// CloudFormationCmd represents the cloudformation command
var CloudFormationCmd = &cobra.Command{
	Use:   "cloudformation",
	Short: "AWS CloudFormation stack management",
	Long: `AWS CloudFormation stack management tools.

Generate and manage CloudFormation stacks with:
- Stack template generation and validation
- Stack deployment and update management
- Change set creation and preview
- Parameter and output management
- Stack monitoring and rollback
- Cross-region deployment support

Available commands:
  generate    Generate CloudFormation templates
  deploy      Deploy and manage CloudFormation stacks
  validate    Validate CloudFormation templates
  changeset   Create and manage change sets`,
	Aliases: []string{"cfn"},
}

func init() {
	CloudFormationCmd.AddCommand(GenerateCmd)
	CloudFormationCmd.AddCommand(DeployCmd)
	CloudFormationCmd.AddCommand(ValidateCmd)
	CloudFormationCmd.AddCommand(ChangeSetCmd)
}
