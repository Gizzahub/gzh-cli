// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp/syntax"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/pkg/github"
)

func TestListRepositoriesRejectsInvalidInputsBeforeHTTP(t *testing.T) {
	for _, verbose := range []bool{false, true} {
		t.Run(map[bool]string{false: "quiet", true: "verbose"}[verbose], func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { calls.Add(1); w.WriteHeader(500) }))
			defer server.Close()
			service := NewRepoConfigServiceWithClient(github.NewRepoConfigClientWithBaseURL("fixture", server.URL))
			result, err := service.ListRepositories(context.Background(), ListOptions{Organization: "org", Filter: "[", Verbose: verbose})
			require.Error(t, err)
			require.Nil(t, result)
			var syntaxError *syntax.Error
			require.ErrorAs(t, err, &syntaxError)
			missing := filepath.Join(t.TempDir(), "missing.yaml")
			result, err = service.ListRepositories(context.Background(), ListOptions{Organization: "org", ConfigFile: missing, Verbose: verbose})
			require.ErrorIs(t, err, os.ErrNotExist)
			require.Nil(t, result)
			for _, body := range []string{"bad: [", "version: '1'\n"} {
				path := filepath.Join(t.TempDir(), "invalid.yaml")
				require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
				result, err = service.ListRepositories(context.Background(), ListOptions{Organization: "org", ConfigFile: path, Verbose: verbose})
				require.Error(t, err)
				require.Nil(t, result)
			}
			require.Zero(t, calls.Load())
		})
	}
}

func TestListRepositoriesFilteringAndLimitRemainReadOnly(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		require.Equal(t, "GET", r.Method)
		require.Equal(t, "/orgs/org/repos", r.URL.Path)
		require.NoError(t, json.NewEncoder(w).Encode([]github.Repository{{Name: "api-a"}, {Name: "web"}, {Name: "api-b"}}))
	}))
	defer server.Close()
	service := NewRepoConfigServiceWithClient(github.NewRepoConfigClientWithBaseURL("fixture", server.URL))
	config := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(config, []byte("version: '1'\norganization: org\n"), 0o600))
	result, err := service.ListRepositories(context.Background(), ListOptions{Organization: "org", ConfigFile: config, Filter: "^api-", Limit: 1})
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "api-a", result[0].Name)
	require.Nil(t, result[0].Config)
	require.Equal(t, int32(1), calls.Load(), "ShowConfig=false must not query details")
}

func TestListRepositoriesDetailFailureReturnsNoPartialSuccess(t *testing.T) {
	var third atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		switch r.URL.Path {
		case "/orgs/org/repos":
			require.NoError(t, json.NewEncoder(w).Encode([]github.Repository{{Name: "first"}, {Name: "second"}, {Name: "third"}}))
		case "/repos/org/first":
			require.NoError(t, json.NewEncoder(w).Encode(github.Repository{Name: "first"}))
		case "/repos/org/first/teams", "/repos/org/first/collaborators":
			_, err := w.Write([]byte(`[]`))
			require.NoError(t, err)
		case "/repos/org/second":
			w.WriteHeader(403)
			_, err := w.Write([]byte(`{"message":"Forbidden"}`))
			require.NoError(t, err)
		default:
			third.Add(1)
			w.WriteHeader(500)
		}
	}))
	defer server.Close()
	service := NewRepoConfigServiceWithClient(github.NewRepoConfigClientWithBaseURL("fixture", server.URL))
	result, err := service.ListRepositories(context.Background(), ListOptions{Organization: "org", ShowConfig: true})
	require.ErrorContains(t, err, "org/second")
	require.Nil(t, result)
	var api *github.APIError
	require.ErrorAs(t, err, &api)
	require.Equal(t, 403, api.StatusCode)
	require.Zero(t, third.Load(), "must stop instead of publishing a partial full-success report")
}
