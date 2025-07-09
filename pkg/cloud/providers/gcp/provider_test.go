package gcp

import (
	"context"
	"os"
	"testing"

	"github.com/gizzahub/gzh-manager-go/pkg/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderRegistration(t *testing.T) {
	// Test that GCP provider is registered
	providers := cloud.GetSupportedProviders()
	hasGCP := false
	for _, p := range providers {
		if p == cloud.ProviderTypeGCP {
			hasGCP = true
			break
		}
	}
	assert.True(t, hasGCP, "GCP provider should be registered")
	assert.True(t, cloud.IsProviderSupported("gcp"))
}

func TestProviderName(t *testing.T) {
	p := &Provider{}
	assert.Equal(t, "gcp", p.Name())
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      cloud.ProviderConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid service account config",
			config: cloud.ProviderConfig{
				Type:   "gcp",
				Region: "us-central1",
				Auth: cloud.AuthConfig{
					Method:          "service_account",
					CredentialsFile: "/path/to/service-account.json",
					Params: map[string]string{
						"project_id": "my-project",
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid ADC config",
			config: cloud.ProviderConfig{
				Type:   "gcp",
				Region: "us-west1",
				Auth: cloud.AuthConfig{
					Method: "adc",
					Params: map[string]string{
						"project_id": "my-project",
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid metadata config",
			config: cloud.ProviderConfig{
				Type:   "gcp",
				Region: "europe-west1",
				Auth: cloud.AuthConfig{
					Method: "metadata",
					Params: map[string]string{
						"project_id": "my-project",
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid provider type",
			config: cloud.ProviderConfig{
				Type:   "aws",
				Region: "us-central1",
				Auth: cloud.AuthConfig{
					Method: "service_account",
					Params: map[string]string{
						"project_id": "my-project",
					},
				},
			},
			expectError: true,
			errorMsg:    "invalid provider type",
		},
		{
			name: "missing region",
			config: cloud.ProviderConfig{
				Type: "gcp",
				Auth: cloud.AuthConfig{
					Method: "service_account",
					Params: map[string]string{
						"project_id": "my-project",
					},
				},
			},
			expectError: true,
			errorMsg:    "region is required",
		},
		{
			name: "invalid region",
			config: cloud.ProviderConfig{
				Type:   "gcp",
				Region: "invalid-region",
				Auth: cloud.AuthConfig{
					Method: "service_account",
					Params: map[string]string{
						"project_id": "my-project",
					},
				},
			},
			expectError: true,
			errorMsg:    "invalid GCP region",
		},
		{
			name: "unsupported auth method",
			config: cloud.ProviderConfig{
				Type:   "gcp",
				Region: "us-central1",
				Auth: cloud.AuthConfig{
					Method: "oauth",
					Params: map[string]string{
						"project_id": "my-project",
					},
				},
			},
			expectError: true,
			errorMsg:    "unsupported auth method",
		},
		{
			name: "service account without credentials",
			config: cloud.ProviderConfig{
				Type:   "gcp",
				Region: "us-central1",
				Auth: cloud.AuthConfig{
					Method: "service_account",
					Params: map[string]string{
						"project_id": "my-project",
					},
				},
			},
			expectError: true,
			errorMsg:    "credentials_file or credentials_json required",
		},
		{
			name: "missing project_id",
			config: cloud.ProviderConfig{
				Type:   "gcp",
				Region: "us-central1",
				Auth: cloud.AuthConfig{
					Method:          "service_account",
					CredentialsFile: "/path/to/service-account.json",
				},
			},
			expectError: true,
			errorMsg:    "project_id required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env var for consistent testing
			originalProject := os.Getenv("GOOGLE_CLOUD_PROJECT")
			_ = os.Unsetenv("GOOGLE_CLOUD_PROJECT")
			defer func() {
				if originalProject != "" {
					_ = os.Setenv("GOOGLE_CLOUD_PROJECT", originalProject)
				}
			}()

			p := &Provider{}
			err := p.ValidateConfig(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateConfigWithEnvProject(t *testing.T) {
	// Test config validation with project ID from environment
	_ = os.Setenv("GOOGLE_CLOUD_PROJECT", "env-project")
	defer func() { _ = os.Unsetenv("GOOGLE_CLOUD_PROJECT") }()

	config := cloud.ProviderConfig{
		Type:   "gcp",
		Region: "us-central1",
		Auth: cloud.AuthConfig{
			Method:          "service_account",
			CredentialsFile: "/path/to/service-account.json",
		},
	}

	p := &Provider{}
	err := p.ValidateConfig(config)
	assert.NoError(t, err)
}

func TestIsValidGCPRegion(t *testing.T) {
	tests := []struct {
		region string
		valid  bool
	}{
		{"us-central1", true},
		{"us-east1", true},
		{"europe-west1", true},
		{"asia-northeast1", true},
		{"invalid-region", false},
		{"", false},
		{"us-central", false},
		{"asia-east1", true},
	}

	for _, tt := range tests {
		t.Run(tt.region, func(t *testing.T) {
			result := isValidGCPRegion(tt.region)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestProfileOperations(t *testing.T) {
	// This is a unit test that doesn't require real GCP connection
	p := &Provider{
		config: cloud.ProviderConfig{
			Type:   "gcp",
			Region: "us-central1",
		},
		projectID: "test-project",
		// Not initializing clients, so GetProfile will fail
	}

	ctx := context.Background()

	// Test GetProfile (will fail without real GCP connection)
	_, err := p.GetProfile(ctx, "test-profile")
	assert.Error(t, err) // Expected to fail without GCP client

	// Test ApplyNetworkPolicy
	policy := &cloud.NetworkPolicy{
		Name:        "test-policy",
		ProfileName: "test-profile",
		Enabled:     true,
		Actions: []cloud.PolicyAction{
			{
				Type: "setup_proxy",
				Params: map[string]string{
					"http_proxy":  "http://proxy.example.com:8080",
					"https_proxy": "https://proxy.example.com:8080",
				},
				Order: 1,
			},
		},
	}

	// This should work as it only sets environment variables
	err = p.ApplyNetworkPolicy(ctx, policy)
	assert.NoError(t, err)
}

func TestNewProvider(t *testing.T) {
	ctx := context.Background()

	// Test with valid config (will fail without real GCP credentials)
	config := cloud.ProviderConfig{
		Type:   "gcp",
		Region: "us-central1",
		Auth: cloud.AuthConfig{
			Method: "adc",
			Params: map[string]string{
				"project_id": "test-project",
			},
		},
	}

	_, err := NewProvider(ctx, config)
	// Expected to fail in test environment
	if err != nil {
		t.Logf("Expected error in test environment: %v", err)
	}
}

func TestLoadServiceAccountKey(t *testing.T) {
	// Create a temporary service account key file
	keyContent := `{
		"type": "service_account",
		"project_id": "test-project",
		"private_key_id": "key-id",
		"private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC...\n-----END PRIVATE KEY-----\n",
		"client_email": "test@test-project.iam.gserviceaccount.com",
		"client_id": "123456789",
		"auth_uri": "https://accounts.google.com/o/oauth2/auth",
		"token_uri": "https://oauth2.googleapis.com/token",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/test%40test-project.iam.gserviceaccount.com"
	}`

	tmpFile := t.TempDir() + "/service-account.json"
	err := os.WriteFile(tmpFile, []byte(keyContent), 0o600)
	require.NoError(t, err)

	// Test loading the key
	key, err := LoadServiceAccountKey(tmpFile)
	require.NoError(t, err)
	assert.Equal(t, "service_account", key.Type)
	assert.Equal(t, "test-project", key.ProjectID)
	assert.Equal(t, "test@test-project.iam.gserviceaccount.com", key.ClientEmail)

	// Test loading non-existent file
	_, err = LoadServiceAccountKey("/non/existent/file.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read service account key file")

	// Test loading invalid JSON
	invalidFile := t.TempDir() + "/invalid.json"
	err = os.WriteFile(invalidFile, []byte("invalid json"), 0o600)
	require.NoError(t, err)

	_, err = LoadServiceAccountKey(invalidFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse service account key")
}

// Integration tests that require GCP credentials
func TestGCPIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping GCP integration test in short mode")
	}

	// Check for GCP credentials
	if _, ok := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); !ok {
		t.Skip("GOOGLE_APPLICATION_CREDENTIALS not set, skipping integration test")
	}

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		t.Skip("GOOGLE_CLOUD_PROJECT not set, skipping integration test")
	}

	ctx := context.Background()

	config := cloud.ProviderConfig{
		Type:   "gcp",
		Region: "us-central1",
		Auth: cloud.AuthConfig{
			Method: "adc",
			Params: map[string]string{
				"project_id": projectID,
			},
		},
	}

	provider, err := NewProvider(ctx, config)
	require.NoError(t, err)

	// Test HealthCheck
	err = provider.HealthCheck(ctx)
	assert.NoError(t, err)

	// Test ListProfiles
	profiles, err := provider.ListProfiles(ctx)
	assert.NoError(t, err)
	t.Logf("Found %d profiles", len(profiles))
}
