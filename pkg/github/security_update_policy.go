package github

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// SecurityUpdatePolicyManager manages security update policies and vulnerability handling.
type SecurityUpdatePolicyManager struct {
	logger            Logger
	apiClient         APIClient
	dependabotManager *DependabotConfigManager
	policies          map[string]*SecurityUpdatePolicy
	vulnerabilityDB   *VulnerabilityDatabase
}

// SecurityUpdatePolicy defines policies for handling security updates.
type SecurityUpdatePolicy struct {
	ID                       string                   `json:"id"`
	Name                     string                   `json:"name"`
	Organization             string                   `json:"organization"`
	Description              string                   `json:"description"`
	Enabled                  bool                     `json:"enabled"`
	AutoApprovalRules        []AutoApprovalRule       `json:"auto_approval_rules"`
	SeverityThresholds       SeverityThresholdConfig  `json:"severity_thresholds"`
	ResponseTimeRequirements ResponseTimeConfig       `json:"response_time_requirements"`
	NotificationSettings     NotificationConfig       `json:"notification_settings"`
	ExclusionRules           []VulnerabilityExclusion `json:"exclusion_rules"`
	EscalationRules          []EscalationRule         `json:"escalation_rules"`
	ComplianceSettings       ComplianceConfig         `json:"compliance_settings"`
	CreatedAt                time.Time                `json:"created_at"`
	UpdatedAt                time.Time                `json:"updated_at"`
	Version                  int                      `json:"version"`
}

// AutoApprovalRule defines when security updates can be automatically approved.
type AutoApprovalRule struct {
	ID                string                `json:"id"`
	Name              string                `json:"name"`
	Enabled           bool                  `json:"enabled"`
	Conditions        []ApprovalCondition   `json:"conditions"`
	Actions           []AutoApprovalAction  `json:"actions"`
	MaxSeverity       VulnerabilitySeverity `json:"max_severity"`
	RequiredChecks    []string              `json:"required_checks"`
	TestingRequired   bool                  `json:"testing_required"`
	MinTestCoverage   float64               `json:"min_test_coverage,omitempty"`
	BusinessHoursOnly bool                  `json:"business_hours_only"`
	CooldownPeriod    time.Duration         `json:"cooldown_period"`
}

// ApprovalCondition defines conditions for auto-approval.
type ApprovalCondition struct {
	Type     ConditionType `json:"type"`
	Field    string        `json:"field"`
	Operator string        `json:"operator"`
	Value    interface{}   `json:"value"`
	Negated  bool          `json:"negated,omitempty"`
}

// AutoApprovalAction defines actions to take when auto-approving.
type AutoApprovalAction struct {
	Type       ActionType        `json:"type"`
	Parameters map[string]string `json:"parameters,omitempty"`
	DelayAfter time.Duration     `json:"delay_after,omitempty"`
}

// SeverityThresholdConfig defines how to handle different severity levels.
type SeverityThresholdConfig struct {
	Critical SeverityThreshold `json:"critical"`
	High     SeverityThreshold `json:"high"`
	Medium   SeverityThreshold `json:"medium"`
	Low      SeverityThreshold `json:"low"`
}

// SeverityThreshold defines response requirements for a severity level.
type SeverityThreshold struct {
	AutoApprove            bool          `json:"auto_approve"`
	RequireManualReview    bool          `json:"require_manual_review"`
	MaxResponseTime        time.Duration `json:"max_response_time"`
	RequiredApprovers      int           `json:"required_approvers"`
	NotifyImmediately      bool          `json:"notify_immediately"`
	EscalateAfter          time.Duration `json:"escalate_after,omitempty"`
	BusinessImpactAnalysis bool          `json:"business_impact_analysis"`
}

// ResponseTimeConfig defines required response times.
type ResponseTimeConfig struct {
	CriticalVulnerabilities time.Duration `json:"critical_vulnerabilities"`
	HighVulnerabilities     time.Duration `json:"high_vulnerabilities"`
	MediumVulnerabilities   time.Duration `json:"medium_vulnerabilities"`
	LowVulnerabilities      time.Duration `json:"low_vulnerabilities"`
	BusinessHours           BusinessHours `json:"business_hours"`
}

// BusinessHours defines when business hours are active.
type BusinessHours struct {
	Timezone  string    `json:"timezone"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Weekdays  []string  `json:"weekdays"`
	Holidays  []string  `json:"holidays,omitempty"`
}

// NotificationConfig defines how notifications should be sent.
type NotificationConfig struct {
	Enabled           bool                  `json:"enabled"`
	Channels          []NotificationChannel `json:"channels"`
	Templates         map[string]string     `json:"templates,omitempty"`
	EscalationTargets []EscalationTarget    `json:"escalation_targets"`
	SummaryFrequency  string                `json:"summary_frequency"`
}

// NotificationChannel defines a notification delivery channel.
type NotificationChannel struct {
	Type       ChannelType             `json:"type"`
	Target     string                  `json:"target"`
	Enabled    bool                    `json:"enabled"`
	Severities []VulnerabilitySeverity `json:"severities"`
	Format     string                  `json:"format,omitempty"`
	RateLimit  *RateLimitConfig        `json:"rate_limit,omitempty"`
}

// EscalationTarget defines who to notify during escalation.
type EscalationTarget struct {
	Level    int      `json:"level"`
	Users    []string `json:"users"`
	Teams    []string `json:"teams,omitempty"`
	External []string `json:"external,omitempty"`
}

// RateLimitConfig defines rate limiting for notifications.
type RateLimitConfig struct {
	MaxPerHour  int           `json:"max_per_hour"`
	MaxPerDay   int           `json:"max_per_day"`
	BurstLimit  int           `json:"burst_limit"`
	ResetPeriod time.Duration `json:"reset_period"`
}

// VulnerabilityExclusion defines vulnerabilities to exclude from policies.
type VulnerabilityExclusion struct {
	ID        string        `json:"id"`
	Type      ExclusionType `json:"type"`
	Pattern   string        `json:"pattern"`
	Reason    string        `json:"reason"`
	ExpiresAt *time.Time    `json:"expires_at,omitempty"`
	Approver  string        `json:"approver"`
	CreatedAt time.Time     `json:"created_at"`
}

// EscalationRule defines when and how to escalate unresolved vulnerabilities.
type EscalationRule struct {
	ID             string                `json:"id"`
	Name           string                `json:"name"`
	Enabled        bool                  `json:"enabled"`
	TriggerAfter   time.Duration         `json:"trigger_after"`
	Conditions     []EscalationCondition `json:"conditions"`
	Actions        []EscalationAction    `json:"actions"`
	MaxEscalations int                   `json:"max_escalations"`
}

// EscalationCondition defines when escalation should occur.
type EscalationCondition struct {
	Type     string      `json:"type"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// EscalationAction defines what to do during escalation.
type EscalationAction struct {
	Type       string            `json:"type"`
	Target     string            `json:"target"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

// ComplianceConfig defines compliance-related settings.
type ComplianceConfig struct {
	Frameworks            []ComplianceFramework `json:"frameworks"`
	AuditTrailRequired    bool                  `json:"audit_trail_required"`
	DocumentationRequired bool                  `json:"documentation_required"`
	ApprovalEvidence      bool                  `json:"approval_evidence"`
	RetentionPeriod       time.Duration         `json:"retention_period"`
}

// ComplianceFramework defines compliance framework requirements.
type ComplianceFramework struct {
	Name         string        `json:"name"`
	Version      string        `json:"version"`
	Requirements []Requirement `json:"requirements"`
}

// Requirement defines a specific compliance requirement.
type Requirement struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Mandatory   bool   `json:"mandatory"`
}

// VulnerabilityDatabase manages vulnerability data and CVE information.
type VulnerabilityDatabase struct {
	vulnerabilities map[string]*VulnerabilityRecord
	cveCache        map[string]*CVERecord
	lastUpdated     time.Time
}

// VulnerabilityRecord represents a vulnerability in the database.
type VulnerabilityRecord struct {
	ID               string                 `json:"id"`
	CVE              string                 `json:"cve,omitempty"`
	Title            string                 `json:"title"`
	Description      string                 `json:"description"`
	Severity         VulnerabilitySeverity  `json:"severity"`
	CVSS             CVSSScore              `json:"cvss"`
	Package          PackageInfo            `json:"package"`
	AffectedVersions []VersionRange         `json:"affected_versions"`
	PatchedVersions  []string               `json:"patched_versions"`
	References       []Reference            `json:"references"`
	PublishedAt      time.Time              `json:"published_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	WithdrawnAt      *time.Time             `json:"withdrawn_at,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// CVERecord represents a CVE record from external sources.
type CVERecord struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	CVSS        CVSSScore              `json:"cvss"`
	References  []Reference            `json:"references"`
	Vendors     []VendorInfo           `json:"vendors"`
	Products    []ProductInfo          `json:"products"`
	Timeline    CVETimeline            `json:"timeline"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// CVSSScore represents CVSS scoring information.
type CVSSScore struct {
	Version     string  `json:"version"`
	Score       float64 `json:"score"`
	Vector      string  `json:"vector"`
	Severity    string  `json:"severity"`
	BaseScore   float64 `json:"base_score"`
	ImpactScore float64 `json:"impact_score,omitempty"`
}

// PackageInfo represents information about a vulnerable package.
type PackageInfo struct {
	Name      string `json:"name"`
	Ecosystem string `json:"ecosystem"`
	Type      string `json:"type,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// VersionRange represents a range of affected versions.
type VersionRange struct {
	Introduced   string `json:"introduced,omitempty"`
	Fixed        string `json:"fixed,omitempty"`
	LastAffected string `json:"last_affected,omitempty"`
}

// Reference represents a reference URL or identifier.
type Reference struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// VendorInfo represents vendor information in CVE records.
type VendorInfo struct {
	Name     string        `json:"name"`
	Products []ProductInfo `json:"products"`
}

// ProductInfo represents product information in CVE records.
type ProductInfo struct {
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
}

// CVETimeline represents the timeline of a CVE.
type CVETimeline struct {
	Published time.Time  `json:"published"`
	Modified  time.Time  `json:"modified"`
	Reserved  *time.Time `json:"reserved,omitempty"`
	Rejected  *time.Time `json:"rejected,omitempty"`
}

// SecurityUpdateStatus represents the status of a security update.
type SecurityUpdateStatus struct {
	UpdateID        string          `json:"update_id"`
	VulnerabilityID string          `json:"vulnerability_id"`
	Repository      string          `json:"repository"`
	Organization    string          `json:"organization"`
	Package         PackageInfo     `json:"package"`
	CurrentVersion  string          `json:"current_version"`
	TargetVersion   string          `json:"target_version"`
	Status          UpdateStatus    `json:"status"`
	Priority        UpdatePriority  `json:"priority"`
	AutoApproved    bool            `json:"auto_approved"`
	ApprovalReason  string          `json:"approval_reason,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeadlineAt      *time.Time      `json:"deadline_at,omitempty"`
	CompletedAt     *time.Time      `json:"completed_at,omitempty"`
	ApprovedBy      []string        `json:"approved_by,omitempty"`
	ReviewNotes     []ReviewNote    `json:"review_notes,omitempty"`
	TestResults     *TestResults    `json:"test_results,omitempty"`
	RiskAssessment  *RiskAssessment `json:"risk_assessment,omitempty"`
}

// ReviewNote represents a review note for a security update.
type ReviewNote struct {
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

// TestResults represents test results for a security update.
type TestResults struct {
	Passed      bool          `json:"passed"`
	TestSuite   string        `json:"test_suite"`
	Coverage    float64       `json:"coverage,omitempty"`
	Duration    time.Duration `json:"duration"`
	FailedTests []string      `json:"failed_tests,omitempty"`
	ExecutedAt  time.Time     `json:"executed_at"`
}

// RiskAssessment represents a risk assessment for a security update.
type RiskAssessment struct {
	OverallRisk    RiskLevel            `json:"overall_risk"`
	BusinessImpact ImpactLevel          `json:"business_impact"`
	TechnicalRisk  RiskLevel            `json:"technical_risk"`
	Factors        []RiskFactor         `json:"factors"`
	Mitigation     []MitigationStrategy `json:"mitigation"`
	Assessor       string               `json:"assessor"`
	AssessedAt     time.Time            `json:"assessed_at"`
}

// RiskFactor represents a factor contributing to risk.
type RiskFactor struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Impact      RiskLevel `json:"impact"`
	Likelihood  string    `json:"likelihood"`
}

// MitigationStrategy represents a strategy to mitigate risk.
type MitigationStrategy struct {
	Type          string `json:"type"`
	Description   string `json:"description"`
	Effort        string `json:"effort"`
	Effectiveness string `json:"effectiveness"`
}

// Enum types.
type VulnerabilitySeverity string

const (
	VulnSeverityCritical VulnerabilitySeverity = "critical"
	VulnSeverityHigh     VulnerabilitySeverity = "high"
	VulnSeverityMedium   VulnerabilitySeverity = "medium"
	VulnSeverityLow      VulnerabilitySeverity = "low"
	VulnSeverityInfo     VulnerabilitySeverity = "info"
)

type ConditionType string

const (
	ConditionTypeSeverity   ConditionType = "severity"
	ConditionTypePackage    ConditionType = "package"
	ConditionTypeVersion    ConditionType = "version"
	ConditionTypeCVSS       ConditionType = "cvss"
	ConditionTypeAge        ConditionType = "age"
	ConditionTypeRepository ConditionType = "repository"
	ConditionTypeEcosystem  ConditionType = "ecosystem"
)

// Security-specific action types (extending ActionType from automation_rule.go).
const (
	ActionTypeSecurityApprove      ActionType = "security_approve"
	ActionTypeSecurityMerge        ActionType = "security_merge"
	ActionTypeSecurityNotify       ActionType = "security_notify"
	ActionTypeSecurityTest         ActionType = "security_test"
	ActionTypeSecurityCreateTicket ActionType = "security_create_ticket"
	ActionTypeSecuritySchedule     ActionType = "security_schedule"
)

type ChannelType string

const (
	ChannelTypeEmail   ChannelType = "email"
	ChannelTypeSlack   ChannelType = "slack"
	ChannelTypeWebhook ChannelType = "webhook"
	ChannelTypeSMS     ChannelType = "sms"
	ChannelTypePager   ChannelType = "pager"
)

type ExclusionType string

const (
	ExclusionTypeCVE        ExclusionType = "cve"
	ExclusionTypePackage    ExclusionType = "package"
	ExclusionTypeRepository ExclusionType = "repository"
	ExclusionTypePattern    ExclusionType = "pattern"
)

type UpdateStatus string

const (
	UpdateStatusPending   UpdateStatus = "pending"
	UpdateStatusReviewing UpdateStatus = "reviewing"
	UpdateStatusApproved  UpdateStatus = "approved"
	UpdateStatusRejected  UpdateStatus = "rejected"
	UpdateStatusTesting   UpdateStatus = "testing"
	UpdateStatusDeploying UpdateStatus = "deploying"
	UpdateStatusCompleted UpdateStatus = "completed"
	UpdateStatusFailed    UpdateStatus = "failed"
	UpdateStatusCancelled UpdateStatus = "cancelled"
)

type UpdatePriority string

const (
	UpdatePriorityCritical UpdatePriority = "critical"
	UpdatePriorityHigh     UpdatePriority = "high"
	UpdatePriorityMedium   UpdatePriority = "medium"
	UpdatePriorityLow      UpdatePriority = "low"
)

// Use existing RiskLevel from interfaces.go
// Additional risk levels for security context.
const (
	SecurityRiskLevelMinimal RiskLevel = "minimal"
)

type ImpactLevel string

const (
	ImpactLevelCritical ImpactLevel = "critical"
	ImpactLevelHigh     ImpactLevel = "high"
	ImpactLevelMedium   ImpactLevel = "medium"
	ImpactLevelLow      ImpactLevel = "low"
	ImpactLevelMinimal  ImpactLevel = "minimal"
)

// NewSecurityUpdatePolicyManager creates a new security update policy manager.
func NewSecurityUpdatePolicyManager(logger Logger, apiClient APIClient, dependabotManager *DependabotConfigManager) *SecurityUpdatePolicyManager {
	return &SecurityUpdatePolicyManager{
		logger:            logger,
		apiClient:         apiClient,
		dependabotManager: dependabotManager,
		policies:          make(map[string]*SecurityUpdatePolicy),
		vulnerabilityDB:   NewVulnerabilityDatabase(),
	}
}

// NewVulnerabilityDatabase creates a new vulnerability database.
func NewVulnerabilityDatabase() *VulnerabilityDatabase {
	return &VulnerabilityDatabase{
		vulnerabilities: make(map[string]*VulnerabilityRecord),
		cveCache:        make(map[string]*CVERecord),
		lastUpdated:     time.Now(),
	}
}

// CreateSecurityPolicy creates a new security update policy.
func (sm *SecurityUpdatePolicyManager) CreateSecurityPolicy(ctx context.Context, policy *SecurityUpdatePolicy) error {
	sm.logger.Info("Creating security update policy", "organization", policy.Organization, "policy", policy.Name)

	// Validate policy
	if err := sm.validateSecurityPolicy(policy); err != nil {
		return fmt.Errorf("invalid security policy: %w", err)
	}

	// Set metadata
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()
	policy.Version = 1

	// Store policy
	sm.policies[policy.ID] = policy

	sm.logger.Info("Security update policy created successfully", "policy_id", policy.ID)

	return nil
}

// EvaluateSecurityUpdate evaluates whether a security update should be auto-approved.
func (sm *SecurityUpdatePolicyManager) EvaluateSecurityUpdate(ctx context.Context, policyID string, update *SecurityUpdateStatus) (*SecurityUpdateDecision, error) {
	sm.logger.Debug("Evaluating security update", "policy_id", policyID, "update_id", update.UpdateID)

	policy, exists := sm.policies[policyID]
	if !exists {
		return nil, fmt.Errorf("security policy not found: %s", policyID)
	}

	if !policy.Enabled {
		return &SecurityUpdateDecision{
			Approved: false,
			Reason:   "Security policy is disabled",
		}, nil
	}

	// Get vulnerability information
	vuln, err := sm.getVulnerabilityInfo(ctx, update.VulnerabilityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vulnerability info: %w", err)
	}

	// Check exclusion rules
	if sm.isExcluded(policy, vuln, update) {
		return &SecurityUpdateDecision{
			Approved: false,
			Reason:   "Update matches exclusion rule",
		}, nil
	}

	// Evaluate auto-approval rules
	decision := sm.evaluateAutoApprovalRules(policy, vuln, update)

	sm.logger.Info("Security update evaluation completed",
		"update_id", update.UpdateID,
		"approved", decision.Approved,
		"reason", decision.Reason)

	return decision, nil
}

// ProcessSecurityUpdates processes pending security updates for an organization.
func (sm *SecurityUpdatePolicyManager) ProcessSecurityUpdates(ctx context.Context, organization string) (*SecurityUpdateProcessResult, error) {
	sm.logger.Info("Processing security updates", "organization", organization)

	result := &SecurityUpdateProcessResult{
		Organization: organization,
		StartedAt:    time.Now(),
		Updates:      make([]SecurityUpdateStatus, 0),
	}

	// Get pending security updates
	pendingUpdates, err := sm.getPendingSecurityUpdates(ctx, organization)
	if err != nil {
		return result, fmt.Errorf("failed to get pending updates: %w", err)
	}

	result.TotalUpdates = len(pendingUpdates)

	// Process each update
	for _, update := range pendingUpdates {
		// Find applicable policy
		policyID := sm.findApplicablePolicy(organization, update.Repository)
		if policyID == "" {
			sm.logger.Warn("No applicable security policy found", "repository", update.Repository)
			continue
		}

		// Evaluate update
		decision, err := sm.EvaluateSecurityUpdate(ctx, policyID, &update)
		if err != nil {
			sm.logger.Error("Failed to evaluate security update", "update_id", update.UpdateID, "error", err)

			result.FailedUpdates++

			continue
		}

		// Apply decision
		if decision.Approved {
			err = sm.approveSecurityUpdate(ctx, &update, decision)
			if err != nil {
				sm.logger.Error("Failed to approve security update", "update_id", update.UpdateID, "error", err)

				result.FailedUpdates++
			} else {
				result.ApprovedUpdates++
			}
		} else {
			result.PendingReview++
		}

		result.Updates = append(result.Updates, update)
	}

	result.CompletedAt = time.Now()
	result.ProcessingTime = result.CompletedAt.Sub(result.StartedAt)

	sm.logger.Info("Security updates processing completed",
		"organization", organization,
		"total", result.TotalUpdates,
		"approved", result.ApprovedUpdates,
		"pending", result.PendingReview,
		"failed", result.FailedUpdates)

	return result, nil
}

// Helper methods

func (sm *SecurityUpdatePolicyManager) validateSecurityPolicy(policy *SecurityUpdatePolicy) error {
	if policy.ID == "" {
		return fmt.Errorf("policy ID is required")
	}

	if policy.Name == "" {
		return fmt.Errorf("policy name is required")
	}

	if policy.Organization == "" {
		return fmt.Errorf("organization is required")
	}

	// Validate auto-approval rules
	for i, rule := range policy.AutoApprovalRules {
		if rule.ID == "" {
			return fmt.Errorf("auto-approval rule %d: ID is required", i)
		}

		if len(rule.Conditions) == 0 {
			return fmt.Errorf("auto-approval rule %d: at least one condition is required", i)
		}
	}

	return nil
}

func (sm *SecurityUpdatePolicyManager) getVulnerabilityInfo(ctx context.Context, vulnID string) (*VulnerabilityRecord, error) {
	// Check cache first
	if vuln, exists := sm.vulnerabilityDB.vulnerabilities[vulnID]; exists {
		return vuln, nil
	}

	// In a real implementation, this would fetch from external vulnerability databases
	// For now, return a mock vulnerability
	vuln := &VulnerabilityRecord{
		ID:          vulnID,
		CVE:         "CVE-2024-" + vulnID[len(vulnID)-4:],
		Title:       "Mock vulnerability for testing",
		Description: "This is a mock vulnerability record",
		Severity:    VulnSeverityMedium,
		CVSS: CVSSScore{
			Version:   "3.1",
			Score:     5.5,
			Vector:    "CVSS:3.1/AV:N/AC:L/PR:N/UI:R/S:U/C:N/I:N/A:H",
			Severity:  "MEDIUM",
			BaseScore: 5.5,
		},
		PublishedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now(),
	}

	sm.vulnerabilityDB.vulnerabilities[vulnID] = vuln

	return vuln, nil
}

func (sm *SecurityUpdatePolicyManager) isExcluded(policy *SecurityUpdatePolicy, vuln *VulnerabilityRecord, update *SecurityUpdateStatus) bool {
	for _, exclusion := range policy.ExclusionRules {
		// Check if exclusion has expired
		if exclusion.ExpiresAt != nil && time.Now().After(*exclusion.ExpiresAt) {
			continue
		}

		switch exclusion.Type {
		case ExclusionTypeCVE:
			if vuln.CVE == exclusion.Pattern {
				return true
			}
		case ExclusionTypePackage:
			if update.Package.Name == exclusion.Pattern {
				return true
			}
		case ExclusionTypeRepository:
			if update.Repository == exclusion.Pattern {
				return true
			}
		case ExclusionTypePattern:
			if strings.Contains(update.Package.Name, exclusion.Pattern) {
				return true
			}
		}
	}

	return false
}

func (sm *SecurityUpdatePolicyManager) evaluateAutoApprovalRules(policy *SecurityUpdatePolicy, vuln *VulnerabilityRecord, update *SecurityUpdateStatus) *SecurityUpdateDecision {
	for _, rule := range policy.AutoApprovalRules {
		if !rule.Enabled {
			continue
		}

		// Check if severity exceeds maximum allowed
		if sm.severityExceeds(vuln.Severity, rule.MaxSeverity) {
			continue
		}

		// Evaluate all conditions
		allConditionsMet := true

		for _, condition := range rule.Conditions {
			if !sm.evaluateCondition(condition, vuln, update) {
				allConditionsMet = false
				break
			}
		}

		if allConditionsMet {
			return &SecurityUpdateDecision{
				Approved:     true,
				Reason:       fmt.Sprintf("Auto-approved by rule: %s", rule.Name),
				RuleID:       rule.ID,
				Actions:      rule.Actions,
				RequiresTest: rule.TestingRequired,
			}
		}
	}

	return &SecurityUpdateDecision{
		Approved: false,
		Reason:   "No auto-approval rules matched",
	}
}

func (sm *SecurityUpdatePolicyManager) severityExceeds(actual, maximum VulnerabilitySeverity) bool {
	severityOrder := map[VulnerabilitySeverity]int{
		VulnSeverityInfo:     0,
		VulnSeverityLow:      1,
		VulnSeverityMedium:   2,
		VulnSeverityHigh:     3,
		VulnSeverityCritical: 4,
	}

	return severityOrder[actual] > severityOrder[maximum]
}

func (sm *SecurityUpdatePolicyManager) evaluateCondition(condition ApprovalCondition, vuln *VulnerabilityRecord, update *SecurityUpdateStatus) bool {
	result := false

	switch condition.Type {
	case ConditionTypeSeverity:
		if condition.Field == "severity" {
			result = sm.compareSeverity(vuln.Severity, condition.Operator, condition.Value)
		}
	case ConditionTypePackage:
		switch condition.Field {
		case "name":
			result = sm.compareString(update.Package.Name, condition.Operator, condition.Value)
		case "ecosystem":
			result = sm.compareString(update.Package.Ecosystem, condition.Operator, condition.Value)
		}
	case ConditionTypeCVSS:
		if condition.Field == "score" {
			result = sm.compareFloat(vuln.CVSS.Score, condition.Operator, condition.Value)
		}
	}

	if condition.Negated {
		result = !result
	}

	return result
}

func (sm *SecurityUpdatePolicyManager) compareSeverity(actual VulnerabilitySeverity, operator string, expected interface{}) bool {
	expectedStr, ok := expected.(string)
	if !ok {
		return false
	}

	switch operator {
	case "eq":
		return actual == VulnerabilitySeverity(expectedStr)
	case "lte":
		return !sm.severityExceeds(actual, VulnerabilitySeverity(expectedStr))
	case "gte":
		return sm.severityExceeds(actual, VulnerabilitySeverity(expectedStr)) || actual == VulnerabilitySeverity(expectedStr)
	}

	return false
}

func (sm *SecurityUpdatePolicyManager) compareString(actual, operator string, expected interface{}) bool {
	expectedStr, ok := expected.(string)
	if !ok {
		return false
	}

	switch operator {
	case "eq":
		return actual == expectedStr
	case "contains":
		return strings.Contains(actual, expectedStr)
	case "starts_with":
		return strings.HasPrefix(actual, expectedStr)
	case "ends_with":
		return strings.HasSuffix(actual, expectedStr)
	}

	return false
}

func (sm *SecurityUpdatePolicyManager) compareFloat(actual float64, operator string, expected interface{}) bool {
	var expectedFloat float64
	switch v := expected.(type) {
	case float64:
		expectedFloat = v
	case int:
		expectedFloat = float64(v)
	default:
		return false
	}

	switch operator {
	case "eq":
		return actual == expectedFloat
	case "lt":
		return actual < expectedFloat
	case "lte":
		return actual <= expectedFloat
	case "gt":
		return actual > expectedFloat
	case "gte":
		return actual >= expectedFloat
	}

	return false
}

func (sm *SecurityUpdatePolicyManager) getPendingSecurityUpdates(ctx context.Context, organization string) ([]SecurityUpdateStatus, error) {
	// In a real implementation, this would query GitHub API for pending security updates
	// For now, return mock data
	return []SecurityUpdateStatus{
		{
			UpdateID:        "update-1",
			VulnerabilityID: "vuln-1",
			Repository:      "test-repo",
			Organization:    organization,
			Package: PackageInfo{
				Name:      "lodash",
				Ecosystem: "npm",
			},
			CurrentVersion: "4.17.20",
			TargetVersion:  "4.17.21",
			Status:         UpdateStatusPending,
			Priority:       UpdatePriorityHigh,
			CreatedAt:      time.Now().Add(-1 * time.Hour),
			UpdatedAt:      time.Now(),
		},
	}, nil
}

func (sm *SecurityUpdatePolicyManager) findApplicablePolicy(organization, repository string) string {
	// Find the most specific policy that applies
	for policyID, policy := range sm.policies {
		if policy.Organization == organization && policy.Enabled {
			return policyID
		}
	}

	return ""
}

func (sm *SecurityUpdatePolicyManager) approveSecurityUpdate(ctx context.Context, update *SecurityUpdateStatus, decision *SecurityUpdateDecision) error {
	// In a real implementation, this would interact with GitHub API to approve the update
	sm.logger.Info("Approving security update",
		"update_id", update.UpdateID,
		"rule_id", decision.RuleID,
		"reason", decision.Reason)

	update.Status = UpdateStatusApproved
	update.AutoApproved = true
	update.ApprovalReason = decision.Reason
	update.UpdatedAt = time.Now()

	return nil
}

// SecurityUpdateDecision represents the result of evaluating a security update.
type SecurityUpdateDecision struct {
	Approved     bool                 `json:"approved"`
	Reason       string               `json:"reason"`
	RuleID       string               `json:"rule_id,omitempty"`
	Actions      []AutoApprovalAction `json:"actions,omitempty"`
	RequiresTest bool                 `json:"requires_test"`
	Conditions   []string             `json:"conditions,omitempty"`
}

// SecurityUpdateProcessResult represents the result of processing security updates.
type SecurityUpdateProcessResult struct {
	Organization    string                 `json:"organization"`
	TotalUpdates    int                    `json:"total_updates"`
	ApprovedUpdates int                    `json:"approved_updates"`
	PendingReview   int                    `json:"pending_review"`
	FailedUpdates   int                    `json:"failed_updates"`
	Updates         []SecurityUpdateStatus `json:"updates"`
	StartedAt       time.Time              `json:"started_at"`
	CompletedAt     time.Time              `json:"completed_at"`
	ProcessingTime  time.Duration          `json:"processing_time"`
}
