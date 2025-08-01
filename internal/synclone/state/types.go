// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package state

import (
	"time"
)

// OperationStatus represents the status of a synclone operation.
type OperationStatus string

const (
	StatusPending    OperationStatus = "pending"
	StatusInProgress OperationStatus = "in_progress"
	StatusCompleted  OperationStatus = "completed"
	StatusFailed     OperationStatus = "failed"
	StatusCanceled   OperationStatus = "canceled"
)

// OperationState represents the comprehensive state of a synclone operation.
type OperationState struct {
	ID           string               `json:"id"`
	StartTime    time.Time            `json:"start_time"`
	LastUpdate   time.Time            `json:"last_update"`
	Status       OperationStatus      `json:"status"`
	Config       *Config              `json:"config"`
	Progress     OperationProgress    `json:"progress"`
	Repositories map[string]RepoState `json:"repositories"`
	Errors       []OperationError     `json:"errors"`
	Metrics      OperationMetrics     `json:"metrics"`
}

// Config represents the configuration used for the operation.
type Config struct {
	Version   string                 `json:"version"`
	Global    map[string]interface{} `json:"global"`
	Providers map[string]interface{} `json:"providers"`
	SyncMode  map[string]interface{} `json:"sync_mode"`
}

// OperationProgress tracks the overall progress of the operation.
type OperationProgress struct {
	TotalRepos      int       `json:"total_repos"`
	CompletedRepos  int       `json:"completed_repos"`
	FailedRepos     int       `json:"failed_repos"`
	PendingRepos    int       `json:"pending_repos"`
	PercentComplete float64   `json:"percent_complete"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time,omitempty"`
	EstimatedETA    time.Time `json:"estimated_eta,omitempty"`
}

// RepoState represents the state of an individual repository.
type RepoState struct {
	Name         string    `json:"name"`
	Status       string    `json:"status"` // pending, cloning, completed, failed
	AttemptCount int       `json:"attempt_count"`
	LastError    string    `json:"last_error,omitempty"`
	StartTime    time.Time `json:"start_time,omitempty"`
	EndTime      time.Time `json:"end_time,omitempty"`
	BytesCloned  int64     `json:"bytes_cloned"`
	URL          string    `json:"url"`
	LocalPath    string    `json:"local_path"`
	Provider     string    `json:"provider"`
	Org          string    `json:"org"`
}

// OperationError represents an error that occurred during the operation.
type OperationError struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	RepoName  string    `json:"repo_name,omitempty"`
	Retryable bool      `json:"retryable"`
}

// OperationMetrics tracks performance metrics for the operation.
type OperationMetrics struct {
	TotalDuration      time.Duration     `json:"total_duration"`
	AvgRepoCloneTime   time.Duration     `json:"avg_repo_clone_time"`
	FastestRepo        RepoMetric        `json:"fastest_repo"`
	SlowestRepo        RepoMetric        `json:"slowest_repo"`
	TotalBytesCloned   int64             `json:"total_bytes_cloned"`
	AvgCloneSpeed      float64           `json:"avg_clone_speed_mbps"`
	ConcurrencyMetrics ConcurrencyMetric `json:"concurrency_metrics"`
	RetryStatistics    RetryStatistics   `json:"retry_statistics"`
}

// RepoMetric represents performance metrics for a single repository.
type RepoMetric struct {
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
	Size     int64         `json:"size"`
}

// ConcurrencyMetric tracks concurrency-related metrics.
type ConcurrencyMetric struct {
	MaxWorkers       int     `json:"max_workers"`
	AvgActiveWorkers float64 `json:"avg_active_workers"`
	WorkerEfficiency float64 `json:"worker_efficiency"`
}

// RetryStatistics tracks retry-related statistics.
type RetryStatistics struct {
	TotalRetries      int            `json:"total_retries"`
	RetriedRepos      int            `json:"retried_repos"`
	RetryReasons      map[string]int `json:"retry_reasons"`
	AvgRetriesPerRepo float64        `json:"avg_retries_per_repo"`
}

// EnvironmentState represents the environment state at the time of operation.
type EnvironmentState struct {
	NetworkStatus    string            `json:"network_status"`
	DiskSpace        int64             `json:"disk_space_bytes"`
	MemoryUsage      int64             `json:"memory_usage_bytes"`
	CPUUsage         float64           `json:"cpu_usage_percent"`
	EnvironmentVars  map[string]string `json:"environment_vars"`
	GitVersion       string            `json:"git_version"`
	WorkingDirectory string            `json:"working_directory"`
}

// UpdateProgress updates the progress information.
func (os *OperationState) UpdateProgress() {
	completed := 0
	failed := 0
	pending := 0

	for _, repo := range os.Repositories {
		switch repo.Status {
		case "completed":
			completed++
		case "failed":
			failed++
		default:
			pending++
		}
	}

	os.Progress.CompletedRepos = completed
	os.Progress.FailedRepos = failed
	os.Progress.PendingRepos = pending
	os.Progress.TotalRepos = len(os.Repositories)

	if os.Progress.TotalRepos > 0 {
		os.Progress.PercentComplete = float64(completed) / float64(os.Progress.TotalRepos) * 100
	}

	os.LastUpdate = time.Now()
}

// AddError adds an error to the operation state.
func (os *OperationState) AddError(errorType, message, repoName string, retryable bool) {
	os.Errors = append(os.Errors, OperationError{
		Timestamp: time.Now(),
		Type:      errorType,
		Message:   message,
		RepoName:  repoName,
		Retryable: retryable,
	})
	os.LastUpdate = time.Now()
}

// IsResumable returns true if the operation can be resumed.
func (os *OperationState) IsResumable() bool {
	return os.Status == StatusInProgress || os.Status == StatusFailed
}

// GetRetryableRepos returns repositories that can be retried.
func (os *OperationState) GetRetryableRepos() []string {
	var retryable []string

	for name, repo := range os.Repositories {
		if repo.Status == "failed" && repo.AttemptCount < 3 {
			retryable = append(retryable, name)
		}
	}

	return retryable
}

// GetPendingRepos returns repositories that haven't been processed.
func (os *OperationState) GetPendingRepos() []string {
	var pending []string

	for name, repo := range os.Repositories {
		if repo.Status == "pending" {
			pending = append(pending, name)
		}
	}

	return pending
}

// CalculateETA estimates the completion time based on current progress.
func (os *OperationState) CalculateETA() time.Time {
	if os.Progress.CompletedRepos == 0 {
		return time.Time{} // Cannot estimate without any completed repos
	}

	elapsedTime := time.Since(os.Progress.StartTime)
	avgTimePerRepo := elapsedTime / time.Duration(os.Progress.CompletedRepos)
	remainingRepos := os.Progress.TotalRepos - os.Progress.CompletedRepos

	estimatedRemainingTime := avgTimePerRepo * time.Duration(remainingRepos)
	return time.Now().Add(estimatedRemainingTime)
}
