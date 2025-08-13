// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-manager-go/internal/netenv/tui"
)

// newTUICmd creates a new TUI command for interactive network environment management.
func newTUICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Launch interactive TUI dashboard for network environment management",
		Long: `Launch an interactive Terminal User Interface (TUI) dashboard for managing
network environments. This provides a real-time view of all network components
including WiFi, VPN, DNS, proxy, and Docker configurations with quick switching
between different network profiles.

The TUI includes:
- Real-time network status monitoring
- Interactive network profile switching
- VPN connection management
- DNS and proxy configuration
- Network performance monitoring
- Quick actions and keyboard shortcuts

Navigation:
  ↑/k, ↓/j     Navigate up/down
  ←/h, →/l     Navigate left/right
  Enter        Select/confirm action
  Esc          Go back to previous view
  q/Q          Quit (from dashboard)
  r            Refresh network status

Network Actions:
  s            Switch network profile
  v            VPN toggle/manager
  d            DNS settings
  p            Proxy toggle
  c            Quick connect VPN
  x            Quick disconnect VPN
  m            Network monitoring view

Views:
  P            Settings/preferences
  /            Search networks/profiles
  ?            Toggle help

Examples:
  # Launch the network TUI dashboard
  gz net-env tui

  # Launch TUI with verbose logging (for debugging)
  gz net-env tui --verbose`,
		SilenceUsage: true,
		RunE:         runTUI,
	}

	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose logging for debugging")

	return cmd
}

// runTUI executes the TUI command.
func runTUI(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Set up context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create TUI model
	model := tui.NewModel(ctx)

	// Configure tea options
	var opts []tea.ProgramOption
	if verbose {
		// Enable debug logging if verbose is set
		opts = append(opts, tea.WithAltScreen())
	} else {
		// Normal operation with alt screen
		opts = append(opts, tea.WithAltScreen())
	}

	// Create and run the TUI program
	p := tea.NewProgram(model, opts...)

	// Handle interrupt signals gracefully
	go func() {
		<-ctx.Done()
		p.Quit()
	}()

	// Run the program
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run network TUI: %w", err)
	}

	// Check if the program exited due to an error
	if m, ok := finalModel.(*tui.Model); ok {
		if verbose {
			fmt.Fprintf(os.Stderr, "Network TUI exited successfully\n")
		}
		_ = m // Use the final model if needed for cleanup
	}

	return nil
}
