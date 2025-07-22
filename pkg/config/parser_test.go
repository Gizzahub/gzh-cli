//nolint:testpackage // White-box testing needed for internal function access
package config

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseYAML(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		setup   func()
		cleanup func()
	}{
		{
			name: "valid YAML with environment variables",
			yaml: `
version: "1.0.0"
default_provider: github
providers:
  github:
    token: "${TEST_TOKEN}"
    orgs:
      - name: "test-org"
        visibility: "public"
`,
			setup: func() {
				_ = os.Setenv("TEST_TOKEN", "test-token-value") // Ignore error
			},
			cleanup: func() {
				_ = os.Unsetenv("TEST_TOKEN") // Ignore error
			},
			wantErr: false,
		},
		{
			name: "invalid YAML",
			yaml: `
version: "1.0.0"
invalid: [unclosed
`,
			wantErr: true,
		},
		{
			name: "missing version",
			yaml: `
providers:
  github:
    token: "test"
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			reader := strings.NewReader(tt.yaml)
			config, err := ParseYAML(reader)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				assert.Equal(t, "1.0.0", config.Version)
			}
		})
	}
}

func TestExpandEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envVar   string
		envValue string
		expected string
	}{
		{
			name:     "simple environment variable",
			input:    "${TEST_VAR}",
			envVar:   "TEST_VAR",
			envValue: "test-value",
			expected: "test-value",
		},
		{
			name:     "environment variable in string",
			input:    "prefix-${TEST_VAR}-suffix",
			envVar:   "TEST_VAR",
			envValue: "middle",
			expected: "prefix-middle-suffix",
		},
		{
			name:     "undefined environment variable",
			input:    "${UNDEFINED_VAR}",
			envVar:   "",
			envValue: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				_ = os.Setenv(tt.envVar, tt.envValue)         // Ignore error
				defer func() { _ = os.Unsetenv(tt.envVar) }() // Ignore error
			}

			result := ExpandEnvironmentVariables(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessDefaultValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envVar   string
		envValue string
		expected string
	}{
		{
			name:     "use environment value",
			input:    "token: ${TEST_TOKEN:default-token}",
			envVar:   "TEST_TOKEN",
			envValue: "real-token",
			expected: "token: real-token",
		},
		{
			name:     "use default value",
			input:    "token: ${MISSING_TOKEN:default-token}",
			envVar:   "",
			envValue: "",
			expected: "token: default-token",
		},
		{
			name:     "no default syntax",
			input:    "token: ${TEST_TOKEN}",
			envVar:   "",
			envValue: "",
			expected: "token: ${TEST_TOKEN}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				_ = os.Setenv(tt.envVar, tt.envValue)         // Ignore error
				defer func() { _ = os.Unsetenv(tt.envVar) }() // Ignore error
			}

			result := processDefaultValues(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseYAMLFile(t *testing.T) {
	// Create a temporary YAML file
	content := `
version: "1.0.0"
default_provider: github
providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "test-org"
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	require.NoError(t, err)

	defer func() { _ = os.Remove(tmpFile.Name()) }() // Ignore cleanup error

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	_ = tmpFile.Close() // Ignore close error

	// Set environment variable
	_ = os.Setenv("GITHUB_TOKEN", "test-token")        // Ignore error
	defer func() { _ = os.Unsetenv("GITHUB_TOKEN") }() // Ignore error

	config, err := ParseYAMLFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "1.0.0", config.Version)
	assert.Equal(t, "github", config.DefaultProvider)
	assert.Len(t, config.Providers["github"].Orgs, 1)
	assert.Equal(t, "test-org", config.Providers["github"].Orgs[0].Name)
}

func TestApplyDefaults(t *testing.T) {
	config := &Config{
		Version: "1.0.0",
		Providers: map[string]Provider{
			"github": {
				Token: "test",
				Orgs: []GitTarget{
					{Name: "test-org"}, // No defaults set
				},
			},
		},
	}

	config.applyDefaults()

	assert.Equal(t, ProviderGitHub, config.DefaultProvider)
	assert.Equal(t, VisibilityAll, config.Providers["github"].Orgs[0].Visibility)
	assert.Equal(t, StrategyReset, config.Providers["github"].Orgs[0].Strategy)
}

func TestParseYAML_ComplexScenarios(t *testing.T) {
	tests := []struct {
		name         string
		yaml         string
		setup        func()
		cleanup      func()
		wantErr      bool
		validateFunc func(*testing.T, *Config)
	}{
		{
			name: "simple valid configuration",
			yaml: `
version: "1.0.0"
default_provider: github
providers:
  github:
    token: "github-token"
    orgs:
      - name: "test-org"
        visibility: "public"
        flatten: true
`,
			wantErr: false,
			validateFunc: func(t *testing.T, config *Config) {
				t.Helper()
				assert.Equal(t, "github", config.DefaultProvider)
				assert.Len(t, config.Providers, 1)

				github := config.Providers["github"]
				assert.Equal(t, "github-token", github.Token)
				assert.Len(t, github.Orgs, 1)
				assert.True(t, github.Orgs[0].Flatten)
			},
		},
		{
			name: "yaml with comments and special characters",
			yaml: `# Configuration file for gzh-manager
version: "1.0.0" # Version number
# Default provider setting
default_provider: github

providers:
  github:
    token: "token-with-special!@#$%chars" # GitHub personal access token
    orgs:
      - name: "org-with-special/chars"
        visibility: "all"
        clone_dir: "/path/with spaces/and-special&chars"
        match: "^test-.*\\.go$"
        exclude:
          - "temp-*"
          - "*-backup"
`,
			wantErr: false,
			validateFunc: func(t *testing.T, config *Config) {
				t.Helper()
				assert.Equal(t, "token-with-special!@#$%chars", config.Providers["github"].Token)
				assert.Equal(t, "org-with-special/chars", config.Providers["github"].Orgs[0].Name)
				assert.Equal(t, "/path/with spaces/and-special&chars", config.Providers["github"].Orgs[0].CloneDir)
				assert.Equal(t, "^test-.*\\.go$", config.Providers["github"].Orgs[0].Match)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			reader := strings.NewReader(tt.yaml)
			config, err := ParseYAML(reader)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)

				if tt.validateFunc != nil {
					tt.validateFunc(t, config)
				}
			}
		})
	}
}

func TestEnvironmentVariableExpansion_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envVars  map[string]string
		expected string
	}{
		{
			name:  "multiple environment variables in one string",
			input: "${VAR1}-${VAR2}-${VAR3}",
			envVars: map[string]string{
				"VAR1": "part1",
				"VAR2": "part2",
				"VAR3": "part3",
			},
			expected: "part1-part2-part3",
		},
		{
			name:     "empty environment variable",
			input:    "${EMPTY_VAR}",
			envVars:  map[string]string{"EMPTY_VAR": ""},
			expected: "",
		},
		{
			name:     "undefined environment variable",
			input:    "${UNDEFINED_VAR}",
			envVars:  map[string]string{},
			expected: "",
		},
		{
			name:     "dollar sign without braces gets expanded",
			input:    "$VAR_WITHOUT_BRACES",
			envVars:  map[string]string{"VAR_WITHOUT_BRACES": "value"},
			expected: "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables with proper cleanup
			var envKeys []string
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value) // Ignore error
				envKeys = append(envKeys, key)
			}
			defer func() {
				for _, key := range envKeys {
					_ = os.Unsetenv(key) // Ignore error
				}
			}()

			result := ExpandEnvironmentVariables(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseYAML_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		expectError bool
	}{
		{
			name: "completely invalid YAML",
			yaml: `
version: "1.0.0"
providers:
  github:
    token: "unclosed string
`,
			expectError: true,
		},
		{
			name: "missing required version field",
			yaml: `
providers:
  github:
    token: "test-token"
`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.yaml)
			config, err := ParseYAML(reader)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
			}
		})
	}
}
