# 데몬 모드에서 CLI 모드로 마이그레이션 가이드

이 가이드는 gzh-manager-go의 네트워크 환경 관리가 데몬 기반 자동 모니터링에서 CLI 기반 수동 전환으로 변경됨에 따라, 기존 사용자가 새로운 방식으로 원활하게 전환할 수 있도록 도움을 제공합니다.

## 📋 변경 사항 개요

### ❌ 제거된 기능

- **WiFi 변경 자동 감지**: 실시간 WiFi 변경 모니터링 기능 제거
- **백그라운드 데몬 모드**: `gz net-env wifi monitor --daemon` 명령어 제거
- **자동 네트워크 전환**: WiFi 변경 시 자동으로 네트워크 설정 변경하는 기능 제거
- **복잡한 이벤트 시스템**: 네트워크 이벤트 기반 자동화 시스템 제거

### ✅ 새로 추가된 기능

- **네트워크 프로필 시스템**: YAML 기반 네트워크 프로필 관리
- **수동 전환 명령어**: `gz net-env switch` 명령어를 통한 즉시 전환
- **상태 확인 도구**: `gz net-env status` 명령어로 현재 네트워크 상태 조회
- **단순화된 설정**: 복잡한 이벤트 설정 대신 간단한 프로필 구성

## 🔄 마이그레이션 단계

### 1단계: 기존 설정 백업

```bash
# 기존 WiFi 설정 백업 (있다면)
cp ~/.config/gzh-manager/wifi-config.yaml ~/.config/gzh-manager/wifi-config.yaml.backup

# 기존 액션 설정 백업 (있다면)
cp ~/.config/gzh-manager/actions-config.yaml ~/.config/gzh-manager/actions-config.yaml.backup

# 기존 통합 설정 백업
cp ~/.config/gzh-manager/gzh.yaml ~/.config/gzh-manager/gzh.yaml.backup
```

### 2단계: 새로운 네트워크 프로필 설정 생성

```bash
# 새로운 네트워크 프로필 설정 파일 생성
gz net-env switch --init

# 생성된 설정 파일 확인
cat ~/.gz/network-profiles.yaml
```

### 3단계: 기존 설정을 새로운 프로필로 변환

#### 기존 WiFi 설정을 프로필로 변환

**이전 방식** (`wifi-config.yaml`):

```yaml
known_networks:
  "Home-WiFi":
    ssid: "Home-WiFi"
    dns_servers:
      - "192.168.1.1"
      - "1.1.1.1"
    on_connect:
      - "update-dns"
  "Office-WiFi":
    ssid: "Office-WiFi"
    vpn_config: "work-vpn"
    dns_servers:
      - "10.0.0.1"
    on_connect:
      - "connect-vpn"
```

**새로운 방식** (`~/.gz/network-profiles.yaml`):

```yaml
default: "home"

profiles:
  - name: "home"
    description: "Home network configuration"
    dns:
      servers:
        - "192.168.1.1"
        - "1.1.1.1"
      method: "resolvectl"
    proxy:
      clear: true
    vpn:
      disconnect:
        - "work-vpn"
    scripts:
      post_switch:
        - "echo 'Switched to home network'"

  - name: "office"
    description: "Office network with VPN"
    vpn:
      connect:
        - name: "work-vpn"
          type: "networkmanager"
    dns:
      servers:
        - "10.0.0.1"
        - "8.8.8.8"
      method: "resolvectl"
    scripts:
      pre_switch:
        - "echo 'Connecting to office network...'"
      post_switch:
        - "echo 'Connected to office network'"
```

### 4단계: 새로운 워크플로우 적용

#### 이전 워크플로우

```bash
# 데몬 시작 (자동 모니터링)
gz net-env wifi monitor --daemon

# 백그라운드에서 자동 실행됨
# 사용자가 직접 관여할 필요 없음
```

#### 새로운 워크플로우

```bash
# 사용 가능한 프로필 확인
gz net-env switch --list

# 현재 네트워크 상태 확인
gz net-env status

# 필요시 수동으로 네트워크 프로필 전환
gz net-env switch home    # 집에서 작업할 때
gz net-env switch office  # 사무실에서 작업할 때
gz net-env switch public  # 카페/공공장소에서 작업할 때
```

### 5단계: 기존 시스템 서비스 정리

데몬 모드를 사용했다면 관련 서비스를 정리해야 합니다:

```bash
# systemd 서비스 중지 (Linux)
sudo systemctl stop gzh-manager-netenv
sudo systemctl disable gzh-manager-netenv

# launchd 서비스 중지 (macOS)
launchctl unload ~/Library/LaunchAgents/com.gizzahub.netenv.plist

# 기존 데몬 프로세스 확인 및 종료
ps aux | grep gzh-manager
# 필요시 kill <PID>
```

## 🛠️ 실전 사용 시나리오

### 시나리오 1: 재택근무 → 사무실 출근

**이전 방식**: 자동으로 WiFi 변경 감지하여 자동 전환

```bash
# 자동으로 실행됨 - 사용자 개입 없음
```

**새로운 방식**: 수동으로 프로필 전환

```bash
# 사무실 도착 후 수동으로 전환
gz net-env switch office

# 실행 결과 확인
gz net-env status --verbose
```

### 시나리오 2: 카페에서 작업

**이전 방식**:

```bash
# WiFi 변경 감지 후 자동으로 공용 WiFi 설정 적용
```

**새로운 방식**:

```bash
# 카페 WiFi 연결 후 보안 프로필로 전환
gz net-env switch public

# VPN 연결 상태 확인
gz net-env status | grep VPN
```

### 시나리오 3: 문제 발생시 디버깅

**이전 방식**: 데몬 로그 확인 필요

```bash
# 로그 파일 확인 필요
tail -f /var/log/gzh-manager-netenv.log
```

**새로운 방식**: 즉시 상태 확인 가능

```bash
# 현재 상태 즉시 확인
gz net-env status --verbose

# 드라이런으로 문제 진단
gz net-env switch office --dry-run --verbose
```

## 🎯 장점과 권장사항

### ✅ 새로운 방식의 장점

1. **예측 가능성**: 사용자가 명시적으로 명령어를 실행하므로 예상치 못한 네트워크 변경 없음
2. **디버깅 용이성**: 문제 발생시 즉시 상태 확인 및 문제 진단 가능
3. **리소스 효율성**: 백그라운드 데몬이 없어 시스템 리소스 절약
4. **설정 단순화**: 복잡한 이벤트 설정 대신 간단한 프로필 구성

### 📋 권장 워크플로우

1. **일과 시작시**: `gz net-env switch office`
2. **점심시간 외부**: `gz net-env switch public`
3. **재택근무 전환**: `gz net-env switch home`
4. **문제 발생시**: `gz net-env status --verbose`

### 🔧 자동화 대안

완전 자동화가 필요한 경우 스크립트나 별칭을 활용할 수 있습니다:

```bash
# ~/.bashrc 또는 ~/.zshrc에 추가
alias work='gz net-env switch office && echo "Office mode activated"'
alias home='gz net-env switch home && echo "Home mode activated"'
alias cafe='gz net-env switch public && echo "Public WiFi mode activated"'

# 사용 예시
work  # 사무실 모드로 전환
home  # 홈 모드로 전환
cafe  # 공용 WiFi 모드로 전환
```

## 🆘 문제 해결

### 일반적인 문제

**Q: 네트워크 전환이 안 됩니다**

```bash
# 1. 설정 파일 확인
gz net-env switch --list

# 2. 권한 문제 확인
gz net-env switch office --dry-run --verbose

# 3. 네트워크 상태 확인
gz net-env status --verbose
```

**Q: VPN 연결이 실패합니다**

```bash
# VPN 연결 상태 확인
systemctl status openvpn@myconfig  # OpenVPN
nmcli connection show --active     # NetworkManager
wg show                           # WireGuard
```

**Q: DNS 설정이 적용되지 않습니다**

```bash
# DNS 설정 확인
resolvectl status
# 또는
cat /etc/resolv.conf
```

### 지원 및 도움

문제가 계속 발생하면 다음을 확인하세요:

1. [GitHub Issues](https://github.com/gizzahub/gzh-manager-go/issues)
2. [사용자 가이드](../USAGE.md#네트워크-환경-관리)
3. [프로젝트 문서](../README.md)

## 📝 요약

데몬 기반 자동 모니터링에서 CLI 기반 수동 전환으로의 변경은 더 예측 가능하고 안정적인 네트워크 관리를 제공합니다. 초기 설정 후에는 간단한 명령어만으로 네트워크 환경을 효율적으로 관리할 수 있습니다.

변경된 워크플로우에 익숙해지는 데 시간이 걸릴 수 있지만, 더 안정적이고 문제 해결이 쉬운 환경을 제공합니다.
