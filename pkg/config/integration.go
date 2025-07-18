package config

import (
	"fmt"
	"path/filepath"
)

// BulkCloneIntegration provides integration between gzh.yaml config and bulk-clone operations.
type BulkCloneIntegration struct {
	config *Config
}

// NewBulkCloneIntegration creates a new integration instance.
func NewBulkCloneIntegration(config *Config) *BulkCloneIntegration {
	return &BulkCloneIntegration{
		config: config,
	}
}

// BulkCloneTarget represents a target for bulk cloning operations.
type BulkCloneTarget struct {
	Provider   string   // github, gitlab, gitea
	Name       string   // organization or group name
	CloneDir   string   // target directory
	Token      string   // authentication token
	Visibility string   // public, private, all
	Strategy   string   // reset, pull, fetch
	Match      string   // regex pattern for filtering
	Exclude    []string // patterns to exclude
	Recursive  bool     // for GitLab groups
	Flatten    bool     // flatten directory structure
}

// GetAllTargets returns all configured targets for bulk cloning.
func (b *BulkCloneIntegration) GetAllTargets() ([]BulkCloneTarget, error) {
	var targets []BulkCloneTarget

	for providerName, provider := range b.config.Providers {
		// Process organizations (GitHub, Gitea)
		for _, org := range provider.Orgs {
			target := BulkCloneTarget{
				Provider:   providerName,
				Name:       org.Name,
				CloneDir:   b.resolveCloneDir(org.CloneDir, providerName, org.Name, org.Flatten),
				Token:      ExpandEnvironmentVariables(provider.Token),
				Visibility: org.Visibility,
				Strategy:   org.Strategy,
				Match:      org.Match,
				Exclude:    org.Exclude,
				Recursive:  false, // Not applicable for orgs
				Flatten:    org.Flatten,
			}
			targets = append(targets, target)
		}

		// Process groups (GitLab)
		for _, group := range provider.Groups {
			target := BulkCloneTarget{
				Provider:   providerName,
				Name:       group.Name,
				CloneDir:   b.resolveCloneDir(group.CloneDir, providerName, group.Name, group.Flatten),
				Token:      ExpandEnvironmentVariables(provider.Token),
				Visibility: group.Visibility,
				Strategy:   group.Strategy,
				Match:      group.Match,
				Exclude:    group.Exclude,
				Recursive:  group.Recursive,
				Flatten:    group.Flatten,
			}
			targets = append(targets, target)
		}
	}

	return targets, nil
}

// GetTargetsByProvider returns targets filtered by provider.
func (b *BulkCloneIntegration) GetTargetsByProvider(providerName string) ([]BulkCloneTarget, error) {
	allTargets, err := b.GetAllTargets()
	if err != nil {
		return nil, err
	}

	var filtered []BulkCloneTarget

	for _, target := range allTargets {
		if target.Provider == providerName {
			filtered = append(filtered, target)
		}
	}

	return filtered, nil
}

// GetTargetByName returns a specific target by provider and name.
func (b *BulkCloneIntegration) GetTargetByName(providerName, targetName string) (*BulkCloneTarget, error) {
	targets, err := b.GetTargetsByProvider(providerName)
	if err != nil {
		return nil, err
	}

	for _, target := range targets {
		if target.Name == targetName {
			return &target, nil
		}
	}

	return nil, fmt.Errorf("target '%s' not found for provider '%s'", targetName, providerName)
}

// resolveCloneDir resolves the clone directory with fallbacks and flatten support.
func (b *BulkCloneIntegration) resolveCloneDir(cloneDir, providerName, targetName string, flatten bool) string {
	if cloneDir != "" {
		return ExpandEnvironmentVariables(cloneDir)
	}

	// Generate default clone directory
	homeDir := "~"
	if expanded := ExpandEnvironmentVariables("${HOME}"); expanded != "${HOME}" {
		homeDir = expanded
	}

	// Create default path based on flatten option
	var defaultPath string
	if flatten {
		// Flatten: ~/repos/{provider}/
		defaultPath = filepath.Join(homeDir, "repos", providerName)
	} else {
		// Normal: ~/repos/{provider}/{target}
		defaultPath = filepath.Join(homeDir, "repos", providerName, targetName)
	}

	return ExpandEnvironmentVariables(defaultPath)
}

// ValidateProvider checks if a provider is configured.
func (b *BulkCloneIntegration) ValidateProvider(providerName string) error {
	if _, exists := b.config.Providers[providerName]; !exists {
		return fmt.Errorf("provider '%s' is not configured", providerName)
	}

	return nil
}

// GetConfiguredProviders returns a list of all configured providers.
func (b *BulkCloneIntegration) GetConfiguredProviders() []string {
	var providers []string
	for name := range b.config.Providers {
		providers = append(providers, name)
	}

	return providers
}

// GetDefaultProvider returns the default provider or the first available one.
func (b *BulkCloneIntegration) GetDefaultProvider() string {
	if b.config.DefaultProvider != "" {
		return b.config.DefaultProvider
	}

	// Return first configured provider as fallback
	for name := range b.config.Providers {
		return name
	}

	return ProviderGitHub // Ultimate fallback
}

// ShouldProcessTarget determines if a target should be processed based on filters.
func (b *BulkCloneIntegration) ShouldProcessTarget(target BulkCloneTarget, filters map[string]interface{}) bool {
	// Check provider filter
	if providerFilter, ok := filters["provider"]; ok {
		if providerStr, ok := providerFilter.(string); ok && providerStr != target.Provider {
			return false
		}
	}

	// Check visibility filter
	if visibilityFilter, ok := filters["visibility"]; ok {
		if visibilityStr, ok := visibilityFilter.(string); ok {
			if visibilityStr != "all" && visibilityStr != target.Visibility {
				return false
			}
		}
	}

	// Check name pattern filter
	if nameFilter, ok := filters["name_pattern"]; ok {
		if pattern, ok := nameFilter.(string); ok && pattern != "" {
			if matched, _ := CompileRegex(pattern); matched != nil {
				if !matched.MatchString(target.Name) {
					return false
				}
			}
		}
	}

	return true
}

// CreateDefaultGZHConfig creates a default gzh.yaml configuration.
func CreateDefaultGZHConfig(filename string) error {
	defaultConfig := &Config{
		Version:         "1.0.0",
		DefaultProvider: ProviderGitHub,
		Providers: map[string]Provider{
			ProviderGitHub: {
				Token: "${GITHUB_TOKEN}",
				Orgs: []GitTarget{
					{
						Name:       "your-org-name",
						Visibility: VisibilityAll,
						Strategy:   StrategyReset,
						CloneDir:   "~/repos/github/your-org-name",
					},
				},
			},
			ProviderGitLab: {
				Token: "${GITLAB_TOKEN}",
				Groups: []GitTarget{
					{
						Name:       "your-group-name",
						Visibility: VisibilityPublic,
						Strategy:   StrategyReset,
						CloneDir:   "~/repos/gitlab/your-group-name",
						Recursive:  true,
					},
				},
			},
		},
	}

	// Apply defaults
	defaultConfig.applyDefaults()

	// Convert to YAML and save
	return SaveConfigToFile(defaultConfig, filename)
}

// SaveConfigToFile saves a configuration to a YAML file.
func SaveConfigToFile(config *Config, filename string) error {
	// This is a placeholder - would need YAML marshaling
	// For now, we'll create the file with a string template
	content := fmt.Sprintf(`version: "%s"
default_provider: %s

providers:
`, config.Version, config.DefaultProvider)

	for providerName, provider := range config.Providers {
		content += fmt.Sprintf(`  %s:
    token: "%s"
`, providerName, provider.Token)

		if len(provider.Orgs) > 0 {
			content += "    orgs:\n"
			for _, org := range provider.Orgs {
				content += fmt.Sprintf(`      - name: "%s"
        visibility: %s
        strategy: %s
        clone_dir: "%s"
`, org.Name, org.Visibility, org.Strategy, org.CloneDir)
			}
		}

		if len(provider.Groups) > 0 {
			content += "    groups:\n"
			for _, group := range provider.Groups {
				content += fmt.Sprintf(`      - name: "%s"
        visibility: %s
        strategy: %s
        clone_dir: "%s"
        recursive: %t
`, group.Name, group.Visibility, group.Strategy, group.CloneDir, group.Recursive)
			}
		}
	}

	return WriteFileContent(filename, content)
}

// WriteFileContent writes content to a file (helper function).
func WriteFileContent(filename, content string) error {
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := CreateDirectory(dir); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	return WriteFile(filename, content)
}
