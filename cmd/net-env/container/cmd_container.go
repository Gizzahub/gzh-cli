// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package container

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// NewCmd creates the main container command that aggregates Docker, Kubernetes and container detection commands
func NewCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "container",
		Short: "Container network environment management",
		Long: `Manage container network environments including Docker, Kubernetes, and container detection.

This command provides comprehensive container networking management with:
- Docker network profiles and management
- Kubernetes network policies and service mesh
- Container environment detection and monitoring
- Multi-container orchestration support

Examples:
  # Docker network management
  gz net-env container docker-network list
  gz net-env container docker-network create myapp --network mynet

  # Kubernetes network management  
  gz net-env container kubernetes-network list
  gz net-env container kubernetes-network create prod-policies

  # Container detection
  gz net-env container container-detection detect
  gz net-env container container-detection list`,
		SilenceUsage: true,
	}

	// Add Docker network command
	cmd.AddCommand(NewDockerNetworkCmd(logger, configDir))

	// Add Kubernetes network command
	cmd.AddCommand(NewKubernetesNetworkCmd(logger, configDir))

	// Add Kubernetes service mesh command
	cmd.AddCommand(NewKubernetesServiceMeshCmd(logger, configDir))

	// Add container detection command
	cmd.AddCommand(NewContainerDetectionCmd(logger, configDir))

	return cmd
}
