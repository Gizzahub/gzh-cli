// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package devenv

import "github.com/spf13/cobra"

// NewDevEnvCmd creates the development environment command.
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
- AWS profile management with SSO support
- Google Cloud (GCloud) configurations and credentials
- GCP project management and gcloud configurations
- Azure subscription management with multi-tenant support
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

  # Manage AWS profiles with SSO support
  gz dev-env aws-profile list
  gz dev-env aws-profile switch production
  gz dev-env aws-profile login production

  # Save current gcloud config
  gz dev-env gcloud save --name production

  # Save current gcloud credentials
  gz dev-env gcloud-credentials save --name production

  # Manage GCP projects and configurations
  gz dev-env gcp-project list
  gz dev-env gcp-project switch my-project-id
  gz dev-env gcp-project config create --name prod --project my-prod-project

  # Manage Azure subscriptions and configurations
  gz dev-env azure-subscription list
  gz dev-env azure-subscription switch my-subscription-id
  gz dev-env azure-subscription show
  gz dev-env azure-subscription login
  gz dev-env azure-subscription validate

  # Save current SSH config
  gz dev-env ssh save --name production

  # Load a saved configuration
  gz dev-env kubeconfig load --name my-cluster

  # List saved configurations
  gz dev-env kubeconfig list`,
		SilenceUsage: true,
	}

	// Add switch-all command for unified environment switching
	cmd.AddCommand(newSwitchAllCmd())

	cmd.AddCommand(newKubeconfigCmd())
	cmd.AddCommand(newDockerCmd())
	cmd.AddCommand(newAwsCmd())
	cmd.AddCommand(newAwsCredentialsCmd())
	cmd.AddCommand(newAWSProfileCmd())
	cmd.AddCommand(newGcloudCmd())
	cmd.AddCommand(newGcloudCredentialsCmd())
	cmd.AddCommand(newGCPProjectCmd())
	cmd.AddCommand(newAzureSubscriptionCmd())
	cmd.AddCommand(newSSHCmd())

	return cmd
}
