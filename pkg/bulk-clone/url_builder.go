// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package bulkclone

import (
	"fmt"
)

// GitURLBuilder provides protocol-aware Git URL construction.
type GitURLBuilder struct {
	Protocol string
	Hostname string
}

// NewGitURLBuilder creates a new GitURLBuilder with the specified protocol and hostname.
func NewGitURLBuilder(protocol, hostname string) *GitURLBuilder {
	return &GitURLBuilder{
		Protocol: protocol,
		Hostname: hostname,
	}
}

// BuildURL constructs a Git clone URL based on the protocol.
func (b *GitURLBuilder) BuildURL(orgName, repoName string) string {
	switch b.Protocol {
	case "ssh":
		return fmt.Sprintf("git@%s:%s/%s.git", b.Hostname, orgName, repoName)
	case "https":
		return fmt.Sprintf("https://%s/%s/%s.git", b.Hostname, orgName, repoName)
	case "http":
		return fmt.Sprintf("http://%s/%s/%s.git", b.Hostname, orgName, repoName)
	default:
		// Default to HTTPS for safety
		return fmt.Sprintf("https://%s/%s/%s.git", b.Hostname, orgName, repoName)
	}
}

// when custom SSH configurations are in use.
func (b *GitURLBuilder) BuildSSHHostAlias(orgName string) string {
	if b.Protocol != "ssh" {
		return b.Hostname
	}

	// For GitHub, GitLab, etc., we might want to use host aliases
	switch b.Hostname {
	case "github.com":
		return fmt.Sprintf("github-%s", orgName)
	case "gitlab.com":
		return fmt.Sprintf("gitlab-%s", orgName)
	case "gitea.com":
		return fmt.Sprintf("gitea-%s", orgName)
	default:
		// For custom hostnames, use as-is
		return b.Hostname
	}
}

// BuildURLWithHostAlias constructs a Git clone URL using SSH host aliases when appropriate.
func (b *GitURLBuilder) BuildURLWithHostAlias(orgName, repoName string) string {
	if b.Protocol == "ssh" {
		hostAlias := b.BuildSSHHostAlias(orgName)
		return fmt.Sprintf("git@%s:%s/%s.git", hostAlias, orgName, repoName)
	}

	return b.BuildURL(orgName, repoName)
}

// GetDefaultHostname returns the default hostname for a given provider.
func GetDefaultHostname(provider string) string {
	switch provider {
	case "github":
		return "github.com"
	case "gitlab":
		return "gitlab.com"
	case "gitea":
		return "gitea.com"
	default:
		return provider // Use provider name as hostname fallback
	}
}

// BuildURLForProvider is a convenience function that builds URLs for common Git providers.
func BuildURLForProvider(provider, protocol, orgName, repoName string) string {
	hostname := GetDefaultHostname(provider)
	builder := NewGitURLBuilder(protocol, hostname)

	return builder.BuildURL(orgName, repoName)
}

// BuildURLWithHostAliasForProvider is a convenience function that builds URLs with host aliases.
func BuildURLWithHostAliasForProvider(provider, protocol, orgName, repoName string) string {
	hostname := GetDefaultHostname(provider)
	builder := NewGitURLBuilder(protocol, hostname)

	return builder.BuildURLWithHostAlias(orgName, repoName)
}
