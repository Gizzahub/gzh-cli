// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package github

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type webhookListingClient struct {
	APIClient
	list func(context.Context, string) ([]RepositoryInfo, error)
}

func (c webhookListingClient) ListOrganizationRepositories(ctx context.Context, org string) ([]RepositoryInfo, error) {
	return c.list(ctx, org)
}

func TestBulkWebhookTargetsAreResolvedBeforeMutation(t *testing.T) {
	sentinel := errors.New("listing failed")
	for _, mode := range []string{"missing-client", "listing-failed", "invalid-later-name", "empty-org", "actual-list", "explicit"} {
		t.Run(mode, func(t *testing.T) {
			requests := make(chan string, 5)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests <- r.Method + " " + r.URL.Path
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":1,"name":"web"}`))
			}))
			defer server.Close()
			var client APIClient
			queried := 0
			if mode != "missing-client" {
				client = webhookListingClient{list: func(_ context.Context, org string) ([]RepositoryInfo, error) {
					queried++
					require.Equal(t, "actual-org", org)
					switch mode {
					case "listing-failed":
						return nil, sentinel
					case "invalid-later-name":
						return []RepositoryInfo{{Name: "valid"}, {Name: ""}}, nil
					case "empty-org":
						return nil, nil
					default:
						return []RepositoryInfo{{Name: "first-real"}, {Name: "last-real"}}, nil
					}
				}}
			}
			service := NewWebhookServiceWithBaseURL(client, &mockWebhookLogger{}, server.URL)
			req := &BulkWebhookRequest{Organization: "actual-org", Template: WebhookCreateRequest{Name: "hook", Events: []string{"push"}, URL: "https://target.invalid/hook", Config: WebhookConfig{URL: "https://target.invalid/hook"}}}
			if mode == "explicit" {
				req.Repositories = []string{"chosen"}
			}
			result, err := service.BulkCreateWebhooks(context.Background(), req)
			switch mode {
			case "missing-client", "listing-failed", "invalid-later-name":
				require.Error(t, err)
				require.Empty(t, requests)
				if mode == "listing-failed" {
					require.ErrorIs(t, err, sentinel)
				}
			case "empty-org":
				require.NoError(t, err)
				require.Zero(t, result.TotalRepositories)
				require.Empty(t, requests)
			case "actual-list":
				require.NoError(t, err)
				require.Equal(t, 2, result.SuccessCount)
				require.Equal(t, "POST /repos/actual-org/first-real/hooks", <-requests)
				require.Equal(t, "POST /repos/actual-org/last-real/hooks", <-requests)
			case "explicit":
				require.NoError(t, err)
				require.Zero(t, queried)
				require.Equal(t, 1, result.SuccessCount)
				require.Len(t, requests, 1)
				require.Equal(t, "POST /repos/actual-org/chosen/hooks", <-requests)
			}
		})
	}
}

func TestWebhookUpdateURLCompatibility(t *testing.T) {
	for _, configURL := range []string{"", "https://config.invalid/hook"} {
		body := make(chan map[string]any, 2)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var data map[string]any
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				t.Error(err)
			}
			body <- data
			_, _ = w.Write([]byte(`{"id":7}`))
		}))
		service := NewWebhookServiceWithBaseURL(nil, &mockWebhookLogger{}, server.URL)
		request := &WebhookUpdateRequest{ID: 7, URL: "https://legacy.invalid/hook", Config: WebhookConfig{URL: configURL}}
		_, err := service.UpdateRepositoryWebhook(context.Background(), "o", "r", request)
		require.NoError(t, err)
		want := request.URL
		if configURL != "" {
			want = configURL
		}
		require.Equal(t, want, (<-body)["config"].(map[string]any)["url"])
		require.Equal(t, configURL, request.Config.URL)
		_, err = service.UpdateOrganizationWebhook(context.Background(), "o", request)
		require.NoError(t, err)
		require.Equal(t, want, (<-body)["config"].(map[string]any)["url"])
		server.Close()
	}
}

func TestBulkWebhookInvalidNamesFailBeforeAnyMutation(t *testing.T) {
	for _, name := range []string{"", " ", " padded", "../other", "owner/repo", "repo?x"} {
		for _, queried := range []bool{false, true} {
			service := &webhookServiceImpl{apiClient: webhookListingClient{list: func(context.Context, string) ([]RepositoryInfo, error) {
				return []RepositoryInfo{{Name: "valid"}, {Name: name}}, nil
			}}}
			explicit := []string{"valid", name}
			if queried {
				explicit = nil
			}
			_, err := service.bulkRepositories(context.Background(), "org", explicit)
			require.ErrorContains(t, err, "invalid repository name")
		}
	}
}

func TestAllBulkOperationsMissingClientFailWithoutHTTP(t *testing.T) {
	service := NewWebhookService(nil, &mockWebhookLogger{})
	_, err := service.BulkCreateWebhooks(context.Background(), &BulkWebhookRequest{Organization: "org"})
	require.ErrorContains(t, err, "API client is not configured")
	_, err = service.BulkUpdateWebhooks(context.Background(), &BulkWebhookUpdateRequest{Organization: "org"})
	require.ErrorContains(t, err, "API client is not configured")
	_, err = service.BulkDeleteWebhooks(context.Background(), &BulkWebhookDeleteRequest{Organization: "org"})
	require.ErrorContains(t, err, "API client is not configured")
}
