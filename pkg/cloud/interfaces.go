package cloud

import (
	"context"
	"time"
)

// Provider represents a cloud provider interface
type Provider interface {
	// Name returns the provider name (aws, gcp, azure)
	Name() string

	// Initialize initializes the provider with given configuration
	Initialize(ctx context.Context, config ProviderConfig) error

	// GetProfile retrieves a specific profile configuration
	GetProfile(ctx context.Context, profileName string) (*Profile, error)

	// ListProfiles lists all available profiles
	ListProfiles(ctx context.Context) ([]*Profile, error)

	// SyncProfile synchronizes a profile configuration
	SyncProfile(ctx context.Context, profile *Profile) error

	// GetNetworkPolicy retrieves network policy for a profile
	GetNetworkPolicy(ctx context.Context, profileName string) (*NetworkPolicy, error)

	// ApplyNetworkPolicy applies network policy settings
	ApplyNetworkPolicy(ctx context.Context, policy *NetworkPolicy) error

	// ValidateConfig validates provider configuration
	ValidateConfig(config ProviderConfig) error

	// HealthCheck performs health check on provider connection
	HealthCheck(ctx context.Context) error
}

// ProviderConfig represents provider-specific configuration
type ProviderConfig struct {
	// Provider type (aws, gcp, azure)
	Type string `yaml:"type" json:"type"`

	// Region or location
	Region string `yaml:"region" json:"region"`

	// Authentication configuration
	Auth AuthConfig `yaml:"auth" json:"auth"`

	// Provider-specific settings
	Settings map[string]interface{} `yaml:"settings,omitempty" json:"settings,omitempty"`

	// Timeout for operations
	Timeout time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	// Authentication method (key, token, iam, service_account, etc.)
	Method string `yaml:"method" json:"method"`

	// Credentials file path (optional)
	CredentialsFile string `yaml:"credentials_file,omitempty" json:"credentials_file,omitempty"`

	// Environment variable prefix for credentials
	EnvPrefix string `yaml:"env_prefix,omitempty" json:"env_prefix,omitempty"`

	// Additional auth parameters
	Params map[string]string `yaml:"params,omitempty" json:"params,omitempty"`
}

// Profile represents a cloud environment profile
type Profile struct {
	// Profile name
	Name string `yaml:"name" json:"name"`

	// Provider type
	Provider string `yaml:"provider" json:"provider"`

	// Environment (dev, staging, prod)
	Environment string `yaml:"environment" json:"environment"`

	// Region or location
	Region string `yaml:"region" json:"region"`

	// Network configuration
	Network NetworkConfig `yaml:"network" json:"network"`

	// Services configuration
	Services map[string]ServiceConfig `yaml:"services,omitempty" json:"services,omitempty"`

	// Tags/labels
	Tags map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`

	// Last sync timestamp
	LastSync time.Time `yaml:"last_sync,omitempty" json:"last_sync,omitempty"`
}

// NetworkConfig represents network configuration for a profile
type NetworkConfig struct {
	// VPC/VNet ID
	VPCId string `yaml:"vpc_id,omitempty" json:"vpc_id,omitempty"`

	// Subnet IDs
	SubnetIds []string `yaml:"subnet_ids,omitempty" json:"subnet_ids,omitempty"`

	// Security groups
	SecurityGroups []string `yaml:"security_groups,omitempty" json:"security_groups,omitempty"`

	// CIDR blocks
	CIDRBlocks []string `yaml:"cidr_blocks,omitempty" json:"cidr_blocks,omitempty"`

	// DNS servers
	DNSServers []string `yaml:"dns_servers,omitempty" json:"dns_servers,omitempty"`

	// Proxy configuration
	Proxy *ProxyConfig `yaml:"proxy,omitempty" json:"proxy,omitempty"`

	// VPN configuration
	VPN *VPNConfig `yaml:"vpn,omitempty" json:"vpn,omitempty"`

	// Custom routes
	Routes []RouteConfig `yaml:"routes,omitempty" json:"routes,omitempty"`
}

// ServiceConfig represents service-specific configuration
type ServiceConfig struct {
	// Service endpoint
	Endpoint string `yaml:"endpoint" json:"endpoint"`

	// Service port
	Port int `yaml:"port,omitempty" json:"port,omitempty"`

	// Authentication required
	AuthRequired bool `yaml:"auth_required,omitempty" json:"auth_required,omitempty"`

	// TLS/SSL configuration
	TLS *TLSConfig `yaml:"tls,omitempty" json:"tls,omitempty"`
}

// ProxyConfig represents proxy configuration
type ProxyConfig struct {
	// HTTP proxy
	HTTP string `yaml:"http,omitempty" json:"http,omitempty"`

	// HTTPS proxy
	HTTPS string `yaml:"https,omitempty" json:"https,omitempty"`

	// No proxy hosts
	NoProxy []string `yaml:"no_proxy,omitempty" json:"no_proxy,omitempty"`

	// Proxy authentication
	Auth *ProxyAuth `yaml:"auth,omitempty" json:"auth,omitempty"`
}

// ProxyAuth represents proxy authentication
type ProxyAuth struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
}

// VPNConfig represents VPN configuration
type VPNConfig struct {
	// VPN type (openvpn, wireguard, ipsec)
	Type string `yaml:"type" json:"type"`

	// VPN server endpoint
	Server string `yaml:"server" json:"server"`

	// VPN port
	Port int `yaml:"port,omitempty" json:"port,omitempty"`

	// Configuration file path
	ConfigFile string `yaml:"config_file,omitempty" json:"config_file,omitempty"`

	// Auto-connect on network change
	AutoConnect bool `yaml:"auto_connect,omitempty" json:"auto_connect,omitempty"`
}

// RouteConfig represents custom route configuration
type RouteConfig struct {
	// Destination CIDR
	Destination string `yaml:"destination" json:"destination"`

	// Gateway IP
	Gateway string `yaml:"gateway" json:"gateway"`

	// Metric/priority
	Metric int `yaml:"metric,omitempty" json:"metric,omitempty"`
}

// TLSConfig represents TLS/SSL configuration
type TLSConfig struct {
	// Skip verification (insecure)
	SkipVerify bool `yaml:"skip_verify,omitempty" json:"skip_verify,omitempty"`

	// CA certificate file
	CAFile string `yaml:"ca_file,omitempty" json:"ca_file,omitempty"`

	// Client certificate file
	CertFile string `yaml:"cert_file,omitempty" json:"cert_file,omitempty"`

	// Client key file
	KeyFile string `yaml:"key_file,omitempty" json:"key_file,omitempty"`
}

// NetworkPolicy represents network policy that can be applied
type NetworkPolicy struct {
	// Policy name
	Name string `yaml:"name" json:"name"`

	// Profile name this policy belongs to
	ProfileName string `yaml:"profile_name,omitempty" json:"profile_name,omitempty"`

	// Environment this policy applies to
	Environment string `yaml:"environment,omitempty" json:"environment,omitempty"`

	// Provider this policy applies to
	Provider string `yaml:"provider,omitempty" json:"provider,omitempty"`

	// Policy rules
	Rules []PolicyRule `yaml:"rules" json:"rules"`

	// Actions to take when policy is applied
	Actions []PolicyAction `yaml:"actions" json:"actions"`

	// Priority (higher number = higher priority)
	Priority int `yaml:"priority" json:"priority"`

	// Enabled status
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// PolicyRule represents a network policy rule
type PolicyRule struct {
	// Rule type (allow, deny, redirect)
	Type string `yaml:"type" json:"type"`

	// Source (CIDR, service name, etc.)
	Source string `yaml:"source,omitempty" json:"source,omitempty"`

	// Destination (CIDR, service name, etc.)
	Destination string `yaml:"destination,omitempty" json:"destination,omitempty"`

	// Protocol (tcp, udp, icmp, any)
	Protocol string `yaml:"protocol,omitempty" json:"protocol,omitempty"`

	// Port or port range
	Port string `yaml:"port,omitempty" json:"port,omitempty"`
}

// PolicyAction represents an action to take when policy is applied
type PolicyAction struct {
	// Action type (configure_dns, setup_proxy, connect_vpn, add_route)
	Type string `yaml:"type" json:"type"`

	// Action parameters
	Params map[string]string `yaml:"params,omitempty" json:"params,omitempty"`

	// Order of execution
	Order int `yaml:"order" json:"order"`
}

// SyncManager manages synchronization between cloud providers
type SyncManager interface {
	// SyncProfiles synchronizes profiles between providers
	SyncProfiles(ctx context.Context, source, target Provider, profileNames []string) error

	// SyncAll synchronizes all profiles between providers
	SyncAll(ctx context.Context, source, target Provider) error

	// GetSyncStatus returns sync status for profiles
	GetSyncStatus(ctx context.Context) ([]SyncStatus, error)

	// ResolveSyncConflicts resolves conflicts during sync
	ResolveSyncConflicts(conflicts []SyncConflict, strategy ConflictStrategy) error
}

// SyncStatus represents synchronization status
type SyncStatus struct {
	ProfileName string    `json:"profile_name"`
	Source      string    `json:"source"`
	Target      string    `json:"target"`
	Status      string    `json:"status"` // synced, pending, conflict, error
	LastSync    time.Time `json:"last_sync"`
	Error       string    `json:"error,omitempty"`
}

// SyncConflict represents a synchronization conflict
type SyncConflict struct {
	ProfileName string      `json:"profile_name"`
	Field       string      `json:"field"`
	SourceValue interface{} `json:"source_value"`
	TargetValue interface{} `json:"target_value"`
}

// ConflictStrategy represents how to resolve sync conflicts
type ConflictStrategy string

const (
	// ConflictStrategySourceWins uses source value in conflicts
	ConflictStrategySourceWins ConflictStrategy = "source_wins"

	// ConflictStrategyTargetWins uses target value in conflicts
	ConflictStrategyTargetWins ConflictStrategy = "target_wins"

	// ConflictStrategyMerge attempts to merge values
	ConflictStrategyMerge ConflictStrategy = "merge"

	// ConflictStrategyAsk prompts user for each conflict
	ConflictStrategyAsk ConflictStrategy = "ask"
)
