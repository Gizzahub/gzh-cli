// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package selfupdate

import (
	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
	"github.com/Gizzahub/gzh-cli/internal/app"
)

type selfUpdateCmdProvider struct {
	appCtx *app.AppContext
}

func (p selfUpdateCmdProvider) Command() *cobra.Command {
	return NewSelfUpdateCmd(p.appCtx)
}

func RegisterSelfUpdateCmd(appCtx *app.AppContext) {
	registry.Register(selfUpdateCmdProvider{appCtx: appCtx})
}
