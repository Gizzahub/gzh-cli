package config

// Config represents the top-level gzh.yaml configuration
type Config struct {
	Version         string              `yaml:"version" json:"version"`
	DefaultProvider string              `yaml:"default_provider,omitempty" json:"default_provider,omitempty"`
	Providers       map[string]Provider `yaml:"providers,omitempty" json:"providers,omitempty"`
}

// Provider represents a Git provider configuration
type Provider struct {
	Token  string      `yaml:"token" json:"token"`
	Orgs   []GitTarget `yaml:"orgs,omitempty" json:"orgs,omitempty"`     // For GitHub/Gitea/Gogs
	Groups []GitTarget `yaml:"groups,omitempty" json:"groups,omitempty"` // For GitLab
}

// GitTarget represents an organization or group configuration
type GitTarget struct {
	Name       string   `yaml:"name" json:"name"`
	Visibility string   `yaml:"visibility,omitempty" json:"visibility,omitempty"` // public, private, all
	Recursive  bool     `yaml:"recursive,omitempty" json:"recursive,omitempty"`   // For GitLab subgroups
	Flatten    bool     `yaml:"flatten,omitempty" json:"flatten,omitempty"`       // Flatten directory structure
	Match      string   `yaml:"match,omitempty" json:"match,omitempty"`           // Regex pattern filter
	CloneDir   string   `yaml:"clone_dir,omitempty" json:"clone_dir,omitempty"`   // Target directory
	Exclude    []string `yaml:"exclude,omitempty" json:"exclude,omitempty"`       // Repos to exclude
	Strategy   string   `yaml:"strategy,omitempty" json:"strategy,omitempty"`     // reset, pull, fetch
}

// Visibility constants
const (
	VisibilityPublic  = "public"
	VisibilityPrivate = "private"
	VisibilityAll     = "all"
)

// Strategy constants
const (
	StrategyReset = "reset"
	StrategyPull  = "pull"
	StrategyFetch = "fetch"
)

// Provider type constants
const (
	ProviderGitHub = "github"
	ProviderGitLab = "gitlab"
	ProviderGitea  = "gitea"
	ProviderGogs   = "gogs"
)

// SetDefaults sets default values for GitTarget
func (g *GitTarget) SetDefaults() {
	if g.Visibility == "" {
		g.Visibility = VisibilityAll
	}
	if g.Strategy == "" {
		g.Strategy = StrategyReset
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Version == "" {
		return ErrMissingVersion
	}
	
	// Validate each provider
	for name, provider := range c.Providers {
		if err := provider.Validate(name); err != nil {
			return err
		}
	}
	
	return nil
}

// Validate checks if the provider configuration is valid
func (p *Provider) Validate(providerType string) error {
	if p.Token == "" {
		return ErrMissingToken
	}
	
	// Validate targets based on provider type
	switch providerType {
	case ProviderGitLab:
		for _, group := range p.Groups {
			if err := group.Validate(); err != nil {
				return err
			}
		}
	case ProviderGitHub, ProviderGitea, ProviderGogs:
		for _, org := range p.Orgs {
			if err := org.Validate(); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// Validate checks if the GitTarget configuration is valid
func (g *GitTarget) Validate() error {
	if g.Name == "" {
		return ErrMissingName
	}
	
	// Validate visibility
	if g.Visibility != "" && g.Visibility != VisibilityAll && 
		g.Visibility != VisibilityPublic && g.Visibility != VisibilityPrivate {
		return ErrInvalidVisibility
	}
	
	// Validate strategy
	if g.Strategy != "" && g.Strategy != StrategyReset && 
		g.Strategy != StrategyPull && g.Strategy != StrategyFetch {
		return ErrInvalidStrategy
	}
	
	// Validate regex pattern if provided
	if g.Match != "" {
		if _, err := CompileRegex(g.Match); err != nil {
			return ErrInvalidRegex
		}
	}
	
	return nil
}