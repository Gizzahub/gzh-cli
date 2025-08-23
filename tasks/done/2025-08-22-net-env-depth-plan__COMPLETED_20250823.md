# net-env depth 구조 분리 제안

## 배경
- `cmd/net-env`는 명령 수와 테스트가 많고, 환경 의존성으로 테스트 실패 가능성이 큼
- 유지보수성과 탐색성을 높이기 위해 기능별 서브패키지화를 제안

## 1차 제안(안전)
- 디렉터리 구조 정리(서브패키지화는 2차)
  - actions/
  - cloud/
  - profile_unified/
  - status_unified/
  - switch_unified/
  - vpn_hierarchy/
  - metrics/
  - analysis/
  - tui/
- 루트 커맨드에서 기존 `newXxxCmd()` 호출부는 유지(파일만 분리 시에는 불가 → 2차에서 서브패키지화 필요)

## 2차 제안(서브패키지화)
- 각 기능을 독립 패키지로 분리하고, 루트에서 `xxx.NewCmd()`로 조립
- 공용 타입/헬퍼는 `internal/netenv` 추출

## 리스크/완화
- 테스트 환경 의존으로 CI에서만 재현되는 실패가 가능 → 변경마다 `go build` + 핵심 경로 테스트 스팟체크

## 실행 순서
1) 파일 매핑표 작성(테스트 파일 포함)
2) `internal/netenv` 초안 생성(공용 옵션/로거/출력 헬퍼)
3) 서브패키지 하나씩 분리(actions → cloud → ...), 각 단계마다 빌드
4) 마지막에 테스트 스위트 스팟체크
