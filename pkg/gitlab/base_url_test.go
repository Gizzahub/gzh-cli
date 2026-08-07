//nolint:testpackage // White-box tests for pure URL helpers
package gitlab

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeAPIBase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "root host", in: "https://gitlab.company.com", want: "https://gitlab.company.com/api/v4"},
		{name: "trailing slash", in: "https://gitlab.company.com/", want: "https://gitlab.company.com/api/v4"},
		{name: "already api v4", in: "https://gitlab.company.com/api/v4", want: "https://gitlab.company.com/api/v4"},
		{name: "already api v4 slash", in: "https://gitlab.company.com/api/v4/", want: "https://gitlab.company.com/api/v4"},
		{name: "custom api path", in: "https://gitlab.company.com/api/v3", want: "https://gitlab.company.com/api/v3"},
		{name: "empty", in: "", want: "/api/v4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, normalizeAPIBase(tt.in))
		})
	}
}

func TestTrimTrailingSlash(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "https://example.com", trimTrailingSlash("https://example.com/"))
	assert.Equal(t, "https://example.com", trimTrailingSlash("https://example.com"))
	assert.Equal(t, "", trimTrailingSlash(""))
	assert.Equal(t, "", trimTrailingSlash("/"))
}

func TestSetBaseURLAndBuildAPIURL(t *testing.T) {
	// Mutates package-level configuredBaseAPIURL; keep serial.
	prev := configuredBaseAPIURL
	t.Cleanup(func() { configuredBaseAPIURL = prev })

	SetBaseURL("https://gitlab.example.com")
	assert.Equal(t, "https://gitlab.example.com/api/v4", getBaseAPIURL())
	assert.Equal(t, "https://gitlab.example.com/api/v4/groups/mygroup", buildAPIURL("groups/mygroup"))
	assert.Equal(t, "https://gitlab.example.com/api/v4/groups/mygroup", buildAPIURL("/groups/mygroup"))
	assert.Equal(t, "https://gitlab.example.com", getWebBaseURL())

	SetBaseURL("https://gitlab.example.com/api/v4/")
	assert.Equal(t, "https://gitlab.example.com/api/v4", getBaseAPIURL())
	assert.Equal(t, "https://gitlab.example.com", getWebBaseURL())
}

func TestGetBaseAPIURL_DefaultAndEnv(t *testing.T) {
	prevConfigured := configuredBaseAPIURL
	configuredBaseAPIURL = ""
	t.Cleanup(func() { configuredBaseAPIURL = prevConfigured })

	t.Setenv("GITLAB_BASE_URL", "")
	t.Setenv("GITLAB_API_URL", "")
	t.Setenv("GZH_GITLAB_API", "")
	assert.Equal(t, "https://gitlab.com/api/v4", getBaseAPIURL())

	t.Setenv("GITLAB_BASE_URL", "https://self-hosted.example")
	assert.Equal(t, "https://self-hosted.example/api/v4", getBaseAPIURL())

	t.Setenv("GITLAB_BASE_URL", "")
	t.Setenv("GITLAB_API_URL", "https://api.example/api/v4/")
	assert.Equal(t, "https://api.example/api/v4", getBaseAPIURL())

	t.Setenv("GITLAB_API_URL", "")
	t.Setenv("GZH_GITLAB_API", "https://gzh.example/api/v4")
	assert.Equal(t, "https://gzh.example/api/v4", getBaseAPIURL())

	// Injected SetBaseURL wins over env.
	SetBaseURL("https://injected.example")
	assert.Equal(t, "https://injected.example/api/v4", getBaseAPIURL())
}

func TestDifference(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b []string
		want []string
	}{
		{name: "basic", a: []string{"a", "b", "c"}, b: []string{"b"}, want: []string{"a", "c"}},
		{name: "none", a: []string{"a"}, b: []string{"a"}, want: nil},
		{name: "empty a", a: nil, b: []string{"x"}, want: nil},
		{name: "empty b", a: []string{"x", "y"}, b: nil, want: []string{"x", "y"}},
		{name: "preserves order", a: []string{"z", "a", "m"}, b: []string{"a"}, want: []string{"z", "m"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, difference(tt.a, tt.b))
		})
	}
}

func TestHasNextPage(t *testing.T) {
	t.Parallel()

	client := &ResilientGitLabClient{}

	tests := []struct {
		name string
		hdr  http.Header
		want bool
	}{
		{
			name: "has next",
			hdr:  http.Header{"X-Total-Pages": []string{"5"}, "X-Page": []string{"2"}},
			want: true,
		},
		{
			name: "last page",
			hdr:  http.Header{"X-Total-Pages": []string{"5"}, "X-Page": []string{"5"}},
			want: false,
		},
		{
			name: "missing headers",
			hdr:  http.Header{},
			want: false,
		},
		{
			name: "invalid numbers",
			hdr:  http.Header{"X-Total-Pages": []string{"x"}, "X-Page": []string{"1"}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, client.hasNextPage(tt.hdr))
		})
	}
}

func TestAddAuthHeader(t *testing.T) {
	prev := configuredToken
	t.Cleanup(func() { configuredToken = prev })

	req, err := http.NewRequest(http.MethodGet, "https://gitlab.com/api/v4/user", nil)
	require.NoError(t, err)

	configuredToken = ""
	addAuthHeader(req)
	assert.Empty(t, req.Header.Get("PRIVATE-TOKEN"))

	configuredToken = "glpat-test-token"
	addAuthHeader(req)
	assert.Equal(t, "glpat-test-token", req.Header.Get("PRIVATE-TOKEN"))
}

func TestTokenGuidanceMessage(t *testing.T) {
	prev := configuredBaseAPIURL
	t.Cleanup(func() { configuredBaseAPIURL = prev })

	SetBaseURL("https://gitlab.example.com")
	msg := tokenGuidanceMessage()
	assert.Contains(t, msg, "read_api")
	assert.Contains(t, msg, "read_repository")
	assert.Contains(t, msg, "https://gitlab.example.com/-/profile/personal_access_tokens")
}

func TestFormatGuidanceBox(t *testing.T) {
	t.Parallel()

	box := formatGuidanceBox("Title", "body content")
	assert.Contains(t, box, "Title")
	assert.Contains(t, box, "body content")
	assert.Contains(t, box, "┌─")
}
