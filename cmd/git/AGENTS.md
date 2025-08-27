# AGENTS.md - git (Git 플랫폼 관리)

> 📋 **공통 규칙**: [cmd/AGENTS_COMMON.md](../AGENTS_COMMON.md) 참조

## 🎯 모듈 특성

**git**은 Git 저장소 관리, 설정, 이벤트 처리, 웹훅을 통합하는 Git 플랫폼 관리 모듈입니다.

### 핵심 기능
- Git 저장소 설정 및 관리 (repo 서브커맨드)
- 저장소 이벤트 처리 (event 서브커맨드)
- 웹훅 관리 (webhook 서브커맨드)
- Git 설정 관리 (config 서브커맨드)

## ⚠️ 개발 시 주의사항

### 1. Git 저장소 상태 안전성
```go
// ✅ 안전한 Git 작업
func (r *RepoManager) SafeOperation(repoPath string, operation func() error) error {
    // 작업 전 상태 확인
    if !r.isValidGitRepo(repoPath) {
        return fmt.Errorf("not a valid git repository")
    }

    // dirty state 체크
    if r.hasUncommittedChanges(repoPath) {
        return fmt.Errorf("uncommitted changes detected")
    }

    return operation()
}
```

### 2. 다중 리모트 처리
```go
// ✅ 리모트 저장소 관리
func (r *RepoManager) HandleMultipleRemotes(repoPath string) error {
    remotes, err := r.listRemotes(repoPath)
    if err != nil {
        return err
    }

    for _, remote := range remotes {
        if err := r.validateRemoteAccess(remote); err != nil {
            logger.Warn("Remote access failed", "remote", remote, "error", err)
            continue // 다른 리모트 계속 처리
        }
    }
}
```

### 3. 웹훅 보안
```go
// ✅ 웹훅 서명 검증
func (w *WebhookHandler) ValidateSignature(payload []byte, signature string) error {
    expectedSig := w.calculateHMAC(payload, w.secret)
    if !hmac.Equal([]byte(signature), expectedSig) {
        return fmt.Errorf("invalid webhook signature")
    }
    return nil
}
```

## 🧪 테스트 고려사항

- **Git 상태 시뮬레이션**: clean, dirty, detached HEAD 등 다양한 상태
- **네트워크 장애**: 리모트 저장소 연결 실패 시나리오
- **권한 문제**: 읽기 전용, 쓰기 권한 등 권한별 테스트
- **웹훅 이벤트**: 다양한 Git 이벤트 유형별 처리

**핵심**: Git 작업은 데이터 손실 위험이 있으므로 항상 저장소 상태를 확인하고 안전한 방식으로 작업해야 합니다.
