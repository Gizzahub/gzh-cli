package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ChangeRecord represents a single configuration change
type ChangeRecord struct {
	ID           string                 `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	User         string                 `json:"user"`
	Organization string                 `json:"organization"`
	Repository   string                 `json:"repository"`
	Operation    string                 `json:"operation"` // create, update, delete
	Category     string                 `json:"category"`  // settings, branch_protection, permissions, etc.
	Before       map[string]interface{} `json:"before,omitempty"`
	After        map[string]interface{} `json:"after,omitempty"`
	Description  string                 `json:"description"`
	Source       string                 `json:"source"` // cli, api, web
	RequestID    string                 `json:"request_id,omitempty"`
}

// ChangeLog manages configuration change history
type ChangeLog struct {
	client *RepoConfigClient
	store  ChangeStore
}

// ChangeStore interface for persisting change records
type ChangeStore interface {
	Store(ctx context.Context, record *ChangeRecord) error
	Get(ctx context.Context, id string) (*ChangeRecord, error)
	List(ctx context.Context, filter ChangeFilter) ([]*ChangeRecord, error)
	Delete(ctx context.Context, id string) error
}

// ChangeFilter for querying change records
type ChangeFilter struct {
	Organization string
	Repository   string
	User         string
	Operation    string
	Category     string
	Since        time.Time
	Until        time.Time
	Limit        int
	Offset       int
}

// RollbackRequest represents a rollback operation
type RollbackRequest struct {
	ChangeID    string `json:"change_id"`
	Repository  string `json:"repository"`
	Category    string `json:"category"`
	DryRun      bool   `json:"dry_run"`
	Description string `json:"description"`
}

// RollbackResult contains the result of a rollback operation
type RollbackResult struct {
	Success     bool     `json:"success"`
	ChangeID    string   `json:"change_id"`
	NewChangeID string   `json:"new_change_id,omitempty"`
	Errors      []string `json:"errors,omitempty"`
	DryRun      bool     `json:"dry_run"`
}

// NewChangeLog creates a new change log manager
func NewChangeLog(client *RepoConfigClient, store ChangeStore) *ChangeLog {
	return &ChangeLog{
		client: client,
		store:  store,
	}
}

// RecordChange creates and stores a change record
func (cl *ChangeLog) RecordChange(ctx context.Context, change *ChangeRecord) error {
	if change.ID == "" {
		change.ID = generateChangeID()
	}
	if change.Timestamp.IsZero() {
		change.Timestamp = time.Now()
	}

	return cl.store.Store(ctx, change)
}

// GetChange retrieves a specific change record
func (cl *ChangeLog) GetChange(ctx context.Context, id string) (*ChangeRecord, error) {
	return cl.store.Get(ctx, id)
}

// ListChanges retrieves change records based on filter criteria
func (cl *ChangeLog) ListChanges(ctx context.Context, filter ChangeFilter) ([]*ChangeRecord, error) {
	return cl.store.List(ctx, filter)
}

// Rollback performs a rollback operation to revert a previous change
func (cl *ChangeLog) Rollback(ctx context.Context, request *RollbackRequest) (*RollbackResult, error) {
	result := &RollbackResult{
		ChangeID: request.ChangeID,
		DryRun:   request.DryRun,
	}

	// Get the original change record
	originalChange, err := cl.store.Get(ctx, request.ChangeID)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to get change record: %v", err))
		return result, err
	}

	if originalChange.Before == nil {
		result.Errors = append(result.Errors, "Cannot rollback: no 'before' state available")
		return result, fmt.Errorf("rollback not possible for change %s", request.ChangeID)
	}

	// Validate rollback is applicable
	if originalChange.Repository != request.Repository {
		result.Errors = append(result.Errors, "Repository mismatch in rollback request")
		return result, fmt.Errorf("repository mismatch")
	}

	if originalChange.Category != request.Category {
		result.Errors = append(result.Errors, "Category mismatch in rollback request")
		return result, fmt.Errorf("category mismatch")
	}

	if request.DryRun {
		result.Success = true
		return result, nil
	}

	// Perform the actual rollback based on category
	err = cl.performRollback(ctx, originalChange, request)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Rollback failed: %v", err))
		return result, err
	}

	// Create a new change record for the rollback
	rollbackChange := &ChangeRecord{
		ID:           generateChangeID(),
		Timestamp:    time.Now(),
		User:         getCurrentUser(),
		Organization: originalChange.Organization,
		Repository:   originalChange.Repository,
		Operation:    "rollback",
		Category:     originalChange.Category,
		Before:       originalChange.After,
		After:        originalChange.Before,
		Description:  fmt.Sprintf("Rollback of change %s: %s", request.ChangeID, request.Description),
		Source:       "cli",
	}

	err = cl.store.Store(ctx, rollbackChange)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to record rollback: %v", err))
		return result, err
	}

	result.Success = true
	result.NewChangeID = rollbackChange.ID
	return result, nil
}

// performRollback executes the actual rollback operation
func (cl *ChangeLog) performRollback(ctx context.Context, change *ChangeRecord, request *RollbackRequest) error {
	switch change.Category {
	case "settings":
		return cl.rollbackRepositorySettings(ctx, change)
	case "branch_protection":
		return cl.rollbackBranchProtection(ctx, change)
	case "permissions":
		return cl.rollbackPermissions(ctx, change)
	default:
		return fmt.Errorf("unsupported rollback category: %s", change.Category)
	}
}

// rollbackRepositorySettings reverts repository settings changes
func (cl *ChangeLog) rollbackRepositorySettings(ctx context.Context, change *ChangeRecord) error {
	owner, repo := parseRepositoryFullName(change.Repository)

	// Convert before state to RepositoryUpdate
	var update RepositoryUpdate
	beforeBytes, err := json.Marshal(change.Before)
	if err != nil {
		return fmt.Errorf("failed to marshal before state: %w", err)
	}

	err = json.Unmarshal(beforeBytes, &update)
	if err != nil {
		return fmt.Errorf("failed to unmarshal before state: %w", err)
	}

	_, err = cl.client.UpdateRepository(ctx, owner, repo, &update)
	return err
}

// rollbackBranchProtection reverts branch protection rule changes
func (cl *ChangeLog) rollbackBranchProtection(ctx context.Context, change *ChangeRecord) error {
	// Implementation for branch protection rollback
	return fmt.Errorf("branch protection rollback not yet implemented")
}

// rollbackPermissions reverts permission changes
func (cl *ChangeLog) rollbackPermissions(ctx context.Context, change *ChangeRecord) error {
	// Implementation for permissions rollback
	return fmt.Errorf("permissions rollback not yet implemented")
}

// RecordRepositoryUpdate creates a change record for repository updates
func (cl *ChangeLog) RecordRepositoryUpdate(ctx context.Context, owner, repo string, before, after *Repository, description string) error {
	beforeData := convertRepositoryToMap(before)
	afterData := convertRepositoryToMap(after)

	change := &ChangeRecord{
		Organization: owner,
		Repository:   fmt.Sprintf("%s/%s", owner, repo),
		Operation:    "update",
		Category:     "settings",
		Before:       beforeData,
		After:        afterData,
		Description:  description,
		Source:       "cli",
		User:         getCurrentUser(),
	}

	return cl.RecordChange(ctx, change)
}

// Helper functions
func generateChangeID() string {
	return fmt.Sprintf("ch_%d", time.Now().UnixNano())
}

func getCurrentUser() string {
	// This would typically get the current user from context or environment
	return "system"
}

func parseRepositoryFullName(fullName string) (owner, repo string) {
	// Simple implementation - should be more robust
	parts := strings.Split(fullName, "/")
	if len(parts) >= 2 {
		return parts[0], parts[1]
	}
	return "owner", "repo"
}

func convertRepositoryToMap(repo *Repository) map[string]interface{} {
	if repo == nil {
		return nil
	}

	data := make(map[string]interface{})
	repoBytes, _ := json.Marshal(repo)
	json.Unmarshal(repoBytes, &data)
	return data
}
