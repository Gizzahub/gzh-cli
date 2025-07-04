package mocking

import (
	"context"
	"testing"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestMockFactoryUsage demonstrates how to use the mock factory
func TestMockFactoryUsage(t *testing.T) {
	// Create controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create factory
	factory := NewMockFactory(ctrl)

	t.Run("GitHub API Client Mock", func(t *testing.T) {
		// Create mock with default expectations
		mockClient := factory.CreateMockGitHubAPIClient()

		// Test rate limit call
		ctx := context.Background()
		rateLimit, err := mockClient.GetRateLimit(ctx)

		require.NoError(t, err)
		assert.Equal(t, 5000, rateLimit.Limit)
		assert.Equal(t, 4999, rateLimit.Remaining)
	})

	t.Run("GitHub API Client with Repository", func(t *testing.T) {
		// Create mock with repository data
		mockClient := factory.CreateMockGitHubAPIClientWithRepo("testorg", "testrepo")

		// Test repository operations
		ctx := context.Background()
		repo, err := mockClient.GetRepository(ctx, "testorg", "testrepo")

		require.NoError(t, err)
		assert.Equal(t, "testrepo", repo.Name)
		assert.Equal(t, "testorg/testrepo", repo.FullName)
		assert.Equal(t, "main", repo.DefaultBranch)

		// Test default branch
		branch, err := mockClient.GetDefaultBranch(ctx, "testorg", "testrepo")
		require.NoError(t, err)
		assert.Equal(t, "main", branch)
	})

	t.Run("Clone Service Mock", func(t *testing.T) {
		// Create clone service mock
		mockCloneService := factory.CreateMockGitHubCloneService()

		// Test cloning operation
		ctx := context.Background()
		repoInfo := github.RepositoryInfo{
			Name:     "testrepo",
			CloneURL: "https://github.com/testorg/testrepo.git",
		}

		err := mockCloneService.CloneRepository(ctx, repoInfo, "/tmp/test", "reset")
		require.NoError(t, err)

		// Test supported strategies
		strategies := mockCloneService.GetSupportedStrategies()
		assert.Contains(t, strategies, "reset")
		assert.Contains(t, strategies, "pull")
		assert.Contains(t, strategies, "fetch")
	})

	t.Run("Token Validator Mock", func(t *testing.T) {
		// Create token validator mock
		mockValidator := factory.CreateMockTokenValidator()

		// Test token validation
		ctx := context.Background()
		tokenInfo, err := mockValidator.ValidateToken(ctx, "test-token")

		require.NoError(t, err)
		assert.True(t, tokenInfo.Valid)
		assert.Contains(t, tokenInfo.Scopes, "repo")
		assert.Equal(t, "testuser", tokenInfo.User)

		// Test operation validation
		err = mockValidator.ValidateForOperation(ctx, "test-token", "clone")
		require.NoError(t, err)

		// Test repository validation
		err = mockValidator.ValidateForRepository(ctx, "test-token", "testorg", "testrepo")
		require.NoError(t, err)
	})

	t.Run("File System Mock", func(t *testing.T) {
		// Create file system mock
		mockFS := factory.CreateMockFileSystem()

		// Test file operations
		ctx := context.Background()
		exists := mockFS.Exists(ctx, "/test/path")
		assert.True(t, exists)

		isDir := mockFS.IsDir(ctx, "/test/path")
		assert.True(t, isDir)

		err := mockFS.MkdirAll(ctx, "/test/new/path", 0o755)
		require.NoError(t, err)
	})

	t.Run("File System with Content", func(t *testing.T) {
		// Create file system mock with content
		testContent := []byte("test file content")
		mockFS := factory.CreateMockFileSystemWithContent("/test/file.txt", testContent)

		// Test reading file
		ctx := context.Background()
		content, err := mockFS.ReadFile(ctx, "/test/file.txt")

		require.NoError(t, err)
		assert.Equal(t, testContent, content)

		// Test writing file
		err = mockFS.WriteFile(ctx, "/test/file.txt", []byte("new content"), 0o644)
		require.NoError(t, err)
	})

	t.Run("HTTP Client Mock", func(t *testing.T) {
		// Create HTTP client mock
		mockHTTP := factory.CreateMockHTTPClient()

		// Test HTTP operations
		ctx := context.Background()
		resp, err := mockHTTP.Get(ctx, nil)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "200 OK", resp.Status)
	})

	t.Run("Git Client Mock", func(t *testing.T) {
		// Create git client mock
		mockGit := factory.CreateMockGitClient()

		// Test git operations
		ctx := context.Background()
		err := mockGit.Clone(ctx, "https://github.com/test/repo.git", "/tmp/test")
		require.NoError(t, err)

		err = mockGit.Pull(ctx, "/tmp/test")
		require.NoError(t, err)

		branch, err := mockGit.GetCurrentBranch(ctx, "/tmp/test")
		require.NoError(t, err)
		assert.Equal(t, "main", branch)

		isClean, err := mockGit.IsClean(ctx, "/tmp/test")
		require.NoError(t, err)
		assert.True(t, isClean)
	})
}

// TestMockFactoryBuilder demonstrates the builder pattern usage
func TestMockFactoryBuilder(t *testing.T) {
	// Create factory using builder
	factoryBuilder := NewMockFactoryBuilder(t)
	defer factoryBuilder.Finish()

	factory := factoryBuilder.Build()

	// Use the factory
	mockClient := factory.CreateMockGitHubAPIClient()

	ctx := context.Background()
	rateLimit, err := mockClient.GetRateLimit(ctx)

	require.NoError(t, err)
	assert.Equal(t, 5000, rateLimit.Limit)
}

// TestCustomMockExpectations demonstrates custom mock expectations
func TestCustomMockExpectations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	factory := NewMockFactory(ctrl)

	t.Run("Custom GitHub Client Expectations", func(t *testing.T) {
		// Create mock with custom expectations
		mockClient := factory.CreateMockGitHubAPIClient()

		// Add custom expectations
		mockClient.EXPECT().
			ListOrganizationRepositories(gomock.Any(), "testorg").
			Return([]github.RepositoryInfo{
				{Name: "repo1", FullName: "testorg/repo1"},
				{Name: "repo2", FullName: "testorg/repo2"},
			}, nil).
			Times(1)

		// Test the custom expectation
		ctx := context.Background()
		repos, err := mockClient.ListOrganizationRepositories(ctx, "testorg")

		require.NoError(t, err)
		assert.Len(t, repos, 2)
		assert.Equal(t, "repo1", repos[0].Name)
		assert.Equal(t, "repo2", repos[1].Name)
	})

	t.Run("Error Simulation", func(t *testing.T) {
		// Create mock with error expectations
		mockClient := factory.CreateMockGitHubAPIClient()

		// Simulate API error
		mockClient.EXPECT().
			GetRepository(gomock.Any(), "badorg", "badrepo").
			Return(nil, assert.AnError).
			Times(1)

		// Test error handling
		ctx := context.Background()
		repo, err := mockClient.GetRepository(ctx, "badorg", "badrepo")

		assert.Error(t, err)
		assert.Nil(t, repo)
	})
}

// TestMockLifecycle demonstrates proper mock lifecycle management
func TestMockLifecycle(t *testing.T) {
	t.Run("Controller Per Test", func(t *testing.T) {
		// Each test should have its own controller
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		factory := NewMockFactory(ctrl)
		mockClient := factory.CreateMockGitHubAPIClient()

		// Test with isolated mock
		ctx, cancel := factory.CreateContextWithTimeout(time.Second)
		defer cancel()

		_, err := mockClient.GetRateLimit(ctx)
		require.NoError(t, err)
	})

	t.Run("Builder Pattern Lifecycle", func(t *testing.T) {
		// Use builder for automatic lifecycle management
		factoryBuilder := NewMockFactoryBuilder(t)
		defer factoryBuilder.Finish()

		factory := factoryBuilder.Build()
		mockClient := factory.CreateMockGitHubAPIClient()

		// Test operations
		ctx := context.Background()
		_, err := mockClient.GetRateLimit(ctx)
		require.NoError(t, err)
	})
}

// BenchmarkMockCreation benchmarks mock creation performance
func BenchmarkMockCreation(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	factory := NewMockFactory(ctrl)

	b.Run("GitHub API Client", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = factory.CreateMockGitHubAPIClient()
		}
	})

	b.Run("File System", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = factory.CreateMockFileSystem()
		}
	})

	b.Run("HTTP Client", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = factory.CreateMockHTTPClient()
		}
	})
}
