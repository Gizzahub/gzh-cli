# 변경 이력

모든 주요 변경사항은 이 파일에 문서화합니다.
이 프로젝트는 [Semantic Versioning](https://semver.org/spec/v2.0.0.html)을 따릅니다.

## [Unreleased]

다음 출시 후보 이후의 변경은 이 절에 기록합니다.

## [0.2.0] - 미출시 후보

아직 공개 태그나 릴리스를 게시하지 않았습니다. 최종 출시 게이트와 승인을 마친 뒤 실제 게시 날짜를 기록합니다.

### ⚠️ 호환성 깨짐

- `gz profile compare`, `gz profile analyze`, `gz profile continuous --auto-analyze`를 제거했다.
  세 기능은 pprof 데이터를 해석하지 않고 파일 이름에 따라 고정된 결과를 출력했다. 프로필 분석은
  `go tool pprof -top <file>`, 비교는 `go tool pprof -diff_base=<base> <current>`를 쓴다.
  공개 명령 제거이므로 다음 릴리스는 minor 버전 이상을 올린다.
- 그룹 명령(`gz profile`, `gz git repo` 등)이 모르는 하위 명령을 받으면 도움말을 찍고 0으로
  끝나지 않고 `unknown command` 오류와 함께 1로 끝난다. 인자 없이 부르면 지금처럼 도움말을 찍는다.
  `gz git help`처럼 그룹 뒤에 붙인 `help`도 모르는 하위 명령이 된다. 도움말은 `--help`로 본다.
  `gz synclone`은 위치 인자를 받지 않는다. 예전에는 `gz synclone githb` 같은 오타가 인자를
  무시한 채 설정 파일 기반 동기화를 실행했다.

### 보안

- accepted-risk 검사가 고정된 gosec 버전과 같은 규칙으로 blanket `#nosec` 주석을 찾는다.
  `make security-gosec-parity`가 CI에서 그 일치를 측정한다.

### 수정됨

- `gz git repo sync`는 일반 push 실패 후 강제로 재시도하지 않습니다. 강제 덮어쓰기는
  명시적 `--force` 옵션에만 따르며, 출력이 비어 있는 태그 push 실패도 원인을 보존해 반환합니다.

- `gz doctor metrics`는 table·JSON·YAML·issues 출력 뒤 같은 threshold를 판정합니다.
  기본75 미만은 실패하며 정확히75는 통과합니다. 저장·출력 오류를 먼저 반환합니다.

- 명시적 `--historical` 요청은 기본값7을 지정한 경우도 미지원 오류입니다. 기본 실행은
  유지하며 history·compliance 결과는 미평가 상태를 별도 표시합니다.

- dev-env v0.2.2를 사용해 dry-run에서 기존 복구 journal을 보존합니다.
  switch-all 실패는 전환 오류와 rollback 오류를 함께 표시합니다.

- Windows checkout의 core coverage 입력은 LF로 고정하고, Azure CLI 부재 시험은
  사용자 환경의 실제 `az`를 실행하지 않습니다.

- 숨김 shell의 `profile start`, `profile stop`, `profile status`와 인자 없는 `profile`은
  프로파일링을 수행하지 않으므로 명시적 미지원 오류를 반환합니다. 성공 메시지를 출력하지 않으며
  shell은 다음 명령을 계속 받을 수 있습니다.

### 배포 준비

- 첫 배포 채널은 GitHub Release와 Homebrew tap입니다.
- 비게시 preview와 태그의 서명 빌드를 분리했습니다. 검증된 기존 산출물만 게시하며 기존 태그나 릴리스를 교체하지 않습니다.
- Docker, Scoop, AUR은 첫 배포에 포함하지 않습니다.

### 검증 상태

- 비게시 preview와 소스 CI의 성공은 core coverage·실제 CLI 성능·서명·공개 설치 검증을 대신하지 않습니다.
- 핵심 로직 coverage 80%와 기동 \<50ms·대다수 명령 응답 \<100ms의 판정 계약 및 실제 결과를 최종 후보에서 확인해야 합니다.
