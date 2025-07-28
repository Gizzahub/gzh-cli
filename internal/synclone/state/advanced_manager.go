// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// StateManager handles advanced state management and cleanup
type StateManager struct {
	StateDir        string
	RetentionPolicy RetentionPolicy
}

// RetentionPolicy defines how long state files should be kept
type RetentionPolicy struct {
	MaxAge          time.Duration `json:"max_age"`
	MaxCompletedOps int           `json:"max_completed_ops"`
	MaxFailedOps    int           `json:"max_failed_ops"`
	AutoCleanup     bool          `json:"auto_cleanup"`
}

// StateFile represents a state file with metadata
type StateFile struct {
	ID         string          `json:"id"`
	FilePath   string          `json:"file_path"`
	State      *OperationState `json:"state"`
	FileInfo   os.FileInfo     `json:"-"`
	Size       int64           `json:"size"`
	CreatedAt  time.Time       `json:"created_at"`
	ModifiedAt time.Time       `json:"modified_at"`
}

// NewStateManager creates a new state manager
func NewStateManager(stateDir string) *StateManager {
	return &StateManager{
		StateDir: stateDir,
		RetentionPolicy: RetentionPolicy{
			MaxAge:          30 * 24 * time.Hour, // 30 days
			MaxCompletedOps: 50,
			MaxFailedOps:    20,
			AutoCleanup:     true,
		},
	}
}

// ListStateFiles returns all state files with metadata
func (sm *StateManager) ListStateFiles() ([]*StateFile, error) {
	files, err := os.ReadDir(sm.StateDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*StateFile{}, nil
		}
		return nil, fmt.Errorf("failed to read state directory: %w", err)
	}

	var stateFiles []*StateFile

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(sm.StateDir, file.Name())
		stateFile, err := sm.loadStateFile(filePath)
		if err != nil {
			// Log error but continue with other files
			continue
		}

		stateFiles = append(stateFiles, stateFile)
	}

	// Sort by modification time (newest first)
	sort.Slice(stateFiles, func(i, j int) bool {
		return stateFiles[i].ModifiedAt.After(stateFiles[j].ModifiedAt)
	})

	return stateFiles, nil
}

// loadStateFile loads a single state file with metadata
func (sm *StateManager) loadStateFile(filePath string) (*StateFile, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var state OperationState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	stateFile := &StateFile{
		ID:         state.ID,
		FilePath:   filePath,
		State:      &state,
		FileInfo:   fileInfo,
		Size:       fileInfo.Size(),
		CreatedAt:  state.StartTime,
		ModifiedAt: fileInfo.ModTime(),
	}

	return stateFile, nil
}

// RunCleanup performs state file cleanup based on retention policy
func (sm *StateManager) RunCleanup() error {
	stateFiles, err := sm.ListStateFiles()
	if err != nil {
		return fmt.Errorf("failed to list state files: %w", err)
	}

	var (
		completedFiles []*StateFile
		failedFiles    []*StateFile
		activeFiles    []*StateFile
	)

	// Categorize files by status
	for _, file := range stateFiles {
		switch file.State.Status {
		case StatusCompleted:
			completedFiles = append(completedFiles, file)
		case StatusFailed:
			failedFiles = append(failedFiles, file)
		default:
			activeFiles = append(activeFiles, file)
		}
	}

	var cleanedFiles []string

	// Clean up old files
	cutoffTime := time.Now().Add(-sm.RetentionPolicy.MaxAge)
	for _, file := range stateFiles {
		if file.ModifiedAt.Before(cutoffTime) && file.State.Status != StatusInProgress {
			if err := sm.deleteStateFile(file); err != nil {
				continue // Log error but continue
			}
			cleanedFiles = append(cleanedFiles, file.ID)
		}
	}

	// Clean up excess completed operations
	if len(completedFiles) > sm.RetentionPolicy.MaxCompletedOps {
		// Sort by modification time (oldest first)
		sort.Slice(completedFiles, func(i, j int) bool {
			return completedFiles[i].ModifiedAt.Before(completedFiles[j].ModifiedAt)
		})

		excessCount := len(completedFiles) - sm.RetentionPolicy.MaxCompletedOps
		for i := 0; i < excessCount; i++ {
			if err := sm.deleteStateFile(completedFiles[i]); err != nil {
				continue
			}
			cleanedFiles = append(cleanedFiles, completedFiles[i].ID)
		}
	}

	// Clean up excess failed operations
	if len(failedFiles) > sm.RetentionPolicy.MaxFailedOps {
		sort.Slice(failedFiles, func(i, j int) bool {
			return failedFiles[i].ModifiedAt.Before(failedFiles[j].ModifiedAt)
		})

		excessCount := len(failedFiles) - sm.RetentionPolicy.MaxFailedOps
		for i := 0; i < excessCount; i++ {
			if err := sm.deleteStateFile(failedFiles[i]); err != nil {
				continue
			}
			cleanedFiles = append(cleanedFiles, failedFiles[i].ID)
		}
	}

	// Optimize remaining state files
	if err := sm.optimizeStateFiles(stateFiles); err != nil {
		return fmt.Errorf("failed to optimize state files: %w", err)
	}

	fmt.Printf("Cleanup completed. Removed %d state files\n", len(cleanedFiles))
	return nil
}

// deleteStateFile safely deletes a state file
func (sm *StateManager) deleteStateFile(file *StateFile) error {
	return os.Remove(file.FilePath)
}

// optimizeStateFiles optimizes state files by compacting and deduplicating
func (sm *StateManager) optimizeStateFiles(files []*StateFile) error {
	for _, file := range files {
		// Skip active operations
		if file.State.Status == StatusInProgress {
			continue
		}

		// Compact state by removing unnecessary data
		optimizedState := sm.compactState(file.State)

		// Save optimized state back to file
		if err := sm.saveOptimizedState(file.FilePath, optimizedState); err != nil {
			continue // Log error but continue with other files
		}
	}

	return nil
}

// compactState removes unnecessary data from completed states
func (sm *StateManager) compactState(state *OperationState) *OperationState {
	compacted := *state

	// For completed operations, keep only essential information
	if state.Status == StatusCompleted || state.Status == StatusFailed {
		// Keep summary metrics but remove detailed progress
		compacted.Progress = OperationProgress{
			TotalRepos:     state.Progress.TotalRepos,
			CompletedRepos: state.Progress.CompletedRepos,
			FailedRepos:    state.Progress.FailedRepos,
			StartTime:      state.Progress.StartTime,
			EndTime:        state.Progress.EndTime,
		}

		// Keep only failed repository states for debugging
		if state.Status == StatusFailed {
			filteredRepos := make(map[string]RepoState)
			for name, repoState := range state.Repositories {
				if repoState.Status == "failed" {
					filteredRepos[name] = repoState
				}
			}
			compacted.Repositories = filteredRepos
		} else {
			// For completed operations, remove individual repo states
			compacted.Repositories = make(map[string]RepoState)
		}
	}

	return &compacted
}

// saveOptimizedState saves the optimized state to file
func (sm *StateManager) saveOptimizedState(filePath string, state *OperationState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal optimized state: %w", err)
	}

	return os.WriteFile(filePath, data, 0o644)
}

// AnalyzeStates provides analysis of all state files
func (sm *StateManager) AnalyzeStates() (*StateAnalysis, error) {
	stateFiles, err := sm.ListStateFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list state files: %w", err)
	}

	analysis := &StateAnalysis{
		TotalFiles:   len(stateFiles),
		TotalSize:    0,
		StatusCounts: make(map[OperationStatus]int),
		OldestFile:   time.Now(),
		NewestFile:   time.Time{},
	}

	for _, file := range stateFiles {
		analysis.TotalSize += file.Size
		analysis.StatusCounts[file.State.Status]++

		if file.CreatedAt.Before(analysis.OldestFile) {
			analysis.OldestFile = file.CreatedAt
		}
		if file.CreatedAt.After(analysis.NewestFile) {
			analysis.NewestFile = file.CreatedAt
		}
	}

	// Calculate recommendations
	analysis.Recommendations = sm.generateRecommendations(stateFiles)

	return analysis, nil
}

// StateAnalysis provides comprehensive analysis of state files
type StateAnalysis struct {
	TotalFiles      int                     `json:"total_files"`
	TotalSize       int64                   `json:"total_size_bytes"`
	StatusCounts    map[OperationStatus]int `json:"status_counts"`
	OldestFile      time.Time               `json:"oldest_file"`
	NewestFile      time.Time               `json:"newest_file"`
	Recommendations []string                `json:"recommendations"`
}

// generateRecommendations generates cleanup and optimization recommendations
func (sm *StateManager) generateRecommendations(files []*StateFile) []string {
	var recommendations []string

	// Check for excessive number of files
	if len(files) > 100 {
		recommendations = append(recommendations, "Consider running cleanup - high number of state files detected")
	}

	// Check for large total size
	totalSizeMB := float64(0)
	for _, file := range files {
		totalSizeMB += float64(file.Size) / (1024 * 1024)
	}
	if totalSizeMB > 100 {
		recommendations = append(recommendations, fmt.Sprintf("Consider cleanup - state files using %.1f MB disk space", totalSizeMB))
	}

	// Check for old files
	cutoffTime := time.Now().Add(-sm.RetentionPolicy.MaxAge)
	oldFileCount := 0
	for _, file := range files {
		if file.ModifiedAt.Before(cutoffTime) {
			oldFileCount++
		}
	}
	if oldFileCount > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Found %d old state files that can be cleaned up", oldFileCount))
	}

	// Check for stuck operations
	stuckCount := 0
	for _, file := range files {
		if file.State.Status == StatusInProgress {
			timeSinceUpdate := time.Since(file.State.LastUpdate)
			if timeSinceUpdate > 24*time.Hour {
				stuckCount++
			}
		}
	}
	if stuckCount > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Found %d potentially stuck operations", stuckCount))
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "State management is healthy")
	}

	return recommendations
}

// RepairCorruptedStates attempts to repair corrupted state files
func (sm *StateManager) RepairCorruptedStates() error {
	files, err := os.ReadDir(sm.StateDir)
	if err != nil {
		return fmt.Errorf("failed to read state directory: %w", err)
	}

	var repairedCount int

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(sm.StateDir, file.Name())

		// Try to load the state file
		if _, err := sm.loadStateFile(filePath); err != nil {
			// File is corrupted, attempt repair
			if repaired := sm.attemptRepair(filePath); repaired {
				repairedCount++
			}
		}
	}

	fmt.Printf("Repaired %d corrupted state files\n", repairedCount)
	return nil
}

// attemptRepair attempts to repair a corrupted state file
func (sm *StateManager) attemptRepair(filePath string) bool {
	// Create backup
	backupPath := filePath + ".backup"
	if err := sm.copyFile(filePath, backupPath); err != nil {
		return false
	}

	// Read file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	// Try to partially parse and reconstruct
	// This is a simplified repair - in production, more sophisticated repair logic would be needed
	var partial map[string]interface{}
	if err := json.Unmarshal(data, &partial); err != nil {
		// File is too corrupted, remove it
		os.Remove(filePath)
		return true
	}

	// Reconstruct basic state structure
	repairedState := &OperationState{
		ID:         fmt.Sprintf("repaired-%d", time.Now().Unix()),
		Status:     StatusFailed,
		StartTime:  time.Now().Add(-time.Hour), // Assume recent
		LastUpdate: time.Now(),
	}

	// Save repaired state
	repairedData, err := json.MarshalIndent(repairedState, "", "  ")
	if err != nil {
		return false
	}

	return os.WriteFile(filePath, repairedData, 0o644) == nil
}

// copyFile creates a copy of a file
func (sm *StateManager) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

// SetRetentionPolicy updates the retention policy
func (sm *StateManager) SetRetentionPolicy(policy RetentionPolicy) {
	sm.RetentionPolicy = policy
}
