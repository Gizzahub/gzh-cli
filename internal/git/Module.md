# internal/git Module

## 목적
- Git 리포지터리 관련 주요 연산(Pull, Push, Fetch, Status 등)을 래핑

## 주요 함수
- CloneRepo()
- PullRepo()
- PushRepo()
- GetStatus()
- ResolveConflict()

## 의존성
- go-git, 또는 exec로 git 바이너리 호출
