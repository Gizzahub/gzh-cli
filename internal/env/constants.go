package env

import "os"

// Get retrieves the value of an environment variable
func Get(key string) string {
	return os.Getenv(key)
}

// Standard environment variable names used by gzh-manager
const (
	// Configuration paths
	ConfigPath      = "GZH_CONFIG_PATH"       // Path to main configuration file
	ConfigDir       = "GZH_CONFIG_DIR"        // Directory containing configuration files
	CloudConfig     = "GZH_CLOUD_CONFIG"      // Path to cloud configuration file
	BulkCloneConfig = "GZH_BULK_CLONE_CONFIG" // Path to bulk-clone configuration file

	// Authentication tokens
	// Note: These follow industry standard naming conventions
	GitHubToken = "GITHUB_TOKEN" // GitHub personal access token
	GitLabToken = "GITLAB_TOKEN" // GitLab personal access token
	GiteaToken  = "GITEA_TOKEN"  // Gitea personal access token

	// Alternative GZH-prefixed tokens (for avoiding conflicts)
	GZHGitHubToken = "GZH_GITHUB_TOKEN" // Alternative GitHub token
	GZHGitLabToken = "GZH_GITLAB_TOKEN" // Alternative GitLab token
	GZHGiteaToken  = "GZH_GITEA_TOKEN"  // Alternative Gitea token

	// Feature flags and settings
	GZHDebug       = "GZH_DEBUG"        // Enable debug mode
	GZHLogLevel    = "GZH_LOG_LEVEL"    // Set log level (debug, info, warn, error)
	GZHNoColor     = "GZH_NO_COLOR"     // Disable colored output
	GZHProgressBar = "GZH_PROGRESS_BAR" // Control progress bar display (auto, always, never)

	// Performance tuning
	GZHMaxWorkers    = "GZH_MAX_WORKERS"    // Maximum number of concurrent workers
	GZHTimeout       = "GZH_TIMEOUT"        // Default timeout for operations
	GZHRetryAttempts = "GZH_RETRY_ATTEMPTS" // Number of retry attempts
	GZHRateLimit     = "GZH_RATE_LIMIT"     // API rate limit per hour

	// Network settings
	GZHHTTPProxy  = "GZH_HTTP_PROXY"  // HTTP proxy URL
	GZHHTTPSProxy = "GZH_HTTPS_PROXY" // HTTPS proxy URL
	GZHNoProxy    = "GZH_NO_PROXY"    // Comma-separated list of hosts to bypass proxy

	// Provider-specific settings
	GZHGitHubAPI = "GZH_GITHUB_API" // GitHub API base URL (for enterprise)
	GZHGitLabAPI = "GZH_GITLAB_API" // GitLab API base URL (for self-hosted)
	GZHGiteaAPI  = "GZH_GITEA_API"  // Gitea API base URL
)

// GetToken returns the token for the specified provider, checking both standard
// and GZH-prefixed environment variables
func GetToken(provider string) string {
	switch provider {
	case "github":
		if token := Get(GZHGitHubToken); token != "" {
			return token
		}
		return Get(GitHubToken)
	case "gitlab":
		if token := Get(GZHGitLabToken); token != "" {
			return token
		}
		return Get(GitLabToken)
	case "gitea":
		if token := Get(GZHGiteaToken); token != "" {
			return token
		}
		return Get(GiteaToken)
	default:
		return ""
	}
}

// GetConfigPath returns the configuration path, checking multiple sources
func GetConfigPath(configType string) string {
	// Check specific config type env var
	envVar := "GZH_" + configType + "_CONFIG"
	if path := Get(envVar); path != "" {
		return path
	}

	// Check general config path
	return Get(ConfigPath)
}
