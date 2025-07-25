// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repoconfig

import (
	"github.com/spf13/cobra"
)

// MakeDeprecatedCommand is no longer needed as repo-config is kept.
func MakeDeprecatedCommand(cmd *cobra.Command) *cobra.Command {
	// repo-config is no longer deprecated since repo-sync was removed
	return cmd
}
