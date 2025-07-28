// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package provider provides a unified abstraction layer for Git hosting platforms.
//
// This package defines common interfaces and types that allow working with different
// Git hosting providers (GitHub, GitLab, Gitea, Gogs) through a consistent API.
//
// Key Components:
//
//   - GitProvider: Main interface implemented by all providers
//   - RepositoryManager: Interface for repository lifecycle operations
//   - WebhookManager: Interface for webhook management
//   - EventManager: Interface for event handling
//   - ProviderFactory: Factory for creating provider instances
//   - ProviderRegistry: Registry with caching and lifecycle management
//
// Usage Example:
//
//	// Create factory and register providers
//	factory := provider.NewProviderFactory()
//	factory.RegisterProvider("github", github.NewProvider)
//	factory.RegisterProvider("gitlab", gitlab.NewProvider)
//
//	// Configure providers
//	factory.RegisterConfig("github-main", &provider.ProviderConfig{
//		Type:    "github",
//		Name:    "github-main",
//		Token:   "ghp_...",
//		Enabled: true,
//	})
//
//	// Create registry
//	registry := provider.NewProviderRegistry(factory, provider.RegistryConfig{
//		EnableCaching:      true,
//		EnableHealthChecks: true,
//	})
//
//	// Use providers
//	ghProvider, err := registry.GetProvider("github-main")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	repos, err := ghProvider.ListRepositories(ctx, provider.ListOptions{
//		Organization: "myorg",
//	})
//
// The package provides comprehensive error handling, rate limiting, retry mechanisms,
// and health monitoring for reliable operation across different Git platforms.
package provider
