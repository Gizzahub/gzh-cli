package testutil

import (
	"testing"

	"github.com/gizzahub/gzh-manager-go/internal/testutil/builders"
	"github.com/gizzahub/gzh-manager-go/internal/testutil/fixtures"
	"github.com/gizzahub/gzh-manager-go/pkg/config"
)

// ExampleBuildersUsage demonstrates how to use the builders package
func ExampleBuildersUsage() {
	// Create a configuration with builders
	config := config.NewConfigBuilder().
		WithVersion("1.0.0").
		WithDefaultProvider("github").
		WithGitHubProvider("${GITHUB_TOKEN}").
		WithOrganization("github", "test-org", "~/repos/test").
		Build()

	// Create a mock environment
	env := builders.NewEnvironmentBuilder().
		WithGitHubToken("test-token").
		WithHome("/home/user").
		Build()

	// Create a mock logger
	logger := builders.NewMockLoggerBuilder().Build()

	// Create a GitHub bulk clone request
	request := builders.NewBulkCloneRequestBuilder().
		WithOrganization("test-org").
		WithTargetPath("/tmp/test").
		WithRepository("repo1").
		WithRepository("repo2").
		Build()

	_ = config
	_ = env
	_ = logger
	_ = request
}

// ExampleFixturesUsage demonstrates how to use the fixtures package
func ExampleFixturesUsage() {
	// Use configuration fixtures
	configFixtures := fixtures.NewConfigFixtures()
	simpleConfig := configFixtures.SimpleGitHubConfig()
	multiProviderConfig := configFixtures.MultiProviderConfig()

	// Use YAML fixtures
	yamlFixtures := fixtures.NewConfigYAMLFixtures()
	yaml := yamlFixtures.SimpleGitHubYAML()

	// Use GitHub fixtures
	githubFixtures := fixtures.NewGitHubFixtures()
	request := githubFixtures.SimpleBulkCloneRequest()
	result := githubFixtures.SuccessfulBulkCloneResult()

	_ = simpleConfig
	_ = multiProviderConfig
	_ = yaml
	_ = request
	_ = result
}

// TestBuildersIntegration demonstrates how builders work together
func TestBuildersIntegration(t *testing.T) {
	// Create test environment
	env := builders.NewEnvironmentBuilder().
		WithGitHubToken("test-token-123").
		WithHome("/home/testuser").
		Build()

	// Create test configuration
	config := config.NewConfigBuilder().
		WithVersion("1.0.0").
		WithDefaultProvider("github").
		WithGitHubProvider("${GITHUB_TOKEN}").
		WithOrganization("github", "test-org", "~/repos/test").
		Build()

	// Create mock logger to track calls
	logger := builders.NewMockLoggerBuilder().Build()

	// Test that environment expansion works
	expandedHome := env.Expand("${HOME}/repos")
	if expandedHome != "/home/testuser/repos" {
		t.Errorf("Expected '/home/testuser/repos', got '%s'", expandedHome)
	}

	// Test that configuration is properly structured
	if config.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", config.Version)
	}

	if config.DefaultProvider != "github" {
		t.Errorf("Expected default provider 'github', got '%s'", config.DefaultProvider)
	}

	if len(config.Providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(config.Providers))
	}

	// Test logger call tracking
	logger.Info("Test message", "key", "value")
	logger.Error("Error message")

	if len(logger.InfoCalls) != 1 {
		t.Errorf("Expected 1 info call, got %d", len(logger.InfoCalls))
	}

	if len(logger.ErrorCalls) != 1 {
		t.Errorf("Expected 1 error call, got %d", len(logger.ErrorCalls))
	}

	if logger.InfoCalls[0].Message != "Test message" {
		t.Errorf("Expected 'Test message', got '%s'", logger.InfoCalls[0].Message)
	}
}

// TestFixturesIntegration demonstrates how fixtures work together
func TestFixturesIntegration(t *testing.T) {
	// Use configuration fixtures
	configFixtures := fixtures.NewConfigFixtures()
	config := configFixtures.SimpleGitHubConfig()

	// Use GitHub fixtures
	githubFixtures := fixtures.NewGitHubFixtures()
	request := githubFixtures.SimpleBulkCloneRequest()
	result := githubFixtures.SuccessfulBulkCloneResult()

	// Test configuration structure
	if config.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", config.Version)
	}

	if len(config.Providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(config.Providers))
	}

	// Test bulk clone request structure
	if request.Organization != "test-org" {
		t.Errorf("Expected organization 'test-org', got '%s'", request.Organization)
	}

	if len(request.Repositories) != 2 {
		t.Errorf("Expected 2 repositories, got %d", len(request.Repositories))
	}

	// Test bulk clone result structure
	if result.TotalRepositories != 3 {
		t.Errorf("Expected 3 total repositories, got %d", result.TotalRepositories)
	}

	if result.SuccessfulOperations != 3 {
		t.Errorf("Expected 3 successful operations, got %d", result.SuccessfulOperations)
	}

	if result.FailedOperations != 0 {
		t.Errorf("Expected 0 failed operations, got %d", result.FailedOperations)
	}
}
