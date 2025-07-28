// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package providers

import (
	"context"
	"fmt"
	"time"

	bulkclone "github.com/gizzahub/gzh-manager-go/pkg/synclone"
)

// SessionManager provides session and state management for git-synclone operations.
// It integrates with the existing synclone state management to provide resumable operations.
type SessionManager struct {
	stateManager *bulkclone.StateManager
}

// NewSessionManager creates a new session manager with state persistence.
func NewSessionManager(stateDir string) *SessionManager {
	return &SessionManager{
		stateManager: bulkclone.NewStateManager(stateDir),
	}
}

// StartSession creates a new clone session and returns the session ID.
func (sm *SessionManager) StartSession(ctx context.Context, request *CloneRequest) (*CloneSession, error) {
	// Create new clone state
	state := bulkclone.NewCloneState(
		getProviderFromContext(ctx),
		request.Organization,
		request.TargetPath,
		request.Strategy,
		request.Options.Parallel,
		request.Options.MaxRetries,
	)

	// Save initial state
	if err := sm.stateManager.SaveState(state); err != nil {
		return nil, fmt.Errorf("failed to save initial session state: %w", err)
	}

	return &CloneSession{
		SessionID:    generateSessionID(getProviderFromContext(ctx), request.Organization),
		State:        state,
		Manager:      sm,
		StartTime:    time.Now(),
		LastActivity: time.Now(),
	}, nil
}

// ResumeSession resumes an existing clone session.
func (sm *SessionManager) ResumeSession(ctx context.Context, sessionID string) (*CloneSession, error) {
	provider, org, err := parseSessionID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	// Load existing state
	state, err := sm.stateManager.LoadState(provider, org)
	if err != nil {
		return nil, fmt.Errorf("failed to load session state: %w", err)
	}

	// Validate state can be resumed
	if state.Status == "completed" {
		return nil, fmt.Errorf("session %s is already completed", sessionID)
	}

	return &CloneSession{
		SessionID:    sessionID,
		State:        state,
		Manager:      sm,
		StartTime:    state.StartTime,
		LastActivity: time.Now(),
		Resumed:      true,
	}, nil
}

// ListSessions returns all available sessions.
func (sm *SessionManager) ListSessions() ([]SessionInfo, error) {
	states, err := sm.stateManager.ListStates()
	if err != nil {
		return nil, fmt.Errorf("failed to list states: %w", err)
	}

	sessions := make([]SessionInfo, 0, len(states))
	for _, state := range states {
		sessionID := generateSessionID(state.Provider, state.Organization)

		completed, failed, pending := state.GetProgress()
		sessions = append(sessions, SessionInfo{
			SessionID:         sessionID,
			Provider:          state.Provider,
			Organization:      state.Organization,
			TargetPath:        state.TargetPath,
			Strategy:          state.Strategy,
			Status:            state.Status,
			StartTime:         state.StartTime,
			LastUpdated:       state.LastUpdated,
			TotalRepositories: state.TotalRepositories,
			CompletedRepos:    completed,
			FailedRepos:       failed,
			PendingRepos:      pending,
			ProgressPercent:   state.GetProgressPercent(),
		})
	}

	return sessions, nil
}

// DeleteSession removes a session and its state.
func (sm *SessionManager) DeleteSession(sessionID string) error {
	provider, org, err := parseSessionID(sessionID)
	if err != nil {
		return fmt.Errorf("invalid session ID: %w", err)
	}

	return sm.stateManager.DeleteState(provider, org)
}

// HasSession checks if a session exists.
func (sm *SessionManager) HasSession(sessionID string) bool {
	provider, org, err := parseSessionID(sessionID)
	if err != nil {
		return false
	}

	return sm.stateManager.HasState(provider, org)
}

// GetLatestSession returns the most recent incomplete session.
func (sm *SessionManager) GetLatestSession() (*SessionInfo, error) {
	sessions, err := sm.ListSessions()
	if err != nil {
		return nil, err
	}

	var latest *SessionInfo
	for i := range sessions {
		session := &sessions[i]
		if session.Status == "in_progress" {
			if latest == nil || session.LastUpdated.After(latest.LastUpdated) {
				latest = session
			}
		}
	}

	if latest == nil {
		return nil, fmt.Errorf("no incomplete sessions found")
	}

	return latest, nil
}

// CloneSession represents an active or resumable clone session.
type CloneSession struct {
	SessionID    string
	State        *bulkclone.CloneState
	Manager      *SessionManager
	StartTime    time.Time
	LastActivity time.Time
	Resumed      bool
}

// UpdateProgress updates the session progress and saves state.
func (cs *CloneSession) UpdateProgress(repoName, operation string, success bool, message string) error {
	cs.LastActivity = time.Now()

	if success {
		cs.State.AddCompletedRepository(repoName, "", operation, message)
	} else {
		cs.State.AddFailedRepository(repoName, "", operation, message, 1)
	}

	return cs.Manager.stateManager.SaveState(cs.State)
}

// SetPendingRepositories sets the list of repositories to be processed.
func (cs *CloneSession) SetPendingRepositories(repos []string) error {
	cs.State.SetPendingRepositories(repos)
	return cs.Manager.stateManager.SaveState(cs.State)
}

// GetRemainingRepositories returns repositories that still need processing.
func (cs *CloneSession) GetRemainingRepositories() []string {
	return cs.State.GetRemainingRepositories()
}

// MarkCompleted marks the session as completed.
func (cs *CloneSession) MarkCompleted() error {
	cs.State.MarkCompleted()
	return cs.Manager.stateManager.SaveState(cs.State)
}

// MarkFailed marks the session as failed.
func (cs *CloneSession) MarkFailed() error {
	cs.State.MarkFailed()
	return cs.Manager.stateManager.SaveState(cs.State)
}

// GetProgress returns current progress statistics.
func (cs *CloneSession) GetProgress() (completed, failed, pending int, percent float64) {
	completed, failed, pending = cs.State.GetProgress()
	percent = cs.State.GetProgressPercent()
	return
}

// SessionInfo provides summary information about a session.
type SessionInfo struct {
	SessionID         string    `json:"sessionId"`
	Provider          string    `json:"provider"`
	Organization      string    `json:"organization"`
	TargetPath        string    `json:"targetPath"`
	Strategy          string    `json:"strategy"`
	Status            string    `json:"status"`
	StartTime         time.Time `json:"startTime"`
	LastUpdated       time.Time `json:"lastUpdated"`
	TotalRepositories int       `json:"totalRepositories"`
	CompletedRepos    int       `json:"completedRepos"`
	FailedRepos       int       `json:"failedRepos"`
	PendingRepos      int       `json:"pendingRepos"`
	ProgressPercent   float64   `json:"progressPercent"`
}

// Helper functions

// generateSessionID creates a session ID from provider and organization.
func generateSessionID(provider, organization string) string {
	return fmt.Sprintf("%s:%s", provider, organization)
}

// parseSessionID extracts provider and organization from session ID.
func parseSessionID(sessionID string) (provider, organization string, err error) {
	parts := []string{}
	current := ""

	for _, r := range sessionID {
		if r == ':' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	if len(parts) != 2 {
		return "", "", fmt.Errorf("session ID must be in format 'provider:organization'")
	}

	return parts[0], parts[1], nil
}

// getProviderFromContext extracts provider name from context.
func getProviderFromContext(ctx context.Context) string {
	if provider, ok := ctx.Value("provider").(string); ok {
		return provider
	}
	return "unknown"
}

// CleanupCompletedSessions removes completed sessions older than the specified duration.
func (sm *SessionManager) CleanupCompletedSessions(olderThan time.Duration) error {
	sessions, err := sm.ListSessions()
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-olderThan)

	for _, session := range sessions {
		if session.Status == "completed" && session.LastUpdated.Before(cutoff) {
			if err := sm.DeleteSession(session.SessionID); err != nil {
				// Log error but continue with other sessions
				continue
			}
		}
	}

	return nil
}

// CleanupFailedSessions removes failed sessions older than the specified duration.
func (sm *SessionManager) CleanupFailedSessions(olderThan time.Duration) error {
	sessions, err := sm.ListSessions()
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-olderThan)

	for _, session := range sessions {
		if session.Status == "failed" && session.LastUpdated.Before(cutoff) {
			if err := sm.DeleteSession(session.SessionID); err != nil {
				// Log error but continue with other sessions
				continue
			}
		}
	}

	return nil
}
