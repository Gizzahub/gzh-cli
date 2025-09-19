// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package update

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// ProgressTracker provides detailed progress tracking with time estimates
// and step-by-step progress reporting for PM update operations.
type ProgressTracker struct {
	mu             sync.Mutex
	totalSteps     int
	currentStep    int
	currentAction  string
	startTime      time.Time
	stepStartTime  time.Time
	stepDurations  []time.Duration
	estimatedEnd   time.Time
	managers       []ManagerProgress
	currentManager int
	showProgress   bool
	formatter      *OutputFormatter
}

// ManagerProgress tracks progress for individual package managers
type ManagerProgress struct {
	Name            string         `json:"name"`
	Status          string         `json:"status"` // "pending", "active", "completed", "failed", "skipped"
	StartTime       time.Time      `json:"startTime"`
	EndTime         time.Time      `json:"endTime"`
	Duration        time.Duration  `json:"duration"`
	Steps           []StepProgress `json:"steps"`
	CurrentStep     int            `json:"currentStep"`
	PackagesUpdated int            `json:"packagesUpdated"`
	Error           string         `json:"error,omitempty"`
}

// StepProgress tracks individual step progress within a manager
type StepProgress struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Status      string        `json:"status"` // "pending", "active", "completed", "failed"
	StartTime   time.Time     `json:"startTime"`
	EndTime     time.Time     `json:"endTime"`
	Duration    time.Duration `json:"duration"`
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(managerNames []string, formatter *OutputFormatter) *ProgressTracker {
	managers := make([]ManagerProgress, len(managerNames))
	for i, name := range managerNames {
		managers[i] = ManagerProgress{
			Name:   name,
			Status: "pending",
			Steps:  getDefaultStepsForManager(name),
		}
	}

	return &ProgressTracker{
		totalSteps:   len(managerNames),
		startTime:    time.Now(),
		managers:     managers,
		showProgress: shouldShowProgress(),
		formatter:    formatter,
	}
}

// shouldShowProgress determines if progress indication should be shown
func shouldShowProgress() bool {
	// Don't show progress in CI environments or when output is redirected
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		return false
	}
	return true
}

// getDefaultStepsForManager returns default steps for each package manager
func getDefaultStepsForManager(manager string) []StepProgress {
	switch manager {
	case "brew":
		return []StepProgress{
			{Name: "update", Description: "Updating Homebrew formulae", Status: "pending"},
			{Name: "upgrade", Description: "Upgrading packages", Status: "pending"},
			{Name: "cleanup", Description: "Cleaning up old versions", Status: "pending"},
		}
	case "asdf":
		return []StepProgress{
			{Name: "plugin_update", Description: "Updating asdf plugins", Status: "pending"},
			{Name: "version_check", Description: "Checking for version updates", Status: "pending"},
			{Name: "install", Description: "Installing updated versions", Status: "pending"},
			{Name: "post_actions", Description: "Running post-install actions", Status: "pending"},
		}
	case "sdkman":
		return []StepProgress{
			{Name: "selfupdate", Description: "Updating SDKMAN itself", Status: "pending"},
			{Name: "update", Description: "Refreshing candidate metadata", Status: "pending"},
		}
	case "npm":
		return []StepProgress{
			{Name: "check", Description: "Checking global packages", Status: "pending"},
			{Name: "update", Description: "Updating global packages", Status: "pending"},
		}
	case "pip":
		return []StepProgress{
			{Name: "upgrade_pip", Description: "Upgrading pip itself", Status: "pending"},
			{Name: "check_outdated", Description: "Finding outdated packages", Status: "pending"},
			{Name: "upgrade_packages", Description: "Upgrading packages", Status: "pending"},
		}
	case "apt":
		return []StepProgress{
			{Name: "update", Description: "Updating package lists", Status: "pending"},
			{Name: "upgrade", Description: "Upgrading packages", Status: "pending"},
		}
	case "pacman":
		return []StepProgress{
			{Name: "sync_update", Description: "Syncing and updating system", Status: "pending"},
			{Name: "cleanup", Description: "Cleaning orphaned packages", Status: "pending"},
		}
	case "yay":
		return []StepProgress{
			{Name: "update", Description: "Updating AUR packages", Status: "pending"},
			{Name: "cleanup", Description: "Cleaning package cache", Status: "pending"},
		}
	default:
		return []StepProgress{
			{Name: "update", Description: "Updating packages", Status: "pending"},
		}
	}
}

// StartManager marks a manager as actively being processed
func (pt *ProgressTracker) StartManager(managerName string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	for i := range pt.managers {
		if pt.managers[i].Name == managerName {
			pt.managers[i].Status = "active"
			pt.managers[i].StartTime = time.Now()
			pt.currentManager = i
			pt.currentStep = i + 1
			pt.stepStartTime = time.Now()

			if pt.showProgress {
				pt.formatter.PrintManagerUpdate(managerName, pt.currentStep, pt.totalSteps, "updating")
				pt.printProgressOverview()
			}
			break
		}
	}
}

// StartManagerStep starts a specific step within a manager
func (pt *ProgressTracker) StartManagerStep(managerName, stepName string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	for i := range pt.managers {
		if pt.managers[i].Name == managerName {
			for j := range pt.managers[i].Steps {
				if pt.managers[i].Steps[j].Name == stepName {
					pt.managers[i].Steps[j].Status = "active"
					pt.managers[i].Steps[j].StartTime = time.Now()
					pt.managers[i].CurrentStep = j

					if pt.showProgress {
						pt.printStepProgress(managerName, pt.managers[i].Steps[j])
					}
					break
				}
			}
			break
		}
	}
}

// CompleteManagerStep marks a step as completed
func (pt *ProgressTracker) CompleteManagerStep(managerName, stepName string, packagesAffected int) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	for i := range pt.managers {
		if pt.managers[i].Name == managerName {
			for j := range pt.managers[i].Steps {
				if pt.managers[i].Steps[j].Name == stepName {
					step := &pt.managers[i].Steps[j]
					step.Status = "completed"
					step.EndTime = time.Now()
					step.Duration = step.EndTime.Sub(step.StartTime)

					pt.managers[i].PackagesUpdated += packagesAffected

					if pt.showProgress {
						pt.printStepCompletion(managerName, *step, packagesAffected)
					}
					break
				}
			}
			break
		}
	}
}

// FailManagerStep marks a step as failed
func (pt *ProgressTracker) FailManagerStep(managerName, stepName, errorMsg string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	for i := range pt.managers {
		if pt.managers[i].Name == managerName {
			for j := range pt.managers[i].Steps {
				if pt.managers[i].Steps[j].Name == stepName {
					step := &pt.managers[i].Steps[j]
					step.Status = "failed"
					step.EndTime = time.Now()
					step.Duration = step.EndTime.Sub(step.StartTime)

					pt.managers[i].Error = errorMsg

					if pt.showProgress {
						pt.printStepFailure(managerName, *step, errorMsg)
					}
					break
				}
			}
			break
		}
	}
}

// CompleteManager marks a manager as completed
func (pt *ProgressTracker) CompleteManager(managerName string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	for i := range pt.managers {
		if pt.managers[i].Name == managerName {
			pt.managers[i].Status = "completed"
			pt.managers[i].EndTime = time.Now()
			pt.managers[i].Duration = pt.managers[i].EndTime.Sub(pt.managers[i].StartTime)

			// Record step duration for ETA calculation
			stepDuration := time.Since(pt.stepStartTime)
			pt.stepDurations = append(pt.stepDurations, stepDuration)
			pt.updateETA()

			if pt.showProgress {
				pt.printManagerCompletion(pt.managers[i])
			}
			break
		}
	}
}

// SkipManager marks a manager as skipped
func (pt *ProgressTracker) SkipManager(managerName, reason string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	for i := range pt.managers {
		if pt.managers[i].Name == managerName {
			pt.managers[i].Status = "skipped"
			pt.managers[i].EndTime = time.Now()
			pt.managers[i].Error = reason
			pt.currentStep = i + 1

			if pt.showProgress {
				pt.formatter.PrintManagerUpdate(managerName, pt.currentStep, pt.totalSteps, "skip")
				fmt.Printf("Reason: %s\n", reason)
			}
			break
		}
	}
}

// updateETA calculates estimated time to completion
func (pt *ProgressTracker) updateETA() {
	if len(pt.stepDurations) == 0 {
		return
	}

	// Calculate average step duration
	var total time.Duration
	for _, duration := range pt.stepDurations {
		total += duration
	}
	avgDuration := total / time.Duration(len(pt.stepDurations))

	// Estimate remaining time
	remainingSteps := pt.totalSteps - pt.currentStep
	if remainingSteps > 0 {
		remainingTime := time.Duration(remainingSteps) * avgDuration
		pt.estimatedEnd = time.Now().Add(remainingTime)
	}
}

// printProgressOverview prints a general progress overview
func (pt *ProgressTracker) printProgressOverview() {
	if !pt.showProgress {
		return
	}

	completed := 0
	skipped := 0
	failed := 0

	for _, manager := range pt.managers {
		switch manager.Status {
		case "completed":
			completed++
		case "skipped":
			skipped++
		case "failed":
			failed++
		}
	}

	fmt.Printf("Progress: %d/%d managers completed", completed, pt.totalSteps)
	if skipped > 0 {
		fmt.Printf(", %d skipped", skipped)
	}
	if failed > 0 {
		fmt.Printf(", %d failed", failed)
	}

	if !pt.estimatedEnd.IsZero() && pt.estimatedEnd.After(time.Now()) {
		remaining := pt.estimatedEnd.Sub(time.Now())
		fmt.Printf(" (ETA: %s)", formatDuration(remaining))
	}
	fmt.Println()
}

// printStepProgress prints progress for a specific step
func (pt *ProgressTracker) printStepProgress(managerName string, step StepProgress) {
	var emoji string
	if pt.formatter.showEmojis {
		emoji = "🔄"
	}
	fmt.Printf("%s %s: %s...\n", emoji, managerName, step.Description)
}

// printStepCompletion prints completion of a specific step
func (pt *ProgressTracker) printStepCompletion(managerName string, step StepProgress, packagesAffected int) {
	duration := step.Duration.Truncate(time.Millisecond)

	var details string
	if packagesAffected > 0 {
		if packagesAffected == 1 {
			details = fmt.Sprintf("1 package affected")
		} else {
			details = fmt.Sprintf("%d packages affected", packagesAffected)
		}
	} else {
		details = fmt.Sprintf("completed in %s", formatDuration(duration))
	}

	pt.formatter.PrintCommandResult(step.Description, true, details)
}

// printStepFailure prints failure of a specific step
func (pt *ProgressTracker) printStepFailure(managerName string, step StepProgress, errorMsg string) {
	pt.formatter.PrintCommandResult(step.Description, false, errorMsg)
}

// printManagerCompletion prints completion summary for a manager
func (pt *ProgressTracker) printManagerCompletion(manager ManagerProgress) {
	duration := manager.Duration.Truncate(time.Millisecond)

	var summary string
	if manager.PackagesUpdated > 0 {
		if manager.PackagesUpdated == 1 {
			summary = fmt.Sprintf("1 package updated in %s", formatDuration(duration))
		} else {
			summary = fmt.Sprintf("%d packages updated in %s", manager.PackagesUpdated, formatDuration(duration))
		}
	} else {
		summary = fmt.Sprintf("completed in %s (no updates needed)", formatDuration(duration))
	}

	var emoji string
	if pt.formatter.showEmojis {
		emoji = "✅"
	}
	fmt.Printf("%s %s: %s\n", emoji, manager.Name, summary)
}

// GetOverallProgress returns overall progress information
func (pt *ProgressTracker) GetOverallProgress() (completed, total int, eta time.Time) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	completed = 0
	for _, manager := range pt.managers {
		if manager.Status == "completed" || manager.Status == "skipped" {
			completed++
		}
	}

	return completed, pt.totalSteps, pt.estimatedEnd
}

// GetManagerProgress returns progress for a specific manager
func (pt *ProgressTracker) GetManagerProgress(managerName string) *ManagerProgress {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	for i := range pt.managers {
		if pt.managers[i].Name == managerName {
			return &pt.managers[i]
		}
	}
	return nil
}

// GetAllManagerProgress returns progress for all managers
func (pt *ProgressTracker) GetAllManagerProgress() []ManagerProgress {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	// Return a copy to avoid race conditions
	result := make([]ManagerProgress, len(pt.managers))
	copy(result, pt.managers)
	return result
}

// GetTotalDuration returns total elapsed time
func (pt *ProgressTracker) GetTotalDuration() time.Duration {
	return time.Since(pt.startTime)
}

// GetSummaryStats returns summary statistics for the update process
func (pt *ProgressTracker) GetSummaryStats() (successful, failed, skipped, totalPackages int) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	for _, manager := range pt.managers {
		switch manager.Status {
		case "completed":
			successful++
		case "failed":
			failed++
		case "skipped":
			skipped++
		}
		totalPackages += manager.PackagesUpdated
	}

	return successful, failed, skipped, totalPackages
}
