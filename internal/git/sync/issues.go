// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Gizzahub/gzh-cli/pkg/git/provider"
)

// IssueSyncer handles synchronization of issues and pull requests.
type IssueSyncer struct {
	source      provider.GitProvider
	destination provider.GitProvider
	mapping     map[string]string // source ID -> destination ID mapping
	httpClient  *http.Client
	sourceToken string
	destToken   string
	baseURL     string // GitHub API base URL
	stats       IssueSyncStats
}

// NewIssueSyncer creates a new IssueSyncer instance.
func NewIssueSyncer(source, destination provider.GitProvider) *IssueSyncer {
	return &IssueSyncer{
		source:      source,
		destination: destination,
		mapping:     make(map[string]string),
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		baseURL:     "https://api.github.com",
	}
}

// SetTokens sets the authentication tokens for source and destination APIs.
func (i *IssueSyncer) SetTokens(sourceToken, destToken string) {
	i.sourceToken = sourceToken
	i.destToken = destToken
}

// SetBaseURL sets the GitHub API base URL (useful for GitHub Enterprise).
func (i *IssueSyncer) SetBaseURL(baseURL string) {
	i.baseURL = strings.TrimSuffix(baseURL, "/")
}

// doRequest performs an HTTP request with authentication.
func (i *IssueSyncer) doRequest(ctx context.Context, method, url, token string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "gzh-cli")

	return i.httpClient.Do(req)
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
	return i.listIssues(ctx, repoID, i.sourceToken)
}

// getDestinationIssues retrieves all issues from the destination repository.
func (i *IssueSyncer) getDestinationIssues(ctx context.Context, repoID string) ([]Issue, error) {
	return i.listIssues(ctx, repoID, i.destToken)
}

// listIssues retrieves all issues from a repository using GitHub API.
// repoID should be in "owner/repo" format.
func (i *IssueSyncer) listIssues(ctx context.Context, repoID, token string) ([]Issue, error) {
	var allIssues []Issue
	page := 1
	perPage := 100

	for {
		url := fmt.Sprintf("%s/repos/%s/issues?state=all&page=%d&per_page=%d",
			i.baseURL, repoID, page, perPage)

		resp, err := i.doRequest(ctx, http.MethodGet, url, token, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to list issues: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("failed to list issues: HTTP %d - %s", resp.StatusCode, string(body))
		}

		var ghIssues []githubIssue
		if err := json.NewDecoder(resp.Body).Decode(&ghIssues); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode issues: %w", err)
		}
		resp.Body.Close()

		// Convert GitHub issues to our Issue type
		for _, gh := range ghIssues {
			// Skip pull requests (GitHub API returns them as issues)
			if gh.PullRequest != nil {
				continue
			}

			issue := Issue{
				ID:        fmt.Sprintf("%d", gh.ID),
				Number:    gh.Number,
				Title:     gh.Title,
				Body:      gh.Body,
				State:     gh.State,
				URL:       gh.HTMLURL,
				CreatedAt: gh.CreatedAt,
				UpdatedAt: gh.UpdatedAt,
				Author:    gh.User.Login,
			}

			// Extract labels
			for _, label := range gh.Labels {
				issue.Labels = append(issue.Labels, label.Name)
			}

			// Extract assignees
			for _, assignee := range gh.Assignees {
				issue.Assignees = append(issue.Assignees, assignee.Login)
			}

			// Extract milestone
			if gh.Milestone != nil {
				issue.Milestone = gh.Milestone.Title
			}

			allIssues = append(allIssues, issue)
		}

		// Check if there are more pages
		if len(ghIssues) < perPage {
			break
		}

		page++

		// Context cancellation check
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}

	i.stats.SourceIssues = len(allIssues)
	return allIssues, nil
}

// githubIssue represents a GitHub issue from the API.
type githubIssue struct {
	ID          int64              `json:"id"`
	Number      int                `json:"number"`
	Title       string             `json:"title"`
	Body        string             `json:"body"`
	State       string             `json:"state"`
	HTMLURL     string             `json:"html_url"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	User        githubUser         `json:"user"`
	Labels      []githubLabel      `json:"labels"`
	Assignees   []githubUser       `json:"assignees"`
	Milestone   *githubMilestone   `json:"milestone"`
	PullRequest *githubPullRequest `json:"pull_request"`
}

type githubUser struct {
	Login string `json:"login"`
}

type githubLabel struct {
	Name string `json:"name"`
}

type githubMilestone struct {
	Title string `json:"title"`
}

type githubPullRequest struct {
	URL string `json:"url"`
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
	body := fmt.Sprintf("%s\n\n---\n_Synced from %s on %s_\n_Original author: @%s_",
		issue.Body, issue.URL, time.Now().Format("2006-01-02"), issue.Author)

	// Build request body for GitHub API
	apiRequest := map[string]interface{}{
		"title": issue.Title,
		"body":  body,
	}

	// Add labels if available
	if len(issue.Labels) > 0 {
		apiRequest["labels"] = issue.Labels
	}

	// Add assignees if available (note: may fail if users don't exist in dest repo)
	// We skip assignees for now to avoid errors
	// if len(issue.Assignees) > 0 {
	// 	apiRequest["assignees"] = issue.Assignees
	// }

	url := fmt.Sprintf("%s/repos/%s/issues", i.baseURL, repoID)
	resp, err := i.doRequest(ctx, http.MethodPost, url, i.destToken, apiRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create issue: HTTP %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var ghIssue githubIssue
	if err := json.NewDecoder(resp.Body).Decode(&ghIssue); err != nil {
		return nil, fmt.Errorf("failed to decode created issue: %w", err)
	}

	// If source issue was closed, close the new issue too
	if issue.State == "closed" {
		if err := i.closeIssue(ctx, repoID, ghIssue.Number); err != nil {
			// Log but don't fail - issue was created successfully
			fmt.Printf("    ⚠️  Failed to close synced issue #%d: %v\n", ghIssue.Number, err)
		}
	}

	newIssue := &Issue{
		ID:        fmt.Sprintf("%d", ghIssue.ID),
		Number:    ghIssue.Number,
		Title:     ghIssue.Title,
		Body:      ghIssue.Body,
		State:     ghIssue.State,
		URL:       ghIssue.HTMLURL,
		CreatedAt: ghIssue.CreatedAt,
		UpdatedAt: ghIssue.UpdatedAt,
	}

	i.stats.CreatedIssues++
	return newIssue, nil
}

// closeIssue closes an issue.
func (i *IssueSyncer) closeIssue(ctx context.Context, repoID string, issueNumber int) error {
	apiRequest := map[string]interface{}{
		"state": "closed",
	}

	url := fmt.Sprintf("%s/repos/%s/issues/%d", i.baseURL, repoID, issueNumber)
	resp, err := i.doRequest(ctx, http.MethodPatch, url, i.destToken, apiRequest)
	if err != nil {
		return fmt.Errorf("failed to close issue: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to close issue: HTTP %d - %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// updateIssue updates an existing issue in the destination repository.
func (i *IssueSyncer) updateIssue(ctx context.Context, repoID string, existing, source *Issue) error {
	// Check if update is needed
	if !i.issueNeedsUpdate(*existing, *source) {
		return nil
	}

	// Build update request
	apiRequest := make(map[string]interface{})

	// Update title if changed
	if existing.Title != source.Title {
		apiRequest["title"] = source.Title
	}

	// Update body (preserving sync marker)
	existingBody := i.removeSyncMarker(existing.Body)
	if existingBody != source.Body {
		body := fmt.Sprintf("%s\n\n---\n_Synced from %s on %s_\n_Original author: @%s_\n_Last updated: %s_",
			source.Body, source.URL, time.Now().Format("2006-01-02"), source.Author, time.Now().Format("2006-01-02 15:04:05"))
		apiRequest["body"] = body
	}

	// Update state if changed
	if existing.State != source.State {
		apiRequest["state"] = source.State
	}

	// Update labels if changed
	if !i.slicesEqual(existing.Labels, source.Labels) {
		apiRequest["labels"] = source.Labels
	}

	// Skip if no updates needed
	if len(apiRequest) == 0 {
		return nil
	}

	url := fmt.Sprintf("%s/repos/%s/issues/%d", i.baseURL, repoID, existing.Number)
	resp, err := i.doRequest(ctx, http.MethodPatch, url, i.destToken, apiRequest)
	if err != nil {
		return fmt.Errorf("failed to update issue: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update issue: HTTP %d - %s", resp.StatusCode, string(bodyBytes))
	}

	i.stats.UpdatedIssues++
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
	// Sync comments for each matched issue pair
	for _, srcIssue := range sourceIssues {
		dstIssue := i.findExistingIssue(srcIssue, existingIssues)
		if dstIssue == nil {
			// Issue doesn't exist in destination, skip comments
			continue
		}

		// Get source comments
		srcComments, err := i.getIssueComments(ctx, srcRepo.FullName, srcIssue.Number, i.sourceToken)
		if err != nil {
			fmt.Printf("    ⚠️  Failed to get comments for issue #%d: %v\n", srcIssue.Number, err)
			continue
		}

		if len(srcComments) == 0 {
			continue
		}

		// Get destination comments to avoid duplicates
		dstComments, err := i.getIssueComments(ctx, dstRepo.FullName, dstIssue.Number, i.destToken)
		if err != nil {
			fmt.Printf("    ⚠️  Failed to get existing comments for issue #%d: %v\n", dstIssue.Number, err)
			continue
		}

		// Create missing comments
		for _, srcComment := range srcComments {
			if i.commentExists(srcComment, dstComments) {
				continue
			}

			if err := i.createComment(ctx, dstRepo.FullName, dstIssue.Number, srcComment); err != nil {
				fmt.Printf("    ⚠️  Failed to create comment on issue #%d: %v\n", dstIssue.Number, err)
			}
		}
	}

	return nil
}

// getIssueComments retrieves comments for an issue.
func (i *IssueSyncer) getIssueComments(ctx context.Context, repoID string, issueNumber int, token string) ([]Comment, error) {
	url := fmt.Sprintf("%s/repos/%s/issues/%d/comments", i.baseURL, repoID, issueNumber)
	resp, err := i.doRequest(ctx, http.MethodGet, url, token, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get comments: HTTP %d - %s", resp.StatusCode, string(body))
	}

	var ghComments []struct {
		ID        int64      `json:"id"`
		Body      string     `json:"body"`
		User      githubUser `json:"user"`
		CreatedAt time.Time  `json:"created_at"`
		UpdatedAt time.Time  `json:"updated_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ghComments); err != nil {
		return nil, fmt.Errorf("failed to decode comments: %w", err)
	}

	var comments []Comment
	for _, gh := range ghComments {
		comments = append(comments, Comment{
			ID:        fmt.Sprintf("%d", gh.ID),
			Body:      gh.Body,
			Author:    gh.User.Login,
			CreatedAt: gh.CreatedAt,
			UpdatedAt: gh.UpdatedAt,
		})
	}

	return comments, nil
}

// commentExists checks if a comment already exists in destination.
func (i *IssueSyncer) commentExists(srcComment Comment, dstComments []Comment) bool {
	// Check by looking for sync marker in existing comments
	syncMarker := fmt.Sprintf("_Synced comment from @%s_", srcComment.Author)
	for _, dstComment := range dstComments {
		if strings.Contains(dstComment.Body, syncMarker) {
			// Further check: compare first 50 chars of original body
			srcPrefix := srcComment.Body
			if len(srcPrefix) > 50 {
				srcPrefix = srcPrefix[:50]
			}
			if strings.Contains(dstComment.Body, srcPrefix) {
				return true
			}
		}
	}
	return false
}

// createComment creates a comment on an issue.
func (i *IssueSyncer) createComment(ctx context.Context, repoID string, issueNumber int, comment Comment) error {
	body := fmt.Sprintf("%s\n\n---\n_Synced comment from @%s on %s_",
		comment.Body, comment.Author, comment.CreatedAt.Format("2006-01-02 15:04:05"))

	apiRequest := map[string]interface{}{
		"body": body,
	}

	url := fmt.Sprintf("%s/repos/%s/issues/%d/comments", i.baseURL, repoID, issueNumber)
	resp, err := i.doRequest(ctx, http.MethodPost, url, i.destToken, apiRequest)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create comment: HTTP %d - %s", resp.StatusCode, string(bodyBytes))
	}

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
	i.stats.TotalMapped = len(i.mapping)
	return i.stats
}

// ResetStats resets the sync statistics.
func (i *IssueSyncer) ResetStats() {
	i.stats = IssueSyncStats{}
}

// IssueSyncStats represents statistics for issue synchronization.
type IssueSyncStats struct {
	TotalMapped   int `json:"total_mapped"`
	SourceIssues  int `json:"source_issues"`
	CreatedIssues int `json:"created_issues"`
	UpdatedIssues int `json:"updated_issues"`
	SkippedIssues int `json:"skipped_issues"`
}
