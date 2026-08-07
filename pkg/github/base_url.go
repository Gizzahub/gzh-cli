package github

import "strings"

// DefaultGitHubAPIBaseURL is the public GitHub REST API host.
// GHES and tests can override via constructor baseURL parameters or config fields.
const DefaultGitHubAPIBaseURL = "https://api.github.com"

// ResolveGitHubAPIBaseURL returns DefaultGitHubAPIBaseURL when baseURL is empty
// or whitespace-only; otherwise returns the trimmed value without a trailing slash.
func ResolveGitHubAPIBaseURL(baseURL string) string {
	trimmed := strings.TrimSpace(baseURL)
	if trimmed == "" {
		return DefaultGitHubAPIBaseURL
	}
	return strings.TrimRight(trimmed, "/")
}
