// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package logger

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Gizzahub/gzh-cli/internal/config"
)

// NewDualLogger creates a logger that outputs to both console (human-readable) and file (JSON).
func NewDualLogger(component string, level LogLevel) (*StructuredLogger, error) {
	// Load global config
	globalConfig, err := config.LoadGlobalConfig()
	if err != nil {
		// Fallback to console-only logging if config loading fails
		return NewConsoleOnlyLogger(component, level), err
	}

	var slogLevel slog.Level
	switch level {
	case LevelDebug:
		slogLevel = slog.LevelDebug
	case LevelInfo:
		slogLevel = slog.LevelInfo
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     slogLevel,
		AddSource: true,
	}

	// Create console handler (human-readable)
	consoleHandler := NewConsoleHandler(os.Stdout, opts)

	var fileLogger *slog.Logger
	hasFileLogger := false

	// Create file handler (JSON) if enabled
	if globalConfig.Logging.Enabled {
		if err := ensureLogDir(globalConfig.Logging.FilePath); err == nil {
			if fileWriter, err := os.OpenFile(globalConfig.Logging.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600); err == nil {
				jsonHandler := slog.NewJSONHandler(fileWriter, opts)
				fileLogger = slog.New(jsonHandler)
				hasFileLogger = true
			}
		}
	}

	// Create a multi-handler that writes to both
	var handler slog.Handler
	if hasFileLogger {
		handler = NewMultiHandler(consoleHandler, fileLogger.Handler())
	} else {
		handler = consoleHandler
	}

	mainLogger := slog.New(handler)

	return &StructuredLogger{
		logger:    mainLogger,
		level:     slogLevel,
		context:   make(map[string]interface{}),
		component: component,
		sessionID: generateSessionID(component),
	}, nil
}

// NewConsoleOnlyLogger creates a logger that only outputs to console.
func NewConsoleOnlyLogger(component string, level LogLevel) *StructuredLogger {
	var slogLevel slog.Level
	switch level {
	case LevelDebug:
		slogLevel = slog.LevelDebug
	case LevelInfo:
		slogLevel = slog.LevelInfo
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     slogLevel,
		AddSource: true,
	}

	consoleHandler := NewConsoleHandler(os.Stdout, opts)
	logger := slog.New(consoleHandler)

	return &StructuredLogger{
		logger:    logger,
		level:     slogLevel,
		context:   make(map[string]interface{}),
		component: component,
		sessionID: generateSessionID(component),
	}
}

// ensureLogDir creates the log directory if it doesn't exist.
func ensureLogDir(logPath string) error {
	dir := filepath.Dir(logPath)
	return os.MkdirAll(dir, 0o750)
}

// MultiHandler implements slog.Handler to write to multiple handlers.
type MultiHandler struct {
	handlers []slog.Handler
}

// NewMultiHandler creates a handler that writes to multiple handlers.
func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

// Enabled returns true if any handler is enabled.
func (mh *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range mh.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle writes the record to all handlers.
func (mh *MultiHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, h := range mh.handlers {
		if h.Enabled(ctx, record.Level) {
			if err := h.Handle(ctx, record); err != nil {
				return err
			}
		}
	}
	return nil
}

// WithAttrs returns a new MultiHandler with attributes added to all handlers.
func (mh *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(mh.handlers))
	for i, h := range mh.handlers {
		newHandlers[i] = h.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: newHandlers}
}

// WithGroup returns a new MultiHandler with a group added to all handlers.
func (mh *MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(mh.handlers))
	for i, h := range mh.handlers {
		newHandlers[i] = h.WithGroup(name)
	}
	return &MultiHandler{handlers: newHandlers}
}
