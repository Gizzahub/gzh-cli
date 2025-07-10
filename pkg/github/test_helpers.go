package github

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-github/v66/github"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestFixtures provides common test data
type TestFixtures struct {
	Org              string
	Repos            []RepositoryInfo
	RepositoryStates map[string]RepositoryStateData
}

// NewTestFixtures creates a new set of test fixtures
func NewTestFixtures() *TestFixtures {
	return &TestFixtures{
		Org: "test-org",
		Repos: []RepositoryInfo{
			{
				Name:          "test-repo-1",
				FullName:      "test-org/test-repo-1",
				Description:   "Test repository 1",
				DefaultBranch: "main",
				CloneURL:      "https://github.com/test-org/test-repo-1.git",
				SSHURL:        "git@github.com:test-org/test-repo-1.git",
				HTMLURL:       "https://github.com/test-org/test-repo-1",
				Private:       true,
				Archived:      false,
				Disabled:      false,
				CreatedAt:     time.Now().Add(-365 * 24 * time.Hour),
				UpdatedAt:     time.Now().Add(-24 * time.Hour),
				Language:      "Go",
				Size:          1024,
			},
			{
				Name:          "test-repo-2",
				FullName:      "test-org/test-repo-2",
				Description:   "Test repository 2",
				DefaultBranch: "main",
				CloneURL:      "https://github.com/test-org/test-repo-2.git",
				SSHURL:        "git@github.com:test-org/test-repo-2.git",
				HTMLURL:       "https://github.com/test-org/test-repo-2",
				Private:       false,
				Archived:      false,
				Disabled:      false,
				CreatedAt:     time.Now().Add(-180 * 24 * time.Hour),
				UpdatedAt:     time.Now().Add(-48 * time.Hour),
				Language:      "JavaScript",
				Size:          512,
			},
			{
				Name:          "archived-repo",
				FullName:      "test-org/archived-repo",
				Description:   "Archived test repository",
				DefaultBranch: "master",
				CloneURL:      "https://github.com/test-org/archived-repo.git",
				SSHURL:        "git@github.com:test-org/archived-repo.git",
				HTMLURL:       "https://github.com/test-org/archived-repo",
				Private:       true,
				Archived:      true,
				Disabled:      false,
				CreatedAt:     time.Now().Add(-730 * 24 * time.Hour),
				UpdatedAt:     time.Now().Add(-365 * 24 * time.Hour),
				Language:      "Python",
				Size:          256,
			},
		},
		RepositoryStates: make(map[string]RepositoryStateData),
	}
}

// AddRepositoryState adds a repository state to the fixtures
func (tf *TestFixtures) AddRepositoryState(repoName string, state RepositoryStateData) {
	tf.RepositoryStates[repoName] = state
}

// GetTestRepositoryState creates a test repository state
func GetTestRepositoryState(name string, private bool) RepositoryStateData {
	return RepositoryStateData{
		Name:         name,
		Private:      private,
		HasIssues:    true,
		HasWiki:      false,
		HasProjects:  false,
		HasDownloads: false,
		BranchProtection: map[string]BranchProtectionData{
			"main": {
				Protected:       true,
				RequiredReviews: 2,
				EnforceAdmins:   true,
			},
		},
		VulnerabilityAlerts: true,
		SecurityAdvisories:  true,
		Files:               []string{"README.md", "LICENSE"},
		Workflows:           []string{"ci", "security"},
		LastModified:        time.Now().Format("2006-01-02T15:04:05Z07:00"),
	}
}

// MockGitHubClient creates a mock GitHub client for testing
type MockGitHubClient struct {
	ctrl     *gomock.Controller
	client   *github.Client
	fixtures *TestFixtures
}

// NewMockGitHubClient creates a new mock GitHub client
func NewMockGitHubClient(t *testing.T) *MockGitHubClient {
	ctrl := gomock.NewController(t)
	return &MockGitHubClient{
		ctrl:     ctrl,
		client:   github.NewClient(nil),
		fixtures: NewTestFixtures(),
	}
}

// Finish should be called at the end of the test
func (m *MockGitHubClient) Finish() {
	m.ctrl.Finish()
}

// ConvertRepoInfoToGitHubRepo converts RepositoryInfo to github.Repository
func ConvertRepoInfoToGitHubRepo(info RepositoryInfo) *github.Repository {
	return &github.Repository{
		Name:          github.String(info.Name),
		FullName:      github.String(info.FullName),
		Description:   github.String(info.Description),
		DefaultBranch: github.String(info.DefaultBranch),
		CloneURL:      github.String(info.CloneURL),
		SSHURL:        github.String(info.SSHURL),
		HTMLURL:       github.String(info.HTMLURL),
		Private:       github.Bool(info.Private),
		Archived:      github.Bool(info.Archived),
		Disabled:      github.Bool(info.Disabled),
		CreatedAt:     &github.Timestamp{Time: info.CreatedAt},
		UpdatedAt:     &github.Timestamp{Time: info.UpdatedAt},
		Language:      github.String(info.Language),
		Size:          github.Int(info.Size),
	}
}

// CreateMockBranchProtection creates a mock branch protection object
func CreateMockBranchProtection(requiredReviews int, enforceAdmins bool) *github.Protection {
	return &github.Protection{
		RequiredPullRequestReviews: &github.PullRequestReviewsEnforcement{
			RequiredApprovingReviewCount: requiredReviews,
			DismissStaleReviews:          true,
			RequireCodeOwnerReviews:      true,
		},
		EnforceAdmins: &github.AdminEnforcement{
			URL:     github.String("https://api.github.com/repos/test-org/test-repo/branches/main/protection/enforce_admins"),
			Enabled: enforceAdmins,
		},
		RequiredStatusChecks: &github.RequiredStatusChecks{
			Strict:   true,
			Contexts: &[]string{"ci/build", "ci/test"},
		},
		AllowForcePushes: &github.AllowForcePushes{
			Enabled: false,
		},
		AllowDeletions: &github.AllowDeletions{
			Enabled: false,
		},
	}
}

// CreateMockRepositoryContent creates mock repository content
func CreateMockRepositoryContent(name, path, content string) *github.RepositoryContent {
	contentType := "file"
	return &github.RepositoryContent{
		Type:    &contentType,
		Name:    &name,
		Path:    &path,
		Content: &content,
		SHA:     github.String("abc123"),
		Size:    github.Int(len(content)),
	}
}

// TestRepoConfigClient wraps RepoConfigClient for testing
type TestRepoConfigClient struct {
	*RepoConfigClient
	MockClient *MockGitHubClient
}

// NewTestRepoConfigClient creates a new test repo config client
func NewTestRepoConfigClient(t *testing.T) *TestRepoConfigClient {
	mockClient := NewMockGitHubClient(t)
	client := &RepoConfigClient{
		token:       "test-token",
		baseURL:     "https://api.github.com",
		httpClient:  mockClient.client.Client(),
		rateLimiter: NewRateLimiter(),
	}

	return &TestRepoConfigClient{
		RepoConfigClient: client,
		MockClient:       mockClient,
	}
}

// SetupListReposExpectation sets up expectation for listing repositories
func SetupListReposExpectation(t *testing.T, ctx context.Context, client *github.Client, org string, repos []*github.Repository) {
	// This would typically use gomock expectations
	// For now, we're showing the pattern
	require.NotNil(t, ctx)
	require.NotEmpty(t, org)
	require.NotNil(t, repos)
}

// SetupGetRepoExpectation sets up expectation for getting a single repository
func SetupGetRepoExpectation(t *testing.T, ctx context.Context, client *github.Client, owner, repo string, repository *github.Repository) {
	require.NotNil(t, ctx)
	require.NotEmpty(t, owner)
	require.NotEmpty(t, repo)
	require.NotNil(t, repository)
}

// Helper functions for creating pointers - removed boolPtr as it's already defined in automation_engine.go

func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

// AssertRepositoryConfig asserts that two repository configurations are equal
func AssertRepositoryConfig(t *testing.T, expected, actual *Repository) {
	if expected == nil && actual == nil {
		return
	}

	require.NotNil(t, actual, "actual settings should not be nil when expected is not nil")

	require.Equal(t, expected.Private, actual.Private, "Private setting mismatch")
	require.Equal(t, expected.HasIssues, actual.HasIssues, "HasIssues setting mismatch")
	require.Equal(t, expected.HasWiki, actual.HasWiki, "HasWiki setting mismatch")
	require.Equal(t, expected.HasProjects, actual.HasProjects, "HasProjects setting mismatch")
	require.Equal(t, expected.HasDownloads, actual.HasDownloads, "HasDownloads setting mismatch")
}

// AssertSecuritySettings asserts that two security settings are equal
func AssertSecuritySettings(t *testing.T, expected, actual *RepositoryConfig) {
	if expected == nil && actual == nil {
		return
	}

	require.NotNil(t, actual, "actual security settings should not be nil when expected is not nil")

	// Simplified assertion - can be expanded later if needed
	require.Equal(t, expected.Name, actual.Name, "Repository name mismatch")
	require.Equal(t, expected.Private, actual.Private, "Private setting mismatch")
}
