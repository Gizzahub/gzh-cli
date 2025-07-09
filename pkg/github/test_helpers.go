package github

import (
	"context"
	"testing"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/config"
	"github.com/google/go-github/v66/github"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestFixtures provides common test data
type TestFixtures struct {
	Org              string
	Repos            []RepositoryInfo
	RepoConfigs      map[string]*config.RepoConfig
	RepositoryStates map[string]config.RepositoryState
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
		RepoConfigs:      make(map[string]*config.RepoConfig),
		RepositoryStates: make(map[string]config.RepositoryState),
	}
}

// AddRepoConfig adds a repository configuration to the fixtures
func (tf *TestFixtures) AddRepoConfig(repoName string, cfg *config.RepoConfig) {
	tf.RepoConfigs[repoName] = cfg
}

// AddRepositoryState adds a repository state to the fixtures
func (tf *TestFixtures) AddRepositoryState(repoName string, state config.RepositoryState) {
	tf.RepositoryStates[repoName] = state
}

// GetDefaultRepoConfig returns a default repository configuration for testing
func GetDefaultRepoConfig() *config.RepoConfig {
	return &config.RepoConfig{
		Version:      "1.0.0",
		Organization: "test-org",
		Templates: map[string]*config.RepoTemplate{
			"standard": {
				Description: "Standard template",
				Settings: &config.RepoSettings{
					Private:   boolPtr(true),
					HasIssues: boolPtr(true),
					HasWiki:   boolPtr(false),
				},
				Security: &config.SecuritySettings{
					VulnerabilityAlerts: boolPtr(true),
					BranchProtection: map[string]*config.BranchProtectionRule{
						"main": {
							RequiredReviews: intPtr(2),
							EnforceAdmins:   boolPtr(true),
						},
					},
				},
			},
		},
		Policies: map[string]*config.PolicyTemplate{
			"security": {
				Description: "Security policy",
				Rules: map[string]config.PolicyRule{
					"private_repos": {
						Type:        "visibility",
						Value:       "private",
						Enforcement: "required",
						Message:     "All repos must be private",
					},
					"branch_protection": {
						Type:        "branch_protection",
						Value:       true,
						Enforcement: "required",
						Message:     "Branch protection required",
					},
				},
			},
		},
	}
}

// GetTestRepositoryState creates a test repository state
func GetTestRepositoryState(name string, private bool) config.RepositoryState {
	return config.RepositoryState{
		Name:         name,
		Private:      private,
		HasIssues:    true,
		HasWiki:      false,
		HasProjects:  false,
		HasDownloads: false,
		BranchProtection: map[string]config.BranchProtectionState{
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
		LastModified:        time.Now(),
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
			Contexts: []string{"ci/build", "ci/test"},
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
		client: mockClient.client,
		org:    "test-org",
		dryRun: false,
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

// Helper functions for creating pointers
func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

// AssertRepositoryConfig asserts that two repository configurations are equal
func AssertRepositoryConfig(t *testing.T, expected, actual *config.RepoSettings) {
	if expected == nil && actual == nil {
		return
	}

	require.NotNil(t, actual, "actual settings should not be nil when expected is not nil")

	if expected.Private != nil {
		require.NotNil(t, actual.Private)
		require.Equal(t, *expected.Private, *actual.Private, "Private setting mismatch")
	}

	if expected.HasIssues != nil {
		require.NotNil(t, actual.HasIssues)
		require.Equal(t, *expected.HasIssues, *actual.HasIssues, "HasIssues setting mismatch")
	}

	if expected.HasWiki != nil {
		require.NotNil(t, actual.HasWiki)
		require.Equal(t, *expected.HasWiki, *actual.HasWiki, "HasWiki setting mismatch")
	}

	if expected.HasProjects != nil {
		require.NotNil(t, actual.HasProjects)
		require.Equal(t, *expected.HasProjects, *actual.HasProjects, "HasProjects setting mismatch")
	}
}

// AssertSecuritySettings asserts that two security settings are equal
func AssertSecuritySettings(t *testing.T, expected, actual *config.SecuritySettings) {
	if expected == nil && actual == nil {
		return
	}

	require.NotNil(t, actual, "actual security settings should not be nil when expected is not nil")

	if expected.VulnerabilityAlerts != nil {
		require.NotNil(t, actual.VulnerabilityAlerts)
		require.Equal(t, *expected.VulnerabilityAlerts, *actual.VulnerabilityAlerts, "VulnerabilityAlerts setting mismatch")
	}

	if expected.BranchProtection != nil {
		require.NotNil(t, actual.BranchProtection)
		require.Equal(t, len(expected.BranchProtection), len(actual.BranchProtection), "BranchProtection count mismatch")

		for branch, expectedRule := range expected.BranchProtection {
			actualRule, exists := actual.BranchProtection[branch]
			require.True(t, exists, "Branch protection rule for %s not found", branch)

			if expectedRule.RequiredReviews != nil {
				require.NotNil(t, actualRule.RequiredReviews)
				require.Equal(t, *expectedRule.RequiredReviews, *actualRule.RequiredReviews, "RequiredReviews mismatch for branch %s", branch)
			}

			if expectedRule.EnforceAdmins != nil {
				require.NotNil(t, actualRule.EnforceAdmins)
				require.Equal(t, *expectedRule.EnforceAdmins, *actualRule.EnforceAdmins, "EnforceAdmins mismatch for branch %s", branch)
			}
		}
	}
}
