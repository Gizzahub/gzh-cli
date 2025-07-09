package audit

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileBasedAuditStore(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "audit-store-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store, err := NewFileBasedAuditStore(tmpDir)
	require.NoError(t, err)

	t.Run("SaveAndRetrieveAuditResult", func(t *testing.T) {
		// Create test audit history
		history := &AuditHistory{
			Timestamp:    time.Now(),
			Organization: "test-org",
			Summary: AuditSummary{
				TotalRepositories:     10,
				CompliantRepositories: 8,
				CompliancePercentage:  80.0,
				TotalViolations:       5,
				CriticalViolations:    2,
			},
			PolicyStats: map[string]PolicyStatistics{
				"security": {
					PolicyName:           "security",
					ViolationCount:       3,
					CompliantRepos:       7,
					ViolatingRepos:       3,
					CompliancePercentage: 70.0,
				},
			},
		}

		// Save the audit result
		err := store.SaveAuditResult(history)
		assert.NoError(t, err)

		// Retrieve historical data
		results, err := store.GetHistoricalData("test-org", 24*time.Hour)
		assert.NoError(t, err)
		assert.Len(t, results, 1)

		// Verify the retrieved data
		assert.Equal(t, "test-org", results[0].Organization)
		assert.Equal(t, 10, results[0].Summary.TotalRepositories)
		assert.Equal(t, 80.0, results[0].Summary.CompliancePercentage)
	})

	t.Run("MultipleSavesOnSameDay", func(t *testing.T) {
		org := "multi-save-org"

		// Save multiple audit results on the same day
		for i := 0; i < 3; i++ {
			history := &AuditHistory{
				Timestamp:    time.Now().Add(time.Duration(i) * time.Hour),
				Organization: org,
				Summary: AuditSummary{
					TotalRepositories:     10 + i,
					CompliantRepositories: 8 + i,
					CompliancePercentage:  80.0,
					TotalViolations:       5 - i,
					CriticalViolations:    2,
				},
				PolicyStats: map[string]PolicyStatistics{},
			}
			err := store.SaveAuditResult(history)
			assert.NoError(t, err)
		}

		// Retrieve and verify we get all saves
		results, err := store.GetHistoricalData(org, 24*time.Hour)
		assert.NoError(t, err)
		assert.Len(t, results, 3)
	})

	t.Run("GetPolicyTrends", func(t *testing.T) {
		org := "policy-trend-org"
		policyName := "branch-protection"

		// Save audit results with policy stats
		for i := 0; i < 5; i++ {
			history := &AuditHistory{
				Timestamp:    time.Now().Add(time.Duration(-i) * 24 * time.Hour),
				Organization: org,
				Summary:      AuditSummary{},
				PolicyStats: map[string]PolicyStatistics{
					policyName: {
						PolicyName:           policyName,
						ViolationCount:       10 - i,
						CompliantRepos:       5 + i,
						ViolatingRepos:       5 - i,
						CompliancePercentage: float64(50 + i*10),
					},
				},
			}
			err := store.SaveAuditResult(history)
			assert.NoError(t, err)
		}

		// Get policy trends
		trends, err := store.GetPolicyTrends(org, policyName, 7*24*time.Hour)
		assert.NoError(t, err)
		assert.Len(t, trends, 5)

		// Verify trend data
		for _, trend := range trends {
			assert.Equal(t, policyName, trend.PolicyName)
		}
	})

	t.Run("NoDataForOrganization", func(t *testing.T) {
		results, err := store.GetHistoricalData("non-existent-org", 24*time.Hour)
		assert.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("FilePathValidation", func(t *testing.T) {
		// Test with organization name containing special characters
		history := &AuditHistory{
			Timestamp:    time.Now(),
			Organization: "test/org/../etc",
			Summary:      AuditSummary{},
			PolicyStats:  map[string]PolicyStatistics{},
		}

		err := store.SaveAuditResult(history)
		assert.NoError(t, err)

		// Verify the file is created with sanitized path
		orgPath := filepath.Join(tmpDir, "test/org/../etc")
		_, err = os.Stat(orgPath)
		assert.NoError(t, err)
	})
}

func TestNewFileBasedAuditStore_DefaultPath(t *testing.T) {
	// Test with empty base path (uses default)
	store, err := NewFileBasedAuditStore("")
	assert.NoError(t, err)
	assert.NotEmpty(t, store.basePath)
	assert.Contains(t, store.basePath, "gzh-manager")
	assert.Contains(t, store.basePath, "audit-history")
}
