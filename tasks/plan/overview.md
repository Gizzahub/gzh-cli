# 패키지 매니저 고도화 로드맵 개요

## 목적
- 통합 패키지 관리에서 빈발 충돌/예외를 자동 완화하고, 일관된 권장 설정을 기본값으로 제공
- 관찰가능성(로그/JSON), 재현성(구성 파일), 제어성(모드/플래그) 강화

## 현재 상태(요약)
- 필터 체인(compat) 기반 호환성 처리: asdf+rust(rustup), asdf+nodejs(corepack)
- 사용자 규칙 파일(~/.gzh/pm/compat.yml) 로딩
- 모드: `--compat=auto|strict|off`
- asdf 드라이런 상세 출력(Env/후속 액션), 최신 버전 스킵
- 선택 실행: `--managers` CSV

## 다음 작업(파일별 상세 계획 참조)
- 출력 고도화: `--output json` 구조화 출력 [feature-compat-json-output.md]
- 진단 연계: `gz pm doctor --check-conflicts` [feature-compat-doctor.md]
- 규칙 스키마 확장: 조건(when), 환경 매칭(match_env), 레벨(level) [feature-compat-conditions.md]
- 추가 필터: pyenv/pip, golang/GOBIN, npm/corepack 고급 옵션 [feature-pyenv-pip.md, feature-golang-gobin.md, feature-npm-corepack.md]
- 지원 범위 분석: OS별/매니저별 매트릭스 [os-support-analysis.md, manager-support-analysis.md]
- API/스키마 정의: JSON 출력 스키마 [api-schema-json-output.md]

---
- 변경 영향: 런타임 옵션/구성 파일만 추가. 기존 워크플로우는 기본값(auto)로 유지.
- 점진 롤아웃: 내장 필터 → 사용자 정의 필터 → doctor/report 통합.
