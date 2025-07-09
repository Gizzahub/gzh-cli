package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/gizzahub/gzh-manager-go/pkg/cloud"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/proto"
)

func init() {
	// Register GCP provider
	cloud.Register(cloud.ProviderTypeGCP, NewProvider)
}

// Provider implements cloud.Provider for Google Cloud Platform
type Provider struct {
	config        cloud.ProviderConfig
	computeClient *compute.InstancesClient
	networkClient *compute.NetworksClient
	zonesClient   *compute.ZonesClient
	subnetsClient *compute.SubnetworksClient
	projectID     string
}

// NewProvider creates a new GCP provider instance
func NewProvider(ctx context.Context, cfg cloud.ProviderConfig) (cloud.Provider, error) {
	p := &Provider{
		config: cfg,
	}

	if err := p.Initialize(ctx, cfg); err != nil {
		return nil, err
	}

	return p, nil
}

// Name returns the provider name
func (p *Provider) Name() string {
	return string(cloud.ProviderTypeGCP)
}

// Initialize initializes the GCP provider
func (p *Provider) Initialize(ctx context.Context, cfg cloud.ProviderConfig) error {
	var opts []option.ClientOption

	// Configure authentication based on method
	switch cfg.Auth.Method {
	case "service_account":
		// Use service account key file
		if credsFile := cfg.Auth.CredentialsFile; credsFile != "" {
			opts = append(opts, option.WithCredentialsFile(credsFile))
		} else if credsJSON, ok := cfg.Auth.Params["credentials_json"]; ok {
			opts = append(opts, option.WithCredentialsJSON([]byte(credsJSON)))
		} else {
			return fmt.Errorf("credentials_file or credentials_json required for service_account method")
		}
	case "adc":
		// Use Application Default Credentials
		// No additional options needed
	case "metadata":
		// Use metadata server (for instances running on GCP)
		// No additional options needed
	default:
		return fmt.Errorf("unsupported auth method: %s", cfg.Auth.Method)
	}

	// Initialize compute client
	computeClient, err := compute.NewInstancesRESTClient(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to create compute client: %w", err)
	}
	p.computeClient = computeClient

	// Initialize network client
	networkClient, err := compute.NewNetworksRESTClient(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to create network client: %w", err)
	}
	p.networkClient = networkClient

	// Initialize zones client
	zonesClient, err := compute.NewZonesRESTClient(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to create zones client: %w", err)
	}
	p.zonesClient = zonesClient

	// Initialize subnets client
	subnetsClient, err := compute.NewSubnetworksRESTClient(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to create subnets client: %w", err)
	}
	p.subnetsClient = subnetsClient

	// Get project ID
	projectID, ok := cfg.Auth.Params["project_id"]
	if !ok {
		// Try to get from environment
		projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
		if projectID == "" {
			return fmt.Errorf("project_id required")
		}
	}
	p.projectID = projectID

	return nil
}

// GetProfile retrieves a specific profile configuration
func (p *Provider) GetProfile(ctx context.Context, profileName string) (*cloud.Profile, error) {
	profile := &cloud.Profile{
		Name:        profileName,
		Provider:    p.Name(),
		Environment: profileName,
		Region:      p.config.Region,
		LastSync:    time.Now(),
	}

	// Check if network client is initialized
	if p.networkClient == nil {
		return nil, fmt.Errorf("network client not initialized")
	}

	// Get VPC network information
	networkName := fmt.Sprintf("%s-network", profileName)
	networkReq := &computepb.GetNetworkRequest{
		Project: p.projectID,
		Network: networkName,
	}

	network, err := p.networkClient.Get(ctx, networkReq)
	if err == nil {
		profile.Network.VPCId = network.GetName()

		// Get subnets
		subnetReq := &computepb.ListSubnetworksRequest{
			Project: p.projectID,
			Region:  p.config.Region,
			Filter:  proto.String(fmt.Sprintf("network eq %s", network.GetSelfLink())),
		}

		subnets := p.subnetsClient.List(ctx, subnetReq)
		for {
			subnet, err := subnets.Next()
			if err != nil {
				break
			}
			profile.Network.SubnetIds = append(profile.Network.SubnetIds, subnet.GetName())
		}
	}

	// Set GCP-specific tags
	profile.Tags = map[string]string{
		"provider":   "gcp",
		"project_id": p.projectID,
		"region":     p.config.Region,
	}

	return profile, nil
}

// ListProfiles lists all available profiles
func (p *Provider) ListProfiles(ctx context.Context) ([]*cloud.Profile, error) {
	var profiles []*cloud.Profile

	// List networks (simplified - in production you'd filter by labels)
	req := &computepb.ListNetworksRequest{
		Project: p.projectID,
	}

	networks := p.networkClient.List(ctx, req)
	for {
		network, err := networks.Next()
		if err != nil {
			break
		}

		// Get profile name from network name (simplified for now)
		// In production, you'd use proper labels or tags
		profileName := network.GetName()
		if strings.Contains(profileName, "profile") {
			profile, err := p.GetProfile(ctx, profileName)
			if err == nil {
				profiles = append(profiles, profile)
			}
		}
	}

	return profiles, nil
}

// SyncProfile synchronizes a profile configuration
func (p *Provider) SyncProfile(ctx context.Context, profile *cloud.Profile) error {
	// In a real implementation, this would:
	// 1. Create/update VPC network if needed
	// 2. Configure subnets
	// 3. Set up firewall rules
	// 4. Configure Cloud NAT
	// 5. Set up VPN gateways

	// For now, we'll just validate the profile
	if profile.Name == "" {
		return fmt.Errorf("profile name is required")
	}

	if profile.Network.VPCId == "" {
		// Would create a new VPC network here
		return fmt.Errorf("VPC network creation not implemented")
	}

	return nil
}

// GetNetworkPolicy retrieves network policy for a profile
func (p *Provider) GetNetworkPolicy(ctx context.Context, profileName string) (*cloud.NetworkPolicy, error) {
	profile, err := p.GetProfile(ctx, profileName)
	if err != nil {
		return nil, err
	}

	policy := &cloud.NetworkPolicy{
		Name:        fmt.Sprintf("%s-policy", profileName),
		ProfileName: profileName,
		Priority:    100,
		Enabled:     true,
	}

	// Add DNS configuration action
	if len(profile.Network.DNSServers) > 0 {
		policy.Actions = append(policy.Actions, cloud.PolicyAction{
			Type: "configure_dns",
			Params: map[string]string{
				"servers": strings.Join(profile.Network.DNSServers, ","),
			},
			Order: 1,
		})
	}

	// Add proxy configuration if present
	if profile.Network.Proxy != nil {
		if profile.Network.Proxy.HTTP != "" {
			policy.Actions = append(policy.Actions, cloud.PolicyAction{
				Type: "setup_proxy",
				Params: map[string]string{
					"http_proxy":  profile.Network.Proxy.HTTP,
					"https_proxy": profile.Network.Proxy.HTTPS,
				},
				Order: 2,
			})
		}
	}

	// Add VPN connection if configured
	if profile.Network.VPN != nil && profile.Network.VPN.AutoConnect {
		policy.Actions = append(policy.Actions, cloud.PolicyAction{
			Type: "connect_vpn",
			Params: map[string]string{
				"server": profile.Network.VPN.Server,
				"type":   profile.Network.VPN.Type,
			},
			Order: 3,
		})
	}

	// Add firewall rules as policy rules
	policy.Rules = append(policy.Rules, cloud.PolicyRule{
		Type:        "allow",
		Source:      "10.0.0.0/8",
		Destination: profile.Network.VPCId,
		Protocol:    "tcp",
		Port:        "443",
	})

	return policy, nil
}

// ApplyNetworkPolicy applies network policy settings
func (p *Provider) ApplyNetworkPolicy(ctx context.Context, policy *cloud.NetworkPolicy) error {
	if !policy.Enabled {
		return nil
	}

	// Execute policy actions in order
	for _, action := range policy.Actions {
		switch action.Type {
		case "configure_dns":
			if servers, ok := action.Params["servers"]; ok {
				// Would configure DNS here
				fmt.Printf("Configuring DNS servers: %s\n", servers)
			}
		case "setup_proxy":
			if httpProxy, ok := action.Params["http_proxy"]; ok {
				_ = os.Setenv("HTTP_PROXY", httpProxy)
				_ = os.Setenv("http_proxy", httpProxy)
			}
			if httpsProxy, ok := action.Params["https_proxy"]; ok {
				_ = os.Setenv("HTTPS_PROXY", httpsProxy)
				_ = os.Setenv("https_proxy", httpsProxy)
			}
		case "connect_vpn":
			// Would connect to VPN here
			fmt.Printf("Connecting to VPN: %s\n", action.Params["server"])
		case "add_route":
			// Would add route here
			fmt.Printf("Adding route: %s via %s\n", action.Params["destination"], action.Params["gateway"])
		}
	}

	return nil
}

// ValidateConfig validates provider configuration
func (p *Provider) ValidateConfig(cfg cloud.ProviderConfig) error {
	if cfg.Type != string(cloud.ProviderTypeGCP) {
		return fmt.Errorf("invalid provider type: expected gcp, got %s", cfg.Type)
	}

	if cfg.Region == "" {
		return fmt.Errorf("region is required")
	}

	// Validate region format
	if !isValidGCPRegion(cfg.Region) {
		return fmt.Errorf("invalid GCP region: %s", cfg.Region)
	}

	// Validate auth method
	switch cfg.Auth.Method {
	case "service_account", "adc", "metadata":
		// Valid methods
	default:
		return fmt.Errorf("unsupported auth method: %s", cfg.Auth.Method)
	}

	// Validate auth params based on method
	if cfg.Auth.Method == "service_account" {
		if cfg.Auth.CredentialsFile == "" {
			if _, ok := cfg.Auth.Params["credentials_json"]; !ok {
				return fmt.Errorf("credentials_file or credentials_json required for service_account method")
			}
		}
	}

	// Validate project ID
	if cfg.Auth.Params["project_id"] == "" && os.Getenv("GOOGLE_CLOUD_PROJECT") == "" {
		return fmt.Errorf("project_id required")
	}

	return nil
}

// HealthCheck performs health check on provider connection
func (p *Provider) HealthCheck(ctx context.Context) error {
	// Check if zones client is initialized
	if p.zonesClient == nil {
		return fmt.Errorf("zones client not initialized")
	}

	// Try to list zones to verify connection
	req := &computepb.ListZonesRequest{
		Project: p.projectID,
	}

	zones := p.zonesClient.List(ctx, req)
	_, err := zones.Next()
	if err != nil {
		return fmt.Errorf("failed to access GCP Compute Engine: %w", err)
	}

	fmt.Printf("Connected to GCP project: %s\n", p.projectID)
	return nil
}

// isValidGCPRegion checks if the region is a valid GCP region
func isValidGCPRegion(region string) bool {
	// This is a simplified check - in production, you'd have a complete list
	validRegions := []string{
		"us-central1", "us-east1", "us-east4", "us-west1", "us-west2", "us-west3", "us-west4",
		"europe-north1", "europe-west1", "europe-west2", "europe-west3", "europe-west4", "europe-west6",
		"asia-east1", "asia-east2", "asia-northeast1", "asia-northeast2", "asia-northeast3",
		"asia-south1", "asia-southeast1", "asia-southeast2",
		"australia-southeast1", "southamerica-east1",
	}

	for _, valid := range validRegions {
		if region == valid {
			return true
		}
	}

	return false
}

// ServiceAccountKey represents a GCP service account key
type ServiceAccountKey struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
}

// LoadServiceAccountKey loads a service account key from file
func LoadServiceAccountKey(keyFile string) (*ServiceAccountKey, error) {
	data, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read service account key file: %w", err)
	}

	var key ServiceAccountKey
	if err := json.Unmarshal(data, &key); err != nil {
		return nil, fmt.Errorf("failed to parse service account key: %w", err)
	}

	return &key, nil
}
