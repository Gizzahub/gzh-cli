# 원격 설정 변경의 지원 범위

## GitHub webhook

`gz git webhook repo`의 CRUD는 실제 GitHub API를 호출한다. update에서 `--active`를
생략하면 기존 활성 상태를 유지하며, `--active=false`는 비활성화를 요청한다. `--url`은
실제 webhook config URL에 반영된다. SDK의 update 요청은 `Config.URL`을 우선하며,
그 값이 비어 있으면 기존 최상위 `URL` 입력을 사용한다.

기본 CLI는 기존 Git 명령과 같은 순서로 `GITHUB_TOKEN`, `GH_TOKEN`을 확인해 webhook
서비스에 전달한다. 토큰이 없으면 인증 없이 요청하며 API의 권한 오류를 반환한다. API
토큰과 webhook signature secret은 서로 다른 값이다. 기본 API host는 유지하며 별도의
host 환경변수를 추측하지 않는다.

bulk 작업은 사용자가 지정하거나 주입된 API client가 조회한 실제 저장소만 대상으로 한다.
기본 CLI에서 조직 조회 client가 구성되지 않은 경우 `--repos`를 명시해야 한다. 대상이
없다고 고정된 예제 저장소를 대신 변경하지 않는다. 일부 생성에 실패하면 CLI도 실패를
반환한다. 이미 성공한 원격 변경을 자동으로 취소하는 트랜잭션은 제공하지 않는다.

webhook monitor와 CLI의 조직 policy/config 실행은 현재 미지원 오류를 반환한다. 명령·인자·
플래그는 남아 있지만 고정 성공, 배송 내역, 정책, 점수나 통계를 실제 결과로 출력하지 않는다.
직접 SDK의 monitor 및 migration/sync/compliance/inventory도 `ErrNotImplemented`를
감싼 오류와 nil 결과를 반환한다.

주입한 `ConfigStorage`를 사용하는 SDK policy CRUD와 실제 configuration 검증은 별도
지원 기능으로 유지한다. 운영 backend가 없는 경로를 연결하지 않으며, 모든 config SDK가
미지원이라고 해석하지 않는다.

## 저장소 설정 조회

명시한 config 파일을 읽거나 해석할 수 없거나 filter가 잘못되면 조회를 시작하기 전에
오류를 반환한다. verbose 설정에 따라 잘못된 입력을 무시하거나 필터 없이 조회하지 않는다.
정상 filter와 limit의 동작은 유지한다.

상세 설정을 요청한 경우 조회 실패를 성공 목록에 숨기지 않고 오류를 반환한다. SDK도
branch protection·permissions 조회가 반환한 전송·해석·취소 오류를 전달한다. 기존 HTTP
오류 처리와 기본 브랜치가 없을 때 보호 조회를 생략하는 동작은 유지한다. 실제 통계나
컴플라이언스 점수의 새 의미를 정의한 것은 아니다.

## 위험 평가와 dashboard

`gz repo-config risk-assessment`와 `dashboard` 및 같은 생성자를 사용하는 `gz git config`
경로는 현재 미지원이다. 기존 명령·플래그·공개 DTO는 유지한다. risk-assessment는 고정
저장소·CVSS 점수를 분석 결과로 만들거나 파일에 쓰지 않으며, 직접 분석 helper도 오류를
반환한다. 실제 입력에 적용하는 필터·표시·포맷 유틸은 분석 backend와 구분한다.

dashboard 실행은 listener를 열거나 파일·자격증명에 접근하기 전에 오류를 반환한다.
직접 호출한 home/repository/compliance HTTP helper도 고정 통계 대신 HTTP 501을 반환한다.
이 변경은 웹 기능을 새로 제공하는 것이 아니다. 실제 risk 산정과 컴플라이언스 의미의
계약을 확정하지 않았으며, 공개 제품 정의의 웹/GUI 제외 조건을 유지한다.

## 저장소 설정 apply

`gz repo-config apply`와 같은 명령을 제공하는 `gz git config apply`는 현재 미지원 오류를
반환한다. dry-run, interactive, force도 실제 적용이나 검증된 계획을 제공하지 않는다.
직접 service의 계획·적용 진입점도 같은 한계를 보고한다. 가짜 변경 목록이나 완료 메시지를
적용 증거로 사용하지 않는다.

저장소 설정의 실제 조회와 저수준 mutation API는 별도 기능이다. YAML에서 생략한 필드의
보존, dry-run의 계획 생성, 확인과 부분 실패 처리 계약을 정하기 전에는 전체 설정 DTO를
CLI 입력에 직접 연결하지 않는다. 기존 플래그와 공개 타입은 유지한다.

부작용 없는 실패의 검증 범위는 각 명령 handler와 직접 service 호출이다. 루트·조상 명령의
공통 초기화까지 파일·환경 접근이 없다고 보장하는 것은 아니다.
