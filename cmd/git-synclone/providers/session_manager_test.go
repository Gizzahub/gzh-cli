// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package providers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// SessionManagerTestSuite provides comprehensive tests for session management.
type SessionManagerTestSuite struct {
	suite.Suite
	tempDir        string
	sessionManager *SessionManager
}

// SetupTest initializes test environment for each test.
func (s *SessionManagerTestSuite) SetupTest() {
	tempDir, err := os.MkdirTemp("", "session-manager-test-*")
	s.Require().NoError(err)

	s.tempDir = tempDir
	s.sessionManager = NewSessionManager(tempDir)
}

// TearDownTest cleans up after each test.
func (s *SessionManagerTestSuite) TearDownTest() {
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}

// TestStartSession tests creating a new session.
func (s *SessionManagerTestSuite) TestStartSession() {
	ctx := context.WithValue(context.Background(), "provider", "github")

	request := &CloneRequest{
		Organization: "testorg",
		TargetPath:   "/tmp/repos",
		Strategy:     "reset",
		Options: &CloneOptions{
			Parallel:   3,
			MaxRetries: 2,
		},
	}

	session, err := s.sessionManager.StartSession(ctx, request)
	s.NoError(err)
	s.NotNil(session)

	// Verify session properties
	s.Equal("github:testorg", session.SessionID)
	s.Equal("in_progress", session.State.Status)
	s.Equal("github", session.State.Provider)
	s.Equal("testorg", session.State.Organization)
	s.Equal("/tmp/repos", session.State.TargetPath)
	s.Equal("reset", session.State.Strategy)
	s.Equal(3, session.State.Parallel)
	s.Equal(2, session.State.MaxRetries)
	s.False(session.Resumed)

	// Verify state file was created
	expectedPath := filepath.Join(s.tempDir, "github_testorg.json")
	s.FileExists(expectedPath)
}

// TestResumeSession tests resuming an existing session.
func (s *SessionManagerTestSuite) TestResumeSession() {
	ctx := context.WithValue(context.Background(), "provider", "github")

	// First, create a session
	request := &CloneRequest{
		Organization: "testorg",
		TargetPath:   "/tmp/repos",
		Strategy:     "reset",
		Options: &CloneOptions{
			Parallel:   3,
			MaxRetries: 2,
		},
	}

	originalSession, err := s.sessionManager.StartSession(ctx, request)
	s.NoError(err)

	// Add some progress
	err = originalSession.SetPendingRepositories([]string{"repo1", "repo2", "repo3"})
	s.NoError(err)

	err = originalSession.UpdateProgress("repo1", "clone", true, "Success")
	s.NoError(err)

	// Now resume the session
	resumedSession, err := s.sessionManager.ResumeSession(ctx, "github:testorg")
	s.NoError(err)
	s.NotNil(resumedSession)

	// Verify resumed session properties
	s.Equal("github:testorg", resumedSession.SessionID)
	s.Equal("in_progress", resumedSession.State.Status)
	s.True(resumedSession.Resumed)

	// Verify progress was preserved
	completed, failed, pending, _ := resumedSession.GetProgress()
	s.Equal(1, completed)
	s.Equal(0, failed)
	s.Equal(2, pending)
}

// TestResumeCompletedSession tests that completed sessions cannot be resumed.
func (s *SessionManagerTestSuite) TestResumeCompletedSession() {
	ctx := context.WithValue(context.Background(), "provider", "github")

	// Create and complete a session
	request := &CloneRequest{
		Organization: "testorg",
		TargetPath:   "/tmp/repos",
		Strategy:     "reset",
		Options: &CloneOptions{
			Parallel:   3,
			MaxRetries: 2,
		},
	}

	session, err := s.sessionManager.StartSession(ctx, request)
	s.NoError(err)

	err = session.MarkCompleted()
	s.NoError(err)

	// Try to resume completed session
	_, err = s.sessionManager.ResumeSession(ctx, "github:testorg")
	s.Error(err)
	s.Contains(err.Error(), "already completed")
}

// TestListSessions tests listing all sessions.
func (s *SessionManagerTestSuite) TestListSessions() {
	ctx := context.WithValue(context.Background(), "provider", "github")

	// Create multiple sessions
	requests := []*CloneRequest{
		{
			Organization: "org1",
			TargetPath:   "/tmp/org1",
			Strategy:     "reset",
			Options:      &CloneOptions{Parallel: 3, MaxRetries: 2},
		},
		{
			Organization: "org2",
			TargetPath:   "/tmp/org2",
			Strategy:     "pull",
			Options:      &CloneOptions{Parallel: 5, MaxRetries: 3},
		},
	}

	for _, req := range requests {
		_, err := s.sessionManager.StartSession(ctx, req)
		s.NoError(err)
	}

	// List sessions
	sessions, err := s.sessionManager.ListSessions()
	s.NoError(err)
	s.Len(sessions, 2)

	// Verify session info
	sessionIDs := make([]string, len(sessions))
	for i, session := range sessions {
		sessionIDs[i] = session.SessionID
	}
	s.Contains(sessionIDs, "github:org1")
	s.Contains(sessionIDs, "github:org2")
}

// TestDeleteSession tests deleting a session.
func (s *SessionManagerTestSuite) TestDeleteSession() {
	ctx := context.WithValue(context.Background(), "provider", "github")

	request := &CloneRequest{
		Organization: "testorg",
		TargetPath:   "/tmp/repos",
		Strategy:     "reset",
		Options:      &CloneOptions{Parallel: 3, MaxRetries: 2},
	}

	_, err := s.sessionManager.StartSession(ctx, request)
	s.NoError(err)

	// Verify session exists
	s.True(s.sessionManager.HasSession("github:testorg"))

	// Delete session
	err = s.sessionManager.DeleteSession("github:testorg")
	s.NoError(err)

	// Verify session is deleted
	s.False(s.sessionManager.HasSession("github:testorg"))

	// Verify state file is deleted
	expectedPath := filepath.Join(s.tempDir, "github_testorg.json")
	s.NoFileExists(expectedPath)
}

// TestGetLatestSession tests getting the most recent incomplete session.
func (s *SessionManagerTestSuite) TestGetLatestSession() {
	ctx := context.WithValue(context.Background(), "provider", "github")

	// Create first session
	req1 := &CloneRequest{
		Organization: "org1",
		TargetPath:   "/tmp/org1",
		Strategy:     "reset",
		Options:      &CloneOptions{Parallel: 3, MaxRetries: 2},
	}

	session1, err := s.sessionManager.StartSession(ctx, req1)
	s.NoError(err)

	// Wait a bit to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	// Create second session (should be latest)
	req2 := &CloneRequest{
		Organization: "org2",
		TargetPath:   "/tmp/org2",
		Strategy:     "pull",
		Options:      &CloneOptions{Parallel: 5, MaxRetries: 3},
	}

	_, err = s.sessionManager.StartSession(ctx, req2)
	s.NoError(err)

	// Complete first session
	err = session1.MarkCompleted()
	s.NoError(err)

	// Get latest session (should be session2 since session1 is completed)
	latest, err := s.sessionManager.GetLatestSession()
	s.NoError(err)
	s.Equal("github:org2", latest.SessionID)
	s.Equal("in_progress", latest.Status)
}

// TestGetLatestSessionNoIncomplete tests when no incomplete sessions exist.
func (s *SessionManagerTestSuite) TestGetLatestSessionNoIncomplete() {
	ctx := context.WithValue(context.Background(), "provider", "github")

	// Create and complete a session
	request := &CloneRequest{
		Organization: "testorg",
		TargetPath:   "/tmp/repos",
		Strategy:     "reset",
		Options:      &CloneOptions{Parallel: 3, MaxRetries: 2},
	}

	session, err := s.sessionManager.StartSession(ctx, request)
	s.NoError(err)

	err = session.MarkCompleted()
	s.NoError(err)

	// Try to get latest session
	_, err = s.sessionManager.GetLatestSession()
	s.Error(err)
	s.Contains(err.Error(), "no incomplete sessions found")
}

// TestCleanupCompletedSessions tests cleaning up old completed sessions.
func (s *SessionManagerTestSuite) TestCleanupCompletedSessions() {
	ctx := context.WithValue(context.Background(), "provider", "github")

	// Create sessions
	req1 := &CloneRequest{
		Organization: "old-org",
		TargetPath:   "/tmp/old",
		Strategy:     "reset",
		Options:      &CloneOptions{Parallel: 3, MaxRetries: 2},
	}

	req2 := &CloneRequest{
		Organization: "new-org",
		TargetPath:   "/tmp/new",
		Strategy:     "reset",
		Options:      &CloneOptions{Parallel: 3, MaxRetries: 2},
	}

	oldSession, err := s.sessionManager.StartSession(ctx, req1)
	s.NoError(err)

	newSession, err := s.sessionManager.StartSession(ctx, req2)
	s.NoError(err)

	// Complete both sessions
	err = oldSession.MarkCompleted()
	s.NoError(err)

	err = newSession.MarkCompleted()
	s.NoError(err)

	// Manually update old session's timestamp to make it appear old
	oldSession.State.LastUpdated = time.Now().Add(-2 * time.Hour)
	err = oldSession.Manager.stateManager.SaveState(oldSession.State)
	s.NoError(err)

	// Cleanup sessions older than 1 hour
	err = s.sessionManager.CleanupCompletedSessions(1 * time.Hour)
	s.NoError(err)

	// Verify old session is deleted, new session remains
	s.False(s.sessionManager.HasSession("github:old-org"))
	s.True(s.sessionManager.HasSession("github:new-org"))
}

// TestCleanupFailedSessions tests cleaning up old failed sessions.
func (s *SessionManagerTestSuite) TestCleanupFailedSessions() {
	ctx := context.WithValue(context.Background(), "provider", "github")

	request := &CloneRequest{
		Organization: "failed-org",
		TargetPath:   "/tmp/failed",
		Strategy:     "reset",
		Options:      &CloneOptions{Parallel: 3, MaxRetries: 2},
	}

	session, err := s.sessionManager.StartSession(ctx, request)
	s.NoError(err)

	// Mark session as failed
	err = session.MarkFailed()
	s.NoError(err)

	// Manually update timestamp to make it appear old
	session.State.LastUpdated = time.Now().Add(-2 * time.Hour)
	err = session.Manager.stateManager.SaveState(session.State)
	s.NoError(err)

	// Cleanup failed sessions older than 1 hour
	err = s.sessionManager.CleanupFailedSessions(1 * time.Hour)
	s.NoError(err)

	// Verify session is deleted
	s.False(s.sessionManager.HasSession("github:failed-org"))
}

// TestSessionProgress tests session progress tracking.
func (s *SessionManagerTestSuite) TestSessionProgress() {
	ctx := context.WithValue(context.Background(), "provider", "github")

	request := &CloneRequest{
		Organization: "testorg",
		TargetPath:   "/tmp/repos",
		Strategy:     "reset",
		Options:      &CloneOptions{Parallel: 3, MaxRetries: 2},
	}

	session, err := s.sessionManager.StartSession(ctx, request)
	s.NoError(err)

	// Set pending repositories
	repos := []string{"repo1", "repo2", "repo3", "repo4", "repo5"}
	err = session.SetPendingRepositories(repos)
	s.NoError(err)

	// Update progress
	err = session.UpdateProgress("repo1", "clone", true, "Success")
	s.NoError(err)

	err = session.UpdateProgress("repo2", "clone", true, "Success")
	s.NoError(err)

	err = session.UpdateProgress("repo3", "clone", false, "Failed to clone")
	s.NoError(err)

	// Check progress
	completed, failed, pending, percent := session.GetProgress()
	s.Equal(2, completed)
	s.Equal(1, failed)
	s.Equal(2, pending)
	s.Equal(60.0, percent) // 3/5 * 100

	// Check remaining repositories
	remaining := session.GetRemainingRepositories()
	s.ElementsMatch([]string{"repo4", "repo5"}, remaining)
}

// TestInvalidSessionID tests handling of invalid session IDs.
func (s *SessionManagerTestSuite) TestInvalidSessionID() {
	ctx := context.Background()

	// Test invalid format
	_, err := s.sessionManager.ResumeSession(ctx, "invalid-format")
	s.Error(err)
	s.Contains(err.Error(), "session ID must be in format 'provider:organization'")

	// Test non-existent session
	_, err = s.sessionManager.ResumeSession(ctx, "github:nonexistent")
	s.Error(err)
	s.Contains(err.Error(), "failed to load session state")

	// Test delete non-existent session
	err = s.sessionManager.DeleteSession("github:nonexistent")
	s.NoError(err) // Should not error for non-existent sessions
}

// Test helper functions

// TestGenerateSessionID tests session ID generation.
func (s *SessionManagerTestSuite) TestGenerateSessionID() {
	sessionID := generateSessionID("github", "testorg")
	s.Equal("github:testorg", sessionID)
}

// TestParseSessionID tests session ID parsing.
func (s *SessionManagerTestSuite) TestParseSessionID() {
	// Valid session ID
	provider, org, err := parseSessionID("github:testorg")
	s.NoError(err)
	s.Equal("github", provider)
	s.Equal("testorg", org)

	// Invalid session ID
	_, _, err = parseSessionID("invalid")
	s.Error(err)
	s.Contains(err.Error(), "session ID must be in format 'provider:organization'")

	// Empty session ID
	_, _, err = parseSessionID("")
	s.Error(err)

	// Session ID with multiple colons
	_, _, err = parseSessionID("github:org:extra")
	s.Error(err)
}

// TestGetProviderFromContext tests provider extraction from context.
func (s *SessionManagerTestSuite) TestGetProviderFromContext() {
	// Context with provider
	ctx := context.WithValue(context.Background(), "provider", "github")
	provider := getProviderFromContext(ctx)
	s.Equal("github", provider)

	// Context without provider
	ctx = context.Background()
	provider = getProviderFromContext(ctx)
	s.Equal("unknown", provider)
}

// TestSessionManagerWithEmptyStateDir tests creating session manager with empty state directory.
func (s *SessionManagerTestSuite) TestSessionManagerWithEmptyStateDir() {
	sm := NewSessionManager("")
	s.NotNil(sm)

	// Should use default state directory
	s.Contains(sm.stateManager.GetStateFilePath("github", "testorg"), ".gzh/state")
}

// Run the test suite
func TestSessionManagerTestSuite(t *testing.T) {
	suite.Run(t, new(SessionManagerTestSuite))
}

// Unit tests for individual functions

func TestGenerateSessionID(t *testing.T) {
	tests := []struct {
		provider     string
		organization string
		expected     string
	}{
		{"github", "testorg", "github:testorg"},
		{"gitlab", "mygroup", "gitlab:mygroup"},
		{"gitea", "team", "gitea:team"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.provider, tt.organization), func(t *testing.T) {
			result := generateSessionID(tt.provider, tt.organization)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseSessionID(t *testing.T) {
	tests := []struct {
		sessionID            string
		expectedProvider     string
		expectedOrganization string
		expectError          bool
	}{
		{"github:testorg", "github", "testorg", false},
		{"gitlab:mygroup", "gitlab", "mygroup", false},
		{"gitea:team", "gitea", "team", false},
		{"invalid", "", "", true},
		{"", "", "", true},
		{"github:org:extra", "", "", true},
		{"github:", "", "", true},
		{":testorg", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.sessionID, func(t *testing.T) {
			provider, org, err := parseSessionID(tt.sessionID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedProvider, provider)
				assert.Equal(t, tt.expectedOrganization, org)
			}
		})
	}
}
