//nolint:testpackage // White-box testing needed for internal function access
package config

import (
	"testing"
)

func TestConfigBuilder(t *testing.T) {
	// Test building a simple config
	config := NewConfigBuilder().
		WithVersion("1.0.0").
		WithDefaultProvider("github").
		WithGitHubProvider("test-token").
		WithOrganization("github", "test-org", "~/repos/test").
		Build()

	if config.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", config.Version)
	}

	if config.DefaultProvider != ProviderGitHub {
		t.Errorf("Expected default provider 'github', got '%s'", config.DefaultProvider)
	}

	if len(config.Providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(config.Providers))
	}

	github := config.Providers["github"]
	if github == nil {
		t.Error("Expected GitHub provider to be configured")
		return
	}

	if github.Token != "test-token" {
		t.Errorf("Expected token 'test-token', got '%s'", github.Token)
	}

	if len(github.Organizations) != 1 {
		t.Errorf("Expected 1 organization, got %d", len(github.Organizations))
	}

	org := github.Organizations[0]
	if org.Name != "test-org" {
		t.Errorf("Expected organization name 'test-org', got '%s'", org.Name)
	}

	if org.CloneDir != "~/repos/test" {
		t.Errorf("Expected clone dir '~/repos/test', got '%s'", org.CloneDir)
	}
}

func TestConfigBuilderWithDetails(t *testing.T) {
	// Test building a config with detailed organization settings
	config := NewConfigBuilder().
		WithVersion("1.0.0").
		WithDefaultProvider("github").
		WithGitHubProvider("test-token").
		WithOrganizationDetails("github", "test-org", "~/repos/test", "private", "pull").
		Build()

	github := config.Providers["github"]
	if github == nil {
		t.Error("Expected GitHub provider to be configured")
		return
	}

	if len(github.Organizations) != 1 {
		t.Errorf("Expected 1 organization, got %d", len(github.Organizations))
	}

	org := github.Organizations[0]
	if org.Visibility != "private" {
		t.Errorf("Expected visibility 'private', got '%s'", org.Visibility)
	}

	if org.Strategy != "pull" {
		t.Errorf("Expected strategy 'pull', got '%s'", org.Strategy)
	}
}

func TestConfigBuilderBuildYAML(t *testing.T) {
	// Test building YAML output
	yaml := NewConfigBuilder().
		WithVersion("1.0.0").
		WithDefaultProvider("github").
		WithGitHubProvider("test-token").
		WithOrganization("github", "test-org", "~/repos/test").
		BuildYAML()

	if yaml == "" {
		t.Error("Expected YAML output to be non-empty")
	}

	// Check that YAML contains expected elements
	if !containsString(yaml, "version: \"1.0.0\"") {
		t.Error("Expected YAML to contain version")
	}

	if !containsString(yaml, "default_provider: github") {
		t.Error("Expected YAML to contain default_provider")
	}

	if !containsString(yaml, "github:") {
		t.Error("Expected YAML to contain github provider")
	}

	if !containsString(yaml, "token: \"test-token\"") {
		t.Error("Expected YAML to contain token")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr || len(s) > len(substr) && containsString(s[1:], substr)
}
