// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package app provides application bootstrapping and lifecycle management.
// It handles signal management, graceful shutdown, and application initialization
// to keep the main function minimal and focused on bootstrapping.
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gizzahub/gzh-manager-go/cmd"
	"github.com/gizzahub/gzh-manager-go/internal/container"
)

// Runner handles application lifecycle and signal management.
type Runner struct {
	version   string
	container *container.Container
}

// NewRunner creates a new application runner with the specified version.
func NewRunner(version string) *Runner {
	// Create and configure the dependency injection container
	appContainer := container.NewContainerBuilder().
		WithHTTPTimeout(30 * time.Second).
		WithMetrics(true).
		WithHealthChecks(true).
		Build()

	return &Runner{
		version:   version,
		container: appContainer,
	}
}

// NewRunnerWithContainer creates a new runner with a custom container.
func NewRunnerWithContainer(version string, appContainer *container.Container) *Runner {
	return &Runner{
		version:   version,
		container: appContainer,
	}
}

// Run starts the application with proper signal handling and graceful shutdown.
func (r *Runner) Run() error {
	// Create a context that will be canceled on interrupt signals
	ctx, cancel := r.setupGracefulShutdown()
	defer cancel()

	// Ensure container cleanup on exit
	defer func() {
		if err := r.container.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to clean up container: %v\n", err)
		}
	}()

	// Create contextual container for command execution
	contextualContainer := container.NewContextualContainer(ctx, r.container)

	// Execute the root command with context and container
	if err := cmd.ExecuteWithContainer(ctx, r.version, contextualContainer); err != nil {
		return fmt.Errorf("application execution failed: %w", err)
	}

	return nil
}

// GetContainer returns the dependency injection container.
func (r *Runner) GetContainer() *container.Container {
	return r.container
}

// setupGracefulShutdown configures signal handling for graceful shutdown.
func (r *Runner) setupGracefulShutdown() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Fprintf(os.Stderr, "\nReceived interrupt signal, shutting down gracefully...\n")
		cancel()
	}()

	return ctx, cancel
}

// GetVersion returns the application version.
func (r *Runner) GetVersion() string {
	return r.version
}
