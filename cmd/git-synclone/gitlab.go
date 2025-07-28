// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-manager-go/pkg/gitlab"
)

type gitlabOptions struct {
	// Group settings
	groupName string
	recursive bool

	// GitLab instance
	apiURL string

	// Repository filters
	match      string
	visibility string
	archived   bool

	// Common options (inherited from parent)
	target         string
	parallel       int
	strategy       string
	resume         bool
	cleanupOrphans bool
	dryRun         bool
	progressMode   string
}

func newGitLabCmd(ctx context.Context) *cobra.Command {
	opts := &gitlabOptions{}

	cmd := &cobra.Command{
		Use:   "gitlab",
		Short: "Clone repositories from GitLab groups",
		Long: `Clone all or filtered repositories from a GitLab group.

Examples:
  # Clone all repositories from a group
  git synclone gitlab -g mygroup -t ~/repos

  # Clone from a self-hosted GitLab instance
  git synclone gitlab -g mygroup --api-url https://gitlab.company.com

  # Clone including subgroups
  git synclone gitlab -g mygroup --recursive

  # Clone only repositories matching a pattern
  git synclone gitlab -g mygroup --match "backend-*"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Inherit common flags from parent
			opts.target, _ = cmd.Flags().GetString("target")
			opts.parallel, _ = cmd.Flags().GetInt("parallel")
			opts.strategy, _ = cmd.Flags().GetString("strategy")
			opts.resume, _ = cmd.Flags().GetBool("resume")
			opts.cleanupOrphans, _ = cmd.Flags().GetBool("cleanup-orphans")
			opts.dryRun, _ = cmd.Flags().GetBool("dry-run")
			opts.progressMode, _ = cmd.Flags().GetString("progress-mode")

			return runGitLab(ctx, opts)
		},
	}

	// GitLab-specific flags
	cmd.Flags().StringVarP(&opts.groupName, "group", "g", "", "GitLab group name (required)")
	cmd.Flags().BoolVar(&opts.recursive, "recursive", false, "Include subgroups")
	cmd.Flags().StringVar(&opts.apiURL, "api-url", "https://gitlab.com", "GitLab API URL")
	cmd.Flags().StringVar(&opts.match, "match", "", "Repository name pattern (regex)")
	cmd.Flags().StringVar(&opts.visibility, "visibility", "all", "Repository visibility: public, private, or all")
	cmd.Flags().BoolVar(&opts.archived, "archived", false, "Include archived repositories")

	cmd.MarkFlagRequired("group")

	return cmd
}

func runGitLab(ctx context.Context, opts *gitlabOptions) error {
	// Validate options
	if err := validateGitLabOptions(opts); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	// Get GitLab token from environment
	token := os.Getenv("GITLAB_TOKEN")
	if token == "" && opts.visibility != "public" {
		fmt.Printf("⚠️ Warning: No GitLab token provided. API rate limits may apply.\n")
		fmt.Printf("   Set GITLAB_TOKEN environment variable for better performance.\n")
	}

	// Set default target directory if not specified
	if opts.target == "" {
		opts.target = opts.groupName
	}

	// Create absolute path for target
	absTarget, err := filepath.Abs(opts.target)
	if err != nil {
		return fmt.Errorf("invalid target path: %w", err)
	}

	fmt.Printf("🚀 Starting GitLab repository synchronization\n")
	fmt.Printf("   Group: %s\n", opts.groupName)
	fmt.Printf("   API URL: %s\n", opts.apiURL)
	fmt.Printf("   Target: %s\n", absTarget)
	fmt.Printf("   Strategy: %s\n", opts.strategy)

	if opts.dryRun {
		fmt.Printf("\n⚠️ DRY RUN MODE - No actual changes will be made\n")
		// TODO: Implement dry run logic
		return nil
	}

	// Use the existing RefreshAll function from the gitlab package
	// TODO: Add support for more advanced options like filtering, recursive, etc.
	return gitlab.RefreshAll(ctx, absTarget, opts.groupName, opts.strategy)
}

func validateGitLabOptions(opts *gitlabOptions) error {
	// Validate visibility
	validVisibility := map[string]bool{
		"all":     true,
		"public":  true,
		"private": true,
	}
	if !validVisibility[opts.visibility] {
		return fmt.Errorf("invalid visibility '%s': must be one of: all, public, private", opts.visibility)
	}

	// Validate strategy
	validStrategy := map[string]bool{
		"reset": true,
		"pull":  true,
		"fetch": true,
	}
	if !validStrategy[opts.strategy] {
		return fmt.Errorf("invalid strategy '%s': must be one of: reset, pull, fetch", opts.strategy)
	}

	// Validate progress mode
	validProgress := map[string]bool{
		"bar":     true,
		"dots":    true,
		"spinner": true,
		"quiet":   true,
	}
	if !validProgress[opts.progressMode] {
		return fmt.Errorf("invalid progress mode '%s': must be one of: bar, dots, spinner, quiet", opts.progressMode)
	}

	return nil
}
