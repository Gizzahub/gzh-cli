# Task: Git Repo Command Structure Design

## 작업 목표
`gz git repo` 명령어의 구조를 설계하고 서브커맨드를 정의합니다. 이는 저장소의 전체 생명주기를 관리하는 통합 인터페이스가 됩니다.

## 선행 조건
- [ ] specs/git.md 검토 완료
- [ ] tasks/gz-unified-cli-design.md의 repo 섹션 검토
- [ ] 기존 cmd/git.go 구조 분석

## 구현 상세

### 1. 명령어 계층 구조
```
gz git repo
├── clone      # 저장소 클론 (synclone 기능 통합)
├── list       # 저장소 목록 조회
├── create     # 새 저장소 생성
├── delete     # 저장소 삭제
├── archive    # 저장소 아카이브
├── sync       # 크로스 플랫폼 동기화
├── migrate    # 저장소 마이그레이션
└── search     # 저장소 검색
```

### 2. 루트 명령어 구현
```go
// cmd/git/repo.go
package git

import (
    "github.com/spf13/cobra"
)

func newGitRepoCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "repo",
        Short: "Repository lifecycle management",
        Long: `Manage repositories across Git platforms including cloning, 
creating, archiving, and synchronizing repositories.`,
        RunE: func(cmd *cobra.Command, args []string) error {
            return cmd.Help()
        },
    }
    
    // 서브커맨드 추가
    cmd.AddCommand(newRepoCloneCmd())
    cmd.AddCommand(newRepoListCmd())
    cmd.AddCommand(newRepoCreateCmd())
    cmd.AddCommand(newRepoDeleteCmd())
    cmd.AddCommand(newRepoArchiveCmd())
    cmd.AddCommand(newRepoSyncCmd())
    cmd.AddCommand(newRepoMigrateCmd())
    cmd.AddCommand(newRepoSearchCmd())
    
    return cmd
}
```

### 3. Clone 서브커맨드 (synclone 통합)
```go
func newRepoCloneCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "clone",
        Short: "Clone repositories from Git platforms",
        Long: `Clone repositories with advanced features like bulk operations,
parallel execution, and resume capability.`,
    }
    
    // Provider 플래그
    cmd.Flags().String("provider", "", "Git provider (github, gitlab, gitea)")
    cmd.Flags().String("org", "", "Organization/Group name")
    cmd.Flags().String("target", "", "Target directory")
    
    // 고급 옵션
    cmd.Flags().Int("parallel", 5, "Parallel clone workers")
    cmd.Flags().String("strategy", "reset", "Clone strategy (reset, pull, fetch)")
    cmd.Flags().Bool("resume", false, "Resume interrupted operation")
    
    // 필터링
    cmd.Flags().String("match", "", "Repository name pattern")
    cmd.Flags().String("visibility", "all", "Visibility (public, private, all)")
    
    return cmd
}
```

### 4. List 서브커맨드
```go
func newRepoListCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "list",
        Short: "List repositories from Git platforms",
        Example: `
  gz git repo list --provider github --org myorg
  gz git repo list --all-providers --format json`,
    }
    
    cmd.Flags().String("provider", "", "Git provider")
    cmd.Flags().String("org", "", "Organization name")
    cmd.Flags().String("format", "table", "Output format (table, json, yaml)")
    cmd.Flags().Bool("all-providers", false, "List from all configured providers")
    
    return cmd
}
```

### 5. Create 서브커맨드
```go
func newRepoCreateCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "create",
        Short: "Create a new repository",
        Example: `
  gz git repo create --provider github --org myorg --name newrepo
  gz git repo create --template api-template --private`,
    }
    
    cmd.Flags().String("provider", "", "Git provider")
    cmd.Flags().String("org", "", "Organization name")
    cmd.Flags().String("name", "", "Repository name")
    cmd.Flags().String("description", "", "Repository description")
    cmd.Flags().Bool("private", false, "Create private repository")
    cmd.Flags().String("template", "", "Template repository")
    cmd.Flags().Bool("auto-init", true, "Initialize with README")
    
    return cmd
}
```

### 6. Sync 서브커맨드
```go
func newRepoSyncCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "sync",
        Short: "Synchronize repositories across providers",
        Long: `Synchronize repositories between different Git platforms,
including code, issues, pull requests, and metadata.`,
    }
    
    cmd.Flags().String("from", "", "Source (provider:org/repo)")
    cmd.Flags().String("to", "", "Destination (provider:org/repo)")
    cmd.Flags().Bool("include-issues", false, "Sync issues")
    cmd.Flags().Bool("include-prs", false, "Sync pull requests")
    cmd.Flags().Bool("include-wiki", false, "Sync wiki")
    cmd.Flags().Bool("dry-run", false, "Preview sync without changes")
    
    return cmd
}
```

### 7. 공통 Provider 인터페이스
```go
// internal/git/provider.go
type RepoProvider interface {
    // 목록 조회
    ListRepositories(ctx context.Context, opts ListOptions) ([]Repository, error)
    
    // 생성/삭제
    CreateRepository(ctx context.Context, repo Repository) (*Repository, error)
    DeleteRepository(ctx context.Context, id string) error
    
    // 상태 변경
    ArchiveRepository(ctx context.Context, id string) error
    UnarchiveRepository(ctx context.Context, id string) error
    
    // 검색
    SearchRepositories(ctx context.Context, query string) ([]Repository, error)
}
```

### 8. 설정 통합
```go
// 기존 git.yaml에 repo 섹션 추가
type RepoConfig struct {
    DefaultProvider string            `yaml:"default_provider"`
    Providers       map[string]Provider `yaml:"providers"`
    CloneDefaults   CloneDefaults      `yaml:"clone_defaults"`
    SyncDefaults    SyncDefaults       `yaml:"sync_defaults"`
}
```

## 구현 체크리스트
- [ ] cmd/git/repo.go 파일 생성
- [ ] 루트 repo 명령어 구현
- [ ] clone 서브커맨드 구현 (synclone 로직 통합)
- [ ] list 서브커맨드 구현
- [ ] create 서브커맨드 구현
- [ ] delete 서브커맨드 구현
- [ ] archive 서브커맨드 구현
- [ ] sync 서브커맨드 구현
- [ ] migrate 서브커맨드 구현
- [ ] search 서브커맨드 구현

## 통합 요구사항
- [ ] cmd/git.go에 repo 명령어 등록
- [ ] 기존 synclone 코드와의 통합 계획
- [ ] Provider 인터페이스 정의
- [ ] 설정 파일 스키마 확장

## 검증 기준
- [ ] `gz git repo --help`가 모든 서브커맨드 표시
- [ ] 각 서브커맨드의 --help가 정상 동작
- [ ] 플래그 파싱이 올바르게 동작
- [ ] 기존 git 명령어와 충돌 없음

## 참고 문서
- specs/git.md의 Future Enhancements 섹션
- tasks/gz-unified-cli-design.md의 Repository Management 섹션
- cmd/synclone/synclone.go (재사용할 로직)

## 완료 후 다음 단계
→ 06-git-repo-provider-abstraction.md