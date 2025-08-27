# AGENTS.md - synclone (대용량 저장소 동기화)

> 📋 **공통 규칙**: [cmd/AGENTS_COMMON.md](../AGENTS_COMMON.md) 참조

## 🎯 모듈 특성

**synclone**은 여러 Git 호스팅 서비스(GitHub, GitLab, Gitea)에서 대용량 저장소를 병렬로 클론/동기화하는 복잡한 모듈입니다.

### 핵심 기능
- 다중 Git 플랫폼 지원 (GitHub, GitLab, Gitea, Gogs)
- 대용량 조직/그룹 저장소 일괄 처리
- 병렬 처리와 메모리 최적화
- 설정 파일 기반 배치 작업
- 재개 가능한 작업과 오류 복구

## ⚠️ 개발 시 핵심 주의사항

### 1. 메모리 관리 (Critical)
```go
// ✅ 스트리밍 처리 사용
err = github.RefreshAllOptimizedStreaming(ctx, targetPath, orgName, strategy, token)

// ❌ 대용량 데이터 한번에 로드 금지
allRepos := loadAllRepositoriesAtOnce() // 메모리 부족 위험
```
- **메모리 제한 옵션** 필수 테스트: `--memory-limit` 플래그
- **스트리밍 API** 사용: 대용량 조직 처리 시
- **점진적 처리**: 저장소 목록을 배치로 나누어 처리

### 2. API 속도 제한 대응
```go
// ✅ 토큰 없이도 동작하도록 설계
if token == "" {
    fmt.Printf("⚠️ Warning: No GitHub token provided. API rate limits may apply.\n")
    // 속도 제한 대응 로직 구현
}
```
- **토큰 없는 환경** 고려: 공개 저장소 클론 지원
- **재시도 로직** 구현: API 속도 제한 시 지수 백오프
- **캐시 활용**: 반복적인 API 호출 최소화

### 3. 네트워크 장애 복구
```go
// ✅ 재개 가능한 작업 설계
err = github.RefreshAllResumable(ctx, targetPath, orgName, strategy, parallel, maxRetries, resume, progressMode)
```
- **상태 저장**: 중단된 작업 재개 가능하도록
- **재시도 정책**: `--max-retries` 옵션 활용
- **부분 실패 처리**: 일부 저장소 실패 시 전체 중단 방지

### 4. 설정 파일 호환성
```yaml
# ✅ 호환성 유지
synclone:
  version: v2  # 버전 명시
  migration_support: true  # 자동 마이그레이션
```
- **버전 호환성**: 기존 설정 파일 자동 마이그레이션
- **검증 로직**: 설정 파일 구문 오류 사전 감지
- **기본값 처리**: 누락된 설정에 대한 안전한 기본값

## 🧪 테스트 요구사항

### 대용량 시나리오 테스트
```bash
# 메모리 제한 테스트
go test ./cmd/synclone -v -run TestMemoryLimit

# 네트워크 장애 시뮬레이션
go test ./cmd/synclone -v -run TestNetworkFailover

# 병렬 처리 안정성
go test ./cmd/synclone -v -run TestParallelStability
```

### 통합 테스트 필수
- **다양한 Git 호스팅 서비스**: GitHub, GitLab, Gitea 모두 테스트
- **대용량 조직**: 100개+ 저장소 처리 성능 검증
- **네트워크 불안정**: 연결 끊김/재연결 시나리오
- **메모리 제약**: 제한된 메모리 환경에서 동작 확인

## 📊 성능 모니터링

### 필수 메트릭
- **메모리 사용량**: RSS, Heap 크기 추적
- **네트워크 대역폭**: 다운로드 속도 모니터링
- **API 호출 횟수**: 속도 제한 근접도 체크
- **병렬 처리 효율**: 워커 풀 활용률

### 최적화 포인트
- **배치 크기 조정**: 메모리와 성능의 균형점 찾기
- **워커 수 최적화**: CPU 코어 수와 네트워크 대역폭 고려
- **캐시 전략**: 메타데이터 캐싱으로 API 호출 최소화

## 🔧 디버깅 팁

```bash
# 상세 로그 활성화
gz synclone --verbose github --org myorg

# 메모리 프로파일링
gz synclone --profile-memory github --org myorg

# 드라이런 모드
gz synclone --dry-run github --org myorg
```

**핵심**: synclone은 대용량 데이터 처리가 핵심이므로, 모든 변경사항은 메모리 사용량과 네트워크 효율성 관점에서 검토해야 합니다.
