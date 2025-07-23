// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package mocking

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gizzahub/gzh-manager-go/internal/filesystem/mocks"
	gitmocks "github.com/gizzahub/gzh-manager-go/internal/git/mocks"
	httpmocks "github.com/gizzahub/gzh-manager-go/internal/httpclient/mocks"
	"github.com/gizzahub/gzh-manager-go/pkg/github"
	githubmocks "github.com/gizzahub/gzh-manager-go/pkg/github/mocks"
	"go.uber.org/mock/gomock"
)

// MockFactory provides factory methods for creating commonly used mocks.
type MockFactory struct {
	ctrl *gomock.Controller
}

// NewMockFactory creates a new mock factory with the given controller.
func NewMockFactory(ctrl *gomock.Controller) *MockFactory {
	return &MockFactory{ctrl: ctrl}
}

// GitHub API Client Factories

// CreateMockGitHubAPIClient creates a GitHub API client mock with common expectations.
func (f *MockFactory) CreateMockGitHubAPIClient() *githubmocks.MockAPIClient {
	mock := githubmocks.NewMockAPIClient(f.ctrl)

	// Add common default expectations
	mock.EXPECT().GetRateLimit(gomock.Any()).
		Return(&github.RateLimit{
			Limit:     5000,
			Remaining: 4999,
			Reset:     time.Now().Add(time.Hour),
			Used:      1,
		}, nil).AnyTimes()

	return mock
}

// CreateMockGitHubAPIClientWithRepo creates a GitHub API client mock with repository data.
func (f *MockFactory) CreateMockGitHubAPIClientWithRepo(owner, repo string) *githubmocks.MockAPIClient {
	mock := f.CreateMockGitHubAPIClient()

	repoInfo := &github.RepositoryInfo{
		Name:          repo,
		FullName:      owner + "/" + repo,
		Description:   "Test repository",
		DefaultBranch: "main",
		CloneURL:      "https://github.com/" + owner + "/" + repo + ".git",
		SSHURL:        "git@github.com:" + owner + "/" + repo + ".git",
		HTMLURL:       "https://github.com/" + owner + "/" + repo,
		Private:       false,
		Archived:      false,
		Disabled:      false,
		CreatedAt:     time.Now().Add(-time.Hour * 24 * 30),
		UpdatedAt:     time.Now(),
		Language:      "Go",
		Size:          1024,
	}

	mock.EXPECT().GetRepository(gomock.Any(), owner, repo).
		Return(repoInfo, nil).AnyTimes()

	mock.EXPECT().GetDefaultBranch(gomock.Any(), owner, repo).
		Return("main", nil).AnyTimes()

	return mock
}

// CreateMockGitHubCloneService creates a clone service mock with success expectations.
func (f *MockFactory) CreateMockGitHubCloneService() *githubmocks.MockCloneService {
	mock := githubmocks.NewMockCloneService(f.ctrl)

	// Default to successful operations
	mock.EXPECT().CloneRepository(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	mock.EXPECT().GetSupportedStrategies().
		Return([]string{"reset", "pull", "fetch"}).AnyTimes()

	return mock
}

// CreateMockTokenValidator creates a token validator mock with valid token.
func (f *MockFactory) CreateMockTokenValidator() *githubmocks.MockTokenValidatorInterface {
	mock := githubmocks.NewMockTokenValidatorInterface(f.ctrl)

	tokenInfo := &github.TokenInfoRecord{
		Valid:       true,
		Scopes:      []string{"repo", "user", "admin:org"},
		RateLimit:   github.RateLimit{Limit: 5000, Remaining: 4999},
		User:        "testuser",
		ExpiresAt:   time.Now().Add(time.Hour * 24 * 365),
		Permissions: []string{"read", "write", "admin"},
	}

	mock.EXPECT().ValidateToken(gomock.Any(), gomock.Any()).
		Return(tokenInfo, nil).AnyTimes()

	mock.EXPECT().ValidateForOperation(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	mock.EXPECT().ValidateForRepository(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	return mock
}

// File System Factories

// CreateMockFileSystem creates a file system mock with common operations.
func (f *MockFactory) CreateMockFileSystem() *mocks.MockFileSystem {
	mock := mocks.NewMockFileSystem(f.ctrl)

	// Default successful operations
	mock.EXPECT().Exists(gomock.Any(), gomock.Any()).
		Return(true).AnyTimes()

	mock.EXPECT().IsDir(gomock.Any(), gomock.Any()).
		Return(true).AnyTimes()

	mock.EXPECT().IsFile(gomock.Any(), gomock.Any()).
		Return(true).AnyTimes()

	mock.EXPECT().MkdirAll(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	return mock
}

// CreateMockFileSystemWithContent creates a file system mock with file content expectations.
func (f *MockFactory) CreateMockFileSystemWithContent(filename string, content []byte) *mocks.MockFileSystem {
	mock := f.CreateMockFileSystem()

	mock.EXPECT().ReadFile(gomock.Any(), filename).
		Return(content, nil).AnyTimes()

	mock.EXPECT().WriteFile(gomock.Any(), filename, gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	return mock
}

// HTTP Client Factories

// CreateMockHTTPClient creates an HTTP client mock with success responses.
func (f *MockFactory) CreateMockHTTPClient() *httpmocks.MockHTTPClient {
	mock := httpmocks.NewMockHTTPClient(f.ctrl)

	// Default successful responses
	successResponse := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Header:     make(http.Header),
	}

	mock.EXPECT().Get(gomock.Any(), gomock.Any()).
		Return(successResponse, nil).AnyTimes()

	mock.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(successResponse, nil).AnyTimes()

	mock.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(successResponse, nil).AnyTimes()

	mock.EXPECT().Delete(gomock.Any(), gomock.Any()).
		Return(successResponse, nil).AnyTimes()

	return mock
}

// CreateMockHTTPClientWithResponse creates HTTP client mock with specific response.
func (f *MockFactory) CreateMockHTTPClientWithResponse(statusCode int, _ string) *httpmocks.MockHTTPClient {
	mock := httpmocks.NewMockHTTPClient(f.ctrl)

	response := &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Header:     make(http.Header),
	}

	mock.EXPECT().Get(gomock.Any(), gomock.Any()).
		Return(response, nil).AnyTimes()

	return mock
}

// Git Client Factories

// CreateMockGitClient creates a git client mock with common operations.
func (f *MockFactory) CreateMockGitClient() *gitmocks.MockGitClient {
	mock := gitmocks.NewMockGitClient(f.ctrl)

	// Default successful operations
	mock.EXPECT().Clone(gomock.Any(), gomock.Any()).
		Return(nil, nil).AnyTimes()

	mock.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil).AnyTimes()

	mock.EXPECT().Fetch(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil).AnyTimes()

	mock.EXPECT().GetCurrentBranch(gomock.Any(), gomock.Any()).
		Return("main", nil).AnyTimes()

	mock.EXPECT().IsRepository(gomock.Any(), gomock.Any()).
		Return(true).AnyTimes()

	return mock
}

// Note: MockRepositoryService no longer exists in the updated interface

// Utility Methods

// CreateContextWithTimeout creates a context with timeout for testing.
func (f *MockFactory) CreateContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// CreateCancelableContext creates a cancelable context for testing.
func (f *MockFactory) CreateCancelableContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

// MockFactoryBuilder provides a fluent interface for building mock factories.
type MockFactoryBuilder struct {
	ctrl     *gomock.Controller
	t        *testing.T
	finished bool
}

// NewMockFactoryBuilder creates a new mock factory builder.
func NewMockFactoryBuilder(t *testing.T) *MockFactoryBuilder {
	t.Helper()
	return &MockFactoryBuilder{
		t:        t,
		finished: false,
	}
}

// WithController sets a custom controller.
func (b *MockFactoryBuilder) WithController(ctrl *gomock.Controller) *MockFactoryBuilder {
	b.ctrl = ctrl
	return b
}

// Build creates the mock factory.
func (b *MockFactoryBuilder) Build() *MockFactory {
	if b.ctrl == nil {
		b.ctrl = gomock.NewController(b.t)
	}

	return NewMockFactory(b.ctrl)
}

// Finish cleans up the controller.
func (b *MockFactoryBuilder) Finish() {
	if b.ctrl != nil && !b.finished {
		b.ctrl.Finish()
		b.finished = true
	}
}
