package cloud_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/cloud"
	// Import cloud providers to register them
	_ "github.com/gizzahub/gzh-manager-go/pkg/cloud/providers/aws"
	_ "github.com/gizzahub/gzh-manager-go/pkg/cloud/providers/azure"
	_ "github.com/gizzahub/gzh-manager-go/pkg/cloud/providers/gcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSDKIntegration tests cloud SDK integration
func TestSDKIntegration(t *testing.T) {
	t.Run("AWS_SDK_Integration", func(t *testing.T) {
		testAWSSDKIntegration(t)
	})

	t.Run("GCP_SDK_Integration", func(t *testing.T) {
		testGCPSDKIntegration(t)
	})

	t.Run("Azure_SDK_Integration", func(t *testing.T) {
		testAzureSDKIntegration(t)
	})
}

func testAWSSDKIntegration(t *testing.T) {
	// Skip if no AWS credentials
	if !hasAWSCredentials() {
		t.Skip("AWS credentials not available, skipping AWS SDK integration test")
		return
	}

	config := cloud.ProviderConfig{
		Type:   "aws",
		Region: "us-east-1",
		Auth: cloud.AuthConfig{
			Method: "default",
		},
		Timeout: 30 * time.Second,
	}

	ctx := context.Background()
	provider, err := cloud.NewProvider(ctx, config)
	require.NoError(t, err)
	assert.NotNil(t, provider)

	// Test basic provider operations
	assert.Equal(t, "aws", provider.Name())

	// Test configuration validation
	err = provider.ValidateConfig(config)
	assert.NoError(t, err)

	// Test health check
	err = provider.HealthCheck(ctx)
	assert.NoError(t, err)

	// Test listing profiles (should work even if empty)
	profiles, err := provider.ListProfiles(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, profiles)
}

func testGCPSDKIntegration(t *testing.T) {
	// Skip if no GCP credentials
	if !hasGCPCredentials() {
		t.Skip("GCP credentials not available, skipping GCP SDK integration test")
		return
	}

	config := cloud.ProviderConfig{
		Type:   "gcp",
		Region: "us-central1",
		Auth: cloud.AuthConfig{
			Method: "default",
		},
		Settings: map[string]interface{}{
			"project_id": getGCPProjectID(),
		},
		Timeout: 30 * time.Second,
	}

	ctx := context.Background()
	provider, err := cloud.NewProvider(ctx, config)
	require.NoError(t, err)
	assert.NotNil(t, provider)

	// Test basic provider operations
	assert.Equal(t, "gcp", provider.Name())

	// Test configuration validation
	err = provider.ValidateConfig(config)
	assert.NoError(t, err)

	// Test health check
	err = provider.HealthCheck(ctx)
	assert.NoError(t, err)

	// Test listing profiles (should work even if empty)
	profiles, err := provider.ListProfiles(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, profiles)
}

func testAzureSDKIntegration(t *testing.T) {
	// Skip if no Azure credentials
	if !hasAzureCredentials() {
		t.Skip("Azure credentials not available, skipping Azure SDK integration test")
		return
	}

	config := cloud.ProviderConfig{
		Type:   "azure",
		Region: "eastus",
		Auth: cloud.AuthConfig{
			Method: "default",
		},
		Settings: map[string]interface{}{
			"subscription_id":     getAzureSubscriptionID(),
			"resource_group_name": getAzureResourceGroupName(),
		},
		Timeout: 30 * time.Second,
	}

	ctx := context.Background()
	provider, err := cloud.NewProvider(ctx, config)
	require.NoError(t, err)
	assert.NotNil(t, provider)

	// Test basic provider operations
	assert.Equal(t, "azure", provider.Name())

	// Test configuration validation
	err = provider.ValidateConfig(config)
	assert.NoError(t, err)

	// Test health check
	err = provider.HealthCheck(ctx)
	assert.NoError(t, err)

	// Test listing profiles (should work even if empty)
	profiles, err := provider.ListProfiles(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, profiles)
}

// TestProviderRegistration tests that all providers are properly registered
func TestProviderRegistration(t *testing.T) {
	// Test that all providers are registered
	registeredProviders := cloud.GetRegisteredProviders()

	expectedProviders := []string{"aws", "gcp", "azure"}
	for _, expected := range expectedProviders {
		assert.Contains(t, registeredProviders, expected, "Provider %s should be registered", expected)
	}
}

// TestMultiProviderSync tests synchronization between multiple providers
func TestMultiProviderSync(t *testing.T) {
	// Create test configuration
	config := &cloud.Config{
		Version: "1.0",
		Providers: map[string]cloud.ProviderConfig{
			"aws-test": {
				Type:   "aws",
				Region: "us-east-1",
				Auth: cloud.AuthConfig{
					Method: "default",
				},
			},
			"gcp-test": {
				Type:   "gcp",
				Region: "us-central1",
				Auth: cloud.AuthConfig{
					Method: "default",
				},
				Settings: map[string]interface{}{
					"project_id": "test-project",
				},
			},
		},
		Profiles: map[string]cloud.Profile{
			"test-profile": {
				Name:        "test-profile",
				Provider:    "aws-test",
				Environment: "test",
				Region:      "us-east-1",
				Network: cloud.NetworkConfig{
					VPCId:      "vpc-12345",
					DNSServers: []string{"8.8.8.8", "8.8.4.4"},
				},
			},
		},
	}

	// Test sync configuration validation
	err := cloud.ValidateSyncConfig(config)
	assert.NoError(t, err)

	// Create sync manager
	syncManager := cloud.NewSyncManager(config)
	assert.NotNil(t, syncManager)

	// Test sync status (should work even without actual providers)
	status, err := syncManager.GetSyncStatus(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, status)
}

// TestPolicyManagerIntegration tests policy manager integration with providers
func TestPolicyManagerIntegration(t *testing.T) {
	// Create test configuration
	config := &cloud.Config{
		Version: "1.0",
		Providers: map[string]cloud.ProviderConfig{
			"test-provider": {
				Type:   "aws",
				Region: "us-east-1",
				Auth: cloud.AuthConfig{
					Method: "env",
				},
			},
		},
		Profiles: map[string]cloud.Profile{
			"test-profile": {
				Name:        "test-profile",
				Provider:    "test-provider",
				Environment: "test",
				Region:      "us-east-1",
				Network: cloud.NetworkConfig{
					VPCId:      "vpc-12345",
					DNSServers: []string{"8.8.8.8", "8.8.4.4"},
				},
			},
		},
		Policies: []cloud.NetworkPolicy{
			{
				Name:        "test-policy",
				ProfileName: "test-profile",
				Environment: "test",
				Provider:    "test-provider",
				Priority:    100,
				Enabled:     true,
				Rules: []cloud.PolicyRule{
					{
						Type:        "allow",
						Source:      "10.0.0.0/8",
						Destination: "vpc-12345",
						Protocol:    "tcp",
						Port:        "443",
					},
				},
				Actions: []cloud.PolicyAction{
					{
						Type: "configure_dns",
						Params: map[string]string{
							"servers": "8.8.8.8,8.8.4.4",
						},
						Order: 1,
					},
				},
			},
		},
	}

	// Create policy manager
	policyManager := cloud.NewPolicyManager(config)
	assert.NotNil(t, policyManager)

	ctx := context.Background()

	// Test getting applicable policies
	policies, err := policyManager.GetApplicablePolicies(ctx, "test-profile")
	// May error due to missing provider implementation, but should not panic
	if err != nil {
		t.Logf("Expected error getting policies (provider not fully implemented): %v", err)
	} else {
		assert.NotNil(t, policies)
	}

	// Test policy validation
	for _, policy := range config.Policies {
		err := policyManager.ValidatePolicy(&policy)
		assert.NoError(t, err)
	}

	// Test policy status
	status, err := policyManager.GetPolicyStatus(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, status)
}

// TestVPNManagerIntegration tests VPN manager integration
func TestVPNManagerIntegration(t *testing.T) {
	// Create test configuration with VPN
	config := &cloud.Config{
		Version: "1.0",
		VPNs: map[string]cloud.VPNConnection{
			"test-vpn": {
				Name:        "test-vpn",
				Type:        "openvpn",
				Server:      "vpn.example.com",
				Port:        1194,
				Priority:    100,
				AutoConnect: true,
				HealthCheck: &cloud.VPNHealthCheck{
					Enabled:  true,
					Interval: 30 * time.Second,
					Timeout:  10 * time.Second,
					Targets:  []string{"8.8.8.8:53"},
				},
				Failover: &cloud.VPNFailover{
					Enabled:       true,
					RetryAttempts: 3,
					RetryInterval: 30 * time.Second,
				},
			},
		},
	}

	// Create VPN manager
	vpnManager := cloud.NewVPNManager()
	assert.NotNil(t, vpnManager)

	// Load VPN connections
	for _, vpn := range config.VPNs {
		err := vpnManager.AddVPNConnection(&vpn)
		assert.NoError(t, err)
	}

	// Test connection status
	status := vpnManager.GetConnectionStatus()
	assert.Len(t, status, 1)
	assert.Equal(t, cloud.VPNStateDisconnected, status["test-vpn"].State)

	// Test validation
	testVPN := config.VPNs["test-vpn"]
	err := vpnManager.ValidateConnection(&testVPN)
	assert.NoError(t, err)

	// Test failover monitoring (should not panic)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = vpnManager.StartFailoverMonitoring(ctx)
	assert.NoError(t, err)

	// Wait a bit then stop
	time.Sleep(100 * time.Millisecond)
	vpnManager.StopFailoverMonitoring()
}

// Helper functions for credential checking

func hasAWSCredentials() bool {
	// Check for AWS credentials in environment or config
	// This is a simplified check - in production you'd want more robust detection
	return getEnvOrDefault("AWS_ACCESS_KEY_ID", "") != "" ||
		getEnvOrDefault("AWS_PROFILE", "") != "" ||
		fileExists("~/.aws/credentials") ||
		fileExists("~/.aws/config")
}

func hasGCPCredentials() bool {
	// Check for GCP credentials
	return getEnvOrDefault("GOOGLE_APPLICATION_CREDENTIALS", "") != "" ||
		getEnvOrDefault("GOOGLE_CLOUD_PROJECT", "") != "" ||
		fileExists("~/.config/gcloud/application_default_credentials.json")
}

func hasAzureCredentials() bool {
	// Check for Azure credentials
	return getEnvOrDefault("AZURE_CLIENT_ID", "") != "" ||
		getEnvOrDefault("AZURE_CLIENT_SECRET", "") != "" ||
		getEnvOrDefault("AZURE_TENANT_ID", "") != "" ||
		fileExists("~/.azure/credentials")
}

func getGCPProjectID() string {
	return getEnvOrDefault("GOOGLE_CLOUD_PROJECT", "test-project")
}

func getAzureSubscriptionID() string {
	return getEnvOrDefault("AZURE_SUBSCRIPTION_ID", "00000000-0000-0000-0000-000000000000")
}

func getAzureResourceGroupName() string {
	return getEnvOrDefault("AZURE_RESOURCE_GROUP", "test-rg")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func fileExists(path string) bool {
	if strings.HasPrefix(path, "~") {
		home := os.Getenv("HOME")
		if home == "" {
			return false
		}
		path = strings.Replace(path, "~", home, 1)
	}

	_, err := os.Stat(path)
	return err == nil
}
