package gitplatform

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gizzahub/gzh-manager-go/internal/httpclient"
)

// BaseClient provides common functionality for all git platform clients.
type BaseClient struct {
	httpClient httpclient.HTTPClient
	token      string
	baseURL    string
	platform   string
}

// NewBaseClient creates a new base client with common configuration.
func NewBaseClient(platform, baseURL, token string) *BaseClient {
	httpClient := httpclient.NewHTTPClient(&httpclient.HTTPClientConfig{
		Timeout: 30 * time.Second,
	}, nil, nil)

	return &BaseClient{
		httpClient: httpClient,
		token:      token,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		platform:   platform,
	}
}

// SetAuthentication sets the authentication token.
func (b *BaseClient) SetAuthentication(token string) {
	b.token = token
}

// GetPlatformName returns the platform name.
func (b *BaseClient) GetPlatformName() string {
	return b.platform
}

// GetHTTPClient returns the configured HTTP client.
func (b *BaseClient) GetHTTPClient() httpclient.HTTPClient {
	return b.httpClient
}

// GetToken returns the authentication token.
func (b *BaseClient) GetToken() string {
	return b.token
}

// GetBaseURL returns the base URL.
func (b *BaseClient) GetBaseURL() string {
	return b.baseURL
}

// CreateAuthenticatedRequest creates an HTTP request with authentication headers.
func (b *BaseClient) CreateAuthenticatedRequest(ctx context.Context, method, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	if b.token != "" {
		switch b.platform {
		case "github":
			req.Header.Set("Authorization", "token "+b.token)
		case "gitlab":
			req.Header.Set("PRIVATE-TOKEN", b.token)
		case "gitea", "gogs":
			req.Header.Set("Authorization", "token "+b.token)
		default:
			req.Header.Set("Authorization", "Bearer "+b.token)
		}
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

// Helper functions that can be shared across platforms

// GetDirectories returns a list of directory names in the given path.
func GetDirectories(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var dirs []string

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	return dirs, nil
}

// Difference returns the elements in 'a' that are not in 'b'.
func Difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}

	var diff []string

	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}

	return diff
}

// Contains checks if a slice contains a specific element.
func Contains(list []string, element string) bool {
	for _, item := range list {
		if item == element {
			return true
		}
	}

	return false
}

// FileExists checks if a file exists.
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// EnsureDir creates a directory if it doesn't exist.
func EnsureDir(dir string) error {
	if !FileExists(dir) {
		return os.MkdirAll(dir, 0o755)
	}

	return nil
}

// BuildCloneURL constructs the clone URL based on protocol and repository info.
func BuildCloneURL(protocol, baseURL, owner, repoName string, _ bool) string {
	switch protocol {
	case "ssh":
		// Extract host from baseURL
		host := strings.TrimPrefix(baseURL, "https://")
		host = strings.TrimPrefix(host, "http://")
		host = strings.Split(host, "/")[0]

		return fmt.Sprintf("git@%s:%s/%s.git", host, owner, repoName)
	case "https":
		return fmt.Sprintf("%s/%s/%s.git", baseURL, owner, repoName)
	default:
		// Default to HTTPS
		return fmt.Sprintf("%s/%s/%s.git", baseURL, owner, repoName)
	}
}

// GetRepoPath returns the full path for a repository.
func GetRepoPath(targetPath, owner, repoName string, flatten bool) string {
	if flatten {
		return filepath.Join(targetPath, repoName)
	}

	return filepath.Join(targetPath, owner, repoName)
}
