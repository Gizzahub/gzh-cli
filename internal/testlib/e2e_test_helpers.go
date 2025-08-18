// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package testlib

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// E2ETestConfig represents common configuration for E2E tests.
type E2ETestConfig struct {
	TestTimeout     time.Duration
	SkipCondition   func() bool
	SkipMessage     string
	SetupFunc       func(t *testing.T) interface{}
	CleanupFunc     func(t *testing.T, data interface{})
	RequiredEnvVars []string
}

// E2ETestSuite provides common E2E test functionality.
type E2ETestSuite struct {
	t      *testing.T
	config E2ETestConfig
	data   interface{}
}

// NewE2ETestSuite creates a new E2E test suite.
func NewE2ETestSuite(t *testing.T, config E2ETestConfig) *E2ETestSuite {
	return &E2ETestSuite{
		t:      t,
		config: config,
	}
}

// Setup performs common E2E test setup.
func (e *E2ETestSuite) Setup() {
	e.t.Helper()

	// Check skip condition
	if e.config.SkipCondition != nil && e.config.SkipCondition() {
		e.t.Skip(e.config.SkipMessage)
	}

	// Check required environment variables
	for _, envVar := range e.config.RequiredEnvVars {
		if os.Getenv(envVar) == "" {
			e.t.Skipf("Required environment variable %s not set", envVar)
		}
	}

	// Set test timeout
	if e.config.TestTimeout > 0 {
		e.t.Deadline() // This would be used with context for timeout
	}

	// Run setup function
	if e.config.SetupFunc != nil {
		e.data = e.config.SetupFunc(e.t)
	}
}

// Cleanup performs common E2E test cleanup.
func (e *E2ETestSuite) Cleanup() {
	e.t.Helper()
	if e.config.CleanupFunc != nil {
		e.config.CleanupFunc(e.t, e.data)
	}
}

// GetSetupData returns the data from setup function.
func (e *E2ETestSuite) GetSetupData() interface{} {
	return e.data
}

// AssertCommandExecution verifies command execution results.
func AssertCommandExecution(t *testing.T, err error, output string, expectedPatterns ...string) {
	t.Helper()
	assert.NoError(t, err, "Command should execute successfully")
	assert.NotEmpty(t, output, "Command output should not be empty")

	for _, pattern := range expectedPatterns {
		assert.Contains(t, output, pattern, "Output should contain pattern: %s", pattern)
	}
}

// AssertConfigFileOperations tests common config file operations.
func AssertConfigFileOperations(t *testing.T, configPath string, operations ...ConfigOperation) {
	t.Helper()

	for _, op := range operations {
		switch op.Type {
		case ConfigOpCreate:
			err := op.Execute(configPath)
			assert.NoError(t, err, "Config creation should succeed")
			assert.FileExists(t, configPath, "Config file should exist after creation")

		case ConfigOpRead:
			assert.FileExists(t, configPath, "Config file should exist before reading")
			err := op.Execute(configPath)
			assert.NoError(t, err, "Config reading should succeed")

		case ConfigOpUpdate:
			assert.FileExists(t, configPath, "Config file should exist before updating")
			err := op.Execute(configPath)
			assert.NoError(t, err, "Config update should succeed")

		case ConfigOpDelete:
			err := op.Execute(configPath)
			assert.NoError(t, err, "Config deletion should succeed")
			assert.NoFileExists(t, configPath, "Config file should not exist after deletion")
		}
	}
}

// ConfigOperationType defines types of config operations.
type ConfigOperationType int

const (
	ConfigOpCreate ConfigOperationType = iota
	ConfigOpRead
	ConfigOpUpdate
	ConfigOpDelete
)

// ConfigOperation represents a configuration operation.
type ConfigOperation struct {
	Type    ConfigOperationType
	Execute func(configPath string) error
}

// AssertDirectoryStructure verifies expected directory structure.
func AssertDirectoryStructure(t *testing.T, basePath string, expectedPaths []string) {
	t.Helper()

	for _, path := range expectedPaths {
		fullPath := basePath + "/" + path
		assert.FileExists(t, fullPath, "Expected path should exist: %s", path)
	}
}

// AssertServiceHealth checks common service health patterns.
func AssertServiceHealth(t *testing.T, healthCheck func() (bool, error), timeout time.Duration) {
	t.Helper()

	start := time.Now()
	for time.Since(start) < timeout {
		healthy, err := healthCheck()
		if err == nil && healthy {
			return // Service is healthy
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("Service did not become healthy within %v", timeout)
}

// AssertIntegrationTest runs a common integration test pattern.
func AssertIntegrationTest(t *testing.T, testName string, testFunc func(t *testing.T) error) {
	t.Helper()
	t.Run(testName, func(t *testing.T) {
		err := testFunc(t)
		assert.NoError(t, err, "Integration test %s should pass", testName)
	})
}

// RetryOperation retries an operation with common retry logic.
func RetryOperation(t *testing.T, operation func() error, maxRetries int, delay time.Duration) error {
	t.Helper()

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		lastErr = operation()
		if lastErr == nil {
			return nil
		}

		if i < maxRetries-1 {
			time.Sleep(delay)
		}
	}

	return lastErr
}

// AssertEnvironmentSetup verifies common environment setup requirements.
func AssertEnvironmentSetup(t *testing.T, requirements EnvironmentRequirements) {
	t.Helper()

	// Check required environment variables
	for _, envVar := range requirements.RequiredEnvVars {
		value := os.Getenv(envVar)
		assert.NotEmpty(t, value, "Required environment variable %s should be set", envVar)
	}

	// Check required directories
	for _, dir := range requirements.RequiredDirectories {
		assert.DirExists(t, dir, "Required directory should exist: %s", dir)
	}

	// Check required files
	for _, file := range requirements.RequiredFiles {
		assert.FileExists(t, file, "Required file should exist: %s", file)
	}
}

// EnvironmentRequirements defines environment setup requirements.
type EnvironmentRequirements struct {
	RequiredEnvVars     []string
	RequiredDirectories []string
	RequiredFiles       []string
}
