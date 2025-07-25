// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-manager-go/cmd/config"
	devenv "github.com/gizzahub/gzh-manager-go/cmd/dev-env"
	"github.com/gizzahub/gzh-manager-go/cmd/docker"
	doctorcmd "github.com/gizzahub/gzh-manager-go/cmd/doctor"
	"github.com/gizzahub/gzh-manager-go/cmd/ide"
	"github.com/gizzahub/gzh-manager-go/cmd/migrate"
	netenv "github.com/gizzahub/gzh-manager-go/cmd/net-env"
	"github.com/gizzahub/gzh-manager-go/cmd/pm"
	repoconfig "github.com/gizzahub/gzh-manager-go/cmd/repo-config"
	reposync "github.com/gizzahub/gzh-manager-go/cmd/repo-sync"
	"github.com/gizzahub/gzh-manager-go/cmd/shell"
	sshconfig "github.com/gizzahub/gzh-manager-go/cmd/ssh-config"
	synclone "github.com/gizzahub/gzh-manager-go/cmd/synclone"
	"github.com/gizzahub/gzh-manager-go/internal/logger"
)

var (
	verbose    bool
	debug      bool
	quiet      bool
	debugShell bool
)

func newRootCmd(ctx context.Context, version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gz",
		Short: "Cli 종합 Manager by Gizzahub",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Set global logging configuration based on flags
			logger.SetGlobalLoggingFlags(verbose, debug, quiet)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newVersionCmd(version))
	cmd.AddCommand(pm.NewPMCmd(ctx))
	cmd.AddCommand(synclone.NewSyncCloneCmd(ctx))
	cmd.AddCommand(config.NewConfigCmd()) //nolint:contextcheck // Command setup doesn't require context propagation
	cmd.AddCommand(doctorcmd.DoctorCmd)
	cmd.AddCommand(devenv.NewDevEnvCmd()) //nolint:contextcheck // Command setup doesn't require context propagation
	cmd.AddCommand(docker.DockerCmd)
	cmd.AddCommand(ide.NewIDECmd(ctx))
	cmd.AddCommand(migrate.NewMigrateCmd()) //nolint:contextcheck // Command setup doesn't require context propagation
	cmd.AddCommand(netenv.NewNetEnvCmd(ctx))
	cmd.AddCommand(repoconfig.NewRepoConfigCmd()) //nolint:contextcheck // Command setup doesn't require context propagation
	cmd.AddCommand(reposync.NewRepoSyncCmd(ctx))
	// Shell command is now hidden - only add if debug mode is enabled
	if debugShell || os.Getenv("GZH_DEBUG_SHELL") == "1" {
		shellCmd := shell.ShellCmd
		shellCmd.Hidden = true
		cmd.AddCommand(shellCmd)
	}
	cmd.AddCommand(sshconfig.NewSSHConfigCmd())
	cmd.AddCommand(NewWebhookCmd())
	cmd.AddCommand(NewEventCmd()) //nolint:contextcheck // Command setup doesn't require context propagation

	// Add global flags
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging (shows all log levels)")
	cmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress all logs except critical errors")

	// Hidden debug shell flag
	cmd.PersistentFlags().BoolVar(&debugShell, "debug-shell", false, "")
	cmd.PersistentFlags().MarkHidden("debug-shell")

	return cmd
}

// Execute invokes the command.
func Execute(ctx context.Context, version string) error {
	// Check if debug shell should be started immediately
	if os.Getenv("GZH_DEBUG_SHELL") == "1" {
		// Run shell directly
		shell.ShellCmd.Run(shell.ShellCmd, []string{})
		return nil
	}

	rootCmd := newRootCmd(ctx, version)

	// Check if --debug-shell flag is present
	for _, arg := range os.Args[1:] {
		if arg == "--debug-shell" {
			// Run shell directly
			shell.ShellCmd.Run(shell.ShellCmd, []string{})
			return nil
		}
	}

	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("error executing root command: %w", err)
	}

	return nil
}
