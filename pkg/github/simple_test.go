package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRateLimit_Struct(t *testing.T) {
	// Test RateLimit struct creation
	rl := &RateLimit{
		Limit:     5000,
		Remaining: 4500,
		Used:      500,
	}

	assert.Equal(t, 5000, rl.Limit)
	assert.Equal(t, 4500, rl.Remaining)
	assert.Equal(t, 500, rl.Used)
}

func TestRepositoryInfo_Struct(t *testing.T) {
	// Test RepositoryInfo struct creation
	repo := &RepositoryInfo{
		Name:          "test-repo",
		FullName:      "test-org/test-repo",
		Description:   "Test repository",
		DefaultBranch: "main",
		CloneURL:      "https://github.com/test-org/test-repo.git",
		SSHURL:        "git@github.com:test-org/test-repo.git",
		HTMLURL:       "https://github.com/test-org/test-repo",
		Private:       true,
		Archived:      false,
		Language:      "Go",
		Size:          1024,
	}

	assert.Equal(t, "test-repo", repo.Name)
	assert.Equal(t, "test-org/test-repo", repo.FullName)
	assert.Equal(t, "Test repository", repo.Description)
	assert.Equal(t, "main", repo.DefaultBranch)
	assert.True(t, repo.Private)
	assert.False(t, repo.Archived)
}

func TestTokenInfoRecord_Struct(t *testing.T) {
	// Test TokenInfoRecord struct creation
	tokenInfo := &TokenInfoRecord{
		Valid:  true,
		Scopes: []string{"repo", "user", "admin:org"},
		RateLimit: RateLimit{
			Limit:     5000,
			Remaining: 4999,
		},
		User:        "testuser",
		Permissions: []string{"read", "write", "admin"},
	}

	assert.True(t, tokenInfo.Valid)
	assert.Equal(t, "testuser", tokenInfo.User)
	assert.Contains(t, tokenInfo.Scopes, "repo")
	assert.Contains(t, tokenInfo.Permissions, "write")
}

func TestLogLevelType_Constants(t *testing.T) {
	// Test LogLevelType constants
	assert.Equal(t, LogLevelType(0), LogLevelTypeDebug)
	assert.Equal(t, LogLevelType(1), LogLevelTypeInfo)
	assert.Equal(t, LogLevelType(2), LogLevelTypeWarn)
	assert.Equal(t, LogLevelType(3), LogLevelTypeError)
}

func TestRiskLevelType_Constants(t *testing.T) {
	// Test RiskLevelType constants
	assert.Equal(t, RiskLevelType(0), RiskLevelLow)
	assert.Equal(t, RiskLevelType(1), RiskLevelMedium)
	assert.Equal(t, RiskLevelType(2), RiskLevelHigh)
	assert.Equal(t, RiskLevelType(3), RiskLevelCritical)
}

func TestConfirmationModeType_Constants(t *testing.T) {
	// Test ConfirmationModeType constants
	assert.Equal(t, ConfirmationModeType(0), ConfirmationModeInteractive)
	assert.Equal(t, ConfirmationModeType(1), ConfirmationModeAutoApprove)
	assert.Equal(t, ConfirmationModeType(2), ConfirmationModeAutoDeny)
	assert.Equal(t, ConfirmationModeType(3), ConfirmationModeDryRun)
}

func TestLogOperationRecord_Struct(t *testing.T) {
	// Test LogOperationRecord struct
	logOp := &LogOperationRecord{
		ID:         "test-123",
		Operation:  "clone",
		Repository: "test-repo",
		User:       "testuser",
		Success:    true,
		Metadata: map[string]interface{}{
			"strategy": "reset",
			"duration": 5.2,
		},
	}

	assert.Equal(t, "test-123", logOp.ID)
	assert.Equal(t, "clone", logOp.Operation)
	assert.True(t, logOp.Success)
	assert.Equal(t, "reset", logOp.Metadata["strategy"])
}

func TestConfirmationPromptRecord_Struct(t *testing.T) {
	// Test ConfirmationPromptRecord struct
	prompt := &ConfirmationPromptRecord{
		Title:       "Confirm Repository Update",
		Description: "This will update repository settings",
		Repository:  "test-repo",
		Operation:   "update",
		Risk:        RiskLevelMedium,
		Impact:      "Repository settings will be changed",
		Metadata: map[string]interface{}{
			"changes": 3,
		},
	}

	assert.Equal(t, "Confirm Repository Update", prompt.Title)
	assert.Equal(t, RiskLevelMedium, prompt.Risk)
	assert.Equal(t, 3, prompt.Metadata["changes"])
}

