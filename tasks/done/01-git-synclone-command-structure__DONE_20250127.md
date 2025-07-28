# Task: Git Synclone Command Structure Design

## 작업 목표
`git synclone` 명령어를 Git extension으로 구현하기 위한 명령어 구조를 설계합니다.

## 선행 조건
- [x] specs/synclone-git-extension.md 검토 완료
- [x] 기존 gz synclone 명령어 구조 분석 완료
- [x] Git extension 작동 방식 이해

## 구현 상세

### 1. 명령어 진입점 생성
`cmd/git-synclone/main.go` 파일 생성:
```go
package main

import (
    "github.com/spf13/cobra"
    // 기존 synclone 패키지 임포트
    "github.com/gizzahub/gzh-manager-go/pkg/synclone"
)
```

### 2. 루트 명령어 정의
```go
var rootCmd = &cobra.Command{
    Use:   "git-synclone",
    Short: "Enhanced Git cloning with provider awareness",
    Long: `git synclone provides intelligent repository cloning with support for
GitHub, GitLab, Gitea, and Gogs platforms. It offers bulk cloning, parallel
execution, and resume capabilities.`,
}
```

### 3. Provider 서브커맨드 구조
각 Git 플랫폼별 서브커맨드 생성:
- `git synclone github` - GitHub 조직 클론
- `git synclone gitlab` - GitLab 그룹 클론
- `git synclone gitea` - Gitea 조직 클론
- `git synclone all` - 설정 파일 기반 전체 클론

### 4. 공통 플래그 정의
```go
// 모든 provider에 공통으로 적용되는 플래그
--target, -t        # 클론 대상 디렉토리
--config, -c        # 설정 파일 경로
--parallel, -p      # 병렬 처리 수
--resume            # 중단된 작업 재개
--cleanup-orphans   # 고아 디렉토리 정리
--strategy          # 클론 전략 (reset/pull/fetch)
--dry-run           # 실제 실행 없이 계획만 표시
```

### 5. Provider별 플래그
#### GitHub
```go
--org, -o           # 조직 이름
--match             # 저장소 이름 패턴
--visibility        # public/private/all
--archived          # 아카이브된 저장소 포함 여부
--protocol          # https/ssh
```

#### GitLab
```go
--group, -g         # 그룹 이름
--recursive         # 하위 그룹 포함
--api-url           # 커스텀 GitLab 인스턴스
```

### 6. 기존 gz synclone과의 호환성
- 동일한 플래그 이름 사용
- 동일한 설정 파일 형식 지원
- 동일한 출력 형식 유지

## 검증 기준
- [x] `git synclone --help`가 올바른 도움말 표시
- [x] 모든 provider 서브커맨드가 정상 동작
- [x] 플래그 파싱이 올바르게 동작
- [x] 기존 gz synclone 플래그와 1:1 매핑 확인

## 참고 문서
- specs/synclone-git-extension.md
- specs/synclone.md
- cmd/synclone/synclone.go (기존 구현)

## 완료 후 다음 단계
→ 02-git-synclone-provider-integration.md
