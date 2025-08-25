# Local Dependency Management

본 프로젝트는 Dependabot 대신 로컬에서 의존성을 관리할 수 있는 make 명령어들을 제공합니다.

## 🎯 Dependabot 문제점

- **소스 트리 오염**: 자동 PR로 인한 브랜치 혼잡
- **제어 불가**: 업데이트 타이밍과 범위 조절 어려움
- **테스트 부족**: 자동 업데이트 후 충분한 검증 없음
- **충돌 가능성**: 여러 의존성 동시 업데이트로 인한 문제

## 🚀 로컬 관리의 장점

- **제어 가능**: 원하는 시점에 선택적 업데이트
- **안전성**: 단계별 업데이트와 충분한 테스트
- **깔끔함**: PR 없이 깨끗한 커밋 히스토리
- **효율성**: 배치 업데이트로 시간 절약

## 📋 사용법

### 일상적인 의존성 관리

```bash
# 1. 업데이트 필요한 의존성 확인
make deps-check

# 2. 안전한 업데이트 (patch + minor)
make deps-update

# 3. 선택적 업데이트 (인터랙티브)
make deps-interactive
```

### 단계별 업데이트

```bash
# 가장 안전 (patch 버전만)
make deps-update-patch

# 중간 수준 (minor 버전까지)
make deps-update-minor

# 주의 필요 (major 버전, 브레이킹 체인지 가능)
make deps-update-major
```

### 정기 유지보수

```bash
# 주간 유지보수 (자동화 가능)
make deps-weekly

# 월간 유지보수 (신중한 업데이트)
make deps-monthly
```

### 보안 및 감사

```bash
# 보안 취약점 검사
make deps-security

# 종합 의존성 감사
make deps-audit

# 의존성 보고서 생성
make deps-report
```

### 기타 의존성 관리

```bash
# GitHub Actions 업데이트 확인
make deps-update-actions

# Docker 이미지 업데이트 확인
make deps-update-docker

# 특정 모듈이 필요한 이유 확인
make deps-why MOD=github.com/pkg/errors
```

## 🔧 Dependabot 비활성화

### 방법 1: Dependabot 설정 파일 제거

```bash
# Dependabot 완전 비활성화
rm .github/dependabot.yml
```

### 방법 2: Dependabot 설정 무력화

`.github/dependabot.yml` 파일을 다음과 같이 수정:

```yaml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 0 # PR 생성 차단
```

### 방법 3: Repository 설정에서 비활성화

1. GitHub 저장소 → Settings
1. Security & analysis
1. Dependabot alerts → Disable
1. Dependabot security updates → Disable

## 📅 권장 워크플로우

### 개발자 개인 워크플로우

```bash
# 매주 금요일
make deps-weekly

# 매월 첫째 주
make deps-monthly
```

### 팀 워크플로우

```bash
# 릴리스 전 점검
make deps-audit
make deps-security

# 의존성 보고서 생성 (문서화용)
make deps-report
```

### CI/CD 통합

```yaml
# .github/workflows/deps-check.yml
name: Dependency Check
on:
  schedule:
    - cron: "0 9 * * 1" # 매주 월요일 9시
jobs:
  deps-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - run: make deps-security
      - run: make deps-audit
```

## 🎯 모범 사례

### 1. 업데이트 우선순위

1. **보안 패치**: 즉시 적용
1. **Patch 버전**: 주간 업데이트
1. **Minor 버전**: 월간 검토
1. **Major 버전**: 분기별 계획적 업데이트

### 2. 테스트 전략

```bash
# 업데이트 후 반드시 실행
make deps-update-patch
make test              # 단위 테스트
make test-integration  # 통합 테스트
make lint             # 린트 검사
```

### 3. 롤백 준비

```bash
# 업데이트 전 백업
cp go.mod go.mod.backup
cp go.sum go.sum.backup

# 문제 발생 시 롤백
mv go.mod.backup go.mod
mv go.sum.backup go.sum
go mod download
```

## 🛠️ 고급 사용법

### 특정 의존성만 업데이트

```bash
# 특정 패키지만 업데이트
go get github.com/spf13/cobra@latest
make deps-verify

# 특정 그룹 업데이트 (AWS SDK)
go list -m all | grep aws | cut -d' ' -f1 | xargs go get -u
```

### 의존성 분석

```bash
# 의존성 트리 시각화
make deps-graph

# 큰 의존성 식별
go mod graph | grep "$(go list -m)" | wc -l

# 라이선스 확인 (별도 도구 필요)
go-licenses report ./...
```

## 🚨 주의사항

1. **Major 버전 업데이트**: 반드시 CHANGELOG 확인
1. **보안 업데이트**: 우선순위 높게 처리
1. **테스트 커버리지**: 업데이트 후 테스트 필수
1. **성능 영향**: 벤치마크 테스트 권장

## 🔗 관련 파일

- `Makefile.deps.mk`: 의존성 관리 명령어 정의
- `.github/dependabot.yml`: Dependabot 설정 (비활성화 권장)
- `go.mod`, `go.sum`: Go 의존성 정의
