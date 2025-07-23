// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

import (
	"time"
)

// BulkCloneConfig represents configuration for bulk clone operations.
type BulkCloneConfig struct {
	Concurrency      int           `json:"concurrency" yaml:"concurrency"`
	Timeout          time.Duration `json:"timeout" yaml:"timeout"`
	RetryAttempts    int           `json:"retry_attempts" yaml:"retry_attempts"`
	RetryDelay       time.Duration `json:"retry_delay" yaml:"retry_delay"`
	BufferSize       int           `json:"buffer_size" yaml:"buffer_size"`
	MaxRepositories  int           `json:"max_repositories" yaml:"max_repositories"`
	EnableStreaming  bool          `json:"enable_streaming" yaml:"enable_streaming"`
	ProgressInterval time.Duration `json:"progress_interval" yaml:"progress_interval"`
}

// HTTPClientConfig represents configuration for HTTP clients.
type HTTPClientConfig struct {
	Timeout               time.Duration `json:"timeout" yaml:"timeout"`
	MaxIdleConns          int           `json:"max_idle_conns" yaml:"max_idle_conns"`
	IdleConnTimeout       time.Duration `json:"idle_conn_timeout" yaml:"idle_conn_timeout"`
	TLSHandshakeTimeout   time.Duration `json:"tls_handshake_timeout" yaml:"tls_handshake_timeout"`
	ExpectContinueTimeout time.Duration `json:"expect_continue_timeout" yaml:"expect_continue_timeout"`
	MinTLSVersion         string        `json:"min_tls_version" yaml:"min_tls_version"`
	UserAgent             string        `json:"user_agent" yaml:"user_agent"`
}

// WorkerPoolConfig represents configuration for worker pools.
type WorkerPoolConfig struct {
	CloneWorkers     int           `json:"clone_workers" yaml:"clone_workers"`
	UpdateWorkers    int           `json:"update_workers" yaml:"update_workers"`
	ConfigWorkers    int           `json:"config_workers" yaml:"config_workers"`
	OperationTimeout time.Duration `json:"operation_timeout" yaml:"operation_timeout"`
	RetryAttempts    int           `json:"retry_attempts" yaml:"retry_attempts"`
	RetryDelay       time.Duration `json:"retry_delay" yaml:"retry_delay"`
}

// AuthConfig represents configuration for authentication.
type AuthConfig struct {
	TokenMinLength    int           `json:"token_min_length" yaml:"token_min_length"`
	ValidationTimeout time.Duration `json:"validation_timeout" yaml:"validation_timeout"`
	CacheDuration     time.Duration `json:"cache_duration" yaml:"cache_duration"`
	MaxRetryAttempts  int           `json:"max_retry_attempts" yaml:"max_retry_attempts"`
}

// LoggingConfig represents configuration for logging.
type LoggingConfig struct {
	Level      string `json:"level" yaml:"level"`
	Format     string `json:"format" yaml:"format"`
	Output     string `json:"output" yaml:"output"`
	MaxSize    int    `json:"max_size" yaml:"max_size"`
	MaxBackups int    `json:"max_backups" yaml:"max_backups"`
	MaxAge     int    `json:"max_age" yaml:"max_age"`
}

// GitHubConfig represents GitHub-specific configuration.
type GitHubConfig struct {
	Token           string        `json:"token" yaml:"token"`
	BaseURL         string        `json:"base_url" yaml:"base_url"`
	Timeout         time.Duration `json:"timeout" yaml:"timeout"`
	RetryAttempts   int           `json:"retry_attempts" yaml:"retry_attempts"`
	RetryDelay      time.Duration `json:"retry_delay" yaml:"retry_delay"`
	RateLimit       int           `json:"rate_limit" yaml:"rate_limit"`
	EnableStreaming bool          `json:"enable_streaming" yaml:"enable_streaming"`
}

// GitLabConfig represents GitLab-specific configuration.
type GitLabConfig struct {
	Token           string        `json:"token" yaml:"token"`
	BaseURL         string        `json:"base_url" yaml:"base_url"`
	Timeout         time.Duration `json:"timeout" yaml:"timeout"`
	RetryAttempts   int           `json:"retry_attempts" yaml:"retry_attempts"`
	RetryDelay      time.Duration `json:"retry_delay" yaml:"retry_delay"`
	RateLimit       int           `json:"rate_limit" yaml:"rate_limit"`
	EnableStreaming bool          `json:"enable_streaming" yaml:"enable_streaming"`
}

// GiteaConfig represents Gitea-specific configuration.
type GiteaConfig struct {
	Token           string        `json:"token" yaml:"token"`
	BaseURL         string        `json:"base_url" yaml:"base_url"`
	Timeout         time.Duration `json:"timeout" yaml:"timeout"`
	RetryAttempts   int           `json:"retry_attempts" yaml:"retry_attempts"`
	RetryDelay      time.Duration `json:"retry_delay" yaml:"retry_delay"`
	RateLimit       int           `json:"rate_limit" yaml:"rate_limit"`
	EnableStreaming bool          `json:"enable_streaming" yaml:"enable_streaming"`
}

// SecurityConfig represents security-related configuration.
type SecurityConfig struct {
	EnableInputValidation   bool          `json:"enable_input_validation" yaml:"enable_input_validation"`
	EnableCommandValidation bool          `json:"enable_command_validation" yaml:"enable_command_validation"`
	EnableHTTPSOnly         bool          `json:"enable_https_only" yaml:"enable_https_only"`
	TokenValidationTimeout  time.Duration `json:"token_validation_timeout" yaml:"token_validation_timeout"`
	MaxRequestSize          int64         `json:"max_request_size" yaml:"max_request_size"`
	AllowedHosts            []string      `json:"allowed_hosts" yaml:"allowed_hosts"`
	TrustedCertificates     []string      `json:"trusted_certificates" yaml:"trusted_certificates"`
}

// PerformanceConfig represents performance-related configuration.
type PerformanceConfig struct {
	EnableStreaming         bool          `json:"enable_streaming" yaml:"enable_streaming"`
	EnableConnectionPooling bool          `json:"enable_connection_pooling" yaml:"enable_connection_pooling"`
	EnableCompression       bool          `json:"enable_compression" yaml:"enable_compression"`
	MaxConcurrentOperations int           `json:"max_concurrent_operations" yaml:"max_concurrent_operations"`
	MemoryLimit             int64         `json:"memory_limit" yaml:"memory_limit"`
	DiskSpaceLimit          int64         `json:"disk_space_limit" yaml:"disk_space_limit"`
	OperationTimeout        time.Duration `json:"operation_timeout" yaml:"operation_timeout"`
	HeartbeatInterval       time.Duration `json:"heartbeat_interval" yaml:"heartbeat_interval"`
}

// MonitoringConfig represents monitoring and metrics configuration.
type MonitoringConfig struct {
	EnableMetrics       bool          `json:"enable_metrics" yaml:"enable_metrics"`
	EnableTracing       bool          `json:"enable_tracing" yaml:"enable_tracing"`
	EnableHealthChecks  bool          `json:"enable_health_checks" yaml:"enable_health_checks"`
	MetricsPort         int           `json:"metrics_port" yaml:"metrics_port"`
	MetricsPath         string        `json:"metrics_path" yaml:"metrics_path"`
	HealthCheckInterval time.Duration `json:"health_check_interval" yaml:"health_check_interval"`
	RetentionPeriod     time.Duration `json:"retention_period" yaml:"retention_period"`
}
