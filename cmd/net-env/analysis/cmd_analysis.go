// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package analysis

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// NewCmd creates the main analysis command that aggregates network analysis, topology, and routing commands
func NewCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analysis",
		Short: "Network analysis and topology management",
		Long: `Network analysis and topology management including performance analysis, network topology discovery, and optimal routing.

This command provides comprehensive network analysis tools with:
- Real-time network performance analysis
- Network topology discovery and visualization
- Optimal routing analysis and recommendations
- Bandwidth utilization and bottleneck detection

Examples:
  # Network performance analysis
  gz net-env analysis network-analysis latency --duration 10m
  gz net-env analysis network-analysis comprehensive

  # Network topology analysis
  gz net-env analysis network-topology analyze
  gz net-env analysis network-topology export --format dot

  # Optimal routing analysis
  gz net-env analysis optimal-routing analyze --destination 8.8.8.8
  gz net-env analysis optimal-routing discover --targets google.com`,
		SilenceUsage: true,
	}

	// Add network analysis command
	cmd.AddCommand(NewNetworkAnalysisCmd(logger, configDir))

	// Add network topology command
	cmd.AddCommand(NewNetworkTopologyCmd(logger, configDir))

	// Add optimal routing command
	cmd.AddCommand(NewOptimalRoutingCmd(logger, configDir))

	return cmd
}
