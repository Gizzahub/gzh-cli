# Dev-Env 아키텍처 방향성 정의: 통합 vs 개별 서비스 접근법

## 🎯 문제점

`specs/dev-env.md` 명세서와 실제 구현이 서로 다른 접근 방식을 취하고 있습니다.

### 명세서 접근법: 통합 환경 관리
```markdown
## 핵심 명령어
- gz dev-env                    # 통합 TUI 대시보드
- gz dev-env switch-all         # 모든 서비스 일괄 전환
- gz dev-env status             # 통합 상태 조회
- gz dev-env validate           # 전체 환경 검증
- gz dev-env sync               # 통합 동기화
- gz dev-env quick              # 사전 정의된 환경 세트로 빠른 전환
```

**특징:**
- **환경 중심** 접근: 개발/스테이징/프로덕션 환경별 관리
- **통합 TUI**: 모든 서비스를 한 화면에서 관리
- **일괄 작업**: 여러 서비스를 동시에 전환
- **환경 프리셋**: 미리 정의된 환경 설정

### 실제 구현: 개별 서비스 관리
```go
// 현재 구현된 명령어들
- gz dev-env gcp-project        # GCP 프로젝트 관리
- gz dev-env aws-profile        # AWS 프로파일 관리
- gz dev-env azure-subscription # Azure 구독 관리
- gz dev-env docker             # Docker 환경 관리
- gz dev-env kubernetes         # Kubernetes 컨텍스트 관리
- gz dev-env ssh                # SSH 설정 관리
- gz dev-env kubeconfig         # Kubeconfig 관리
```

**특징:**
- **서비스 중심** 접근: 각 클라우드/도구별 개별 관리
- **세밀한 제어**: 서비스별 상세 설정 관리
- **독립적 작업**: 각 서비스를 개별적으로 관리
- **전문성**: 각 도구의 고유 기능 완전 활용

## 🤔 각 접근법의 장단점

### 통합 접근법 (명세서)
**장점:**
- 🚀 **사용자 경험**: 한 곳에서 모든 환경 관리
- ⚡ **효율성**: 환경 전체를 한 번에 전환
- 🎯 **일관성**: 통일된 인터페이스
- 🛡️ **안전성**: 환경 간 불일치 방지

**단점:**
- 🔧 **복잡성**: 통합 로직의 구현 복잡도 증가
- 🎛️ **제약**: 서비스별 고유 기능 제한 가능
- 🐛 **장애 확산**: 한 서비스 문제가 전체에 영향
- 📚 **학습 곡선**: 새로운 추상화 레벨 학습 필요

### 개별 서비스 접근법 (현재 구현)
**장점:**
- 🔧 **전문성**: 각 도구의 모든 기능 활용
- 🎛️ **유연성**: 서비스별 세밀한 제어
- 🏗️ **단순성**: 구현 및 유지보수 용이
- 🔍 **디버깅**: 문제 격리 및 해결 용이

**단점:**
- 🔄 **반복 작업**: 환경 전환 시 여러 명령어 실행 필요
- 🤝 **일관성**: 서비스 간 설정 불일치 가능
- 📚 **학습 부담**: 각 서비스별 명령어 학습 필요
- ⏰ **시간 소모**: 환경 전환에 더 많은 시간 필요

## 💡 하이브리드 접근법 제안

### 1단계: 현재 구현 기반 확장
기존 개별 서비스 명령어들을 유지하면서 통합 기능 추가:

```bash
# 개별 서비스 (현재 구현 유지)
gz dev-env gcp-project switch my-prod-project
gz dev-env aws-profile switch production
gz dev-env docker context use prod-context

# 통합 기능 (신규 추가)
gz dev-env switch-env production    # 환경 프리셋으로 일괄 전환
gz dev-env status                   # 모든 서비스 통합 상태
gz dev-env                          # TUI 대시보드 (선택적)
```

### 2단계: 환경 프리셋 시스템
환경별 설정 파일을 통한 일괄 관리:

```yaml
# ~/.gzh/dev-env/environments/production.yaml
name: "Production Environment"
services:
  gcp:
    project: "my-company-prod"
    region: "us-central1"
  aws:
    profile: "production"
    region: "us-west-2"
  kubernetes:
    context: "prod-cluster"
    namespace: "default"
  docker:
    context: "prod-remote"
```

### 3단계: 점진적 통합 기능 추가
사용자 피드백에 따라 추가 기능 개발:
- 환경 간 차이점 비교
- 설정 검증 및 헬스체크
- 환경 백업 및 복원

## 📋 구체적인 제안사항

### 즉시 실행 가능한 작업들

#### A. 명세서 업데이트 (단기)
현재 구현을 반영하여 명세서 수정:
- 개별 서비스 중심 구조로 재작성
- 통합 기능을 "향후 계획"으로 이동
- 실제 사용 가능한 명령어들로 예시 업데이트

#### B. 기본적인 통합 기능 추가 (중기)
최소한의 통합 기능부터 시작:
```go
// 새로 추가할 간단한 명령어들
cmd.AddCommand(status.NewUnifiedStatusCmd())    // 통합 상태 조회
cmd.AddCommand(env.NewEnvironmentCmd())         // 환경 프리셋 관리
```

#### C. 환경 프리셋 시스템 (장기)
사용자 요구가 확인되면 본격적인 환경 관리 시스템 구축

## 🎯 권장 결정사항

### 1차 결정: 방향성 선택
- **옵션 A**: 명세서를 현실에 맞게 수정 (개별 서비스 중심)
- **옵션 B**: 통합 기능 개발 추진 (명세서 기준)
- **옵션 C**: 하이브리드 접근법 채택 (권장)

### 2차 결정: 구현 우선순위
통합 기능 개발 시 우선순위:
1. **통합 상태 조회** (`gz dev-env status`)
2. **환경 프리셋 관리** (`gz dev-env switch-env`)
3. **TUI 대시보드** (`gz dev-env`)
4. **고급 검증 기능** (`gz dev-env validate`)

## 📊 사용자 관점 비교

### 현재 사용 패턴 (개별)
```bash
# 프로덕션 환경으로 전환하려면...
gz dev-env gcp-project switch prod-project
gz dev-env aws-profile switch production
gz dev-env kubernetes context use prod-cluster
gz dev-env docker context use prod-remote
# 총 4개 명령어 실행 필요
```

### 제안된 패턴 (하이브리드)
```bash
# 간단한 경우
gz dev-env switch-env production    # 한 번에 모든 서비스 전환

# 세밀한 제어가 필요한 경우
gz dev-env gcp-project switch specific-project  # 개별 서비스만 변경
```

## ✅ 결정이 필요한 항목
- [ ] 기본 아키텍처 접근법 결정 (통합 vs 개별 vs 하이브리드)
- [ ] 명세서 업데이트 범위 결정
- [ ] 통합 기능 개발 우선순위 설정
- [ ] 환경 프리셋 시스템 필요성 검토
- [ ] 개발 리소스 할당 계획
