// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"github.com/spf13/cobra"

	netenvlib "github.com/gizzahub/gzh-cli-net-env/cmd/netenv"
)

// LibraryNetEnvCmd creates the network environment command using the gzh-cli-net-env library.
// This replaces the older, heavier implementation with a library-based approach.
//
// Library features:
//   - status: Display network status dashboard
//   - watch: Continuously monitor network changes
//   - profile: List and show profile configurations
func LibraryNetEnvCmd() *cobra.Command {
	return netenvlib.NewRootCmd()
}
