package reposync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

// QualityRulesConfig represents custom quality rules configuration
type QualityRulesConfig struct {
	Version  string                    `yaml:"version" json:"version"`
	Rules    map[string]QualityRule    `yaml:"rules" json:"rules"`
	Profiles map[string]QualityProfile `yaml:"profiles" json:"profiles"`
	Excludes []string                  `yaml:"excludes" json:"excludes"`
}

// QualityRule represents a custom quality rule
type QualityRule struct {
	ID          string               `yaml:"id" json:"id"`
	Name        string               `yaml:"name" json:"name"`
	Description string               `yaml:"description" json:"description"`
	Severity    string               `yaml:"severity" json:"severity"`
	Language    string               `yaml:"language" json:"language"`
	Type        string               `yaml:"type" json:"type"` // code-smell, bug, vulnerability, security
	Pattern     string               `yaml:"pattern" json:"pattern"`
	FilePattern string               `yaml:"file_pattern" json:"file_pattern"`
	Message     string               `yaml:"message" json:"message"`
	Tags        []string             `yaml:"tags" json:"tags"`
	Examples    RuleExamples         `yaml:"examples" json:"examples"`
	Parameters  map[string]RuleParam `yaml:"parameters" json:"parameters"`
}

// RuleExamples contains good and bad examples for a rule
type RuleExamples struct {
	Good []string `yaml:"good" json:"good"`
	Bad  []string `yaml:"bad" json:"bad"`
}

// RuleParam represents a rule parameter
type RuleParam struct {
	Type         string      `yaml:"type" json:"type"`
	DefaultValue interface{} `yaml:"default" json:"default"`
	Description  string      `yaml:"description" json:"description"`
	Min          *int        `yaml:"min,omitempty" json:"min,omitempty"`
	Max          *int        `yaml:"max,omitempty" json:"max,omitempty"`
}

// QualityProfile represents a set of rules with specific configurations
type QualityProfile struct {
	Name        string                `yaml:"name" json:"name"`
	Description string                `yaml:"description" json:"description"`
	Extends     string                `yaml:"extends" json:"extends"`
	Rules       map[string]RuleConfig `yaml:"rules" json:"rules"`
}

// RuleConfig represents rule configuration in a profile
type RuleConfig struct {
	Enabled    bool                   `yaml:"enabled" json:"enabled"`
	Severity   string                 `yaml:"severity,omitempty" json:"severity,omitempty"`
	Parameters map[string]interface{} `yaml:"parameters,omitempty" json:"parameters,omitempty"`
}

// CustomRuleEngine applies custom rules to code analysis
type CustomRuleEngine struct {
	config *QualityRulesConfig
	rules  map[string]*compiledRule
}

// compiledRule represents a compiled rule ready for execution
type compiledRule struct {
	rule    QualityRule
	pattern *regexp.Regexp
	filePat *regexp.Regexp
}

// NewCustomRuleEngine creates a new custom rule engine
func NewCustomRuleEngine(configPath string) (*CustomRuleEngine, error) {
	config, err := loadQualityRulesConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load rules config: %w", err)
	}

	engine := &CustomRuleEngine{
		config: config,
		rules:  make(map[string]*compiledRule),
	}

	// Compile rules
	for id, rule := range config.Rules {
		compiled, err := compileRule(rule)
		if err != nil {
			return nil, fmt.Errorf("failed to compile rule %s: %w", id, err)
		}
		engine.rules[id] = compiled
	}

	return engine, nil
}

// ApplyRules applies custom rules to files
func (cre *CustomRuleEngine) ApplyRules(files []string, profile string) ([]QualityIssue, error) {
	issues := make([]QualityIssue, 0)

	// Get active rules for profile
	activeRules := cre.getActiveRules(profile)

	for _, file := range files {
		// Skip excluded files
		if cre.isExcluded(file) {
			continue
		}

		// Read file content
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		// Apply each active rule
		for ruleID, ruleConfig := range activeRules {
			if compiled, exists := cre.rules[ruleID]; exists {
				if compiled.filePat == nil || compiled.filePat.MatchString(file) {
					fileIssues := cre.applyRule(compiled, file, string(content), ruleConfig)
					issues = append(issues, fileIssues...)
				}
			}
		}
	}

	return issues, nil
}

// applyRule applies a single rule to file content
func (cre *CustomRuleEngine) applyRule(rule *compiledRule, file string, content string, config RuleConfig) []QualityIssue {
	issues := make([]QualityIssue, 0)

	if rule.pattern == nil {
		return issues
	}

	// Find all matches
	lines := splitLines(content)
	for lineNum, line := range lines {
		if rule.pattern.MatchString(line) {
			severity := rule.rule.Severity
			if config.Severity != "" {
				severity = config.Severity
			}

			issue := QualityIssue{
				Type:       rule.rule.Type,
				Severity:   severity,
				File:       file,
				Line:       lineNum + 1,
				Column:     0,
				Message:    cre.formatMessage(rule.rule.Message, line),
				Rule:       rule.rule.ID,
				Tool:       "custom-rules",
				Suggestion: cre.getSuggestion(rule.rule),
			}
			issues = append(issues, issue)
		}
	}

	return issues
}

// getActiveRules returns active rules for a profile
func (cre *CustomRuleEngine) getActiveRules(profileName string) map[string]RuleConfig {
	activeRules := make(map[string]RuleConfig)

	// Start with default profile if it exists
	if defaultProfile, exists := cre.config.Profiles["default"]; exists {
		for ruleID, config := range defaultProfile.Rules {
			if config.Enabled {
				activeRules[ruleID] = config
			}
		}
	}

	// Apply specific profile
	if profile, exists := cre.config.Profiles[profileName]; exists {
		// Handle inheritance
		if profile.Extends != "" && profile.Extends != "default" {
			if parentProfile, exists := cre.config.Profiles[profile.Extends]; exists {
				for ruleID, config := range parentProfile.Rules {
					if config.Enabled {
						activeRules[ruleID] = config
					}
				}
			}
		}

		// Apply profile rules
		for ruleID, config := range profile.Rules {
			if config.Enabled {
				activeRules[ruleID] = config
			} else {
				delete(activeRules, ruleID)
			}
		}
	}

	return activeRules
}

// isExcluded checks if a file should be excluded
func (cre *CustomRuleEngine) isExcluded(file string) bool {
	for _, pattern := range cre.config.Excludes {
		if matched, _ := filepath.Match(pattern, file); matched {
			return true
		}
	}
	return false
}

// formatMessage formats rule message with context
func (cre *CustomRuleEngine) formatMessage(template string, context string) string {
	// Simple template replacement
	message := template
	if context != "" {
		message = regexp.MustCompile(`\{\{\.Line\}\}`).ReplaceAllString(message, context)
	}
	return message
}

// getSuggestion returns suggestion for a rule
func (cre *CustomRuleEngine) getSuggestion(rule QualityRule) string {
	if len(rule.Examples.Good) > 0 {
		return fmt.Sprintf("Consider: %s", rule.Examples.Good[0])
	}
	return ""
}

// loadQualityRulesConfig loads rules configuration from file
func loadQualityRulesConfig(configPath string) (*QualityRulesConfig, error) {
	// Try to find config file
	if configPath == "" {
		configPath = findQualityRulesConfig()
	}

	if configPath == "" {
		// Return default config
		return getDefaultQualityRulesConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config QualityRulesConfig
	ext := filepath.Ext(configPath)

	switch ext {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &config)
	case ".json":
		err = json.Unmarshal(data, &config)
	default:
		return nil, fmt.Errorf("unsupported config format: %s", ext)
	}

	if err != nil {
		return nil, err
	}

	return &config, nil
}

// findQualityRulesConfig searches for quality rules config file
func findQualityRulesConfig() string {
	configNames := []string{
		".quality-rules.yaml",
		".quality-rules.yml",
		".quality-rules.json",
		"quality-rules.yaml",
		"quality-rules.yml",
		"quality-rules.json",
	}

	// Check current directory
	for _, name := range configNames {
		if _, err := os.Stat(name); err == nil {
			return name
		}
	}

	// Check config directory
	configDir, _ := os.UserConfigDir()
	if configDir != "" {
		gzhDir := filepath.Join(configDir, "gzh-manager")
		for _, name := range configNames {
			path := filepath.Join(gzhDir, name)
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	return ""
}

// compileRule compiles a rule for execution
func compileRule(rule QualityRule) (*compiledRule, error) {
	compiled := &compiledRule{
		rule: rule,
	}

	// Compile pattern
	if rule.Pattern != "" {
		pattern, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern: %w", err)
		}
		compiled.pattern = pattern
	}

	// Compile file pattern
	if rule.FilePattern != "" {
		filePat, err := regexp.Compile(rule.FilePattern)
		if err != nil {
			return nil, fmt.Errorf("invalid file pattern: %w", err)
		}
		compiled.filePat = filePat
	}

	return compiled, nil
}

// splitLines splits content into lines
func splitLines(content string) []string {
	// Handle different line endings
	content = regexp.MustCompile(`\r\n`).ReplaceAllString(content, "\n")
	content = regexp.MustCompile(`\r`).ReplaceAllString(content, "\n")
	return regexp.MustCompile(`\n`).Split(content, -1)
}

// getDefaultQualityRulesConfig returns default quality rules configuration
func getDefaultQualityRulesConfig() *QualityRulesConfig {
	return &QualityRulesConfig{
		Version: "1.0",
		Rules: map[string]QualityRule{
			"no-console": {
				ID:          "no-console",
				Name:        "No Console Statements",
				Description: "Avoid console statements in production code",
				Severity:    "minor",
				Language:    "javascript",
				Type:        "code-smell",
				Pattern:     `console\.(log|error|warn|info|debug)`,
				Message:     "Remove console statement",
				Tags:        []string{"production", "logging"},
			},
			"no-print": {
				ID:          "no-print",
				Name:        "No Print Statements",
				Description: "Avoid print statements in production code",
				Severity:    "minor",
				Language:    "python",
				Type:        "code-smell",
				Pattern:     `print\s*\(`,
				Message:     "Remove print statement",
				Tags:        []string{"production", "logging"},
			},
			"no-fmt-print": {
				ID:          "no-fmt-print",
				Name:        "No fmt.Print Statements",
				Description: "Use structured logging instead of fmt.Print",
				Severity:    "minor",
				Language:    "go",
				Type:        "code-smell",
				Pattern:     `fmt\.(Print|Printf|Println)`,
				FilePattern: `\.go$`,
				Message:     "Use structured logging (zap, logrus) instead of fmt.Print",
				Tags:        []string{"production", "logging"},
			},
			"todo-fixme": {
				ID:          "todo-fixme",
				Name:        "TODO/FIXME Comments",
				Description: "Track TODO and FIXME comments",
				Severity:    "info",
				Language:    "all",
				Type:        "code-smell",
				Pattern:     `(TODO|FIXME|XXX|HACK|BUG):\s*(.+)`,
				Message:     "Unresolved TODO/FIXME comment",
				Tags:        []string{"maintenance", "technical-debt"},
			},
		},
		Profiles: map[string]QualityProfile{
			"default": {
				Name:        "Default",
				Description: "Default quality profile",
				Rules: map[string]RuleConfig{
					"no-console":   {Enabled: true},
					"no-print":     {Enabled: true},
					"no-fmt-print": {Enabled: true},
					"todo-fixme":   {Enabled: true, Severity: "info"},
				},
			},
			"strict": {
				Name:        "Strict",
				Description: "Strict quality profile",
				Extends:     "default",
				Rules: map[string]RuleConfig{
					"todo-fixme": {Enabled: true, Severity: "minor"},
				},
			},
		},
		Excludes: []string{
			"**/vendor/**",
			"**/node_modules/**",
			"**/.git/**",
			"**/dist/**",
			"**/build/**",
			"**/*_test.go",
			"**/*.test.js",
			"**/*.spec.js",
		},
	}
}
