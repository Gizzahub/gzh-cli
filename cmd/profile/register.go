// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package profile

import (
	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
	"github.com/Gizzahub/gzh-cli/internal/app"
)

type profileCmdProvider struct {
	appCtx *app.AppContext
}

func (p profileCmdProvider) Command() *cobra.Command {
	return NewProfileCmd(p.appCtx)
}

func RegisterProfileCmd(appCtx *app.AppContext) {
	registry.Register(profileCmdProvider{appCtx: appCtx})
}
