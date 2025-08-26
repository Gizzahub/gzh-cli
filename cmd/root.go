// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	_ "github.com/Gizzahub/gzh-cli/cmd/dev-env"
	_ "github.com/Gizzahub/gzh-cli/cmd/doctor"
	_ "github.com/Gizzahub/gzh-cli/cmd/git"
	_ "github.com/Gizzahub/gzh-cli/cmd/ide"
	_ "github.com/Gizzahub/gzh-cli/cmd/net-env"
	_ "github.com/Gizzahub/gzh-cli/cmd/pm"
	_ "github.com/Gizzahub/gzh-cli/cmd/profile"
	_ "github.com/Gizzahub/gzh-cli/cmd/quality"
	_ "github.com/Gizzahub/gzh-cli/cmd/repo-config"
	_ "github.com/Gizzahub/gzh-cli/cmd/synclone"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
	"github.com/Gizzahub/gzh-cli/cmd/shell"
	versioncmd "github.com/Gizzahub/gzh-cli/cmd/version"
	"github.com/Gizzahub/gzh-cli/internal/logger"
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
		Short: "개발 환경 및 Git 플랫폼 통합 관리 도구",
		Long: `gz는 개발자를 위한 종합 CLI 도구입니다.

개발 환경 설정, Git 플랫폼 관리, IDE 모니터링, 네트워크 환경 전환 등
다양한 개발 워크플로우를 통합적으로 관리할 수 있습니다.

Utility Commands: doctor, version`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Set global logging configuration based on flags
			logger.SetGlobalLoggingFlags(verbose, debug, quiet)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	for _, provider := range registry.List() {
		cmd.AddCommand(provider.Command())
	}

	// Utility commands - set as hidden to reduce clutter in main help
	versionCmd := versioncmd.NewVersionCmd(version)
	versionCmd.Hidden = true
	cmd.AddCommand(versionCmd)

	// Shell command is hidden - only add if debug mode is enabled
	if debugShell || os.Getenv("GZH_DEBUG_SHELL") == "1" {
		shellCmd := shell.ShellCmd
		shellCmd.Hidden = true
		cmd.AddCommand(shellCmd)
	}

	// Hide completion command and help command
	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

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
