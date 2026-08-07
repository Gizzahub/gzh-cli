//nolint:testpackage // White-box testing needed for internal field access
package github

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveGitHubAPIBaseURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: DefaultGitHubAPIBaseURL},
		{name: "whitespace", input: "   ", want: DefaultGitHubAPIBaseURL},
		{name: "custom", input: "https://ghes.example.com/api/v3", want: "https://ghes.example.com/api/v3"},
		{name: "trailing slash", input: "https://ghes.example.com/api/v3/", want: "https://ghes.example.com/api/v3"},
		{name: "trailing slashes", input: "https://ghes.example.com/api/v3///", want: "https://ghes.example.com/api/v3"},
		{name: "spaces around", input: "  https://ghes.example.com/api/v3/  ", want: "https://ghes.example.com/api/v3"},
		{name: "default constant", input: DefaultGitHubAPIBaseURL, want: DefaultGitHubAPIBaseURL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ResolveGitHubAPIBaseURL(tt.input))
		})
	}
}

func TestDefaultAPIClientConfig_BaseURL(t *testing.T) {
	cfg := DefaultAPIClientConfig()
	assert.Equal(t, DefaultGitHubAPIBaseURL, cfg.BaseURL)
}

func TestNewAPIClient_EmptyBaseURL(t *testing.T) {
	client := NewAPIClient(&APIClientConfig{Token: "t"}, nil, &mockWebhookLogger{})
	api, ok := client.(*GitHubAPIClient)
	require.True(t, ok)
	assert.Equal(t, DefaultGitHubAPIBaseURL, api.config.BaseURL)
}

func TestNewAPIClient_CustomBaseURL(t *testing.T) {
	client := NewAPIClient(&APIClientConfig{
		BaseURL: "https://ghes.example.com/api/v3/",
		Token:   "t",
	}, nil, &mockWebhookLogger{})
	api, ok := client.(*GitHubAPIClient)
	require.True(t, ok)
	assert.Equal(t, "https://ghes.example.com/api/v3", api.config.BaseURL)
}

func TestNewWebhookService_DefaultAndCustomBaseURL(t *testing.T) {
	logger := &mockWebhookLogger{}

	defaultSvc := NewWebhookService(nil, logger)
	impl, ok := defaultSvc.(*webhookServiceImpl)
	require.True(t, ok)
	assert.Equal(t, DefaultGitHubAPIBaseURL, impl.baseURL)

	custom := NewWebhookServiceWithBaseURL(nil, logger, "https://ghes.example.com/api/v3/")
	cimpl, ok := custom.(*webhookServiceImpl)
	require.True(t, ok)
	assert.Equal(t, "https://ghes.example.com/api/v3", cimpl.baseURL)

	tokenSvc := NewWebhookServiceWithToken(nil, "tok", logger)
	timpl, ok := tokenSvc.(*webhookServiceImpl)
	require.True(t, ok)
	assert.Equal(t, DefaultGitHubAPIBaseURL, timpl.baseURL)
	assert.Equal(t, "tok", timpl.token)

	tokenCustom := NewWebhookServiceWithTokenAndBaseURL(nil, "tok2", logger, "https://enterprise.local/api/v3")
	tcimpl, ok := tokenCustom.(*webhookServiceImpl)
	require.True(t, ok)
	assert.Equal(t, "https://enterprise.local/api/v3", tcimpl.baseURL)
	assert.Equal(t, "tok2", tcimpl.token)
}

func TestNewResilientGitHubClient_DefaultAndCustomBaseURL(t *testing.T) {
	client := NewResilientGitHubClient("token")
	assert.Equal(t, DefaultGitHubAPIBaseURL, client.baseURL)

	custom := NewResilientGitHubClientWithBaseURL("token", "https://ghes.example.com/api/v3/")
	assert.Equal(t, "https://ghes.example.com/api/v3", custom.baseURL)

	withConfig := NewResilientGitHubClientWithConfig("token", 10*time.Second)
	assert.Equal(t, DefaultGitHubAPIBaseURL, withConfig.baseURL)
	assert.Equal(t, 10*time.Second, withConfig.httpClient.Timeout)

	withConfigCustom := NewResilientGitHubClientWithConfigAndBaseURL("token", 5*time.Second, "https://enterprise.local/api/v3/")
	assert.Equal(t, "https://enterprise.local/api/v3", withConfigCustom.baseURL)
	assert.Equal(t, 5*time.Second, withConfigCustom.httpClient.Timeout)
}

func TestNewGitHubProvider_DefaultAndCustomBaseURL(t *testing.T) {
	p := NewGitHubProvider(nil, nil)
	assert.Equal(t, DefaultGitHubAPIBaseURL, p.GetBaseURL())

	custom := NewGitHubProviderWithBaseURL(nil, nil, "https://ghes.example.com/api/v3/")
	assert.Equal(t, "https://ghes.example.com/api/v3", custom.GetBaseURL())
}

func TestNewTokenAwareGitHubClient_ResolvesBaseURL(t *testing.T) {
	emptyCfg := TokenAwareGitHubClientConfig{PrimaryToken: "t"}
	client, err := NewTokenAwareGitHubClient(emptyCfg)
	require.NoError(t, err)
	assert.Equal(t, DefaultGitHubAPIBaseURL, client.baseURL)

	customCfg := TokenAwareGitHubClientConfig{
		BaseURL:      "https://ghes.example.com/api/v3/",
		PrimaryToken: "t",
	}
	custom, err := NewTokenAwareGitHubClient(customCfg)
	require.NoError(t, err)
	assert.Equal(t, "https://ghes.example.com/api/v3", custom.baseURL)
}
