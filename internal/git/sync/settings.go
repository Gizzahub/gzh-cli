// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package sync

import (
	"context"
	"fmt"

	"github.com/gizzahub/gzh-cli/pkg/git/provider"
)

// SettingsSyncer handles synchronization of repository settings.
type SettingsSyncer struct {
	source      provider.GitProvider
	destination provider.GitProvider
	options     SyncOptions
}

// NewSettingsSyncer creates a new settings syncer.
func NewSettingsSyncer(src, dst provider.GitProvider, opts SyncOptions) *SettingsSyncer {
	return &SettingsSyncer{
		source:      src,
		destination: dst,
		options:     opts,
	}
}

// Sync synchronizes repository settings from source to destination.
func (s *SettingsSyncer) Sync(ctx context.Context, source, destination provider.Repository) error {
	if s.options.Verbose {
		fmt.Printf("  ⚙️ Syncing settings from %s to %s\n", source.FullName, destination.FullName)
	}

	// 설정 변경사항 수집
	updates := s.buildSettingsUpdate(source, destination)
	if updates == nil {
		if s.options.Verbose {
			fmt.Printf("    - No settings changes needed\n")
		}
		return nil
	}

	// 설정 업데이트 적용
	_, err := s.destination.UpdateRepository(ctx, destination.FullName, *updates)
	if err != nil {
		return fmt.Errorf("failed to update repository settings: %w", err)
	}

	if s.options.Verbose {
		s.printSettingsChanges(source, destination)
	}

	return nil
}

// buildSettingsUpdate creates an update request based on differences between repositories.
func (s *SettingsSyncer) buildSettingsUpdate(source, destination provider.Repository) *provider.UpdateRepoRequest {
	updates := &provider.UpdateRepoRequest{}
	hasChanges := false

	// 설명 동기화
	if source.Description != destination.Description {
		updates.Description = &source.Description
		hasChanges = true
	}

	// 기본 브랜치 동기화 (강제 모드에서만)
	if s.options.Force && source.DefaultBranch != destination.DefaultBranch {
		updates.DefaultBranch = &source.DefaultBranch
		hasChanges = true
	}

	// 토픽 동기화
	if !equalStringSlices(source.Topics, destination.Topics) {
		updates.Topics = source.Topics
		hasChanges = true
	}

	// 가시성 동기화 (강제 모드에서만 - 위험할 수 있음)
	if s.options.Force && source.Visibility != destination.Visibility {
		updates.Visibility = source.Visibility
		hasChanges = true
	}

	// 아카이브 상태 동기화
	if source.Archived != destination.Archived {
		updates.Archived = &source.Archived
		hasChanges = true
	}

	if !hasChanges {
		return nil
	}

	return updates
}

// printSettingsChanges prints the settings changes made.
func (s *SettingsSyncer) printSettingsChanges(source, destination provider.Repository) {
	if source.Description != destination.Description {
		fmt.Printf("    ✓ Updated description\n")
	}
	if source.DefaultBranch != destination.DefaultBranch && s.options.Force {
		fmt.Printf("    ✓ Updated default branch: %s → %s\n", destination.DefaultBranch, source.DefaultBranch)
	}
	if !equalStringSlices(source.Topics, destination.Topics) {
		fmt.Printf("    ✓ Updated topics: %v\n", source.Topics)
	}
	if source.Visibility != destination.Visibility && s.options.Force {
		fmt.Printf("    ✓ Updated visibility: %s → %s\n", destination.Visibility, source.Visibility)
	}
	if source.Archived != destination.Archived {
		fmt.Printf("    ✓ Updated archived status: %v → %v\n", destination.Archived, source.Archived)
	}
}

// equalStringSlices compares two string slices for equality.
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// 슬라이스를 맵으로 변환하여 비교
	setA := make(map[string]struct{})
	for _, v := range a {
		setA[v] = struct{}{}
	}

	for _, v := range b {
		if _, exists := setA[v]; !exists {
			return false
		}
	}

	return true
}
