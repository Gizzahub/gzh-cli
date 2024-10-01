package github

import (
	"fmt"
	"testing"
)

func TestList(t *testing.T) {
	t.SkipNow()
	org := "ScriptonBasestar"

	repos, err := List(org)
	if err != nil {
		t.Fatalf("Failed to list repositories: %v", err)
	}

	fmt.Printf("Repositories in organization %s:\n", org)
	for _, repo := range repos {
		fmt.Println(repo)
	}
}
