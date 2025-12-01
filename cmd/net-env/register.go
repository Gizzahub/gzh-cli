// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
	"github.com/Gizzahub/gzh-cli/internal/app"
)

type netEnvCmdProvider struct {
	appCtx *app.AppContext
}

func (p netEnvCmdProvider) Command() *cobra.Command {
	return NewNetEnvCmd(context.Background(), p.appCtx)
}

func (p netEnvCmdProvider) Metadata() registry.CommandMetadata {
	return registry.CommandMetadata{
		Name:         "net-env",
		Category:     registry.CategoryNetwork,
		Version:      "1.0.0",
		Priority:     50,
		Experimental: false,
		Dependencies: []string{},
		Tags:         []string{"network", "environment", "proxy", "vpn", "switch"},
		Lifecycle:    registry.LifecycleStable,
	}
}

func RegisterNetEnvCmd(appCtx *app.AppContext) {
	registry.Register(netEnvCmdProvider{appCtx: appCtx})
}
