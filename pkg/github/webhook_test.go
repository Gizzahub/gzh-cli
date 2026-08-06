//nolint:testpackage // White-box testing needed for internal function access
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockWebhookLogger implements the Logger interface for testing.
type mockWebhookLogger struct {
	logs []mockWebhookLogEntry
}

type mockWebhookLogEntry struct {
	level  string
	msg    string
	fields []any
}

func (l *mockWebhookLogger) Debug(msg string, fields ...any) {
	l.logs = append(l.logs, mockWebhookLogEntry{"debug", msg, fields})
}

func (l *mockWebhookLogger) Info(msg string, fields ...any) {
	l.logs = append(l.logs, mockWebhookLogEntry{"info", msg, fields})
}

func (l *mockWebhookLogger) Warn(msg string, fields ...any) {
	l.logs = append(l.logs, mockWebhookLogEntry{"warn", msg, fields})
}

func (l *mockWebhookLogger) Error(msg string, fields ...any) {
	l.logs = append(l.logs, mockWebhookLogEntry{"error", msg, fields})
}

// newFakeWebhookService는 가짜 GitHub 웹훅 API를 띄우고 그쪽을 보는
// 서비스를 만든다.
//
// 이 시험들은 원래 캔값을 돌려주던 가짜 구현을 대상으로 쓰였다
// (webhook_test.go에 남은 "Based on the mock implementation" 주석이 그
// 흔적이다). 그 뒤 구현이 실제 HTTP 호출로 바뀌었는데 시험은 그대로
// 남았고, 그 결과 단위 시험이 api.github.com에 진짜 요청을 보내
// "HTTP 401 - Requires authentication"을 받았다. 이어지는 assert.NotNil
// 뒤의 역참조가 nil 포인터로 죽어서 패키지 전체 실행이 그 자리에서
// 멈췄다 -- 뒤쪽 시험들은 아예 돌지 않았다.
//
// NewWebhookService는 baseURL을 https://api.github.com으로 박아 두므로
// 같은 패키지 안에서 구조체를 직접 만든다.
func newFakeWebhookService(t *testing.T) *webhookServiceImpl {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	// hookBody는 GitHub이 실제로 돌려주는 훅 객체 모양이다. name은 언제나
	// "web"이고 url은 훅 자원의 API 주소다. 사용자가 지정한 수신 주소는
	// config.url에 들어간다. insecure_ssl은 불리언이 아니라 문자열이다.
	hookBody := func(id int64, resource, targetURL string, events []string, active bool) map[string]any {
		return map[string]any{
			"type":       "Repository",
			"id":         id,
			"name":       "web",
			"active":     active,
			"events":     events,
			"config":     map[string]any{"url": targetURL, "content_type": "json", "insecure_ssl": "0"},
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-02T00:00:00Z",
			"url":        fmt.Sprintf("%s/%s/%d", server.URL, resource, id),
		}
	}

	writeJSON := func(w http.ResponseWriter, status int, body any) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)

		if err := json.NewEncoder(w).Encode(body); err != nil {
			t.Errorf("가짜 서버 응답을 쓰지 못했다: %v", err)
		}
	}

	// createdHook은 요청 본문을 그대로 반영해 돌려준다. 그래야 시험이
	// 보내는 쪽과 받는 쪽을 함께 확인할 수 있다.
	createdHook := func(w http.ResponseWriter, r *http.Request, id int64, resource string) {
		var body struct {
			Active bool     `json:"active"`
			Events []string `json:"events"`
			Config struct {
				URL string `json:"url"`
			} `json:"config"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("가짜 서버가 요청 본문을 읽지 못했다: %v", err)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		writeJSON(w, http.StatusCreated, hookBody(id, resource, body.Config.URL, body.Events, body.Active))
	}

	mux.HandleFunc("POST /repos/{owner}/{repo}/hooks", func(w http.ResponseWriter, r *http.Request) {
		createdHook(w, r, 1, fmt.Sprintf("repos/%s/%s/hooks", r.PathValue("owner"), r.PathValue("repo")))
	})

	mux.HandleFunc("POST /orgs/{org}/hooks", func(w http.ResponseWriter, r *http.Request) {
		createdHook(w, r, 2, fmt.Sprintf("orgs/%s/hooks", r.PathValue("org")))
	})

	mux.HandleFunc("GET /repos/{owner}/{repo}/hooks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		resource := fmt.Sprintf("repos/%s/%s/hooks", r.PathValue("owner"), r.PathValue("repo"))
		writeJSON(w, http.StatusOK, hookBody(id, resource, "https://example.com/webhook",
			[]string{"push", "pull_request"}, true))
	})

	mux.HandleFunc("GET /repos/{owner}/{repo}/hooks", func(w http.ResponseWriter, r *http.Request) {
		resource := fmt.Sprintf("repos/%s/%s/hooks", r.PathValue("owner"), r.PathValue("repo"))
		writeJSON(w, http.StatusOK, []map[string]any{
			hookBody(123456, resource, "https://ci.example.com/webhook", []string{"push"}, true),
			hookBody(789012, resource, "https://deploy.example.com/webhook", []string{"release"}, true),
		})
	})

	mux.HandleFunc("PATCH /repos/{owner}/{repo}/hooks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		var body struct {
			Active *bool    `json:"active"`
			Events []string `json:"events"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("가짜 서버가 요청 본문을 읽지 못했다: %v", err)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		active := true
		if body.Active != nil {
			active = *body.Active
		}

		resource := fmt.Sprintf("repos/%s/%s/hooks", r.PathValue("owner"), r.PathValue("repo"))
		writeJSON(w, http.StatusOK, hookBody(id, resource, "https://example.com/webhook", body.Events, active))
	})

	mux.HandleFunc("DELETE /repos/{owner}/{repo}/hooks/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 시험이 예상하지 못한 경로를 부르면 조용히 넘기지 않고 알린다.
		t.Errorf("가짜 서버가 모르는 요청을 받았다: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	})

	return &webhookServiceImpl{
		httpClient: server.Client(),
		baseURL:    server.URL,
		token:      "test-token",
		logger:     &mockWebhookLogger{},
	}
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
			service := newFakeWebhookService(t)

			webhook, err := service.CreateRepositoryWebhook(context.Background(), tt.owner, tt.repo, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, webhook)

				return
			}

			// 실패하면 여기서 멈춰야 한다. 예전에는 assert.NotNil이라
			// nil을 확인하고도 다음 줄에서 역참조해 SIGSEGV로 죽었다.
			require.NoError(t, err)
			require.NotNil(t, webhook)

			// GitHub의 훅 객체에서 name은 언제나 "web"이고 url은 훅 자원의
			// API 주소다. 사용자가 지정한 수신 주소는 config.url에 있다.
			// 예전 시험은 요청에 넣은 Name/URL이 그대로 돌아온다고 봤는데,
			// 구현은 그 둘을 API에 보내지도 않는다("name": "web" 고정).
			assert.Equal(t, "web", webhook.Name)
			assert.Equal(t, tt.request.Config.URL, webhook.Config.URL)
			assert.Equal(t, tt.request.Events, webhook.Events)
			assert.Equal(t, tt.request.Active, webhook.Active)
			assert.Equal(t, "testowner/testrepo", webhook.Repository)
			assert.NotZero(t, webhook.ID)
			assert.NotZero(t, webhook.CreatedAt)
			assert.NotZero(t, webhook.UpdatedAt)
		})
	}
}

func TestWebhookService_GetRepositoryWebhook(t *testing.T) {
	service := newFakeWebhookService(t)

	webhook, err := service.GetRepositoryWebhook(context.Background(), "testowner", "testrepo", 123456)

	require.NoError(t, err)
	require.NotNil(t, webhook)
	assert.Equal(t, int64(123456), webhook.ID)
	assert.Equal(t, "web", webhook.Name)
	assert.Equal(t, "https://example.com/webhook", webhook.Config.URL)
	assert.Equal(t, []string{"push", "pull_request"}, webhook.Events)
	assert.True(t, webhook.Active)
	assert.Equal(t, "testowner/testrepo", webhook.Repository)
	// insecure_ssl은 응답에서 문자열 "0"으로 온다. 예전에는 이 값 하나
	// 때문에 디코딩이 통째로 실패했다.
	assert.False(t, webhook.Config.InsecureSSL)
}

func TestWebhookService_ListRepositoryWebhooks(t *testing.T) {
	service := newFakeWebhookService(t)

	webhooks, err := service.ListRepositoryWebhooks(context.Background(), "testowner", "testrepo", nil)

	require.NoError(t, err)
	require.Len(t, webhooks, 2)

	// Verify first webhook
	webhook1 := webhooks[0]
	assert.Equal(t, int64(123456), webhook1.ID)
	assert.Equal(t, "https://ci.example.com/webhook", webhook1.Config.URL)

	// Verify second webhook
	webhook2 := webhooks[1]
	assert.Equal(t, int64(789012), webhook2.ID)
	assert.Equal(t, "https://deploy.example.com/webhook", webhook2.Config.URL)
}

func TestWebhookService_UpdateRepositoryWebhook(t *testing.T) {
	service := newFakeWebhookService(t)

	updateRequest := &WebhookUpdateRequest{
		ID:     123456,
		Name:   "updated-webhook",
		URL:    "https://updated.example.com/webhook",
		Events: []string{"push", "release"},
		Active: boolPtr(false),
	}

	webhook, err := service.UpdateRepositoryWebhook(context.Background(), "testowner", "testrepo", updateRequest)

	require.NoError(t, err)
	require.NotNil(t, webhook)
	assert.Equal(t, updateRequest.ID, webhook.ID)
	assert.Equal(t, updateRequest.Events, webhook.Events)
	assert.Equal(t, *updateRequest.Active, webhook.Active)
	assert.Equal(t, "testowner/testrepo", webhook.Repository)
	// Name과 URL은 단언하지 않는다. UpdateRepositoryWebhook은 이 둘을
	// PATCH 본문에 담지 않으므로 요청에 무엇을 넣든 응답에 반영되지
	// 않는다. 예전 시험이 반영을 기대한 것 자체가 잘못이었다.
}

func TestWebhookService_DeleteRepositoryWebhook(t *testing.T) {
	service := newFakeWebhookService(t)

	err := service.DeleteRepositoryWebhook(context.Background(), "testowner", "testrepo", 123456)

	assert.NoError(t, err)
}

func TestWebhookService_CreateOrganizationWebhook(t *testing.T) {
	service := newFakeWebhookService(t)

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

	require.NoError(t, err)
	require.NotNil(t, webhook)
	assert.Equal(t, "web", webhook.Name)
	assert.Equal(t, request.Config.URL, webhook.Config.URL)
	assert.Equal(t, request.Events, webhook.Events)
	assert.Equal(t, request.Active, webhook.Active)
	assert.Equal(t, "testorg", webhook.Organization)
	assert.NotZero(t, webhook.ID)
}

func TestWebhookService_BulkCreateWebhooks(t *testing.T) {
	service := newFakeWebhookService(t)

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

	require.NoError(t, err)
	require.NotNil(t, result)
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
