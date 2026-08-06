// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package tui

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	netenvtui "github.com/gizzahub/gzh-cli/internal/netenv/tui"
)

// NewCmd creates a new TUI command for interactive network environment management.
func NewCmd() *cobra.Command {
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

	// 실행 맥락에서 갈라 쓴다. 예전에는 context.Background()에서 갈랐는데,
	// 그러면 아래 "Handle interrupt signals gracefully" 고리가 제 때 돌 수
	// 없다 -- 그 맥락을 끊는 것은 defer한 cancel뿐이고 그것은 이 함수가
	// 끝나야, 즉 p.Run()이 이미 돌아온 뒤에야 불린다. 신호를 받아 TUI를
	// 닫으려고 둔 자리가 TUI가 닫힌 뒤에 도는 셈이었다.
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	// Create TUI model
	model := netenvtui.NewModel(ctx)

	// Configure tea options
	//
	// verbose로 갈라 두었지만 두 갈래가 같은 일을 했다("Enable debug logging
	// if verbose is set"이라고 적혀 있을 뿐 실제로 다른 것은 없었다).
	// verbose는 아래 종료 후 보고에서만 쓴다. dev-env의 같은 자리는 이미
	// 갈래 없이 한 줄이다.
	opts := []tea.ProgramOption{tea.WithAltScreen()}

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
	if m, ok := finalModel.(*netenvtui.Model); ok {
		if verbose {
			fmt.Fprintf(os.Stderr, "Network TUI exited successfully\n")
		}
		_ = m // Use the final model if needed for cleanup
	}

	return nil
}
