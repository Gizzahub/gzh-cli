# Task: Git Repo Clone Implementation

## 작업 목표
`gz git repo clone` 명령어를 구현하여 기존 synclone 기능을 통합하고 확장합니다.

## 선행 조건
- [ ] 05-git-repo-command-structure.md 완료
- [ ] 06-git-repo-provider-abstraction.md 완료
- [ ] 기존 cmd/synclone 로직 분석
- [ ] pkg/synclone 상태 관리 이해

## 구현 상세

### 1. Clone 명령어 구현
```go
// cmd/git/repo_clone.go
package git

import (
    "context"
    "fmt"
    "github.com/spf13/cobra"
    "github.com/gizzahub/gzh-manager-go/pkg/git/provider"
    "github.com/gizzahub/gzh-manager-go/internal/git/clone"
)

func newRepoCloneCmd() *cobra.Command {
    var opts CloneOptions
    
    cmd := &cobra.Command{
        Use:   "clone",
        Short: "Clone repositories from Git platforms",
        Long: `Clone repositories with advanced features:
- Bulk operations for entire organizations
- Parallel execution with configurable workers
- Resume capability for interrupted operations
- Multiple clone strategies (reset, pull, fetch)`,
        Example: `
  # Clone all repos from GitHub organization
  gz git repo clone --provider github --org myorg
  
  # Clone with filters
  gz git repo clone --provider gitlab --group mygroup --match "api-*"
  
  # Resume interrupted operation
  gz git repo clone --resume abc123`,
        RunE: func(cmd *cobra.Command, args []string) error {
            return runClone(cmd.Context(), opts)
        },
    }
    
    // Provider 옵션
    cmd.Flags().StringVar(&opts.Provider, "provider", "", "Git provider (github, gitlab, gitea, gogs)")
    cmd.Flags().StringVar(&opts.Org, "org", "", "Organization/Group name")
    cmd.Flags().StringVar(&opts.Target, "target", ".", "Target directory")
    
    // 고급 옵션
    cmd.Flags().IntVar(&opts.Parallel, "parallel", 5, "Number of parallel workers")
    cmd.Flags().StringVar(&opts.Strategy, "strategy", "reset", "Clone strategy (reset, pull, fetch)")
    cmd.Flags().StringVar(&opts.Resume, "resume", "", "Resume session ID")
    
    // 필터링
    cmd.Flags().StringVar(&opts.Match, "match", "", "Repository name pattern (regex)")
    cmd.Flags().StringVar(&opts.Visibility, "visibility", "all", "Visibility (public, private, all)")
    cmd.Flags().BoolVar(&opts.IncludeArchived, "include-archived", false, "Include archived repositories")
    
    // 출력
    cmd.Flags().StringVar(&opts.Format, "format", "progress", "Output format (progress, json, quiet)")
    cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Preview without cloning")
    
    return cmd
}
```

### 2. Clone Options 구조체
```go
// internal/git/clone/options.go
package clone

type CloneOptions struct {
    // Provider 설정
    Provider string
    Org      string
    Target   string
    
    // 실행 옵션
    Parallel int
    Strategy CloneStrategy
    Resume   string
    
    // 필터
    Match           string
    Visibility      string
    IncludeArchived bool
    
    // 출력
    Format string
    DryRun bool
}

type CloneStrategy string

const (
    StrategyReset CloneStrategy = "reset" // git reset --hard && git pull
    StrategyPull  CloneStrategy = "pull"  // git pull (merge)
    StrategyFetch CloneStrategy = "fetch" // git fetch only
)
```

### 3. Clone 실행 로직
```go
// internal/git/clone/executor.go
package clone

import (
    "context"
    "sync"
    "github.com/gizzahub/gzh-manager-go/pkg/git/provider"
)

type CloneExecutor struct {
    provider provider.GitProvider
    options  CloneOptions
    session  *Session
}

func NewCloneExecutor(p provider.GitProvider, opts CloneOptions) *CloneExecutor {
    return &CloneExecutor{
        provider: p,
        options:  opts,
        session:  NewSession(opts),
    }
}

func (e *CloneExecutor) Execute(ctx context.Context) error {
    // 1. 세션 초기화 또는 복원
    if e.options.Resume != "" {
        if err := e.session.Load(e.options.Resume); err != nil {
            return fmt.Errorf("failed to resume session: %w", err)
        }
    }
    
    // 2. 저장소 목록 조회
    repos, err := e.listRepositories(ctx)
    if err != nil {
        return err
    }
    
    // 3. 필터링 적용
    filtered := e.filterRepositories(repos)
    
    // 4. Dry run 처리
    if e.options.DryRun {
        return e.printDryRun(filtered)
    }
    
    // 5. 병렬 클론 실행
    return e.cloneParallel(ctx, filtered)
}

func (e *CloneExecutor) cloneParallel(ctx context.Context, repos []provider.Repository) error {
    sem := make(chan struct{}, e.options.Parallel)
    errChan := make(chan error, len(repos))
    var wg sync.WaitGroup
    
    for _, repo := range repos {
        // 이미 클론된 저장소 스킵
        if e.session.IsCompleted(repo.FullName) {
            continue
        }
        
        wg.Add(1)
        go func(r provider.Repository) {
            defer wg.Done()
            sem <- struct{}{}
            defer func() { <-sem }()
            
            if err := e.cloneRepository(ctx, r); err != nil {
                errChan <- fmt.Errorf("%s: %w", r.FullName, err)
                e.session.MarkFailed(r.FullName, err)
            } else {
                e.session.MarkCompleted(r.FullName)
            }
            
            // 진행상황 저장
            e.session.Save()
        }(repo)
    }
    
    wg.Wait()
    close(errChan)
    
    // 에러 수집
    var errs []error
    for err := range errChan {
        errs = append(errs, err)
    }
    
    if len(errs) > 0 {
        return fmt.Errorf("clone failed for %d repositories", len(errs))
    }
    
    return nil
}
```

### 4. 저장소 클론 로직
```go
// internal/git/clone/repository.go
package clone

func (e *CloneExecutor) cloneRepository(ctx context.Context, repo provider.Repository) error {
    targetPath := filepath.Join(e.options.Target, repo.FullName)
    
    // 디렉토리 존재 확인
    if exists, err := dirExists(targetPath); err != nil {
        return err
    } else if exists {
        return e.handleExistingRepo(ctx, targetPath, repo)
    }
    
    // 새로운 클론
    return e.cloneNewRepo(ctx, targetPath, repo)
}

func (e *CloneExecutor) cloneNewRepo(ctx context.Context, path string, repo provider.Repository) error {
    // 부모 디렉토리 생성
    if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
        return err
    }
    
    // git clone 실행
    cloneURL := e.getCloneURL(repo)
    cmd := exec.CommandContext(ctx, "git", "clone", cloneURL, path)
    
    if output, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("git clone failed: %w\n%s", err, output)
    }
    
    return nil
}

func (e *CloneExecutor) handleExistingRepo(ctx context.Context, path string, repo provider.Repository) error {
    switch e.options.Strategy {
    case StrategyReset:
        return e.resetAndPull(ctx, path)
    case StrategyPull:
        return e.pull(ctx, path)
    case StrategyFetch:
        return e.fetch(ctx, path)
    default:
        return fmt.Errorf("unknown strategy: %s", e.options.Strategy)
    }
}
```

### 5. 세션 관리
```go
// internal/git/clone/session.go
package clone

import (
    "encoding/json"
    "time"
)

type Session struct {
    ID          string                 `json:"id"`
    StartedAt   time.Time              `json:"started_at"`
    Options     CloneOptions           `json:"options"`
    Repositories map[string]RepoStatus `json:"repositories"`
}

type RepoStatus struct {
    Status    string    `json:"status"` // pending, cloning, completed, failed
    Error     string    `json:"error,omitempty"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (s *Session) Save() error {
    data, err := json.MarshalIndent(s, "", "  ")
    if err != nil {
        return err
    }
    
    sessionFile := filepath.Join(getSessionDir(), s.ID+".json")
    return os.WriteFile(sessionFile, data, 0644)
}

func (s *Session) Load(id string) error {
    sessionFile := filepath.Join(getSessionDir(), id+".json")
    data, err := os.ReadFile(sessionFile)
    if err != nil {
        return err
    }
    
    return json.Unmarshal(data, s)
}

func getSessionDir() string {
    // ~/.config/gzh-manager/sessions/
    configDir, _ := os.UserConfigDir()
    return filepath.Join(configDir, "gzh-manager", "sessions")
}
```

### 6. 진행상황 출력
```go
// internal/git/clone/progress.go
package clone

import (
    "fmt"
    "github.com/schollz/progressbar/v3"
)

type ProgressReporter struct {
    bar     *progressbar.ProgressBar
    format  string
    session *Session
}

func NewProgressReporter(total int, format string) *ProgressReporter {
    bar := progressbar.NewOptions(total,
        progressbar.OptionSetDescription("Cloning repositories"),
        progressbar.OptionShowCount(),
        progressbar.OptionShowIts(),
        progressbar.OptionSetPredictTime(true),
    )
    
    return &ProgressReporter{
        bar:    bar,
        format: format,
    }
}

func (p *ProgressReporter) Update(repo string, status RepoStatus) {
    switch p.format {
    case "progress":
        p.bar.Add(1)
    case "json":
        p.printJSON(repo, status)
    case "quiet":
        // No output
    }
}
```

### 7. 기존 synclone 코드 재사용
```go
// internal/git/clone/legacy.go
package clone

import (
    "github.com/gizzahub/gzh-manager-go/cmd/synclone"
    "github.com/gizzahub/gzh-manager-go/pkg/synclone"
)

// 기존 synclone 로직을 래핑하여 재사용
type LegacyAdapter struct {
    config *synclone.Config
}

func (l *LegacyAdapter) CloneWithLegacy(ctx context.Context, opts CloneOptions) error {
    // 옵션을 synclone.Config로 변환
    legacyConfig := l.convertToLegacyConfig(opts)
    
    // 기존 synclone 실행
    return synclone.Execute(ctx, legacyConfig)
}
```

## 구현 체크리스트
- [ ] Clone 명령어 구조 구현
- [ ] CloneOptions 구조체 정의
- [ ] CloneExecutor 구현
- [ ] 병렬 클론 로직 구현
- [ ] 세션 관리 시스템 구현
- [ ] 진행상황 리포터 구현
- [ ] 기존 synclone 코드 어댑터 구현
- [ ] 에러 처리 및 재시도 로직

## 테스트 요구사항
- [ ] 단일 저장소 클론 테스트
- [ ] 병렬 클론 테스트
- [ ] 세션 저장/복원 테스트
- [ ] 각 전략별 동작 테스트
- [ ] 필터링 테스트
- [ ] Dry-run 테스트

## 검증 기준
- [ ] `gz git repo clone --provider github --org myorg` 정상 동작
- [ ] 중단된 작업 재개 가능
- [ ] 병렬 실행으로 성능 향상
- [ ] 기존 synclone과 동일한 결과
- [ ] 진행상황이 실시간으로 표시

## 참고 문서
- cmd/synclone/synclone.go
- pkg/synclone/executor.go
- pkg/github/bulk_clone.go
- Go 동시성 패턴

## 완료 후 다음 단계
→ 08-git-repo-lifecycle-commands.md