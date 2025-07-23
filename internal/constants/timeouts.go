// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package constants provides common constants for timeouts, limits, and configuration values
// throughout the application to avoid magic numbers and ensure consistency.
package constants

import "time"

// HTTP and Network Timeouts.
const (
	// DefaultHTTPTimeout is the default timeout for HTTP requests.
	DefaultHTTPTimeout = 30 * time.Second

	// DefaultDialTimeout is the default timeout for network dialing.
	DefaultDialTimeout = 10 * time.Second

	// DefaultKeepAlive is the default keep-alive duration for connections.
	DefaultKeepAlive = 30 * time.Second

	// DefaultIdleConnTimeout is the default timeout for idle connections.
	DefaultIdleConnTimeout = 90 * time.Second

	// LongHTTPTimeout is used for operations that may take longer.
	LongHTTPTimeout = 60 * time.Second

	// MediumHTTPTimeout is used for medium-duration operations.
	MediumHTTPTimeout = 45 * time.Second

	// ShortHTTPTimeout is used for quick operations.
	ShortHTTPTimeout = 10 * time.Second

	// VeryLongTimeout is used for very long operations like large file downloads.
	VeryLongTimeout = 5 * time.Minute

	// ExtraLongTimeout is used for extremely long operations.
	ExtraLongTimeout = 10 * time.Minute
)

// Operation Timeouts.
const (
	// DefaultTestTimeout is the default timeout for tests.
	DefaultTestTimeout = 5 * time.Second

	// LongTestTimeout is used for longer-running tests.
	LongTestTimeout = 30 * time.Second

	// ConfigSyncInterval is the default configuration sync interval.
	ConfigSyncInterval = 1 * time.Hour

	// HealthCheckInterval is the default health check interval.
	HealthCheckInterval = 60 * time.Second

	// ShortSleepDuration is used for brief waits in operations.
	ShortSleepDuration = 100 * time.Millisecond

	// VeryShortSleepDuration is used for very brief waits.
	VeryShortSleepDuration = 10 * time.Millisecond

	// RetryDelay is the default delay between retry attempts.
	RetryDelay = 1 * time.Second

	// BackoffMultiplier is used for exponential backoff calculations.
	BackoffMultiplier = 2
)

// Resource Limits.
const (
	// MaxConcurrentConnections is the maximum number of concurrent connections.
	MaxConcurrentConnections = 100

	// MaxIdleConnections is the maximum number of idle connections to maintain.
	MaxIdleConnections = 100

	// MaxIdleConnectionsPerHost is the maximum idle connections per host.
	MaxIdleConnectionsPerHost = 10

	// MaxConnectionsPerHost is the maximum connections per host.
	MaxConnectionsPerHost = 50

	// GitHubMaxIdleConnectionsPerHost is optimized for GitHub API.
	GitHubMaxIdleConnectionsPerHost = 20

	// GitLabMaxIdleConnectionsPerHost is optimized for GitLab API.
	GitLabMaxIdleConnectionsPerHost = 15

	// GiteaMaxIdleConnectionsPerHost is optimized for Gitea API.
	GiteaMaxIdleConnectionsPerHost = 10

	// DefaultMaxIdleConns is the default maximum idle connections.
	DefaultMaxIdleConns = 100

	// DefaultTLSHandshakeTimeout is the default TLS handshake timeout.
	DefaultTLSHandshakeTimeout = 10 * time.Second

	// DefaultExpectContinueTimeout is the default expect continue timeout.
	DefaultExpectContinueTimeout = 1 * time.Second

	// DefaultUserAgent is the default user agent string.
	DefaultUserAgent = "gzh-manager-go/1.0.0"
)

// Retry Configuration.
const (
	// DefaultMaxRetries is the default maximum number of retry attempts.
	DefaultMaxRetries = 3

	// MaxRetryAttempts is the maximum allowed retry attempts.
	MaxRetryAttempts = 10

	// DefaultParallelism is the default number of parallel operations.
	DefaultParallelism = 5

	// MaxParallelism is the maximum allowed parallelism.
	MaxParallelism = 100

	// DefaultConcurrency is the default concurrency level.
	DefaultConcurrency = 5

	// DefaultRetryAttempts is the default number of retry attempts.
	DefaultRetryAttempts = 3

	// DefaultRetryDelay is the default delay between retries.
	DefaultRetryDelay = 2 * time.Second

	// DefaultBufferSize is the default buffer size for operations.
	DefaultBufferSize = 100

	// MaxRepositories is the maximum number of repositories to process.
	MaxRepositories = 10000

	// ProgressUpdateInterval is the interval for progress updates.
	ProgressUpdateInterval = 1 * time.Second
)

// Validation Limits.
const (
	// MinTokenLength is the minimum length for authentication tokens.
	MinTokenLength = 8

	// MaxTokenLength is the maximum length for authentication tokens.
	MaxTokenLength = 500

	// MaxOrganizationNameLength is the maximum length for organization names.
	MaxOrganizationNameLength = 100

	// MaxRepositoryNameLength is the maximum length for repository names.
	MaxRepositoryNameLength = 100

	// MaxTopicLength is the maximum length for repository topics.
	MaxTopicLength = 50

	// MaxLanguageFilterLength is the maximum length for language filters.
	MaxLanguageFilterLength = 50

	// MaxConfigValueLength is the maximum length for configuration values.
	MaxConfigValueLength = 10000
)

// Memory and Storage Limits.
const (
	// BytesPerKB represents bytes in a kilobyte.
	BytesPerKB = 1024

	// BytesPerMB represents bytes in a megabyte.
	BytesPerMB = 1024 * 1024

	// BytesPerGB represents bytes in a gigabyte.
	BytesPerGB = 1024 * 1024 * 1024

	// DefaultMemoryLimitMB is the default memory limit in megabytes.
	DefaultMemoryLimitMB = 256

	// MaxMemoryLimitMB is the maximum memory limit in megabytes.
	MaxMemoryLimitMB = 4096
)

// Default Port Numbers.
const (
	// DefaultRedisPort is the default Redis port.
	DefaultRedisPort = 6379

	// DefaultHTTPPort is the default HTTP port.
	DefaultHTTPPort = 80

	// DefaultHTTPSPort is the default HTTPS port.
	DefaultHTTPSPort = 443

	// MinPortNumber is the minimum valid port number.
	MinPortNumber = 1

	// MaxPortNumber is the maximum valid port number.
	MaxPortNumber = 65535
)

// TLS and Security Constants.
const (
	// MaxRedirectsAllowed is the maximum number of redirects to follow.
	MaxRedirectsAllowed = 10

	// DefaultTLSVersion represents the minimum TLS version to use.
	// This maps to tls.VersionTLS12.
	DefaultTLSVersion = 0x0303
)
