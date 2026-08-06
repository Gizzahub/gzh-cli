//nolint:testpackage // White-box testing needed for internal function access
package github

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock storage implementation for testing.
type mockConfigStorage struct {
	mock.Mock
}

func (m *mockConfigStorage) SavePolicy(ctx context.Context, policy *WebhookPolicy) error {
	args := m.Called(ctx, policy)
	return args.Error(0)
}

func (m *mockConfigStorage) GetPolicy(ctx context.Context, org, policyID string) (*WebhookPolicy, error) {
	args := m.Called(ctx, org, policyID)
	if policy, ok := args.Get(0).(*WebhookPolicy); ok {
		return policy, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockConfigStorage) ListPolicies(ctx context.Context, org string) ([]*WebhookPolicy, error) {
	args := m.Called(ctx, org)
	if policies, ok := args.Get(0).([]*WebhookPolicy); ok {
		return policies, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockConfigStorage) DeletePolicy(ctx context.Context, org, policyID string) error {
	args := m.Called(ctx, org, policyID)
	return args.Error(0)
}

func (m *mockConfigStorage) SaveOrganizationConfig(ctx context.Context, config *OrganizationWebhookConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *mockConfigStorage) GetOrganizationConfig(ctx context.Context, org string) (*OrganizationWebhookConfig, error) {
	args := m.Called(ctx, org)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	if config, ok := args.Get(0).(*OrganizationWebhookConfig); ok {
		return config, args.Error(1)
	}
	return nil, args.Error(1)
}

// Test helper functions.
func createTestPolicy() *WebhookPolicy {
	return &WebhookPolicy{
		ID:           "test-policy",
		Name:         "Test Policy",
		Description:  "Test webhook policy",
		Organization: "testorg",
		Enabled:      true,
		Priority:     100,
		Rules: []WebhookPolicyRule{
			{
				ID:      "test-rule",
				Name:    "Test Rule",
				Enabled: true,
				Conditions: WebhookConditions{
					RepositoryName: []string{"test-repo"},
				},
				Action: WebhookActionCreate,
				Template: WebhookTemplate{
					Name:   "test-webhook",
					URL:    "https://test.example.com/webhook",
					Events: []string{"push"},
					Active: true,
					Config: WebhookConfigTemplate{
						URL:         "https://test.example.com/webhook",
						ContentType: "json",
					},
				},
				OnConflict: ConflictResolutionSkip,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: "test-user",
	}
}

func createTestOrganizationConfig() *OrganizationWebhookConfig {
	return &OrganizationWebhookConfig{
		Organization: "testorg",
		Version:      "1.0",
		Metadata: ConfigMetadata{
			Name:        "Test Organization Config",
			Description: "Test configuration",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Defaults: WebhookDefaults{
			Events: []string{"push"},
			Active: true,
			Config: WebhookConfigTemplate{
				ContentType: "json",
			},
		},
		Policies: []WebhookPolicy{*createTestPolicy()},
		Settings: OrganizationWebhookSettings{
			AllowRepositoryOverride: true,
			MaxWebhooksPerRepo:      5,
		},
		Validation: ValidationConfig{
			RequireSSL: true,
		},
	}
}

func TestWebhookConfigurationService_CreatePolicy(t *testing.T) {
	mockStorage := &mockConfigStorage{}
	mockLogger := &mockLogger{}
	service := NewWebhookConfigurationService(nil, nil, mockLogger, mockStorage)

	policy := createTestPolicy()

	mockStorage.On("SavePolicy", mock.Anything, policy).Return(nil)

	err := service.CreatePolicy(context.Background(), policy)

	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}

func TestWebhookConfigurationService_CreatePolicy_InvalidPolicy(t *testing.T) {
	mockStorage := &mockConfigStorage{}
	mockLogger := &mockLogger{}
	service := NewWebhookConfigurationService(nil, nil, mockLogger, mockStorage)

	// Create invalid policy (missing name)
	policy := createTestPolicy()
	policy.Name = ""

	err := service.CreatePolicy(context.Background(), policy)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid policy")
	mockStorage.AssertNotCalled(t, "SavePolicy")
}

func TestWebhookConfigurationService_GetPolicy(t *testing.T) {
	mockStorage := &mockConfigStorage{}
	mockLogger := &mockLogger{}
	service := NewWebhookConfigurationService(nil, nil, mockLogger, mockStorage)

	expectedPolicy := createTestPolicy()
	mockStorage.On("GetPolicy", mock.Anything, "testorg", "test-policy").Return(expectedPolicy, nil)

	policy, err := service.GetPolicy(context.Background(), "testorg", "test-policy")

	assert.NoError(t, err)
	assert.Equal(t, expectedPolicy, policy)
	mockStorage.AssertExpectations(t)
}

func TestWebhookConfigurationService_ListPolicies(t *testing.T) {
	mockStorage := &mockConfigStorage{}
	mockLogger := &mockLogger{}
	service := NewWebhookConfigurationService(nil, nil, mockLogger, mockStorage)

	expectedPolicies := []*WebhookPolicy{createTestPolicy()}
	mockStorage.On("ListPolicies", mock.Anything, "testorg").Return(expectedPolicies, nil)

	policies, err := service.ListPolicies(context.Background(), "testorg")

	assert.NoError(t, err)
	assert.Equal(t, expectedPolicies, policies)
	mockStorage.AssertExpectations(t)
}

func TestWebhookConfigurationService_UpdatePolicy(t *testing.T) {
	mockStorage := &mockConfigStorage{}
	mockLogger := &mockLogger{}
	service := NewWebhookConfigurationService(nil, nil, mockLogger, mockStorage)

	policy := createTestPolicy()
	policy.Description = testUpdatedDescription

	mockStorage.On("SavePolicy", mock.Anything, policy).Return(nil)

	err := service.UpdatePolicy(context.Background(), policy)

	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}

func TestWebhookConfigurationService_DeletePolicy(t *testing.T) {
	mockStorage := &mockConfigStorage{}
	mockLogger := &mockLogger{}
	service := NewWebhookConfigurationService(nil, nil, mockLogger, mockStorage)

	mockStorage.On("DeletePolicy", mock.Anything, "testorg", "test-policy").Return(nil)

	err := service.DeletePolicy(context.Background(), "testorg", "test-policy")

	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}

func TestWebhookConfigurationService_GetOrganizationConfig(t *testing.T) {
	mockStorage := &mockConfigStorage{}
	mockLogger := &mockLogger{}
	service := NewWebhookConfigurationService(nil, nil, mockLogger, mockStorage)

	expectedConfig := createTestOrganizationConfig()
	mockStorage.On("GetOrganizationConfig", mock.Anything, "testorg").Return(expectedConfig, nil)

	config, err := service.GetOrganizationConfig(context.Background(), "testorg")

	assert.NoError(t, err)
	assert.Equal(t, expectedConfig, config)
	mockStorage.AssertExpectations(t)
}

func TestWebhookConfigurationService_GetOrganizationConfig_DefaultConfig(t *testing.T) {
	mockStorage := &mockConfigStorage{}
	mockLogger := &mockLogger{}
	service := NewWebhookConfigurationService(nil, nil, mockLogger, mockStorage)

	// 없음은 ErrOrganizationConfigNotFound로 알린다. 예전에는 아무 오류나
	// (assert.AnError) 기본값으로 대체됐는데, 그러면 저장소 장애까지
	// "설정이 비어 있다"로 둔갑한다.
	mockStorage.On("GetOrganizationConfig", mock.Anything, "testorg").
		Return((*OrganizationWebhookConfig)(nil), ErrOrganizationConfigNotFound)

	config, err := service.GetOrganizationConfig(context.Background(), "testorg")

	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, "testorg", config.Organization)
	mockStorage.AssertExpectations(t)
}

func TestWebhookConfigurationService_GetOrganizationConfig_StorageError(t *testing.T) {
	mockStorage := &mockConfigStorage{}
	mockLogger := &mockLogger{}
	service := NewWebhookConfigurationService(nil, nil, mockLogger, mockStorage)

	// 없음이 아닌 오류는 기본값으로 가리지 않고 그대로 올린다.
	mockStorage.On("GetOrganizationConfig", mock.Anything, "testorg").
		Return((*OrganizationWebhookConfig)(nil), assert.AnError)

	config, err := service.GetOrganizationConfig(context.Background(), "testorg")

	require.Error(t, err)
	assert.Nil(t, config)
	mockStorage.AssertExpectations(t)
}

func TestWebhookConfigurationService_UpdateOrganizationConfig(t *testing.T) {
	mockStorage := &mockConfigStorage{}
	mockLogger := &mockLogger{}
	service := NewWebhookConfigurationService(nil, nil, mockLogger, mockStorage)

	config := createTestOrganizationConfig()
	mockStorage.On("SaveOrganizationConfig", mock.Anything, config).Return(nil)

	err := service.UpdateOrganizationConfig(context.Background(), config)

	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}

func TestWebhookConfigurationService_ValidateConfiguration(t *testing.T) {
	mockStorage := &mockConfigStorage{}
	mockLogger := &mockLogger{}
	service := NewWebhookConfigurationService(nil, nil, mockLogger, mockStorage)

	config := createTestOrganizationConfig()

	result, err := service.ValidateConfiguration(context.Background(), config)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.Equal(t, 100, result.Score)
}

func TestWebhookConfigurationService_ValidateConfiguration_InvalidConfig(t *testing.T) {
	mockStorage := &mockConfigStorage{}
	mockLogger := &mockLogger{}
	service := NewWebhookConfigurationService(nil, nil, mockLogger, mockStorage)

	config := createTestOrganizationConfig()
	config.Organization = "" // Invalid

	result, err := service.ValidateConfiguration(context.Background(), config)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.True(t, len(result.Errors) > 0)
	assert.Equal(t, 0, result.Score)
}

func TestWebhookConfigurationService_ApplyPolicies(t *testing.T) {
	mockStorage := &mockConfigStorage{}
	mockLogger := &mockLogger{}
	service := NewWebhookConfigurationService(nil, nil, mockLogger, mockStorage)

	config := createTestOrganizationConfig()
	policies := []*WebhookPolicy{createTestPolicy()}

	mockStorage.On("GetOrganizationConfig", mock.Anything, "testorg").Return(config, nil)
	mockStorage.On("ListPolicies", mock.Anything, "testorg").Return(policies, nil)

	// 대상 저장소를 직접 준다. test-repo만 정책 조건에 걸리므로 조건
	// 필터가 실제로 동작하는지도 같이 확인된다.
	request := &ApplyPoliciesRequest{
		Organization:    "testorg",
		RepositoryNames: []string{"test-repo", "unrelated-repo"},
		DryRun:          true,
	}

	result, err := service.ApplyPolicies(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "testorg", result.Organization)
	assert.Equal(t, 2, result.TotalRepositories)
	assert.Equal(t, 1, result.SuccessCount)
	require.Len(t, result.Results, 1)
	assert.Equal(t, "test-repo", result.Results[0].Repository)
	mockStorage.AssertExpectations(t)
}

func TestWebhookConfigurationService_ApplyPolicies_ListsOrganizationRepositories(t *testing.T) {
	mockStorage := &mockConfigStorage{}
	mockLogger := &mockLogger{}
	// 대상 저장소를 지정하지 않으면 조직 목록을 API로 가져온다. 예전에는
	// getTargetRepositories가 repo1/repo2/repo3을 지어냈다.
	apiClient := &orgReposAPIClient{repos: []RepositoryInfo{
		{Name: "test-repo", FullName: "testorg/test-repo"},
		{Name: "unrelated-repo", FullName: "testorg/unrelated-repo"},
	}}
	service := NewWebhookConfigurationService(nil, apiClient, mockLogger, mockStorage)

	config := createTestOrganizationConfig()
	policies := []*WebhookPolicy{createTestPolicy()}

	mockStorage.On("GetOrganizationConfig", mock.Anything, "testorg").Return(config, nil)
	mockStorage.On("ListPolicies", mock.Anything, "testorg").Return(policies, nil)

	request := &ApplyPoliciesRequest{Organization: "testorg", DryRun: true}

	result, err := service.ApplyPolicies(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.TotalRepositories)
	require.Len(t, result.Results, 1)
	assert.Equal(t, "test-repo", result.Results[0].Repository)
	mockStorage.AssertExpectations(t)
}

func TestWebhookConfigurationService_ApplyPolicies_NoAPIClient(t *testing.T) {
	mockStorage := &mockConfigStorage{}
	mockLogger := &mockLogger{}
	service := NewWebhookConfigurationService(nil, nil, mockLogger, mockStorage)

	mockStorage.On("GetOrganizationConfig", mock.Anything, "testorg").
		Return(createTestOrganizationConfig(), nil)
	mockStorage.On("ListPolicies", mock.Anything, "testorg").
		Return([]*WebhookPolicy{createTestPolicy()}, nil)

	request := &ApplyPoliciesRequest{Organization: "testorg", DryRun: true}

	// 저장소 목록을 알 수 없으면 지어내지 말고 즉시 실패해야 한다.
	result, err := service.ApplyPolicies(context.Background(), request)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "API client is not configured")
}

func TestWebhookConfigurationService_PreviewPolicyApplication(t *testing.T) {
	mockStorage := &mockConfigStorage{}
	mockLogger := &mockLogger{}
	service := NewWebhookConfigurationService(nil, nil, mockLogger, mockStorage)

	config := createTestOrganizationConfig()
	policies := []*WebhookPolicy{createTestPolicy()}

	mockStorage.On("GetOrganizationConfig", mock.Anything, "testorg").Return(config, nil)
	mockStorage.On("ListPolicies", mock.Anything, "testorg").Return(policies, nil)

	request := &ApplyPoliciesRequest{
		Organization:    "testorg",
		RepositoryNames: []string{"test-repo", "unrelated-repo"},
	}

	preview, err := service.PreviewPolicyApplication(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, preview)
	assert.Equal(t, "testorg", preview.Organization)
	// 조건에 걸리는 저장소만 계획에 오른다.
	require.Len(t, preview.PlannedActions, 1)
	assert.Equal(t, "test-repo", preview.PlannedActions[0].Repository)
	assert.Equal(t, WebhookActionCreate, preview.PlannedActions[0].Action)
	mockStorage.AssertExpectations(t)
}

func TestWebhookConfigurationService_GenerateComplianceReport(t *testing.T) {
	mockStorage := &mockConfigStorage{}
	mockLogger := &mockLogger{}
	apiClient := &orgReposAPIClient{repos: []RepositoryInfo{
		{Name: "test-repo", FullName: "testorg/test-repo"},
	}}
	service := NewWebhookConfigurationService(nil, apiClient, mockLogger, mockStorage)

	config := createTestOrganizationConfig()
	mockStorage.On("GetOrganizationConfig", mock.Anything, "testorg").Return(config, nil)

	report, err := service.GenerateComplianceReport(context.Background(), "testorg")

	require.NoError(t, err)
	require.NotNil(t, report)
	assert.Equal(t, "testorg", report.Organization)
	assert.True(t, report.ComplianceScore >= 0 && report.ComplianceScore <= 100)
	mockStorage.AssertExpectations(t)
}

func TestWebhookConfigurationService_GetWebhookInventory(t *testing.T) {
	mockStorage := &mockConfigStorage{}
	mockLogger := &mockLogger{}
	apiClient := &orgReposAPIClient{repos: []RepositoryInfo{
		{Name: "test-repo", FullName: "testorg/test-repo"},
	}}
	service := NewWebhookConfigurationService(nil, apiClient, mockLogger, mockStorage)

	inventory, err := service.GetWebhookInventory(context.Background(), "testorg")

	require.NoError(t, err)
	require.NotNil(t, inventory)
	assert.Equal(t, "testorg", inventory.Organization)
	assert.NotNil(t, inventory.Summary)
}

func TestValidatePolicy(t *testing.T) {
	mockStorage := &mockConfigStorage{}
	mockLogger := &mockLogger{}
	service := &webhookConfigurationServiceImpl{
		storage: mockStorage,
		logger:  mockLogger,
	}

	tests := []struct {
		name    string
		policy  *WebhookPolicy
		wantErr bool
	}{
		{
			name:    "valid policy",
			policy:  createTestPolicy(),
			wantErr: false,
		},
		{
			name: "policy without name",
			policy: func() *WebhookPolicy {
				p := createTestPolicy()
				p.Name = ""
				return p
			}(),
			wantErr: true,
		},
		{
			name: "policy without organization",
			policy: func() *WebhookPolicy {
				p := createTestPolicy()
				p.Organization = ""
				return p
			}(),
			wantErr: true,
		},
		{
			name: "policy without rules",
			policy: func() *WebhookPolicy {
				p := createTestPolicy()
				p.Rules = []WebhookPolicyRule{}
				return p
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validatePolicy(tt.policy)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRuleAppliesTo(t *testing.T) {
	mockStorage := &mockConfigStorage{}
	mockLogger := &mockLogger{}
	service := &webhookConfigurationServiceImpl{
		storage: mockStorage,
		logger:  mockLogger,
	}

	tests := []struct {
		name       string
		repo       string
		conditions *WebhookConditions
		expected   bool
		expectErr  bool
	}{
		{
			name: "exact repository name match",
			repo: "test-repo",
			conditions: &WebhookConditions{
				RepositoryName: []string{"test-repo", "other-repo"},
			},
			expected: true,
		},
		{
			name: "repository name no match",
			repo: "different-repo",
			conditions: &WebhookConditions{
				RepositoryName: []string{"test-repo", "other-repo"},
			},
			expected: false,
		},
		{
			name: "pattern match",
			repo: "api-service",
			conditions: &WebhookConditions{
				RepositoryPattern: []string{"^api-.*", "^web-.*"},
			},
			expected: true,
		},
		{
			name: "pattern no match",
			repo: "tool-something",
			conditions: &WebhookConditions{
				RepositoryPattern: []string{"^api-.*", "^web-.*"},
			},
			expected: false,
		},
		{
			name:       "no conditions (applies to all)",
			repo:       "any-repo",
			conditions: &WebhookConditions{},
			expected:   true,
		},
		{
			// Go의 정규식 엔진(RE2)에는 전방탐색이 없어 이 무늬는 컴파일되지
			// 않는다. 예전에는 오류를 버려서 "안 걸림"으로 조용히 넘어갔고,
			// 정책이 통째로 적용되지 않아도 보고서에 아무것도 남지 않았다.
			name: "invalid pattern reports an error",
			repo: "api-service",
			conditions: &WebhookConditions{
				RepositoryPattern: []string{"^(?!test-).*"},
			},
			expected:  false,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ruleAppliesTo(tt.repo, tt.conditions)
			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid repository pattern")
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests.
func BenchmarkValidatePolicyWebhook(b *testing.B) {
	mockStorage := &mockConfigStorage{}
	mockLogger := &mockLogger{}
	service := &webhookConfigurationServiceImpl{
		storage: mockStorage,
		logger:  mockLogger,
	}

	policy := createTestPolicy()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = service.validatePolicy(policy)
	}
}

func BenchmarkRuleAppliesTo(b *testing.B) {
	mockStorage := &mockConfigStorage{}
	mockLogger := &mockLogger{}
	service := &webhookConfigurationServiceImpl{
		storage: mockStorage,
		logger:  mockLogger,
	}

	conditions := &WebhookConditions{
		RepositoryName:    []string{"repo1", "repo2", "repo3"},
		RepositoryPattern: []string{"^api-.*", "^web-.*", "^service-.*"},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = service.ruleAppliesTo("api-service", conditions)
	}
}
