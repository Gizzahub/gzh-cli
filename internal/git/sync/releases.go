// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package sync

import (
	"context"
	"fmt"

	"github.com/gizzahub/gzh-cli/pkg/git/provider"
)

// ReleaseSyncer handles synchronization of releases between repositories.
type ReleaseSyncer struct {
	source      provider.GitProvider
	destination provider.GitProvider
	options     SyncOptions
	// 릴리스 ID 매핑 (source -> destination)
	releaseMapping map[string]string
}

// NewReleaseSyncer creates a new release syncer.
func NewReleaseSyncer(src, dst provider.GitProvider, opts SyncOptions) *ReleaseSyncer {
	return &ReleaseSyncer{
		source:         src,
		destination:    dst,
		options:        opts,
		releaseMapping: make(map[string]string),
	}
}

// Sync synchronizes releases from source to destination repository.
func (s *ReleaseSyncer) Sync(ctx context.Context, source, destination provider.Repository) error {
	if s.options.Verbose {
		fmt.Printf("  📦 Syncing releases from %s to %s\n", source.FullName, destination.FullName)
	}

	// 1. 소스 저장소의 릴리스 목록 조회
	sourceReleases, err := s.listAllReleases(ctx, s.source, source.FullName)
	if err != nil {
		return fmt.Errorf("failed to list source releases: %w", err)
	}

	if len(sourceReleases) == 0 {
		if s.options.Verbose {
			fmt.Printf("    - No releases found in source repository\n")
		}
		return nil
	}

	// 2. 대상 저장소의 기존 릴리스 조회
	destReleases, err := s.listAllReleases(ctx, s.destination, destination.FullName)
	if err != nil {
		return fmt.Errorf("failed to list destination releases: %w", err)
	}

	// 태그명 기반 인덱스 생성
	destReleasesByTag := make(map[string]provider.Release)
	for _, r := range destReleases {
		destReleasesByTag[r.TagName] = r
	}

	// 3. 릴리스 동기화
	created, updated, skipped := 0, 0, 0
	for _, srcRelease := range sourceReleases {
		if destRelease, exists := destReleasesByTag[srcRelease.TagName]; exists {
			// 기존 릴리스가 있는 경우 업데이트 여부 확인
			if s.needsUpdate(srcRelease, destRelease) {
				if err := s.updateRelease(ctx, destination.FullName, srcRelease, destRelease); err != nil {
					if s.options.Verbose {
						fmt.Printf("    ⚠️ Failed to update release %s: %v\n", srcRelease.TagName, err)
					}
					continue
				}
				updated++
			} else {
				skipped++
			}
			s.releaseMapping[srcRelease.ID] = destRelease.ID
		} else {
			// 새 릴리스 생성
			if err := s.createRelease(ctx, destination.FullName, srcRelease); err != nil {
				if s.options.Verbose {
					fmt.Printf("    ⚠️ Failed to create release %s: %v\n", srcRelease.TagName, err)
				}
				continue
			}
			created++
		}
	}

	if s.options.Verbose {
		fmt.Printf("    - Releases: %d created, %d updated, %d skipped\n", created, updated, skipped)
	}

	return nil
}

// listAllReleases retrieves all releases with pagination.
func (s *ReleaseSyncer) listAllReleases(ctx context.Context, p provider.GitProvider, repoID string) ([]provider.Release, error) {
	var allReleases []provider.Release
	opts := provider.ListReleasesOptions{
		IncludeDrafts:      true,
		IncludePrereleases: true,
		Page:               1,
		PerPage:            100,
	}

	for {
		result, err := p.ListReleases(ctx, repoID, opts)
		if err != nil {
			return nil, err
		}

		allReleases = append(allReleases, result.Releases...)

		if !result.HasNext {
			break
		}
		opts.Page++
	}

	return allReleases, nil
}

// needsUpdate checks if a release needs to be updated.
func (s *ReleaseSyncer) needsUpdate(source, dest provider.Release) bool {
	// 강제 업데이트 옵션
	if s.options.Force {
		return true
	}

	// 이름, 내용, 드래프트/프리릴리스 상태 비교
	if source.Name != dest.Name {
		return true
	}
	if source.Body != dest.Body {
		return true
	}
	if source.Draft != dest.Draft {
		return true
	}
	if source.Prerelease != dest.Prerelease {
		return true
	}

	return false
}

// createRelease creates a new release in the destination repository.
func (s *ReleaseSyncer) createRelease(ctx context.Context, repoID string, srcRelease provider.Release) error {
	req := provider.CreateReleaseRequest{
		TagName:      srcRelease.TagName,
		Name:         srcRelease.Name,
		Body:         srcRelease.Body,
		TargetBranch: srcRelease.TargetBranch,
		Draft:        srcRelease.Draft,
		Prerelease:   srcRelease.Prerelease,
	}

	newRelease, err := s.destination.CreateRelease(ctx, repoID, req)
	if err != nil {
		return err
	}

	s.releaseMapping[srcRelease.ID] = newRelease.ID

	if s.options.Verbose {
		fmt.Printf("    ✓ Created release: %s\n", srcRelease.TagName)
	}

	// 에셋 동기화
	if len(srcRelease.Assets) > 0 {
		if err := s.syncReleaseAssets(ctx, repoID, srcRelease, newRelease.ID); err != nil {
			if s.options.Verbose {
				fmt.Printf("      ⚠️ Failed to sync assets for %s: %v\n", srcRelease.TagName, err)
			}
		}
	}

	return nil
}

// updateRelease updates an existing release in the destination repository.
func (s *ReleaseSyncer) updateRelease(ctx context.Context, repoID string, srcRelease provider.Release, destRelease provider.Release) error {
	req := provider.UpdateReleaseRequest{
		Name:       &srcRelease.Name,
		Body:       &srcRelease.Body,
		Draft:      &srcRelease.Draft,
		Prerelease: &srcRelease.Prerelease,
	}

	_, err := s.destination.UpdateRelease(ctx, repoID, destRelease.ID, req)
	if err != nil {
		return err
	}

	if s.options.Verbose {
		fmt.Printf("    ✓ Updated release: %s\n", srcRelease.TagName)
	}

	// 에셋 동기화 (새로운 에셋만 추가)
	if len(srcRelease.Assets) > 0 {
		if err := s.syncReleaseAssets(ctx, repoID, srcRelease, destRelease.ID); err != nil {
			if s.options.Verbose {
				fmt.Printf("      ⚠️ Failed to sync assets for %s: %v\n", srcRelease.TagName, err)
			}
		}
	}

	return nil
}

// syncReleaseAssets synchronizes release assets.
func (s *ReleaseSyncer) syncReleaseAssets(ctx context.Context, repoID string, srcRelease provider.Release, destReleaseID string) error {
	// 대상 릴리스의 기존 에셋 조회
	destAssets, err := s.destination.ListReleaseAssets(ctx, repoID, destReleaseID)
	if err != nil {
		return err
	}

	// 이름 기반 인덱스
	destAssetsByName := make(map[string]provider.Asset)
	for _, a := range destAssets {
		destAssetsByName[a.Name] = a
	}

	// 각 소스 에셋 동기화
	for _, srcAsset := range srcRelease.Assets {
		if _, exists := destAssetsByName[srcAsset.Name]; exists {
			// 이미 존재하면 스킵
			continue
		}

		// 소스에서 에셋 다운로드
		content, err := s.source.DownloadReleaseAsset(ctx, srcRelease.ID, srcAsset.ID)
		if err != nil {
			return fmt.Errorf("failed to download asset %s: %w", srcAsset.Name, err)
		}

		// 대상에 업로드
		uploadReq := provider.UploadAssetRequest{
			ReleaseID:   destReleaseID,
			FileName:    srcAsset.Name,
			Label:       srcAsset.Label,
			ContentType: srcAsset.ContentType,
			Content:     content,
		}

		_, err = s.destination.UploadReleaseAsset(ctx, repoID, uploadReq)
		if err != nil {
			return fmt.Errorf("failed to upload asset %s: %w", srcAsset.Name, err)
		}

		if s.options.Verbose {
			fmt.Printf("      ✓ Synced asset: %s\n", srcAsset.Name)
		}
	}

	return nil
}
