// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package event

import "github.com/gizzahub/gzh-manager-go/pkg/github"

// LoggerAdapter adapts SimpleLogger to github.Logger interface.
type LoggerAdapter struct {
	logger *SimpleLogger
}

// NewLoggerAdapter creates a new logger adapter.
func NewLoggerAdapter() github.Logger {
	return &LoggerAdapter{
		logger: NewSimpleLogger(),
	}
}

// Debug logs a debug message.
func (l *LoggerAdapter) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, args...)
}

// Info logs an info message.
func (l *LoggerAdapter) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

// Warn logs a warning message.
func (l *LoggerAdapter) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, args...)
}

// Error logs an error message.
func (l *LoggerAdapter) Error(msg string, args ...interface{}) {
	// Convert to SimpleLogger's Error signature
	if len(args) > 0 {
		if err, ok := args[0].(error); ok {
			l.logger.Error(msg, err, args[1:]...)
			return
		}
	}

	l.logger.Error(msg, nil, args...)
}
