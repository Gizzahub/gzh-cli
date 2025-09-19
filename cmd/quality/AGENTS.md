# AGENTS.md - quality (코드 품질 관리)

> 📋 **공통 규칙**: [cmd/AGENTS_COMMON.md](../AGENTS_COMMON.md) 참조

## 🎯 모듈 특성

**quality**는 코드 품질 검증, 린팅, 포맷팅을 관리하는 모듈입니다.

### 핵심 기능

- 코드 린팅 (golint, eslint, pylint 등)
- 코드 포맷팅 (gofmt, prettier, black 등)
- 품질 메트릭 수집
- CI/CD 통합

## ⚠️ 개발 시 주의사항

### 1. 다양한 언어 지원

```go
// ✅ 언어별 품질 도구 관리
type QualityTool interface {
    Check(files []string) ([]Issue, error)
    Fix(files []string) error
    Configure(config Config) error
}

// Go 구현
type GoLinter struct{}
func (g *GoLinter) Check(files []string) ([]Issue, error) {
    return g.runGolangCI(files)
}

// JavaScript 구현
type ESLinter struct{}
func (e *ESLinter) Check(files []string) ([]Issue, error) {
    return e.runESLint(files)
}
```

### 2. 설정 파일 관리

```go
// ✅ 품질 도구 설정 통합
func (q *QualityManager) LoadConfigurations() error {
    configs := map[string]string{
        "golangci":  ".golangci.yml",
        "eslint":    ".eslintrc.js",
        "prettier":  ".prettierrc",
        "pytest":    "pytest.ini",
    }

    for tool, configFile := range configs {
        if err := q.validateConfig(tool, configFile); err != nil {
            logger.Warn("Invalid config", "tool", tool, "error", err)
        }
    }
}
```

### 3. 성능 최적화

```go
// ✅ 병렬 품질 검사
func (q *QualityManager) RunChecksParallel(files []string) error {
    var wg sync.WaitGroup
    results := make(chan CheckResult, len(q.tools))

    for _, tool := range q.tools {
        wg.Add(1)
        go func(t QualityTool) {
            defer wg.Done()
            issues, err := t.Check(files)
            results <- CheckResult{Tool: t, Issues: issues, Error: err}
        }(tool)
    }

    wg.Wait()
    close(results)

    return q.aggregateResults(results)
}
```

## 🧪 테스트 요구사항

- **다양한 언어 파일**: Go, JavaScript, Python, YAML 등
- **설정 파일 변형**: 다양한 린터 설정 조합
- **대용량 코드베이스**: 성능 및 메모리 사용량 테스트
- **CI 환경**: 다양한 CI 시스템에서의 동작 검증

**핵심**: 개발자 워크플로우에 통합되므로 빠른 실행 속도와 정확한 결과가 중요합니다.
