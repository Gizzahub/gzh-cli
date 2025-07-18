package testutil_test

import (
	"net/http"
	"testing"

	"github.com/gizzahub/gzh-manager-go/internal/testutil/fixtures"
	"github.com/gizzahub/gzh-manager-go/internal/testutil/helpers"
	"github.com/gizzahub/gzh-manager-go/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
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
			// Return different responses based on URL
			switch req.URL.Path {
			case "/api/v3/user":
				return mocks.NewMockJSONResponse(200, `{"login":"testuser"}`), nil
			case "/api/v3/repos":
				return mocks.NewMockJSONResponse(200, `[{"name":"repo1"},{"name":"repo2"}]`), nil
			default:
				return mocks.NewMockResponse(404, "Not Found"), nil
			}
		},
	}

	// Make some requests
	req1, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req2, _ := http.NewRequest("GET", "https://api.github.com/repos", nil)

	resp1, err := mockClient.Do(req1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)
	resp1.Body.Close()

	resp2, err := mockClient.Do(req2)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)
	resp2.Body.Close()

	// Verify calls were recorded
	assert.Len(t, mockClient.Calls, 2)
	assert.Equal(t, "/user", mockClient.Calls[0].URL.Path)
	assert.Equal(t, "/repos", mockClient.Calls[1].URL.Path)
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
