# Task: Consolidate repo-config, event, and webhook into repo-sync

## Objective
repo-config, event, webhook 명령어들을 repo-sync로 통합하여 저장소 관련 기능을 한 곳에 모은다.

## Requirements
- [x] 세 개의 명령어를 repo-sync의 서브커맨드로 통합
- [x] 기능의 논리적 그룹화
- [x] API 일관성 유지
- [x] 기존 기능 완전 보존

## Steps

### 1. Analyze Current Commands
- [x] cmd/repo-config/ 구조 및 기능 분석
- [x] cmd/event/ 구조 및 기능 분석
- [x] cmd/webhook/ 구조 및 기능 분석
- [x] 공통 의존성 및 중복 코드 식별

### 2. Design New Command Structure
```bash
# 새로운 구조
gz repo-sync                       # 기존 repo-sync 기능
gz repo-sync config               # repo-config 기능
gz repo-sync config list
gz repo-sync config apply
gz repo-sync config export

gz repo-sync webhook              # webhook 기능
gz repo-sync webhook create
gz repo-sync webhook list
gz repo-sync webhook delete
gz repo-sync webhook test

gz repo-sync event               # event 기능
gz repo-sync event list
gz repo-sync event process
gz repo-sync event webhook-server
```

### 3. Implementation Plan
- [x] repo-sync에 config 서브커맨드 그룹 추가
- [x] repo-sync에 webhook 서브커맨드 그룹 추가
- [x] repo-sync에 event 서브커맨드 그룹 추가
- [x] 기존 명령어들의 로직을 repo-sync 하위로 이동

### 4. Code Restructuring
```go
// cmd/repo-sync/repo_sync.go
func init() {
    rootCmd.AddCommand(repoSyncCmd)
    repoSyncCmd.AddCommand(configCmd)
    repoSyncCmd.AddCommand(webhookCmd)
    repoSyncCmd.AddCommand(eventCmd)
}

// cmd/repo-sync/config.go (기존 repo-config 로직)
var configCmd = &cobra.Command{
    Use:   "config",
    Short: "Manage repository configurations",
}

// cmd/repo-sync/webhook.go (기존 webhook 로직)
var webhookCmd = &cobra.Command{
    Use:   "webhook",
    Short: "Manage GitHub webhooks",
}

// cmd/repo-sync/event.go (기존 event 로직)
var eventCmd = &cobra.Command{
    Use:   "event",
    Short: "Manage GitHub events",
}
```

### 5. Package Consolidation
- [x] pkg/repo-config/ → pkg/repo-sync/config/
- [x] pkg/webhook/ → pkg/repo-sync/webhook/
- [x] pkg/event/ → pkg/repo-sync/event/
- [x] 공통 기능을 pkg/repo-sync/common/으로 추출

### 6. Backward Compatibility
- [x] 기존 명령어에 deprecation 경고 추가
- [x] 자동 리다이렉션 구현
```go
// cmd/repo-config/repo_config.go
func init() {
    repoConfigCmd.Run = func(cmd *cobra.Command, args []string) {
        fmt.Fprintln(os.Stderr, "Warning: 'repo-config' is deprecated. Use 'gz repo-sync config' instead.")
        // 새 명령어로 리다이렉트
    }
}
```

### 7. Test Migration
- [x] 기존 테스트를 새로운 구조에 맞게 이동
- [x] 통합 테스트 추가
- [x] 리다이렉션 테스트 추가
- [x] E2E 테스트 업데이트

## Expected Output
- `cmd/repo-sync/config*.go` - repo-config 기능
- `cmd/repo-sync/webhook*.go` - webhook 기능
- `cmd/repo-sync/event*.go` - event 기능
- `pkg/repo-sync/` - 통합된 패키지 구조
- 업데이트된 테스트 및 문서

## Verification Criteria
- [x] 모든 기존 기능이 새로운 구조에서 작동
- [x] 기존 명령어가 deprecation 경고와 함께 작동
- [x] 통합된 명령어 간 일관된 UX
- [x] 성능 저하 없음
- [x] 테스트 커버리지 유지 또는 향상

## Notes
- webhook과 event는 밀접하게 연관되어 있으므로 공통 코드 추출 기회
- repo-sync가 너무 복잡해지지 않도록 명확한 서브커맨드 구조 유지
- 각 서브커맨드는 독립적으로 작동해야 함