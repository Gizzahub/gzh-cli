// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

// manpagegen은 사용자 확장을 로드하지 않은 공개 CLI command tree에서
// 릴리스용 gzip manpage를 생성한다.
package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gizzahub/gzh-cli/cmd"
	"github.com/gizzahub/gzh-cli/internal/app"
	"github.com/gizzahub/gzh-cli/internal/config"
	"github.com/gizzahub/gzh-cli/internal/logger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "generate manpage: %v\n", err)
		os.Exit(1)
	}
}

func run() (runErr error) {
	tempHome, err := os.MkdirTemp("", "gzh-cli-manpagegen-")
	if err != nil {
		return fmt.Errorf("create isolated home: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tempHome); err != nil && runErr == nil {
			runErr = fmt.Errorf("remove isolated home: %w", err)
		}
	}()

	for _, key := range []string{"HOME", "USERPROFILE", "XDG_CONFIG_HOME"} {
		if err := os.Setenv(key, tempHome); err != nil {
			return fmt.Errorf("isolate %s: %w", key, err)
		}
	}

	epochText := os.Getenv("SOURCE_DATE_EPOCH")
	epoch, err := strconv.ParseInt(epochText, 10, 64)
	if err != nil {
		return fmt.Errorf("parse SOURCE_DATE_EPOCH %q: %w", epochText, err)
	}
	date := time.Unix(epoch, 0).UTC()

	appCtx := &app.AppContext{
		Logger: logger.NewStructuredLogger("manpagegen", logger.LevelInfo),
		Config: config.DefaultGlobalConfig(),
	}
	rootCmd := cmd.NewRootCmd(context.Background(), "dev", appCtx)

	var roff bytes.Buffer
	fmt.Fprintf(&roff, ".nh\n.TH \"GZ\" \"1\" \"%s\" \"gzh-cli\" \"gz Manual\"\n", date.Format("2006-01-02"))
	fmt.Fprintf(&roff, ".SH NAME\n%s\n", wrapRoff("gz \\- "+roffEscape(rootCmd.Short), 60))
	fmt.Fprintf(&roff, ".SH SYNOPSIS\n\\fBgz [flags] <command>\\fP\n")
	fmt.Fprintf(&roff, ".SH DESCRIPTION\n%s\n", wrapRoff(roffEscape(rootCmd.Long), 60))
	fmt.Fprintln(&roff, ".SH COMMANDS")
	for _, subcommand := range rootCmd.Commands() {
		if subcommand.Hidden {
			continue
		}
		fmt.Fprintf(&roff, ".TP\n\\fBgz %s\\fP\n%s\n", roffEscape(subcommand.Name()), wrapRoff(roffEscape(subcommand.Short), 60))
	}
	fmt.Fprintf(&roff, ".SH OPTIONS\n.nf\n%s.fi\n", roffEscape(rootCmd.PersistentFlags().FlagUsagesWrapped(60)))

	writer := gzip.NewWriter(os.Stdout)
	if _, err := writer.Write(roff.Bytes()); err != nil {
		_ = writer.Close()
		return fmt.Errorf("write compressed roff: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close gzip stream: %w", err)
	}

	return nil
}

func roffEscape(value string) string {
	return strings.NewReplacer("\\", "\\\\", "-", "\\-").Replace(value)
}

func wrapRoff(value string, width int) string {
	words := strings.Fields(value)
	if len(words) == 0 {
		return ""
	}

	var output strings.Builder
	lineBytes := 0
	for _, word := range words {
		separator := 0
		if lineBytes > 0 {
			separator = 1
		}
		wordWidth := roffInputWidth(word)
		if lineBytes > 0 && lineBytes+separator+wordWidth > width {
			output.WriteByte('\n')
			lineBytes = 0
			separator = 0
		}
		if separator > 0 {
			output.WriteByte(' ')
			lineBytes++
		}
		output.WriteString(word)
		lineBytes += wordWidth
	}

	return output.String()
}

func roffInputWidth(value string) int {
	width := 0
	for _, character := range value {
		if character > 127 {
			width += len(`\[uFFFF]`)
			continue
		}
		width++
	}

	return width
}
