package github

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// DependencyVersionPolicyManager manages dependency version policies for repositories.
type DependencyVersionPolicyManager struct {
	logger                Logger
	apiClient             APIClient
	dependabotManager     *DependabotConfigManager
	securityPolicyManager *SecurityUpdatePolicyManager
	policies              map[string]*DependencyVersionPolicy
	versionConstraints    *VersionConstraintEngine
}

// DependencyVersionPolicy defines version management policies for dependencies.
type DependencyVersionPolicy struct {
	ID                   string                            `json:"id"`
	Name                 string                            `json:"name"`
	Organization         string                            `json:"organization"`
	Description          string                            `json:"description"`
	Enabled              bool                              `json:"enabled"`
	VersionConstraints   map[string]VersionConstraintRule  `json:"versionConstraints"`
	EcosystemPolicies    map[string]EcosystemVersionPolicy `json:"ecosystemPolicies"`
	BreakingChangePolicy BreakingChangePolicy              `json:"breakingChangePolicy"`
	CompatibilityChecks  CompatibilityCheckConfig          `json:"compatibilityChecks"`
	RollbackPolicy       RollbackPolicy                    `json:"rollbackPolicy"`
	ApprovalRequirements VersionUpdateApprovalRequirements `json:"approvalRequirements"`
	NotificationSettings VersionPolicyNotificationConfig   `json:"notificationSettings"`
	TestingRequirements  TestingRequirements               `json:"testingRequirements"`
	ReleaseWindows       []ReleaseWindow                   `json:"releaseWindows"`
	MetricsTracking      MetricsTrackingConfig             `json:"metricsTracking"`
	CreatedAt            time.Time                         `json:"createdAt"`
	UpdatedAt            time.Time                         `json:"updatedAt"`
	Version              int                               `json:"version"`
}

// VersionConstraintRule defines version constraints for dependencies.
type VersionConstraintRule struct {
	RuleID            string                       `json:"ruleId"`
	DependencyPattern string                       `json:"dependencyPattern"`
	Ecosystem         string                       `json:"ecosystem"`
	AllowedVersions   []VersionRange               `json:"allowedVersions"`
	BlockedVersions   []VersionRange               `json:"blockedVersions"`
	PreferredVersions []VersionRange               `json:"preferredVersions"`
	MinimumVersion    string                       `json:"minimumVersion,omitempty"`
	MaximumVersion    string                       `json:"maximumVersion,omitempty"`
	VersionPattern    string                       `json:"versionPattern,omitempty"`
	AllowPrerelease   bool                         `json:"allowPrerelease"`
	AllowBetaVersions bool                         `json:"allowBetaVersions"`
	UpdateStrategy    DependencyUpdateStrategy     `json:"updateStrategy"`
	AutoUpdateEnabled bool                         `json:"autoUpdateEnabled"`
	UpdateFrequency   UpdateFrequency              `json:"updateFrequency"`
	Priority          ConstraintPriority           `json:"priority"`
	ExpirationDate    *time.Time                   `json:"expirationDate,omitempty"`
	Justification     string                       `json:"justification"`
	Exceptions        []VersionConstraintException `json:"exceptions,omitempty"`
}

// EcosystemVersionPolicy defines version policies specific to package ecosystems.
type EcosystemVersionPolicy struct {
	Ecosystem               string                   `json:"ecosystem"`
	Enabled                 bool                     `json:"enabled"`
	DefaultUpdateStrategy   DependencyUpdateStrategy `json:"defaultUpdateStrategy"`
	AllowMajorUpdates       bool                     `json:"allowMajorUpdates"`
	AllowMinorUpdates       bool                     `json:"allowMinorUpdates"`
	AllowPatchUpdates       bool                     `json:"allowPatchUpdates"`
	RequireSecurityUpdates  bool                     `json:"requireSecurityUpdates"`
	MaxVersionAge           time.Duration            `json:"maxVersionAge"`
	DeprecationPolicy       DeprecationPolicy        `json:"deprecationPolicy"`
	LicenseRestrictions     []LicenseRestriction     `json:"licenseRestrictions"`
	PerformanceRequirements PerformanceRequirements  `json:"performanceRequirements"`
	QualityGates            []QualityGate            `json:"qualityGates"`
	CustomValidationRules   []CustomValidationRule   `json:"customValidationRules"`
}

// BreakingChangePolicy defines how to handle breaking changes.
type BreakingChangePolicy struct {
	AllowBreakingChanges        bool                    `json:"allowBreakingChanges"`
	BreakingChangeDetection     BreakingChangeDetection `json:"breakingChangeDetection"`
	ImpactAnalysisRequired      bool                    `json:"impactAnalysisRequired"`
	DeprecationNoticePeriod     time.Duration           `json:"deprecationNoticePeriod"`
	MigrationGuidanceRequired   bool                    `json:"migrationGuidanceRequired"`
	BackwardCompatibilityPeriod time.Duration           `json:"backwardCompatibilityPeriod"`
	BreakingChangeApprovers     []string                `json:"breakingChangeApprovers"`
	CommunicationPlan           CommunicationPlan       `json:"communicationPlan"`
}

// BreakingChangeDetection configures how breaking changes are detected.
type BreakingChangeDetection struct {
	Enabled                bool              `json:"enabled"`
	Methods                []DetectionMethod `json:"methods"`
	SemverStrictMode       bool              `json:"semverStrictMode"`
	APIChangeDetection     bool              `json:"apiChangeDetection"`
	SchemaChangeDetection  bool              `json:"schemaChangeDetection"`
	CustomDetectionRules   []DetectionRule   `json:"customDetectionRules"`
	IgnorePatterns         []string          `json:"ignorePatterns"`
	ThresholdConfiguration ThresholdConfig   `json:"thresholdConfiguration"`
}

// CompatibilityCheckConfig defines compatibility checking requirements.
type CompatibilityCheckConfig struct {
	Enabled                   bool                       `json:"enabled"`
	MatrixTesting             MatrixTestingConfig        `json:"matrixTesting"`
	DependencyGraphAnalysis   bool                       `json:"dependencyGraphAnalysis"`
	ConflictDetection         ConflictDetectionConfig    `json:"conflictDetection"`
	IntegrationTesting        IntegrationTestingConfig   `json:"integrationTesting"`
	PerformanceImpactAnalysis bool                       `json:"performanceImpactAnalysis"`
	SecurityImpactAnalysis    bool                       `json:"securityImpactAnalysis"`
	CompatibilityMatrix       []CompatibilityMatrixEntry `json:"compatibilityMatrix"`
	RegressionTesting         RegressionTestingConfig    `json:"regressionTesting"`
}

// RollbackPolicy defines rollback procedures and conditions.
type RollbackPolicy struct {
	Enabled                 bool                     `json:"enabled"`
	AutoRollbackTriggers    []RollbackTrigger        `json:"autoRollbackTriggers"`
	ManualRollbackProcedure ManualRollbackProcedure  `json:"manualRollbackProcedure"`
	RollbackTimeframe       time.Duration            `json:"rollbackTimeframe"`
	HealthCheckRequirements []HealthCheck            `json:"healthCheckRequirements"`
	RollbackApprovers       []string                 `json:"rollbackApprovers"`
	DataMigrationHandling   DataMigrationHandling    `json:"dataMigrationHandling"`
	NotificationPlan        RollbackNotificationPlan `json:"notificationPlan"`
	PostRollbackAnalysis    bool                     `json:"postRollbackAnalysis"`
}

// VersionUpdateApprovalRequirements defines approval requirements for version updates.
type VersionUpdateApprovalRequirements struct {
	MajorVersionUpdates VersionApprovalRule            `json:"majorVersionUpdates"`
	MinorVersionUpdates VersionApprovalRule            `json:"minorVersionUpdates"`
	PatchVersionUpdates VersionApprovalRule            `json:"patchVersionUpdates"`
	SecurityUpdates     VersionApprovalRule            `json:"securityUpdates"`
	PreReleaseUpdates   VersionApprovalRule            `json:"preReleaseUpdates"`
	EmergencyUpdates    EmergencyApprovalRule          `json:"emergencyUpdates"`
	BulkUpdates         BulkUpdateApprovalRule         `json:"bulkUpdates"`
	DependencySpecific  map[string]VersionApprovalRule `json:"dependencySpecific,omitempty"`
}

// VersionApprovalRule defines approval rules for version updates.
type VersionApprovalRule struct {
	RequiredApprovers          int                      `json:"requiredApprovers"`
	RequiredApprovalTeams      []string                 `json:"requiredApprovalTeams"`
	AutoApprovalConditions     []AutoApprovalCondition  `json:"autoApprovalConditions"`
	ManualReviewRequired       bool                     `json:"manualReviewRequired"`
	SecurityReviewRequired     bool                     `json:"securityReviewRequired"`
	ArchitectureReviewRequired bool                     `json:"architectureReviewRequired"`
	BusinessApprovalRequired   bool                     `json:"businessApprovalRequired"`
	TestingGateRequired        bool                     `json:"testingGateRequired"`
	WaitingPeriod              time.Duration            `json:"waitingPeriod,omitempty"`
	ApprovalTimeLimit          time.Duration            `json:"approvalTimeLimit,omitempty"`
	EscalationRules            []ApprovalEscalationRule `json:"escalationRules"`
}

// TestingRequirements defines testing requirements for version updates.
type TestingRequirements struct {
	Enabled                    bool                   `json:"enabled"`
	UnitTestingRequired        bool                   `json:"unitTestingRequired"`
	IntegrationTestingRequired bool                   `json:"integrationTestingRequired"`
	E2ETestingRequired         bool                   `json:"e2eTestingRequired"`
	PerformanceTestingRequired bool                   `json:"performanceTestingRequired"`
	SecurityTestingRequired    bool                   `json:"securityTestingRequired"`
	MinimumTestCoverage        float64                `json:"minimumTestCoverage"`
	TestSuiteConfiguration     TestSuiteConfiguration `json:"testSuiteConfiguration"`
	AutomatedTesting           AutomatedTestingConfig `json:"automatedTesting"`
	ManualTestingChecklist     []ManualTestingItem    `json:"manual_testing_checklist"`
	TestEnvironments           []TestEnvironment      `json:"test_environments"`
	TestDataRequirements       TestDataRequirements   `json:"test_data_requirements"`
}

// ReleaseWindow defines allowed time windows for dependency updates.
type ReleaseWindow struct {
	ID                   string                     `json:"id"`
	Name                 string                     `json:"name"`
	Description          string                     `json:"description"`
	Enabled              bool                       `json:"enabled"`
	Schedule             ReleaseSchedule            `json:"schedule"`
	AllowedUpdateTypes   []string                   `json:"allowed_update_types"`
	RestrictedEcosystems []string                   `json:"restricted_ecosystems"`
	ApprovalRequired     bool                       `json:"approval_required"`
	Approvers            []string                   `json:"approvers"`
	NotificationSettings WindowNotificationSettings `json:"notification_settings"`
	BlackoutPeriods      []BlackoutPeriod           `json:"blackout_periods"`
	EmergencyOverride    EmergencyOverride          `json:"emergency_override"`
}

// VersionConstraintEngine handles version constraint evaluation and resolution.
type VersionConstraintEngine struct {
	logger          Logger
	semverParser    SemverParser
	constraintCache map[string]*ConstraintEvaluationResult
	cacheTTL        time.Duration
}

// DependencyVersionAnalysis represents analysis results for a dependency version update.
type DependencyVersionAnalysis struct {
	DependencyName         string                       `json:"dependency_name"`
	Ecosystem              string                       `json:"ecosystem"`
	CurrentVersion         string                       `json:"current_version"`
	ProposedVersion        string                       `json:"proposed_version"`
	UpdateType             string                       `json:"update_type"`
	VersionConstraintCheck VersionConstraintCheckResult `json:"version_constraint_check"`
	CompatibilityAnalysis  CompatibilityAnalysisResult  `json:"compatibility_analysis"`
	SecurityImpact         SecurityImpactAnalysis       `json:"security_impact"`
	PerformanceImpact      PerformanceImpactAnalysis    `json:"performance_impact"`
	BreakingChangeAnalysis BreakingChangeAnalysisResult `json:"breaking_change_analysis"`
	LicenseCompatibility   LicenseCompatibilityResult   `json:"license_compatibility"`
	RiskAssessment         DependencyRiskAssessment     `json:"risk_assessment"`
	RecommendedAction      RecommendedAction            `json:"recommended_action"`
	TestingRecommendations []TestingRecommendation      `json:"testing_recommendations"`
	RollbackPlan           RollbackPlan                 `json:"rollback_plan"`
	Timeline               UpdateTimeline               `json:"timeline"`
	ApprovalWorkflow       ApprovalWorkflow             `json:"approval_workflow"`
}

// Supporting types and enums.
type DependencyUpdateStrategy string

const (
	UpdateStrategyConservative DependencyUpdateStrategy = "conservative"
	UpdateStrategyModerate     DependencyUpdateStrategy = "moderate"
	UpdateStrategyAggressive   DependencyUpdateStrategy = "aggressive"
	UpdateStrategySecurityOnly DependencyUpdateStrategy = "security_only"
	UpdateStrategyCustom       DependencyUpdateStrategy = "custom"
)

type UpdateFrequency string

const (
	UpdateFrequencyImmediate UpdateFrequency = "immediate"
	UpdateFrequencyDaily     UpdateFrequency = "daily"
	UpdateFrequencyWeekly    UpdateFrequency = "weekly"
	UpdateFrequencyBiWeekly  UpdateFrequency = "bi_weekly"
	UpdateFrequencyMonthly   UpdateFrequency = "monthly"
	UpdateFrequencyQuarterly UpdateFrequency = "quarterly"
	UpdateFrequencyManual    UpdateFrequency = "manual"
)

type ConstraintPriority string

const (
	ConstraintPriorityLow      ConstraintPriority = "low"
	ConstraintPriorityMedium   ConstraintPriority = "medium"
	ConstraintPriorityHigh     ConstraintPriority = "high"
	ConstraintPriorityCritical ConstraintPriority = "critical"
)

type DetectionMethod string

const (
	DetectionMethodSemver    DetectionMethod = "semver"
	DetectionMethodAPI       DetectionMethod = "api_diff"
	DetectionMethodSchema    DetectionMethod = "schema_diff"
	DetectionMethodCustom    DetectionMethod = "custom_rules"
	DetectionMethodChangeLog DetectionMethod = "changelog_analysis"
	DetectionMethodBinary    DetectionMethod = "binary_diff"
)

// Supporting structs for complex configurations.
type VersionConstraintException struct {
	Repository    string    `json:"repository"`
	Justification string    `json:"justification"`
	ExpiresAt     time.Time `json:"expires_at"`
	Approver      string    `json:"approver"`
}

type DeprecationPolicy struct {
	AllowDeprecatedVersions  bool          `json:"allow_deprecated_versions"`
	DeprecationWarningPeriod time.Duration `json:"deprecation_warning_period"`
	ForceUpgradeAfterEOL     bool          `json:"force_upgrade_after_eol"`
	EOLNotificationPeriod    time.Duration `json:"eol_notification_period"`
}

type LicenseRestriction struct {
	BlockedLicenses            []string            `json:"blocked_licenses"`
	RequiredLicenses           []string            `json:"required_licenses,omitempty"`
	LicenseCompatibilityMatrix map[string][]string `json:"license_compatibility_matrix,omitempty"`
}

type PerformanceRequirements struct {
	MaxPerformanceRegression float64            `json:"max_performance_regression"`
	BenchmarkSuites          []string           `json:"benchmark_suites"`
	PerformanceThresholds    map[string]float64 `json:"performance_thresholds"`
}

type QualityGate struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Threshold  float64                `json:"threshold"`
	Required   bool                   `json:"required"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

type CustomValidationRule struct {
	Name       string                 `json:"name"`
	Script     string                 `json:"script"`
	Language   string                 `json:"language"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Required   bool                   `json:"required"`
}

// Additional supporting types.
type DetectionRule struct {
	Pattern     string  `json:"pattern"`
	Severity    string  `json:"severity"`
	Description string  `json:"description"`
	Weight      float64 `json:"weight"`
}

type ThresholdConfig struct {
	MinorChangeThreshold    float64 `json:"minor_change_threshold"`
	MajorChangeThreshold    float64 `json:"major_change_threshold"`
	BreakingChangeThreshold float64 `json:"breaking_change_threshold"`
}

type MatrixTestingConfig struct {
	Enabled          bool                `json:"enabled"`
	OperatingSystems []string            `json:"operating_systems"`
	RuntimeVersions  []string            `json:"runtime_versions"`
	DatabaseVersions []string            `json:"database_versions,omitempty"`
	BrowserVersions  []string            `json:"browser_versions,omitempty"`
	CustomDimensions map[string][]string `json:"custom_dimensions,omitempty"`
}

type ConflictDetectionConfig struct {
	Enabled                       bool     `json:"enabled"`
	CheckTransitiveDependencies   bool     `json:"check_transitive_dependencies"`
	ResolveConflictsAutomatically bool     `json:"resolve_conflicts_automatically"`
	ConflictResolutionStrategy    string   `json:"conflict_resolution_strategy"`
	IgnoredConflicts              []string `json:"ignored_conflicts,omitempty"`
}

type IntegrationTestingConfig struct {
	Enabled          bool              `json:"enabled"`
	TestSuites       []string          `json:"test_suites"`
	RequiredCoverage float64           `json:"required_coverage"`
	Timeout          time.Duration     `json:"timeout"`
	Environment      string            `json:"environment"`
	PreTestSetup     []string          `json:"pre_test_setup"`
	PostTestCleanup  []string          `json:"post_test_cleanup"`
	TestData         map[string]string `json:"test_data,omitempty"`
}

type CompatibilityMatrixEntry struct {
	Dependency1          string   `json:"dependency1"`
	Dependency2          string   `json:"dependency2"`
	CompatibleVersions   []string `json:"compatible_versions"`
	IncompatibleVersions []string `json:"incompatible_versions"`
	Notes                string   `json:"notes,omitempty"`
}

type RegressionTestingConfig struct {
	Enabled                   bool     `json:"enabled"`
	BaselineVersion           string   `json:"baseline_version"`
	TestSuites                []string `json:"test_suites"`
	AutomatedRegression       bool     `json:"automated_regression"`
	ManualRegressionChecklist []string `json:"manual_regression_checklist"`
	RegressionThreshold       float64  `json:"regression_threshold"`
	TestEnvironment           string   `json:"test_environment"`
}

// NewDependencyVersionPolicyManager creates a new dependency version policy manager.
func NewDependencyVersionPolicyManager(logger Logger, apiClient APIClient, dependabotManager *DependabotConfigManager, securityPolicyManager *SecurityUpdatePolicyManager) *DependencyVersionPolicyManager {
	return &DependencyVersionPolicyManager{
		logger:                logger,
		apiClient:             apiClient,
		dependabotManager:     dependabotManager,
		securityPolicyManager: securityPolicyManager,
		policies:              make(map[string]*DependencyVersionPolicy),
		versionConstraints:    NewVersionConstraintEngine(logger),
	}
}

// NewVersionConstraintEngine creates a new version constraint engine.
func NewVersionConstraintEngine(logger Logger) *VersionConstraintEngine {
	return &VersionConstraintEngine{
		logger:          logger,
		semverParser:    NewSemverParser(),
		constraintCache: make(map[string]*ConstraintEvaluationResult),
		cacheTTL:        time.Hour,
	}
}

// CreateDependencyVersionPolicy creates a new dependency version policy.
func (dvm *DependencyVersionPolicyManager) CreateDependencyVersionPolicy(ctx context.Context, policy *DependencyVersionPolicy) error {
	dvm.logger.Info("Creating dependency version policy", "organization", policy.Organization, "policy", policy.Name)

	// Validate policy
	if err := dvm.validateDependencyVersionPolicy(policy); err != nil {
		return fmt.Errorf("invalid dependency version policy: %w", err)
	}

	// Set metadata
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()
	policy.Version = 1

	// Store policy
	dvm.policies[policy.ID] = policy

	dvm.logger.Info("Dependency version policy created successfully", "policy_id", policy.ID)

	return nil
}

// AnalyzeDependencyVersionUpdate analyzes a proposed dependency version update.
func (dvm *DependencyVersionPolicyManager) AnalyzeDependencyVersionUpdate(ctx context.Context, policyID string, dependencyName, currentVersion, proposedVersion, ecosystem string) (*DependencyVersionAnalysis, error) {
	dvm.logger.Debug("Analyzing dependency version update",
		"policy_id", policyID,
		"dependency", dependencyName,
		"current_version", currentVersion,
		"proposed_version", proposedVersion,
		"ecosystem", ecosystem)

	policy, exists := dvm.policies[policyID]
	if !exists {
		return nil, fmt.Errorf("dependency version policy not found: %s", policyID)
	}

	if !policy.Enabled {
		return &DependencyVersionAnalysis{
			DependencyName:  dependencyName,
			Ecosystem:       ecosystem,
			CurrentVersion:  currentVersion,
			ProposedVersion: proposedVersion,
			RecommendedAction: RecommendedAction{
				Action: "skip",
				Reason: "Dependency version policy is disabled",
			},
		}, nil
	}

	analysis := &DependencyVersionAnalysis{
		DependencyName:  dependencyName,
		Ecosystem:       ecosystem,
		CurrentVersion:  currentVersion,
		ProposedVersion: proposedVersion,
	}

	// Determine update type
	analysis.UpdateType = dvm.determineUpdateType(currentVersion, proposedVersion)

	// Check version constraints
	constraintResult, err := dvm.checkVersionConstraints(policy, dependencyName, proposedVersion, ecosystem)
	if err != nil {
		return nil, fmt.Errorf("failed to check version constraints: %w", err)
	}

	analysis.VersionConstraintCheck = *constraintResult

	// Perform compatibility analysis
	compatibilityResult, err := dvm.performCompatibilityAnalysis(policy, dependencyName, currentVersion, proposedVersion, ecosystem)
	if err != nil {
		return nil, fmt.Errorf("failed to perform compatibility analysis: %w", err)
	}

	analysis.CompatibilityAnalysis = *compatibilityResult

	// Analyze security impact
	securityImpact, err := dvm.analyzeSecurityImpact(dependencyName, currentVersion, proposedVersion, ecosystem)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze security impact: %w", err)
	}

	analysis.SecurityImpact = *securityImpact

	// Analyze breaking changes
	breakingChangeResult, err := dvm.analyzeBreakingChanges(policy, dependencyName, currentVersion, proposedVersion, ecosystem)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze breaking changes: %w", err)
	}

	analysis.BreakingChangeAnalysis = *breakingChangeResult

	// Determine recommended action
	analysis.RecommendedAction = dvm.determineRecommendedAction(analysis)

	// Generate approval workflow
	analysis.ApprovalWorkflow = dvm.generateApprovalWorkflow(policy, analysis)

	dvm.logger.Info("Dependency version analysis completed",
		"dependency", dependencyName,
		"recommended_action", analysis.RecommendedAction.Action,
		"risk_level", analysis.RiskAssessment.OverallRisk)

	return analysis, nil
}

// ApplyVersionConstraints applies version constraints to a list of dependency updates.
func (dvm *DependencyVersionPolicyManager) ApplyVersionConstraints(ctx context.Context, policyID string, updates []DependencyUpdate) (*VersionConstraintApplicationResult, error) {
	dvm.logger.Info("Applying version constraints", "policy_id", policyID, "updates_count", len(updates))

	_, exists := dvm.policies[policyID]
	if !exists {
		return nil, fmt.Errorf("dependency version policy not found: %s", policyID)
	}

	result := &VersionConstraintApplicationResult{
		PolicyID:        policyID,
		TotalUpdates:    len(updates),
		ApprovedUpdates: make([]DependencyUpdate, 0),
		RejectedUpdates: make([]DependencyUpdateRejection, 0),
		PendingReview:   make([]DependencyUpdate, 0),
		ProcessedAt:     time.Now(),
	}

	for _, update := range updates {
		analysis, err := dvm.AnalyzeDependencyVersionUpdate(ctx, policyID, update.Name, update.CurrentVersion, update.ProposedVersion, update.Ecosystem)
		if err != nil {
			dvm.logger.Error("Failed to analyze dependency update", "dependency", update.Name, "error", err)
			result.RejectedUpdates = append(result.RejectedUpdates, DependencyUpdateRejection{
				Update: update,
				Reason: fmt.Sprintf("Analysis failed: %s", err.Error()),
			})

			continue
		}

		switch analysis.RecommendedAction.Action {
		case "approve":
			result.ApprovedUpdates = append(result.ApprovedUpdates, update)
		case "reject":
			result.RejectedUpdates = append(result.RejectedUpdates, DependencyUpdateRejection{
				Update: update,
				Reason: analysis.RecommendedAction.Reason,
			})
		case "review":
			result.PendingReview = append(result.PendingReview, update)
		}
	}

	result.ApprovedCount = len(result.ApprovedUpdates)
	result.RejectedCount = len(result.RejectedUpdates)
	result.PendingReviewCount = len(result.PendingReview)

	dvm.logger.Info("Version constraints applied",
		"total", result.TotalUpdates,
		"approved", result.ApprovedCount,
		"rejected", result.RejectedCount,
		"pending_review", result.PendingReviewCount)

	return result, nil
}

// Helper methods

func (dvm *DependencyVersionPolicyManager) validateDependencyVersionPolicy(policy *DependencyVersionPolicy) error {
	if policy.ID == "" {
		return fmt.Errorf("policy ID is required")
	}

	if policy.Name == "" {
		return fmt.Errorf("policy name is required")
	}

	if policy.Organization == "" {
		return fmt.Errorf("organization is required")
	}

	// Validate version constraints
	for constraintID := range policy.VersionConstraints {
		constraint := policy.VersionConstraints[constraintID]
		if err := dvm.validateVersionConstraintRule(&constraint); err != nil {
			return fmt.Errorf("invalid version constraint %s: %w", constraintID, err)
		}
	}

	// Validate ecosystem policies
	for ecosystem := range policy.EcosystemPolicies {
		ecosystemPolicy := policy.EcosystemPolicies[ecosystem]
		if err := dvm.validateEcosystemVersionPolicy(&ecosystemPolicy); err != nil {
			return fmt.Errorf("invalid ecosystem policy for %s: %w", ecosystem, err)
		}
	}

	return nil
}

func (dvm *DependencyVersionPolicyManager) validateVersionConstraintRule(rule *VersionConstraintRule) error {
	if rule.RuleID == "" {
		return fmt.Errorf("rule ID is required")
	}

	if rule.DependencyPattern == "" {
		return fmt.Errorf("dependency pattern is required")
	}

	if rule.Ecosystem == "" {
		return fmt.Errorf("ecosystem is required")
	}

	// Validate version pattern if provided
	if rule.VersionPattern != "" {
		_, err := regexp.Compile(rule.VersionPattern)
		if err != nil {
			return fmt.Errorf("invalid version pattern: %w", err)
		}
	}

	return nil
}

func (dvm *DependencyVersionPolicyManager) validateEcosystemVersionPolicy(policy *EcosystemVersionPolicy) error {
	if policy.Ecosystem == "" {
		return fmt.Errorf("ecosystem is required")
	}

	// Validate performance requirements
	if policy.PerformanceRequirements.MaxPerformanceRegression < 0 || policy.PerformanceRequirements.MaxPerformanceRegression > 1 {
		return fmt.Errorf("max performance regression must be between 0 and 1")
	}

	return nil
}

func (dvm *DependencyVersionPolicyManager) determineUpdateType(currentVersion, proposedVersion string) string {
	// In a real implementation, this would use semantic versioning to determine update type
	if strings.Contains(proposedVersion, "-") {
		return "prerelease"
	}

	// Simple heuristic for now
	currentParts := strings.Split(currentVersion, ".")
	proposedParts := strings.Split(proposedVersion, ".")

	if len(currentParts) >= 1 && len(proposedParts) >= 1 && currentParts[0] != proposedParts[0] {
		return "major"
	}

	if len(currentParts) >= 2 && len(proposedParts) >= 2 && currentParts[1] != proposedParts[1] {
		return "minor"
	}

	return "patch"
}

func (dvm *DependencyVersionPolicyManager) checkVersionConstraints(policy *DependencyVersionPolicy, dependencyName, proposedVersion, ecosystem string) (*VersionConstraintCheckResult, error) {
	result := &VersionConstraintCheckResult{
		DependencyName:      dependencyName,
		ProposedVersion:     proposedVersion,
		Ecosystem:           ecosystem,
		Allowed:             true,
		ViolatedConstraints: make([]string, 0),
	}

	// Check ecosystem-specific policies
	if ecosystemPolicy, exists := policy.EcosystemPolicies[ecosystem]; exists && ecosystemPolicy.Enabled {
		if !ecosystemPolicy.AllowMajorUpdates && dvm.determineUpdateType("1.0.0", proposedVersion) == "major" {
			result.Allowed = false
			result.ViolatedConstraints = append(result.ViolatedConstraints, "Major updates not allowed for ecosystem")
		}
	}

	// Check dependency-specific constraints
	for i := range policy.VersionConstraints {
		constraint := policy.VersionConstraints[i]
		if constraint.Ecosystem == ecosystem {
			matched, err := dvm.matchesDependencyPattern(dependencyName, constraint.DependencyPattern)
			if err != nil {
				return nil, err
			}

			if matched {
				dvm.evaluateVersionConstraint(&constraint, proposedVersion, result)
			}
		}
	}

	return result, nil
}

func (dvm *DependencyVersionPolicyManager) matchesDependencyPattern(dependencyName, pattern string) (bool, error) {
	// Simple pattern matching for now
	if pattern == "*" {
		return true, nil
	}

	// Use regex for pattern matching
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return false, fmt.Errorf("invalid dependency pattern: %w", err)
	}

	return regex.MatchString(dependencyName), nil
}

func (dvm *DependencyVersionPolicyManager) evaluateVersionConstraint(constraint *VersionConstraintRule, proposedVersion string, result *VersionConstraintCheckResult) {
	// Check blocked versions
	for _, blockedRange := range constraint.BlockedVersions {
		if dvm.versionInRange(proposedVersion, blockedRange) {
			result.Allowed = false
			result.ViolatedConstraints = append(result.ViolatedConstraints, fmt.Sprintf("Version %s is in blocked range", proposedVersion))
		}
	}

	// Check minimum version
	if constraint.MinimumVersion != "" {
		if dvm.compareVersions(proposedVersion, constraint.MinimumVersion) < 0 {
			result.Allowed = false
			result.ViolatedConstraints = append(result.ViolatedConstraints, fmt.Sprintf("Version %s is below minimum %s", proposedVersion, constraint.MinimumVersion))
		}
	}

	// Check maximum version
	if constraint.MaximumVersion != "" {
		if dvm.compareVersions(proposedVersion, constraint.MaximumVersion) > 0 {
			result.Allowed = false
			result.ViolatedConstraints = append(result.ViolatedConstraints, fmt.Sprintf("Version %s exceeds maximum %s", proposedVersion, constraint.MaximumVersion))
		}
	}

	// Check prerelease policy
	if !constraint.AllowPrerelease && strings.Contains(proposedVersion, "-") {
		result.Allowed = false
		result.ViolatedConstraints = append(result.ViolatedConstraints, "Prerelease versions not allowed")
	}
}

func (dvm *DependencyVersionPolicyManager) versionInRange(version string, versionRange VersionRange) bool {
	// Simplified version range checking
	if versionRange.Introduced != "" && dvm.compareVersions(version, versionRange.Introduced) < 0 {
		return false
	}

	if versionRange.Fixed != "" && dvm.compareVersions(version, versionRange.Fixed) >= 0 {
		return false
	}

	return true
}

func (dvm *DependencyVersionPolicyManager) compareVersions(v1, v2 string) int {
	// Simplified version comparison
	return strings.Compare(v1, v2)
}

func (dvm *DependencyVersionPolicyManager) performCompatibilityAnalysis(policy *DependencyVersionPolicy, dependencyName, currentVersion, proposedVersion, ecosystem string) (*CompatibilityAnalysisResult, error) {
	result := &CompatibilityAnalysisResult{
		Compatible: true,
		Issues:     make([]CompatibilityIssue, 0),
	}

	// Check if compatibility checks are enabled
	if !policy.CompatibilityChecks.Enabled {
		result.ChecksSkipped = true
		result.Reason = "Compatibility checks disabled"

		return result, nil
	}

	// Perform various compatibility checks
	if policy.CompatibilityChecks.DependencyGraphAnalysis {
		// Mock dependency graph analysis
		result.Issues = append(result.Issues, CompatibilityIssue{
			Type:        "dependency_conflict",
			Severity:    "low",
			Description: "Minor version conflict detected in transitive dependency",
		})
	}

	return result, nil
}

func (dvm *DependencyVersionPolicyManager) analyzeSecurityImpact(dependencyName, currentVersion, proposedVersion, ecosystem string) (*SecurityImpactAnalysis, error) {
	return &SecurityImpactAnalysis{
		SecurityImprovements: true,
		VulnerabilitiesFixed: []string{"CVE-2024-example"},
		NewVulnerabilities:   make([]string, 0),
		SecurityScore:        85.5,
	}, nil
}

func (dvm *DependencyVersionPolicyManager) analyzeBreakingChanges(policy *DependencyVersionPolicy, dependencyName, currentVersion, proposedVersion, ecosystem string) (*BreakingChangeAnalysisResult, error) {
	result := &BreakingChangeAnalysisResult{
		HasBreakingChanges: false,
		DetectedChanges:    make([]DetectedChange, 0),
	}

	if !policy.BreakingChangePolicy.AllowBreakingChanges {
		updateType := dvm.determineUpdateType(currentVersion, proposedVersion)
		if updateType == "major" {
			result.HasBreakingChanges = true
			result.DetectedChanges = append(result.DetectedChanges, DetectedChange{
				Type:        "major_version",
				Description: "Major version update may contain breaking changes",
				Severity:    "medium",
			})
		}
	}

	return result, nil
}

func (dvm *DependencyVersionPolicyManager) determineRecommendedAction(analysis *DependencyVersionAnalysis) RecommendedAction {
	if !analysis.VersionConstraintCheck.Allowed {
		return RecommendedAction{
			Action: "reject",
			Reason: "Version constraint violations: " + strings.Join(analysis.VersionConstraintCheck.ViolatedConstraints, ", "),
		}
	}

	if analysis.BreakingChangeAnalysis.HasBreakingChanges {
		return RecommendedAction{
			Action: "review",
			Reason: "Breaking changes detected, manual review required",
		}
	}

	if analysis.SecurityImpact.SecurityImprovements {
		return RecommendedAction{
			Action: "approve",
			Reason: "Security improvements detected",
		}
	}

	return RecommendedAction{
		Action: "approve",
		Reason: "Update meets all policy requirements",
	}
}

func (dvm *DependencyVersionPolicyManager) generateApprovalWorkflow(policy *DependencyVersionPolicy, analysis *DependencyVersionAnalysis) ApprovalWorkflow {
	workflow := ApprovalWorkflow{
		Required: false,
		Steps:    make([]ApprovalStep, 0),
	}

	updateType := analysis.UpdateType

	var approvalRule VersionApprovalRule

	switch updateType {
	case "major":
		approvalRule = policy.ApprovalRequirements.MajorVersionUpdates
	case "minor":
		approvalRule = policy.ApprovalRequirements.MinorVersionUpdates
	case "patch":
		approvalRule = policy.ApprovalRequirements.PatchVersionUpdates
	default:
		return workflow
	}

	if approvalRule.RequiredApprovers > 0 || approvalRule.ManualReviewRequired {
		workflow.Required = true
		workflow.Steps = append(workflow.Steps, ApprovalStep{
			Type:        "manual_review",
			Description: "Manual review required for version update",
			Approvers:   approvalRule.RequiredApprovalTeams,
			Required:    true,
		})
	}

	return workflow
}

// Supporting types for analysis results.
type VersionConstraintCheckResult struct {
	DependencyName      string    `json:"dependency_name"`
	ProposedVersion     string    `json:"proposed_version"`
	Ecosystem           string    `json:"ecosystem"`
	Allowed             bool      `json:"allowed"`
	ViolatedConstraints []string  `json:"violated_constraints"`
	AppliedRules        []string  `json:"applied_rules"`
	CheckedAt           time.Time `json:"checked_at"`
}

type CompatibilityAnalysisResult struct {
	Compatible    bool                      `json:"compatible"`
	Issues        []CompatibilityIssue      `json:"issues"`
	ChecksSkipped bool                      `json:"checks_skipped"`
	Reason        string                    `json:"reason,omitempty"`
	TestResults   []CompatibilityTestResult `json:"test_results,omitempty"`
}

type CompatibilityIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Solution    string `json:"solution,omitempty"`
}

type SecurityImpactAnalysis struct {
	SecurityImprovements bool     `json:"security_improvements"`
	VulnerabilitiesFixed []string `json:"vulnerabilities_fixed"`
	NewVulnerabilities   []string `json:"new_vulnerabilities"`
	SecurityScore        float64  `json:"security_score"`
	RiskLevel            string   `json:"risk_level"`
}

type BreakingChangeAnalysisResult struct {
	HasBreakingChanges bool             `json:"has_breaking_changes"`
	DetectedChanges    []DetectedChange `json:"detected_changes"`
	ImpactAssessment   string           `json:"impact_assessment"`
	MigrationRequired  bool             `json:"migration_required"`
}

type DetectedChange struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Impact      string `json:"impact,omitempty"`
}

type RecommendedAction struct {
	Action     string                 `json:"action"`
	Reason     string                 `json:"reason"`
	Priority   string                 `json:"priority,omitempty"`
	Timeline   string                 `json:"timeline,omitempty"`
	Conditions []string               `json:"conditions,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type ApprovalWorkflow struct {
	Required             bool           `json:"required"`
	Steps                []ApprovalStep `json:"steps"`
	EstimatedTime        time.Duration  `json:"estimated_time"`
	AutoApprovalEligible bool           `json:"auto_approval_eligible"`
}

type ApprovalStep struct {
	Type        string        `json:"type"`
	Description string        `json:"description"`
	Approvers   []string      `json:"approvers"`
	Required    bool          `json:"required"`
	Timeout     time.Duration `json:"timeout,omitempty"`
}

// Result types.
type DependencyUpdate struct {
	Name            string `json:"name"`
	Ecosystem       string `json:"ecosystem"`
	CurrentVersion  string `json:"current_version"`
	ProposedVersion string `json:"proposed_version"`
}

type DependencyUpdateRejection struct {
	Update DependencyUpdate `json:"update"`
	Reason string           `json:"reason"`
}

type VersionConstraintApplicationResult struct {
	PolicyID           string                      `json:"policy_id"`
	TotalUpdates       int                         `json:"total_updates"`
	ApprovedUpdates    []DependencyUpdate          `json:"approved_updates"`
	RejectedUpdates    []DependencyUpdateRejection `json:"rejected_updates"`
	PendingReview      []DependencyUpdate          `json:"pending_review"`
	ApprovedCount      int                         `json:"approved_count"`
	RejectedCount      int                         `json:"rejected_count"`
	PendingReviewCount int                         `json:"pending_review_count"`
	ProcessedAt        time.Time                   `json:"processed_at"`
}

type ConstraintEvaluationResult struct {
	Satisfied           bool      `json:"satisfied"`
	ViolatedConstraints []string  `json:"violated_constraints"`
	EvaluatedAt         time.Time `json:"evaluated_at"`
}

type SemverParser struct{}

func NewSemverParser() SemverParser {
	return SemverParser{}
}

// Additional supporting types for comprehensive functionality.
type CompatibilityTestResult struct {
	TestName string `json:"test_name"`
	Passed   bool   `json:"passed"`
	Details  string `json:"details,omitempty"`
}

type DependencyRiskAssessment struct {
	OverallRisk     string       `json:"overall_risk"`
	RiskFactors     []RiskFactor `json:"risk_factors"`
	Mitigations     []string     `json:"mitigations"`
	BusinessImpact  string       `json:"business_impact"`
	TechnicalImpact string       `json:"technical_impact"`
}

type TestingRecommendation struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Automated   bool   `json:"automated"`
}

type RollbackPlan struct {
	Supported          bool           `json:"supported"`
	EstimatedTime      time.Duration  `json:"estimated_time"`
	Steps              []RollbackStep `json:"steps"`
	RequiredApprovals  []string       `json:"required_approvals"`
	DataBackupRequired bool           `json:"data_backup_required"`
}

type RollbackStep struct {
	Order       int    `json:"order"`
	Description string `json:"description"`
	Command     string `json:"command,omitempty"`
	Validation  string `json:"validation,omitempty"`
}

type UpdateTimeline struct {
	EstimatedDuration time.Duration   `json:"estimated_duration"`
	Phases            []TimelinePhase `json:"phases"`
	Dependencies      []string        `json:"dependencies,omitempty"`
	Blockers          []string        `json:"blockers,omitempty"`
}

type TimelinePhase struct {
	Name        string        `json:"name"`
	Duration    time.Duration `json:"duration"`
	Description string        `json:"description"`
	Required    bool          `json:"required"`
}

// Additional complex supporting types.
type PerformanceImpactAnalysis struct {
	ExpectedImpact        string            `json:"expected_impact"`
	BenchmarkResults      []BenchmarkResult `json:"benchmark_results"`
	PerformanceRegression float64           `json:"performance_regression"`
	RecommendedActions    []string          `json:"recommended_actions"`
}

type BenchmarkResult struct {
	TestName     string  `json:"test_name"`
	CurrentScore float64 `json:"current_score"`
	NewScore     float64 `json:"new_score"`
	Change       float64 `json:"change"`
	Unit         string  `json:"unit"`
}

type LicenseCompatibilityResult struct {
	Compatible          bool     `json:"compatible"`
	LicenseChanges      []string `json:"license_changes"`
	ConflictingLicenses []string `json:"conflicting_licenses"`
	RequiredActions     []string `json:"required_actions"`
}

// Additional supporting configuration types.
type CommunicationPlan struct {
	Channels             []string      `json:"channels"`
	NotificationTemplate string        `json:"notification_template"`
	EscalationContacts   []string      `json:"escalation_contacts"`
	AdvanceNoticePeriod  time.Duration `json:"advance_notice_period"`
}

type RollbackTrigger struct {
	Type       string                 `json:"type"`
	Condition  string                 `json:"condition"`
	Threshold  float64                `json:"threshold,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

type ManualRollbackProcedure struct {
	Documentation     string   `json:"documentation"`
	RequiredSteps     []string `json:"required_steps"`
	VerificationSteps []string `json:"verification_steps"`
	EmergencyContacts []string `json:"emergency_contacts"`
}

type HealthCheck struct {
	Name     string        `json:"name"`
	Type     string        `json:"type"`
	Endpoint string        `json:"endpoint,omitempty"`
	Timeout  time.Duration `json:"timeout"`
	Required bool          `json:"required"`
}

type DataMigrationHandling struct {
	BackupRequired    bool     `json:"backup_required"`
	RollbackSupported bool     `json:"rollback_supported"`
	MigrationSteps    []string `json:"migration_steps"`
	ValidationSteps   []string `json:"validation_steps"`
}

type RollbackNotificationPlan struct {
	Immediate           []string `json:"immediate"`
	PostRollback        []string `json:"post_rollback"`
	StakeholderUpdate   []string `json:"stakeholder_update"`
	DocumentationUpdate bool     `json:"documentation_update"`
}

type EmergencyApprovalRule struct {
	Enabled               bool          `json:"enabled"`
	EmergencyApprovers    []string      `json:"emergency_approvers"`
	MaxEmergencyDuration  time.Duration `json:"max_emergency_duration"`
	PostEmergencyReview   bool          `json:"post_emergency_review"`
	JustificationRequired bool          `json:"justification_required"`
}

type BulkUpdateApprovalRule struct {
	MaxBulkSize         int           `json:"max_bulk_size"`
	RequiredApprovers   int           `json:"required_approvers"`
	StaggeredDeployment bool          `json:"staggered_deployment"`
	TestingBatchSize    int           `json:"testing_batch_size"`
	CooldownPeriod      time.Duration `json:"cooldown_period"`
}

type AutoApprovalCondition struct {
	Type     string      `json:"type"`
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
	Required bool        `json:"required"`
}

type ApprovalEscalationRule struct {
	TriggerAfter   time.Duration `json:"trigger_after"`
	EscalateTo     []string      `json:"escalate_to"`
	Action         string        `json:"action"`
	MaxEscalations int           `json:"max_escalations"`
}

type TestSuiteConfiguration struct {
	DefaultSuites     []string            `json:"default_suites"`
	EcosystemSpecific map[string][]string `json:"ecosystem_specific"`
	CustomSuites      []CustomTestSuite   `json:"custom_suites"`
}

type CustomTestSuite struct {
	Name        string        `json:"name"`
	Commands    []string      `json:"commands"`
	Environment string        `json:"environment"`
	Timeout     time.Duration `json:"timeout"`
	Required    bool          `json:"required"`
}

type AutomatedTestingConfig struct {
	Enabled               bool          `json:"enabled"`
	TriggerOnUpdate       bool          `json:"trigger_on_update"`
	ParallelExecution     bool          `json:"parallel_execution"`
	MaxConcurrentTests    int           `json:"max_concurrent_tests"`
	TestEnvironments      []string      `json:"test_environments"`
	NotificationOnFailure bool          `json:"notification_on_failure"`
	AutoRetryOnFailure    bool          `json:"auto_retry_on_failure"`
	MaxRetries            int           `json:"max_retries"`
	TestResultsRetention  time.Duration `json:"test_results_retention"`
}

type ManualTestingItem struct {
	ID            string        `json:"id"`
	Description   string        `json:"description"`
	Category      string        `json:"category"`
	Required      bool          `json:"required"`
	EstimatedTime time.Duration `json:"estimated_time"`
}

type TestEnvironment struct {
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	Configuration map[string]string `json:"configuration"`
	Available     bool              `json:"available"`
	Priority      int               `json:"priority"`
}

type TestDataRequirements struct {
	DataSets          []string      `json:"data_sets"`
	SyntheticData     bool          `json:"synthetic_data"`
	ProductionData    bool          `json:"production_data"`
	DataMasking       bool          `json:"data_masking"`
	DataRetention     time.Duration `json:"data_retention"`
	PrivacyCompliance bool          `json:"privacy_compliance"`
}

type ReleaseSchedule struct {
	Type       string        `json:"type"`
	DaysOfWeek []string      `json:"days_of_week,omitempty"`
	TimeOfDay  string        `json:"time_of_day,omitempty"`
	Timezone   string        `json:"timezone"`
	StartDate  time.Time     `json:"start_date,omitempty"`
	EndDate    time.Time     `json:"end_date,omitempty"`
	Frequency  string        `json:"frequency"`
	Duration   time.Duration `json:"duration"`
}

type WindowNotificationSettings struct {
	Enabled              bool          `json:"enabled"`
	AdvanceNotice        time.Duration `json:"advance_notice"`
	ReminderInterval     time.Duration `json:"reminder_interval"`
	NotificationChannels []string      `json:"notification_channels"`
	Recipients           []string      `json:"recipients"`
}

type BlackoutPeriod struct {
	Name        string    `json:"name"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Recurring   bool      `json:"recurring"`
	Description string    `json:"description"`
}

type EmergencyOverride struct {
	Enabled               bool     `json:"enabled"`
	AuthorizedUsers       []string `json:"authorized_users"`
	JustificationRequired bool     `json:"justification_required"`
	AuditTrail            bool     `json:"audit_trail"`
	PostEmergencyReview   bool     `json:"post_emergency_review"`
}

type VersionPolicyNotificationConfig struct {
	Enabled              bool                         `json:"enabled"`
	Channels             []VersionNotificationChannel `json:"channels"`
	EventTypes           []string                     `json:"event_types"`
	NotificationTemplate string                       `json:"notification_template"`
	Frequency            string                       `json:"frequency"`
	Recipients           []NotificationRecipient      `json:"recipients"`
}

type VersionNotificationChannel struct {
	Type        string            `json:"type"`
	Target      string            `json:"target"`
	Enabled     bool              `json:"enabled"`
	EventFilter []string          `json:"event_filter"`
	Template    string            `json:"template,omitempty"`
	Parameters  map[string]string `json:"parameters,omitempty"`
}

type NotificationRecipient struct {
	Type       string   `json:"type"`
	Identifier string   `json:"identifier"`
	EventTypes []string `json:"event_types"`
	Active     bool     `json:"active"`
}

type MetricsTrackingConfig struct {
	Enabled           bool           `json:"enabled"`
	MetricsCollectors []string       `json:"metrics_collectors"`
	TrackingFrequency time.Duration  `json:"tracking_frequency"`
	RetentionPeriod   time.Duration  `json:"retention_period"`
	AlertingEnabled   bool           `json:"alerting_enabled"`
	DashboardEnabled  bool           `json:"dashboard_enabled"`
	CustomMetrics     []CustomMetric `json:"custom_metrics"`
}

type CustomMetric struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Query      string            `json:"query"`
	Threshold  float64           `json:"threshold,omitempty"`
	Alerting   bool              `json:"alerting"`
	Parameters map[string]string `json:"parameters,omitempty"`
}
