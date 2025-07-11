package reposync

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewBranchValidator(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name     string
		template string
	}{
		{"GitFlow", "gitflow"},
		{"GitHub Flow", "github-flow"},
		{"GitLab Flow", "gitlab-flow"},
		{"Custom", "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewBranchValidator(logger, tt.template)
			assert.NotNil(t, validator)
			assert.Equal(t, tt.template, validator.rules.Template)
			assert.NotEmpty(t, validator.rules.Patterns)
		})
	}
}

func TestValidateBranchName_GitFlow(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewBranchValidator(logger, "gitflow")

	tests := []struct {
		name       string
		branchName string
		valid      bool
		branchType string
	}{
		// Valid branches
		{"valid feature", "feature/user-authentication", true, "feature"},
		{"valid release", "release/1.2.0", true, "release"},
		{"valid hotfix", "hotfix/critical-bug", true, "hotfix"},
		{"valid bugfix", "bugfix/login-error", true, "bugfix"},
		{"protected main", "main", true, "protected"},
		{"protected develop", "develop", true, "protected"},

		// Invalid branches
		{"invalid feature caps", "feature/User-Auth", false, ""},
		{"invalid feature special", "feature/user@auth", false, ""},
		{"invalid release format", "release/v1.2.0", false, ""},
		{"missing prefix", "user-authentication", false, ""},
		{"wrong separator", "feature_user_auth", false, ""},
		{"too long", "feature/" + string(make([]byte, 100)), false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateBranchName(tt.branchName)
			assert.Equal(t, tt.valid, result.Valid)
			if tt.branchType != "" {
				assert.Equal(t, tt.branchType, result.BranchType)
			}
			if !tt.valid && result.FixedName != "" {
				// Verify the fixed name is valid
				fixedResult := validator.ValidateBranchName(result.FixedName)
				assert.True(t, fixedResult.Valid, "Fixed name should be valid")
			}
		})
	}
}

func TestValidateBranchName_GitHubFlow(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewBranchValidator(logger, "github-flow")

	tests := []struct {
		name       string
		branchName string
		valid      bool
		branchType string
	}{
		// Valid branches
		{"valid feature", "john/add-user-profile", true, "feature"},
		{"valid fix", "fix/navigation-error", true, "fix"},
		{"protected main", "main", true, "protected"},

		// Invalid branches
		{"invalid caps", "John/add-feature", false, ""},
		{"invalid special char", "john/add@feature", false, ""},
		{"no prefix", "add-feature", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateBranchName(tt.branchName)
			assert.Equal(t, tt.valid, result.Valid)
			if tt.branchType != "" {
				assert.Equal(t, tt.branchType, result.BranchType)
			}
		})
	}
}

func TestSuggestBranchName(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewBranchValidator(logger, "gitflow")

	tests := []struct {
		name        string
		branchType  string
		description string
		expected    string
		shouldError bool
	}{
		{
			name:        "feature branch",
			branchType:  "feature",
			description: "Add user authentication",
			expected:    "feature/add-user-authentication",
			shouldError: false,
		},
		{
			name:        "hotfix with special chars",
			branchType:  "hotfix",
			description: "Fix critical bug in payment!",
			expected:    "hotfix/fix-critical-bug-in-payment",
			shouldError: false,
		},
		{
			name:        "long description",
			branchType:  "feature",
			description: "This is a very long description that should be truncated to fit within the maximum length limit",
			expected:    "feature/this-is-a-very-long-description-that-should-be",
			shouldError: false,
		},
		{
			name:        "invalid branch type",
			branchType:  "invalid",
			description: "Some description",
			expected:    "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestion, err := validator.SuggestBranchName(tt.branchType, tt.description)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, suggestion)

				// Verify suggested name is valid
				result := validator.ValidateBranchName(suggestion)
				assert.True(t, result.Valid, "Suggested branch name should be valid")
			}
		})
	}
}

func TestBatchValidate(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewBranchValidator(logger, "gitflow")

	branches := []string{
		"feature/valid-feature",
		"invalid-branch-name",
		"release/1.0.0",
		"Feature/Invalid-Caps",
		"main",
	}

	results := validator.BatchValidate(branches)

	assert.Len(t, results, len(branches))
	assert.True(t, results["feature/valid-feature"].Valid)
	assert.False(t, results["invalid-branch-name"].Valid)
	assert.True(t, results["release/1.0.0"].Valid)
	assert.False(t, results["Feature/Invalid-Caps"].Valid)
	assert.True(t, results["main"].Valid)
	assert.Equal(t, "protected", results["main"].BranchType)
}

func TestSuggestFixedName(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewBranchValidator(logger, "gitflow")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "add feature prefix",
			input:    "user-authentication",
			expected: "feature/user-authentication",
		},
		{
			name:     "fix caps and special chars",
			input:    "Feature/User@Authentication!",
			expected: "feature/user-authentication",
		},
		{
			name:     "detect fix type",
			input:    "bugfix-login-error",
			expected: "bugfix/login-error",
		},
		{
			name:     "detect release type",
			input:    "release-1.2.0",
			expected: "release/1.2.0",
		},
		{
			name:     "clean up multiple hyphens",
			input:    "feature///user---auth",
			expected: "feature/user-auth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateBranchName(tt.input)
			assert.False(t, result.Valid)
			assert.Equal(t, tt.expected, result.FixedName)

			// Verify fixed name is valid
			fixedResult := validator.ValidateBranchName(result.FixedName)
			assert.True(t, fixedResult.Valid)
		})
	}
}

func TestCleanDescription(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewBranchValidator(logger, "gitflow")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple description",
			input:    "Add user authentication",
			expected: "add-user-authentication",
		},
		{
			name:     "special characters",
			input:    "Fix bug #123 (critical!)",
			expected: "fix-bug-123-critical",
		},
		{
			name:     "multiple spaces",
			input:    "Add   multiple   spaces   handling",
			expected: "add-multiple-spaces-handling",
		},
		{
			name:     "very long description",
			input:    "This is a very long description that exceeds fifty characters limit",
			expected: "this-is-a-very-long-description-that-exceeds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaned := validator.cleanDescription(tt.input)
			assert.Equal(t, tt.expected, cleaned)
		})
	}
}

func TestGetNamingConventions(t *testing.T) {
	logger := zaptest.NewLogger(t)

	templates := []string{"gitflow", "github-flow", "gitlab-flow", "custom"}

	for _, template := range templates {
		t.Run(template, func(t *testing.T) {
			validator := NewBranchValidator(logger, template)
			conventions := validator.GetNamingConventions()

			assert.NotEmpty(t, conventions)

			// Verify each convention has required fields
			for _, convention := range conventions {
				assert.NotEmpty(t, convention.Pattern)
				assert.NotEmpty(t, convention.Description)
			}
		})
	}
}

func TestProtectedBranches(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewBranchValidator(logger, "gitflow")

	protectedBranches := []string{"main", "master", "develop", "staging", "production"}

	for _, branch := range protectedBranches {
		t.Run(branch, func(t *testing.T) {
			result := validator.ValidateBranchName(branch)
			assert.True(t, result.Valid)
			assert.Equal(t, "protected", result.BranchType)
			assert.Empty(t, result.Errors)
		})
	}
}
