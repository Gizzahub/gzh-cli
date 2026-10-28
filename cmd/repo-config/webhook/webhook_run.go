// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package webhook

import (
	"context"
	"fmt"

	"github.com/google/go-github/v66/github"
	"golang.org/x/oauth2"
)

// runWebhookList lists all webhooks for a repository.
func runWebhookList(ctx context.Context, flags *WebhookFlags) error {
	client := createGitHubClient(flags.Token)

	webhooks, _, err := client.Repositories.ListHooks(ctx, flags.Organization, flags.Repository, nil)
	if err != nil {
		return fmt.Errorf("failed to list webhooks: %w", err)
	}

	return displayWebhooks(webhooks, flags.OutputFormat)
}

// runWebhookCreate creates a new webhook.
func runWebhookCreate(ctx context.Context, flags *WebhookFlags) error {
	client := createGitHubClient(flags.Token)

	config := &github.HookConfig{
		URL:         &flags.URL,
		ContentType: &flags.ContentType,
	}
	if flags.Secret != "" {
		config.Secret = &flags.Secret
	}

	hook := &github.Hook{
		Events: flags.Events,
		Active: &flags.Active,
		Config: config,
	}

	if flags.DryRun {
		fmt.Printf("Would create webhook with URL: %s, Events: %v\n", flags.URL, flags.Events)
		return nil
	}

	createdHook, _, err := client.Repositories.CreateHook(ctx, flags.Organization, flags.Repository, hook)
	if err != nil {
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	fmt.Printf("Successfully created webhook with ID: %d\n", createdHook.GetID())

	return displayWebhook(createdHook, "table")
}

// runWebhookUpdate updates an existing webhook.
func runWebhookUpdate(ctx context.Context, flags *WebhookFlags) error {
	client := createGitHubClient(flags.Token)

	// Get existing webhook to preserve unmodified fields
	existingHook, _, err := client.Repositories.GetHook(ctx, flags.Organization, flags.Repository, flags.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing webhook: %w", err)
	}

	// Create updated hook with existing values as defaults
	config := existingHook.Config
	if config == nil {
		config = &github.HookConfig{}
	}

	// Update fields if provided
	if flags.URL != "" {
		config.URL = &flags.URL
	}

	if flags.ContentType != "" {
		config.ContentType = &flags.ContentType
	}

	if flags.Secret != "" {
		config.Secret = &flags.Secret
	}

	hook := &github.Hook{
		Config: config,
		Active: &flags.Active,
	}

	// Update events if provided
	if len(flags.Events) > 0 {
		hook.Events = flags.Events
	} else {
		hook.Events = existingHook.Events
	}

	if flags.DryRun {
		fmt.Printf("Would update webhook ID %d\n", flags.ID)
		return nil
	}

	updatedHook, _, err := client.Repositories.EditHook(ctx, flags.Organization, flags.Repository, flags.ID, hook)
	if err != nil {
		return fmt.Errorf("failed to update webhook: %w", err)
	}

	fmt.Printf("Successfully updated webhook with ID: %d\n", flags.ID)

	return displayWebhook(updatedHook, "table")
}

// runWebhookDelete deletes a webhook.
func runWebhookDelete(ctx context.Context, flags *WebhookFlags) error {
	client := createGitHubClient(flags.Token)

	if flags.DryRun {
		fmt.Printf("Would delete webhook ID %d\n", flags.ID)
		return nil
	}

	_, err := client.Repositories.DeleteHook(ctx, flags.Organization, flags.Repository, flags.ID)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	fmt.Printf("Successfully deleted webhook with ID: %d\n", flags.ID)

	return nil
}

// runWebhookGet gets details of a specific webhook.
func runWebhookGet(ctx context.Context, flags *WebhookFlags) error {
	client := createGitHubClient(flags.Token)

	hook, _, err := client.Repositories.GetHook(ctx, flags.Organization, flags.Repository, flags.ID)
	if err != nil {
		return fmt.Errorf("failed to get webhook: %w", err)
	}

	return displayWebhook(hook, flags.OutputFormat)
}

// createGitHubClient creates a GitHub API client.
func createGitHubClient(token string) *github.Client {
	if token == "" {
		return github.NewClient(nil)
	}

	// 여기의 context.Background()는 그대로 둔다. oauth2.NewClient는 이 맥락을
	// 취소 통로로 쓰지 않고 밑에 깔 http.Client를 찾는 데만 쓴다(내부의
	// ContextClient). StaticTokenSource라 토큰을 다시 받아 올 일도 없다.
	// 실행 맥락은 이 client를 쓰는 각 호출(ListHooks 등)에 직접 넘긴다.
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)

	return github.NewClient(tc)
}
