package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Gizzahub/gzh-cli/internal/testlib"
)

// setupBasicRepos creates basic test repositories
func setupBasicRepos(ctx context.Context, baseDir string, count int, withData bool, branches []string) error {
	factory := testlib.NewMockRepoFactory()

	log.Printf("Creating %d basic repositories in %s", count, baseDir)

	for i := 1; i <= count; i++ {
		repoName := fmt.Sprintf("basic-repo-%02d", i)
		opts := testlib.BasicRepoOptions{
			BaseDir:     baseDir,
			RepoName:    repoName,
			InitialData: withData,
			Branches:    branches,
		}

		if err := factory.CreateBasicRepos(ctx, opts); err != nil {
			return fmt.Errorf("failed to create basic repository %s: %w", repoName, err)
		}

		log.Printf("✓ Created basic repository: %s", repoName)
	}

	return nil
}

// setupConflictRepos creates conflict scenario repositories
func setupConflictRepos(ctx context.Context, baseDir string, conflictTypes []string) error {
	factory := testlib.NewMockRepoFactory()

	log.Printf("Creating conflict repositories in %s", baseDir)

	for i, conflictType := range conflictTypes {
		repoName := fmt.Sprintf("conflict-%s-%02d", conflictType, i+1)
		opts := testlib.ConflictRepoOptions{
			BaseDir:      baseDir,
			RepoName:     repoName,
			ConflictType: conflictType,
			LocalChanges: true,
		}

		if err := factory.CreateConflictRepos(ctx, opts); err != nil {
			return fmt.Errorf("failed to create conflict repository %s: %w", repoName, err)
		}

		log.Printf("✓ Created conflict repository: %s (%s)", repoName, conflictType)
	}

	return nil
}

// setupSpecialRepos creates special scenario repositories (placeholder for Phase 1C)
func setupSpecialRepos(ctx context.Context, baseDir string, specialTypes []string) error {
	log.Println("Special repositories will be implemented in Phase 1C")
	return nil
}
