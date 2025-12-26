// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

//go:build devenv_external
// +build devenv_external

package cmd

import (
	"github.com/spf13/cobra"

	devenvcmd "github.com/gizzahub/gzh-cli-dev-env/cmd/devenv"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
	"github.com/Gizzahub/gzh-cli/internal/app"
)

// NewDevEnvCmd creates the development environment command by wrapping gzh-cli-dev-env.
// This delegates all dev-env functionality to the external gzh-cli-dev-env package,
// avoiding code duplication and ensuring consistency with the standalone dev-env CLI.
//
// The wrapper allows customization of the command metadata while preserving all
// subcommands and functionality from the gzh-cli-dev-env implementation.
func NewDevEnvCmd(appCtx *app.AppContext) *cobra.Command {
	_ = appCtx // Reserved for future app context integration

	// Use the external dev-env implementation
	cmd := devenvcmd.NewRootCmd()

	// Customize command metadata for gzh-cli context
	cmd.Use = "dev-env"
	cmd.Short = "Manage development environment configurations"
	cmd.Aliases = []string{"de", "devenv"}
	cmd.Long = `Save and load development environment configurations.

This command helps you backup, restore, and manage various development
environment configurations including:
- Kubernetes configurations (kubeconfig)
- Docker configurations
- AWS configurations and credentials
- AWS profile management with SSO support
- Google Cloud (GCloud) configurations and credentials
- GCP project management and gcloud configurations
- Azure subscription management with multi-tenant support
- SSH configurations
- And more...

This is useful when setting up new development machines, switching between
projects, or maintaining consistent environments across multiple machines.

Examples:
  # Show status of all development environment services
  gz dev-env status

  # Launch interactive TUI dashboard
  gz dev-env tui

  # Switch all services to a named environment
  gz dev-env switch-all --env production

  # Save current kubeconfig
  gz dev-env kubeconfig save --name my-cluster

  # Manage AWS profiles with SSO support
  gz dev-env aws-profile list
  gz dev-env aws-profile switch production

For detailed configuration, see: ~/.gzh/dev-env/`

	return cmd
}

// devEnvCmdProvider implements the command provider interface for dev-env.
type devEnvCmdProvider struct {
	appCtx *app.AppContext
}

func (p devEnvCmdProvider) Command() *cobra.Command {
	return NewDevEnvCmd(p.appCtx)
}

func (p devEnvCmdProvider) Metadata() registry.CommandMetadata {
	return registry.CommandMetadata{
		Name:         "dev-env",
		Category:     registry.CategoryDevelopment,
		Version:      "1.0.0",
		Priority:     20,
		Experimental: false,
		Dependencies: []string{}, // 동적으로 확인 (aws, gcloud, docker 등)
		Tags:         []string{"development", "environment", "aws", "gcp", "azure", "docker", "kubernetes"},
		Lifecycle:    registry.LifecycleStable,
	}
}

// RegisterDevEnvCmd registers the development environment command with the command registry.
func RegisterDevEnvCmd(appCtx *app.AppContext) {
	registry.Register(devEnvCmdProvider{appCtx: appCtx})
}
