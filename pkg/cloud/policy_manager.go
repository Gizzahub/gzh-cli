package cloud

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

// PolicyManager manages network policy automation
type PolicyManager interface {
	// ApplyEnvironmentPolicies applies policies for a specific environment
	ApplyEnvironmentPolicies(ctx context.Context, environment string) error

	// ApplyPoliciesForProfile applies policies for a specific profile
	ApplyPoliciesForProfile(ctx context.Context, profileName string) error

	// GetApplicablePolicies returns policies applicable to a profile
	GetApplicablePolicies(ctx context.Context, profileName string) ([]*NetworkPolicy, error)

	// ValidatePolicy validates a network policy
	ValidatePolicy(policy *NetworkPolicy) error

	// GetPolicyStatus returns status of applied policies
	GetPolicyStatus(ctx context.Context) ([]PolicyStatus, error)

	// UpdatePolicyPriority updates policy priority
	UpdatePolicyPriority(policyName string, priority int) error
}

// DefaultPolicyManager implements PolicyManager interface
type DefaultPolicyManager struct {
	config    *Config
	providers map[string]Provider
}

// PolicyStatus represents the status of a policy application
type PolicyStatus struct {
	PolicyName  string    `json:"policy_name"`
	ProfileName string    `json:"profile_name"`
	Provider    string    `json:"provider"`
	Status      string    `json:"status"` // applied, failed, pending
	Applied     time.Time `json:"applied"`
	Error       string    `json:"error,omitempty"`
}

// NewPolicyManager creates a new policy manager
func NewPolicyManager(config *Config) PolicyManager {
	return &DefaultPolicyManager{
		config:    config,
		providers: make(map[string]Provider),
	}
}

// ApplyEnvironmentPolicies applies policies for a specific environment
func (pm *DefaultPolicyManager) ApplyEnvironmentPolicies(ctx context.Context, environment string) error {
	// Get all profiles for the environment
	profiles := pm.getProfilesForEnvironment(environment)
	if len(profiles) == 0 {
		return fmt.Errorf("no profiles found for environment: %s", environment)
	}

	var allErrors []string
	successCount := 0

	for _, profile := range profiles {
		if err := pm.ApplyPoliciesForProfile(ctx, profile.Name); err != nil {
			allErrors = append(allErrors, fmt.Sprintf("profile %s: %v", profile.Name, err))
		} else {
			successCount++
		}
	}

	if len(allErrors) > 0 {
		return fmt.Errorf("failed to apply policies for %d profiles: %s", len(allErrors), strings.Join(allErrors, "; "))
	}

	fmt.Printf("Successfully applied policies for %d profiles in environment %s\n", successCount, environment)
	return nil
}

// ApplyPoliciesForProfile applies policies for a specific profile
func (pm *DefaultPolicyManager) ApplyPoliciesForProfile(ctx context.Context, profileName string) error {
	// Get profile
	profile, exists := pm.config.GetProfile(profileName)
	if !exists {
		return fmt.Errorf("profile not found: %s", profileName)
	}

	// Get provider
	provider, err := pm.getProvider(profile.Provider)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Get applicable policies
	policies, err := pm.GetApplicablePolicies(ctx, profileName)
	if err != nil {
		return fmt.Errorf("failed to get applicable policies: %w", err)
	}

	if len(policies) == 0 {
		fmt.Printf("No policies found for profile: %s\n", profileName)
		return nil
	}

	// Sort policies by priority (higher priority first)
	sort.Slice(policies, func(i, j int) bool {
		return policies[i].Priority > policies[j].Priority
	})

	// Apply policies in order
	for _, policy := range policies {
		if !policy.Enabled {
			continue
		}

		fmt.Printf("Applying policy %s to profile %s...\n", policy.Name, profileName)

		if err := pm.ValidatePolicy(policy); err != nil {
			fmt.Printf("  ✗ Policy validation failed: %v\n", err)
			continue
		}

		if err := provider.ApplyNetworkPolicy(ctx, policy); err != nil {
			fmt.Printf("  ✗ Policy application failed: %v\n", err)
			continue
		}

		fmt.Printf("  ✓ Policy applied successfully\n")
	}

	return nil
}

// GetApplicablePolicies returns policies applicable to a profile
func (pm *DefaultPolicyManager) GetApplicablePolicies(ctx context.Context, profileName string) ([]*NetworkPolicy, error) {
	// Get profile
	profile, exists := pm.config.GetProfile(profileName)
	if !exists {
		return nil, fmt.Errorf("profile not found: %s", profileName)
	}

	// Get provider
	provider, err := pm.getProvider(profile.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	var policies []*NetworkPolicy

	// Get provider-specific policies
	providerPolicy, err := provider.GetNetworkPolicy(ctx, profileName)
	if err == nil && providerPolicy != nil {
		policies = append(policies, providerPolicy)
	}

	// Get environment-specific policies from config
	if pm.config.Policies != nil {
		for _, policy := range pm.config.Policies {
			if pm.isPolicyApplicable(&policy, &profile) {
				policies = append(policies, &policy)
			}
		}
	}

	// Get default policies
	defaultPolicies := pm.getDefaultPolicies(&profile)
	policies = append(policies, defaultPolicies...)

	return policies, nil
}

// ValidatePolicy validates a network policy
func (pm *DefaultPolicyManager) ValidatePolicy(policy *NetworkPolicy) error {
	if policy == nil {
		return fmt.Errorf("policy cannot be nil")
	}

	if policy.Name == "" {
		return fmt.Errorf("policy name is required")
	}

	if policy.ProfileName == "" {
		return fmt.Errorf("profile name is required")
	}

	// Validate rules
	for i, rule := range policy.Rules {
		if err := pm.validatePolicyRule(&rule); err != nil {
			return fmt.Errorf("rule %d: %w", i, err)
		}
	}

	// Validate actions
	for i, action := range policy.Actions {
		if err := pm.validatePolicyAction(&action); err != nil {
			return fmt.Errorf("action %d: %w", i, err)
		}
	}

	return nil
}

// GetPolicyStatus returns status of applied policies
func (pm *DefaultPolicyManager) GetPolicyStatus(ctx context.Context) ([]PolicyStatus, error) {
	var statuses []PolicyStatus

	// Iterate through all profiles
	for profileName, profile := range pm.config.Profiles {
		// Get applicable policies
		policies, err := pm.GetApplicablePolicies(ctx, profileName)
		if err != nil {
			continue
		}

		// Check status of each policy
		for _, policy := range policies {
			status := PolicyStatus{
				PolicyName:  policy.Name,
				ProfileName: profileName,
				Provider:    profile.Provider,
				Status:      "unknown",
				Applied:     time.Now(),
			}

			// Try to determine actual status
			if policy.Enabled {
				status.Status = "applied"
			} else {
				status.Status = "disabled"
			}

			statuses = append(statuses, status)
		}
	}

	return statuses, nil
}

// UpdatePolicyPriority updates policy priority
func (pm *DefaultPolicyManager) UpdatePolicyPriority(policyName string, priority int) error {
	if pm.config.Policies == nil {
		return fmt.Errorf("no policies configured")
	}

	for i := range pm.config.Policies {
		if pm.config.Policies[i].Name == policyName {
			pm.config.Policies[i].Priority = priority
			return nil
		}
	}

	return fmt.Errorf("policy not found: %s", policyName)
}

// Helper methods

func (pm *DefaultPolicyManager) getProfilesForEnvironment(environment string) []Profile {
	var profiles []Profile
	for _, profile := range pm.config.Profiles {
		if profile.Environment == environment {
			profiles = append(profiles, profile)
		}
	}
	return profiles
}

func (pm *DefaultPolicyManager) getProvider(providerName string) (Provider, error) {
	if provider, exists := pm.providers[providerName]; exists {
		return provider, nil
	}

	// Get provider config
	providerConfig, exists := pm.config.GetProvider(providerName)
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", providerName)
	}

	// Create provider instance
	provider, err := NewProvider(context.Background(), providerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	// Cache provider
	pm.providers[providerName] = provider

	return provider, nil
}

func (pm *DefaultPolicyManager) isPolicyApplicable(policy *NetworkPolicy, profile *Profile) bool {
	// Check if policy is for this profile
	if policy.ProfileName != "" && policy.ProfileName != profile.Name {
		return false
	}

	// Check if policy is for this environment
	if policy.Environment != "" && policy.Environment != profile.Environment {
		return false
	}

	// Check if policy is for this provider
	if policy.Provider != "" && policy.Provider != profile.Provider {
		return false
	}

	return true
}

func (pm *DefaultPolicyManager) getDefaultPolicies(profile *Profile) []*NetworkPolicy {
	var policies []*NetworkPolicy

	// Default security policy
	securityPolicy := &NetworkPolicy{
		Name:        fmt.Sprintf("%s-security", profile.Name),
		ProfileName: profile.Name,
		Priority:    50,
		Enabled:     true,
		Rules: []PolicyRule{
			{
				Type:        "deny",
				Source:      "0.0.0.0/0",
				Destination: profile.Network.VPCId,
				Protocol:    "tcp",
				Port:        "22",
			},
			{
				Type:        "allow",
				Source:      "10.0.0.0/8",
				Destination: profile.Network.VPCId,
				Protocol:    "tcp",
				Port:        "443",
			},
		},
	}

	// Default DNS policy
	if len(profile.Network.DNSServers) > 0 {
		dnsPolicy := &NetworkPolicy{
			Name:        fmt.Sprintf("%s-dns", profile.Name),
			ProfileName: profile.Name,
			Priority:    100,
			Enabled:     true,
			Actions: []PolicyAction{
				{
					Type: "configure_dns",
					Params: map[string]string{
						"servers": strings.Join(profile.Network.DNSServers, ","),
					},
					Order: 1,
				},
			},
		}
		policies = append(policies, dnsPolicy)
	}

	// Default proxy policy
	if profile.Network.Proxy != nil {
		proxyPolicy := &NetworkPolicy{
			Name:        fmt.Sprintf("%s-proxy", profile.Name),
			ProfileName: profile.Name,
			Priority:    75,
			Enabled:     true,
			Actions: []PolicyAction{
				{
					Type: "setup_proxy",
					Params: map[string]string{
						"http_proxy":  profile.Network.Proxy.HTTP,
						"https_proxy": profile.Network.Proxy.HTTPS,
					},
					Order: 1,
				},
			},
		}
		policies = append(policies, proxyPolicy)
	}

	// Default VPN policy
	if profile.Network.VPN != nil && profile.Network.VPN.AutoConnect {
		vpnPolicy := &NetworkPolicy{
			Name:        fmt.Sprintf("%s-vpn", profile.Name),
			ProfileName: profile.Name,
			Priority:    90,
			Enabled:     true,
			Actions: []PolicyAction{
				{
					Type: "connect_vpn",
					Params: map[string]string{
						"server": profile.Network.VPN.Server,
						"type":   profile.Network.VPN.Type,
					},
					Order: 1,
				},
			},
		}
		policies = append(policies, vpnPolicy)
	}

	policies = append(policies, securityPolicy)
	return policies
}

func (pm *DefaultPolicyManager) validatePolicyRule(rule *PolicyRule) error {
	if rule.Type == "" {
		return fmt.Errorf("rule type is required")
	}

	validTypes := []string{"allow", "deny", "redirect"}
	valid := false
	for _, validType := range validTypes {
		if rule.Type == validType {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid rule type: %s", rule.Type)
	}

	if rule.Protocol != "" {
		validProtocols := []string{"tcp", "udp", "icmp", "any"}
		valid = false
		for _, validProtocol := range validProtocols {
			if rule.Protocol == validProtocol {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid protocol: %s", rule.Protocol)
		}
	}

	return nil
}

func (pm *DefaultPolicyManager) validatePolicyAction(action *PolicyAction) error {
	if action.Type == "" {
		return fmt.Errorf("action type is required")
	}

	validTypes := []string{"configure_dns", "setup_proxy", "connect_vpn", "add_route", "configure_firewall"}
	valid := false
	for _, validType := range validTypes {
		if action.Type == validType {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid action type: %s", action.Type)
	}

	// Validate action-specific parameters
	switch action.Type {
	case "configure_dns":
		if _, ok := action.Params["servers"]; !ok {
			return fmt.Errorf("dns action requires 'servers' parameter")
		}
	case "setup_proxy":
		if _, ok := action.Params["http_proxy"]; !ok {
			return fmt.Errorf("proxy action requires 'http_proxy' parameter")
		}
	case "connect_vpn":
		if _, ok := action.Params["server"]; !ok {
			return fmt.Errorf("vpn action requires 'server' parameter")
		}
	case "add_route":
		if _, ok := action.Params["destination"]; !ok {
			return fmt.Errorf("route action requires 'destination' parameter")
		}
		if _, ok := action.Params["gateway"]; !ok {
			return fmt.Errorf("route action requires 'gateway' parameter")
		}
	}

	return nil
}

// PolicyScheduler handles automatic policy application
type PolicyScheduler struct {
	policyManager PolicyManager
	config        *Config
}

// NewPolicyScheduler creates a new policy scheduler
func NewPolicyScheduler(policyManager PolicyManager, config *Config) *PolicyScheduler {
	return &PolicyScheduler{
		policyManager: policyManager,
		config:        config,
	}
}

// StartScheduler starts the policy scheduler
func (ps *PolicyScheduler) StartScheduler(ctx context.Context) {
	// Check if scheduling is enabled
	if !ps.config.Sync.Enabled {
		return
	}

	interval := ps.config.Sync.Interval
	if interval == 0 {
		interval = 5 * time.Minute // Default interval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ps.applyScheduledPolicies(ctx)
		}
	}
}

// applyScheduledPolicies applies policies based on schedule
func (ps *PolicyScheduler) applyScheduledPolicies(ctx context.Context) {
	// Get all environments
	environments := ps.getEnvironments()

	for _, env := range environments {
		if err := ps.policyManager.ApplyEnvironmentPolicies(ctx, env); err != nil {
			fmt.Printf("Failed to apply policies for environment %s: %v\n", env, err)
		}
	}
}

// getEnvironments returns all unique environments
func (ps *PolicyScheduler) getEnvironments() []string {
	envSet := make(map[string]bool)
	for _, profile := range ps.config.Profiles {
		if profile.Environment != "" {
			envSet[profile.Environment] = true
		}
	}

	var environments []string
	for env := range envSet {
		environments = append(environments, env)
	}

	return environments
}
