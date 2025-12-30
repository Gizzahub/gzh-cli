// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package ide

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/cmd/registry"
	"github.com/gizzahub/gzh-cli/internal/app"
)

type ideCmdProvider struct {
	appCtx *app.AppContext
}

func (p ideCmdProvider) Command() *cobra.Command {
	return NewIDECmd(context.Background(), p.appCtx)
}

func (p ideCmdProvider) Metadata() registry.CommandMetadata {
	return registry.CommandMetadata{
		Name:         "ide",
		Category:     registry.CategoryDevelopment,
		Version:      "1.0.0",
		Priority:     25,
		Experimental: false,
		Dependencies: []string{}, // IDE는 선택적
		Tags:         []string{"ide", "jetbrains", "intellij", "vscode", "monitor", "settings"},
		Lifecycle:    registry.LifecycleStable,
	}
}

func RegisterIDECmd(appCtx *app.AppContext) {
	registry.Register(ideCmdProvider{appCtx: appCtx})
}
