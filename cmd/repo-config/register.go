// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repoconfig

import (
	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
	"github.com/Gizzahub/gzh-cli/internal/app"
)

type repoConfigCmdProvider struct {
	appCtx *app.AppContext
}

func (p repoConfigCmdProvider) Command() *cobra.Command {
	return NewRepoConfigCmd(p.appCtx)
}

func RegisterRepoConfigCmd(appCtx *app.AppContext) {
	registry.Register(repoConfigCmdProvider{appCtx: appCtx})
}
