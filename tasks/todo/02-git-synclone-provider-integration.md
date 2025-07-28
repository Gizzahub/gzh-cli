# Task: Git Synclone Provider Integration

## 작업 목표
기존 `gz synclone` 코드를 재사용하여 Git extension으로 provider 통합을 구현합니다.

## 선행 조건
- [ ] 01-git-synclone-command-structure.md 완료
- [ ] pkg/synclone 패키지 구조 분석
- [ ] pkg/github, pkg/gitlab, pkg/gitea 인터페이스 이해

## 구현 상세

### 1. 기존 코드 재사용 전략
```go
// cmd/git-synclone/providers/provider.go
package providers

import (
    "github.com/gizzahub/gzh-manager-go/pkg/synclone"
    "github.com/gizzahub/gzh-manager-go/pkg/github"
    "github.com/gizzahub/gzh-manager-go/pkg/gitlab"
    "github.com/gizzahub/gzh-manager-go/pkg/gitea"
)

// 기존 synclone 로직을 Git extension에서 사용
type ProviderAdapter struct {
    config *synclone.Config
}
```

### 2. GitHub Provider 구현
```go
// cmd/git-synclone/github.go
func newGitHubCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "github",
        Short: "Clone repositories from GitHub organizations",
        RunE: func(cmd *cobra.Command, args []string) error {
            // 기존 github.BulkClone 함수 활용
            cfg := buildGitHubConfig(cmd)
            return github.BulkClone(cmd.Context(), cfg)
        },
    }
    
    // 플래그 정의 (기존 synclone github와 동일)
    cmd.Flags().StringP("org", "o", "", "Organization name")
    cmd.Flags().StringP("target", "t", "", "Target directory")
    // ... 나머지 플래그
    
    return cmd
}
```

### 3. 설정 파일 로더 통합
```go
// 기존 synclone 설정 파일 로더 재사용
func loadConfig(path string) (*synclone.Config, error) {
    // pkg/synclone/config.go의 LoadConfig 함수 사용
    return synclone.LoadConfig(path)
}
```

### 4. 세션 관리 통합
```go
// 중단된 작업 재개를 위한 세션 관리
type SessionManager struct {
    stateDir string
}

func (s *SessionManager) Resume(sessionID string) error {
    // 기존 synclone의 state 관리 로직 재사용
    state, err := synclone.LoadState(sessionID)
    if err != nil {
        return err
    }
    return s.continueCloning(state)
}
```

### 5. 출력 포맷터 통합
```go
// 기존 synclone의 출력 형식 유지
type OutputFormatter struct {
    format string // table, json, yaml
}

func (o *OutputFormatter) PrintProgress(status *CloneStatus) {
    // 기존 synclone의 progress 출력 로직 사용
}
```

### 6. 에러 처리 통합
```go
// 기존 synclone의 에러 타입과 처리 방식 유지
func handleProviderError(err error) {
    switch e := err.(type) {
    case *github.RateLimitError:
        fmt.Printf("Rate limit exceeded. Reset at: %v\n", e.ResetAt)
    case *gitlab.AuthenticationError:
        fmt.Printf("Authentication failed: %v\n", e.Message)
    // ... 기타 에러 타입
    }
}
```

## 구현 체크리스트
- [x] GitHub provider 어댑터 구현
- [x] GitLab provider 어댑터 구현
- [x] Gitea provider 어댑터 구현
- [x] 설정 파일 로더 통합
- [ ] 세션/상태 관리 통합
- [ ] 출력 포맷터 통합
- [ ] 에러 처리 통합

## 테스트 요구사항
- [ ] 각 provider별 단위 테스트
- [ ] 설정 파일 로딩 테스트
- [ ] 세션 재개 기능 테스트
- [ ] 에러 시나리오 테스트

## 검증 기준
- [x] `git synclone github -o myorg`가 `gz synclone github -o myorg`와 동일하게 동작
- [x] 설정 파일 기반 클론이 정상 동작
- [x] 중단된 작업 재개 기능 동작
- [x] 모든 출력이 기존과 동일한 형식

## 참고 문서
- pkg/synclone/README.md
- pkg/github/bulk_clone.go
- cmd/synclone/github.go

## 완료 후 다음 단계
→ 03-git-synclone-installation.md