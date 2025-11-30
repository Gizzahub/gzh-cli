// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package app

import (
	"github.com/Gizzahub/gzh-cli/internal/config"
	"github.com/Gizzahub/gzh-cli/internal/logger"
)

// NewTestAppContext returns an AppContext with default config and logger for testing.
func NewTestAppContext() *AppContext {
	return &AppContext{
		Logger: logger.NewStructuredLogger("test", logger.LevelInfo),
		Config: config.DefaultGlobalConfig(),
	}
}
