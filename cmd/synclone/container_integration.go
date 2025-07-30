// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package synclone

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-manager-go/internal/container"
)

// NewSyncCloneCmdWithContainer creates a new synclone command with dependency injection support.
func NewSyncCloneCmdWithContainer(ctx context.Context, containerInstance *container.ContextualContainer) *cobra.Command {
	// For now, use the regular command but store container reference for future use
	cmd := NewSyncCloneCmd(ctx)

	// Store container in command context for later use
	cmd.SetContext(containerInstance.GetContext())

	// Add container-specific functionality if needed
	// This is where we would wire up container-managed dependencies

	return cmd
}
