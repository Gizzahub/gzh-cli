package gitlabci

import (
	"github.com/spf13/cobra"
)

// GitLabCICmd represents the gitlab-ci command
var GitLabCICmd = &cobra.Command{
	Use:   "gitlab-ci",
	Short: "GitLab CI/CD pipeline generation and management",
	Long: `GitLab CI/CD pipeline generation and management tools.

Generate and manage GitLab CI/CD pipelines with:
- .gitlab-ci.yml template generation
- Pipeline stage templates
- GitLab Runner configuration
- Multi-environment deployments
- Pipeline optimization and caching
- Security scanning integration

Available commands:
  generate    Generate GitLab CI/CD pipelines
  validate    Validate pipeline configurations
  deploy      Deploy and manage pipelines`,
	Aliases: []string{"gitlab", "glci"},
}

func init() {
	GitLabCICmd.AddCommand(GenerateCmd)
	GitLabCICmd.AddCommand(ValidateCmd)
	GitLabCICmd.AddCommand(DeployCmd)
}
