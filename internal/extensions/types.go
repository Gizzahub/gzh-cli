// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package extensions provides user-extensible alias and external command integration.
package extensions

// Config is the user-defined extension configuration.
type Config struct {
	Aliases  map[string]AliasConfig  `yaml:"aliases"`  // Command aliases
	External []ExternalCommandConfig `yaml:"external"` // External command integration
}

// AliasConfig is the command alias configuration.
type AliasConfig struct {
	Command     string   `yaml:"command,omitempty"` // Simple alias command
	Description string   `yaml:"description"`       // Description
	Steps       []string `yaml:"steps,omitempty"`   // Multi-step workflow
	Params      []Param  `yaml:"params,omitempty"`  // Parameters
}

// Param is a parameter for aliased commands.
type Param struct {
	Name        string `yaml:"name"`        // Parameter name
	Description string `yaml:"description"` // Parameter description
	Required    bool   `yaml:"required"`    // Required flag
}

// ExternalCommandConfig is the external command integration configuration.
type ExternalCommandConfig struct {
	Name        string   `yaml:"name"`                  // Command name
	Command     string   `yaml:"command"`               // Command path to execute
	Description string   `yaml:"description"`           // Description
	Passthrough bool     `yaml:"passthrough,omitempty"` // Pass arguments through
	Args        []string `yaml:"args,omitempty"`        // Default arguments
}
