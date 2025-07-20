//nolint:tagliatelle // GitHub API response formatting may require specific JSON field naming conventions
package github

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

// ConfirmationPrompt handles user confirmation for sensitive operations.
type ConfirmationPrompt struct {
	// AutoConfirm bypasses prompts when true (useful for automation)
	AutoConfirm bool
	// Input reader for testing
	inputReader *bufio.Scanner
}

// SensitiveChange represents a potentially sensitive configuration change.
type SensitiveChange struct {
	Repository  string      `json:"repository"`
	Category    string      `json:"category"`  // settings, branch_protection, permissions
	Operation   string      `json:"operation"` // update, create, delete
	Field       string      `json:"field"`     // specific field being changed
	OldValue    interface{} `json:"old_value"`
	NewValue    interface{} `json:"new_value"`
	Risk        RiskLevel   `json:"risk"`
	Description string      `json:"description"`
	Impact      string      `json:"impact"`
}

// RiskLevel represents the risk level of a change.
type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

// ConfirmationRequest contains details for a confirmation request.
type ConfirmationRequest struct {
	Changes     []SensitiveChange `json:"changes"`
	Operation   string            `json:"operation"` // bulk_update, rollback, etc.
	Target      string            `json:"target"`    // organization or repository name
	DryRun      bool              `json:"dry_run"`
	BatchSize   int               `json:"batch_size"` // number of repositories affected
	Description string            `json:"description"`
}

// ConfirmationResult contains the result of a confirmation request.
type ConfirmationResult struct {
	Confirmed    bool        `json:"confirmed"`
	UserChoice   string      `json:"user_choice"` // yes, no, skip, abort
	SkippedRisks []RiskLevel `json:"skipped_risks,omitempty"`
	Reason       string      `json:"reason,omitempty"`
}

// NewConfirmationPrompt creates a new confirmation prompt handler.
func NewConfirmationPrompt() *ConfirmationPrompt {
	return &ConfirmationPrompt{
		inputReader: bufio.NewScanner(os.Stdin),
	}
}

// NewAutoConfirmationPrompt creates a confirmation prompt that auto-confirms all prompts.
func NewAutoConfirmationPrompt() *ConfirmationPrompt {
	return &ConfirmationPrompt{
		AutoConfirm: true,
	}
}

// RequestConfirmation requests user confirmation for sensitive changes.
func (cp *ConfirmationPrompt) RequestConfirmation(ctx context.Context, request *ConfirmationRequest) (*ConfirmationResult, error) {
	result := &ConfirmationResult{}

	// Auto-confirm if enabled
	if cp.AutoConfirm || request.DryRun {
		result.Confirmed = true

		result.UserChoice = "auto"
		if request.DryRun {
			result.Reason = "Dry run mode - no actual changes will be made"
		} else {
			result.Reason = "Auto-confirmation enabled"
		}

		return result, nil
	}

	// Display operation summary
	fmt.Printf("\n🔧 Repository Configuration Change Request\n")
	fmt.Printf("Operation: %s\n", request.Operation)
	fmt.Printf("Target: %s\n", request.Target)

	if request.BatchSize > 1 {
		fmt.Printf("Repositories affected: %d\n", request.BatchSize)
	}

	if request.Description != "" {
		fmt.Printf("Description: %s\n", request.Description)
	}

	// Categorize changes by risk level
	riskCategories := cp.categorizeChangesByRisk(request.Changes)

	// Display changes by risk level
	for _, risk := range []RiskLevel{RiskCritical, RiskHigh, RiskMedium, RiskLow} {
		if changes, exists := riskCategories[risk]; exists && len(changes) > 0 {
			cp.displayRiskCategory(risk, changes)
		}
	}

	// Get user confirmation
	return cp.promptForConfirmation(request, riskCategories)
}

// categorizeChangesByRisk groups changes by their risk level.
func (cp *ConfirmationPrompt) categorizeChangesByRisk(changes []SensitiveChange) map[RiskLevel][]SensitiveChange {
	categories := make(map[RiskLevel][]SensitiveChange)

	for _, change := range changes {
		categories[change.Risk] = append(categories[change.Risk], change)
	}

	return categories
}

// displayRiskCategory displays changes for a specific risk level.
func (cp *ConfirmationPrompt) displayRiskCategory(risk RiskLevel, changes []SensitiveChange) {
	riskColor := cp.getRiskColor(risk)
	riskIcon := cp.getRiskIcon(risk)

	fmt.Printf("\n%s %s%s Risk Changes:%s\n", riskIcon, riskColor, strings.ToUpper(string(risk)), cp.getResetColor())

	for i, change := range changes {
		fmt.Printf("  %d. %s (%s)\n", i+1, change.Description, change.Repository)

		if change.Field != "" {
			fmt.Printf("     Field: %s\n", change.Field)
			fmt.Printf("     Change: %v → %v\n", change.OldValue, change.NewValue)
		}

		if change.Impact != "" {
			fmt.Printf("     Impact: %s\n", change.Impact)
		}
	}
}

// promptForConfirmation handles the interactive confirmation process.
func (cp *ConfirmationPrompt) promptForConfirmation(request *ConfirmationRequest, riskCategories map[RiskLevel][]SensitiveChange) (*ConfirmationResult, error) {
	result := &ConfirmationResult{}

	// Check for critical or high-risk changes
	hasCritical := len(riskCategories[RiskCritical]) > 0
	hasHigh := len(riskCategories[RiskHigh]) > 0

	fmt.Printf("\n")

	if hasCritical {
		fmt.Printf("⚠️  CRITICAL risk changes detected! These changes may have severe impact.\n")
	} else if hasHigh {
		fmt.Printf("⚠️  HIGH risk changes detected! Please review carefully.\n")
	}

	// Provide options
	fmt.Printf("\nOptions:\n")
	fmt.Printf("  [y]es - Proceed with all changes\n")
	fmt.Printf("  [n]o  - Cancel operation\n")

	if hasCritical || hasHigh {
		fmt.Printf("  [s]kip - Skip high/critical risk changes only\n")
	}

	fmt.Printf("  [a]bort - Abort and exit\n")
	fmt.Printf("\nYour choice: ")

	// Read user input
	if !cp.inputReader.Scan() {
		return result, fmt.Errorf("failed to read user input")
	}

	choice := strings.ToLower(strings.TrimSpace(cp.inputReader.Text()))
	result.UserChoice = choice

	switch choice {
	case "y", "yes":
		result.Confirmed = true
		result.Reason = "User confirmed all changes"

	case "n", "no":
		result.Confirmed = false
		result.Reason = "User declined changes"

	case "s", "skip":
		if hasCritical || hasHigh {
			result.Confirmed = true

			result.SkippedRisks = []RiskLevel{}
			if hasCritical {
				result.SkippedRisks = append(result.SkippedRisks, RiskCritical)
			}

			if hasHigh {
				result.SkippedRisks = append(result.SkippedRisks, RiskHigh)
			}

			result.Reason = "User chose to skip high/critical risk changes"
		} else {
			result.Confirmed = false
			result.Reason = "No high/critical risk changes to skip"
		}

	case "a", "abort":
		result.Confirmed = false
		result.Reason = "User aborted operation"

		return result, fmt.Errorf("operation aborted by user")

	default:
		result.Confirmed = false
		result.Reason = fmt.Sprintf("Invalid choice: %s", choice)

		return result, fmt.Errorf("invalid choice: %s", choice)
	}

	return result, nil
}

// AnalyzeRepositoryChanges analyzes repository configuration changes for sensitivity.
func (cp *ConfirmationPrompt) AnalyzeRepositoryChanges(ctx context.Context, owner, repo string, before, after *RepositoryConfig) []SensitiveChange {
	var changes []SensitiveChange

	repoName := fmt.Sprintf("%s/%s", owner, repo)

	// Check visibility changes
	if before.Private != after.Private {
		var risk RiskLevel
		var impact string

		if !before.Private && after.Private {
			impact = "Repository will become private - public access will be lost"
			risk = RiskCritical
		} else {
			impact = "Repository will become public - private content will be exposed"
			risk = RiskCritical
		}

		changes = append(changes, SensitiveChange{
			Repository:  repoName,
			Category:    "settings",
			Operation:   "update",
			Field:       "private",
			OldValue:    before.Private,
			NewValue:    after.Private,
			Risk:        risk,
			Description: "Repository visibility change",
			Impact:      impact,
		})
	}

	// Check archive status changes
	if before.Archived != after.Archived {
		var risk RiskLevel
		var impact string

		if after.Archived {
			impact = "Repository will be archived - no further commits allowed"
			risk = RiskHigh
		} else {
			impact = "Repository will be unarchived - write access restored"
			risk = RiskMedium
		}

		changes = append(changes, SensitiveChange{
			Repository:  repoName,
			Category:    "settings",
			Operation:   "update",
			Field:       "archived",
			OldValue:    before.Archived,
			NewValue:    after.Archived,
			Risk:        risk,
			Description: "Repository archive status change",
			Impact:      impact,
		})
	}

	// Check default branch changes
	if before.Settings.DefaultBranch != after.Settings.DefaultBranch &&
		before.Settings.DefaultBranch != "" && after.Settings.DefaultBranch != "" {
		changes = append(changes, SensitiveChange{
			Repository:  repoName,
			Category:    "settings",
			Operation:   "update",
			Field:       "default_branch",
			OldValue:    before.Settings.DefaultBranch,
			NewValue:    after.Settings.DefaultBranch,
			Risk:        RiskHigh,
			Description: "Default branch change",
			Impact:      "May affect CI/CD pipelines and developer workflows",
		})
	}

	// Check branch protection changes
	changes = append(changes, cp.analyzeBranchProtectionChanges(repoName, before.BranchProtection, after.BranchProtection)...)

	// Check permission changes
	changes = append(changes, cp.analyzePermissionChanges(repoName, before.Permissions, after.Permissions)...)

	return changes
}

// analyzeBranchProtectionChanges analyzes branch protection rule changes.
func (cp *ConfirmationPrompt) analyzeBranchProtectionChanges(repoName string, before, after map[string]BranchProtectionConfig) []SensitiveChange {
	var changes []SensitiveChange

	// Check for removed branch protection
	for branch := range before {
		if _, exists := after[branch]; !exists {
			changes = append(changes, SensitiveChange{
				Repository:  repoName,
				Category:    "branch_protection",
				Operation:   "delete",
				Field:       "branch_protection",
				OldValue:    fmt.Sprintf("Protected: %s", branch),
				NewValue:    "Unprotected",
				Risk:        RiskHigh,
				Description: fmt.Sprintf("Branch protection removed for %s", branch),
				Impact:      "Branch will allow direct pushes and forced updates",
			})
		}
	}

	// Check for modified branch protection
	for branch, afterConfig := range after {
		if beforeConfig, exists := before[branch]; exists {
			// Check if required reviews decreased
			if beforeConfig.RequiredReviews > afterConfig.RequiredReviews {
				changes = append(changes, SensitiveChange{
					Repository:  repoName,
					Category:    "branch_protection",
					Operation:   "update",
					Field:       "required_reviews",
					OldValue:    beforeConfig.RequiredReviews,
					NewValue:    afterConfig.RequiredReviews,
					Risk:        RiskMedium,
					Description: fmt.Sprintf("Required reviews decreased for %s", branch),
					Impact:      "Lower review requirements may reduce code quality oversight",
				})
			}

			// Check if admin enforcement is disabled
			if beforeConfig.EnforceAdmins && !afterConfig.EnforceAdmins {
				changes = append(changes, SensitiveChange{
					Repository:  repoName,
					Category:    "branch_protection",
					Operation:   "update",
					Field:       "enforce_admins",
					OldValue:    true,
					NewValue:    false,
					Risk:        RiskHigh,
					Description: fmt.Sprintf("Admin enforcement disabled for %s", branch),
					Impact:      "Administrators can now bypass branch protection rules",
				})
			}
		}
	}

	return changes
}

// analyzePermissionChanges analyzes permission changes.
func (cp *ConfirmationPrompt) analyzePermissionChanges(repoName string, before, after PermissionsConfig) []SensitiveChange {
	var changes []SensitiveChange

	// Check team permission changes
	for team, afterPerm := range after.Teams {
		if beforePerm, exists := before.Teams[team]; exists {
			if cp.isPermissionEscalation(beforePerm, afterPerm) {
				changes = append(changes, SensitiveChange{
					Repository:  repoName,
					Category:    "permissions",
					Operation:   "update",
					Field:       "team_permission",
					OldValue:    fmt.Sprintf("%s: %s", team, beforePerm),
					NewValue:    fmt.Sprintf("%s: %s", team, afterPerm),
					Risk:        RiskMedium,
					Description: fmt.Sprintf("Team permission escalation: %s", team),
					Impact:      "Team will have increased access to repository",
				})
			}
		} else {
			// New team permission
			changes = append(changes, SensitiveChange{
				Repository:  repoName,
				Category:    "permissions",
				Operation:   "create",
				Field:       "team_permission",
				OldValue:    nil,
				NewValue:    fmt.Sprintf("%s: %s", team, afterPerm),
				Risk:        RiskLow,
				Description: fmt.Sprintf("New team permission: %s", team),
				Impact:      "Team will gain access to repository",
			})
		}
	}

	// Check user permission changes
	for user, afterPerm := range after.Users {
		if beforePerm, exists := before.Users[user]; exists {
			if cp.isPermissionEscalation(beforePerm, afterPerm) {
				changes = append(changes, SensitiveChange{
					Repository:  repoName,
					Category:    "permissions",
					Operation:   "update",
					Field:       "user_permission",
					OldValue:    fmt.Sprintf("%s: %s", user, beforePerm),
					NewValue:    fmt.Sprintf("%s: %s", user, afterPerm),
					Risk:        RiskMedium,
					Description: fmt.Sprintf("User permission escalation: %s", user),
					Impact:      "User will have increased access to repository",
				})
			}
		} else {
			// New user permission
			changes = append(changes, SensitiveChange{
				Repository:  repoName,
				Category:    "permissions",
				Operation:   "create",
				Field:       "user_permission",
				OldValue:    nil,
				NewValue:    fmt.Sprintf("%s: %s", user, afterPerm),
				Risk:        RiskLow,
				Description: fmt.Sprintf("New user permission: %s", user),
				Impact:      "User will gain access to repository",
			})
		}
	}

	return changes
}

// isPermissionEscalation checks if a permission change represents an escalation.
func (cp *ConfirmationPrompt) isPermissionEscalation(before, after string) bool {
	levels := map[string]int{
		"read":     1,
		"triage":   2,
		"write":    3,
		"maintain": 4,
		"admin":    5,
	}

	beforeLevel, beforeExists := levels[before]
	afterLevel, afterExists := levels[after]

	if !beforeExists || !afterExists {
		return false
	}

	return afterLevel > beforeLevel
}

// Helper functions for display formatting.
func (cp *ConfirmationPrompt) getRiskColor(risk RiskLevel) string {
	switch risk {
	case RiskCritical:
		return "\033[1;91m" // Bright red
	case RiskHigh:
		return "\033[1;31m" // Red
	case RiskMedium:
		return "\033[1;33m" // Yellow
	case RiskLow:
		return "\033[1;32m" // Green
	case SecurityRiskLevelMinimal:
		return "\033[1;90m" // Bright black (gray)
	default:
		return ""
	}
}

func (cp *ConfirmationPrompt) getRiskIcon(risk RiskLevel) string {
	switch risk {
	case RiskCritical:
		return "🚨"
	case RiskHigh:
		return "⚠️"
	case RiskMedium:
		return "⚡"
	case RiskLow:
		return "ℹ️"
	case SecurityRiskLevelMinimal:
		return "🔸"
	default:
		return "•"
	}
}

func (cp *ConfirmationPrompt) getResetColor() string {
	return "\033[0m"
}
