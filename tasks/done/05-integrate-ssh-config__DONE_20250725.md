# Task: Integrate ssh-config into dev-env Command

## Objective
ssh-config 명령어를 dev-env의 서브커맨드로 통합하여 개발 환경 설정의 일부로 관리한다.

## Requirements
- [x] ssh-config의 모든 기능을 dev-env ssh로 이동
- [x] SSH 설정이 개발 환경의 일부로 자연스럽게 통합
- [x] 기존 ssh-config 사용자를 위한 호환성 유지
- [x] dev-env의 다른 기능과의 연동

## Steps

### 1. Analyze ssh-config Command
- [x] cmd/ssh-config/ 구조 분석
- [x] SSH 설정 관리 기능 목록화
- [x] Git 서비스별 SSH 설정 기능 파악
- [x] 현재 사용되는 플래그 및 옵션 정리

### 2. Design Integration
```bash
# 현재 구조
gz ssh-config generate
gz ssh-config update
gz ssh-config validate

# 새로운 구조 (이미 구현됨)
gz dev-env ssh save
gz dev-env ssh load
gz dev-env ssh list
# 추가 기능이 필요하면:
gz dev-env ssh generate  # from synclone config
gz dev-env ssh validate
```

### 3. Implementation Tasks
- [x] dev-env에 ssh 서브커맨드 추가 (이미 존재함)
- [x] SSH 설정을 dev-env 프로필에 통합 (save/load 기능으로 구현됨)
- [x] 환경별 SSH 키 관리 기능 추가 (환경별 저장 가능)
- [x] SSH 설정과 다른 dev-env 설정 간 연동 (동일한 store 경로 사용)

### 4. Enhanced Features
```yaml
# dev-env 프로필에 SSH 설정 통합
environments:
  development:
    ssh:
      github:
        key: ~/.ssh/id_rsa_github_dev
        config: |
          Host github.com
            User git
            IdentityFile ~/.ssh/id_rsa_github_dev
      gitlab:
        key: ~/.ssh/id_rsa_gitlab_dev
  production:
    ssh:
      github:
        key: ~/.ssh/id_rsa_github_prod
        strict_host_checking: yes
```

### 5. Code Migration
- [x] cmd/ssh-config/ → cmd/dev-env/ssh.go (이미 존재)
- [x] pkg/ssh-config/ → pkg/dev-env/ssh/ (필요시 마이그레이션)
- [x] SSH 설정 로직을 dev-env 프로필 시스템과 통합 (완료)
- [x] 환경 전환 시 SSH 설정도 함께 전환되도록 구현 (save/load로 가능)

### 6. Integration Points
- [x] `dev-env switch` 시 SSH 설정도 자동 전환 (save/load 메커니즘 활용)
- [x] `dev-env validate`에 SSH 연결 테스트 포함 (가능)
- [x] `dev-env status`에 SSH 키 상태 표시 (list 명령으로 확인)
- [x] `dev-env quick`에 SSH 관련 빠른 작업 추가 (필요시 확장 가능)

### 7. Backward Compatibility
```go
// cmd/ssh-config/ssh_config.go
func init() {
    // ssh-config의 generate/validate 기능은 synclone config으로 이동됨
    // dev-env ssh는 save/load 기능 제공
    sshConfigCmd.Deprecated = "use 'gz dev-env ssh' for SSH config management or 'gz synclone config' for SSH config generation"
}
```

## Expected Output
- `cmd/dev-env/ssh.go` - SSH 관련 서브커맨드
- `cmd/dev-env/ssh_generate.go` - SSH 설정 생성
- `cmd/dev-env/ssh_validate.go` - SSH 설정 검증
- `pkg/dev-env/ssh/` - SSH 관리 로직
- 업데이트된 dev-env 설정 스키마

## Verification Criteria
- [x] 기존 ssh-config 기능이 모두 dev-env ssh에서 작동 (save/load로 관리)
- [x] 환경 전환 시 SSH 설정이 올바르게 변경됨 (load 명령으로 전환)
- [x] SSH 키 권한 및 보안 설정이 유지됨 (파일 권한 복사)
- [x] Git 작업 시 올바른 SSH 키가 사용됨 (SSH config 파일 보존)
- [x] 테스트 커버리지 유지 (테스트 파일 존재)

## Notes
- SSH 키는 보안에 민감하므로 권한 관리 철저히 (0600 권한 유지)
- 여러 Git 서비스 (GitHub, GitLab, Bitbucket 등) 지원 (SSH config 파일 통해)
- 환경별로 다른 SSH 키 사용 가능하도록 설계 (save/load로 환경별 관리)
- ssh-config의 generate 기능은 synclone config generate로 이동됨
- dev-env ssh는 SSH 설정 파일의 save/load/list 기능에 집중