//nolint:testpackage // White-box testing for provider mutation wiring
package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/pkg/git/provider"
)

// mutationAPIClient is a controllable APIClient for provider mutation tests.
type mutationAPIClient struct {
	createFn    func(ctx context.Context, owner string, opts *CreateRepositoryOptions) (*RepositoryInfo, error)
	deleteFn    func(ctx context.Context, owner, repo string) error
	archiveFn   func(ctx context.Context, owner, repo string) error
	unarchiveFn func(ctx context.Context, owner, repo string) error
	searchFn    func(ctx context.Context, query string, opts *SearchRepositoriesOptions) (*RepositorySearchResult, error)
}

func (m *mutationAPIClient) GetRepository(ctx context.Context, owner, repo string) (*RepositoryInfo, error) {
	return nil, nil
}
func (m *mutationAPIClient) ListOrganizationRepositories(ctx context.Context, org string) ([]RepositoryInfo, error) {
	return nil, nil
}
func (m *mutationAPIClient) GetDefaultBranch(ctx context.Context, owner, repo string) (string, error) {
	return "main", nil
}
func (m *mutationAPIClient) SetToken(ctx context.Context, token string) error { return nil }
func (m *mutationAPIClient) GetRateLimit(ctx context.Context) (*RateLimit, error) {
	return &RateLimit{}, nil
}
func (m *mutationAPIClient) GetRepositoryConfiguration(ctx context.Context, owner, repo string) (*RepositoryConfig, error) {
	return nil, nil
}
func (m *mutationAPIClient) UpdateRepositoryConfiguration(ctx context.Context, owner, repo string, config *RepositoryConfig) error {
	return nil
}
func (m *mutationAPIClient) CreateRepository(ctx context.Context, owner string, opts *CreateRepositoryOptions) (*RepositoryInfo, error) {
	if m.createFn != nil {
		return m.createFn(ctx, owner, opts)
	}
	return nil, nil
}
func (m *mutationAPIClient) DeleteRepository(ctx context.Context, owner, repo string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, owner, repo)
	}
	return nil
}
func (m *mutationAPIClient) ArchiveRepository(ctx context.Context, owner, repo string) error {
	if m.archiveFn != nil {
		return m.archiveFn(ctx, owner, repo)
	}
	return nil
}
func (m *mutationAPIClient) UnarchiveRepository(ctx context.Context, owner, repo string) error {
	if m.unarchiveFn != nil {
		return m.unarchiveFn(ctx, owner, repo)
	}
	return nil
}
func (m *mutationAPIClient) SearchRepositories(ctx context.Context, query string, opts *SearchRepositoriesOptions) (*RepositorySearchResult, error) {
	if m.searchFn != nil {
		return m.searchFn(ctx, query, opts)
	}
	return &RepositorySearchResult{}, nil
}

type noopCloneService struct{}

func (n *noopCloneService) CloneRepository(ctx context.Context, repo RepositoryInfo, targetPath, strategy string) error {
	return nil
}
func (n *noopCloneService) RefreshAll(ctx context.Context, targetPath, orgName, strategy string) error {
	return nil
}
func (n *noopCloneService) CloneOrganization(ctx context.Context, orgName, targetPath, strategy string) error {
	return nil
}
func (n *noopCloneService) SetStrategy(ctx context.Context, strategy string) error { return nil }
func (n *noopCloneService) GetSupportedStrategies(ctx context.Context) ([]string, error) {
	return []string{"reset"}, nil
}

func TestGitHubProvider_CreateRepository(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		req       provider.CreateRepoRequest
		client    *mutationAPIClient
		wantErr   string
		wantName  string
		wantOwner string
	}{
		{
			name: "owner set creates under owner",
			req: provider.CreateRepoRequest{
				Owner:  "acme",
				Name:   "widget",
				Private: true,
			},
			client: &mutationAPIClient{
				createFn: func(_ context.Context, owner string, opts *CreateRepositoryOptions) (*RepositoryInfo, error) {
					assert.Equal(t, "acme", owner)
					assert.Equal(t, "widget", opts.Name)
					assert.True(t, opts.Private)
					return &RepositoryInfo{
						Name:     "widget",
						FullName: "acme/widget",
						Private:  true,
					}, nil
				},
			},
			wantName:  "widget",
			wantOwner: "acme/widget",
		},
		{
			name: "empty owner fails fast",
			req: provider.CreateRepoRequest{
				Name: "widget",
			},
			client:  &mutationAPIClient{},
			wantErr: "owner is required",
		},
		{
			name: "empty name fails fast",
			req: provider.CreateRepoRequest{
				Owner: "acme",
			},
			client:  &mutationAPIClient{},
			wantErr: "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := NewGitHubProvider(tt.client, &noopCloneService{})
			repo, err := p.CreateRepository(context.Background(), tt.req)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, repo)
			assert.Equal(t, tt.wantName, repo.Name)
			assert.Equal(t, tt.wantOwner, repo.FullName)
		})
	}
}

func TestGitHubProvider_DeleteArchiveSearch(t *testing.T) {
	t.Parallel()

	var deleted, archived, unarchived bool
	client := &mutationAPIClient{
		deleteFn: func(_ context.Context, owner, repo string) error {
			assert.Equal(t, "acme", owner)
			assert.Equal(t, "widget", repo)
			deleted = true
			return nil
		},
		archiveFn: func(_ context.Context, owner, repo string) error {
			assert.Equal(t, "acme", owner)
			assert.Equal(t, "widget", repo)
			archived = true
			return nil
		},
		unarchiveFn: func(_ context.Context, owner, repo string) error {
			assert.Equal(t, "acme", owner)
			assert.Equal(t, "widget", repo)
			unarchived = true
			return nil
		},
		searchFn: func(_ context.Context, query string, _ *SearchRepositoriesOptions) (*RepositorySearchResult, error) {
			assert.Contains(t, query, "widget")
			return &RepositorySearchResult{
				TotalCount: 1,
				Repositories: []RepositoryInfo{
					{Name: "widget", FullName: "acme/widget"},
				},
			}, nil
		},
	}

	p := NewGitHubProvider(client, &noopCloneService{})
	ctx := context.Background()

	require.NoError(t, p.DeleteRepository(ctx, "acme/widget"))
	require.NoError(t, p.ArchiveRepository(ctx, "acme/widget"))
	require.NoError(t, p.UnarchiveRepository(ctx, "acme/widget"))

	result, err := p.SearchRepositories(ctx, provider.SearchQuery{Query: "widget", Organization: "acme"})
	require.NoError(t, err)
	require.Equal(t, 1, result.TotalCount)
	require.Len(t, result.Repositories, 1)
	assert.Equal(t, "acme/widget", result.Repositories[0].FullName)

	assert.True(t, deleted)
	assert.True(t, archived)
	assert.True(t, unarchived)
}

func TestResilientGitHubClient_Mutations_HTTP(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/acme/repos", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "token test-token", r.Header.Get("Authorization"))
		var body CreateRepositoryOptions
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "widget", body.Name)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(RepositoryInfo{
			Name:     "widget",
			FullName: "acme/widget",
			Private:  body.Private,
		})
	})
	mux.HandleFunc("/repos/acme/widget", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodPatch:
			var body map[string]bool
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Contains(t, body, "archived")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"full_name": "acme/widget", "archived": body["archived"]})
		default:
			http.NotFound(w, r)
		}
	})
	mux.HandleFunc("/search/repositories", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "widget", r.URL.Query().Get("q"))
		_ = json.NewEncoder(w).Encode(RepositorySearchResult{
			TotalCount: 1,
			Repositories: []RepositoryInfo{
				{Name: "widget", FullName: "acme/widget"},
			},
		})
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewResilientGitHubClientWithBaseURL("test-token", server.URL)
	ctx := context.Background()

	created, err := client.CreateRepository(ctx, "acme", &CreateRepositoryOptions{Name: "widget", Private: true})
	require.NoError(t, err)
	assert.Equal(t, "acme/widget", created.FullName)

	require.NoError(t, client.DeleteRepository(ctx, "acme", "widget"))
	require.NoError(t, client.ArchiveRepository(ctx, "acme", "widget"))
	require.NoError(t, client.UnarchiveRepository(ctx, "acme", "widget"))

	search, err := client.SearchRepositories(ctx, "widget", nil)
	require.NoError(t, err)
	require.Equal(t, 1, search.TotalCount)
}

func TestGitHubAPIClient_CreateRequiresOwner(t *testing.T) {
	t.Parallel()
	c := &GitHubAPIClient{
		config: DefaultAPIClientConfig(),
		logger: &testLogger{},
	}
	_, err := c.CreateRepository(context.Background(), "", &CreateRepositoryOptions{Name: "x"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "owner is required")
}

type testLogger struct{}

func (l *testLogger) Debug(msg string, args ...any) {}
func (l *testLogger) Info(msg string, args ...any)  {}
func (l *testLogger) Warn(msg string, args ...any)  {}
func (l *testLogger) Error(msg string, args ...any) {}

func TestBuildGitHubSearchQuery(t *testing.T) {
	t.Parallel()
	fork := false
	q := buildGitHubSearchQuery(provider.SearchQuery{
		Query:        "cli",
		Organization: "gizzahub",
		Language:     "go",
		Fork:         &fork,
	})
	assert.True(t, strings.Contains(q, "cli"))
	assert.True(t, strings.Contains(q, "org:gizzahub"))
	assert.True(t, strings.Contains(q, "language:go"))
	assert.True(t, strings.Contains(q, "fork:false"))
}
