//nolint:testpackage // White-box testing needed for internal function access
package github

import (
	"context"
	"testing"

	"github.com/Gizzahub/gzh-manager-go/internal/env"
)

func TestGitHubProviderFactory(t *testing.T) {
	mockEnv := env.NewMockEnvironment(map[string]string{
		"GITHUB_TOKEN": "test-github-token",
	})

	factory := NewGitHubProviderFactory(mockEnv)

	// Test GetProviderName
	if factory.GetProviderName() != "github" {
		t.Errorf("Expected provider name to be 'github', got %s", factory.GetProviderName())
	}

	ctx := context.Background()

	// Test CreateCloner with token
	cloner, err := factory.CreateCloner(ctx, "explicit-token")
	if err != nil {
		t.Errorf("Failed to create cloner: %v", err)
	}

	if cloner == nil {
		t.Error("Cloner should not be nil")
	}

	if cloner.GetProviderName() != "github" {
		t.Errorf("Expected cloner provider name to be 'github', got %s", cloner.GetProviderName())
	}

	// Test CreateCloner without token (should use environment)
	cloner2, err := factory.CreateCloner(ctx, "")
	if err != nil {
		t.Errorf("Failed to create cloner from environment: %v", err)
	}

	if cloner2 == nil {
		t.Error("Cloner should not be nil")
	}

	// Test CreateClonerWithEnv
	customEnv := env.NewMockEnvironment(map[string]string{
		"GITHUB_TOKEN": "custom-token",
	})

	cloner3, err := factory.CreateClonerWithEnv(ctx, "", customEnv)
	if err != nil {
		t.Errorf("Failed to create cloner with custom environment: %v", err)
	}

	if cloner3 == nil {
		t.Error("Cloner should not be nil")
	}

	// Test CreateCloner with empty token and no environment token
	emptyEnv := env.NewMockEnvironment(map[string]string{})

	cloner4, err := factory.CreateClonerWithEnv(ctx, "", emptyEnv)
	if err == nil {
		t.Error("Expected error when no token is provided")
	}

	if cloner4 != nil {
		t.Error("Cloner should be nil when no token is provided")
	}
}

func TestGitHubProviderFactoryWithConfig(t *testing.T) {
	mockEnv := env.NewMockEnvironment(map[string]string{
		"GITHUB_TOKEN": "test-github-token",
	})

	config := &GitHubFactoryConfig{
		DefaultToken: "config-token",
		Environment:  mockEnv,
	}

	factory := NewGitHubProviderFactoryWithConfig(config)

	if factory == nil {
		t.Error("Factory should not be nil")
	}

	if factory.GetProviderName() != "github" {
		t.Errorf("Expected provider name to be 'github', got %s", factory.GetProviderName())
	}
}

func TestDefaultGitHubFactoryConfig(t *testing.T) {
	config := DefaultGitHubFactoryConfig()

	if config == nil {
		t.Error("Default config should not be nil")
		return
	}

	if config.Environment == nil {
		t.Error("Default environment should not be nil")
	}
}

func TestGitHubClonerImpl(t *testing.T) {
	mockEnv := env.NewMockEnvironment(map[string]string{})

	cloner := &gitHubClonerImpl{
		Token:       "test-token",
		Environment: mockEnv,
	}

	// Test GetProviderName
	if cloner.GetProviderName() != "github" {
		t.Errorf("Expected provider name to be 'github', got %s", cloner.GetProviderName())
	}

	// Test SetToken and GetToken
	cloner.SetToken("new-token")

	if cloner.GetToken() != "new-token" {
		t.Errorf("Expected token to be 'new-token', got %s", cloner.GetToken())
	}

	// Test CloneRepository (should return not implemented error)
	ctx := context.Background()

	err := cloner.CloneRepository(ctx, "owner", "repo", "/tmp/test", "reset")
	if err == nil {
		t.Error("Expected error for unimplemented CloneRepository")
	}

	if err.Error() != "CloneRepository not yet implemented" {
		t.Errorf("Expected 'CloneRepository not yet implemented' error, got %s", err.Error())
	}
}
