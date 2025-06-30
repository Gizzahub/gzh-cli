package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

	client.updateRateLimit(resp)

	assert.Equal(t, 4999, client.rateLimiter.remaining)
	assert.Equal(t, 5000, client.rateLimiter.limit)
	assert.Equal(t, time.Unix(1640995200, 0), client.rateLimiter.resetTime)
}

func TestCheckRateLimit(t *testing.T) {
	client := NewRepoConfigClient("test-token")

	// Test no rate limiting
	err := client.checkRateLimit()
	assert.NoError(t, err)

	// Test rate limit exceeded with future reset time
	client.rateLimiter.remaining = 0
	client.rateLimiter.resetTime = time.Now().Add(10 * time.Minute)

	err = client.checkRateLimit()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit exceeded")
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

func TestGetRateLimit(t *testing.T) {
	client := NewRepoConfigClient("test-token")

	// Set some rate limit values
	client.rateLimiter.remaining = 4999
	client.rateLimiter.limit = 5000
	client.rateLimiter.resetTime = time.Unix(1640995200, 0)

	rateLimit := client.GetRateLimit()

	assert.Equal(t, 4999, rateLimit.remaining)
	assert.Equal(t, 5000, rateLimit.limit)
	assert.Equal(t, time.Unix(1640995200, 0), rateLimit.resetTime)
}
