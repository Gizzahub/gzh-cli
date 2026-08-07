// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

// Package apprunner provides application bootstrapping and lifecycle management.
// It handles signal management, graceful shutdown, and application initialization
// to keep the main function minimal and focused on bootstrapping.
//
// Interrupt exit convention (POSIX):
//
//	Ctrl+C / SIGINT → context cancel → long-running commands return an error
//	matching context.Canceled (or wrap it) → ExitCode maps that to 130
//	(128 + SIGINT). Other errors stay exit 1. Do not map unrelated failures
//	to 130. SIGTERM also cancels the root context; we use 130 for any user
//	interrupt cancel for script-friendly consistency.
package apprunner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gizzahub/gzh-cli/cmd"
)

// ExitInterrupted is the POSIX shell status for a process killed by SIGINT
// (128 + 2). Runner maps context.Canceled to this code via ExitCode.
const ExitInterrupted = 130

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

// ExitCode maps a Run error to a process exit status.
// context.Canceled (user interrupt) → ExitInterrupted (130).
// All other errors → 1. nil → 0.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	if errors.Is(err, context.Canceled) {
		return ExitInterrupted
	}
	return 1
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
