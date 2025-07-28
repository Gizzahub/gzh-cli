// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package git

import (
	"context"
	"fmt"

	"github.com/gizzahub/gzh-manager-go/internal/git/sync"
)

// runSync executes the repository synchronization operation.
func runSync(ctx context.Context, opts sync.Options) error {
	// Validate options
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	// Parse source and destination targets
	sourceTarget, err := opts.GetSourceTarget()
	if err != nil {
		return fmt.Errorf("invalid source target: %w", err)
	}

	destTarget, err := opts.GetDestinationTarget()
	if err != nil {
		return fmt.Errorf("invalid destination target: %w", err)
	}

	// Get provider instances
	sourceProvider, err := getGitProvider(sourceTarget.Provider, sourceTarget.Org)
	if err != nil {
		return fmt.Errorf("failed to get source provider: %w", err)
	}

	destProvider, err := getGitProvider(destTarget.Provider, destTarget.Org)
	if err != nil {
		return fmt.Errorf("failed to get destination provider: %w", err)
	}

	// Create sync engine
	engine := sync.NewSyncEngine(sourceProvider, destProvider, opts)

	// Execute synchronization
	if err := engine.Sync(ctx); err != nil {
		return fmt.Errorf("synchronization failed: %w", err)
	}

	return nil
}
