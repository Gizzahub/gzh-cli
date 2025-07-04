package config

import (
	"context"
	"testing"

	"github.com/gizzahub/gzh-manager-go/internal/env"
)

// MockLogger implements the Logger interface for testing
type MockLogger struct{}

func (m *MockLogger) Debug(msg string, args ...interface{}) {}
func (m *MockLogger) Info(msg string, args ...interface{})  {}
func (m *MockLogger) Warn(msg string, args ...interface{})  {}
func (m *MockLogger) Error(msg string, args ...interface{}) {}

func TestProviderFactory(t *testing.T) {
	mockEnv := env.NewMockEnvironment(map[string]string{
		"GITHUB_TOKEN": "test-github-token",
		"GITLAB_TOKEN": "test-gitlab-token",
		"GITEA_TOKEN":  "test-gitea-token",
	})

	logger := &MockLogger{}
	factory := NewProviderFactory(mockEnv, logger)

	// Test GetSupportedProviders
	supportedProviders := factory.GetSupportedProviders()
	expectedProviders := []string{"github", "gitlab", "gitea"}

	if len(supportedProviders) != len(expectedProviders) {
		t.Errorf("Expected %d providers, got %d", len(expectedProviders), len(supportedProviders))
	}

	for i, provider := range expectedProviders {
		if supportedProviders[i] != provider {
			t.Errorf("Expected provider %s, got %s", provider, supportedProviders[i])
		}
	}

	// Test IsProviderSupported
	if !factory.IsProviderSupported("github") {
		t.Error("GitHub should be supported")
	}

	if !factory.IsProviderSupported("gitlab") {
		t.Error("GitLab should be supported")
	}

	if !factory.IsProviderSupported("gitea") {
		t.Error("Gitea should be supported")
	}

	if factory.IsProviderSupported("invalid") {
		t.Error("Invalid provider should not be supported")
	}

	ctx := context.Background()

	// Test CreateCloner for GitHub
	githubCloner, err := factory.CreateCloner(ctx, "github", "test-token")
	if err != nil {
		t.Errorf("Failed to create GitHub cloner: %v", err)
	}
	if githubCloner == nil {
		t.Error("GitHub cloner should not be nil")
	}
	if githubCloner.GetName() != "github" {
		t.Errorf("Expected GitHub cloner name to be 'github', got %s", githubCloner.GetName())
	}

	// Test CreateCloner for GitLab
	gitlabCloner, err := factory.CreateCloner(ctx, "gitlab", "test-token")
	if err != nil {
		t.Errorf("Failed to create GitLab cloner: %v", err)
	}
	if gitlabCloner == nil {
		t.Error("GitLab cloner should not be nil")
	}
	if gitlabCloner.GetName() != "gitlab" {
		t.Errorf("Expected GitLab cloner name to be 'gitlab', got %s", gitlabCloner.GetName())
	}

	// Test CreateCloner for Gitea
	giteaCloner, err := factory.CreateCloner(ctx, "gitea", "test-token")
	if err != nil {
		t.Errorf("Failed to create Gitea cloner: %v", err)
	}
	if giteaCloner == nil {
		t.Error("Gitea cloner should not be nil")
	}
	if giteaCloner.GetName() != "gitea" {
		t.Errorf("Expected Gitea cloner name to be 'gitea', got %s", giteaCloner.GetName())
	}

	// Test CreateCloner for unsupported provider
	_, err = factory.CreateCloner(ctx, "invalid", "test-token")
	if err == nil {
		t.Error("Expected error for unsupported provider")
	}
}

func TestProviderFactoryWithConfig(t *testing.T) {
	mockEnv := env.NewMockEnvironment(map[string]string{
		"GITHUB_TOKEN": "test-github-token",
	})

	logger := &MockLogger{}
	config := &ProviderFactoryConfig{
		DefaultEnvironment: mockEnv,
		EnableLogging:      true,
	}

	factory := NewProviderFactoryWithConfig(config, logger)

	// Test that the factory was created successfully
	if factory == nil {
		t.Error("Factory should not be nil")
	}

	// Test creating a cloner
	ctx := context.Background()
	cloner, err := factory.CreateCloner(ctx, "github", "test-token")
	if err != nil {
		t.Errorf("Failed to create cloner: %v", err)
	}
	if cloner == nil {
		t.Error("Cloner should not be nil")
	}
}

func TestProviderFactoryWithEnvironment(t *testing.T) {
	mockEnv := env.NewMockEnvironment(map[string]string{
		"GITHUB_TOKEN": "env-github-token",
		"GITLAB_TOKEN": "env-gitlab-token",
		"GITEA_TOKEN":  "env-gitea-token",
	})

	logger := &MockLogger{}
	factory := NewProviderFactory(mockEnv, logger)

	ctx := context.Background()

	// Test CreateClonerWithEnv
	customEnv := env.NewMockEnvironment(map[string]string{
		"GITHUB_TOKEN": "custom-github-token",
	})

	cloner, err := factory.CreateClonerWithEnv(ctx, "github", "test-token", customEnv)
	if err != nil {
		t.Errorf("Failed to create cloner with custom environment: %v", err)
	}
	if cloner == nil {
		t.Error("Cloner should not be nil")
	}
}

func TestDefaultProviderFactoryConfig(t *testing.T) {
	config := DefaultProviderFactoryConfig()

	if config == nil {
		t.Error("Default config should not be nil")
	}

	if config.DefaultEnvironment == nil {
		t.Error("Default environment should not be nil")
	}

	if !config.EnableLogging {
		t.Error("Logging should be enabled by default")
	}
}
