# gzh-cli v0.2.0 출시 노트 초안

**미출시 후보입니다.** 아직 정식 tag·GitHub Release·Homebrew cask를 게시하지 않았습니다.
게시 날짜와 실제 다운로드 링크는 최종 검증·승인 후 기록합니다.

## 사용자 변경점

- `gz profile compare`, `gz profile analyze`, `gz profile continuous --auto-analyze`는 제거됐습니다.
  실제 pprof 분석은 `go tool pprof -top <file>`, 비교는 `go tool pprof -diff_base=<base> <current>`를 사용합니다.
- 그룹 뒤의 알 수 없는 명령과 `gz synclone`의 불필요한 위치 인자를 오류로 처리합니다.
  도움말은 `--help`로 확인합니다.
- 숨김 shell의 profile 기능은 미구현입니다. 시작·중지·상태 조회가 미지원 오류를 표시하며
  프로파일링이 시작됐다는 성공 메시지를 출력하지 않습니다.

`gz doctor metrics`의 threshold는 출력 형식에 관계없이 적용합니다. 명시적 history 요청은
미지원 오류이며, 미평가 history·compliance 상태를 실제 분석 성공으로 표시하지 않습니다.
dev-env v0.2.2의 dry-run journal 보존 수리와 서비스 오류 전달을 포함합니다.

공개 명령 제거가 포함되므로 채택된 첫 공개 후보는 v0.2.0입니다.
구체적인 전체 변경점은 [CHANGELOG](../../CHANGELOG.md)를 따릅니다.

## 배포 범위

첫 공개 채널은 GitHub Release와 Homebrew tap입니다. Docker·Scoop·AUR 배포는 포함하지 않습니다.
artifact 지원 목록과 게시 절차는 [출시 절차](70-releases.md)를 따릅니다.
archive·SBOM·checksum 생성 성공은 실제 플랫폼 동작이나 성능 통과를 대신하지 않습니다.

macOS Cask는 Apple code-signing·notarization이 없는 `gz`의 `com.apple.quarantine`
속성을 staged binary 경로에서 제거합니다. 이는 Cosign artifact 서명과 별개입니다.
Apple 서명·notarization 채택 시 이 호환 후크를 제거해야 하며, 실제 tap·설치 검증은
공개 릴리스 단계에서 수행합니다.

## 공개 전 남은 조건

채택된 core coverage·성능 계약의 실제 결과, 최종 source 리뷰, 실제 서명·게시·설치 검증이 남아 있습니다.
비게시 preview의 snapshot 버전은 정식 v0.2.0이 아닙니다.
공개 후 정식 실행파일의 `gz version` 결과가 `gz version 0.2.0`인지 확인합니다.

사용자 설치본을 바꾸지 않는 격리 환경에서 새 설치를 검증합니다.
이 후보는 첫 공개 출시이므로 이전 공개 버전에서의 upgrade 검증은 해당하지 않습니다.
