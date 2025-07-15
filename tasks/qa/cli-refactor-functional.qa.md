# title: CLI 중심 재편 기능 QA 시나리오

## related_tasks
- /tasks/done/20250712__cli_focused_refactor__DONE_20250712.md

## purpose
데몬 모드 제거 후 CLI 중심으로 재편된 네트워크 관리 기능이 정상 동작하는지 검증

## scenarios

### 1. 네트워크 관리 명령어 검증
1. **네트워크 상태 확인 기능**
   ```bash
   gz net-env status
   ```
   - 현재 VPN 연결 상태 표시 확인
   - DNS/프록시 설정 정보 출력 검증
   - 네트워크 연결성 테스트 결과 확인
   - 출력 형식이 사용자 친화적인지 검증

2. **네트워크 프로필 전환 기능**
   ```bash
   gz net-env switch --profile office
   gz net-env switch --profile home
   gz net-env switch --profile mobile
   ```
   - VPN, DNS, 프록시, 호스트 파일 일괄 전환 확인
   - 프로필별 설정 템플릿 정상 적용 검증
   - 전환 후 `gz net-env status`로 변경 사항 확인

3. **Hook 시스템 검증**
   ```bash
   gz net-env switch --profile office --hook /path/to/script.sh
   ```
   - 네트워크 변경 시 사용자 정의 스크립트 실행 확인
   - Hook 실행 성공/실패 로그 확인
   - 여러 Hook 스크립트 순차 실행 검증

### 2. 설정 파일 시스템 검증
1. **YAML 설정 파일 구조**
   - 데몬 관련 설정이 제거되었는지 확인
   - 네트워크 프로필 설정만 포함되는지 검증
   - YAML 스키마 유효성 검사
   - 기존 설정 파일 호환성 테스트

2. **설정 파일 로딩 테스트**
   ```bash
   gz net-env --config /custom/path/config.yaml status
   ```
   - 커스텀 설정 파일 경로 지원 확인
   - 설정 파일 오류 시 명확한 에러 메시지 출력
   - 기본 설정 파일 fallback 동작 검증

### 3. 데몬 모드 제거 검증
1. **백그라운드 프로세스 확인**
   ```bash
   ps aux | grep gz
   systemctl status gzh-manager || echo "데몬 서비스 없음 확인"
   ```
   - 백그라운드에서 실행되는 gzh-manager 프로세스가 없는지 확인
   - systemd/launchd 서비스가 제거되었는지 검증
   - 시스템 리소스 사용량 감소 확인

2. **실시간 감지 기능 제거 확인**
   - 네트워크 변경 시 자동 감지/전환이 발생하지 않는지 확인
   - fsnotify 관련 코드가 제거되었는지 검증
   - 이벤트 기반 자동화가 비활성화되었는지 확인

### 4. 워크플로우 시나리오 테스트
1. **일반적인 사용 패턴**
   ```bash
   # 1. 현재 상태 확인
   gz net-env status
   
   # 2. 작업 환경으로 전환
   gz net-env switch --profile office
   
   # 3. 상태 재확인
   gz net-env status
   
   # 4. 집에서 작업할 때 전환
   gz net-env switch --profile home
   ```

2. **오류 복구 시나리오**
   ```bash
   # 잘못된 프로필 이름 사용
   gz net-env switch --profile nonexistent
   
   # 권한 부족 시나리오
   sudo chmod 000 /path/to/config/file
   gz net-env status
   
   # 네트워크 연결 오류 시나리오
   gz net-env switch --profile vpn-disconnected
   ```

### 5. 크로스 플랫폼 호환성
1. **Linux 환경 테스트**
   - systemd-resolve DNS 설정 변경
   - NetworkManager 프로필 전환
   - iptables 프록시 설정

2. **macOS 환경 테스트**
   - networksetup 명령어를 통한 DNS 변경
   - 시스템 프록시 설정 변경
   - /etc/hosts 파일 수정

3. **Windows 환경 테스트**
   - netsh 명령어를 통한 네트워크 설정
   - 레지스트리 프록시 설정 변경
   - hosts 파일 관리자 권한 처리

## expected_results
- **CLI 명령어**: 모든 `gz net-env` 명령어가 정상 동작
- **데몬 제거**: 백그라운드 프로세스 및 시스템 서비스 완전 제거
- **설정 관리**: 단순화된 YAML 구조로 설정 관리
- **워크플로우**: 사용자 명령 기반 네트워크 전환 워크플로우 정상 동작
- **오류 처리**: 명확하고 도움이 되는 에러 메시지 제공
- **성능**: 시스템 리소스 사용량 감소 및 응답성 향상
- **호환성**: Linux, macOS, Windows 환경에서 정상 동작

## test_environment
- **운영체제**: Linux (Ubuntu/CentOS), macOS, Windows 10/11
- **네트워크**: VPN 서버, 프록시 서버, 다양한 DNS 서버
- **권한**: 관리자 권한 및 일반 사용자 권한 시나리오
- **설정**: 다양한 네트워크 프로필 구성 파일

## automation_level
- **자동화 가능**: 명령어 실행 테스트, 설정 파일 유효성 검사, 프로세스 존재 확인
- **수동 검증 필요**: 실제 네트워크 설정 변경, 사용자 워크플로우 경험, 크로스 플랫폼 동작

## tags
[qa], [functional], [cli], [networking], [manual], [cross-platform]