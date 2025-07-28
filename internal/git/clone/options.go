// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package clone

import (
	"regexp"
	"time"
)

// CloneOptions represents options for repository cloning operations.
type CloneOptions struct {
	// Provider configuration
	Provider string `json:"provider"`
	Org      string `json:"org"`
	Target   string `json:"target"`
	Config   string `json:"config,omitempty"`

	// Execution options
	Parallel   int           `json:"parallel"`
	Strategy   CloneStrategy `json:"strategy"`
	Resume     string        `json:"resume,omitempty"`
	Timeout    time.Duration `json:"timeout"`
	MaxRetries int           `json:"max_retries"`
	RetryDelay time.Duration `json:"retry_delay"`

	// Filtering options
	Match           string   `json:"match,omitempty"`
	Exclude         string   `json:"exclude,omitempty"`
	Visibility      string   `json:"visibility"`
	IncludeArchived bool     `json:"include_archived"`
	IncludeForks    bool     `json:"include_forks"`
	Language        string   `json:"language,omitempty"`
	Topics          []string `json:"topics,omitempty"`
	MinStars        int      `json:"min_stars"`
	MaxStars        int      `json:"max_stars"`
	UpdatedSince    string   `json:"updated_since,omitempty"`

	// Output and behavior
	Format         string `json:"format"`
	DryRun         bool   `json:"dry_run"`
	Quiet          bool   `json:"quiet"`
	Verbose        bool   `json:"verbose"`
	CleanupOrphans bool   `json:"cleanup_orphans"`
	CreateGZHFile  bool   `json:"create_gzh_file"`

	// Authentication
	Token    string `json:"token,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`

	// Git options
	Protocol     string `json:"protocol"`
	Depth        int    `json:"depth"`
	SingleBranch bool   `json:"single_branch"`
	Branch       string `json:"branch,omitempty"`

	// Compiled patterns (internal use)
	matchPattern   *regexp.Regexp `json:"-"`
	excludePattern *regexp.Regexp `json:"-"`
}

// CloneStrategy represents the strategy to use when cloning repositories.
type CloneStrategy string

const (
	// StrategyReset performs git reset --hard && git pull for existing repos
	StrategyReset CloneStrategy = "reset"
	// StrategyPull performs git pull (merge) for existing repos
	StrategyPull CloneStrategy = "pull"
	// StrategyFetch performs git fetch only for existing repos
	StrategyFetch CloneStrategy = "fetch"
)

// OutputFormat represents the output format for clone operations.
type OutputFormat string

const (
	// FormatProgress shows progress bars and status
	FormatProgress OutputFormat = "progress"
	// FormatJSON outputs structured JSON
	FormatJSON OutputFormat = "json"
	// FormatTable outputs tabular format
	FormatTable OutputFormat = "table"
	// FormatQuiet suppresses most output
	FormatQuiet OutputFormat = "quiet"
)

// Validate validates the clone options and compiles regex patterns.
func (opts *CloneOptions) Validate() error {
	// Validate required fields
	if opts.Provider == "" {
		return ErrMissingProvider
	}
	if opts.Org == "" {
		return ErrMissingOrganization
	}
	if opts.Target == "" {
		opts.Target = "."
	}

	// Validate strategy
	switch opts.Strategy {
	case StrategyReset, StrategyPull, StrategyFetch:
		// Valid strategies
	case "":
		opts.Strategy = StrategyReset // Default
	default:
		return ErrInvalidStrategy
	}

	// Validate format
	switch OutputFormat(opts.Format) {
	case FormatProgress, FormatJSON, FormatTable, FormatQuiet:
		// Valid formats
	case "":
		opts.Format = string(FormatProgress) // Default
	default:
		return ErrInvalidFormat
	}

	// Validate parallel workers
	if opts.Parallel <= 0 {
		opts.Parallel = 5 // Default
	}
	if opts.Parallel > 50 {
		opts.Parallel = 50 // Max limit
	}

	// Validate timeout
	if opts.Timeout <= 0 {
		opts.Timeout = 30 * time.Minute // Default
	}

	// Validate retry settings
	if opts.MaxRetries < 0 {
		opts.MaxRetries = 3 // Default
	}
	if opts.RetryDelay <= 0 {
		opts.RetryDelay = 1 * time.Second // Default
	}

	// Validate protocol
	if opts.Protocol == "" {
		opts.Protocol = "https" // Default
	}
	if opts.Protocol != "https" && opts.Protocol != "ssh" {
		return ErrInvalidProtocol
	}

	// Validate visibility
	if opts.Visibility == "" {
		opts.Visibility = "all" // Default
	}
	if opts.Visibility != "all" && opts.Visibility != "public" && opts.Visibility != "private" {
		return ErrInvalidVisibility
	}

	// Compile regex patterns
	if opts.Match != "" {
		pattern, err := regexp.Compile(opts.Match)
		if err != nil {
			return ErrInvalidMatchPattern
		}
		opts.matchPattern = pattern
	}

	if opts.Exclude != "" {
		pattern, err := regexp.Compile(opts.Exclude)
		if err != nil {
			return ErrInvalidExcludePattern
		}
		opts.excludePattern = pattern
	}

	return nil
}

// GetMatchPattern returns the compiled match pattern.
func (opts *CloneOptions) GetMatchPattern() *regexp.Regexp {
	return opts.matchPattern
}

// GetExcludePattern returns the compiled exclude pattern.
func (opts *CloneOptions) GetExcludePattern() *regexp.Regexp {
	return opts.excludePattern
}

// IsValidStrategy checks if the given strategy is valid.
func IsValidStrategy(strategy string) bool {
	switch CloneStrategy(strategy) {
	case StrategyReset, StrategyPull, StrategyFetch:
		return true
	default:
		return false
	}
}

// GetValidStrategies returns a list of valid clone strategies.
func GetValidStrategies() []string {
	return []string{
		string(StrategyReset),
		string(StrategyPull),
		string(StrategyFetch),
	}
}

// GetValidFormats returns a list of valid output formats.
func GetValidFormats() []string {
	return []string{
		string(FormatProgress),
		string(FormatJSON),
		string(FormatTable),
		string(FormatQuiet),
	}
}

// DefaultCloneOptions returns a new CloneOptions with default values.
func DefaultCloneOptions() *CloneOptions {
	return &CloneOptions{
		Parallel:        5,
		Strategy:        StrategyReset,
		Format:          string(FormatProgress),
		Timeout:         30 * time.Minute,
		MaxRetries:      3,
		RetryDelay:      1 * time.Second,
		Visibility:      "all",
		IncludeArchived: false,
		IncludeForks:    true,
		Protocol:        "https",
		CreateGZHFile:   true,
	}
}

// CloneRequest represents a single repository clone request.
type CloneRequest struct {
	Repository  RepositoryInfo `json:"repository"`
	TargetPath  string         `json:"target_path"`
	Options     *CloneOptions  `json:"options"`
	SessionID   string         `json:"session_id"`
	Attempt     int            `json:"attempt"`
	StartedAt   time.Time      `json:"started_at,omitempty"`
	CompletedAt time.Time      `json:"completed_at,omitempty"`
	Error       string         `json:"error,omitempty"`
}

// RepositoryInfo represents basic repository information for cloning.
type RepositoryInfo struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	FullName      string    `json:"full_name"`
	CloneURL      string    `json:"clone_url"`
	SSHURL        string    `json:"ssh_url"`
	Private       bool      `json:"private"`
	Archived      bool      `json:"archived"`
	Fork          bool      `json:"fork"`
	Language      string    `json:"language,omitempty"`
	Topics        []string  `json:"topics,omitempty"`
	Stars         int       `json:"stars"`
	Forks         int       `json:"forks"`
	UpdatedAt     time.Time `json:"updated_at"`
	DefaultBranch string    `json:"default_branch"`
}

// GetCloneURL returns the appropriate clone URL based on protocol.
func (r *RepositoryInfo) GetCloneURL(protocol string) string {
	switch protocol {
	case "ssh":
		if r.SSHURL != "" {
			return r.SSHURL
		}
		// Fallback to HTTPS if SSH URL is not available
		return r.CloneURL
	case "https":
		fallthrough
	default:
		return r.CloneURL
	}
}

// Matches checks if the repository matches the given options filters.
func (r *RepositoryInfo) Matches(opts *CloneOptions) bool {
	// Check visibility
	if opts.Visibility != "all" {
		if opts.Visibility == "public" && r.Private {
			return false
		}
		if opts.Visibility == "private" && !r.Private {
			return false
		}
	}

	// Check archived repositories
	if !opts.IncludeArchived && r.Archived {
		return false
	}

	// Check forks
	if !opts.IncludeForks && r.Fork {
		return false
	}

	// Check language
	if opts.Language != "" && r.Language != opts.Language {
		return false
	}

	// Check stars
	if opts.MinStars > 0 && r.Stars < opts.MinStars {
		return false
	}
	if opts.MaxStars > 0 && r.Stars > opts.MaxStars {
		return false
	}

	// Check topics
	if len(opts.Topics) > 0 {
		hasRequiredTopic := false
		for _, requiredTopic := range opts.Topics {
			for _, repoTopic := range r.Topics {
				if repoTopic == requiredTopic {
					hasRequiredTopic = true
					break
				}
			}
			if hasRequiredTopic {
				break
			}
		}
		if !hasRequiredTopic {
			return false
		}
	}

	// Check match pattern
	if opts.GetMatchPattern() != nil {
		if !opts.GetMatchPattern().MatchString(r.Name) && !opts.GetMatchPattern().MatchString(r.FullName) {
			return false
		}
	}

	// Check exclude pattern
	if opts.GetExcludePattern() != nil {
		if opts.GetExcludePattern().MatchString(r.Name) || opts.GetExcludePattern().MatchString(r.FullName) {
			return false
		}
	}

	return true
}
