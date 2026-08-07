// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package netenv

import (
	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/internal/app"
)

// NewNetEnvCmd returns the network environment command tree.
// Implementation is provided by gzh-cli-net-env (status, watch, profile).
func NewNetEnvCmd(appCtx *app.AppContext) *cobra.Command {
	_ = appCtx
	return LibraryNetEnvCmd()
}
