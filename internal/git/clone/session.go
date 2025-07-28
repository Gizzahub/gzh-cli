// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package clone

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Session represents a clone operation session for resumability.
type Session struct {
	ID           string                       `json:"id"`
	StartedAt    time.Time                    `json:"started_at"`
	UpdatedAt    time.Time                    `json:"updated_at"`
	Options      *CloneOptions                `json:"options"`
	Repositories map[string]*RepositoryStatus `json:"repositories"`
	Statistics   *SessionStatistics           `json:"statistics"`
}

// RepositoryStatus represents the status of a repository in a session.
type RepositoryStatus struct {
	Status      string    `json:"status"` // pending, cloning, completed, failed
	Error       string    `json:"error,omitempty"`
	StartedAt   time.Time `json:"started_at,omitempty"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	Attempts    int       `json:"attempts"`
	LastAttempt time.Time `json:"last_attempt,omitempty"`
}

// SessionStatistics contains session-level statistics.
type SessionStatistics struct {
	TotalRepositories int `json:"total_repositories"`
	CompletedCount    int `json:"completed_count"`
	FailedCount       int `json:"failed_count"`
	PendingCount      int `json:"pending_count"`
	InProgressCount   int `json:"in_progress_count"`
}

// NewSession creates a new clone session.
func NewSession(opts *CloneOptions) *Session {
	sessionID := generateSessionID()
	now := time.Now()

	return &Session{
		ID:           sessionID,
		StartedAt:    now,
		UpdatedAt:    now,
		Options:      opts,
		Repositories: make(map[string]*RepositoryStatus),
		Statistics: &SessionStatistics{
			TotalRepositories: 0,
			CompletedCount:    0,
			FailedCount:       0,
			PendingCount:      0,
			InProgressCount:   0,
		},
	}
}

// Initialize initializes the session and creates the session directory.
func (s *Session) Initialize() error {
	sessionDir := getSessionDir()
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		return fmt.Errorf("failed to create session directory: %w", err)
	}

	return s.Save()
}

// Load loads a session from disk.
func (s *Session) Load(sessionID string) error {
	sessionFile := getSessionFile(sessionID)
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return fmt.Errorf("failed to read session file: %w", err)
	}

	if err := json.Unmarshal(data, s); err != nil {
		return fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return nil
}

// Save saves the session to disk.
func (s *Session) Save() error {
	s.UpdatedAt = time.Now()
	s.updateStatistics()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	sessionFile := getSessionFile(s.ID)
	if err := os.WriteFile(sessionFile, data, 0o644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// AddRepository adds a repository to the session.
func (s *Session) AddRepository(repoName string) {
	if s.Repositories == nil {
		s.Repositories = make(map[string]*RepositoryStatus)
	}

	s.Repositories[repoName] = &RepositoryStatus{
		Status:    "pending",
		StartedAt: time.Time{},
		Attempts:  0,
	}
}

// MarkStarted marks a repository as started.
func (s *Session) MarkStarted(repoName string) {
	if status, exists := s.Repositories[repoName]; exists {
		status.Status = "cloning"
		status.StartedAt = time.Now()
		status.Attempts++
		status.LastAttempt = time.Now()
	}
}

// MarkCompleted marks a repository as completed.
func (s *Session) MarkCompleted(repoName string) {
	if status, exists := s.Repositories[repoName]; exists {
		status.Status = "completed"
		status.CompletedAt = time.Now()
		status.Error = ""
	}
}

// MarkFailed marks a repository as failed.
func (s *Session) MarkFailed(repoName string, err error) {
	if status, exists := s.Repositories[repoName]; exists {
		status.Status = "failed"
		status.Error = err.Error()
		status.CompletedAt = time.Now()
	}
}

// IsCompleted checks if a repository is completed.
func (s *Session) IsCompleted(repoName string) bool {
	if status, exists := s.Repositories[repoName]; exists {
		return status.Status == "completed"
	}
	return false
}

// IsFailed checks if a repository is failed.
func (s *Session) IsFailed(repoName string) bool {
	if status, exists := s.Repositories[repoName]; exists {
		return status.Status == "failed"
	}
	return false
}

// GetStatus returns the status of a repository.
func (s *Session) GetStatus(repoName string) *RepositoryStatus {
	if status, exists := s.Repositories[repoName]; exists {
		return status
	}
	return nil
}

// GetCompletedRepositories returns a list of completed repositories.
func (s *Session) GetCompletedRepositories() []string {
	var completed []string
	for repoName, status := range s.Repositories {
		if status.Status == "completed" {
			completed = append(completed, repoName)
		}
	}
	return completed
}

// GetFailedRepositories returns a list of failed repositories.
func (s *Session) GetFailedRepositories() []string {
	var failed []string
	for repoName, status := range s.Repositories {
		if status.Status == "failed" {
			failed = append(failed, repoName)
		}
	}
	return failed
}

// GetPendingRepositories returns a list of pending repositories.
func (s *Session) GetPendingRepositories() []string {
	var pending []string
	for repoName, status := range s.Repositories {
		if status.Status == "pending" {
			pending = append(pending, repoName)
		}
	}
	return pending
}

// GetProgress returns the current progress as a percentage.
func (s *Session) GetProgress() float64 {
	if s.Statistics.TotalRepositories == 0 {
		return 0.0
	}
	completed := s.Statistics.CompletedCount + s.Statistics.FailedCount
	return float64(completed) / float64(s.Statistics.TotalRepositories) * 100.0
}

// updateStatistics updates the session statistics.
func (s *Session) updateStatistics() {
	s.Statistics.TotalRepositories = len(s.Repositories)
	s.Statistics.CompletedCount = 0
	s.Statistics.FailedCount = 0
	s.Statistics.PendingCount = 0
	s.Statistics.InProgressCount = 0

	for _, status := range s.Repositories {
		switch status.Status {
		case "completed":
			s.Statistics.CompletedCount++
		case "failed":
			s.Statistics.FailedCount++
		case "pending":
			s.Statistics.PendingCount++
		case "cloning":
			s.Statistics.InProgressCount++
		}
	}
}

// Delete removes the session file from disk.
func (s *Session) Delete() error {
	sessionFile := getSessionFile(s.ID)
	return os.Remove(sessionFile)
}

// GetDuration returns the duration since the session started.
func (s *Session) GetDuration() time.Duration {
	return time.Since(s.StartedAt)
}

// IsActive checks if the session is active (has pending or in-progress repositories).
func (s *Session) IsActive() bool {
	s.updateStatistics()
	return s.Statistics.PendingCount > 0 || s.Statistics.InProgressCount > 0
}

// Cleanup removes old session files (older than specified duration).
func CleanupOldSessions(olderThan time.Duration) error {
	sessionDir := getSessionDir()
	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		return fmt.Errorf("failed to read session directory: %w", err)
	}

	cutoff := time.Now().Add(-olderThan)
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			if info.ModTime().Before(cutoff) {
				sessionFile := filepath.Join(sessionDir, entry.Name())
				if err := os.Remove(sessionFile); err != nil {
					// Log error but continue cleanup
					continue
				}
			}
		}
	}

	return nil
}

// ListSessions returns a list of existing session IDs.
func ListSessions() ([]string, error) {
	sessionDir := getSessionDir()
	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read session directory: %w", err)
	}

	var sessionIDs []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			sessionID := entry.Name()[:len(entry.Name())-5] // Remove .json extension
			sessionIDs = append(sessionIDs, sessionID)
		}
	}

	return sessionIDs, nil
}

// Helper functions

// generateSessionID generates a random session ID.
func generateSessionID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// getSessionDir returns the session directory path.
func getSessionDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to current directory if user config dir is not available
		return filepath.Join(".", ".gzh-sessions")
	}
	return filepath.Join(configDir, "gzh-manager", "sessions")
}

// getSessionFile returns the session file path for a given session ID.
func getSessionFile(sessionID string) string {
	return filepath.Join(getSessionDir(), sessionID+".json")
}

// SessionExists checks if a session file exists.
func SessionExists(sessionID string) bool {
	sessionFile := getSessionFile(sessionID)
	_, err := os.Stat(sessionFile)
	return err == nil
}

// LoadSessionInfo loads basic session information without full session data.
func LoadSessionInfo(sessionID string) (*SessionInfo, error) {
	sessionFile := getSessionFile(sessionID)
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	session.updateStatistics()

	return &SessionInfo{
		ID:           session.ID,
		StartedAt:    session.StartedAt,
		UpdatedAt:    session.UpdatedAt,
		Provider:     session.Options.Provider,
		Organization: session.Options.Org,
		Target:       session.Options.Target,
		Statistics:   session.Statistics,
	}, nil
}

// SessionInfo contains basic information about a session.
type SessionInfo struct {
	ID           string             `json:"id"`
	StartedAt    time.Time          `json:"started_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	Provider     string             `json:"provider"`
	Organization string             `json:"organization"`
	Target       string             `json:"target"`
	Statistics   *SessionStatistics `json:"statistics"`
}
