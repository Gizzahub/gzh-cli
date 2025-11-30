// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package pm

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
	"github.com/Gizzahub/gzh-cli/internal/app"
)

type pmCmdProvider struct {
	appCtx *app.AppContext
}

func (p pmCmdProvider) Command() *cobra.Command {
	return NewPMCmd(context.Background(), p.appCtx)
}

func RegisterPMCmd(appCtx *app.AppContext) {
	registry.Register(pmCmdProvider{appCtx: appCtx})
}
