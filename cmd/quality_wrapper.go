// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package cmd

import (
	qualitypkg "github.com/Gizzahub/gzh-cli-quality"
	"github.com/Gizzahub/gzh-cli/cmd/registry"
	"github.com/Gizzahub/gzh-cli/internal/app"
	"github.com/spf13/cobra"
)

// NewQualityCmd creates the quality command by wrapping gzh-cli-quality.
// This delegates all quality tool functionality to the external gzh-cli-quality package,
// avoiding code duplication and ensuring consistency with the standalone quality CLI.
//
// The wrapper allows customization of the command metadata while preserving all
// subcommands and functionality from the gzh-cli-quality implementation.
func NewQualityCmd(appCtx *app.AppContext) *cobra.Command {
	_ = appCtx // Reserved for future context integration

	// Use the external quality tool implementation
	cmd := qualitypkg.NewQualityCmd()

	// Customize command metadata for gzh-cli context
	cmd.Use = "quality"
	cmd.Short = "통합 코드 품질 도구 (포매팅 + 린팅)"
	cmd.Aliases = []string{"q", "qual"}

	return cmd
}

// qualityCmdProvider implements the command provider interface for quality tools.
type qualityCmdProvider struct {
	appCtx *app.AppContext
}

func (p qualityCmdProvider) Command() *cobra.Command {
	return NewQualityCmd(p.appCtx)
}

func (p qualityCmdProvider) Metadata() registry.CommandMetadata {
	return registry.CommandMetadata{
		Name:         "quality",
		Category:     registry.CategoryQuality,
		Version:      "1.0.0",
		Priority:     40,
		Experimental: false,
		Dependencies: []string{}, // 언어별 도구들은 동적으로 확인
		Tags:         []string{"quality", "lint", "format", "code", "check"},
		Lifecycle:    registry.LifecycleStable,
	}
}

// RegisterQualityCmd registers the quality command with the command registry.
func RegisterQualityCmd(appCtx *app.AppContext) {
	registry.Register(qualityCmdProvider{appCtx: appCtx})
}
