//nolint:testpackage // White-box testing needed for internal function access
package github

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestResilientGitHubClient_GetDefaultBranch(t *testing.T) {
	// Test server that returns repository info
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/testorg/testrepo" {
			t.Errorf("Expected path /repos/testorg/testrepo, got %s", r.URL.Path)
		}

		repoInfo := RepoInfo{DefaultBranch: "main"}
		json.NewEncoder(w).Encode(repoInfo)
	}))
	defer server.Close()

	client := NewResilientGitHubClient("test-token")
	client.SetBaseURL(server.URL)

	branch, err := client.GetDefaultBranch(context.Background(), "testorg", "testrepo")
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if branch != "main" {
		t.Errorf("Expected branch 'main', got '%s'", branch)
	}
}

func TestResilientGitHubClient_GetDefaultBranch_NotFound(t *testing.T) {
	// Test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewResilientGitHubClient("test-token")
	client.SetBaseURL(server.URL)

	_, err := client.GetDefaultBranch(context.Background(), "testorg", "nonexistent")
	if err == nil {
		t.Fatal("Expected error for 404 response, got success")
	}

	if !containsString(err.Error(), "not found") {
		t.Errorf("Expected 'not found' in error, got: %v", err)
	}
}

func TestResilientGitHubClient_ListRepositories(t *testing.T) {
	// Test server that returns paginated repository list
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")

		var repos []struct {
			Name string `json:"name"`
		}

		switch page {
		case "1", "":
			repos = []struct {
				Name string `json:"name"`
			}{
				{Name: "repo1"},
				{Name: "repo2"},
			}
			// Add link header for pagination
			w.Header().Set("Link", `<http://example.com?page=2>; rel="next"`)
		case "2":
			repos = []struct {
				Name string `json:"name"`
			}{
				{Name: "repo3"},
			}
			// No next link for last page
		}

		json.NewEncoder(w).Encode(repos)
	}))
	defer server.Close()

	client := NewResilientGitHubClient("test-token")
	client.SetBaseURL(server.URL)

	repos, err := client.ListRepositories(context.Background(), "testorg")
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	expectedRepos := []string{"repo1", "repo2", "repo3"}
	if len(repos) != len(expectedRepos) {
		t.Errorf("Expected %d repos, got %d", len(expectedRepos), len(repos))
	}

	for i, expected := range expectedRepos {
		if i >= len(repos) || repos[i] != expected {
			t.Errorf("Expected repo[%d] = '%s', got '%s'", i, expected, repos[i])
		}
	}
}

func TestResilientGitHubClient_GetRateLimit(t *testing.T) {
	// Test server that returns rate limit info
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rate_limit" {
			t.Errorf("Expected path /rate_limit, got %s", r.URL.Path)
		}

		rateLimitResponse := struct {
			Rate struct {
				Limit     int   `json:"limit"`
				Remaining int   `json:"remaining"`
				Reset     int64 `json:"reset"`
			} `json:"rate"`
		}{
			Rate: struct {
				Limit     int   `json:"limit"`
				Remaining int   `json:"remaining"`
				Reset     int64 `json:"reset"`
			}{
				Limit:     5000,
				Remaining: 4999,
				Reset:     time.Now().Add(time.Hour).Unix(),
			},
		}

		json.NewEncoder(w).Encode(rateLimitResponse)
	}))
	defer server.Close()

	client := NewResilientGitHubClient("test-token")
	client.SetBaseURL(server.URL)

	rateLimitInfo, err := client.GetRateLimit(context.Background())
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if rateLimitInfo.Limit != 5000 {
		t.Errorf("Expected limit 5000, got %d", rateLimitInfo.Limit)
	}

	if rateLimitInfo.Remaining != 4999 {
		t.Errorf("Expected remaining 4999, got %d", rateLimitInfo.Remaining)
	}
}

func TestResilientGitHubClient_Authentication(t *testing.T) {
	// Test server that checks for authentication header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "token test-token" {
			t.Errorf("Expected 'token test-token', got '%s'", authHeader)
		}

		userAgent := r.Header.Get("User-Agent")
		if userAgent != "gzh-manager-go" {
			t.Errorf("Expected 'gzh-manager-go' user agent, got '%s'", userAgent)
		}

		accept := r.Header.Get("Accept")
		if accept != "application/vnd.github.v3+json" {
			t.Errorf("Expected GitHub API accept header, got '%s'", accept)
		}

		// Return minimal response
		json.NewEncoder(w).Encode(RepoInfo{DefaultBranch: "main"})
	}))
	defer server.Close()

	client := NewResilientGitHubClient("test-token")
	client.SetBaseURL(server.URL)

	_, err := client.GetDefaultBranch(context.Background(), "testorg", "testrepo")
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
}

func TestResilientGitHubClient_RateLimitHandling(t *testing.T) {
	resetTime := time.Now().Add(time.Hour).Unix()

	// Test server that returns rate limit error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Reset", string(rune(resetTime)))
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewResilientGitHubClient("test-token")
	client.SetBaseURL(server.URL)

	_, err := client.GetDefaultBranch(context.Background(), "testorg", "testrepo")
	if err == nil {
		t.Fatal("Expected rate limit error, got success")
	}

	if !containsString(err.Error(), "rate limited") {
		t.Errorf("Expected 'rate limited' in error, got: %v", err)
	}
}

func TestResilientGitHubClient_ContextCancellation(t *testing.T) {
	// Test server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		json.NewEncoder(w).Encode(RepoInfo{DefaultBranch: "main"})
	}))
	defer server.Close()

	client := NewResilientGitHubClient("test-token")
	client.SetBaseURL(server.URL)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.GetDefaultBranch(ctx, "testorg", "testrepo")
	if err == nil {
		t.Fatal("Expected timeout error, got success")
	}

	if !errors.Is(err, context.DeadlineExceeded) && !containsString(err.Error(), "timeout") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func TestRateLimitInfo_IsRateLimited(t *testing.T) {
	tests := []struct {
		remaining  int
		shouldRate bool
		name       string
	}{
		{100, false, "plenty of requests"},
		{10, false, "exactly at threshold"},
		{5, true, "below threshold"},
		{0, true, "no requests left"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &RateLimitInfo{
				Limit:     5000,
				Remaining: tt.remaining,
				ResetTime: time.Now().Add(time.Hour),
			}

			if info.IsRateLimited() != tt.shouldRate {
				t.Errorf("Expected IsRateLimited() = %v, got %v", tt.shouldRate, info.IsRateLimited())
			}
		})
	}
}

func TestRateLimitInfo_TimeUntilReset(t *testing.T) {
	resetTime := time.Now().Add(time.Hour)
	info := &RateLimitInfo{
		Limit:     5000,
		Remaining: 100,
		ResetTime: resetTime,
	}

	duration := info.TimeUntilReset()
	if duration <= 59*time.Minute || duration > time.Hour {
		t.Errorf("Expected ~1 hour until reset, got %v", duration)
	}
}

func TestResilientGitHubClient_SetToken(t *testing.T) {
	client := NewResilientGitHubClient("initial-token")
	client.SetToken("new-token")

	// Test server that checks for new token
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "token new-token" {
			t.Errorf("Expected 'token new-token', got '%s'", authHeader)
		}

		json.NewEncoder(w).Encode(RepoInfo{DefaultBranch: "main"})
	}))
	defer server.Close()

	client.SetBaseURL(server.URL)

	_, err := client.GetDefaultBranch(context.Background(), "testorg", "testrepo")
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
}

func TestResilientGitHubClient_Stats(t *testing.T) {
	client := NewResilientGitHubClient("test-token")
	stats := client.GetStats()

	if stats == nil {
		t.Error("Expected stats, got nil")
	}

	if _, exists := stats["config"]; !exists {
		t.Error("Expected config in stats")
	}
}

// Helper function for string containment check.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				stringContains(s, substr))))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
