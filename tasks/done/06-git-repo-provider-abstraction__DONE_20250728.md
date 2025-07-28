# Task: Git Repo Provider Abstraction Layer

## 작업 목표
모든 Git 플랫폼(GitHub, GitLab, Gitea, Gogs)에서 동일하게 작동하는 Provider 추상화 계층을 설계하고 구현합니다.

## 선행 조건
- [ ] 05-git-repo-command-structure.md 완료
- [ ] 기존 pkg/github, pkg/gitlab, pkg/gitea 인터페이스 분석
- [ ] tasks/gz-unified-cli-design.md의 Provider Abstraction 섹션 검토

## 구현 상세

### 1. 핵심 인터페이스 정의
```go
// pkg/git/provider/interface.go
package provider

import (
    "context"
    "time"
)

// GitProvider는 모든 Git 플랫폼이 구현해야 하는 인터페이스
type GitProvider interface {
    // 기본 정보
    GetName() string
    GetCapabilities() []Capability
    Authenticate(ctx context.Context, creds Credentials) error
    
    // Repository 작업
    RepositoryManager
    
    // Webhook 작업
    WebhookManager
    
    // Event 작업
    EventManager
}

// RepositoryManager는 저장소 관련 작업 인터페이스
type RepositoryManager interface {
    ListRepositories(ctx context.Context, opts ListOptions) ([]Repository, error)
    GetRepository(ctx context.Context, id string) (*Repository, error)
    CreateRepository(ctx context.Context, repo CreateRepoRequest) (*Repository, error)
    UpdateRepository(ctx context.Context, id string, updates UpdateRepoRequest) error
    DeleteRepository(ctx context.Context, id string) error
    ArchiveRepository(ctx context.Context, id string) error
    CloneRepository(ctx context.Context, repo Repository, target string) error
    SearchRepositories(ctx context.Context, query SearchQuery) ([]Repository, error)
}
```

### 2. 공통 데이터 타입
```go
// pkg/git/provider/types.go
package provider

// Repository는 플랫폼 독립적인 저장소 정보
type Repository struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    FullName    string                 `json:"full_name"`
    Owner       Owner                  `json:"owner"`
    Description string                 `json:"description"`
    Private     bool                   `json:"private"`
    Archived    bool                   `json:"archived"`
    CloneURL    string                 `json:"clone_url"`
    SSHURL      string                 `json:"ssh_url"`
    DefaultBranch string               `json:"default_branch"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
    
    // Provider별 추가 정보
    ProviderType string                `json:"provider_type"`
    ProviderData map[string]interface{} `json:"provider_data"`
}

// Capability는 Provider가 지원하는 기능
type Capability string

const (
    CapabilityWebhooks      Capability = "webhooks"
    CapabilityEvents        Capability = "events"
    CapabilityIssues        Capability = "issues"
    CapabilityPullRequests  Capability = "pull_requests"
    CapabilityWiki          Capability = "wiki"
    CapabilityProjects      Capability = "projects"
    CapabilityActions       Capability = "actions"
)

// ListOptions는 목록 조회 옵션
type ListOptions struct {
    Organization string
    User         string
    Visibility   string // public, private, all
    Archived     *bool
    Sort         string // created, updated, name
    Direction    string // asc, desc
    PerPage      int
    Page         int
}
```

### 3. GitHub Provider 구현
```go
// pkg/git/provider/github/provider.go
package github

import (
    "github.com/gizzahub/gzh-manager-go/pkg/git/provider"
    "github.com/google/go-github/v50/github"
)

type GitHubProvider struct {
    client *github.Client
    config *Config
}

func NewGitHubProvider(config *Config) (*GitHubProvider, error) {
    // 기존 pkg/github 코드 재사용
    client := createClient(config.Token)
    
    return &GitHubProvider{
        client: client,
        config: config,
    }, nil
}

func (g *GitHubProvider) GetName() string {
    return "github"
}

func (g *GitHubProvider) GetCapabilities() []provider.Capability {
    return []provider.Capability{
        provider.CapabilityWebhooks,
        provider.CapabilityEvents,
        provider.CapabilityIssues,
        provider.CapabilityPullRequests,
        provider.CapabilityProjects,
        provider.CapabilityActions,
    }
}

func (g *GitHubProvider) ListRepositories(ctx context.Context, opts provider.ListOptions) ([]provider.Repository, error) {
    // GitHub API를 provider.Repository로 변환
    var repos []provider.Repository
    
    opt := &github.RepositoryListByOrgOptions{
        Type:        opts.Visibility,
        Sort:        opts.Sort,
        Direction:   opts.Direction,
        ListOptions: github.ListOptions{
            PerPage: opts.PerPage,
            Page:    opts.Page,
        },
    }
    
    githubRepos, _, err := g.client.Repositories.ListByOrg(ctx, opts.Organization, opt)
    if err != nil {
        return nil, err
    }
    
    for _, gr := range githubRepos {
        repos = append(repos, g.convertToProviderRepo(gr))
    }
    
    return repos, nil
}

func (g *GitHubProvider) convertToProviderRepo(gr *github.Repository) provider.Repository {
    return provider.Repository{
        ID:          gr.GetNodeID(),
        Name:        gr.GetName(),
        FullName:    gr.GetFullName(),
        Description: gr.GetDescription(),
        Private:     gr.GetPrivate(),
        Archived:    gr.GetArchived(),
        CloneURL:    gr.GetCloneURL(),
        SSHURL:      gr.GetSSHURL(),
        DefaultBranch: gr.GetDefaultBranch(),
        CreatedAt:   gr.GetCreatedAt().Time,
        UpdatedAt:   gr.GetUpdatedAt().Time,
        ProviderType: "github",
        ProviderData: map[string]interface{}{
            "stars":    gr.GetStargazersCount(),
            "forks":    gr.GetForksCount(),
            "language": gr.GetLanguage(),
        },
    }
}
```

### 4. Provider Factory
```go
// pkg/git/provider/factory.go
package provider

import (
    "fmt"
    "github.com/gizzahub/gzh-manager-go/pkg/git/provider/github"
    "github.com/gizzahub/gzh-manager-go/pkg/git/provider/gitlab"
    "github.com/gizzahub/gzh-manager-go/pkg/git/provider/gitea"
)

type Factory struct {
    configs map[string]interface{}
}

func NewFactory(configs map[string]interface{}) *Factory {
    return &Factory{configs: configs}
}

func (f *Factory) GetProvider(name string) (GitProvider, error) {
    config, ok := f.configs[name]
    if !ok {
        return nil, fmt.Errorf("provider %s not configured", name)
    }
    
    switch name {
    case "github":
        cfg := config.(*github.Config)
        return github.NewGitHubProvider(cfg)
    case "gitlab":
        cfg := config.(*gitlab.Config)
        return gitlab.NewGitLabProvider(cfg)
    case "gitea":
        cfg := config.(*gitea.Config)
        return gitea.NewGiteaProvider(cfg)
    default:
        return nil, fmt.Errorf("unknown provider: %s", name)
    }
}

func (f *Factory) ListProviders() []string {
    providers := make([]string, 0, len(f.configs))
    for name := range f.configs {
        providers = append(providers, name)
    }
    return providers
}
```

### 5. Provider Registry
```go
// pkg/git/provider/registry.go
package provider

import (
    "context"
    "sync"
)

type Registry struct {
    mu        sync.RWMutex
    providers map[string]GitProvider
    factory   *Factory
}

func NewRegistry(factory *Factory) *Registry {
    return &Registry{
        providers: make(map[string]GitProvider),
        factory:   factory,
    }
}

func (r *Registry) Get(name string) (GitProvider, error) {
    r.mu.RLock()
    provider, exists := r.providers[name]
    r.mu.RUnlock()
    
    if exists {
        return provider, nil
    }
    
    // Lazy loading
    r.mu.Lock()
    defer r.mu.Unlock()
    
    provider, err := r.factory.GetProvider(name)
    if err != nil {
        return nil, err
    }
    
    r.providers[name] = provider
    return provider, nil
}

func (r *Registry) ExecuteAcrossProviders(ctx context.Context, fn func(GitProvider) error) error {
    for _, name := range r.factory.ListProviders() {
        provider, err := r.Get(name)
        if err != nil {
            return err
        }
        
        if err := fn(provider); err != nil {
            return fmt.Errorf("provider %s: %w", name, err)
        }
    }
    return nil
}
```

### 6. 에러 처리
```go
// pkg/git/provider/errors.go
package provider

import "fmt"

// ProviderError는 Provider별 에러를 래핑
type ProviderError struct {
    Provider string
    Op       string
    Err      error
}

func (e *ProviderError) Error() string {
    return fmt.Sprintf("%s: %s: %v", e.Provider, e.Op, e.Err)
}

func (e *ProviderError) Unwrap() error {
    return e.Err
}

// 공통 에러 타입
var (
    ErrNotFound          = fmt.Errorf("not found")
    ErrUnauthorized      = fmt.Errorf("unauthorized")
    ErrRateLimitExceeded = fmt.Errorf("rate limit exceeded")
    ErrNotSupported      = fmt.Errorf("operation not supported")
)
```

## 구현 체크리스트
- [ ] Provider 인터페이스 정의
- [ ] 공통 데이터 타입 정의
- [ ] GitHub Provider 구현
- [ ] GitLab Provider 구현
- [ ] Gitea Provider 구현
- [ ] Provider Factory 구현
- [ ] Provider Registry 구현
- [ ] 에러 처리 표준화

## 테스트 요구사항
- [ ] 각 Provider의 인터페이스 구현 테스트
- [ ] Factory 패턴 테스트
- [ ] Registry 동시성 테스트
- [ ] 에러 처리 테스트

## 검증 기준
- [ ] 모든 Provider가 동일한 인터페이스 구현
- [ ] Provider 간 전환이 코드 변경 없이 가능
- [ ] 기존 pkg/github 등의 코드 재사용
- [ ] 새로운 Provider 추가가 용이

## 참고 문서
- pkg/github/client.go
- pkg/gitlab/client.go
- pkg/gitea/client.go
- Go 인터페이스 베스트 프랙티스

## 완료 후 다음 단계
→ 07-git-repo-clone-implementation.md