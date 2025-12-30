package gitlab_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gizzahub/gzh-cli/pkg/gitlab"
)

// ExampleGetDefaultBranch demonstrates how to retrieve the default branch
// of a GitLab project.
func ExampleGetDefaultBranch() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get default branch for a project
	branch, err := gitlab.GetDefaultBranch(ctx, "gitlab-org", "gitlab")
	if err != nil {
		log.Printf("Error getting default branch: %v", err)
		return
	}

	fmt.Printf("Default branch: %s", branch)
	// Output: Default branch: master
}

// ExampleList demonstrates how to list all projects in a GitLab group.
func ExampleList() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// List all projects in a group
	projects, err := gitlab.List(ctx, "gitlab-org")
	if err != nil {
		log.Printf("Error listing projects: %v", err)
		return
	}

	fmt.Printf("Found %d projects", len(projects))

	if len(projects) > 0 {
		fmt.Printf("\nFirst project: %s", projects[0])
	}
	// Output: Found projects in group
}

// ExampleClone demonstrates how to clone a GitLab project to a local directory.
func ExampleClone() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create a temporary directory for cloning
	tempDir := "/tmp/gitlab-clone-example"

	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		log.Printf("Warning: failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			log.Printf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	// Clone a project with specific branch
	err := gitlab.Clone(ctx, tempDir, "gitlab-org", "gitlab", "master")
	if err != nil {
		log.Printf("Error cloning project: %v", err)
		return
	}

	fmt.Println("Project cloned successfully")
	// Output: Project cloned successfully
}

// ExampleClone_defaultBranch demonstrates cloning with automatic default branch detection.
func ExampleClone_defaultBranch() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create a temporary directory for cloning
	tempDir := "/tmp/gitlab-default-branch-example"

	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		log.Printf("Warning: failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			log.Printf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	// Clone a project without specifying branch (uses default)
	err := gitlab.Clone(ctx, tempDir, "gitlab-org", "gitlab", "")
	if err != nil {
		log.Printf("Error cloning project: %v", err)
		return
	}

	fmt.Println("Project cloned with default branch")
	// Output: Project cloned with default branch
}

// ExampleList_groupWorkflow demonstrates a complete workflow of discovering and cloning
// projects from a GitLab group.
func ExampleList_groupWorkflow() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	groupName := "gitlab-org"
	targetDir := "/tmp/gitlab-workflow-example"

	// Step 1: Create target directory
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		log.Printf("Warning: failed to create target dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(targetDir); err != nil {
			log.Printf("Warning: failed to remove target dir: %v", err)
		}
	}()

	// Step 2: List all projects in the group
	projects, err := gitlab.List(ctx, groupName)
	if err != nil {
		log.Printf("Error listing projects: %v", err)
		return
	}

	fmt.Printf("Found %d projects in %s group\n", len(projects), groupName)

	// Step 3: Clone the first few projects (limit for example)
	maxProjects := 2
	if len(projects) > maxProjects {
		projects = projects[:maxProjects]
	}

	for _, project := range projects {
		fmt.Printf("Processing %s...\n", project)

		// Get default branch first
		branch, err := gitlab.GetDefaultBranch(ctx, groupName, project)
		if err != nil {
			log.Printf("Warning: Could not get default branch for %s: %v", project, err)

			branch = "main" // fallback
		}

		fmt.Printf("  Default branch: %s\n", branch)

		// Clone the project
		err = gitlab.Clone(ctx, targetDir, groupName, project, branch)
		if err != nil {
			log.Printf("Error cloning %s: %v", project, err)
			continue
		}

		fmt.Printf("  ✓ Successfully cloned %s\n", project)
	}

	fmt.Println("Group workflow completed")
	// Output: Group workflow demonstrates GitLab project management
}

// ExampleGetDefaultBranch_authentication demonstrates handling authentication for private projects.
func ExampleGetDefaultBranch_authentication() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Set GitLab token (in real usage, this would come from environment)
	// os.Setenv("GITLAB_TOKEN", "your-gitlab-token")

	// This example would work with proper authentication
	fmt.Println("Authentication setup:")
	fmt.Println("1. Set GITLAB_TOKEN environment variable")
	fmt.Println("2. Ensure token has read access to repositories")
	fmt.Println("3. Use same API calls as public repositories")

	// Example with authentication context
	_, err := gitlab.GetDefaultBranch(ctx, "private-group", "private-project")
	if err != nil {
		fmt.Printf("Expected error without proper authentication: %v\n", err)
	}

	fmt.Println("Authentication example completed")
	// Output: Authentication setup guide for private repositories
}

// ExampleGetDefaultBranch_errorHandling demonstrates proper error handling when working
// with GitLab API operations.
func ExampleGetDefaultBranch_errorHandling() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to get default branch for a non-existent project
	_, err := gitlab.GetDefaultBranch(ctx, "nonexistent", "project")
	if err != nil {
		fmt.Printf("Expected error for non-existent project: %v\n", err)
	}

	// Attempt to list projects for a non-existent group
	_, err = gitlab.List(ctx, "definitely-does-not-exist-group-12345")
	if err != nil {
		fmt.Printf("Expected error for non-existent group: %v\n", err)
	}

	// Attempt to clone with invalid parameters
	err = gitlab.Clone(ctx, "/invalid/path", "group", "project", "nonexistent-branch")
	if err != nil {
		fmt.Printf("Expected error for invalid clone: %v\n", err)
	}

	fmt.Println("Error handling examples completed")
	// Output: Error handling examples demonstrate proper error management
}
