# AGENTS.md - git-sync (저장소 동기화 진입점)

> 📋 **공통 규칙**: [cmd/AGENTS_COMMON.md](../AGENTS_COMMON.md) 참조

## 모듈 역할

`git-sync`는 `gzh-cli-gitforge/pkg/reposync`와 `pkg/reposynccli`가 제공하는 공용
Cobra command tree를 `gz git-sync`로 조립·등록하는 얇은 진입점입니다.

- 동기화 계획과 실행 로직은 `gzh-cli-gitforge`에 둡니다.
- 이 모듈은 registry metadata와 공개 command 이름을 연결합니다. `app.AppContext`는
  registry 경계에서 전달되지만 현재 command factory 구성에는 사용하지 않습니다.
- `git` command dependency와 stable lifecycle을 유지합니다.

## 변경 규칙

- 기능 로직을 이 패키지에 복제하지 말고 gitforge의 공개 API를 확장합니다.
- `Use`, command 이름, flags를 바꿀 때는 기존 `gz git-sync` 호환성을 먼저 검토합니다.
- filesystem/Git 실행을 추가한다면 dry-run, 부분 실패, 재시도와 상태 저장 계약을
  gitforge 구현 및 테스트에서 다룹니다.
- 토큰, remote URL의 credential, 환경 변수 값을 로그에 출력하지 않습니다.

## 검증

```bash
go test ./cmd/git-sync ./cmd/registry
go test -race ./cmd/git-sync ./cmd/registry
```

command tree나 gitforge 연동을 변경했다면 `make check`도 실행합니다.
