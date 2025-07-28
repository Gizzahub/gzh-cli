// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-manager-go/pkg/github"
)

type githubOptions struct {
	// Organization settings
	orgName string

	// Repository filters
	match      string
	visibility string
	archived   bool

	// Clone settings
	protocol string

	// Common options (inherited from parent)
	target         string
	parallel       int
	strategy       string
	resume         bool
	cleanupOrphans bool
	dryRun         bool
	progressMode   string
}

func newGitHubCmd(ctx context.Context) *cobra.Command {
	opts := &githubOptions{}

	cmd := &cobra.Command{
		Use:   "github",
		Short: "Clone repositories from GitHub organizations",
		Long: `Clone all or filtered repositories from a GitHub organization.

Examples:
  # Clone all repositories from an organization
  git synclone github -o myorg -t ~/repos

  # Clone only public repositories matching a pattern
  git synclone github -o myorg --match "api-*" --visibility public

  # Resume an interrupted clone operation
  git synclone github -o myorg --resume

  # Use SSH protocol instead of HTTPS
  git synclone github -o myorg --protocol ssh`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Inherit common flags from parent
			opts.target, _ = cmd.Flags().GetString("target")
			opts.parallel, _ = cmd.Flags().GetInt("parallel")
			opts.strategy, _ = cmd.Flags().GetString("strategy")
			opts.resume, _ = cmd.Flags().GetBool("resume")
			opts.cleanupOrphans, _ = cmd.Flags().GetBool("cleanup-orphans")
			opts.dryRun, _ = cmd.Flags().GetBool("dry-run")
			opts.progressMode, _ = cmd.Flags().GetString("progress-mode")

			return runGitHub(ctx, opts)
		},
	}

	// GitHub-specific flags
	cmd.Flags().StringVarP(&opts.orgName, "org", "o", "", "GitHub organization name (required)")
	cmd.Flags().StringVar(&opts.match, "match", "", "Repository name pattern (regex)")
	cmd.Flags().StringVar(&opts.visibility, "visibility", "all", "Repository visibility: public, private, or all")
	cmd.Flags().BoolVar(&opts.archived, "archived", false, "Include archived repositories")
	cmd.Flags().StringVar(&opts.protocol, "protocol", "https", "Clone protocol: https or ssh")

	cmd.MarkFlagRequired("org")

	return cmd
}

func runGitHub(ctx context.Context, opts *githubOptions) error {
	// Validate options
	if err := validateGitHubOptions(opts); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	// Get GitHub token from environment
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" && opts.visibility != "public" {
		fmt.Printf("⚠️ Warning: No GitHub token provided. API rate limits may apply.\n")
		fmt.Printf("   Set GITHUB_TOKEN environment variable for better performance.\n")
	}

	// Set default target directory if not specified
	if opts.target == "" {
		opts.target = opts.orgName
	}

	// Create absolute path for target
	absTarget, err := filepath.Abs(opts.target)
	if err != nil {
		return fmt.Errorf("invalid target path: %w", err)
	}

	// For now, use the existing github.RefreshAll function
	// This maintains compatibility with the existing implementation
	fmt.Printf("🚀 Starting GitHub repository synchronization\n")
	fmt.Printf("   Organization: %s\n", opts.orgName)
	fmt.Printf("   Target: %s\n", absTarget)
	fmt.Printf("   Strategy: %s\n", opts.strategy)

	if opts.dryRun {
		fmt.Printf("\n⚠️ DRY RUN MODE - No actual changes will be made\n")
		// TODO: Implement dry run logic
		return nil
	}

	// Use the existing RefreshAll function from the github package
	// TODO: Add support for more advanced options like filtering, resume, etc.
	if opts.parallel > 1 {
		// Use the resumable version if parallel is specified
		return github.RefreshAllResumable(ctx, absTarget, opts.orgName, opts.strategy, opts.parallel, 3, opts.resume, opts.progressMode)
	}

	return github.RefreshAll(ctx, absTarget, opts.orgName, opts.strategy)
}

func validateGitHubOptions(opts *githubOptions) error {
	// Validate visibility
	validVisibility := map[string]bool{
		"all":     true,
		"public":  true,
		"private": true,
	}
	if !validVisibility[opts.visibility] {
		return fmt.Errorf("invalid visibility '%s': must be one of: all, public, private", opts.visibility)
	}

	// Validate protocol
	validProtocol := map[string]bool{
		"https": true,
		"ssh":   true,
	}
	if !validProtocol[opts.protocol] {
		return fmt.Errorf("invalid protocol '%s': must be one of: https, ssh", opts.protocol)
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
