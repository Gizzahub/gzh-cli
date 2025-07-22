//nolint:testpackage // White-box testing needed for internal function access
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gizzahub/gzh-manager-go/internal/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name          string
		configContent string
		setupEnv      func()
		cleanupEnv    func()
		strict        bool
		expectError   bool
		errorContains string
	}{
		{
			name: "valid configuration",
			configContent: `
version: "1.0.0"
default_provider: github
providers:
  github:
    token: "test-token"
    orgs:
      - name: "test-org"
        visibility: "all"
        clone_dir: "/tmp/test-repos"
`,
			expectError: false,
		},
		{
			name: "invalid YAML syntax",
			configContent: `
version: "1.0.0"
providers:
  github:
    token: "unclosed string
`,
			expectError:   true,
			errorContains: "parsing failed",
		},
		{
			name: "missing required field",
			configContent: `
providers:
  github:
    token: "test-token"
`,
			expectError:   true,
			errorContains: "version",
		},
		{
			name: "invalid regex pattern",
			configContent: `
version: "1.0.0"
providers:
  github:
    token: "test-token"
    orgs:
      - name: "test-org"
        match: "[invalid"
`,
			expectError:   true,
			errorContains: "invalid regex pattern",
		},
		{
			name: "environment variable configuration",
			configContent: `
version: "1.0.0"
providers:
  github:
    token: "${TEST_GITHUB_TOKEN}"
    orgs:
      - name: "test-org"
        clone_dir: "${HOME}/repos"
`,
			setupEnv: func() {
				_ = os.Setenv("TEST_GITHUB_TOKEN", "test-token-value")
				_ = os.Setenv("HOME", "/home/testuser")
			},
			cleanupEnv: func() {
				if err := os.Unsetenv("TEST_GITHUB_TOKEN"); err != nil {
					// Environment cleanup error in test is non-critical
					t.Logf("Failed to unset TEST_GITHUB_TOKEN: %v", err)
				}
				if err := os.Unsetenv("HOME"); err != nil {
					// Environment cleanup error in test is non-critical
					t.Logf("Failed to unset HOME: %v", err)
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment if needed
			if tt.setupEnv != nil {
				tt.setupEnv()
			}

			if tt.cleanupEnv != nil {
				defer tt.cleanupEnv()
			}

			// Create temporary config file
			tmpDir, err := os.MkdirTemp("", "config-validate-test-*")
			require.NoError(t, err)

			defer func() {
				if err := os.RemoveAll(tmpDir); err != nil {
					t.Logf("Warning: failed to remove temp dir: %v", err)
				}
			}()

			configFile := filepath.Join(tmpDir, "gzh.yaml")
			err = os.WriteFile(configFile, []byte(tt.configContent), 0o644)
			require.NoError(t, err)

			// Run validation
			err = validateConfig(configFile, tt.strict, false)

			if tt.expectError {
				assert.Error(t, err)

				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFindConfigFile(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "find-config-test-*")
	require.NoError(t, err)

	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: failed to change back to original dir: %v", err)
		}
	}()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	tests := []struct {
		name         string
		setupFiles   func()
		setupEnv     func()
		cleanupEnv   func()
		expectError  bool
		expectedFile string
	}{
		{
			name: "find gzh.yaml in current directory",
			setupFiles: func() {
				if err := os.WriteFile("gzh.yaml", []byte("version: 1.0.0"), 0o644); err != nil {
					t.Errorf("failed to write test file: %v", err)
				}
			},
			expectError:  false,
			expectedFile: "gzh.yaml",
		},
		{
			name: "find gzh.yml in current directory",
			setupFiles: func() {
				if err := os.WriteFile("gzh.yml", []byte("version: 1.0.0"), 0o644); err != nil {
					t.Errorf("failed to write test file: %v", err)
				}
			},
			expectError:  false,
			expectedFile: "gzh.yml",
		},
		{
			name: "find via environment variable",
			setupFiles: func() {
				customFile := filepath.Join(tmpDir, "custom-config.yaml")
				if err := os.WriteFile(customFile, []byte("version: 1.0.0"), 0o644); err != nil {
					t.Errorf("failed to write test file: %v", err)
				}
			},
			setupEnv: func() {
				customFile := filepath.Join(tmpDir, "custom-config.yaml")
				if err := os.Setenv("GZH_CONFIG_PATH", customFile); err != nil {
					t.Errorf("failed to set environment variable: %v", err)
				}
			},
			cleanupEnv: func() {
				if err := os.Unsetenv("GZH_CONFIG_PATH"); err != nil {
					t.Logf("Warning: failed to unset GZH_CONFIG_PATH: %v", err)
				}
			},
			expectError:  false,
			expectedFile: "custom-config.yaml",
		},
		{
			name:        "no config file found",
			setupFiles:  func() {}, // No files
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing files
			_ = os.Remove("gzh.yaml") //nolint:errcheck // Test cleanup, errors are non-critical
			_ = os.Remove("gzh.yml")  //nolint:errcheck // Test cleanup, errors are non-critical

			if tt.setupEnv != nil {
				tt.setupEnv()
			}

			if tt.cleanupEnv != nil {
				defer tt.cleanupEnv()
			}

			tt.setupFiles()

			configFile, err := findConfigFile()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, configFile, tt.expectedFile)
			}
		})
	}
}

func TestValidateFileAccess(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "file-access-test-*")
	require.NoError(t, err)

	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	tests := []struct {
		name          string
		setupFile     func() string
		expectError   bool
		errorContains string
	}{
		{
			name: "valid file",
			setupFile: func() string {
				file := filepath.Join(tmpDir, "valid.yaml")
				_ = os.WriteFile(file, []byte("test"), 0o644) //nolint:errcheck // Test setup, errors handled by test framework
				return file
			},
			expectError: false,
		},
		{
			name: "non-existent file",
			setupFile: func() string {
				return filepath.Join(tmpDir, "nonexistent.yaml")
			},
			expectError:   true,
			errorContains: "does not exist",
		},
		{
			name: "directory instead of file",
			setupFile: func() string {
				dir := filepath.Join(tmpDir, "notafile")
				_ = os.Mkdir(dir, 0o755) //nolint:errcheck // Test setup, errors handled by test framework
				return dir
			},
			expectError:   true,
			errorContains: "not a regular file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setupFile()
			err := validateFileAccess(filePath)

			if tt.expectError {
				assert.Error(t, err)

				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCloneDirectory(t *testing.T) {
	tests := []struct {
		name          string
		cloneDir      string
		strict        bool
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid directory",
			cloneDir:    "/tmp/test-repos",
			strict:      false,
			expectError: false,
		},
		{
			name:          "unsafe path",
			cloneDir:      "/tmp/../etc/passwd",
			strict:        false,
			expectError:   true,
			errorContains: "potentially unsafe path",
		},
		{
			name:        "environment variable",
			cloneDir:    "${HOME}/repos",
			strict:      false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCloneDirectory(tt.cloneDir, tt.strict)

			if tt.expectError {
				assert.Error(t, err)

				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckPathEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		setupEnv      func()
		cleanupEnv    func()
		expectError   bool
		errorContains string
	}{
		{
			name:        "no environment variables",
			path:        "/simple/path",
			expectError: false,
		},
		{
			name: "existing environment variable",
			path: "${TEST_VAR}/repos",
			setupEnv: func() {
				_ = os.Setenv("TEST_VAR", "/home/user") //nolint:errcheck // Test setup, errors handled by test framework
			},
			cleanupEnv: func() {
				_ = os.Unsetenv("TEST_VAR") //nolint:errcheck // Test cleanup, errors are non-critical
			},
			expectError: false,
		},
		{
			name:          "missing environment variable",
			path:          "${MISSING_VAR}/repos",
			expectError:   true,
			errorContains: "not found",
		},
		{
			name: "environment variable with default",
			path: "${TEST_VAR:default}/repos",
			setupEnv: func() {
				_ = os.Setenv("TEST_VAR", "/home/user") //nolint:errcheck // Test setup, errors handled by test framework
			},
			cleanupEnv: func() {
				_ = os.Unsetenv("TEST_VAR") //nolint:errcheck // Test cleanup, errors are non-critical
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupEnv != nil {
				tt.setupEnv()
			}

			if tt.cleanupEnv != nil {
				defer tt.cleanupEnv()
			}

			err := checkPathEnvironmentVariables(tt.path)

			if tt.expectError {
				assert.Error(t, err)

				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateConfigWithEnvironmentAbstraction tests that environment abstraction works properly.
func TestValidateConfigWithEnvironmentAbstraction(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "env-abstraction-test-*")
	require.NoError(t, err)

	defer func() { _ = os.RemoveAll(tmpDir) }() // Ignore cleanup error

	// Create a test config file
	configContent := `
version: "1.0.0"
default_provider: github
providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "test-org"
        visibility: "all"
        clone_dir: "${HOME}/repos"
`
	configFile := filepath.Join(tmpDir, "test-config.yaml")
	err = os.WriteFile(configFile, []byte(configContent), 0o644)
	require.NoError(t, err)

	t.Run("with mock environment", func(t *testing.T) {
		// Create mock environment with required variables
		mockEnv := env.NewMockEnvironment(map[string]string{
			"GITHUB_TOKEN": "mock-token-123",
			"HOME":         "/home/testuser",
		})

		// Test findConfigFileWithEnv function
		testHome := "/home/testuser"
		mockEnvForFind := env.NewMockEnvironment(map[string]string{
			"HOME":            testHome,
			"GZH_CONFIG_PATH": configFile,
		})

		foundFile, err := findConfigFileWithEnv(mockEnvForFind)
		assert.NoError(t, err)
		assert.Equal(t, configFile, foundFile)

		// Test validateConfigWithEnv function
		err = validateConfigWithEnv(configFile, false, false, mockEnv)
		assert.NoError(t, err)
	})

	t.Run("with missing environment variables", func(t *testing.T) {
		// Create mock environment without required variables
		mockEnv := env.NewMockEnvironment(map[string]string{})

		// This should succeed but generate warnings since we're not in strict mode
		err = validateConfigWithEnv(configFile, false, false, mockEnv)
		assert.NoError(t, err)
	})

	t.Run("environment variable expansion", func(t *testing.T) {
		mockEnv := env.NewMockEnvironment(map[string]string{
			"TEST_VAR": "test-value",
		})

		expanded := mockEnv.Expand("${TEST_VAR}/subdir")
		assert.Equal(t, "test-value/subdir", expanded)

		// Test path environment validation
		err := checkPathEnvironmentVariablesWithEnv("${TEST_VAR}/path", mockEnv)
		assert.NoError(t, err)

		// Test with missing variable
		err = checkPathEnvironmentVariablesWithEnv("${MISSING_VAR}/path", mockEnv)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "MISSING_VAR")
	})
}
