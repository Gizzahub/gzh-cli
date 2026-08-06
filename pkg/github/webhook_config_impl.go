package github

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
)

// webhookConfigurationServiceImpl implements WebhookConfigurationService.
type webhookConfigurationServiceImpl struct {
	webhookService WebhookService
	apiClient      APIClient
	logger         Logger
	storage        ConfigStorage // Interface for storing configuration data
}

// ConfigStorage defines the interface for storing webhook configuration data.
type ConfigStorage interface {
	SavePolicy(ctx context.Context, policy *WebhookPolicy) error
	GetPolicy(ctx context.Context, org, policyID string) (*WebhookPolicy, error)
	ListPolicies(ctx context.Context, org string) ([]*WebhookPolicy, error)
	DeletePolicy(ctx context.Context, org, policyID string) error

	SaveOrganizationConfig(ctx context.Context, config *OrganizationWebhookConfig) error
	GetOrganizationConfig(ctx context.Context, org string) (*OrganizationWebhookConfig, error)
}

// NewWebhookConfigurationService creates a new webhook configuration service.
func NewWebhookConfigurationService(webhookService WebhookService, apiClient APIClient, logger Logger, storage ConfigStorage) WebhookConfigurationService {
	return &webhookConfigurationServiceImpl{
		webhookService: webhookService,
		apiClient:      apiClient,
		logger:         logger,
		storage:        storage,
	}
}

// Policy Management

func (w *webhookConfigurationServiceImpl) CreatePolicy(ctx context.Context, policy *WebhookPolicy) error {
	w.logger.Info("Creating webhook policy", "org", policy.Organization, "policy_id", policy.ID)

	// Validate policy
	if err := w.validatePolicy(policy); err != nil {
		return fmt.Errorf("invalid policy: %w", err)
	}

	// Set timestamps
	now := time.Now()
	policy.CreatedAt = now
	policy.UpdatedAt = now

	return w.storage.SavePolicy(ctx, policy)
}

func (w *webhookConfigurationServiceImpl) GetPolicy(ctx context.Context, org, policyID string) (*WebhookPolicy, error) {
	return w.storage.GetPolicy(ctx, org, policyID)
}

func (w *webhookConfigurationServiceImpl) ListPolicies(ctx context.Context, org string) ([]*WebhookPolicy, error) {
	return w.storage.ListPolicies(ctx, org)
}

func (w *webhookConfigurationServiceImpl) UpdatePolicy(ctx context.Context, policy *WebhookPolicy) error {
	w.logger.Info("Updating webhook policy", "org", policy.Organization, "policy_id", policy.ID)

	// Validate policy
	if err := w.validatePolicy(policy); err != nil {
		return fmt.Errorf("invalid policy: %w", err)
	}

	// Update timestamp
	policy.UpdatedAt = time.Now()

	return w.storage.SavePolicy(ctx, policy)
}

func (w *webhookConfigurationServiceImpl) DeletePolicy(ctx context.Context, org, policyID string) error {
	w.logger.Info("Deleting webhook policy", "org", org, "policy_id", policyID)
	return w.storage.DeletePolicy(ctx, org, policyID)
}

// Configuration Management

// ErrOrganizationConfigNotFound는 조직 설정이 아직 저장되지 않았음을
// 알린다. ConfigStorage 구현은 "없음"을 이 오류로 감싸서 알려야 한다.
var ErrOrganizationConfigNotFound = errors.New("organization webhook configuration not found")

func (w *webhookConfigurationServiceImpl) GetOrganizationConfig(ctx context.Context, org string) (*OrganizationWebhookConfig, error) {
	config, err := w.storage.GetOrganizationConfig(ctx, org)
	if err == nil {
		return config, nil
	}

	// 설정이 없을 때만 기본값으로 대신한다.
	//
	// 예전에는 어떤 오류든 기본 설정과 오류를 함께 돌려줬다. 주석은
	// "없으면 기본값"이라고 했지만 오류도 같이 나가므로 호출자는 보통
	// err != nil에서 되돌아가고 기본값은 버려진다. 즉 선언한 대체 동작이
	// 실제로는 일어나지 않았다.
	//
	// 반대로 오류를 통째로 삼키면 저장소 장애(권한 없음, 연결 끊김)가
	// "설정이 비어 있다"로 둔갑한다. 그래서 없음만 골라 대신한다.
	if errors.Is(err, ErrOrganizationConfigNotFound) {
		return w.getDefaultOrganizationConfig(org), nil
	}

	return nil, fmt.Errorf("failed to get organization config for %s: %w", org, err)
}

func (w *webhookConfigurationServiceImpl) UpdateOrganizationConfig(ctx context.Context, config *OrganizationWebhookConfig) error {
	w.logger.Info("Updating organization webhook config", "org", config.Organization)

	// Validate configuration
	validationResult, err := w.ValidateConfiguration(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to validate configuration: %w", err)
	}

	if !validationResult.Valid {
		return fmt.Errorf("configuration validation failed: %v", validationResult.Errors)
	}

	// Update metadata
	config.Metadata.UpdatedAt = time.Now()

	return w.storage.SaveOrganizationConfig(ctx, config)
}

func (w *webhookConfigurationServiceImpl) ValidateConfiguration(ctx context.Context, config *OrganizationWebhookConfig) (*WebhookValidationResult, error) {
	result := &WebhookValidationResult{
		Valid:    true,
		Errors:   []WebhookValidationError{},
		Warnings: []WebhookValidationWarning{},
		Score:    100,
	}

	// Validate organization
	if config.Organization == "" {
		result.Valid = false
		result.Errors = append(result.Errors, WebhookValidationError{
			Field:    "organization",
			Message:  "Organization is required",
			Severity: "error",
		})
	}

	// Validate policies
	for i, policy := range config.Policies {
		if err := w.validatePolicy(&policy); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, WebhookValidationError{
				Field:    fmt.Sprintf("policies[%d]", i),
				Message:  err.Error(),
				Severity: "error",
			})
		}
	}

	// Validate defaults
	if err := w.validateDefaultsTemplate(&config.Defaults.Config); err != nil {
		result.Warnings = append(result.Warnings, WebhookValidationWarning{
			Field:   "defaults.config",
			Message: err.Error(),
		})
	}

	// Calculate final score
	//
	// 감점은 여기 한 곳에서만 한다. 예전에는 위에서 result.Score -= 10을
	// 하고 아래에서 다시 경고 수만큼 빼서 경고 하나에 20점이 깎였다.
	if len(result.Errors) > 0 {
		result.Score = 0
	} else if len(result.Warnings) > 0 {
		result.Score = max(50, result.Score-len(result.Warnings)*10)
	}

	return result, nil
}

// Policy Application

func (w *webhookConfigurationServiceImpl) ApplyPolicies(ctx context.Context, request *ApplyPoliciesRequest) (*ApplyPoliciesResult, error) {
	w.logger.Info("Starting policy application", "org", request.Organization, "dry_run", request.DryRun)

	startTime := time.Now()
	result := &ApplyPoliciesResult{
		Organization: request.Organization,
		Results:      []PolicyApplicationResult{},
		Summary:      PolicyApplicationSummary{},
	}

	// Get organization configuration
	orgConfig, err := w.GetOrganizationConfig(ctx, request.Organization)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization config: %w", err)
	}

	// Get applicable policies
	policies, err := w.getApplicablePolicies(ctx, request.Organization, request.PolicyIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get policies: %w", err)
	}

	// Get target repositories
	repositories, err := w.getTargetRepositories(ctx, request.Organization, request.RepositoryNames)
	if err != nil {
		return nil, fmt.Errorf("failed to get repositories: %w", err)
	}

	result.TotalRepositories = len(repositories)

	// Apply policies to each repository
	for _, repo := range repositories {
		repoResults := w.applyPoliciesToRepository(ctx, repo, policies, orgConfig, request.DryRun, request.Force)
		result.Results = append(result.Results, repoResults...)
		result.ProcessedRepositories++

		// Update counters
		for _, r := range repoResults {
			if r.Success {
				result.SuccessCount++
				w.updateSummaryForAction(&result.Summary, r.Action)
			} else if r.Skipped {
				result.SkippedCount++
			} else {
				result.FailureCount++
			}
		}
	}

	result.ExecutionTime = time.Since(startTime).String()

	w.logger.Info("Completed policy application",
		"org", request.Organization,
		"total", result.TotalRepositories,
		"success", result.SuccessCount,
		"failures", result.FailureCount,
		"skipped", result.SkippedCount)

	return result, nil
}

func (w *webhookConfigurationServiceImpl) PreviewPolicyApplication(ctx context.Context, request *ApplyPoliciesRequest) (*PolicyApplicationPreview, error) {
	w.logger.Info("Generating policy application preview", "org", request.Organization)

	// Force dry run for preview
	previewRequest := *request
	previewRequest.DryRun = true

	result, err := w.ApplyPolicies(ctx, &previewRequest)
	if err != nil {
		return nil, err
	}

	preview := &PolicyApplicationPreview{
		Organization:      result.Organization,
		TotalRepositories: result.TotalRepositories,
		PlannedActions:    []PlannedAction{},
		Conflicts:         []PolicyConflict{},
		Warnings:          []string{},
		Summary:           result.Summary,
	}

	// Convert results to planned actions
	for _, r := range result.Results {
		action := PlannedAction{
			Repository:  r.Repository,
			PolicyID:    r.PolicyID,
			RuleID:      r.RuleID,
			Action:      r.Action,
			WebhookName: fmt.Sprintf("webhook-%s", r.PolicyID),
			Changes:     r.Changes,
		}

		if !r.Success && !r.Skipped {
			action.Conflicts = []string{r.Error}
		}

		preview.PlannedActions = append(preview.PlannedActions, action)
	}

	return preview, nil
}

// Migration and Sync

func (w *webhookConfigurationServiceImpl) MigrateExistingWebhooks(ctx context.Context, request *MigrationRequest) (*MigrationResult, error) {
	w.logger.Info("Starting webhook migration", "org", request.Organization, "policy_id", request.TargetPolicyID)

	startTime := time.Now()
	result := &MigrationResult{
		Organization: request.Organization,
		Results:      []WebhookMigrationResult{},
	}

	// Get target policy
	policy, err := w.GetPolicy(ctx, request.Organization, request.TargetPolicyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target policy: %w", err)
	}

	// Get all repositories in organization
	repositories, err := w.getTargetRepositories(ctx, request.Organization, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get repositories: %w", err)
	}

	// Migrate webhooks for each repository
	for _, repo := range repositories {
		migrationResults := w.migrateRepositoryWebhooks(ctx, repo, policy, request)
		result.Results = append(result.Results, migrationResults...)

		// Update counters
		for _, r := range migrationResults {
			result.TotalWebhooks++
			if r.Success {
				result.MigratedWebhooks++
			} else {
				result.FailedWebhooks++
			}
		}
	}

	result.ExecutionTime = time.Since(startTime).String()

	return result, nil
}

func (w *webhookConfigurationServiceImpl) SyncOrganizationWebhooks(ctx context.Context, org string) (*SyncResult, error) {
	w.logger.Info("Starting webhook synchronization", "org", org)

	startTime := time.Now()
	result := &SyncResult{
		Organization:  org,
		Discrepancies: []WebhookDiscrepancy{},
	}

	// Get organization config
	orgConfig, err := w.GetOrganizationConfig(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization config: %w", err)
	}

	// Get all repositories
	repositories, err := w.getTargetRepositories(ctx, org, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get repositories: %w", err)
	}

	result.TotalRepositories = len(repositories)

	// Check each repository for discrepancies
	for _, repo := range repositories {
		discrepancies := w.checkRepositoryCompliance(ctx, repo, orgConfig)

		result.Discrepancies = append(result.Discrepancies, discrepancies...)
		if len(discrepancies) == 0 {
			result.SyncedRepositories++
		}
	}

	result.ExecutionTime = time.Since(startTime).String()

	return result, nil
}

// Reporting and Audit

func (w *webhookConfigurationServiceImpl) GenerateComplianceReport(ctx context.Context, org string) (*ComplianceReport, error) {
	w.logger.Info("Generating compliance report", "org", org)

	report := &ComplianceReport{
		Organization:    org,
		GeneratedAt:     time.Now(),
		Violations:      []ComplianceViolation{},
		Recommendations: []string{},
	}

	// Get organization config
	orgConfig, err := w.GetOrganizationConfig(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization config: %w", err)
	}

	// Get all repositories
	repositories, err := w.getTargetRepositories(ctx, org, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get repositories: %w", err)
	}

	report.TotalRepositories = len(repositories)

	// Check compliance for each repository
	for _, repo := range repositories {
		violations := w.checkRepositoryViolations(ctx, repo, orgConfig)
		if len(violations) == 0 {
			report.CompliantRepos++
		} else {
			report.NonCompliantRepos++
			report.Violations = append(report.Violations, violations...)
		}
	}

	// Calculate compliance score
	if report.TotalRepositories > 0 {
		report.ComplianceScore = float64(report.CompliantRepos) / float64(report.TotalRepositories) * 100
	}

	// Generate recommendations
	report.Recommendations = w.generateComplianceRecommendations(report)

	return report, nil
}

func (w *webhookConfigurationServiceImpl) GetWebhookInventory(ctx context.Context, org string) (*WebhookInventory, error) {
	w.logger.Info("Generating webhook inventory", "org", org)

	inventory := &WebhookInventory{
		Organization:    org,
		GeneratedAt:     time.Now(),
		WebhooksByType:  make(map[string]int),
		WebhooksByEvent: make(map[string]int),
		Repositories:    []RepositoryWebhookInfo{},
	}

	// Get all repositories
	repositories, err := w.getTargetRepositories(ctx, org, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get repositories: %w", err)
	}

	// Collect webhook information for each repository
	for _, repo := range repositories {
		repoInfo := w.collectRepositoryWebhookInfo(ctx, org, repo)
		inventory.Repositories = append(inventory.Repositories, repoInfo)

		// Update counters
		for _, webhook := range repoInfo.Webhooks {
			inventory.TotalWebhooks++

			// Count by type (based on URL domain)
			webhookType := w.categorizeWebhook(webhook)
			inventory.WebhooksByType[webhookType]++

			// Count by events
			for _, event := range webhook.Events {
				inventory.WebhooksByEvent[event]++
			}
		}
	}

	// Generate summary
	inventory.Summary = w.generateInventorySummary(inventory)

	return inventory, nil
}

// Helper methods

func (w *webhookConfigurationServiceImpl) validatePolicy(policy *WebhookPolicy) error {
	if policy.Name == "" {
		return fmt.Errorf("policy name is required")
	}

	if policy.Organization == "" {
		return fmt.Errorf("organization is required")
	}

	if len(policy.Rules) == 0 {
		return fmt.Errorf("at least one rule is required")
	}

	for i, rule := range policy.Rules {
		if err := w.validatePolicyRule(&rule); err != nil {
			return fmt.Errorf("rule %d: %w", i, err)
		}
	}

	return nil
}

func (w *webhookConfigurationServiceImpl) validatePolicyRule(rule *WebhookPolicyRule) error {
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}

	if rule.Action == "" {
		return fmt.Errorf("action is required")
	}

	if rule.Action == WebhookActionCreate || rule.Action == WebhookActionEnsure {
		if err := w.validateWebhookTemplate(&rule.Template.Config); err != nil {
			return fmt.Errorf("invalid template: %w", err)
		}
	}

	return nil
}

// validateDefaultsTemplate는 조직 기본값 템플릿을 검사한다.
//
// 여기서는 URL이 필수가 아니다. 기본값 템플릿은 개별 웹훅이 채우지 않은
// 항목만 메우는 자리이고 수신 주소는 웹훅마다 다르다. 실제로 이 파일의
// getDefaultOrganizationConfig가 만드는 기본 설정에도 URL이 없어서, 예전
// 규칙대로면 제품이 스스로 내놓은 기본 설정이 언제나 자기 검사에서 경고를
// 받았다. 규칙이 웹훅을 새로 만들 때(validatePolicyRule)는 주소가 정말
// 필요하므로 그쪽은 validateWebhookTemplate을 그대로 쓴다.
func (w *webhookConfigurationServiceImpl) validateDefaultsTemplate(config *WebhookConfigTemplate) error {
	if config.URL == "" {
		return nil
	}

	return w.validateWebhookTemplate(config)
}

func (w *webhookConfigurationServiceImpl) validateWebhookTemplate(config *WebhookConfigTemplate) error {
	if config.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}

	// Validate URL format
	if !strings.HasPrefix(config.URL, "http://") && !strings.HasPrefix(config.URL, "https://") {
		return fmt.Errorf("webhook URL must start with http:// or https://")
	}

	return nil
}

func (w *webhookConfigurationServiceImpl) getDefaultOrganizationConfig(org string) *OrganizationWebhookConfig {
	return &OrganizationWebhookConfig{
		Organization: org,
		Version:      "1.0",
		Metadata: ConfigMetadata{
			Name:        fmt.Sprintf("%s Webhook Configuration", org),
			Description: "Default webhook configuration",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Version:     "1.0",
		},
		Defaults: WebhookDefaults{
			Events: []string{"push", "pull_request"},
			Active: true,
			Config: WebhookConfigTemplate{
				ContentType: "json",
				InsecureSSL: false,
			},
		},
		Policies: []WebhookPolicy{},
		Settings: OrganizationWebhookSettings{
			AllowRepositoryOverride: true,
			RequireApproval:         false,
			MaxWebhooksPerRepo:      5,
			RetryOnFailure:          true,
		},
		Validation: ValidationConfig{
			RequireSSL:    true,
			RequireSecret: false,
		},
	}
}

func (w *webhookConfigurationServiceImpl) getApplicablePolicies(ctx context.Context, org string, policyIDs []string) ([]*WebhookPolicy, error) {
	if len(policyIDs) > 0 {
		// Get specific policies
		policies := make([]*WebhookPolicy, 0, len(policyIDs))
		for _, id := range policyIDs {
			policy, err := w.GetPolicy(ctx, org, id)
			if err != nil {
				return nil, fmt.Errorf("failed to get policy %s: %w", id, err)
			}

			if policy.Enabled {
				policies = append(policies, policy)
			}
		}

		return policies, nil
	}

	// Get all enabled policies
	allPolicies, err := w.ListPolicies(ctx, org)
	if err != nil {
		return nil, err
	}

	enabledPolicies := make([]*WebhookPolicy, 0)

	for _, policy := range allPolicies {
		if policy.Enabled {
			enabledPolicies = append(enabledPolicies, policy)
		}
	}

	return enabledPolicies, nil
}

func (w *webhookConfigurationServiceImpl) getTargetRepositories(ctx context.Context, org string, repoNames []string) ([]string, error) {
	if len(repoNames) > 0 {
		return repoNames, nil
	}

	// 조직 전체를 대상으로 할 때는 실제 목록을 가져온다.
	//
	// 예전에는 repo1/repo2/repo3을 그대로 돌려줬다. 어떤 조직을 넣어도
	// 저장소가 셋이라고 답했고, ApplyPolicies는 존재하지도 않는 이름에
	// 정책을 적용했다고 보고했다. 있지도 않은 결과를 지어내는 쪽이
	// 아무것도 못 한다고 말하는 쪽보다 훨씬 위험하다.
	if w.apiClient == nil {
		return nil, fmt.Errorf("cannot list repositories of %s: API client is not configured", org)
	}

	repos, err := w.apiClient.ListOrganizationRepositories(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories of %s: %w", org, err)
	}

	names := make([]string, 0, len(repos))
	for _, repo := range repos {
		names = append(names, repo.Name)
	}

	return names, nil
}

func (w *webhookConfigurationServiceImpl) applyPoliciesToRepository(ctx context.Context, repo string, policies []*WebhookPolicy, orgConfig *OrganizationWebhookConfig, dryRun, force bool) []PolicyApplicationResult {
	results := []PolicyApplicationResult{}

	for _, policy := range policies {
		for _, rule := range policy.Rules {
			if !rule.Enabled {
				continue
			}

			// Check if rule applies to this repository
			applies, err := w.ruleAppliesTo(repo, &rule.Conditions)
			if err != nil {
				// 무늬가 깨졌다는 사실을 조용히 넘기지 않는다. 예전에는
				// 컴파일 오류를 버려서 "안 걸림"과 구분되지 않았고, 정책이
				// 통째로 건너뛰어져도 보고서는 깨끗했다.
				results = append(results, PolicyApplicationResult{
					Repository: repo,
					PolicyID:   policy.ID,
					RuleID:     rule.ID,
					Action:     rule.Action,
					Success:    false,
					Error:      err.Error(),
				})

				continue
			}

			if applies {
				result := w.applyRuleToRepository(ctx, repo, policy.ID, &rule, dryRun, force)
				results = append(results, result)
			}
		}
	}

	return results
}

// ruleAppliesTo는 규칙의 조건이 이 저장소에 걸리는지 본다.
//
// 무늬는 정책에 사용자가 적어 넣은 값이라 컴파일되지 않을 수 있다. 오류를
// 버리면 "무늬가 깨졌다"와 "안 걸린다"가 같은 결과가 되어, 정책이 통째로
// 적용되지 않아도 아무 데도 흔적이 남지 않는다. 같은 성격의 판단을 하는
// condition_evaluator.go의 EvaluateRepositoryConditions도 오류를 올려보낸다.
func (w *webhookConfigurationServiceImpl) ruleAppliesTo(repo string, conditions *WebhookConditions) (bool, error) {
	// Check repository name exact match
	if len(conditions.RepositoryName) > 0 {
		found := slices.Contains(conditions.RepositoryName, repo)

		if !found {
			return false, nil
		}
	}

	// Check repository pattern match
	if len(conditions.RepositoryPattern) > 0 {
		found := false

		for _, pattern := range conditions.RepositoryPattern {
			matched, err := regexp.MatchString(pattern, repo)
			if err != nil {
				return false, fmt.Errorf("invalid repository pattern %q: %w", pattern, err)
			}

			if matched {
				found = true
				break
			}
		}

		if !found {
			return false, nil
		}
	}

	// Additional conditions would be checked here (language, topics, etc.)
	// For now, we'll assume the rule applies if basic conditions are met

	return true, nil
}

func (w *webhookConfigurationServiceImpl) applyRuleToRepository(ctx context.Context, repo, policyID string, rule *WebhookPolicyRule, dryRun, force bool) PolicyApplicationResult {
	startTime := time.Now()

	result := PolicyApplicationResult{
		Repository: repo,
		PolicyID:   policyID,
		RuleID:     rule.ID,
		Action:     rule.Action,
		Duration:   time.Since(startTime).String(),
	}

	if dryRun {
		result.Success = true
		result.Changes = []string{fmt.Sprintf("Would %s webhook '%s'", rule.Action, rule.Template.Name)}

		return result
	}

	// Apply the rule (mock implementation)
	switch rule.Action {
	case WebhookActionCreate:
		// Create webhook
		result.Success = true
		result.WebhookID = int64Ptr(123456)
		result.Changes = []string{fmt.Sprintf("Created webhook '%s'", rule.Template.Name)}
	case WebhookActionUpdate:
		// Update webhook
		result.Success = true
		result.Changes = []string{fmt.Sprintf("Updated webhook '%s'", rule.Template.Name)}
	case WebhookActionDelete:
		// Delete webhook
		result.Success = true
		result.Changes = []string{fmt.Sprintf("Deleted webhook '%s'", rule.Template.Name)}
	case WebhookActionEnsure:
		// Ensure webhook exists with correct configuration
		result.Success = true
		result.Changes = []string{fmt.Sprintf("Ensured webhook '%s' exists", rule.Template.Name)}
	}

	result.Duration = time.Since(startTime).String()

	return result
}

func (w *webhookConfigurationServiceImpl) updateSummaryForAction(summary *PolicyApplicationSummary, action WebhookAction) {
	switch action {
	case WebhookActionCreate, WebhookActionEnsure:
		summary.WebhooksCreated++
	case WebhookActionUpdate:
		summary.WebhooksUpdated++
	case WebhookActionDelete:
		summary.WebhooksDeleted++
	}
}

func (w *webhookConfigurationServiceImpl) migrateRepositoryWebhooks(ctx context.Context, repo string, policy *WebhookPolicy, request *MigrationRequest) []WebhookMigrationResult {
	// Mock implementation for webhook migration
	return []WebhookMigrationResult{
		{
			Repository:   repo,
			OldWebhookID: 123,
			NewWebhookID: 456,
			Success:      true,
			Changes:      []string{"Migrated webhook to new policy"},
		},
	}
}

func (w *webhookConfigurationServiceImpl) checkRepositoryCompliance(ctx context.Context, repo string, orgConfig *OrganizationWebhookConfig) []WebhookDiscrepancy {
	// Mock implementation for compliance checking
	return []WebhookDiscrepancy{}
}

func (w *webhookConfigurationServiceImpl) checkRepositoryViolations(ctx context.Context, repo string, orgConfig *OrganizationWebhookConfig) []ComplianceViolation {
	// Mock implementation for violation checking
	return []ComplianceViolation{}
}

func (w *webhookConfigurationServiceImpl) generateComplianceRecommendations(report *ComplianceReport) []string {
	recommendations := []string{}

	if report.ComplianceScore < 80 {
		recommendations = append(recommendations, "Consider implementing standardized webhook policies")
	}

	if len(report.Violations) > 0 {
		recommendations = append(recommendations, "Review and address compliance violations")
	}

	return recommendations
}

func (w *webhookConfigurationServiceImpl) collectRepositoryWebhookInfo(ctx context.Context, org, repo string) RepositoryWebhookInfo {
	// Mock implementation - would call actual webhook service
	return RepositoryWebhookInfo{
		Repository: repo,
		Webhooks:   []*WebhookInfo{},
		Compliance: "compliant",
		Issues:     []string{},
	}
}

func (w *webhookConfigurationServiceImpl) categorizeWebhook(webhook *WebhookInfo) string {
	// Simple categorization based on URL
	if strings.Contains(webhook.URL, "slack") {
		return "slack"
	} else if strings.Contains(webhook.URL, "teams") {
		return "teams"
	} else if strings.Contains(webhook.URL, "discord") {
		return "discord"
	}

	return "other"
}

func (w *webhookConfigurationServiceImpl) generateInventorySummary(inventory *WebhookInventory) WebhookInventorySummary {
	summary := WebhookInventorySummary{}

	for _, repoInfo := range inventory.Repositories {
		for _, webhook := range repoInfo.Webhooks {
			if webhook.Active {
				summary.ActiveWebhooks++
			} else {
				summary.InactiveWebhooks++
			}
		}
	}

	// Calculate health score (simplified)
	if inventory.TotalWebhooks > 0 {
		summary.HealthScore = float64(summary.ActiveWebhooks) / float64(inventory.TotalWebhooks) * 100
	}

	return summary
}

// Helper functions

func int64Ptr(i int64) *int64 {
	return &i
}
