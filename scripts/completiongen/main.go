// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

// completiongen은 공개 CLI에 completion 명령을 노출하지 않고 릴리스용
// 셸 완성 파일을 생성한다.
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

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
		return writeWithNote(output, rootCmd.GenZshCompletion, "compdef _gz gz", true)
	case "fish":
		return writeWithNote(output, func(w io.Writer) error {
			return rootCmd.GenFishCompletion(w, true)
		}, "# fish completion for gz", false)
	default:
		return fmt.Errorf("unsupported shell %q (want bash, zsh, or fish)", args[0])
	}
}

// noteLines explain an asymmetry that a reader of the zsh and fish files trips
// over otherwise. The bash file is cobra's legacy static form: it enumerates
// every flag by name, so `grep -- --known-hosts completions/gzh-manager.bash`
// finds eight hits. zsh and fish are the dynamic form and ask the binary at
// completion time, so in those two files the only hit that same grep finds is
// this note itself, and without it the files look truncated.
//
// The two lines existed as a hand edit in the tracked zsh and fish files while
// `scripts/completions.sh` began with `rm -rf completions`, so no release ever
// shipped them. Emitting them from the generator is what makes the tracked file
// and the packaged file the same artifact.
var noteLines = []string{
	"# Command-specific flags, including --known-hosts and --accept-new-host-key,",
	"# are requested dynamically from gz at completion time, not listed in this file.",
}

// writeWithNote inserts noteLines directly after the line that anchor prefixes.
// A missing anchor is an error rather than a silent skip: if a cobra upgrade
// reshapes the preamble, this must fail loudly instead of quietly dropping the
// note, which is the precise failure the whole change exists to prevent.
func writeWithNote(output io.Writer, gen func(io.Writer) error, anchor string, blankBefore bool) error {
	var buf bytes.Buffer
	if err := gen(&buf); err != nil {
		return err
	}

	lines := strings.Split(buf.String(), "\n")
	for i, line := range lines {
		if !strings.HasPrefix(line, anchor) {
			continue
		}

		insert := noteLines
		if blankBefore {
			insert = append([]string{""}, noteLines...)
		}

		out := make([]string, 0, len(lines)+len(insert))
		out = append(out, lines[:i+1]...)
		out = append(out, insert...)
		out = append(out, lines[i+1:]...)

		_, err := io.WriteString(output, strings.Join(out, "\n"))

		return err
	}

	return fmt.Errorf("anchor %q not found in generated completion; cobra's preamble changed", anchor)
}
