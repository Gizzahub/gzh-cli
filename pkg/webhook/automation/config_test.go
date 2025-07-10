package automation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	configContent := `
version: "1.0"
global:
  enabled: true
  default_timeout: "60s"
  max_concurrency: 5
  notification_urls:
    slack: "https://hooks.slack.com/test"
  variables:
    test_var: "test_value"

rules:
  - id: "test-rule"
    name: "Test Rule"
    description: "A test automation rule"
    enabled: true
    priority: 100
    conditions:
      - type: "event_type"
        operator: "equals"
        value: "push"
    actions:
      - type: "create_issue"
        parameters:
          title: "Test Issue"
          body: "Test issue body"
`

	tmpFile, err := os.CreateTemp("", "automation-config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Load and verify config
	config, err := LoadConfig(tmpFile.Name())
	require.NoError(t, err)

	assert.Equal(t, "1.0", config.Version)
	assert.True(t, config.Global.Enabled)
	assert.Equal(t, "60s", config.Global.DefaultTimeout)
	assert.Equal(t, 5, config.Global.MaxConcurrency)
	assert.Equal(t, "https://hooks.slack.com/test", config.Global.NotificationURLs["slack"])
	assert.Equal(t, "test_value", config.Global.Variables["test_var"])

	assert.Len(t, config.Rules, 1)
	rule := config.Rules[0]
	assert.Equal(t, "test-rule", rule.ID)
	assert.Equal(t, "Test Rule", rule.Name)
	assert.True(t, rule.Enabled)
	assert.Equal(t, 100, rule.Priority)
	assert.Len(t, rule.Conditions, 1)
	assert.Len(t, rule.Actions, 1)
}

func TestLoadConfigDefaults(t *testing.T) {
	// Create a minimal config file
	configContent := `
rules:
  - id: "minimal-rule"
    name: "Minimal Rule"
    enabled: true
    conditions:
      - type: "event_type"
        operator: "equals"
        value: "push"
    actions:
      - type: "create_issue"
        parameters:
          title: "Test Issue"
`

	tmpFile, err := os.CreateTemp("", "minimal-config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Load and verify defaults are set
	config, err := LoadConfig(tmpFile.Name())
	require.NoError(t, err)

	assert.Equal(t, "1.0", config.Version)
	assert.Equal(t, 10, config.Global.MaxConcurrency)
	assert.Equal(t, "30s", config.Global.DefaultTimeout)
}

func TestLoadConfigFromDirectory(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "automation-configs-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create multiple config files
	configs := []string{
		`
version: "1.0"
rules:
  - id: "rule-1"
    name: "Rule 1"
    enabled: true
    conditions:
      - type: "event_type"
        operator: "equals"
        value: "push"
    actions:
      - type: "create_issue"
        parameters:
          title: "Issue 1"
`,
		`
version: "1.0"
rules:
  - id: "rule-2"
    name: "Rule 2"
    enabled: true
    conditions:
      - type: "event_type"
        operator: "equals"
        value: "pull_request"
    actions:
      - type: "add_label"
        parameters:
          labels: ["test"]
`,
	}

	for i, content := range configs {
		filename := filepath.Join(tmpDir, "config"+string(rune('1'+i))+".yaml")
		err := os.WriteFile(filename, []byte(content), 0o644)
		require.NoError(t, err)
	}

	// Create a non-YAML file (should be ignored)
	txtFile := filepath.Join(tmpDir, "readme.txt")
	err = os.WriteFile(txtFile, []byte("This should be ignored"), 0o644)
	require.NoError(t, err)

	// Load configs from directory
	loadedConfigs, err := LoadConfigFromDirectory(tmpDir)
	require.NoError(t, err)

	assert.Len(t, loadedConfigs, 2)

	// Verify both configs were loaded
	ruleIDs := make([]string, len(loadedConfigs))
	for i, config := range loadedConfigs {
		assert.Len(t, config.Rules, 1)
		ruleIDs[i] = config.Rules[0].ID
	}

	assert.Contains(t, ruleIDs, "rule-1")
	assert.Contains(t, ruleIDs, "rule-2")
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				Version: "1.0",
				Rules: []Rule{
					{
						ID:          "valid-rule",
						Name:        "Valid Rule",
						Description: "A valid rule",
						Enabled:     true,
						Priority:    100,
						Conditions: []Condition{
							{
								Type:     "event_type",
								Operator: "equals",
								Value:    "push",
							},
						},
						Actions: []Action{
							{
								Type:       "create_issue",
								Parameters: map[string]interface{}{"title": "Test"},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "unsupported version",
			config: &Config{
				Version: "2.0",
				Rules: []Rule{
					{
						ID:         "test-rule",
						Name:       "Test Rule",
						Enabled:    true,
						Conditions: []Condition{{Type: "event_type", Operator: "equals", Value: "push"}},
						Actions:    []Action{{Type: "create_issue", Parameters: map[string]interface{}{"title": "Test"}}},
					},
				},
			},
			wantErr: true,
			errMsg:  "unsupported config version: 2.0",
		},
		{
			name: "duplicate rule IDs",
			config: &Config{
				Version: "1.0",
				Rules: []Rule{
					{
						ID:         "duplicate-id",
						Name:       "Rule 1",
						Enabled:    true,
						Conditions: []Condition{{Type: "event_type", Operator: "equals", Value: "push"}},
						Actions:    []Action{{Type: "create_issue", Parameters: map[string]interface{}{"title": "Test"}}},
					},
					{
						ID:         "duplicate-id",
						Name:       "Rule 2",
						Enabled:    true,
						Conditions: []Condition{{Type: "event_type", Operator: "equals", Value: "push"}},
						Actions:    []Action{{Type: "create_issue", Parameters: map[string]interface{}{"title": "Test"}}},
					},
				},
			},
			wantErr: true,
			errMsg:  "duplicate rule ID: duplicate-id",
		},
		{
			name: "rule missing ID",
			config: &Config{
				Version: "1.0",
				Rules: []Rule{
					{
						Name:       "Rule without ID",
						Enabled:    true,
						Conditions: []Condition{{Type: "event_type", Operator: "equals", Value: "push"}},
						Actions:    []Action{{Type: "create_issue", Parameters: map[string]interface{}{"title": "Test"}}},
					},
				},
			},
			wantErr: true,
			errMsg:  "rule[0] missing ID",
		},
		{
			name: "rule missing name",
			config: &Config{
				Version: "1.0",
				Rules: []Rule{
					{
						ID:         "test-rule",
						Enabled:    true,
						Conditions: []Condition{{Type: "event_type", Operator: "equals", Value: "push"}},
						Actions:    []Action{{Type: "create_issue", Parameters: map[string]interface{}{"title": "Test"}}},
					},
				},
			},
			wantErr: true,
			errMsg:  "rule[0] missing name",
		},
		{
			name: "rule with no conditions",
			config: &Config{
				Version: "1.0",
				Rules: []Rule{
					{
						ID:         "test-rule",
						Name:       "Test Rule",
						Enabled:    true,
						Conditions: []Condition{},
						Actions:    []Action{{Type: "create_issue", Parameters: map[string]interface{}{"title": "Test"}}},
					},
				},
			},
			wantErr: true,
			errMsg:  "rule[0] has no conditions",
		},
		{
			name: "rule with no actions",
			config: &Config{
				Version: "1.0",
				Rules: []Rule{
					{
						ID:         "test-rule",
						Name:       "Test Rule",
						Enabled:    true,
						Conditions: []Condition{{Type: "event_type", Operator: "equals", Value: "push"}},
						Actions:    []Action{},
					},
				},
			},
			wantErr: true,
			errMsg:  "rule[0] has no actions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCondition(t *testing.T) {
	tests := []struct {
		name      string
		condition Condition
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid condition",
			condition: Condition{
				Type:     "event_type",
				Operator: "equals",
				Value:    "push",
			},
			wantErr: false,
		},
		{
			name: "missing type",
			condition: Condition{
				Operator: "equals",
				Value:    "push",
			},
			wantErr: true,
			errMsg:  "missing type",
		},
		{
			name: "invalid type",
			condition: Condition{
				Type:     "invalid_type",
				Operator: "equals",
				Value:    "push",
			},
			wantErr: true,
			errMsg:  "invalid type: invalid_type",
		},
		{
			name: "missing operator",
			condition: Condition{
				Type:  "event_type",
				Value: "push",
			},
			wantErr: true,
			errMsg:  "missing operator",
		},
		{
			name: "invalid operator",
			condition: Condition{
				Type:     "event_type",
				Operator: "invalid_op",
				Value:    "push",
			},
			wantErr: true,
			errMsg:  "invalid operator: invalid_op",
		},
		{
			name: "missing value",
			condition: Condition{
				Type:     "event_type",
				Operator: "equals",
			},
			wantErr: true,
			errMsg:  "missing value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCondition(tt.condition)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAction(t *testing.T) {
	tests := []struct {
		name    string
		action  Action
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid action",
			action: Action{
				Type:       "create_issue",
				Parameters: map[string]interface{}{"title": "Test"},
			},
			wantErr: false,
		},
		{
			name: "missing type",
			action: Action{
				Parameters: map[string]interface{}{"title": "Test"},
			},
			wantErr: true,
			errMsg:  "missing type",
		},
		{
			name: "invalid type",
			action: Action{
				Type:       "invalid_action",
				Parameters: map[string]interface{}{"title": "Test"},
			},
			wantErr: true,
			errMsg:  "invalid type: invalid_action",
		},
		{
			name: "missing parameters",
			action: Action{
				Type: "create_issue",
			},
			wantErr: true,
			errMsg:  "missing parameters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAction(tt.action)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMergeConfigs(t *testing.T) {
	config1 := &Config{
		Version: "1.0",
		Rules: []Rule{
			{
				ID:         "rule-1",
				Name:       "Rule 1",
				Enabled:    true,
				Conditions: []Condition{{Type: "event_type", Operator: "equals", Value: "push"}},
				Actions:    []Action{{Type: "create_issue", Parameters: map[string]interface{}{"title": "Test 1"}}},
			},
		},
		Global: Global{
			Enabled:        true,
			DefaultTimeout: "30s",
			MaxConcurrency: 5,
			NotificationURLs: map[string]string{
				"slack": "https://slack.com/webhook1",
			},
			Variables: map[string]interface{}{
				"var1": "value1",
			},
		},
	}

	config2 := &Config{
		Version: "1.0",
		Rules: []Rule{
			{
				ID:         "rule-2",
				Name:       "Rule 2",
				Enabled:    true,
				Conditions: []Condition{{Type: "event_type", Operator: "equals", Value: "pull_request"}},
				Actions:    []Action{{Type: "add_label", Parameters: map[string]interface{}{"labels": []string{"test"}}}},
			},
		},
		Global: Global{
			Enabled:        true,
			DefaultTimeout: "60s",
			MaxConcurrency: 10,
			NotificationURLs: map[string]string{
				"discord": "https://discord.com/webhook1",
			},
			Variables: map[string]interface{}{
				"var2": "value2",
			},
		},
	}

	merged := MergeConfigs(config1, config2)

	assert.Equal(t, "1.0", merged.Version)
	assert.Len(t, merged.Rules, 2)

	// Check that both rules are present
	ruleIDs := make([]string, len(merged.Rules))
	for i, rule := range merged.Rules {
		ruleIDs[i] = rule.ID
	}
	assert.Contains(t, ruleIDs, "rule-1")
	assert.Contains(t, ruleIDs, "rule-2")

	// Check global settings (last config wins for simple values)
	assert.True(t, merged.Global.Enabled)
	assert.Equal(t, "60s", merged.Global.DefaultTimeout)
	assert.Equal(t, 10, merged.Global.MaxConcurrency)

	// Check merged notification URLs
	assert.Equal(t, "https://slack.com/webhook1", merged.Global.NotificationURLs["slack"])
	assert.Equal(t, "https://discord.com/webhook1", merged.Global.NotificationURLs["discord"])

	// Check merged variables
	assert.Equal(t, "value1", merged.Global.Variables["var1"])
	assert.Equal(t, "value2", merged.Global.Variables["var2"])
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "item exists",
			slice:    []string{"a", "b", "c"},
			item:     "b",
			expected: true,
		},
		{
			name:     "item does not exist",
			slice:    []string{"a", "b", "c"},
			item:     "d",
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			item:     "a",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}
