// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package provider

import (
	"time"
)

// Release represents a repository release/tag.
type Release struct {
	ID           string    `json:"id"`
	TagName      string    `json:"tag_name"`
	Name         string    `json:"name"`
	Body         string    `json:"body"`
	Draft        bool      `json:"draft"`
	Prerelease   bool      `json:"prerelease"`
	TargetBranch string    `json:"target_branch,omitempty"`
	HTMLURL      string    `json:"html_url"`
	TarballURL   string    `json:"tarball_url"`
	ZipballURL   string    `json:"zipball_url"`
	CreatedAt    time.Time `json:"created_at"`
	PublishedAt  time.Time `json:"published_at"`
	Author       Actor     `json:"author"`
	Assets       []Asset   `json:"assets"`

	// 프로바이더별 메타데이터
	ProviderType string                 `json:"provider_type"`
	ProviderData map[string]interface{} `json:"provider_data"`
}

// Asset represents a release asset (file attachment).
type Asset struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Label         string    `json:"label,omitempty"`
	ContentType   string    `json:"content_type"`
	Size          int64     `json:"size"`
	DownloadCount int       `json:"download_count"`
	DownloadURL   string    `json:"download_url"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ReleaseList represents a paginated list of releases.
type ReleaseList struct {
	Releases   []Release `json:"releases"`
	TotalCount int       `json:"total_count"`
	Page       int       `json:"page"`
	PerPage    int       `json:"per_page"`
	HasNext    bool      `json:"has_next"`
	HasPrev    bool      `json:"has_prev"`
}

// ListReleasesOptions defines options for listing releases.
type ListReleasesOptions struct {
	// 필터링
	IncludeDrafts      bool `json:"include_drafts,omitempty"`
	IncludePrereleases bool `json:"include_prereleases,omitempty"`

	// 페이지네이션
	Page    int `json:"page,omitempty"`
	PerPage int `json:"per_page,omitempty"`
}

// CreateReleaseRequest represents a request to create a release.
type CreateReleaseRequest struct {
	TagName       string `json:"tag_name"`
	Name          string `json:"name,omitempty"`
	Body          string `json:"body,omitempty"`
	TargetBranch  string `json:"target_commitish,omitempty"`
	Draft         bool   `json:"draft,omitempty"`
	Prerelease    bool   `json:"prerelease,omitempty"`
	GenerateNotes bool   `json:"generate_release_notes,omitempty"`
}

// UpdateReleaseRequest represents a request to update a release.
type UpdateReleaseRequest struct {
	TagName    *string `json:"tag_name,omitempty"`
	Name       *string `json:"name,omitempty"`
	Body       *string `json:"body,omitempty"`
	Draft      *bool   `json:"draft,omitempty"`
	Prerelease *bool   `json:"prerelease,omitempty"`
}

// UploadAssetRequest represents a request to upload a release asset.
type UploadAssetRequest struct {
	ReleaseID   string `json:"release_id"`
	FileName    string `json:"file_name"`
	Label       string `json:"label,omitempty"`
	ContentType string `json:"content_type"`
	Content     []byte `json:"content"`
}
