# Task: Git Repo Lifecycle Commands Implementation

## 작업 목표
`gz git repo`의 생명주기 관리 명령어들(create, delete, archive, list)을 구현합니다.

## 선행 조건
- [ ] 05-git-repo-command-structure.md 완료
- [ ] 06-git-repo-provider-abstraction.md 완료
- [ ] Provider 인터페이스 구현 완료

## 구현 상세

### 1. Create 명령어 구현
```go
// cmd/git/repo_create.go
package git

import (
    "context"
    "fmt"
    "github.com/spf13/cobra"
    "github.com/gizzahub/gzh-manager-go/pkg/git/provider"
)

func newRepoCreateCmd() *cobra.Command {
    var opts CreateOptions
    
    cmd := &cobra.Command{
        Use:   "create",
        Short: "Create a new repository",
        Long: `Create a new repository on the specified Git platform with
customizable settings including visibility, templates, and initialization options.`,
        Example: `
  # Create a public repository
  gz git repo create --provider github --org myorg --name newrepo
  
  # Create from template
  gz git repo create --provider github --template myorg/template-repo --name myapp
  
  # Create with full options
  gz git repo create --provider gitlab --group mygroup --name api \
    --private --description "API service" --auto-init`,
        RunE: func(cmd *cobra.Command, args []string) error {
            return runCreate(cmd.Context(), opts)
        },
    }
    
    // 필수 옵션
    cmd.Flags().StringVar(&opts.Provider, "provider", "", "Git provider")
    cmd.Flags().StringVar(&opts.Org, "org", "", "Organization/Group name")
    cmd.Flags().StringVar(&opts.Name, "name", "", "Repository name")
    
    // 저장소 설정
    cmd.Flags().StringVar(&opts.Description, "description", "", "Repository description")
    cmd.Flags().BoolVar(&opts.Private, "private", false, "Create as private repository")
    cmd.Flags().StringVar(&opts.Template, "template", "", "Template repository (provider:org/repo)")
    
    // 초기화 옵션
    cmd.Flags().BoolVar(&opts.AutoInit, "auto-init", true, "Initialize with README")
    cmd.Flags().StringVar(&opts.GitignoreTemplate, "gitignore", "", "Gitignore template")
    cmd.Flags().StringVar(&opts.License, "license", "", "License template (mit, apache2, etc)")
    cmd.Flags().StringVar(&opts.DefaultBranch, "default-branch", "main", "Default branch name")
    
    // 추가 설정
    cmd.Flags().BoolVar(&opts.Issues, "issues", true, "Enable issues")
    cmd.Flags().BoolVar(&opts.Wiki, "wiki", false, "Enable wiki")
    cmd.Flags().BoolVar(&opts.Projects, "projects", false, "Enable projects")
    
    cmd.MarkFlagRequired("provider")
    cmd.MarkFlagRequired("name")
    
    return cmd
}

type CreateOptions struct {
    Provider          string
    Org               string
    Name              string
    Description       string
    Private           bool
    Template          string
    AutoInit          bool
    GitignoreTemplate string
    License           string
    DefaultBranch     string
    Issues            bool
    Wiki              bool
    Projects          bool
}
```

### 2. Delete 명령어 구현
```go
// cmd/git/repo_delete.go
package git

func newRepoDeleteCmd() *cobra.Command {
    var opts DeleteOptions
    
    cmd := &cobra.Command{
        Use:   "delete",
        Short: "Delete repositories",
        Long: `Delete one or more repositories with safety checks and confirmation.`,
        Example: `
  # Delete a single repository
  gz git repo delete --provider github --repo myorg/oldrepo
  
  # Delete multiple repositories
  gz git repo delete --provider gitlab --repo "group/repo1,group/repo2"
  
  # Delete with pattern matching (requires --force)
  gz git repo delete --provider github --org myorg --match "test-*" --force`,
        RunE: func(cmd *cobra.Command, args []string) error {
            return runDelete(cmd.Context(), opts)
        },
    }
    
    cmd.Flags().StringVar(&opts.Provider, "provider", "", "Git provider")
    cmd.Flags().StringSliceVar(&opts.Repos, "repo", nil, "Repository to delete (provider:org/repo)")
    cmd.Flags().StringVar(&opts.Org, "org", "", "Organization for pattern matching")
    cmd.Flags().StringVar(&opts.Match, "match", "", "Repository name pattern")
    cmd.Flags().BoolVar(&opts.Force, "force", false, "Skip confirmation")
    cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Preview without deleting")
    
    return cmd
}

func runDelete(ctx context.Context, opts DeleteOptions) error {
    // 안전 검사
    if opts.Match != "" && !opts.Force {
        return fmt.Errorf("pattern matching requires --force flag for safety")
    }
    
    p, err := getProvider(opts.Provider)
    if err != nil {
        return err
    }
    
    // 삭제할 저장소 목록 확인
    repos, err := opts.getTargetRepos(ctx, p)
    if err != nil {
        return err
    }
    
    // Dry run
    if opts.DryRun {
        fmt.Printf("Would delete %d repositories:\n", len(repos))
        for _, repo := range repos {
            fmt.Printf("  - %s\n", repo.FullName)
        }
        return nil
    }
    
    // 확인 프롬프트
    if !opts.Force {
        if !confirmDeletion(repos) {
            return fmt.Errorf("deletion cancelled")
        }
    }
    
    // 삭제 실행
    return deleteRepositories(ctx, p, repos)
}
```

### 3. Archive 명령어 구현
```go
// cmd/git/repo_archive.go
package git

func newRepoArchiveCmd() *cobra.Command {
    var opts ArchiveOptions
    
    cmd := &cobra.Command{
        Use:   "archive",
        Short: "Archive repositories",
        Long: `Archive repositories to make them read-only while preserving all data.`,
        Example: `
  # Archive a single repository
  gz git repo archive --provider github --repo myorg/oldproject
  
  # Archive multiple repositories
  gz git repo archive --provider github --org myorg --match "deprecated-*"
  
  # Unarchive a repository
  gz git repo archive --provider github --repo myorg/project --unarchive`,
        RunE: func(cmd *cobra.Command, args []string) error {
            return runArchive(cmd.Context(), opts)
        },
    }
    
    cmd.Flags().StringVar(&opts.Provider, "provider", "", "Git provider")
    cmd.Flags().StringSliceVar(&opts.Repos, "repo", nil, "Repository to archive")
    cmd.Flags().StringVar(&opts.Org, "org", "", "Organization for pattern matching")
    cmd.Flags().StringVar(&opts.Match, "match", "", "Repository name pattern")
    cmd.Flags().BoolVar(&opts.Unarchive, "unarchive", false, "Unarchive instead of archive")
    cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Preview without archiving")
    
    return cmd
}

func runArchive(ctx context.Context, opts ArchiveOptions) error {
    p, err := getProvider(opts.Provider)
    if err != nil {
        return err
    }
    
    repos, err := opts.getTargetRepos(ctx, p)
    if err != nil {
        return err
    }
    
    for _, repo := range repos {
        if opts.Unarchive {
            fmt.Printf("Unarchiving %s...\n", repo.FullName)
            err = p.UnarchiveRepository(ctx, repo.ID)
        } else {
            fmt.Printf("Archiving %s...\n", repo.FullName)
            err = p.ArchiveRepository(ctx, repo.ID)
        }
        
        if err != nil {
            fmt.Printf("  Error: %v\n", err)
        } else {
            fmt.Printf("  ✓ Done\n")
        }
    }
    
    return nil
}
```

### 4. List 명령어 구현
```go
// cmd/git/repo_list.go
package git

func newRepoListCmd() *cobra.Command {
    var opts ListOptions
    
    cmd := &cobra.Command{
        Use:   "list",
        Short: "List repositories",
        Long: `List repositories from one or more Git platforms with filtering and formatting options.`,
        Example: `
  # List all repositories from an organization
  gz git repo list --provider github --org myorg
  
  # List with filters
  gz git repo list --provider gitlab --group mygroup --visibility private
  
  # List from all configured providers
  gz git repo list --all-providers --format json
  
  # List archived repositories
  gz git repo list --provider github --org myorg --archived-only`,
        RunE: func(cmd *cobra.Command, args []string) error {
            return runList(cmd.Context(), opts)
        },
    }
    
    // Provider 옵션
    cmd.Flags().StringVar(&opts.Provider, "provider", "", "Git provider")
    cmd.Flags().BoolVar(&opts.AllProviders, "all-providers", false, "List from all providers")
    cmd.Flags().StringVar(&opts.Org, "org", "", "Organization/Group name")
    
    // 필터링
    cmd.Flags().StringVar(&opts.Visibility, "visibility", "all", "Filter by visibility (public, private, all)")
    cmd.Flags().BoolVar(&opts.ArchivedOnly, "archived-only", false, "Show only archived repos")
    cmd.Flags().BoolVar(&opts.NoArchived, "no-archived", false, "Exclude archived repos")
    cmd.Flags().StringVar(&opts.Match, "match", "", "Name pattern filter")
    cmd.Flags().StringVar(&opts.Language, "language", "", "Filter by primary language")
    
    // 정렬
    cmd.Flags().StringVar(&opts.Sort, "sort", "name", "Sort by (name, created, updated, stars)")
    cmd.Flags().StringVar(&opts.Order, "order", "asc", "Sort order (asc, desc)")
    
    // 출력
    cmd.Flags().StringVar(&opts.Format, "format", "table", "Output format (table, json, yaml, csv)")
    cmd.Flags().IntVar(&opts.Limit, "limit", 0, "Limit number of results")
    
    return cmd
}

func runList(ctx context.Context, opts ListOptions) error {
    var allRepos []provider.Repository
    
    if opts.AllProviders {
        // 모든 provider에서 조회
        providers, err := getAllProviders()
        if err != nil {
            return err
        }
        
        for name, p := range providers {
            repos, err := listFromProvider(ctx, p, opts)
            if err != nil {
                fmt.Printf("Error from %s: %v\n", name, err)
                continue
            }
            allRepos = append(allRepos, repos...)
        }
    } else {
        // 단일 provider에서 조회
        p, err := getProvider(opts.Provider)
        if err != nil {
            return err
        }
        
        repos, err := listFromProvider(ctx, p, opts)
        if err != nil {
            return err
        }
        allRepos = repos
    }
    
    // 필터링 및 정렬
    filtered := applyFilters(allRepos, opts)
    sorted := applySorting(filtered, opts)
    
    // 제한 적용
    if opts.Limit > 0 && len(sorted) > opts.Limit {
        sorted = sorted[:opts.Limit]
    }
    
    // 출력
    return outputRepos(sorted, opts.Format)
}
```

### 5. 출력 포맷터
```go
// internal/git/output/formatter.go
package output

import (
    "encoding/json"
    "fmt"
    "github.com/olekukonko/tablewriter"
    "gopkg.in/yaml.v3"
)

type RepoFormatter struct {
    format string
}

func (f *RepoFormatter) Output(repos []provider.Repository) error {
    switch f.format {
    case "table":
        return f.outputTable(repos)
    case "json":
        return f.outputJSON(repos)
    case "yaml":
        return f.outputYAML(repos)
    case "csv":
        return f.outputCSV(repos)
    default:
        return fmt.Errorf("unknown format: %s", f.format)
    }
}

func (f *RepoFormatter) outputTable(repos []provider.Repository) error {
    table := tablewriter.NewWriter(os.Stdout)
    table.SetHeader([]string{"Provider", "Name", "Visibility", "Language", "Updated"})
    
    for _, repo := range repos {
        visibility := "public"
        if repo.Private {
            visibility = "private"
        }
        
        table.Append([]string{
            repo.ProviderType,
            repo.FullName,
            visibility,
            repo.ProviderData["language"].(string),
            repo.UpdatedAt.Format("2006-01-02"),
        })
    }
    
    table.Render()
    return nil
}
```

### 6. 필터링 및 정렬
```go
// internal/git/filter/filter.go
package filter

func ApplyFilters(repos []provider.Repository, opts ListOptions) []provider.Repository {
    var filtered []provider.Repository
    
    for _, repo := range repos {
        // Visibility 필터
        if opts.Visibility != "all" {
            if opts.Visibility == "private" && !repo.Private {
                continue
            }
            if opts.Visibility == "public" && repo.Private {
                continue
            }
        }
        
        // Archive 필터
        if opts.ArchivedOnly && !repo.Archived {
            continue
        }
        if opts.NoArchived && repo.Archived {
            continue
        }
        
        // 이름 패턴 필터
        if opts.Match != "" {
            matched, _ := regexp.MatchString(opts.Match, repo.Name)
            if !matched {
                continue
            }
        }
        
        // 언어 필터
        if opts.Language != "" {
            lang, _ := repo.ProviderData["language"].(string)
            if !strings.EqualFold(lang, opts.Language) {
                continue
            }
        }
        
        filtered = append(filtered, repo)
    }
    
    return filtered
}
```

## 구현 체크리스트
- [ ] Create 명령어 구현
- [ ] Delete 명령어 구현
- [ ] Archive 명령어 구현
- [ ] List 명령어 구현
- [ ] 출력 포맷터 구현
- [ ] 필터링 로직 구현
- [ ] 정렬 로직 구현
- [ ] 확인 프롬프트 구현

## 테스트 요구사항
- [ ] 각 명령어별 단위 테스트
- [ ] Provider 통합 테스트
- [ ] 필터링 로직 테스트
- [ ] 출력 형식 테스트
- [ ] 에러 처리 테스트

## 검증 기준
- [ ] 모든 생명주기 명령어가 정상 동작
- [ ] 다중 provider 지원
- [ ] 안전한 삭제 프로세스
- [ ] 다양한 출력 형식 지원
- [ ] 직관적인 필터링 옵션

## 참고 문서
- GitHub API v3 documentation
- GitLab API documentation
- Go 테이블 출력 라이브러리

## 완료 후 다음 단계
→ 09-git-repo-sync-implementation.md