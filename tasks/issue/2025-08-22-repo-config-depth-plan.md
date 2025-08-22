# repo-config depth 구조 분리 제안

## 배경
- `cmd/repo-config`는 파일/라인 수가 많아 가독성과 유지보수성이 저하됨
- 공통 전역 플래그(`GlobalFlags`)와 내부 함수 공유가 많아 즉시 서브패키지 분리는 리스크 존재

## 제안
- 1차: 단일 패키지 유지, 파일만 기능별 디렉터리로 이동(패키지명 그대로 유지)
  - 디렉터리: audit/, diff/, apply/, list/, validate/, webhook/, dashboard/, risk/, template/
  - 각 파일의 package는 `repoconfig` 유지 → 컴파일 영향 최소화(Go는 동일 패키지를 여러 디렉터리에 둘 수 있음)
  - 루트 `repo_config.go`는 `newXxxCmd()` 조립 유지
- 2차: 공용 타입/헬퍼(`GlobalFlags`, addGlobalFlags, run* helpers)를 `internal/repoconfig`로 추출
  - 의존성 단절 후 필요한 대상만 서브패키지화 검토

## 이점
- 기능별 탐색성 향상, PR/코드리뷰 범위 축소
- 점진적 분리로 리스크 최소화

## 리스크/완화
- 경로 변경으로 일부 도구/스크립트가 경로 의존 시 수정 필요 → 리포 루트 빌드/CI 우선 확인
- 2차 분리 시 import 경로 변경에 따른 대량 변경 가능 → 일괄 리팩터 도구 사용

## 실행 계획(1차)
1) 디렉터리 생성
2) 해당 파일 이동(패키지명 유지)
3) `go build ./...` 확인
4) 필요 시 `go test` 일부 스팟체크
