# 🗂️ BACKLOG - gzh.yaml 스키마 정의 및 파서 확장 방향

## 📘 목적

- 다양한 Git provider(GitHub, GitLab 등)에 대한 리포지터리 복제 동작을 세분화된 설정으로 제어
- CLI 옵션 없이도 `gzh.yaml` 기반으로 반복 작업 자동화
- 계층 구조(group/subgroup), visibility(public/private), regex 패턴, flatten 여부 등 정의

---

## 📄 gzh.yaml 스키마 예시

```yaml
# 공통 설정
default_provider: github

providers:
  github:
    token: ${GITHUB_TOKEN}
    orgs:
      - name: gizzahub
        visibility: all           # public | private | all
        match: "^gzh-.*"          # 정규식 필터 (optional)
        clone_dir: ./github       # 복제 위치 (optional)

  gitlab:
    token: ${GITLAB_TOKEN}
    groups:
      - name: gizzahub/infra
        visibility: public        # public | private | all
        recursive: true
        flatten: true             # true: 평평한 디렉토리 구조
        match: ".*-manager$"      # 선택적 정규식 필터
        clone_dir: ./gitlab/infra

      - name: gizzahub/labs
        visibility: all
        recursive: false
```

---

## 🧩 파서 확장 설계 방향 (Go 기준)

```go
// config.Config
type Config struct {
	DefaultProvider string              `yaml:"default_provider"`
	Providers       map[string]Provider `yaml:"providers"`
}

// config.Provider
type Provider struct {
	Token  string      `yaml:"token"`
	Orgs   []GitTarget `yaml:"orgs,omitempty"`   // GitHub
	Groups []GitTarget `yaml:"groups,omitempty"` // GitLab
}

// config.GitTarget
type GitTarget struct {
	Name       string `yaml:"name"`
	Visibility string `yaml:"visibility"` // "public", "private", "all"
	Recursive  bool   `yaml:"recursive,omitempty"`
	Flatten    bool   `yaml:"flatten,omitempty"`
	Match      string `yaml:"match,omitempty"`
	CloneDir   string `yaml:"clone_dir,omitempty"`
}
```

---

## ✅ 구현 시 고려사항

- `.yaml` 또는 `.yml` 우선순위 탐색 (`gzh.yaml`, `gzh.yml`)
- `~/.config/gzh.yaml` → 실행 경로 탐색 순서 유지
- `token`은 환경변수 치환(`os.ExpandEnv`) 가능하도록 처리
- `flatten`이 true이면 경로를 `group-subgroup-subgroup...` 형식으로 합성

---

이 스키마는 Claude Code에서 바로 파서 구조, validation, CLI 바인딩 등에 사용할 수 있도록 구성되어 있습니다.

👉 원하시면 이 스키마 기반으로 `config` 모듈 코드, 디렉토리 경로 구성 유틸리티, 에러 메시지 포맷 등도 바로 만들어드릴 수 있습니다. 어떤 방식으로 진행해볼까요?

---

## 📋 GitHub Organization & Repository Management

### 🎯 목적
- GitHub 조직 및 리포지터리의 기본 설정을 일괄 관리
- 리포지터리 정책 및 설정의 표준화
- 새 프로젝트 생성 시 자동화된 설정 적용

### 📋 관리 대상 설정들
- 기본 브랜치 설정
- 머지 정책 (squash, merge commit, rebase)
- 보안 및 분석 설정
- Issues, Projects, Wiki 활성화 여부
- 가시성 설정 (public/private)
- 포킹 및 자동 머지 정책
- 커밋 서명 요구사항

### 🛠️ 구현 방향
- **CLI 명령어**: `gz github-org config` 또는 `gz repo-config`
- **설정 방식**: YAML 기반 정책 파일
- **API 활용**: GitHub REST API `repos.update` 엔드포인트
- **대안 고려**: Terraform 사용 검토 (Infrastructure as Code)

### 📚 참고 자료
상세한 구현 참고 자료는 `docs/github-org-management-research.md` 참조

### ⚠️ 고려사항
- GitHub Actions보다는 CLI 도구로 구현이 더 적합
- 토큰 권한 관리 (repos, admin:org 권한 필요)
- 대량 업데이트 시 API Rate Limiting 고려
