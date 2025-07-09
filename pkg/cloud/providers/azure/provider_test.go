package azure

import (
	"context"
	"os"
	"testing"

	"github.com/gizzahub/gzh-manager-go/pkg/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderRegistration(t *testing.T) {
	// Test that Azure provider is registered
	providers := cloud.GetSupportedProviders()
	hasAzure := false
	for _, p := range providers {
		if p == cloud.ProviderTypeAzure {
			hasAzure = true
			break
		}
	}
	assert.True(t, hasAzure, "Azure provider should be registered")
	assert.True(t, cloud.IsProviderSupported("azure"))
}

func TestProviderName(t *testing.T) {
	p := &Provider{}
	assert.Equal(t, "azure", p.Name())
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      cloud.ProviderConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid service principal config",
			config: cloud.ProviderConfig{
				Type:   "azure",
				Region: "eastus",
				Auth: cloud.AuthConfig{
					Method: "service_principal",
					Params: map[string]string{
						"client_id":       "12345678-1234-1234-1234-123456789012",
						"client_secret":   "secret-value",
						"tenant_id":       "12345678-1234-1234-1234-123456789012",
						"subscription_id": "12345678-1234-1234-1234-123456789012",
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid default config",
			config: cloud.ProviderConfig{
				Type:   "azure",
				Region: "westus2",
				Auth: cloud.AuthConfig{
					Method: "default",
					Params: map[string]string{
						"subscription_id": "12345678-1234-1234-1234-123456789012",
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid CLI config",
			config: cloud.ProviderConfig{
				Type:   "azure",
				Region: "northeurope",
				Auth: cloud.AuthConfig{
					Method: "cli",
					Params: map[string]string{
						"subscription_id": "12345678-1234-1234-1234-123456789012",
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid managed identity config",
			config: cloud.ProviderConfig{
				Type:   "azure",
				Region: "eastasia",
				Auth: cloud.AuthConfig{
					Method: "managed_identity",
					Params: map[string]string{
						"subscription_id": "12345678-1234-1234-1234-123456789012",
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid provider type",
			config: cloud.ProviderConfig{
				Type:   "aws",
				Region: "eastus",
				Auth: cloud.AuthConfig{
					Method: "default",
					Params: map[string]string{
						"subscription_id": "12345678-1234-1234-1234-123456789012",
					},
				},
			},
			expectError: true,
			errorMsg:    "invalid provider type",
		},
		{
			name: "missing region",
			config: cloud.ProviderConfig{
				Type: "azure",
				Auth: cloud.AuthConfig{
					Method: "default",
					Params: map[string]string{
						"subscription_id": "12345678-1234-1234-1234-123456789012",
					},
				},
			},
			expectError: true,
			errorMsg:    "region is required",
		},
		{
			name: "invalid region",
			config: cloud.ProviderConfig{
				Type:   "azure",
				Region: "invalid-region",
				Auth: cloud.AuthConfig{
					Method: "default",
					Params: map[string]string{
						"subscription_id": "12345678-1234-1234-1234-123456789012",
					},
				},
			},
			expectError: true,
			errorMsg:    "invalid Azure region",
		},
		{
			name: "unsupported auth method",
			config: cloud.ProviderConfig{
				Type:   "azure",
				Region: "eastus",
				Auth: cloud.AuthConfig{
					Method: "oauth",
					Params: map[string]string{
						"subscription_id": "12345678-1234-1234-1234-123456789012",
					},
				},
			},
			expectError: true,
			errorMsg:    "unsupported auth method",
		},
		{
			name: "service principal missing client_id",
			config: cloud.ProviderConfig{
				Type:   "azure",
				Region: "eastus",
				Auth: cloud.AuthConfig{
					Method: "service_principal",
					Params: map[string]string{
						"client_secret":   "secret-value",
						"tenant_id":       "12345678-1234-1234-1234-123456789012",
						"subscription_id": "12345678-1234-1234-1234-123456789012",
					},
				},
			},
			expectError: true,
			errorMsg:    "client_id required",
		},
		{
			name: "service principal missing client_secret",
			config: cloud.ProviderConfig{
				Type:   "azure",
				Region: "eastus",
				Auth: cloud.AuthConfig{
					Method: "service_principal",
					Params: map[string]string{
						"client_id":       "12345678-1234-1234-1234-123456789012",
						"tenant_id":       "12345678-1234-1234-1234-123456789012",
						"subscription_id": "12345678-1234-1234-1234-123456789012",
					},
				},
			},
			expectError: true,
			errorMsg:    "client_secret required",
		},
		{
			name: "service principal missing tenant_id",
			config: cloud.ProviderConfig{
				Type:   "azure",
				Region: "eastus",
				Auth: cloud.AuthConfig{
					Method: "service_principal",
					Params: map[string]string{
						"client_id":       "12345678-1234-1234-1234-123456789012",
						"client_secret":   "secret-value",
						"subscription_id": "12345678-1234-1234-1234-123456789012",
					},
				},
			},
			expectError: true,
			errorMsg:    "tenant_id required",
		},
		{
			name: "missing subscription_id",
			config: cloud.ProviderConfig{
				Type:   "azure",
				Region: "eastus",
				Auth: cloud.AuthConfig{
					Method: "default",
				},
			},
			expectError: true,
			errorMsg:    "subscription_id required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env var for consistent testing
			originalSubscription := os.Getenv("AZURE_SUBSCRIPTION_ID")
			_ = os.Unsetenv("AZURE_SUBSCRIPTION_ID")
			defer func() {
				if originalSubscription != "" {
					_ = os.Setenv("AZURE_SUBSCRIPTION_ID", originalSubscription)
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

func TestValidateConfigWithEnvSubscription(t *testing.T) {
	// Test config validation with subscription ID from environment
	_ = os.Setenv("AZURE_SUBSCRIPTION_ID", "12345678-1234-1234-1234-123456789012")
	defer func() { _ = os.Unsetenv("AZURE_SUBSCRIPTION_ID") }()

	config := cloud.ProviderConfig{
		Type:   "azure",
		Region: "eastus",
		Auth: cloud.AuthConfig{
			Method: "default",
		},
	}

	p := &Provider{}
	err := p.ValidateConfig(config)
	assert.NoError(t, err)
}

func TestIsValidAzureRegion(t *testing.T) {
	tests := []struct {
		region string
		valid  bool
	}{
		{"eastus", true},
		{"westus2", true},
		{"northeurope", true},
		{"eastasia", true},
		{"invalid-region", false},
		{"", false},
		{"east-us", false},
		{"centralus", true},
		{"japaneast", true},
		{"australiaeast", true},
	}

	for _, tt := range tests {
		t.Run(tt.region, func(t *testing.T) {
			result := isValidAzureRegion(tt.region)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestProfileOperations(t *testing.T) {
	// This is a unit test that doesn't require real Azure connection
	p := &Provider{
		config: cloud.ProviderConfig{
			Type:   "azure",
			Region: "eastus",
		},
		subscriptionID:    "12345678-1234-1234-1234-123456789012",
		resourceGroupName: "test-rg",
		// Not initializing clients, so GetProfile will fail
	}

	ctx := context.Background()

	// Test GetProfile (will fail without real Azure connection)
	_, err := p.GetProfile(ctx, "test-profile")
	assert.Error(t, err) // Expected to fail without Azure client

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

	// Test with valid config (will fail without real Azure credentials)
	config := cloud.ProviderConfig{
		Type:   "azure",
		Region: "eastus",
		Auth: cloud.AuthConfig{
			Method: "default",
			Params: map[string]string{
				"subscription_id": "12345678-1234-1234-1234-123456789012",
			},
		},
	}

	_, err := NewProvider(ctx, config)
	// Expected to fail in test environment
	if err != nil {
		t.Logf("Expected error in test environment: %v", err)
	}
}

// Integration tests that require Azure credentials
func TestAzureIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Azure integration test in short mode")
	}

	// Check for Azure credentials
	if _, ok := os.LookupEnv("AZURE_SUBSCRIPTION_ID"); !ok {
		t.Skip("AZURE_SUBSCRIPTION_ID not set, skipping integration test")
	}

	ctx := context.Background()

	config := cloud.ProviderConfig{
		Type:   "azure",
		Region: "eastus",
		Auth: cloud.AuthConfig{
			Method: "default",
			Params: map[string]string{
				"subscription_id": os.Getenv("AZURE_SUBSCRIPTION_ID"),
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
