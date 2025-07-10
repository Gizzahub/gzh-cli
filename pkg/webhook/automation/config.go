package automation

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the automation configuration
type Config struct {
	Version string `yaml:"version"`
	Rules   []Rule `yaml:"rules"`
	Global  Global `yaml:"global,omitempty"`
}

// Global represents global configuration settings
type Global struct {
	Enabled          bool                   `yaml:"enabled"`
	DefaultTimeout   string                 `yaml:"default_timeout,omitempty"`
	MaxConcurrency   int                    `yaml:"max_concurrency,omitempty"`
	NotificationURLs map[string]string      `yaml:"notification_urls,omitempty"`
	Variables        map[string]interface{} `yaml:"variables,omitempty"`
}

// LoadConfig loads automation configuration from a file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if config.Version == "" {
		config.Version = "1.0"
	}
	if config.Global.MaxConcurrency == 0 {
		config.Global.MaxConcurrency = 10
	}
	if config.Global.DefaultTimeout == "" {
		config.Global.DefaultTimeout = "30s"
	}

	return &config, nil
}

// LoadConfigFromDirectory loads all automation configs from a directory
func LoadConfigFromDirectory(dir string) ([]*Config, error) {
	var configs []*Config

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if file has .yaml or .yml extension
		ext := filepath.Ext(path)
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		config, err := LoadConfig(path)
		if err != nil {
			return fmt.Errorf("failed to load config %s: %w", path, err)
		}

		configs = append(configs, config)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return configs, nil
}

// ValidateConfig validates the configuration
func ValidateConfig(config *Config) error {
	if config.Version != "1.0" {
		return fmt.Errorf("unsupported config version: %s", config.Version)
	}

	// Validate rules
	ruleIDs := make(map[string]bool)
	for i, rule := range config.Rules {
		if rule.ID == "" {
			return fmt.Errorf("rule[%d] missing ID", i)
		}
		if ruleIDs[rule.ID] {
			return fmt.Errorf("duplicate rule ID: %s", rule.ID)
		}
		ruleIDs[rule.ID] = true

		if rule.Name == "" {
			return fmt.Errorf("rule[%d] missing name", i)
		}

		if len(rule.Conditions) == 0 {
			return fmt.Errorf("rule[%d] has no conditions", i)
		}

		if len(rule.Actions) == 0 {
			return fmt.Errorf("rule[%d] has no actions", i)
		}

		// Validate conditions
		for j, condition := range rule.Conditions {
			if err := validateCondition(condition); err != nil {
				return fmt.Errorf("rule[%d].condition[%d]: %w", i, j, err)
			}
		}

		// Validate actions
		for j, action := range rule.Actions {
			if err := validateAction(action); err != nil {
				return fmt.Errorf("rule[%d].action[%d]: %w", i, j, err)
			}
		}
	}

	return nil
}

// validateCondition validates a single condition
func validateCondition(condition Condition) error {
	if condition.Type == "" {
		return fmt.Errorf("missing type")
	}

	validTypes := []string{"event_type", "repository", "sender", "payload", "time"}
	if !contains(validTypes, condition.Type) {
		return fmt.Errorf("invalid type: %s", condition.Type)
	}

	if condition.Operator == "" {
		return fmt.Errorf("missing operator")
	}

	validOperators := []string{"equals", "==", "not_equals", "!=", "contains", "starts_with", "ends_with", "matches", "in"}
	if !contains(validOperators, condition.Operator) {
		return fmt.Errorf("invalid operator: %s", condition.Operator)
	}

	if condition.Value == nil {
		return fmt.Errorf("missing value")
	}

	return nil
}

// validateAction validates a single action
func validateAction(action Action) error {
	if action.Type == "" {
		return fmt.Errorf("missing type")
	}

	validTypes := []string{"create_issue", "add_label", "create_comment", "merge_pr", "notification", "run_workflow"}
	if !contains(validTypes, action.Type) {
		return fmt.Errorf("invalid type: %s", action.Type)
	}

	if action.Parameters == nil {
		return fmt.Errorf("missing parameters")
	}

	return nil
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// MergeConfigs merges multiple configs into one
func MergeConfigs(configs ...*Config) *Config {
	merged := &Config{
		Version: "1.0",
		Rules:   []Rule{},
		Global: Global{
			Enabled:          true,
			NotificationURLs: make(map[string]string),
			Variables:        make(map[string]interface{}),
		},
	}

	for _, config := range configs {
		if config == nil {
			continue
		}

		// Merge rules
		merged.Rules = append(merged.Rules, config.Rules...)

		// Merge global settings (last one wins for simple values)
		if config.Global.Enabled {
			merged.Global.Enabled = config.Global.Enabled
		}
		if config.Global.DefaultTimeout != "" {
			merged.Global.DefaultTimeout = config.Global.DefaultTimeout
		}
		if config.Global.MaxConcurrency > 0 {
			merged.Global.MaxConcurrency = config.Global.MaxConcurrency
		}

		// Merge notification URLs
		for k, v := range config.Global.NotificationURLs {
			merged.Global.NotificationURLs[k] = v
		}

		// Merge variables
		for k, v := range config.Global.Variables {
			merged.Global.Variables[k] = v
		}
	}

	return merged
}
