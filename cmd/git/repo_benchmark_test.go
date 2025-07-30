// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package git

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	mockprovider "github.com/gizzahub/gzh-manager-go/internal/git/provider/mock"
	"github.com/gizzahub/gzh-manager-go/pkg/git/provider"
)

// BenchmarkCloneParallel benchmarks parallel cloning operations.
func BenchmarkCloneParallel(b *testing.B) {
	benchmarks := []struct {
		name      string
		repoCount int
		workers   int
	}{
		{"10repos_1worker", 10, 1},
		{"10repos_5workers", 10, 5},
		{"100repos_1worker", 100, 1},
		{"100repos_10workers", 100, 10},
		{"1000repos_1worker", 1000, 1},
		{"1000repos_20workers", 1000, 20},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Generate test repositories
			repos := generateTestRepos(bm.repoCount)

			// Create mock provider
			mockProvider := mockprovider.NewProvider("github")
			mockProvider.SetupListResponse("testorg", repos)

			// Setup clone operations for all repos
			for _, repo := range repos {
				mockProvider.On("CloneRepository", mock.Anything, repo, mock.AnythingOfType("string"), mock.Anything).Return(nil)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Simulate clone executor execution
				ctx := context.Background()
				_ = simulateCloneExecution(ctx, mockProvider, repos, bm.workers)
			}
		})
	}
}

// BenchmarkListWithFilters benchmarks repository listing with various filters.
func BenchmarkListWithFilters(b *testing.B) {
	repoCount := 10000
	repos := generateTestRepos(repoCount)

	benchmarks := []struct {
		name string
		opts provider.ListOptions
	}{
		{
			name: "NoFilter",
			opts: provider.ListOptions{},
		},
		{
			name: "VisibilityFilter",
			opts: provider.ListOptions{
				Visibility: provider.VisibilityPrivate,
			},
		},
		{
			name: "LanguageFilter",
			opts: provider.ListOptions{
				Language: "Go",
			},
		},
		{
			name: "TopicFilter",
			opts: provider.ListOptions{
				Topic: "microservice",
			},
		},
		{
			name: "MultipleFilters",
			opts: provider.ListOptions{
				Visibility: provider.VisibilityPublic,
				Language:   "Go",
				Topic:      "api",
				MinStars:   10,
			},
		},
		{
			name: "DateRangeFilter",
			opts: provider.ListOptions{
				UpdatedSince: time.Now().AddDate(-1, 0, 0),
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			mockProvider := mockprovider.NewProvider("github")

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				filtered := mockProvider.FilterRepos(repos, bm.opts)
				_ = filtered // Prevent optimization
			}
		})
	}
}

// BenchmarkSyncOperations benchmarks sync operations.
func BenchmarkSyncOperations(b *testing.B) {
	benchmarks := []struct {
		name      string
		repoCount int
		workers   int
	}{
		{"10repos_sync", 10, 1},
		{"10repos_parallel_sync", 10, 3},
		{"100repos_sync", 100, 1},
		{"100repos_parallel_sync", 100, 5},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			sourceRepos := generateTestRepos(bm.repoCount)
			destRepos := []provider.Repository{} // Empty destination

			srcProvider := mockprovider.NewProvider("github")
			dstProvider := mockprovider.NewProvider("gitlab")

			srcProvider.SetupListResponse("sourceorg", sourceRepos)
			dstProvider.SetupListResponse("destorg", destRepos)

			// Setup create operations
			for i, repo := range sourceRepos {
				dstRepo := repo
				dstRepo.ID = fmt.Sprintf("dst-%d", i)
				dstProvider.SetupCreateResponse(func(repoName string) func(provider.CreateRepoRequest) bool {
					return func(req provider.CreateRepoRequest) bool {
						return req.Name == repoName
					}
				}(repo.Name), &dstRepo, nil)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Simulate sync execution
				ctx := context.Background()
				_ = simulateSyncExecution(ctx, srcProvider, dstProvider, sourceRepos, bm.workers)
			}
		})
	}
}

// BenchmarkProviderOperations benchmarks basic provider operations.
func BenchmarkProviderOperations(b *testing.B) {
	mockProvider := mockprovider.NewProvider("github")
	testRepo := generateTestRepos(1)[0]

	b.Run("GetRepository", func(b *testing.B) {
		mockProvider.SetupGetResponse("testorg/testrepo", &testRepo, nil)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			_, _ = mockProvider.GetRepository(ctx, "testorg/testrepo")
		}
	})

	b.Run("CreateRepository", func(b *testing.B) {
		createReq := provider.CreateRepoRequest{
			Name:        "benchmark-repo",
			Description: "Benchmark repository",
			Private:     false,
		}

		mockProvider.SetupCreateResponse(func(req provider.CreateRepoRequest) bool {
			return req.Name == "benchmark-repo"
		}, &testRepo, nil)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			_, _ = mockProvider.CreateRepository(ctx, createReq)
		}
	})

	b.Run("ListRepositories", func(b *testing.B) {
		repos := generateTestRepos(100)
		mockProvider.SetupListResponse("testorg", repos)

		opts := provider.ListOptions{
			Organization: "testorg",
			PerPage:      100,
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			_, _ = mockProvider.ListRepositories(ctx, opts)
		}
	})
}

// BenchmarkMemoryUsage benchmarks memory usage patterns.
func BenchmarkMemoryUsage(b *testing.B) {
	b.Run("LargeRepositoryList", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			repos := generateTestRepos(10000)
			_ = repos // Prevent optimization
		}
	})

	b.Run("MockProviderSetup", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			mockProvider := mockprovider.NewProvider("github")
			repos := generateTestRepos(1000)
			mockProvider.SetupListResponse("testorg", repos)
			_ = mockProvider
		}
	})

	b.Run("FilteringOperations", func(b *testing.B) {
		repos := generateTestRepos(5000)
		mockProvider := mockprovider.NewProvider("github")

		opts := provider.ListOptions{
			Language:   "Go",
			Visibility: provider.VisibilityPublic,
			MinStars:   10,
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			filtered := mockProvider.FilterRepos(repos, opts)
			_ = filtered
		}
	})
}

// BenchmarkConcurrentOperations benchmarks concurrent operations.
func BenchmarkConcurrentOperations(b *testing.B) {
	b.Run("ConcurrentListOperations", func(b *testing.B) {
		mockProvider := mockprovider.NewProvider("github")
		repos := generateTestRepos(100)
		mockProvider.SetupListResponse("testorg", repos)

		opts := provider.ListOptions{
			Organization: "testorg",
		}

		b.ResetTimer()
		b.ReportAllocs()
		b.SetParallelism(10)

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				ctx := context.Background()
				_, _ = mockProvider.ListRepositories(ctx, opts)
			}
		})
	})

	b.Run("ConcurrentFilterOperations", func(b *testing.B) {
		repos := generateTestRepos(1000)
		mockProvider := mockprovider.NewProvider("github")

		opts := provider.ListOptions{
			Language: "Go",
		}

		b.ResetTimer()
		b.ReportAllocs()
		b.SetParallelism(5)

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				filtered := mockProvider.FilterRepos(repos, opts)
				_ = filtered
			}
		})
	})
}

// Helper functions for benchmarks

// generateTestRepos generates a specified number of test repositories for benchmarking.
func generateTestRepos(count int) []provider.Repository {
	repos := make([]provider.Repository, count)
	languages := []string{"Go", "Python", "JavaScript", "TypeScript", "Java", "C++", "Rust", "Ruby"}
	topics := []string{"api", "web", "mobile", "cli", "microservice", "database", "frontend", "backend"}

	for i := 0; i < count; i++ {
		repos[i] = provider.Repository{
			ID:          fmt.Sprintf("repo-%d", i),
			Name:        fmt.Sprintf("repo-%d", i),
			FullName:    fmt.Sprintf("testorg/repo-%d", i),
			Description: fmt.Sprintf("Test repository %d", i),
			Private:     i%3 == 0, // Every 3rd repo is private
			Language:    languages[i%len(languages)],
			Stars:       i * 5,
			Forks:       i * 2,
			Topics:      []string{topics[i%len(topics)]},
			CreatedAt:   time.Now().AddDate(0, -i%12, 0),
			UpdatedAt:   time.Now().AddDate(0, 0, -i%30),
			Visibility: func() provider.VisibilityType {
				if i%3 == 0 {
					return provider.VisibilityPrivate
				}
				return provider.VisibilityPublic
			}(),
		}
	}

	return repos
}

// simulateCloneExecution simulates clone execution for benchmarking.
func simulateCloneExecution(ctx context.Context, mockProvider *mockprovider.Provider, repos []provider.Repository, workers int) error {
	// Simulate clone executor logic
	if workers <= 1 {
		// Sequential execution
		for _, repo := range repos {
			if err := mockProvider.CloneRepository(ctx, repo, "/tmp/test", provider.CloneOptions{}); err != nil {
				return err
			}
		}
	} else {
		// Parallel execution simulation
		semaphore := make(chan struct{}, workers)
		errors := make(chan error, len(repos))

		for _, repo := range repos {
			go func(r provider.Repository) {
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				if err := mockProvider.CloneRepository(ctx, r, "/tmp/test", provider.CloneOptions{}); err != nil {
					errors <- err
					return
				}
				errors <- nil
			}(repo)
		}

		// Wait for all operations to complete
		for i := 0; i < len(repos); i++ {
			if err := <-errors; err != nil {
				return err
			}
		}
	}

	return nil
}

// simulateSyncExecution simulates sync execution for benchmarking.
func simulateSyncExecution(ctx context.Context, srcProvider, dstProvider *mockprovider.Provider, repos []provider.Repository, workers int) error {
	// Simulate sync engine logic
	for _, repo := range repos {
		// Create repository in destination
		createReq := provider.CreateRepoRequest{
			Name:        repo.Name,
			Description: repo.Description,
			Private:     repo.Private,
		}

		if _, err := dstProvider.CreateRepository(ctx, createReq); err != nil {
			return err
		}
	}

	return nil
}
