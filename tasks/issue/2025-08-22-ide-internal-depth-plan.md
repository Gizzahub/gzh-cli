# ide internal 분리 및 depth 구조 제안

## 배경
- `cmd/ide`는 `detector*` 등 무거운 로직과 커맨드 조립이 같은 패키지에 혼재
- 서브패키지화를 하려면 공용 심볼(`IDE`, `NewIDEDetector`, `joinStrings` 등) 의존 해소 필요

## 1차 제안 (internal 추출)
- `internal/idecore` 생성
  - 공용 타입/헬퍼: `IDE`, `IDEDetector`, path/alias 유틸, 포맷터 등 이전
  - `cmd/ide`는 커맨드 조립 및 플래그/IO에 집중

## 2차 제안 (서브패키지화)
- `cmd/ide/open`, `cmd/ide/status`, `cmd/ide/scan` 패키지 생성 후
  - 각 패키지에서 `NewCmd()` 제공
  - 루트 `NewIDECmd`는 하위 `NewCmd()`를 조립

## 리스크/완화
- 대량 이동으로 git blame 영향 → 단계별 커밋, 문서화
- 테스트 부재 영역은 스냅샷 출력 기반 스팟테스트 작성 권장

## 실행 순서
1) `internal/idecore` 초안 작성 및 컴파일 이동(단계별)
2) `open/status/scan`를 서브패키지화, 루트 조립 갱신
3) `go build` 및 주요 경로 스팟테스트
