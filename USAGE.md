<!-- 🚫 AI_MODIFY_PROHIBITED -->
<!-- This file should not be modified by AI agents -->

# gzh-manager-go 사용법

이 문서는 gzh-manager-go (`gz` 명령어)의 실제 사용법을 단계별로 안내합니다.

## 📋 목차

1. [기본 설정](#기본-설정)
2. [리포지토리 대량 클론](#리포지토리-대량-클론)
3. [GitHub 조직 관리](#github-조직-관리)
4. [네트워크 환경 관리](#네트워크-환경-관리)
5. [개발 환경 관리](#개발-환경-관리)
6. [설정 관리](#설정-관리)

## 🔧 기본 설정

### 1. 도구 설치 및 초기 설정

```bash
# 프로젝트 빌드
make build

# 또는 설치
make install

# 통합 설정 파일 생성
gz config init

# 생성된 설정 파일 확인
cat ~/.config/gzh-manager/gzh.yaml
```

### 2. 기본 명령어 구조

```bash
# 도움말 확인
gz --help

# 특정 명령어 도움말
gz bulk-clone --help
gz net-env --help
gz repo-config --help
```

## 📦 리포지토리 대량 클론

### 기본 클론 작업

```bash
# GitHub 조직 전체 클론
gz bulk-clone github -o myorganization -t ~/repos/myorg

# GitLab 그룹 클론
gz bulk-clone gitlab -g mygroup -t ~/repos/gitlab

# Gitea 조직 클론
gz bulk-clone gitea -o myorg -t ~/repos/gitea
```

### 고급 클론 옵션

```bash
# 병렬 처리 (기본값: 10개)
gz bulk-clone github -o myorg -t ~/repos -p 20

# 클론 전략 선택
gz bulk-clone github -o myorg -t ~/repos -s reset   # 기본값: 로컬 변경사항 삭제 후 동기화
gz bulk-clone github -o myorg -t ~/repos -s pull    # 병합 시도
gz bulk-clone github -o myorg -t ~/repos -s fetch   # 원격 정보만 가져오기

# 중단된 작업 재개
gz bulk-clone github -o myorg -t ~/repos --resume

# 프라이빗 저장소 포함 (토큰 필요)
export GITHUB_TOKEN=your_token
gz bulk-clone github -o myorg -t ~/repos --private
```

### 설정 파일 사용

```bash
# 설정 파일 생성
gz bulk-clone config init

# 설정 파일을 사용한 클론
gz bulk-clone github --use-config -o myorg

# 특정 설정 파일 사용
gz bulk-clone github -c /path/to/config.yaml -o myorg
```

### 상태 관리

```bash
# 저장된 상태 목록 확인
gz bulk-clone state list

# 특정 상태 상세 정보
gz bulk-clone state show -p github -o myorg

# 상태 정리
gz bulk-clone state clean -p github -o myorg
gz bulk-clone state clean --all
```

## 🏢 GitHub 조직 관리

### 리포지토리 설정 관리

```bash
# 현재 조직의 리포지토리 설정 확인
gz repo-config list -o myorg

# 설정 파일 생성
gz repo-config init -o myorg

# 설정 적용 (미리보기)
gz repo-config apply --config repo-config.yaml --dry-run

# 설정 적용 (실제 적용)
gz repo-config apply --config repo-config.yaml

# 정책 준수 감사
gz repo-config audit --config repo-config.yaml
```

### 정책 템플릿 사용

```bash
# 보안 강화 템플릿 적용
gz repo-config template apply --type security -o myorg

# 오픈소스 템플릿 적용
gz repo-config template apply --type opensource -o myorg

# 엔터프라이즈 템플릿 적용
gz repo-config template apply --type enterprise -o myorg
```

## 🌐 네트워크 환경 관리

### 네트워크 프로필 설정

```bash
# 네트워크 프로필 설정 파일 생성
gz net-env switch --init

# 사용 가능한 네트워크 프로필 확인
gz net-env switch --list

# 현재 네트워크 상태 확인
gz net-env status

# 상세 네트워크 정보 확인
gz net-env status --verbose
```

### 네트워크 환경 전환

```bash
# 특정 네트워크 프로필로 전환
gz net-env switch home
gz net-env switch office
gz net-env switch public

# 실행 전 미리보기 (dry-run)
gz net-env switch office --dry-run

# 강제 전환 (조건 확인 건너뛰기)
gz net-env switch office --force

# 상세 로그와 함께 전환
gz net-env switch office --verbose
```

### 네트워크 프로필 구성 예시

```bash
# 홈 네트워크 프로필로 전환
# - VPN 연결 해제
# - DNS를 홈 라우터로 설정
# - 프록시 비활성화
gz net-env switch home

# 오피스 네트워크 프로필로 전환
# - 회사 VPN 연결
# - 회사 DNS 서버 설정
# - 프록시 설정 적용
# - 회사 내부 호스트 파일 추가
gz net-env switch office

# 공용 WiFi 프로필로 전환
# - 개인 VPN 연결 (보안)
# - 안전한 DNS 서버 사용
# - 프록시 비활성화
gz net-env switch public
```

### 네트워크 프로필 설정 파일 예시

```yaml
# ~/.gz/network-profiles.yaml
default: "home"

profiles:
  - name: "home"
    description: "홈 네트워크 설정"
    dns:
      servers: ["192.168.1.1", "1.1.1.1"]
      method: "resolvectl"
    proxy:
      clear: true
    vpn:
      disconnect: ["office-vpn"]
    scripts:
      post_switch: ["echo '홈 네트워크로 전환 완료'"]

  - name: "office"
    description: "오피스 네트워크 설정"
    vpn:
      connect:
        - name: "office-vpn"
          type: "networkmanager"
    dns:
      servers: ["10.0.0.1", "10.0.0.2"]
    proxy:
      http: "http://proxy.company.com:8080"
      https: "http://proxy.company.com:8080"
    hosts:
      add:
        - ip: "10.0.1.100"
          host: "intranet.company.com"
```

### 네트워크 구성 요소별 상태 확인

```bash
# 현재 DNS 설정 상태 확인
gz net-env status --verbose | grep -A 5 "DNS Configuration"

# 현재 VPN 연결 상태 확인
gz net-env status --verbose | grep -A 5 "VPN Connections"

# 현재 프록시 설정 확인
gz net-env status --verbose | grep -A 5 "Proxy Configuration"

# 전체 네트워크 인터페이스 정보
gz net-env status --verbose | grep -A 10 "Network Interfaces"
```

> **💡 권장사항**: 개별 네트워크 구성 요소를 직접 수정하는 대신, 네트워크 프로필을 사용하여 일괄적으로 관리하는 것을 권장합니다. 이렇게 하면 설정의 일관성을 유지하고 복잡한 네트워크 환경 간 전환을 쉽게 할 수 있습니다.

## 🏠 개발 환경 관리

### 패키지 관리자 업데이트

```bash
# 모든 패키지 관리자 업데이트
gz always-latest all

# 특정 패키지 관리자만 업데이트
gz always-latest asdf
gz always-latest brew
gz always-latest sdkman

# 마이너 버전만 업데이트
gz always-latest asdf --strategy minor-latest
```

### 개발 환경 설정 백업/복원

```bash
# AWS 설정 백업
gz dev-env backup aws --description "production aws config"

# Docker 설정 백업
gz dev-env backup docker --description "current docker setup"

# 설정 복원
gz dev-env restore aws --id backup_id

# 백업 목록 확인
gz dev-env list aws
```

### JetBrains IDE 관리

```bash
# 설치된 IDE 목록 확인
gz ide list

# IDE 설정 모니터링
gz ide monitor

# 설정 동기화 문제 수정
gz ide fix-sync

# 특정 IDE 설정 확인
gz ide status --ide IntelliJ
```

## ⚙️ 설정 관리

### 통합 설정 시스템

```bash
# 설정 파일 생성
gz config init

# 설정 파일 검증
gz config validate

# 설정 파일 위치 확인
gz config show --paths

# 특정 설정 파일 사용
gz config validate --config /path/to/gzh.yaml
```

### 설정 마이그레이션

```bash
# 설정 파일 검증
gz synclone config validate --file synclone.yaml

# 설정 파일 생성
gz synclone config generate
```

### 설정 우선순위 테스트

```bash
# CLI 플래그가 최우선 (다른 설정 무시)
gz bulk-clone github -o myorg --parallel 20

# 환경 변수 설정
export GITHUB_TOKEN=your_token
export GZH_PARALLEL=15

# 설정 파일 사용 (환경 변수와 CLI 플래그에 의해 재정의됨)
gz bulk-clone github --use-config -o myorg
```

## 🔄 실제 사용 시나리오

### 시나리오 1: 사무실 네트워크 환경 설정

```bash
# 1. WiFi 설정 파일에 사무실 네트워크 추가
gz net-env wifi config init

# 2. 사무실 네트워크 액션 설정
gz net-env actions config init

# 3. 사무실 WiFi 연결 시 자동 실행될 액션 설정
# ~/.gz/wifi-hooks.yaml 파일 편집:
# actions:
#   - name: "office-setup"
#     conditions:
#       ssid: ["Office-WiFi"]
#       state: ["connected"]
#     commands:
#       - "gz net-env actions vpn connect --name office"
#       - "gz net-env actions dns set --servers 10.0.0.1,10.0.0.2"
#       - "gz net-env actions proxy set --http http://proxy.company.com:8080"

# 4. 모니터링 시작
gz net-env wifi monitor --daemon
```

### 시나리오 2: 대규모 조직 리포지토리 관리

```bash
# 1. 통합 설정 파일 생성
gz config init

# 2. 조직 리포지토리 대량 클론
gz bulk-clone github -o large-org -t ~/repos/large-org -p 30 --resume

# 3. 리포지토리 설정 정책 적용
gz repo-config apply --config enterprise-policy.yaml --dry-run
gz repo-config apply --config enterprise-policy.yaml

# 4. 정책 준수 감사
gz repo-config audit --config enterprise-policy.yaml
```

### 시나리오 3: 개발 환경 완전 자동화

```bash
# 1. 모든 설정 파일 초기화
gz config init
gz net-env wifi config init
gz net-env actions config init

# 2. 개발 환경 백업
gz dev-env backup aws --description "current setup"
gz dev-env backup docker --description "current docker config"

# 3. 패키지 관리자 업데이트
gz always-latest all

# 4. 네트워크 환경 자동화 시작
gz net-env wifi monitor --daemon

# 5. IDE 설정 모니터링
gz ide monitor
```

## 🐛 문제 해결

### 일반적인 문제

```bash
# 설정 파일 검증
gz config validate

# 상세 로그 확인
gz bulk-clone github -o myorg -t ~/repos --verbose

# 드라이런 모드로 테스트
gz net-env actions run --dry-run
```

### 권한 문제

```bash
# 일부 네트워크 액션은 sudo 권한 필요
sudo gz net-env actions vpn connect --name office
sudo gz net-env actions dns set --servers 1.1.1.1,1.0.0.1
```

### 설정 파일 문제

```bash
# 설정 파일 위치 확인
gz config show --paths

# 설정 파일 재생성
gz config init --force

# 마이그레이션 문제 시 백업에서 복원
# 백업에서 복원 (수동으로 파일 복사)
```

## 📚 추가 자료

- [상세한 기능 목록](FEATURES.md)
- [설정 우선순위 가이드](docs/configuration-priority.md)
- [네트워크 액션 문서](docs/network-actions.md)
- [GitHub 조직 관리 가이드](docs/repo-config-user-guide.md)
- [정책 템플릿 예제](docs/repo-config-policy-examples.md)

---

> 💡 **팁**: 모든 명령어는 `--help` 옵션을 지원합니다. 자세한 옵션은 `gz <command> --help`를 참고하세요.
