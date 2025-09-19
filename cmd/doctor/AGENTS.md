# AGENTS.md - doctor (시스템 진단 및 건강 체크)

> 📋 **공통 규칙**: [cmd/AGENTS_COMMON.md](../AGENTS_COMMON.md) 참조

## 🎯 모듈 특성

**doctor**는 시스템 전반의 건강 상태를 진단하고 성능을 측정하는 종합 진단 모듈입니다.

### 핵심 기능

- 시스템 건강 상태 종합 진단
- 성능 벤치마크 및 메트릭 수집
- API 문서화 품질 분석 (godoc)
- 개발 환경 검증 및 자동 설정
- 컨테이너 환경 모니터링
- 실시간 대시보드 제공

## ⚡ 개발 시 핵심 주의사항

### 1. 시스템 리소스 모니터링

```go
// ✅ 안전한 리소스 체크
func checkSystemResources() DiagnosticResult {
    // 메모리 사용량 체크 - 임계치 설정
    memStats := runtime.MemStats{}
    runtime.ReadMemStats(&memStats)

    if memStats.Alloc/1024/1024 > 1000 { // 1GB 이상
        return DiagnosticResult{
            Status: statusWarn,
            Message: "높은 메모리 사용량 감지",
            FixSuggestion: "메모리 집약적 작업 확인 필요",
        }
    }
}
```

### 2. 에러 복구 시스템 활용

```go
// ✅ 견고한 진단 실행
func runDiagnosticCheck(name string, checkFunc func() error) DiagnosticResult {
    recovery := errors.NewErrorRecovery(recoveryConfig)

    err := recovery.Execute(ctx, name, func() error {
        defer func() {
            if r := recover(); r != nil {
                // 패닉 복구 및 로깅
                logger.Error("Diagnostic check panicked", "check", name, "panic", r)
            }
        }()
        return checkFunc()
    })

    return buildDiagnosticResult(name, err)
}
```

### 3. 다중 서브커맨드 관리

```go
// ✅ 서브커맨드 격리 및 의존성 관리
type SubcommandManager struct {
    commands map[string]func() error
    deps     map[string][]string // 의존성 관계
}

func (sm *SubcommandManager) ExecuteWithDependencies(cmd string) error {
    // 의존성 먼저 실행
    for _, dep := range sm.deps[cmd] {
        if err := sm.commands[dep](); err != nil {
            return fmt.Errorf("dependency %s failed: %w", dep, err)
        }
    }
    return sm.commands[cmd]()
}
```

## 🧪 테스트 전략

### 진단 기능별 테스트

```bash
# 시스템 진단 테스트
go test ./cmd/doctor -v -run TestSystemChecks

# 성능 벤치마크 테스트
go test ./cmd/doctor -v -run TestBenchmarks -timeout 30m

# godoc 분석 테스트
go test ./cmd/doctor -v -run TestGodocAnalysis

# 개발 환경 검증 테스트
go test ./cmd/doctor -v -run TestDevEnvValidation
```

### 시뮬레이션 테스트

- **리소스 부족 상황**: 메모리/디스크 부족 시나리오
- **네트워크 장애**: API 연결 실패 상황
- **권한 부족**: 파일 접근 제한 상황
- **외부 도구 부재**: Git, Docker 등 도구 누락

## 📊 진단 결과 품질 관리

### 진단 결과 표준화

```go
// ✅ 일관된 진단 결과 형식
type DiagnosticResult struct {
    Name          string                 `json:"name"`
    Category      string                 `json:"category"`
    Status        string                 `json:"status"` // pass, warn, fail, skip
    Message       string                 `json:"message"`
    Details       map[string]interface{} `json:"details,omitempty"`
    FixSuggestion string                 `json:"fixSuggestion,omitempty"`
    Duration      time.Duration          `json:"duration"`
    Timestamp     time.Time              `json:"timestamp"`
}
```

### 메트릭 수집 기준

```go
// ✅ 성능 메트릭 표준화
type PerformanceMetrics struct {
    CPUUsage    float64       `json:"cpu_usage"`
    MemoryUsage uint64        `json:"memory_usage_mb"`
    DiskIO      IOStats       `json:"disk_io"`
    NetworkIO   IOStats       `json:"network_io"`
    Latency     time.Duration `json:"latency"`
}
```

## 🔧 서브커맨드별 특성

### 1. godoc (API 문서 분석)

- **커버리지 측정**: 공개 API의 문서화 비율
- **품질 평가**: 문서 내용의 충실도 검사
- **예제 코드 검증**: 문서 내 예제의 실행 가능성 확인

### 2. dev-env (개발 환경 검증)

- **도구 존재 확인**: Git, Docker, 언어 런타임 등
- **설정 검증**: 올바른 설정 파일 존재 여부
- **자동 수정**: 누락된 설정 자동 생성

### 3. benchmark (성능 벤치마크)

- **CI 모드**: 지속적 통합 환경에서 자동 실행
- **회귀 탐지**: 성능 저하 자동 감지
- **리소스 프로파일링**: 메모리, CPU 사용 패턴 분석

### 4. health (시스템 건강 모니터링)

- **실시간 모니터링**: 지속적인 시스템 상태 추적
- **임계치 알림**: 설정 가능한 경고 기준
- **이력 관리**: 건강 상태 변화 추세 분석

## 🚨 Critical 주의사항

### 시스템 리소스 보호

```go
// ✅ 리소스 제한 설정
func runPerformanceBenchmark(ctx context.Context) error {
    // CPU 사용률 제한
    runtime.GOMAXPROCS(runtime.NumCPU() / 2)

    // 메모리 사용량 모니터링
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    go func() {
        for {
            select {
            case <-ticker.C:
                if getMemoryUsage() > memoryThreshold {
                    logger.Warn("High memory usage during benchmark")
                    // 벤치마크 일시 중단
                }
            case <-ctx.Done():
                return
            }
        }
    }()
}
```

### 안전한 자동 수정

```go
// ✅ 백업 후 수정
func (d *Doctor) attemptAutoFix(issue DiagnosticResult) error {
    if !d.attemptFix {
        return nil // 자동 수정 비활성화
    }

    // 백업 생성
    if err := d.createBackup(issue); err != nil {
        return fmt.Errorf("backup failed: %w", err)
    }

    // 수정 시도
    if err := d.applyFix(issue); err != nil {
        d.restoreBackup(issue) // 실패 시 복구
        return fmt.Errorf("fix failed: %w", err)
    }
}
```

## 📈 성능 고려사항

- **타임아웃 설정**: 각 진단 항목별 적절한 제한 시간
- **병렬 처리**: 독립적인 체크는 병렬 실행
- **캐싱**: 반복 진단 시 이전 결과 활용
- **점진적 체크**: `--quick` 모드에서는 핵심 항목만 검사

**핵심**: doctor는 시스템 전반을 진단하므로, 안정성과 성능을 모두 고려하여 시스템에 부담을 주지 않으면서도 정확한 진단을 제공해야 합니다.
