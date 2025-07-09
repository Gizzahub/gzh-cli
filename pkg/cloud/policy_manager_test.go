package cloud

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPolicyManager(t *testing.T) {
	config := &Config{
		Providers: map[string]ProviderConfig{
			"aws": {Type: "aws", Region: "us-east-1"},
		},
		Profiles: map[string]Profile{
			"test-profile": {
				Name:        "test-profile",
				Provider:    "aws",
				Environment: "dev",
				Network: NetworkConfig{
					VPCId:      "vpc-123",
					DNSServers: []string{"8.8.8.8"},
				},
			},
		},
	}

	manager := NewPolicyManager(config)
	assert.NotNil(t, manager)

	// Test type assertion
	defaultManager, ok := manager.(*DefaultPolicyManager)
	assert.True(t, ok)
	assert.Equal(t, config, defaultManager.config)
}

func TestApplyEnvironmentPolicies(t *testing.T) {
	config := &Config{
		Providers: map[string]ProviderConfig{
			"aws": {Type: "aws", Region: "us-east-1"},
		},
		Profiles: map[string]Profile{
			"dev-profile": {
				Name:        "dev-profile",
				Provider:    "aws",
				Environment: "dev",
				Network: NetworkConfig{
					VPCId:      "vpc-123",
					DNSServers: []string{"8.8.8.8"},
				},
			},
			"prod-profile": {
				Name:        "prod-profile",
				Provider:    "aws",
				Environment: "prod",
				Network: NetworkConfig{
					VPCId:      "vpc-456",
					DNSServers: []string{"1.1.1.1"},
				},
			},
		},
	}

	manager := NewPolicyManager(config)
	ctx := context.Background()

	// Test applying policies for dev environment
	err := manager.ApplyEnvironmentPolicies(ctx, "dev")
	// Should not error but may not find provider implementation
	assert.Error(t, err) // Expected due to missing provider implementation

	// Test non-existent environment
	err = manager.ApplyEnvironmentPolicies(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no profiles found for environment")
}

func TestApplyPoliciesForProfile(t *testing.T) {
	config := &Config{
		Providers: map[string]ProviderConfig{
			"aws": {Type: "aws", Region: "us-east-1"},
		},
		Profiles: map[string]Profile{
			"test-profile": {
				Name:        "test-profile",
				Provider:    "aws",
				Environment: "dev",
				Network: NetworkConfig{
					VPCId:      "vpc-123",
					DNSServers: []string{"8.8.8.8"},
				},
			},
		},
	}

	manager := NewPolicyManager(config)
	ctx := context.Background()

	// Test applying policies for existing profile
	err := manager.ApplyPoliciesForProfile(ctx, "test-profile")
	// Should error due to missing provider implementation
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get provider")

	// Test non-existent profile
	err = manager.ApplyPoliciesForProfile(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "profile not found")
}

func TestGetApplicablePolicies(t *testing.T) {
	config := &Config{
		Providers: map[string]ProviderConfig{
			"aws": {Type: "aws", Region: "us-east-1"},
		},
		Profiles: map[string]Profile{
			"test-profile": {
				Name:        "test-profile",
				Provider:    "aws",
				Environment: "dev",
				Network: NetworkConfig{
					VPCId:      "vpc-123",
					DNSServers: []string{"8.8.8.8"},
					Proxy: &ProxyConfig{
						HTTP:  "http://proxy.example.com:8080",
						HTTPS: "https://proxy.example.com:8080",
					},
					VPN: &VPNConfig{
						Type:        "openvpn",
						Server:      "vpn.example.com",
						AutoConnect: true,
					},
				},
			},
		},
		Policies: []NetworkPolicy{
			{
				Name:        "custom-policy",
				ProfileName: "test-profile",
				Environment: "dev",
				Enabled:     true,
				Priority:    200,
				Rules: []PolicyRule{
					{
						Type:        "allow",
						Source:      "10.0.0.0/8",
						Destination: "vpc-123",
						Protocol:    "tcp",
						Port:        "80",
					},
				},
			},
		},
	}

	manager := NewPolicyManager(config)
	ctx := context.Background()

	// Test getting applicable policies
	_, err := manager.GetApplicablePolicies(ctx, "test-profile")
	// Should error due to missing provider implementation
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get provider")

	// Test non-existent profile
	_, err = manager.GetApplicablePolicies(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "profile not found")

	// Test with a profile that has no provider configured
	config.Profiles["test-profile"] = Profile{
		Name:        "test-profile",
		Provider:    "non-existent",
		Environment: "dev",
	}
	_, err = manager.GetApplicablePolicies(ctx, "test-profile")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider not found")
}

func TestValidatePolicy(t *testing.T) {
	config := &Config{}
	manager := NewPolicyManager(config)

	// Test valid policy
	validPolicy := &NetworkPolicy{
		Name:        "test-policy",
		ProfileName: "test-profile",
		Priority:    100,
		Enabled:     true,
		Rules: []PolicyRule{
			{
				Type:        "allow",
				Source:      "10.0.0.0/8",
				Destination: "vpc-123",
				Protocol:    "tcp",
				Port:        "80",
			},
		},
		Actions: []PolicyAction{
			{
				Type: "configure_dns",
				Params: map[string]string{
					"servers": "8.8.8.8,8.8.4.4",
				},
				Order: 1,
			},
		},
	}

	err := manager.ValidatePolicy(validPolicy)
	assert.NoError(t, err)

	// Test nil policy
	err = manager.ValidatePolicy(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "policy cannot be nil")

	// Test policy with empty name
	invalidPolicy := &NetworkPolicy{
		ProfileName: "test-profile",
	}
	err = manager.ValidatePolicy(invalidPolicy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "policy name is required")

	// Test policy with empty profile name
	invalidPolicy = &NetworkPolicy{
		Name: "test-policy",
	}
	err = manager.ValidatePolicy(invalidPolicy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "profile name is required")

	// Test policy with invalid rule type
	invalidPolicy = &NetworkPolicy{
		Name:        "test-policy",
		ProfileName: "test-profile",
		Rules: []PolicyRule{
			{
				Type: "invalid-type",
			},
		},
	}
	err = manager.ValidatePolicy(invalidPolicy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid rule type")

	// Test policy with invalid action type
	invalidPolicy = &NetworkPolicy{
		Name:        "test-policy",
		ProfileName: "test-profile",
		Actions: []PolicyAction{
			{
				Type: "invalid-action",
			},
		},
	}
	err = manager.ValidatePolicy(invalidPolicy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid action type")
}

func TestValidatePolicyRule(t *testing.T) {
	config := &Config{}
	manager := NewPolicyManager(config).(*DefaultPolicyManager)

	// Test valid rule
	validRule := &PolicyRule{
		Type:        "allow",
		Source:      "10.0.0.0/8",
		Destination: "vpc-123",
		Protocol:    "tcp",
		Port:        "80",
	}

	err := manager.validatePolicyRule(validRule)
	assert.NoError(t, err)

	// Test rule with empty type
	invalidRule := &PolicyRule{
		Source: "10.0.0.0/8",
	}
	err = manager.validatePolicyRule(invalidRule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule type is required")

	// Test rule with invalid type
	invalidRule = &PolicyRule{
		Type: "invalid-type",
	}
	err = manager.validatePolicyRule(invalidRule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid rule type")

	// Test rule with invalid protocol
	invalidRule = &PolicyRule{
		Type:     "allow",
		Protocol: "invalid-protocol",
	}
	err = manager.validatePolicyRule(invalidRule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid protocol")
}

func TestValidatePolicyAction(t *testing.T) {
	config := &Config{}
	manager := NewPolicyManager(config).(*DefaultPolicyManager)

	// Test valid DNS action
	validAction := &PolicyAction{
		Type: "configure_dns",
		Params: map[string]string{
			"servers": "8.8.8.8,8.8.4.4",
		},
		Order: 1,
	}

	err := manager.validatePolicyAction(validAction)
	assert.NoError(t, err)

	// Test action with empty type
	invalidAction := &PolicyAction{
		Params: map[string]string{
			"servers": "8.8.8.8",
		},
	}
	err = manager.validatePolicyAction(invalidAction)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "action type is required")

	// Test action with invalid type
	invalidAction = &PolicyAction{
		Type: "invalid-action",
	}
	err = manager.validatePolicyAction(invalidAction)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid action type")

	// Test DNS action without servers parameter
	invalidAction = &PolicyAction{
		Type:   "configure_dns",
		Params: map[string]string{},
	}
	err = manager.validatePolicyAction(invalidAction)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dns action requires 'servers' parameter")

	// Test proxy action without http_proxy parameter
	invalidAction = &PolicyAction{
		Type:   "setup_proxy",
		Params: map[string]string{},
	}
	err = manager.validatePolicyAction(invalidAction)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "proxy action requires 'http_proxy' parameter")

	// Test VPN action without server parameter
	invalidAction = &PolicyAction{
		Type:   "connect_vpn",
		Params: map[string]string{},
	}
	err = manager.validatePolicyAction(invalidAction)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "vpn action requires 'server' parameter")

	// Test route action without destination parameter
	invalidAction = &PolicyAction{
		Type: "add_route",
		Params: map[string]string{
			"gateway": "10.0.0.1",
		},
	}
	err = manager.validatePolicyAction(invalidAction)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "route action requires 'destination' parameter")

	// Test route action without gateway parameter
	invalidAction = &PolicyAction{
		Type: "add_route",
		Params: map[string]string{
			"destination": "192.168.1.0/24",
		},
	}
	err = manager.validatePolicyAction(invalidAction)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "route action requires 'gateway' parameter")
}

func TestGetPolicyStatus(t *testing.T) {
	config := &Config{
		Providers: map[string]ProviderConfig{
			"aws": {Type: "aws", Region: "us-east-1"},
		},
		Profiles: map[string]Profile{
			"test-profile": {
				Name:        "test-profile",
				Provider:    "aws",
				Environment: "dev",
				Network: NetworkConfig{
					VPCId:      "vpc-123",
					DNSServers: []string{"8.8.8.8"},
				},
			},
		},
	}

	manager := NewPolicyManager(config)
	ctx := context.Background()

	// Test getting policy status
	status, err := manager.GetPolicyStatus(ctx)
	// Should not error but may have empty status due to missing provider
	assert.NoError(t, err)
	assert.IsType(t, []PolicyStatus{}, status)
}

func TestUpdatePolicyPriority(t *testing.T) {
	config := &Config{
		Policies: []NetworkPolicy{
			{
				Name:        "test-policy",
				ProfileName: "test-profile",
				Priority:    100,
				Enabled:     true,
			},
		},
	}

	manager := NewPolicyManager(config)

	// Test updating existing policy priority
	err := manager.UpdatePolicyPriority("test-policy", 200)
	assert.NoError(t, err)
	assert.Equal(t, 200, config.Policies[0].Priority)

	// Test updating non-existent policy
	err = manager.UpdatePolicyPriority("non-existent", 150)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "policy not found")

	// Test with no policies configured
	config.Policies = nil
	err = manager.UpdatePolicyPriority("test-policy", 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no policies configured")
}

func TestIsPolicyApplicable(t *testing.T) {
	config := &Config{}
	manager := NewPolicyManager(config).(*DefaultPolicyManager)

	profile := &Profile{
		Name:        "test-profile",
		Provider:    "aws",
		Environment: "dev",
	}

	// Test policy applicable to any profile
	policy := &NetworkPolicy{
		Name:        "general-policy",
		ProfileName: "",
		Environment: "",
		Provider:    "",
	}
	assert.True(t, manager.isPolicyApplicable(policy, profile))

	// Test policy specific to profile
	policy = &NetworkPolicy{
		Name:        "specific-policy",
		ProfileName: "test-profile",
	}
	assert.True(t, manager.isPolicyApplicable(policy, profile))

	// Test policy for different profile
	policy = &NetworkPolicy{
		Name:        "other-policy",
		ProfileName: "other-profile",
	}
	assert.False(t, manager.isPolicyApplicable(policy, profile))

	// Test policy specific to environment
	policy = &NetworkPolicy{
		Name:        "env-policy",
		Environment: "dev",
	}
	assert.True(t, manager.isPolicyApplicable(policy, profile))

	// Test policy for different environment
	policy = &NetworkPolicy{
		Name:        "prod-policy",
		Environment: "prod",
	}
	assert.False(t, manager.isPolicyApplicable(policy, profile))

	// Test policy specific to provider
	policy = &NetworkPolicy{
		Name:     "provider-policy",
		Provider: "aws",
	}
	assert.True(t, manager.isPolicyApplicable(policy, profile))

	// Test policy for different provider
	policy = &NetworkPolicy{
		Name:     "gcp-policy",
		Provider: "gcp",
	}
	assert.False(t, manager.isPolicyApplicable(policy, profile))
}

func TestGetDefaultPolicies(t *testing.T) {
	config := &Config{}
	manager := NewPolicyManager(config).(*DefaultPolicyManager)

	// Test profile with full network configuration
	profile := &Profile{
		Name:        "test-profile",
		Provider:    "aws",
		Environment: "dev",
		Network: NetworkConfig{
			VPCId:      "vpc-123",
			DNSServers: []string{"8.8.8.8", "8.8.4.4"},
			Proxy: &ProxyConfig{
				HTTP:  "http://proxy.example.com:8080",
				HTTPS: "https://proxy.example.com:8080",
			},
			VPN: &VPNConfig{
				Type:        "openvpn",
				Server:      "vpn.example.com",
				AutoConnect: true,
			},
		},
	}

	policies := manager.getDefaultPolicies(profile)

	// Should have 4 policies: security, DNS, proxy, and VPN
	assert.Len(t, policies, 4)

	// Check security policy
	securityPolicy := findPolicyByName(policies, "test-profile-security")
	assert.NotNil(t, securityPolicy)
	assert.Equal(t, 50, securityPolicy.Priority)
	assert.True(t, securityPolicy.Enabled)
	assert.Len(t, securityPolicy.Rules, 2)

	// Check DNS policy
	dnsPolicy := findPolicyByName(policies, "test-profile-dns")
	assert.NotNil(t, dnsPolicy)
	assert.Equal(t, 100, dnsPolicy.Priority)
	assert.True(t, dnsPolicy.Enabled)
	assert.Len(t, dnsPolicy.Actions, 1)
	assert.Equal(t, "configure_dns", dnsPolicy.Actions[0].Type)

	// Check proxy policy
	proxyPolicy := findPolicyByName(policies, "test-profile-proxy")
	assert.NotNil(t, proxyPolicy)
	assert.Equal(t, 75, proxyPolicy.Priority)
	assert.True(t, proxyPolicy.Enabled)
	assert.Len(t, proxyPolicy.Actions, 1)
	assert.Equal(t, "setup_proxy", proxyPolicy.Actions[0].Type)

	// Check VPN policy
	vpnPolicy := findPolicyByName(policies, "test-profile-vpn")
	assert.NotNil(t, vpnPolicy)
	assert.Equal(t, 90, vpnPolicy.Priority)
	assert.True(t, vpnPolicy.Enabled)
	assert.Len(t, vpnPolicy.Actions, 1)
	assert.Equal(t, "connect_vpn", vpnPolicy.Actions[0].Type)
}

func TestGetDefaultPolicies_Minimal(t *testing.T) {
	config := &Config{}
	manager := NewPolicyManager(config).(*DefaultPolicyManager)

	// Test profile with minimal network configuration
	profile := &Profile{
		Name:        "minimal-profile",
		Provider:    "aws",
		Environment: "dev",
		Network: NetworkConfig{
			VPCId: "vpc-123",
		},
	}

	policies := manager.getDefaultPolicies(profile)

	// Should have only 1 policy: security
	assert.Len(t, policies, 1)

	// Check security policy
	securityPolicy := findPolicyByName(policies, "minimal-profile-security")
	assert.NotNil(t, securityPolicy)
	assert.Equal(t, 50, securityPolicy.Priority)
	assert.True(t, securityPolicy.Enabled)
}

func TestNewPolicyScheduler(t *testing.T) {
	config := &Config{
		Sync: SyncConfig{
			Enabled:  true,
			Interval: 5 * time.Minute,
		},
	}

	manager := NewPolicyManager(config)
	scheduler := NewPolicyScheduler(manager, config)

	assert.NotNil(t, scheduler)
	assert.Equal(t, manager, scheduler.policyManager)
	assert.Equal(t, config, scheduler.config)
}

func TestPolicyScheduler_GetEnvironments(t *testing.T) {
	config := &Config{
		Profiles: map[string]Profile{
			"prod-profile": {
				Name:        "prod-profile",
				Environment: "prod",
			},
			"dev-profile": {
				Name:        "dev-profile",
				Environment: "dev",
			},
			"test-profile": {
				Name:        "test-profile",
				Environment: "dev", // Duplicate environment
			},
			"no-env-profile": {
				Name: "no-env-profile",
				// No environment set
			},
		},
	}

	scheduler := &PolicyScheduler{
		config: config,
	}

	environments := scheduler.getEnvironments()

	// Should have 2 unique environments: prod and dev
	assert.Len(t, environments, 2)
	assert.Contains(t, environments, "prod")
	assert.Contains(t, environments, "dev")
}

// Helper function to find policy by name
func findPolicyByName(policies []*NetworkPolicy, name string) *NetworkPolicy {
	for _, policy := range policies {
		if policy.Name == name {
			return policy
		}
	}
	return nil
}

func TestGetProfilesForEnvironment(t *testing.T) {
	config := &Config{
		Profiles: map[string]Profile{
			"prod-profile1": {
				Name:        "prod-profile1",
				Environment: "prod",
			},
			"prod-profile2": {
				Name:        "prod-profile2",
				Environment: "prod",
			},
			"dev-profile": {
				Name:        "dev-profile",
				Environment: "dev",
			},
		},
	}

	manager := NewPolicyManager(config).(*DefaultPolicyManager)
	profiles := manager.getProfilesForEnvironment("prod")

	assert.Len(t, profiles, 2)
	assert.Contains(t, []string{profiles[0].Name, profiles[1].Name}, "prod-profile1")
	assert.Contains(t, []string{profiles[0].Name, profiles[1].Name}, "prod-profile2")

	// Test non-existent environment
	profiles = manager.getProfilesForEnvironment("non-existent")
	assert.Len(t, profiles, 0)
}

func TestMockPolicyProvider(t *testing.T) {
	// Create a mock provider for testing policy functionality
	mockProvider := &MockPolicyProvider{
		policies: map[string]*NetworkPolicy{
			"test-profile": {
				Name:        "provider-policy",
				ProfileName: "test-profile",
				Priority:    150,
				Enabled:     true,
				Rules: []PolicyRule{
					{
						Type:        "allow",
						Source:      "10.0.0.0/8",
						Destination: "vpc-123",
						Protocol:    "tcp",
						Port:        "443",
					},
				},
			},
		},
	}

	config := &Config{
		Providers: map[string]ProviderConfig{
			"mock": {Type: "mock", Region: "us-east-1"},
		},
		Profiles: map[string]Profile{
			"test-profile": {
				Name:        "test-profile",
				Provider:    "mock",
				Environment: "dev",
				Network: NetworkConfig{
					VPCId: "vpc-123",
				},
			},
		},
	}

	manager := NewPolicyManager(config).(*DefaultPolicyManager)
	manager.providers["mock"] = mockProvider

	ctx := context.Background()

	// Test getting applicable policies with mock provider
	policies, err := manager.GetApplicablePolicies(ctx, "test-profile")
	assert.NoError(t, err)
	assert.Len(t, policies, 2) // Provider policy + default security policy

	// Find provider policy
	var providerPolicy *NetworkPolicy
	for _, policy := range policies {
		if policy.Name == "provider-policy" {
			providerPolicy = policy
			break
		}
	}

	assert.NotNil(t, providerPolicy)
	assert.Equal(t, "provider-policy", providerPolicy.Name)
	assert.Equal(t, 150, providerPolicy.Priority)
}

// MockPolicyProvider for testing
type MockPolicyProvider struct {
	*MockProvider
	policies map[string]*NetworkPolicy
}

func (m *MockPolicyProvider) GetNetworkPolicy(ctx context.Context, profileName string) (*NetworkPolicy, error) {
	if policy, exists := m.policies[profileName]; exists {
		return policy, nil
	}
	return nil, fmt.Errorf("policy not found for profile: %s", profileName)
}

func (m *MockPolicyProvider) ApplyNetworkPolicy(ctx context.Context, policy *NetworkPolicy) error {
	// Mock implementation - just return success
	return nil
}
