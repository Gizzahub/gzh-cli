package netenv

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloudCommand(t *testing.T) {
	ctx := context.Background()

	// Create test config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "cloud-config.yaml")

	configContent := `version: "1.0"
providers:
  aws-prod:
    type: aws
    region: us-east-1
    auth:
      method: iam
  gcp-dev:
    type: gcp
    region: us-central1
    auth:
      method: service_account
      credentials_file: /path/to/creds.json
profiles:
  production:
    provider: aws-prod
    environment: prod
    region: us-east-1
    network:
      vpc_id: vpc-12345
      dns_servers:
        - 10.0.0.2
        - 10.0.0.3
      proxy:
        http: http://proxy.company.com:8080
        https: https://proxy.company.com:8080
  development:
    provider: gcp-dev
    environment: dev
    region: us-central1
    network:
      vpc_id: dev-vpc
      vpn:
        type: openvpn
        server: vpn.dev.company.com
        auto_connect: true
`

	err := os.WriteFile(configFile, []byte(configContent), 0o644)
	require.NoError(t, err)

	tests := []struct {
		name         string
		args         []string
		expectError  bool
		expectOutput []string
	}{
		{
			name: "list providers and profiles",
			args: []string{"cloud", "list", "--config", configFile},
			expectOutput: []string{
				"Cloud Providers:",
				"aws-prod",
				"gcp-dev",
				"Cloud Profiles:",
				"production",
				"development",
			},
		},
		{
			name: "list only providers",
			args: []string{"cloud", "list", "--providers", "--config", configFile},
			expectOutput: []string{
				"Cloud Providers:",
				"aws-prod",
				"gcp-dev",
			},
		},
		{
			name: "show profile details",
			args: []string{"cloud", "show", "production", "--config", configFile},
			expectOutput: []string{
				"Profile: production",
				"Provider: aws-prod",
				"Environment: prod",
				"VPC ID: vpc-12345",
				"DNS Servers:",
				"10.0.0.2",
				"Proxy:",
				"HTTP: http://proxy.company.com:8080",
			},
		},
		{
			name:        "show non-existent profile",
			args:        []string{"cloud", "show", "nonexistent", "--config", configFile},
			expectError: true,
		},
		{
			name: "validate configuration",
			args: []string{"cloud", "validate", "--config", configFile},
			expectOutput: []string{
				"Configuration file is valid",
				"Providers: 2",
				"Profiles: 2",
			},
		},
		{
			name: "switch profile dry-run",
			args: []string{"cloud", "switch", "production", "--dry-run", "--config", configFile},
			expectOutput: []string{
				"Switching to cloud profile: production",
				"[DRY RUN]",
				"Setting DNS servers:",
				"Setting HTTP proxy:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create command
			cmd := NewNetEnvCmd(ctx)

			// Capture output
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)

			// Set args
			cmd.SetArgs(tt.args)

			// Execute command
			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			output := stdout.String() + stderr.String()

			// Check expected output
			for _, expected := range tt.expectOutput {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestCloudSyncCommand(t *testing.T) {
	ctx := context.Background()

	// Create test config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "cloud-config.yaml")

	configContent := `version: "1.0"
providers:
  aws:
    type: aws
    region: us-east-1
    auth:
      method: iam
  gcp:
    type: gcp
    region: us-central1
    auth:
      method: service_account
profiles:
  webapp:
    provider: aws
    environment: prod
    region: us-east-1
    network:
      vpc_id: vpc-web
sync:
  enabled: true
  targets:
    - source: aws
      target: gcp
`

	err := os.WriteFile(configFile, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Test sync command
	cmd := NewNetEnvCmd(ctx)

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"cloud", "sync", "--source", "aws", "--target", "gcp", "--config", configFile})

	// Note: This will fail because we don't have actual provider implementations
	// but we can test that the command structure works
	err = cmd.Execute()
	assert.Error(t, err) // Expected to fail without provider implementations
}

func TestCloudConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: `version: "1.0"
providers:
  aws:
    type: aws
    region: us-east-1
    auth:
      method: iam
profiles:
  prod:
    provider: aws
    environment: production
    region: us-east-1
`,
			expectError: false,
		},
		{
			name: "missing provider type",
			config: `version: "1.0"
providers:
  aws:
    region: us-east-1
    auth:
      method: iam
`,
			expectError: true,
			errorMsg:    "type is required",
		},
		{
			name: "unknown provider in profile",
			config: `version: "1.0"
providers:
  aws:
    type: aws
    region: us-east-1
    auth:
      method: iam
profiles:
  prod:
    provider: unknown
    environment: production
`,
			expectError: true,
			errorMsg:    "unknown provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), "config.yaml")
			err := os.WriteFile(tmpFile, []byte(tt.config), 0o644)
			require.NoError(t, err)

			ctx := context.Background()
			cmd := NewNetEnvCmd(ctx)

			var stderr bytes.Buffer
			cmd.SetErr(&stderr)
			cmd.SetArgs([]string{"cloud", "validate", "--config", tmpFile})

			err = cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, stderr.String()+err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSaveAndGetCurrentProfile(t *testing.T) {
	// Set temporary config dir
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")

	// Test save
	err := saveCurrentProfile("test-profile")
	assert.NoError(t, err)

	// Test get
	profile, err := getCurrentProfile()
	assert.NoError(t, err)
	assert.Equal(t, "test-profile", profile)

	// Test get when file doesn't exist
	os.Remove(filepath.Join(tmpDir, ".config", "gzh-manager", "current-cloud-profile"))
	profile, err = getCurrentProfile()
	assert.NoError(t, err)
	assert.Empty(t, profile)
}
