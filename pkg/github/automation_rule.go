package github

import (
	"context"
	"encoding/json"
	"time"
)

// AutomationRule represents a complete automation rule for GitHub events
type AutomationRule struct {
	ID           string                 `json:"id" yaml:"id"`
	Name         string                 `json:"name" yaml:"name"`
	Description  string                 `json:"description" yaml:"description"`
	Organization string                 `json:"organization" yaml:"organization"`
	Enabled      bool                   `json:"enabled" yaml:"enabled"`
	Priority     int                    `json:"priority" yaml:"priority"` // Higher number = higher priority
	Conditions   AutomationConditions   `json:"conditions" yaml:"conditions"`
	Actions      []AutomationAction     `json:"actions" yaml:"actions"`
	Schedule     *AutomationSchedule    `json:"schedule,omitempty" yaml:"schedule,omitempty"`
	Metadata     AutomationRuleMetadata `json:"metadata" yaml:"metadata"`
	CreatedAt    time.Time              `json:"created_at" yaml:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" yaml:"updated_at"`
	CreatedBy    string                 `json:"created_by" yaml:"created_by"`
	Tags         map[string]string      `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// AutomationConditions defines the conditions that must be met for a rule to trigger
type AutomationConditions struct {
	// Event-based conditions
	EventTypes   []EventType   `json:"event_types,omitempty" yaml:"event_types,omitempty"`
	Actions      []EventAction `json:"actions,omitempty" yaml:"actions,omitempty"`
	Organization string        `json:"organization,omitempty" yaml:"organization,omitempty"`
	Repository   string        `json:"repository,omitempty" yaml:"repository,omitempty"`
	Sender       string        `json:"sender,omitempty" yaml:"sender,omitempty"`

	// Repository-based conditions
	RepositoryPatterns []string `json:"repository_patterns,omitempty" yaml:"repository_patterns,omitempty"`
	Languages          []string `json:"languages,omitempty" yaml:"languages,omitempty"`
	Topics             []string `json:"topics,omitempty" yaml:"topics,omitempty"`
	Visibility         []string `json:"visibility,omitempty" yaml:"visibility,omitempty"` // public, private, internal
	IsArchived         *bool    `json:"is_archived,omitempty" yaml:"is_archived,omitempty"`
	IsTemplate         *bool    `json:"is_template,omitempty" yaml:"is_template,omitempty"`

	// Content-based conditions
	BranchPatterns []string `json:"branch_patterns,omitempty" yaml:"branch_patterns,omitempty"`
	FilePatterns   []string `json:"file_patterns,omitempty" yaml:"file_patterns,omitempty"`
	PathPatterns   []string `json:"path_patterns,omitempty" yaml:"path_patterns,omitempty"`

	// Time-based conditions
	TimeRange     *TimeRange `json:"time_range,omitempty" yaml:"time_range,omitempty"`
	DaysOfWeek    []int      `json:"days_of_week,omitempty" yaml:"days_of_week,omitempty"`     // 0=Sunday, 1=Monday, etc.
	HoursOfDay    []int      `json:"hours_of_day,omitempty" yaml:"hours_of_day,omitempty"`     // 0-23
	BusinessHours bool       `json:"business_hours,omitempty" yaml:"business_hours,omitempty"` // 9-17 weekdays

	// Advanced conditions
	CustomFilters map[string]interface{} `json:"custom_filters,omitempty" yaml:"custom_filters,omitempty"`
	PayloadMatch  []PayloadMatcher       `json:"payload_match,omitempty" yaml:"payload_match,omitempty"`

	// Logical operators
	LogicalOperator ConditionOperator      `json:"logical_operator,omitempty" yaml:"logical_operator,omitempty"`
	SubConditions   []AutomationConditions `json:"sub_conditions,omitempty" yaml:"sub_conditions,omitempty"`
}

// PayloadMatcher defines conditions for matching against event payload
type PayloadMatcher struct {
	Path          string        `json:"path" yaml:"path"`         // JSONPath expression (e.g., "$.pull_request.title")
	Operator      MatchOperator `json:"operator" yaml:"operator"` // equals, contains, regex, etc.
	Value         interface{}   `json:"value" yaml:"value"`       // Value to match against
	CaseSensitive bool          `json:"case_sensitive,omitempty" yaml:"case_sensitive,omitempty"`
}

// ConditionOperator defines how multiple conditions are combined
type ConditionOperator string

const (
	ConditionOperatorAND ConditionOperator = "AND"
	ConditionOperatorOR  ConditionOperator = "OR"
	ConditionOperatorNOT ConditionOperator = "NOT"
)

// MatchOperator defines how payload matching is performed
type MatchOperator string

const (
	MatchOperatorEquals      MatchOperator = "equals"
	MatchOperatorNotEquals   MatchOperator = "not_equals"
	MatchOperatorContains    MatchOperator = "contains"
	MatchOperatorNotContains MatchOperator = "not_contains"
	MatchOperatorStartsWith  MatchOperator = "starts_with"
	MatchOperatorEndsWith    MatchOperator = "ends_with"
	MatchOperatorRegex       MatchOperator = "regex"
	MatchOperatorGreaterThan MatchOperator = "greater_than"
	MatchOperatorLessThan    MatchOperator = "less_than"
	MatchOperatorExists      MatchOperator = "exists"
	MatchOperatorNotExists   MatchOperator = "not_exists"
	MatchOperatorEmpty       MatchOperator = "empty"
	MatchOperatorNotEmpty    MatchOperator = "not_empty"
)

// AutomationAction defines an action to be executed when conditions are met
type AutomationAction struct {
	ID          string                 `json:"id" yaml:"id"`
	Type        ActionType             `json:"type" yaml:"type"`
	Name        string                 `json:"name" yaml:"name"`
	Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Enabled     bool                   `json:"enabled" yaml:"enabled"`
	Parameters  map[string]interface{} `json:"parameters" yaml:"parameters"`
	Timeout     time.Duration          `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	RetryPolicy *ActionRetryPolicy     `json:"retry_policy,omitempty" yaml:"retry_policy,omitempty"`
	OnFailure   ActionFailurePolicy    `json:"on_failure,omitempty" yaml:"on_failure,omitempty"`
}

// ActionType defines the type of action to be executed
type ActionType string

const (
	// Webhook actions
	ActionTypeWebhook     ActionType = "webhook"
	ActionTypeHTTPRequest ActionType = "http_request"

	// GitHub API actions
	ActionTypeCreateIssue    ActionType = "create_issue"
	ActionTypeCreatePR       ActionType = "create_pr"
	ActionTypeAddLabel       ActionType = "add_label"
	ActionTypeRemoveLabel    ActionType = "remove_label"
	ActionTypeAssignReviewer ActionType = "assign_reviewer"
	ActionTypeMergePR        ActionType = "merge_pr"
	ActionTypeClosePR        ActionType = "close_pr"
	ActionTypeCloseIssue     ActionType = "close_issue"

	// Repository actions
	ActionTypeCreateBranch  ActionType = "create_branch"
	ActionTypeDeleteBranch  ActionType = "delete_branch"
	ActionTypeProtectBranch ActionType = "protect_branch"
	ActionTypeCreateTag     ActionType = "create_tag"
	ActionTypeCreateRelease ActionType = "create_release"

	// Notification actions
	ActionTypeSlackMessage ActionType = "slack_message"
	ActionTypeTeamsMessage ActionType = "teams_message"
	ActionTypeEmail        ActionType = "email"
	ActionTypeSMS          ActionType = "sms"

	// Workflow actions
	ActionTypeTriggerWorkflow ActionType = "trigger_workflow"
	ActionTypeRunScript       ActionType = "run_script"
	ActionTypeDeployment      ActionType = "deployment"

	// Custom actions
	ActionTypeCustom ActionType = "custom"
)

// ActionRetryPolicy defines retry behavior for failed actions
type ActionRetryPolicy struct {
	MaxRetries    int           `json:"max_retries" yaml:"max_retries"`
	RetryInterval time.Duration `json:"retry_interval" yaml:"retry_interval"`
	BackoffFactor float64       `json:"backoff_factor,omitempty" yaml:"backoff_factor,omitempty"`
	MaxInterval   time.Duration `json:"max_interval,omitempty" yaml:"max_interval,omitempty"`
}

// ActionFailurePolicy defines what to do when an action fails
type ActionFailurePolicy string

const (
	ActionFailurePolicyStop     ActionFailurePolicy = "stop"     // Stop processing remaining actions
	ActionFailurePolicyContinue ActionFailurePolicy = "continue" // Continue with remaining actions
	ActionFailurePolicyRetry    ActionFailurePolicy = "retry"    // Retry the failed action
	ActionFailurePolicySkip     ActionFailurePolicy = "skip"     // Skip and mark as failed
)

// AutomationSchedule defines when a rule should be evaluated (for scheduled rules)
type AutomationSchedule struct {
	Type       ScheduleType `json:"type" yaml:"type"`
	Expression string       `json:"expression" yaml:"expression"` // Cron expression
	Timezone   string       `json:"timezone,omitempty" yaml:"timezone,omitempty"`
	StartDate  *time.Time   `json:"start_date,omitempty" yaml:"start_date,omitempty"`
	EndDate    *time.Time   `json:"end_date,omitempty" yaml:"end_date,omitempty"`
}

// ScheduleType defines the type of schedule
type ScheduleType string

const (
	ScheduleTypeCron     ScheduleType = "cron"
	ScheduleTypeInterval ScheduleType = "interval"
	ScheduleTypeOneTime  ScheduleType = "one_time"
)

// AutomationRuleMetadata contains metadata about the rule
type AutomationRuleMetadata struct {
	Version        string            `json:"version" yaml:"version"`
	Category       string            `json:"category,omitempty" yaml:"category,omitempty"`
	Environment    string            `json:"environment,omitempty" yaml:"environment,omitempty"`
	Owner          string            `json:"owner,omitempty" yaml:"owner,omitempty"`
	Team           string            `json:"team,omitempty" yaml:"team,omitempty"`
	Documentation  string            `json:"documentation,omitempty" yaml:"documentation,omitempty"`
	ExamplePayload json.RawMessage   `json:"example_payload,omitempty" yaml:"example_payload,omitempty"`
	CustomMetadata map[string]string `json:"custom_metadata,omitempty" yaml:"custom_metadata,omitempty"`
}

// AutomationRuleExecution represents an execution instance of an automation rule
type AutomationRuleExecution struct {
	ID             string                     `json:"id"`
	RuleID         string                     `json:"rule_id"`
	TriggerEventID string                     `json:"trigger_event_id,omitempty"`
	StartedAt      time.Time                  `json:"started_at"`
	CompletedAt    *time.Time                 `json:"completed_at,omitempty"`
	Status         ExecutionStatus            `json:"status"`
	TriggerType    ExecutionTriggerType       `json:"trigger_type"`
	Context        AutomationExecutionContext `json:"context"`
	Actions        []ActionExecutionResult    `json:"actions"`
	Error          string                     `json:"error,omitempty"`
	Duration       time.Duration              `json:"duration,omitempty"`
	Metadata       map[string]interface{}     `json:"metadata,omitempty"`
}

// ExecutionStatus defines the status of a rule execution
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
	ExecutionStatusTimeout   ExecutionStatus = "timeout"
)

// ExecutionTriggerType defines what triggered the rule execution
type ExecutionTriggerType string

const (
	ExecutionTriggerTypeEvent    ExecutionTriggerType = "event"
	ExecutionTriggerTypeSchedule ExecutionTriggerType = "schedule"
	ExecutionTriggerTypeManual   ExecutionTriggerType = "manual"
	ExecutionTriggerTypeAPI      ExecutionTriggerType = "api"
)

// AutomationExecutionContext provides context for rule execution
type AutomationExecutionContext struct {
	Event        *GitHubEvent           `json:"event,omitempty"`
	Repository   *RepositoryInfo        `json:"repository,omitempty"`
	Organization string                 `json:"organization,omitempty"`
	User         string                 `json:"user,omitempty"`
	Variables    map[string]interface{} `json:"variables,omitempty"`
	Environment  string                 `json:"environment,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ActionExecutionResult represents the result of executing a single action
type ActionExecutionResult struct {
	ActionID    string                 `json:"action_id"`
	ActionType  ActionType             `json:"action_type"`
	Status      ExecutionStatus        `json:"status"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Duration    time.Duration          `json:"duration,omitempty"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	RetryCount  int                    `json:"retry_count,omitempty"`
}

// AutomationRuleSet represents a collection of related automation rules
type AutomationRuleSet struct {
	ID           string            `json:"id" yaml:"id"`
	Name         string            `json:"name" yaml:"name"`
	Description  string            `json:"description" yaml:"description"`
	Organization string            `json:"organization" yaml:"organization"`
	Rules        []AutomationRule  `json:"rules" yaml:"rules"`
	Enabled      bool              `json:"enabled" yaml:"enabled"`
	Tags         map[string]string `json:"tags,omitempty" yaml:"tags,omitempty"`
	CreatedAt    time.Time         `json:"created_at" yaml:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at" yaml:"updated_at"`
	CreatedBy    string            `json:"created_by" yaml:"created_by"`
}

// AutomationRuleTemplate represents a reusable rule template
type AutomationRuleTemplate struct {
	ID          string             `json:"id" yaml:"id"`
	Name        string             `json:"name" yaml:"name"`
	Description string             `json:"description" yaml:"description"`
	Category    string             `json:"category" yaml:"category"`
	Template    AutomationRule     `json:"template" yaml:"template"`
	Variables   []TemplateVariable `json:"variables" yaml:"variables"`
	Examples    []TemplateExample  `json:"examples,omitempty" yaml:"examples,omitempty"`
	Tags        map[string]string  `json:"tags,omitempty" yaml:"tags,omitempty"`
	CreatedAt   time.Time          `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" yaml:"updated_at"`
	CreatedBy   string             `json:"created_by" yaml:"created_by"`
}

// TemplateVariable defines a variable that can be customized in a template
type TemplateVariable struct {
	Name         string      `json:"name" yaml:"name"`
	Type         string      `json:"type" yaml:"type"` // string, number, boolean, array, object
	Description  string      `json:"description" yaml:"description"`
	Required     bool        `json:"required" yaml:"required"`
	DefaultValue interface{} `json:"default_value,omitempty" yaml:"default_value,omitempty"`
	Options      []string    `json:"options,omitempty" yaml:"options,omitempty"`
	Validation   string      `json:"validation,omitempty" yaml:"validation,omitempty"` // Regex or validation rule
}

// TemplateExample provides example configurations for a template
type TemplateExample struct {
	Name        string                 `json:"name" yaml:"name"`
	Description string                 `json:"description" yaml:"description"`
	Variables   map[string]interface{} `json:"variables" yaml:"variables"`
}

// AutomationRuleService defines the interface for managing automation rules
type AutomationRuleService interface {
	// Rule Management
	CreateRule(ctx context.Context, rule *AutomationRule) error
	GetRule(ctx context.Context, org, ruleID string) (*AutomationRule, error)
	ListRules(ctx context.Context, org string, filter *RuleFilter) ([]*AutomationRule, error)
	UpdateRule(ctx context.Context, rule *AutomationRule) error
	DeleteRule(ctx context.Context, org, ruleID string) error
	EnableRule(ctx context.Context, org, ruleID string) error
	DisableRule(ctx context.Context, org, ruleID string) error

	// Rule Evaluation
	EvaluateConditions(ctx context.Context, rule *AutomationRule, event *GitHubEvent) (bool, error)
	ExecuteRule(ctx context.Context, rule *AutomationRule, context *AutomationExecutionContext) (*AutomationRuleExecution, error)

	// Rule Sets
	CreateRuleSet(ctx context.Context, ruleSet *AutomationRuleSet) error
	GetRuleSet(ctx context.Context, org, setID string) (*AutomationRuleSet, error)
	ListRuleSets(ctx context.Context, org string) ([]*AutomationRuleSet, error)
	UpdateRuleSet(ctx context.Context, ruleSet *AutomationRuleSet) error
	DeleteRuleSet(ctx context.Context, org, setID string) error

	// Templates
	CreateTemplate(ctx context.Context, template *AutomationRuleTemplate) error
	GetTemplate(ctx context.Context, templateID string) (*AutomationRuleTemplate, error)
	ListTemplates(ctx context.Context, category string) ([]*AutomationRuleTemplate, error)
	UpdateTemplate(ctx context.Context, template *AutomationRuleTemplate) error
	DeleteTemplate(ctx context.Context, templateID string) error
	InstantiateTemplate(ctx context.Context, templateID string, variables map[string]interface{}) (*AutomationRule, error)

	// Execution History
	GetExecution(ctx context.Context, executionID string) (*AutomationRuleExecution, error)
	ListExecutions(ctx context.Context, org string, filter *ExecutionFilter) ([]*AutomationRuleExecution, error)
	CancelExecution(ctx context.Context, executionID string) error

	// Validation and Testing
	ValidateRule(ctx context.Context, rule *AutomationRule) (*RuleValidationResult, error)
	TestRule(ctx context.Context, rule *AutomationRule, testEvent *GitHubEvent) (*RuleTestResult, error)
	DryRunRule(ctx context.Context, ruleID string, event *GitHubEvent) (*RuleTestResult, error)
}

// RuleFilter defines criteria for filtering automation rules
type RuleFilter struct {
	Organization  string      `json:"organization,omitempty"`
	Enabled       *bool       `json:"enabled,omitempty"`
	Tags          []string    `json:"tags,omitempty"`
	Category      string      `json:"category,omitempty"`
	EventTypes    []EventType `json:"event_types,omitempty"`
	CreatedBy     string      `json:"created_by,omitempty"`
	CreatedAfter  *time.Time  `json:"created_after,omitempty"`
	CreatedBefore *time.Time  `json:"created_before,omitempty"`
}

// ExecutionFilter defines criteria for filtering rule executions
type ExecutionFilter struct {
	RuleID        string               `json:"rule_id,omitempty"`
	Status        ExecutionStatus      `json:"status,omitempty"`
	TriggerType   ExecutionTriggerType `json:"trigger_type,omitempty"`
	StartedAfter  *time.Time           `json:"started_after,omitempty"`
	StartedBefore *time.Time           `json:"started_before,omitempty"`
}

// RuleValidationResult represents the result of rule validation
type RuleValidationResult struct {
	Valid    bool                    `json:"valid"`
	Errors   []RuleValidationError   `json:"errors,omitempty"`
	Warnings []RuleValidationWarning `json:"warnings,omitempty"`
	Score    int                     `json:"score"` // 0-100
}

// RuleValidationError represents a validation error
type RuleValidationError struct {
	Field      string `json:"field"`
	Message    string `json:"message"`
	Severity   string `json:"severity"`
	Suggestion string `json:"suggestion,omitempty"`
}

// RuleValidationWarning represents a validation warning
type RuleValidationWarning struct {
	Field      string `json:"field"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// RuleTestResult represents the result of testing a rule
type RuleTestResult struct {
	RuleID            string                     `json:"rule_id"`
	ConditionsMatched bool                       `json:"conditions_matched"`
	ActionsExecuted   []ActionExecutionResult    `json:"actions_executed"`
	ExecutionTime     time.Duration              `json:"execution_time"`
	Errors            []string                   `json:"errors,omitempty"`
	Context           AutomationExecutionContext `json:"context"`
}
