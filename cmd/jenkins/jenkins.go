package jenkins

import (
	"github.com/spf13/cobra"
)

// JenkinsCmd represents the jenkins command
var JenkinsCmd = &cobra.Command{
	Use:   "jenkins",
	Short: "Jenkins pipeline generation and management",
	Long: `Jenkins pipeline generation and management tools.

Generate and manage Jenkins pipelines with:
- Jenkinsfile generation for various project types
- Shared library development and deployment
- Plugin management and automated installation
- Pipeline configuration and optimization
- Multi-branch pipeline support
- Blue Ocean integration

Available commands:
  generate    Generate Jenkins pipelines and configurations
  validate    Validate pipeline configurations
  deploy      Deploy and manage pipelines`,
	Aliases: []string{"jen"},
}

func init() {
	JenkinsCmd.AddCommand(GenerateCmd)
	JenkinsCmd.AddCommand(ValidateCmd)
	JenkinsCmd.AddCommand(DeployCmd)
}
