// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package profile

import (
	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/cmd/registry"
	"github.com/gizzahub/gzh-cli/internal/app"
)

type profileCmdProvider struct {
	appCtx *app.AppContext
}

func (p profileCmdProvider) Command() *cobra.Command {
	return NewProfileCmd(p.appCtx)
}

func (p profileCmdProvider) Metadata() registry.CommandMetadata {
	return registry.CommandMetadata{
		Name:         "profile",
		Category:     registry.CategoryUtility,
		Version:      "1.0.0",
		Priority:     60,
		Experimental: false,
		Dependencies: []string{},
		Tags:         []string{"profile", "performance", "pprof", "cpu", "memory", "benchmark"},
		Lifecycle:    registry.LifecycleStable,
	}
}

func RegisterProfileCmd(appCtx *app.AppContext) {
	registry.Register(profileCmdProvider{appCtx: appCtx})
}
