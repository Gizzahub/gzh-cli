package bulkclone

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CloneState represents the state of a bulk clone operation
type CloneState struct {
	// Operation metadata
	StartTime    time.Time `json:"start_time"`
	LastUpdated  time.Time `json:"last_updated"`
	Provider     string    `json:"provider"`
	Organization string    `json:"organization"`
	TargetPath   string    `json:"target_path"`
	Strategy     string    `json:"strategy"`

	// Progress tracking
	TotalRepositories int                   `json:"total_repositories"`
	CompletedRepos    []CompletedRepository `json:"completed_repos"`
	FailedRepos       []FailedRepository    `json:"failed_repos"`
	PendingRepos      []string              `json:"pending_repos"`

	// Configuration
	Parallel   int `json:"parallel"`
	MaxRetries int `json:"max_retries"`

	// Status
	Status string `json:"status"` // "in_progress", "completed", "failed", "cancelled"
}

// CompletedRepository represents a successfully processed repository
type CompletedRepository struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Operation   string    `json:"operation"` // "clone", "pull", "reset", "fetch"
	CompletedAt time.Time `json:"completed_at"`
	Message     string    `json:"message,omitempty"`
}

// FailedRepository represents a failed repository operation
type FailedRepository struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Operation   string    `json:"operation"`
	Error       string    `json:"error"`
	Attempts    int       `json:"attempts"`
	LastAttempt time.Time `json:"last_attempt"`
}

// StateManager handles saving and loading clone states
type StateManager struct {
	stateDir string
}

// NewStateManager creates a new state manager
func NewStateManager(stateDir string) *StateManager {
	if stateDir == "" {
		// Default to ~/.gzh/state
		homeDir, _ := os.UserHomeDir()
		stateDir = filepath.Join(homeDir, ".gzh", "state")
	}

	return &StateManager{
		stateDir: stateDir,
	}
}

// GetStateFilePath returns the path to the state file for a given operation
func (sm *StateManager) GetStateFilePath(provider, organization string) string {
	filename := fmt.Sprintf("%s_%s.json", provider, organization)
	return filepath.Join(sm.stateDir, filename)
}

// SaveState saves the clone state to disk
func (sm *StateManager) SaveState(state *CloneState) error {
	// Ensure state directory exists
	if err := os.MkdirAll(sm.stateDir, 0o755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Update last updated time
	state.LastUpdated = time.Now()

	// Get state file path
	statePath := sm.GetStateFilePath(state.Provider, state.Organization)

	// Marshal state to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to file
	if err := os.WriteFile(statePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// LoadState loads the clone state from disk
func (sm *StateManager) LoadState(provider, organization string) (*CloneState, error) {
	statePath := sm.GetStateFilePath(provider, organization)

	// Check if state file exists
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no state file found for %s/%s", provider, organization)
	}

	// Read state file
	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	// Unmarshal state
	var state CloneState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

// DeleteState removes the state file
func (sm *StateManager) DeleteState(provider, organization string) error {
	statePath := sm.GetStateFilePath(provider, organization)

	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return nil // Already deleted
	}

	if err := os.Remove(statePath); err != nil {
		return fmt.Errorf("failed to delete state file: %w", err)
	}

	return nil
}

// HasState checks if a state file exists for the given operation
func (sm *StateManager) HasState(provider, organization string) bool {
	statePath := sm.GetStateFilePath(provider, organization)
	_, err := os.Stat(statePath)
	return err == nil
}

// ListStates returns all saved states
func (sm *StateManager) ListStates() ([]CloneState, error) {
	// Check if state directory exists
	if _, err := os.Stat(sm.stateDir); os.IsNotExist(err) {
		return []CloneState{}, nil
	}

	// Read directory
	entries, err := os.ReadDir(sm.stateDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read state directory: %w", err)
	}

	var states []CloneState
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			// Read and parse each state file
			data, err := os.ReadFile(filepath.Join(sm.stateDir, entry.Name()))
			if err != nil {
				continue // Skip invalid files
			}

			var state CloneState
			if err := json.Unmarshal(data, &state); err != nil {
				continue // Skip invalid JSON
			}

			states = append(states, state)
		}
	}

	return states, nil
}

// NewCloneState creates a new clone state
func NewCloneState(provider, organization, targetPath, strategy string, parallel, maxRetries int) *CloneState {
	return &CloneState{
		StartTime:         time.Now(),
		LastUpdated:       time.Now(),
		Provider:          provider,
		Organization:      organization,
		TargetPath:        targetPath,
		Strategy:          strategy,
		TotalRepositories: 0,
		CompletedRepos:    []CompletedRepository{},
		FailedRepos:       []FailedRepository{},
		PendingRepos:      []string{},
		Parallel:          parallel,
		MaxRetries:        maxRetries,
		Status:            "in_progress",
	}
}

// AddCompletedRepository adds a completed repository to the state
func (cs *CloneState) AddCompletedRepository(name, path, operation, message string) {
	cs.CompletedRepos = append(cs.CompletedRepos, CompletedRepository{
		Name:        name,
		Path:        path,
		Operation:   operation,
		CompletedAt: time.Now(),
		Message:     message,
	})

	// Remove from pending if it exists
	cs.removePendingRepo(name)

	// Update total repositories count
	cs.updateTotalRepositories()
}

// AddFailedRepository adds a failed repository to the state
func (cs *CloneState) AddFailedRepository(name, path, operation, errorMsg string, attempts int) {
	// Check if this repo already failed and update it
	for i, failed := range cs.FailedRepos {
		if failed.Name == name {
			cs.FailedRepos[i].Error = errorMsg
			cs.FailedRepos[i].Attempts = attempts
			cs.FailedRepos[i].LastAttempt = time.Now()
			return
		}
	}

	// Add new failed repo
	cs.FailedRepos = append(cs.FailedRepos, FailedRepository{
		Name:        name,
		Path:        path,
		Operation:   operation,
		Error:       errorMsg,
		Attempts:    attempts,
		LastAttempt: time.Now(),
	})

	// Remove from pending if it exists
	cs.removePendingRepo(name)

	// Update total repositories count
	cs.updateTotalRepositories()
}

// IsCompleted checks if a repository has been completed
func (cs *CloneState) IsCompleted(name string) bool {
	for _, completed := range cs.CompletedRepos {
		if completed.Name == name {
			return true
		}
	}
	return false
}

// IsFailed checks if a repository has failed
func (cs *CloneState) IsFailed(name string) bool {
	for _, failed := range cs.FailedRepos {
		if failed.Name == name {
			return true
		}
	}
	return false
}

// GetProgress returns the current progress statistics
func (cs *CloneState) GetProgress() (completed, failed, pending int) {
	return len(cs.CompletedRepos), len(cs.FailedRepos), len(cs.PendingRepos)
}

// GetProgressPercent returns the progress as a percentage
func (cs *CloneState) GetProgressPercent() float64 {
	if cs.TotalRepositories == 0 {
		return 0
	}

	completed, failed, _ := cs.GetProgress()
	processed := completed + failed
	return float64(processed) / float64(cs.TotalRepositories) * 100
}

// SetPendingRepositories sets the list of pending repositories
func (cs *CloneState) SetPendingRepositories(repos []string) {
	cs.PendingRepos = repos
	cs.updateTotalRepositories()
}

// updateTotalRepositories updates the total repository count
func (cs *CloneState) updateTotalRepositories() {
	cs.TotalRepositories = len(cs.PendingRepos) + len(cs.CompletedRepos) + len(cs.FailedRepos)
}

// removePendingRepo removes a repository from the pending list
func (cs *CloneState) removePendingRepo(name string) {
	for i, repo := range cs.PendingRepos {
		if repo == name {
			cs.PendingRepos = append(cs.PendingRepos[:i], cs.PendingRepos[i+1:]...)
			break
		}
	}
}

// GetRemainingRepositories returns repositories that still need to be processed
func (cs *CloneState) GetRemainingRepositories() []string {
	return cs.PendingRepos
}

// MarkCompleted marks the operation as completed
func (cs *CloneState) MarkCompleted() {
	cs.Status = "completed"
	cs.LastUpdated = time.Now()
}

// MarkFailed marks the operation as failed
func (cs *CloneState) MarkFailed() {
	cs.Status = "failed"
	cs.LastUpdated = time.Now()
}

// MarkCancelled marks the operation as cancelled
func (cs *CloneState) MarkCancelled() {
	cs.Status = "cancelled"
	cs.LastUpdated = time.Now()
}
