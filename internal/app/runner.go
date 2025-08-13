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

	"github.com/Gizzahub/gzh-manager-go/cmd"
)

// Runner handles application lifecycle and signal management.
type Runner struct {
	version string
}

// NewRunner creates a new application runner with the specified version.
func NewRunner(version string) *Runner {
	return &Runner{
		version: version,
	}
}

// Run starts the application with proper signal handling and graceful shutdown.
func (r *Runner) Run() error {
	// Create a context that will be canceled on interrupt signals
	ctx, cancel := r.setupGracefulShutdown()
	defer cancel()

	// Execute the root command with context
	if err := cmd.Execute(ctx, r.version); err != nil {
		return fmt.Errorf("application execution failed: %w", err)
	}

	return nil
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
