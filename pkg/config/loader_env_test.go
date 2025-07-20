//nolint:testpackage // White-box testing needed for internal function access
package config

import (
	"testing"

	"github.com/gizzahub/gzh-manager-go/internal/env"
)

func TestConfigLoaderWithEnvironment(t *testing.T) {
	// Create a mock environment for testing
	mockEnv := env.NewMockEnvironment(map[string]string{
		"GZH_CONFIG_PATH": "/test/config.yaml",
		"HOME":            "/test/home",
	})

	// Test FindConfigFileWithEnv
	_, err := FindConfigFileWithEnv(mockEnv)
	// We expect an error because the file doesn't exist, but this demonstrates
	// that the environment abstraction is working
	if err == nil {
		t.Error("Expected error for non-existent config file")
	}

	// Test path expansion with mock environment
	testPath := "~/config/test.yaml"
	expandedPath := expandPathWithEnv(testPath, mockEnv)
	expectedPath := "/test/home/config/test.yaml"

	if expandedPath != expectedPath {
		t.Errorf("Expected %s, got %s", expectedPath, expandedPath)
	}
}

func TestEnvironmentDependencyInjection(t *testing.T) {
	// Demonstrate that we can inject different environments
	osEnv := env.NewOSEnvironment()
	mockEnv := env.NewMockEnvironment(map[string]string{
		"GZH_CONFIG_PATH": "/mock/path/config.yaml",
	})

	// Both should use their respective environments
	osConfigPath := osEnv.Get("GZH_CONFIG_PATH")     // May be empty
	mockConfigPath := mockEnv.Get("GZH_CONFIG_PATH") // Will be "/mock/path/config.yaml"

	// The mock environment should return our test value
	if mockConfigPath != "/mock/path/config.yaml" {
		t.Errorf("Mock environment should return test value, got %s", mockConfigPath)
	}

	// They should be different (unless OS env happens to have same value)
	if osConfigPath == mockConfigPath && osConfigPath != "" {
		t.Log("OS environment happened to have same value as mock - this is unlikely but possible")
	}
}
