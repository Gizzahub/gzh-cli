// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"fmt"
	"os"
	"slices"

	"github.com/spf13/cobra"

	devenv "github.com/gizzahub/gzh-cli/cmd/dev-env"
	_ "github.com/gizzahub/gzh-cli/cmd/doctor"
	"github.com/gizzahub/gzh-cli/cmd/git"
	gitsync "github.com/gizzahub/gzh-cli/cmd/git-sync"
	"github.com/gizzahub/gzh-cli/cmd/ide"
	netenv "github.com/gizzahub/gzh-cli/cmd/net-env"
	"github.com/gizzahub/gzh-cli/cmd/profile"
	repoconfig "github.com/gizzahub/gzh-cli/cmd/repo-config"
	"github.com/gizzahub/gzh-cli/cmd/selfupdate"
	"github.com/gizzahub/gzh-cli/cmd/synclone"

	"github.com/gizzahub/gzh-cli/cmd/registry"
	"github.com/gizzahub/gzh-cli/cmd/shell"
	versioncmd "github.com/gizzahub/gzh-cli/cmd/version"
	"github.com/gizzahub/gzh-cli/internal/app"
	"github.com/gizzahub/gzh-cli/internal/config"
	"github.com/gizzahub/gzh-cli/internal/extensions"
	"github.com/gizzahub/gzh-cli/internal/logger"
)

var (
	verbose      bool
	debug        bool
	quiet        bool
	debugShell   bool
	experimental bool
)

type extensionRegistrar func(*cobra.Command) error

// NewRootCmd creates the root command and wires up subcommands with shared context.
func NewRootCmd(ctx context.Context, version string, appCtx *app.AppContext) *cobra.Command {
	return newRootCmd(ctx, version, appCtx, registerUserExtensions)
}

// NewRootCmdForGeneration은 사용자 확장 없이 배포 문서용 command tree를 만든다.
func NewRootCmdForGeneration(ctx context.Context, version string, appCtx *app.AppContext) *cobra.Command {
	return newRootCmd(ctx, version, appCtx, nil)
}

func newRootCmd(
	ctx context.Context,
	version string,
	appCtx *app.AppContext,
	registerExtensions extensionRegistrar,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gz",
		Short: "개발 환경 및 Git 플랫폼 통합 관리 도구",
		Long: `gz는 개발자를 위한 종합 CLI 도구입니다.

개발 환경 설정, Git 플랫폼 관리, IDE 모니터링, 네트워크 환경 전환 등
다양한 개발 워크플로우를 통합적으로 관리할 수 있습니다.

Utility Commands: doctor, version`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// 여기 온 시점에 깃발 파싱과 인자 검사는 이미 지났다. 그러니
			// 지금부터 나는 오류는 사용법 오류가 아니라 실행 중 오류다.
			// 도움말을 통째로 찍을 이유가 없다.
			//
			// cobra는 ExecuteC에서 뿌리와 잎만 본다(command.go:1165).
			// 중간 명령에 SilenceUsage를 걸어도 자식에게는 아무 효과가 없다 --
			// synclone.go:60이 그랬고 `gz synclone config generate discover`가
			// Ctrl+C에 도움말을 다 찍었다. PersistentPreRun은 실행된 잎을
			// 받으므로 여기서 걸면 나무 전체가 덮인다.
			//
			// 잃는 것: --required 깃발 누락과 깃발 묶음 위반은 cobra가 이
			// 함수보다 뒤에 검사해서 사용법이 안 나온다. 대신 그쪽은 무엇이
			// 빠졌는지 오류 문구가 이미 분명하다.
			cmd.SilenceUsage = true

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
	gitsync.RegisterGitSyncCmd(appCtx)
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

	// 사용자 확장은 선택 사항이므로 실패해도 기본 command tree를 유지한다.
	if registerExtensions != nil {
		if err := registerExtensions(cmd); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Failed to load extensions: %v\n", err)
		}
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

func registerUserExtensions(cmd *cobra.Command) error {
	extensionLoader := extensions.NewLoader()
	return extensionLoader.RegisterAll(cmd)
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
	if slices.Contains(os.Args[1:], "--debug-shell") {
		// Run shell directly
		shell.ShellCmd.Run(shell.ShellCmd, []string{})
		return nil
	}

	// ExecuteContext여야 한다. Execute()는 ctx가 비어 있으면 cobra가
	// context.Background()를 대신 넣는다(command.go의 ExecuteC). 그러면
	// cmd.Context()가 취소되지 않는 맥락을 돌려주고, 그것을 쓰는 77곳이
	// 전부 헛돌게 된다 -- apprunner가 SIGINT/SIGTERM에서 cancel()을
	// 불러도 아무 데도 닿지 않는다.
	//
	// 실제로 `gz ide monitor`는 Ctrl+C로 멈추지 않았다. "shutting down
	// gracefully"만 찍고 계속 살아 있었다. ctx.Done()을 기다리는 자리가
	// 영원히 열리지 않는 통로를 보고 있었기 때문이다.
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		return fmt.Errorf("error executing root command: %w", err)
	}

	return nil
}
