# Task: Git Synclone Testing Strategy

## 작업 목표
`git synclone` 명령어의 품질을 보장하기 위한 포괄적인 테스트 전략을 구현합니다.

## 선행 조건
- [ ] 01-git-synclone-command-structure.md 완료
- [ ] 02-git-synclone-provider-integration.md 완료
- [ ] 03-git-synclone-installation.md 완료
- [ ] 테스트 프레임워크 (testify) 이해

## 구현 상세

### 1. 단위 테스트 구조
```go
// cmd/git-synclone/main_test.go
package main

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCommandStructure(t *testing.T) {
    cmd := newRootCmd()
    
    // 명령어 구조 검증
    assert.Equal(t, "git-synclone", cmd.Use)
    assert.NotNil(t, cmd.Commands())
    
    // 서브커맨드 존재 확인
    githubCmd := findSubCommand(cmd, "github")
    require.NotNil(t, githubCmd, "github subcommand should exist")
}

func TestGitHubFlags(t *testing.T) {
    cmd := newGitHubCmd()
    
    // 필수 플래그 확인
    assert.NotNil(t, cmd.Flags().Lookup("org"))
    assert.NotNil(t, cmd.Flags().Lookup("target"))
    assert.NotNil(t, cmd.Flags().Lookup("parallel"))
}
```

### 2. Provider 통합 테스트
```go
// cmd/git-synclone/providers/github_test.go
func TestGitHubProviderIntegration(t *testing.T) {
    if os.Getenv("GITHUB_TOKEN") == "" {
        t.Skip("GITHUB_TOKEN not set")
    }
    
    tests := []struct {
        name     string
        org      string
        expected int
    }{
        {
            name:     "Public repos only",
            org:      "github",
            expected: 10, // 최소 기대 저장소 수
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            repos, err := listRepositories(tt.org)
            require.NoError(t, err)
            assert.GreaterOrEqual(t, len(repos), tt.expected)
        })
    }
}
```

### 3. 설정 파일 테스트
```go
// cmd/git-synclone/config_test.go
func TestConfigLoading(t *testing.T) {
    // 테스트 설정 파일 생성
    configContent := `
version: "1.0.0"
providers:
  github:
    token: "test-token"
    organizations:
      - name: "test-org"
        clone_dir: "/tmp/test"
`
    
    tmpFile := createTempFile(t, configContent)
    defer os.Remove(tmpFile)
    
    config, err := loadConfig(tmpFile)
    require.NoError(t, err)
    assert.Equal(t, "1.0.0", config.Version)
    assert.Len(t, config.Providers.GitHub.Organizations, 1)
}
```

### 4. Git 통합 테스트
```go
// cmd/git-synclone/git_integration_test.go
func TestGitCommandIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // git-synclone이 PATH에 있는지 확인
    gitSynclonePath, err := exec.LookPath("git-synclone")
    require.NoError(t, err, "git-synclone should be in PATH")
    
    // git synclone 명령어 실행
    cmd := exec.Command("git", "synclone", "--help")
    output, err := cmd.CombinedOutput()
    
    require.NoError(t, err)
    assert.Contains(t, string(output), "Enhanced Git cloning")
}
```

### 5. 세션/상태 관리 테스트
```go
// cmd/git-synclone/session_test.go
func TestSessionManagement(t *testing.T) {
    tempDir := t.TempDir()
    
    // 세션 생성
    session := NewSession(tempDir)
    session.AddRepository("org/repo1", StatusPending)
    session.AddRepository("org/repo2", StatusCloned)
    
    // 세션 저장
    err := session.Save()
    require.NoError(t, err)
    
    // 세션 복원
    restored, err := LoadSession(session.ID)
    require.NoError(t, err)
    
    assert.Equal(t, 2, len(restored.Repositories))
    assert.Equal(t, StatusPending, restored.Repositories["org/repo1"])
}
```

### 6. 에러 시나리오 테스트
```go
// cmd/git-synclone/error_test.go
func TestErrorScenarios(t *testing.T) {
    tests := []struct {
        name        string
        args        []string
        expectedErr string
    }{
        {
            name:        "Missing organization",
            args:        []string{"github"},
            expectedErr: "organization name is required",
        },
        {
            name:        "Invalid config file",
            args:        []string{"--config", "nonexistent.yaml"},
            expectedErr: "failed to load config",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := newRootCmd()
            cmd.SetArgs(tt.args)
            err := cmd.Execute()
            require.Error(t, err)
            assert.Contains(t, err.Error(), tt.expectedErr)
        })
    }
}
```

### 7. 성능 테스트
```go
// cmd/git-synclone/performance_test.go
func BenchmarkParallelCloning(b *testing.B) {
    repos := generateMockRepos(100)
    
    b.Run("Sequential", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            cloneSequential(repos)
        }
    })
    
    b.Run("Parallel-5", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            cloneParallel(repos, 5)
        }
    })
}
```

### 8. E2E 테스트 시나리오
```bash
# scripts/e2e-test.sh
#!/bin/bash
set -e

echo "Running E2E tests for git synclone..."

# 1. 설치 테스트
./scripts/install-git-extensions.sh

# 2. 기본 클론 테스트
git synclone github -o gizzahub -t /tmp/test-clone --dry-run

# 3. 설정 파일 테스트
cat > /tmp/test-config.yaml << EOF
version: "1.0.0"
providers:
  github:
    organizations:
      - name: "gizzahub"
        clone_dir: "/tmp/test-config-clone"
EOF

git synclone --config /tmp/test-config.yaml --dry-run

# 4. 도움말 테스트
git synclone --help
git synclone github --help

echo "✓ All E2E tests passed"
```

## 구현 체크리스트
- [ ] 명령어 구조 단위 테스트
- [ ] Provider별 통합 테스트
- [ ] 설정 파일 로딩 테스트
- [ ] Git 명령어 통합 테스트
- [ ] 세션 관리 테스트
- [ ] 에러 시나리오 테스트
- [ ] 성능 벤치마크
- [ ] E2E 테스트 스크립트

## 테스트 커버리지 목표
- 단위 테스트: 80% 이상
- 통합 테스트: 주요 시나리오 100% 커버
- E2E 테스트: 사용자 워크플로우 커버

## 검증 기준
- [ ] `go test ./cmd/git-synclone/...` 통과
- [ ] 테스트 커버리지 80% 이상
- [ ] CI/CD 파이프라인 통합
- [ ] 모든 플랫폼에서 E2E 테스트 통과

## 참고 문서
- testing package documentation
- testify framework guide
- Git command testing best practices

## 완료 후 다음 단계
→ 05-git-repo-command-structure.md (Phase 2 시작)