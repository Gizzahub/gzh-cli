// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package repo

import (
	"context"
	"fmt"

	gitinternal "github.com/gizzahub/gzh-cli/internal/git"
	"github.com/gizzahub/gzh-cli/internal/git/sync"
)

// syncGitRunner is a test seam. When non-nil, runSync injects it into the
// SyncEngine so code sync never shells out to real git / the network.
var syncGitRunner gitinternal.GitRunner

// runSync executes the repository synchronization operation.
func runSync(ctx context.Context, opts sync.SyncOptions) error {
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

	// Create sync engine; inject test runner when the seam is set.
	engine := sync.NewSyncEngine(sourceProvider, destProvider, opts)
	if syncGitRunner != nil {
		engine.SetRunner(syncGitRunner)
	}

	// Execute synchronization
	if err := engine.Sync(ctx); err != nil {
		return fmt.Errorf("synchronization failed: %w", err)
	}

	return nil
}
