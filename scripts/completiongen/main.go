// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

// completiongen은 공개 CLI에 completion 명령을 노출하지 않고 릴리스용
// 셸 완성 파일을 생성한다.
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/cmd"
	"github.com/gizzahub/gzh-cli/internal/app"
	"github.com/gizzahub/gzh-cli/internal/config"
	"github.com/gizzahub/gzh-cli/internal/logger"
)

type rootCommandFactory func(context.Context, string, *app.AppContext) *cobra.Command

var newRootCommand rootCommandFactory = cmd.NewRootCmdForGeneration

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "generate completion: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, output io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: go run ./scripts/completiongen <bash|zsh|fish>")
	}

	appCtx := &app.AppContext{
		Logger: logger.NewStructuredLogger("completiongen", logger.LevelInfo),
		Config: config.DefaultGlobalConfig(),
	}
	rootCmd := newRootCommand(context.Background(), "dev", appCtx)

	switch args[0] {
	case "bash":
		return rootCmd.GenBashCompletion(output)
	case "zsh":
		return rootCmd.GenZshCompletion(output)
	case "fish":
		return rootCmd.GenFishCompletion(output, true)
	default:
		return fmt.Errorf("unsupported shell %q (want bash, zsh, or fish)", args[0])
	}
}
