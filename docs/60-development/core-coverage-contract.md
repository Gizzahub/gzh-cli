# Core coverage 출시 계약

이 도구는 실제 profile을 수집하는 runner가 아니라 소스 선택과 profile 완전성·수치를
검증하는 main 소유 도구다. 전체 `coverage.out`와 기존 test/race gate는 유지한다.

## 범위와 조합

검토된 `scripts/coveragecontract/core-files.txt`는 유지되는 cmd/internal/pkg의 core 파일을
명시한다. `reviewed-exclusions.txt`는 별도 시험 대역·harness·개발 도구의 검토된 경계다.
파일 추가·제거·분류 변경은 이 목록과 소스 검토에 반영해야 한다. 낮은 수치를 이유로
목록에서 빼지 않는다. 알려지지 않은 소스나 변경된 목록은 조용히 제외하지 않는다.

필수 조합은 Darwin arm64·Linux amd64·Windows amd64 각각의
devenv_external/netenv_external/pm_external 8조합, 총24개다.
release profile은 기존 `.goreleaser.yml`과 같은 `CGO_ENABLED=0`을 사용한다.
전체 선택 union에서 core 파일이 빠지면 전체 합격을 거절한다.
배포7플랫폼의 cross-build는 별도이며24개 native profile의 대체 증거가 아니다.

각 조합은 instrumented statement 기준으로 독립적으로80% 이상이어야 한다.
정확히80%는 수용하고 직전 값은 거절한다. 조합별 결과를 pooling하지 않는다.
기존 전체 profile bytes/hash와 전체 분모는 보존하고 core 통계만 별도로 계산한다.

## 계측과 입력 검증

선택기는 Go build 조건을 적용하며 실제 CLI package를 import하지 않는다.
기준 블록은 같은 Go 버전의 `go tool cover`를 소스에 적용하고 생성 AST의
Pos/NumStmt를 읽어 만든다. 제품 binary나 package initializer를 실행하지 않는다.
reference와 sidecar의 해시만 믿지 않고 소스에서 블록을 다시 생성해 대조한다.

공통 manifest는 source/tree·설정·Go 버전·selector digest를 묶어 세 native host에서
동일한 identity를 사용한다. host마다 다른 Go/cover binary SHA와 OS/arch는 각 조합의
기준 블록·binding provenance에 기록한다. native binder는 실제 로컬 바이너리와 기준
생성 도구를 대조하고, 집계기는 두 provenance의 native identity 일치를 요구한다.

계측 블록이 없는 파일은 파일 hash·선택 정보와 함께 무계측 사유를 남긴다.
함수 밖 initializer가 실행되거나 검증됐다는 뜻은 아니다. 보고서 유실을0 statement로
취급하지 않는다. 누락된 조합·profile·계측 블록, 중복·알 수 없는 경로·잘못된 count,
소스·selector·Go 도구·OS/arch/tag/CGO의 불일치를 거절한다.

manifest/기준 블록의 dirty 준비는 진단 자료다. 실제 release profile binding·검증은
clean source를 요구한다. counter mode는 Go 표준 set/count/atomic을 지원하며
측정 profile과 대응 기준의 mode가 같아야 한다.

## 선언된 진입점과 한계

- `make coverage-contract-manifest`: 소스·설정 hash와24개 선택 조합을 준비한다.
- `make coverage-contract-instrument`: source-only 기준 블록과 provenance를 생성한다.
- `make coverage-contract-verify`: 별도 수집·binding된24개 profile을 검증한다.

`Release Core Coverage Qualification` workflow는 수동 비게시 수집 경로다.
선택한 ref와 정확한 `source_sha`가 같아야 하며, 세 native runner에서 tag8조합을
실행한다. 조합별 실제 `go test` 명령·종료 코드·원문 로그와 전체 profile을 보존하고,
모두 성공하면 집계 job이24개 독립 threshold·완전성을 검증한다. 실패한 테스트의
부분 profile을 성공으로 바인딩하지 않는다. 기존 기본/race CI와 배포7플랫폼 빌드는 별도다.

수집 job은 임시 HOME·설정·빈 자격증명 경로를 사용하며 상속한 credential 환경을
테스트 자식에서 제거한다. 이는 hosted runner의 OS/network 격리 증명이 아니고,
실제 CLI 성능 측정용 격리 qualification을 대체하지 않는다.

source-only 준비가 성공해도 실제24개 테스트 실행이나 coverage80% 합격은 아니다.
helper의 `bind`는 제공된 profile hash와 binding 환경을 묶으며, 그 파일을 생성한
`go test` 실행 자체를 관측하지 않는다. 실제 측정 provenance에는 신뢰된 CI 수집의
명령·source·flags·tool identity·exit·원문 로그가 함께 있어야 한다.
로컬 hash/provenance는 악의적인 같은 권한의 host가 꾸민 실행의 전자서명이 아니다.

CLI 성능 측정은 별도의 실행 격리·성공 oracle·측정 계약을 따른다. 이 source-only
검사기가 credential/file/network/process 격리를 제공한다고 주장하지 않는다.
