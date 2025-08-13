// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package sync

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Gizzahub/gzh-manager-go/pkg/git/provider"
)

// IssueSyncer handles synchronization of issues and pull requests.
type IssueSyncer struct {
	source      provider.GitProvider
	destination provider.GitProvider
	mapping     map[string]string // source ID -> destination ID mapping
}

// Issue represents a repository issue.
type Issue struct {
	ID        string
	Number    int
	Title     string
	Body      string
	State     string
	Labels    []string
	Assignees []string
	Milestone string
	URL       string
	CreatedAt time.Time
	UpdatedAt time.Time
	Author    string
	Comments  []Comment
}

// Comment represents an issue comment.
type Comment struct {
	ID        string
	Body      string
	Author    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateIssueRequest represents a request to create an issue.
type CreateIssueRequest struct {
	Title     string
	Body      string
	Labels    []string
	Assignees []string
	Milestone string
	State     string
}

// Sync synchronizes issues between source and destination repositories.
func (i *IssueSyncer) Sync(ctx context.Context, srcRepo, dstRepo provider.Repository) error {
	// Check provider capability for issues
	if !i.hasIssueSupport() {
		return fmt.Errorf("one or both providers don't support issues")
	}

	fmt.Printf("  🐛 Synchronizing issues for %s...\n", srcRepo.FullName)

	// Get source issues
	sourceIssues, err := i.getSourceIssues(ctx, srcRepo.ID)
	if err != nil {
		return fmt.Errorf("failed to get source issues: %w", err)
	}

	if len(sourceIssues) == 0 {
		fmt.Printf("    No issues found in source repository\n")
		return nil
	}

	// Get existing issues in destination
	existingIssues, err := i.getDestinationIssues(ctx, dstRepo.ID)
	if err != nil {
		return fmt.Errorf("failed to get destination issues: %w", err)
	}

	// Synchronize issues
	synced := 0
	created := 0
	updated := 0
	skipped := 0

	for _, issue := range sourceIssues {
		if existing := i.findExistingIssue(issue, existingIssues); existing != nil {
			// Update existing issue
			if err := i.updateIssue(ctx, dstRepo.ID, existing, &issue); err != nil {
				fmt.Printf("    ⚠️  Failed to update issue #%d: %v\n", issue.Number, err)
				skipped++
			} else {
				updated++
				synced++
			}
		} else {
			// Create new issue
			newIssue, err := i.createIssue(ctx, dstRepo.ID, issue)
			if err != nil {
				fmt.Printf("    ⚠️  Failed to create issue #%d: %v\n", issue.Number, err)
				skipped++
			} else {
				i.mapping[issue.ID] = newIssue.ID
				created++
				synced++
			}
		}
	}

	// Sync comments for existing issues
	if err := i.syncComments(ctx, srcRepo, dstRepo, sourceIssues, existingIssues); err != nil {
		fmt.Printf("    ⚠️  Failed to sync comments: %v\n", err)
	}

	fmt.Printf("    ✅ Issues synchronized: %d total (%d created, %d updated, %d skipped)\n",
		synced, created, updated, skipped)

	return nil
}

// hasIssueSupport checks if both providers support issues.
func (i *IssueSyncer) hasIssueSupport() bool {
	// TODO: Implement capability checking
	// For now, assume all providers support issues
	return true
}

// getSourceIssues retrieves all issues from the source repository.
func (i *IssueSyncer) getSourceIssues(ctx context.Context, repoID string) ([]Issue, error) {
	// TODO: Implement actual issue retrieval from provider
	// This would use the provider's ListIssues method
	// For now, return empty slice
	return []Issue{}, nil
}

// getDestinationIssues retrieves all issues from the destination repository.
func (i *IssueSyncer) getDestinationIssues(ctx context.Context, repoID string) ([]Issue, error) {
	// TODO: Implement actual issue retrieval from provider
	// This would use the provider's ListIssues method
	// For now, return empty slice
	return []Issue{}, nil
}

// findExistingIssue finds an existing issue in the destination that matches the source issue.
func (i *IssueSyncer) findExistingIssue(sourceIssue Issue, existingIssues []Issue) *Issue {
	// Try to match by title first (most reliable for migrated issues)
	for _, existing := range existingIssues {
		if i.issuesMatch(sourceIssue, existing) {
			return &existing
		}
	}
	return nil
}

// issuesMatch determines if two issues are the same.
func (i *IssueSyncer) issuesMatch(source, existing Issue) bool {
	// Match by title and creation date (within reasonable tolerance)
	if source.Title == existing.Title {
		// Check if creation dates are close (within 1 hour)
		timeDiff := source.CreatedAt.Sub(existing.CreatedAt)
		if timeDiff < 0 {
			timeDiff = -timeDiff
		}
		if timeDiff < time.Hour {
			return true
		}
	}

	// Match by sync marker in body
	if strings.Contains(existing.Body, fmt.Sprintf("_Synced from %s_", source.URL)) {
		return true
	}

	return false
}

// createIssue creates a new issue in the destination repository.
func (i *IssueSyncer) createIssue(ctx context.Context, repoID string, issue Issue) (*Issue, error) {
	// Add sync marker to body
	body := fmt.Sprintf("%s\n\n---\n_Synced from %s on %s_",
		issue.Body, issue.URL, time.Now().Format("2006-01-02"))

	request := CreateIssueRequest{
		Title:     issue.Title,
		Body:      body,
		Labels:    issue.Labels,
		Assignees: issue.Assignees,
		Milestone: issue.Milestone,
		State:     issue.State,
	}

	// TODO: Implement actual issue creation using provider
	// For now, return a mock issue
	newIssue := &Issue{
		ID:        fmt.Sprintf("new-%s", issue.ID),
		Title:     request.Title,
		Body:      request.Body,
		State:     request.State,
		Labels:    request.Labels,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return newIssue, nil
}

// updateIssue updates an existing issue in the destination repository.
func (i *IssueSyncer) updateIssue(ctx context.Context, repoID string, existing, source *Issue) error {
	// Check if update is needed
	if i.issueNeedsUpdate(*existing, *source) {
		// TODO: Implement actual issue update using provider
		// For now, just return success
		return nil
	}
	return nil
}

// issueNeedsUpdate determines if an issue needs to be updated.
func (i *IssueSyncer) issueNeedsUpdate(existing, source Issue) bool {
	// Check if title, body, state, or labels have changed
	if existing.Title != source.Title {
		return true
	}
	if existing.State != source.State {
		return true
	}
	if !i.slicesEqual(existing.Labels, source.Labels) {
		return true
	}
	// Check if body changed (excluding sync marker)
	existingBody := i.removeSyncMarker(existing.Body)
	if existingBody != source.Body {
		return true
	}
	return false
}

// syncComments synchronizes comments for issues.
func (i *IssueSyncer) syncComments(ctx context.Context, srcRepo, dstRepo provider.Repository,
	sourceIssues, existingIssues []Issue,
) error {
	// TODO: Implement comment synchronization
	// This would involve:
	// 1. Getting comments for each source issue
	// 2. Finding corresponding destination issue
	// 3. Creating missing comments
	// 4. Updating existing comments if needed

	return nil
}

// removeSyncMarker removes the sync marker from issue body.
func (i *IssueSyncer) removeSyncMarker(body string) string {
	// Remove the sync marker line
	lines := strings.Split(body, "\n")
	var cleaned []string

	for _, line := range lines {
		if !strings.Contains(line, "_Synced from") && !strings.Contains(line, "---") {
			cleaned = append(cleaned, line)
		}
	}

	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

// slicesEqual compares two string slices for equality.
func (i *IssueSyncer) slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// Create maps for comparison
	aMap := make(map[string]bool)
	bMap := make(map[string]bool)

	for _, item := range a {
		aMap[item] = true
	}
	for _, item := range b {
		bMap[item] = true
	}

	// Compare maps
	for item := range aMap {
		if !bMap[item] {
			return false
		}
	}
	for item := range bMap {
		if !aMap[item] {
			return false
		}
	}

	return true
}

// GetSyncStats returns statistics about issue synchronization.
func (i *IssueSyncer) GetSyncStats() IssueSyncStats {
	return IssueSyncStats{
		TotalMapped:   len(i.mapping),
		SourceIssues:  0, // TODO: Track during sync
		CreatedIssues: 0, // TODO: Track during sync
		UpdatedIssues: 0, // TODO: Track during sync
		SkippedIssues: 0, // TODO: Track during sync
	}
}

// IssueSyncStats represents statistics for issue synchronization.
type IssueSyncStats struct {
	TotalMapped   int `json:"total_mapped"`
	SourceIssues  int `json:"source_issues"`
	CreatedIssues int `json:"created_issues"`
	UpdatedIssues int `json:"updated_issues"`
	SkippedIssues int `json:"skipped_issues"`
}
