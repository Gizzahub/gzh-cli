// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	configservice "github.com/gizzahub/gzh-manager-go/internal/config"
	configpkg "github.com/gizzahub/gzh-manager-go/pkg/config"
	"github.com/spf13/cobra"
)

// newWatchCmd creates the config watch subcommand.
func newWatchCmd() *cobra.Command {
	var (
		configFile string
		verbose    bool
		interval   time.Duration
	)

	cmd := &cobra.Command{
		Use:   "watch [config-file]",
		Short: "Watch configuration file for changes and reload automatically",
		Long: `Watch gzh.yaml configuration file for changes and automatically reload.

This command demonstrates configuration hot-reloading functionality:
- Monitors the configuration file for changes using file system events
- Automatically reloads and validates configuration when changes are detected
- Shows current configuration status and any validation errors
- Gracefully handles interruption signals (Ctrl+C)

Examples:
  gz config watch                     # Watch default gzh.yaml
  gz config watch my-config.yaml     # Watch specific file
  gz config watch --verbose          # Verbose output with detailed changes
  gz config watch --interval 5s      # Custom status display interval`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			// Determine config file path
			if len(args) > 0 {
				configFile = args[0]
			}

			return watchConfig(configFile, verbose, interval)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Configuration file path (default: auto-detect)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output with detailed change information")
	cmd.Flags().DurationVar(&interval, "interval", 30*time.Second, "Status display interval")

	return cmd
}

// watchConfig implements the configuration watching functionality.
func watchConfig(configFile string, verbose bool, interval time.Duration) error { //nolint:gocognit // Complex file watching logic with multiple state checks
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create configuration service with watching enabled
	options := configservice.DefaultConfigServiceOptions()
	options.WatchEnabled = true
	options.ValidationEnabled = true

	service, err := configservice.NewConfigService(options)
	if err != nil {
		return fmt.Errorf("failed to create configuration service: %w", err)
	}

	// Load initial configuration
	if configFile == "" {
		var findErr error

		configFile, findErr = findConfigFile()
		if findErr != nil {
			return fmt.Errorf("failed to find configuration file: %w", findErr)
		}
	}

	fmt.Printf("🔍 Loading configuration from: %s\n", configFile)

	config, err := service.LoadConfiguration(ctx, configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	fmt.Printf("✅ Configuration loaded successfully\n")

	if verbose {
		printConfigSummary(config)
	}

	// Set up configuration change callback
	changeCount := 0
	lastChangeTime := time.Now()

	changeCallback := func(newConfig *configpkg.UnifiedConfig) {
		changeCount++
		lastChangeTime = time.Now()

		fmt.Printf("\n🔄 Configuration changed (change #%d at %s)\n",
			changeCount, lastChangeTime.Format("15:04:05"))

		// Validate the new configuration
		validationResult := service.GetValidationResult()
		if validationResult != nil {
			if validationResult.IsValid {
				fmt.Printf("✅ Configuration reloaded and validated successfully\n")

				if len(validationResult.Warnings) > 0 {
					fmt.Printf("⚠️  %d warnings found:\n", len(validationResult.Warnings))

					for _, warning := range validationResult.Warnings {
						fmt.Printf("   - [%s] %s\n", warning.Field, warning.Message)
					}
				}
			} else {
				fmt.Printf("❌ Configuration validation failed with %d errors:\n", len(validationResult.Errors))

				for _, err := range validationResult.Errors {
					fmt.Printf("   - [%s] %s\n", err.Field, err.Message)
				}
			}
		}

		if verbose {
			printConfigSummary(newConfig)
		}
	}

	// Start watching for configuration changes
	fmt.Printf("👀 Watching for configuration changes...\n")
	fmt.Printf("   Press Ctrl+C to stop watching\n\n")

	err = service.WatchConfiguration(ctx, changeCallback)
	if err != nil {
		return fmt.Errorf("failed to start watching configuration: %w", err)
	}
	defer service.StopWatching()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Set up status ticker
	statusTicker := time.NewTicker(interval)
	defer statusTicker.Stop()

	startTime := time.Now()

	// Main watch loop
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("\n🛑 Context canceled, stopping watch\n")
			return nil

		case sig := <-sigChan:
			fmt.Printf("\n🛑 Received signal %v, stopping watch\n", sig)
			cancel()

			return nil

		case <-statusTicker.C:
			uptime := time.Since(startTime)
			fmt.Printf("📊 Status: watching %s | uptime: %v | changes: %d",
				configFile, uptime.Round(time.Second), changeCount)

			if changeCount > 0 {
				fmt.Printf(" | last change: %v ago",
					time.Since(lastChangeTime).Round(time.Second))
			}

			fmt.Printf("\n")
		}
	}
}

// printConfigSummary prints a summary of the configuration.
func printConfigSummary(config *configpkg.UnifiedConfig) {
	fmt.Printf("📋 Configuration Summary:\n")
	fmt.Printf("   Version: %s\n", config.Version)
	fmt.Printf("   Default Provider: %s\n", config.DefaultProvider)
	fmt.Printf("   Providers: %d\n", len(config.Providers))

	for providerName, provider := range config.Providers {
		fmt.Printf("   📌 %s: %d organizations\n", providerName, len(provider.Organizations))

		if provider.APIURL != "" {
			fmt.Printf("      API URL: %s\n", provider.APIURL)
		}
	}

	if config.Global != nil {
		fmt.Printf("   🌐 Global Settings:\n")

		if config.Global.CloneBaseDir != "" {
			fmt.Printf("      Clone Base Dir: %s\n", config.Global.CloneBaseDir)
		}

		if config.Global.DefaultStrategy != "" {
			fmt.Printf("      Default Strategy: %s\n", config.Global.DefaultStrategy)
		}
	}

	fmt.Printf("\n")
}
