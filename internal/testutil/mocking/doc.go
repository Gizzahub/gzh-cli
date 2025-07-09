// Package mocking provides a comprehensive mocking strategy for the gzh-manager-go project.
//
// This package implements a dual-approach mocking strategy that combines the power of
// gomock for interface-based mocking with testify/mock for complex stateful scenarios.
//
// # Architecture
//
// The mocking strategy is organized into several layers:
//
// 1. **Generated Mocks (gomock)**: Automatic interface mocks for clean, type-safe testing
// 2. **Manual Mocks (testify/mock)**: Complex scenarios requiring stateful behavior
// 3. **Mock Factories**: Convenient factory methods for common mock setups
// 4. **Test Utilities**: Helper functions for mock management and verification
//
// # Generated Mocks
//
// Generated mocks are created using gomock and stored in package-specific directories:
//
//	pkg/github/mocks/          - GitHub API client mocks
//	internal/filesystem/mocks/ - File system operation mocks
//	internal/httpclient/mocks/ - HTTP client interface mocks
//	internal/git/mocks/        - Git operation interface mocks
//
// Generate all mocks using:
//
//	make generate-mocks
//
// # Mock Factories
//
// The MockFactory provides convenient methods for creating commonly used mocks
// with sensible defaults:
//
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	factory := mocking.NewMockFactory(ctrl)
//	mockClient := factory.CreateMockGitHubAPIClient()
//
// # Testify Mocks for Complex Scenarios
//
// For scenarios requiring stateful behavior, custom logic, or complex interactions,
// use the testify/mock-based implementations:
//
//	mockService := mocking.NewMockComplexGitHubService()
//	mockService.On("ProcessRepositories", ctx, repos).Return(result, nil)
//
// # Integration Example
//
//	func TestMyFeature(t *testing.T) {
//		// Setup gomock controller
//		ctrl := gomock.NewController(t)
//		defer ctrl.Finish()
//
//		// Create factory and mocks
//		factory := mocking.NewMockFactory(ctrl)
//		mockGitHub := factory.CreateMockGitHubAPIClient()
//		mockFS := factory.CreateMockFileSystem()
//
//		// Create testify mocks for complex scenarios
//		mockComplex := mocking.NewMockComplexGitHubService()
//		mockComplex.On("ProcessRepositories", mock.Anything, mock.Anything).
//			Return(&mocking.ProcessResult{ProcessedCount: 5}, nil)
//
//		// Use mocks in your test...
//		// Verify expectations
//		mockComplex.AssertExpectations(t)
//	}
//
// # Best Practices
//
// 1. **Isolation**: Use separate controllers for each test
// 2. **Clarity**: Be explicit about mock expectations
// 3. **Verification**: Always verify mock expectations
// 4. **Cleanup**: Use defer to ensure cleanup
// 5. **Factories**: Use factories for common setups
//
// # Mock Types
//
// ## Interface Mocks (gomock)
// - APIClient: GitHub API operations
// - FileSystem: File system operations
// - HTTPClient: HTTP request/response handling
// - GitClient: Git repository operations
//
// ## Complex Mocks (testify/mock)
// - MockComplexGitHubService: Stateful repository processing
// - MockStatefulFileSystem: File system with operation tracking
// - MockRateLimitedClient: Rate limiting simulation
//
// # Utilities
//
// The package provides utility functions for common testing patterns:
//
//	helpers := mocking.NewMockTestHelpers()
//	helpers.AssertMockExpectations(t, mock1, mock2, mock3)
//	helpers.SetupCommonExpectations(map[string]interface{}{
//		"service": mockService,
//		"client":  mockClient,
//	})
//
// # Performance
//
// Mock creation and usage is optimized for test performance:
// - Factory methods cache common setups
// - Lightweight mock objects
// - Efficient expectation matching
// - Minimal overhead for mock verification
//
// # Maintenance
//
// The mocking strategy is designed for long-term maintainability:
// - Automatic mock generation from interfaces
// - Clear separation between mock types
// - Comprehensive documentation
// - Example-driven API design
//
// For more examples and advanced usage patterns, see the test files in this package.
package mocking
