// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package synclone

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/cmd/registry"
	"github.com/gizzahub/gzh-cli/internal/app"
)

type syncCloneCmdProvider struct {
	appCtx *app.AppContext
}

func (p syncCloneCmdProvider) Command() *cobra.Command {
	return NewSyncCloneCmd(context.Background(), p.appCtx)
}

func (p syncCloneCmdProvider) Metadata() registry.CommandMetadata {
	return registry.CommandMetadata{
		Name:         "synclone",
		Category:     registry.CategoryConfig,
		Version:      "1.0.0",
		Priority:     30,
		Experimental: false,
		Dependencies: []string{"git"},
		Tags:         []string{"sync", "clone", "multi-platform", "github", "gitlab", "gitea"},
		Lifecycle:    registry.LifecycleStable,
	}
}

func RegisterSyncCloneCmd(appCtx *app.AppContext) {
	registry.Register(syncCloneCmdProvider{appCtx: appCtx})
}
