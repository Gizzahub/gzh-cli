// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

//nolint:testpackage // White-box access pins the package's HTTP service behavior.
package github

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWebhookCreateURLCompatibility(t *testing.T) {
	for _, target := range []string{"repository", "organization"} {
		for _, configURL := range []string{"", "https://config.invalid/hook"} {
			t.Run(target+"/config="+configURL, func(t *testing.T) {
				type observation struct {
					method string
					path   string
					url    string
					err    error
				}
				observed := make(chan observation, 1)
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var body struct {
						Config struct {
							URL string `json:"url"`
						} `json:"config"`
					}
					err := json.NewDecoder(r.Body).Decode(&body)
					observed <- observation{method: r.Method, path: r.URL.Path, url: body.Config.URL, err: err}
					if err != nil {
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusCreated)
					_, _ = w.Write([]byte(`{"id":7,"name":"web","url":"https://api.invalid/hook/7","config":{"url":"https://target.invalid/hook","insecure_ssl":"0"}}`))
				}))
				defer server.Close()

				service := NewWebhookServiceWithBaseURL(nil, &mockWebhookLogger{}, server.URL).(*webhookServiceImpl)
				service.httpClient = server.Client()
				request := &WebhookCreateRequest{
					Name:   "deploy",
					URL:    "https://legacy.invalid/hook",
					Events: []string{"push"},
					Active: true,
					Config: WebhookConfig{URL: configURL},
				}

				var err error
				if target == "repository" {
					_, err = service.CreateRepositoryWebhook(context.Background(), "org", "repo", request)
				} else {
					_, err = service.CreateOrganizationWebhook(context.Background(), "org", request)
				}
				require.NoError(t, err)
				got := <-observed
				require.NoError(t, got.err)
				require.Equal(t, http.MethodPost, got.method)
				if target == "repository" {
					require.Equal(t, "/repos/org/repo/hooks", got.path)
				} else {
					require.Equal(t, "/orgs/org/hooks", got.path)
				}
				wantURL := request.URL
				if configURL != "" {
					wantURL = configURL
				}
				require.Equal(t, wantURL, got.url)
			})
		}
	}
}

func TestBulkWebhookListFailuresAreReportedAndLaterTargetsContinue(t *testing.T) {
	for _, operation := range []string{"update", "delete"} {
		t.Run(operation, func(t *testing.T) {
			var goodMutationRequests int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					goodMutationRequests++
				}
				if strings.HasPrefix(r.URL.Path, "/repos/org/bad/") {
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"message":"list failed"}`))
					return
				}
				switch r.Method {
				case http.MethodGet:
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`[ {"id":7,"name":"web","url":"https://api.invalid/hook/7","active":true,"events":["push"],"config":{"url":"https://target.invalid/hook","insecure_ssl":"0"}} ]`))
				case http.MethodPatch:
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"id":7,"name":"web","url":"https://api.invalid/hook/7","active":true,"events":["push"],"config":{"url":"https://target.invalid/hook","insecure_ssl":"0"}}`))
				case http.MethodDelete:
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusMethodNotAllowed)
				}
			}))
			defer server.Close()
			service := NewWebhookServiceWithBaseURL(nil, &mockWebhookLogger{}, server.URL).(*webhookServiceImpl)
			service.httpClient = server.Client()

			var result *BulkWebhookResult
			var err error
			if operation == "update" {
				result, err = service.BulkUpdateWebhooks(context.Background(), &BulkWebhookUpdateRequest{
					Organization: "org", Repositories: []string{"bad", "good"},
					Template: WebhookUpdateRequest{Events: []string{"push"}},
					SelectBy: WebhookSelector{ByName: "web"},
				})
			} else {
				result, err = service.BulkDeleteWebhooks(context.Background(), &BulkWebhookDeleteRequest{
					Organization: "org", Repositories: []string{"bad", "good"},
					SelectBy: WebhookSelector{ByName: "web"},
				})
			}
			require.NoError(t, err)
			require.Equal(t, 2, result.TotalRepositories)
			require.Equal(t, 1, result.FailureCount)
			require.Equal(t, 1, result.SuccessCount)
			require.Len(t, result.Results, 2)
			require.Equal(t, "bad", result.Results[0].Repository)
			require.Equal(t, operation, result.Results[0].Operation)
			require.False(t, result.Results[0].Success)
			require.Contains(t, result.Results[0].Error, "list failed")
			require.Equal(t, "good", result.Results[1].Repository)
			require.True(t, result.Results[1].Success)
			require.Equal(t, 1, goodMutationRequests)
		})
	}
}

func TestBulkWebhookMutationFailuresAreCountedAndLaterTargetsContinue(t *testing.T) {
	for _, operation := range []string{"update", "delete"} {
		t.Run(operation, func(t *testing.T) {
			var successfulMutations int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`[ {"id":7,"name":"web","url":"https://api.invalid/hook/7","active":true,"events":["push"],"config":{"url":"https://target.invalid/hook","insecure_ssl":"0"}} ]`))
					return
				}
				if strings.HasPrefix(r.URL.Path, "/repos/org/bad/") {
					w.WriteHeader(http.StatusUnprocessableEntity)
					_, _ = w.Write([]byte(`{"message":"mutation failed"}`))
					return
				}
				successfulMutations++
				if operation == "update" {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"id":7,"name":"web","url":"https://api.invalid/hook/7","active":true,"events":["push"],"config":{"url":"https://target.invalid/hook","insecure_ssl":"0"}}`))
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()
			service := NewWebhookServiceWithBaseURL(nil, &mockWebhookLogger{}, server.URL).(*webhookServiceImpl)
			service.httpClient = server.Client()

			var result *BulkWebhookResult
			var err error
			if operation == "update" {
				result, err = service.BulkUpdateWebhooks(context.Background(), &BulkWebhookUpdateRequest{
					Organization: "org", Repositories: []string{"bad", "good"},
					Template: WebhookUpdateRequest{Events: []string{"push"}},
					SelectBy: WebhookSelector{ByName: "web"},
				})
			} else {
				result, err = service.BulkDeleteWebhooks(context.Background(), &BulkWebhookDeleteRequest{
					Organization: "org", Repositories: []string{"bad", "good"},
					SelectBy: WebhookSelector{ByName: "web"},
				})
			}
			require.NoError(t, err)
			require.Equal(t, 1, result.FailureCount)
			require.Equal(t, 1, result.SuccessCount)
			require.Len(t, result.Results, 2)
			require.False(t, result.Results[0].Success)
			require.Contains(t, result.Results[0].Error, "mutation failed")
			require.True(t, result.Results[1].Success)
			require.Equal(t, 1, successfulMutations)
		})
	}
}

func TestWebhookCreateReturnsHTTPDecodeAndCancellationErrors(t *testing.T) {
	t.Run("malformed response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("{"))
		}))
		defer server.Close()
		service := NewWebhookServiceWithBaseURL(nil, &mockWebhookLogger{}, server.URL).(*webhookServiceImpl)
		service.httpClient = server.Client()
		_, err := service.CreateRepositoryWebhook(context.Background(), "org", "repo", validWebhookCreateRequest())
		require.ErrorContains(t, err, "failed to decode webhook response")
	})

	t.Run("HTTP rejection includes response body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_, _ = w.Write([]byte("invalid hook"))
		}))
		defer server.Close()
		service := NewWebhookServiceWithBaseURL(nil, &mockWebhookLogger{}, server.URL).(*webhookServiceImpl)
		service.httpClient = server.Client()
		_, err := service.CreateRepositoryWebhook(context.Background(), "org", "repo", validWebhookCreateRequest())
		require.ErrorContains(t, err, "HTTP 422 - invalid hook")
	})

	t.Run("request cancellation is preserved", func(t *testing.T) {
		started := make(chan struct{})
		release := make(chan struct{})
		var releaseOnce sync.Once
		releaseHandler := func() { releaseOnce.Do(func() { close(release) }) }
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := io.Copy(io.Discard, r.Body); err != nil {
				t.Errorf("read fixture request body: %v", err)
			}
			close(started)
			<-release
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()
		defer releaseHandler()
		service := NewWebhookServiceWithBaseURL(nil, &mockWebhookLogger{}, server.URL).(*webhookServiceImpl)
		service.httpClient = server.Client()
		ctx, cancel := context.WithCancel(context.Background())
		resultCh := make(chan error, 1)
		go func() {
			_, err := service.CreateRepositoryWebhook(ctx, "org", "repo", validWebhookCreateRequest())
			resultCh <- err
		}()
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			cancel()
			t.Fatal("request did not reach the fixture")
		}
		cancel()
		releaseHandler()
		select {
		case err := <-resultCh:
			require.ErrorIs(t, err, context.Canceled)
		case <-time.After(2 * time.Second):
			t.Fatal("request did not return after cancellation")
		}
	})
}

func validWebhookCreateRequest() *WebhookCreateRequest {
	return &WebhookCreateRequest{
		Name: "deploy", URL: "https://legacy.invalid/hook", Events: []string{"push"},
		Config: WebhookConfig{ContentType: "json"},
	}
}
