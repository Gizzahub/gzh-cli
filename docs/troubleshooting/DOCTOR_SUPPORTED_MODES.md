# Doctor 실행 모드의 지원 범위

## Container

`doctor container`의 기본 단발 진단과 `--output` JSON 저장은 유지한다. `--watch`, `--server`, `--generate-report`는 실행 backend에 연결돼 있지 않아 명시적 미지원 오류를 반환한다. 이 검사는 진단 logger·Docker 확인·진단 수집·파일 저장 전에 수행한다. 여러 미지원 플래그를 함께 켜도 단발 진단으로 대신 성공하지 않는다.

필드·플래그와 기본값은 유지한다. 비활성 모드의 `--watch-interval`, `--server-port` 설정값만으로 단발 진단을 거부하지 않는다. 별도 `--generate-report`와 실제 진단 결과를 저장하는 `--output`은 다른 옵션이다. HTTP 서버나 반복 주기·출력 누적 정책을 새로 구현한 변경은 아니다.

## Benchmark

현재 `doctor benchmark`의 실행 모드는 지원되지 않는다. 기존 package 발견 경로는 요청한 package/filter 대신 고정 sleep 함수를 측정했고, compare 모드에는 유효한 current 데이터 입력이 없었다. profile 옵션도 실제 생성 파일과 광고하는 경로가 연결돼 있지 않았다.

명령과 플래그·공개 보고서 타입은 유지하되 실행은 profiler·Git 조회·디렉터리 생성·보고서/snapshot 저장 전에 오류로 종료한다. 직접 package 실행 helper도 가상 측정 대신 오류를 반환한다. artifact helper의 CPU/memory profile 요청은 파일 작업 전에 거부하며 생성하지 않은 profile 경로를 결과에 넣지 않는다.

실제 함수 측정용 `BenchmarkSuite`, `SnapshotManager`, JSON 저장·기본 비교 계산 API는 유지한다. 이 보존이 고정 p-value·heuristic significance·confidence 표현이나 메모리 제한값의 통계적 정당성을 입증하지는 않는다. CLI의 저장 전용 추세 분석 입력·출력·통계 계약과 실제 package 실행 방식은 별도 결정 대상이다. 다른 명령의 기본 CPU/disk 진단이나 profiling API를 이 benchmark CLI의 가짜 발견 동작과 혼동하지 않는다.

## 검증 범위

신규 회귀는 임시 cwd/HOME, 통제된 진단 fixture와 실행 seam을 사용한다. 미지원 요청의 provider/진단 호출 0회·새 파일 부재·기존 출력 bytes 보존, 단발 진단의 실제 JSON 저장/출력 경로와 저장 데이터의 기본 계산을 확인한다. Cobra 범위는 각 독립 생성자이며 root 초기화나 조상 hook 전체를 검증한 결과는 아니다. 실제 Docker·Go package benchmark·사용자 저장소를 새 회귀에서 실행하지 않는다.
