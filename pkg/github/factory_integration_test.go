//nolint:testpackage // White-box testing needed for internal function access
package github

import (
	"context"
	"testing"

	"github.com/Gizzahub/gzh-cli/internal/env"
)

// TestFactoryPatternIntegration demonstrates the complete factory pattern working.
func TestFactoryPatternIntegration(t *testing.T) {
	// Create a mock environment for testing
	mockEnv := env.NewMockEnvironment(map[string]string{
		env.CommonEnvironmentKeys.GitHubToken: "test-github-token",
		env.CommonEnvironmentKeys.HomeDir:     "/test/home",
	})

	// Create the GitHub provider factory
	factory := NewGitHubProviderFactory(mockEnv)

	// Test that we can create different types of instances
	ctx := context.Background()

	// Create a cloner with explicit token
	cloner1, err := factory.CreateCloner(ctx, "explicit-token")
	if err != nil {
		t.Errorf("Failed to create cloner with explicit token: %v", err)
	}

	if cloner1.GetToken() != "explicit-token" {
		t.Errorf("Expected token 'explicit-token', got %s", cloner1.GetToken())
	}

	// Create a cloner using environment token
	cloner2, err := factory.CreateCloner(ctx, "")
	if err != nil {
		t.Errorf("Failed to create cloner with environment token: %v", err)
	}

	if cloner2.GetToken() != "test-github-token" {
		t.Errorf("Expected token 'test-github-token', got %s", cloner2.GetToken())
	}

	// Create a cloner with different environment
	customEnv := env.NewMockEnvironment(map[string]string{
		env.CommonEnvironmentKeys.GitHubToken: "custom-token",
	})

	cloner3, err := factory.CreateClonerWithEnv(ctx, "", customEnv)
	if err != nil {
		t.Errorf("Failed to create cloner with custom environment: %v", err)
	}

	if cloner3.GetToken() != "custom-token" {
		t.Errorf("Expected token 'custom-token', got %s", cloner3.GetToken())
	}

	// Verify that all cloners have the correct provider name
	for i, cloner := range []GitHubCloner{cloner1, cloner2, cloner3} {
		if cloner.GetProviderName() != "github" {
			t.Errorf("Cloner %d: expected provider name 'github', got %s", i+1, cloner.GetProviderName())
		}
	}
}

// TestFactoryPatternVersusDirectInstantiation shows the benefit of the factory pattern.
func TestFactoryPatternVersusDirectInstantiation(t *testing.T) {
	mockEnv := env.NewMockEnvironment(map[string]string{
		env.CommonEnvironmentKeys.GitHubToken: "env-token",
	})

	ctx := context.Background()

	// Old way: Direct instantiation (harder to test, less flexible)
	oldCloner := &gitHubClonerImpl{
		Token:       "direct-token",
		Environment: mockEnv,
	}

	// New way: Factory pattern (easier to test, more flexible)
	factory := NewGitHubProviderFactory(mockEnv)

	newCloner, err := factory.CreateCloner(ctx, "factory-token")
	if err != nil {
		t.Errorf("Failed to create cloner via factory: %v", err)
	}

	// Both should work, but factory provides more flexibility
	if oldCloner.GetProviderName() != newCloner.GetProviderName() {
		t.Error("Both cloners should have the same provider name")
	}

	// Factory allows easy swapping of implementations and dependency injection
	// while direct instantiation couples the code to specific implementations

	// Factory pattern also allows for configuration-driven creation
	configEnv := env.NewMockEnvironment(map[string]string{
		env.CommonEnvironmentKeys.GitHubToken: "config-default-token",
	})
	config := &GitHubFactoryConfig{
		DefaultToken: "config-default-token",
		Environment:  configEnv,
	}

	configuredFactory := NewGitHubProviderFactoryWithConfig(config)

	configuredCloner, err := configuredFactory.CreateCloner(ctx, "")
	if err != nil {
		t.Errorf("Failed to create cloner with configured factory: %v", err)
	}

	// This demonstrates the power of factory pattern for configuration-driven instantiation
	if configuredCloner == nil {
		t.Error("Configured factory should create valid cloner")
	}
}
