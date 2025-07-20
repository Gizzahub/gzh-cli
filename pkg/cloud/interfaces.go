package cloud

import (
	"context"
	"time"
)

// Provider represents a cloud provider interface.
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

// ProviderConfig represents provider-specific configuration.
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

// AuthConfig represents authentication configuration.
type AuthConfig struct {
	// Authentication method (key, token, iam, service_account, etc.)
	Method string `yaml:"method" json:"method"`

	// Credentials file path (optional)
	CredentialsFile string `yaml:"credentialsFile,omitempty" json:"credentialsFile,omitempty"`

	// Environment variable prefix for credentials
	EnvPrefix string `yaml:"envPrefix,omitempty" json:"envPrefix,omitempty"`

	// Additional auth parameters
	Params map[string]string `yaml:"params,omitempty" json:"params,omitempty"`
}

// Profile represents a cloud environment profile.
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
	LastSync time.Time `yaml:"lastSync,omitempty" json:"lastSync,omitempty"`
}

// NetworkConfig represents network configuration for a profile.
type NetworkConfig struct {
	// VPC/VNet ID
	VPCId string `yaml:"vpcId,omitempty" json:"vpcId,omitempty"`

	// Subnet IDs
	SubnetIds []string `yaml:"subnetIds,omitempty" json:"subnetIds,omitempty"`

	// Security groups
	SecurityGroups []string `yaml:"securityGroups,omitempty" json:"securityGroups,omitempty"`

	// CIDR blocks
	CIDRBlocks []string `yaml:"cidrBlocks,omitempty" json:"cidrBlocks,omitempty"`

	// DNS servers
	DNSServers []string `yaml:"dnsServers,omitempty" json:"dnsServers,omitempty"`

	// Proxy configuration
	Proxy *ProxyConfig `yaml:"proxy,omitempty" json:"proxy,omitempty"`

	// VPN configuration
	VPN *VPNConfig `yaml:"vpn,omitempty" json:"vpn,omitempty"`

	// Custom routes
	Routes []RouteConfig `yaml:"routes,omitempty" json:"routes,omitempty"`
}

// ServiceConfig represents service-specific configuration.
type ServiceConfig struct {
	// Service endpoint
	Endpoint string `yaml:"endpoint" json:"endpoint"`

	// Service port
	Port int `yaml:"port,omitempty" json:"port,omitempty"`

	// Authentication required
	AuthRequired bool `yaml:"authRequired,omitempty" json:"authRequired,omitempty"`

	// TLS/SSL configuration
	TLS *TLSConfig `yaml:"tls,omitempty" json:"tls,omitempty"`
}

// ProxyConfig represents proxy configuration.
type ProxyConfig struct {
	// HTTP proxy
	HTTP string `yaml:"http,omitempty" json:"http,omitempty"`

	// HTTPS proxy
	HTTPS string `yaml:"https,omitempty" json:"https,omitempty"`

	// No proxy hosts
	NoProxy []string `yaml:"noProxy,omitempty" json:"noProxy,omitempty"`

	// Proxy authentication
	Auth *ProxyAuth `yaml:"auth,omitempty" json:"auth,omitempty"`
}

// ProxyAuth represents proxy authentication.
type ProxyAuth struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
}

// VPNConfig represents VPN configuration.
type VPNConfig struct {
	// VPN type (openvpn, wireguard, ipsec)
	Type string `yaml:"type" json:"type"`

	// VPN server endpoint
	Server string `yaml:"server" json:"server"`

	// VPN port
	Port int `yaml:"port,omitempty" json:"port,omitempty"`

	// Configuration file path
	ConfigFile string `yaml:"configFile,omitempty" json:"configFile,omitempty"`

	// Auto-connect on network change
	AutoConnect bool `yaml:"autoConnect,omitempty" json:"autoConnect,omitempty"`
}

// VPNConnection represents a VPN connection configuration.
type VPNConnection struct {
	// Connection name
	Name string `yaml:"name" json:"name"`

	// VPN type (openvpn, wireguard, ipsec)
	Type string `yaml:"type" json:"type"`

	// VPN server endpoint
	Server string `yaml:"server" json:"server"`

	// VPN port
	Port int `yaml:"port,omitempty" json:"port,omitempty"`

	// Configuration file path
	ConfigFile string `yaml:"configFile,omitempty" json:"configFile,omitempty"`

	// Username for authentication
	Username string `yaml:"username,omitempty" json:"username,omitempty"`

	// Password for authentication (optional, can use keychain/env)
	Password string `yaml:"password,omitempty" json:"password,omitempty"`

	// Certificate files for authentication
	CertFile string `yaml:"certFile,omitempty" json:"certFile,omitempty"`
	KeyFile  string `yaml:"keyFile,omitempty" json:"keyFile,omitempty"`
	CAFile   string `yaml:"caFile,omitempty" json:"caFile,omitempty"`

	// Auto-connect on network change
	AutoConnect bool `yaml:"autoConnect,omitempty" json:"autoConnect,omitempty"`

	// Connection timeout
	Timeout time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`

	// Retry configuration
	MaxRetries    int           `yaml:"maxRetries,omitempty" json:"maxRetries,omitempty"`
	RetryInterval time.Duration `yaml:"retryInterval,omitempty" json:"retryInterval,omitempty"`

	// Health check configuration
	HealthCheck *VPNHealthCheck `yaml:"healthCheck,omitempty" json:"healthCheck,omitempty"`

	// Route configuration for this VPN
	Routes []RouteConfig `yaml:"routes,omitempty" json:"routes,omitempty"`

	// DNS servers to use when connected
	DNSServers []string `yaml:"dnsServers,omitempty" json:"dnsServers,omitempty"`

	// Environment this VPN connection belongs to
	Environment string `yaml:"environment,omitempty" json:"environment,omitempty"`

	// Priority for connection ordering (higher = more priority)
	Priority int `yaml:"priority,omitempty" json:"priority,omitempty"`

	// Tags for categorization and filtering
	Tags map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// VPNHealthCheck represents health check configuration for VPN connections.
type VPNHealthCheck struct {
	// Enable health checking
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Interval between health checks
	Interval time.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`

	// Timeout for each health check
	Timeout time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`

	// Target host/IP to ping for health check
	Target string `yaml:"target,omitempty" json:"target,omitempty"`

	// Number of failed checks before marking as unhealthy
	FailureThreshold int `yaml:"failureThreshold,omitempty" json:"failureThreshold,omitempty"`

	// Number of successful checks before marking as healthy
	SuccessThreshold int `yaml:"successThreshold,omitempty" json:"successThreshold,omitempty"`
}

// RouteConfig represents custom route configuration.
type RouteConfig struct {
	// Destination CIDR
	Destination string `yaml:"destination" json:"destination"`

	// Gateway IP
	Gateway string `yaml:"gateway" json:"gateway"`

	// Metric/priority
	Metric int `yaml:"metric,omitempty" json:"metric,omitempty"`
}

// TLSConfig represents TLS/SSL configuration.
type TLSConfig struct {
	// Skip verification (insecure)
	SkipVerify bool `yaml:"skipVerify,omitempty" json:"skipVerify,omitempty"`

	// CA certificate file
	CAFile string `yaml:"caFile,omitempty" json:"caFile,omitempty"`

	// Client certificate file
	CertFile string `yaml:"certFile,omitempty" json:"certFile,omitempty"`

	// Client key file
	KeyFile string `yaml:"keyFile,omitempty" json:"keyFile,omitempty"`
}

// NetworkPolicy represents network policy that can be applied.
type NetworkPolicy struct {
	// Policy name
	Name string `yaml:"name" json:"name"`

	// Profile name this policy belongs to
	ProfileName string `yaml:"profileName,omitempty" json:"profileName,omitempty"`

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

// PolicyRule represents a network policy rule.
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

// PolicyAction represents an action to take when policy is applied.
type PolicyAction struct {
	// Action type (configure_dns, setup_proxy, connect_vpn, add_route)
	Type string `yaml:"type" json:"type"`

	// Action parameters
	Params map[string]string `yaml:"params,omitempty" json:"params,omitempty"`

	// Order of execution
	Order int `yaml:"order" json:"order"`
}

// VPNManager manages VPN connections.
type VPNManager interface {
	// AddVPNConnection adds a VPN connection
	AddVPNConnection(conn *VPNConnection) error

	// RemoveVPNConnection removes a VPN connection
	RemoveVPNConnection(name string) error

	// GetVPNConnection retrieves a VPN connection by name
	GetVPNConnection(name string) (*VPNConnection, error)

	// ListVPNConnections lists all VPN connections
	ListVPNConnections() ([]*VPNConnection, error)

	// ConnectVPN connects to a VPN
	ConnectVPN(ctx context.Context, name string) error

	// DisconnectVPN disconnects from a VPN
	DisconnectVPN(ctx context.Context, name string) error

	// GetVPNStatus returns the status of a VPN connection
	GetVPNStatus(ctx context.Context, name string) (*VPNStatus, error)

	// GetAllVPNStatuses returns statuses of all VPN connections
	GetAllVPNStatuses(ctx context.Context) (map[string]*VPNStatus, error)

	// GetConnectionStatus returns the status of a VPN connection (alias for GetVPNStatus)
	GetConnectionStatus(ctx context.Context, name string) (*VPNStatus, error)

	// GetActiveConnections returns all active VPN connections
	GetActiveConnections(ctx context.Context) (map[string]*VPNStatus, error)

	// ConnectByPriority connects VPN connections by priority order
	ConnectByPriority(ctx context.Context, connectionNames []string) error
}

// HierarchicalVPNManager manages hierarchical VPN connections.
type HierarchicalVPNManager interface {
	VPNManager

	// AddVPNHierarchy adds a VPN hierarchy
	AddVPNHierarchy(hierarchy *VPNHierarchy) error

	// RemoveVPNHierarchy removes a VPN hierarchy
	RemoveVPNHierarchy(name string) error

	// GetVPNHierarchy retrieves a VPN hierarchy by name
	GetVPNHierarchy(name string) (*VPNHierarchy, error)

	// ListVPNHierarchies lists all VPN hierarchies
	ListVPNHierarchies() ([]*VPNHierarchy, error)

	// ConnectVPNHierarchy connects to a VPN hierarchy
	ConnectVPNHierarchy(ctx context.Context, name string) error

	// DisconnectVPNHierarchy disconnects from a VPN hierarchy
	DisconnectVPNHierarchy(ctx context.Context, name string) error

	// GetVPNHierarchyStatus returns the status of a VPN hierarchy
	GetVPNHierarchyStatus(ctx context.Context, name string) (*VPNHierarchyStatus, error)
}

// PolicyManager manages network policies.
type PolicyManager interface {
	// AddPolicy adds a network policy
	AddPolicy(policy *NetworkPolicy) error

	// RemovePolicy removes a network policy
	RemovePolicy(name string) error

	// GetPolicy retrieves a network policy by name
	GetPolicy(name string) (*NetworkPolicy, error)

	// ListPolicies lists all network policies
	ListPolicies() ([]*NetworkPolicy, error)

	// ApplyPolicy applies a network policy
	ApplyPolicy(ctx context.Context, name string) error

	// RemovePolicy removes a network policy
	RemoveAppliedPolicy(ctx context.Context, name string) error

	// ApplyEnvironmentPolicies applies policies for an environment
	ApplyEnvironmentPolicies(ctx context.Context, environment string) error

	// GetApplicablePolicies gets applicable policies for a profile
	GetApplicablePolicies(ctx context.Context, profileName string) ([]*NetworkPolicy, error)

	// ApplyPoliciesForProfile applies policies for a specific profile
	ApplyPoliciesForProfile(ctx context.Context, profileName string) error

	// GetPolicyStatus gets the status of applied policies
	GetPolicyStatus(ctx context.Context) ([]*PolicyStatus, error)

	// GetPolicyStatusForProfile gets the status of applied policies for a specific profile
	GetPolicyStatusForProfile(ctx context.Context, profileName string) (map[string]string, error)

	// ValidatePolicy validates a network policy
	ValidatePolicy(ctx context.Context, policy *NetworkPolicy) error
}

// SyncManager manages synchronization between cloud providers.
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

// SyncStatus represents synchronization status.
type SyncStatus struct {
	ProfileName string    `json:"profileName"`
	Source      string    `json:"source"`
	Target      string    `json:"target"`
	Status      string    `json:"status"` // synced, pending, conflict, error
	LastSync    time.Time `json:"lastSync"`
	Error       string    `json:"error,omitempty"`
}

// SyncConflict represents a synchronization conflict.
type SyncConflict struct {
	ProfileName string      `json:"profileName"`
	Field       string      `json:"field"`
	SourceValue interface{} `json:"sourceValue"`
	TargetValue interface{} `json:"targetValue"`
}

// ConflictStrategy represents how to resolve sync conflicts.
type ConflictStrategy string

const (
	// ConflictStrategySourceWins uses source value in conflicts.
	ConflictStrategySourceWins ConflictStrategy = "sourceWins"

	// ConflictStrategyTargetWins uses target value in conflicts.
	ConflictStrategyTargetWins ConflictStrategy = "targetWins"

	// ConflictStrategyMerge attempts to merge values.
	ConflictStrategyMerge ConflictStrategy = "merge"

	// ConflictStrategyAsk prompts user for each conflict.
	ConflictStrategyAsk ConflictStrategy = "ask"
)

// VPN connection state constants.
const (
	VPNStateDisconnected = "disconnected"
	VPNStateConnected    = "connected"
	VPNStateConnecting   = "connecting"
	VPNStateError        = "error"
)

// VPNStatus represents the status of a VPN connection.
type VPNStatus struct {
	// Connection name
	Name string `json:"name"`

	// Connection status (connected, disconnected, connecting, error)
	Status string `json:"status"`

	// IP address assigned to the VPN connection
	IPAddress string `json:"ipAddress,omitempty"`

	// Connection uptime
	Uptime time.Duration `json:"uptime,omitempty"`

	// Data transferred
	BytesReceived uint64 `json:"bytesReceived,omitempty"`
	BytesSent     uint64 `json:"bytesSent,omitempty"`

	// Last error (if any)
	LastError string `json:"lastError,omitempty"`

	// Connection timestamp
	ConnectedAt time.Time `json:"connectedAt,omitempty"`

	// Health check status
	HealthCheck *VPNHealthStatus `json:"healthCheck,omitempty"`
}

// VPNHealthStatus represents health check status for VPN.
type VPNHealthStatus struct {
	// Health status (healthy, unhealthy, unknown)
	Status string `json:"status"`

	// Last health check timestamp
	LastCheck time.Time `json:"lastCheck"`

	// Health check target
	Target string `json:"target"`

	// Response time
	ResponseTime time.Duration `json:"responseTime,omitempty"`

	// Failure count
	FailureCount int `json:"failureCount"`

	// Success count
	SuccessCount int `json:"successCount"`
}

// VPNHierarchy represents a hierarchical VPN configuration.
type VPNHierarchy struct {
	// Hierarchy name
	Name string `yaml:"name" json:"name"`

	// Description
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Hierarchy nodes organized by layer
	Layers map[int][]*VPNHierarchyNode `yaml:"layers" json:"layers"`

	// Connection policy
	Policy VPNHierarchyPolicy `yaml:"policy" json:"policy"`

	// Environment this hierarchy belongs to
	Environment string `yaml:"environment,omitempty" json:"environment,omitempty"`

	// Tags for categorization
	Tags map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// VPNHierarchyNode represents a node in the VPN hierarchy.
type VPNHierarchyNode struct {
	// Node name
	Name string `yaml:"name" json:"name"`

	// VPN connection configuration
	Connection *VPNConnection `yaml:"connection" json:"connection"`

	// Layer in the hierarchy (0 = first layer)
	Layer int `yaml:"layer" json:"layer"`

	// Dependencies (nodes that must be connected before this one)
	Dependencies []string `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`

	// Failover configuration
	Failover *VPNFailoverConfig `yaml:"failover,omitempty" json:"failover,omitempty"`

	// Health check configuration
	HealthCheck *VPNHealthCheck `yaml:"healthCheck,omitempty" json:"healthCheck,omitempty"`

	// Auto-reconnect configuration
	AutoReconnect bool `yaml:"autoReconnect,omitempty" json:"autoReconnect,omitempty"`
}

// VPNHierarchyPolicy represents policy for VPN hierarchy connections.
type VPNHierarchyPolicy struct {
	// Connection strategy (sequential, parallel, smart)
	Strategy string `yaml:"strategy" json:"strategy"`

	// Timeout for each connection attempt
	ConnectionTimeout time.Duration `yaml:"connectionTimeout,omitempty" json:"connectionTimeout,omitempty"`

	// Maximum retries per connection
	MaxRetries int `yaml:"maxRetries,omitempty" json:"maxRetries,omitempty"`

	// Failure handling (stop, continue, failover)
	FailureHandling string `yaml:"failureHandling,omitempty" json:"failureHandling,omitempty"`
}

// VPNFailoverConfig represents failover configuration for VPN connections.
type VPNFailoverConfig struct {
	// Enable failover
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Backup VPN connections (in order of preference)
	BackupConnections []string `yaml:"backupConnections,omitempty" json:"backupConnections,omitempty"`

	// Failover trigger conditions
	TriggerConditions []string `yaml:"triggerConditions,omitempty" json:"triggerConditions,omitempty"`

	// Failover timeout
	FailoverTimeout time.Duration `yaml:"failoverTimeout,omitempty" json:"failoverTimeout,omitempty"`

	// Auto-failback configuration
	AutoFailback bool `yaml:"autoFailback,omitempty" json:"autoFailback,omitempty"`
}

// VPNHierarchyStatus represents the status of a VPN hierarchy.
type VPNHierarchyStatus struct {
	// Hierarchy name
	Name string `json:"name"`

	// Overall status (connected, disconnected, partial, error)
	Status string `json:"status"`

	// Status of each layer
	LayerStatuses map[int]*VPNLayerStatus `json:"layerStatuses"`

	// Node statuses
	NodeStatuses map[string]*VPNStatus `json:"nodeStatuses"`

	// Last connection attempt
	LastConnectionAttempt time.Time `json:"lastConnectionAttempt"`

	// Active connections count
	ActiveConnections int `json:"activeConnections"`

	// Total connections count
	TotalConnections int `json:"totalConnections"`
}

// VPNLayerStatus represents the status of a layer in VPN hierarchy.
type VPNLayerStatus struct {
	// Layer number
	Layer int `json:"layer"`

	// Layer status (connected, disconnected, partial, error)
	Status string `json:"status"`

	// Connected nodes count
	ConnectedNodes int `json:"connectedNodes"`

	// Total nodes count
	TotalNodes int `json:"totalNodes"`

	// Layer connection time
	ConnectedAt time.Time `json:"connectedAt,omitempty"`
}

// NetworkEnvironment represents a network environment configuration.
type NetworkEnvironment struct {
	// Environment name
	Name string `yaml:"name" json:"name"`

	// Environment description
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Network configuration
	Network NetworkConfig `yaml:"network" json:"network"`

	// VPN connections for this environment
	VPNConnections map[string]*VPNConnection `yaml:"vpnConnections,omitempty" json:"vpnConnections,omitempty"`

	// Network policies for this environment
	NetworkPolicies []*NetworkPolicy `yaml:"networkPolicies,omitempty" json:"networkPolicies,omitempty"`

	// Environment variables
	Environment map[string]string `yaml:"environment,omitempty" json:"environment,omitempty"`

	// Tags for categorization
	Tags map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// PolicyStatus represents the status of a network policy.
type PolicyStatus struct {
	// Policy name
	PolicyName string `json:"policyName"`

	// Profile name
	ProfileName string `json:"profileName"`

	// Provider name
	Provider string `json:"provider"`

	// Status (applied, failed, pending)
	Status string `json:"status"`

	// Applied timestamp
	Applied time.Time `json:"applied"`

	// Error message if any
	Error string `json:"error,omitempty"`
}
