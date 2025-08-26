package gitlab

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Gizzahub/gzh-cli/internal/httpclient"
)

// PreflightCheckGroupAccess checks if the group exists and is accessible.
// If the instance or group is private and no token is configured, it returns a helpful error.
func PreflightCheckGroupAccess(ctx context.Context, group string) error {
	if group == "" {
		return fmt.Errorf("group is required")
	}

	encodedGroup := url.PathEscape(group)
	reqURL := buildAPIURL(fmt.Sprintf("groups/%s", encodedGroup))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create preflight request: %w", err)
	}
	addAuthHeader(req)

	client := httpclient.GetGlobalClient("gitlab")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to contact gitlab: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized, http.StatusForbidden:
		if configuredToken == "" {
			return fmt.Errorf("private instance or group detected; provide token via --token or GITLAB_TOKEN")
		}
		return fmt.Errorf("access denied (status %d)", resp.StatusCode)
	case http.StatusNotFound:
		// Could be wrong path or hidden private group
		if configuredToken == "" {
			return fmt.Errorf("group not found (404). Check group path or use numeric group ID; if private, provide token via --token or GITLAB_TOKEN")
		}
		return fmt.Errorf("group not found (404). Check group path or use numeric group ID")
	default:
		return fmt.Errorf("unexpected response: %s", resp.Status)
	}
}
