// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repoconfig

import (
	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/cmd/registry"
	"github.com/gizzahub/gzh-cli/internal/app"
)

type repoConfigCmdProvider struct {
	appCtx *app.AppContext
}

func (p repoConfigCmdProvider) Command() *cobra.Command {
	return NewRepoConfigCmd(p.appCtx)
}

func (p repoConfigCmdProvider) Metadata() registry.CommandMetadata {
	return registry.CommandMetadata{
		Name:         "repo-config",
		Category:     registry.CategoryConfig,
		Version:      "1.0.0",
		Priority:     35,
		Experimental: false,
		Dependencies: []string{},
		Tags:         []string{"repository", "config", "github", "settings", "template"},
		Lifecycle:    registry.LifecycleStable,
	}
}

func RegisterRepoConfigCmd(appCtx *app.AppContext) {
	registry.Register(repoConfigCmdProvider{appCtx: appCtx})
}
