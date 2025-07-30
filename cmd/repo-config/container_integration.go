// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repoconfig

import (
	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-manager-go/internal/container"
)

// NewRepoConfigCmdWithContainer creates a new repo-config command with dependency injection support.
func NewRepoConfigCmdWithContainer(containerInstance *container.ContextualContainer) *cobra.Command {
	// For now, use the regular command but store container reference for future use
	cmd := NewRepoConfigCmd()

	// Store container in command context for later use
	cmd.SetContext(containerInstance.GetContext())

	// Add container-specific functionality if needed
	// This is where we would wire up container-managed dependencies

	return cmd
}
