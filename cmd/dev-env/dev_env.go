package devenv

import "github.com/spf13/cobra"

func NewDevEnvCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dev-env",
		Short: "Manage development environment configurations",
		Long: `Save and load development environment configurations.

This command helps you backup, restore, and manage various development 
environment configurations including:
- Kubernetes configurations (kubeconfig)
- Docker configurations
- AWS configurations and credentials
- Google Cloud (GCloud) configurations and credentials
- SSH configurations
- And more...

This is useful when setting up new development machines, switching between
projects, or maintaining consistent environments across multiple machines.

Examples:
  # Save current kubeconfig
  gz dev-env kubeconfig save --name my-cluster
  
  # Save current Docker config
  gz dev-env docker save --name production
  
  # Save current AWS config
  gz dev-env aws save --name production
  
  # Save current AWS credentials
  gz dev-env aws-credentials save --name production
  
  # Save current gcloud config
  gz dev-env gcloud save --name production
  
  # Save current gcloud credentials
  gz dev-env gcloud-credentials save --name production
  
  # Save current SSH config
  gz dev-env ssh save --name production
  
  # Load a saved configuration
  gz dev-env kubeconfig load --name my-cluster
  
  # List saved configurations
  gz dev-env kubeconfig list`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newKubeconfigCmd())
	cmd.AddCommand(newDockerCmd())
	cmd.AddCommand(newAwsCmd())
	cmd.AddCommand(newAwsCredentialsCmd())
	cmd.AddCommand(newGcloudCmd())
	cmd.AddCommand(newGcloudCredentialsCmd())
	cmd.AddCommand(newSshCmd())

	return cmd
}
