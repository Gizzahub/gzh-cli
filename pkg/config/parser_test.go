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
				os.Setenv("TEST_TOKEN", "test-token-value")
			},
			cleanup: func() {
				os.Unsetenv("TEST_TOKEN")
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
				os.Setenv(tt.envVar, tt.envValue)
				defer os.Unsetenv(tt.envVar)
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
				os.Setenv(tt.envVar, tt.envValue)
				defer os.Unsetenv(tt.envVar)
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
	defer os.Remove(tmpFile.Name())
	
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	// Set environment variable
	os.Setenv("GITHUB_TOKEN", "test-token")
	defer os.Unsetenv("GITHUB_TOKEN")

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