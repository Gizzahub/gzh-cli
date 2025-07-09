package azure

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/gizzahub/gzh-manager-go/pkg/cloud"
)

func init() {
	// Register Azure provider
	cloud.Register(cloud.ProviderTypeAzure, NewProvider)
}

// Provider implements cloud.Provider for Microsoft Azure
type Provider struct {
	config            cloud.ProviderConfig
	credential        azcore.TokenCredential
	networkClient     *armnetwork.VirtualNetworksClient
	subnetClient      *armnetwork.SubnetsClient
	computeClient     *armcompute.VirtualMachinesClient
	subscriptionID    string
	resourceGroupName string
}

// NewProvider creates a new Azure provider instance
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
	return string(cloud.ProviderTypeAzure)
}

// Initialize initializes the Azure provider
func (p *Provider) Initialize(ctx context.Context, cfg cloud.ProviderConfig) error {
	var err error

	// Configure Azure authentication based on method
	switch cfg.Auth.Method {
	case "service_principal":
		// Use service principal credentials
		clientID, ok := cfg.Auth.Params["client_id"]
		if !ok {
			return fmt.Errorf("client_id required for service_principal method")
		}
		clientSecret, ok := cfg.Auth.Params["client_secret"]
		if !ok {
			return fmt.Errorf("client_secret required for service_principal method")
		}
		tenantID, ok := cfg.Auth.Params["tenant_id"]
		if !ok {
			return fmt.Errorf("tenant_id required for service_principal method")
		}

		p.credential, err = azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)
		if err != nil {
			return fmt.Errorf("failed to create service principal credential: %w", err)
		}

	case "default":
		// Use Azure Default Credential (managed identity, CLI, etc.)
		p.credential, err = azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return fmt.Errorf("failed to create default credential: %w", err)
		}

	case "cli":
		// Use Azure CLI credentials
		p.credential, err = azidentity.NewAzureCLICredential(nil)
		if err != nil {
			return fmt.Errorf("failed to create CLI credential: %w", err)
		}

	case "managed_identity":
		// Use managed identity
		p.credential, err = azidentity.NewManagedIdentityCredential(nil)
		if err != nil {
			return fmt.Errorf("failed to create managed identity credential: %w", err)
		}

	default:
		return fmt.Errorf("unsupported auth method: %s", cfg.Auth.Method)
	}

	// Get subscription ID
	subscriptionID, ok := cfg.Auth.Params["subscription_id"]
	if !ok {
		// Try to get from environment
		subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
		if subscriptionID == "" {
			return fmt.Errorf("subscription_id required")
		}
	}
	p.subscriptionID = subscriptionID

	// Get resource group name
	resourceGroupName, ok := cfg.Auth.Params["resource_group"]
	if !ok {
		resourceGroupName = "default"
	}
	p.resourceGroupName = resourceGroupName

	// Initialize Azure clients
	p.networkClient, err = armnetwork.NewVirtualNetworksClient(subscriptionID, p.credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create network client: %w", err)
	}

	p.subnetClient, err = armnetwork.NewSubnetsClient(subscriptionID, p.credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create subnet client: %w", err)
	}

	p.computeClient, err = armcompute.NewVirtualMachinesClient(subscriptionID, p.credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create compute client: %w", err)
	}

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

	// Get Virtual Network information
	vnetName := fmt.Sprintf("%s-vnet", profileName)
	vnet, err := p.networkClient.Get(ctx, p.resourceGroupName, vnetName, nil)
	if err == nil {
		profile.Network.VPCId = *vnet.Name

		// Get subnets
		subnets, err := p.subnetClient.NewListPager(p.resourceGroupName, vnetName, nil).NextPage(ctx)
		if err == nil {
			for _, subnet := range subnets.Value {
				profile.Network.SubnetIds = append(profile.Network.SubnetIds, *subnet.Name)
			}
		}

		// Get address spaces
		if vnet.Properties != nil && vnet.Properties.AddressSpace != nil {
			for _, addressPrefix := range vnet.Properties.AddressSpace.AddressPrefixes {
				profile.Network.CIDRBlocks = append(profile.Network.CIDRBlocks, *addressPrefix)
			}
		}
	}

	// Set Azure-specific tags
	profile.Tags = map[string]string{
		"provider":        "azure",
		"subscription_id": p.subscriptionID,
		"resource_group":  p.resourceGroupName,
		"region":          p.config.Region,
	}

	return profile, nil
}

// ListProfiles lists all available profiles
func (p *Provider) ListProfiles(ctx context.Context) ([]*cloud.Profile, error) {
	var profiles []*cloud.Profile

	// List Virtual Networks (simplified - in production you'd filter by tags)
	pager := p.networkClient.NewListPager(p.resourceGroupName, nil)
	for pager.More() {
		vnets, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list virtual networks: %w", err)
		}

		for _, vnet := range vnets.Value {
			// Get profile name from virtual network name (simplified)
			vnetName := *vnet.Name
			if strings.Contains(vnetName, "profile") || strings.Contains(vnetName, "vnet") {
				profileName := strings.TrimSuffix(vnetName, "-vnet")
				profile, err := p.GetProfile(ctx, profileName)
				if err == nil {
					profiles = append(profiles, profile)
				}
			}
		}
	}

	return profiles, nil
}

// SyncProfile synchronizes a profile configuration
func (p *Provider) SyncProfile(ctx context.Context, profile *cloud.Profile) error {
	// In a real implementation, this would:
	// 1. Create/update Virtual Network if needed
	// 2. Configure subnets
	// 3. Set up Network Security Groups
	// 4. Configure route tables
	// 5. Set up VPN gateways

	// For now, we'll just validate the profile
	if profile.Name == "" {
		return fmt.Errorf("profile name is required")
	}

	if profile.Network.VPCId == "" {
		// Would create a new Virtual Network here
		return fmt.Errorf("virtual network creation not implemented")
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

	// Add Network Security Group rules as policy rules
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
	if cfg.Type != string(cloud.ProviderTypeAzure) {
		return fmt.Errorf("invalid provider type: expected azure, got %s", cfg.Type)
	}

	if cfg.Region == "" {
		return fmt.Errorf("region is required")
	}

	// Validate region format
	if !isValidAzureRegion(cfg.Region) {
		return fmt.Errorf("invalid Azure region: %s", cfg.Region)
	}

	// Validate auth method
	switch cfg.Auth.Method {
	case "service_principal", "default", "cli", "managed_identity":
		// Valid methods
	default:
		return fmt.Errorf("unsupported auth method: %s", cfg.Auth.Method)
	}

	// Validate auth params based on method
	if cfg.Auth.Method == "service_principal" {
		if _, ok := cfg.Auth.Params["client_id"]; !ok {
			return fmt.Errorf("client_id required for service_principal method")
		}
		if _, ok := cfg.Auth.Params["client_secret"]; !ok {
			return fmt.Errorf("client_secret required for service_principal method")
		}
		if _, ok := cfg.Auth.Params["tenant_id"]; !ok {
			return fmt.Errorf("tenant_id required for service_principal method")
		}
	}

	// Validate subscription ID
	if cfg.Auth.Params["subscription_id"] == "" && os.Getenv("AZURE_SUBSCRIPTION_ID") == "" {
		return fmt.Errorf("subscription_id required")
	}

	return nil
}

// HealthCheck performs health check on provider connection
func (p *Provider) HealthCheck(ctx context.Context) error {
	// Check if network client is initialized
	if p.networkClient == nil {
		return fmt.Errorf("network client not initialized")
	}

	// Try to list virtual networks to verify connection
	pager := p.networkClient.NewListPager(p.resourceGroupName, nil)
	if pager.More() {
		_, err := pager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to access Azure Virtual Networks: %w", err)
		}
	}

	fmt.Printf("Connected to Azure subscription: %s\n", p.subscriptionID)
	return nil
}

// isValidAzureRegion checks if the region is a valid Azure region
func isValidAzureRegion(region string) bool {
	// This is a simplified check - in production, you'd have a complete list
	validRegions := []string{
		"eastus", "eastus2", "westus", "westus2", "westus3", "centralus", "southcentralus", "northcentralus",
		"canadacentral", "canadaeast",
		"brazilsouth",
		"northeurope", "westeurope", "uksouth", "ukwest", "francecentral", "germanywestcentral",
		"norwayeast", "switzerlandnorth",
		"eastasia", "southeastasia", "japaneast", "japanwest", "koreacentral", "koreasouth",
		"australiaeast", "australiasoutheast", "australiacentral",
		"centralindia", "southindia", "westindia",
		"southafricanorth",
		"uaenorth",
	}

	for _, valid := range validRegions {
		if region == valid {
			return true
		}
	}

	return false
}
