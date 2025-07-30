// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package template

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// TemplateEngine handles configuration template processing
type TemplateEngine struct {
	TemplateDir string
	Variables   map[string]interface{}
}

// TemplateConfig represents a template configuration
type TemplateConfig struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Template    map[string]interface{} `yaml:"template"`
	Variables   []TemplateVariable     `yaml:"variables"`
}

// TemplateVariable represents a template variable definition
type TemplateVariable struct {
	Name         string      `yaml:"name"`
	Description  string      `yaml:"description"`
	Required     bool        `yaml:"required"`
	Type         string      `yaml:"type"`
	DefaultValue interface{} `yaml:"default,omitempty"`
	Options      []string    `yaml:"options,omitempty"`
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine(templateDir string) *TemplateEngine {
	return &TemplateEngine{
		TemplateDir: templateDir,
		Variables:   make(map[string]interface{}),
	}
}

// ListTemplates returns available template names
func (te *TemplateEngine) ListTemplates() ([]string, error) {
	var templates []string

	err := filepath.Walk(te.TemplateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if !info.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
			rel, err := filepath.Rel(te.TemplateDir, path)
			if err == nil {
				name := strings.TrimSuffix(rel, filepath.Ext(rel))
				templates = append(templates, name)
			}
		}

		return nil
	})

	return templates, err
}

// LoadTemplate loads a template configuration by name
func (te *TemplateEngine) LoadTemplate(name string) (*TemplateConfig, error) {
	templatePath := te.getTemplatePath(name)

	data, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	var config TemplateConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse template configuration: %w", err)
	}

	return &config, nil
}

// GenerateConfig generates a configuration from a template
func (te *TemplateEngine) GenerateConfig(templateName string, variables map[string]interface{}) (map[string]interface{}, error) {
	templateConfig, err := te.LoadTemplate(templateName)
	if err != nil {
		return nil, err
	}

	// Validate required variables
	if err := te.validateVariables(templateConfig, variables); err != nil {
		return nil, err
	}

	// Apply default values for missing optional variables
	mergedVars := te.mergeWithDefaults(templateConfig, variables)

	// Process template
	processedConfig, err := te.processTemplate(templateConfig.Template, mergedVars)
	if err != nil {
		return nil, fmt.Errorf("failed to process template: %w", err)
	}

	return processedConfig, nil
}

// validateVariables validates that all required variables are provided
func (te *TemplateEngine) validateVariables(config *TemplateConfig, variables map[string]interface{}) error {
	var missingVars []string

	for _, variable := range config.Variables {
		if variable.Required {
			if _, exists := variables[variable.Name]; !exists {
				missingVars = append(missingVars, variable.Name)
			}
		}
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("missing required variables: %s", strings.Join(missingVars, ", "))
	}

	return nil
}

// mergeWithDefaults merges provided variables with template defaults
func (te *TemplateEngine) mergeWithDefaults(config *TemplateConfig, variables map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// Start with defaults
	for _, variable := range config.Variables {
		if variable.DefaultValue != nil {
			merged[variable.Name] = variable.DefaultValue
		}
	}

	// Override with provided variables
	for key, value := range variables {
		merged[key] = value
	}

	return merged
}

// processTemplate processes a template with variables
func (te *TemplateEngine) processTemplate(templateData map[string]interface{}, variables map[string]interface{}) (map[string]interface{}, error) {
	// Convert template data to YAML string for processing
	yamlData, err := yaml.Marshal(templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal template data: %w", err)
	}

	// Process template string
	tmpl, err := template.New("config").Parse(string(yamlData))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// Parse processed YAML back to map
	var result map[string]interface{}
	if err := yaml.Unmarshal(buf.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse processed template: %w", err)
	}

	return result, nil
}

// GetTemplateInfo returns information about a template
func (te *TemplateEngine) GetTemplateInfo(name string) (*TemplateConfig, error) {
	return te.LoadTemplate(name)
}

// SaveTemplate saves a template configuration
func (te *TemplateEngine) SaveTemplate(name string, config *TemplateConfig) error {
	templatePath := te.getTemplatePath(name)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(templatePath), 0o755); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal template configuration: %w", err)
	}

	if err := os.WriteFile(templatePath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	return nil
}

// DeleteTemplate deletes a template
func (te *TemplateEngine) DeleteTemplate(name string) error {
	templatePath := te.getTemplatePath(name)
	return os.Remove(templatePath)
}

// getTemplatePath returns the full path to a template file
func (te *TemplateEngine) getTemplatePath(name string) string {
	if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
		name += ".yaml"
	}
	return filepath.Join(te.TemplateDir, name)
}

// CreateBuiltinTemplates creates built-in template files
func (te *TemplateEngine) CreateBuiltinTemplates() error {
	builtinTemplates := GetBuiltinTemplates()

	for name, config := range builtinTemplates {
		templatePath := te.getTemplatePath(name)

		// Skip if template already exists
		if _, err := os.Stat(templatePath); err == nil {
			continue
		}

		if err := te.SaveTemplate(name, config); err != nil {
			return fmt.Errorf("failed to create builtin template %s: %w", name, err)
		}
	}

	return nil
}

// ValidateTemplate validates a template configuration
func (te *TemplateEngine) ValidateTemplate(config *TemplateConfig) error {
	if config.Name == "" {
		return fmt.Errorf("template name is required")
	}

	if config.Template == nil {
		return fmt.Errorf("template configuration is required")
	}

	// Validate variable definitions
	for _, variable := range config.Variables {
		if variable.Name == "" {
			return fmt.Errorf("variable name is required")
		}

		if variable.Type == "" {
			variable.Type = "string" // Default type
		}

		// Validate type
		validTypes := []string{"string", "int", "bool", "array"}
		isValidType := false
		for _, validType := range validTypes {
			if variable.Type == validType {
				isValidType = true
				break
			}
		}

		if !isValidType {
			return fmt.Errorf("invalid variable type %s for variable %s", variable.Type, variable.Name)
		}
	}

	return nil
}

// InterpolateString interpolates template variables in a string
func (te *TemplateEngine) InterpolateString(s string, variables map[string]interface{}) (string, error) {
	tmpl, err := template.New("string").Parse(s)
	if err != nil {
		return "", fmt.Errorf("failed to parse string template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("failed to execute string template: %w", err)
	}

	return buf.String(), nil
}
