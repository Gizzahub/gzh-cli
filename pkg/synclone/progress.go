// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package bulkclone

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// RepositoryProgress represents the progress of a single repository operation.
type RepositoryProgress struct {
	Name        string
	Status      ProgressStatus
	StartTime   time.Time
	UpdatedTime time.Time
	Message     string
	Error       string
	Progress    float64 // 0.0 to 1.0
}

// ProgressStatus represents the status of a repository operation.
type ProgressStatus string

const (
	// StatusPending indicates the repository operation is pending.
	StatusPending ProgressStatus = "pending"
	// StatusStarted indicates the repository operation has started.
	StatusStarted ProgressStatus = "started"
	// StatusCloning indicates the repository is being cloned.
	StatusCloning ProgressStatus = "cloning"
	// StatusPulling indicates the repository is being pulled.
	StatusPulling ProgressStatus = "pulling"
	// StatusFetching indicates the repository is being fetched.
	StatusFetching ProgressStatus = "fetching"
	// StatusResetting indicates the repository is being reset.
	StatusResetting ProgressStatus = "resetting"
	// StatusCompleted indicates the repository operation completed successfully.
	StatusCompleted ProgressStatus = "completed"
	// StatusFailed indicates the repository operation failed.
	StatusFailed ProgressStatus = "failed"
	// StatusSkipped indicates the repository operation was skipped.
	StatusSkipped ProgressStatus = "skipped"
)

// ProgressTracker tracks progress for multiple repositories.
type ProgressTracker struct {
	mu           sync.RWMutex
	repositories map[string]*RepositoryProgress
	totalRepos   int
	startTime    time.Time
	displayMode  DisplayMode
}

// DisplayMode controls how progress is displayed.
type DisplayMode string

const (
	// DisplayModeCompact shows single line with overall progress.
	DisplayModeCompact DisplayMode = "compact" // Single line with overall progress
	// DisplayModeDetailed shows multiple lines with per-repo status.
	DisplayModeDetailed DisplayMode = "detailed" // Multiple lines with per-repo status
	// DisplayModeQuiet shows no progress display.
	DisplayModeQuiet DisplayMode = "quiet" // No progress display
)

// NewProgressTracker creates a new progress tracker.
func NewProgressTracker(repoNames []string, displayMode DisplayMode) *ProgressTracker {
	tracker := &ProgressTracker{
		repositories: make(map[string]*RepositoryProgress),
		totalRepos:   len(repoNames),
		startTime:    time.Now(),
		displayMode:  displayMode,
	}

	// Initialize all repositories as pending
	for _, name := range repoNames {
		tracker.repositories[name] = &RepositoryProgress{
			Name:      name,
			Status:    StatusPending,
			StartTime: time.Now(),
			Progress:  0.0,
		}
	}

	return tracker
}

// UpdateRepository updates the progress for a specific repository.
func (pt *ProgressTracker) UpdateRepository(name string, status ProgressStatus, message string, progress float64) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if repo, exists := pt.repositories[name]; exists {
		repo.Status = status
		repo.Message = message
		repo.Progress = progress
		repo.UpdatedTime = time.Now()

		// Set start time when operation begins
		if status == StatusStarted && repo.StartTime.IsZero() {
			repo.StartTime = time.Now()
		}
	}
}

// SetRepositoryError sets an error for a specific repository.
func (pt *ProgressTracker) SetRepositoryError(name, err string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if repo, exists := pt.repositories[name]; exists {
		repo.Status = StatusFailed
		repo.Error = err
		repo.Progress = 0.0
		repo.UpdatedTime = time.Now()
	}
}

// CompleteRepository marks a repository as completed.
func (pt *ProgressTracker) CompleteRepository(name, message string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if repo, exists := pt.repositories[name]; exists {
		repo.Status = StatusCompleted
		repo.Message = message
		repo.Progress = 1.0
		repo.UpdatedTime = time.Now()
	}
}

// GetOverallProgress returns the overall progress statistics.
func (pt *ProgressTracker) GetOverallProgress() (completed, failed, pending int, progressPercent float64) {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	var totalProgress float64

	for _, repo := range pt.repositories {
		switch repo.Status {
		case StatusCompleted:
			completed++
			totalProgress += 1.0
		case StatusFailed:
			failed++
		case StatusPending:
			pending++
		case StatusStarted, StatusCloning, StatusPulling, StatusFetching, StatusResetting:
			// In-progress repositories contribute partial progress
			totalProgress += repo.Progress
		case StatusSkipped:
			// Skipped repositories are considered completed for progress calculation
			completed++
			totalProgress += 1.0
		default:
			// Unknown status, treat as pending
			pending++
		}
	}

	if pt.totalRepos > 0 {
		progressPercent = totalProgress / float64(pt.totalRepos) * 100
	}

	return completed, failed, pending, progressPercent
}

// GetRepositoryProgress returns progress for a specific repository.
func (pt *ProgressTracker) GetRepositoryProgress(name string) (*RepositoryProgress, bool) {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	repo, exists := pt.repositories[name]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid race conditions
	return &RepositoryProgress{
		Name:        repo.Name,
		Status:      repo.Status,
		StartTime:   repo.StartTime,
		UpdatedTime: repo.UpdatedTime,
		Message:     repo.Message,
		Error:       repo.Error,
		Progress:    repo.Progress,
	}, true
}

// GetAllRepositories returns all repository progress information.
func (pt *ProgressTracker) GetAllRepositories() []RepositoryProgress {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	repos := make([]RepositoryProgress, 0, len(pt.repositories))
	for _, repo := range pt.repositories {
		repos = append(repos, *repo) // Copy the struct
	}

	return repos
}

// RenderProgress renders the current progress based on display mode.
func (pt *ProgressTracker) RenderProgress() string {
	switch pt.displayMode {
	case DisplayModeQuiet:
		return ""
	case DisplayModeDetailed:
		return pt.renderDetailedProgress()
	case DisplayModeCompact:
		return pt.renderCompactProgress()
	default:
		return pt.renderCompactProgress()
	}
}

// renderCompactProgress renders a single-line progress display.
func (pt *ProgressTracker) renderCompactProgress() string {
	completed, failed, pending, progressPercent := pt.GetOverallProgress()

	elapsed := time.Since(pt.startTime).Round(time.Second)

	// Create progress bar
	barWidth := 30
	filledWidth := int(float64(barWidth) * progressPercent / 100)
	emptyWidth := barWidth - filledWidth

	bar := strings.Repeat("█", filledWidth) + strings.Repeat("░", emptyWidth)

	return fmt.Sprintf("[%s] %.1f%% (%d/%d) • ✓ %d • ✗ %d • ⏳ %d • %s",
		bar, progressPercent, completed+failed, pt.totalRepos,
		completed, failed, pending, elapsed)
}

// renderDetailedProgress renders multi-line progress with per-repository status.
func (pt *ProgressTracker) renderDetailedProgress() string {
	completed, failed, _, progressPercent := pt.GetOverallProgress()
	elapsed := time.Since(pt.startTime).Round(time.Second)

	var output strings.Builder

	// Overall progress header
	output.WriteString(fmt.Sprintf("Overall Progress: %.1f%% (%d/%d) • %s\n",
		progressPercent, completed+failed, pt.totalRepos, elapsed))

	// Repository details
	repos := pt.GetAllRepositories()

	// Group by status
	statusGroups := make(map[ProgressStatus][]RepositoryProgress)
	for _, repo := range repos {
		statusGroups[repo.Status] = append(statusGroups[repo.Status], repo)
	}

	// Show active operations first
	activeStatuses := []ProgressStatus{StatusStarted, StatusCloning, StatusPulling, StatusFetching, StatusResetting}
	for _, status := range activeStatuses {
		if repos, exists := statusGroups[status]; exists {
			output.WriteString(fmt.Sprintf("\n%s (%d):\n", getStatusEmoji(status), len(repos)))

			for _, repo := range repos {
				duration := time.Since(repo.StartTime).Round(time.Second)
				output.WriteString(fmt.Sprintf("  %s (%.0f%%) - %s\n",
					repo.Name, repo.Progress*100, duration))
			}
		}
	}

	// Show completed repositories (limited to last 5)
	if completedRepos, exists := statusGroups[StatusCompleted]; exists {
		output.WriteString(fmt.Sprintf("\n✅ Completed (%d):\n", len(completedRepos)))

		// Show last 5 completed
		start := len(completedRepos) - 5
		if start < 0 {
			start = 0
		}

		for i := start; i < len(completedRepos); i++ {
			repo := completedRepos[i]
			duration := repo.UpdatedTime.Sub(repo.StartTime).Round(time.Second)
			output.WriteString(fmt.Sprintf("  %s - %s\n", repo.Name, duration))
		}

		if len(completedRepos) > 5 {
			output.WriteString(fmt.Sprintf("  ... and %d more\n", len(completedRepos)-5))
		}
	}

	// Show failed repositories
	if failedRepos, exists := statusGroups[StatusFailed]; exists {
		output.WriteString(fmt.Sprintf("\n❌ Failed (%d):\n", len(failedRepos)))

		for _, repo := range failedRepos {
			output.WriteString(fmt.Sprintf("  %s - %s\n", repo.Name, repo.Error))
		}
	}

	// Show pending count
	if pendingRepos, exists := statusGroups[StatusPending]; exists {
		output.WriteString(fmt.Sprintf("\n⏳ Pending: %d repositories\n", len(pendingRepos)))
	}

	return output.String()
}

// getStatusEmoji returns an emoji for the given status.
func getStatusEmoji(status ProgressStatus) string {
	switch status {
	case StatusPending:
		return "⏳ Pending"
	case StatusStarted:
		return "🚀 Started"
	case StatusCloning:
		return "📥 Cloning"
	case StatusPulling:
		return "🔄 Pulling"
	case StatusFetching:
		return "📡 Fetching"
	case StatusResetting:
		return "🔄 Resetting"
	case StatusCompleted:
		return "✅ Completed"
	case StatusFailed:
		return "❌ Failed"
	case StatusSkipped:
		return "⏭️ Skipped"
	default:
		return "❓ Unknown"
	}
}

// GetDuration returns the total duration of the operation.
func (pt *ProgressTracker) GetDuration() time.Duration {
	return time.Since(pt.startTime)
}

// GetETA estimates the time to completion based on current progress.
func (pt *ProgressTracker) GetETA() time.Duration {
	_, _, _, progressPercent := pt.GetOverallProgress()

	if progressPercent <= 0 {
		return 0
	}

	elapsed := time.Since(pt.startTime)
	totalEstimated := elapsed * 100 / time.Duration(progressPercent)
	remaining := totalEstimated - elapsed

	if remaining < 0 {
		return 0
	}

	return remaining
}

// IsCompleted returns true if all repositories have been processed.
func (pt *ProgressTracker) IsCompleted() bool {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	for _, repo := range pt.repositories {
		if repo.Status != StatusCompleted && repo.Status != StatusFailed && repo.Status != StatusSkipped {
			return false
		}
	}

	return true
}

// GetSummary returns a summary of the operation.
func (pt *ProgressTracker) GetSummary() string {
	completed, failed, pending, progressPercent := pt.GetOverallProgress()
	duration := pt.GetDuration().Round(time.Second)

	return fmt.Sprintf("Summary: %.1f%% complete • %d successful • %d failed • %d pending • Duration: %s",
		progressPercent, completed, failed, pending, duration)
}
