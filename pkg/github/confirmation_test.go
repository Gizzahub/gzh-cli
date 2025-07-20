//nolint:testpackage // White-box testing needed for internal function access
package github

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfirmationPrompt(t *testing.T) {
	prompt := NewConfirmationPrompt()
	assert.NotNil(t, prompt)
	assert.False(t, prompt.AutoConfirm)
	assert.NotNil(t, prompt.inputReader)
}

func TestNewAutoConfirmationPrompt(t *testing.T) {
	prompt := NewAutoConfirmationPrompt()
	assert.NotNil(t, prompt)
	assert.True(t, prompt.AutoConfirm)
}

func TestRequestConfirmation_AutoConfirm(t *testing.T) {
	prompt := NewAutoConfirmationPrompt()
	ctx := context.Background()

	request := &ConfirmationRequest{
		Changes: []SensitiveChange{
			{
				Repository:  "test/repo",
				Category:    "settings",
				Operation:   "update",
				Field:       "private",
				OldValue:    false,
				NewValue:    true,
				Risk:        RiskHigh,
				Description: "Make repository private",
			},
		},
		Operation: "bulk_update",
		Target:    "testorg",
	}

	result, err := prompt.RequestConfirmation(ctx, request)
	require.NoError(t, err)
	assert.True(t, result.Confirmed)
	assert.Equal(t, "auto", result.UserChoice)
	assert.Equal(t, "Auto-confirmation enabled", result.Reason)
}

func TestRequestConfirmation_DryRun(t *testing.T) {
	prompt := NewConfirmationPrompt()
	ctx := context.Background()

	request := &ConfirmationRequest{
		Changes: []SensitiveChange{
			{
				Repository:  "test/repo",
				Category:    "settings",
				Operation:   "update",
				Field:       "private",
				OldValue:    false,
				NewValue:    true,
				Risk:        RiskHigh,
				Description: "Make repository private",
			},
		},
		Operation: "bulk_update",
		Target:    "testorg",
		DryRun:    true,
	}

	result, err := prompt.RequestConfirmation(ctx, request)
	require.NoError(t, err)
	assert.True(t, result.Confirmed)
	assert.Equal(t, "auto", result.UserChoice)
	assert.Equal(t, "Dry run mode - no actual changes will be made", result.Reason)
}

func TestRequestConfirmation_UserInput(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectConfirmed bool
		expectChoice    string
		expectError     bool
	}{
		{
			name:            "yes confirmation",
			input:           "y",
			expectConfirmed: true,
			expectChoice:    "y",
		},
		{
			name:            "yes full word",
			input:           "yes",
			expectConfirmed: true,
			expectChoice:    "yes",
		},
		{
			name:            "no confirmation",
			input:           "n",
			expectConfirmed: false,
			expectChoice:    "n",
		},
		{
			name:            "no full word",
			input:           "no",
			expectConfirmed: false,
			expectChoice:    "no",
		},
		{
			name:            "abort operation",
			input:           "a",
			expectConfirmed: false,
			expectChoice:    "a",
			expectError:     true,
		},
		{
			name:            "invalid choice",
			input:           "invalid",
			expectConfirmed: false,
			expectChoice:    "invalid",
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := NewConfirmationPrompt()
			prompt.inputReader = bufio.NewScanner(strings.NewReader(tt.input))

			ctx := context.Background()
			request := &ConfirmationRequest{
				Changes: []SensitiveChange{
					{
						Repository:  "test/repo",
						Category:    "settings",
						Operation:   "update",
						Field:       "private",
						OldValue:    false,
						NewValue:    true,
						Risk:        RiskMedium,
						Description: "Make repository private",
					},
				},
				Operation: "bulk_update",
				Target:    "testorg",
			}

			result, err := prompt.RequestConfirmation(ctx, request)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectConfirmed, result.Confirmed)
			assert.Equal(t, tt.expectChoice, result.UserChoice)
		})
	}
}

func TestRequestConfirmation_SkipHighRisk(t *testing.T) {
	prompt := NewConfirmationPrompt()
	prompt.inputReader = bufio.NewScanner(strings.NewReader("s"))

	ctx := context.Background()
	request := &ConfirmationRequest{
		Changes: []SensitiveChange{
			{
				Repository:  "test/repo",
				Category:    "settings",
				Operation:   "update",
				Field:       "private",
				OldValue:    false,
				NewValue:    true,
				Risk:        RiskHigh,
				Description: "Make repository private",
			},
			{
				Repository:  "test/repo",
				Category:    "settings",
				Operation:   "update",
				Field:       "archived",
				OldValue:    false,
				NewValue:    true,
				Risk:        RiskCritical,
				Description: "Archive repository",
			},
		},
		Operation: "bulk_update",
		Target:    "testorg",
	}

	result, err := prompt.RequestConfirmation(ctx, request)
	require.NoError(t, err)
	assert.True(t, result.Confirmed)
	assert.Equal(t, "s", result.UserChoice)
	assert.Contains(t, result.SkippedRisks, RiskHigh)
	assert.Contains(t, result.SkippedRisks, RiskCritical)
}

func TestAnalyzeRepositoryChanges(t *testing.T) {
	prompt := NewConfirmationPrompt()
	ctx := context.Background()

	before := &RepositoryConfig{
		Private:  false,
		Archived: false,
		Settings: RepoConfigSettings{
			DefaultBranch: "main",
		},
		BranchProtection: map[string]BranchProtectionConfig{
			"main": {
				RequiredReviews: 2,
				EnforceAdmins:   true,
			},
		},
		Permissions: PermissionsConfig{
			Teams: map[string]string{"dev": "write"},
			Users: map[string]string{"user1": "read"},
		},
	}

	after := &RepositoryConfig{
		Private:  true, // Changed to private - HIGH RISK
		Archived: true, // Changed to archived - HIGH RISK
		Settings: RepoConfigSettings{
			DefaultBranch: "master", // Changed default branch - HIGH RISK
		},
		BranchProtection: map[string]BranchProtectionConfig{
			"main": {
				RequiredReviews: 1,     // Decreased reviews - MEDIUM RISK
				EnforceAdmins:   false, // Disabled admin enforcement - HIGH RISK
			},
		},
		Permissions: PermissionsConfig{
			Teams: map[string]string{"dev": "admin"},                    // Permission escalation - MEDIUM RISK
			Users: map[string]string{"user1": "write", "user2": "read"}, // New user + escalation - LOW/MEDIUM RISK
		},
	}

	changes := prompt.AnalyzeRepositoryChanges(ctx, "testorg", "testrepo", before, after)

	// Should detect multiple sensitive changes
	assert.Greater(t, len(changes), 5)

	// Check for specific critical/high risk changes
	foundPrivacyChange := false
	foundArchiveChange := false
	foundBranchChange := false
	foundAdminEnforcement := false

	for _, change := range changes {
		switch change.Field {
		case "private":
			foundPrivacyChange = true

			assert.Equal(t, RiskCritical, change.Risk)
		case "archived":
			foundArchiveChange = true

			assert.Equal(t, RiskHigh, change.Risk)
		case "default_branch":
			foundBranchChange = true

			assert.Equal(t, RiskHigh, change.Risk)
		case "enforce_admins":
			foundAdminEnforcement = true

			assert.Equal(t, RiskHigh, change.Risk)
		}
	}

	assert.True(t, foundPrivacyChange, "Should detect privacy change")
	assert.True(t, foundArchiveChange, "Should detect archive change")
	assert.True(t, foundBranchChange, "Should detect default branch change")
	assert.True(t, foundAdminEnforcement, "Should detect admin enforcement change")
}

func TestAnalyzeBranchProtectionChanges(t *testing.T) {
	prompt := NewConfirmationPrompt()

	before := map[string]BranchProtectionConfig{
		"main": {
			RequiredReviews: 2,
			EnforceAdmins:   true,
		},
		"develop": {
			RequiredReviews: 1,
			EnforceAdmins:   false,
		},
	}

	after := map[string]BranchProtectionConfig{
		"main": {
			RequiredReviews: 1,     // Decreased
			EnforceAdmins:   false, // Disabled
		},
		// develop branch protection removed
	}

	changes := prompt.analyzeBranchProtectionChanges("test/repo", before, after)

	assert.Len(t, changes, 3) // Two changes to main + one removal of develop

	// Check for removal of develop branch protection
	foundRemoval := false
	foundReviewDecrease := false
	foundAdminDisable := false

	for _, change := range changes {
		if change.Operation == "delete" && strings.Contains(change.Description, "develop") {
			foundRemoval = true

			assert.Equal(t, RiskHigh, change.Risk)
		}

		if change.Field == "required_reviews" {
			foundReviewDecrease = true

			assert.Equal(t, RiskMedium, change.Risk)
		}

		if change.Field == "enforce_admins" {
			foundAdminDisable = true

			assert.Equal(t, RiskHigh, change.Risk)
		}
	}

	assert.True(t, foundRemoval, "Should detect branch protection removal")
	assert.True(t, foundReviewDecrease, "Should detect review requirement decrease")
	assert.True(t, foundAdminDisable, "Should detect admin enforcement disable")
}

func TestAnalyzePermissionChanges(t *testing.T) {
	prompt := NewConfirmationPrompt()

	before := PermissionsConfig{
		Teams: map[string]string{"dev": "write", "qa": "read"},
		Users: map[string]string{"user1": "read"},
	}

	after := PermissionsConfig{
		Teams: map[string]string{"dev": "admin", "qa": "read", "ops": "write"}, // dev escalated, ops added
		Users: map[string]string{"user1": "write", "user2": "read"},            // user1 escalated, user2 added
	}

	changes := prompt.analyzePermissionChanges("test/repo", before, after)

	assert.Len(t, changes, 4) // 2 escalations + 2 new permissions

	foundTeamEscalation := false
	foundUserEscalation := false
	foundNewTeam := false
	foundNewUser := false

	for _, change := range changes {
		if change.Field == "team_permission" && change.Operation == "update" && strings.Contains(change.Description, "dev") {
			foundTeamEscalation = true

			assert.Equal(t, RiskMedium, change.Risk)
		}

		if change.Field == "user_permission" && change.Operation == "update" && strings.Contains(change.Description, "user1") {
			foundUserEscalation = true

			assert.Equal(t, RiskMedium, change.Risk)
		}

		if change.Field == "team_permission" && change.Operation == "create" && strings.Contains(change.Description, "ops") {
			foundNewTeam = true

			assert.Equal(t, RiskLow, change.Risk)
		}

		if change.Field == "user_permission" && change.Operation == "create" && strings.Contains(change.Description, "user2") {
			foundNewUser = true

			assert.Equal(t, RiskLow, change.Risk)
		}
	}

	assert.True(t, foundTeamEscalation, "Should detect team permission escalation")
	assert.True(t, foundUserEscalation, "Should detect user permission escalation")
	assert.True(t, foundNewTeam, "Should detect new team permission")
	assert.True(t, foundNewUser, "Should detect new user permission")
}

func TestIsPermissionEscalation(t *testing.T) {
	prompt := NewConfirmationPrompt()

	tests := []struct {
		before   string
		after    string
		expected bool
	}{
		{"read", "write", true},
		{"write", "admin", true},
		{"read", "admin", true},
		{"admin", "write", false},
		{"write", "read", false},
		{"read", "read", false},
		{"unknown", "write", false},
		{"read", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_to_%s", tt.before, tt.after), func(t *testing.T) {
			result := prompt.isPermissionEscalation(tt.before, tt.after)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCategorizeChangesByRisk(t *testing.T) {
	prompt := NewConfirmationPrompt()

	changes := []SensitiveChange{
		{Risk: RiskCritical, Description: "Critical change"},
		{Risk: RiskHigh, Description: "High change 1"},
		{Risk: RiskHigh, Description: "High change 2"},
		{Risk: RiskMedium, Description: "Medium change"},
		{Risk: RiskLow, Description: "Low change 1"},
		{Risk: RiskLow, Description: "Low change 2"},
		{Risk: RiskLow, Description: "Low change 3"},
	}

	categories := prompt.categorizeChangesByRisk(changes)

	assert.Len(t, categories[RiskCritical], 1)
	assert.Len(t, categories[RiskHigh], 2)
	assert.Len(t, categories[RiskMedium], 1)
	assert.Len(t, categories[RiskLow], 3)
}

func TestGetRiskColor(t *testing.T) {
	prompt := NewConfirmationPrompt()

	assert.Equal(t, "\033[1;91m", prompt.getRiskColor(RiskCritical))
	assert.Equal(t, "\033[1;31m", prompt.getRiskColor(RiskHigh))
	assert.Equal(t, "\033[1;33m", prompt.getRiskColor(RiskMedium))
	assert.Equal(t, "\033[1;32m", prompt.getRiskColor(RiskLow))
	assert.Equal(t, "", prompt.getRiskColor("unknown"))
}

func TestGetRiskIcon(t *testing.T) {
	prompt := NewConfirmationPrompt()

	assert.Equal(t, "🚨", prompt.getRiskIcon(RiskCritical))
	assert.Equal(t, "⚠️", prompt.getRiskIcon(RiskHigh))
	assert.Equal(t, "⚡", prompt.getRiskIcon(RiskMedium))
	assert.Equal(t, "ℹ️", prompt.getRiskIcon(RiskLow))
	assert.Equal(t, "•", prompt.getRiskIcon("unknown"))
}
