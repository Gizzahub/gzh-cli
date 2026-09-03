// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package cmd

import (
	"github.com/spf13/cobra"

	shellcmd "github.com/gizzahub/gzh-cli-shellforge/pkg/cmd"
	"github.com/gizzahub/gzh-cli/cmd/registry"
	"github.com/gizzahub/gzh-cli/internal/app"
)

// NewShellforgeCmd creates the shellforge command by wrapping gzh-cli-shellforge.
// This delegates all shell configuration management functionality to the external
// gzh-cli-shellforge package, avoiding code duplication and ensuring consistency
// with the standalone shellforge CLI.
//
// The wrapper allows customization of the command metadata while preserving all
// subcommands and functionality from the gzh-cli-shellforge implementation.
// ctx는 받지 않는다. "Reserved for future context integration"이라고 적어 두고
// 쓰지 않았는데, 정작 넣어 주던 값은 register.go의 context.Background()라
// 취소되지도 않는 것이었다. 하위 명령이 맥락을 필요로 하면 각 RunE에서
// cmd.Context()를 쓰면 된다 -- root의 ExecuteContext가 넣어 준 진짜 맥락이다.
func NewShellforgeCmd(appCtx *app.AppContext) *cobra.Command {
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

// shellforgeCmdProvider implements the command provider interface for shellforge.
type shellforgeCmdProvider struct {
	appCtx *app.AppContext
}

func (p shellforgeCmdProvider) Command() *cobra.Command {
	return NewShellforgeCmd(p.appCtx)
}

func (p shellforgeCmdProvider) Metadata() registry.CommandMetadata {
	return registry.CommandMetadata{
		Name:         "shellforge",
		Category:     registry.CategoryUtility,
		Version:      "1.0.0",
		Priority:     80,
		Experimental: false,
		Dependencies: []string{},
		Tags:         []string{"shell", "config", "bash", "zsh", "modules", "build"},
		Lifecycle:    registry.LifecycleStable,
	}
}

// RegisterShellforgeCmd registers the shellforge command with the command registry.
func RegisterShellforgeCmd(appCtx *app.AppContext) {
	registry.Register(shellforgeCmdProvider{appCtx: appCtx})
}
