# Webhook Monitoring Dashboard

이 문서는 GitHub 웹훅 상태 모니터링 대시보드의 기능과 사용법을 설명합니다.

## 개요

웹훅 모니터링 대시보드는 GitHub 웹훅의 상태와 성능을 실시간으로 모니터링하는 종합적인 솔루션입니다. 웹 기반 대시보드, REST API, CLI 도구를 제공하여 웹훅 관리를 효율화합니다.

## 주요 기능

### 🔍 실시간 모니터링
- 웹훅 상태 실시간 추적 (Healthy, Degraded, Unhealthy)
- 배송 성공률 및 오류율 모니터링
- 응답 시간 및 업타임 추적
- 조직별/리포지토리별 메트릭 수집

### 📊 시각화 대시보드
- 직관적인 웹 기반 인터페이스
- 실시간 메트릭 및 차트
- 상태 분포 및 트렌드 분석
- 조직별 필터링 및 드릴다운

### 🚨 알림 시스템
- 임계값 기반 자동 알림
- 다양한 알림 유형 (오류율, 응답시간, 연속 실패 등)
- 심각도별 알림 분류 (Info, Warning, Error, Critical)
- 알림 확인 및 해결 추적

### 🛠 관리 도구
- REST API를 통한 프로그래밍 방식 접근
- CLI 도구로 명령줄에서 상태 확인
- 웹훅 설정 및 구성 관리
- 히스토리 및 트렌드 데이터 보관

## 시스템 구성

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Dashboard │    │   REST API      │    │   CLI Tool      │
│                 │    │                 │    │                 │
│ - Real-time UI  │    │ - HTTP endpoints│    │ - Status check  │
│ - Charts/Graphs │    │ - JSON responses│    │ - Alert mgmt    │
│ - Filtering     │    │ - Authentication│    │ - Monitoring    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │ Webhook Monitor │
                    │                 │
                    │ - Health checks │
                    │ - Metrics       │
                    │ - Alerting      │
                    │ - Storage       │
                    └─────────────────┘
                                 │
                    ┌─────────────────┐
                    │  GitHub API     │
                    │                 │
                    │ - Webhook data  │
                    │ - Delivery logs │
                    │ - Configuration │
                    └─────────────────┘
```

## 설치 및 실행

### 1. 대시보드 서버 시작

```bash
# 기본 설정으로 시작
webhook-dashboard start

# 사용자 정의 설정
webhook-dashboard start \
  --host 0.0.0.0 \
  --port 8080 \
  --token $GITHUB_TOKEN \
  --check-interval 5m
```

### 2. 웹 대시보드 접속

브라우저에서 `http://localhost:8080`에 접속하여 대시보드를 확인할 수 있습니다.

### 3. CLI를 통한 상태 확인

```bash
# 전체 웹훅 상태 확인
webhook-dashboard status

# 특정 조직 필터링
webhook-dashboard status --org myorg

# JSON 형식으로 출력
webhook-dashboard status --format json

# 상세 메트릭 포함
webhook-dashboard status --show-metrics
```

### 4. 알림 관리

```bash
# 활성 알림 목록
webhook-dashboard alerts list

# 심각도별 필터링
webhook-dashboard alerts list --severity error

# 알림 확인
webhook-dashboard alerts ack alert-12345
```

## API 엔드포인트

### 대시보드 데이터
- `GET /api/v1/dashboard` - 전체 대시보드 개요
- `GET /api/v1/dashboard/organization/{org}` - 조직별 대시보드

### 웹훅 관리
- `GET /api/v1/webhooks` - 웹훅 목록
- `GET /api/v1/webhooks/{id}` - 특정 웹훅 상세정보
- `GET /api/v1/webhooks/{id}/status` - 웹훅 상태
- `GET /api/v1/webhooks/{id}/history` - 웹훅 히스토리

### 메트릭 및 분석
- `GET /api/v1/metrics` - 전체 메트릭
- `GET /api/v1/metrics/organization/{org}` - 조직별 메트릭
- `GET /api/v1/metrics/trends` - 트렌드 데이터

### 알림 관리
- `GET /api/v1/alerts` - 알림 목록
- `GET /api/v1/alerts/active` - 활성 알림
- `POST /api/v1/alerts/{id}/acknowledge` - 알림 확인

### 시스템 상태
- `GET /api/v1/health` - 헬스 체크
- `GET /api/v1/status` - 시스템 상태

## 설정 옵션

### 모니터 설정
```yaml
webhook_monitor:
  check_interval: "5m"           # 상태 확인 간격
  health_check_timeout: "30s"    # 헬스 체크 타임아웃
  retention_period: "24h"        # 데이터 보관 기간
  enable_notifications: true     # 알림 활성화
  max_history_size: 1000        # 히스토리 최대 크기

  alert_thresholds:
    error_rate: 10.0             # 오류율 임계값 (%)
    response_time: "5s"          # 응답시간 임계값
    failure_count: 5             # 연속 실패 횟수
    delivery_failure_age: "1h"   # 배송 실패 허용 시간
```

### API 서버 설정
```yaml
dashboard_api:
  host: "0.0.0.0"               # 바인딩 호스트
  port: 8080                    # 포트 번호
  enable_cors: true             # CORS 활성화
  request_timeout: "30s"        # 요청 타임아웃
  enable_auth: false            # 인증 활성화
  auth_token: ""                # 인증 토큰
```

## 모니터링 메트릭

### 전역 메트릭
- **Total Webhooks**: 전체 웹훅 수
- **Active Webhooks**: 활성 웹훅 수
- **Healthy/Unhealthy**: 상태별 웹훅 수
- **Total Deliveries**: 전체 배송 횟수
- **Success Rate**: 성공율
- **Average Response Time**: 평균 응답시간
- **Active Alerts**: 활성 알림 수

### 웹훅별 메트릭
- **Status**: 현재 상태 (Healthy, Degraded, Unhealthy)
- **Uptime**: 업타임 백분율
- **Error Rate**: 오류율
- **Response Time**: 평균 응답시간
- **Last Delivery**: 마지막 배송 시간
- **Consecutive Failures**: 연속 실패 횟수

### 조직별 메트릭
- **Organization Summary**: 조직별 요약 통계
- **Repository Breakdown**: 리포지토리별 분석
- **Event Type Distribution**: 이벤트 유형별 분포
- **Performance Trends**: 성능 트렌드

## 알림 유형

### 1. High Error Rate
- **설명**: 웹훅 오류율이 임계값을 초과했을 때
- **임계값**: 설정 가능한 백분율 (기본: 10%)
- **심각도**: Warning 또는 Error

### 2. Slow Response
- **설명**: 응답시간이 임계값을 초과했을 때
- **임계값**: 설정 가능한 시간 (기본: 5초)
- **심각도**: Warning

### 3. Consecutive Failures
- **설명**: 연속 배송 실패가 임계값을 초과했을 때
- **임계값**: 설정 가능한 횟수 (기본: 5회)
- **심각도**: Error

### 4. Configuration Issue
- **설명**: 웹훅 설정에 문제가 발견되었을 때
- **심각도**: Warning

### 5. Delivery Failure
- **설명**: 배송 실패가 지속되고 있을 때
- **심각도**: Error

### 6. Endpoint Down
- **설명**: 웹훅 엔드포인트가 응답하지 않을 때
- **심각도**: Critical

## 트러블슈팅

### 일반적인 문제

1. **웹훅이 모니터링되지 않음**
   - GitHub 토큰 권한 확인
   - 조직/리포지토리 접근 권한 확인
   - 네트워크 연결 상태 확인

2. **대시보드에 데이터가 표시되지 않음**
   - 모니터링 서비스 실행 상태 확인
   - API 엔드포인트 응답 확인
   - 브라우저 콘솔 오류 확인

3. **알림이 작동하지 않음**
   - 알림 설정 활성화 확인
   - 임계값 설정 확인
   - 로그에서 오류 메시지 확인

### 로그 확인

```bash
# 대시보드 로그 확인 (실행 중일 때)
webhook-dashboard start --verbose

# 시스템 상태 확인
webhook-dashboard status --show-metrics
```

## 개발 및 확장

### 커스텀 알림 추가
웹훅 모니터에 새로운 알림 유형을 추가하려면:

1. `WebhookAlertType`에 새 상수 추가
2. 알림 처리 로직 구현
3. 대시보드 UI 업데이트

### API 확장
새로운 API 엔드포인트를 추가하려면:

1. `WebhookDashboardAPI`에 핸들러 추가
2. 라우팅 설정 업데이트
3. 프론트엔드 연동

### 메트릭 확장
새로운 메트릭을 추가하려면:

1. `WebhookMetrics` 구조체 확장
2. 수집 로직 구현
3. 대시보드 표시 로직 추가

## 보안 고려사항

1. **API 인증**: 프로덕션 환경에서는 API 인증을 활성화하세요
2. **네트워크 보안**: 방화벽 설정으로 접근을 제한하세요
3. **토큰 관리**: GitHub 토큰을 안전하게 보관하세요
4. **HTTPS**: 프로덕션에서는 HTTPS를 사용하세요

## 성능 최적화

1. **캐싱**: 자주 조회되는 데이터는 캐싱을 활용하세요
2. **배치 처리**: 대량의 웹훅 처리 시 배치 방식을 사용하세요
3. **인덱싱**: 데이터베이스 사용 시 적절한 인덱스를 설정하세요
4. **리소스 모니터링**: CPU, 메모리 사용량을 모니터링하세요

## 라이센스

이 소프트웨어는 MIT 라이센스 하에 배포됩니다.