// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package provider

import (
	"time"
)

// Repository represents a platform-independent repository.
type Repository struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	FullName      string         `json:"full_name"`
	Owner         Owner          `json:"owner"`
	Description   string         `json:"description"`
	Private       bool           `json:"private"`
	Archived      bool           `json:"archived"`
	Disabled      bool           `json:"disabled"`
	Fork          bool           `json:"fork"`
	Template      bool           `json:"template"`
	CloneURL      string         `json:"clone_url"`
	SSHURL        string         `json:"ssh_url"`
	HTMLURL       string         `json:"html_url"`
	DefaultBranch string         `json:"default_branch"`
	Language      string         `json:"language"`
	Size          int64          `json:"size"`
	Topics        []string       `json:"topics"`
	Visibility    VisibilityType `json:"visibility"`
	License       License        `json:"license"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	PushedAt      time.Time      `json:"pushed_at"`

	// Statistics
	Stars    int `json:"stars"`
	Forks    int `json:"forks"`
	Watchers int `json:"watchers"`
	Issues   int `json:"open_issues"`

	// Provider-specific data
	ProviderType string                 `json:"provider_type"`
	ProviderData map[string]interface{} `json:"provider_data"`
}

// Owner represents the owner of a repository (user or organization).
type Owner struct {
	ID        string    `json:"id"`
	Login     string    `json:"login"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Type      OwnerType `json:"type"`
	HTMLURL   string    `json:"html_url"`
	AvatarURL string    `json:"avatar_url"`
}

// OwnerType represents the type of repository owner.
type OwnerType string

const (
	OwnerTypeUser         OwnerType = "user"
	OwnerTypeOrganization OwnerType = "organization"
	OwnerTypeGroup        OwnerType = "group"
)

// VisibilityType represents repository visibility.
type VisibilityType string

const (
	VisibilityPublic   VisibilityType = "public"
	VisibilityPrivate  VisibilityType = "private"
	VisibilityInternal VisibilityType = "internal"
)

// License represents repository license information.
type License struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	SPDXID string `json:"spdx_id"`
	URL    string `json:"url"`
	NodeID string `json:"node_id"`
}

// RepositoryList represents a paginated list of repositories.
type RepositoryList struct {
	Repositories []Repository `json:"repositories"`
	TotalCount   int          `json:"total_count"`
	Page         int          `json:"page"`
	PerPage      int          `json:"per_page"`
	HasNext      bool         `json:"has_next"`
	HasPrev      bool         `json:"has_prev"`
}

// ListOptions defines options for listing repositories.
type ListOptions struct {
	// Scope
	Organization string `json:"organization,omitempty"`
	User         string `json:"user,omitempty"`

	// Filtering
	Visibility   VisibilityType `json:"visibility,omitempty"`
	Type         string         `json:"type,omitempty"` // all, owner, member
	Archived     *bool          `json:"archived,omitempty"`
	Fork         *bool          `json:"fork,omitempty"`
	Language     string         `json:"language,omitempty"`
	Topic        string         `json:"topic,omitempty"`
	MinStars     int            `json:"min_stars,omitempty"`
	MaxStars     int            `json:"max_stars,omitempty"`
	UpdatedSince time.Time      `json:"updated_since,omitempty"`

	// Sorting and pagination
	Sort      string `json:"sort,omitempty"`      // created, updated, pushed, full_name
	Direction string `json:"direction,omitempty"` // asc, desc
	Page      int    `json:"page,omitempty"`
	PerPage   int    `json:"per_page,omitempty"`
}

// CreateRepoRequest represents a request to create a repository.
type CreateRepoRequest struct {
	Name              string         `json:"name"`
	Description       string         `json:"description"`
	Homepage          string         `json:"homepage,omitempty"`
	Private           bool           `json:"private"`
	Visibility        VisibilityType `json:"visibility,omitempty"`
	HasIssues         bool           `json:"has_issues"`
	HasProjects       bool           `json:"has_projects"`
	HasWiki           bool           `json:"has_wiki"`
	HasDownloads      bool           `json:"has_downloads"`
	TeamID            int            `json:"team_id,omitempty"`
	AutoInit          bool           `json:"auto_init"`
	GitignoreTemplate string         `json:"gitignore_template,omitempty"`
	LicenseTemplate   string         `json:"license_template,omitempty"`
	AllowSquashMerge  bool           `json:"allow_squash_merge"`
	AllowMergeCommit  bool           `json:"allow_merge_commit"`
	AllowRebaseMerge  bool           `json:"allow_rebase_merge"`
	AllowAutoMerge    bool           `json:"allow_auto_merge"`
	DefaultBranch     string         `json:"default_branch,omitempty"`
	Topics            []string       `json:"topics,omitempty"`

	// Template options
	TemplateOwner      string `json:"template_owner,omitempty"`
	TemplateRepo       string `json:"template_repo,omitempty"`
	IncludeAllBranches bool   `json:"include_all_branches,omitempty"`
}

// UpdateRepoRequest represents a request to update a repository.
type UpdateRepoRequest struct {
	Name             *string        `json:"name,omitempty"`
	Description      *string        `json:"description,omitempty"`
	Homepage         *string        `json:"homepage,omitempty"`
	Private          *bool          `json:"private,omitempty"`
	Visibility       VisibilityType `json:"visibility,omitempty"`
	HasIssues        *bool          `json:"has_issues,omitempty"`
	HasProjects      *bool          `json:"has_projects,omitempty"`
	HasWiki          *bool          `json:"has_wiki,omitempty"`
	HasDownloads     *bool          `json:"has_downloads,omitempty"`
	DefaultBranch    *string        `json:"default_branch,omitempty"`
	AllowSquashMerge *bool          `json:"allow_squash_merge,omitempty"`
	AllowMergeCommit *bool          `json:"allow_merge_commit,omitempty"`
	AllowRebaseMerge *bool          `json:"allow_rebase_merge,omitempty"`
	AllowAutoMerge   *bool          `json:"allow_auto_merge,omitempty"`
	Archived         *bool          `json:"archived,omitempty"`
	Topics           []string       `json:"topics,omitempty"`
}

// CloneOptions defines options for cloning repositories.
type CloneOptions struct {
	Strategy     string        `json:"strategy"`      // reset, pull, fetch
	Protocol     string        `json:"protocol"`      // https, ssh
	Parallel     int           `json:"parallel"`      // number of parallel workers
	Depth        int           `json:"depth"`         // clone depth (0 = full)
	SingleBranch bool          `json:"single_branch"` // clone only default branch
	Branch       string        `json:"branch"`        // specific branch to clone
	Resume       bool          `json:"resume"`        // resume interrupted operation
	DryRun       bool          `json:"dry_run"`       // preview without cloning
	Timeout      time.Duration `json:"timeout"`       // operation timeout
}

// ForkOptions defines options for forking repositories.
type ForkOptions struct {
	Organization      string `json:"organization,omitempty"`
	Name              string `json:"name,omitempty"`
	DefaultBranchOnly bool   `json:"default_branch_only"`
}

// SearchQuery represents a repository search query.
type SearchQuery struct {
	Query        string         `json:"query"`
	Sort         string         `json:"sort,omitempty"`  // stars, forks, updated
	Order        string         `json:"order,omitempty"` // asc, desc
	Language     string         `json:"language,omitempty"`
	User         string         `json:"user,omitempty"`
	Organization string         `json:"organization,omitempty"`
	Repository   string         `json:"repository,omitempty"`
	Topic        string         `json:"topic,omitempty"`
	License      string         `json:"license,omitempty"`
	Fork         *bool          `json:"fork,omitempty"`
	Archived     *bool          `json:"archived,omitempty"`
	Visibility   VisibilityType `json:"visibility,omitempty"`
	Created      string         `json:"created,omitempty"` // date range like ">2021-01-01"
	Updated      string         `json:"updated,omitempty"` // date range like ">2021-01-01"
	Size         string         `json:"size,omitempty"`    // size range like ">1000"
	Stars        string         `json:"stars,omitempty"`   // star range like ">100"
	Forks        string         `json:"forks,omitempty"`   // fork range like ">10"
	Page         int            `json:"page,omitempty"`
	PerPage      int            `json:"per_page,omitempty"`
}

// SearchResult represents search results.
type SearchResult struct {
	TotalCount        int          `json:"total_count"`
	IncompleteResults bool         `json:"incomplete_results"`
	Repositories      []Repository `json:"repositories"`
	Page              int          `json:"page"`
	PerPage           int          `json:"per_page"`
	HasNext           bool         `json:"has_next"`
	HasPrev           bool         `json:"has_prev"`
}

// Webhook represents a webhook configuration.
type Webhook struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	URL          string          `json:"url"`
	Events       []string        `json:"events"`
	Active       bool            `json:"active"`
	Config       WebhookConfig   `json:"config"`
	LastResponse WebhookResponse `json:"last_response,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// WebhookConfig represents webhook configuration.
type WebhookConfig struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Secret      string `json:"secret,omitempty"`
	InsecureSSL bool   `json:"insecure_ssl"`
}

// WebhookResponse represents the last webhook response.
type WebhookResponse struct {
	Code      int       `json:"code"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// CreateWebhookRequest represents a request to create a webhook.
type CreateWebhookRequest struct {
	Name   string        `json:"name"`
	Config WebhookConfig `json:"config"`
	Events []string      `json:"events"`
	Active bool          `json:"active"`
}

// UpdateWebhookRequest represents a request to update a webhook.
type UpdateWebhookRequest struct {
	Config *WebhookConfig `json:"config,omitempty"`
	Events []string       `json:"events,omitempty"`
	Active *bool          `json:"active,omitempty"`
}

// Event represents a platform event.
type Event struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Actor      Actor                  `json:"actor"`
	Repository Repository             `json:"repository"`
	Payload    map[string]interface{} `json:"payload"`
	Public     bool                   `json:"public"`
	CreatedAt  time.Time              `json:"created_at"`

	// Provider-specific data
	ProviderType string                 `json:"provider_type"`
	ProviderData map[string]interface{} `json:"provider_data"`
}

// Actor represents the user who triggered an event.
type Actor struct {
	ID        string `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
}

// EventListOptions defines options for listing events.
type EventListOptions struct {
	Repository   string    `json:"repository,omitempty"`
	Organization string    `json:"organization,omitempty"`
	User         string    `json:"user,omitempty"`
	EventType    string    `json:"event_type,omitempty"`
	Since        time.Time `json:"since,omitempty"`
	Until        time.Time `json:"until,omitempty"`
	Page         int       `json:"page,omitempty"`
	PerPage      int       `json:"per_page,omitempty"`
}
