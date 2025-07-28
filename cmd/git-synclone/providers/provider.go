// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package providers provides adapters for Git hosting platforms.
// It reuses the existing synclone package implementation to provide
// Git extension functionality for repository cloning and management.
package providers

import (
	"context"
	"fmt"

	bulkclone "github.com/gizzahub/gzh-manager-go/pkg/synclone"
)

// ProviderAdapter provides a unified interface for all Git hosting platforms.
// It adapts the existing synclone functionality for use in Git extensions.
type ProviderAdapter interface {
	// Clone repositories from the provider using the specified strategy
	CloneRepositories(ctx context.Context, request *CloneRequest) (*CloneResult, error)

	// List repositories from the provider without cloning
	ListRepositories(ctx context.Context, request *ListRequest) (*ListResult, error)

	// Validate provider-specific options
	ValidateOptions(options *CloneOptions) error

	// Get provider name for logging and identification
	GetProviderName() string
}

// CloneRequest represents a request to clone repositories from a provider.
type CloneRequest struct {
	Organization string
	TargetPath   string
	Strategy     string
	Filters      *RepositoryFilters
	Options      *CloneOptions
}

// ListRequest represents a request to list repositories from a provider.
type ListRequest struct {
	Organization string
	Filters      *RepositoryFilters
	Options      *CloneOptions
}

// RepositoryFilters contains filtering criteria for repositories.
type RepositoryFilters struct {
	NamePattern     string
	Visibility      string // public, private, all
	IncludeArchived bool
	IncludeForks    bool
	Language        string
	Topics          []string
	MinStars        int
	MaxStars        int
}

// CloneOptions contains configuration options for cloning operations.
type CloneOptions struct {
	Protocol       string
	Strategy       string
	Parallel       int
	MaxRetries     int
	Resume         bool
	DryRun         bool
	ProgressMode   string
	Token          string
	ConfigFile     string
	UseConfig      bool
	CleanupOrphans bool
}

// CloneResult represents the result of a clone operation.
type CloneResult struct {
	TotalRepositories int
	ClonesSuccessful  int
	ClonesFailed      int
	ClonesSkipped     int
	Repositories      []RepositoryResult
	Errors            []error
}

// ListResult represents the result of a list repositories operation.
type ListResult struct {
	TotalRepositories int
	Repositories      []RepositoryInfo
}

// RepositoryResult represents the result of a single repository operation.
type RepositoryResult struct {
	Name       string
	URL        string
	Path       string
	Success    bool
	Error      string
	Skipped    bool
	SkipReason string
}

// RepositoryInfo represents basic information about a repository.
type RepositoryInfo struct {
	Name        string
	FullName    string
	CloneURL    string
	SSHURL      string
	Description string
	Language    string
	Private     bool
	Archived    bool
	Fork        bool
	Stars       int
	Topics      []string
}

// BaseProviderAdapter provides common functionality for all provider adapters.
type BaseProviderAdapter struct {
	config *bulkclone.BulkCloneConfig
}

// NewBaseProviderAdapter creates a new base provider adapter.
func NewBaseProviderAdapter() *BaseProviderAdapter {
	return &BaseProviderAdapter{}
}

// LoadConfig loads the synclone configuration from file or standard locations.
func (b *BaseProviderAdapter) LoadConfig(configFile string) error {
	var err error
	if configFile != "" {
		b.config, err = bulkclone.LoadConfig(configFile)
	} else {
		b.config, err = bulkclone.LoadConfig("")
	}
	return err
}

// GetConfig returns the loaded configuration.
func (b *BaseProviderAdapter) GetConfig() *bulkclone.BulkCloneConfig {
	return b.config
}

// ValidateCommonOptions validates common options across all providers.
func (b *BaseProviderAdapter) ValidateCommonOptions(options *CloneOptions) error {
	if options == nil {
		return fmt.Errorf("options cannot be nil")
	}

	// Validate strategy
	validStrategies := map[string]bool{
		"reset": true,
		"pull":  true,
		"fetch": true,
	}
	if !validStrategies[options.Strategy] {
		return fmt.Errorf("invalid strategy '%s': must be one of: reset, pull, fetch", options.Strategy)
	}

	// Validate protocol
	validProtocols := map[string]bool{
		"https": true,
		"ssh":   true,
	}
	if !validProtocols[options.Protocol] {
		return fmt.Errorf("invalid protocol '%s': must be one of: https, ssh", options.Protocol)
	}

	// Validate progress mode
	validProgress := map[string]bool{
		"bar":     true,
		"dots":    true,
		"spinner": true,
		"quiet":   true,
	}
	if !validProgress[options.ProgressMode] {
		return fmt.Errorf("invalid progress mode '%s': must be one of: bar, dots, spinner, quiet", options.ProgressMode)
	}

	// Validate parallel workers
	if options.Parallel < 1 {
		return fmt.Errorf("parallel workers must be at least 1")
	}

	return nil
}
