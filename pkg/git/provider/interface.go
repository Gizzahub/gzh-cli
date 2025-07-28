// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"time"
)

// GitProvider is the main interface that all Git platforms must implement.
// It provides a unified interface for repository lifecycle management across
// different Git hosting platforms like GitHub, GitLab, Gitea, and Gogs.
type GitProvider interface {
	// Basic provider information
	GetName() string
	GetCapabilities() []Capability
	GetBaseURL() string

	// Authentication
	Authenticate(ctx context.Context, creds Credentials) error
	ValidateToken(ctx context.Context) (*TokenInfo, error)

	// Repository management
	RepositoryManager

	// Webhook management (if supported)
	WebhookManager

	// Event management (if supported)
	EventManager

	// Health and monitoring
	HealthChecker
}

// RepositoryManager defines repository-related operations.
type RepositoryManager interface {
	// Repository CRUD operations
	ListRepositories(ctx context.Context, opts ListOptions) (*RepositoryList, error)
	GetRepository(ctx context.Context, id string) (*Repository, error)
	CreateRepository(ctx context.Context, req CreateRepoRequest) (*Repository, error)
	UpdateRepository(ctx context.Context, id string, updates UpdateRepoRequest) (*Repository, error)
	DeleteRepository(ctx context.Context, id string) error

	// Repository state management
	ArchiveRepository(ctx context.Context, id string) error
	UnarchiveRepository(ctx context.Context, id string) error

	// Repository operations
	CloneRepository(ctx context.Context, repo Repository, target string, opts CloneOptions) error
	ForkRepository(ctx context.Context, id string, opts ForkOptions) (*Repository, error)

	// Search and discovery
	SearchRepositories(ctx context.Context, query SearchQuery) (*SearchResult, error)
}

// WebhookManager defines webhook-related operations.
type WebhookManager interface {
	// Webhook CRUD operations
	ListWebhooks(ctx context.Context, repoID string) ([]Webhook, error)
	GetWebhook(ctx context.Context, repoID, webhookID string) (*Webhook, error)
	CreateWebhook(ctx context.Context, repoID string, webhook CreateWebhookRequest) (*Webhook, error)
	UpdateWebhook(ctx context.Context, repoID, webhookID string, updates UpdateWebhookRequest) (*Webhook, error)
	DeleteWebhook(ctx context.Context, repoID, webhookID string) error

	// Webhook testing and validation
	TestWebhook(ctx context.Context, repoID, webhookID string) (*WebhookTestResult, error)
	ValidateWebhookURL(ctx context.Context, url string) error
}

// EventManager defines event-related operations.
type EventManager interface {
	// Event querying
	ListEvents(ctx context.Context, opts EventListOptions) ([]Event, error)
	GetEvent(ctx context.Context, eventID string) (*Event, error)

	// Event handling
	ProcessEvent(ctx context.Context, event Event) error
	RegisterEventHandler(eventType string, handler EventHandler) error

	// Event streaming (if supported)
	StreamEvents(ctx context.Context, opts StreamOptions) (<-chan Event, error)
}

// HealthChecker defines health monitoring operations.
type HealthChecker interface {
	// Health status
	HealthCheck(ctx context.Context) (*HealthStatus, error)
	GetRateLimit(ctx context.Context) (*RateLimit, error)

	// Performance metrics
	GetMetrics(ctx context.Context) (*ProviderMetrics, error)
}

// EventHandler is a function that processes events.
type EventHandler func(ctx context.Context, event Event) error

// Capability represents a feature that a provider supports.
type Capability string

const (
	CapabilityRepositories     Capability = "repositories"
	CapabilityWebhooks         Capability = "webhooks"
	CapabilityEvents           Capability = "events"
	CapabilityIssues           Capability = "issues"
	CapabilityPullRequests     Capability = "pull_requests"
	CapabilityMergeRequests    Capability = "merge_requests"
	CapabilityWiki             Capability = "wiki"
	CapabilityProjects         Capability = "projects"
	CapabilityActions          Capability = "actions"
	CapabilityCICD             Capability = "ci_cd"
	CapabilityPackages         Capability = "packages"
	CapabilityReleases         Capability = "releases"
	CapabilityOrganizations    Capability = "organizations"
	CapabilityUsers            Capability = "users"
	CapabilityTeams            Capability = "teams"
	CapabilityPermissions      Capability = "permissions"
	CapabilityBranchProtection Capability = "branch_protection"
	CapabilitySecurityAlerts   Capability = "security_alerts"
	CapabilityDependabot       Capability = "dependabot"
)

// Credentials represents authentication credentials for a provider.
type Credentials struct {
	Type     CredentialType         `json:"type"`
	Token    string                 `json:"token,omitempty"`
	Username string                 `json:"username,omitempty"`
	Password string                 `json:"password,omitempty"`
	APIKey   string                 `json:"api_key,omitempty"`
	Extra    map[string]interface{} `json:"extra,omitempty"`
}

// CredentialType represents the type of authentication.
type CredentialType string

const (
	CredentialTypeToken  CredentialType = "token"
	CredentialTypeBasic  CredentialType = "basic"
	CredentialTypeAPIKey CredentialType = "api_key"
	CredentialTypeOAuth  CredentialType = "oauth"
	CredentialTypeSSHKey CredentialType = "ssh_key"
)

// TokenInfo represents information about an authentication token.
type TokenInfo struct {
	Valid       bool      `json:"valid"`
	Scopes      []string  `json:"scopes"`
	User        string    `json:"user"`
	Email       string    `json:"email"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	Permissions []string  `json:"permissions"`
	RateLimit   RateLimit `json:"rate_limit"`
}

// HealthStatus represents the health status of a provider.
type HealthStatus struct {
	Status      HealthStatusType       `json:"status"`
	LastChecked time.Time              `json:"last_checked"`
	Latency     time.Duration          `json:"latency"`
	Message     string                 `json:"message,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// HealthStatusType represents the health status.
type HealthStatusType string

const (
	HealthStatusHealthy   HealthStatusType = "healthy"
	HealthStatusDegraded  HealthStatusType = "degraded"
	HealthStatusUnhealthy HealthStatusType = "unhealthy"
	HealthStatusUnknown   HealthStatusType = "unknown"
)

// RateLimit represents API rate limiting information.
type RateLimit struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	Reset     time.Time `json:"reset"`
	Used      int       `json:"used"`
	Resource  string    `json:"resource,omitempty"`
}

// ProviderMetrics represents performance metrics for a provider.
type ProviderMetrics struct {
	RequestCount   int64         `json:"request_count"`
	ErrorCount     int64         `json:"error_count"`
	AverageLatency time.Duration `json:"average_latency"`
	SuccessRate    float64       `json:"success_rate"`
	LastError      string        `json:"last_error,omitempty"`
	LastErrorTime  time.Time     `json:"last_error_time,omitempty"`
	CollectedAt    time.Time     `json:"collected_at"`
}

// StreamOptions defines options for event streaming.
type StreamOptions struct {
	EventTypes []string               `json:"event_types,omitempty"`
	Filters    map[string]interface{} `json:"filters,omitempty"`
	BufferSize int                    `json:"buffer_size,omitempty"`
}

// WebhookTestResult represents the result of a webhook test.
type WebhookTestResult struct {
	Success      bool          `json:"success"`
	StatusCode   int           `json:"status_code"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error,omitempty"`
	RequestID    string        `json:"request_id,omitempty"`
}
