// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package selfupdate

import (
	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/cmd/registry"
	"github.com/gizzahub/gzh-cli/internal/app"
)

type selfUpdateCmdProvider struct {
	appCtx *app.AppContext
}

func (p selfUpdateCmdProvider) Command() *cobra.Command {
	return NewSelfUpdateCmd(p.appCtx)
}

func (p selfUpdateCmdProvider) Metadata() registry.CommandMetadata {
	return registry.CommandMetadata{
		Name:         "selfupdate",
		Category:     registry.CategoryUtility,
		Version:      "1.0.0",
		Priority:     90,
		Experimental: false,
		Dependencies: []string{},
		Tags:         []string{"update", "upgrade", "self-update", "version"},
		Lifecycle:    registry.LifecycleStable,
	}
}

func RegisterSelfUpdateCmd(appCtx *app.AppContext) {
	registry.Register(selfUpdateCmdProvider{appCtx: appCtx})
}
