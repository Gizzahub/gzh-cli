// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package devenv

import (
	"github.com/spf13/cobra"
)

func newKubeconfigCmd() *cobra.Command {
	baseCmd := NewBaseCommand(
		"kubeconfig",
		"yaml",
		".kube/config",
		`This command helps you backup and restore Kubernetes configuration files, which contain:
- Kubernetes cluster configurations
- User authentication credentials
- Context and namespace settings
- Certificate and token information

This is useful when:
- Setting up new development machines
- Switching between different Kubernetes environments
- Backing up Kubernetes configurations before changes
- Managing multiple Kubernetes configurations for different projects`,
		[]string{
			"# Save current kubeconfig with a name",
			"gz dev-env kubeconfig save --name production",
			"",
			"# Save with description",
			"gz dev-env kubeconfig save --name staging --description \"Staging K8s config\"",
			"",
			"# Load a saved kubeconfig",
			"gz dev-env kubeconfig load --name production",
			"",
			"# List all saved configurations",
			"gz dev-env kubeconfig list",
			"",
			"# Save from specific path",
			"gz dev-env kubeconfig save --name custom --config-path /path/to/kubeconfig",
		},
	)

	return baseCmd.CreateMainCommand()
}
