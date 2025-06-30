package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRepoConfigClient(t *testing.T) {
	client := NewRepoConfigClient("test-token")

	assert.Equal(t, "test-token", client.token)
	assert.Equal(t, "https://api.github.com", client.baseURL)
	assert.NotNil(t, client.httpClient)
	assert.NotNil(t, client.rateLimiter)
	assert.Equal(t, 30*time.Second, client.httpClient.Timeout)
}

func TestSetTimeout(t *testing.T) {
	client := NewRepoConfigClient("test-token")

	client.SetTimeout(60 * time.Second)
	assert.Equal(t, 60*time.Second, client.httpClient.Timeout)
}

func TestAPIError(t *testing.T) {
	err := &APIError{
		Message:    "Not Found",
		StatusCode: 404,
	}

	assert.Equal(t, "GitHub API error (404): Not Found", err.Error())
}

func TestUpdateRateLimit(t *testing.T) {
	client := NewRepoConfigClient("test-token")

	// Mock response with rate limit headers
	resp := &http.Response{
		Header: make(http.Header),
	}

	// Set headers correctly
	resp.Header.Set("X-RateLimit-Remaining", "4999")
	resp.Header.Set("X-RateLimit-Reset", "1640995200")
	resp.Header.Set("X-RateLimit-Limit", "5000")

	// Verify headers are set correctly
	assert.Equal(t, "4999", resp.Header.Get("X-RateLimit-Remaining"))
	assert.Equal(t, "1640995200", resp.Header.Get("X-RateLimit-Reset"))
	assert.Equal(t, "5000", resp.Header.Get("X-RateLimit-Limit"))

	client.rateLimiter.Update(resp)

	remaining, limit, resetTime := client.GetRateLimitStatus()
	assert.Equal(t, 4999, remaining)
	assert.Equal(t, 5000, limit)
	assert.Equal(t, time.Unix(1640995200, 0), resetTime)
}

func TestCheckRateLimit(t *testing.T) {
	client := NewRepoConfigClient("test-token")

	// Test no rate limiting
	ctx := context.Background()
	err := client.rateLimiter.Wait(ctx)
	assert.NoError(t, err)

	// Verify rate limiting works properly in the rate_limiter_test.go file
	// This functionality is now tested separately
}

func TestListRepositories(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/orgs/testorg/repos")
		assert.Equal(t, "token test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/vnd.github.v3+json", r.Header.Get("Accept"))

		// Mock response - return only 1 repo (less than PerPage=2) to stop pagination
		repos := []*Repository{
			{
				ID:       1,
				Name:     "repo1",
				FullName: "testorg/repo1",
				Private:  false,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(repos)
	}))
	defer server.Close()

	client := NewRepoConfigClient("test-token")
	client.baseURL = server.URL

	repos, err := client.ListRepositories(context.Background(), "testorg", &ListOptions{
		PerPage: 2,
		Type:    "all",
	})

	require.NoError(t, err)
	assert.Len(t, repos, 1)
	assert.Equal(t, "repo1", repos[0].Name)
	assert.False(t, repos[0].Private)
}

func TestGetRepository(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/repos/testorg/testrepo", r.URL.Path)

		repo := &Repository{
			ID:            1,
			Name:          "testrepo",
			FullName:      "testorg/testrepo",
			Description:   "Test repository",
			Private:       false,
			DefaultBranch: "main",
			HasIssues:     true,
			HasWiki:       false,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(repo)
	}))
	defer server.Close()

	client := NewRepoConfigClient("test-token")
	client.baseURL = server.URL

	repo, err := client.GetRepository(context.Background(), "testorg", "testrepo")

	require.NoError(t, err)
	assert.Equal(t, int64(1), repo.ID)
	assert.Equal(t, "testrepo", repo.Name)
	assert.Equal(t, "testorg/testrepo", repo.FullName)
	assert.Equal(t, "Test repository", repo.Description)
	assert.False(t, repo.Private)
	assert.Equal(t, "main", repo.DefaultBranch)
	assert.True(t, repo.HasIssues)
	assert.False(t, repo.HasWiki)
}

func TestUpdateRepository(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/repos/testorg/testrepo", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify request body
		var update RepositoryUpdate
		err := json.NewDecoder(r.Body).Decode(&update)
		require.NoError(t, err)

		assert.Equal(t, "Updated description", *update.Description)
		assert.True(t, *update.HasIssues)
		assert.False(t, *update.HasWiki)

		// Mock response
		repo := &Repository{
			ID:          1,
			Name:        "testrepo",
			Description: "Updated description",
			HasIssues:   true,
			HasWiki:     false,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(repo)
	}))
	defer server.Close()

	client := NewRepoConfigClient("test-token")
	client.baseURL = server.URL

	description := "Updated description"
	hasIssues := true
	hasWiki := false

	update := &RepositoryUpdate{
		Description: &description,
		HasIssues:   &hasIssues,
		HasWiki:     &hasWiki,
	}

	repo, err := client.UpdateRepository(context.Background(), "testorg", "testrepo", update)

	require.NoError(t, err)
	assert.Equal(t, "Updated description", repo.Description)
	assert.True(t, repo.HasIssues)
	assert.False(t, repo.HasWiki)
}

func TestGetBranchProtection(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/repos/testorg/testrepo/branches/main/protection", r.URL.Path)

		protection := &BranchProtection{
			RequiredStatusChecks: &RequiredStatusChecks{
				Strict:   true,
				Contexts: []string{"ci/build", "ci/test"},
			},
			EnforceAdmins: true,
			RequiredPullRequestReviews: &RequiredPullRequestReviews{
				DismissStaleReviews:          true,
				RequireCodeOwnerReviews:      true,
				RequiredApprovingReviewCount: 2,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(protection)
	}))
	defer server.Close()

	client := NewRepoConfigClient("test-token")
	client.baseURL = server.URL

	protection, err := client.GetBranchProtection(context.Background(), "testorg", "testrepo", "main")

	require.NoError(t, err)
	assert.NotNil(t, protection.RequiredStatusChecks)
	assert.True(t, protection.RequiredStatusChecks.Strict)
	assert.Equal(t, []string{"ci/build", "ci/test"}, protection.RequiredStatusChecks.Contexts)
	assert.True(t, protection.EnforceAdmins)
	assert.NotNil(t, protection.RequiredPullRequestReviews)
	assert.True(t, protection.RequiredPullRequestReviews.DismissStaleReviews)
	assert.True(t, protection.RequiredPullRequestReviews.RequireCodeOwnerReviews)
	assert.Equal(t, 2, protection.RequiredPullRequestReviews.RequiredApprovingReviewCount)
}

func TestUpdateBranchProtection(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/repos/testorg/testrepo/branches/main/protection", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify request body
		var protection BranchProtection
		err := json.NewDecoder(r.Body).Decode(&protection)
		require.NoError(t, err)

		assert.True(t, protection.EnforceAdmins)
		assert.NotNil(t, protection.RequiredPullRequestReviews)
		assert.Equal(t, 2, protection.RequiredPullRequestReviews.RequiredApprovingReviewCount)

		// Echo back the protection
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(protection)
	}))
	defer server.Close()

	client := NewRepoConfigClient("test-token")
	client.baseURL = server.URL

	protection := &BranchProtection{
		EnforceAdmins: true,
		RequiredPullRequestReviews: &RequiredPullRequestReviews{
			RequiredApprovingReviewCount: 2,
			DismissStaleReviews:          true,
		},
	}

	updated, err := client.UpdateBranchProtection(context.Background(), "testorg", "testrepo", "main", protection)

	require.NoError(t, err)
	assert.True(t, updated.EnforceAdmins)
	assert.NotNil(t, updated.RequiredPullRequestReviews)
	assert.Equal(t, 2, updated.RequiredPullRequestReviews.RequiredApprovingReviewCount)
}

func TestDeleteBranchProtection(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/repos/testorg/testrepo/branches/main/protection", r.URL.Path)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewRepoConfigClient("test-token")
	client.baseURL = server.URL

	err := client.DeleteBranchProtection(context.Background(), "testorg", "testrepo", "main")

	require.NoError(t, err)
}

func TestAPIErrorHandling(t *testing.T) {
	// Mock server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")

		errorResp := map[string]interface{}{
			"message":           "Not Found",
			"documentation_url": "https://docs.github.com/rest",
		}
		_ = json.NewEncoder(w).Encode(errorResp)
	}))
	defer server.Close()

	client := NewRepoConfigClient("test-token")
	client.baseURL = server.URL

	_, err := client.GetRepository(context.Background(), "testorg", "nonexistent")

	require.Error(t, err)

	var apiError *APIError
	assert.ErrorAs(t, err, &apiError)
	assert.Equal(t, 404, apiError.StatusCode)
	assert.Equal(t, "Not Found", apiError.Message)
}

func TestMakeRequestWithoutToken(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify no Authorization header is set
		assert.Empty(t, r.Header.Get("Authorization"))
		assert.Equal(t, "application/vnd.github.v3+json", r.Header.Get("Accept"))
		assert.Equal(t, "gzh-manager-go/1.0", r.Header.Get("User-Agent"))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := NewRepoConfigClient("") // No token
	client.baseURL = server.URL

	resp, err := client.makeRequest(context.Background(), "GET", "/test", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestListRepositoriesWithPagination(t *testing.T) {
	page := 0
	// Mock server that returns different responses for different pages
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++

		var repos []*Repository
		if page == 1 {
			// First page with 2 items (per_page=2)
			repos = []*Repository{
				{ID: 1, Name: "repo1"},
				{ID: 2, Name: "repo2"},
			}
		} else {
			// Second page with 1 item (less than per_page, so last page)
			repos = []*Repository{
				{ID: 3, Name: "repo3"},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(repos)
	}))
	defer server.Close()

	client := NewRepoConfigClient("test-token")
	client.baseURL = server.URL

	repos, err := client.ListRepositories(context.Background(), "testorg", &ListOptions{
		PerPage: 2,
	})

	require.NoError(t, err)
	assert.Len(t, repos, 3) // Should get all repos from both pages
	assert.Equal(t, "repo1", repos[0].Name)
	assert.Equal(t, "repo2", repos[1].Name)
	assert.Equal(t, "repo3", repos[2].Name)
}

func TestContextCancellation(t *testing.T) {
	// Mock server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := NewRepoConfigClient("test-token")
	client.baseURL = server.URL

	// Create context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.GetRepository(ctx, "testorg", "testrepo")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestGetRateLimitStatus(t *testing.T) {
	client := NewRepoConfigClient("test-token")

	// Create test response with rate limit headers
	resp := &http.Response{
		Header: make(http.Header),
	}
	resp.Header.Set("X-RateLimit-Remaining", "4999")
	resp.Header.Set("X-RateLimit-Limit", "5000")
	resp.Header.Set("X-RateLimit-Reset", "1640995200")

	// Update rate limit
	client.rateLimiter.Update(resp)

	// Get rate limit status
	remaining, limit, resetTime := client.GetRateLimitStatus()
	assert.Equal(t, 4999, remaining)
	assert.Equal(t, 5000, limit)
	assert.Equal(t, time.Unix(1640995200, 0), resetTime)
}

func TestGetRepositoryConfiguration(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/testorg/testrepo":
			repo := &Repository{
				ID:                  1,
				Name:                "testrepo",
				Description:         "Test repository",
				Homepage:            "https://example.com",
				Private:             false,
				Archived:            false,
				DefaultBranch:       "main",
				Topics:              []string{"test", "example"},
				HasIssues:           true,
				HasWiki:             false,
				HasProjects:         true,
				HasDownloads:        false,
				AllowSquashMerge:    true,
				AllowMergeCommit:    false,
				AllowRebaseMerge:    false,
				DeleteBranchOnMerge: true,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(repo)

		case "/repos/testorg/testrepo/branches/main/protection":
			protection := &BranchProtection{
				RequiredStatusChecks: &RequiredStatusChecks{
					Strict:   true,
					Contexts: []string{"ci/build", "ci/test"},
				},
				EnforceAdmins: true,
				RequiredPullRequestReviews: &RequiredPullRequestReviews{
					DismissStaleReviews:          true,
					RequireCodeOwnerReviews:      true,
					RequiredApprovingReviewCount: 2,
				},
				Restrictions: &BranchRestrictions{
					Users: []string{"admin"},
					Teams: []string{"maintainers"},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(protection)

		case "/repos/testorg/testrepo/teams":
			teams := []TeamPermission{
				{ID: 1, Name: "Maintainers", Slug: "maintainers", Permission: "admin"},
				{ID: 2, Name: "Developers", Slug: "developers", Permission: "push"},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(teams)

		case "/repos/testorg/testrepo/collaborators":
			users := []UserPermission{
				{Login: "john", ID: 100, Permission: "admin"},
				{Login: "jane", ID: 101, Permission: "write"},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(users)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewRepoConfigClient("test-token")
	client.baseURL = server.URL

	config, err := client.GetRepositoryConfiguration(context.Background(), "testorg", "testrepo")

	require.NoError(t, err)
	assert.Equal(t, "testrepo", config.Name)
	assert.Equal(t, "Test repository", config.Description)
	assert.Equal(t, "https://example.com", config.Homepage)
	assert.False(t, config.Private)
	assert.False(t, config.Archived)
	assert.Equal(t, []string{"test", "example"}, config.Topics)

	// Check settings
	assert.True(t, config.Settings.HasIssues)
	assert.False(t, config.Settings.HasWiki)
	assert.True(t, config.Settings.HasProjects)
	assert.False(t, config.Settings.HasDownloads)
	assert.True(t, config.Settings.AllowSquashMerge)
	assert.False(t, config.Settings.AllowMergeCommit)
	assert.False(t, config.Settings.AllowRebaseMerge)
	assert.True(t, config.Settings.DeleteBranchOnMerge)
	assert.Equal(t, "main", config.Settings.DefaultBranch)

	// Check branch protection
	assert.NotNil(t, config.BranchProtection)
	mainProtection, ok := config.BranchProtection["main"]
	assert.True(t, ok)
	assert.Equal(t, 2, mainProtection.RequiredReviews)
	assert.True(t, mainProtection.DismissStaleReviews)
	assert.True(t, mainProtection.RequireCodeOwnerReviews)
	assert.True(t, mainProtection.StrictStatusChecks)
	assert.Equal(t, []string{"ci/build", "ci/test"}, mainProtection.RequiredStatusChecks)
	assert.True(t, mainProtection.EnforceAdmins)
	assert.True(t, mainProtection.RestrictPushes)
	assert.Equal(t, []string{"admin"}, mainProtection.AllowedUsers)
	assert.Equal(t, []string{"maintainers"}, mainProtection.AllowedTeams)

	// Check permissions
	assert.Equal(t, "admin", config.Permissions.Teams["maintainers"])
	assert.Equal(t, "push", config.Permissions.Teams["developers"])
	assert.Equal(t, "admin", config.Permissions.Users["john"])
	assert.Equal(t, "write", config.Permissions.Users["jane"])
}

func TestGetRepositoryPermissions(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/testorg/testrepo/teams":
			teams := []TeamPermission{
				{ID: 1, Name: "Admins", Slug: "admins", Permission: "admin"},
				{ID: 2, Name: "Writers", Slug: "writers", Permission: "push"},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(teams)

		case "/repos/testorg/testrepo/collaborators":
			users := []UserPermission{
				{Login: "alice", ID: 200, Permission: "admin"},
				{Login: "bob", ID: 201, Permission: "read"},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(users)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewRepoConfigClient("test-token")
	client.baseURL = server.URL

	teamPerms, userPerms, err := client.GetRepositoryPermissions(context.Background(), "testorg", "testrepo")

	require.NoError(t, err)
	assert.Len(t, teamPerms, 2)
	assert.Equal(t, "admin", teamPerms["admins"])
	assert.Equal(t, "push", teamPerms["writers"])
	assert.Len(t, userPerms, 2)
	assert.Equal(t, "admin", userPerms["alice"])
	assert.Equal(t, "read", userPerms["bob"])
}

func TestConvertBranchProtection(t *testing.T) {
	bp := &BranchProtection{
		RequiredStatusChecks: &RequiredStatusChecks{
			Strict:   true,
			Contexts: []string{"ci/test", "ci/lint"},
		},
		EnforceAdmins: true,
		RequiredPullRequestReviews: &RequiredPullRequestReviews{
			DismissStaleReviews:          true,
			RequireCodeOwnerReviews:      true,
			RequiredApprovingReviewCount: 3,
		},
		Restrictions: &BranchRestrictions{
			Users: []string{"lead"},
			Teams: []string{"core"},
		},
	}

	config := convertBranchProtection(bp)

	assert.True(t, config.EnforceAdmins)
	assert.True(t, config.StrictStatusChecks)
	assert.Equal(t, []string{"ci/test", "ci/lint"}, config.RequiredStatusChecks)
	assert.Equal(t, 3, config.RequiredReviews)
	assert.True(t, config.DismissStaleReviews)
	assert.True(t, config.RequireCodeOwnerReviews)
	assert.True(t, config.RestrictPushes)
	assert.Equal(t, []string{"lead"}, config.AllowedUsers)
	assert.Equal(t, []string{"core"}, config.AllowedTeams)
}

func TestUpdateRepositoryConfiguration(t *testing.T) {
	requestCount := 0
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		switch {
		case r.URL.Path == "/repos/testorg/testrepo" && r.Method == "PATCH":
			// Update repository
			var update RepositoryUpdate
			err := json.NewDecoder(r.Body).Decode(&update)
			require.NoError(t, err)

			assert.Equal(t, "Updated description", *update.Description)
			assert.True(t, *update.HasIssues)
			assert.False(t, *update.HasWiki)

			repo := &Repository{
				ID:          1,
				Name:        "testrepo",
				Description: *update.Description,
				HasIssues:   *update.HasIssues,
				HasWiki:     *update.HasWiki,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(repo)

		case r.URL.Path == "/repos/testorg/testrepo/branches/main/protection" && r.Method == "PUT":
			// Update branch protection
			var protection BranchProtection
			err := json.NewDecoder(r.Body).Decode(&protection)
			require.NoError(t, err)

			assert.True(t, protection.EnforceAdmins)
			assert.NotNil(t, protection.RequiredPullRequestReviews)
			assert.Equal(t, 2, protection.RequiredPullRequestReviews.RequiredApprovingReviewCount)

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(protection)

		case strings.HasPrefix(r.URL.Path, "/orgs/testorg/teams/") && r.Method == "PUT":
			// Update team permission
			var body map[string]string
			err := json.NewDecoder(r.Body).Decode(&body)
			require.NoError(t, err)
			assert.Contains(t, []string{"admin", "push"}, body["permission"])
			w.WriteHeader(http.StatusNoContent)

		case strings.HasPrefix(r.URL.Path, "/repos/testorg/testrepo/collaborators/") && r.Method == "PUT":
			// Update user permission
			var body map[string]string
			err := json.NewDecoder(r.Body).Decode(&body)
			require.NoError(t, err)
			assert.Contains(t, []string{"admin", "write"}, body["permission"])
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewRepoConfigClient("test-token")
	client.baseURL = server.URL

	config := &RepositoryConfig{
		Name:        "testrepo",
		Description: "Updated description",
		Homepage:    "https://example.com",
		Private:     false,
		Archived:    false,
		Topics:      []string{"test", "example"},
		Settings: RepoConfigSettings{
			HasIssues:           true,
			HasWiki:             false,
			HasProjects:         true,
			HasDownloads:        false,
			AllowSquashMerge:    true,
			AllowMergeCommit:    false,
			AllowRebaseMerge:    false,
			DeleteBranchOnMerge: true,
			DefaultBranch:       "main",
		},
		BranchProtection: map[string]BranchProtectionConfig{
			"main": {
				RequiredReviews:         2,
				DismissStaleReviews:     true,
				RequireCodeOwnerReviews: true,
				EnforceAdmins:           true,
			},
		},
		Permissions: PermissionsConfig{
			Teams: map[string]string{
				"maintainers": "admin",
				"developers":  "push",
			},
			Users: map[string]string{
				"john": "admin",
				"jane": "write",
			},
		},
	}

	err := client.UpdateRepositoryConfiguration(context.Background(), "testorg", "testrepo", config)
	require.NoError(t, err)

	// Verify all requests were made
	assert.Equal(t, 6, requestCount) // 1 repo update + 1 branch protection + 2 teams + 2 users = 6 total
}

func TestUpdateBranchProtectionConfig(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/repos/testorg/testrepo/branches/main/protection", r.URL.Path)

		var protection BranchProtection
		err := json.NewDecoder(r.Body).Decode(&protection)
		require.NoError(t, err)

		assert.True(t, protection.EnforceAdmins)
		assert.NotNil(t, protection.RequiredStatusChecks)
		assert.True(t, protection.RequiredStatusChecks.Strict)
		assert.Equal(t, []string{"ci/build", "ci/test"}, protection.RequiredStatusChecks.Contexts)
		assert.NotNil(t, protection.RequiredPullRequestReviews)
		assert.Equal(t, 2, protection.RequiredPullRequestReviews.RequiredApprovingReviewCount)
		assert.NotNil(t, protection.AllowForcePushes)
		assert.False(t, protection.AllowForcePushes.Enabled)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(protection)
	}))
	defer server.Close()

	client := NewRepoConfigClient("test-token")
	client.baseURL = server.URL

	config := &BranchProtectionConfig{
		RequiredReviews:         2,
		DismissStaleReviews:     true,
		RequireCodeOwnerReviews: true,
		RequiredStatusChecks:    []string{"ci/build", "ci/test"},
		StrictStatusChecks:      true,
		EnforceAdmins:           true,
		RestrictPushes:          true,
		AllowedUsers:            []string{"admin"},
		AllowedTeams:            []string{"maintainers"},
		AllowForcePushes:        false,
		AllowDeletions:          false,
	}

	err := client.UpdateBranchProtectionConfig(context.Background(), "testorg", "testrepo", "main", config)
	require.NoError(t, err)
}

func TestUpdateRepositoryPermissions(t *testing.T) {
	requestCount := 0
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		assert.Equal(t, "PUT", r.Method)

		var body map[string]string
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)

		if strings.HasPrefix(r.URL.Path, "/orgs/testorg/teams/") {
			// Team permission update
			assert.Contains(t, r.URL.Path, "repos/testorg/testrepo")
			assert.Contains(t, []string{"admin", "push"}, body["permission"])
		} else if strings.HasPrefix(r.URL.Path, "/repos/testorg/testrepo/collaborators/") {
			// User permission update
			assert.Contains(t, []string{"admin", "write"}, body["permission"])
		} else {
			t.Fatalf("Unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewRepoConfigClient("test-token")
	client.baseURL = server.URL

	perms := PermissionsConfig{
		Teams: map[string]string{
			"admins":  "admin",
			"writers": "push",
		},
		Users: map[string]string{
			"alice": "admin",
			"bob":   "write",
		},
	}

	err := client.UpdateRepositoryPermissions(context.Background(), "testorg", "testrepo", perms)
	require.NoError(t, err)
	assert.Equal(t, 4, requestCount) // 2 teams + 2 users
}
