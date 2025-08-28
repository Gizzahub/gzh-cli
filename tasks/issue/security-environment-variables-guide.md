# 🔐 보안 환경변수 설정 가이드

**생성일:** 2024-12-19 23:46
**관련 이슈:** [security-audit-report-2024-12-19.md](./security-audit-report-2024-12-19.md)

---

## 📋 개요

보안 감사 결과에 따라 실제 이메일 주소들을 환경변수로 변경하였습니다.
이 문서는 해당 환경변수들의 설정 방법을 안내합니다.

---

## 🌟 필수 환경변수 목록

### 1. GitHub 조직 설정용 환경변수

**파일:** `examples/github/org-settings.yaml`

```bash
# GitHub 조직 관리자 이메일
export GITHUB_ORG_ADMIN_EMAIL="admin@your-domain.com"

# GitHub 조직 결제 담당자 이메일
export GITHUB_ORG_BILLING_EMAIL="billing@your-domain.com"
```

### 2. GoReleaser 설정용 환경변수

**파일:** `.goreleaser.yml`

```bash
# 패키지 메인테이너 이메일
export GORELEASER_MAINTAINER_EMAIL="support@your-domain.com"
```

---

## 🛠️ 설정 방법

### 로컬 개발 환경

1. **환경변수 파일 생성**
   ```bash
   # .env.security 파일 생성 (Git에서 제외됨)
   cat > .env.security << EOF
   GITHUB_ORG_ADMIN_EMAIL="admin@your-domain.com"
   GITHUB_ORG_BILLING_EMAIL="billing@your-domain.com"
   GORELEASER_MAINTAINER_EMAIL="support@your-domain.com"
   EOF
   ```

2. **환경변수 로드**
   ```bash
   # 현재 세션에 로드
   source .env.security

   # 또는 프로필에 추가
   echo "source $(pwd)/.env.security" >> ~/.bashrc
   ```

### CI/CD 환경

**GitHub Actions 예시:**
```yaml
env:
  GITHUB_ORG_ADMIN_EMAIL: ${{ secrets.GITHUB_ORG_ADMIN_EMAIL }}
  GITHUB_ORG_BILLING_EMAIL: ${{ secrets.GITHUB_ORG_BILLING_EMAIL }}
  GORELEASER_MAINTAINER_EMAIL: ${{ secrets.GORELEASER_MAINTAINER_EMAIL }}
```

**GitLab CI 예시:**
```yaml
variables:
  GITHUB_ORG_ADMIN_EMAIL: $GITHUB_ORG_ADMIN_EMAIL
  GITHUB_ORG_BILLING_EMAIL: $GITHUB_ORG_BILLING_EMAIL
  GORELEASER_MAINTAINER_EMAIL: $GORELEASER_MAINTAINER_EMAIL
```

---

## 🔒 보안 모범 사례

### 1. .gitignore 확인
```gitignore
# 환경변수 파일들
.env.security
.env.local
*.env
```

### 2. 권한 설정
```bash
# 환경변수 파일 권한 제한
chmod 600 .env.security
```

### 3. 검증
```bash
# 환경변수 설정 확인
echo "GitHub Admin Email: ${GITHUB_ORG_ADMIN_EMAIL}"
echo "GitHub Billing Email: ${GITHUB_ORG_BILLING_EMAIL}"
echo "GoReleaser Maintainer: ${GORELEASER_MAINTAINER_EMAIL}"
```

---

## 🚨 주의사항

1. **실제 이메일 사용 금지**
   - 테스트나 개발시에는 더미 이메일 사용
   - 예: `test-admin@example.com`

2. **환경변수 누락시 처리**
   - 빌드 스크립트에서 필수 변수 검증 로직 추가
   - 기본값 설정 고려

3. **정기적 검토**
   - 분기별 이메일 주소 유효성 확인
   - 권한 변경시 즉시 업데이트

---

## 📞 문의

환경변수 설정에 문제가 있을 경우 시스템 관리자에게 문의하시기 바랍니다.
