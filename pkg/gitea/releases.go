// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package gitea

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Gizzahub/gzh-cli/internal/httpclient"
	"github.com/Gizzahub/gzh-cli/pkg/git/provider"
)

// giteaRelease represents a Gitea release from the API.
// nolint:tagliatelle // External API format - must match Gitea JSON output
type giteaRelease struct {
	ID           int64        `json:"id"`
	TagName      string       `json:"tag_name"`
	TargetBranch string       `json:"target_commitish"`
	Name         string       `json:"name"`
	Body         string       `json:"body"`
	Draft        bool         `json:"draft"`
	Prerelease   bool         `json:"prerelease"`
	CreatedAt    time.Time    `json:"created_at"`
	PublishedAt  time.Time    `json:"published_at"`
	Author       giteaAuthor  `json:"author"`
	Assets       []giteaAsset `json:"assets"`
	TarballURL   string       `json:"tarball_url"`
	ZipballURL   string       `json:"zipball_url"`
	HTMLURL      string       `json:"html_url"`
}

// giteaAuthor represents a Gitea user.
type giteaAuthor struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// giteaAsset represents a release asset.
type giteaAsset struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Size          int64     `json:"size"`
	DownloadCount int       `json:"download_count"`
	CreatedAt     time.Time `json:"created_at"`
	UUID          string    `json:"uuid"`
	BrowserURL    string    `json:"browser_download_url"`
}

// giteaCreateReleaseRequest represents a request to create a release.
type giteaCreateReleaseRequest struct {
	TagName      string `json:"tag_name"`
	TargetBranch string `json:"target_commitish,omitempty"`
	Name         string `json:"name,omitempty"`
	Body         string `json:"body,omitempty"`
	Draft        bool   `json:"draft,omitempty"`
	Prerelease   bool   `json:"prerelease,omitempty"`
}

// giteaUpdateReleaseRequest represents a request to update a release.
type giteaUpdateReleaseRequest struct {
	TagName    string `json:"tag_name,omitempty"`
	Name       string `json:"name,omitempty"`
	Body       string `json:"body,omitempty"`
	Draft      *bool  `json:"draft,omitempty"`
	Prerelease *bool  `json:"prerelease,omitempty"`
}

// buildAPIURL constructs a Gitea API URL.
func buildAPIURL(path string) string {
	return fmt.Sprintf("https://gitea.com/api/v1/%s", path)
}

// ListReleases retrieves all releases for a repository.
func ListReleases(ctx context.Context, owner, repo string, opts provider.ListReleasesOptions) (*provider.ReleaseList, error) {
	apiURL := buildAPIURL(fmt.Sprintf("repos/%s/%s/releases", url.PathEscape(owner), url.PathEscape(repo)))

	// 페이지네이션 쿼리 파라미터 추가
	if opts.Page > 0 || opts.PerPage > 0 {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", fmt.Sprintf("%d", opts.Page))
		}
		if opts.PerPage > 0 {
			params.Set("limit", fmt.Sprintf("%d", opts.PerPage))
		}
		apiURL = apiURL + "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	addAuthHeader(req)

	client := httpclient.GetGlobalClient("gitea")
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

	var giteaReleases []giteaRelease
	if err := json.Unmarshal(body, &giteaReleases); err != nil {
		return nil, fmt.Errorf("failed to parse releases: %w", err)
	}

	// Gitea 릴리스를 provider.Release로 변환
	releases := make([]provider.Release, 0, len(giteaReleases))
	for _, r := range giteaReleases {
		// Draft/Prerelease 필터링
		if !opts.IncludeDrafts && r.Draft {
			continue
		}
		if !opts.IncludePrereleases && r.Prerelease {
			continue
		}
		releases = append(releases, convertGiteaRelease(r))
	}

	// 페이지네이션 헤더 파싱 (Gitea uses X-HasMore header)
	hasNext := resp.Header.Get("X-HasMore") == "true"
	hasPrev := opts.Page > 1

	return &provider.ReleaseList{
		Releases:   releases,
		TotalCount: len(releases),
		Page:       opts.Page,
		PerPage:    opts.PerPage,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}, nil
}

// GetRelease retrieves a specific release by ID.
func GetRelease(ctx context.Context, owner, repo string, releaseID int64) (*provider.Release, error) {
	apiURL := buildAPIURL(fmt.Sprintf("repos/%s/%s/releases/%d", url.PathEscape(owner), url.PathEscape(repo), releaseID))

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	addAuthHeader(req)

	client := httpclient.GetGlobalClient("gitea")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("release not found: %d", releaseID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get release: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var r giteaRelease
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	release := convertGiteaRelease(r)
	return &release, nil
}

// GetReleaseByTag retrieves a release by tag name.
func GetReleaseByTag(ctx context.Context, owner, repo, tagName string) (*provider.Release, error) {
	apiURL := buildAPIURL(fmt.Sprintf("repos/%s/%s/releases/tags/%s",
		url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(tagName)))

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	addAuthHeader(req)

	client := httpclient.GetGlobalClient("gitea")
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

	var r giteaRelease
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	release := convertGiteaRelease(r)
	return &release, nil
}

// CreateRelease creates a new release.
func CreateRelease(ctx context.Context, owner, repo string, req provider.CreateReleaseRequest) (*provider.Release, error) {
	apiURL := buildAPIURL(fmt.Sprintf("repos/%s/%s/releases", url.PathEscape(owner), url.PathEscape(repo)))

	createReq := giteaCreateReleaseRequest{
		TagName:      req.TagName,
		Name:         req.Name,
		Body:         req.Body,
		TargetBranch: req.TargetBranch,
		Draft:        req.Draft,
		Prerelease:   req.Prerelease,
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

	client := httpclient.GetGlobalClient("gitea")
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

	var r giteaRelease
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	release := convertGiteaRelease(r)
	return &release, nil
}

// UpdateRelease updates an existing release.
func UpdateRelease(ctx context.Context, owner, repo string, releaseID int64, req provider.UpdateReleaseRequest) (*provider.Release, error) {
	apiURL := buildAPIURL(fmt.Sprintf("repos/%s/%s/releases/%d",
		url.PathEscape(owner), url.PathEscape(repo), releaseID))

	updateReq := giteaUpdateReleaseRequest{}
	if req.TagName != nil {
		updateReq.TagName = *req.TagName
	}
	if req.Name != nil {
		updateReq.Name = *req.Name
	}
	if req.Body != nil {
		updateReq.Body = *req.Body
	}
	if req.Draft != nil {
		updateReq.Draft = req.Draft
	}
	if req.Prerelease != nil {
		updateReq.Prerelease = req.Prerelease
	}

	jsonBody, err := json.Marshal(updateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PATCH", apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	addAuthHeader(httpReq)
	httpReq.Header.Set("Content-Type", "application/json")

	client := httpclient.GetGlobalClient("gitea")
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

	var r giteaRelease
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	release := convertGiteaRelease(r)
	return &release, nil
}

// DeleteRelease deletes a release.
func DeleteRelease(ctx context.Context, owner, repo string, releaseID int64) error {
	apiURL := buildAPIURL(fmt.Sprintf("repos/%s/%s/releases/%d",
		url.PathEscape(owner), url.PathEscape(repo), releaseID))

	req, err := http.NewRequestWithContext(ctx, "DELETE", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	addAuthHeader(req)

	client := httpclient.GetGlobalClient("gitea")
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

// ListReleaseAssets lists assets for a release.
func ListReleaseAssets(ctx context.Context, owner, repo string, releaseID int64) ([]provider.Asset, error) {
	apiURL := buildAPIURL(fmt.Sprintf("repos/%s/%s/releases/%d/assets",
		url.PathEscape(owner), url.PathEscape(repo), releaseID))

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	addAuthHeader(req)

	client := httpclient.GetGlobalClient("gitea")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list assets: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list assets: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var giteaAssets []giteaAsset
	if err := json.Unmarshal(body, &giteaAssets); err != nil {
		return nil, fmt.Errorf("failed to parse assets: %w", err)
	}

	assets := make([]provider.Asset, 0, len(giteaAssets))
	for _, a := range giteaAssets {
		assets = append(assets, convertGiteaAsset(a))
	}

	return assets, nil
}

// DownloadReleaseAsset downloads a release asset.
func DownloadReleaseAsset(ctx context.Context, owner, repo string, assetID int64) ([]byte, error) {
	// Gitea API endpoint for downloading asset
	apiURL := buildAPIURL(fmt.Sprintf("repos/%s/%s/releases/assets/%d",
		url.PathEscape(owner), url.PathEscape(repo), assetID))

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	addAuthHeader(req)
	req.Header.Set("Accept", "application/octet-stream")

	client := httpclient.GetGlobalClient("gitea")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download asset: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download asset: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// DeleteReleaseAsset deletes a release asset.
func DeleteReleaseAsset(ctx context.Context, owner, repo string, assetID int64) error {
	apiURL := buildAPIURL(fmt.Sprintf("repos/%s/%s/releases/assets/%d",
		url.PathEscape(owner), url.PathEscape(repo), assetID))

	req, err := http.NewRequestWithContext(ctx, "DELETE", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	addAuthHeader(req)

	client := httpclient.GetGlobalClient("gitea")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete asset: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete asset: %s", resp.Status)
	}

	return nil
}

// convertGiteaRelease converts a Gitea release to the provider.Release format.
func convertGiteaRelease(r giteaRelease) provider.Release {
	// 에셋 변환
	assets := make([]provider.Asset, 0, len(r.Assets))
	for _, a := range r.Assets {
		assets = append(assets, convertGiteaAsset(a))
	}

	return provider.Release{
		ID:           fmt.Sprintf("%d", r.ID),
		TagName:      r.TagName,
		Name:         r.Name,
		Body:         r.Body,
		Draft:        r.Draft,
		Prerelease:   r.Prerelease,
		TargetBranch: r.TargetBranch,
		HTMLURL:      r.HTMLURL,
		TarballURL:   r.TarballURL,
		ZipballURL:   r.ZipballURL,
		CreatedAt:    r.CreatedAt,
		PublishedAt:  r.PublishedAt,
		Author: provider.Actor{
			ID:        fmt.Sprintf("%d", r.Author.ID),
			Login:     r.Author.Login,
			Name:      r.Author.FullName,
			Email:     r.Author.Email,
			AvatarURL: r.Author.AvatarURL,
		},
		Assets:       assets,
		ProviderType: "gitea",
	}
}

// convertGiteaAsset converts a Gitea asset to the provider.Asset format.
func convertGiteaAsset(a giteaAsset) provider.Asset {
	return provider.Asset{
		ID:            fmt.Sprintf("%d", a.ID),
		Name:          a.Name,
		Size:          a.Size,
		DownloadCount: a.DownloadCount,
		DownloadURL:   a.BrowserURL,
		CreatedAt:     a.CreatedAt,
	}
}
