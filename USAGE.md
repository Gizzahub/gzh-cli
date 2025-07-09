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

### 데몬 모니터링
```bash
# 시스템 데몬 목록 확인
gz net-env daemon list

# 네트워크 관련 서비스만 확인
gz net-env daemon list --network-services

# 특정 서비스 상태 확인
gz net-env daemon status --service ssh

# 실시간 모니터링
gz net-env daemon monitor --network-services

# 서비스 관리
gz net-env daemon manage --service nginx --action start
gz net-env daemon manage --service nginx --action stop
gz net-env daemon manage --service nginx --action restart
```

### WiFi 변경 감지 및 자동화
```bash
# WiFi 설정 파일 생성
gz net-env wifi config init

# 현재 WiFi 상태 확인
gz net-env wifi status

# WiFi 변경 모니터링 시작
gz net-env wifi monitor

# 데몬 모드로 백그라운드 실행
gz net-env wifi monitor --daemon

# 설정 파일 검증
gz net-env wifi config validate

# 설정 파일 내용 확인
gz net-env wifi config show
```

### 네트워크 액션 실행
```bash
# 네트워크 액션 설정 파일 생성
gz net-env actions config init

# 모든 네트워크 액션 실행
gz net-env actions run

# 드라이런 모드 (실제 변경 없이 테스트)
gz net-env actions run --dry-run

# 자세한 로그와 함께 실행
gz net-env actions run --verbose
```

### VPN 관리
```bash
# VPN 연결
gz net-env actions vpn connect --name office --type networkmanager
gz net-env actions vpn connect --name home --type openvpn --config /etc/openvpn/home.conf
gz net-env actions vpn connect --name mobile --type wireguard

# VPN 해제
gz net-env actions vpn disconnect --name office

# VPN 상태 확인
gz net-env actions vpn status
```

### DNS 설정 관리
```bash
# DNS 서버 변경
gz net-env actions dns set --servers 1.1.1.1,1.0.0.1
gz net-env actions dns set --servers 8.8.8.8,8.8.4.4 --interface wlan0

# 현재 DNS 설정 확인
gz net-env actions dns status

# DNS 설정 초기화
gz net-env actions dns reset
```

### 프록시 설정 관리
```bash
# 프록시 설정
gz net-env actions proxy set --http http://proxy.company.com:8080
gz net-env actions proxy set --https https://proxy.company.com:8080 --socks socks5://proxy.company.com:1080

# 프록시 설정 제거
gz net-env actions proxy clear

# 현재 프록시 상태 확인
gz net-env actions proxy status
```

### 호스트 파일 관리
```bash
# 호스트 엔트리 추가
gz net-env actions hosts add --ip 192.168.1.100 --host server.local
gz net-env actions hosts add --ip 10.0.0.50 --host dev-server.local

# 호스트 엔트리 제거
gz net-env actions hosts remove --host server.local

# 호스트 파일 내용 확인
gz net-env actions hosts show
```

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
# 기존 설정을 통합 설정으로 마이그레이션
gz migrate config --from bulk-clone.yaml --to gzh.yaml

# 배치 마이그레이션
gz migrate config --batch --auto

# 드라이런 마이그레이션 (실제 변경 없이 테스트)
gz migrate config --dry-run --from bulk-clone.yaml
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
gz migrate config --restore --backup-id backup_timestamp
```

## 📚 추가 자료

- [상세한 기능 목록](FEATURES.md)
- [설정 우선순위 가이드](docs/configuration-priority.md)
- [네트워크 액션 문서](docs/network-actions.md)
- [GitHub 조직 관리 가이드](docs/repo-config-user-guide.md)
- [정책 템플릿 예제](docs/repo-config-policy-examples.md)

---

> 💡 **팁**: 모든 명령어는 `--help` 옵션을 지원합니다. 자세한 옵션은 `gz <command> --help`를 참고하세요.