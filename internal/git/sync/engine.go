// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package sync

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/Gizzahub/gzh-cli/pkg/git/provider"
)

// SyncEngine manages repository synchronization between Git platforms.
type SyncEngine struct {
	source      provider.GitProvider
	destination provider.GitProvider
	options     SyncOptions
}

// NewSyncEngine creates a new sync engine.
func NewSyncEngine(src, dst provider.GitProvider, opts SyncOptions) *SyncEngine {
	return &SyncEngine{
		source:      src,
		destination: dst,
		options:     opts,
	}
}

// Sync executes the synchronization process.
func (e *SyncEngine) Sync(ctx context.Context) error {
	// 1. Analyze source repositories
	sourceRepos, err := e.analyzeSource(ctx)
	if err != nil {
		return fmt.Errorf("failed to analyze source: %w", err)
	}

	if len(sourceRepos) == 0 {
		fmt.Println("No repositories found in source")
		return nil
	}

	// 2. Analyze destination repositories
	destRepos, err := e.analyzeDestination(ctx)
	if err != nil {
		return fmt.Errorf("failed to analyze destination: %w", err)
	}

	// 3. Create synchronization plan
	plan := e.createSyncPlan(sourceRepos, destRepos)

	// 4. Handle dry run
	if e.options.DryRun {
		return e.printSyncPlan(plan)
	}

	// 5. Execute synchronization plan
	return e.executeSyncPlan(ctx, plan)
}

// analyzeSource analyzes the source to get repositories to sync.
func (e *SyncEngine) analyzeSource(ctx context.Context) ([]provider.Repository, error) {
	sourceTarget, err := e.options.GetSourceTarget()
	if err != nil {
		return nil, err
	}

	if sourceTarget.IsRepository() {
		// Get single repository
		repo, err := e.source.GetRepository(ctx, sourceTarget.FullName())
		if err != nil {
			return nil, fmt.Errorf("failed to get repository %s: %w", sourceTarget.FullName(), err)
		}
		return []provider.Repository{*repo}, nil
	}

	// Get all repositories from organization
	listOpts := provider.ListOptions{
		Organization: sourceTarget.Org,
		Type:         "all",
		Sort:         "full_name",
		Direction:    "asc",
	}

	result, err := e.source.ListRepositories(ctx, listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	// Apply filtering
	return e.filterRepositories(result.Repositories), nil
}

// analyzeDestination analyzes the destination to get existing repositories.
func (e *SyncEngine) analyzeDestination(ctx context.Context) (map[string]provider.Repository, error) {
	destTarget, err := e.options.GetDestinationTarget()
	if err != nil {
		return nil, err
	}

	destRepos := make(map[string]provider.Repository)

	if destTarget.IsRepository() {
		// Check if single repository exists
		repo, err := e.destination.GetRepository(ctx, destTarget.FullName())
		if err == nil {
			destRepos[repo.Name] = *repo
		}
		// If error, repository doesn't exist, which is fine
		return destRepos, nil
	}

	// Get all repositories from destination organization
	listOpts := provider.ListOptions{
		Organization: destTarget.Org,
		Type:         "all",
		Sort:         "full_name",
		Direction:    "asc",
	}

	result, err := e.destination.ListRepositories(ctx, listOpts)
	if err != nil {
		// If organization doesn't exist, that's fine
		return destRepos, nil
	}

	// Index by repository name
	for _, repo := range result.Repositories {
		destRepos[repo.Name] = repo
	}

	return destRepos, nil
}

// filterRepositories applies filtering options to repositories.
func (e *SyncEngine) filterRepositories(repos []provider.Repository) []provider.Repository {
	var filtered []provider.Repository

	for _, repo := range repos {
		// Apply match pattern
		if e.options.Match != "" {
			matched, err := filepath.Match(e.options.Match, repo.Name)
			if err != nil || !matched {
				continue
			}
		}

		// Apply exclude pattern
		if e.options.Exclude != "" {
			matched, err := filepath.Match(e.options.Exclude, repo.Name)
			if err == nil && matched {
				continue
			}
		}

		filtered = append(filtered, repo)
	}

	return filtered
}

// createSyncPlan creates a synchronization plan based on source and destination analysis.
func (e *SyncEngine) createSyncPlan(sourceRepos []provider.Repository, destRepos map[string]provider.Repository) SyncPlan {
	var plan SyncPlan

	for _, sourceRepo := range sourceRepos {
		if destRepo, exists := destRepos[sourceRepo.Name]; exists {
			// Repository exists - plan update
			if e.options.UpdateExisting {
				repoSync := RepoSync{
					Source:      sourceRepo,
					Destination: &destRepo,
					Actions:     e.createSyncActions(sourceRepo, &destRepo),
				}
				plan.Update = append(plan.Update, repoSync)
			} else {
				repoSync := RepoSync{
					Source:      sourceRepo,
					Destination: &destRepo,
					Actions:     []SyncAction{{Type: "skip", Description: "Update disabled"}},
				}
				plan.Skip = append(plan.Skip, repoSync)
			}
		} else {
			// Repository doesn't exist - plan creation
			if e.options.CreateMissing {
				repoSync := RepoSync{
					Source:      sourceRepo,
					Destination: nil,
					Actions:     e.createSyncActions(sourceRepo, nil),
				}
				plan.Create = append(plan.Create, repoSync)
			} else {
				repoSync := RepoSync{
					Source:      sourceRepo,
					Destination: nil,
					Actions:     []SyncAction{{Type: "skip", Description: "Creation disabled"}},
				}
				plan.Skip = append(plan.Skip, repoSync)
			}
		}
	}

	return plan
}

// createSyncActions creates the list of sync actions for a repository.
func (e *SyncEngine) createSyncActions(source provider.Repository, destination *provider.Repository) []SyncAction {
	var actions []SyncAction

	// Code synchronization
	if e.options.IncludeCode {
		actions = append(actions, SyncAction{
			Type:        "code",
			Description: "Sync repository code and branches",
			Handler: func(ctx context.Context) error {
				syncer := &CodeSyncer{
					source:      source,
					destination: destination,
					options:     e.options,
				}
				return syncer.Sync(ctx)
			},
		})
	}

	// Issues synchronization
	if e.options.IncludeIssues {
		actions = append(actions, SyncAction{
			Type:        "issues",
			Description: "Sync issues and comments",
			Handler: func(ctx context.Context) error {
				if destination == nil {
					return fmt.Errorf("cannot sync issues: destination repository not created yet")
				}
				syncer := &IssueSyncer{
					source:      e.source,
					destination: e.destination,
					mapping:     make(map[string]string),
				}
				return syncer.Sync(ctx, source, *destination)
			},
		})
	}

	// Wiki synchronization
	if e.options.IncludeWiki {
		actions = append(actions, SyncAction{
			Type:        "wiki",
			Description: "Sync wiki content",
			Handler: func(ctx context.Context) error {
				if destination == nil {
					return fmt.Errorf("cannot sync wiki: destination repository not created yet")
				}
				syncer := &WikiSyncer{
					source:      e.source,
					destination: e.destination,
				}
				// 사전 접근성 검증이 포함된 동기화 사용
				return syncer.SyncWithValidation(ctx, source, *destination, e.options.Verbose)
			},
		})
	}

	// Releases synchronization
	if e.options.IncludeReleases {
		actions = append(actions, SyncAction{
			Type:        "releases",
			Description: "Sync releases and tags",
			Handler: func(ctx context.Context) error {
				if destination == nil {
					return fmt.Errorf("cannot sync releases: destination repository not created yet")
				}
				syncer := NewReleaseSyncer(e.source, e.destination, e.options)
				return syncer.Sync(ctx, source, *destination)
			},
		})
	}

	// Settings synchronization
	if e.options.IncludeSettings {
		actions = append(actions, SyncAction{
			Type:        "settings",
			Description: "Sync repository settings",
			Handler: func(ctx context.Context) error {
				if destination == nil {
					return fmt.Errorf("cannot sync settings: destination repository not created yet")
				}
				syncer := NewSettingsSyncer(e.source, e.destination, e.options)
				return syncer.Sync(ctx, source, *destination)
			},
		})
	}

	return actions
}

// printSyncPlan prints the synchronization plan for dry run.
func (e *SyncEngine) printSyncPlan(plan SyncPlan) error {
	fmt.Printf("\n🔍 Synchronization Plan (Dry Run)\n")
	fmt.Printf("================================\n\n")

	if len(plan.Create) > 0 {
		fmt.Printf("📝 Repositories to CREATE (%d):\n", len(plan.Create))
		for _, sync := range plan.Create {
			fmt.Printf("  + %s\n", sync.Source.FullName)
			for _, action := range sync.Actions {
				fmt.Printf("    - %s: %s\n", action.Type, action.Description)
			}
		}
		fmt.Println()
	}

	if len(plan.Update) > 0 {
		fmt.Printf("🔄 Repositories to UPDATE (%d):\n", len(plan.Update))
		for _, sync := range plan.Update {
			fmt.Printf("  ~ %s\n", sync.Source.FullName)
			for _, action := range sync.Actions {
				fmt.Printf("    - %s: %s\n", action.Type, action.Description)
			}
		}
		fmt.Println()
	}

	if len(plan.Skip) > 0 {
		fmt.Printf("⏭️  Repositories to SKIP (%d):\n", len(plan.Skip))
		for _, sync := range plan.Skip {
			reason := "Unknown reason"
			if len(sync.Actions) > 0 {
				reason = sync.Actions[0].Description
			}
			fmt.Printf("  - %s (%s)\n", sync.Source.FullName, reason)
		}
		fmt.Println()
	}

	total := len(plan.Create) + len(plan.Update) + len(plan.Skip)
	fmt.Printf("Total repositories: %d\n", total)
	fmt.Printf("\nRun without --dry-run to execute the synchronization.\n")

	return nil
}

// SyncPlan represents a synchronization plan.
type SyncPlan struct {
	Create []RepoSync // Repositories to create
	Update []RepoSync // Repositories to update
	Skip   []RepoSync // Repositories to skip
}

// RepoSync represents a repository synchronization task.
type RepoSync struct {
	Source      provider.Repository
	Destination *provider.Repository // nil if creating new
	Actions     []SyncAction
}

// SyncAction represents a single synchronization action.
type SyncAction struct {
	Type        string // code, issues, wiki, etc.
	Description string
	Handler     func(context.Context) error
}
