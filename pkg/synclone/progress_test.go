//nolint:testpackage // White-box testing needed for internal function access
package bulkclone

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewProgressTracker(t *testing.T) {
	repos := []string{"repo1", "repo2", "repo3"}
	tracker := NewProgressTracker(repos, DisplayModeCompact)

	assert.Equal(t, 3, tracker.totalRepos)
	assert.Equal(t, DisplayModeCompact, tracker.displayMode)
	assert.Len(t, tracker.repositories, 3)

	// Check that all repositories start as pending
	for _, repoName := range repos {
		repo, exists := tracker.GetRepositoryProgress(repoName)
		assert.True(t, exists)
		assert.Equal(t, StatusPending, repo.Status)
		assert.Equal(t, 0.0, repo.Progress)
	}
}

func TestUpdateRepository(t *testing.T) {
	repos := []string{"repo1", "repo2"}
	tracker := NewProgressTracker(repos, DisplayModeCompact)

	// Update repository status
	tracker.UpdateRepository("repo1", StatusCloning, "Cloning repository", 0.5)

	repo, exists := tracker.GetRepositoryProgress("repo1")
	assert.True(t, exists)
	assert.Equal(t, StatusCloning, repo.Status)
	assert.Equal(t, "Cloning repository", repo.Message)
	assert.Equal(t, 0.5, repo.Progress)

	// Update non-existent repository should not crash
	tracker.UpdateRepository("nonexistent", StatusCompleted, "Done", 1.0)
}

func TestSetRepositoryError(t *testing.T) {
	repos := []string{"repo1"}
	tracker := NewProgressTracker(repos, DisplayModeCompact)

	tracker.SetRepositoryError("repo1", "Network error")

	repo, exists := tracker.GetRepositoryProgress("repo1")
	assert.True(t, exists)
	assert.Equal(t, StatusFailed, repo.Status)
	assert.Equal(t, "Network error", repo.Error)
	assert.Equal(t, 0.0, repo.Progress)
}

func TestCompleteRepository(t *testing.T) {
	repos := []string{"repo1"}
	tracker := NewProgressTracker(repos, DisplayModeCompact)

	tracker.CompleteRepository("repo1", "Successfully cloned")

	repo, exists := tracker.GetRepositoryProgress("repo1")
	assert.True(t, exists)
	assert.Equal(t, StatusCompleted, repo.Status)
	assert.Equal(t, "Successfully cloned", repo.Message)
	assert.Equal(t, 1.0, repo.Progress)
}

func TestGetOverallProgress(t *testing.T) {
	repos := []string{"repo1", "repo2", "repo3", "repo4"}
	tracker := NewProgressTracker(repos, DisplayModeCompact)

	// Initial state: all pending
	completed, failed, pending, progressPercent := tracker.GetOverallProgress()
	assert.Equal(t, 0, completed)
	assert.Equal(t, 0, failed)
	assert.Equal(t, 4, pending)
	assert.Equal(t, 0.0, progressPercent)

	// Complete one repository
	tracker.CompleteRepository("repo1", "Done")
	completed, failed, pending, progressPercent = tracker.GetOverallProgress()
	assert.Equal(t, 1, completed)
	assert.Equal(t, 0, failed)
	assert.Equal(t, 3, pending)
	assert.Equal(t, 25.0, progressPercent)

	// Fail one repository
	tracker.SetRepositoryError("repo2", "Error")
	completed, failed, pending, progressPercent = tracker.GetOverallProgress()
	assert.Equal(t, 1, completed)
	assert.Equal(t, 1, failed)
	assert.Equal(t, 2, pending)
	assert.Equal(t, 25.0, progressPercent)

	// Partially complete one repository
	tracker.UpdateRepository("repo3", StatusCloning, "In progress", 0.5)
	completed, failed, pending, progressPercent = tracker.GetOverallProgress()
	assert.Equal(t, 1, completed)
	assert.Equal(t, 1, failed)
	assert.Equal(t, 1, pending)            // repo4 is still pending
	assert.Equal(t, 37.5, progressPercent) // (1.0 + 0.0 + 0.5 + 0.0) / 4 * 100
}

func TestGetAllRepositories(t *testing.T) {
	repos := []string{"repo1", "repo2"}
	tracker := NewProgressTracker(repos, DisplayModeCompact)

	// Update some repositories
	tracker.UpdateRepository("repo1", StatusCloning, "Cloning", 0.3)
	tracker.CompleteRepository("repo2", "Done")

	allRepos := tracker.GetAllRepositories()
	assert.Len(t, allRepos, 2)

	// Check that we get copies, not references
	for _, repo := range allRepos {
		switch repo.Name {
		case "repo1":
			assert.Equal(t, StatusCloning, repo.Status)
			assert.Equal(t, 0.3, repo.Progress)
		case "repo2":
			assert.Equal(t, StatusCompleted, repo.Status)
			assert.Equal(t, 1.0, repo.Progress)
		}
	}
}

func TestRenderCompactProgress(t *testing.T) {
	repos := []string{"repo1", "repo2", "repo3", "repo4"}
	tracker := NewProgressTracker(repos, DisplayModeCompact)

	// Complete some repositories
	tracker.CompleteRepository("repo1", "Done")
	tracker.CompleteRepository("repo2", "Done")
	tracker.SetRepositoryError("repo3", "Failed")
	// repo4 remains pending

	progress := tracker.RenderProgress()

	// Should contain progress percentage
	assert.Contains(t, progress, "50.0%")
	// Should contain counts
	assert.Contains(t, progress, "3/4") // 2 completed + 1 failed out of 4
	// Should contain success/failure counts
	assert.Contains(t, progress, "✓ 2")
	assert.Contains(t, progress, "✗ 1")
	assert.Contains(t, progress, "⏳ 1")
	// Should contain progress bar
	assert.Contains(t, progress, "█")
	assert.Contains(t, progress, "░")
}

func TestRenderDetailedProgress(t *testing.T) {
	repos := []string{"repo1", "repo2", "repo3"}
	tracker := NewProgressTracker(repos, DisplayModeDetailed)

	// Set up different statuses
	tracker.UpdateRepository("repo1", StatusCloning, "Cloning", 0.7)
	tracker.CompleteRepository("repo2", "Done")
	tracker.SetRepositoryError("repo3", "Network error")

	progress := tracker.RenderProgress()

	// Should contain overall progress
	assert.Contains(t, progress, "Overall Progress:")
	assert.Contains(t, progress, "56.7%") // (0.7 + 1.0 + 0.0) / 3 * 100

	// Should contain status sections
	assert.Contains(t, progress, "📥 Cloning")
	assert.Contains(t, progress, "✅ Completed")
	assert.Contains(t, progress, "❌ Failed")

	// Should contain repository names
	assert.Contains(t, progress, "repo1")
	assert.Contains(t, progress, "repo2")
	assert.Contains(t, progress, "repo3")
}

func TestRenderQuietProgress(t *testing.T) {
	repos := []string{"repo1", "repo2"}
	tracker := NewProgressTracker(repos, DisplayModeQuiet)

	progress := tracker.RenderProgress()
	assert.Empty(t, progress)
}

func TestGetStatusEmoji(t *testing.T) {
	tests := []struct {
		status   ProgressStatus
		expected string
	}{
		{StatusPending, "⏳ Pending"},
		{StatusStarted, "🚀 Started"},
		{StatusCloning, "📥 Cloning"},
		{StatusPulling, "🔄 Pulling"},
		{StatusFetching, "📡 Fetching"},
		{StatusResetting, "🔄 Resetting"},
		{StatusCompleted, "✅ Completed"},
		{StatusFailed, "❌ Failed"},
		{StatusSkipped, "⏭️ Skipped"},
	}

	for _, test := range tests {
		result := getStatusEmoji(test.status)
		assert.Equal(t, test.expected, result)
	}
}

func TestGetDuration(t *testing.T) {
	repos := []string{"repo1"}
	tracker := NewProgressTracker(repos, DisplayModeCompact)

	// Small delay to ensure duration is non-zero
	time.Sleep(10 * time.Millisecond)

	duration := tracker.GetDuration()
	assert.True(t, duration > 0)
	assert.True(t, duration < time.Second) // Should be much less than a second
}

func TestGetETA(t *testing.T) {
	repos := []string{"repo1", "repo2", "repo3", "repo4"}
	tracker := NewProgressTracker(repos, DisplayModeCompact)

	// With 0% progress, ETA should be 0
	eta := tracker.GetETA()
	assert.Equal(t, time.Duration(0), eta)

	// Wait a bit then complete 50% of repositories
	time.Sleep(10 * time.Millisecond)
	tracker.CompleteRepository("repo1", "Done")
	tracker.CompleteRepository("repo2", "Done")

	eta = tracker.GetETA()
	assert.True(t, eta > 0)
	// ETA should be roughly equal to elapsed time (since we're 50% done)
	elapsed := tracker.GetDuration()
	assert.True(t, eta > elapsed/3) // Should be at least 1/3 of elapsed time
	assert.True(t, eta < elapsed*3) // Should be less than 3x elapsed time
}

func TestIsCompleted(t *testing.T) {
	repos := []string{"repo1", "repo2"}
	tracker := NewProgressTracker(repos, DisplayModeCompact)

	// Initially not completed
	assert.False(t, tracker.IsCompleted())

	// Complete one repository
	tracker.CompleteRepository("repo1", "Done")
	assert.False(t, tracker.IsCompleted())

	// Complete all repositories
	tracker.CompleteRepository("repo2", "Done")
	assert.True(t, tracker.IsCompleted())

	// Test with failed repositories
	tracker.SetRepositoryError("repo1", "Error")
	assert.True(t, tracker.IsCompleted()) // Failed repos also count as "completed"
}

func TestGetSummary(t *testing.T) {
	repos := []string{"repo1", "repo2", "repo3"}
	tracker := NewProgressTracker(repos, DisplayModeCompact)

	// Complete some repositories
	tracker.CompleteRepository("repo1", "Done")
	tracker.SetRepositoryError("repo2", "Failed")
	// repo3 remains pending

	summary := tracker.GetSummary()

	// Should contain progress percentage
	assert.Contains(t, summary, "33.3%")
	// Should contain counts
	assert.Contains(t, summary, "1 successful")
	assert.Contains(t, summary, "1 failed")
	assert.Contains(t, summary, "1 pending")
	// Should contain duration
	assert.Contains(t, summary, "Duration:")
}

func TestConcurrentAccess(t *testing.T) {
	repos := []string{"repo1", "repo2", "repo3"}
	tracker := NewProgressTracker(repos, DisplayModeCompact)

	// Test concurrent access to avoid race conditions
	go func() {
		for i := 0; i < 100; i++ {
			tracker.UpdateRepository("repo1", StatusCloning, "Cloning", 0.5)
		}
	}()

	go func() {
		for i := 0; i < 100; i++ {
			tracker.GetOverallProgress()
		}
	}()

	go func() {
		for i := 0; i < 100; i++ {
			tracker.RenderProgress()
		}
	}()

	// Wait for goroutines to complete
	time.Sleep(100 * time.Millisecond)

	// Should not crash and should be in a consistent state
	completed, failed, pending, progressPercent := tracker.GetOverallProgress()
	assert.True(t, completed >= 0)
	assert.True(t, failed >= 0)
	assert.True(t, pending >= 0)
	assert.True(t, progressPercent >= 0)
	assert.True(t, progressPercent <= 100)
}

func TestProgressBarRendering(t *testing.T) {
	repos := []string{"repo1", "repo2", "repo3", "repo4"}
	tracker := NewProgressTracker(repos, DisplayModeCompact)

	// Complete 50% of repositories
	tracker.CompleteRepository("repo1", "Done")
	tracker.CompleteRepository("repo2", "Done")

	progress := tracker.RenderProgress()

	// Should contain filled and empty characters
	assert.Contains(t, progress, "█")
	assert.Contains(t, progress, "░")

	// Count the characters in the progress bar
	barStart := strings.Index(progress, "[")
	barEnd := strings.Index(progress, "]")

	assert.True(t, barStart >= 0)
	assert.True(t, barEnd > barStart)

	progressBar := progress[barStart+1 : barEnd]
	filledCount := strings.Count(progressBar, "█")
	emptyCount := strings.Count(progressBar, "░")

	// Should have roughly 50% filled (allowing for rounding)
	totalWidth := filledCount + emptyCount
	assert.True(t, totalWidth > 0)

	filledPercent := float64(filledCount) / float64(totalWidth) * 100
	assert.True(t, filledPercent >= 40 && filledPercent <= 60) // 50% ± 10%
}
