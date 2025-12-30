// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

//go:build !pm_external
// +build !pm_external

package cmd

import (
	"github.com/gizzahub/gzh-cli/internal/app"
)

// RegisterPMCmd is a stub when pm_external is not enabled.
// The actual PM command integration requires the external library.
func RegisterPMCmd(appCtx *app.AppContext) {
	// PM command is disabled (external library not available)
	// To enable, build with: go build -tags pm_external
	_ = appCtx
}
