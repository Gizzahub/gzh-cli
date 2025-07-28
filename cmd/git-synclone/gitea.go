// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-manager-go/pkg/gitea"
)

type giteaOptions struct {
	// Organization settings
	orgName string

	// Gitea instance
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

func newGiteaCmd(ctx context.Context) *cobra.Command {
	opts := &giteaOptions{}

	cmd := &cobra.Command{
		Use:   "gitea",
		Short: "Clone repositories from Gitea organizations",
		Long: `Clone all or filtered repositories from a Gitea organization.

Examples:
  # Clone all repositories from an organization
  git synclone gitea -o myorg -t ~/repos --api-url https://gitea.company.com

  # Clone only public repositories
  git synclone gitea -o myorg --visibility public

  # Clone repositories matching a pattern
  git synclone gitea -o myorg --match "service-*"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Inherit common flags from parent
			opts.target, _ = cmd.Flags().GetString("target")
			opts.parallel, _ = cmd.Flags().GetInt("parallel")
			opts.strategy, _ = cmd.Flags().GetString("strategy")
			opts.resume, _ = cmd.Flags().GetBool("resume")
			opts.cleanupOrphans, _ = cmd.Flags().GetBool("cleanup-orphans")
			opts.dryRun, _ = cmd.Flags().GetBool("dry-run")
			opts.progressMode, _ = cmd.Flags().GetString("progress-mode")

			return runGitea(ctx, opts)
		},
	}

	// Gitea-specific flags
	cmd.Flags().StringVarP(&opts.orgName, "org", "o", "", "Gitea organization name (required)")
	cmd.Flags().StringVar(&opts.apiURL, "api-url", "", "Gitea API URL (required)")
	cmd.Flags().StringVar(&opts.match, "match", "", "Repository name pattern (regex)")
	cmd.Flags().StringVar(&opts.visibility, "visibility", "all", "Repository visibility: public, private, or all")
	cmd.Flags().BoolVar(&opts.archived, "archived", false, "Include archived repositories")

	cmd.MarkFlagRequired("org")
	cmd.MarkFlagRequired("api-url")

	return cmd
}

func runGitea(ctx context.Context, opts *giteaOptions) error {
	// Validate options
	if err := validateGiteaOptions(opts); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	// Get Gitea token from environment
	token := os.Getenv("GITEA_TOKEN")
	if token == "" && opts.visibility != "public" {
		fmt.Printf("⚠️ Warning: No Gitea token provided. API rate limits may apply.\n")
		fmt.Printf("   Set GITEA_TOKEN environment variable for better performance.\n")
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

	fmt.Printf("🚀 Starting Gitea repository synchronization\n")
	fmt.Printf("   Organization: %s\n", opts.orgName)
	fmt.Printf("   API URL: %s\n", opts.apiURL)
	fmt.Printf("   Target: %s\n", absTarget)
	fmt.Printf("   Strategy: %s\n", opts.strategy)

	if opts.dryRun {
		fmt.Printf("\n⚠️ DRY RUN MODE - No actual changes will be made\n")
		// TODO: Implement dry run logic
		return nil
	}

	// Use the existing RefreshAll function from the gitea package
	// TODO: Add support for more advanced options like filtering, etc.
	return gitea.RefreshAll(ctx, absTarget, opts.orgName)
}

func validateGiteaOptions(opts *giteaOptions) error {
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
