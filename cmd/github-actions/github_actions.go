package githubactions

import (
	"github.com/spf13/cobra"
)

// GitHubActionsCmd represents the github-actions command
var GitHubActionsCmd = &cobra.Command{
	Use:   "github-actions",
	Short: "GitHub Actions workflow generation and management",
	Long: `GitHub Actions workflow generation and management tools.

Generate and manage GitHub Actions workflows with:
- Workflow template library
- Reusable action development
- Secret management and security scanning
- CI/CD pipeline automation
- Multi-platform builds
- Deployment automation

Available commands:
  generate    Generate GitHub Actions workflows
  validate    Validate workflow configurations
  deploy      Deploy and manage workflows`,
	Aliases: []string{"gha", "actions"},
}

func init() {
	GitHubActionsCmd.AddCommand(GenerateCmd)
	GitHubActionsCmd.AddCommand(ValidateCmd)
	GitHubActionsCmd.AddCommand(DeployCmd)
}
