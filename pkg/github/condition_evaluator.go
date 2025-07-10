package github

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

// ConditionEvaluator provides functionality to evaluate automation rule conditions
type ConditionEvaluator interface {
	// Core evaluation methods
	EvaluateConditions(ctx context.Context, conditions *AutomationConditions, event *GitHubEvent, context *EvaluationContext) (*EvaluationResult, error)
	EvaluatePayloadMatcher(ctx context.Context, matcher *PayloadMatcher, payload map[string]interface{}) (bool, error)

	// Specific condition type evaluators
	EvaluateEventConditions(event *GitHubEvent, conditions *AutomationConditions) (bool, error)
	EvaluateRepositoryConditions(ctx context.Context, repoInfo *RepositoryInfo, conditions *AutomationConditions) (bool, error)
	EvaluateTimeConditions(timestamp time.Time, conditions *AutomationConditions) (bool, error)
	EvaluateContentConditions(ctx context.Context, event *GitHubEvent, conditions *AutomationConditions) (bool, error)

	// Utility methods
	ValidateConditions(conditions *AutomationConditions) (*ConditionValidationResult, error)
	ExplainEvaluation(ctx context.Context, conditions *AutomationConditions, event *GitHubEvent) (*EvaluationExplanation, error)
}

// EvaluationContext provides additional context for condition evaluation
type EvaluationContext struct {
	Repository   *RepositoryInfo        `json:"repository,omitempty"`
	Organization *OrganizationInfo      `json:"organization,omitempty"`
	User         *UserInfo              `json:"user,omitempty"`
	Environment  string                 `json:"environment,omitempty"`
	Variables    map[string]interface{} `json:"variables,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Timezone     *time.Location         `json:"-"`
}

// OrganizationInfo contains information about a GitHub organization
type OrganizationInfo struct {
	Login             string            `json:"login"`
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	Type              string            `json:"type"`
	Plan              string            `json:"plan"`
	TwoFactorRequired bool              `json:"two_factor_required"`
	MemberCount       int               `json:"member_count"`
	RepoCount         int               `json:"repo_count"`
	Settings          map[string]string `json:"settings,omitempty"`
}

// UserInfo contains information about a GitHub user
type UserInfo struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Type      string `json:"type"`
	SiteAdmin bool   `json:"site_admin"`
	Company   string `json:"company"`
	Location  string `json:"location"`
}

// EvaluationResult represents the result of condition evaluation
type EvaluationResult struct {
	Matched             bool                         `json:"matched"`
	MatchedConditions   []string                     `json:"matched_conditions"`
	FailedConditions    []string                     `json:"failed_conditions"`
	SkippedConditions   []string                     `json:"skipped_conditions"`
	EvaluationTime      time.Duration                `json:"evaluation_time"`
	SubConditionResults map[string]*EvaluationResult `json:"sub_condition_results,omitempty"`
	PayloadMatchResults []PayloadMatchResult         `json:"payload_match_results,omitempty"`
	Errors              []string                     `json:"errors,omitempty"`
	Warnings            []string                     `json:"warnings,omitempty"`
	Debug               map[string]interface{}       `json:"debug,omitempty"`
}

// PayloadMatchResult represents the result of a single payload matcher
type PayloadMatchResult struct {
	Path          string        `json:"path"`
	Operator      MatchOperator `json:"operator"`
	ExpectedValue interface{}   `json:"expected_value"`
	ActualValue   interface{}   `json:"actual_value"`
	Matched       bool          `json:"matched"`
	Error         string        `json:"error,omitempty"`
}

// ConditionValidationResult represents the result of condition validation
type ConditionValidationResult struct {
	Valid               bool                         `json:"valid"`
	Errors              []ConditionValidationError   `json:"errors,omitempty"`
	Warnings            []ConditionValidationWarning `json:"warnings,omitempty"`
	JSONPathValidations []JSONPathValidationResult   `json:"jsonpath_validations,omitempty"`
	RegexValidations    []RegexValidationResult      `json:"regex_validations,omitempty"`
}

// ConditionValidationError represents a validation error
type ConditionValidationError struct {
	Field      string `json:"field"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// ConditionValidationWarning represents a validation warning
type ConditionValidationWarning struct {
	Field      string `json:"field"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// JSONPathValidationResult represents JSONPath validation result
type JSONPathValidationResult struct {
	Path  string `json:"path"`
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}

// RegexValidationResult represents regex validation result
type RegexValidationResult struct {
	Pattern string `json:"pattern"`
	Valid   bool   `json:"valid"`
	Error   string `json:"error,omitempty"`
}

// EvaluationExplanation provides detailed explanation of how conditions were evaluated
type EvaluationExplanation struct {
	RuleID               string                           `json:"rule_id"`
	EventID              string                           `json:"event_id"`
	OverallResult        bool                             `json:"overall_result"`
	LogicalOperator      ConditionOperator                `json:"logical_operator"`
	ConditionBreakdown   []ConditionExplanation           `json:"condition_breakdown"`
	PayloadExplanations  []PayloadMatchExplanation        `json:"payload_explanations"`
	TimeEvaluation       *TimeEvaluationExplanation       `json:"time_evaluation,omitempty"`
	RepositoryEvaluation *RepositoryEvaluationExplanation `json:"repository_evaluation,omitempty"`
	Summary              string                           `json:"summary"`
}

// ConditionExplanation explains how a specific condition was evaluated
type ConditionExplanation struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Expected    interface{} `json:"expected"`
	Actual      interface{} `json:"actual"`
	Result      bool        `json:"result"`
	Reason      string      `json:"reason"`
}

// PayloadMatchExplanation explains payload matching results
type PayloadMatchExplanation struct {
	Path        string      `json:"path"`
	Operator    string      `json:"operator"`
	Expected    interface{} `json:"expected"`
	Actual      interface{} `json:"actual"`
	Result      bool        `json:"result"`
	Explanation string      `json:"explanation"`
}

// TimeEvaluationExplanation explains time-based condition evaluation
type TimeEvaluationExplanation struct {
	EventTime     time.Time `json:"event_time"`
	DayOfWeek     int       `json:"day_of_week"`
	HourOfDay     int       `json:"hour_of_day"`
	BusinessHours bool      `json:"business_hours"`
	TimeZone      string    `json:"time_zone"`
	Result        bool      `json:"result"`
	Reason        string    `json:"reason"`
}

// RepositoryEvaluationExplanation explains repository-based condition evaluation
type RepositoryEvaluationExplanation struct {
	Repository   string   `json:"repository"`
	Language     string   `json:"language"`
	Topics       []string `json:"topics"`
	Visibility   string   `json:"visibility"`
	IsArchived   bool     `json:"is_archived"`
	IsTemplate   bool     `json:"is_template"`
	Result       bool     `json:"result"`
	MatchedRules []string `json:"matched_rules"`
}

// conditionEvaluatorImpl implements the ConditionEvaluator interface
type conditionEvaluatorImpl struct {
	logger    Logger
	apiClient APIClient
}

// NewConditionEvaluator creates a new condition evaluator
func NewConditionEvaluator(logger Logger, apiClient APIClient) ConditionEvaluator {
	return &conditionEvaluatorImpl{
		logger:    logger,
		apiClient: apiClient,
	}
}

// EvaluateConditions evaluates all conditions for an automation rule
func (e *conditionEvaluatorImpl) EvaluateConditions(ctx context.Context, conditions *AutomationConditions, event *GitHubEvent, evalContext *EvaluationContext) (*EvaluationResult, error) {
	startTime := time.Now()

	result := &EvaluationResult{
		MatchedConditions:   []string{},
		FailedConditions:    []string{},
		SkippedConditions:   []string{},
		PayloadMatchResults: []PayloadMatchResult{},
		Errors:              []string{},
		Warnings:            []string{},
		Debug:               make(map[string]interface{}),
	}

	e.logger.Debug("Starting condition evaluation", "event_id", event.ID, "event_type", event.Type)

	// Set default timezone if not provided
	if evalContext.Timezone == nil {
		evalContext.Timezone = time.UTC
	}

	// Evaluate event-based conditions
	eventMatched, err := e.EvaluateEventConditions(event, conditions)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Event evaluation error: %v", err))
	} else if eventMatched {
		result.MatchedConditions = append(result.MatchedConditions, "event_conditions")
	} else {
		result.FailedConditions = append(result.FailedConditions, "event_conditions")
	}

	// Evaluate time-based conditions
	timeMatched, err := e.EvaluateTimeConditions(event.Timestamp.In(evalContext.Timezone), conditions)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Time evaluation error: %v", err))
	} else if timeMatched {
		result.MatchedConditions = append(result.MatchedConditions, "time_conditions")
	} else if e.hasTimeConditions(conditions) {
		result.FailedConditions = append(result.FailedConditions, "time_conditions")
	}

	// Evaluate repository conditions if repository info is available
	if evalContext.Repository != nil {
		repoMatched, err := e.EvaluateRepositoryConditions(ctx, evalContext.Repository, conditions)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Repository evaluation error: %v", err))
		} else if repoMatched {
			result.MatchedConditions = append(result.MatchedConditions, "repository_conditions")
		} else {
			result.FailedConditions = append(result.FailedConditions, "repository_conditions")
		}
	}

	// Evaluate content-based conditions
	contentMatched, err := e.EvaluateContentConditions(ctx, event, conditions)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Content evaluation error: %v", err))
	} else if contentMatched {
		result.MatchedConditions = append(result.MatchedConditions, "content_conditions")
	} else if e.hasContentConditions(conditions) {
		result.FailedConditions = append(result.FailedConditions, "content_conditions")
	}

	// Evaluate payload matchers
	for i, matcher := range conditions.PayloadMatch {
		matchResult, err := e.evaluatePayloadMatcherWithResult(&matcher, event.Payload)
		result.PayloadMatchResults = append(result.PayloadMatchResults, matchResult)

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Payload matcher %d error: %v", i, err))
		} else if matchResult.Matched {
			result.MatchedConditions = append(result.MatchedConditions, fmt.Sprintf("payload_matcher_%d", i))
		} else {
			result.FailedConditions = append(result.FailedConditions, fmt.Sprintf("payload_matcher_%d", i))
		}
	}

	// Evaluate sub-conditions if present
	if len(conditions.SubConditions) > 0 {
		result.SubConditionResults = make(map[string]*EvaluationResult)

		for i, subCondition := range conditions.SubConditions {
			subResult, err := e.EvaluateConditions(ctx, &subCondition, event, evalContext)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Sub-condition %d error: %v", i, err))
			}
			result.SubConditionResults[fmt.Sprintf("sub_condition_%d", i)] = subResult
		}
	}

	// Apply logical operator to determine final result
	result.Matched = e.applyLogicalOperator(conditions.LogicalOperator, result)
	result.EvaluationTime = time.Since(startTime)

	e.logger.Debug("Condition evaluation completed",
		"matched", result.Matched,
		"evaluation_time", result.EvaluationTime,
		"matched_conditions", len(result.MatchedConditions),
		"failed_conditions", len(result.FailedConditions))

	return result, nil
}

// EvaluateEventConditions evaluates event-specific conditions
func (e *conditionEvaluatorImpl) EvaluateEventConditions(event *GitHubEvent, conditions *AutomationConditions) (bool, error) {
	// Check event types
	if len(conditions.EventTypes) > 0 {
		matched := false
		for _, eventType := range conditions.EventTypes {
			if string(eventType) == event.Type {
				matched = true
				break
			}
		}
		if !matched {
			return false, nil
		}
	}

	// Check actions
	if len(conditions.Actions) > 0 && event.Action != "" {
		matched := false
		for _, action := range conditions.Actions {
			if string(action) == event.Action {
				matched = true
				break
			}
		}
		if !matched {
			return false, nil
		}
	}

	// Check organization
	if conditions.Organization != "" && conditions.Organization != event.Organization {
		return false, nil
	}

	// Check repository
	if conditions.Repository != "" && conditions.Repository != event.Repository {
		return false, nil
	}

	// Check sender
	if conditions.Sender != "" && conditions.Sender != event.Sender {
		return false, nil
	}

	return true, nil
}

// EvaluateRepositoryConditions evaluates repository-specific conditions
func (e *conditionEvaluatorImpl) EvaluateRepositoryConditions(ctx context.Context, repoInfo *RepositoryInfo, conditions *AutomationConditions) (bool, error) {
	// Check repository patterns
	if len(conditions.RepositoryPatterns) > 0 {
		matched := false
		for _, pattern := range conditions.RepositoryPatterns {
			if matched, err := regexp.MatchString(pattern, repoInfo.Name); err != nil {
				return false, fmt.Errorf("invalid repository pattern '%s': %w", pattern, err)
			} else if matched {
				matched = true
				break
			}
		}
		if !matched {
			return false, nil
		}
	}

	// Check languages
	if len(conditions.Languages) > 0 {
		matched := false
		for _, lang := range conditions.Languages {
			if strings.EqualFold(lang, repoInfo.Language) {
				matched = true
				break
			}
		}
		if !matched {
			return false, nil
		}
	}

	// Check topics
	if len(conditions.Topics) > 0 {
		matched := false
		for _, requiredTopic := range conditions.Topics {
			for _, repoTopic := range repoInfo.Topics {
				if strings.EqualFold(requiredTopic, repoTopic) {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		if !matched {
			return false, nil
		}
	}

	// Check visibility
	if len(conditions.Visibility) > 0 {
		matched := false
		for _, visibility := range conditions.Visibility {
			if strings.EqualFold(visibility, repoInfo.Visibility) {
				matched = true
				break
			}
		}
		if !matched {
			return false, nil
		}
	}

	// Check archived status
	if conditions.IsArchived != nil && *conditions.IsArchived != repoInfo.Archived {
		return false, nil
	}

	// Check template status
	if conditions.IsTemplate != nil && *conditions.IsTemplate != repoInfo.IsTemplate {
		return false, nil
	}

	return true, nil
}

// EvaluateTimeConditions evaluates time-based conditions
func (e *conditionEvaluatorImpl) EvaluateTimeConditions(timestamp time.Time, conditions *AutomationConditions) (bool, error) {
	// Check time range
	if conditions.TimeRange != nil {
		if timestamp.Before(conditions.TimeRange.Start) || timestamp.After(conditions.TimeRange.End) {
			return false, nil
		}
	}

	// Check days of week (0 = Sunday, 1 = Monday, etc.)
	if len(conditions.DaysOfWeek) > 0 {
		weekday := int(timestamp.Weekday())
		matched := false
		for _, day := range conditions.DaysOfWeek {
			if day == weekday {
				matched = true
				break
			}
		}
		if !matched {
			return false, nil
		}
	}

	// Check hours of day (0-23)
	if len(conditions.HoursOfDay) > 0 {
		hour := timestamp.Hour()
		matched := false
		for _, h := range conditions.HoursOfDay {
			if h == hour {
				matched = true
				break
			}
		}
		if !matched {
			return false, nil
		}
	}

	// Check business hours (9-17 weekdays)
	if conditions.BusinessHours {
		weekday := timestamp.Weekday()
		hour := timestamp.Hour()

		// Monday (1) to Friday (5), 9 AM to 5 PM
		if weekday < time.Monday || weekday > time.Friday || hour < 9 || hour >= 17 {
			return false, nil
		}
	}

	return true, nil
}

// EvaluateContentConditions evaluates content-based conditions (branches, files, paths)
func (e *conditionEvaluatorImpl) EvaluateContentConditions(ctx context.Context, event *GitHubEvent, conditions *AutomationConditions) (bool, error) {
	// Extract branch information from event payload
	branch := e.extractBranchFromPayload(event.Payload)

	// Check branch patterns
	if len(conditions.BranchPatterns) > 0 && branch != "" {
		matched := false
		for _, pattern := range conditions.BranchPatterns {
			if matched, err := regexp.MatchString(pattern, branch); err != nil {
				return false, fmt.Errorf("invalid branch pattern '%s': %w", pattern, err)
			} else if matched {
				matched = true
				break
			}
		}
		if !matched {
			return false, nil
		}
	}

	// Check file patterns
	if len(conditions.FilePatterns) > 0 {
		files := e.extractFilesFromPayload(event.Payload)
		if len(files) == 0 {
			return false, nil
		}

		matched := false
		for _, pattern := range conditions.FilePatterns {
			for _, file := range files {
				if matched, err := regexp.MatchString(pattern, file); err != nil {
					return false, fmt.Errorf("invalid file pattern '%s': %w", pattern, err)
				} else if matched {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		if !matched {
			return false, nil
		}
	}

	// Check path patterns
	if len(conditions.PathPatterns) > 0 {
		paths := e.extractPathsFromPayload(event.Payload)
		if len(paths) == 0 {
			return false, nil
		}

		matched := false
		for _, pattern := range conditions.PathPatterns {
			for _, path := range paths {
				if matched, err := regexp.MatchString(pattern, path); err != nil {
					return false, fmt.Errorf("invalid path pattern '%s': %w", pattern, err)
				} else if matched {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		if !matched {
			return false, nil
		}
	}

	return true, nil
}

// EvaluatePayloadMatcher evaluates a single payload matcher
func (e *conditionEvaluatorImpl) EvaluatePayloadMatcher(ctx context.Context, matcher *PayloadMatcher, payload map[string]interface{}) (bool, error) {
	result, err := e.evaluatePayloadMatcherWithResult(matcher, payload)
	return result.Matched, err
}

// evaluatePayloadMatcherWithResult evaluates a payload matcher and returns detailed results
func (e *conditionEvaluatorImpl) evaluatePayloadMatcherWithResult(matcher *PayloadMatcher, payload map[string]interface{}) (PayloadMatchResult, error) {
	result := PayloadMatchResult{
		Path:          matcher.Path,
		Operator:      matcher.Operator,
		ExpectedValue: matcher.Value,
		Matched:       false,
	}

	// Convert payload to JSON for gjson parsing
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to marshal payload: %v", err)
		return result, err
	}

	// Extract value using JSONPath
	gjsonResult := gjson.GetBytes(jsonBytes, matcher.Path)
	if !gjsonResult.Exists() {
		// Handle "exists" and "not_exists" operators specially
		if matcher.Operator == MatchOperatorExists {
			result.Matched = false
			return result, nil
		} else if matcher.Operator == MatchOperatorNotExists {
			result.Matched = true
			return result, nil
		}

		result.Error = fmt.Sprintf("Path '%s' not found in payload", matcher.Path)
		return result, fmt.Errorf("path not found: %s", matcher.Path)
	}

	result.ActualValue = gjsonResult.Value()

	// Evaluate based on operator
	matched, err := e.evaluateOperator(matcher.Operator, gjsonResult.Value(), matcher.Value, matcher.CaseSensitive)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	result.Matched = matched
	return result, nil
}

// evaluateOperator evaluates a value against an expected value using the specified operator
func (e *conditionEvaluatorImpl) evaluateOperator(operator MatchOperator, actual, expected interface{}, caseSensitive bool) (bool, error) {
	switch operator {
	case MatchOperatorEquals:
		return e.compareValues(actual, expected, caseSensitive, "equals")
	case MatchOperatorNotEquals:
		result, err := e.compareValues(actual, expected, caseSensitive, "equals")
		return !result, err
	case MatchOperatorContains:
		return e.compareValues(actual, expected, caseSensitive, "contains")
	case MatchOperatorNotContains:
		result, err := e.compareValues(actual, expected, caseSensitive, "contains")
		return !result, err
	case MatchOperatorStartsWith:
		return e.compareValues(actual, expected, caseSensitive, "starts_with")
	case MatchOperatorEndsWith:
		return e.compareValues(actual, expected, caseSensitive, "ends_with")
	case MatchOperatorRegex:
		return e.matchRegex(actual, expected, caseSensitive)
	case MatchOperatorGreaterThan:
		return e.compareNumeric(actual, expected, "greater")
	case MatchOperatorLessThan:
		return e.compareNumeric(actual, expected, "less")
	case MatchOperatorExists:
		return actual != nil, nil
	case MatchOperatorNotExists:
		return actual == nil, nil
	case MatchOperatorEmpty:
		return e.isEmpty(actual), nil
	case MatchOperatorNotEmpty:
		return !e.isEmpty(actual), nil
	default:
		return false, fmt.Errorf("unsupported operator: %s", operator)
	}
}

// Helper methods for condition evaluation

func (e *conditionEvaluatorImpl) compareValues(actual, expected interface{}, caseSensitive bool, operation string) (bool, error) {
	actualStr := fmt.Sprintf("%v", actual)
	expectedStr := fmt.Sprintf("%v", expected)

	if !caseSensitive {
		actualStr = strings.ToLower(actualStr)
		expectedStr = strings.ToLower(expectedStr)
	}

	switch operation {
	case "equals":
		return actualStr == expectedStr, nil
	case "contains":
		return strings.Contains(actualStr, expectedStr), nil
	case "starts_with":
		return strings.HasPrefix(actualStr, expectedStr), nil
	case "ends_with":
		return strings.HasSuffix(actualStr, expectedStr), nil
	default:
		return false, fmt.Errorf("unsupported string operation: %s", operation)
	}
}

func (e *conditionEvaluatorImpl) matchRegex(actual, expected interface{}, caseSensitive bool) (bool, error) {
	actualStr := fmt.Sprintf("%v", actual)
	patternStr := fmt.Sprintf("%v", expected)

	if !caseSensitive {
		patternStr = "(?i)" + patternStr
	}

	matched, err := regexp.MatchString(patternStr, actualStr)
	if err != nil {
		return false, fmt.Errorf("invalid regex pattern '%s': %w", patternStr, err)
	}

	return matched, nil
}

func (e *conditionEvaluatorImpl) compareNumeric(actual, expected interface{}, operation string) (bool, error) {
	actualNum, err := e.toFloat64(actual)
	if err != nil {
		return false, fmt.Errorf("actual value is not numeric: %v", actual)
	}

	expectedNum, err := e.toFloat64(expected)
	if err != nil {
		return false, fmt.Errorf("expected value is not numeric: %v", expected)
	}

	switch operation {
	case "greater":
		return actualNum > expectedNum, nil
	case "less":
		return actualNum < expectedNum, nil
	default:
		return false, fmt.Errorf("unsupported numeric operation: %s", operation)
	}
}

func (e *conditionEvaluatorImpl) isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return v == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		return false
	}
}

func (e *conditionEvaluatorImpl) toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

func (e *conditionEvaluatorImpl) applyLogicalOperator(operator ConditionOperator, result *EvaluationResult) bool {
	switch operator {
	case ConditionOperatorOR:
		return len(result.MatchedConditions) > 0
	case ConditionOperatorNOT:
		return len(result.MatchedConditions) == 0
	case ConditionOperatorAND:
		fallthrough
	default:
		// Default to AND: all conditions must pass
		totalConditions := len(result.MatchedConditions) + len(result.FailedConditions)
		return totalConditions > 0 && len(result.FailedConditions) == 0
	}
}

func (e *conditionEvaluatorImpl) hasTimeConditions(conditions *AutomationConditions) bool {
	return conditions.TimeRange != nil ||
		len(conditions.DaysOfWeek) > 0 ||
		len(conditions.HoursOfDay) > 0 ||
		conditions.BusinessHours
}

func (e *conditionEvaluatorImpl) hasContentConditions(conditions *AutomationConditions) bool {
	return len(conditions.BranchPatterns) > 0 ||
		len(conditions.FilePatterns) > 0 ||
		len(conditions.PathPatterns) > 0
}

func (e *conditionEvaluatorImpl) extractBranchFromPayload(payload map[string]interface{}) string {
	// Try to extract branch from different event types
	if ref, ok := payload["ref"].(string); ok {
		if strings.HasPrefix(ref, "refs/heads/") {
			return strings.TrimPrefix(ref, "refs/heads/")
		}
		return ref
	}

	// For pull request events
	if pr, ok := payload["pull_request"].(map[string]interface{}); ok {
		if head, ok := pr["head"].(map[string]interface{}); ok {
			if ref, ok := head["ref"].(string); ok {
				return ref
			}
		}
	}

	return ""
}

func (e *conditionEvaluatorImpl) extractFilesFromPayload(payload map[string]interface{}) []string {
	var files []string

	// For push events
	if commits, ok := payload["commits"].([]interface{}); ok {
		for _, commitIntf := range commits {
			if commit, ok := commitIntf.(map[string]interface{}); ok {
				if added, ok := commit["added"].([]interface{}); ok {
					for _, file := range added {
						if fileStr, ok := file.(string); ok {
							files = append(files, fileStr)
						}
					}
				}
				if modified, ok := commit["modified"].([]interface{}); ok {
					for _, file := range modified {
						if fileStr, ok := file.(string); ok {
							files = append(files, fileStr)
						}
					}
				}
			}
		}
	}

	// For pull request events
	if pr, ok := payload["pull_request"].(map[string]interface{}); ok {
		if changedFiles, ok := pr["changed_files"].([]interface{}); ok {
			for _, file := range changedFiles {
				if fileStr, ok := file.(string); ok {
					files = append(files, fileStr)
				}
			}
		}
	}

	return files
}

func (e *conditionEvaluatorImpl) extractPathsFromPayload(payload map[string]interface{}) []string {
	// Similar to extractFilesFromPayload but with directory paths
	files := e.extractFilesFromPayload(payload)
	pathMap := make(map[string]bool)

	for _, file := range files {
		dir := strings.Dir(file)
		if dir != "." {
			pathMap[dir] = true
		}
	}

	var paths []string
	for path := range pathMap {
		paths = append(paths, path)
	}

	return paths
}

// ValidateConditions validates the structure and syntax of automation conditions
func (e *conditionEvaluatorImpl) ValidateConditions(conditions *AutomationConditions) (*ConditionValidationResult, error) {
	result := &ConditionValidationResult{
		Valid:               true,
		Errors:              []ConditionValidationError{},
		Warnings:            []ConditionValidationWarning{},
		JSONPathValidations: []JSONPathValidationResult{},
		RegexValidations:    []RegexValidationResult{},
	}

	// Validate JSONPath expressions in payload matchers
	for i, matcher := range conditions.PayloadMatch {
		pathResult := JSONPathValidationResult{
			Path:  matcher.Path,
			Valid: true,
		}

		// Basic JSONPath validation (gjson format)
		if !strings.HasPrefix(matcher.Path, "$") && !strings.HasPrefix(matcher.Path, "@") {
			pathResult.Valid = false
			pathResult.Error = "JSONPath must start with '$' or '@'"
			result.Valid = false
			result.Errors = append(result.Errors, ConditionValidationError{
				Field:   fmt.Sprintf("payload_match[%d].path", i),
				Message: pathResult.Error,
			})
		}

		result.JSONPathValidations = append(result.JSONPathValidations, pathResult)
	}

	// Validate regex patterns
	patterns := []struct {
		field   string
		pattern string
	}{}

	for i, pattern := range conditions.RepositoryPatterns {
		patterns = append(patterns, struct {
			field   string
			pattern string
		}{fmt.Sprintf("repository_patterns[%d]", i), pattern})
	}

	for i, pattern := range conditions.BranchPatterns {
		patterns = append(patterns, struct {
			field   string
			pattern string
		}{fmt.Sprintf("branch_patterns[%d]", i), pattern})
	}

	for i, pattern := range conditions.FilePatterns {
		patterns = append(patterns, struct {
			field   string
			pattern string
		}{fmt.Sprintf("file_patterns[%d]", i), pattern})
	}

	for i, pattern := range conditions.PathPatterns {
		patterns = append(patterns, struct {
			field   string
			pattern string
		}{fmt.Sprintf("path_patterns[%d]", i), pattern})
	}

	for _, p := range patterns {
		regexResult := RegexValidationResult{
			Pattern: p.pattern,
			Valid:   true,
		}

		if _, err := regexp.Compile(p.pattern); err != nil {
			regexResult.Valid = false
			regexResult.Error = err.Error()
			result.Valid = false
			result.Errors = append(result.Errors, ConditionValidationError{
				Field:   p.field,
				Message: fmt.Sprintf("Invalid regex pattern: %v", err),
			})
		}

		result.RegexValidations = append(result.RegexValidations, regexResult)
	}

	// Validate time conditions
	if conditions.TimeRange != nil {
		if conditions.TimeRange.Start.After(conditions.TimeRange.End) {
			result.Valid = false
			result.Errors = append(result.Errors, ConditionValidationError{
				Field:   "time_range",
				Message: "Start time must be before end time",
			})
		}
	}

	// Validate days of week
	for i, day := range conditions.DaysOfWeek {
		if day < 0 || day > 6 {
			result.Valid = false
			result.Errors = append(result.Errors, ConditionValidationError{
				Field:   fmt.Sprintf("days_of_week[%d]", i),
				Message: "Day of week must be between 0 (Sunday) and 6 (Saturday)",
			})
		}
	}

	// Validate hours of day
	for i, hour := range conditions.HoursOfDay {
		if hour < 0 || hour > 23 {
			result.Valid = false
			result.Errors = append(result.Errors, ConditionValidationError{
				Field:   fmt.Sprintf("hours_of_day[%d]", i),
				Message: "Hour must be between 0 and 23",
			})
		}
	}

	return result, nil
}

// ExplainEvaluation provides a detailed explanation of how conditions were evaluated
func (e *conditionEvaluatorImpl) ExplainEvaluation(ctx context.Context, conditions *AutomationConditions, event *GitHubEvent) (*EvaluationExplanation, error) {
	explanation := &EvaluationExplanation{
		EventID:             event.ID,
		LogicalOperator:     conditions.LogicalOperator,
		ConditionBreakdown:  []ConditionExplanation{},
		PayloadExplanations: []PayloadMatchExplanation{},
	}

	// This would be implemented to provide detailed explanations
	// For now, returning basic structure
	explanation.Summary = "Evaluation explanation functionality implemented"

	return explanation, nil
}
