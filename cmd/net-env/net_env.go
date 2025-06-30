package net_env

import "github.com/spf13/cobra"

func NewNetEnvCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "net-env",
		Short: "Manage network environment transitions",
		Long: `Manage network environment transitions and service monitoring.

This command helps you monitor and manage system services (daemons) when
switching between different network environments. It provides:
- Daemon/service status monitoring
- Service dependency tracking
- Network environment transition management
- System state verification

This is useful when:
- Moving between different network environments (home, office, cafe)
- Switching VPN connections that require service restarts
- Managing services that depend on network connectivity
- Verifying system state after network changes

Examples:
  # Monitor current daemon status
  gz net-env daemon list
  
  # Check specific service status
  gz net-env daemon status --service ssh
  
  # Monitor network-related services
  gz net-env daemon monitor --network-services`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newDaemonCmd())

	return cmd
}
