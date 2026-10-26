// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"crypto/sha1" //nolint:gosec // Git blob 식별자는 Git 프로토콜의 SHA-1 형식이다.
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// API 클라이언트는 토큰별로 분리하며 고정 GitHub 호스트 외에는 요청하지 않는다.
type apiClient struct {
	client   *http.Client
	getToken func(bool) string
}

func productionAPI() apiClient {
	transport := &http.Transport{TLSHandshakeTimeout: 10 * time.Second, ResponseHeaderTimeout: time.Minute, IdleConnTimeout: 90 * time.Second}
	return apiClient{
		client: &http.Client{
			Transport: transport, Timeout: 2 * time.Minute,
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
		getToken: func(tap bool) string {
			if tap {
				return os.Getenv("HOMEBREW_TAP_TOKEN")
			}
			return os.Getenv("GITHUB_TOKEN")
		},
	}
}

func (api apiClient) request(ctx context.Context, tap bool, method, host, path string, body []byte, contentType string) (int, []byte, error) {
	if host != "api.github.com" && host != "uploads.github.com" {
		return 0, nil, fmt.Errorf("API host rejected")
	}
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return 0, nil, fmt.Errorf("API path rejected")
	}
	u := "https://" + host + path
	req, err := http.NewRequestWithContext(ctx, method, u, bytes.NewReader(body))
	if err != nil {
		return 0, nil, fmt.Errorf("invalid GitHub request")
	}
	req.Header.Set("Authorization", "Bearer "+api.getToken(tap))
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Content-Type", contentType)
	resp, err := api.client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("GitHub %s transport failure", method)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxJSON+1))
	if err != nil || len(b) > maxJSON {
		return resp.StatusCode, nil, fmt.Errorf("GitHub %s response limit/read failure", method)
	}
	return resp.StatusCode, b, nil
}

func (api apiClient) json(ctx context.Context, tap bool, method, path string, body any, expected int, target any) error {
	var b []byte
	if body != nil {
		var err error
		b, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode GitHub request")
		}
	}
	status, data, err := api.request(ctx, tap, method, "api.github.com", path, b, "application/json")
	if err != nil {
		return err
	}
	if status != expected {
		return fmt.Errorf("GitHub %s HTTP %d", method, status)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("GitHub %s invalid JSON", method)
	}
	return nil
}

type githubRelease struct {
	ID         int64  `json:"id"`
	Tag        string `json:"tag_name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

func (r *githubRelease) UnmarshalJSON(data []byte) error {
	// 필수 bool의 부재를 false로 해석해 공개 상태를 증명하지 않는다.
	var wire struct {
		ID         *int64  `json:"id"`
		Tag        *string `json:"tag_name"`
		Draft      *bool   `json:"draft"`
		Prerelease *bool   `json:"prerelease"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	if wire.ID == nil || *wire.ID <= 0 || wire.Tag == nil || len(*wire.Tag) > 256 || !githubTag.MatchString(*wire.Tag) || strings.Contains(*wire.Tag, "..") || strings.Contains(*wire.Tag, "//") || strings.Contains(*wire.Tag, "/.") || strings.HasSuffix(*wire.Tag, "/") || strings.HasSuffix(*wire.Tag, ".") || strings.HasSuffix(*wire.Tag, ".lock") || wire.Draft == nil || wire.Prerelease == nil {
		return fmt.Errorf("release required fields missing/invalid")
	}
	*r = githubRelease{ID: *wire.ID, Tag: *wire.Tag, Draft: *wire.Draft, Prerelease: *wire.Prerelease}
	return nil
}

func publish(ctx context.Context, o options, m releaseManifest, files map[string][]byte, run commandRunner, api apiClient) (err error) {
	if m.Snapshot {
		return fmt.Errorf("snapshot cannot publish")
	}
	first, second := api.getToken(false), api.getToken(true)
	if first == "" || second == "" || first == second {
		return fmt.Errorf("separate release and tap installation tokens required")
	}
	if err := publisherSource(ctx, o, run); err != nil {
		return err
	}
	if err := api.requireReleaseAbsent(ctx, o.Tag); err != nil {
		return err
	}
	old, err := api.tapPreflight(ctx)
	if err != nil {
		return err
	}
	created, err := api.createDraft(ctx, o, files["release-notes.md"])
	if err != nil {
		return err
	}
	stage := "draft retained"
	defer func() {
		if err != nil {
			err = fmt.Errorf("release %d: %s; inspect release/tap before retry: %w", created.ID, stage, err)
		}
	}()
	if err := api.uploadAssets(ctx, created.ID, files); err != nil {
		return err
	}
	stage = "GitHub publication uncertain; tap unchanged"
	if err := api.publishDraft(ctx, created.ID, o.Tag); err != nil {
		return err
	}
	stage = "valid public release retained; tap update uncertain"
	return api.commitCask(ctx, old, o.Tag, files["gz.rb"])
}

func publisherSource(ctx context.Context, o options, run commandRunner) error {
	git := func(args ...string) (string, error) {
		b, e := run(ctx, "git", args, cleanEnvironment(), o.Repo)
		return strings.TrimSpace(string(b)), e
	}
	head, err := git("rev-parse", "HEAD")
	if err != nil || head != o.SHA {
		return fmt.Errorf("publisher HEAD/source mismatch")
	}
	status, err := git("status", "--porcelain", "--untracked-files=normal")
	if err != nil || status != "" {
		return fmt.Errorf("publisher source must be clean")
	}
	refs, err := git("ls-remote", "--tags", "https://github.com/Gizzahub/gzh-cli.git", "refs/tags/"+o.Tag, "refs/tags/"+o.Tag+"^{}")
	if err != nil {
		return fmt.Errorf("remote tag lookup failed")
	}
	return validateRemoteTag(refs, o.Tag, o.SHA)
}

func validateRemoteTag(refs, tag, sha string) error {
	var direct, peeled string
	for _, line := range strings.Split(refs, "\n") {
		f := strings.Fields(line)
		if len(f) != 2 || !commitSHA.MatchString(f[0]) {
			return fmt.Errorf("remote tag response invalid")
		}
		switch f[1] {
		case "refs/tags/" + tag:
			if direct != "" {
				return fmt.Errorf("duplicate remote tag")
			}
			direct = f[0]
		case "refs/tags/" + tag + "^{}":
			if peeled != "" {
				return fmt.Errorf("duplicate remote peeled tag")
			}
			peeled = f[0]
		default:
			return fmt.Errorf("remote tag response unexpected ref")
		}
	}
	if direct == "" || (peeled == "" && direct != sha) || (peeled != "" && peeled != sha) {
		return fmt.Errorf("remote tag source mismatch")
	}
	return nil
}

func (api apiClient) requireReleaseAbsent(ctx context.Context, tag string) error {
	code, b, err := api.request(ctx, false, http.MethodGet, "api.github.com", "/repos/Gizzahub/gzh-cli/releases/tags/"+tag, nil, "application/json")
	if err != nil {
		return err
	}
	var missing struct {
		Message string `json:"message"`
	}
	if code != http.StatusNotFound || json.Unmarshal(b, &missing) != nil || missing.Message != "Not Found" {
		return fmt.Errorf("release absence not proven (HTTP %d); existing releases including drafts cannot be changed", code)
	}
	// 태그 lookup이 draft를 숨길 수 있으므로 전체 목록도 보수적으로 확인한다.
	for page := 1; ; page++ {
		if page > 100 {
			return fmt.Errorf("release listing exceeds limit")
		}
		var releases []githubRelease
		if err := api.json(ctx, false, "GET", "/repos/Gizzahub/gzh-cli/releases?per_page=100&page="+strconv.Itoa(page), nil, 200, &releases); err != nil {
			return err
		}
		if releases == nil {
			return fmt.Errorf("release absence requires a non-null array")
		}
		for _, r := range releases {
			if r.Tag == tag {
				return fmt.Errorf("release already exists, including draft; no writes performed")
			}
		}
		if len(releases) < 100 {
			break
		}
	}
	return nil
}

type tapState struct {
	SHA    string `json:"sha"`
	Path   string `json:"path"`
	Type   string `json:"type"`
	Create bool   `json:"-"`
}

func (api apiClient) tapPreflight(ctx context.Context) (tapState, error) {
	var old tapState
	code, b, err := api.request(ctx, true, "GET", "api.github.com", "/repos/Gizzahub/homebrew-tap/contents/Casks/gz.rb?ref=main", nil, "application/json")
	if err != nil {
		return old, err
	}
	createCask := code == 404
	if createCask {
		var missing struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(b, &missing) != nil || missing.Message != "Not Found" {
			return old, fmt.Errorf("tap absence not proven")
		}
	} else if code != 200 || json.Unmarshal(b, &old) != nil || !commitSHA.MatchString(old.SHA) || old.Path != "Casks/gz.rb" || old.Type != "file" {
		return old, fmt.Errorf("tap preflight identity not proven (HTTP %d)", code)
	}
	old.Create = createCask
	return old, nil
}

func (api apiClient) createDraft(ctx context.Context, o options, notes []byte) (githubRelease, error) {
	var created githubRelease
	if err := api.json(ctx, false, "POST", "/repos/Gizzahub/gzh-cli/releases", map[string]any{"tag_name": o.Tag, "target_commitish": o.SHA, "name": "gz " + o.Tag, "body": string(notes), "draft": true, "prerelease": false}, 201, &created); err != nil {
		return created, fmt.Errorf("draft creation uncertain; inspect remote state before retry: %w", err)
	}
	if created.ID <= 0 || created.Tag != o.Tag || !created.Draft || created.Prerelease {
		return created, fmt.Errorf("draft creation response invalid; inspect remote state before retry")
	}
	return created, nil
}

func (api apiClient) uploadAssets(ctx context.Context, id int64, files map[string][]byte) error {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		data := files[name]
		code, response, err := api.request(ctx, false, "POST", "uploads.github.com", "/repos/Gizzahub/gzh-cli/releases/"+strconv.FormatInt(id, 10)+"/assets?name="+url.QueryEscape(name), data, "application/octet-stream")
		if err != nil {
			return err
		}
		var uploaded struct {
			Name   string `json:"name"`
			Size   int64  `json:"size"`
			Digest string `json:"digest"`
			State  string `json:"state"`
		}
		if code != 201 || json.Unmarshal(response, &uploaded) != nil || uploaded.Name != name || uploaded.Size != int64(len(data)) || uploaded.Digest != "sha256:"+digest(data) || uploaded.State != "uploaded" {
			return fmt.Errorf("asset upload/digest verification failed (HTTP %d)", code)
		}
	}
	return nil
}

func (api apiClient) publishDraft(ctx context.Context, id int64, tag string) error {
	var published githubRelease
	if err := api.json(ctx, false, "PATCH", "/repos/Gizzahub/gzh-cli/releases/"+strconv.FormatInt(id, 10), map[string]bool{"draft": false}, 200, &published); err != nil {
		return err
	}
	if published.ID != id || published.Tag != tag || published.Draft || published.Prerelease {
		return fmt.Errorf("release publication response invalid")
	}
	return nil
}

func (api apiClient) commitCask(ctx context.Context, old tapState, tag string, cask []byte) error {
	var committed struct {
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
		Content struct {
			SHA  string `json:"sha"`
			Path string `json:"path"`
		} `json:"content"`
	}
	tapBody := map[string]string{"branch": "main", "message": "gz " + tag, "content": base64.StdEncoding.EncodeToString(cask)}
	expected := 201
	if !old.Create {
		tapBody["sha"] = old.SHA
		expected = 200
	}
	if err := api.json(ctx, true, "PUT", "/repos/Gizzahub/homebrew-tap/contents/Casks/gz.rb", tapBody, expected, &committed); err != nil {
		return err
	}
	if !commitSHA.MatchString(committed.Commit.SHA) || committed.Content.SHA != gitBlobSHA(cask) || committed.Content.Path != "Casks/gz.rb" {
		return fmt.Errorf("tap CAS response invalid")
	}
	return nil
}

func gitBlobSHA(data []byte) string {
	// 서명이 아니라 Git contents API가 저장한 정확한 bytes의 프로토콜 식별자다.
	object := append([]byte(fmt.Sprintf("blob %d\x00", len(data))), data...)
	sum := sha1.Sum(object) //nolint:gosec // Git blob identity interoperability.
	return fmt.Sprintf("%x", sum)
}

func decodeVersion(b []byte, result *string) error {
	var version struct {
		GitVersion string `json:"gitVersion"`
	}
	if err := json.Unmarshal(b, &version); err != nil {
		return err
	}
	*result = version.GitVersion
	return nil
}
