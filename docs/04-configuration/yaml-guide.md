# gzh.yaml 설정 가이드

<!-- 
통합된 파일 출처:
- yaml-quick-reference.md (빠른 참조)
- yaml-usage-guide.md (상세 사용 가이드)
통합일: 2025-07-16
-->

## 📋 목차
1. [빠른 시작](#빠른-시작)
2. [기본 설정](#기본-설정)
3. [고급 설정](#고급-설정)
4. [예제 모음](#예제-모음)
5. [문제 해결](#문제-해결)

## 🚀 빠른 시작

### 최소 설정
```yaml
# gzh.yaml
version: "1.0"
providers:
  github:
    token: "${GITHUB_TOKEN}"
```

### 기본 설정 템플릿
```yaml
# gzh.yaml
version: "1.0"
metadata:
  name: "my-development-setup"
  description: "개인 개발 환경 설정"

providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations: ["my-org"]
  
  gitlab:
    token: "${GITLAB_TOKEN}"
    groups: ["my-group"]

clone:
  destination: "./repositories"
  strategy: "reset"  # reset, pull, fetch
  concurrent: 5

network:
  profiles:
    - name: "home"
      dns: ["8.8.8.8", "1.1.1.1"]
    - name: "office" 
      proxy: "http://proxy.company.com:8080"
```

---

## ⚙️ 상세 설정 옵션

이 가이드는 gzh-manager-go의 `gzh.yaml` 설정 시스템에 대한 종합적인 문서입니다.

### 프로바이더 설정

#### GitHub 설정
```yaml
providers:
  github:
    token: "${GITHUB_TOKEN}"
    api_url: "https://api.github.com"  # Enterprise의 경우 변경
    organizations: 
      - "org1"
      - "org2"
    exclude_repos:
      - "archived-repo"
      - "private-test-*"
    include_forks: false
    rate_limit:
      requests_per_hour: 5000
      concurrent_requests: 10
```

#### GitLab 설정
```yaml
providers:
  gitlab:
    token: "${GITLAB_TOKEN}"
    api_url: "https://gitlab.com/api/v4"
    groups:
      - "group1"
      - "group2"
    include_subgroups: true
    exclude_archived: true
```

### 클론 설정
```yaml
clone:
  destination: "./repos"
  create_org_dirs: true
  strategy: "reset"
  concurrent: 3
  timeout: "10m"
  git_config:
    user.name: "Your Name"
    user.email: "your.email@example.com"
  ssh_key: "~/.ssh/id_rsa"
```

### 네트워크 환경 설정
```yaml
network:
  auto_switch: true
  profiles:
    - name: "home"
      dns: ["8.8.8.8", "1.1.1.1"]
      routes:
        - destination: "192.168.1.0/24"
          gateway: "192.168.1.1"
    
    - name: "office"
      proxy: "http://proxy.company.com:8080"
      no_proxy: "localhost,127.0.0.1,.company.com"
      dns: ["192.168.10.1"]
      
    - name: "vpn"
      vpn:
        provider: "openvpn"
        config: "/etc/openvpn/client.conf"
        auto_connect: true
```

### 개발 환경 설정
```yaml
development:
  cloud_profiles:
    aws:
      default_region: "ap-northeast-2"
      profiles:
        - name: "dev"
          access_key_id: "${AWS_DEV_ACCESS_KEY}"
          secret_access_key: "${AWS_DEV_SECRET_KEY}"
        - name: "prod"
          role_arn: "arn:aws:iam::123456789012:role/ProductionRole"
    
    gcp:
      default_project: "my-project-dev"
      service_account_key: "${GCP_SERVICE_ACCOUNT_KEY}"
      
  containers:
    docker:
      network: "development"
      compose_files: 
        - "docker-compose.dev.yml"
    kubernetes:
      context: "minikube"
      namespace: "development"
```

## 📚 설정 예제

### 개인 개발자용 설정
```yaml
version: "1.0"
metadata:
  name: "personal-dev"
  
providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations: ["my-username"]

clone:
  destination: "~/Development"
  create_org_dirs: true
  strategy: "pull"
  concurrent: 3
```

### 팀 개발용 설정
```yaml
version: "1.0"
metadata:
  name: "team-development"
  
providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations: 
      - "company-org"
      - "open-source-org"
    exclude_repos:
      - "archived-*"
      - "*-backup"
      
  gitlab:
    token: "${GITLAB_TOKEN}"
    groups: ["internal-projects"]

clone:
  destination: "./team-repos"
  strategy: "reset"
  concurrent: 5
  
network:
  auto_switch: true
  profiles:
    - name: "office"
      proxy: "http://proxy.company.com:8080"
      dns: ["192.168.1.1"]
    - name: "home"
      dns: ["8.8.8.8", "1.1.1.1"]
```

### 엔터프라이즈용 설정
```yaml
version: "1.0"
metadata:
  name: "enterprise-setup"
  organization: "company"
  
providers:
  github:
    api_url: "https://github.company.com/api/v3"
    token: "${GITHUB_ENTERPRISE_TOKEN}"
    organizations: ["platform", "security", "infrastructure"]
    rate_limit:
      requests_per_hour: 10000
      
security:
  allowed_domains: ["*.company.com", "github.company.com"]
  require_ssl: true
  audit_log: "/var/log/gzh/audit.log"
  
monitoring:
  prometheus:
    enabled: true
    port: 9090
  logging:
    level: "info"
    format: "json"
```

## 🔧 환경 변수

### 필수 환경 변수
```bash
# GitHub
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"

# GitLab
export GITLAB_TOKEN="glpat-xxxxxxxxxxxx"

# AWS (선택사항)
export AWS_ACCESS_KEY_ID="AKIAXXXXXXXX"
export AWS_SECRET_ACCESS_KEY="xxxxxxxx"

# 설정 파일 경로 (선택사항)
export GZH_CONFIG_PATH="/path/to/gzh.yaml"
```

### 설정 파일 우선순위
1. `GZH_CONFIG_PATH` 환경 변수로 지정된 경로
2. 현재 디렉토리의 `gzh.yaml` 또는 `gzh.yml`
3. `~/.config/gzh-manager/gzh.yaml`
4. `/etc/gzh-manager/gzh.yaml`

## 🛠️ 문제 해결

### 일반적인 문제

#### 1. 토큰 권한 오류
```bash
# 토큰 검증
gz config validate

# 권한 확인
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user
```

#### 2. 설정 파일 검증
```bash
# 설정 파일 문법 검사
gz config validate

# 상세 설정 정보 출력
gz config show --verbose
```

#### 3. 네트워크 연결 문제
```bash
# 네트워크 프로필 확인
gz net-env status

# 프록시 설정 확인
gz net-env proxy status
```

### 디버깅 모드
```yaml
debug:
  enabled: true
  log_level: "debug"
  log_file: "/tmp/gzh-debug.log"
```

## 📖 추가 참고자료

- [설정 우선순위 시스템](priority-system.md)
- [핫 리로딩 기능](hot-reloading.md)
- [호환성 분석](compatibility-analysis.md)
- [스키마 참조](schemas/)