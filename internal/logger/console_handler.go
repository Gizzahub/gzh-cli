// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

// ConsoleHandler provides human-readable console output.
type ConsoleHandler struct {
	writer io.Writer
	level  slog.Level
	attrs  []slog.Attr
	group  string
}

// NewConsoleHandler creates a new console handler with human-readable output.
func NewConsoleHandler(w io.Writer, opts *slog.HandlerOptions) *ConsoleHandler {
	level := slog.LevelInfo
	if opts != nil && opts.Level != nil {
		level = opts.Level.Level()
	}

	return &ConsoleHandler{
		writer: w,
		level:  level,
		attrs:  make([]slog.Attr, 0),
	}
}

// Enabled returns whether the handler is enabled for the given level.
func (ch *ConsoleHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= ch.level
}

// Handle processes a log record and outputs it in human-readable format.
func (ch *ConsoleHandler) Handle(_ context.Context, record slog.Record) error {
	timestamp := record.Time.Format("15:04:05")
	level := ch.formatLevel(record.Level)
	
	// Extract key information from attributes
	var component, operation, orgName string
	var errorMsg string
	
	record.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "component":
			component = a.Value.String()
		case "operation":
			operation = a.Value.String()
		case "org_name":
			orgName = a.Value.String()
		case "error":
			errorMsg = a.Value.String()
		}
		return true
	})

	// Build context string
	var contextParts []string
	if component != "" {
		contextParts = append(contextParts, fmt.Sprintf("component=%s", component))
	}
	if orgName != "" {
		contextParts = append(contextParts, fmt.Sprintf("org=%s", orgName))
	}
	if operation != "" {
		contextParts = append(contextParts, fmt.Sprintf("op=%s", operation))
	}
	
	contextStr := ""
	if len(contextParts) > 0 {
		contextStr = fmt.Sprintf(" [%s]", strings.Join(contextParts, " "))
	}

	// Format the message
	message := record.Message
	if errorMsg != "" {
		message = fmt.Sprintf("%s: %s", message, errorMsg)
	}

	// Output format: TIME LEVEL[CONTEXT] MESSAGE
	output := fmt.Sprintf("%s %s%s %s\n", timestamp, level, contextStr, message)
	
	_, err := ch.writer.Write([]byte(output))
	return err
}

// WithAttrs returns a new handler with the given attributes.
func (ch *ConsoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(ch.attrs)+len(attrs))
	copy(newAttrs, ch.attrs)
	copy(newAttrs[len(ch.attrs):], attrs)

	return &ConsoleHandler{
		writer: ch.writer,
		level:  ch.level,
		attrs:  newAttrs,
		group:  ch.group,
	}
}

// WithGroup returns a new handler with the given group.
func (ch *ConsoleHandler) WithGroup(name string) slog.Handler {
	return &ConsoleHandler{
		writer: ch.writer,
		level:  ch.level,
		attrs:  ch.attrs,
		group:  name,
	}
}

// formatLevel formats the log level with colors and padding.
func (ch *ConsoleHandler) formatLevel(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return "\033[36mDEBUG\033[0m" // Cyan
	case slog.LevelInfo:
		return "\033[32mINFO \033[0m" // Green
	case slog.LevelWarn:
		return "\033[33mWARN \033[0m" // Yellow
	case slog.LevelError:
		return "\033[31mERROR\033[0m" // Red
	default:
		return "INFO "
	}
}