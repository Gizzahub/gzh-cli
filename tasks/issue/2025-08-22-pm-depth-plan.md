# pm depth 구조 제안

## 배경
- `cmd/pm`는 파일별 기능이 명확(install/export/status/update/advanced/cache/doctor)
- 루트 `pm.go`에서 서브커맨드 조립 중. 현재도 관리 용이하나, 탐색성 개선 차원에서 폴더링 가능

## 제안
- 디렉터리 구조: `install/`, `export/`, `status/`, `update/`, `advanced/`, `cache/`, `doctor/`
- 패키지명은 유지(`package pm`), 파일만 분리(1차). 2차에서 서브패키지화 검토

## 리스크/완화
- 파일 이동으로 PR 범위 확대 → 1차는 파일만 분리, 조립 로직 변경 없음

## 실행 순서
1) 디렉터리 생성 및 파일 이동(패키지명 유지)
2) `go build` 확인
3) 필요 시 서브패키지화로 확장
