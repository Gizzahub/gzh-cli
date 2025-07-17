// Package main is the entry point for the gzh-manager-go CLI application.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gizzahub/gzh-manager-go/cmd"
)

var (
	version = "dev"
)

func main() {
	// Create a context that will be cancelled on interrupt signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Fprintf(os.Stderr, "\nReceived interrupt signal, shutting down gracefully...\n")
		cancel()
	}()

	if err := cmd.Execute(ctx, version); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}
