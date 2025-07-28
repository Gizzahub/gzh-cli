# Task: Git Repo Sync Implementation

## 작업 목표
`gz git repo sync` 명령어를 구현하여 서로 다른 Git 플랫폼 간 저장소 동기화 기능을 제공합니다.

## 선행 조건
- [ ] 05-git-repo-command-structure.md 완료
- [ ] 06-git-repo-provider-abstraction.md 완료
- [ ] Provider 인터페이스의 기본 기능 구현

## 구현 상세

### 1. Sync 명령어 구조
```go
// cmd/git/repo_sync.go
package git

import (
    "context"
    "fmt"
    "github.com/spf13/cobra"
    "github.com/gizzahub/gzh-manager-go/internal/git/sync"
)

func newRepoSyncCmd() *cobra.Command {
    var opts sync.Options
    
    cmd := &cobra.Command{
        Use:   "sync",
        Short: "Synchronize repositories across Git platforms",
        Long: `Synchronize repositories between different Git platforms including:
- Repository code and branches
- Issues and pull requests (if supported)
- Wiki content
- Releases and tags
- Repository settings and metadata`,
        Example: `
  # Sync a single repository
  gz git repo sync --from github:myorg/repo --to gitlab:mygroup/repo
  
  # Sync entire organization
  gz git repo sync --from github:myorg --to gitea:myorg --create-missing
  
  # Sync with specific features
  gz git repo sync --from github:org/repo --to gitlab:group/repo \
    --include-issues --include-wiki --include-releases
  
  # Dry run to preview changes
  gz git repo sync --from github:org/repo --to gitlab:group/repo --dry-run`,
        RunE: func(cmd *cobra.Command, args []string) error {
            return runSync(cmd.Context(), opts)
        },
    }
    
    // 소스와 대상
    cmd.Flags().StringVar(&opts.From, "from", "", "Source (provider:org/repo or provider:org)")
    cmd.Flags().StringVar(&opts.To, "to", "", "Destination (provider:org/repo or provider:org)")
    
    // 동기화 옵션
    cmd.Flags().BoolVar(&opts.CreateMissing, "create-missing", false, "Create repos that don't exist in destination")
    cmd.Flags().BoolVar(&opts.UpdateExisting, "update-existing", true, "Update existing repositories")
    cmd.Flags().BoolVar(&opts.Force, "force", false, "Force push (destructive)")
    
    // 포함할 항목
    cmd.Flags().BoolVar(&opts.IncludeCode, "include-code", true, "Sync repository code")
    cmd.Flags().BoolVar(&opts.IncludeIssues, "include-issues", false, "Sync issues")
    cmd.Flags().BoolVar(&opts.IncludePRs, "include-prs", false, "Sync pull/merge requests")
    cmd.Flags().BoolVar(&opts.IncludeWiki, "include-wiki", false, "Sync wiki")
    cmd.Flags().BoolVar(&opts.IncludeReleases, "include-releases", false, "Sync releases")
    cmd.Flags().BoolVar(&opts.IncludeSettings, "include-settings", false, "Sync repository settings")
    
    // 필터링
    cmd.Flags().StringVar(&opts.Match, "match", "", "Repository name pattern")
    cmd.Flags().StringVar(&opts.Exclude, "exclude", "", "Exclude pattern")
    
    // 실행 옵션
    cmd.Flags().IntVar(&opts.Parallel, "parallel", 1, "Parallel sync workers")
    cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Preview without making changes")
    cmd.Flags().BoolVar(&opts.Verbose, "verbose", false, "Verbose output")
    
    cmd.MarkFlagRequired("from")
    cmd.MarkFlagRequired("to")
    
    return cmd
}
```

### 2. Sync 엔진 구현
```go
// internal/git/sync/engine.go
package sync

import (
    "context"
    "fmt"
    "github.com/gizzahub/gzh-manager-go/pkg/git/provider"
)

type SyncEngine struct {
    source      provider.GitProvider
    destination provider.GitProvider
    options     Options
}

func NewSyncEngine(src, dst provider.GitProvider, opts Options) *SyncEngine {
    return &SyncEngine{
        source:      src,
        destination: dst,
        options:     opts,
    }
}

func (e *SyncEngine) Sync(ctx context.Context) error {
    // 1. 소스 분석
    sourceRepos, err := e.analyzeSource(ctx)
    if err != nil {
        return fmt.Errorf("failed to analyze source: %w", err)
    }
    
    // 2. 대상 분석
    destRepos, err := e.analyzeDestination(ctx)
    if err != nil {
        return fmt.Errorf("failed to analyze destination: %w", err)
    }
    
    // 3. 동기화 계획 생성
    plan := e.createSyncPlan(sourceRepos, destRepos)
    
    // 4. Dry run 처리
    if e.options.DryRun {
        return e.printSyncPlan(plan)
    }
    
    // 5. 동기화 실행
    return e.executeSyncPlan(ctx, plan)
}

type SyncPlan struct {
    Create []RepoSync // 생성할 저장소
    Update []RepoSync // 업데이트할 저장소
    Skip   []RepoSync // 건너뛸 저장소
}

type RepoSync struct {
    Source      provider.Repository
    Destination *provider.Repository // nil if creating new
    Actions     []SyncAction
}

type SyncAction struct {
    Type        string // code, issues, wiki, etc.
    Description string
    Handler     func(context.Context) error
}
```

### 3. Repository 코드 동기화
```go
// internal/git/sync/code.go
package sync

import (
    "os/exec"
    "path/filepath"
)

type CodeSyncer struct {
    source      provider.Repository
    destination provider.Repository
    options     Options
}

func (c *CodeSyncer) Sync(ctx context.Context) error {
    tempDir, err := os.MkdirTemp("", "gzh-sync-*")
    if err != nil {
        return err
    }
    defer os.RemoveAll(tempDir)
    
    // 1. 소스에서 클론
    if err := c.cloneSource(ctx, tempDir); err != nil {
        return fmt.Errorf("failed to clone source: %w", err)
    }
    
    // 2. 대상 remote 추가
    if err := c.addDestinationRemote(ctx, tempDir); err != nil {
        return fmt.Errorf("failed to add destination remote: %w", err)
    }
    
    // 3. 모든 브랜치와 태그 푸시
    if err := c.pushToDestination(ctx, tempDir); err != nil {
        return fmt.Errorf("failed to push to destination: %w", err)
    }
    
    return nil
}

func (c *CodeSyncer) cloneSource(ctx context.Context, dir string) error {
    cmd := exec.CommandContext(ctx, "git", "clone", "--mirror", c.source.CloneURL, dir)
    if output, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("git clone failed: %w\n%s", err, output)
    }
    return nil
}

func (c *CodeSyncer) pushToDestination(ctx context.Context, dir string) error {
    // 모든 브랜치 푸시
    pushArgs := []string{"push", "destination", "--all"}
    if c.options.Force {
        pushArgs = append(pushArgs, "--force")
    }
    
    cmd := exec.CommandContext(ctx, "git", pushArgs...)
    cmd.Dir = dir
    if output, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("git push branches failed: %w\n%s", err, output)
    }
    
    // 모든 태그 푸시
    cmd = exec.CommandContext(ctx, "git", "push", "destination", "--tags")
    cmd.Dir = dir
    if output, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("git push tags failed: %w\n%s", err, output)
    }
    
    return nil
}
```

### 4. Issues 동기화
```go
// internal/git/sync/issues.go
package sync

type IssueSyncer struct {
    source      provider.GitProvider
    destination provider.GitProvider
    mapping     map[string]string // 소스 ID -> 대상 ID 매핑
}

func (i *IssueSyncer) Sync(ctx context.Context, srcRepo, dstRepo provider.Repository) error {
    // Provider capability 확인
    if !i.hasIssueSupport() {
        return fmt.Errorf("one or both providers don't support issues")
    }
    
    // 1. 소스에서 모든 이슈 가져오기
    sourceIssues, err := i.source.ListIssues(ctx, srcRepo.ID)
    if err != nil {
        return err
    }
    
    // 2. 대상의 기존 이슈 확인
    existingIssues, err := i.destination.ListIssues(ctx, dstRepo.ID)
    if err != nil {
        return err
    }
    
    // 3. 이슈 동기화
    for _, issue := range sourceIssues {
        if existing := i.findExistingIssue(issue, existingIssues); existing != nil {
            // 업데이트
            if err := i.updateIssue(ctx, dstRepo.ID, existing, issue); err != nil {
                fmt.Printf("Failed to update issue #%d: %v\n", issue.Number, err)
            }
        } else {
            // 생성
            newIssue, err := i.createIssue(ctx, dstRepo.ID, issue)
            if err != nil {
                fmt.Printf("Failed to create issue #%d: %v\n", issue.Number, err)
            } else {
                i.mapping[issue.ID] = newIssue.ID
            }
        }
    }
    
    // 4. 코멘트 동기화
    return i.syncComments(ctx, srcRepo, dstRepo)
}

func (i *IssueSyncer) createIssue(ctx context.Context, repoID string, issue Issue) (*Issue, error) {
    // 본문에 원본 참조 추가
    body := fmt.Sprintf("%s\n\n---\n_Synced from %s_", issue.Body, issue.URL)
    
    return i.destination.CreateIssue(ctx, repoID, CreateIssueRequest{
        Title:  issue.Title,
        Body:   body,
        Labels: issue.Labels,
        State:  issue.State,
    })
}
```

### 5. Wiki 동기화
```go
// internal/git/sync/wiki.go
package sync

type WikiSyncer struct {
    source      provider.GitProvider
    destination provider.GitProvider
}

func (w *WikiSyncer) Sync(ctx context.Context, srcRepo, dstRepo provider.Repository) error {
    // Wiki는 별도의 Git 저장소로 관리됨
    srcWikiURL := w.getWikiURL(srcRepo)
    dstWikiURL := w.getWikiURL(dstRepo)
    
    tempDir, err := os.MkdirTemp("", "gzh-wiki-sync-*")
    if err != nil {
        return err
    }
    defer os.RemoveAll(tempDir)
    
    // Wiki 저장소 클론
    cmd := exec.CommandContext(ctx, "git", "clone", srcWikiURL, tempDir)
    if err := cmd.Run(); err != nil {
        // Wiki가 없을 수 있음
        return nil
    }
    
    // 대상에 푸시
    cmd = exec.CommandContext(ctx, "git", "remote", "add", "destination", dstWikiURL)
    cmd.Dir = tempDir
    cmd.Run()
    
    cmd = exec.CommandContext(ctx, "git", "push", "destination", "--all", "--force")
    cmd.Dir = tempDir
    return cmd.Run()
}
```

### 6. 동기화 상태 추적
```go
// internal/git/sync/tracker.go
package sync

import (
    "encoding/json"
    "time"
)

type SyncTracker struct {
    ID          string                 `json:"id"`
    StartedAt   time.Time              `json:"started_at"`
    Source      string                 `json:"source"`
    Destination string                 `json:"destination"`
    Status      string                 `json:"status"`
    Progress    map[string]SyncProgress `json:"progress"`
}

type SyncProgress struct {
    Total     int       `json:"total"`
    Completed int       `json:"completed"`
    Failed    int       `json:"failed"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (t *SyncTracker) UpdateProgress(component string, completed, failed int) {
    if t.Progress == nil {
        t.Progress = make(map[string]SyncProgress)
    }
    
    progress := t.Progress[component]
    progress.Completed = completed
    progress.Failed = failed
    progress.UpdatedAt = time.Now()
    t.Progress[component] = progress
    
    t.save()
}

func (t *SyncTracker) save() error {
    data, err := json.MarshalIndent(t, "", "  ")
    if err != nil {
        return err
    }
    
    trackingFile := filepath.Join(getSyncDir(), t.ID+".json")
    return os.WriteFile(trackingFile, data, 0644)
}
```

### 7. 병렬 동기화
```go
// internal/git/sync/parallel.go
package sync

func (e *SyncEngine) executeSyncPlan(ctx context.Context, plan SyncPlan) error {
    if e.options.Parallel <= 1 {
        return e.executeSequential(ctx, plan)
    }
    
    // 병렬 실행을 위한 작업 큐
    tasks := make(chan RepoSync, len(plan.Create)+len(plan.Update))
    errors := make(chan error, len(plan.Create)+len(plan.Update))
    
    // 작업 큐에 추가
    for _, sync := range plan.Create {
        tasks <- sync
    }
    for _, sync := range plan.Update {
        tasks <- sync
    }
    close(tasks)
    
    // Worker 시작
    var wg sync.WaitGroup
    for i := 0; i < e.options.Parallel; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for task := range tasks {
                if err := e.syncRepository(ctx, task); err != nil {
                    errors <- fmt.Errorf("%s: %w", task.Source.FullName, err)
                }
            }
        }()
    }
    
    wg.Wait()
    close(errors)
    
    // 에러 수집
    var allErrors []error
    for err := range errors {
        allErrors = append(allErrors, err)
    }
    
    if len(allErrors) > 0 {
        return fmt.Errorf("sync failed for %d repositories", len(allErrors))
    }
    
    return nil
}
```

## 구현 체크리스트
- [ ] Sync 명령어 구조 구현
- [ ] SyncEngine 핵심 로직 구현
- [ ] Repository 코드 동기화 구현
- [ ] Issues 동기화 구현
- [ ] Wiki 동기화 구현
- [ ] Releases 동기화 구현
- [ ] 동기화 상태 추적 구현
- [ ] 병렬 동기화 구현
- [ ] 충돌 해결 전략 구현

## 테스트 요구사항
- [ ] 단일 저장소 동기화 테스트
- [ ] 조직 전체 동기화 테스트
- [ ] 각 컴포넌트별 동기화 테스트
- [ ] 병렬 실행 테스트
- [ ] 에러 복구 테스트
- [ ] Dry-run 테스트

## 검증 기준
- [ ] GitHub → GitLab 동기화 성공
- [ ] GitLab → Gitea 동기화 성공
- [ ] 코드, 이슈, Wiki가 정확히 동기화됨
- [ ] 중복 실행 시 멱등성 보장
- [ ] 대용량 저장소 동기화 가능

## 참고 문서
- Git mirror and push strategies
- Provider API documentation
- Issue migration best practices

## 완료 후 다음 단계
→ 10-git-repo-testing.md