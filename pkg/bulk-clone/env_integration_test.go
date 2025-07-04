package bulkclone

import (
	"testing"

	"github.com/gizzahub/gzh-manager-go/internal/env"
)

// TestEnvironmentAbstraction demonstrates the environment abstraction working
func TestEnvironmentAbstraction(t *testing.T) {
	// Test with mock environment
	mockEnv := env.NewMockEnvironment(map[string]string{
		"GZH_CONFIG_PATH": "/test/config.yaml",
		"HOME":            "/test/home",
	})

	// Test FindConfigFileWithEnv
	_, err := FindConfigFileWithEnv(mockEnv)
	if err == nil {
		t.Error("Expected error for non-existent config file, but got nil")
	}

	expectedError := "config file specified in GZH_CONFIG_PATH not found: /test/config.yaml"
	if err.Error() != expectedError {
		t.Errorf("Expected error: %s, got: %s", expectedError, err.Error())
	}

	// Test path expansion
	testPath := "~/config/test.yaml"
	expanded := ExpandPathWithEnv(testPath, mockEnv)
	expected := "/test/home/config/test.yaml"

	if expanded != expected {
		t.Errorf("Expected expanded path: %s, got: %s", expected, expanded)
	}
}

// TestEnvironmentAbstractionBefore demonstrates what the code looked like before abstraction
func TestEnvironmentAbstractionBefore(t *testing.T) {
	// Before: Direct os.Getenv() call (this would be hard to test)
	// configPath := os.Getenv("GZH_CONFIG_PATH")

	// After: We can inject a mock environment for testing
	mockEnv := env.NewMockEnvironment(map[string]string{
		"GZH_CONFIG_PATH": "/mock/path",
		"HOME":            "/mock/home",
	})

	// Now we can predictably test environment-dependent behavior
	configPath := mockEnv.Get("GZH_CONFIG_PATH")
	if configPath != "/mock/path" {
		t.Errorf("Expected /mock/path, got %s", configPath)
	}

	homeDir := mockEnv.Get("HOME")
	if homeDir != "/mock/home" {
		t.Errorf("Expected /mock/home, got %s", homeDir)
	}
}
