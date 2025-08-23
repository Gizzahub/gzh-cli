// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package metrics

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// NewCmd creates the main metrics command that aggregates network metrics and monitoring commands
func NewCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Network metrics and monitoring",
		Long: `Network metrics and monitoring including real-time performance monitoring and metrics collection.

This command provides comprehensive network metrics tools with:
- Real-time network performance monitoring
- Bandwidth, latency, and packet loss measurement
- Historical metrics collection and analysis
- Performance trend analysis and reporting

Examples:
  # Network metrics monitoring
  gz net-env metrics network-metrics monitor
  gz net-env metrics network-metrics show
  gz net-env metrics network-metrics latency --targets 8.8.8.8

  # Unified monitoring dashboard
  gz net-env metrics monitor --performance
  gz net-env metrics monitor --changes --interval 30s`,
		SilenceUsage: true,
	}

	// Add network metrics command
	cmd.AddCommand(NewNetworkMetricsCmd(logger, configDir))

	// Add unified monitoring command
	cmd.AddCommand(NewMonitorCmd())

	return cmd
}
