// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package builders

import (
	"github.com/Gizzahub/gzh-cli/internal/env"
)

// EnvironmentBuilder provides a fluent interface for building test environments.
type EnvironmentBuilder struct {
	vars map[string]string
}

// NewEnvironmentBuilder creates a new EnvironmentBuilder.
func NewEnvironmentBuilder() *EnvironmentBuilder {
	return &EnvironmentBuilder{
		vars: make(map[string]string),
	}
}

// WithVar adds an environment variable.
func (b *EnvironmentBuilder) WithVar(key, value string) *EnvironmentBuilder {
	b.vars[key] = value
	return b
}

// WithGitHubToken adds a GitHub token.
func (b *EnvironmentBuilder) WithGitHubToken(token string) *EnvironmentBuilder {
	b.vars["GITHUB_TOKEN"] = token
	return b
}

// WithGitLabToken adds a GitLab token.
func (b *EnvironmentBuilder) WithGitLabToken(token string) *EnvironmentBuilder {
	b.vars["GITLAB_TOKEN"] = token
	return b
}

// WithGiteaToken adds a Gitea token.
func (b *EnvironmentBuilder) WithGiteaToken(token string) *EnvironmentBuilder {
	b.vars["GITEA_TOKEN"] = token
	return b
}

// WithHome sets the HOME environment variable.
func (b *EnvironmentBuilder) WithHome(home string) *EnvironmentBuilder {
	b.vars["HOME"] = home
	return b
}

// WithConfigPath sets the GZH_CONFIG_PATH environment variable.
func (b *EnvironmentBuilder) WithConfigPath(path string) *EnvironmentBuilder {
	b.vars["GZH_CONFIG_PATH"] = path
	return b
}

// Build returns the constructed mock environment.
func (b *EnvironmentBuilder) Build() env.Environment {
	return env.NewMockEnvironment(b.vars)
}

// BuildMap returns the environment variables as a map.
func (b *EnvironmentBuilder) BuildMap() map[string]string {
	result := make(map[string]string)
	for k, v := range b.vars {
		result[k] = v
	}

	return result
}
