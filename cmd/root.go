// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	devenv "github.com/Gizzahub/gzh-cli/cmd/dev-env"
	_ "github.com/Gizzahub/gzh-cli/cmd/doctor"
	"github.com/Gizzahub/gzh-cli/cmd/git"
	"github.com/Gizzahub/gzh-cli/cmd/ide"
	netenv "github.com/Gizzahub/gzh-cli/cmd/net-env"
	"github.com/Gizzahub/gzh-cli/cmd/profile"
	repoconfig "github.com/Gizzahub/gzh-cli/cmd/repo-config"
	"github.com/Gizzahub/gzh-cli/cmd/selfupdate"
	"github.com/Gizzahub/gzh-cli/cmd/synclone"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
	"github.com/Gizzahub/gzh-cli/cmd/shell"
	versioncmd "github.com/Gizzahub/gzh-cli/cmd/version"
	"github.com/Gizzahub/gzh-cli/internal/app"
	"github.com/Gizzahub/gzh-cli/internal/config"
	"github.com/Gizzahub/gzh-cli/internal/extensions"
	"github.com/Gizzahub/gzh-cli/internal/logger"
)

var (
	verbose      bool
	debug        bool
	quiet        bool
	debugShell   bool
	experimental bool
)

// NewRootCmd creates the root command and wires up subcommands with shared context.
func NewRootCmd(ctx context.Context, version string, appCtx *app.AppContext) *cobra.Command {
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
			// Propagate verbose to env for deep packages that can't import logger
			if verbose {
				_ = os.Setenv("GZH_VERBOSE", "1")
			} else {
				_ = os.Unsetenv("GZH_VERBOSE")
			}
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	// Register all core feature commands with AppContext
	RegisterPMCmd(appCtx)         // Package manager (from pm_wrapper.go)
	RegisterQualityCmd(appCtx)    // Code quality (from quality_wrapper.go)
	RegisterShellforgeCmd(appCtx) // Shell config builder (from shellforge_wrapper.go)
	synclone.RegisterSyncCloneCmd(appCtx)
	devenv.RegisterDevEnvCmd(appCtx)
	ide.RegisterIDECmd(appCtx)
	netenv.RegisterNetEnvCmd(appCtx)
	repoconfig.RegisterRepoConfigCmd(appCtx)
	profile.RegisterProfileCmd(appCtx)
	git.RegisterGitCmd(appCtx)
	selfupdate.RegisterSelfUpdateCmd(appCtx)

	// Initialize lifecycle manager and filter commands
	lifecycleManager := registry.NewLifecycleManager()
	if experimental {
		lifecycleManager.EnableExperimental()
	}
	filteredProviders := lifecycleManager.FilterCommands(registry.List())

	// Add all registered commands to root with lifecycle checks
	for _, provider := range filteredProviders {
		providerCmd := provider.Command()

		// Wrap the command execution with lifecycle validation
		if registry.HasMetadata(provider) {
			meta := registry.GetMetadata(provider)
			originalRunE := providerCmd.RunE
			originalRun := providerCmd.Run

			// Wrap RunE if exists
			if originalRunE != nil {
				providerCmd.RunE = func(cmd *cobra.Command, args []string) error {
					if err := lifecycleManager.CheckCommand(meta); err != nil {
						return err
					}
					return originalRunE(cmd, args)
				}
			} else if originalRun != nil {
				// Wrap Run if exists
				providerCmd.Run = func(cmd *cobra.Command, args []string) {
					if err := lifecycleManager.CheckCommand(meta); err != nil {
						fmt.Fprintf(os.Stderr, "Error: %v\n", err)
						os.Exit(1)
					}
					originalRun(cmd, args)
				}
			}
		}

		cmd.AddCommand(providerCmd)
	}

	// Load user extensions (aliases and external commands)
	// 실패해도 계속 진행 (사용자 확장은 선택적)
	extensionLoader := extensions.NewLoader()
	if err := extensionLoader.RegisterAll(cmd); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Failed to load extensions: %v\n", err)
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
	cmd.PersistentFlags().BoolVar(&experimental, "experimental", false, "Enable experimental features")

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

	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		cfg = config.DefaultGlobalConfig()
	}

	log := logger.NewStructuredLogger("gzh-cli", logger.LevelInfo)
	appCtx := &app.AppContext{
		Logger: log,
		Config: cfg,
	}

	rootCmd := NewRootCmd(ctx, version, appCtx)

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
