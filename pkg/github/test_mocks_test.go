//nolint:testpackage // White-box testing needed for internal function access
package github

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type mockAPIClient struct {
	mock.Mock
}

func (m *mockAPIClient) GetRepository(ctx context.Context, owner, repo string) (*RepositoryInfo, error) {
	args := m.Called(ctx, owner, repo)
	if repo, ok := args.Get(0).(*RepositoryInfo); ok {
		return repo, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAPIClient) ListOrganizationRepositories(ctx context.Context, org string) ([]RepositoryInfo, error) {
	args := m.Called(ctx, org)
	if repos, ok := args.Get(0).([]RepositoryInfo); ok {
		return repos, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAPIClient) GetDefaultBranch(ctx context.Context, owner, repo string) (string, error) {
	args := m.Called(ctx, owner, repo)
	return args.String(0), args.Error(1)
}

func (m *mockAPIClient) SetToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *mockAPIClient) GetRateLimit(ctx context.Context) (*RateLimit, error) {
	args := m.Called(ctx)
	if rateLimit, ok := args.Get(0).(*RateLimit); ok {
		return rateLimit, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAPIClient) GetRepositoryConfiguration(ctx context.Context, owner, repo string) (*RepositoryConfig, error) {
	args := m.Called(ctx, owner, repo)
	if config, ok := args.Get(0).(*RepositoryConfig); ok {
		return config, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAPIClient) UpdateRepositoryConfiguration(ctx context.Context, owner, repo string, config *RepositoryConfig) error {
	args := m.Called(ctx, owner, repo, config)
	return args.Error(0)
}

func (m *mockAPIClient) CreateRepository(ctx context.Context, owner string, opts *CreateRepositoryOptions) (*RepositoryInfo, error) {
	args := m.Called(ctx, owner, opts)
	if info, ok := args.Get(0).(*RepositoryInfo); ok {
		return info, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAPIClient) UpdateRepository(ctx context.Context, owner, repo string, opts *UpdateRepositoryOptions) (*RepositoryInfo, error) {
	args := m.Called(ctx, owner, repo, opts)
	if info, ok := args.Get(0).(*RepositoryInfo); ok {
		return info, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAPIClient) ForkRepository(ctx context.Context, owner, repo string, opts *ForkRepositoryOptions) (*RepositoryInfo, error) {
	args := m.Called(ctx, owner, repo, opts)
	if info, ok := args.Get(0).(*RepositoryInfo); ok {
		return info, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAPIClient) DeleteRepository(ctx context.Context, owner, repo string) error {
	args := m.Called(ctx, owner, repo)
	return args.Error(0)
}

func (m *mockAPIClient) ArchiveRepository(ctx context.Context, owner, repo string) error {
	args := m.Called(ctx, owner, repo)
	return args.Error(0)
}

func (m *mockAPIClient) UnarchiveRepository(ctx context.Context, owner, repo string) error {
	args := m.Called(ctx, owner, repo)
	return args.Error(0)
}

func (m *mockAPIClient) SearchRepositories(ctx context.Context, query string, opts *SearchRepositoriesOptions) (*RepositorySearchResult, error) {
	args := m.Called(ctx, query, opts)
	if result, ok := args.Get(0).(*RepositorySearchResult); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
}

// mockLogger discards logs. Tests do not assert log calls.
type mockLogger struct{}

func (m *mockLogger) Debug(_ string, _ ...any) {}
func (m *mockLogger) Info(_ string, _ ...any)  {}
func (m *mockLogger) Warn(_ string, _ ...any)  {}
func (m *mockLogger) Error(_ string, _ ...any) {}
func (m *mockLogger) Fatal(_ string, _ ...any) {}

func boolPtr(v bool) *bool { return &v }
