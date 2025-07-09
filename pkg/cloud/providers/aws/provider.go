package aws

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/gizzahub/gzh-manager-go/pkg/cloud"
)

func init() {
	// Register AWS provider
	cloud.Register(cloud.ProviderTypeAWS, NewProvider)
}

// Provider implements cloud.Provider for AWS
type Provider struct {
	config    cloud.ProviderConfig
	awsConfig aws.Config
	ec2Client *ec2.Client
	stsClient *sts.Client
}

// NewProvider creates a new AWS provider instance
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
	return string(cloud.ProviderTypeAWS)
}

// Initialize initializes the AWS provider
func (p *Provider) Initialize(ctx context.Context, cfg cloud.ProviderConfig) error {
	// Configure AWS SDK based on auth method
	var awsCfg aws.Config
	var err error

	switch cfg.Auth.Method {
	case "iam":
		// Use IAM instance profile or default credentials
		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
		)
	case "key":
		// Use access key credentials
		accessKey, ok := cfg.Auth.Params["access_key"]
		if !ok {
			return fmt.Errorf("access_key required for key auth method")
		}
		secretKey, ok := cfg.Auth.Params["secret_key"]
		if !ok {
			return fmt.Errorf("secret_key required for key auth method")
		}

		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
			),
		)
	case "profile":
		// Use AWS profile
		profileName, ok := cfg.Auth.Params["profile"]
		if !ok {
			profileName = "default"
		}

		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
			config.WithSharedConfigProfile(profileName),
		)
	case "env":
		// Use environment variables
		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
		)
	default:
		return fmt.Errorf("unsupported auth method: %s", cfg.Auth.Method)
	}

	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	p.awsConfig = awsCfg
	p.ec2Client = ec2.NewFromConfig(awsCfg)
	p.stsClient = sts.NewFromConfig(awsCfg)

	return nil
}

// GetProfile retrieves a specific profile configuration
func (p *Provider) GetProfile(ctx context.Context, profileName string) (*cloud.Profile, error) {
	// In AWS context, profiles map to VPCs/environments
	// This is a simplified implementation
	profile := &cloud.Profile{
		Name:        profileName,
		Provider:    p.Name(),
		Environment: profileName,
		Region:      p.config.Region,
		LastSync:    time.Now(),
	}

	// Get VPC information
	vpcs, err := p.ec2Client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
		Filters: []ec2Types.Filter{
			{
				Name:   aws.String("tag:Profile"),
				Values: []string{profileName},
			},
		},
	})
	if err == nil && len(vpcs.Vpcs) > 0 {
		vpc := vpcs.Vpcs[0]
		profile.Network.VPCId = aws.ToString(vpc.VpcId)

		// Get subnets
		subnets, err := p.ec2Client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
			Filters: []ec2Types.Filter{
				{
					Name:   aws.String("vpc-id"),
					Values: []string{profile.Network.VPCId},
				},
			},
		})
		if err == nil {
			for _, subnet := range subnets.Subnets {
				profile.Network.SubnetIds = append(profile.Network.SubnetIds, aws.ToString(subnet.SubnetId))
			}
		}

		// Get security groups
		sgs, err := p.ec2Client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
			Filters: []ec2Types.Filter{
				{
					Name:   aws.String("vpc-id"),
					Values: []string{profile.Network.VPCId},
				},
			},
		})
		if err == nil {
			for _, sg := range sgs.SecurityGroups {
				profile.Network.SecurityGroups = append(profile.Network.SecurityGroups, aws.ToString(sg.GroupId))
			}
		}
	}

	// Set AWS-specific tags
	profile.Tags = map[string]string{
		"provider": "aws",
		"region":   p.config.Region,
	}

	return profile, nil
}

// ListProfiles lists all available profiles
func (p *Provider) ListProfiles(ctx context.Context) ([]*cloud.Profile, error) {
	var profiles []*cloud.Profile

	// List VPCs with Profile tag
	vpcs, err := p.ec2Client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
		Filters: []ec2Types.Filter{
			{
				Name:   aws.String("tag-key"),
				Values: []string{"Profile"},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe VPCs: %w", err)
	}

	// Create profile for each VPC
	for _, vpc := range vpcs.Vpcs {
		profileName := ""
		for _, tag := range vpc.Tags {
			if aws.ToString(tag.Key) == "Profile" {
				profileName = aws.ToString(tag.Value)
				break
			}
		}

		if profileName != "" {
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
	// 1. Create/update VPC if needed
	// 2. Configure subnets
	// 3. Set up security groups
	// 4. Configure route tables
	// 5. Set up VPC endpoints

	// For now, we'll just validate the profile
	if profile.Name == "" {
		return fmt.Errorf("profile name is required")
	}

	if profile.Network.VPCId == "" {
		// Would create a new VPC here
		return fmt.Errorf("VPC creation not implemented")
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

	// Add security group rules as policy rules
	if len(profile.Network.SecurityGroups) > 0 {
		// Would fetch actual security group rules here
		policy.Rules = append(policy.Rules, cloud.PolicyRule{
			Type:        "allow",
			Source:      "0.0.0.0/0",
			Destination: profile.Network.VPCId,
			Protocol:    "tcp",
			Port:        "443",
		})
	}

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
				os.Setenv("HTTP_PROXY", httpProxy)
				os.Setenv("http_proxy", httpProxy)
			}
			if httpsProxy, ok := action.Params["https_proxy"]; ok {
				os.Setenv("HTTPS_PROXY", httpsProxy)
				os.Setenv("https_proxy", httpsProxy)
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
	if cfg.Type != string(cloud.ProviderTypeAWS) {
		return fmt.Errorf("invalid provider type: expected aws, got %s", cfg.Type)
	}

	if cfg.Region == "" {
		return fmt.Errorf("region is required")
	}

	// Validate region format
	if !isValidAWSRegion(cfg.Region) {
		return fmt.Errorf("invalid AWS region: %s", cfg.Region)
	}

	// Validate auth method
	switch cfg.Auth.Method {
	case "iam", "key", "profile", "env":
		// Valid methods
	default:
		return fmt.Errorf("unsupported auth method: %s", cfg.Auth.Method)
	}

	// Validate auth params based on method
	if cfg.Auth.Method == "key" {
		if _, ok := cfg.Auth.Params["access_key"]; !ok {
			return fmt.Errorf("access_key required for key auth method")
		}
		if _, ok := cfg.Auth.Params["secret_key"]; !ok {
			return fmt.Errorf("secret_key required for key auth method")
		}
	}

	return nil
}

// HealthCheck performs health check on provider connection
func (p *Provider) HealthCheck(ctx context.Context) error {
	// Try to get caller identity
	result, err := p.stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("failed to verify AWS credentials: %w", err)
	}

	// Verify we can access EC2
	_, err = p.ec2Client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	if err != nil {
		return fmt.Errorf("failed to access EC2 service: %w", err)
	}

	if result.Account != nil {
		fmt.Printf("Connected to AWS account: %s\n", aws.ToString(result.Account))
	}

	return nil
}

// isValidAWSRegion checks if the region is a valid AWS region
func isValidAWSRegion(region string) bool {
	// This is a simplified check - in production, you'd have a complete list
	validRegions := []string{
		"us-east-1", "us-east-2", "us-west-1", "us-west-2",
		"eu-west-1", "eu-west-2", "eu-west-3", "eu-central-1",
		"ap-northeast-1", "ap-northeast-2", "ap-northeast-3",
		"ap-southeast-1", "ap-southeast-2", "ap-south-1",
		"sa-east-1", "ca-central-1",
	}

	for _, valid := range validRegions {
		if region == valid {
			return true
		}
	}

	return false
}
