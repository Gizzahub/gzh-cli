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
