// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package devenv provides development environment management capabilities
package devenv

import (
	"context"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// Environment represents a complete development environment configuration
type Environment struct {
	Name         string                   `yaml:"name"`
	Description  string                   `yaml:"description"`
	Services     map[string]ServiceConfig `yaml:"services"`
	Dependencies []string                 `yaml:"dependencies"`
	PreHooks     []Hook                   `yaml:"pre_hooks,omitempty"`
	PostHooks    []Hook                   `yaml:"post_hooks,omitempty"`
}

// ServiceConfig contains configuration for a specific service
type ServiceConfig struct {
	AWS        *AWSConfig        `yaml:"aws,omitempty"`
	GCP        *GCPConfig        `yaml:"gcp,omitempty"`
	Azure      *AzureConfig      `yaml:"azure,omitempty"`
	Docker     *DockerConfig     `yaml:"docker,omitempty"`
	Kubernetes *KubernetesConfig `yaml:"kubernetes,omitempty"`
	SSH        *SSHConfig        `yaml:"ssh,omitempty"`
}

// AWSConfig represents AWS service configuration
type AWSConfig struct {
	Profile   string `yaml:"profile"`
	Region    string `yaml:"region"`
	AccountID string `yaml:"account_id,omitempty"`
}

// GCPConfig represents GCP service configuration
type GCPConfig struct {
	Project string `yaml:"project"`
	Account string `yaml:"account,omitempty"`
	Region  string `yaml:"region,omitempty"`
}

// AzureConfig represents Azure service configuration
type AzureConfig struct {
	Subscription string `yaml:"subscription"`
	Tenant       string `yaml:"tenant,omitempty"`
}

// DockerConfig represents Docker service configuration
type DockerConfig struct {
	Context string `yaml:"context"`
}

// KubernetesConfig represents Kubernetes service configuration
type KubernetesConfig struct {
	Context   string `yaml:"context"`
	Namespace string `yaml:"namespace,omitempty"`
}

// SSHConfig represents SSH service configuration
type SSHConfig struct {
	Config string `yaml:"config"`
}

// Hook represents a command to execute before or after environment switching
type Hook struct {
	Command string        `yaml:"command"`
	Timeout time.Duration `yaml:"timeout,omitempty"`
	OnError string        `yaml:"on_error,omitempty"` // continue, fail, rollback
}

// ServiceSwitcher interface for switching individual services
type ServiceSwitcher interface {
	Name() string
	Switch(ctx context.Context, config interface{}) error
	GetCurrentState(ctx context.Context) (interface{}, error)
	Rollback(ctx context.Context, previousState interface{}) error
}

// SwitchProgress represents the progress of environment switching
type SwitchProgress struct {
	TotalServices     int           `json:"total_services"`
	CompletedServices int           `json:"completed_services"`
	CurrentService    string        `json:"current_service"`
	Status            string        `json:"status"`
	StartTime         time.Time     `json:"start_time"`
	EstimatedEnd      time.Time     `json:"estimated_end"`
	Errors            []SwitchError `json:"errors,omitempty"`
}

// SwitchError represents an error during environment switching
type SwitchError struct {
	Service string    `json:"service"`
	Error   string    `json:"error"`
	Time    time.Time `json:"time"`
}

// SwitchResult represents the result of environment switching
type SwitchResult struct {
	Success           bool          `json:"success"`
	SwitchedServices  []string      `json:"switched_services"`
	FailedServices    []string      `json:"failed_services"`
	RollbackPerformed bool          `json:"rollback_performed"`
	Duration          time.Duration `json:"duration"`
	Errors            []SwitchError `json:"errors,omitempty"`
}

// LoadEnvironment loads an environment configuration from YAML
func LoadEnvironment(data []byte) (*Environment, error) {
	var env Environment
	if err := yaml.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("failed to parse environment configuration: %w", err)
	}

	// Validate required fields
	if env.Name == "" {
		return nil, fmt.Errorf("environment name is required")
	}

	return &env, nil
}

// LoadEnvironmentFromFile loads an environment configuration from a file
func LoadEnvironmentFromFile(filepath string) (*Environment, error) {
	// This will be implemented when we add file reading capabilities
	return nil, fmt.Errorf("file loading not yet implemented")
}

// Validate validates the environment configuration
func (e *Environment) Validate() error {
	if e.Name == "" {
		return fmt.Errorf("environment name is required")
	}

	if len(e.Services) == 0 {
		return fmt.Errorf("at least one service must be configured")
	}

	// Validate dependencies
	for _, dep := range e.Dependencies {
		if _, exists := e.Services[dep]; !exists {
			return fmt.Errorf("dependency service '%s' is not configured", dep)
		}
	}

	return nil
}

// GetServiceNames returns a list of configured service names
func (e *Environment) GetServiceNames() []string {
	services := make([]string, 0, len(e.Services))
	for name := range e.Services {
		services = append(services, name)
	}
	return services
}

// HasService checks if a service is configured in this environment
func (e *Environment) HasService(serviceName string) bool {
	_, exists := e.Services[serviceName]
	return exists
}
