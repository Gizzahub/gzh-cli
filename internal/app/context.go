// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package app

import (
	"github.com/Gizzahub/gzh-cli/internal/config"
	"github.com/Gizzahub/gzh-cli/internal/logger"
)

// AppContext holds application-wide dependencies.
type AppContext struct {
	Logger *logger.StructuredLogger
	Config *config.GlobalConfig
}
