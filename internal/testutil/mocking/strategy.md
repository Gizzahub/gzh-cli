# Comprehensive Mocking Strategy

This document outlines the comprehensive mocking strategy for the gzh-manager-go project.

## 🎯 Overview

The mocking strategy combines **gomock** for interface mocking and **testify/mock** for struct-based mocking, providing comprehensive test coverage and isolation.

## 📁 Mock Organization

### Generated Mocks (gomock)
- `pkg/github/mocks/` - GitHub API client mocks
- `internal/filesystem/mocks/` - File system operation mocks  
- `internal/httpclient/mocks/` - HTTP client interface mocks
- `internal/git/mocks/` - Git operation interface mocks
- `pkg/config/mocks/` - Configuration service mocks

### Manual Mocks (testify/mock)
- `internal/testutil/mocks/` - Custom mocks for complex scenarios
- `internal/testutil/builders/` - Builder-based mocks (existing)

## 🔧 Mock Generation

### Automatic Generation
```bash
# Generate all mocks
make generate-mocks

# Generate specific package mocks
mockgen -source=pkg/github/interfaces.go -destination=pkg/github/mocks/github_mocks.go -package=mocks
```

### Manual Mock Implementation
Use testify/mock for complex scenarios that require custom behavior or stateful mocking.

## 📋 Interface Coverage

### GitHub Package
- ✅ `APIClient` - GitHub API operations
- ✅ `CloneService` - Repository cloning
- ✅ `TokenValidatorInterface` - Token validation

### Filesystem Package  
- ✅ `FileSystem` - File system operations
- ✅ `File` - File handle operations

### HTTP Client Package
- ✅ `HTTPClient` - HTTP request/response handling
- ✅ `RequestBuilder` - HTTP request construction

### Git Package
- ✅ `GitClient` - Git repository operations
- ✅ `RepositoryService` - Repository management

### Config Package
- ✅ `ConfigService` - Configuration management
- ✅ `ConfigLoader` - Configuration loading
- ✅ `ConfigValidator` - Configuration validation

## 🧪 Testing Patterns

### 1. Basic Interface Mocking
```go
func TestExample(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    
    mockClient := mocks.NewMockAPIClient(ctrl)
    mockClient.EXPECT().
        GetRepository(gomock.Any(), "owner", "repo").
        Return(&github.RepositoryInfo{Name: "repo"}, nil)
    
    // Test with mock
}
```

### 2. Builder-Based Mocking
```go
func TestExample(t *testing.T) {
    mockLogger := builders.NewMockLoggerBuilder().Build()
    mockEnv := builders.NewEnvironmentBuilder().
        WithGitHubToken("token").
        Build()
    
    // Test with builders
}
```

### 3. Testify Mock for Complex Scenarios
```go
type MockComplexService struct {
    mock.Mock
}

func (m *MockComplexService) ProcessData(data []byte) error {
    args := m.Called(data)
    return args.Error(0)
}

func TestComplexScenario(t *testing.T) {
    mockService := new(MockComplexService)
    mockService.On("ProcessData", mock.AnythingOfType("[]byte")).Return(nil)
    
    // Test with complex mock
    mockService.AssertExpectations(t)
}
```

## 🔄 Mock Lifecycle

### Setup
1. **Controller Creation**: `ctrl := gomock.NewController(t)`
2. **Mock Creation**: `mock := mocks.NewMockInterface(ctrl)`
3. **Expectation Setting**: `mock.EXPECT().Method().Return()`

### Execution
1. **Test Execution**: Run test with mocked dependencies
2. **Assertion**: Verify behavior and interactions

### Cleanup
1. **Controller Finish**: `defer ctrl.Finish()`
2. **Expectation Verification**: Automatic with gomock

## 📊 Mock Coverage Goals

### High Priority Interfaces
- **GitHub API Client** - Core business logic
- **File System** - I/O operations
- **HTTP Client** - Network operations
- **Git Client** - Version control operations
- **Config Service** - Configuration management

### Medium Priority Interfaces
- **Token Validator** - Security operations
- **Clone Service** - Repository management
- **Directory Resolver** - Path operations

### Low Priority Interfaces
- **Filter Service** - Data filtering
- **Schema Validator** - Structure validation

## 🛠 Utilities and Helpers

### Mock Factories
Create factory functions for common mock setups:

```go
// CreateMockGitHubClient creates a GitHub client mock with common expectations
func CreateMockGitHubClient(ctrl *gomock.Controller) *mocks.MockAPIClient {
    mock := mocks.NewMockAPIClient(ctrl)
    // Add common expectations
    return mock
}
```

### Mock Builders
Extend existing builders to support gomock integration:

```go
type MockAPIClientBuilder struct {
    ctrl *gomock.Controller
    mock *mocks.MockAPIClient
}

func NewMockAPIClientBuilder(ctrl *gomock.Controller) *MockAPIClientBuilder {
    return &MockAPIClientBuilder{
        ctrl: ctrl,
        mock: mocks.NewMockAPIClient(ctrl),
    }
}
```

## 📝 Best Practices

### 1. Mock Isolation
- Each test should have isolated mocks
- Use fresh controller per test
- Avoid shared mock state

### 2. Expectation Clarity
- Be explicit about expected calls
- Use descriptive parameter matchers
- Document complex expectations

### 3. Error Simulation
- Test both success and failure paths
- Use realistic error scenarios
- Test edge cases with mocks

### 4. Mock Maintenance
- Keep mocks in sync with interfaces
- Regenerate after interface changes
- Document breaking changes

## 🔍 Mock Verification

### Automatic Verification
- gomock automatically verifies expectations
- Use `ctrl.Finish()` for complete verification

### Manual Verification
- Use `AssertExpectations(t)` with testify/mock
- Verify call counts and parameters
- Check interaction order when relevant

## 📚 Documentation

### Interface Documentation
Each mocked interface should have:
- Purpose and responsibilities
- Key methods and their behavior
- Usage examples
- Mock generation commands

### Test Documentation
Each test using mocks should include:
- Mock setup explanation
- Expected interactions
- Verification strategy
- Error scenarios covered

---

This strategy ensures comprehensive test coverage while maintaining clean, maintainable test code.