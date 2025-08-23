// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SetupLogger configures the logger for net-env components with consistent settings
func SetupLogger(verbose bool) (*zap.Logger, error) {
	config := zap.NewProductionConfig()

	if verbose {
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	} else {
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	// Use console encoder for better readability in CLI
	config.Encoding = "console"
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	return config.Build()
}

// GetDefaultLogger returns a pre-configured logger for net-env
func GetDefaultLogger() *zap.Logger {
	verbose := os.Getenv("GZH_NET_ENV_VERBOSE") == "true"
	logger, err := SetupLogger(verbose)
	if err != nil {
		// Fallback to noop logger if setup fails
		return zap.NewNop()
	}
	return logger
}
