// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package webhook

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/pkg/github"
)

func TestCobraWebhookUpdatePreservesOptionalActive(t *testing.T) {
	for _, active := range []string{"", "false", "true"} {
		t.Run("active="+active, func(t *testing.T) {
			bodies := make(chan map[string]any, 1)
			paths := make(chan string, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				paths <- r.Method + " " + r.URL.Path
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				bodies <- body
				_, _ = w.Write([]byte(`{"id":7,"name":"web"}`))
			}))
			defer server.Close()
			service := github.NewWebhookServiceWithBaseURL(nil, &mockLogger{}, server.URL)
			command := newWebhookCmdWithFactory(func() github.WebhookService { return service })
			command.SetOut(io.Discard)
			command.SetErr(io.Discard)
			args := []string{"repo", "update", "owner", "repo", "7", "--url", "https://new.invalid/hook"}
			if active != "" {
				args = append(args, "--active="+active)
			}
			command.SetArgs(args)
			require.NoError(t, command.Execute())
			require.Equal(t, "PATCH /repos/owner/repo/hooks/7", <-paths)
			body := <-bodies
			require.Equal(t, "https://new.invalid/hook", body["config"].(map[string]any)["url"])
			value, present := body["active"]
			require.Equal(t, active != "", present)
			if present {
				require.Equal(t, active == "true", value)
			}
		})
	}
}

func TestCobraBulkWebhookFailureIsVisible(t *testing.T) {
	for _, mode := range []string{"missing-list", "partial-failure", "success"} {
		t.Run(mode, func(t *testing.T) {
			paths := make(chan string, 3)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				paths <- r.URL.Path
				if r.URL.Path == "/repos/org/bad/hooks" {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":7,"name":"web"}`))
			}))
			defer server.Close()
			service := github.NewWebhookServiceWithBaseURL(nil, &mockLogger{}, server.URL)
			command := newWebhookCmdWithFactory(func() github.WebhookService { return service })
			command.SetOut(io.Discard)
			command.SetErr(io.Discard)
			command.SilenceUsage, command.SilenceErrors = true, true
			args := []string{"bulk", "create", "org", "--name", "fixture", "--url", "https://target.invalid/hook"}
			if mode == "partial-failure" {
				args = append(args, "--repos", "good,bad")
			}
			if mode == "success" {
				args = append(args, "--repos", "good")
			}
			command.SetArgs(args)
			err := command.Execute()
			switch mode {
			case "missing-list":
				require.ErrorContains(t, err, "explicit repositories")
				require.Empty(t, paths)
			case "partial-failure":
				require.ErrorContains(t, err, "1 of 2")
				require.Len(t, paths, 2)
			default:
				require.NoError(t, err)
				require.Len(t, paths, 1)
			}
		})
	}
}

func TestWebhookFactoryUsesExistingTokenPrecedence(t *testing.T) {
	for _, tokens := range [][3]string{{"github-fixture", "gh-fixture", "token github-fixture"}, {"", "gh-fixture", "token gh-fixture"}, {"", "", ""}} {
		t.Setenv("GITHUB_TOKEN", tokens[0])
		t.Setenv("GH_TOKEN", tokens[1])
		auth := make(chan string, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth <- r.Header.Get("Authorization")
			_, _ = w.Write([]byte(`[]`))
		}))
		service := createWebhookServiceWithBaseURL(server.URL)
		_, err := service.ListRepositoryWebhooks(context.Background(), "owner", "repo", nil)
		require.NoError(t, err)
		require.Equal(t, tokens[2], <-auth)
		server.Close()
	}
}
