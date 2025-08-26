// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package ide

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/ide/fixsync"
	"github.com/Gizzahub/gzh-cli/cmd/ide/list"
	"github.com/Gizzahub/gzh-cli/cmd/ide/monitor"
	"github.com/Gizzahub/gzh-cli/cmd/ide/open"
	"github.com/Gizzahub/gzh-cli/cmd/ide/scan"
	"github.com/Gizzahub/gzh-cli/cmd/ide/status"
	"github.com/Gizzahub/gzh-cli/internal/app"
)

// NewIDECmd creates the IDE subcommand for monitoring and managing IDE configuration changes.
func NewIDECmd(ctx context.Context, appCtx *app.AppContext) *cobra.Command {
	_ = appCtx
	cmd := &cobra.Command{
		Use:   "ide",
		Short: "Monitor and manage IDE configuration changes",
		Long: `Monitor and manage IDE configuration changes, particularly JetBrains products.

This command provides comprehensive IDE management capabilities:
- IDE detection and scanning across multiple platforms
- Status monitoring with detailed information
- Easy IDE launching from command line
- Real-time monitoring of JetBrains settings directories
- Cross-platform support for Linux, macOS, and Windows
- Settings synchronization issue detection and fixes

Supported IDEs:
- JetBrains family: IntelliJ IDEA, PyCharm, WebStorm, GoLand, CLion, etc.
- VS Code family: VS Code, VS Code Insiders, Cursor, VSCodium
- Other editors: Sublime Text, Vim, Neovim, Emacs

Examples:
  # Scan for installed IDEs
  gz ide scan

  # Show IDE status information
  gz ide status

  # Open an IDE in current directory
  gz ide open pycharm

  # Open IDE in specific directory
  gz ide open code ~/projects/myapp

  # Monitor all JetBrains settings
  gz ide monitor

  # Monitor specific product
  gz ide monitor --product IntelliJIdea2023.2

  # Fix settings sync issues
  gz ide fix-sync

  # List detected JetBrains installations
  gz ide list`,
		SilenceUsage: true,
	}

	cmd.AddCommand(monitor.NewCmd(ctx))
	cmd.AddCommand(list.NewCmd())
	cmd.AddCommand(fixsync.NewCmd())
	cmd.AddCommand(scan.NewCmd())
	cmd.AddCommand(status.NewCmd())
	cmd.AddCommand(open.NewCmd())

	return cmd
}
