package main

import (
	"context"
	"log"

	"github.com/spf13/cobra"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		log.Fatal(err)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup-test-repos",
		Short: "Setup test repositories for synclone testing",
		Long: `Setup test repositories with various Git scenarios for comprehensive
synclone testing including basic repos, conflict scenarios, and special cases.`,
	}

	cmd.AddCommand(newBasicCmd())
	cmd.AddCommand(newConflictCmd())
	cmd.AddCommand(newSpecialCmd())
	cmd.AddCommand(newAllCmd())

	return cmd
}

func newBasicCmd() *cobra.Command {
	var (
		baseDir  string
		count    int
		withData bool
		branches []string
	)

	cmd := &cobra.Command{
		Use:   "basic",
		Short: "Create basic test repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setupBasicRepos(context.Background(), baseDir, count, withData, branches)
		},
	}

	cmd.Flags().StringVar(&baseDir, "base-dir", "./test-repos", "Base directory for test repositories")
	cmd.Flags().IntVar(&count, "count", 5, "Number of repositories to create")
	cmd.Flags().BoolVar(&withData, "with-data", true, "Create repositories with initial data")
	cmd.Flags().StringSliceVar(&branches, "branches", []string{"main", "develop"}, "Branches to create")

	return cmd
}

func newConflictCmd() *cobra.Command {
	var (
		baseDir string
		types   []string
	)

	cmd := &cobra.Command{
		Use:   "conflict",
		Short: "Create conflict scenario repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setupConflictRepos(context.Background(), baseDir, types)
		},
	}

	cmd.Flags().StringVar(&baseDir, "base-dir", "./test-repos", "Base directory for test repositories")
	cmd.Flags().StringSliceVar(&types, "types", []string{"merge", "rebase", "diverged"}, "Conflict types to create")

	return cmd
}

func newSpecialCmd() *cobra.Command {
	var (
		baseDir string
		types   []string
	)

	cmd := &cobra.Command{
		Use:   "special",
		Short: "Create special scenario repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setupSpecialRepos(context.Background(), baseDir, types)
		},
	}

	cmd.Flags().StringVar(&baseDir, "base-dir", "./test-repos", "Base directory for test repositories")
	cmd.Flags().StringSliceVar(&types, "types", []string{"lfs", "submodule", "large"}, "Special types to create")

	return cmd
}

func newAllCmd() *cobra.Command {
	var baseDir string

	cmd := &cobra.Command{
		Use:   "all",
		Short: "Create all test repository scenarios",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Create basic repos
			if err := setupBasicRepos(ctx, baseDir, 5, true, []string{"main", "develop"}); err != nil {
				return err
			}

			// Create conflict repos
			if err := setupConflictRepos(ctx, baseDir, []string{"merge", "rebase", "diverged"}); err != nil {
				return err
			}

			// Create special repos (placeholder for Phase 1C)
			log.Println("Special repositories will be implemented in Phase 1C")

			return nil
		},
	}

	cmd.Flags().StringVar(&baseDir, "base-dir", "./test-repos", "Base directory for test repositories")

	return cmd
}
