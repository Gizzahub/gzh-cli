// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package devenv

import (
	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
	"github.com/Gizzahub/gzh-cli/internal/app"
)

type devEnvCmdProvider struct {
	appCtx *app.AppContext
}

func (p devEnvCmdProvider) Command() *cobra.Command {
	return NewDevEnvCmd(p.appCtx)
}

func RegisterDevEnvCmd(appCtx *app.AppContext) {
	registry.Register(devEnvCmdProvider{appCtx: appCtx})
}
