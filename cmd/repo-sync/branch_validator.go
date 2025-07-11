package reposync

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

// BranchValidator handles branch name validation and suggestion
type BranchValidator struct {
	logger *zap.Logger
	rules  *BranchNamingRules
}

// BranchNamingRules defines branch naming rules
type BranchNamingRules struct {
	Template     string                      `json:"template"`
	Patterns     map[string]*regexp.Regexp   `json:"patterns"`
	Prefixes     []string                    `json:"prefixes"`
	MaxLength    int                         `json:"max_length"`
	AllowedChars *regexp.Regexp              `json:"allowed_chars"`
	Conventions  map[string]NamingConvention `json:"conventions"`
}

// NamingConvention represents a naming convention for a branch type
type NamingConvention struct {
	Prefix      string   `json:"prefix"`
	Pattern     string   `json:"pattern"`
	Description string   `json:"description"`
	Examples    []string `json:"examples"`
}

// ValidationResult represents the result of branch name validation
type ValidationResult struct {
	Valid       bool     `json:"valid"`
	BranchName  string   `json:"branch_name"`
	BranchType  string   `json:"branch_type"`
	Errors      []string `json:"errors"`
	Suggestions []string `json:"suggestions"`
	FixedName   string   `json:"fixed_name,omitempty"`
}

// NewBranchValidator creates a new branch validator
func NewBranchValidator(logger *zap.Logger, template string) *BranchValidator {
	rules := createNamingRules(template)
	return &BranchValidator{
		logger: logger,
		rules:  rules,
	}
}

// ValidateBranchName validates a branch name against the rules
func (bv *BranchValidator) ValidateBranchName(branchName string) *ValidationResult {
	result := &ValidationResult{
		BranchName:  branchName,
		Valid:       true,
		Errors:      make([]string, 0),
		Suggestions: make([]string, 0),
	}

	// Check if branch is protected and should not be validated
	if bv.isProtectedBranch(branchName) {
		result.BranchType = "protected"
		return result
	}

	// Check length
	if bv.rules.MaxLength > 0 && len(branchName) > bv.rules.MaxLength {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Branch name exceeds maximum length of %d characters", bv.rules.MaxLength))
		result.Suggestions = append(result.Suggestions, "Shorten the branch name")
	}

	// Check allowed characters
	if bv.rules.AllowedChars != nil && !bv.rules.AllowedChars.MatchString(branchName) {
		result.Valid = false
		result.Errors = append(result.Errors, "Branch name contains invalid characters")
		result.Suggestions = append(result.Suggestions, "Use only lowercase letters, numbers, hyphens, and forward slashes")
	}

	// Check against patterns
	branchType, matched := bv.matchBranchType(branchName)
	if !matched {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Branch name doesn't match any pattern for %s template", bv.rules.Template))
		result.Suggestions = bv.generateSuggestions(branchName)
	} else {
		result.BranchType = branchType
	}

	// Generate fixed name if validation failed
	if !result.Valid {
		result.FixedName = bv.suggestFixedName(branchName)
	}

	return result
}

// SuggestBranchName suggests a valid branch name based on input
func (bv *BranchValidator) SuggestBranchName(branchType, description string) (string, error) {
	convention, exists := bv.rules.Conventions[branchType]
	if !exists {
		return "", fmt.Errorf("unknown branch type: %s", branchType)
	}

	// Clean and format description
	cleanDesc := bv.cleanDescription(description)

	// Generate branch name based on convention
	branchName := fmt.Sprintf("%s/%s", convention.Prefix, cleanDesc)

	// Validate the generated name
	result := bv.ValidateBranchName(branchName)
	if !result.Valid && result.FixedName != "" {
		branchName = result.FixedName
	}

	return branchName, nil
}

// BatchValidate validates multiple branch names
func (bv *BranchValidator) BatchValidate(branchNames []string) map[string]*ValidationResult {
	results := make(map[string]*ValidationResult)

	for _, branchName := range branchNames {
		results[branchName] = bv.ValidateBranchName(branchName)
	}

	return results
}

// GetNamingConventions returns the naming conventions for the current template
func (bv *BranchValidator) GetNamingConventions() map[string]NamingConvention {
	return bv.rules.Conventions
}

// Helper methods

func (bv *BranchValidator) isProtectedBranch(branchName string) bool {
	protected := []string{"main", "master", "develop", "staging", "production"}
	for _, p := range protected {
		if branchName == p {
			return true
		}
	}
	return false
}

func (bv *BranchValidator) matchBranchType(branchName string) (string, bool) {
	for branchType, pattern := range bv.rules.Patterns {
		if pattern.MatchString(branchName) {
			return branchType, true
		}
	}
	return "", false
}

func (bv *BranchValidator) generateSuggestions(branchName string) []string {
	suggestions := make([]string, 0)

	for branchType, convention := range bv.rules.Conventions {
		suggestions = append(suggestions, fmt.Sprintf("For %s branches, use format: %s", branchType, convention.Pattern))
		if len(convention.Examples) > 0 {
			suggestions = append(suggestions, fmt.Sprintf("Example: %s", convention.Examples[0]))
		}
	}

	return suggestions
}

func (bv *BranchValidator) suggestFixedName(branchName string) string {
	// Clean the branch name
	fixed := strings.ToLower(branchName)
	fixed = regexp.MustCompile(`[^a-z0-9-/]+`).ReplaceAllString(fixed, "-")
	fixed = regexp.MustCompile(`-+`).ReplaceAllString(fixed, "-")
	fixed = strings.Trim(fixed, "-/")

	// Try to detect branch type from name
	if strings.Contains(fixed, "feature") || strings.Contains(fixed, "feat") {
		fixed = regexp.MustCompile(`^(feature|feat)[-/]*`).ReplaceAllString(fixed, "feature/")
	} else if strings.Contains(fixed, "fix") || strings.Contains(fixed, "bug") {
		fixed = regexp.MustCompile(`^(fix|bug|bugfix)[-/]*`).ReplaceAllString(fixed, "fix/")
	} else if strings.Contains(fixed, "release") || strings.Contains(fixed, "rel") {
		fixed = regexp.MustCompile(`^(release|rel)[-/]*`).ReplaceAllString(fixed, "release/")
	} else if strings.Contains(fixed, "hotfix") || strings.Contains(fixed, "hf") {
		fixed = regexp.MustCompile(`^(hotfix|hf)[-/]*`).ReplaceAllString(fixed, "hotfix/")
	} else {
		// Default to feature if no type detected
		if !strings.Contains(fixed, "/") {
			fixed = "feature/" + fixed
		}
	}

	// Ensure it matches at least one pattern
	if _, matched := bv.matchBranchType(fixed); !matched && strings.Contains(fixed, "/") {
		// If still doesn't match, use the most permissive pattern
		parts := strings.SplitN(fixed, "/", 2)
		if len(parts) == 2 {
			fixed = fmt.Sprintf("%s/%s", parts[0], parts[1])
		}
	}

	// Truncate if too long
	if bv.rules.MaxLength > 0 && len(fixed) > bv.rules.MaxLength {
		fixed = fixed[:bv.rules.MaxLength]
		// Ensure we don't cut off in the middle of a word
		if idx := strings.LastIndex(fixed, "-"); idx > 0 && idx > len(fixed)-10 {
			fixed = fixed[:idx]
		}
	}

	return fixed
}

func (bv *BranchValidator) cleanDescription(description string) string {
	// Convert to lowercase
	clean := strings.ToLower(description)

	// Replace spaces and special characters with hyphens
	clean = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(clean, "-")

	// Remove multiple consecutive hyphens
	clean = regexp.MustCompile(`-+`).ReplaceAllString(clean, "-")

	// Trim hyphens from start and end
	clean = strings.Trim(clean, "-")

	// Limit length
	if len(clean) > 50 {
		clean = clean[:50]
		// Ensure we don't cut off in the middle of a word
		if idx := strings.LastIndex(clean, "-"); idx > 0 && idx > 40 {
			clean = clean[:idx]
		}
	}

	return clean
}

// createNamingRules creates naming rules based on template
func createNamingRules(template string) *BranchNamingRules {
	switch strings.ToLower(template) {
	case "gitflow":
		return createGitFlowNamingRules()
	case "github-flow":
		return createGitHubFlowNamingRules()
	case "gitlab-flow":
		return createGitLabFlowNamingRules()
	default:
		return createCustomNamingRules()
	}
}

func createGitFlowNamingRules() *BranchNamingRules {
	patterns := make(map[string]*regexp.Regexp)
	patterns["feature"] = regexp.MustCompile(`^feature/[a-z0-9-]+$`)
	patterns["release"] = regexp.MustCompile(`^release/\d+\.\d+\.\d+$`)
	patterns["hotfix"] = regexp.MustCompile(`^hotfix/[a-z0-9-]+$`)
	patterns["bugfix"] = regexp.MustCompile(`^bugfix/[a-z0-9-]+$`)

	conventions := map[string]NamingConvention{
		"feature": {
			Prefix:      "feature",
			Pattern:     "feature/description-of-feature",
			Description: "New features or enhancements",
			Examples:    []string{"feature/user-authentication", "feature/payment-integration"},
		},
		"release": {
			Prefix:      "release",
			Pattern:     "release/X.Y.Z",
			Description: "Release preparation branches",
			Examples:    []string{"release/1.2.0", "release/2.0.0-beta"},
		},
		"hotfix": {
			Prefix:      "hotfix",
			Pattern:     "hotfix/description-of-fix",
			Description: "Emergency fixes for production",
			Examples:    []string{"hotfix/critical-security-patch", "hotfix/payment-bug"},
		},
		"bugfix": {
			Prefix:      "bugfix",
			Pattern:     "bugfix/description-of-bug",
			Description: "Bug fixes for development",
			Examples:    []string{"bugfix/login-validation", "bugfix/memory-leak"},
		},
	}

	return &BranchNamingRules{
		Template:     "gitflow",
		Patterns:     patterns,
		Prefixes:     []string{"feature", "release", "hotfix", "bugfix"},
		MaxLength:    80,
		AllowedChars: regexp.MustCompile(`^[a-z0-9-/]+$`),
		Conventions:  conventions,
	}
}

func createGitHubFlowNamingRules() *BranchNamingRules {
	patterns := make(map[string]*regexp.Regexp)
	patterns["feature"] = regexp.MustCompile(`^[a-z0-9-]+/[a-z0-9-]+$`)
	patterns["fix"] = regexp.MustCompile(`^fix/[a-z0-9-]+$`)

	conventions := map[string]NamingConvention{
		"feature": {
			Prefix:      "username",
			Pattern:     "username/description",
			Description: "Feature branches with username prefix",
			Examples:    []string{"john/add-user-profile", "mary/update-dashboard"},
		},
		"fix": {
			Prefix:      "fix",
			Pattern:     "fix/description",
			Description: "Bug fix branches",
			Examples:    []string{"fix/navigation-error", "fix/data-validation"},
		},
	}

	return &BranchNamingRules{
		Template:     "github-flow",
		Patterns:     patterns,
		Prefixes:     []string{"fix"},
		MaxLength:    63,
		AllowedChars: regexp.MustCompile(`^[a-z0-9-/]+$`),
		Conventions:  conventions,
	}
}

func createGitLabFlowNamingRules() *BranchNamingRules {
	patterns := make(map[string]*regexp.Regexp)
	patterns["feature"] = regexp.MustCompile(`^feature/[a-z0-9-]+$`)
	patterns["environment"] = regexp.MustCompile(`^(staging|production)$`)

	conventions := map[string]NamingConvention{
		"feature": {
			Prefix:      "feature",
			Pattern:     "feature/issue-description",
			Description: "Feature branches linked to issues",
			Examples:    []string{"feature/123-user-auth", "feature/456-api-refactor"},
		},
	}

	return &BranchNamingRules{
		Template:     "gitlab-flow",
		Patterns:     patterns,
		Prefixes:     []string{"feature"},
		MaxLength:    100,
		AllowedChars: regexp.MustCompile(`^[a-z0-9-/]+$`),
		Conventions:  conventions,
	}
}

func createCustomNamingRules() *BranchNamingRules {
	patterns := make(map[string]*regexp.Regexp)
	patterns["any"] = regexp.MustCompile(`^[a-zA-Z0-9-_/]+$`)

	conventions := map[string]NamingConvention{
		"custom": {
			Prefix:      "",
			Pattern:     "any-valid-branch-name",
			Description: "Custom branch naming",
			Examples:    []string{"my-feature", "team/sprint-1/task-123"},
		},
	}

	return &BranchNamingRules{
		Template:     "custom",
		Patterns:     patterns,
		Prefixes:     []string{},
		MaxLength:    255,
		AllowedChars: regexp.MustCompile(`^[a-zA-Z0-9-_/]+$`),
		Conventions:  conventions,
	}
}
