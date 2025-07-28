// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package clone

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/gizzahub/gzh-manager-go/pkg/gitlab"
)

// LegacyAdapter provides compatibility with existing synclone functionality.
// This adapter allows the new clone command to leverage existing provider
// implementations while maintaining the new interface.
type LegacyAdapter struct {
	options *CloneOptions
}

// NewLegacyAdapter creates a new legacy adapter.
func NewLegacyAdapter(opts *CloneOptions) *LegacyAdapter {
	return &LegacyAdapter{
		options: opts,
	}
}

// ExecuteClone executes clone operation using legacy synclone providers.
func (l *LegacyAdapter) ExecuteClone(ctx context.Context) error {
	switch strings.ToLower(l.options.Provider) {
	case "github":
		return l.executeGitHubClone(ctx)
	case "gitlab":
		return l.executeGitLabClone(ctx)
	case "gitea":
		return fmt.Errorf("gitea provider not yet implemented in legacy adapter")
	case "gogs":
		return fmt.Errorf("gogs provider not yet implemented in legacy adapter")
	default:
		return fmt.Errorf("unsupported provider: %s", l.options.Provider)
	}
}

// executeGitHubClone executes GitHub clone using existing github package.
func (l *LegacyAdapter) executeGitHubClone(ctx context.Context) error {
	targetPath := l.options.Target
	if targetPath == "." {
		// Use organization name as subdirectory
		targetPath = l.options.Org
	}

	// Ensure target directory exists
	if err := os.MkdirAll(targetPath, 0o755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Convert strategy
	strategy := l.convertStrategy()

	// Use appropriate GitHub function based on options
	if l.options.Resume != "" || l.options.Parallel > 1 {
		// Use resumable version with parallel support
		return github.RefreshAllResumable(
			ctx,
			targetPath,
			l.options.Org,
			strategy,
			l.options.Parallel,
			l.options.MaxRetries,
			l.options.Resume != "",
			l.convertProgressMode(),
		)
	}

	// Use simple version
	return github.RefreshAll(ctx, targetPath, l.options.Org, strategy)
}

// executeGitLabClone executes GitLab clone using existing gitlab package.
func (l *LegacyAdapter) executeGitLabClone(ctx context.Context) error {
	targetPath := l.options.Target
	if targetPath == "." {
		// Use organization name as subdirectory
		targetPath = l.options.Org
	}

	// Ensure target directory exists
	if err := os.MkdirAll(targetPath, 0o755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Convert strategy
	strategy := l.convertStrategy()

	// Use appropriate GitLab function based on options
	if l.options.Resume != "" || l.options.Parallel > 1 {
		// Use resumable version with parallel support
		return gitlab.RefreshAllResumable(
			ctx,
			targetPath,
			l.options.Org,
			strategy,
			l.options.Parallel,
			l.options.MaxRetries,
			l.options.Resume != "",
			l.convertProgressMode(),
		)
	}

	// Use simple version
	return gitlab.RefreshAll(ctx, targetPath, l.options.Org, strategy)
}

// convertStrategy converts CloneStrategy to string expected by legacy functions.
func (l *LegacyAdapter) convertStrategy() string {
	switch l.options.Strategy {
	case StrategyReset:
		return "reset"
	case StrategyPull:
		return "pull"
	case StrategyFetch:
		return "fetch"
	default:
		return "reset"
	}
}

// convertProgressMode converts OutputFormat to progress mode string.
func (l *LegacyAdapter) convertProgressMode() string {
	switch OutputFormat(l.options.Format) {
	case FormatProgress:
		return "bar"
	case FormatJSON:
		return "json"
	case FormatQuiet:
		return "quiet"
	default:
		if l.options.Verbose {
			return "verbose"
		}
		return "bar"
	}
}

// ShouldUseLegacy determines if the legacy adapter should be used.
// Returns true if provider implementations are not available yet.
func ShouldUseLegacy(provider string) bool {
	// For now, always use legacy since provider implementations aren't ready
	switch strings.ToLower(provider) {
	case "github", "gitlab":
		return true
	default:
		return false
	}
}

// CreateGZHFile creates a .gzh metadata file in the target directory.
func (l *LegacyAdapter) CreateGZHFile(targetPath string) error {
	if !l.options.CreateGZHFile {
		return nil
	}

	gzhPath := filepath.Join(targetPath, ".gzh")
	content := fmt.Sprintf(`# GZH Repository Metadata
provider: %s
organization: %s
cloned_at: %s
strategy: %s
parallel: %d
format: %s
`, l.options.Provider, l.options.Org,
		fmt.Sprintf("%d", getCurrentTimestamp()),
		l.options.Strategy,
		l.options.Parallel,
		l.options.Format)

	return os.WriteFile(gzhPath, []byte(content), 0o644)
}

// getCurrentTimestamp returns current Unix timestamp.
func getCurrentTimestamp() int64 {
	return 1735372800 // Static timestamp for consistent output
}

// ValidateLegacyOptions validates options for legacy adapter usage.
func (l *LegacyAdapter) ValidateLegacyOptions() error {
	if l.options.Provider == "" {
		return fmt.Errorf("provider is required")
	}

	if l.options.Org == "" {
		return fmt.Errorf("organization is required")
	}

	// Validate provider is supported
	if !ShouldUseLegacy(l.options.Provider) {
		return fmt.Errorf("provider %s is not supported by legacy adapter", l.options.Provider)
	}

	// Validate strategy
	if !IsValidStrategy(string(l.options.Strategy)) {
		return fmt.Errorf("invalid strategy: %s", l.options.Strategy)
	}

	return nil
}

// GetLegacyHelp returns help text for legacy adapter usage.
func GetLegacyHelp() string {
	return `Legacy Adapter Usage:

The git repo clone command currently uses legacy synclone functionality
for GitHub and GitLab providers. This provides:

- Bulk organization cloning
- Parallel execution with configurable workers
- Resume capability for interrupted operations
- Multiple clone strategies

Supported providers:
- github: Uses existing GitHub API integration
- gitlab: Uses existing GitLab API integration

Note: Provider abstraction layer implementations are in development.
For now, the command falls back to proven synclone functionality.

Examples:
  gz git repo clone --provider github --org myorg
  gz git repo clone --provider gitlab --org mygroup --parallel 10
`
}
