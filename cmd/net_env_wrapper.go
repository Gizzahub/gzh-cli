// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

//go:build netenv_external
// +build netenv_external

package cmd

import (
	"github.com/spf13/cobra"

	netenvcmd "github.com/gizzahub/gzh-cli-net-env/cmd/netenv"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
	"github.com/Gizzahub/gzh-cli/internal/app"
)

// NewNetEnvCmd creates the network environment command by wrapping gzh-cli-net-env.
// This delegates all net-env functionality to the external gzh-cli-net-env package,
// avoiding code duplication and ensuring consistency with the standalone net-env CLI.
//
// The wrapper allows customization of the command metadata while preserving all
// subcommands and functionality from the gzh-cli-net-env implementation.
func NewNetEnvCmd(appCtx *app.AppContext) *cobra.Command {
	_ = appCtx // Reserved for future app context integration

	// Use the external net-env implementation
	cmd := netenvcmd.NewRootCmd()

	// Customize command metadata for gzh-cli context
	cmd.Use = "net-env"
	cmd.Short = "Manage network environment configurations"
	cmd.Aliases = []string{"ne", "netenv"}
	cmd.Long = `Manage network environment transitions on-demand.

This command helps you manage network configurations when
switching between different network environments. It provides:
- Network status checking (WiFi, VPN, DNS, Proxy)
- Real-time network monitoring with dashboard
- Network profile management and automatic switching
- Cross-platform support (macOS and Linux)

This is useful when:
- Moving between different network environments (home, office, cafe)
- Switching VPN connections based on network conditions
- Monitoring network status in real-time
- Automatically applying network profiles based on SSID

Examples:
  # Show current network status
  gz net-env status

  # Monitor network changes in real-time
  gz net-env watch

  # List configured network profiles
  gz net-env profile list

  # Show profile details
  gz net-env profile show office

For detailed configuration, see: ~/.config/gzh-cli/net-env.json`

	return cmd
}

// netEnvCmdProvider implements the command provider interface for net-env.
type netEnvCmdProvider struct {
	appCtx *app.AppContext
}

func (p netEnvCmdProvider) Command() *cobra.Command {
	return NewNetEnvCmd(p.appCtx)
}

func (p netEnvCmdProvider) Metadata() registry.CommandMetadata {
	return registry.CommandMetadata{
		Name:         "net-env",
		Category:     registry.CategoryNetwork,
		Version:      "2.0.0",
		Priority:     50,
		Experimental: false,
		Dependencies: []string{}, // Dynamically checks (networksetup, nmcli, etc.)
		Tags:         []string{"network", "environment", "wifi", "vpn", "dns", "proxy"},
		Lifecycle:    registry.LifecycleStable,
	}
}

// RegisterNetEnvCmd registers the network environment command with the command registry.
func RegisterNetEnvCmd(appCtx *app.AppContext) {
	registry.Register(netEnvCmdProvider{appCtx: appCtx})
}
