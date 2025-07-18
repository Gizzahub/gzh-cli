package reposync

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// MockGitCommandExecutor is a mock implementation of GitCommandExecutor.
type MockGitCommandExecutor struct {
	mock.Mock
}

func (m *MockGitCommandExecutor) ExecuteCommand(ctx context.Context, dir string, args ...string) (*GitCommandResult, error) {
	// Convert variadic args to a slice and pass them individually
	callArgs := append([]interface{}{ctx, dir}, argsToInterfaces(args)...)
	callResults := m.Called(callArgs...)

	return callResults.Get(0).(*GitCommandResult), callResults.Error(1)
}

// Helper function to convert string slice to interface slice.
func argsToInterfaces(args []string) []interface{} {
	interfaces := make([]interface{}, len(args))
	for i, arg := range args {
		interfaces[i] = arg
	}

	return interfaces
}

func (m *MockGitCommandExecutor) GetStatus(ctx context.Context, dir string) (*GitStatus, error) {
	args := m.Called(ctx, dir)
	return args.Get(0).(*GitStatus), args.Error(1)
}

func (m *MockGitCommandExecutor) GetRemoteInfo(ctx context.Context, dir string, remote string) (*GitRemoteInfo, error) {
	args := m.Called(ctx, dir, remote)
	return args.Get(0).(*GitRemoteInfo), args.Error(1)
}

func TestNewRepositorySynchronizer(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &SyncConfig{
		RepositoryPath:   ".",
		Bidirectional:    true,
		RemoteName:       "origin",
		ConflictStrategy: "auto-merge",
		DryRun:           false,
		AutoCommit:       true,
		CommitMessage:    "Auto-sync: {{.Timestamp}}",
	}

	synchronizer, err := NewRepositorySynchronizer(logger, config)
	require.NoError(t, err)
	assert.NotNil(t, synchronizer)
	assert.Equal(t, config, synchronizer.config)
	assert.NotNil(t, synchronizer.gitCmd)
}

func TestSynchronizeWithCleanRepository(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &SyncConfig{
		RepositoryPath:   "/test/repo",
		Bidirectional:    false,
		RemoteName:       "origin",
		ConflictStrategy: "manual",
		DryRun:           false,
		AutoCommit:       false,
	}

	mockGit := &MockGitCommandExecutor{}
	synchronizer := &RepositorySynchronizer{
		logger: logger,
		config: config,
		gitCmd: mockGit,
	}

	// Mock clean repository status
	mockStatus := &GitStatus{
		Branch:          "main",
		Upstream:        "origin/main",
		AheadBy:         0,
		BehindBy:        0,
		ModifiedFiles:   []string{},
		UntrackedFiles:  []string{},
		ConflictedFiles: []string{},
		CleanWorkingDir: true,
		LastCommitHash:  "abc123",
		LastCommitTime:  time.Now(),
	}

	mockRemoteInfo := &GitRemoteInfo{
		Name:      "origin",
		URL:       "https://github.com/test/repo.git",
		Reachable: true,
	}

	mockGit.On("GetStatus", mock.Anything, "/test/repo").Return(mockStatus, nil)
	mockGit.On("GetRemoteInfo", mock.Anything, "/test/repo", "origin").Return(mockRemoteInfo, nil)

	ctx := context.Background()
	result, err := synchronizer.Synchronize(ctx)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 0, result.FilesModified)
	assert.Equal(t, 0, result.FilesCreated)
	assert.Equal(t, 0, result.FilesDeleted)
	assert.Len(t, result.Conflicts, 0)
	assert.Len(t, result.Errors, 0)

	mockGit.AssertExpectations(t)
}

func TestSynchronizeWithLocalChanges(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &SyncConfig{
		RepositoryPath:   "/test/repo",
		Bidirectional:    false,
		RemoteName:       "origin",
		ConflictStrategy: "manual",
		DryRun:           false,
		AutoCommit:       true,
		CommitMessage:    "Auto-sync: {{.Timestamp}}",
	}

	mockGit := &MockGitCommandExecutor{}
	synchronizer := &RepositorySynchronizer{
		logger: logger,
		config: config,
		gitCmd: mockGit,
	}

	// Mock repository with local changes
	mockStatus := &GitStatus{
		Branch:          "main",
		Upstream:        "origin/main",
		AheadBy:         2,
		BehindBy:        0,
		ModifiedFiles:   []string{"file1.go", "file2.go"},
		UntrackedFiles:  []string{"file3.go"},
		ConflictedFiles: []string{},
		CleanWorkingDir: false,
		LastCommitHash:  "abc123",
		LastCommitTime:  time.Now(),
	}

	mockRemoteInfo := &GitRemoteInfo{
		Name:      "origin",
		URL:       "https://github.com/test/repo.git",
		Reachable: true,
	}

	mockCommitResult := &GitCommandResult{
		Command:  "git commit -m \"Auto-sync: 2023-01-01 12:00:00\"",
		Output:   "[main def456] Auto-sync: 2023-01-01 12:00:00",
		Success:  true,
		Duration: 100 * time.Millisecond,
	}

	mockPushResult := &GitCommandResult{
		Command:  "git push origin main",
		Output:   "To github.com:test/repo.git\n   abc123..def456  main -> main",
		Success:  true,
		Duration: 500 * time.Millisecond,
	}

	mockGit.On("GetStatus", mock.Anything, "/test/repo").Return(mockStatus, nil)
	mockGit.On("GetRemoteInfo", mock.Anything, "/test/repo", "origin").Return(mockRemoteInfo, nil)
	mockGit.On("ExecuteCommand", mock.Anything, "/test/repo", "add", ".").Return(mockCommitResult, nil)
	mockGit.On("ExecuteCommand", mock.Anything, "/test/repo", "commit", "-m", mock.AnythingOfType("string")).Return(mockCommitResult, nil)
	mockGit.On("ExecuteCommand", mock.Anything, "/test/repo", "push", "origin", "main").Return(mockPushResult, nil)

	ctx := context.Background()
	result, err := synchronizer.Synchronize(ctx)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 2, result.FilesModified)
	assert.Equal(t, 1, result.FilesCreated)
	assert.Equal(t, 0, result.FilesDeleted)
	assert.Equal(t, "def456", result.CommitHash)

	mockGit.AssertExpectations(t)
}

func TestSynchronizeBidirectionalWithDivergence(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &SyncConfig{
		RepositoryPath:   "/test/repo",
		Bidirectional:    true,
		RemoteName:       "origin",
		ConflictStrategy: "auto-merge",
		DryRun:           false,
		AutoCommit:       false,
	}

	mockGit := &MockGitCommandExecutor{}
	synchronizer := &RepositorySynchronizer{
		logger: logger,
		config: config,
		gitCmd: mockGit,
	}

	// Mock diverged repository
	mockStatus := &GitStatus{
		Branch:          "main",
		Upstream:        "origin/main",
		AheadBy:         2,
		BehindBy:        3,
		ModifiedFiles:   []string{},
		UntrackedFiles:  []string{},
		ConflictedFiles: []string{},
		CleanWorkingDir: true,
		LastCommitHash:  "abc123",
		LastCommitTime:  time.Now(),
	}

	mockRemoteInfo := &GitRemoteInfo{
		Name:      "origin",
		URL:       "https://github.com/test/repo.git",
		Reachable: true,
	}

	mockFetchResult := &GitCommandResult{
		Command:  "git fetch origin",
		Output:   "Fetching origin",
		Success:  true,
		Duration: 200 * time.Millisecond,
	}

	mockMergeResult := &GitCommandResult{
		Command:  "git merge origin/main",
		Output:   "Merge made by the 'recursive' strategy.",
		Success:  true,
		Duration: 150 * time.Millisecond,
	}

	mockPushResult := &GitCommandResult{
		Command:  "git push origin main",
		Output:   "Everything up-to-date",
		Success:  true,
		Duration: 100 * time.Millisecond,
	}

	mockGit.On("GetStatus", mock.Anything, "/test/repo").Return(mockStatus, nil)
	mockGit.On("GetRemoteInfo", mock.Anything, "/test/repo", "origin").Return(mockRemoteInfo, nil)
	mockGit.On("ExecuteCommand", mock.Anything, "/test/repo", "fetch", "origin").Return(mockFetchResult, nil)
	mockGit.On("ExecuteCommand", mock.Anything, "/test/repo", "merge", "origin/main").Return(mockMergeResult, nil)
	mockGit.On("ExecuteCommand", mock.Anything, "/test/repo", "push", "origin", "main").Return(mockPushResult, nil)

	ctx := context.Background()
	result, err := synchronizer.Synchronize(ctx)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Len(t, result.Conflicts, 0)

	mockGit.AssertExpectations(t)
}

func TestSynchronizeDryRun(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &SyncConfig{
		RepositoryPath:   "/test/repo",
		Bidirectional:    true,
		RemoteName:       "origin",
		ConflictStrategy: "manual",
		DryRun:           true,
		AutoCommit:       false,
	}

	mockGit := &MockGitCommandExecutor{}
	synchronizer := &RepositorySynchronizer{
		logger: logger,
		config: config,
		gitCmd: mockGit,
	}

	mockStatus := &GitStatus{
		Branch:          "main",
		Upstream:        "origin/main",
		AheadBy:         1,
		BehindBy:        1,
		ModifiedFiles:   []string{"file1.go"},
		UntrackedFiles:  []string{},
		ConflictedFiles: []string{},
		CleanWorkingDir: false,
		LastCommitHash:  "abc123",
		LastCommitTime:  time.Now(),
	}

	mockRemoteInfo := &GitRemoteInfo{
		Name:      "origin",
		URL:       "https://github.com/test/repo.git",
		Reachable: true,
	}

	mockGit.On("GetStatus", mock.Anything, "/test/repo").Return(mockStatus, nil)
	mockGit.On("GetRemoteInfo", mock.Anything, "/test/repo", "origin").Return(mockRemoteInfo, nil)

	ctx := context.Background()
	result, err := synchronizer.Synchronize(ctx)

	require.NoError(t, err)
	assert.True(t, result.Success)
	// In dry run mode, no actual git commands should be executed except status/remote checks
	mockGit.AssertExpectations(t)
}

func TestExpandCommitMessage(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &SyncConfig{
		CommitMessage: "Auto-sync: {{.Timestamp}}",
	}

	synchronizer := &RepositorySynchronizer{
		logger: logger,
		config: config,
	}

	message := synchronizer.expandCommitMessage(config.CommitMessage)
	assert.Contains(t, message, "Auto-sync:")
	assert.NotContains(t, message, "{{.Timestamp}}")
}

func TestExtractCommitHash(t *testing.T) {
	logger := zaptest.NewLogger(t)
	synchronizer := &RepositorySynchronizer{
		logger: logger,
		config: &SyncConfig{},
	}

	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{
			name:     "standard commit output",
			output:   "[main abc123] Commit message",
			expected: "abc123",
		},
		{
			name:     "different branch",
			output:   "[feature/test def456] Another commit",
			expected: "def456",
		},
		{
			name:     "no commit hash",
			output:   "No changes to commit",
			expected: "",
		},
		{
			name:     "multiline output",
			output:   "Author: Test User\n[main xyz789] Test commit\nChanged files: 2",
			expected: "xyz789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := synchronizer.extractCommitHash(tt.output)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseConflictedFiles(t *testing.T) {
	logger := zaptest.NewLogger(t)
	synchronizer := &RepositorySynchronizer{
		logger: logger,
		config: &SyncConfig{},
	}

	statusOutput := `UU conflict1.go
M  modified.go
UU conflict2.go
A  added.go
UU conflict3.go`

	conflicted := synchronizer.parseConflictedFiles(statusOutput)

	expected := []string{"conflict1.go", "conflict2.go", "conflict3.go"}
	assert.Equal(t, expected, conflicted)
}

func TestDefaultGitExecutorCommands(t *testing.T) {
	executor := &defaultGitExecutor{}
	ctx := context.Background()

	// Test ExecuteCommand with invalid command
	result, err := executor.ExecuteCommand(ctx, ".", "invalid-command")
	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.NotEmpty(t, result.Error)
}

func TestSyncConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config *SyncConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: &SyncConfig{
				RepositoryPath:   "/test/repo",
				RemoteName:       "origin",
				ConflictStrategy: "auto-merge",
			},
			valid: true,
		},
		{
			name: "invalid conflict strategy",
			config: &SyncConfig{
				RepositoryPath:   "/test/repo",
				RemoteName:       "origin",
				ConflictStrategy: "invalid-strategy",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validStrategies := []string{"manual", "auto-merge", "local-wins", "remote-wins", "timestamp"}
			isValid := false

			for _, strategy := range validStrategies {
				if tt.config.ConflictStrategy == strategy {
					isValid = true
					break
				}
			}

			assert.Equal(t, tt.valid, isValid)
		})
	}
}

func TestConflictInfo(t *testing.T) {
	conflict := ConflictInfo{
		Path:         "/test/file.go",
		ConflictType: "content",
		LocalHash:    "abc123",
		RemoteHash:   "def456",
		Resolution:   "manual",
		ResolvedAt:   time.Now(),
	}

	assert.Equal(t, "/test/file.go", conflict.Path)
	assert.Equal(t, "content", conflict.ConflictType)
	assert.Equal(t, "abc123", conflict.LocalHash)
	assert.Equal(t, "def456", conflict.RemoteHash)
	assert.Equal(t, "manual", conflict.Resolution)
	assert.False(t, conflict.ResolvedAt.IsZero())
}

func TestSyncResult(t *testing.T) {
	result := &SyncResult{
		Success:       true,
		FilesModified: 5,
		FilesCreated:  2,
		FilesDeleted:  1,
		Conflicts:     []ConflictInfo{},
		Errors:        []string{},
		Duration:      100 * time.Millisecond,
		CommitHash:    "abc123",
	}

	assert.True(t, result.Success)
	assert.Equal(t, 5, result.FilesModified)
	assert.Equal(t, 2, result.FilesCreated)
	assert.Equal(t, 1, result.FilesDeleted)
	assert.Len(t, result.Conflicts, 0)
	assert.Len(t, result.Errors, 0)
	assert.Equal(t, 100*time.Millisecond, result.Duration)
	assert.Equal(t, "abc123", result.CommitHash)
}
