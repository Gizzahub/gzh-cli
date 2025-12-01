// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"

	shellcmd "github.com/gizzahub/gzh-cli-shellforge/pkg/cmd"
	"github.com/Gizzahub/gzh-cli/cmd/registry"
	"github.com/Gizzahub/gzh-cli/internal/app"
	"github.com/spf13/cobra"
)

// NewShellforgeCmd creates the shellforge command by wrapping gzh-cli-shellforge.
// This delegates all shell configuration management functionality to the external
// gzh-cli-shellforge package, avoiding code duplication and ensuring consistency
// with the standalone shellforge CLI.
//
// The wrapper allows customization of the command metadata while preserving all
// subcommands and functionality from the gzh-cli-shellforge implementation.
func NewShellforgeCmd(ctx context.Context, appCtx *app.AppContext) *cobra.Command {
	_ = ctx    // Reserved for future context integration
	_ = appCtx // Reserved for future app context integration

	// Use the external shellforge implementation
	cmd := shellcmd.NewRootCmd()

	// Customize command metadata for gzh-cli context
	cmd.Use = "shellforge"
	cmd.Short = "Build tool for modular shell configurations"
	cmd.Long = `Build tool for modular shell configurations with automatic dependency resolution.

This command provides unified management for modular shell scripts including:
- Automatic dependency resolution via topological sort
- OS-specific filtering (macOS/Linux)
- Validation and dry-run support
- Backup/restore system with Git-backed versioning
- Template generation for common modules

Examples:
  # Validate shell configuration
  gz shellforge validate --manifest manifest.yaml --config-dir modules

  # Build shell config (dry-run to preview)
  gz shellforge build --manifest manifest.yaml --config-dir modules --os Mac --dry-run

  # Build and save to file
  gz shellforge build --manifest manifest.yaml --config-dir modules --os Mac --output ~/.zshrc

  # Create backup before changes
  gz shellforge backup --file ~/.zshrc --backup-dir ~/.shellforge/backups

For detailed configuration, see: https://github.com/gizzahub/gzh-cli-shellforge`

	return cmd
}

// shellorgeCmdProvider implements the command provider interface for shellforge.
type shellorgeCmdProvider struct {
	appCtx *app.AppContext
}

func (p shellorgeCmdProvider) Command() *cobra.Command {
	return NewShellforgeCmd(context.Background(), p.appCtx)
}

// RegisterShellforgeCmd registers the shellforge command with the command registry.
func RegisterShellforgeCmd(appCtx *app.AppContext) {
	registry.Register(shellorgeCmdProvider{appCtx: appCtx})
}
