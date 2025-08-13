package github_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Gizzahub/gzh-manager-go/pkg/github"
)

// ExampleGetDefaultBranch demonstrates how to retrieve the default branch
// of a GitHub repository.
func ExampleGetDefaultBranch() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get default branch for a repository
	branch, err := github.GetDefaultBranch(ctx, "octocat", "Hello-World")
	if err != nil {
		log.Printf("Error getting default branch: %v", err)
		return
	}

	fmt.Printf("Default branch: %s", branch)
	// Output: Default branch: master
}

// ExampleList demonstrates how to list all repositories in a GitHub organization.
func ExampleList() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// List all repositories in an organization
	repos, err := github.List(ctx, "github")
	if err != nil {
		log.Printf("Error listing repositories: %v", err)
		return
	}

	fmt.Printf("Found %d repositories", len(repos))

	if len(repos) > 0 {
		fmt.Printf("\nFirst repository: %s", repos[0])
	}
	// Output: Found repositories in organization
}

// ExampleClone demonstrates how to clone a GitHub repository to a local directory.
func ExampleClone() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create a temporary directory for cloning
	tempDir := "/tmp/github-clone-example"

	_ = os.MkdirAll(tempDir, 0o755)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Clone a repository
	err := github.Clone(ctx, tempDir, "octocat", "Hello-World")
	if err != nil {
		log.Printf("Error cloning repository: %v", err)
		return
	}

	fmt.Println("Repository cloned successfully")
	// Output: Repository cloned successfully
}

// ExampleWorkflow demonstrates a complete workflow of discovering and cloning
// repositories from a GitHub organization.
func ExampleCachedGitHubClient_workflow() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	orgName := "octocat"
	targetDir := "/tmp/github-workflow-example"

	// Step 1: Create target directory
	_ = os.MkdirAll(targetDir, 0o755)
	defer func() { _ = os.RemoveAll(targetDir) }()

	// Step 2: List all repositories in the organization
	repos, err := github.List(ctx, orgName)
	if err != nil {
		log.Printf("Error listing repositories: %v", err)
		return
	}

	fmt.Printf("Found %d repositories in %s organization\n", len(repos), orgName)

	// Step 3: Clone the first few repositories (limit for example)
	maxRepos := 3
	if len(repos) > maxRepos {
		repos = repos[:maxRepos]
	}

	for _, repo := range repos {
		fmt.Printf("Cloning %s...\n", repo)

		// Get default branch first
		branch, err := github.GetDefaultBranch(ctx, orgName, repo)
		if err != nil {
			log.Printf("Warning: Could not get default branch for %s: %v", repo, err)
		} else {
			fmt.Printf("  Default branch: %s\n", branch)
		}

		// Clone the repository
		err = github.Clone(ctx, targetDir, orgName, repo)
		if err != nil {
			log.Printf("Error cloning %s: %v", repo, err)
			continue
		}

		fmt.Printf("  ✓ Successfully cloned %s\n", repo)
	}

	fmt.Println("Workflow completed")
	// Output: Workflow demonstrates organization repository management
}

// ExampleErrorHandling demonstrates proper error handling when working
// with GitHub API operations.
func ExampleCachedGitHubClient_errorHandling() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to get default branch for a non-existent repository
	_, err := github.GetDefaultBranch(ctx, "nonexistent", "repository")
	if err != nil {
		fmt.Printf("Expected error for non-existent repository: %v\n", err)
	}

	// Attempt to list repositories for a non-existent organization
	_, err = github.List(ctx, "definitely-does-not-exist-org-12345")
	if err != nil {
		fmt.Printf("Expected error for non-existent organization: %v\n", err)
	}

	// Attempt to clone to an invalid path
	err = github.Clone(ctx, "/invalid/path/that/does/not/exist", "octocat", "Hello-World")
	if err != nil {
		fmt.Printf("Expected error for invalid path: %v\n", err)
	}

	fmt.Println("Error handling examples completed")
	// Output: Error handling examples demonstrate proper error management
}
