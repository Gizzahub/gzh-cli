// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/gizzahub/gzh-manager-go/internal/env"
)

// ProviderCloner defines the interface for provider-specific cloning operations.
type ProviderCloner interface {
	CloneOrganization(orgName, targetPath, strategy string) error
	CloneGroup(groupName, targetPath, strategy string) error
	SetToken(token string)
	GetName() string
}

// GitHubCloner implements ProviderCloner for GitHub.
type GitHubCloner struct {
	token       string
	environment env.Environment
}

// NewGitHubCloner creates a new GitHub cloner.
func NewGitHubCloner(token string) *GitHubCloner {
	return NewGitHubClonerWithEnv(token, env.NewOSEnvironment())
}

// NewGitHubClonerWithEnv creates a new GitHub cloner with the provided environment.
func NewGitHubClonerWithEnv(token string, environment env.Environment) *GitHubCloner {
	return &GitHubCloner{
		token:       token,
		environment: environment,
	}
}

func (g *GitHubCloner) CloneOrganization(orgName, targetPath, strategy string) error {
	// Set token as environment variable if provided
	if g.token != "" && !strings.HasPrefix(g.token, "$") {
		_ = g.environment.Set(env.CommonEnvironmentKeys.GitHubToken, g.token) // Ignore error
	}
	// Use the new provider service interface
	factory := NewDefaultProviderFactory(g.environment)

	provider, err := factory.CreateProvider(context.Background(), ProviderGitHub, ProviderConfig{Token: g.token})
	if err != nil {
		return fmt.Errorf("failed to create GitHub provider: %w", err)
	}

	return provider.RefreshAll(context.Background(), targetPath, orgName, strategy)
}

func (g *GitHubCloner) CloneGroup(groupName, targetPath, strategy string) error {
	// GitHub doesn't have groups, use organization instead
	return g.CloneOrganization(groupName, targetPath, strategy)
}

func (g *GitHubCloner) SetToken(token string) {
	g.token = token
}

func (g *GitHubCloner) GetName() string {
	return ProviderGitHub
}

// GitLabCloner implements ProviderCloner for GitLab.
type GitLabCloner struct {
	token       string
	environment env.Environment
}

// NewGitLabCloner creates a new GitLab cloner.
func NewGitLabCloner(token string) *GitLabCloner {
	return NewGitLabClonerWithEnv(token, env.NewOSEnvironment())
}

// NewGitLabClonerWithEnv creates a new GitLab cloner with the provided environment.
func NewGitLabClonerWithEnv(token string, environment env.Environment) *GitLabCloner {
	return &GitLabCloner{
		token:       token,
		environment: environment,
	}
}

func (g *GitLabCloner) CloneOrganization(orgName, targetPath, strategy string) error {
	// GitLab organizations are groups
	return g.CloneGroup(orgName, targetPath, strategy)
}

func (g *GitLabCloner) CloneGroup(groupName, targetPath, strategy string) error {
	// Set token as environment variable if provided
	if g.token != "" && !strings.HasPrefix(g.token, "$") {
		_ = g.environment.Set(env.CommonEnvironmentKeys.GitLabToken, g.token) // Ignore error
	}
	// Use the new provider service interface
	factory := NewDefaultProviderFactory(g.environment)

	provider, err := factory.CreateProvider(context.Background(), ProviderGitLab, ProviderConfig{Token: g.token})
	if err != nil {
		return fmt.Errorf("failed to create GitLab provider: %w", err)
	}

	return provider.RefreshAll(context.Background(), targetPath, groupName, strategy)
}

func (g *GitLabCloner) SetToken(token string) {
	g.token = token
}

func (g *GitLabCloner) GetName() string {
	return ProviderGitLab
}

// GiteaCloner implements ProviderCloner for Gitea.
type GiteaCloner struct {
	token       string
	environment env.Environment
}

// NewGiteaCloner creates a new Gitea cloner.
func NewGiteaCloner(token string) *GiteaCloner {
	return NewGiteaClonerWithEnv(token, env.NewOSEnvironment())
}

// NewGiteaClonerWithEnv creates a new Gitea cloner with the provided environment.
func NewGiteaClonerWithEnv(token string, environment env.Environment) *GiteaCloner {
	return &GiteaCloner{
		token:       token,
		environment: environment,
	}
}

func (g *GiteaCloner) CloneOrganization(orgName, targetPath, strategy string) error {
	// Set token as environment variable if provided
	if g.token != "" && !strings.HasPrefix(g.token, "$") {
		_ = g.environment.Set(env.CommonEnvironmentKeys.GiteaToken, g.token) // Ignore error
	}
	// Use the new provider service interface
	factory := NewDefaultProviderFactory(g.environment)

	provider, err := factory.CreateProvider(context.Background(), ProviderGitea, ProviderConfig{Token: g.token})
	if err != nil {
		return fmt.Errorf("failed to create Gitea provider: %w", err)
	}

	return provider.RefreshAll(context.Background(), targetPath, orgName, strategy)
}

func (g *GiteaCloner) CloneGroup(groupName, targetPath, strategy string) error {
	// Gitea doesn't have groups, use organization instead
	return g.CloneOrganization(groupName, targetPath, strategy)
}

func (g *GiteaCloner) SetToken(token string) {
	g.token = token
}

func (g *GiteaCloner) GetName() string {
	return ProviderGitea
}

// CreateProviderCloner creates a cloner for the specified provider (deprecated: use ProviderFactory).
func CreateProviderCloner(providerName, token string) (ProviderCloner, error) {
	environment := env.NewOSEnvironment()

	switch strings.ToLower(providerName) {
	case ProviderGitHub:
		return NewGitHubClonerWithEnv(token, environment), nil
	case ProviderGitLab:
		return NewGitLabClonerWithEnv(token, environment), nil
	case ProviderGitea:
		return NewGiteaClonerWithEnv(token, environment), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}
}

// BulkCloneExecutor handles bulk cloning operations with filtering and processing.
type BulkCloneExecutor struct {
	integration *BulkCloneIntegration
	cloners     map[string]ProviderCloner
}

// NewBulkCloneExecutor creates a new bulk clone executor.
func NewBulkCloneExecutor(config *Config) (*BulkCloneExecutor, error) {
	return NewBulkCloneExecutorWithFactory(config, nil)
}

// NewBulkCloneExecutorWithFactory creates a new bulk clone executor with a custom factory.
func NewBulkCloneExecutorWithFactory(config *Config, factory ProviderFactory) (*BulkCloneExecutor, error) {
	integration := NewBulkCloneIntegration(config)
	cloners := make(map[string]ProviderCloner)

	// Create cloners for each configured provider using the legacy interface
	for providerName, provider := range config.Providers {
		cloner, err := CreateProviderCloner(providerName, provider.Token)
		if err != nil {
			return nil, fmt.Errorf("failed to create cloner for %s: %w", providerName, err)
		}

		cloners[providerName] = cloner
	}

	return &BulkCloneExecutor{
		integration: integration,
		cloners:     cloners,
	}, nil
}

// ExecuteAll executes bulk cloning for all configured targets.
func (e *BulkCloneExecutor) ExecuteAll(filters map[string]interface{}) (*BulkCloneResult, error) {
	targets, err := e.integration.GetAllTargets()
	if err != nil {
		return nil, fmt.Errorf("failed to get targets: %w", err)
	}

	return e.executeTargets(targets, filters)
}

// ExecuteByProvider executes bulk cloning for a specific provider.
func (e *BulkCloneExecutor) ExecuteByProvider(providerName string, filters map[string]interface{}) (*BulkCloneResult, error) {
	targets, err := e.integration.GetTargetsByProvider(providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get targets for provider %s: %w", providerName, err)
	}

	return e.executeTargets(targets, filters)
}

// executeTargets executes cloning for a list of targets.
func (e *BulkCloneExecutor) executeTargets(targets []BulkCloneTarget, filters map[string]interface{}) (*BulkCloneResult, error) {
	result := &BulkCloneResult{
		TotalTargets: len(targets),
		Results:      make([]TargetResult, 0, len(targets)),
	}

	for _, target := range targets {
		// Apply filters
		if !e.integration.ShouldProcessTarget(target, filters) {
			result.SkippedTargets++
			continue
		}

		targetResult := e.executeTarget(target)
		result.Results = append(result.Results, targetResult)

		if targetResult.Success {
			result.SuccessfulTargets++
		} else {
			result.FailedTargets++
		}
	}

	return result, nil
}

// executeTarget executes cloning for a single target.
func (e *BulkCloneExecutor) executeTarget(target BulkCloneTarget) TargetResult {
	result := TargetResult{
		Provider: target.Provider,
		Name:     target.Name,
		CloneDir: target.CloneDir,
		Strategy: target.Strategy,
	}

	// Get cloner for this provider
	cloner, exists := e.cloners[target.Provider]
	if !exists {
		result.Error = fmt.Sprintf("no cloner available for provider %s", target.Provider)
		return result
	}

	// Expand target directory
	targetPath := ExpandEnvironmentVariables(target.CloneDir)
	result.CloneDir = targetPath

	// Execute cloning based on target type
	var err error
	if target.Provider == ProviderGitLab {
		err = cloner.CloneGroup(target.Name, targetPath, target.Strategy)
	} else {
		err = cloner.CloneOrganization(target.Name, targetPath, target.Strategy)
	}

	if err != nil {
		result.Error = err.Error()
	} else {
		result.Success = true
	}

	return result
}

// BulkCloneResult contains the results of a bulk clone operation.
type BulkCloneResult struct {
	TotalTargets      int            `json:"totalTargets"`
	SuccessfulTargets int            `json:"successfulTargets"`
	FailedTargets     int            `json:"failedTargets"`
	SkippedTargets    int            `json:"skippedTargets"`
	Results           []TargetResult `json:"results"`
}

// TargetResult contains the result of cloning a single target.
type TargetResult struct {
	Provider string `json:"provider"`
	Name     string `json:"name"`
	CloneDir string `json:"cloneDir"`
	Strategy string `json:"strategy"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
}

// GetSummary returns a summary of the bulk clone operation.
func (r *BulkCloneResult) GetSummary() string {
	return fmt.Sprintf("Total: %d, Successful: %d, Failed: %d, Skipped: %d",
		r.TotalTargets, r.SuccessfulTargets, r.FailedTargets, r.SkippedTargets)
}
