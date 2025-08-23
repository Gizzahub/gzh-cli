// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package vpn

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// NewCmd creates the main VPN command that aggregates VPN hierarchy, profile, and failover commands
func NewCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vpn",
		Short: "VPN connection and management",
		Long: `VPN connection and management including hierarchical VPN, profile management, and failover.

This command provides comprehensive VPN management tools with:
- Hierarchical VPN connection management
- VPN profile and priority configuration
- Automatic VPN failover and backup connections
- Multi-VPN environment support

Examples:
  # VPN hierarchy management
  gz net-env vpn vpn-hierarchy show
  gz net-env vpn vpn-hierarchy connect --root corp-vpn

  # VPN profile management  
  gz net-env vpn vpn-profile list
  gz net-env vpn vpn-profile create office --network "Office WiFi"

  # VPN failover management
  gz net-env vpn vpn-failover start
  gz net-env vpn vpn-failover backup add --primary corp-vpn`,
		SilenceUsage: true,
	}

	// Add VPN hierarchy command
	cmd.AddCommand(NewHierarchyCmd(logger, configDir))

	// Add VPN profile command
	cmd.AddCommand(NewProfileCmd(logger, configDir))

	// Add VPN failover command
	cmd.AddCommand(NewFailoverCmd(logger, configDir))

	return cmd
}
