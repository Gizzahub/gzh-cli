// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SyncTracker tracks synchronization progress and state.
type SyncTracker struct {
	ID          string                  `json:"id"`
	StartedAt   time.Time               `json:"started_at"`
	CompletedAt *time.Time              `json:"completed_at,omitempty"`
	Source      string                  `json:"source"`
	Destination string                  `json:"destination"`
	Status      SyncStatus              `json:"status"`
	Progress    map[string]SyncProgress `json:"progress"`
	Statistics  SyncStatistics          `json:"statistics"`
	SyncOptions     SyncOptions                 `json:"options"`
	Errors      []SyncError             `json:"errors,omitempty"`
}

// SyncStatus represents the overall status of a sync operation.
type SyncStatus string

const (
	StatusPending    SyncStatus = "pending"
	StatusInProgress SyncStatus = "in_progress"
	StatusCompleted  SyncStatus = "completed"
	StatusFailed     SyncStatus = "failed"
	StatusCancelled  SyncStatus = "cancelled"
)

// SyncStatistics contains overall sync statistics.
type SyncStatistics struct {
	TotalRepositories     int           `json:"total_repositories"`
	CompletedRepositories int           `json:"completed_repositories"`
	FailedRepositories    int           `json:"failed_repositories"`
	SkippedRepositories   int           `json:"skipped_repositories"`
	TotalDuration         time.Duration `json:"total_duration"`
	AverageDuration       time.Duration `json:"average_duration"`
	BytesTransferred      int64         `json:"bytes_transferred"`
}

// SyncError represents an error that occurred during synchronization.
type SyncError struct {
	Repository string    `json:"repository"`
	Component  string    `json:"component"`
	Message    string    `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
}

// NewSyncTracker creates a new sync tracker.
func NewSyncTracker(id, source, destination string, opts SyncOptions) *SyncTracker {
	return &SyncTracker{
		ID:          id,
		StartedAt:   time.Now(),
		Source:      source,
		Destination: destination,
		Status:      StatusPending,
		Progress:    make(map[string]SyncProgress),
		SyncOptions:     opts,
		Statistics:  SyncStatistics{},
	}
}

// UpdateProgress updates the progress for a specific component.
func (t *SyncTracker) UpdateProgress(component string, completed, failed int) {
	if t.Progress == nil {
		t.Progress = make(map[string]SyncProgress)
	}

	progress := t.Progress[component]
	progress.Completed = completed
	progress.Failed = failed
	progress.UpdatedAt = time.Now()
	t.Progress[component] = progress

	// Update overall statistics
	t.updateStatistics()

	// Save to disk
	if err := t.save(); err != nil {
		fmt.Printf("Warning: Failed to save sync progress: %v\n", err)
	}
}

// SetStatus updates the sync status.
func (t *SyncTracker) SetStatus(status SyncStatus) {
	t.Status = status
	if status == StatusCompleted || status == StatusFailed || status == StatusCancelled {
		now := time.Now()
		t.CompletedAt = &now
		t.Statistics.TotalDuration = now.Sub(t.StartedAt)
	}

	if err := t.save(); err != nil {
		fmt.Printf("Warning: Failed to save sync status: %v\n", err)
	}
}

// AddError adds an error to the tracker.
func (t *SyncTracker) AddError(repository, component, message string) {
	error := SyncError{
		Repository: repository,
		Component:  component,
		Message:    message,
		Timestamp:  time.Now(),
	}
	t.Errors = append(t.Errors, error)

	if err := t.save(); err != nil {
		fmt.Printf("Warning: Failed to save sync error: %v\n", err)
	}
}

// updateStatistics updates the overall statistics based on progress.
func (t *SyncTracker) updateStatistics() {
	totalCompleted := 0
	totalFailed := 0
	totalRepositories := 0

	for _, progress := range t.Progress {
		totalCompleted += progress.Completed
		totalFailed += progress.Failed
		if progress.Total > totalRepositories {
			totalRepositories = progress.Total
		}
	}

	t.Statistics.CompletedRepositories = totalCompleted
	t.Statistics.FailedRepositories = totalFailed
	t.Statistics.TotalRepositories = totalRepositories
	t.Statistics.SkippedRepositories = totalRepositories - totalCompleted - totalFailed

	// Calculate average duration
	if totalCompleted > 0 && !t.StartedAt.IsZero() {
		elapsed := time.Since(t.StartedAt)
		t.Statistics.AverageDuration = elapsed / time.Duration(totalCompleted)
	}
}

// GetOverallProgress returns the overall progress percentage.
func (t *SyncTracker) GetOverallProgress() float64 {
	if t.Statistics.TotalRepositories == 0 {
		return 0
	}
	completed := t.Statistics.CompletedRepositories + t.Statistics.FailedRepositories
	return float64(completed) / float64(t.Statistics.TotalRepositories) * 100
}

// IsCompleted returns true if the sync operation is completed.
func (t *SyncTracker) IsCompleted() bool {
	return t.Status == StatusCompleted || t.Status == StatusFailed || t.Status == StatusCancelled
}

// HasErrors returns true if there are any errors.
func (t *SyncTracker) HasErrors() bool {
	return len(t.Errors) > 0 || t.Statistics.FailedRepositories > 0
}

// GetSummary returns a summary of the sync operation.
func (t *SyncTracker) GetSummary() SyncSummary {
	return SyncSummary{
		ID:                t.ID,
		Source:            t.Source,
		Destination:       t.Destination,
		Status:            t.Status,
		StartedAt:         t.StartedAt,
		CompletedAt:       t.CompletedAt,
		Duration:          t.Statistics.TotalDuration,
		TotalRepositories: t.Statistics.TotalRepositories,
		Completed:         t.Statistics.CompletedRepositories,
		Failed:            t.Statistics.FailedRepositories,
		Skipped:           t.Statistics.SkippedRepositories,
		Progress:          t.GetOverallProgress(),
		ErrorCount:        len(t.Errors),
	}
}

// save saves the tracker state to disk.
func (t *SyncTracker) save() error {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tracker: %w", err)
	}

	trackingFile := filepath.Join(getSyncDir(), t.ID+".json")
	if err := os.WriteFile(trackingFile, data, 0o644); err != nil {
		return fmt.Errorf("failed to write tracker file: %w", err)
	}

	return nil
}

// LoadSyncTracker loads a sync tracker from disk.
func LoadSyncTracker(id string) (*SyncTracker, error) {
	trackingFile := filepath.Join(getSyncDir(), id+".json")

	data, err := os.ReadFile(trackingFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read tracker file: %w", err)
	}

	var tracker SyncTracker
	if err := json.Unmarshal(data, &tracker); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tracker: %w", err)
	}

	return &tracker, nil
}

// ListSyncTrackers lists all sync trackers.
func ListSyncTrackers() ([]*SyncTracker, error) {
	syncDir := getSyncDir()

	entries, err := os.ReadDir(syncDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*SyncTracker{}, nil
		}
		return nil, fmt.Errorf("failed to read sync directory: %w", err)
	}

	var trackers []*SyncTracker
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		id := strings.TrimSuffix(entry.Name(), ".json")
		tracker, err := LoadSyncTracker(id)
		if err != nil {
			fmt.Printf("Warning: Failed to load tracker %s: %v\n", id, err)
			continue
		}

		trackers = append(trackers, tracker)
	}

	return trackers, nil
}

// DeleteSyncTracker deletes a sync tracker from disk.
func DeleteSyncTracker(id string) error {
	trackingFile := filepath.Join(getSyncDir(), id+".json")
	if err := os.Remove(trackingFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete tracker file: %w", err)
	}
	return nil
}

// CleanupOldTrackers removes trackers older than the specified duration.
func CleanupOldTrackers(maxAge time.Duration) error {
	trackers, err := ListSyncTrackers()
	if err != nil {
		return fmt.Errorf("failed to list trackers: %w", err)
	}

	cutoff := time.Now().Add(-maxAge)
	cleaned := 0

	for _, tracker := range trackers {
		if tracker.StartedAt.Before(cutoff) {
			if err := DeleteSyncTracker(tracker.ID); err != nil {
				fmt.Printf("Warning: Failed to delete old tracker %s: %v\n", tracker.ID, err)
			} else {
				cleaned++
			}
		}
	}

	if cleaned > 0 {
		fmt.Printf("Cleaned up %d old sync trackers\n", cleaned)
	}

	return nil
}

// getSyncDir returns the directory for storing sync tracking files.
func getSyncDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "gzh-sync")
	}

	syncDir := filepath.Join(homeDir, ".config", "gzh-manager", "sync")
	if err := os.MkdirAll(syncDir, 0o755); err != nil {
		return filepath.Join(os.TempDir(), "gzh-sync")
	}

	return syncDir
}

// SyncSummary provides a summary view of sync operations.
type SyncSummary struct {
	ID                string        `json:"id"`
	Source            string        `json:"source"`
	Destination       string        `json:"destination"`
	Status            SyncStatus    `json:"status"`
	StartedAt         time.Time     `json:"started_at"`
	CompletedAt       *time.Time    `json:"completed_at,omitempty"`
	Duration          time.Duration `json:"duration"`
	TotalRepositories int           `json:"total_repositories"`
	Completed         int           `json:"completed"`
	Failed            int           `json:"failed"`
	Skipped           int           `json:"skipped"`
	Progress          float64       `json:"progress"`
	ErrorCount        int           `json:"error_count"`
}

// PrintSyncSummary prints a formatted summary of sync operations.
func PrintSyncSummary(summaries []SyncSummary) {
	if len(summaries) == 0 {
		fmt.Println("No sync operations found")
		return
	}

	fmt.Printf("Recent Sync Operations:\n")
	fmt.Printf("%-10s %-20s %-20s %-12s %-8s %-10s\n",
		"ID", "Source", "Destination", "Status", "Progress", "Duration")
	fmt.Printf("%-10s %-20s %-20s %-12s %-8s %-10s\n",
		"----------", "--------------------", "--------------------",
		"------------", "--------", "----------")

	for _, summary := range summaries {
		duration := "N/A"
		if summary.Duration > 0 {
			duration = summary.Duration.Truncate(time.Second).String()
		}

		progress := fmt.Sprintf("%.1f%%", summary.Progress)

		fmt.Printf("%-10s %-20s %-20s %-12s %-8s %-10s\n",
			summary.ID[:min(10, len(summary.ID))],
			truncateString(summary.Source, 20),
			truncateString(summary.Destination, 20),
			string(summary.Status),
			progress,
			duration)
	}
}

// truncateString truncates a string to the specified length.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
