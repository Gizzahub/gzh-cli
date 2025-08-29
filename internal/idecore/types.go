// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package idecore provides core types and interfaces for IDE detection and management
package idecore

import "time"

// IDE represents an IDE installation.
type IDE struct {
	Name          string    `json:"name"`
	Executable    string    `json:"executable"`
	Version       string    `json:"version"`
	Type          string    `json:"type"`
	InstallMethod string    `json:"install_method"`
	InstallPath   string    `json:"install_path"`
	LastUpdated   time.Time `json:"last_updated"`
	Aliases       []string  `json:"aliases"`
}

// IDECache represents cached IDE scan results
type IDECache struct {
	Timestamp time.Time `json:"timestamp"`
	IDEs      []IDE     `json:"ides"`
}

// IDEDetectorInterface defines the interface for IDE detection
type IDEDetectorInterface interface {
	DetectIDEs(useCache bool) ([]IDE, error)
	FindIDEByAlias(ides []IDE, nameOrAlias string) *IDE
}

// DetectorFunc is a function type for creating IDE detectors
type DetectorFunc func() IDEDetectorInterface
