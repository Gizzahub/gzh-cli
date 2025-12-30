package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gizzahub/gzh-cli/internal/httpclient"
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
			return fmt.Errorf("private instance or group detected\n%s", accessGuidanceMessage())
		}
		return fmt.Errorf("access denied (status %d)\n%s", resp.StatusCode, accessGuidanceMessage())
	case http.StatusNotFound:
		// Could be wrong path or hidden private group
		if configuredToken == "" {
			return fmt.Errorf("group not found (404). Check group path or use numeric group ID\n%s", accessGuidanceMessage())
		}
		return fmt.Errorf("group not found (404). Check group path or use numeric group ID\n%s", accessGuidanceMessage())
	default:
		return fmt.Errorf("unexpected response: %s", resp.Status)
	}
}

// PreflightCheckGitAccess checks if HTTP git access is available for the group's first project.
// Returns warning messages if HTTP clone is not available.
func PreflightCheckGitAccess(ctx context.Context, group string) []string {
	var warnings []string

	// First get group projects to check at least one
	encodedGroup := url.PathEscape(group)
	reqURL := buildAPIURL(fmt.Sprintf("groups/%s/projects?per_page=1", encodedGroup))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		warnings = append(warnings, "Could not check git access availability")
		return warnings
	}
	addAuthHeader(req)

	client := httpclient.GetGlobalClient("gitlab")
	resp, err := client.Do(req)
	if err != nil {
		warnings = append(warnings, "Could not check git access availability")
		return warnings
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return warnings // Skip check if can't get projects
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return warnings
	}

	var projects []struct {
		HTTPURLToRepo string `json:"http_url_to_repo"`
		SSHURLToRepo  string `json:"ssh_url_to_repo"`
		Name          string `json:"name"`
	}

	if err := json.Unmarshal(body, &projects); err != nil {
		return warnings
	}

	if len(projects) == 0 {
		warnings = append(warnings, "No projects found in group to verify git access")
		return warnings
	}

	// Check if HTTP URL is available
	project := projects[0]
	if project.HTTPURLToRepo == "" {
		warnings = append(warnings, fmt.Sprintf(
			"⚠️  HTTP git access appears disabled on this GitLab instance.\n"+
				"   Consider using SSH keys or contact administrator.\n"+
				"   Example project '%s' has no HTTP clone URL.", project.Name))
	}

	return warnings
}
