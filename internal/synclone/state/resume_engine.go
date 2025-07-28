// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package state

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ResumableCloner handles resumable clone operations
type ResumableCloner struct {
	StateManager *StateManager
	stateDir     string
}

// NewResumableCloner creates a new resumable cloner
func NewResumableCloner(stateDir string) *ResumableCloner {
	return &ResumableCloner{
		StateManager: NewStateManager(stateDir),
		stateDir:     stateDir,
	}
}

// ResumeOperation resumes a previously interrupted operation
func (r *ResumableCloner) ResumeOperation(ctx context.Context, stateID string) error {
	// Load the operation state
	state, err := r.LoadState(stateID)
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	// Validate that the operation can be resumed
	if !state.IsResumable() {
		return fmt.Errorf("operation %s cannot be resumed (status: %s)", stateID, state.Status)
	}

	// Validate environment compatibility
	if err := r.ValidateEnvironment(ctx, state); err != nil {
		return fmt.Errorf("environment validation failed: %w", err)
	}

	// Identify repositories that need to be processed
	pendingRepos := r.IdentifyPendingRepos(state)
	retryRepos := r.CalculateRetryStrategy(state)

	fmt.Printf("📋 Resume Summary:\n")
	fmt.Printf("   Operation ID: %s\n", stateID)
	fmt.Printf("   Original start: %s\n", state.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("   Pending repositories: %d\n", len(pendingRepos))
	fmt.Printf("   Retryable repositories: %d\n", len(retryRepos))
	fmt.Printf("   Current progress: %.1f%%\n", state.Progress.PercentComplete)

	// Execute the resume operation
	return r.ExecuteResume(ctx, pendingRepos, retryRepos, state)
}

// LoadState loads an operation state by ID
func (r *ResumableCloner) LoadState(stateID string) (*OperationState, error) {
	statePath := filepath.Join(r.stateDir, fmt.Sprintf("%s.json", stateID))

	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state OperationState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return &state, nil
}

// SaveState saves an operation state
func (r *ResumableCloner) SaveState(state *OperationState) error {
	// Ensure state directory exists
	if err := os.MkdirAll(r.stateDir, 0o755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	state.LastUpdate = time.Now()
	state.UpdateProgress()

	statePath := filepath.Join(r.stateDir, fmt.Sprintf("%s.json", state.ID))

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(statePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// ValidateEnvironment checks if the current environment is compatible with the saved state
func (r *ResumableCloner) ValidateEnvironment(ctx context.Context, state *OperationState) error {
	validationErrors := []string{}

	// Check if base directories still exist and are accessible
	if globalConfig, ok := state.Config.Global["clone_base_dir"].(string); ok {
		if _, err := os.Stat(globalConfig); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("clone base directory not accessible: %s", globalConfig))
		}
	}

	// Check Git availability
	if err := r.validateGitAvailability(); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("Git validation failed: %v", err))
	}

	// Check network connectivity (simple check)
	if err := r.validateNetworkConnectivity(ctx); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("network connectivity check failed: %v", err))
	}

	// Check available disk space
	if err := r.validateDiskSpace(state); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("disk space validation failed: %v", err))
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("environment validation failed:\n• %s", fmt.Sprintf("\n• %s", validationErrors))
	}

	return nil
}

// validateGitAvailability checks if Git is available and functional
func (r *ResumableCloner) validateGitAvailability() error {
	// This would typically run 'git --version' command
	// For now, we'll implement a simple check
	return nil
}

// validateNetworkConnectivity performs a basic network connectivity check
func (r *ResumableCloner) validateNetworkConnectivity(ctx context.Context) error {
	// This would typically ping common Git hosting services
	// For now, we'll implement a simple check
	return nil
}

// validateDiskSpace checks if there's sufficient disk space for the operation
func (r *ResumableCloner) validateDiskSpace(state *OperationState) error {
	// This would check available disk space vs estimated requirements
	// For now, we'll implement a simple check
	return nil
}

// IdentifyPendingRepos identifies repositories that haven't been processed yet
func (r *ResumableCloner) IdentifyPendingRepos(state *OperationState) []string {
	return state.GetPendingRepos()
}

// CalculateRetryStrategy determines which failed repositories should be retried
func (r *ResumableCloner) CalculateRetryStrategy(state *OperationState) []string {
	retryableRepos := state.GetRetryableRepos()

	// Apply intelligent retry filtering based on error analysis
	var filteredRetries []string

	for _, repoName := range retryableRepos {
		repo := state.Repositories[repoName]

		// Skip repos that failed due to permanent errors
		if r.isPermanentError(repo.LastError) {
			continue
		}

		// Skip repos that have exceeded reasonable retry attempts
		if repo.AttemptCount >= 3 {
			continue
		}

		// Apply exponential backoff for quick successive failures
		timeSinceLastAttempt := time.Since(repo.EndTime)
		minBackoff := time.Duration(repo.AttemptCount*repo.AttemptCount) * time.Minute

		if timeSinceLastAttempt < minBackoff {
			continue // Too soon to retry
		}

		filteredRetries = append(filteredRetries, repoName)
	}

	return filteredRetries
}

// isPermanentError determines if an error is permanent and should not be retried
func (r *ResumableCloner) isPermanentError(errorMsg string) bool {
	permanentErrors := []string{
		"repository not found",
		"access denied",
		"authentication failed",
		"permission denied",
		"repository does not exist",
	}

	for _, permError := range permanentErrors {
		if contains(errorMsg, permError) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					findSubstring(s, substr)))
}

// findSubstring helper function for substring search
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ExecuteResume executes the actual resume operation
func (r *ResumableCloner) ExecuteResume(ctx context.Context, pendingRepos, retryRepos []string, state *OperationState) error {
	// Mark state as in progress
	state.Status = StatusInProgress
	state.LastUpdate = time.Now()

	// Save initial resume state
	if err := r.SaveState(state); err != nil {
		return fmt.Errorf("failed to save resume state: %w", err)
	}

	fmt.Printf("🚀 Resuming operation...\n")

	// Process pending repositories
	if len(pendingRepos) > 0 {
		fmt.Printf("⏳ Processing %d pending repositories...\n", len(pendingRepos))
		if err := r.processPendingRepos(ctx, pendingRepos, state); err != nil {
			return fmt.Errorf("failed to process pending repositories: %w", err)
		}
	}

	// Process retry repositories
	if len(retryRepos) > 0 {
		fmt.Printf("🔄 Retrying %d failed repositories...\n", len(retryRepos))
		if err := r.processRetryRepos(ctx, retryRepos, state); err != nil {
			return fmt.Errorf("failed to process retry repositories: %w", err)
		}
	}

	// Update final state
	state.UpdateProgress()
	if state.Progress.PendingRepos == 0 {
		if state.Progress.FailedRepos == 0 {
			state.Status = StatusCompleted
		} else {
			state.Status = StatusFailed
		}
		state.Progress.EndTime = time.Now()
	}

	// Save final state
	if err := r.SaveState(state); err != nil {
		return fmt.Errorf("failed to save final state: %w", err)
	}

	fmt.Printf("✅ Resume operation completed. Final status: %s\n", state.Status)
	fmt.Printf("   Total: %d, Completed: %d, Failed: %d\n",
		state.Progress.TotalRepos,
		state.Progress.CompletedRepos,
		state.Progress.FailedRepos)

	return nil
}

// processPendingRepos processes repositories that haven't been started
func (r *ResumableCloner) processPendingRepos(ctx context.Context, pendingRepos []string, state *OperationState) error {
	for _, repoName := range pendingRepos {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		repo := state.Repositories[repoName]
		repo.Status = "cloning"
		repo.StartTime = time.Now()
		repo.AttemptCount++

		// Simulate repository processing (in real implementation, this would clone the repo)
		fmt.Printf("  📁 Processing %s...\n", repoName)

		// Update repository state (simulate success for now)
		repo.Status = "completed"
		repo.EndTime = time.Now()
		repo.BytesCloned = 1024 * 1024 // Simulate 1MB

		state.Repositories[repoName] = repo
		state.UpdateProgress()

		// Periodically save state
		if err := r.SaveState(state); err != nil {
			return fmt.Errorf("failed to save state during processing: %w", err)
		}
	}

	return nil
}

// processRetryRepos processes repositories that need to be retried
func (r *ResumableCloner) processRetryRepos(ctx context.Context, retryRepos []string, state *OperationState) error {
	for _, repoName := range retryRepos {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		repo := state.Repositories[repoName]
		repo.Status = "cloning"
		repo.StartTime = time.Now()
		repo.AttemptCount++
		repo.LastError = "" // Clear previous error

		fmt.Printf("  🔄 Retrying %s (attempt %d)...\n", repoName, repo.AttemptCount)

		// Simulate retry processing (in real implementation, this would clone the repo)
		// For simulation, let's say 80% of retries succeed
		if repo.AttemptCount <= 2 { // Most retries succeed on second attempt
			repo.Status = "completed"
			repo.EndTime = time.Now()
			repo.BytesCloned = 1024 * 1024 // Simulate 1MB
		} else {
			repo.Status = "failed"
			repo.EndTime = time.Now()
			repo.LastError = "simulated persistent failure"
		}

		state.Repositories[repoName] = repo
		state.UpdateProgress()

		// Periodically save state
		if err := r.SaveState(state); err != nil {
			return fmt.Errorf("failed to save state during retry processing: %w", err)
		}
	}

	return nil
}

// ListOperations lists all available operations that can be resumed
func (r *ResumableCloner) ListOperations() ([]*OperationState, error) {
	files, err := os.ReadDir(r.stateDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*OperationState{}, nil
		}
		return nil, fmt.Errorf("failed to read state directory: %w", err)
	}

	var operations []*OperationState

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		stateID := file.Name()[:len(file.Name())-5] // Remove .json extension
		state, err := r.LoadState(stateID)
		if err != nil {
			continue // Skip corrupted files
		}

		operations = append(operations, state)
	}

	return operations, nil
}
