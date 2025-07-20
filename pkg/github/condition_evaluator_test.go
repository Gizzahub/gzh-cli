//nolint:testpackage // White-box testing needed for internal function access
package github

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper functions.
func createTestEvaluationContext() *EvaluationContext {
	return &EvaluationContext{
		Repository: &RepositoryInfo{
			Name:       "test-repo",
			Language:   "Go",
			Topics:     []string{"api", "backend"},
			Visibility: "public",
			Archived:   false,
			IsTemplate: false,
		},
		Organization: &OrganizationInfo{
			Login:       "testorg",
			Name:        "Test Organization",
			Type:        "Organization",
			Plan:        "free",
			MemberCount: 10,
			RepoCount:   50,
		},
		User: &UserInfo{
			Login: "testuser",
			Name:  "Test User",
			Type:  "User",
		},
		Environment: "production",
		Variables: map[string]interface{}{
			"test_var": "test_value",
		},
		Timezone: time.UTC,
	}
}

func createTestEvent() *GitHubEvent {
	return &GitHubEvent{
		ID:           "test-event-001",
		Type:         "pull_request",
		Action:       "opened",
		Organization: "testorg",
		Repository:   "test-repo",
		Sender:       "testuser",
		Timestamp:    time.Now(),
		Payload: map[string]interface{}{
			"action": "opened",
			"pull_request": map[string]interface{}{
				"title": "Fix bug in authentication",
				"head": map[string]interface{}{
					"ref": "feature/fix-auth",
				},
				"base": map[string]interface{}{
					"ref": "main",
				},
				"changed_files": []interface{}{
					"auth/handler.go",
					"auth/middleware.go",
					"tests/auth_test.go",
				},
			},
			"repository": map[string]interface{}{
				"name": "test-repo",
				"owner": map[string]interface{}{
					"login": "testorg",
				},
			},
			"sender": map[string]interface{}{
				"login": "testuser",
			},
		},
	}
}

func createTestConditions() *AutomationConditions {
	return &AutomationConditions{
		EventTypes:         []EventType{EventTypePullRequest},
		Actions:            []EventAction{ActionOpened, ActionSynchronize},
		Organization:       "testorg",
		Repository:         "test-repo",
		RepositoryPatterns: []string{"^test-.*", "^api-.*"},
		Languages:          []string{"Go", "JavaScript"},
		Topics:             []string{"api", "backend"},
		Visibility:         []string{"public"},
		BranchPatterns:     []string{"^feature/.*", "^fix/.*"},
		FilePatterns:       []string{"*.go", "*.js"},
		LogicalOperator:    ConditionOperatorAND,
		PayloadMatch: []PayloadMatcher{
			{
				Path:          "$.pull_request.title",
				Operator:      MatchOperatorContains,
				Value:         "fix",
				CaseSensitive: false,
			},
			{
				Path:     "$.pull_request.base.ref",
				Operator: MatchOperatorEquals,
				Value:    "main",
			},
		},
	}
}

func TestNewConditionEvaluator(t *testing.T) {
	logger := &mockLogger{}
	apiClient := &mockAPIClient{}

	evaluator := NewConditionEvaluator(logger, apiClient)

	assert.NotNil(t, evaluator)
	assert.IsType(t, &conditionEvaluatorImpl{}, evaluator)
}

func TestConditionEvaluator_EvaluateConditions_Success(t *testing.T) {
	logger := &mockLogger{}
	apiClient := &mockAPIClient{}
	evaluator := NewConditionEvaluator(logger, apiClient)

	conditions := createTestConditions()
	event := createTestEvent()
	evalContext := createTestEvaluationContext()

	result, err := evaluator.EvaluateConditions(context.Background(), conditions, event, evalContext)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Matched)
	assert.Greater(t, len(result.MatchedConditions), 0)
	assert.Equal(t, 0, len(result.FailedConditions))
	assert.Equal(t, 2, len(result.PayloadMatchResults))
	assert.Greater(t, result.EvaluationTime, time.Duration(0))
}

func TestConditionEvaluator_EvaluateEventConditions(t *testing.T) {
	logger := &mockLogger{}
	apiClient := &mockAPIClient{}
	evaluator, ok := NewConditionEvaluator(logger, apiClient).(*conditionEvaluatorImpl)
	require.True(t, ok, "evaluator should be of correct type")

	tests := []struct {
		name       string
		event      *GitHubEvent
		conditions *AutomationConditions
		expected   bool
	}{
		{
			name:  "matching event type and action",
			event: createTestEvent(),
			conditions: &AutomationConditions{
				EventTypes: []EventType{EventTypePullRequest},
				Actions:    []EventAction{ActionOpened},
			},
			expected: true,
		},
		{
			name:  "non-matching event type",
			event: createTestEvent(),
			conditions: &AutomationConditions{
				EventTypes: []EventType{EventTypePush},
			},
			expected: false,
		},
		{
			name:  "non-matching action",
			event: createTestEvent(),
			conditions: &AutomationConditions{
				EventTypes: []EventType{EventTypePullRequest},
				Actions:    []EventAction{ActionClosed},
			},
			expected: false,
		},
		{
			name:  "matching organization",
			event: createTestEvent(),
			conditions: &AutomationConditions{
				Organization: "testorg",
			},
			expected: true,
		},
		{
			name:  "non-matching organization",
			event: createTestEvent(),
			conditions: &AutomationConditions{
				Organization: "other-org",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateEventConditions(tt.event, tt.conditions)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConditionEvaluator_EvaluateRepositoryConditions(t *testing.T) {
	logger := &mockLogger{}
	apiClient := &mockAPIClient{}
	evaluator, ok := NewConditionEvaluator(logger, apiClient).(*conditionEvaluatorImpl)
	require.True(t, ok, "evaluator should be of correct type")

	repoInfo := &RepositoryInfo{
		Name:       "test-repo",
		Language:   "Go",
		Topics:     []string{"api", "backend"},
		Visibility: "public",
		Archived:   false,
		IsTemplate: false,
	}

	tests := []struct {
		name       string
		repoInfo   *RepositoryInfo
		conditions *AutomationConditions
		expected   bool
	}{
		{
			name:     "matching repository pattern",
			repoInfo: repoInfo,
			conditions: &AutomationConditions{
				RepositoryPatterns: []string{"^test-.*"},
			},
			expected: true,
		},
		{
			name:     "non-matching repository pattern",
			repoInfo: repoInfo,
			conditions: &AutomationConditions{
				RepositoryPatterns: []string{"^api-.*"},
			},
			expected: false,
		},
		{
			name:     "matching language",
			repoInfo: repoInfo,
			conditions: &AutomationConditions{
				Languages: []string{"Go", "JavaScript"},
			},
			expected: true,
		},
		{
			name:     "non-matching language",
			repoInfo: repoInfo,
			conditions: &AutomationConditions{
				Languages: []string{"Python", "Java"},
			},
			expected: false,
		},
		{
			name:     "matching topics",
			repoInfo: repoInfo,
			conditions: &AutomationConditions{
				Topics: []string{"api"},
			},
			expected: true,
		},
		{
			name:     "non-matching topics",
			repoInfo: repoInfo,
			conditions: &AutomationConditions{
				Topics: []string{"frontend", "mobile"},
			},
			expected: false,
		},
		{
			name:     "matching visibility",
			repoInfo: repoInfo,
			conditions: &AutomationConditions{
				Visibility: []string{"public", "private"},
			},
			expected: true,
		},
		{
			name:     "non-matching visibility",
			repoInfo: repoInfo,
			conditions: &AutomationConditions{
				Visibility: []string{"private"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateRepositoryConditions(context.Background(), tt.repoInfo, tt.conditions)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConditionEvaluator_EvaluateTimeConditions(t *testing.T) {
	logger := &mockLogger{}
	apiClient := &mockAPIClient{}
	evaluator, ok := NewConditionEvaluator(logger, apiClient).(*conditionEvaluatorImpl)
	require.True(t, ok, "evaluator should be of correct type")

	// Tuesday, 2:00 PM UTC
	testTime := time.Date(2024, 1, 9, 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		timestamp  time.Time
		conditions *AutomationConditions
		expected   bool
	}{
		{
			name:      "within time range",
			timestamp: testTime,
			conditions: &AutomationConditions{
				TimeRange: &TimeRange{
					Start: testTime.Add(-1 * time.Hour),
					End:   testTime.Add(1 * time.Hour),
				},
			},
			expected: true,
		},
		{
			name:      "outside time range",
			timestamp: testTime,
			conditions: &AutomationConditions{
				TimeRange: &TimeRange{
					Start: testTime.Add(1 * time.Hour),
					End:   testTime.Add(2 * time.Hour),
				},
			},
			expected: false,
		},
		{
			name:      "matching day of week (Tuesday = 2)",
			timestamp: testTime,
			conditions: &AutomationConditions{
				DaysOfWeek: []int{1, 2, 3}, // Monday, Tuesday, Wednesday
			},
			expected: true,
		},
		{
			name:      "non-matching day of week",
			timestamp: testTime,
			conditions: &AutomationConditions{
				DaysOfWeek: []int{0, 6}, // Sunday, Saturday
			},
			expected: false,
		},
		{
			name:      "matching hour of day",
			timestamp: testTime,
			conditions: &AutomationConditions{
				HoursOfDay: []int{13, 14, 15}, // 1 PM, 2 PM, 3 PM
			},
			expected: true,
		},
		{
			name:      "non-matching hour of day",
			timestamp: testTime,
			conditions: &AutomationConditions{
				HoursOfDay: []int{9, 10, 11}, // 9 AM, 10 AM, 11 AM
			},
			expected: false,
		},
		{
			name:      "business hours - within",
			timestamp: testTime, // Tuesday 2 PM
			conditions: &AutomationConditions{
				BusinessHours: true,
			},
			expected: true,
		},
		{
			name:      "business hours - outside (weekend)",
			timestamp: time.Date(2024, 1, 7, 14, 0, 0, 0, time.UTC), // Sunday 2 PM
			conditions: &AutomationConditions{
				BusinessHours: true,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateTimeConditions(tt.timestamp, tt.conditions)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConditionEvaluator_EvaluateContentConditions(t *testing.T) {
	logger := &mockLogger{}
	apiClient := &mockAPIClient{}
	evaluator, ok := NewConditionEvaluator(logger, apiClient).(*conditionEvaluatorImpl)
	require.True(t, ok, "evaluator should be of correct type")

	event := createTestEvent()

	tests := []struct {
		name       string
		event      *GitHubEvent
		conditions *AutomationConditions
		expected   bool
	}{
		{
			name:  "matching branch pattern",
			event: event,
			conditions: &AutomationConditions{
				BranchPatterns: []string{"^feature/.*"},
			},
			expected: true,
		},
		{
			name:  "non-matching branch pattern",
			event: event,
			conditions: &AutomationConditions{
				BranchPatterns: []string{"^hotfix/.*"},
			},
			expected: false,
		},
		{
			name:  "matching file pattern",
			event: event,
			conditions: &AutomationConditions{
				FilePatterns: []string{"*.go"},
			},
			expected: true,
		},
		{
			name:  "non-matching file pattern",
			event: event,
			conditions: &AutomationConditions{
				FilePatterns: []string{"*.py"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateContentConditions(context.Background(), tt.event, tt.conditions)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConditionEvaluator_EvaluatePayloadMatcher(t *testing.T) {
	logger := &mockLogger{}
	apiClient := &mockAPIClient{}
	evaluator := NewConditionEvaluator(logger, apiClient)

	payload := map[string]interface{}{
		"pull_request": map[string]interface{}{
			"title":  "Fix bug in authentication",
			"number": 123,
			"state":  "open",
		},
		"action": "opened",
	}

	tests := []struct {
		name     string
		matcher  *PayloadMatcher
		payload  map[string]interface{}
		expected bool
		wantErr  bool
	}{
		{
			name: "string equals match",
			matcher: &PayloadMatcher{
				Path:     "$.action",
				Operator: MatchOperatorEquals,
				Value:    "opened",
			},
			payload:  payload,
			expected: true,
		},
		{
			name: "string contains match",
			matcher: &PayloadMatcher{
				Path:          "$.pull_request.title",
				Operator:      MatchOperatorContains,
				Value:         "fix",
				CaseSensitive: false,
			},
			payload:  payload,
			expected: true,
		},
		{
			name: "string contains no match",
			matcher: &PayloadMatcher{
				Path:          "$.pull_request.title",
				Operator:      MatchOperatorContains,
				Value:         "feature",
				CaseSensitive: false,
			},
			payload:  payload,
			expected: false,
		},
		{
			name: "number greater than",
			matcher: &PayloadMatcher{
				Path:     "$.pull_request.number",
				Operator: MatchOperatorGreaterThan,
				Value:    100,
			},
			payload:  payload,
			expected: true,
		},
		{
			name: "number less than",
			matcher: &PayloadMatcher{
				Path:     "$.pull_request.number",
				Operator: MatchOperatorLessThan,
				Value:    100,
			},
			payload:  payload,
			expected: false,
		},
		{
			name: "regex match",
			matcher: &PayloadMatcher{
				Path:     "$.pull_request.title",
				Operator: MatchOperatorRegex,
				Value:    "^Fix.*authentication$",
			},
			payload:  payload,
			expected: true,
		},
		{
			name: "exists operator - path exists",
			matcher: &PayloadMatcher{
				Path:     "$.pull_request.title",
				Operator: MatchOperatorExists,
				Value:    nil,
			},
			payload:  payload,
			expected: true,
		},
		{
			name: "not exists operator - path exists",
			matcher: &PayloadMatcher{
				Path:     "$.pull_request.title",
				Operator: MatchOperatorNotExists,
				Value:    nil,
			},
			payload:  payload,
			expected: false,
		},
		{
			name: "path not found",
			matcher: &PayloadMatcher{
				Path:     "$.nonexistent.field",
				Operator: MatchOperatorEquals,
				Value:    "value",
			},
			payload: payload,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluatePayloadMatcher(context.Background(), tt.matcher, tt.payload)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestConditionEvaluator_ValidateConditions(t *testing.T) {
	logger := &mockLogger{}
	apiClient := &mockAPIClient{}
	evaluator, ok := NewConditionEvaluator(logger, apiClient).(*conditionEvaluatorImpl)
	require.True(t, ok, "evaluator should be of correct type")

	tests := []struct {
		name       string
		conditions *AutomationConditions
		wantValid  bool
		wantErrors int
	}{
		{
			name:       "valid conditions",
			conditions: createTestConditions(),
			wantValid:  true,
			wantErrors: 0,
		},
		{
			name: "invalid regex pattern",
			conditions: &AutomationConditions{
				RepositoryPatterns: []string{"[invalid"},
			},
			wantValid:  false,
			wantErrors: 1,
		},
		{
			name: "invalid JSONPath",
			conditions: &AutomationConditions{
				PayloadMatch: []PayloadMatcher{
					{
						Path:     "invalid.path", // Missing $ prefix
						Operator: MatchOperatorEquals,
						Value:    "test",
					},
				},
			},
			wantValid:  false,
			wantErrors: 1,
		},
		{
			name: "invalid time range",
			conditions: &AutomationConditions{
				TimeRange: &TimeRange{
					Start: time.Now().Add(1 * time.Hour),
					End:   time.Now(),
				},
			},
			wantValid:  false,
			wantErrors: 1,
		},
		{
			name: "invalid day of week",
			conditions: &AutomationConditions{
				DaysOfWeek: []int{7, 8}, // Valid range is 0-6
			},
			wantValid:  false,
			wantErrors: 2,
		},
		{
			name: "invalid hour of day",
			conditions: &AutomationConditions{
				HoursOfDay: []int{24, 25}, // Valid range is 0-23
			},
			wantValid:  false,
			wantErrors: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.ValidateConditions(tt.conditions)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.wantValid, result.Valid)
			assert.Equal(t, tt.wantErrors, len(result.Errors))
		})
	}
}

func TestConditionEvaluator_LogicalOperators(t *testing.T) {
	logger := &mockLogger{}
	apiClient := &mockAPIClient{}
	evaluator, ok := NewConditionEvaluator(logger, apiClient).(*conditionEvaluatorImpl)
	require.True(t, ok, "evaluator should be of correct type")

	// Create test evaluation results
	tests := []struct {
		name           string
		operator       ConditionOperator
		matchedCount   int
		failedCount    int
		expectedResult bool
	}{
		{
			name:           "AND operator - all pass",
			operator:       ConditionOperatorAND,
			matchedCount:   3,
			failedCount:    0,
			expectedResult: true,
		},
		{
			name:           "AND operator - some fail",
			operator:       ConditionOperatorAND,
			matchedCount:   2,
			failedCount:    1,
			expectedResult: false,
		},
		{
			name:           "OR operator - some pass",
			operator:       ConditionOperatorOR,
			matchedCount:   1,
			failedCount:    2,
			expectedResult: true,
		},
		{
			name:           "OR operator - none pass",
			operator:       ConditionOperatorOR,
			matchedCount:   0,
			failedCount:    3,
			expectedResult: false,
		},
		{
			name:           "NOT operator - some conditions",
			operator:       ConditionOperatorNOT,
			matchedCount:   0,
			failedCount:    3,
			expectedResult: true,
		},
		{
			name:           "NOT operator - conditions match",
			operator:       ConditionOperatorNOT,
			matchedCount:   2,
			failedCount:    1,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &EvaluationResult{
				MatchedConditions: make([]string, tt.matchedCount),
				FailedConditions:  make([]string, tt.failedCount),
			}

			// Fill with dummy condition names
			for i := 0; i < tt.matchedCount; i++ {
				result.MatchedConditions[i] = fmt.Sprintf("matched_%d", i)
			}

			for i := 0; i < tt.failedCount; i++ {
				result.FailedConditions[i] = fmt.Sprintf("failed_%d", i)
			}

			finalResult := evaluator.applyLogicalOperator(tt.operator, result)
			assert.Equal(t, tt.expectedResult, finalResult)
		})
	}
}

func TestConditionEvaluator_ExtractDataFromPayload(t *testing.T) {
	logger := &mockLogger{}
	apiClient := &mockAPIClient{}
	evaluator, ok := NewConditionEvaluator(logger, apiClient).(*conditionEvaluatorImpl)
	require.True(t, ok, "evaluator should be of correct type")

	// Test branch extraction
	t.Run("extract branch from push event", func(t *testing.T) {
		payload := map[string]interface{}{
			"ref": "refs/heads/feature/test-branch",
		}

		branch := evaluator.extractBranchFromPayload(payload)
		assert.Equal(t, "feature/test-branch", branch)
	})

	t.Run("extract branch from pull request event", func(t *testing.T) {
		payload := map[string]interface{}{
			"pull_request": map[string]interface{}{
				"head": map[string]interface{}{
					"ref": "feature/pr-branch",
				},
			},
		}

		branch := evaluator.extractBranchFromPayload(payload)
		assert.Equal(t, "feature/pr-branch", branch)
	})

	// Test file extraction
	t.Run("extract files from push event", func(t *testing.T) {
		payload := map[string]interface{}{
			"commits": []interface{}{
				map[string]interface{}{
					"added":    []interface{}{"file1.go", "file2.go"},
					"modified": []interface{}{"file3.go"},
				},
			},
		}

		files := evaluator.extractFilesFromPayload(payload)
		assert.Contains(t, files, "file1.go")
		assert.Contains(t, files, "file2.go")
		assert.Contains(t, files, "file3.go")
		assert.Len(t, files, 3)
	})
}

func TestConditionEvaluator_HelperMethods(t *testing.T) {
	logger := &mockLogger{}
	apiClient := &mockAPIClient{}
	evaluator, ok := NewConditionEvaluator(logger, apiClient).(*conditionEvaluatorImpl)
	require.True(t, ok, "evaluator should be of correct type")

	// Test isEmpty
	t.Run("isEmpty checks", func(t *testing.T) {
		assert.True(t, evaluator.isEmpty(nil))
		assert.True(t, evaluator.isEmpty(""))
		assert.True(t, evaluator.isEmpty([]interface{}{}))
		assert.True(t, evaluator.isEmpty(map[string]interface{}{}))
		assert.False(t, evaluator.isEmpty("not empty"))
		assert.False(t, evaluator.isEmpty([]interface{}{"item"}))
		assert.False(t, evaluator.isEmpty(map[string]interface{}{"key": "value"}))
	})

	// Test toFloat64
	t.Run("toFloat64 conversions", func(t *testing.T) {
		val, err := evaluator.toFloat64(123)
		assert.NoError(t, err)
		assert.Equal(t, 123.0, val)

		val, err = evaluator.toFloat64(123.45)
		assert.NoError(t, err)
		assert.Equal(t, 123.45, val)

		val, err = evaluator.toFloat64("123.45")
		assert.NoError(t, err)
		assert.Equal(t, 123.45, val)

		_, err = evaluator.toFloat64("not a number")
		assert.Error(t, err)
	})
}

// Integration test.
func TestConditionEvaluator_Integration(t *testing.T) {
	logger := &mockLogger{}
	apiClient := &mockAPIClient{}
	evaluator := NewConditionEvaluator(logger, apiClient)

	// Create a comprehensive test scenario
	conditions := &AutomationConditions{
		EventTypes:         []EventType{EventTypePullRequest},
		Actions:            []EventAction{ActionOpened},
		Organization:       "testorg",
		RepositoryPatterns: []string{"^test-.*"},
		Languages:          []string{"Go"},
		BranchPatterns:     []string{"^feature/.*"},
		FilePatterns:       []string{"*.go"},
		LogicalOperator:    ConditionOperatorAND,
		DaysOfWeek:         []int{1, 2, 3, 4, 5},                 // Weekdays
		HoursOfDay:         []int{9, 10, 11, 12, 13, 14, 15, 16}, // Business hours
		PayloadMatch: []PayloadMatcher{
			{
				Path:          "$.pull_request.title",
				Operator:      MatchOperatorContains,
				Value:         "fix",
				CaseSensitive: false,
			},
		},
	}

	event := createTestEvent()
	// Set event timestamp to a weekday during business hours
	event.Timestamp = time.Date(2024, 1, 9, 14, 0, 0, 0, time.UTC) // Tuesday 2 PM

	evalContext := createTestEvaluationContext()

	result, err := evaluator.EvaluateConditions(context.Background(), conditions, event, evalContext)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Matched, "All conditions should match")
	assert.Greater(t, len(result.MatchedConditions), 0)
	assert.Equal(t, 0, len(result.FailedConditions))
	assert.Greater(t, result.EvaluationTime, time.Duration(0))
}

// Benchmark tests.
func BenchmarkConditionEvaluator_EvaluateConditions(b *testing.B) {
	logger := &mockLogger{}
	apiClient := &mockAPIClient{}
	evaluator := NewConditionEvaluator(logger, apiClient)

	conditions := createTestConditions()
	event := createTestEvent()
	evalContext := createTestEvaluationContext()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = evaluator.EvaluateConditions(context.Background(), conditions, event, evalContext) //nolint:errcheck // Benchmark test
	}
}

func BenchmarkConditionEvaluator_EvaluatePayloadMatcher(b *testing.B) {
	logger := &mockLogger{}
	apiClient := &mockAPIClient{}
	evaluator := NewConditionEvaluator(logger, apiClient)

	matcher := &PayloadMatcher{
		Path:          "$.pull_request.title",
		Operator:      MatchOperatorContains,
		Value:         "fix",
		CaseSensitive: false,
	}

	payload := map[string]interface{}{
		"pull_request": map[string]interface{}{
			"title": "Fix bug in authentication",
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = evaluator.EvaluatePayloadMatcher(context.Background(), matcher, payload) //nolint:errcheck // Benchmark test
	}
}
