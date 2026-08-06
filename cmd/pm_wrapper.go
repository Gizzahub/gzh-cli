// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

//go:build pm_external

package cmd

import (
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
// ctx는 받지 않는다. "Reserved for future context integration"이라고 적어 두고
// 쓰지 않았는데, 정작 넣어 주던 값은 register.go의 context.Background()라
// 취소되지도 않는 것이었다. 하위 명령이 맥락을 필요로 하면 각 RunE에서
// cmd.Context()를 쓰면 된다 -- root의 ExecuteContext가 넣어 준 진짜 맥락이다.
func NewPMCmd(appCtx *app.AppContext) *cobra.Command {
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
	return NewPMCmd(p.appCtx)
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
