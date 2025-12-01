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

func (p devEnvCmdProvider) Metadata() registry.CommandMetadata {
	return registry.CommandMetadata{
		Name:         "dev-env",
		Category:     registry.CategoryDevelopment,
		Version:      "1.0.0",
		Priority:     20,
		Experimental: false,
		Dependencies: []string{}, // 동적으로 확인 (aws, gcloud, docker 등)
		Tags:         []string{"development", "environment", "aws", "gcp", "azure", "docker", "kubernetes"},
		Lifecycle:    registry.LifecycleStable,
	}
}

func RegisterDevEnvCmd(appCtx *app.AppContext) {
	registry.Register(devEnvCmdProvider{appCtx: appCtx})
}
