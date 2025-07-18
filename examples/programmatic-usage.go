// Package main demonstrates programmatic usage of the gzh-manager-go library
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/archmagece/gzh-manager-go/pkg/bulk-clone"
	"github.com/archmagece/gzh-manager-go/pkg/github"
	"github.com/archmagece/gzh-manager-go/pkg/gitlab"
)

func main() {
	// Example 1: Using the bulk clone manager
	bulkCloneExample()

	// Example 2: Using GitHub client directly
	githubExample()

	// Example 3: Using GitLab client directly
	gitlabExample()
}

func bulkCloneExample() {
	fmt.Println("=== Bulk Clone Example ===")

	// Load configuration from file
	config, err := bulkclone.LoadConfigFromFile("bulk-clone.yaml")
	if err != nil {
		log.Printf("Error loading config: %v", err)
		return
	}

	// Create bulk clone manager
	manager := bulkclone.NewDefaultManager(config)

	// Clone an organization
	ctx := context.Background()
	request := &bulkclone.OrganizationCloneRequest{
		Provider:     "github",
		Organization: "kubernetes",
		TargetPath:   "/tmp/repos",
		Strategy:     "fetch",
		Concurrency:  5,
		DryRun:       true,
		Token:        os.Getenv("GITHUB_TOKEN"),
	}

	result, err := manager.CloneOrganization(ctx, request)
	if err != nil {
		log.Printf("Clone failed: %v", err)
		return
	}

	fmt.Printf("Clone completed: %d successful, %d failed\n",
		result.ClonesSuccessful, result.ClonesFailed)
}

func githubExample() {
	fmt.Println("\n=== GitHub Direct Example ===")

	// Create GitHub client
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Println("GITHUB_TOKEN not set")
		return
	}

	client := github.NewCachedGitHubClient(token)

	// List repositories for an organization
	ctx := context.Background()
	repos, err := client.ListOrganizationRepositories(ctx, "golang")
	if err != nil {
		log.Printf("Error listing repos: %v", err)
		return
	}

	fmt.Printf("Found %d repositories in golang org\n", len(repos))
	for i, repo := range repos {
		if i >= 5 {
			fmt.Println("...")
			break
		}
		fmt.Printf("  - %s: %s\n", repo.Name, repo.Description)
	}
}

func gitlabExample() {
	fmt.Println("\n=== GitLab Direct Example ===")

	// Create GitLab client
	token := os.Getenv("GITLAB_TOKEN")
	if token == "" {
		log.Println("GITLAB_TOKEN not set")
		return
	}

	config := gitlab.ClientConfig{
		Token:   token,
		BaseURL: "https://gitlab.com", // or your GitLab instance
	}

	client, err := gitlab.NewClient(config)
	if err != nil {
		log.Printf("Error creating GitLab client: %v", err)
		return
	}

	// List groups
	ctx := context.Background()
	groups, err := client.ListGroups(ctx)
	if err != nil {
		log.Printf("Error listing groups: %v", err)
		return
	}

	fmt.Printf("Found %d accessible groups\n", len(groups))
}

// Example: Custom repository filter
type CustomFilter struct{}

func (f *CustomFilter) ShouldClone(repo *bulkclone.DiscoveredRepository) bool {
	// Clone only Go repositories
	return repo.Language == "Go" && !repo.Archived
}

func customFilterExample() {
	fmt.Println("\n=== Custom Filter Example ===")

	// Use custom filter in clone request
	request := &bulkclone.OrganizationCloneRequest{
		Provider:     "github",
		Organization: "kubernetes",
		TargetPath:   "/tmp/go-repos",
		Filters: &bulkclone.RepositoryFilters{
			Language:      []string{"Go"},
			ExcludeTopics: []string{"deprecated"},
			MinStars:      10,
		},
		Concurrency: 5,
		DryRun:      true,
	}

	// The manager will apply these filters during discovery
	_ = request
}