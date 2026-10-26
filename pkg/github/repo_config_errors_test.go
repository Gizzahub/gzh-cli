// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package github

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// 외부 HTTP 없이 이미 반환된 transport·decode·취소·API 오류의 전파를 확인한다.
type configurationHTTP struct {
	HTTPClient
	do func(context.Context, *http.Request) (*http.Response, error)
}

func (c configurationHTTP) Do(ctx context.Context, request *http.Request) (*http.Response, error) {
	return c.do(ctx, request)
}

func configurationResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Status: http.StatusText(status), Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}

func TestRepositoryConfigurationPropagatesEveryReadError(t *testing.T) {
	for _, stage := range []string{"repository", "protection", "teams", "collaborators"} {
		for _, kind := range []string{"transport", "canceled", "decode", "403", "404"} {
			t.Run(stage+"/"+kind, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				sentinel := errors.New("fixture transport failure")
				reached := false
				client := NewRepoConfigClient("fixture-not-secret")
				client.httpClient = configurationHTTP{do: func(ctx context.Context, request *http.Request) (*http.Response, error) {
					path := request.URL.Path
					current := "repository"
					body := `{"name":"repo","default_branch":"main"}`
					switch {
					case strings.HasSuffix(path, "/protection"):
						current, body = "protection", `{}`
					case strings.HasSuffix(path, "/teams"):
						current, body = "teams", `[]`
					case strings.HasSuffix(path, "/collaborators"):
						current, body = "collaborators", `[]`
					}
					if current != stage {
						return configurationResponse(200, body), nil
					}
					reached = true
					switch kind {
					case "transport":
						return nil, sentinel
					case "canceled":
						cancel()
						return nil, ctx.Err()
					case "decode":
						return configurationResponse(200, "!invalid-json"), nil
					case "403":
						return configurationResponse(403, `{"message":"Forbidden"}`), nil
					default:
						return configurationResponse(404, `{"message":"Not Found"}`), nil
					}
				}}
				config, err := client.GetRepositoryConfiguration(ctx, "org", "repo")
				require.True(t, reached)
				require.Error(t, err)
				require.Nil(t, config, "read failure must not become a complete-looking configuration")
				switch kind {
				case "transport":
					require.ErrorIs(t, err, sentinel)
				case "canceled":
					require.ErrorIs(t, err, context.Canceled)
				case "decode":
					var syntax *json.SyntaxError
					require.ErrorAs(t, err, &syntax)
				default:
					var api *APIError
					require.ErrorAs(t, err, &api)
					if kind == "403" {
						require.Equal(t, 403, api.StatusCode)
					} else {
						require.Equal(t, 404, api.StatusCode)
					}
				}
			})
		}
	}
}

// default branch가 없는 저장소와 빈 권한 목록은 여전히 정상 데이터다.
func TestRepositoryConfigurationEmptyOptionalData(t *testing.T) {
	client := NewRepoConfigClient("fixture-not-secret")
	calls := 0
	client.httpClient = configurationHTTP{do: func(_ context.Context, request *http.Request) (*http.Response, error) {
		calls++
		if strings.HasSuffix(request.URL.Path, "/protection") {
			t.Fatal("no default branch: protection lookup must remain skipped")
		}
		body := `[]`
		if request.URL.Path == "/repos/org/repo" {
			body = `{"name":"repo","default_branch":""}`
		}
		return configurationResponse(200, body), nil
	}}
	config, err := client.GetRepositoryConfiguration(context.Background(), "org", "repo")
	require.NoError(t, err)
	require.Equal(t, "repo", config.Name)
	require.Empty(t, config.BranchProtection)
	require.Empty(t, config.Permissions.Teams)
	require.Empty(t, config.Permissions.Users)
	require.Equal(t, 3, calls)
}

// 연속 요청의 응답 변수가 바뀌어도 각 body를 한 번씩 닫아야 한다.
type trackedConfigurationBody struct {
	io.Reader
	closes int
}

func (b *trackedConfigurationBody) Close() error { b.closes++; return nil }

func TestRepositoryConfigurationClosesEachResponseBody(t *testing.T) {
	var bodies []*trackedConfigurationBody
	client := NewRepoConfigClient("fixture-not-secret")
	client.httpClient = configurationHTTP{do: func(_ context.Context, request *http.Request) (*http.Response, error) {
		text := `[]`
		if request.URL.Path == "/repos/org/repo" {
			text = `{"name":"repo","default_branch":"main"}`
		}
		if strings.HasSuffix(request.URL.Path, "/protection") {
			text = `{}`
		}
		body := &trackedConfigurationBody{Reader: strings.NewReader(text)}
		bodies = append(bodies, body)
		response := configurationResponse(200, text)
		response.Body = body
		return response, nil
	}}
	_, err := client.GetRepositoryConfiguration(context.Background(), "org", "repo")
	require.NoError(t, err)
	require.Len(t, bodies, 4)
	for _, body := range bodies {
		require.Equal(t, 1, body.closes)
	}
}
