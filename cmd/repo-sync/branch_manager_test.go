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

func TestNewBranchManager(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &BranchManagementConfig{
		RepositoryPath:    "/test/repo",
		BranchNamingRules: "gitflow",
		RemoteName:        "origin",
		StaleBranchDays:   30,
	}

	manager := NewBranchManager(logger, config)
	assert.NotNil(t, manager)
	assert.Equal(t, config, manager.config)
	assert.NotNil(t, manager.validator)
	assert.NotNil(t, manager.gitCmd)
}

func TestCreateBranchFromIssue(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &BranchManagementConfig{
		RepositoryPath:    "/test/repo",
		BranchNamingRules: "gitflow",
		RemoteName:        "origin",
		DryRun:            false,
	}

	mockGit := &MockGitCommandExecutor{}
	manager := &BranchManager{
		logger:    logger,
		config:    config,
		gitCmd:    mockGit,
		validator: NewBranchValidator(logger, "gitflow"),
	}

	tests := []struct {
		name        string
		issue       *IssueInfo
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "create feature branch",
			issue: &IssueInfo{
				ID:    "123",
				Title: "Add user authentication",
				Type:  "feature",
			},
			setupMocks: func() {
				// Check branch existence - doesn't exist
				mockGit.On("ExecuteCommand", mock.Anything, "/test/repo",
					[]string{"show-ref", "--verify", "--quiet", "refs/heads/feature/123-add-user-authentication"}).
					Return(&GitCommandResult{Success: false}, nil)

				// Get current branch
				mockGit.On("ExecuteCommand", mock.Anything, "/test/repo",
					[]string{"branch", "--show-current"}).
					Return(&GitCommandResult{Output: "main\n", Success: true}, nil)

				// Create branch
				mockGit.On("ExecuteCommand", mock.Anything, "/test/repo",
					[]string{"checkout", "-b", "feature/123-add-user-authentication", "develop"}).
					Return(&GitCommandResult{Success: true}, nil)

				// Push to remote
				mockGit.On("ExecuteCommand", mock.Anything, "/test/repo",
					[]string{"push", "-u", "origin", "feature/123-add-user-authentication"}).
					Return(&GitCommandResult{Success: true}, nil)
			},
			expectError: false,
		},
		{
			name: "branch already exists",
			issue: &IssueInfo{
				ID:    "456",
				Title: "Fix login bug",
				Type:  "bug",
			},
			setupMocks: func() {
				// Check branch existence - exists
				mockGit.On("ExecuteCommand", mock.Anything, "/test/repo",
					[]string{"show-ref", "--verify", "--quiet", "refs/heads/bugfix/456-fix-login-bug"}).
					Return(&GitCommandResult{Success: true}, nil)
			},
			expectError: true,
			errorMsg:    "branch already exists",
		},
		{
			name: "dry run mode",
			issue: &IssueInfo{
				ID:    "789",
				Title: "Critical hotfix",
				Type:  "hotfix",
			},
			setupMocks: func() {
				config.DryRun = true
				// Check branch existence
				mockGit.On("ExecuteCommand", mock.Anything, "/test/repo",
					[]string{"show-ref", "--verify", "--quiet", "refs/heads/hotfix/789-critical-hotfix"}).
					Return(&GitCommandResult{Success: false}, nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGit.ExpectedCalls = nil
			mockGit.Calls = nil
			tt.setupMocks()

			ctx := context.Background()
			err := manager.CreateBranchFromIssue(ctx, tt.issue)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			// Reset dry run
			config.DryRun = false
		})
	}
}

func TestDeleteMergedBranches(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &BranchManagementConfig{
		RepositoryPath:    "/test/repo",
		BranchNamingRules: "gitflow",
		RemoteName:        "origin",
		DryRun:            false,
	}

	mockGit := &MockGitCommandExecutor{}
	manager := &BranchManager{
		logger: logger,
		config: config,
		gitCmd: mockGit,
	}

	ctx := context.Background()

	// Mock branch listing
	mockGit.On("ExecuteCommand", ctx, "/test/repo",
		[]string{"for-each-ref", "--format=%(refname:short)|%(committerdate:iso8601)|%(authoremail)", "refs/heads/"}).
		Return(&GitCommandResult{
			Output: `feature/old-feature|2023-01-01 10:00:00 +0000|test@example.com
main|2023-12-01 10:00:00 +0000|test@example.com
develop|2023-12-01 10:00:00 +0000|test@example.com`,
			Success: true,
		}, nil)

	// Mock merged check for feature/old-feature
	mockGit.On("ExecuteCommand", ctx, "/test/repo",
		[]string{"show-ref", "--verify", "--quiet", "refs/heads/main"}).
		Return(&GitCommandResult{Success: true}, nil)
	mockGit.On("ExecuteCommand", ctx, "/test/repo",
		[]string{"branch", "--merged", "main"}).
		Return(&GitCommandResult{
			Output:  "  feature/old-feature\n* main",
			Success: true,
		}, nil)

	// Mock branch deletion
	mockGit.On("ExecuteCommand", ctx, "/test/repo",
		[]string{"branch", "-D", "feature/old-feature"}).
		Return(&GitCommandResult{Success: true}, nil)

	// Mock merged check for develop (not merged)
	mockGit.On("ExecuteCommand", ctx, "/test/repo",
		[]string{"show-ref", "--verify", "--quiet", "refs/heads/master"}).
		Return(&GitCommandResult{Success: false}, nil)
	mockGit.On("ExecuteCommand", ctx, "/test/repo",
		[]string{"show-ref", "--verify", "--quiet", "refs/heads/develop"}).
		Return(&GitCommandResult{Success: true}, nil)
	mockGit.On("ExecuteCommand", ctx, "/test/repo",
		[]string{"branch", "--merged", "develop"}).
		Return(&GitCommandResult{
			Output:  "* develop",
			Success: true,
		}, nil)

	err := manager.DeleteMergedBranches(ctx)
	assert.NoError(t, err)
	mockGit.AssertExpectations(t)
}

func TestCleanupStaleBranches(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &BranchManagementConfig{
		RepositoryPath:    "/test/repo",
		BranchNamingRules: "gitflow",
		RemoteName:        "origin",
		StaleBranchDays:   30,
		DryRun:            false,
	}

	mockGit := &MockGitCommandExecutor{}
	manager := &BranchManager{
		logger: logger,
		config: config,
		gitCmd: mockGit,
	}

	ctx := context.Background()

	// Create dates
	oldDate := time.Now().AddDate(0, 0, -60).Format("2006-01-02 15:04:05 -0700")
	recentDate := time.Now().AddDate(0, 0, -10).Format("2006-01-02 15:04:05 -0700")

	// Mock branch listing
	mockGit.On("ExecuteCommand", ctx, "/test/repo",
		[]string{"for-each-ref", "--format=%(refname:short)|%(committerdate:iso8601)|%(authoremail)", "refs/heads/"}).
		Return(&GitCommandResult{
			Output: `feature/old-feature|` + oldDate + `|test@example.com
feature/recent-feature|` + recentDate + `|test@example.com
main|` + recentDate + `|test@example.com`,
			Success: true,
		}, nil)

	// Mock merged checks (all return not merged)
	for _, branch := range []string{"main", "master", "develop"} {
		mockGit.On("ExecuteCommand", ctx, "/test/repo",
			[]string{"show-ref", "--verify", "--quiet", fmt.Sprintf("refs/heads/%s", branch)}).
			Return(&GitCommandResult{Success: branch == "main"}, nil).Maybe()

		if branch == "main" {
			mockGit.On("ExecuteCommand", ctx, "/test/repo",
				[]string{"branch", "--merged", branch}).
				Return(&GitCommandResult{
					Output:  fmt.Sprintf("* %s", branch),
					Success: true,
				}, nil).Maybe()
		}
	}

	// Mock deletion of old branch
	mockGit.On("ExecuteCommand", ctx, "/test/repo",
		[]string{"branch", "-D", "feature/old-feature"}).
		Return(&GitCommandResult{Success: true}, nil)

	err := manager.CleanupStaleBranches(ctx)
	assert.NoError(t, err)
	mockGit.AssertExpectations(t)
}

func TestDetermineBranchType(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := &BranchManager{
		logger: logger,
	}

	tests := []struct {
		name     string
		issue    *IssueInfo
		expected string
	}{
		{
			name: "bug type",
			issue: &IssueInfo{
				Type: "bug",
			},
			expected: "bugfix",
		},
		{
			name: "feature type",
			issue: &IssueInfo{
				Type: "feature",
			},
			expected: "feature",
		},
		{
			name: "hotfix type",
			issue: &IssueInfo{
				Type: "hotfix",
			},
			expected: "hotfix",
		},
		{
			name: "bug label",
			issue: &IssueInfo{
				Type:   "unknown",
				Labels: []string{"bug", "priority-high"},
			},
			expected: "bugfix",
		},
		{
			name: "feature label",
			issue: &IssueInfo{
				Type:   "unknown",
				Labels: []string{"enhancement", "ui"},
			},
			expected: "feature",
		},
		{
			name: "default to feature",
			issue: &IssueInfo{
				Type: "unknown",
			},
			expected: "feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.determineBranchType(tt.issue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCleanIssueTitle(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := &BranchManager{
		logger: logger,
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple title",
			input:    "Add user authentication",
			expected: "add-user-authentication",
		},
		{
			name:     "special characters",
			input:    "Fix bug #123 (critical!)",
			expected: "fix-bug-123-critical",
		},
		{
			name:     "multiple spaces",
			input:    "Add   multiple   spaces   handling",
			expected: "add-multiple-spaces-handling",
		},
		{
			name:     "very long title",
			input:    "This is a very long issue title that exceeds the maximum allowed length",
			expected: "this-is-a-very-long-issue-title-that",
		},
		{
			name:     "mixed case with numbers",
			input:    "JIRA-1234: Update API v2.0",
			expected: "jira-1234-update-api-v20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.cleanIssueTitle(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseIssueFromString(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedID    string
		expectedTitle string
		expectError   bool
	}{
		{
			name:          "valid format with hyphen",
			input:         "JIRA-123: Add user authentication",
			expectedID:    "JIRA-123",
			expectedTitle: "Add user authentication",
			expectError:   false,
		},
		{
			name:          "valid format simple",
			input:         "123: Fix login bug",
			expectedID:    "123",
			expectedTitle: "Fix login bug",
			expectError:   false,
		},
		{
			name:          "valid format with spaces",
			input:         "GH-456:   Multiple   spaces   in title",
			expectedID:    "GH-456",
			expectedTitle: "Multiple   spaces   in title",
			expectError:   false,
		},
		{
			name:        "invalid format - no colon",
			input:       "123 Add feature",
			expectError: true,
		},
		{
			name:        "invalid format - empty",
			input:       "",
			expectError: true,
		},
		{
			name:        "invalid format - only ID",
			input:       "123:",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue, err := ParseIssueFromString(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedID, issue.ID)
				assert.Equal(t, tt.expectedTitle, issue.Title)
				assert.Equal(t, "feature", issue.Type) // Default type
			}
		})
	}
}

func TestIsProtectedBranch(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &BranchManagementConfig{
		ProtectedBranches: []string{"release/1.0", "custom-protected"},
	}

	manager := &BranchManager{
		logger: logger,
		config: config,
	}

	tests := []struct {
		name       string
		branchName string
		expected   bool
	}{
		{"default main", "main", true},
		{"default master", "master", true},
		{"default develop", "develop", true},
		{"default staging", "staging", true},
		{"default production", "production", true},
		{"configured branch", "release/1.0", true},
		{"custom protected", "custom-protected", true},
		{"regular feature", "feature/new-feature", false},
		{"regular bugfix", "bugfix/fix-bug", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.isProtectedBranch(tt.branchName)
			assert.Equal(t, tt.expected, result)
		})
	}
}
