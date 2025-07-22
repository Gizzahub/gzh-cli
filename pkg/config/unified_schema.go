// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

import (
	"time"
)

// UnifiedConfig represents the new unified configuration format
// This merges the functionality of both bulk-clone.yaml and gzh.yaml formats.
type UnifiedConfig struct {
	// Schema version for the configuration format
	Version string `yaml:"version" json:"version" validate:"required,oneof=1.0.0"`

	// Default provider to use when not specified
	DefaultProvider string `yaml:"defaultProvider,omitempty" json:"defaultProvider,omitempty" validate:"omitempty,oneof=github gitlab gitea gogs"` //nolint:revive // Custom validation tags are valid

	// Global settings that apply to all providers
	Global *GlobalSettings `yaml:"global,omitempty" json:"global,omitempty"`

	// Provider-specific configurations
	Providers map[string]*ProviderConfig `yaml:"providers" json:"providers" validate:"required,min=1"`

	// Migration information from legacy formats
	Migration *MigrationInfo `yaml:"migration,omitempty" json:"migration,omitempty"`

	// IDE monitoring configuration
	IDE *IDEConfig `yaml:"ide,omitempty" json:"ide,omitempty"`

	// Development environment configuration
	DevEnv *DevEnvConfig `yaml:"devEnv,omitempty" json:"devEnv,omitempty"`

	// Network environment configuration
	NetEnv *NetEnvConfig `yaml:"netEnv,omitempty" json:"netEnv,omitempty"`

	// SSH configuration management
	SSHConfig *SSHConfigSettings `yaml:"sshConfig,omitempty" json:"sshConfig,omitempty"`
}

// GlobalSettings contains settings that apply across all providers.
type GlobalSettings struct {
	// Default clone directory base path
	CloneBaseDir string `yaml:"clone_base_dir,omitempty" json:"cloneBaseDir,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Default strategy for repository operations
	DefaultStrategy string `yaml:"default_strategy,omitempty" json:"defaultStrategy,omitempty" validate:"omitempty,oneof=reset pull fetch"` //nolint:tagliatelle,revive // YAML compatibility and custom validation tags

	// Global ignore patterns (regex)
	GlobalIgnores []string `yaml:"global_ignores,omitempty" json:"globalIgnores,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Default visibility filter
	DefaultVisibility string `yaml:"default_visibility,omitempty" json:"defaultVisibility,omitempty" validate:"omitempty,oneof=public private all"` //nolint:tagliatelle,revive // YAML compatibility and custom validation tags

	// Timeout settings
	Timeouts *TimeoutSettings `yaml:"timeouts,omitempty" json:"timeouts,omitempty"`

	// Concurrency settings
	Concurrency *ConcurrencySettings `yaml:"concurrency,omitempty" json:"concurrency,omitempty"`
}

// TimeoutSettings contains timeout configurations.
type TimeoutSettings struct {
	// HTTP request timeout
	HTTPTimeout time.Duration `yaml:"http_timeout,omitempty" json:"httpTimeout,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Git operation timeout
	GitTimeout time.Duration `yaml:"git_timeout,omitempty" json:"gitTimeout,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// API rate limit timeout
	RateLimitTimeout time.Duration `yaml:"rate_limit_timeout,omitempty" json:"rateLimitTimeout,omitempty"` //nolint:tagliatelle // YAML compatibility required
}

// ConcurrencySettings contains concurrency configurations.
type ConcurrencySettings struct {
	// Maximum concurrent clone operations
	CloneWorkers int `yaml:"clone_workers,omitempty" json:"cloneWorkers,omitempty" validate:"omitempty,min=1,max=50"` //nolint:tagliatelle,revive // YAML compatibility and custom validation tags

	// Maximum concurrent update operations
	UpdateWorkers int `yaml:"update_workers,omitempty" json:"updateWorkers,omitempty" validate:"omitempty,min=1,max=50"` //nolint:tagliatelle,revive // YAML compatibility and custom validation tags

	// Maximum concurrent API operations
	APIWorkers int `yaml:"api_workers,omitempty" json:"apiWorkers,omitempty" validate:"omitempty,min=1,max=20"` //nolint:tagliatelle // YAML compatibility required
}

// ProviderConfig represents configuration for a specific Git provider.
type ProviderConfig struct {
	// Authentication token (supports environment variables)
	Token string `yaml:"token,omitempty" json:"token,omitempty" validate:"required,envtoken"` //nolint:revive // Custom validation tag for environment token

	// API endpoint URL (for self-hosted instances)
	APIURL string `yaml:"api_url,omitempty" json:"apiUrl,omitempty" validate:"omitempty,url"` //nolint:tagliatelle // YAML compatibility required

	// Organizations/groups to manage
	Organizations []*OrganizationConfig `yaml:"organizations,omitempty" json:"organizations,omitempty" validate:"min=1"`

	// Provider-specific settings
	Settings *ProviderSettings `yaml:"settings,omitempty" json:"settings,omitempty"`

	// Legacy support for bulk-clone.yaml format
	Legacy *LegacyProviderConfig `yaml:"legacy,omitempty" json:"legacy,omitempty"`
}

// OrganizationConfig represents configuration for an organization/group.
type OrganizationConfig struct {
	// Organization/group name
	Name string `yaml:"name" json:"name" validate:"required"`

	// Clone directory for this organization
	CloneDir string `yaml:"clone_dir" json:"cloneDir" validate:"required,dirpath"` //nolint:tagliatelle // YAML compatibility required

	// Repository visibility filter
	Visibility string `yaml:"visibility,omitempty" json:"visibility,omitempty" validate:"omitempty,oneof=public private all"`

	// Update strategy for existing repositories
	Strategy string `yaml:"strategy,omitempty" json:"strategy,omitempty" validate:"omitempty,oneof=reset pull fetch"`

	// Include pattern (regex)
	Include string `yaml:"include,omitempty" json:"include,omitempty" validate:"omitempty,regexpattern"`

	// Exclude patterns (regex)
	Exclude []string `yaml:"exclude,omitempty" json:"exclude,omitempty" validate:"dive,regexpattern"` //nolint:revive // Custom validation tag for regex patterns

	// Whether to flatten directory structure
	Flatten bool `yaml:"flatten,omitempty" json:"flatten,omitempty"`

	// Recursive processing (for GitLab groups)
	Recursive bool `yaml:"recursive,omitempty" json:"recursive,omitempty"`

	// Repository management settings
	RepoManagement *RepoManagementConfig `yaml:"repo_management,omitempty" json:"repoManagement,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Custom labels for organization
	Labels map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// ProviderSettings contains provider-specific settings.
type ProviderSettings struct {
	// Rate limiting settings
	RateLimit *RateLimitConfig `yaml:"rate_limit,omitempty" json:"rateLimit,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Retry settings
	Retry *RetryConfig `yaml:"retry,omitempty" json:"retry,omitempty"`

	// Authentication settings
	Auth *AuthConfig `yaml:"auth,omitempty" json:"auth,omitempty"`
}

// RateLimitConfig contains rate limiting configuration.
type RateLimitConfig struct {
	// Requests per hour
	RequestsPerHour int `yaml:"requests_per_hour,omitempty" json:"requestsPerHour,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Burst limit
	BurstLimit int `yaml:"burst_limit,omitempty" json:"burstLimit,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Enable automatic rate limit detection
	AutoDetect bool `yaml:"auto_detect,omitempty" json:"autoDetect,omitempty"` //nolint:tagliatelle // YAML compatibility required
}

// RetryConfig contains retry configuration.
type RetryConfig struct {
	// Maximum retry attempts
	MaxAttempts int `yaml:"max_attempts,omitempty" json:"maxAttempts,omitempty" validate:"omitempty,min=0,max=10"` //nolint:tagliatelle // YAML compatibility required

	// Base delay between retries
	BaseDelay time.Duration `yaml:"base_delay,omitempty" json:"baseDelay,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Maximum delay between retries
	MaxDelay time.Duration `yaml:"max_delay,omitempty" json:"maxDelay,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Enable exponential backoff
	ExponentialBackoff bool `yaml:"exponential_backoff,omitempty" json:"exponentialBackoff,omitempty"` //nolint:tagliatelle // YAML compatibility required
}

// AuthConfig contains authentication configuration.
type AuthConfig struct {
	// Token environment variable name
	TokenEnvVar string `yaml:"token_env_var,omitempty" json:"tokenEnvVar,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// SSH key path for Git operations
	SSHKeyPath string `yaml:"ssh_key_path,omitempty" json:"sshKeyPath,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Use SSH for Git operations
	UseSSH bool `yaml:"use_ssh,omitempty" json:"useSsh,omitempty"` //nolint:tagliatelle // YAML compatibility required
}

// RepoManagementConfig contains repository management settings.
type RepoManagementConfig struct {
	// Enable repository configuration management
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`

	// Configuration templates to apply
	Templates []string `yaml:"templates,omitempty" json:"templates,omitempty"`

	// Branch protection settings
	BranchProtection *BranchProtectionConfig `yaml:"branch_protection,omitempty" json:"branchProtection,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Security settings
	Security *SecurityConfig `yaml:"security,omitempty" json:"security,omitempty"`
}

// BranchProtectionConfig contains branch protection settings.
type BranchProtectionConfig struct {
	// Enable branch protection
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`

	// Protected branches (patterns)
	Branches []string `yaml:"branches,omitempty" json:"branches,omitempty"`

	// Require status checks
	RequireStatusChecks bool `yaml:"require_status_checks,omitempty" json:"requireStatusChecks,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Require pull request reviews
	RequirePRReviews bool `yaml:"require_pr_reviews,omitempty" json:"requirePrReviews,omitempty"` //nolint:tagliatelle // YAML compatibility required
}

// SecurityConfig contains security settings.
type SecurityConfig struct {
	// Enable vulnerability alerts
	VulnerabilityAlerts bool `yaml:"vulnerability_alerts,omitempty" json:"vulnerabilityAlerts,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Enable automated security fixes
	AutomatedSecurityFixes bool `yaml:"automated_security_fixes,omitempty" json:"automatedSecurityFixes,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Required security policies
	RequiredPolicies []string `yaml:"required_policies,omitempty" json:"requiredPolicies,omitempty"` //nolint:tagliatelle // YAML compatibility required
}

// LegacyProviderConfig supports migration from bulk-clone.yaml format.
type LegacyProviderConfig struct {
	// Legacy root path
	RootPath string `yaml:"root_path,omitempty" json:"rootPath,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Legacy protocol
	Protocol string `yaml:"protocol,omitempty" json:"protocol,omitempty"`

	// Legacy organization name
	OrgName string `yaml:"org_name,omitempty" json:"orgName,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Legacy group name (GitLab)
	GroupName string `yaml:"group_name,omitempty" json:"groupName,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Legacy URL (for GitLab)
	URL string `yaml:"url,omitempty" json:"url,omitempty"`
}

// MigrationInfo contains information about configuration migration.
type MigrationInfo struct {
	// Source format that was migrated from
	SourceFormat string `yaml:"source_format,omitempty" json:"sourceFormat,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Migration date
	MigrationDate time.Time `yaml:"migration_date,omitempty" json:"migrationDate,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Original configuration file path
	SourcePath string `yaml:"source_path,omitempty" json:"sourcePath,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Migration tool version
	ToolVersion string `yaml:"tool_version,omitempty" json:"toolVersion,omitempty"` //nolint:tagliatelle // YAML compatibility required
}

// DefaultUnifiedConfig returns a default unified configuration.
func DefaultUnifiedConfig() *UnifiedConfig {
	return &UnifiedConfig{
		Version:         "1.0.0",
		DefaultProvider: "github",
		Global: &GlobalSettings{
			CloneBaseDir:      "$HOME/repos",
			DefaultStrategy:   "reset",
			DefaultVisibility: "all",
			Timeouts: &TimeoutSettings{
				HTTPTimeout:      30 * time.Second,
				GitTimeout:       5 * time.Minute,
				RateLimitTimeout: 1 * time.Hour,
			},
			Concurrency: &ConcurrencySettings{
				CloneWorkers:  10,
				UpdateWorkers: 15,
				APIWorkers:    5,
			},
		},
		Providers: make(map[string]*ProviderConfig),
		IDE: &IDEConfig{
			Enabled:           true,
			WatchDirectories:  []string{"$HOME/.config", "$HOME/.local/share/JetBrains"},
			ExcludePatterns:   []string{`\.git/.*`, `node_modules/.*`, `\.DS_Store`},
			JetBrainsProducts: []string{"IntelliJ", "PyCharm", "GoLand", "WebStorm"},
			AutoFixSync:       true,
			SyncSettings: &IDESyncSettings{
				Enabled:          true,
				Interval:         5 * time.Minute,
				SyncTypes:        []string{"keymap", "editor", "ui", "plugins"},
				BackupBeforeSync: true,
			},
			Logging: &IDELoggingConfig{
				Level:    "info",
				FilePath: "$HOME/.local/share/gzh-manager/logs/ide.log",
				Console:  true,
				Rotation: &LogRotationConfig{
					MaxSizeMB:  10,
					MaxBackups: 5,
					MaxAgeDays: 30,
					Compress:   true,
				},
			},
		},
		DevEnv: &DevEnvConfig{
			Enabled:        true,
			BackupLocation: "$HOME/.gz/backups",
			AutoBackup:     true,
			Providers: &DevEnvProviders{
				AWS: &AWSConfig{
					DefaultProfile:   "default",
					PreferredRegions: []string{"us-west-2", "us-east-1"},
					CredentialsFile:  "$HOME/.aws/credentials",
					ConfigFile:       "$HOME/.aws/config",
					EnableMFA:        false,
				},
				GCP: &GCPConfig{
					DefaultProject:   "",
					PreferredRegions: []string{"us-central1", "us-west1"},
					UseADC:           true,
				},
				Azure: &AzureConfig{
					PreferredRegions:   []string{"westus2", "eastus"},
					UseManagedIdentity: false,
				},
			},
			Containers: &ContainerConfig{
				DefaultRuntime: "docker",
				Docker: &DockerConfig{
					SocketPath:      "/var/run/docker.sock",
					DefaultRegistry: "docker.io",
					BuildOptions: &DockerBuildOptions{
						DefaultContext: ".",
						EnableBuildKit: true,
					},
				},
			},
			Kubernetes: &KubernetesConfig{
				KubeconfigPath:   "$HOME/.kube/config",
				DefaultNamespace: "default",
				AutoDiscovery:    true,
			},
			Backup: &BackupConfig{
				Enabled:         false,
				Interval:        24 * time.Hour,
				RetentionPeriod: 30 * 24 * time.Hour, // 30 days
				Compression:     "gzip",
				Encryption: &BackupEncryption{
					Enabled: false,
					Method:  "aes256",
				},
			},
		},
		NetEnv: &NetEnvConfig{
			Enabled: true,
			WiFiMonitoring: &WiFiMonitoringConfig{
				Enabled:  true,
				Interval: 5 * time.Second,
			},
			VPN: &VPNConfig{
				AutoConnect: &VPNAutoConnect{
					Enabled:             true,
					OnUntrustedNetworks: true,
					RetryAttempts:       3,
					RetryDelay:          5 * time.Second,
				},
			},
			DNS: &DNSConfig{
				DefaultServers: []string{"1.1.1.1", "1.0.0.1"},
				EnableDoH:      false,
				DoHProvider:    "cloudflare",
			},
			Proxy: &ProxyConfig{
				AutoConfigure: false,
			},
			Actions: &NetworkActions{
				OnNetworkChange:  []string{"update-dns", "check-vpn"},
				OnWiFiConnect:    []string{"sync-time"},
				OnWiFiDisconnect: []string{"pause-sync"},
			},
			Daemon: &DaemonConfig{
				Enabled:            false,
				PIDFile:            "/var/run/gzh-manager-netenv.pid",
				LogFile:            "/var/log/gzh-manager-netenv.log",
				LogLevel:           "info",
				WorkingDir:         "/",
				SystemdIntegration: true,
			},
		},
		SSHConfig: &SSHConfigSettings{
			Enabled:       true,
			ConfigFile:    "$HOME/.ssh/config",
			BackupEnabled: true,
			BackupDir:     "$HOME/.ssh/backups",
			KeyManagement: &SSHKeyManagement{
				Enabled:        true,
				KeyDir:         "$HOME/.ssh",
				DefaultKeyType: "ed25519",
				UseSSHAgent:    true,
			},
		},
	}
}

// SupportedProviders returns a list of supported Git providers.
func SupportedProviders() []string {
	return []string{"github", "gitlab", "gitea", "gogs"}
}

// ValidateProvider checks if a provider name is supported.
func ValidateProvider(provider string) bool {
	for _, supported := range SupportedProviders() {
		if provider == supported {
			return true
		}
	}

	return false
}

// IDEConfig contains configuration for IDE monitoring and management.
type IDEConfig struct {
	// Enable IDE monitoring
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`

	// Directories to watch for IDE settings changes
	WatchDirectories []string `yaml:"watch_directories,omitempty" json:"watchDirectories,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Patterns to exclude from monitoring (regex)
	ExcludePatterns []string `yaml:"exclude_patterns,omitempty" json:"excludePatterns,omitempty" validate:"dive,regexpattern"` //nolint:tagliatelle // YAML compatibility required

	// JetBrains products to monitor
	JetBrainsProducts []string `yaml:"jetbrains_products,omitempty" json:"jetbrainsProducts,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Enable automatic sync fixes
	AutoFixSync bool `yaml:"auto_fix_sync,omitempty" json:"autoFixSync,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Settings synchronization configuration
	SyncSettings *IDESyncSettings `yaml:"sync_settings,omitempty" json:"syncSettings,omitempty"` //nolint:tagliatelle // YAML compatibility required

	// Logging configuration
	Logging *IDELoggingConfig `yaml:"logging,omitempty" json:"logging,omitempty"`
}

// IDESyncSettings contains IDE synchronization settings.
type IDESyncSettings struct {
	// Enable settings sync
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`

	// Sync interval
	Interval time.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`

	// Settings types to sync
	SyncTypes []string `yaml:"syncTypes,omitempty" json:"syncTypes,omitempty"`

	// Backup settings before sync
	BackupBeforeSync bool `yaml:"backupBeforeSync,omitempty" json:"backupBeforeSync,omitempty"`
}

// IDELoggingConfig contains IDE logging configuration.
type IDELoggingConfig struct {
	// Log level
	Level string `yaml:"level,omitempty" json:"level,omitempty" validate:"omitempty,oneof=debug info warn error"`

	// Log file path
	FilePath string `yaml:"filePath,omitempty" json:"filePath,omitempty"`

	// Enable console logging
	Console bool `yaml:"console,omitempty" json:"console,omitempty"`

	// Log rotation settings
	Rotation *LogRotationConfig `yaml:"rotation,omitempty" json:"rotation,omitempty"`
}

// LogRotationConfig contains log rotation settings.
type LogRotationConfig struct {
	// Maximum log file size in MB
	MaxSizeMB int `yaml:"maxSizeMb,omitempty" json:"maxSizeMb,omitempty"`

	// Maximum number of backup files
	MaxBackups int `yaml:"maxBackups,omitempty" json:"maxBackups,omitempty"`

	// Maximum age in days
	MaxAgeDays int `yaml:"maxAgeDays,omitempty" json:"maxAgeDays,omitempty"`

	// Compress backup files
	Compress bool `yaml:"compress,omitempty" json:"compress,omitempty"`
}

// DevEnvConfig contains configuration for development environment management.
type DevEnvConfig struct {
	// Enable development environment management
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`

	// Default backup location
	BackupLocation string `yaml:"backupLocation,omitempty" json:"backupLocation,omitempty"`

	// Enable automatic backups
	AutoBackup bool `yaml:"autoBackup,omitempty" json:"autoBackup,omitempty"`

	// Cloud provider configurations
	Providers *DevEnvProviders `yaml:"providers,omitempty" json:"providers,omitempty"`

	// Container configurations
	Containers *ContainerConfig `yaml:"containers,omitempty" json:"containers,omitempty"`

	// Kubernetes configurations
	Kubernetes *KubernetesConfig `yaml:"kubernetes,omitempty" json:"kubernetes,omitempty"`

	// Backup settings
	Backup *BackupConfig `yaml:"backup,omitempty" json:"backup,omitempty"`
}

// DevEnvProviders contains cloud provider configurations.
type DevEnvProviders struct {
	// AWS configuration
	AWS *AWSConfig `yaml:"aws,omitempty" json:"aws,omitempty"`

	// Google Cloud configuration
	GCP *GCPConfig `yaml:"gcp,omitempty" json:"gcp,omitempty"`

	// Azure configuration
	Azure *AzureConfig `yaml:"azure,omitempty" json:"azure,omitempty"`
}

// AWSConfig contains AWS configuration settings.
type AWSConfig struct {
	// Default AWS profile
	DefaultProfile string `yaml:"defaultProfile,omitempty" json:"defaultProfile,omitempty"`

	// AWS region preferences
	PreferredRegions []string `yaml:"preferredRegions,omitempty" json:"preferredRegions,omitempty"`

	// Credential file path
	CredentialsFile string `yaml:"credentialsFile,omitempty" json:"credentialsFile,omitempty"`

	// Config file path
	ConfigFile string `yaml:"configFile,omitempty" json:"configFile,omitempty"`

	// Enable MFA
	EnableMFA bool `yaml:"enableMfa,omitempty" json:"enableMfa,omitempty"`
}

// GCPConfig contains Google Cloud configuration settings.
type GCPConfig struct {
	// Default GCP project
	DefaultProject string `yaml:"defaultProject,omitempty" json:"defaultProject,omitempty"`

	// Service account key file
	ServiceAccountKey string `yaml:"serviceAccountKey,omitempty" json:"serviceAccountKey,omitempty"`

	// Preferred regions
	PreferredRegions []string `yaml:"preferredRegions,omitempty" json:"preferredRegions,omitempty"`

	// Enable application default credentials
	UseADC bool `yaml:"useAdc,omitempty" json:"useAdc,omitempty"`
}

// AzureConfig contains Azure configuration settings.
type AzureConfig struct {
	// Default subscription ID
	DefaultSubscription string `yaml:"defaultSubscription,omitempty" json:"defaultSubscription,omitempty"`

	// Default tenant ID
	DefaultTenant string `yaml:"defaultTenant,omitempty" json:"defaultTenant,omitempty"`

	// Preferred regions
	PreferredRegions []string `yaml:"preferredRegions,omitempty" json:"preferredRegions,omitempty"`

	// Use managed identity
	UseManagedIdentity bool `yaml:"useManagedIdentity,omitempty" json:"useManagedIdentity,omitempty"`
}

// ContainerConfig contains container configuration settings.
type ContainerConfig struct {
	// Docker configuration
	Docker *DockerConfig `yaml:"docker,omitempty" json:"docker,omitempty"`

	// Podman configuration
	Podman *PodmanConfig `yaml:"podman,omitempty" json:"podman,omitempty"`

	// Default container runtime
	DefaultRuntime string `yaml:"defaultRuntime,omitempty" json:"defaultRuntime,omitempty" validate:"omitempty,oneof=docker podman"`
}

// DockerConfig contains Docker configuration settings.
type DockerConfig struct {
	// Docker daemon socket
	SocketPath string `yaml:"socket_path,omitempty" json:"socket_path,omitempty"`

	// Default registry
	DefaultRegistry string `yaml:"default_registry,omitempty" json:"default_registry,omitempty"`

	// Registry authentication
	RegistryAuth map[string]*RegistryAuthConfig `yaml:"registry_auth,omitempty" json:"registry_auth,omitempty"`

	// Build options
	BuildOptions *DockerBuildOptions `yaml:"build_options,omitempty" json:"build_options,omitempty"`
}

// PodmanConfig contains Podman configuration settings.
type PodmanConfig struct {
	// Podman socket path
	SocketPath string `yaml:"socket_path,omitempty" json:"socket_path,omitempty"`

	// Remote connections
	RemoteConnections map[string]*PodmanRemoteConfig `yaml:"remote_connections,omitempty" json:"remote_connections,omitempty"`

	// Default connection
	DefaultConnection string `yaml:"default_connection,omitempty" json:"default_connection,omitempty"`
}

// RegistryAuthConfig contains registry authentication settings.
type RegistryAuthConfig struct {
	// Username
	Username string `yaml:"username,omitempty" json:"username,omitempty"`

	// Password (supports environment variables)
	Password string `yaml:"password,omitempty" json:"password,omitempty" validate:"omitempty,envtoken"`

	// Email
	Email string `yaml:"email,omitempty" json:"email,omitempty"`

	// Auth token
	Auth string `yaml:"auth,omitempty" json:"auth,omitempty"`
}

// DockerBuildOptions contains Docker build options.
type DockerBuildOptions struct {
	// Default build context
	DefaultContext string `yaml:"default_context,omitempty" json:"default_context,omitempty"`

	// Build arguments
	BuildArgs map[string]string `yaml:"build_args,omitempty" json:"build_args,omitempty"`

	// Target stage for multi-stage builds
	Target string `yaml:"target,omitempty" json:"target,omitempty"`

	// Enable BuildKit
	EnableBuildKit bool `yaml:"enable_buildkit,omitempty" json:"enable_buildkit,omitempty"`
}

// PodmanRemoteConfig contains Podman remote connection settings.
type PodmanRemoteConfig struct {
	// Remote host
	Host string `yaml:"host,omitempty" json:"host,omitempty"`

	// SSH identity file
	Identity string `yaml:"identity,omitempty" json:"identity,omitempty"`

	// SSH port
	Port int `yaml:"port,omitempty" json:"port,omitempty"`

	// Use SSH tunnel
	UseSSH bool `yaml:"use_ssh,omitempty" json:"useSsh,omitempty"` //nolint:tagliatelle // YAML compatibility required
}

// KubernetesConfig contains Kubernetes configuration settings.
type KubernetesConfig struct {
	// Default context
	DefaultContext string `yaml:"default_context,omitempty" json:"default_context,omitempty"`

	// Kubeconfig file path
	KubeconfigPath string `yaml:"kubeconfig_path,omitempty" json:"kubeconfig_path,omitempty"`

	// Cluster configurations
	Clusters map[string]*K8sClusterConfig `yaml:"clusters,omitempty" json:"clusters,omitempty"`

	// Namespace preferences
	DefaultNamespace string `yaml:"default_namespace,omitempty" json:"default_namespace,omitempty"`

	// Enable auto-discovery
	AutoDiscovery bool `yaml:"auto_discovery,omitempty" json:"auto_discovery,omitempty"`
}

// K8sClusterConfig contains Kubernetes cluster configuration.
type K8sClusterConfig struct {
	// Cluster endpoint
	Server string `yaml:"server,omitempty" json:"server,omitempty"`

	// Certificate authority data
	CertificateAuthority string `yaml:"certificate_authority,omitempty" json:"certificate_authority,omitempty"`

	// Skip TLS verification
	InsecureSkipTLSVerify bool `yaml:"insecure_skip_tls_verify,omitempty" json:"insecure_skip_tls_verify,omitempty"`

	// Authentication method
	AuthMethod string `yaml:"auth_method,omitempty" json:"auth_method,omitempty" validate:"omitempty,oneof=token certificate exec"`

	// Authentication configuration
	AuthConfig map[string]interface{} `yaml:"auth_config,omitempty" json:"auth_config,omitempty"`
}

// BackupConfig contains backup configuration settings.
type BackupConfig struct {
	// Enable automatic backups
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`

	// Backup interval
	Interval time.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`

	// Backup retention period
	RetentionPeriod time.Duration `yaml:"retention_period,omitempty" json:"retention_period,omitempty"`

	// Backup compression
	Compression string `yaml:"compression,omitempty" json:"compression,omitempty" validate:"omitempty,oneof=none gzip xz"`

	// Backup destinations
	Destinations []string `yaml:"destinations,omitempty" json:"destinations,omitempty"`

	// Encryption settings
	Encryption *BackupEncryption `yaml:"encryption,omitempty" json:"encryption,omitempty"`
}

// BackupEncryption contains backup encryption settings.
type BackupEncryption struct {
	// Enable encryption
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`

	// Encryption method
	Method string `yaml:"method,omitempty" json:"method,omitempty" validate:"omitempty,oneof=aes256 gpg"`

	// Encryption key path
	KeyPath string `yaml:"key_path,omitempty" json:"key_path,omitempty"`

	// GPG recipient (for GPG encryption)
	GPGRecipient string `yaml:"gpg_recipient,omitempty" json:"gpg_recipient,omitempty"`
}

// NetEnvConfig contains configuration for network environment management.
type NetEnvConfig struct {
	// Enable network environment monitoring
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`

	// WiFi monitoring settings
	WiFiMonitoring *WiFiMonitoringConfig `yaml:"wifi_monitoring,omitempty" json:"wifi_monitoring,omitempty"`

	// VPN configurations
	VPN *VPNConfig `yaml:"vpn,omitempty" json:"vpn,omitempty"`

	// DNS configurations
	DNS *DNSConfig `yaml:"dns,omitempty" json:"dns,omitempty"`

	// Proxy configurations
	Proxy *ProxyConfig `yaml:"proxy,omitempty" json:"proxy,omitempty"`

	// Network actions
	Actions *NetworkActions `yaml:"actions,omitempty" json:"actions,omitempty"`

	// Daemon settings
	Daemon *DaemonConfig `yaml:"daemon,omitempty" json:"daemon,omitempty"`
}

// WiFiMonitoringConfig contains WiFi monitoring settings.
type WiFiMonitoringConfig struct {
	// Enable WiFi monitoring
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`

	// Monitoring interval
	Interval time.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`

	// Known networks with configurations
	KnownNetworks map[string]*NetworkProfile `yaml:"known_networks,omitempty" json:"known_networks,omitempty"`

	// Default actions on network change
	DefaultActions []string `yaml:"default_actions,omitempty" json:"default_actions,omitempty"`
}

// NetworkProfile contains network-specific configuration.
type NetworkProfile struct {
	// Network SSID
	SSID string `yaml:"ssid,omitempty" json:"ssid,omitempty"`

	// Network type (home, work, public)
	Type string `yaml:"type,omitempty" json:"type,omitempty" validate:"omitempty,oneof=home work public"`

	// VPN configuration for this network
	VPNConfig string `yaml:"vpn_config,omitempty" json:"vpn_config,omitempty"`

	// DNS servers for this network
	DNSServers []string `yaml:"dns_servers,omitempty" json:"dns_servers,omitempty"`

	// Proxy configuration for this network
	ProxyConfig string `yaml:"proxy_config,omitempty" json:"proxy_config,omitempty"`

	// Actions to execute when connecting to this network
	OnConnect []string `yaml:"on_connect,omitempty" json:"on_connect,omitempty"`

	// Actions to execute when disconnecting from this network
	OnDisconnect []string `yaml:"on_disconnect,omitempty" json:"on_disconnect,omitempty"`
}

// VPNConfig contains VPN configuration settings.
type VPNConfig struct {
	// VPN profiles
	Profiles map[string]*VPNProfile `yaml:"profiles,omitempty" json:"profiles,omitempty"`

	// Default VPN profile
	DefaultProfile string `yaml:"defaultProfile,omitempty" json:"defaultProfile,omitempty"`

	// Auto-connect settings
	AutoConnect *VPNAutoConnect `yaml:"auto_connect,omitempty" json:"auto_connect,omitempty"`
}

// VPNProfile contains VPN profile settings.
type VPNProfile struct {
	// VPN type (openvpn, wireguard, etc.)
	Type string `yaml:"type,omitempty" json:"type,omitempty" validate:"omitempty,oneof=openvpn wireguard ipsec"`

	// Configuration file path
	ConfigFile string `yaml:"config_file,omitempty" json:"config_file,omitempty"`

	// Connection command
	ConnectCommand string `yaml:"connect_command,omitempty" json:"connect_command,omitempty"`

	// Disconnection command
	DisconnectCommand string `yaml:"disconnect_command,omitempty" json:"disconnect_command,omitempty"`

	// Status check command
	StatusCommand string `yaml:"status_command,omitempty" json:"status_command,omitempty"`

	// Auto-connect on specific networks
	AutoConnectNetworks []string `yaml:"auto_connect_networks,omitempty" json:"auto_connect_networks,omitempty"`
}

// VPNAutoConnect contains VPN auto-connect settings.
type VPNAutoConnect struct {
	// Enable auto-connect
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`

	// Connect on untrusted networks
	OnUntrustedNetworks bool `yaml:"on_untrusted_networks,omitempty" json:"on_untrusted_networks,omitempty"`

	// Trusted network SSIDs (won't auto-connect)
	TrustedNetworks []string `yaml:"trusted_networks,omitempty" json:"trusted_networks,omitempty"`

	// Retry attempts
	RetryAttempts int `yaml:"retry_attempts,omitempty" json:"retry_attempts,omitempty"`

	// Retry delay
	RetryDelay time.Duration `yaml:"retry_delay,omitempty" json:"retry_delay,omitempty"`
}

// DNSConfig contains DNS configuration settings.
type DNSConfig struct {
	// Default DNS servers
	DefaultServers []string `yaml:"default_servers,omitempty" json:"default_servers,omitempty"`

	// DNS profiles for different networks
	Profiles map[string]*DNSProfile `yaml:"profiles,omitempty" json:"profiles,omitempty"`

	// Enable DNS over HTTPS
	EnableDoH bool `yaml:"enable_doh,omitempty" json:"enable_doh,omitempty"`

	// DoH provider
	DoHProvider string `yaml:"doh_provider,omitempty" json:"doh_provider,omitempty"`

	// Custom DNS mappings
	CustomMappings map[string]string `yaml:"custom_mappings,omitempty" json:"custom_mappings,omitempty"`
}

// DNSProfile contains DNS profile settings.
type DNSProfile struct {
	// DNS servers
	Servers []string `yaml:"servers,omitempty" json:"servers,omitempty"`

	// Search domains
	SearchDomains []string `yaml:"search_domains,omitempty" json:"search_domains,omitempty"`

	// Enable IPv6
	EnableIPv6 bool `yaml:"enable_ipv6,omitempty" json:"enable_ipv6,omitempty"`

	// DNS timeout
	Timeout time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// ProxyConfig contains proxy configuration settings.
type ProxyConfig struct {
	// Proxy profiles
	Profiles map[string]*ProxyProfile `yaml:"profiles,omitempty" json:"profiles,omitempty"`

	// Default proxy profile
	DefaultProfile string `yaml:"defaultProfile,omitempty" json:"defaultProfile,omitempty"`

	// Auto-configure proxy
	AutoConfigure bool `yaml:"auto_configure,omitempty" json:"auto_configure,omitempty"`

	// Proxy auto-config URL
	PACUrl string `yaml:"pac_url,omitempty" json:"pac_url,omitempty"`
}

// ProxyProfile contains proxy profile settings.
type ProxyProfile struct {
	// Proxy type (http, https, socks5)
	Type string `yaml:"type,omitempty" json:"type,omitempty" validate:"omitempty,oneof=http https socks4 socks5"`

	// Proxy host
	Host string `yaml:"host,omitempty" json:"host,omitempty"`

	// Proxy port
	Port int `yaml:"port,omitempty" json:"port,omitempty"`

	// Username for authentication
	Username string `yaml:"username,omitempty" json:"username,omitempty"`

	// Password for authentication
	Password string `yaml:"password,omitempty" json:"password,omitempty" validate:"omitempty,envtoken"`

	// Bypass proxy for these hosts
	NoProxy []string `yaml:"no_proxy,omitempty" json:"no_proxy,omitempty"`

	// Enable proxy for HTTPS
	HTTPSProxy bool `yaml:"https_proxy,omitempty" json:"https_proxy,omitempty"`
}

// NetworkActions contains network action configurations.
type NetworkActions struct {
	// Actions to execute on network change
	OnNetworkChange []string `yaml:"on_network_change,omitempty" json:"on_network_change,omitempty"`

	// Actions to execute on WiFi connect
	OnWiFiConnect []string `yaml:"on_wifi_connect,omitempty" json:"on_wifi_connect,omitempty"`

	// Actions to execute on WiFi disconnect
	OnWiFiDisconnect []string `yaml:"on_wifi_disconnect,omitempty" json:"on_wifi_disconnect,omitempty"`

	// Actions to execute on VPN connect
	OnVPNConnect []string `yaml:"on_vpn_connect,omitempty" json:"on_vpn_connect,omitempty"`

	// Actions to execute on VPN disconnect
	OnVPNDisconnect []string `yaml:"on_vpn_disconnect,omitempty" json:"on_vpn_disconnect,omitempty"`

	// Custom action scripts
	CustomActions map[string]*CustomAction `yaml:"custom_actions,omitempty" json:"custom_actions,omitempty"`
}

// CustomAction contains custom action configuration.
type CustomAction struct {
	// Action name
	Name string `yaml:"name,omitempty" json:"name,omitempty"`

	// Action description
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Command to execute
	Command string `yaml:"command,omitempty" json:"command,omitempty"`

	// Working directory
	WorkingDir string `yaml:"working_dir,omitempty" json:"working_dir,omitempty"`

	// Environment variables
	Environment map[string]string `yaml:"environment,omitempty" json:"environment,omitempty"`

	// Timeout for action execution
	Timeout time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`

	// Run as specific user
	RunAsUser string `yaml:"run_as_user,omitempty" json:"run_as_user,omitempty"`

	// Retry settings
	Retry *ActionRetryConfig `yaml:"retry,omitempty" json:"retry,omitempty"`
}

// ActionRetryConfig contains action retry settings.
type ActionRetryConfig struct {
	// Maximum retry attempts
	MaxAttempts int `yaml:"max_attempts,omitempty" json:"max_attempts,omitempty"`

	// Delay between retries
	Delay time.Duration `yaml:"delay,omitempty" json:"delay,omitempty"`

	// Exponential backoff multiplier
	BackoffMultiplier float64 `yaml:"backoff_multiplier,omitempty" json:"backoff_multiplier,omitempty"`
}

// DaemonConfig contains daemon configuration settings.
type DaemonConfig struct {
	// Enable daemon mode
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`

	// Daemon process ID file
	PIDFile string `yaml:"pid_file,omitempty" json:"pid_file,omitempty"`

	// Daemon log file
	LogFile string `yaml:"log_file,omitempty" json:"log_file,omitempty"`

	// Daemon log level
	LogLevel string `yaml:"log_level,omitempty" json:"log_level,omitempty" validate:"omitempty,oneof=debug info warn error"`

	// Daemon user
	User string `yaml:"user,omitempty" json:"user,omitempty"`

	// Daemon group
	Group string `yaml:"group,omitempty" json:"group,omitempty"`

	// Working directory
	WorkingDir string `yaml:"working_dir,omitempty" json:"working_dir,omitempty"`

	// Enable systemd integration
	SystemdIntegration bool `yaml:"systemd_integration,omitempty" json:"systemd_integration,omitempty"`
}

// SSHConfigSettings contains SSH configuration management settings.
type SSHConfigSettings struct {
	// Enable SSH configuration management
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`

	// SSH config file path
	ConfigFile string `yaml:"config_file,omitempty" json:"config_file,omitempty"`

	// Enable automatic backups
	BackupEnabled bool `yaml:"backup_enabled,omitempty" json:"backup_enabled,omitempty"`

	// Backup directory
	BackupDir string `yaml:"backup_dir,omitempty" json:"backup_dir,omitempty"`

	// Git provider SSH configurations
	ProviderConfigs map[string]*SSHProviderConfig `yaml:"provider_configs,omitempty" json:"provider_configs,omitempty"`

	// SSH key management
	KeyManagement *SSHKeyManagement `yaml:"key_management,omitempty" json:"key_management,omitempty"`

	// Host alias configurations
	HostAliases map[string]*SSHHostAlias `yaml:"host_aliases,omitempty" json:"host_aliases,omitempty"`
}

// SSHProviderConfig contains SSH configuration for Git providers.
type SSHProviderConfig struct {
	// Provider hostname
	Hostname string `yaml:"hostname,omitempty" json:"hostname,omitempty"`

	// SSH user
	User string `yaml:"user,omitempty" json:"user,omitempty"`

	// SSH port
	Port int `yaml:"port,omitempty" json:"port,omitempty"`

	// Identity file (SSH key)
	IdentityFile string `yaml:"identity_file,omitempty" json:"identity_file,omitempty"`

	// Host alias
	HostAlias string `yaml:"host_alias,omitempty" json:"host_alias,omitempty"`

	// SSH options
	Options map[string]string `yaml:"options,omitempty" json:"options,omitempty"`
}

// SSHKeyManagement contains SSH key management settings.
type SSHKeyManagement struct {
	// Enable key management
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`

	// SSH key directory
	KeyDir string `yaml:"key_dir,omitempty" json:"key_dir,omitempty"`

	// Default key type
	DefaultKeyType string `yaml:"default_key_type,omitempty" json:"default_key_type,omitempty" validate:"omitempty,oneof=rsa ed25519 ecdsa"`

	// Default key size (for RSA keys)
	DefaultKeySize int `yaml:"default_key_size,omitempty" json:"default_key_size,omitempty"`

	// Enable SSH agent
	UseSSHAgent bool `yaml:"use_ssh_agent,omitempty" json:"use_ssh_agent,omitempty"`

	// SSH agent socket path
	SSHAgentSocket string `yaml:"ssh_agent_socket,omitempty" json:"ssh_agent_socket,omitempty"`
}

// SSHHostAlias contains SSH host alias configuration.
type SSHHostAlias struct {
	// Real hostname
	RealHostname string `yaml:"real_hostname,omitempty" json:"real_hostname,omitempty"`

	// SSH user
	User string `yaml:"user,omitempty" json:"user,omitempty"`

	// SSH port
	Port int `yaml:"port,omitempty" json:"port,omitempty"`

	// Identity file
	IdentityFile string `yaml:"identity_file,omitempty" json:"identity_file,omitempty"`

	// Additional SSH options
	Options map[string]string `yaml:"options,omitempty" json:"options,omitempty"`
}
