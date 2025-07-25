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
	RetryAttempts    int           `json:"retry_attempts" yaml:"retryAttempts"`
	RetryDelay       time.Duration `json:"retry_delay" yaml:"retryDelay"`
	BufferSize       int           `json:"buffer_size" yaml:"bufferSize"`
	MaxRepositories  int           `json:"max_repositories" yaml:"maxRepositories"`
	EnableStreaming  bool          `json:"enable_streaming" yaml:"enableStreaming"`
	ProgressInterval time.Duration `json:"progress_interval" yaml:"progressInterval"`
}

// HTTPClientConfig represents configuration for HTTP clients.
type HTTPClientConfig struct {
	Timeout               time.Duration `json:"timeout" yaml:"timeout"`
	MaxIdleConns          int           `json:"maxIdleConns" yaml:"maxIdleConns"`
	IdleConnTimeout       time.Duration `json:"idleConnTimeout" yaml:"idleConnTimeout"`
	TLSHandshakeTimeout   time.Duration `json:"tlsHandshakeTimeout" yaml:"tlsHandshakeTimeout"`
	ExpectContinueTimeout time.Duration `json:"expectContinueTimeout" yaml:"expectContinueTimeout"`
	MinTLSVersion         string        `json:"minTlsVersion" yaml:"minTlsVersion"`
	UserAgent             string        `json:"userAgent" yaml:"userAgent"`
}

// WorkerPoolConfig represents configuration for worker pools.
type WorkerPoolConfig struct {
	CloneWorkers     int           `json:"cloneWorkers" yaml:"cloneWorkers"`
	UpdateWorkers    int           `json:"updateWorkers" yaml:"updateWorkers"`
	ConfigWorkers    int           `json:"configWorkers" yaml:"configWorkers"`
	OperationTimeout time.Duration `json:"operationTimeout" yaml:"operationTimeout"`
	RetryAttempts    int           `json:"retryAttempts" yaml:"retryAttempts"`
	RetryDelay       time.Duration `json:"retryDelay" yaml:"retryDelay"`
}

// AuthConfig represents configuration for authentication.
type AuthConfig struct {
	TokenMinLength    int           `json:"tokenMinLength" yaml:"tokenMinLength"`
	ValidationTimeout time.Duration `json:"validationTimeout" yaml:"validationTimeout"`
	CacheDuration     time.Duration `json:"cacheDuration" yaml:"cacheDuration"`
	MaxRetryAttempts  int           `json:"maxRetryAttempts" yaml:"maxRetryAttempts"`
}

// LoggingConfig represents configuration for logging.
type LoggingConfig struct {
	Level      string `json:"level" yaml:"level"`
	Format     string `json:"format" yaml:"format"`
	Output     string `json:"output" yaml:"output"`
	MaxSize    int    `json:"maxSize" yaml:"maxSize"`
	MaxBackups int    `json:"maxBackups" yaml:"maxBackups"`
	MaxAge     int    `json:"maxAge" yaml:"maxAge"`
}

// GitHubConfig represents GitHub-specific configuration.
type GitHubConfig struct {
	Token           string        `json:"token" yaml:"token"`
	BaseURL         string        `json:"baseUrl" yaml:"baseUrl"`
	Timeout         time.Duration `json:"timeout" yaml:"timeout"`
	RetryAttempts   int           `json:"retryAttempts" yaml:"retryAttempts"`
	RetryDelay      time.Duration `json:"retryDelay" yaml:"retryDelay"`
	RateLimit       int           `json:"rateLimit" yaml:"rateLimit"`
	EnableStreaming bool          `json:"enableStreaming" yaml:"enableStreaming"`
}

// GitLabConfig represents GitLab-specific configuration.
type GitLabConfig struct {
	Token           string        `json:"token" yaml:"token"`
	BaseURL         string        `json:"baseUrl" yaml:"baseUrl"`
	Timeout         time.Duration `json:"timeout" yaml:"timeout"`
	RetryAttempts   int           `json:"retryAttempts" yaml:"retryAttempts"`
	RetryDelay      time.Duration `json:"retryDelay" yaml:"retryDelay"`
	RateLimit       int           `json:"rateLimit" yaml:"rateLimit"`
	EnableStreaming bool          `json:"enableStreaming" yaml:"enableStreaming"`
}

// GiteaConfig represents Gitea-specific configuration.
type GiteaConfig struct {
	Token           string        `json:"token" yaml:"token"`
	BaseURL         string        `json:"baseUrl" yaml:"baseUrl"`
	Timeout         time.Duration `json:"timeout" yaml:"timeout"`
	RetryAttempts   int           `json:"retryAttempts" yaml:"retryAttempts"`
	RetryDelay      time.Duration `json:"retryDelay" yaml:"retryDelay"`
	RateLimit       int           `json:"rateLimit" yaml:"rateLimit"`
	EnableStreaming bool          `json:"enableStreaming" yaml:"enableStreaming"`
}

// SecurityConfig represents security-related configuration.
type SecurityConfig struct {
	EnableInputValidation   bool          `json:"enableInputValidation" yaml:"enableInputValidation"`
	EnableCommandValidation bool          `json:"enableCommandValidation" yaml:"enableCommandValidation"`
	EnableHTTPSOnly         bool          `json:"enableHttpsOnly" yaml:"enableHttpsOnly"`
	TokenValidationTimeout  time.Duration `json:"tokenValidationTimeout" yaml:"tokenValidationTimeout"`
	MaxRequestSize          int64         `json:"maxRequestSize" yaml:"maxRequestSize"`
	AllowedHosts            []string      `json:"allowedHosts" yaml:"allowedHosts"`
	TrustedCertificates     []string      `json:"trustedCertificates" yaml:"trustedCertificates"`
}

// PerformanceConfig represents performance-related configuration.
type PerformanceConfig struct {
	EnableStreaming         bool          `json:"enableStreaming" yaml:"enableStreaming"`
	EnableConnectionPooling bool          `json:"enableConnectionPooling" yaml:"enableConnectionPooling"`
	EnableCompression       bool          `json:"enableCompression" yaml:"enableCompression"`
	MaxConcurrentOperations int           `json:"maxConcurrentOperations" yaml:"maxConcurrentOperations"`
	MemoryLimit             int64         `json:"memoryLimit" yaml:"memoryLimit"`
	DiskSpaceLimit          int64         `json:"diskSpaceLimit" yaml:"diskSpaceLimit"`
	OperationTimeout        time.Duration `json:"operationTimeout" yaml:"operationTimeout"`
	HeartbeatInterval       time.Duration `json:"heartbeatInterval" yaml:"heartbeatInterval"`
}

// MonitoringConfig represents monitoring and metrics configuration.
type MonitoringConfig struct {
	EnableMetrics       bool          `json:"enableMetrics" yaml:"enableMetrics"`
	EnableTracing       bool          `json:"enableTracing" yaml:"enableTracing"`
	EnableHealthChecks  bool          `json:"enableHealthChecks" yaml:"enableHealthChecks"`
	MetricsPort         int           `json:"metricsPort" yaml:"metricsPort"`
	MetricsPath         string        `json:"metricsPath" yaml:"metricsPath"`
	HealthCheckInterval time.Duration `json:"healthCheckInterval" yaml:"healthCheckInterval"`
	RetentionPeriod     time.Duration `json:"retentionPeriod" yaml:"retentionPeriod"`
}
