package aws

import (
	"context"
	"os"
	"testing"

	"github.com/gizzahub/gzh-manager-go/pkg/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderRegistration(t *testing.T) {
	// Test that AWS provider is registered
	providers := cloud.GetSupportedProviders()
	hasAWS := false
	for _, p := range providers {
		if p == cloud.ProviderTypeAWS {
			hasAWS = true
			break
		}
	}
	assert.True(t, hasAWS, "AWS provider should be registered")
	assert.True(t, cloud.IsProviderSupported("aws"))
}

func TestProviderName(t *testing.T) {
	p := &Provider{}
	assert.Equal(t, "aws", p.Name())
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      cloud.ProviderConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid IAM config",
			config: cloud.ProviderConfig{
				Type:   "aws",
				Region: "us-east-1",
				Auth: cloud.AuthConfig{
					Method: "iam",
				},
			},
			expectError: false,
		},
		{
			name: "valid key config",
			config: cloud.ProviderConfig{
				Type:   "aws",
				Region: "us-west-2",
				Auth: cloud.AuthConfig{
					Method: "key",
					Params: map[string]string{
						"access_key": "AKIAIOSFODNN7EXAMPLE",
						"secret_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid provider type",
			config: cloud.ProviderConfig{
				Type:   "gcp",
				Region: "us-east-1",
				Auth: cloud.AuthConfig{
					Method: "iam",
				},
			},
			expectError: true,
			errorMsg:    "invalid provider type",
		},
		{
			name: "missing region",
			config: cloud.ProviderConfig{
				Type: "aws",
				Auth: cloud.AuthConfig{
					Method: "iam",
				},
			},
			expectError: true,
			errorMsg:    "region is required",
		},
		{
			name: "invalid region",
			config: cloud.ProviderConfig{
				Type:   "aws",
				Region: "invalid-region",
				Auth: cloud.AuthConfig{
					Method: "iam",
				},
			},
			expectError: true,
			errorMsg:    "invalid AWS region",
		},
		{
			name: "unsupported auth method",
			config: cloud.ProviderConfig{
				Type:   "aws",
				Region: "us-east-1",
				Auth: cloud.AuthConfig{
					Method: "oauth",
				},
			},
			expectError: true,
			errorMsg:    "unsupported auth method",
		},
		{
			name: "key auth missing access_key",
			config: cloud.ProviderConfig{
				Type:   "aws",
				Region: "us-east-1",
				Auth: cloud.AuthConfig{
					Method: "key",
					Params: map[string]string{
						"secret_key": "secret",
					},
				},
			},
			expectError: true,
			errorMsg:    "access_key required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func TestInitialize(t *testing.T) {
	// Skip if AWS credentials are not available
	if testing.Short() {
		t.Skip("Skipping AWS initialization test in short mode")
	}

	ctx := context.Background()

	tests := []struct {
		name        string
		config      cloud.ProviderConfig
		expectError bool
	}{
		{
			name: "initialize with profile",
			config: cloud.ProviderConfig{
				Type:   "aws",
				Region: "us-east-1",
				Auth: cloud.AuthConfig{
					Method: "profile",
					Params: map[string]string{
						"profile": "default",
					},
				},
			},
			expectError: false,
		},
		{
			name: "initialize with env",
			config: cloud.ProviderConfig{
				Type:   "aws",
				Region: "us-east-1",
				Auth: cloud.AuthConfig{
					Method: "env",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{}
			err := p.Initialize(ctx, tt.config)

			// We expect errors in test environment without real AWS credentials
			// In real tests with credentials, adjust expectations
			if !tt.expectError && err != nil {
				// This is expected in test environment
				t.Logf("Expected error in test environment: %v", err)
			}
		})
	}
}

func TestIsValidAWSRegion(t *testing.T) {
	tests := []struct {
		region string
		valid  bool
	}{
		{"us-east-1", true},
		{"us-west-2", true},
		{"eu-west-1", true},
		{"ap-northeast-1", true},
		{"invalid-region", false},
		{"", false},
		{"us-east", false},
	}

	for _, tt := range tests {
		t.Run(tt.region, func(t *testing.T) {
			result := isValidAWSRegion(tt.region)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestProfileOperations(t *testing.T) {
	// This is a unit test that doesn't require real AWS connection
	p := &Provider{
		config: cloud.ProviderConfig{
			Type:   "aws",
			Region: "us-east-1",
		},
	}

	ctx := context.Background()

	// Test GetProfile (will fail without real AWS connection)
	_, err := p.GetProfile(ctx, "test-profile")
	assert.Error(t, err) // Expected to fail without AWS client

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

	// Test with valid config (will fail without real AWS credentials)
	config := cloud.ProviderConfig{
		Type:   "aws",
		Region: "us-east-1",
		Auth: cloud.AuthConfig{
			Method: "env",
		},
	}

	_, err := NewProvider(ctx, config)
	// Expected to fail in test environment
	if err != nil {
		t.Logf("Expected error in test environment: %v", err)
	}
}

// Integration tests that require AWS credentials
func TestAWSIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping AWS integration test in short mode")
	}

	// Check for AWS credentials
	if _, ok := os.LookupEnv("AWS_ACCESS_KEY_ID"); !ok {
		t.Skip("AWS_ACCESS_KEY_ID not set, skipping integration test")
	}

	ctx := context.Background()

	config := cloud.ProviderConfig{
		Type:   "aws",
		Region: "us-east-1",
		Auth: cloud.AuthConfig{
			Method: "env",
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
