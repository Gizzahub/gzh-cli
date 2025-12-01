// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package wrapper provides standard interface and utilities for integrating external libraries.
package wrapper

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/internal/app"
)

// ExternalLibrary is an interface for wrapping external libraries.
type ExternalLibrary interface {
	// Name returns the library name.
	Name() string

	// Version returns the library version.
	Version() string

	// CreateCommand creates a cobra command from the external library.
	CreateCommand(appCtx *app.AppContext) (*cobra.Command, error)

	// Validate checks if the library is properly configured.
	Validate() error

	// Dependencies returns the list of required external tools.
	Dependencies() []string

	// Repository returns the external library repository URL.
	Repository() string
}

// BaseWrapper provides common wrapper functionality.
type BaseWrapper struct {
	name         string
	version      string
	dependencies []string
	repository   string
}

// NewBaseWrapper creates a new BaseWrapper.
func NewBaseWrapper(name, version, repository string, deps []string) *BaseWrapper {
	return &BaseWrapper{
		name:         name,
		version:      version,
		dependencies: deps,
		repository:   repository,
	}
}

// Name returns the library name.
func (w *BaseWrapper) Name() string {
	return w.name
}

// Version returns the library version.
func (w *BaseWrapper) Version() string {
	return w.version
}

// Dependencies returns the required dependencies.
func (w *BaseWrapper) Dependencies() []string {
	return w.dependencies
}

// Repository returns the repository URL.
func (w *BaseWrapper) Repository() string {
	return w.repository
}

// Validate checks dependencies.
func (w *BaseWrapper) Validate() error {
	for _, dep := range w.dependencies {
		if _, err := exec.LookPath(dep); err != nil {
			return fmt.Errorf("missing dependency: %s", dep)
		}
	}
	return nil
}

// Metadata contains wrapper metadata information.
type Metadata struct {
	Name         string   // Library name
	Version      string   // Version
	Repository   string   // Repository URL
	Dependencies []string // Dependency list
	Description  string   // Description
	Status       string   // Status (active, deprecated, experimental)
}

// GetMetadata extracts metadata from a wrapper.
func GetMetadata(lib ExternalLibrary) Metadata {
	return Metadata{
		Name:         lib.Name(),
		Version:      lib.Version(),
		Repository:   lib.Repository(),
		Dependencies: lib.Dependencies(),
		Status:       "active",
	}
}
