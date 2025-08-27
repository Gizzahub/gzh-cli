# Net-Env 명세서와 구현 간의 차이점 해결 방안

## 🎯 문제점

`specs/net-env.md` 명세서(1001줄)와 실제 구현 간에 상당한 차이가 발견되었습니다.

### 명세서의 범위 (매우 포괄적)
- **1001줄**의 상세한 기능 명세
- 15개 이상의 고급 서브커맨드
- TUI 대시보드, VPN 계층화, 네트워크 분석, 최적 라우팅 등
- 복잡한 컨테이너/쿠버네티스 네트워크 관리
- 실시간 모니터링 및 분석 도구

### 실제 구현 상태 (부분적)
현재 활성화된 기능:
- ✅ `gz net-env` (TUI)
- ✅ `gz net-env status`
- ✅ `gz net-env profile`
- ✅ `gz net-env actions`
- ✅ `gz net-env cloud`

주석처리된 기능들:
```go
// cmd.AddCommand(switchcmd.NewCmd())                      // 네트워크 프로파일 전환
// cmd.AddCommand(container.NewCmd(logger, configDir))     // Docker + Kubernetes + 컨테이너 감지
// cmd.AddCommand(vpn.NewCmd(logger, configDir))          // VPN 계층화 + 프로파일 + 장애조치
// cmd.AddCommand(analysis.NewCmd(logger, configDir))     // 네트워크 분석 + 토폴로지 + 라우팅
// cmd.AddCommand(metrics.NewCmd(logger, configDir))      // 네트워크 메트릭 + 모니터링
```

## 🤔 검토가 필요한 결정사항

### 옵션 A: 명세서 간소화 (권장)
**장점:**
- 실제 구현과 일치
- 유지보수 부담 감소
- 사용자 혼란 방지
- 핵심 기능에 집중

**단점:**
- 향후 확장성 제한 가능

**작업 범위:**
- 명세서를 현재 구현 기능으로 축소
- 주석처리된 기능들을 "향후 계획" 섹션으로 이동
- 약 600-700줄 정도로 축소 예상

### 옵션 B: 추가 구현 진행
**장점:**
- 완전한 네트워크 관리 솔루션
- 명세서 대로 모든 기능 제공

**단점:**
- 상당한 개발 리소스 필요
- 복잡성 증가
- 테스트 및 유지보수 부담 증가

**작업 범위:**
- 주석처리된 모든 기능 구현
- VPN 관리, 네트워크 분석, 메트릭 모니터링 등
- 대규모 개발 프로젝트

### 옵션 C: 단계별 구현 접근법
**장점:**
- 점진적 기능 확장
- 우선순위에 따른 개발

**단점:**
- 중기적인 불일치 상태 지속

## 📊 현재 기능 분석

### Core 기능 (구현 완료)
- **TUI Dashboard**: 네트워크 상태 시각화
- **Status Command**: 통합 네트워크 상태 확인
- **Profile Management**: 네트워크 프로파일 관리
- **Quick Actions**: 빠른 네트워크 작업
- **Cloud Integration**: 클라우드 제공업체 관리

### Advanced 기능 (주석처리됨)
- **Container Detection**: Docker/Kubernetes 환경 감지
- **VPN Hierarchy**: 계층적 VPN 연결 관리
- **Network Analysis**: 토폴로지 분석 및 최적 라우팅
- **Metrics Monitoring**: 실시간 네트워크 성능 모니터링
- **Network Switching**: 자동 프로파일 전환

## 💡 권장사항

### 1단계: 명세서 정리 (즉시)
- 현재 구현된 기능만 상세히 기술
- 주석처리된 기능들을 "Future Enhancements" 섹션으로 이동
- 명세서 크기를 40-50% 축소

### 2단계: 우선순위 결정 (중기)
향후 구현할 기능의 우선순위를 다음 기준으로 평가:
- **사용자 요구도**
- **구현 복잡도**
- **유지보수 비용**
- **보안 고려사항**

### 3단계: 단계별 로드맵 (장기)
필요시 다음 순서로 기능 추가:
1. **Network Switching** (사용자 요청 높음)
2. **Basic VPN Management** (보안 중요성)
3. **Container Integration** (DevOps 지원)
4. **Advanced Analytics** (고급 사용자)

## 📋 제안된 명세서 구조 (간소화 버전)

```markdown
# Network Environment Management Specification

## Core Commands
- gz net-env            # TUI Dashboard
- gz net-env status     # Network Status
- gz net-env profile    # Profile Management
- gz net-env actions    # Quick Actions
- gz net-env cloud      # Cloud Integration

## Future Enhancements
(현재 주석처리된 고급 기능들)
- Container network management
- Advanced VPN hierarchy
- Network performance analytics
- Automated routing optimization
```

## 🎯 우선순위
**High** - 사용자 혼란 방지 및 명세서 일치성을 위해 조속한 결정 필요

## ✅ 결정이 필요한 항목
- [ ] 명세서 간소화 vs 추가 구현 방향 결정
- [ ] 향후 기능 개발 우선순위 설정
- [ ] 명세서 업데이트 범위 확정
- [ ] 개발팀 리소스 할당 검토
