//nolint:testpackage // White-box testing needed for internal function access
package bulkclone

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloneState_NewCloneState(t *testing.T) {
	state := NewCloneState("github", "myorg", "/tmp/repos", "reset", 10, 3)

	assert.Equal(t, "github", state.Provider)
	assert.Equal(t, "myorg", state.Organization)
	assert.Equal(t, "/tmp/repos", state.TargetPath)
	assert.Equal(t, "reset", state.Strategy)
	assert.Equal(t, 10, state.Parallel)
	assert.Equal(t, 3, state.MaxRetries)
	assert.Equal(t, "in_progress", state.Status)
	assert.Equal(t, 0, state.TotalRepositories)
	assert.Empty(t, state.CompletedRepos)
	assert.Empty(t, state.FailedRepos)
	assert.Empty(t, state.PendingRepos)
}

func TestCloneState_AddCompletedRepository(t *testing.T) {
	state := NewCloneState("github", "myorg", "/tmp/repos", "reset", 10, 3)

	state.AddCompletedRepository("repo1", "/tmp/repos/repo1", "clone", "Successfully cloned")

	assert.Len(t, state.CompletedRepos, 1)
	assert.Equal(t, "repo1", state.CompletedRepos[0].Name)
	assert.Equal(t, "/tmp/repos/repo1", state.CompletedRepos[0].Path)
	assert.Equal(t, "clone", state.CompletedRepos[0].Operation)
	assert.Equal(t, "Successfully cloned", state.CompletedRepos[0].Message)
}

func TestCloneState_AddFailedRepository(t *testing.T) {
	state := NewCloneState("github", "myorg", "/tmp/repos", "reset", 10, 3)

	state.AddFailedRepository("repo1", "/tmp/repos/repo1", "clone", "Network error", 2)

	assert.Len(t, state.FailedRepos, 1)
	assert.Equal(t, "repo1", state.FailedRepos[0].Name)
	assert.Equal(t, "/tmp/repos/repo1", state.FailedRepos[0].Path)
	assert.Equal(t, "clone", state.FailedRepos[0].Operation)
	assert.Equal(t, "Network error", state.FailedRepos[0].Error)
	assert.Equal(t, 2, state.FailedRepos[0].Attempts)
}

func TestCloneState_IsCompleted(t *testing.T) {
	state := NewCloneState("github", "myorg", "/tmp/repos", "reset", 10, 3)

	assert.False(t, state.IsCompleted("repo1"))

	state.AddCompletedRepository("repo1", "/tmp/repos/repo1", "clone", "")

	assert.True(t, state.IsCompleted("repo1"))
	assert.False(t, state.IsCompleted("repo2"))
}

func TestCloneState_IsFailed(t *testing.T) {
	state := NewCloneState("github", "myorg", "/tmp/repos", "reset", 10, 3)

	assert.False(t, state.IsFailed("repo1"))

	state.AddFailedRepository("repo1", "/tmp/repos/repo1", "clone", "Error", 1)

	assert.True(t, state.IsFailed("repo1"))
	assert.False(t, state.IsFailed("repo2"))
}

func TestCloneState_GetProgress(t *testing.T) {
	state := NewCloneState("github", "myorg", "/tmp/repos", "reset", 10, 3)
	state.SetPendingRepositories([]string{"repo1", "repo2", "repo3", "repo4"})

	// Initial progress
	completed, failed, pending := state.GetProgress()
	assert.Equal(t, 0, completed)
	assert.Equal(t, 0, failed)
	assert.Equal(t, 4, pending)

	// Add completed repo
	state.AddCompletedRepository("repo1", "/tmp/repos/repo1", "clone", "")
	completed, failed, pending = state.GetProgress()
	assert.Equal(t, 1, completed)
	assert.Equal(t, 0, failed)
	assert.Equal(t, 3, pending)

	// Add failed repo
	state.AddFailedRepository("repo2", "/tmp/repos/repo2", "clone", "Error", 1)
	completed, failed, pending = state.GetProgress()
	assert.Equal(t, 1, completed)
	assert.Equal(t, 1, failed)
	assert.Equal(t, 2, pending)
}

func TestCloneState_GetProgressPercent(t *testing.T) {
	state := NewCloneState("github", "myorg", "/tmp/repos", "reset", 10, 3)
	state.SetPendingRepositories([]string{"repo1", "repo2", "repo3", "repo4"})

	// Initial progress
	assert.Equal(t, 0.0, state.GetProgressPercent())

	// 25% complete
	state.AddCompletedRepository("repo1", "/tmp/repos/repo1", "clone", "")
	assert.Equal(t, 25.0, state.GetProgressPercent())

	// 50% complete (25% completed + 25% failed)
	state.AddFailedRepository("repo2", "/tmp/repos/repo2", "clone", "Error", 1)
	assert.Equal(t, 50.0, state.GetProgressPercent())
}

func TestStateManager_SaveAndLoadState(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gzh-test-*")
	require.NoError(t, err)

	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	sm := NewStateManager(tempDir)

	// Create test state
	state := NewCloneState("github", "myorg", "/tmp/repos", "reset", 10, 3)
	state.SetPendingRepositories([]string{"repo1", "repo2"})
	state.AddCompletedRepository("repo3", "/tmp/repos/repo3", "clone", "")

	// Save state
	err = sm.SaveState(state)
	require.NoError(t, err)

	// Verify state file exists
	statePath := sm.GetStateFilePath("github", "myorg")
	assert.FileExists(t, statePath)

	// Load state
	loadedState, err := sm.LoadState("github", "myorg")
	require.NoError(t, err)

	// Verify loaded state
	assert.Equal(t, state.Provider, loadedState.Provider)
	assert.Equal(t, state.Organization, loadedState.Organization)
	assert.Equal(t, state.TargetPath, loadedState.TargetPath)
	assert.Equal(t, state.Strategy, loadedState.Strategy)
	assert.Equal(t, state.Parallel, loadedState.Parallel)
	assert.Equal(t, state.MaxRetries, loadedState.MaxRetries)
	assert.Equal(t, state.Status, loadedState.Status)
	assert.Equal(t, len(state.PendingRepos), len(loadedState.PendingRepos))
	assert.Equal(t, len(state.CompletedRepos), len(loadedState.CompletedRepos))
}

func TestStateManager_HasState(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gzh-test-*")
	require.NoError(t, err)

	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	sm := NewStateManager(tempDir)

	// Initially no state
	assert.False(t, sm.HasState("github", "myorg"))

	// Save state
	state := NewCloneState("github", "myorg", "/tmp/repos", "reset", 10, 3)
	err = sm.SaveState(state)
	require.NoError(t, err)

	// Now state exists
	assert.True(t, sm.HasState("github", "myorg"))
	assert.False(t, sm.HasState("gitlab", "myorg"))
}

func TestStateManager_DeleteState(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gzh-test-*")
	require.NoError(t, err)

	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	sm := NewStateManager(tempDir)

	// Save state
	state := NewCloneState("github", "myorg", "/tmp/repos", "reset", 10, 3)
	err = sm.SaveState(state)
	require.NoError(t, err)

	// Verify state exists
	assert.True(t, sm.HasState("github", "myorg"))

	// Delete state
	err = sm.DeleteState("github", "myorg")
	require.NoError(t, err)

	// Verify state is deleted
	assert.False(t, sm.HasState("github", "myorg"))
}

func TestStateManager_ListStates(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gzh-test-*")
	require.NoError(t, err)

	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	sm := NewStateManager(tempDir)

	// Initially no states
	states, err := sm.ListStates()
	require.NoError(t, err)
	assert.Empty(t, states)

	// Save multiple states
	state1 := NewCloneState("github", "org1", "/tmp/repos1", "reset", 10, 3)
	state2 := NewCloneState("gitlab", "org2", "/tmp/repos2", "pull", 5, 2)

	err = sm.SaveState(state1)
	require.NoError(t, err)
	err = sm.SaveState(state2)
	require.NoError(t, err)

	// List states
	states, err = sm.ListStates()
	require.NoError(t, err)
	assert.Len(t, states, 2)

	// Check that we have both states (order might vary)
	providers := []string{states[0].Provider, states[1].Provider}
	assert.Contains(t, providers, "github")
	assert.Contains(t, providers, "gitlab")
}

func TestStateManager_DefaultStateDir(t *testing.T) {
	sm := NewStateManager("")

	homeDir, _ := os.UserHomeDir()
	expectedDir := filepath.Join(homeDir, ".gzh", "state")

	assert.Equal(t, expectedDir, sm.stateDir)
}

func TestCloneState_SetPendingRepositories(t *testing.T) {
	state := NewCloneState("github", "myorg", "/tmp/repos", "reset", 10, 3)

	repos := []string{"repo1", "repo2", "repo3"}
	state.SetPendingRepositories(repos)

	assert.Equal(t, repos, state.PendingRepos)
	assert.Equal(t, 3, state.TotalRepositories)

	// Add completed repo and check total is updated
	state.AddCompletedRepository("repo4", "/tmp/repos/repo4", "clone", "")
	assert.Equal(t, 4, state.TotalRepositories) // 3 pending + 1 completed
}

func TestCloneState_MarkStatus(t *testing.T) {
	state := NewCloneState("github", "myorg", "/tmp/repos", "reset", 10, 3)

	initialTime := state.LastUpdated

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Mark as completed
	state.MarkCompleted()
	assert.Equal(t, "completed", state.Status)
	assert.True(t, state.LastUpdated.After(initialTime))

	// Mark as failed
	state.MarkFailed()
	assert.Equal(t, "failed", state.Status)

	// Mark as canceled
	state.MarkCancelled()
	assert.Equal(t, "canceled", state.Status)
}
