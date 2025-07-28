// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package sync

import (
	"fmt"
	"strings"
)

// Options contains options for repository synchronization.
type Options struct {
	// Source and destination
	From string
	To   string

	// Sync options
	CreateMissing  bool
	UpdateExisting bool
	Force          bool

	// Include options
	IncludeCode     bool
	IncludeIssues   bool
	IncludePRs      bool
	IncludeWiki     bool
	IncludeReleases bool
	IncludeSettings bool

	// Filtering
	Match   string
	Exclude string

	// Execution options
	Parallel int
	DryRun   bool
	Verbose  bool
}

// SyncTarget represents a parsed sync target (provider:org/repo or provider:org).
type SyncTarget struct {
	Provider string
	Org      string
	Repo     string // empty if syncing entire organization
}

// ParseTarget parses a sync target string into components.
func ParseTarget(target string) (*SyncTarget, error) {
	if target == "" {
		return nil, fmt.Errorf("target cannot be empty")
	}

	// Split by first colon to separate provider from path
	parts := strings.SplitN(target, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid target format: %s (expected provider:org or provider:org/repo)", target)
	}

	provider := strings.TrimSpace(parts[0])
	path := strings.TrimSpace(parts[1])

	if provider == "" {
		return nil, fmt.Errorf("provider cannot be empty")
	}

	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	// Parse path (org or org/repo)
	pathParts := strings.SplitN(path, "/", 2)
	org := strings.TrimSpace(pathParts[0])
	if org == "" {
		return nil, fmt.Errorf("organization cannot be empty")
	}

	syncTarget := &SyncTarget{
		Provider: provider,
		Org:      org,
	}

	// If repo is specified
	if len(pathParts) == 2 {
		repo := strings.TrimSpace(pathParts[1])
		if repo == "" {
			return nil, fmt.Errorf("repository name cannot be empty")
		}
		syncTarget.Repo = repo
	}

	return syncTarget, nil
}

// IsOrganization returns true if this target represents an entire organization.
func (t *SyncTarget) IsOrganization() bool {
	return t.Repo == ""
}

// IsRepository returns true if this target represents a specific repository.
func (t *SyncTarget) IsRepository() bool {
	return t.Repo != ""
}

// String returns the string representation of the target.
func (t *SyncTarget) String() string {
	if t.IsRepository() {
		return fmt.Sprintf("%s:%s/%s", t.Provider, t.Org, t.Repo)
	}
	return fmt.Sprintf("%s:%s", t.Provider, t.Org)
}

// FullName returns the full repository name (org/repo) or just org for organization targets.
func (t *SyncTarget) FullName() string {
	if t.IsRepository() {
		return fmt.Sprintf("%s/%s", t.Org, t.Repo)
	}
	return t.Org
}

// Validate validates the sync options.
func (opts *Options) Validate() error {
	if opts.From == "" {
		return fmt.Errorf("source (--from) is required")
	}

	if opts.To == "" {
		return fmt.Errorf("destination (--to) is required")
	}

	// Validate source target
	sourceTarget, err := ParseTarget(opts.From)
	if err != nil {
		return fmt.Errorf("invalid source target: %w", err)
	}

	// Validate destination target
	destTarget, err := ParseTarget(opts.To)
	if err != nil {
		return fmt.Errorf("invalid destination target: %w", err)
	}

	// Check that both targets are of the same type (org or repo)
	if sourceTarget.IsOrganization() != destTarget.IsOrganization() {
		return fmt.Errorf("source and destination must be of the same type (both organizations or both repositories)")
	}

	// Validate parallel workers
	if opts.Parallel < 1 {
		opts.Parallel = 1
	}
	if opts.Parallel > 20 {
		return fmt.Errorf("parallel workers cannot exceed 20")
	}

	// Validate that at least one sync feature is enabled
	if !opts.IncludeCode && !opts.IncludeIssues && !opts.IncludePRs &&
		!opts.IncludeWiki && !opts.IncludeReleases && !opts.IncludeSettings {
		return fmt.Errorf("at least one sync feature must be enabled")
	}

	return nil
}

// GetSourceTarget parses and returns the source target.
func (opts *Options) GetSourceTarget() (*SyncTarget, error) {
	return ParseTarget(opts.From)
}

// GetDestinationTarget parses and returns the destination target.
func (opts *Options) GetDestinationTarget() (*SyncTarget, error) {
	return ParseTarget(opts.To)
}
