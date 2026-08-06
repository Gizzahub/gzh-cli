package testutil_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gizzahub/gzh-cli/internal/testutil/fixtures"
	"github.com/gizzahub/gzh-cli/internal/testutil/helpers"
	"github.com/gizzahub/gzh-cli/internal/testutil/mocks"
)

const (
	// userEndpoint is the GitHub API user endpoint.
	userEndpoint = "/user"
	// reposEndpoint is the GitHub API repositories endpoint.
	reposEndpoint = "/repos"
)

// Example of using test helpers.
func TestExampleWithHelpers(t *testing.T) {
	// Create temporary directory
	tempDir, cleanup := helpers.TempDir(t, "test-*")
	defer cleanup()

	// Set environment variables
	cleanupEnv := helpers.SetEnvs(t, map[string]string{
		"GITHUB_TOKEN": "test-token",
		"GZH_DEBUG":    "true",
	})
	defer cleanupEnv()

	// Create test configuration
	configPath := helpers.CreateTestConfig(t, tempDir, fixtures.MinimalConfig)

	// Assert file exists
	helpers.AssertFileExists(t, configPath)
	helpers.AssertFileContains(t, configPath, "test-org")

	// Create test repository structure
	repo := helpers.CreateTestRepo(t, tempDir, "test-repo", map[string]string{
		"README.md":   "# Test Repo",
		"src/main.go": "package main",
		".gitignore":  "*.tmp",
	})

	// Assert repository structure
	helpers.AssertGitRepository(t, repo)
	helpers.AssertFileExists(t, repo+"/README.md")
}

// Example of using mock HTTP client.
func TestExampleWithMockHTTP(t *testing.T) {
	// Create mock HTTP client
	mockClient := &mocks.MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Return different responses based on URL.
			//
			// 아래 요청은 api.github.com으로 나가므로 경로가 /user, /repos다.
			// 예전에는 GitHub Enterprise 접두어(/api/v3/...)로 분기해서 두
			// 요청 모두 default로 떨어져 404가 나왔다. 85~86줄의 어설션도
			// /user, /repos를 기대하고 있어 분기 쪽이 틀린 것이 분명하다.
			switch req.URL.Path {
			case userEndpoint:
				return mocks.NewMockJSONResponse(200, `{"login":"testuser"}`), nil
			case reposEndpoint:
				return mocks.NewMockJSONResponse(200, `[{"name":"repo1"},{"name":"repo2"}]`), nil
			default:
				return mocks.NewMockResponse(404, "Not Found"), nil
			}
		},
	}

	// Make some requests
	req1, _ := http.NewRequestWithContext(context.Background(), "GET", "https://api.github.com/user", http.NoBody)
	req2, _ := http.NewRequestWithContext(context.Background(), "GET", "https://api.github.com/repos", http.NoBody)

	resp1, err := mockClient.Do(req1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)
	_ = resp1.Body.Close()

	resp2, err := mockClient.Do(req2)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)
	_ = resp2.Body.Close()

	// Verify calls were recorded
	assert.Len(t, mockClient.Calls, 2)
	assert.Equal(t, userEndpoint, mockClient.Calls[0].URL.Path)
	assert.Equal(t, reposEndpoint, mockClient.Calls[1].URL.Path)
}

// Example of using fixtures.
func TestExampleWithFixtures(t *testing.T) {
	tempDir, cleanup := helpers.TempDir(t, "config-test-*")
	defer cleanup()

	// Test with different fixture configurations
	testCases := []struct {
		name   string
		config string
		valid  bool
	}{
		{"minimal", fixtures.MinimalConfig, true},
		{"complex", fixtures.ComplexConfig, true},
		{"invalid", fixtures.InvalidConfig, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configPath := helpers.CreateTestConfig(t, tempDir, tc.config)
			helpers.AssertFileExists(t, configPath)

			// In a real test, you would load and validate the config
			// cfg, err := config.LoadConfigFromFile(configPath)
			// if tc.valid {
			//     assert.NoError(t, err)
			// } else {
			//     assert.Error(t, err)
			// }
		})
	}
}
