// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package synclone

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
	"github.com/Gizzahub/gzh-cli/internal/app"
)

type syncCloneCmdProvider struct {
	appCtx *app.AppContext
}

func (p syncCloneCmdProvider) Command() *cobra.Command {
	return NewSyncCloneCmd(context.Background(), p.appCtx)
}

func RegisterSyncCloneCmd(appCtx *app.AppContext) {
	registry.Register(syncCloneCmdProvider{appCtx: appCtx})
}
