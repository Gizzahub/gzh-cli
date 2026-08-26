// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

// manpagegen은 사용자 확장을 제외한 릴리스 command tree에서 gz(1)을 생성한다.
package main

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/cmd"
	"github.com/gizzahub/gzh-cli/internal/app"
	"github.com/gizzahub/gzh-cli/internal/config"
	"github.com/gizzahub/gzh-cli/internal/logger"
)

type rootCommandFactory func(context.Context, string, *app.AppContext) *cobra.Command

var newRootCommand rootCommandFactory = cmd.NewRootCmdForGeneration

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "generate manpage: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: go run ./scripts/manpagegen <output-path>")
	}

	generatedAt, err := sourceDate()
	if err != nil {
		return err
	}

	appCtx := &app.AppContext{
		Logger: logger.NewStructuredLogger("manpagegen", logger.LevelInfo),
		Config: config.DefaultGlobalConfig(),
	}
	rootCmd := newRootCommand(context.Background(), "dev", appCtx)

	return writeCompressedManpage(args[0], rootCmd, generatedAt)
}

func sourceDate() (time.Time, error) {
	value := os.Getenv("SOURCE_DATE_EPOCH")
	if value == "" {
		return time.Time{}, fmt.Errorf("SOURCE_DATE_EPOCH is required")
	}

	epoch, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse SOURCE_DATE_EPOCH: %w", err)
	}

	return time.Unix(epoch, 0).UTC(), nil
}

func writeCompressedManpage(path string, root *cobra.Command, generatedAt time.Time) (returnErr error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	temp, err := os.CreateTemp(dir, ".gz-manpage-*")
	if err != nil {
		return fmt.Errorf("create temporary manpage: %w", err)
	}
	tempPath := temp.Name()
	defer func() {
		if returnErr != nil {
			_ = temp.Close()
			_ = os.Remove(tempPath)
		}
	}()

	if err := writeGzip(temp, root, generatedAt); err != nil {
		return err
	}
	if err := temp.Sync(); err != nil {
		return fmt.Errorf("sync temporary manpage: %w", err)
	}
	if err := temp.Chmod(0o644); err != nil {
		return fmt.Errorf("set manpage mode: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close temporary manpage: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace manpage: %w", err)
	}

	return nil
}

func writeGzip(w io.Writer, root *cobra.Command, generatedAt time.Time) error {
	compressed, err := gzip.NewWriterLevel(w, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("create gzip writer: %w", err)
	}
	compressed.Header.ModTime = generatedAt
	compressed.Header.OS = 255

	if err := writeManpage(compressed, root, generatedAt); err != nil {
		_ = compressed.Close()
		return err
	}
	if err := compressed.Close(); err != nil {
		return fmt.Errorf("close gzip writer: %w", err)
	}

	return nil
}

func writeManpage(w io.Writer, root *cobra.Command, generatedAt time.Time) error {
	commands := append([]*cobra.Command(nil), root.Commands()...)
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name() < commands[j].Name()
	})

	var page strings.Builder
	fmt.Fprintf(&page, ".TH \"GZ\" \"1\" \"%s\" \"gz dev\" \"gz Manual\"\n", generatedAt.Format("2006-01-02"))
	page.WriteString(".SH NAME\n")
	fmt.Fprintf(&page, "gz \\- %s\n", escapeRoff(root.Short))
	page.WriteString(".SH SYNOPSIS\n.B gz\n.RI \"[flags] command\"\n")
	page.WriteString(".SH DESCRIPTION\n")
	writeParagraphs(&page, root.Long)
	page.WriteString(".SH COMMANDS\n")
	for _, command := range commands {
		if command.Hidden || !command.IsAvailableCommand() {
			continue
		}
		page.WriteString(".TP\n")
		fmt.Fprintf(&page, ".B \"gz %s\"\n%s\n", escapeRoff(command.Name()), escapeRoff(command.Short))
	}
	page.WriteString(".SH OPTIONS\n.nf\n")
	page.WriteString(escapeRoff(strings.TrimSpace(root.PersistentFlags().FlagUsagesWrapped(80))))
	page.WriteString("\n.fi\n")
	page.WriteString(".SH SEE ALSO\n")
	page.WriteString("Project documentation: https://github.com/Gizzahub/gzh-cli\n")

	if _, err := io.WriteString(w, page.String()); err != nil {
		return fmt.Errorf("write manpage: %w", err)
	}

	return nil
}

func writeParagraphs(page *strings.Builder, text string) {
	for index, paragraph := range strings.Split(text, "\n\n") {
		if index > 0 {
			page.WriteString(".PP\n")
		}
		page.WriteString(escapeRoff(strings.ReplaceAll(paragraph, "\n", " ")))
		page.WriteByte('\n')
	}
}

func escapeRoff(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\e")
	value = strings.ReplaceAll(value, "-", "\\-")
	if strings.HasPrefix(value, ".") || strings.HasPrefix(value, "'") {
		value = "\\&" + value
	}

	return value
}
