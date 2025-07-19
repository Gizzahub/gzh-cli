package env

import (
	"testing"
)

func TestOSEnvironment(t *testing.T) {
	env := NewOSEnvironment()

	// Test setting and getting a value
	key := "TEST_ENV_VAR"
	value := "test_value"

	err := env.Set(key, value)
	if err != nil {
		t.Errorf("Failed to set environment variable: %v", err)
	}

	retrievedValue := env.Get(key)
	if retrievedValue != value {
		t.Errorf("Expected %s, got %s", value, retrievedValue)
	}

	// Test LookupEnv
	lookedUpValue, exists := env.LookupEnv(key)
	if !exists {
		t.Error("Environment variable should exist")
	}

	if lookedUpValue != value {
		t.Errorf("Expected %s, got %s", value, lookedUpValue)
	}

	// Clean up
	if err := env.Unset(key); err != nil {
		t.Logf("Warning: failed to unset environment variable: %v", err)
	}
}

func TestMockEnvironment(t *testing.T) {
	initialVars := map[string]string{
		"TEST_VAR": "initial_value",
	}

	env := NewMockEnvironment(initialVars)

	// Test getting initial value
	value := env.Get("TEST_VAR")
	if value != "initial_value" {
		t.Errorf("Expected initial_value, got %s", value)
	}

	// Test setting new value
	if err := env.Set("NEW_VAR", "new_value"); err != nil {
		t.Fatalf("Failed to set NEW_VAR: %v", err)
	}

	newValue := env.Get("NEW_VAR")
	if newValue != "new_value" {
		t.Errorf("Expected new_value, got %s", newValue)
	}

	// Test LookupEnv for non-existing variable
	_, exists := env.LookupEnv("NON_EXISTING")
	if exists {
		t.Error("Non-existing variable should not exist")
	}

	// Test Expand
	if err := env.Set("GREETING", "Hello"); err != nil {
		t.Fatalf("Failed to set GREETING: %v", err)
	}

	expanded := env.Expand("$GREETING World")
	if expanded != "Hello World" {
		t.Errorf("Expected 'Hello World', got %s", expanded)
	}

	// Test GetAll
	allVars := env.GetAll()
	expectedKeys := map[string]bool{
		"TEST_VAR": true,
		"NEW_VAR":  true,
		"GREETING": true,
	}

	if len(allVars) != len(expectedKeys) {
		t.Errorf("Expected %d variables, got %d: %v", len(expectedKeys), len(allVars), allVars)
	}

	for key := range expectedKeys {
		if _, exists := allVars[key]; !exists {
			t.Errorf("Expected key %s to exist in environment", key)
		}
	}
}

func TestCommonEnvironmentKeys(t *testing.T) {
	keys := CommonEnvironmentKeys

	if keys.GitHubToken != "GITHUB_TOKEN" {
		t.Errorf("Expected GITHUB_TOKEN, got %s", keys.GitHubToken)
	}

	if keys.GitLabToken != "GITLAB_TOKEN" {
		t.Errorf("Expected GITLAB_TOKEN, got %s", keys.GitLabToken)
	}

	if keys.GZHConfigPath != "GZH_CONFIG_PATH" {
		t.Errorf("Expected GZH_CONFIG_PATH, got %s", keys.GZHConfigPath)
	}
}
