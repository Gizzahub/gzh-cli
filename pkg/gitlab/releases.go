// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gizzahub/gzh-cli/internal/httpclient"
	"github.com/gizzahub/gzh-cli/pkg/git/provider"
)

// gitlabRelease represents a GitLab release from the API.
// nolint:tagliatelle // External API format - must match GitLab JSON output
type gitlabRelease struct {
	Name        string        `json:"name"`
	TagName     string        `json:"tag_name"`
	Description string        `json:"description"`
	CreatedAt   time.Time     `json:"created_at"`
	ReleasedAt  time.Time     `json:"released_at"`
	Author      gitlabAuthor  `json:"author"`
	Assets      gitlabAssets  `json:"assets"`
	Links       gitlabLinks   `json:"_links"`
	Evidences   []interface{} `json:"evidences"`
	Milestones  []interface{} `json:"milestones"`
	CommitPath  string        `json:"commit_path"`
	TagPath     string        `json:"tag_path"`
}

// gitlabAuthor represents a GitLab user.
type gitlabAuthor struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	State     string `json:"state"`
	AvatarURL string `json:"avatar_url"`
	WebURL    string `json:"web_url"`
}

// gitlabAssets represents release assets.
type gitlabAssets struct {
	Count   int               `json:"count"`
	Sources []gitlabSource    `json:"sources"`
	Links   []gitlabAssetLink `json:"links"`
}

// gitlabSource represents a source archive.
type gitlabSource struct {
	Format string `json:"format"`
	URL    string `json:"url"`
}

// gitlabAssetLink represents a linked asset.
type gitlabAssetLink struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	LinkType string `json:"link_type"`
}

// gitlabLinks represents release links.
type gitlabLinks struct {
	ClosedIssuesURL        string `json:"closed_issues_url"`
	ClosedMergeRequestsURL string `json:"closed_merge_requests_url"`
	EditURL                string `json:"edit_url"`
	MergedMergeRequestsURL string `json:"merged_merge_requests_url"`
	OpenedIssuesURL        string `json:"opened_issues_url"`
	OpenedMergeRequestsURL string `json:"opened_merge_requests_url"`
	Self                   string `json:"self"`
}

// gitlabCreateReleaseRequest represents a request to create a release.
type gitlabCreateReleaseRequest struct {
	Name        string `json:"name,omitempty"`
	TagName     string `json:"tag_name"`
	Description string `json:"description,omitempty"`
	Ref         string `json:"ref,omitempty"`
	ReleasedAt  string `json:"released_at,omitempty"`
}

// gitlabUpdateReleaseRequest represents a request to update a release.
type gitlabUpdateReleaseRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	ReleasedAt  string `json:"released_at,omitempty"`
}

// ListReleases retrieves all releases for a project.
func ListReleases(ctx context.Context, projectPath string, opts provider.ListReleasesOptions) (*provider.ReleaseList, error) {
	encoded := url.PathEscape(projectPath)
	apiURL := buildAPIURL(fmt.Sprintf("projects/%s/releases", encoded))

	// 페이지네이션 쿼리 파라미터 추가
	if opts.Page > 0 || opts.PerPage > 0 {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", fmt.Sprintf("%d", opts.Page))
		}
		if opts.PerPage > 0 {
			params.Set("per_page", fmt.Sprintf("%d", opts.PerPage))
		}
		apiURL = apiURL + "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	addAuthHeader(req)

	client := httpclient.GetGlobalClient("gitlab")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list releases: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list releases: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var gitlabReleases []gitlabRelease
	if err := json.Unmarshal(body, &gitlabReleases); err != nil {
		return nil, fmt.Errorf("failed to parse releases: %w", err)
	}

	// GitLab 릴리스를 provider.Release로 변환
	releases := make([]provider.Release, 0, len(gitlabReleases))
	for _, r := range gitlabReleases {
		releases = append(releases, convertGitLabRelease(r))
	}

	// 페이지네이션 헤더 파싱
	hasNext := resp.Header.Get("X-Next-Page") != ""
	hasPrev := resp.Header.Get("X-Prev-Page") != ""

	return &provider.ReleaseList{
		Releases:   releases,
		TotalCount: len(releases),
		Page:       opts.Page,
		PerPage:    opts.PerPage,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}, nil
}

// GetRelease retrieves a specific release by tag name.
func GetRelease(ctx context.Context, projectPath, tagName string) (*provider.Release, error) {
	encoded := url.PathEscape(projectPath)
	tagEncoded := url.PathEscape(tagName)
	apiURL := buildAPIURL(fmt.Sprintf("projects/%s/releases/%s", encoded, tagEncoded))

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	addAuthHeader(req)

	client := httpclient.GetGlobalClient("gitlab")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("release not found: %s", tagName)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get release: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var r gitlabRelease
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	release := convertGitLabRelease(r)
	return &release, nil
}

// CreateRelease creates a new release.
func CreateRelease(ctx context.Context, projectPath string, req provider.CreateReleaseRequest) (*provider.Release, error) {
	encoded := url.PathEscape(projectPath)
	apiURL := buildAPIURL(fmt.Sprintf("projects/%s/releases", encoded))

	createReq := gitlabCreateReleaseRequest{
		TagName:     req.TagName,
		Name:        req.Name,
		Description: req.Body,
		Ref:         req.TargetBranch,
	}

	jsonBody, err := json.Marshal(createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	addAuthHeader(httpReq)
	httpReq.Header.Set("Content-Type", "application/json")

	client := httpclient.GetGlobalClient("gitlab")
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create release: %s - %s", resp.Status, string(respBody))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var r gitlabRelease
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	release := convertGitLabRelease(r)
	return &release, nil
}

// UpdateRelease updates an existing release.
func UpdateRelease(ctx context.Context, projectPath, tagName string, req provider.UpdateReleaseRequest) (*provider.Release, error) {
	encoded := url.PathEscape(projectPath)
	tagEncoded := url.PathEscape(tagName)
	apiURL := buildAPIURL(fmt.Sprintf("projects/%s/releases/%s", encoded, tagEncoded))

	updateReq := gitlabUpdateReleaseRequest{}
	if req.Name != nil {
		updateReq.Name = *req.Name
	}
	if req.Body != nil {
		updateReq.Description = *req.Body
	}

	jsonBody, err := json.Marshal(updateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	addAuthHeader(httpReq)
	httpReq.Header.Set("Content-Type", "application/json")

	client := httpclient.GetGlobalClient("gitlab")
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to update release: %s - %s", resp.Status, string(respBody))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var r gitlabRelease
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	release := convertGitLabRelease(r)
	return &release, nil
}

// DeleteRelease deletes a release.
func DeleteRelease(ctx context.Context, projectPath, tagName string) error {
	encoded := url.PathEscape(projectPath)
	tagEncoded := url.PathEscape(tagName)
	apiURL := buildAPIURL(fmt.Sprintf("projects/%s/releases/%s", encoded, tagEncoded))

	req, err := http.NewRequestWithContext(ctx, "DELETE", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	addAuthHeader(req)

	client := httpclient.GetGlobalClient("gitlab")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete release: %s", resp.Status)
	}

	return nil
}

// convertGitLabRelease converts a GitLab release to the provider.Release format.
func convertGitLabRelease(r gitlabRelease) provider.Release {
	// 에셋 변환
	assets := make([]provider.Asset, 0, len(r.Assets.Links))
	for _, link := range r.Assets.Links {
		assets = append(assets, provider.Asset{
			ID:          fmt.Sprintf("%d", link.ID),
			Name:        link.Name,
			DownloadURL: link.URL,
		})
	}

	// 소스 아카이브도 에셋으로 추가
	for _, src := range r.Assets.Sources {
		assets = append(assets, provider.Asset{
			Name:        fmt.Sprintf("Source code (%s)", src.Format),
			DownloadURL: src.URL,
			ContentType: fmt.Sprintf("application/%s", src.Format),
		})
	}

	return provider.Release{
		ID:          r.TagName, // GitLab은 태그명을 ID로 사용
		TagName:     r.TagName,
		Name:        r.Name,
		Body:        r.Description,
		Draft:       false, // GitLab doesn't have draft releases
		Prerelease:  false, // GitLab doesn't have prerelease flag
		HTMLURL:     r.Links.Self,
		CreatedAt:   r.CreatedAt,
		PublishedAt: r.ReleasedAt,
		Author: provider.Actor{
			ID:        fmt.Sprintf("%d", r.Author.ID),
			Login:     r.Author.Username,
			Name:      r.Author.Name,
			AvatarURL: r.Author.AvatarURL,
			HTMLURL:   r.Author.WebURL,
		},
		Assets:       assets,
		ProviderType: "gitlab",
	}
}
