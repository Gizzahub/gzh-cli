package compat

import (
	"os/exec"
)

// 필터 체인용 포스트 액션 정의
// 실제 실행할 커맨드와 환경변수, 설명, 에러 무시 여부를 포함한다.
type PostAction struct {
	Command     []string
	Env         map[string]string
	Description string
	IgnoreError bool
}

// 필터 심각도
type FilterKind string

const (
	FilterKindAdvisory FilterKind = "advisory"
	FilterKindConflict FilterKind = "conflict"
)

// CompatibilityFilter 는 특정 매니저/플러그인 조합에 적용되는 호환성 필터 인터페이스다.
// Env 는 자식 프로세스에 주입할 환경변수를 제공하고,
// Warning 은 사용자에게 알릴 경고/가이드 메시지를 반환하며,
// PostActions 는 설치 직후 수행할 선택적 후속 작업을 반환한다.
// Kind 는 충돌(conflict)인지 권고(advisory)인지 구분한다.
type CompatibilityFilter interface {
	Applies(manager, plugin string) bool
	Env() map[string]string
	Warning() string
	PostActions() []PostAction
	Kind() FilterKind
}

// build-in 필터 목록
var builtinFilters = []CompatibilityFilter{
	&asdfRustRustupFilter{},
	&asdfNodejsCorepackFilter{},
}

// BuildFilterChain 은 매니저/플러그인 조합에 적용 가능한 필터를 순서대로 반환한다.
func BuildFilterChain(manager, plugin string) []CompatibilityFilter {
	chain := make([]CompatibilityFilter, 0, len(builtinFilters))
	for _, f := range builtinFilters {
		if f.Applies(manager, plugin) {
			chain = append(chain, f)
		}
	}
	// 사용자 정의 필터 병합
	for _, uf := range loadUserFilters() {
		if uf.Applies(manager, plugin) {
			chain = append(chain, uf)
		}
	}
	return chain
}

// MergeEnvFromFilters 는 여러 필터의 Env를 순서대로 병합한다. 나중 필터가 우선한다.
func MergeEnvFromFilters(filters []CompatibilityFilter) map[string]string {
	merged := make(map[string]string)
	for _, f := range filters {
		for k, v := range f.Env() {
			merged[k] = v
		}
	}
	return merged
}

// CollectWarnings 는 모든 필터의 경고 메시지를 수집해 반환한다.
func CollectWarnings(filters []CompatibilityFilter) []string {
	warnings := make([]string, 0, len(filters))
	for _, f := range filters {
		if w := f.Warning(); w != "" {
			warnings = append(warnings, w)
		}
	}
	return warnings
}

// CountConflicts 는 체인 내 conflict 필터 개수를 반환한다.
func CountConflicts(filters []CompatibilityFilter) int {
	count := 0
	for _, f := range filters {
		if f.Kind() == FilterKindConflict {
			count++
		}
	}
	return count
}

// CollectPostActions 는 모든 필터의 후속 작업을 순서대로 수집한다.
func CollectPostActions(filters []CompatibilityFilter) []PostAction {
	actions := make([]PostAction, 0)
	for _, f := range filters {
		actions = append(actions, f.PostActions()...)
	}
	return actions
}

// NewCommandWithEnv 는 주어진 커맨드에 필터 체인의 환경변수를 병합하여 exec.Cmd 를 생성한다.
func NewCommandWithEnv(name string, arg []string, filters []CompatibilityFilter) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	env := MergeEnvWithProcessEnv(MergeEnvFromFilters(filters))
	cmd.Env = env
	return cmd
}

// MergeEnvWithProcessEnv 는 기존 함수 MergeWithProcessEnv 의 alias로 제공한다.
func MergeEnvWithProcessEnv(custom map[string]string) []string {
	return MergeWithProcessEnv(custom)
}

// === 개별 필터 구현 ===

// asdf + rust: rustup PATH 체크 우회 및 비대화형 진행
type asdfRustRustupFilter struct{}

func (f *asdfRustRustupFilter) Applies(manager, plugin string) bool {
	return manager == "asdf" && plugin == "rust"
}

func (f *asdfRustRustupFilter) Env() map[string]string {
	return map[string]string{
		"RUSTUP_INIT_SKIP_PATH_CHECK": "yes",
		"RUSTUP_INIT_YES":             "1",
	}
}

func (f *asdfRustRustupFilter) Warning() string {
	return "호환성: rustup PATH 체크를 우회합니다. 권장: rustup 단일 관리로 전환하거나 asdf rust 중복 PATH를 정리하세요. (rustup default stable; asdf plugin remove rust)"
}

func (f *asdfRustRustupFilter) PostActions() []PostAction { return nil }

func (f *asdfRustRustupFilter) Kind() FilterKind { return FilterKindConflict }

// asdf + nodejs: corepack 권장 활성화
type asdfNodejsCorepackFilter struct{}

func (f *asdfNodejsCorepackFilter) Applies(manager, plugin string) bool {
	return manager == "asdf" && plugin == "nodejs"
}

func (f *asdfNodejsCorepackFilter) Env() map[string]string {
	return map[string]string{
		// Node 16.9+ 에서 corepack을 활성화해 yarn/pnpm 관리 일관성 확보
		"COREPACK_ENABLE": "1",
	}
}

func (f *asdfNodejsCorepackFilter) Warning() string {
	return "권장: corepack를 활성화해 yarn/pnpm 전역 설치/버전 충돌을 방지합니다. 필요 시 'corepack enable'을 실행합니다."
}

func (f *asdfNodejsCorepackFilter) PostActions() []PostAction {
	return []PostAction{
		{
			Command:     []string{"bash", "-lc", "corepack enable"},
			Env:         nil,
			Description: "Enable corepack",
			IgnoreError: true,
		},
	}
}

func (f *asdfNodejsCorepackFilter) Kind() FilterKind { return FilterKindAdvisory }
