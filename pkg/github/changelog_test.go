package github

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStore(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "changelog_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := NewFileStore(tempDir)
	require.NoError(t, err)

	ctx := context.Background()

	// Test storing a change record
	record := &ChangeRecord{
		ID:           "test-123",
		Timestamp:    time.Now(),
		User:         "testuser",
		Organization: "testorg",
		Repository:   "testorg/testrepo",
		Operation:    "update",
		Category:     "settings",
		Before:       map[string]interface{}{"private": false},
		After:        map[string]interface{}{"private": true},
		Description:  "Changed repository visibility",
		Source:       "cli",
	}

	err = store.Store(ctx, record)
	require.NoError(t, err)

	// Test retrieving the record
	retrieved, err := store.Get(ctx, "test-123")
	require.NoError(t, err)
	assert.Equal(t, record.ID, retrieved.ID)
	assert.Equal(t, record.User, retrieved.User)
	assert.Equal(t, record.Organization, retrieved.Organization)

	// Test listing records
	filter := ChangeFilter{
		Organization: "testorg",
	}
	records, err := store.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, "test-123", records[0].ID)

	// Test deleting a record
	err = store.Delete(ctx, "test-123")
	require.NoError(t, err)

	// Verify record is deleted
	_, err = store.Get(ctx, "test-123")
	assert.Error(t, err)
}

func TestChangeLog(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "changelog_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := NewFileStore(tempDir)
	require.NoError(t, err)

	// Create a mock client (we'll use nil for this test)
	changelog := NewChangeLog(nil, store)

	ctx := context.Background()

	// Test recording a change
	record := &ChangeRecord{
		User:         "testuser",
		Organization: "testorg",
		Repository:   "testorg/testrepo",
		Operation:    "update",
		Category:     "settings",
		Before:       map[string]interface{}{"private": false},
		After:        map[string]interface{}{"private": true},
		Description:  "Changed repository visibility",
		Source:       "cli",
	}

	err = changelog.RecordChange(ctx, record)
	require.NoError(t, err)
	assert.NotEmpty(t, record.ID)
	assert.False(t, record.Timestamp.IsZero())

	// Test retrieving the change
	retrieved, err := changelog.GetChange(ctx, record.ID)
	require.NoError(t, err)
	assert.Equal(t, record.ID, retrieved.ID)

	// Test listing changes
	filter := ChangeFilter{
		Organization: "testorg",
		Limit:        10,
	}
	changes, err := changelog.ListChanges(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, changes, 1)
}

func TestChangeFilter(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "changelog_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := NewFileStore(tempDir)
	require.NoError(t, err)

	ctx := context.Background()

	// Create multiple test records
	records := []*ChangeRecord{
		{
			ID:           "test-1",
			Timestamp:    time.Now().Add(-2 * time.Hour),
			User:         "user1",
			Organization: "org1",
			Repository:   "org1/repo1",
			Operation:    "create",
			Category:     "settings",
		},
		{
			ID:           "test-2",
			Timestamp:    time.Now().Add(-1 * time.Hour),
			User:         "user2",
			Organization: "org1",
			Repository:   "org1/repo2",
			Operation:    "update",
			Category:     "branch_protection",
		},
		{
			ID:           "test-3",
			Timestamp:    time.Now(),
			User:         "user1",
			Organization: "org2",
			Repository:   "org2/repo1",
			Operation:    "delete",
			Category:     "permissions",
		},
	}

	// Store all records
	for _, record := range records {
		err = store.Store(ctx, record)
		require.NoError(t, err)
	}

	// Test filtering by organization
	filter := ChangeFilter{Organization: "org1"}
	results, err := store.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Test filtering by user
	filter = ChangeFilter{User: "user1"}
	results, err = store.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Test filtering by operation
	filter = ChangeFilter{Operation: "update"}
	results, err = store.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "test-2", results[0].ID)

	// Test filtering by category
	filter = ChangeFilter{Category: "settings"}
	results, err = store.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "test-1", results[0].ID)

	// Test time-based filtering
	filter = ChangeFilter{
		Since: time.Now().Add(-90 * time.Minute),
		Until: time.Now().Add(-30 * time.Minute),
	}
	results, err = store.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "test-2", results[0].ID)

	// Test limit and offset
	filter = ChangeFilter{Limit: 1}
	results, err = store.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	// Should be the newest record (test-3)
	assert.Equal(t, "test-3", results[0].ID)

	filter = ChangeFilter{Limit: 1, Offset: 1}
	results, err = store.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	// Should be the second newest record (test-2)
	assert.Equal(t, "test-2", results[0].ID)
}

func TestFileStoreStats(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "changelog_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := NewFileStore(tempDir)
	require.NoError(t, err)

	ctx := context.Background()

	// Initial stats should show empty store
	stats, err := store.GetStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, stats["total_records"])
	assert.Equal(t, int64(0), stats["total_size_bytes"])

	// Add a record
	record := &ChangeRecord{
		ID:          "test-stats",
		Timestamp:   time.Now(),
		User:        "testuser",
		Operation:   "test",
		Category:    "test",
		Description: "Test record for stats",
	}

	err = store.Store(ctx, record)
	require.NoError(t, err)

	// Stats should now show one record
	stats, err = store.GetStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, stats["total_records"])
	assert.Greater(t, stats["total_size_bytes"].(int64), int64(0))
	assert.Equal(t, tempDir, stats["storage_path"])
}

func TestRollbackRequest(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "changelog_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := NewFileStore(tempDir)
	require.NoError(t, err)

	changelog := NewChangeLog(nil, store)
	ctx := context.Background()

	// Create a change record with before/after state
	record := &ChangeRecord{
		ID:           "rollback-test",
		Timestamp:    time.Now(),
		User:         "testuser",
		Organization: "testorg",
		Repository:   "testorg/testrepo",
		Operation:    "update",
		Category:     "settings",
		Before:       map[string]interface{}{"private": false, "description": "Old description"},
		After:        map[string]interface{}{"private": true, "description": "New description"},
		Description:  "Test change for rollback",
		Source:       "cli",
	}

	err = changelog.RecordChange(ctx, record)
	require.NoError(t, err)

	// Test dry run rollback
	rollbackReq := &RollbackRequest{
		ChangeID:    record.ID,
		Repository:  "testorg/testrepo",
		Category:    "settings",
		DryRun:      true,
		Description: "Test rollback",
	}

	result, err := changelog.Rollback(ctx, rollbackReq)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.True(t, result.DryRun)
	assert.Empty(t, result.NewChangeID)

	// Test repository mismatch
	rollbackReq.Repository = "different/repo"
	rollbackReq.DryRun = false
	result, err = changelog.Rollback(ctx, rollbackReq)
	assert.Error(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Errors, "Repository mismatch in rollback request")

	// Test category mismatch
	rollbackReq.Repository = "testorg/testrepo"
	rollbackReq.Category = "permissions"
	result, err = changelog.Rollback(ctx, rollbackReq)
	assert.Error(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Errors, "Category mismatch in rollback request")
}

func TestGenerateChangeID(t *testing.T) {
	id1 := generateChangeID()
	id2 := generateChangeID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "ch_")
	assert.Contains(t, id2, "ch_")
}
