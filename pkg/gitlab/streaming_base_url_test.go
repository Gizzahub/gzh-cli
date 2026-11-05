// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package gitlab

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// 테스트 서버 두 개로 명시한 API 주소와 전역 fallback 주소의 선택을 고정한다.
func TestStreamingClientUsesConfiguredBaseURL(t *testing.T) {
	tests := []struct {
		name          string
		clientBaseURL string
		wantBase      string
		wantGlobal    bool
		globalBase    string
	}{
		{
			name:          "explicit web root with trailing slash",
			clientBaseURL: "{server}/",
			wantBase:      "/api/v4",
		},
		{
			name:          "explicit API path with trailing slash",
			clientBaseURL: "{server}/gitlab/api/v4/",
			wantBase:      "/gitlab/api/v4",
		},
		{
			name:       "empty URL preserves configured custom API version",
			wantBase:   "/gitlab/api/v3",
			wantGlobal: true,
			globalBase: "/gitlab/api/v3/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var explicitRequests, globalRequests int
			var requestURI, token string

			newServer := func(onRequest func(*http.Request)) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if onRequest != nil {
						onRequest(r)
					}
					w.Header().Set("X-Total-Pages", "1")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("[]"))
				}))
			}
			explicitServer := newServer(func(r *http.Request) {
				explicitRequests++
				requestURI = r.RequestURI
				token = r.Header.Get("PRIVATE-TOKEN")
			})
			defer explicitServer.Close()
			globalServer := newServer(func(r *http.Request) {
				globalRequests++
				requestURI = r.RequestURI
				token = r.Header.Get("PRIVATE-TOKEN")
			})
			defer globalServer.Close()

			previousBaseURL := configuredBaseAPIURL
			globalBase := tt.globalBase
			if globalBase == "" {
				globalBase = "/api/v4/"
			}
			SetBaseURL(globalServer.URL + globalBase)
			t.Cleanup(func() { configuredBaseAPIURL = previousBaseURL })

			clientBaseURL := tt.clientBaseURL
			if clientBaseURL != "" {
				clientBaseURL = strings.Replace(clientBaseURL, "{server}", explicitServer.URL, 1)
			}
			client := NewStreamingClient("test-token", clientBaseURL, DefaultStreamingConfig())
			client.httpClient = explicitServer.Client()
			if tt.wantGlobal {
				client.httpClient = globalServer.Client()
			}

			_, _, err := client.fetchProjectPage(context.Background(), "engineering/platform", CursorPagination{
				Page:    1,
				PerPage: 25,
			})
			require.NoError(t, err)
			if tt.wantGlobal {
				require.Zero(t, explicitRequests)
				require.Equal(t, 1, globalRequests)
			} else {
				require.Equal(t, 1, explicitRequests)
				require.Zero(t, globalRequests)
			}
			require.Equal(t, tt.wantBase+"/groups/engineering%2Fplatform/projects?per_page=25&page=1", requestURI)
			require.Equal(t, "test-token", token)
		})
	}
}
