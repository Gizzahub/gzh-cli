// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/net-env/actions"
	"github.com/Gizzahub/gzh-cli/cmd/net-env/cloud"
	"github.com/Gizzahub/gzh-cli/cmd/net-env/profile"
	"github.com/Gizzahub/gzh-cli/cmd/net-env/status"
	"github.com/Gizzahub/gzh-cli/cmd/net-env/tui"
	"github.com/Gizzahub/gzh-cli/internal/app"
)

func NewNetEnvCmd(ctx context.Context, appCtx *app.AppContext) *cobra.Command {
	_ = appCtx
	_ = ctx

	// Use library-based implementation instead of legacy local implementation
	// The gzh-cli-net-env library provides:
	// - status: Display network status dashboard
	// - watch: Continuously monitor network changes
	// - profile: List and show profile configurations
	//
	// To switch back to the legacy implementation, comment out the return below
	// and uncomment the _legacyNetEnvCmd() call instead.
	return LibraryNetEnvCmd()

	// Legacy implementation (commented out for now)
	// return _legacyNetEnvCmd(ctx)
}

// _legacyNetEnvCmd returns the legacy net-env command implementation.
// This is kept for reference and can be re-enabled if needed.
// Prefer the library-based implementation (LibraryNetEnvCmd) for new features.
func _legacyNetEnvCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "net-env",
		Short: "Manage network environment transitions",
		Long: `Manage network environment transitions on-demand.

This command helps you manage network configurations when
switching between different network environments. It provides:
- Network configuration switching (VPN, DNS, proxy, hosts)
- Network status verification
- Container environment management
- Network performance monitoring`,
		SilenceUsage: true,
	}

	// Add organized subcommand packages (working packages first)
	cmd.AddCommand(tui.NewCmd())      // Interactive TUI dashboard
	cmd.AddCommand(status.NewCmd())   // Network status (unified command)
	cmd.AddCommand(profile.NewCmd())  // Profile management + quick actions
	cmd.AddCommand(actions.NewCmd())  // Network configuration actions
	cmd.AddCommand(cloud.NewCmd(ctx)) // Cloud provider management

	return cmd
}
