# mdformat 패밀리 표준 적용

2026-10-20, gzh-cli-devbox ISSUE-352. 구현·독립 리뷰 요구 strong.
기준 source: `5ea55a3d86ef84e32f9531c4474f14a978ea7347`.

## 변경

패밀리 GUIDELINES §3와 `prmt/ref/makefile-patterns.md`의 채택 표준을 적용한다.
mdformat 1.0.0, mdformat-gfm 1.0.0, mdformat-frontmatter 2.1.2를
`.make/tools.mk`에서 고정한다. gfm이 표를 처리하므로 mdformat\<0.8을 요구하는
mdformat-tables는 제거한다. upstream 버전 근거는
[mdformat PyPI](https://pypi.org/project/mdformat/),
[frontmatter PyPI](https://pypi.org/project/mdformat-frontmatter/)다.

`uv tool install`은 worktree의 `bin/tools` 아래 격리 환경과 launcher를 설치한다.
전역 PATH의 mdformat으로 게이트를 만족시키지 않는다. `--version`은 플러그인 순서에
관계없이 정확한 묶음을 확인하며, `.make/mdformat-constraints.txt`로 렌더링 의존성을
고정·대조한다. 기준 구현은 gzh-cli-gitforge의 동일 설치 계약이다.
Windows launcher/interpreter suffix는 GOEXE로 선택하되 실제 Windows 실행은 별도다.

`format-simplify`도 `install-mdformat`을 선행하고 formatter 실패를 그대로 전파한다.
Go formatter·lint·test의 선택이나 timeout은 변경하지 않는다.

## 좁은 검증

- 실제 설치와 재확인 성공, 새 도구는 `bin/tools` 소유 환경에서 실행.
- README.md, AGENTS.md, 표가 있는 94-gosec-suppression-inventory.md의
  이전/이후 formatter stdout은 byte 동일.
- YAML frontmatter와 표 fixture: 이전 묶음은 머리말을 바꾸지만 새 묶음은 보존.
  포맷 후 재검사 성공, 미정렬 입력의 `--check`는 실패.
- 정확한 버전 두 플러그인 순서 모두 허용. post-release 버전, 추가·누락 플러그인 거부.
- 설치 의존성 대조 성공, 잘못된 mdurl pin 거부.
- 격리된 잘못된 mdformat 버전 설치가 Make 실패로 전파됨.
- `make format-md-check`, `make format-check` 성공.

실행 로그·출력 hash와 재현 진단은 task worktree의 `tmp/issue352`에 보존한다.
임시 진단은 제품 entrypoint나 영구 테스트를 추가하지 않는다.
필수 검사 결과: `make fmt` exit0(67.395초), `make lint` exit0·0 issues(91.448초),
`make test` exit0(51.297초). 시간은 해당 작업의 단일 표본이며 성능 기준이 아니다.
테스트의 기존 SKIP 7건은 그대로 남았고 외부 연동·Docker 수용을 증명하지 않는다.
기존 도구별 timeout은 유지했고 진단 wrapper의 초기 외부 예산만 단계별 1800초로 뒀다.
독립 리뷰·native 통합은 이 기록으로 대체하지 않으며 실제 고정 후보의 결과를 따른다.
primary의 기존 `internal/version/version.go` 수정은 이번 후보에 포함하지 않는다.
