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

// TestComprehensiveMockingStrategy demonstrates a complete mocking strategy
// integrating gomock, testify/mock, and custom builders
func TestComprehensiveMockingStrategy(t *testing.T) {
	// Setup test context
	ctx := context.Background()
	timeout := time.Second * 30

	t.Run("Full Integration Test", func(t *testing.T) {
		// Create gomock controller
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mock factory
		factory := NewMockFactory(ctrl)

		// Create gomock-based mocks
		mockGitHubClient := factory.CreateMockGitHubAPIClientWithRepo("testorg", "testrepo")
		mockFileSystem := factory.CreateMockFileSystem()
		mockHTTPClient := factory.CreateMockHTTPClient()

		// Create testify-based mocks for complex scenarios
		mockComplexService := NewMockComplexGitHubService()
		mockStatefulFS := NewMockStatefulFileSystem()
		mockRateLimitedClient := NewMockRateLimitedClient(5000)

		// Setup testify mock expectations
		repos := []github.RepositoryInfo{
			{Name: "repo1", FullName: "testorg/repo1"},
			{Name: "repo2", FullName: "testorg/repo2"},
		}

		expectedResult := &ProcessResult{
			ProcessedCount: 2,
			SkippedCount:   0,
			ErrorCount:     0,
			ProcessingTime: time.Millisecond * 200,
			State:          1,
		}

		mockComplexService.On("ProcessRepositories", ctx, repos).Return(expectedResult, nil)
		mockComplexService.On("BulkCloneWithCallback", ctx, repos, nil).Return(nil)

		// Setup stateful file system mock
		mockStatefulFS.On("WriteFile", ctx, "/tmp/test.txt", []byte("test"), 0o644).Return(nil)
		mockStatefulFS.On("ReadFile", ctx, "/tmp/test.txt").Return([]byte("test"), nil)
		mockStatefulFS.On("Exists", ctx, "/tmp/test.txt").Return(true)

		// Simulate a complex workflow
		testCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// Step 1: Get repository information using gomock
		repo, err := mockGitHubClient.GetRepository(testCtx, "testorg", "testrepo")
		require.NoError(t, err)
		assert.Equal(t, "testrepo", repo.Name)

		// Step 2: Process repositories using testify mock
		result, err := mockComplexService.ProcessRepositories(testCtx, repos)
		require.NoError(t, err)
		assert.Equal(t, 2, result.ProcessedCount)
		assert.Equal(t, 1, result.State)

		// Step 3: File operations using stateful mock
		err = mockStatefulFS.WriteFile(testCtx, "/tmp/test.txt", []byte("test"), 0o644)
		require.NoError(t, err)

		content, err := mockStatefulFS.ReadFile(testCtx, "/tmp/test.txt")
		require.NoError(t, err)
		assert.Equal(t, []byte("test"), content)

		exists := mockStatefulFS.Exists(testCtx, "/tmp/test.txt")
		assert.True(t, exists)

		// Step 4: Bulk clone with progress callback
		var progressUpdates []ProgressUpdate
		progressCallback := func(update ProgressUpdate) {
			progressUpdates = append(progressUpdates, update)
		}

		err = mockComplexService.BulkCloneWithCallback(testCtx, repos, progressCallback)
		require.NoError(t, err)
		assert.Len(t, progressUpdates, 2)

		// Step 5: Rate limited API calls
		response1, err := mockRateLimitedClient.MakeRequest(testCtx, "/api/repos")
		require.NoError(t, err)
		assert.Equal(t, 200, response1.Status)
		assert.Equal(t, 4999, mockRateLimitedClient.GetRateLimit().Remaining)

		// Verify all expectations
		mockComplexService.AssertExpectations(t)
		mockStatefulFS.AssertExpectations(t)

		// Verify call logs and state
		assert.Contains(t, mockComplexService.GetProcessingHistory(), "ProcessRepositories")
		assert.Contains(t, mockComplexService.GetProcessingHistory(), "BulkCloneWithCallback")
		assert.Equal(t, 1, mockComplexService.GetCurrentState())

		assert.Equal(t, 3, mockStatefulFS.GetOperationCount())
		assert.Equal(t, 1, mockStatefulFS.GetFileCount())

		accessLog := mockStatefulFS.GetAccessLog()
		assert.Len(t, accessLog, 3)
		assert.Equal(t, "write", accessLog[0].Operation)
		assert.Equal(t, "read", accessLog[1].Operation)
		assert.Equal(t, "exists", accessLog[2].Operation)
	})

	t.Run("Error Handling and Edge Cases", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		factory := NewMockFactory(ctrl)

		// Test error scenarios with gomock
		mockGitHubClient := factory.CreateMockGitHubAPIClient()
		mockGitHubClient.EXPECT().
			GetRepository(ctx, "badorg", "badrepo").
			Return(nil, assert.AnError).
			Times(1)

		// Test rate limiting with testify mock
		mockRateLimitedClient := NewMockRateLimitedClient(1) // Very low limit

		// First request should succeed
		response1, err := mockRateLimitedClient.MakeRequest(ctx, "/api/test")
		require.NoError(t, err)
		assert.Equal(t, 200, response1.Status)
		assert.Equal(t, 0, mockRateLimitedClient.GetRateLimit().Remaining)

		// Second request should fail due to rate limit
		response2, err := mockRateLimitedClient.MakeRequest(ctx, "/api/test")
		assert.Error(t, err)
		assert.Nil(t, response2)
		assert.True(t, IsRateLimitError(err))

		// Test error from GitHub client
		repo, err := mockGitHubClient.GetRepository(ctx, "badorg", "badrepo")
		assert.Error(t, err)
		assert.Nil(t, repo)
	})

	t.Run("Performance and Concurrency", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		factory := NewMockFactory(ctrl)
		mockGitHubClient := factory.CreateMockGitHubAPIClient()

		// Test concurrent access to mocks
		concurrency := 10
		done := make(chan bool, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(id int) {
				defer func() { done <- true }()

				// Each goroutine makes API calls
				rateLimit, err := mockGitHubClient.GetRateLimit(ctx)
				assert.NoError(t, err)
				assert.Equal(t, 5000, rateLimit.Limit)
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < concurrency; i++ {
			select {
			case <-done:
				// Success
			case <-time.After(time.Second * 5):
				t.Fatal("Timeout waiting for concurrent operations")
			}
		}
	})

	t.Run("Mock State Management", func(t *testing.T) {
		// Test state management across multiple operations
		mockStatefulFS := NewMockStatefulFileSystem()

		// Setup expectations
		mockStatefulFS.On("WriteFile", ctx, "/file1.txt", []byte("content1"), 0o644).Return(nil)
		mockStatefulFS.On("WriteFile", ctx, "/file2.txt", []byte("content2"), 0o644).Return(nil)
		mockStatefulFS.On("ReadFile", ctx, "/file1.txt").Return([]byte("content1"), nil)
		mockStatefulFS.On("ReadFile", ctx, "/file2.txt").Return([]byte("content2"), nil)
		mockStatefulFS.On("Exists", ctx, "/file1.txt").Return(true)
		mockStatefulFS.On("Exists", ctx, "/file2.txt").Return(true)

		// Perform operations
		err := mockStatefulFS.WriteFile(ctx, "/file1.txt", []byte("content1"), 0o644)
		require.NoError(t, err)

		err = mockStatefulFS.WriteFile(ctx, "/file2.txt", []byte("content2"), 0o644)
		require.NoError(t, err)

		// Verify state
		assert.Equal(t, 2, mockStatefulFS.GetFileCount())
		assert.Equal(t, 2, mockStatefulFS.GetOperationCount())

		// Read files back
		content1, err := mockStatefulFS.ReadFile(ctx, "/file1.txt")
		require.NoError(t, err)
		assert.Equal(t, []byte("content1"), content1)

		content2, err := mockStatefulFS.ReadFile(ctx, "/file2.txt")
		require.NoError(t, err)
		assert.Equal(t, []byte("content2"), content2)

		// Verify final state
		assert.Equal(t, 4, mockStatefulFS.GetOperationCount())

		exists1 := mockStatefulFS.Exists(ctx, "/file1.txt")
		exists2 := mockStatefulFS.Exists(ctx, "/file2.txt")
		assert.True(t, exists1)
		assert.True(t, exists2)

		// Verify access log
		accessLog := mockStatefulFS.GetAccessLog()
		assert.Len(t, accessLog, 6) // 2 writes + 2 reads + 2 exists

		// Verify expectations
		mockStatefulFS.AssertExpectations(t)

		// Reset and verify clean state
		mockStatefulFS.ResetState()
		assert.Equal(t, 0, mockStatefulFS.GetOperationCount())
		assert.Equal(t, 0, mockStatefulFS.GetFileCount())
		assert.Len(t, mockStatefulFS.GetAccessLog(), 0)
	})
}

// TestMockingBestPractices demonstrates best practices for mock usage
func TestMockingBestPractices(t *testing.T) {
	t.Run("Mock Isolation", func(t *testing.T) {
		// Each test should have its own mocks
		ctrl1 := gomock.NewController(t)
		defer ctrl1.Finish()

		ctrl2 := gomock.NewController(t)
		defer ctrl2.Finish()

		factory1 := NewMockFactory(ctrl1)
		factory2 := NewMockFactory(ctrl2)

		mock1 := factory1.CreateMockGitHubAPIClient()
		mock2 := factory2.CreateMockGitHubAPIClient()

		// Mocks should be independent
		ctx := context.Background()

		rateLimit1, err1 := mock1.GetRateLimit(ctx)
		rateLimit2, err2 := mock2.GetRateLimit(ctx)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.Equal(t, 5000, rateLimit1.Limit)
		assert.Equal(t, 5000, rateLimit2.Limit)
	})

	t.Run("Expectation Clarity", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		factory := NewMockFactory(ctrl)
		mockClient := factory.CreateMockGitHubAPIClient()

		// Be explicit about expectations
		mockClient.EXPECT().
			ListOrganizationRepositories(
				gomock.Any(),              // context
				gomock.Eq("specific-org"), // specific organization
			).
			Return([]github.RepositoryInfo{
				{Name: "repo1", FullName: "specific-org/repo1"},
				{Name: "repo2", FullName: "specific-org/repo2"},
			}, nil).
			Times(1) // Expect exactly one call

		// Test with the specific expectation
		ctx := context.Background()
		repos, err := mockClient.ListOrganizationRepositories(ctx, "specific-org")

		require.NoError(t, err)
		assert.Len(t, repos, 2)
		assert.Equal(t, "repo1", repos[0].Name)
		assert.Equal(t, "repo2", repos[1].Name)
	})

	t.Run("Mock Helpers Usage", func(t *testing.T) {
		// Use mock helpers for complex scenarios
		helpers := NewMockTestHelpers()

		// Create multiple mocks
		mockComplex := NewMockComplexGitHubService()
		mockStateful := NewMockStatefulFileSystem()

		// Setup common expectations using helpers
		mocks := map[string]interface{}{
			"complex":  mockComplex,
			"stateful": mockStateful,
		}
		helpers.SetupCommonExpectations(mocks)

		// Use the mocks (expectations are already set up)
		ctx := context.Background()
		exists := mockStateful.Exists(ctx, "/test/path")
		assert.True(t, exists)

		// Verify all expectations at once
		helpers.AssertMockExpectations(t, mockComplex, mockStateful)
	})
}

// BenchmarkMockingPerformance benchmarks different mocking approaches
func BenchmarkMockingPerformance(b *testing.B) {
	b.Run("Gomock Creation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ctrl := gomock.NewController(b)
			factory := NewMockFactory(ctrl)
			_ = factory.CreateMockGitHubAPIClient()
			ctrl.Finish()
		}
	})

	b.Run("Testify Mock Creation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewMockComplexGitHubService()
		}
	})

	b.Run("Mock Factory Usage", func(b *testing.B) {
		ctrl := gomock.NewController(b)
		defer ctrl.Finish()

		factory := NewMockFactory(ctrl)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mock := factory.CreateMockGitHubAPIClient()
			_, _ = mock.GetRateLimit(context.Background())
		}
	})

	b.Run("Stateful Mock Operations", func(b *testing.B) {
		mockFS := NewMockStatefulFileSystem()
		ctx := context.Background()

		// Setup expectations
		mockFS.On("WriteFile", ctx, "/test.txt", []byte("data"), 0o644).Return(nil)
		mockFS.On("ReadFile", ctx, "/test.txt").Return([]byte("data"), nil)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = mockFS.WriteFile(ctx, "/test.txt", []byte("data"), 0o644)
			_, _ = mockFS.ReadFile(ctx, "/test.txt")
		}
	})
}
