// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

//go:build pm_external
// +build pm_external

package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/cmd/registry"
	"github.com/gizzahub/gzh-cli/internal/app"

	pmcmd "github.com/gizzahub/gzh-cli-package-manager/cmd/pm/command"
)

// NewPMCmd creates the package manager command by wrapping gzh-cli-package-manager.
// This delegates all package manager functionality to the external gzh-cli-package-manager package,
// avoiding code duplication and ensuring consistency with the standalone pm CLI.
//
// The wrapper allows customization of the command metadata while preserving all
// subcommands and functionality from the gzh-cli-package-manager implementation.
func NewPMCmd(ctx context.Context, appCtx *app.AppContext) *cobra.Command {
	_ = ctx    // Reserved for future context integration
	_ = appCtx // Reserved for future app context integration

	// Use the external package manager implementation
	cmd := pmcmd.NewRootCmd()

	// Customize command metadata for gzh-cli context
	cmd.Use = "pm"
	cmd.Short = "Package manager operations"
	cmd.Long = `Manage multiple package managers with unified commands.

This command provides centralized management for multiple package managers including:
- System package managers: brew, apt, port, yum, dnf, pacman, winget (Windows)
- Version managers: asdf, rbenv, pyenv, nvm, sdkman
- Language package managers: pip, gem, npm, cargo, go, composer

Examples:
  # Show status of all package managers
  gz pm status

  # Update all packages
  gz pm update --all

  # Bootstrap missing package managers
  gz pm bootstrap

For detailed configuration, see: ~/.gzh/pm/`

	return cmd
}

// pmCmdProvider implements the command provider interface for package manager.
type pmCmdProvider struct {
	appCtx *app.AppContext
}

func (p pmCmdProvider) Command() *cobra.Command {
	return NewPMCmd(context.Background(), p.appCtx)
}

func (p pmCmdProvider) Metadata() registry.CommandMetadata {
	return registry.CommandMetadata{
		Name:         "pm",
		Category:     registry.CategoryDevelopment,
		Version:      "1.0.0",
		Priority:     30,
		Experimental: false,
		Dependencies: []string{}, // 패키지 관리자들은 동적으로 확인
		Tags:         []string{"package", "manager", "brew", "apt", "npm", "pip", "update"},
		Lifecycle:    registry.LifecycleStable,
	}
}

// RegisterPMCmd registers the package manager command with the command registry.
func RegisterPMCmd(appCtx *app.AppContext) {
	registry.Register(pmCmdProvider{appCtx: appCtx})
}
