//nolint:testpackage // White-box testing needed for internal function access
package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBulkCloneResult_GetSummary(t *testing.T) {
	result := &BulkCloneResult{
		TotalTargets:      10,
		SuccessfulTargets: 7,
		FailedTargets:     2,
		SkippedTargets:    1,
	}

	expected := "Total: 10, Successful: 7, Failed: 2, Skipped: 1"
	assert.Equal(t, expected, result.GetSummary())
}

func TestTargetResult(t *testing.T) {
	result := TargetResult{
		Provider: ProviderGitHub,
		Name:     "test-org",
		CloneDir: "/path/to/clone",
		Strategy: StrategyReset,
		Success:  true,
	}

	assert.Equal(t, ProviderGitHub, result.Provider)
	assert.Equal(t, "test-org", result.Name)
	assert.Equal(t, "/path/to/clone", result.CloneDir)
	assert.Equal(t, StrategyReset, result.Strategy)
	assert.True(t, result.Success)
	assert.Empty(t, result.Error)
}
