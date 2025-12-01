// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package registry

import (
	"sort"
	"sync"

	"github.com/spf13/cobra"
)

// CommandCategory represents a command grouping category.
type CommandCategory string

const (
	CategoryGit         CommandCategory = "git"         // Git 작업 관련 명령어
	CategoryDevelopment CommandCategory = "development" // 개발 환경 관련 (dev-env, ide, pm)
	CategoryQuality     CommandCategory = "quality"     // 코드 품질 관련 (quality, doctor)
	CategoryNetwork     CommandCategory = "network"     // 네트워크 환경 관련 (net-env)
	CategoryUtility     CommandCategory = "utility"     // 유틸리티 (profile, shell, version)
	CategoryConfig      CommandCategory = "config"      // 설정 관리 (repo-config, synclone)
)

// LifecycleStage represents the development stage of a command.
type LifecycleStage string

const (
	LifecycleStable       LifecycleStage = "stable"       // 프로덕션 준비 완료
	LifecycleBeta         LifecycleStage = "beta"         // 기능 완성, 테스트 중
	LifecycleExperimental LifecycleStage = "experimental" // 초기 개발 단계
	LifecycleDeprecated   LifecycleStage = "deprecated"   // 제거 예정
)

// CommandMetadata contains metadata about a command.
type CommandMetadata struct {
	Name         string          // 명령어 이름 (예: "git")
	Category     CommandCategory // 명령어 카테고리
	Version      string          // 명령어 버전
	Priority     int             // 표시/실행 순서 (낮을수록 높은 우선순위)
	Experimental bool            // 실험적 기능 여부
	Dependencies []string        // 필요한 외부 도구 목록
	Tags         []string        // 검색 가능한 태그
	Lifecycle    LifecycleStage  // 개발 단계
}

// CommandProvider defines an interface that exposes a Cobra command.
// 향후 호환성을 위해 Metadata() 메서드는 선택적으로 구현 가능.
type CommandProvider interface {
	Command() *cobra.Command
}

// CommandProviderWithMetadata extends CommandProvider with metadata support.
type CommandProviderWithMetadata interface {
	CommandProvider
	Metadata() CommandMetadata
}

var (
	mu        sync.RWMutex
	providers []CommandProvider
)

// Register adds a command provider to the registry.
func Register(p CommandProvider) {
	mu.Lock()
	providers = append(providers, p)
	mu.Unlock()
}

// List returns all registered command providers.
func List() []CommandProvider {
	mu.RLock()
	defer mu.RUnlock()
	return append([]CommandProvider(nil), providers...)
}

// ByCategory returns command providers filtered by category.
func ByCategory(cat CommandCategory) []CommandProvider {
	mu.RLock()
	defer mu.RUnlock()

	var result []CommandProvider
	for _, p := range providers {
		if mp, ok := p.(CommandProviderWithMetadata); ok {
			if mp.Metadata().Category == cat {
				result = append(result, p)
			}
		}
	}
	return result
}

// StableCommands returns only stable command providers.
func StableCommands() []CommandProvider {
	mu.RLock()
	defer mu.RUnlock()

	var result []CommandProvider
	for _, p := range providers {
		if mp, ok := p.(CommandProviderWithMetadata); ok {
			if mp.Metadata().Lifecycle == LifecycleStable {
				result = append(result, p)
			}
		} else {
			// 메타데이터 없는 명령어는 stable로 간주
			result = append(result, p)
		}
	}
	return result
}

// ExperimentalCommands returns experimental command providers.
func ExperimentalCommands() []CommandProvider {
	mu.RLock()
	defer mu.RUnlock()

	var result []CommandProvider
	for _, p := range providers {
		if mp, ok := p.(CommandProviderWithMetadata); ok {
			if mp.Metadata().Experimental || mp.Metadata().Lifecycle == LifecycleExperimental {
				result = append(result, p)
			}
		}
	}
	return result
}

// AllByPriority returns all providers sorted by priority (lower number = higher priority).
func AllByPriority() []CommandProvider {
	mu.RLock()
	defer mu.RUnlock()

	result := append([]CommandProvider(nil), providers...)

	sort.Slice(result, func(i, j int) bool {
		// 메타데이터가 있는 경우 우선순위 사용
		iMeta, iOK := result[i].(CommandProviderWithMetadata)
		jMeta, jOK := result[j].(CommandProviderWithMetadata)

		if iOK && jOK {
			return iMeta.Metadata().Priority < jMeta.Metadata().Priority
		}
		// 메타데이터 없는 경우 등록 순서 유지
		return false
	})

	return result
}

// HasMetadata checks if a provider supports metadata.
func HasMetadata(p CommandProvider) bool {
	_, ok := p.(CommandProviderWithMetadata)
	return ok
}

// GetMetadata safely retrieves metadata from a provider
// Returns zero-value metadata if provider doesn't support it.
func GetMetadata(p CommandProvider) CommandMetadata {
	if mp, ok := p.(CommandProviderWithMetadata); ok {
		return mp.Metadata()
	}
	// 기본 메타데이터 반환
	return CommandMetadata{
		Name:      p.Command().Name(),
		Category:  CategoryUtility,
		Version:   "unknown",
		Priority:  999,
		Lifecycle: LifecycleStable,
	}
}
