# ISSUE.md - 프로젝트 기능 검토 결과

이 문서는 `FEATURES.md`에 명시된 기능과 실제 구현 간의 불일치 사항을 기록합니다.

---

## 🚨 ISSUE: monitoring 명령어가 구현되지 않음

- 📌 관련 기능: README.md의 Available Commands 섹션
- 📁 관련 파일/모듈: `cmd/root.go`, `cmd/monitoring/` (존재하지 않음)
- 📎 문제 요약: README.md에는 "monitoring - Run monitoring and alerting system" 명령어가 나열되어 있지만, 실제로는 구현되지 않았음
- 🛠️ 제안: TODO 항목으로 전환하여 monitoring 명령어 구현 필요 또는 문서에서 제거

! cli도구로 모니터링 기능 없음 문서에서 제거

---

## 🚨 ISSUE: 병렬 처리가 50개가 아닌 5개로 제한됨

- 📌 관련 기능: 성능 개선 사항 - 병렬 처리
- 📁 관련 파일/모듈: `pkg/github/github_org_clone.go`
- 📎 문제 요약: `FEATURES.md`에는 "최대 50개 리포지토리 동시 클론 지원"이라고 명시되어 있으나, 실제 코드는 `semaphore.NewWeighted(5)`로 5개만 동시 처리
- 🛠️ 제안: DOC_FIX로 문서를 실제 구현(5개)에 맞게 수정하거나, TODO로 전환하여 50개 지원 구현

! 병렬처리5개로 확정

---

## 🚨 ISSUE: WiFi 변경 감지 자동화가 구현되지 않음

- 📌 관련 기능: 네트워크 환경 자동화 - 이벤트 기반 자동화
- 📁 관련 파일/모듈: `cmd/net-env/`, WiFi 모니터링 데몬 없음
- 📎 문제 요약: `FEATURES.md`에는 "WiFi 변경 감지 → 자동 VPN/DNS/프록시 설정 전환"이라고 명시되어 있으나, 실제로는 수동 명령어만 존재
- 🛠️ 제안: TODO 항목으로 전환하여 WiFi 이벤트 모니터링 데몬 구현 필요 또는 문서 수정

! cli도구로 수동으로 전환하는 기능이면 충분. 문서에 반영

---

## 🚨 ISSUE: gzh.yaml 통합 설정 파일 지원이 부분적임

- 📌 관련 기능: 통합 설정 시스템 - gzh.yaml 통합 설정
- 📁 관련 파일/모듈: `pkg/config/`
- 📎 문제 요약: `FEATURES.md`에는 "모든 도구의 설정을 하나의 파일로 통합 관리"라고 되어 있으나, 실제로는 bulk-clone에만 `--use-gzh-config` 옵션이 있음
- 🛠️ 제안: TODO로 모든 명령어에 gzh.yaml 지원 추가 또는 문서에서 "부분 지원"으로 수정

! 설정기반의 cli도구로 다른 도구도 지원가능한 부부은 지원하도록. '부분 지원'으로 수정

---

## 🚨 ISSUE: 설정 마이그레이션 도구가 migrate 명령어와 혼동됨

- 📌 관련 기능: 통합 설정 시스템 - 설정 마이그레이션 도구
- 📁 관련 파일/모듈: `cmd/migrate/`
- 📎 문제 요약: `FEATURES.md`에는 "기존 bulk-clone.yaml을 gzh.yaml로 자동 변환"이라고 되어 있으나, migrate 명령어의 실제 기능 확인 필요
- 🛠️ 제안: migrate 명령어의 실제 기능 검증 후 문서 업데이트 필요

! 제안수용

---

## 🚨 ISSUE: event 명령어가 문서화되지 않음

- 📌 관련 기능: GitHub 이벤트 관리
- 📁 관련 파일/모듈: `cmd/event.go`, README.md
- 📎 문제 요약: root.go에는 event 명령어가 등록되어 있고 README에도 나와 있으나, FEATURES.md에는 이에 대한 설명이 없음
- 🛠️ 제안: DOC_FIX로 FEATURES.md에 event 명령어 기능 추가

---

## 🚨 ISSUE: task-runner와 webhook 명령어가 문서화되지 않음

- 📌 관련 기능: 작업 자동화 및 웹훅 관리
- 📁 관련 파일/모듈: `cmd/task-runner.go`, `cmd/webhook.go`
- 📎 문제 요약: README.md에는 나와 있으나 FEATURES.md에는 이들 명령어에 대한 기능 설명이 없음
- 🛠️ 제안: DOC_FIX로 FEATURES.md에 해당 기능들 추가

! 제안수용. task-runner는 관련코드 문서 모두 제거.

---

## 🚨 ISSUE: 네트워크 환경 "완료" 표시가 오해의 소지가 있음

- 📌 관련 기능: 완료된 네트워크 환경 관리 기능
- 📁 관련 파일/모듈: FEATURES.md 126-131줄
- 📎 문제 요약: "✅ WiFi 이벤트 훅"이 완료됐다고 표시되어 있으나, 실제로는 자동 이벤트 감지가 아닌 수동 명령어만 구현됨
- 🛠️ 제안: DOC_FIX로 "수동 네트워크 액션 시스템"으로 수정 또는 자동화 기능 구현

! 수동으로 구현하고 문서에도 반영

---

## 🚨 ISSUE: GitHub API 클라이언트의 속도 제한이 문서와 다름

- 📌 관련 기능: Repository Management - API integration
- 📁 관련 파일/모듈: `pkg/github/`
- 📎 문제 요약: FEATURES.md에는 "Full GitHub API client with rate limiting"이라고 되어 있으나, 실제 동시 처리는 5개로 제한됨
- 🛠️ 제안: DOC_FIX로 실제 제한 사항 명시

! 제안수용

---

## 🚨 ISSUE: Go SDK 예제 코드 경로가 잘못됨

- 📌 관련 기능: Go SDK Documentation
- 📁 관련 파일/모듈: README.md 676줄
- 📎 문제 요약: `pkg/gzhclient/examples_test.go` 파일을 참조하고 있으나 해당 경로가 존재하는지 확인 필요
- 🛠️ 제안: 실제 예제 파일 경로 확인 후 문서 업데이트

---

## 🚨 ISSUE: 프로젝트 레이아웃 설명이 일관성이 없음

- 📌 관련 기능: Project Layout
- 📁 관련 파일/모듈: README.md 690-696줄
- 📎 문제 요약: internal 패키지 설명 URL이 pkg로 잘못 지정됨 (694줄)
- 🛠️ 제안: DOC_FIX로 URL 수정 필요

! 제안수용

---

## 📊 요약

전반적으로 프로젝트는 견고한 기능을 가지고 있으나, 문서와 실제 구현 간에 여러 불일치가 발견되었습니다:

1. **과장된 기능 설명**: 병렬 처리 50개, WiFi 자동 감지 등
2. **누락된 구현**: monitoring 명령어
3. **불완전한 문서화**: event, webhook, task-runner 명령어
4. **오해의 소지가 있는 표현**: 자동화라고 표현했으나 실제로는 수동 실행

이러한 이슈들은 대부분 문서 수정(DOC_FIX)으로 해결 가능하며, 일부는 TODO로 전환하여 추가 구현이 필요합니다.
