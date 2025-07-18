package github

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockWebhookLogger implements the Logger interface for testing.
type mockWebhookLogger struct {
	logs []mockWebhookLogEntry
}

type mockWebhookLogEntry struct {
	level  string
	msg    string
	fields []interface{}
}

func (l *mockWebhookLogger) Debug(msg string, fields ...interface{}) {
	l.logs = append(l.logs, mockWebhookLogEntry{"debug", msg, fields})
}

func (l *mockWebhookLogger) Info(msg string, fields ...interface{}) {
	l.logs = append(l.logs, mockWebhookLogEntry{"info", msg, fields})
}

func (l *mockWebhookLogger) Warn(msg string, fields ...interface{}) {
	l.logs = append(l.logs, mockWebhookLogEntry{"warn", msg, fields})
}

func (l *mockWebhookLogger) Error(msg string, fields ...interface{}) {
	l.logs = append(l.logs, mockWebhookLogEntry{"error", msg, fields})
}

func TestWebhookService_CreateRepositoryWebhook(t *testing.T) {
	tests := []struct {
		name    string
		owner   string
		repo    string
		request *WebhookCreateRequest
		wantErr bool
	}{
		{
			name:  "valid webhook creation",
			owner: "testowner",
			repo:  "testrepo",
			request: &WebhookCreateRequest{
				Name:   "test-webhook",
				URL:    "https://example.com/webhook",
				Events: []string{"push", "pull_request"},
				Active: true,
				Config: WebhookConfig{
					URL:         "https://example.com/webhook",
					ContentType: "json",
				},
			},
			wantErr: false,
		},
		{
			name:  "missing webhook name",
			owner: "testowner",
			repo:  "testrepo",
			request: &WebhookCreateRequest{
				URL:    "https://example.com/webhook",
				Events: []string{"push"},
				Active: true,
			},
			wantErr: true,
		},
		{
			name:  "missing webhook URL",
			owner: "testowner",
			repo:  "testrepo",
			request: &WebhookCreateRequest{
				Name:   "test-webhook",
				Events: []string{"push"},
				Active: true,
			},
			wantErr: true,
		},
		{
			name:  "missing events",
			owner: "testowner",
			repo:  "testrepo",
			request: &WebhookCreateRequest{
				Name:   "test-webhook",
				URL:    "https://example.com/webhook",
				Events: []string{},
				Active: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &mockWebhookLogger{}
			service := NewWebhookService(nil, logger)

			webhook, err := service.CreateRepositoryWebhook(context.Background(), tt.owner, tt.repo, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, webhook)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, webhook)
				assert.Equal(t, tt.request.Name, webhook.Name)
				assert.Equal(t, tt.request.URL, webhook.URL)
				assert.Equal(t, tt.request.Events, webhook.Events)
				assert.Equal(t, tt.request.Active, webhook.Active)
				assert.NotZero(t, webhook.ID)
				assert.NotZero(t, webhook.CreatedAt)
				assert.NotZero(t, webhook.UpdatedAt)
			}
		})
	}
}

func TestWebhookService_GetRepositoryWebhook(t *testing.T) {
	logger := &mockWebhookLogger{}
	service := NewWebhookService(nil, logger)

	webhook, err := service.GetRepositoryWebhook(context.Background(), "testowner", "testrepo", 123456)

	assert.NoError(t, err)
	assert.NotNil(t, webhook)
	assert.Equal(t, int64(123456), webhook.ID)
	assert.Equal(t, "example-webhook", webhook.Name)
	assert.Equal(t, "https://example.com/webhook", webhook.URL)
	assert.Equal(t, []string{"push", "pull_request"}, webhook.Events)
	assert.True(t, webhook.Active)
	assert.Equal(t, "testowner/testrepo", webhook.Repository)
}

func TestWebhookService_ListRepositoryWebhooks(t *testing.T) {
	logger := &mockWebhookLogger{}
	service := NewWebhookService(nil, logger)

	webhooks, err := service.ListRepositoryWebhooks(context.Background(), "testowner", "testrepo", nil)

	assert.NoError(t, err)
	assert.NotNil(t, webhooks)
	assert.Len(t, webhooks, 2) // Based on the mock implementation

	// Verify first webhook
	webhook1 := webhooks[0]
	assert.Equal(t, int64(123456), webhook1.ID)
	assert.Equal(t, "ci-webhook", webhook1.Name)
	assert.Equal(t, "https://ci.example.com/webhook", webhook1.URL)

	// Verify second webhook
	webhook2 := webhooks[1]
	assert.Equal(t, int64(789012), webhook2.ID)
	assert.Equal(t, "deploy-webhook", webhook2.Name)
	assert.Equal(t, "https://deploy.example.com/webhook", webhook2.URL)
}

func TestWebhookService_UpdateRepositoryWebhook(t *testing.T) {
	logger := &mockWebhookLogger{}
	service := NewWebhookService(nil, logger)

	updateRequest := &WebhookUpdateRequest{
		ID:     123456,
		Name:   "updated-webhook",
		URL:    "https://updated.example.com/webhook",
		Events: []string{"push", "release"},
		Active: boolPtr(false),
	}

	webhook, err := service.UpdateRepositoryWebhook(context.Background(), "testowner", "testrepo", updateRequest)

	assert.NoError(t, err)
	assert.NotNil(t, webhook)
	assert.Equal(t, updateRequest.ID, webhook.ID)
	assert.Equal(t, updateRequest.Name, webhook.Name)
	assert.Equal(t, updateRequest.URL, webhook.URL)
	assert.Equal(t, updateRequest.Events, webhook.Events)
	assert.Equal(t, *updateRequest.Active, webhook.Active)
}

func TestWebhookService_DeleteRepositoryWebhook(t *testing.T) {
	logger := &mockWebhookLogger{}
	service := NewWebhookService(nil, logger)

	err := service.DeleteRepositoryWebhook(context.Background(), "testowner", "testrepo", 123456)

	assert.NoError(t, err)
}

func TestWebhookService_CreateOrganizationWebhook(t *testing.T) {
	logger := &mockWebhookLogger{}
	service := NewWebhookService(nil, logger)

	request := &WebhookCreateRequest{
		Name:   "org-webhook",
		URL:    "https://org.example.com/webhook",
		Events: []string{"repository", "member"},
		Active: true,
		Config: WebhookConfig{
			URL:         "https://org.example.com/webhook",
			ContentType: "json",
		},
	}

	webhook, err := service.CreateOrganizationWebhook(context.Background(), "testorg", request)

	assert.NoError(t, err)
	assert.NotNil(t, webhook)
	assert.Equal(t, request.Name, webhook.Name)
	assert.Equal(t, request.URL, webhook.URL)
	assert.Equal(t, request.Events, webhook.Events)
	assert.Equal(t, request.Active, webhook.Active)
	assert.Equal(t, "testorg", webhook.Organization)
	assert.NotZero(t, webhook.ID)
}

func TestWebhookService_BulkCreateWebhooks(t *testing.T) {
	logger := &mockWebhookLogger{}
	service := NewWebhookService(nil, logger)

	request := &BulkWebhookRequest{
		Organization: "testorg",
		Repositories: []string{"repo1", "repo2"},
		Template: WebhookCreateRequest{
			Name:   "bulk-webhook",
			URL:    "https://bulk.example.com/webhook",
			Events: []string{"push"},
			Active: true,
			Config: WebhookConfig{
				URL:         "https://bulk.example.com/webhook",
				ContentType: "json",
			},
		},
	}

	result, err := service.BulkCreateWebhooks(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.TotalRepositories)
	assert.Equal(t, 2, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Len(t, result.Results, 2)

	// Verify individual results
	for _, opResult := range result.Results {
		assert.True(t, opResult.Success)
		assert.Equal(t, "create", opResult.Operation)
		assert.NotNil(t, opResult.WebhookInfo)
		assert.Empty(t, opResult.Error)
		assert.NotEmpty(t, opResult.Duration)
	}
}

func TestWebhookService_TestWebhook(t *testing.T) {
	logger := &mockWebhookLogger{}
	service := NewWebhookService(nil, logger)

	result, err := service.TestWebhook(context.Background(), "testowner", "testrepo", 123456)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, 200, result.StatusCode)
	assert.Equal(t, "OK", result.Response)
	assert.NotEmpty(t, result.Duration)
	assert.NotEmpty(t, result.DeliveryID)
	assert.NotZero(t, result.TestedAt)
}

func TestWebhookService_GetWebhookDeliveries(t *testing.T) {
	logger := &mockWebhookLogger{}
	service := NewWebhookService(nil, logger)

	deliveries, err := service.GetWebhookDeliveries(context.Background(), "testowner", "testrepo", 123456)

	assert.NoError(t, err)
	assert.NotNil(t, deliveries)
	assert.Len(t, deliveries, 2) // Based on mock implementation

	// Verify first delivery
	delivery1 := deliveries[0]
	assert.NotEmpty(t, delivery1.ID)
	assert.Equal(t, "push", delivery1.Event)
	assert.Equal(t, "synchronize", delivery1.Action)
	assert.Equal(t, 200, delivery1.StatusCode)
	assert.True(t, delivery1.Success)
	assert.False(t, delivery1.Redelivered)

	// Verify second delivery
	delivery2 := deliveries[1]
	assert.NotEmpty(t, delivery2.ID)
	assert.Equal(t, "pull_request", delivery2.Event)
	assert.Equal(t, "opened", delivery2.Action)
	assert.Equal(t, 200, delivery2.StatusCode)
	assert.True(t, delivery2.Success)
}

func TestWebhookMatchesSelector(t *testing.T) {
	logger := &mockWebhookLogger{}
	service := &webhookServiceImpl{
		apiClient: nil,
		logger:    logger,
	}

	webhook := &WebhookInfo{
		ID:     123,
		Name:   "test-webhook",
		URL:    "https://example.com/webhook",
		Events: []string{"push", "pull_request"},
		Active: true,
	}

	tests := []struct {
		name     string
		selector WebhookSelector
		expected bool
	}{
		{
			name: "match by name",
			selector: WebhookSelector{
				ByName: "test-webhook",
			},
			expected: true,
		},
		{
			name: "no match by name",
			selector: WebhookSelector{
				ByName: "other-webhook",
			},
			expected: false,
		},
		{
			name: "match by URL",
			selector: WebhookSelector{
				ByURL: "https://example.com/webhook",
			},
			expected: true,
		},
		{
			name: "match by active status",
			selector: WebhookSelector{
				Active: boolPtr(true),
			},
			expected: true,
		},
		{
			name: "no match by active status",
			selector: WebhookSelector{
				Active: boolPtr(false),
			},
			expected: false,
		},
		{
			name: "match by events",
			selector: WebhookSelector{
				ByEvents: []string{"push"},
			},
			expected: true,
		},
		{
			name: "no match by events",
			selector: WebhookSelector{
				ByEvents: []string{"release"},
			},
			expected: false,
		},
		{
			name: "multiple criteria match",
			selector: WebhookSelector{
				ByName: "test-webhook",
				Active: boolPtr(true),
			},
			expected: true,
		},
		{
			name: "multiple criteria no match",
			selector: WebhookSelector{
				ByName: "test-webhook",
				Active: boolPtr(false),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.webhookMatchesSelector(webhook, &tt.selector)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateWebhookRequest(t *testing.T) {
	logger := &mockWebhookLogger{}
	service := &webhookServiceImpl{
		apiClient: nil,
		logger:    logger,
	}

	tests := []struct {
		name    string
		request *WebhookCreateRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: &WebhookCreateRequest{
				Name:   "test-webhook",
				URL:    "https://example.com/webhook",
				Events: []string{"push"},
				Active: true,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			request: &WebhookCreateRequest{
				URL:    "https://example.com/webhook",
				Events: []string{"push"},
				Active: true,
			},
			wantErr: true,
			errMsg:  "webhook name is required",
		},
		{
			name: "missing URL",
			request: &WebhookCreateRequest{
				Name:   "test-webhook",
				Events: []string{"push"},
				Active: true,
			},
			wantErr: true,
			errMsg:  "webhook URL is required",
		},
		{
			name: "missing events",
			request: &WebhookCreateRequest{
				Name:   "test-webhook",
				URL:    "https://example.com/webhook",
				Events: []string{},
				Active: true,
			},
			wantErr: true,
			errMsg:  "at least one event must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateWebhookRequest(tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to create bool pointer
// boolPtr is defined in automation_engine.go

// Benchmark tests.
func BenchmarkWebhookService_CreateRepositoryWebhook(b *testing.B) {
	logger := &mockWebhookLogger{}
	service := NewWebhookService(nil, logger)

	request := &WebhookCreateRequest{
		Name:   "benchmark-webhook",
		URL:    "https://benchmark.example.com/webhook",
		Events: []string{"push"},
		Active: true,
		Config: WebhookConfig{
			URL:         "https://benchmark.example.com/webhook",
			ContentType: "json",
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := service.CreateRepositoryWebhook(context.Background(), "testowner", "testrepo", request)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkWebhookService_ListRepositoryWebhooks(b *testing.B) {
	logger := &mockWebhookLogger{}
	service := NewWebhookService(nil, logger)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := service.ListRepositoryWebhooks(context.Background(), "testowner", "testrepo", nil)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkWebhookMatchesSelector(b *testing.B) {
	logger := &mockWebhookLogger{}
	service := &webhookServiceImpl{
		apiClient: nil,
		logger:    logger,
	}

	webhook := &WebhookInfo{
		ID:     123,
		Name:   "test-webhook",
		URL:    "https://example.com/webhook",
		Events: []string{"push", "pull_request"},
		Active: true,
	}

	selector := &WebhookSelector{
		ByName:   "test-webhook",
		ByEvents: []string{"push"},
		Active:   boolPtr(true),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		service.webhookMatchesSelector(webhook, selector)
	}
}
