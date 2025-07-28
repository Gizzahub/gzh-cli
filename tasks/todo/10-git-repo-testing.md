# Task: Git Repo Testing Strategy

## 작업 목표
`gz git repo` 명령어의 모든 기능에 대한 포괄적인 테스트 전략을 구현합니다.

## 선행 조건
- [ ] 05-git-repo-command-structure.md 완료
- [ ] 06-git-repo-provider-abstraction.md 완료
- [ ] 07-git-repo-clone-implementation.md 완료
- [ ] 08-git-repo-lifecycle-commands.md 완료
- [ ] 09-git-repo-sync-implementation.md 완료

## 구현 상세

### 1. 테스트 구조 설정
```go
// cmd/git/repo_test.go
package git

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/suite"
)

// 테스트 스위트 정의
type GitRepoTestSuite struct {
    suite.Suite
    mockProviders map[string]*MockProvider
    testRepos     []provider.Repository
}

func (s *GitRepoTestSuite) SetupSuite() {
    // Mock providers 초기화
    s.mockProviders = map[string]*MockProvider{
        "github": NewMockGitHubProvider(),
        "gitlab": NewMockGitLabProvider(),
        "gitea":  NewMockGiteaProvider(),
    }
    
    // 테스트 저장소 데이터
    s.testRepos = generateTestRepos()
}

func TestGitRepoSuite(t *testing.T) {
    suite.Run(t, new(GitRepoTestSuite))
}
```

### 2. Provider Mock 구현
```go
// internal/git/provider/mock/provider.go
package mock

import (
    "context"
    "github.com/stretchr/testify/mock"
    "github.com/gizzahub/gzh-manager-go/pkg/git/provider"
)

type MockProvider struct {
    mock.Mock
    repos []provider.Repository
}

func NewMockProvider() *MockProvider {
    return &MockProvider{
        repos: []provider.Repository{},
    }
}

func (m *MockProvider) GetName() string {
    args := m.Called()
    return args.String(0)
}

func (m *MockProvider) ListRepositories(ctx context.Context, opts provider.ListOptions) ([]provider.Repository, error) {
    args := m.Called(ctx, opts)
    return args.Get(0).([]provider.Repository), args.Error(1)
}

func (m *MockProvider) CreateRepository(ctx context.Context, repo provider.CreateRepoRequest) (*provider.Repository, error) {
    args := m.Called(ctx, repo)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*provider.Repository), args.Error(1)
}

// 테스트 헬퍼
func (m *MockProvider) AddTestRepo(repo provider.Repository) {
    m.repos = append(m.repos, repo)
}

func (m *MockProvider) SetupListResponse(org string, repos []provider.Repository) {
    m.On("ListRepositories", mock.Anything, mock.MatchedBy(func(opts provider.ListOptions) bool {
        return opts.Organization == org
    })).Return(repos, nil)
}
```

### 3. Clone 명령어 테스트
```go
// cmd/git/repo_clone_test.go
package git

func (s *GitRepoTestSuite) TestCloneCommand() {
    tests := []struct {
        name      string
        args      []string
        setup     func()
        validate  func(t *testing.T)
        expectErr bool
    }{
        {
            name: "Basic clone from GitHub",
            args: []string{"clone", "--provider", "github", "--org", "testorg"},
            setup: func() {
                s.mockProviders["github"].SetupListResponse("testorg", s.testRepos[:3])
            },
            validate: func(t *testing.T) {
                // 클론된 디렉토리 확인
                assert.DirExists(t, "testorg/repo1")
                assert.DirExists(t, "testorg/repo2")
                assert.DirExists(t, "testorg/repo3")
            },
        },
        {
            name: "Clone with pattern matching",
            args: []string{"clone", "--provider", "github", "--org", "testorg", "--match", "api-.*"},
            setup: func() {
                s.mockProviders["github"].SetupListResponse("testorg", s.testRepos)
            },
            validate: func(t *testing.T) {
                // api-로 시작하는 저장소만 클론됨
                assert.DirExists(t, "testorg/api-service")
                assert.NoDirExists(t, "testorg/web-app")
            },
        },
        {
            name:      "Clone with invalid provider",
            args:      []string{"clone", "--provider", "invalid", "--org", "testorg"},
            expectErr: true,
        },
    }
    
    for _, tt := range tests {
        s.Run(tt.name, func() {
            // Setup
            if tt.setup != nil {
                tt.setup()
            }
            
            // Execute
            cmd := newGitRepoCmd()
            cmd.SetArgs(tt.args)
            err := cmd.Execute()
            
            // Validate
            if tt.expectErr {
                s.Error(err)
            } else {
                s.NoError(err)
                if tt.validate != nil {
                    tt.validate(s.T())
                }
            }
        })
    }
}
```

### 4. 생명주기 명령어 테스트
```go
// cmd/git/repo_lifecycle_test.go
package git

func (s *GitRepoTestSuite) TestCreateCommand() {
    mockProvider := s.mockProviders["github"]
    
    // 성공 케이스 설정
    mockProvider.On("CreateRepository", mock.Anything, mock.MatchedBy(func(req provider.CreateRepoRequest) bool {
        return req.Name == "newrepo" && req.Private == true
    })).Return(&provider.Repository{
        ID:       "123",
        Name:     "newrepo",
        FullName: "testorg/newrepo",
        Private:  true,
    }, nil)
    
    cmd := newGitRepoCmd()
    cmd.SetArgs([]string{
        "create",
        "--provider", "github",
        "--org", "testorg",
        "--name", "newrepo",
        "--private",
        "--description", "Test repository",
    })
    
    err := cmd.Execute()
    s.NoError(err)
    
    // Mock 호출 확인
    mockProvider.AssertExpectations(s.T())
}

func (s *GitRepoTestSuite) TestListCommand() {
    // 다양한 필터링 시나리오 테스트
    testCases := []struct {
        name     string
        args     []string
        expected int
    }{
        {
            name:     "List all repos",
            args:     []string{"list", "--provider", "github", "--org", "testorg"},
            expected: 10,
        },
        {
            name:     "List private repos only",
            args:     []string{"list", "--provider", "github", "--org", "testorg", "--visibility", "private"},
            expected: 3,
        },
        {
            name:     "List with language filter",
            args:     []string{"list", "--provider", "github", "--org", "testorg", "--language", "Go"},
            expected: 5,
        },
    }
    
    for _, tc := range testCases {
        s.Run(tc.name, func() {
            // 테스트별 Mock 설정
            // ... 구현
        })
    }
}
```

### 5. Sync 명령어 테스트
```go
// cmd/git/repo_sync_test.go
package git

func (s *GitRepoTestSuite) TestSyncCommand() {
    srcProvider := s.mockProviders["github"]
    dstProvider := s.mockProviders["gitlab"]
    
    // 소스 저장소 설정
    srcRepo := provider.Repository{
        ID:       "gh-123",
        Name:     "myapp",
        FullName: "myorg/myapp",
    }
    
    srcProvider.On("GetRepository", mock.Anything, "myorg/myapp").Return(&srcRepo, nil)
    
    // 대상 저장소 생성 예상
    dstProvider.On("CreateRepository", mock.Anything, mock.Anything).Return(&provider.Repository{
        ID:       "gl-456",
        Name:     "myapp",
        FullName: "mygroup/myapp",
    }, nil)
    
    // Sync 실행
    cmd := newGitRepoCmd()
    cmd.SetArgs([]string{
        "sync",
        "--from", "github:myorg/myapp",
        "--to", "gitlab:mygroup/myapp",
        "--create-missing",
    })
    
    err := cmd.Execute()
    s.NoError(err)
    
    // 검증
    srcProvider.AssertExpectations(s.T())
    dstProvider.AssertExpectations(s.T())
}
```

### 6. 통합 테스트
```go
// cmd/git/repo_integration_test.go
// +build integration

package git

import (
    "os"
    "testing"
)

func TestGitRepoIntegration(t *testing.T) {
    if os.Getenv("INTEGRATION_TEST") != "true" {
        t.Skip("Skipping integration test")
    }
    
    // 실제 Git 서비스와 통합 테스트
    t.Run("GitHub Integration", func(t *testing.T) {
        if os.Getenv("GITHUB_TOKEN") == "" {
            t.Skip("GITHUB_TOKEN not set")
        }
        
        // 테스트 조직에서 실제 작업 수행
        cmd := newGitRepoCmd()
        cmd.SetArgs([]string{
            "list",
            "--provider", "github",
            "--org", "gizzahub-test",
        })
        
        err := cmd.Execute()
        assert.NoError(t, err)
    })
}
```

### 7. 성능 테스트
```go
// cmd/git/repo_benchmark_test.go
package git

func BenchmarkCloneParallel(b *testing.B) {
    // Mock provider with many repos
    mockProvider := NewMockProvider()
    repos := generateTestRepos(1000) // 1000개 저장소
    mockProvider.SetupListResponse("largecorp", repos)
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        executor := &CloneExecutor{
            provider: mockProvider,
            options: CloneOptions{
                Parallel: 10,
            },
        }
        
        executor.Execute(context.Background())
    }
}

func BenchmarkListWithFilters(b *testing.B) {
    mockProvider := NewMockProvider()
    repos := generateTestRepos(10000) // 10,000개 저장소
    
    b.Run("NoFilter", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            applyFilters(repos, ListOptions{})
        }
    })
    
    b.Run("WithRegexFilter", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            applyFilters(repos, ListOptions{
                Match: "api-.*-service",
            })
        }
    })
}
```

### 8. E2E 테스트 스크립트
```bash
#!/bin/bash
# scripts/test-git-repo-e2e.sh

set -e

echo "Running Git Repo E2E Tests..."

# 환경 준비
export TEST_DIR="/tmp/gzh-git-repo-test"
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

# 1. List 명령어 테스트
echo "Testing list command..."
gz git repo list --provider github --org gizzahub --limit 5

# 2. Clone 테스트
echo "Testing clone command..."
gz git repo clone --provider github --org gizzahub --match "gzh-.*" --dry-run

# 3. Create 테스트 (dry-run)
echo "Testing create command..."
gz git repo create --provider github --org test-org --name test-repo --dry-run

# 4. Sync 테스트 준비
echo "Testing sync command..."
gz git repo sync --from github:gizzahub/test-repo --to gitlab:test-group/test-repo --dry-run

# 5. 도움말 테스트
echo "Testing help..."
gz git repo --help
gz git repo clone --help

echo "✓ All E2E tests passed!"
```

### 9. 테스트 커버리지 설정
```makefile
# Makefile 추가
test-git-repo:
	@echo "Running git repo tests..."
	go test -v ./cmd/git/... -coverprofile=coverage-git-repo.out
	go tool cover -html=coverage-git-repo.out -o coverage-git-repo.html

test-git-repo-integration:
	@echo "Running git repo integration tests..."
	INTEGRATION_TEST=true go test -v ./cmd/git/... -tags=integration

benchmark-git-repo:
	@echo "Running git repo benchmarks..."
	go test -bench=. -benchmem ./cmd/git/...
```

## 구현 체크리스트
- [ ] 테스트 스위트 구조 설정
- [ ] Provider Mock 구현
- [ ] Clone 명령어 단위 테스트
- [ ] 생명주기 명령어 단위 테스트
- [ ] Sync 명령어 단위 테스트
- [ ] 통합 테스트 작성
- [ ] 성능 벤치마크 작성
- [ ] E2E 테스트 스크립트
- [ ] 테스트 커버리지 설정

## 테스트 커버리지 목표
- 단위 테스트: 85% 이상
- 통합 테스트: 주요 워크플로우 100%
- E2E 테스트: 사용자 시나리오 커버

## 검증 기준
- [ ] 모든 명령어에 대한 테스트 존재
- [ ] Mock을 사용한 빠른 단위 테스트
- [ ] 실제 서비스와의 통합 테스트 (선택적)
- [ ] 성능 회귀 방지를 위한 벤치마크
- [ ] CI/CD 파이프라인 통합

## 참고 문서
- testify framework documentation
- Go testing best practices
- Mock 패턴 구현 가이드

## 완료 후 다음 단계
→ Phase 3: Migration and Deprecation 시작