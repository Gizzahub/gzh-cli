# TODO: 코드 품질 개선 - Lint 이슈 해결

- status: [ ]
- priority: high (P1)
- category: code-quality
- estimated_effort: 15분
- depends_on: []
- spec_reference: golangci-lint 출력 결과

## 📋 작업 개요

현재 golangci-lint에서 발견된 코드 품질 이슈들을 해결하여 CI/CD 파이프라인을 통과하고 전반적인 코드 품질을 향상시킵니다.

## 🎯 해결해야 할 이슈들

### 1. **높은 복잡도 함수 리팩토링**
- [ ] **파일**: `internal/analysis/godoc/analyzer.go:431`
- [ ] **이슈**: `calculateCoverageStats` 함수의 cognitive complexity가 47 (기준: 30 이하)
- [ ] **해결책**: 함수를 더 작은 단위로 분할하거나 복잡한 로직 단순화

### 2. **반복 문자열 상수화**
- [ ] **파일**: `internal/pm/compat/filters.go:128`
- [ ] **이슈**: `"asdf"` 문자열이 4번 반복 사용됨
- [ ] **해결책**: 상수로 정의하여 재사용

### 3. **if-else 체인을 switch문으로 변경**
- [ ] **파일**: `cmd/doctor/dev_env.go:509`
- [ ] **이슈**: 긴 if-else 체인
- [ ] **해결책**: switch문으로 리팩토링하여 가독성 향상

- [ ] **파일**: `cmd/doctor/godoc.go:103`
- [ ] **이슈**: 긴 if-else 체인
- [ ] **해결책**: switch문으로 리팩토링하여 가독성 향상

## 🔧 구체적인 수정 방법

### 1. 복잡도 높은 함수 개선
```go
// Before: 하나의 큰 함수
func (a *Analyzer) calculateCoverageStats(pkgInfo *PackageInfo) CoverageStats {
    // 47줄의 복잡한 로직...
}

// After: 작은 함수들로 분할
func (a *Analyzer) calculateCoverageStats(pkgInfo *PackageInfo) CoverageStats {
    return CoverageStats{
        Total:       a.calculateTotalCoverage(pkgInfo),
        Statements:  a.calculateStatementCoverage(pkgInfo),
        Functions:   a.calculateFunctionCoverage(pkgInfo),
        Branches:    a.calculateBranchCoverage(pkgInfo),
    }
}

func (a *Analyzer) calculateTotalCoverage(pkgInfo *PackageInfo) float64 {
    // 단순화된 로직
}
// ... 기타 헬퍼 함수들
```

### 2. 문자열 상수화
```go
// Before
return manager == "asdf" && plugin == "rust"

// After
const ManagerAsdf = "asdf"

return manager == ManagerAsdf && plugin == "rust"
```

### 3. Switch문으로 변경
```go
// Before
if err != nil {
    // handle error
} else if condition1 {
    // handle case 1
} else if condition2 {
    // handle case 2
} else {
    // default case
}

// After
switch {
case err != nil:
    // handle error
case condition1:
    // handle case 1
case condition2:
    // handle case 2
default:
    // default case
}
```

## 📁 관련 파일들

### 수정할 파일
- `internal/analysis/godoc/analyzer.go` - 복잡도 개선
- `internal/pm/compat/filters.go` - 상수화
- `cmd/doctor/dev_env.go` - switch문 변경
- `cmd/doctor/godoc.go` - switch문 변경

## 🧪 테스트 방법

### 1. Lint 검사
```bash
make lint
```

### 2. 기능 테스트
```bash
make test
```

### 3. 특정 패키지 테스트
```bash
go test ./internal/analysis/godoc/...
go test ./internal/pm/compat/...
go test ./cmd/doctor/...
```

## ✅ 완료 기준

- [ ] 모든 golangci-lint 이슈 해결
- [ ] 기존 테스트 모두 통과
- [ ] 리팩토링한 함수들이 동일한 결과 반환
- [ ] 코드 가독성 향상 확인

## 🚀 커밋 메시지 가이드

```
refactor(claude-opus): 코드 품질 개선 - lint 이슈 해결

- calculateCoverageStats 함수 복잡도 47→30 이하로 개선
- 반복 문자열 "asdf" 상수화
- if-else 체인을 switch문으로 리팩토링 (dev_env.go, godoc.go)
- 코드 가독성 및 유지보수성 향상

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## 💡 구현 힌트

1. **단계별 접근**: 한 파일씩 수정하고 테스트하여 회귀 방지
2. **함수 분할 시**: 의미 있는 단위로 분할하고 명확한 함수명 사용
3. **상수 정의**: 패키지 레벨이나 적절한 scope에서 정의
4. **Switch문**: 조건이 복잡한 경우 switch true 패턴 활용

## ⚠️ 주의사항

- 기존 로직의 동작 변경 금지 - 순수 리팩토링만 수행
- 테스트 커버리지 유지 또는 개선
- 공개 API 변경 시 호환성 확인
